package bridge

import (
	"testing"
)

func TestSessionState_Values(t *testing.T) {
	states := []SessionState{
		SessionStateIdle,
		SessionStateRunning,
		SessionStateRequiresAction,
	}
	expected := []string{"idle", "running", "requires_action"}
	for i, s := range states {
		if string(s) != expected[i] {
			t.Errorf("state[%d] = %q, want %q", i, s, expected[i])
		}
	}
}

func TestAttachBridgeSession_RequiredFields(t *testing.T) {
	tests := []struct {
		name string
		opts AttachBridgeSessionOptions
	}{
		{"missing sessionId", AttachBridgeSessionOptions{IngressToken: "tok", APIBaseURL: "http://x"}},
		{"missing ingressToken", AttachBridgeSessionOptions{SessionID: "cse_1", APIBaseURL: "http://x"}},
		{"missing apiBaseUrl", AttachBridgeSessionOptions{SessionID: "cse_1", IngressToken: "tok"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := AttachBridgeSession(tt.opts)
			if err == nil {
				t.Error("expected error for missing required field")
			}
		})
	}
}

func TestAttachBridgeSession_Success(t *testing.T) {
	seqNum := 42
	handle, err := AttachBridgeSession(AttachBridgeSessionOptions{
		SessionID:          "cse_test123",
		IngressToken:       "jwt-token",
		APIBaseURL:         "https://api.example.com",
		InitialSequenceNum: &seqNum,
	})
	if err != nil {
		t.Fatalf("AttachBridgeSession: %v", err)
	}
	if handle.SessionID() != "cse_test123" {
		t.Errorf("SessionID = %q", handle.SessionID())
	}
	if handle.GetSequenceNum() != 42 {
		t.Errorf("GetSequenceNum = %d, want 42", handle.GetSequenceNum())
	}
	if !handle.IsConnected() {
		t.Error("expected IsConnected=true")
	}
}

func TestBridgeSessionHandle_Close(t *testing.T) {
	handle, err := AttachBridgeSession(AttachBridgeSessionOptions{
		SessionID:    "cse_close",
		IngressToken: "tok",
		APIBaseURL:   "http://localhost",
	})
	if err != nil {
		t.Fatal(err)
	}
	handle.Close()
	if handle.IsConnected() {
		t.Error("expected IsConnected=false after Close")
	}
}

func TestBridgeSessionHandle_Methods(t *testing.T) {
	handle, err := AttachBridgeSession(AttachBridgeSessionOptions{
		SessionID:    "cse_methods",
		IngressToken: "tok",
		APIBaseURL:   "http://localhost",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer handle.Close()

	// Verify all methods exist and don't panic.
	handle.Write(nil)
	handle.SendResult()
	handle.SendControlRequest(nil)
	handle.SendControlResponse(nil)
	handle.SendControlCancelRequest("req-1")
	handle.ReportState(SessionStateRunning)
	handle.ReportMetadata(map[string]interface{}{"branch": "main"})
	handle.ReportDelivery("evt-1", "processing")
	if err := handle.Flush(); err != nil {
		t.Errorf("Flush: %v", err)
	}
	if err := handle.ReconnectTransport(ReconnectOptions{
		IngressToken: "new-tok",
		APIBaseURL:   "http://new",
	}); err != nil {
		t.Errorf("ReconnectTransport: %v", err)
	}
}

func TestAttachBridgeSessionOptions_Callbacks(t *testing.T) {
	var setCalled bool
	var modelCalled bool

	handle, err := AttachBridgeSession(AttachBridgeSessionOptions{
		SessionID:    "cse_cb",
		IngressToken: "tok",
		APIBaseURL:   "http://localhost",
		OnSetModel: func(model *string) {
			modelCalled = true
		},
		OnSetPermissionMode: func(mode string) *BridgePermissionModeResult {
			setCalled = true
			return &BridgePermissionModeResult{OK: true}
		},
		OnClose: func(code *int) {},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer handle.Close()

	// Verify callbacks are stored and callable.
	if handle.opts.OnSetModel != nil {
		handle.opts.OnSetModel(nil)
	}
	if handle.opts.OnSetPermissionMode != nil {
		result := handle.opts.OnSetPermissionMode("plan")
		if !result.OK {
			t.Error("expected OK=true")
		}
	}
	if !modelCalled {
		t.Error("OnSetModel not called")
	}
	if !setCalled {
		t.Error("OnSetPermissionMode not called")
	}
}

func TestBridgePermissionModeResult(t *testing.T) {
	ok := BridgePermissionModeResult{OK: true}
	if !ok.OK {
		t.Error("expected OK=true")
	}

	fail := BridgePermissionModeResult{OK: false, Error: "not supported"}
	if fail.OK {
		t.Error("expected OK=false")
	}
	if fail.Error != "not supported" {
		t.Errorf("Error = %q", fail.Error)
	}
}
