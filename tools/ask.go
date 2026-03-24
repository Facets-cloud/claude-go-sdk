package tools

// AskQuestionOption represents a selectable option for a question.
type AskQuestionOption struct {
	// Label is the display text for the option.
	Label string `json:"label"`
	// Description explains what this option means.
	Description string `json:"description"`
	// Preview is optional content rendered when this option is focused.
	Preview *string `json:"preview,omitempty"`
}

// AskQuestion represents a single question to ask the user.
type AskQuestion struct {
	// Question is the complete question text.
	Question string `json:"question"`
	// Header is a short label displayed as a chip/tag (max 12 chars).
	Header string `json:"header"`
	// Options is the list of available choices (2-4 options).
	Options []AskQuestionOption `json:"options"`
	// MultiSelect allows selecting multiple options when true.
	MultiSelect bool `json:"multiSelect"`
}

// AskUserQuestionInput is the input for the AskUserQuestion tool.
type AskUserQuestionInput struct {
	// Questions is the list of questions to ask (1-4 questions).
	Questions []AskQuestion `json:"questions"`
}

// AskAnnotation contains per-question annotations from the user.
type AskAnnotation struct {
	// Preview is the preview content of the selected option.
	Preview *string `json:"preview,omitempty"`
	// Notes is free-text notes the user added.
	Notes *string `json:"notes,omitempty"`
}

// AskUserQuestionOutput is the output from the AskUserQuestion tool.
type AskUserQuestionOutput struct {
	// Questions contains the questions that were asked.
	Questions []AskQuestion `json:"questions"`
	// Answers maps question text to answer string.
	Answers map[string]string `json:"answers"`
	// Annotations contains optional per-question annotations.
	Annotations map[string]AskAnnotation `json:"annotations,omitempty"`
}
