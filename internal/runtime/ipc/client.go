package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Client is a tiny HTTP-over-Unix-socket dialer that talks to the
// daemon. It is intentionally bare: one transport, one Get helper.
type Client struct {
	socket string
	http   *http.Client
}

// NewClient returns a Client bound to the Unix socket at path.
func NewClient(socket string) *Client {
	return &Client{
		socket: socket,
		http: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					var d net.Dialer
					return d.DialContext(ctx, "unix", socket)
				},
			},
		},
	}
}

// Status fetches the service list from the daemon. The host portion
// of the URL is irrelevant; the socket is fixed by the transport.
func (c *Client) Status(ctx context.Context) (StatusResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://daemon"+PathStatus, nil)
	if err != nil {
		return StatusResponse{}, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return StatusResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return StatusResponse{}, fmt.Errorf("ipc: GET %s: %s", PathStatus, resp.Status)
	}
	var out StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return StatusResponse{}, fmt.Errorf("ipc: decode status: %w", err)
	}
	return out, nil
}

// Shutdown asks the daemon to stop all services and exit. The daemon
// acknowledges immediately; callers should poll the socket to confirm
// the process has actually exited.
func (c *Client) Shutdown(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://daemon"+PathShutdown, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("ipc: POST %s: %s", PathShutdown, resp.Status)
	}
	return nil
}

// StartService asks the daemon to start the named service of the given kind.
// Idempotent: if the service is already running, the daemon returns 202 with no effect.
func (c *Client) StartService(ctx context.Context, name string, kind Kind) error {
	u := "http://daemon" + PathStart + "?" + url.Values{"name": {name}, "kind": {string(kind)}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("ipc: POST %s: %s", PathStart, resp.Status)
	}
	return nil
}

// StopService asks the daemon to stop the named service of the given kind
// without disturbing other services. Idempotent: already-stopped is a no-op.
func (c *Client) StopService(ctx context.Context, name string, kind Kind) error {
	u := "http://daemon" + PathStop + "?" + url.Values{"name": {name}, "kind": {string(kind)}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("ipc: POST %s: %s", PathStop, resp.Status)
	}
	return nil
}
