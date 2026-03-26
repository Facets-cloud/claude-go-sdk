package claudeagent

import (
	"encoding/json"
	"testing"
)

func TestSdkMcpTransport_Send(t *testing.T) {
	var received json.RawMessage
	transport := NewSdkMcpTransport(func(msg json.RawMessage) {
		received = msg
	})

	msg := json.RawMessage(`{"jsonrpc":"2.0","method":"test","id":1}`)
	if err := transport.Send(msg); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if string(received) != string(msg) {
		t.Errorf("received = %s, want %s", received, msg)
	}
}

func TestSdkMcpTransport_Receive(t *testing.T) {
	var received json.RawMessage
	transport := NewSdkMcpTransport(nil)
	transport.OnMessage = func(msg json.RawMessage) {
		received = msg
	}

	msg := json.RawMessage(`{"jsonrpc":"2.0","result":"ok","id":1}`)
	transport.Receive(msg)

	if string(received) != string(msg) {
		t.Errorf("received = %s, want %s", received, msg)
	}
}

func TestSdkMcpTransport_Close(t *testing.T) {
	closeCalled := false
	transport := NewSdkMcpTransport(nil)
	transport.OnClose = func() { closeCalled = true }

	transport.Close()

	if !closeCalled {
		t.Error("OnClose not called")
	}
	if !transport.IsClosed() {
		t.Error("expected IsClosed() to be true")
	}

	// Send after close should not panic or error.
	if err := transport.Send(json.RawMessage(`{}`)); err != nil {
		t.Errorf("Send after close: %v", err)
	}
}

func TestSdkMcpTransport_CloseIdempotent(t *testing.T) {
	closeCount := 0
	transport := NewSdkMcpTransport(nil)
	transport.OnClose = func() { closeCount++ }

	transport.Close()
	transport.Close()

	if closeCount != 1 {
		t.Errorf("OnClose called %d times, want 1", closeCount)
	}
}

func TestSdkMcpServerManager_RegisterAndGet(t *testing.T) {
	mgr := NewSdkMcpServerManager()
	transport := NewSdkMcpTransport(nil)

	mgr.Register("test-server", transport)

	got, ok := mgr.Get("test-server")
	if !ok {
		t.Fatal("expected to find test-server")
	}
	if got != transport {
		t.Error("returned wrong transport")
	}

	_, ok = mgr.Get("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent server")
	}
}

func TestSdkMcpServerManager_Names(t *testing.T) {
	mgr := NewSdkMcpServerManager()
	mgr.Register("a", NewSdkMcpTransport(nil))
	mgr.Register("b", NewSdkMcpTransport(nil))

	names := mgr.Names()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
}

func TestSdkMcpServerManager_CloseAll(t *testing.T) {
	mgr := NewSdkMcpServerManager()
	t1 := NewSdkMcpTransport(nil)
	t2 := NewSdkMcpTransport(nil)
	mgr.Register("a", t1)
	mgr.Register("b", t2)

	mgr.CloseAll()

	if !t1.IsClosed() || !t2.IsClosed() {
		t.Error("expected all transports to be closed")
	}
}

func TestHandleMcpRequestResponse_Request(t *testing.T) {
	// Simulate an in-process MCP server that echoes requests.
	transport := NewSdkMcpTransport(nil)
	transport.OnMessage = func(msg json.RawMessage) {
		// The "MCP server" receives the request and sends a response via Send().
		var req struct {
			Method string          `json:"method"`
			ID     json.RawMessage `json:"id"`
		}
		json.Unmarshal(msg, &req)
		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"result":  map[string]interface{}{"tools": []string{}},
			"id":      json.RawMessage(req.ID),
		}
		respJSON, _ := json.Marshal(resp)
		transport.Send(respJSON)
	}

	request := json.RawMessage(`{"jsonrpc":"2.0","method":"tools/list","id":42}`)
	resp, err := handleMcpRequestResponse(transport, request)
	if err != nil {
		t.Fatalf("handleMcpRequestResponse: %v", err)
	}

	var parsed struct {
		Result struct {
			Tools []string `json:"tools"`
		} `json:"result"`
		ID int `json:"id"`
	}
	if err := json.Unmarshal(resp, &parsed); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if parsed.ID != 42 {
		t.Errorf("response id = %d, want 42", parsed.ID)
	}
}

func TestHandleMcpRequestResponse_Notification(t *testing.T) {
	var received json.RawMessage
	transport := NewSdkMcpTransport(nil)
	transport.OnMessage = func(msg json.RawMessage) {
		received = msg
	}

	// Notification has method but no id.
	notification := json.RawMessage(`{"jsonrpc":"2.0","method":"notifications/cancelled"}`)
	resp, err := handleMcpRequestResponse(transport, notification)
	if err != nil {
		t.Fatalf("handleMcpRequestResponse: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response for notification")
	}

	if string(received) != string(notification) {
		t.Errorf("notification not forwarded: got %s", received)
	}
}

func TestHandleMcpRequestResponse_NotificationWithNullID(t *testing.T) {
	var received json.RawMessage
	transport := NewSdkMcpTransport(nil)
	transport.OnMessage = func(msg json.RawMessage) {
		received = msg
	}

	// Notification with id:null should be treated as a notification.
	notification := json.RawMessage(`{"jsonrpc":"2.0","method":"test","id":null}`)
	resp, err := handleMcpRequestResponse(transport, notification)
	if err != nil {
		t.Fatalf("handleMcpRequestResponse: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	if string(received) != string(notification) {
		t.Errorf("message not forwarded: got %s", received)
	}
}
