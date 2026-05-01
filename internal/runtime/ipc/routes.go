// Package ipc defines the wire shapes and route constants shared between
// the daemon and its clients over the per-stack Unix socket.
package ipc

// Route paths.
const (
	// PathStatus returns one StatusLine per service.
	PathStatus = "/status"
	// PathShutdown asks the daemon to stop all services and exit.
	PathShutdown = "/shutdown"
	// PathStart starts a single named service in the running daemon.
	PathStart = "/start"
	// PathStop stops a single named service without touching siblings.
	PathStop = "/stop"
)

// State is the lifecycle state of a single service inside the daemon's runner.
type State string

// Service states.
const (
	StateRunning State = "running"
	StateStopped State = "stopped"
)

// Kind identifies the type of service (mcp, agent, …).
type Kind string

// Service kinds.
const (
	KindMCP   Kind = "mcp"
	KindAgent Kind = "agent"
)

// StatusLine is one row in the response to GET /status.
type StatusLine struct {
	Name  string `json:"name"`
	Kind  Kind   `json:"kind"`
	State State  `json:"state"`
}

// StatusResponse is the JSON body returned by GET /status.
type StatusResponse struct {
	Services []StatusLine `json:"services"`
}
