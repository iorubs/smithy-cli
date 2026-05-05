package daemon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/iorubs/smithy-cli/internal/runtime"
	"github.com/iorubs/smithy-cli/internal/runtime/ipc"
)

// ErrAlreadyRunning is returned by SpawnDetached when the per-stack
// socket already responds to /status. Callers can treat this as a successful no-op (idempotent up).
var ErrAlreadyRunning = errors.New("daemon: already running")

// ErrNameConflict is returned by SpawnDetached when a stack with the
// requested name is already running but was started from a different
// stack file. Callers should ask the user to disambiguate.
var ErrNameConflict = errors.New("daemon: stack name in use by a different stack file")

// SpawnDetached re-execs the current binary in daemon mode for the
// named stack and returns once the daemon's socket is accepting
// connections (or the timeout elapses). The returned PID is the
// daemon process; the parent is free to exit.
//
// If a live daemon is already serving the stack's socket, the
// existing PID is returned with ErrAlreadyRunning so callers can decide whether to attach or fail.
//
// The child is fully detached: new session, stdio redirected to the
// daemon log file (created if missing), no parent pipes.
func SpawnDetached(ctx context.Context, name, stackPath string, timeout time.Duration, startAll bool) (int, error) {
	paths, err := PathsFor(name)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(paths.Dir, 0o755); err != nil {
		return 0, fmt.Errorf("daemon: mkdir %s: %w", paths.Dir, err)
	}

	if probeLive(paths.Socket) {
		meta, ok := ReadMeta(paths)
		pid := 0
		if ok {
			pid = meta.PID
			absRequested, _ := filepath.Abs(stackPath)
			if meta.ConfigPath != "" && absRequested != "" && meta.ConfigPath != absRequested {
				return pid, fmt.Errorf("%w: %q already running from %s", ErrNameConflict, name, meta.ConfigPath)
			}
		}
		return pid, ErrAlreadyRunning
	}

	if _, err := os.Stat(paths.Socket); err == nil {
		_ = os.Remove(paths.Socket)
	}

	exe, err := os.Executable()
	if err != nil {
		return 0, fmt.Errorf("daemon: locate self: %w", err)
	}

	logFile, err := os.OpenFile(paths.DaemonLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return 0, fmt.Errorf("daemon: open log: %w", err)
	}
	defer logFile.Close()

	args := []string{"__daemon__", "--name", name, "-c", stackPath}
	if startAll {
		args = append(args, "--start-all")
	}
	args = append(args, "--log-level", runtime.LogLevelFromCtx(ctx))
	cmd := exec.Command(exe, args...)
	cmd.Stdin = nil
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("daemon: start child: %w", err)
	}
	pid := cmd.Process.Pid
	if err := cmd.Process.Release(); err != nil {
		return pid, fmt.Errorf("daemon: release child: %w", err)
	}

	if err := WaitForSocket(ctx, paths.Socket, timeout); err != nil {
		return pid, err
	}
	return pid, nil
}

// WaitForSocket polls until path responds with 200 to GET /status,
// ctx is done, or timeout elapses. Polling /status (not just dialing
// the socket) ensures the http.Server has started accepting and
// dispatching, not just that net.Listen returned.
func WaitForSocket(ctx context.Context, path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := ipc.NewClient(path)
	for {
		probeCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
		_, err := client.Status(probeCtx)
		cancel()
		if err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("daemon: socket %s did not become ready within %s", path, timeout)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
}
