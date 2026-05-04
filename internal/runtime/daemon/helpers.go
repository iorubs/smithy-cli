package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"syscall"
	"time"

	"github.com/iorubs/smithy-cli/internal/runtime/ipc"
)

// IsStackRunning returns true if the daemon socket exists and responds to a status probe.
func IsStackRunning(ctx context.Context, paths Paths, probeTimeout time.Duration) bool {
	if _, err := os.Stat(paths.Socket); err != nil {
		return false
	}
	client := ipc.NewClient(paths.Socket)
	probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	_, err := client.Status(probeCtx)
	return err == nil
}

// WaitForExit polls until the daemon's socket is gone or ctx fires.
func WaitForExit(ctx context.Context, paths Paths) error {
	for {
		if _, err := os.Stat(paths.Socket); errors.Is(err, os.ErrNotExist) {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("daemon did not exit: %w", ctx.Err())
		case <-time.After(100 * time.Millisecond):
		}
	}
}

// SignalFromPID is the fallback when the socket is unresponsive: read
// the daemon PID from stack metadata and SIGTERM the process directly.
func SignalFromPID(paths Paths) error {
	meta, ok := ReadMeta(paths)
	if !ok || meta.PID <= 0 {
		return fmt.Errorf("read pid from meta: missing or invalid")
	}
	proc, err := os.FindProcess(meta.PID)
	if err != nil {
		return fmt.Errorf("find pid %d: %w", meta.PID, err)
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("signal pid %d: %w", meta.PID, err)
	}
	return nil
}

// CleanupArtifacts removes the socket, daemon log, and metadata file.
// The per-stack directory is also removed if empty.
func CleanupArtifacts(paths Paths) {
	for _, p := range []string{paths.Socket, paths.DaemonLog, paths.Meta} {
		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			slog.Warn("stack: cleanup failed", "path", p, "error", err)
		}
	}
	if err := os.Remove(paths.Dir); err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Debug("stack: stack dir not removed", "path", paths.Dir, "error", err)
	}
}

// ReadMeta loads the metadata file written by Run. Missing or
// malformed files return ok=false; callers should treat the stack as alive-but-without-metadata.
func ReadMeta(paths Paths) (StackMeta, bool) {
	data, err := os.ReadFile(paths.Meta)
	if err != nil {
		return StackMeta{}, false
	}
	var m StackMeta
	if err := json.Unmarshal(data, &m); err != nil {
		return StackMeta{}, false
	}
	return m, true
}

// SetChatContextID writes contextID into the stack metadata under
// agent. Pass an empty contextID to remove the entry. The write is
// atomic (rename) so concurrent readers never see a torn file.
func SetChatContextID(paths Paths, agent, contextID string) error {
	meta, ok := ReadMeta(paths)
	if !ok {
		return fmt.Errorf("read stack meta")
	}
	if meta.Chats == nil {
		meta.Chats = map[string]string{}
	}
	if contextID == "" {
		delete(meta.Chats, agent)
	} else {
		meta.Chats[agent] = contextID
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	tmp := paths.Meta + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, paths.Meta)
}
