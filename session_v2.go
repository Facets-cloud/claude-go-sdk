package claudeagent

import (
	"encoding/json"
	"fmt"
)

// SDKSession is the interface for multi-turn V2 sessions (unstable/alpha API).
type SDKSession interface {
	// SessionID returns the session UUID.
	SessionID() string

	// Send sends a message to the agent. Accepts a string or SDKUserMessage.
	Send(message interface{}) error

	// Stream returns a channel that streams SDKMessage values from the agent.
	Stream() <-chan SDKMessage

	// Close terminates the session.
	Close()
}

// sdkSession is the concrete V2 session implementation backed by a Query.
type sdkSession struct {
	query       *Query
	sessionID   string
	firstSend   bool
	pendingOpts SDKSessionOptions
}

// CreateSession creates a new V2 session.
// The session is lazy — the underlying query starts on the first Send() call.
func CreateSession(opts SDKSessionOptions) (SDKSession, error) {
	return &sdkSession{pendingOpts: opts, firstSend: true}, nil
}

// ResumeSession resumes an existing V2 session by ID.
func ResumeSession(sessionID string, opts SDKSessionOptions) (SDKSession, error) {
	return &sdkSession{
		pendingOpts: opts,
		sessionID:   sessionID,
		firstSend:   true,
	}, nil
}

// Prompt is a convenience function for single-turn V2 queries.
// Creates a session, sends the prompt, collects the result, and closes.
func Prompt(message string, opts SDKSessionOptions) (SDKMessage, error) {
	sess, err := CreateSession(opts)
	if err != nil {
		return nil, err
	}
	defer sess.Close()

	if err := sess.Send(message); err != nil {
		return nil, err
	}

	var result SDKMessage
	for msg := range sess.Stream() {
		if IsResultMessage(msg) {
			result = msg
		}
	}
	if result == nil {
		return nil, fmt.Errorf("no result received")
	}
	return result, nil
}

func (s *sdkSession) SessionID() string {
	return s.sessionID
}

func (s *sdkSession) Send(message interface{}) error {
	var prompt string
	switch m := message.(type) {
	case string:
		prompt = m
	case SDKUserMessage:
		// Extract text content from structured message
		var parsed struct {
			Content string `json:"content"`
		}
		json.Unmarshal(m.Message, &parsed)
		prompt = parsed.Content
	case *SDKUserMessage:
		var parsed struct {
			Content string `json:"content"`
		}
		json.Unmarshal(m.Message, &parsed)
		prompt = parsed.Content
	default:
		return fmt.Errorf("Send: unsupported message type %T, expected string or SDKUserMessage", message)
	}

	if s.firstSend {
		// Start the query with the first message as the prompt
		s.firstSend = false
		opts := &Options{
			Model:           &s.pendingOpts.Model,
			AllowedTools:    s.pendingOpts.AllowedTools,
			DisallowedTools: s.pendingOpts.DisallowedTools,
			PermissionMode:  s.pendingOpts.PermissionMode,
		}
		if s.sessionID != "" {
			opts.Resume = &s.sessionID
		}
		s.query = NewQuery(QueryParams{
			Prompt:  prompt,
			Options: opts,
		})
		return nil
	}

	// For subsequent messages, we'd need streaming input mode.
	// In --print mode, multi-turn is not supported within a single CLI invocation.
	return fmt.Errorf("multi-turn Send not yet supported in print mode; use NewQuery with channel input")
}

func (s *sdkSession) Stream() <-chan SDKMessage {
	if s.query == nil {
		ch := make(chan SDKMessage)
		close(ch)
		return ch
	}
	return s.query.Messages()
}

func (s *sdkSession) Close() {
	if s.query != nil {
		s.query.Close()
	}
}
