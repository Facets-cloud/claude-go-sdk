package claudeagent

// AbortError is returned when an operation is cancelled via context cancellation.
type AbortError struct {
	Message string
}

func (e *AbortError) Error() string {
	return e.Message
}
