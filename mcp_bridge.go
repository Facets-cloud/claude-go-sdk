package claudeagent

import (
	"encoding/json"
	"sync"
)

// SdkMcpTransport relays JSON-RPC messages between the CLI and an in-process
// MCP server. It implements the transport interface expected by MCP server
// libraries: incoming messages (from CLI) arrive via OnMessage, outgoing
// messages (from the MCP server) are forwarded via the sendToQuery callback.
//
// This mirrors the TS SDK's oQ class.
type SdkMcpTransport struct {
	// OnMessage is called when the CLI sends a JSON-RPC message to this server.
	OnMessage func(msg json.RawMessage)

	// OnClose is called when the transport is closed.
	OnClose func()

	// sendToQuery forwards a JSON-RPC message from the MCP server back to the CLI.
	sendToQuery func(msg json.RawMessage)

	mu     sync.Mutex
	closed bool
}

// NewSdkMcpTransport creates a transport that forwards messages via the given callback.
func NewSdkMcpTransport(sendToQuery func(msg json.RawMessage)) *SdkMcpTransport {
	return &SdkMcpTransport{
		sendToQuery: sendToQuery,
	}
}

// Send sends a JSON-RPC message from the MCP server to the CLI.
func (t *SdkMcpTransport) Send(msg json.RawMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return nil
	}
	if t.sendToQuery != nil {
		t.sendToQuery(msg)
	}
	return nil
}

// Close closes the transport.
func (t *SdkMcpTransport) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return
	}
	t.closed = true
	if t.OnClose != nil {
		t.OnClose()
	}
}

// IsClosed returns whether the transport has been closed.
func (t *SdkMcpTransport) IsClosed() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.closed
}

// Receive delivers a message from the CLI to the in-process MCP server.
func (t *SdkMcpTransport) Receive(msg json.RawMessage) {
	if t.OnMessage != nil {
		t.OnMessage(msg)
	}
}

// SdkMcpServerManager manages in-process MCP server transports.
type SdkMcpServerManager struct {
	mu         sync.RWMutex
	transports map[string]*SdkMcpTransport
}

// NewSdkMcpServerManager creates a new server manager.
func NewSdkMcpServerManager() *SdkMcpServerManager {
	return &SdkMcpServerManager{
		transports: make(map[string]*SdkMcpTransport),
	}
}

// Register registers a transport for the given server name.
func (m *SdkMcpServerManager) Register(name string, transport *SdkMcpTransport) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.transports[name] = transport
}

// Get returns the transport for the given server name.
func (m *SdkMcpServerManager) Get(name string) (*SdkMcpTransport, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.transports[name]
	return t, ok
}

// Names returns the registered server names.
func (m *SdkMcpServerManager) Names() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	names := make([]string, 0, len(m.transports))
	for name := range m.transports {
		names = append(names, name)
	}
	return names
}

// CloseAll closes all transports.
func (m *SdkMcpServerManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.transports {
		t.Close()
	}
}

// handleMcpRequestResponse routes a JSON-RPC message to the in-process MCP server.
// For requests (method+id): forwards to server via OnMessage, intercepts the server's
// response via the sendToQuery path, and returns it.
// For notifications (no id): just forwards and returns a stub.
func handleMcpRequestResponse(transport *SdkMcpTransport, message json.RawMessage) (json.RawMessage, error) {
	var env struct {
		Method string          `json:"method,omitempty"`
		ID     json.RawMessage `json:"id,omitempty"`
	}
	json.Unmarshal(message, &env)

	isRequest := env.Method != "" && env.ID != nil && string(env.ID) != "null"

	if isRequest {
		// This is a JSON-RPC request — wait for the server's response.
		// Temporarily intercept the sendToQuery path to capture the response.
		responseCh := make(chan json.RawMessage, 1)
		origSend := transport.sendToQuery
		transport.sendToQuery = func(msg json.RawMessage) {
			var respEnv struct {
				ID json.RawMessage `json:"id,omitempty"`
			}
			json.Unmarshal(msg, &respEnv)
			if string(respEnv.ID) == string(env.ID) {
				responseCh <- msg
			} else if origSend != nil {
				origSend(msg)
			}
		}
		defer func() { transport.sendToQuery = origSend }()

		// Forward to MCP server.
		transport.Receive(message)

		// Wait for server response (comes through Send → sendToQuery).
		resp := <-responseCh
		return resp, nil
	}

	// Notification — just forward, no response expected.
	transport.Receive(message)
	return json.RawMessage(`{"jsonrpc":"2.0","result":{},"id":0}`), nil
}
