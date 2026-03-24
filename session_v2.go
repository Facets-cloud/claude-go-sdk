package claudeagent

// SDKSession is the interface for multi-turn V2 sessions (unstable/alpha API).
// Implementation requires Query (Task #13), so methods are defined but not yet
// wired to a concrete implementation.
type SDKSession interface {
	// SessionID returns the session UUID. Available after first message;
	// for resumed sessions, available immediately.
	SessionID() string

	// Send sends a message to the agent.
	Send(message string) error

	// Messages returns a channel that streams SDKMessage values from the agent.
	Messages() <-chan SDKMessage

	// Close terminates the session.
	Close()
}

// SDKSessionOptions is defined in session.go.

// TODO: Implement these functions once Query (Task #13) is complete.
//
// func CreateSession(opts SDKSessionOptions) SDKSession { ... }
// func ResumeSession(sessionID string, opts SDKSessionOptions) SDKSession { ... }
// func Prompt(message string, opts SDKSessionOptions) (*SDKResultMessage, error) { ... }
