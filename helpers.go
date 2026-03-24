package claudeagent

import (
	"encoding/json"
	"fmt"
)

// String returns a pointer to the given string. Convenience for optional fields.
func String(s string) *string { return &s }

// Int returns a pointer to the given int.
func Int(i int) *int { return &i }

// Bool returns a pointer to the given bool.
func Bool(b bool) *bool { return &b }

// Float64 returns a pointer to the given float64.
func Float64(f float64) *float64 { return &f }

// mustMarshal marshals v to JSON, panicking on error. Internal helper.
func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("mustMarshal: %v", err))
	}
	return data
}
