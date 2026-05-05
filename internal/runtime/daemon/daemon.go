package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/iorubs/smithy-cli/internal/config"
	"github.com/iorubs/smithy-cli/internal/runtime"
	"github.com/iorubs/smithy-cli/internal/runtime/ipc"
)

// shutdownGrace bounds how long Run waits for Launch to return after ctx is cancelled.
const shutdownGrace = 5 * time.Second

// restartDelay throttles auto-restart so a crash-looping service
// can't fill the log file with thousands of identical errors per
// second. Cheap, simple, and enough to keep logs readable.
const restartDelay = time.Second

// SpawnTimeout bounds how long an `up` command waits for the daemon's
// socket to come online when it has to spawn the daemon itself.
const SpawnTimeout = 10 * time.Second

// StackMeta describes a backgrounded stack. Written to Paths.Meta on
// Run and read by ls/down for display and config recovery. The
// `Chats` map is mutated by `smithy agent chat` while the daemon is
// running to persist a2a contextIDs across CLI invocations; it is
// cleared together with the rest of the file on daemon shutdown,
// which matches the lifetime of in-memory agent sessions.
type StackMeta struct {
	Name       string            `json:"name"`
	ConfigPath string            `json:"config_path"`
	PID        int               `json:"pid"`
	StartedAt  string            `json:"started_at"`
	Chats      map[string]string `json:"chats,omitempty"`
}

// Run is the body of the hidden `__daemon__` Kong subcommand. It
// loads the stack file, translates it, opens the per-stack socket,
// serves /status, and runs the launcher. It blocks until ctx is
// cancelled or Launch returns. When startAll is true every service
// in the plan is started immediately; otherwise services only start
// when an explicit /start request arrives.
func Run(ctx context.Context, name, stackPath string, startAll bool, logLevel string) error {
	// The daemon writes to a log file that the TUI parses as JSON.
	// Use the level forwarded from the parent CLI via --log-level.
	var level slog.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		level = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
	runtime.WrapDefaultWithCtx()
	ctx = runtime.WithServiceKind(ctx, name, "daemon")
	slog.InfoContext(ctx, "daemon starting", "config", stackPath)

	paths, err := PathsFor(name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(paths.Dir, 0o755); err != nil {
		return fmt.Errorf("daemon: mkdir %s: %w", paths.Dir, err)
	}

	absConfig, err := filepath.Abs(stackPath)
	if err != nil {
		absConfig = stackPath
	}
	meta := StackMeta{
		Name:       name,
		ConfigPath: absConfig,
		PID:        os.Getpid(),
		StartedAt:  time.Now().UTC().Format(time.RFC3339),
	}
	if metaBytes, err := json.MarshalIndent(meta, "", "  "); err == nil {
		if werr := os.WriteFile(paths.Meta, metaBytes, 0o644); werr != nil {
			slog.WarnContext(ctx, "daemon: write stack meta", "path", paths.Meta, "error", werr)
		}
	}
	defer os.Remove(paths.Meta)

	data, err := os.ReadFile(stackPath)
	if err != nil {
		return fmt.Errorf("daemon: read stack: %w", err)
	}
	cfg, err := config.Parse(data)
	if err != nil {
		return fmt.Errorf("daemon: parse stack: %w", err)
	}
	if err := loadStackEnv(stackPath, cfg); err != nil {
		return fmt.Errorf("daemon: load env: %w", err)
	}
	plan, err := runtime.Translate(cfg, stackPath)
	if err != nil {
		return fmt.Errorf("daemon: translate stack: %w", err)
	}
	if len(plan.MCPs)+len(plan.Agents) == 0 {
		return fmt.Errorf("daemon: no services declared in %s", stackPath)
	}

	initialState := ipc.StateStopped
	if startAll {
		initialState = ipc.StateRunning
	}
	state := newStateTable(plan, initialState)
	sm := newServiceManagerWithRunners(plan, state, runtime.RunMCP, runtime.RunAgent)

	ln, err := listenSocket(paths.Socket)
	if err != nil {
		return err
	}
	defer ln.Close()
	defer os.Remove(paths.Socket)

	runCtx, cancelRun := context.WithCancel(ctx)
	defer cancelRun()

	mux := http.NewServeMux()
	mux.HandleFunc(ipc.PathStatus, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		if err := json.NewEncoder(w).Encode(state.snapshot()); err != nil {
			slog.WarnContext(r.Context(), "daemon: encode status", "error", err)
		}
	})
	mux.HandleFunc(ipc.PathShutdown, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		slog.InfoContext(runCtx, "daemon: shutdown requested via socket")
		cancelRun()
	})
	mux.HandleFunc(ipc.PathStart, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		svcName := r.URL.Query().Get("name")
		if svcName == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		if err := sm.start(runCtx, svcName); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc(ipc.PathStop, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		svcName := r.URL.Query().Get("name")
		if svcName == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		sm.stop(svcName)
		w.WriteHeader(http.StatusAccepted)
	})

	srv := &http.Server{Handler: mux}
	srvDone := make(chan error, 1)
	go func() {
		err := srv.Serve(ln)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		srvDone <- err
	}()

	slog.InfoContext(ctx, "daemon ready", "name", name, "socket", paths.Socket, "services", len(plan.MCPs)+len(plan.Agents))

	if startAll {
		sm.startAll(runCtx)
	}

	<-runCtx.Done()
	sm.stopAll()
	state.mu.Lock()
	for i := range state.rows {
		state.rows[i].State = ipc.StateStopped
	}
	state.mu.Unlock()

	shutCtx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()
	_ = srv.Shutdown(shutCtx)
	<-srvDone
	return nil
}

// listenSocket opens a Unix listener at path. If the address is in
// use, it probes the socket: a live peer means another daemon owns
// the stack (error); a dead socket file is removed and listen is retried once.
func listenSocket(path string) (net.Listener, error) {
	if ln, err := net.Listen("unix", path); err == nil {
		return ln, nil
	}
	if probeLive(path) {
		return nil, fmt.Errorf("daemon: another instance is already running at %s", path)
	}
	_ = os.Remove(path)
	ln, err := net.Listen("unix", path)
	if err != nil {
		return nil, fmt.Errorf("daemon: listen %s: %w", path, err)
	}
	return ln, nil
}

func probeLive(socket string) bool {
	conn, err := net.DialTimeout("unix", socket, 200*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// stateTable tracks per-service lifecycle for /status responses.
type stateTable struct {
	mu   sync.RWMutex
	rows []ipc.StatusLine
}

func newStateTable(plan runtime.Plan, initialState ipc.State) *stateTable {
	t := &stateTable{}
	for _, m := range plan.MCPs {
		t.rows = append(t.rows, ipc.StatusLine{
			Name:  m.Name,
			Kind:  ipc.KindMCP,
			State: initialState,
		})
	}
	for _, a := range plan.Agents {
		t.rows = append(t.rows, ipc.StatusLine{
			Name:  a.Name,
			Kind:  ipc.KindAgent,
			State: initialState,
		})
	}
	return t
}

func (t *stateTable) snapshot() ipc.StatusResponse {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]ipc.StatusLine, len(t.rows))
	copy(out, t.rows)
	return ipc.StatusResponse{Services: out}
}

func (t *stateTable) setState(name string, s ipc.State) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i := range t.rows {
		if t.rows[i].Name == name {
			t.rows[i].State = s
			return
		}
	}
}

// serviceManager owns the per-service goroutines. Each service runs in
// its own goroutine with its own cancel so it can be started and stopped
// independently without touching siblings.
type serviceManager struct {
	mu      sync.Mutex
	entries map[string]serviceEntry
	names   []string // preserves plan ordering for startAll
	cancels map[string]context.CancelFunc
	state   *stateTable
	wg      sync.WaitGroup
}

// serviceEntry holds the per-service runtime hooks. The closure
// captures the typed spec and the runner that knows how to drive it,
// so the manager can treat MCPs and agents uniformly.
type serviceEntry struct {
	autoRestart bool
	run         func(ctx context.Context) error
}

func newServiceManagerWithRunners(
	plan runtime.Plan,
	state *stateTable,
	mcpRunner runtime.Runner,
	agentRunner runtime.AgentRunner,
) *serviceManager {
	total := len(plan.MCPs) + len(plan.Agents)
	sm := &serviceManager{
		entries: make(map[string]serviceEntry, total),
		names:   make([]string, 0, total),
		cancels: make(map[string]context.CancelFunc, total),
		state:   state,
	}
	for _, spec := range plan.MCPs {
		sm.entries[spec.Name] = serviceEntry{
			autoRestart: spec.AutoRestart,
			run: func(ctx context.Context) error {
				return mcpRunner(ctx, spec, os.Stdout, os.Stderr)
			},
		}
		sm.names = append(sm.names, spec.Name)
	}
	for _, spec := range plan.Agents {
		sm.entries[spec.Name] = serviceEntry{
			autoRestart: spec.AutoRestart,
			run: func(ctx context.Context) error {
				if agentRunner == nil {
					return fmt.Errorf("agent runner not configured")
				}
				return agentRunner(ctx, spec, os.Stdout, os.Stderr)
			},
		}
		sm.names = append(sm.names, spec.Name)
	}
	return sm
}

// start spawns a service goroutine if one is not already running.
// Idempotent: already-running is a no-op.
func (sm *serviceManager) start(ctx context.Context, name string) error {
	sm.mu.Lock()
	entry, ok := sm.entries[name]
	if !ok {
		sm.mu.Unlock()
		return fmt.Errorf("unknown service %q", name)
	}
	if _, running := sm.cancels[name]; running {
		sm.mu.Unlock()
		return nil
	}
	svcCtx, cancel := context.WithCancel(ctx)
	sm.cancels[name] = cancel
	sm.state.setState(name, ipc.StateRunning)
	sm.mu.Unlock()

	sm.wg.Go(func() {
		for {
			err := entry.run(svcCtx)
			if err == nil {
				slog.InfoContext(ctx, "daemon: service finished", "service", name)
				break
			}
			if svcCtx.Err() != nil {
				break
			}
			slog.WarnContext(ctx, "daemon: service exited with error", "service", name, "error", err)
			if !entry.autoRestart {
				break
			}
			select {
			case <-time.After(restartDelay):
			case <-svcCtx.Done():
				return
			}
		}
		sm.mu.Lock()
		delete(sm.cancels, name)
		sm.mu.Unlock()
		sm.state.setState(name, ipc.StateStopped)
	})
	return nil
}

// stop cancels a running service. Idempotent: already-stopped is a no-op.
func (sm *serviceManager) stop(name string) {
	sm.mu.Lock()
	cancel, ok := sm.cancels[name]
	if ok {
		delete(sm.cancels, name)
	}
	sm.mu.Unlock()
	if ok {
		cancel()
	}
}

// startAll starts every registered service in plan order.
func (sm *serviceManager) startAll(ctx context.Context) {
	for _, name := range sm.names {
		if err := sm.start(ctx, name); err != nil {
			slog.WarnContext(ctx, "daemon: failed to start service", "service", name, "error", err)
		}
	}
}

// stopAll waits for all running service goroutines to exit. The caller
// is responsible for cancelling the context passed to start/startAll.
func (sm *serviceManager) stopAll() {
	sm.wg.Wait()
}
