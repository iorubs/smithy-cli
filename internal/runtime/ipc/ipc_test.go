package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"testing"
)

// sockPath returns a short socket path under dir. macOS limits unix
// socket paths to ~104 bytes so we keep the filename short.
func sockPath(dir, name string) string {
	return filepath.Join(dir, name+".sock")
}

func newTestServer(t *testing.T, sock string, mux *http.ServeMux) {
	t.Helper()
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() {
		srv.Close()
		ln.Close()
	})
}

func TestClientStatus(t *testing.T) {
	sock := sockPath(t.TempDir(), "s")
	mux := http.NewServeMux()
	mux.HandleFunc(PathStatus, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		if err := json.NewEncoder(w).Encode(StatusResponse{
			Services: []StatusLine{
				{Name: "a", Kind: KindMCP, State: StateRunning},
			},
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	newTestServer(t, sock, mux)

	resp, err := NewClient(sock).Status(t.Context())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(resp.Services) != 1 || resp.Services[0].Name != "a" {
		t.Fatalf("unexpected status: %+v", resp)
	}
	if resp.Services[0].State != StateRunning {
		t.Errorf("state: got %q, want %q", resp.Services[0].State, StateRunning)
	}
	if resp.Services[0].Kind != KindMCP {
		t.Errorf("kind: got %q, want %q", resp.Services[0].Kind, KindMCP)
	}
}

func TestClientShutdown(t *testing.T) {
	called := make(chan struct{}, 1)
	sock := sockPath(t.TempDir(), "s")
	mux := http.NewServeMux()
	mux.HandleFunc(PathShutdown, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		called <- struct{}{}
		w.WriteHeader(http.StatusAccepted)
	})
	newTestServer(t, sock, mux)

	if err := NewClient(sock).Shutdown(t.Context()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	select {
	case <-called:
	default:
		t.Error("handler was not called")
	}
}

func TestClientStartService(t *testing.T) {
	cases := []struct {
		name string
		kind Kind
		path string
		call func(*Client, context.Context, string, Kind) error
	}{
		{"start-mcp", KindMCP, PathStart, func(c *Client, ctx context.Context, n string, k Kind) error {
			return c.StartService(ctx, n, k)
		}},
		{"start-agent", KindAgent, PathStart, func(c *Client, ctx context.Context, n string, k Kind) error {
			return c.StartService(ctx, n, k)
		}},
		{"stop-mcp", KindMCP, PathStop, func(c *Client, ctx context.Context, n string, k Kind) error {
			return c.StopService(ctx, n, k)
		}},
	}
	dir := t.TempDir()
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotName, gotKind string
			sock := sockPath(dir, fmt.Sprintf("st%d", i))
			mux := http.NewServeMux()
			mux.HandleFunc(tc.path, func(w http.ResponseWriter, r *http.Request) {
				gotName = r.URL.Query().Get("name")
				gotKind = r.URL.Query().Get("kind")
				w.WriteHeader(http.StatusAccepted)
			})
			newTestServer(t, sock, mux)

			if err := tc.call(NewClient(sock), t.Context(), tc.name, tc.kind); err != nil {
				t.Fatalf("call: %v", err)
			}
			if gotName != tc.name {
				t.Errorf("name: got %q, want %q", gotName, tc.name)
			}
			if gotKind != string(tc.kind) {
				t.Errorf("kind: got %q, want %q", gotKind, tc.kind)
			}
		})
	}
}

func TestClientErrorOnBadStatus(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		doCall func(*Client) error
	}{
		{"status", PathStatus, func(c *Client) error {
			_, err := c.Status(context.Background())
			return err
		}},
		{"shutdown", PathShutdown, func(c *Client) error {
			return c.Shutdown(context.Background())
		}},
		{"start", PathStart, func(c *Client) error {
			return c.StartService(context.Background(), "x", KindMCP)
		}},
		{"stop", PathStop, func(c *Client) error {
			return c.StopService(context.Background(), "x", KindMCP)
		}},
	}
	dir := t.TempDir()
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sock := sockPath(dir, fmt.Sprintf("e%d", i))
			mux := http.NewServeMux()
			mux.HandleFunc(tc.path, func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "server error", http.StatusInternalServerError)
			})
			newTestServer(t, sock, mux)

			if err := tc.doCall(NewClient(sock)); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}
