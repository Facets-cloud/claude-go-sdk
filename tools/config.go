package tools

// ConfigInput is the input for the Config tool.
type ConfigInput struct {
	// Setting is the setting key (e.g., "theme", "model").
	Setting string `json:"setting"`
	// Value is the new value. Omit to get current value.
	Value interface{} `json:"value,omitempty"`
}

// ConfigOutput is the output from the Config tool.
type ConfigOutput struct {
	// Success indicates whether the operation succeeded.
	Success bool `json:"success"`
	// Operation is "get" or "set".
	Operation *string `json:"operation,omitempty"`
	// Setting is the setting key that was accessed.
	Setting *string `json:"setting,omitempty"`
	// Value is the current value (for get operations).
	Value interface{} `json:"value,omitempty"`
	// PreviousValue is the value before the set operation.
	PreviousValue interface{} `json:"previousValue,omitempty"`
	// NewValue is the value after the set operation.
	NewValue interface{} `json:"newValue,omitempty"`
	// Error is the error message if the operation failed.
	Error *string `json:"error,omitempty"`
}
