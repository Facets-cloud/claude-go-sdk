package claudeagent

import "fmt"

// SDKSession is the interface for multi-turn V2 sessions (unstable/alpha API).
type SDKSession interface {
	// SessionID returns the session UUID.
	SessionID() string

	// Send sends a message to the agent.
	Send(message string) error

	// Stream returns a channel that streams SDKMessage values from the agent.
	Stream() <-chan SDKMessage

	// Close terminates the session.
	Close()
}

// sdkSession is the concrete V2 session implementation backed by a Query.
type sdkSession struct {
	query     *Query
	input     chan SDKUserMessage
	sessionID string
}

// CreateSession creates a new V2 session.
func CreateSession(opts SDKSessionOptions) (SDKSession, error) {
	input := make(chan SDKUserMessage, 1)
	q := NewQuery(QueryParams{
		Prompt: input,
		Options: &Options{
			Model:           &opts.Model,
			AllowedTools:    opts.AllowedTools,
			DisallowedTools: opts.DisallowedTools,
			PermissionMode:  opts.PermissionMode,
		},
	})
	return &sdkSession{query: q, input: input}, nil
}

// ResumeSession resumes an existing V2 session by ID.
func ResumeSession(sessionID string, opts SDKSessionOptions) (SDKSession, error) {
	input := make(chan SDKUserMessage, 1)
	q := NewQuery(QueryParams{
		Prompt: input,
		Options: &Options{
			Model:          &opts.Model,
			Resume:         &sessionID,
			AllowedTools:   opts.AllowedTools,
			DisallowedTools: opts.DisallowedTools,
			PermissionMode: opts.PermissionMode,
		},
	})
	return &sdkSession{query: q, input: input, sessionID: sessionID}, nil
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

func (s *sdkSession) Send(message string) error {
	s.input <- SDKUserMessage{
		Type:      "user",
		Message:   mustMarshal(map[string]interface{}{"role": "user", "content": message}),
		SessionID: s.sessionID,
	}
	return nil
}

func (s *sdkSession) Stream() <-chan SDKMessage {
	return s.query.Messages()
}

func (s *sdkSession) Close() {
	close(s.input)
	s.query.Close()
}
