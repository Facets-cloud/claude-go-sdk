package claudeagent

import (
	"errors"
	"testing"
)

func TestAbortError(t *testing.T) {
	err := &AbortError{Message: "operation cancelled"}
	if err.Error() != "operation cancelled" {
		t.Errorf("got %q, want %q", err.Error(), "operation cancelled")
	}

	var target *AbortError
	if !errors.As(err, &target) {
		t.Error("AbortError should be matchable with errors.As")
	}
}
