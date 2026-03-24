// Package bridge provides the alpha Bridge API for attaching to existing
// bridge sessions. This API is in a separate versioning universe from the
// main query() surface — breaking changes here do NOT bump the package major.
package bridge

import "context"

// SessionState represents bridge session states.
type SessionState string

const (
	SessionStateIdle           SessionState = "idle"
	SessionStateRunning        SessionState = "running"
	SessionStateRequiresAction SessionState = "requires_action"
)

// BridgePermissionModeResult is the return type for OnSetPermissionMode callbacks.
type BridgePermissionModeResult struct {
	// OK is true if the permission mode was accepted.
	OK bool
	// Error is the error message if OK is false.
	Error string
}

// ReconnectOptions contains options for reconnecting the transport.
type ReconnectOptions struct {
	// IngressToken is the new worker JWT.
	IngressToken string
	// APIBaseURL is the session ingress URL.
	APIBaseURL string
	// Epoch is omitted to call registerWorker; provide if server already bumped.
	Epoch *int
}

// AttachBridgeSessionOptions configures a bridge session attachment.
type AttachBridgeSessionOptions struct {
	// SessionID is the session ID (cse_* form).
	SessionID string
	// IngressToken is the worker JWT.
	IngressToken string
	// APIBaseURL is the session ingress URL.
	APIBaseURL string
	// Epoch is the worker epoch if already known.
	Epoch *int
	// InitialSequenceNum seeds the first SSE connect's from_sequence_num.
	InitialSequenceNum *int
	// HeartbeatIntervalMs is the CCRClient heartbeat interval (default 20s).
	HeartbeatIntervalMs *int

	// OnInboundMessage is called for user messages from claude.ai.
	// Filtered echoes and re-deliveries are excluded.
	OnInboundMessage func(msg interface{}) error
	// OnPermissionResponse is called when the user answers a can_use_tool prompt.
	OnPermissionResponse func(res interface{})
	// OnInterrupt is called on interrupt control_request from claude.ai.
	OnInterrupt func()
	// OnSetModel is called when the model is changed.
	OnSetModel func(model *string)
	// OnSetMaxThinkingTokens is called when max thinking tokens are changed.
	OnSetMaxThinkingTokens func(tokens *int)
	// OnSetPermissionMode is called when the permission mode is changed.
	// Return a result to accept or reject with an error.
	OnSetPermissionMode func(mode string) *BridgePermissionModeResult
	// OnClose is called when the transport dies permanently.
	// Code meanings: 401=JWT expired, 4090=epoch superseded, 4091=init failed.
	OnClose func(code *int)
}

// BridgeSessionHandle is a per-session bridge transport handle.
// Auth is instance-scoped so multiple handles can coexist.
type BridgeSessionHandle struct {
	sessionID   string
	sequenceNum int
	connected   bool
	opts        AttachBridgeSessionOptions
	ctx         context.Context
	cancel      context.CancelFunc
}

// SessionID returns the session ID.
func (h *BridgeSessionHandle) SessionID() string {
	return h.sessionID
}

// GetSequenceNum returns the live SSE event-stream high-water mark.
func (h *BridgeSessionHandle) GetSequenceNum() int {
	return h.sequenceNum
}

// IsConnected returns true once the write path is ready.
func (h *BridgeSessionHandle) IsConnected() bool {
	return h.connected
}

// Write sends a single SDKMessage. session_id is injected automatically.
func (h *BridgeSessionHandle) Write(msg interface{}) {
	// Implementation will be filled in when process management is ready.
}

// SendResult signals a turn boundary.
func (h *BridgeSessionHandle) SendResult() {
	// Implementation placeholder.
}

// SendControlRequest forwards a permission request to claude.ai.
func (h *BridgeSessionHandle) SendControlRequest(req interface{}) {
	// Implementation placeholder.
}

// SendControlResponse forwards a permission response back through the bridge.
func (h *BridgeSessionHandle) SendControlResponse(res interface{}) {
	// Implementation placeholder.
}

// SendControlCancelRequest tells claude.ai to dismiss a pending permission prompt.
func (h *BridgeSessionHandle) SendControlCancelRequest(requestID string) {
	// Implementation placeholder.
}

// ReconnectTransport swaps the underlying transport with a fresh JWT.
func (h *BridgeSessionHandle) ReconnectTransport(opts ReconnectOptions) error {
	// Implementation placeholder.
	return nil
}

// ReportState reports session state to the CCR /worker endpoint.
func (h *BridgeSessionHandle) ReportState(state SessionState) {
	// Implementation placeholder.
}

// ReportMetadata reports external metadata (branch, dir shown on claude.ai).
func (h *BridgeSessionHandle) ReportMetadata(metadata map[string]interface{}) {
	// Implementation placeholder.
}

// ReportDelivery reports event delivery status.
func (h *BridgeSessionHandle) ReportDelivery(eventID string, status string) {
	// Implementation placeholder.
}

// Flush drains the write queue. Call before Close when delivery matters.
func (h *BridgeSessionHandle) Flush() error {
	// Implementation placeholder.
	return nil
}

// Close closes the bridge session handle.
func (h *BridgeSessionHandle) Close() {
	if h.cancel != nil {
		h.cancel()
	}
	h.connected = false
}
