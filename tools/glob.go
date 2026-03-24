package tools

// GlobInput is the input for the Glob tool.
type GlobInput struct {
	// Pattern is the glob pattern to match files against.
	Pattern string `json:"pattern"`
	// Path is the directory to search in. Defaults to current working directory.
	Path *string `json:"path,omitempty"`
}

// GlobOutput is the output from the Glob tool.
type GlobOutput struct {
	// DurationMs is the time taken in milliseconds.
	DurationMs float64 `json:"durationMs"`
	// NumFiles is the total number of files found.
	NumFiles int `json:"numFiles"`
	// Filenames is the array of matching file paths.
	Filenames []string `json:"filenames"`
	// Truncated indicates whether results were truncated.
	Truncated bool `json:"truncated"`
}
