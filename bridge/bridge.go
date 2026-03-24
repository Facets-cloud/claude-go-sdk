package bridge

import (
	"context"
	"fmt"
)

// AttachBridgeSession attaches to an existing bridge session and returns
// a handle scoped to that session.
//
// This is an ALPHA API. Breaking changes here do NOT bump the package major.
func AttachBridgeSession(opts AttachBridgeSessionOptions) (*BridgeSessionHandle, error) {
	if opts.SessionID == "" {
		return nil, fmt.Errorf("bridge: sessionId is required")
	}
	if opts.IngressToken == "" {
		return nil, fmt.Errorf("bridge: ingressToken is required")
	}
	if opts.APIBaseURL == "" {
		return nil, fmt.Errorf("bridge: apiBaseUrl is required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	handle := &BridgeSessionHandle{
		sessionID: opts.SessionID,
		connected: false,
		opts:      opts,
		ctx:       ctx,
		cancel:    cancel,
	}

	if opts.InitialSequenceNum != nil {
		handle.sequenceNum = *opts.InitialSequenceNum
	}

	// TODO: Initialize v2 transport (SSETransport + CCRClient),
	// wire ingress routing and control dispatch.
	// For now, mark as connected for type/API completeness.
	handle.connected = true

	return handle, nil
}
