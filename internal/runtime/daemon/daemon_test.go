package daemon

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/iorubs/smithy-cli/internal/runtime"
	"github.com/iorubs/smithy-cli/internal/runtime/ipc"
)

func makePlan(names ...string) runtime.Plan {
	specs := make([]runtime.MCPSpec, len(names))
	for i, n := range names {
		specs[i] = runtime.MCPSpec{Name: n, AutoRestart: true}
	}
	return runtime.Plan{MCPs: specs}
}

// newServiceManagerWithRunner wires only an MCP runner, for tests that
// exercise MCP fan-out behaviour without bringing the agent runtime in.
func newServiceManagerWithRunner(plan runtime.Plan, state *stateTable, runner runtime.Runner) *serviceManager {
	return newServiceManagerWithRunners(plan, state, runner, nil)
}

func TestStateTable(t *testing.T) {
	t.Run("initial state is running for all services", func(t *testing.T) {
		st := newStateTable(makePlan("a", "b"), ipc.StateRunning)
		snap := st.snapshot()
		if len(snap.Services) != 2 {
			t.Fatalf("got %d services, want 2", len(snap.Services))
		}
		for _, s := range snap.Services {
			if s.State != ipc.StateRunning {
				t.Errorf("service %q: state = %q, want running", s.Name, s.State)
			}
			if s.Kind != ipc.KindMCP {
				t.Errorf("service %q: kind = %q, want mcp", s.Name, s.Kind)
			}
		}
	})

	t.Run("setState updates correct row", func(t *testing.T) {
		st := newStateTable(makePlan("a", "b"), ipc.StateRunning)
		st.setState("a", ipc.StateStopped)
		snap := st.snapshot()
		for _, s := range snap.Services {
			if s.Name == "a" && s.State != ipc.StateStopped {
				t.Errorf("a: got %q, want stopped", s.State)
			}
			if s.Name == "b" && s.State != ipc.StateRunning {
				t.Errorf("b: got %q, want running", s.State)
			}
		}
	})

	t.Run("setState unknown name is a no-op", func(t *testing.T) {
		st := newStateTable(makePlan("a"), ipc.StateRunning)
		st.setState("nope", ipc.StateStopped) // must not panic
		snap := st.snapshot()
		if snap.Services[0].State != ipc.StateRunning {
			t.Errorf("a unexpectedly changed")
		}
	})

	t.Run("agents land with KindAgent", func(t *testing.T) {
		plan := runtime.Plan{
			MCPs:   []runtime.MCPSpec{{Name: "m", AutoRestart: true}},
			Agents: []runtime.AgentSpec{{Name: "a", AutoRestart: true}},
		}
		st := newStateTable(plan, ipc.StateRunning)
		snap := st.snapshot()
		if len(snap.Services) != 2 {
			t.Fatalf("got %d services, want 2", len(snap.Services))
		}
		var gotMCP, gotAgent bool
		for _, s := range snap.Services {
			switch s.Name {
			case "m":
				if s.Kind != ipc.KindMCP {
					t.Errorf("m kind = %q, want mcp", s.Kind)
				}
				gotMCP = true
			case "a":
				if s.Kind != ipc.KindAgent {
					t.Errorf("a kind = %q, want agent", s.Kind)
				}
				gotAgent = true
			}
		}
		if !gotMCP || !gotAgent {
			t.Errorf("missing rows: gotMCP=%v gotAgent=%v", gotMCP, gotAgent)
		}
	})
}

func TestServiceManager(t *testing.T) {
	blocker := func(ctx context.Context, _ runtime.MCPSpec, _, _ io.Writer) error {
		<-ctx.Done()
		return nil
	}

	t.Run("start runs service and stop cancels it", func(t *testing.T) {
		plan := makePlan("svc")
		st := newStateTable(plan, ipc.StateStopped)
		sm := newServiceManagerWithRunner(plan, st, blocker)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		if err := sm.start(ctx, "svc"); err != nil {
			t.Fatalf("start: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
		if snap := st.snapshot(); snap.Services[0].State != ipc.StateRunning {
			t.Errorf("expected running, got %q", snap.Services[0].State)
		}

		sm.stop("svc")
		sm.stopAll()
		if snap := st.snapshot(); snap.Services[0].State != ipc.StateStopped {
			t.Errorf("expected stopped, got %q", snap.Services[0].State)
		}
	})

	t.Run("start is idempotent", func(t *testing.T) {
		plan := makePlan("svc")
		st := newStateTable(plan, ipc.StateStopped)
		sm := newServiceManagerWithRunner(plan, st, blocker)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		if err := sm.start(ctx, "svc"); err != nil {
			t.Fatalf("first start: %v", err)
		}
		if err := sm.start(ctx, "svc"); err != nil {
			t.Fatalf("second start: %v", err)
		}
		cancel()
		sm.stopAll()
	})

	t.Run("start unknown service returns error", func(t *testing.T) {
		plan := makePlan("svc")
		st := newStateTable(plan, ipc.StateStopped)
		sm := newServiceManagerWithRunner(plan, st, blocker)
		if err := sm.start(t.Context(), "nope"); err == nil {
			t.Fatal("expected error for unknown service")
		}
	})

	t.Run("stop is idempotent", func(t *testing.T) {
		plan := makePlan("svc")
		st := newStateTable(plan, ipc.StateStopped)
		sm := newServiceManagerWithRunner(plan, st, blocker)
		sm.stop("svc") // not started, must not panic
		sm.stop("svc")
	})

	t.Run("autorestart true restarts on error", func(t *testing.T) {
		var mu sync.Mutex
		calls := 0
		runner := func(ctx context.Context, _ runtime.MCPSpec, _, _ io.Writer) error {
			mu.Lock()
			calls++
			n := calls
			mu.Unlock()
			if n < 3 {
				return errors.New("crash")
			}
			<-ctx.Done()
			return nil
		}
		spec := runtime.MCPSpec{Name: "svc", AutoRestart: true}
		plan := runtime.Plan{MCPs: []runtime.MCPSpec{spec}}
		st := newStateTable(plan, ipc.StateStopped)
		sm := newServiceManagerWithRunner(plan, st, runner)

		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		defer cancel()

		if err := sm.start(ctx, "svc"); err != nil {
			t.Fatalf("start: %v", err)
		}
		for {
			mu.Lock()
			n := calls
			mu.Unlock()
			if n >= 3 {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		cancel()
		sm.stopAll()
	})

	t.Run("autorestart false does not restart on error", func(t *testing.T) {
		var mu sync.Mutex
		calls := 0
		runner := func(_ context.Context, _ runtime.MCPSpec, _, _ io.Writer) error {
			mu.Lock()
			calls++
			mu.Unlock()
			return errors.New("crash")
		}
		spec := runtime.MCPSpec{Name: "svc", AutoRestart: false}
		plan := runtime.Plan{MCPs: []runtime.MCPSpec{spec}}
		st := newStateTable(plan, ipc.StateStopped)
		sm := newServiceManagerWithRunner(plan, st, runner)

		if err := sm.start(t.Context(), "svc"); err != nil {
			t.Fatalf("start: %v", err)
		}
		sm.stopAll()

		mu.Lock()
		n := calls
		mu.Unlock()
		if n != 1 {
			t.Errorf("expected 1 call, got %d", n)
		}
		if snap := st.snapshot(); snap.Services[0].State != ipc.StateStopped {
			t.Errorf("expected stopped, got %q", snap.Services[0].State)
		}
	})

	t.Run("clean exit does not restart even with autorestart true", func(t *testing.T) {
		var mu sync.Mutex
		calls := 0
		runner := func(_ context.Context, _ runtime.MCPSpec, _, _ io.Writer) error {
			mu.Lock()
			calls++
			mu.Unlock()
			return nil
		}
		spec := runtime.MCPSpec{Name: "svc", AutoRestart: true}
		plan := runtime.Plan{MCPs: []runtime.MCPSpec{spec}}
		st := newStateTable(plan, ipc.StateStopped)
		sm := newServiceManagerWithRunner(plan, st, runner)

		if err := sm.start(t.Context(), "svc"); err != nil {
			t.Fatalf("start: %v", err)
		}
		sm.stopAll()

		mu.Lock()
		n := calls
		mu.Unlock()
		if n != 1 {
			t.Errorf("expected 1 call (clean exit, no restart), got %d", n)
		}
	})
}
