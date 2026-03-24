package tools

// GrepInput is the input for the Grep tool.
type GrepInput struct {
	// Pattern is the regex pattern to search for.
	Pattern string `json:"pattern"`
	// Path is the file or directory to search in.
	Path *string `json:"path,omitempty"`
	// Glob is a glob pattern to filter files.
	Glob *string `json:"glob,omitempty"`
	// OutputMode controls output format: "content", "files_with_matches", or "count".
	OutputMode *string `json:"output_mode,omitempty"`
	// BeforeContext is lines to show before each match (rg -B).
	BeforeContext *int `json:"-B,omitempty"`
	// AfterContext is lines to show after each match (rg -A).
	AfterContext *int `json:"-A,omitempty"`
	// ContextAlias is an alias for Context (rg -C).
	ContextAlias *int `json:"-C,omitempty"`
	// Context is lines to show before and after each match.
	Context *int `json:"context,omitempty"`
	// LineNumbers shows line numbers in output (rg -n).
	LineNumbers *bool `json:"-n,omitempty"`
	// CaseInsensitive enables case-insensitive search (rg -i).
	CaseInsensitive *bool `json:"-i,omitempty"`
	// Type is the file type to search (rg --type).
	Type *string `json:"type,omitempty"`
	// HeadLimit limits output to first N lines/entries.
	HeadLimit *int `json:"head_limit,omitempty"`
	// Offset skips first N lines/entries before applying HeadLimit.
	Offset *int `json:"offset,omitempty"`
	// Multiline enables multiline matching mode.
	Multiline *bool `json:"multiline,omitempty"`
}

// GrepOutput is the output from the Grep tool.
type GrepOutput struct {
	// Mode is the output mode that was used.
	Mode *string `json:"mode,omitempty"`
	// NumFiles is the number of files with matches.
	NumFiles int `json:"numFiles"`
	// Filenames is the list of matching file paths.
	Filenames []string `json:"filenames"`
	// Content is the matching content (when mode is "content").
	Content *string `json:"content,omitempty"`
	// NumLines is the number of matching lines.
	NumLines *int `json:"numLines,omitempty"`
	// NumMatches is the number of matches found.
	NumMatches *int `json:"numMatches,omitempty"`
	// AppliedLimit is the limit that was applied.
	AppliedLimit *int `json:"appliedLimit,omitempty"`
	// AppliedOffset is the offset that was applied.
	AppliedOffset *int `json:"appliedOffset,omitempty"`
}
