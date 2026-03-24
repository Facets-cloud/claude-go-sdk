package tools

// WebFetchInput is the input for the WebFetch tool.
type WebFetchInput struct {
	// URL is the URL to fetch content from.
	URL string `json:"url"`
	// Prompt is the prompt to run on the fetched content.
	Prompt string `json:"prompt"`
}

// WebFetchOutput is the output from the WebFetch tool.
type WebFetchOutput struct {
	// Bytes is the size of the fetched content in bytes.
	Bytes int `json:"bytes"`
	// Code is the HTTP response code.
	Code int `json:"code"`
	// CodeText is the HTTP response code text.
	CodeText string `json:"codeText"`
	// Result is the processed result from applying the prompt.
	Result string `json:"result"`
	// DurationMs is the time taken to fetch and process.
	DurationMs float64 `json:"durationMs"`
	// URL is the URL that was fetched.
	URL string `json:"url"`
}

// WebSearchInput is the input for the WebSearch tool.
type WebSearchInput struct {
	// Query is the search query.
	Query string `json:"query"`
	// AllowedDomains limits results to these domains.
	AllowedDomains []string `json:"allowed_domains,omitempty"`
	// BlockedDomains excludes results from these domains.
	BlockedDomains []string `json:"blocked_domains,omitempty"`
}

// WebSearchHit represents a single search result.
type WebSearchHit struct {
	// Title is the title of the search result.
	Title string `json:"title"`
	// URL is the URL of the search result.
	URL string `json:"url"`
}

// WebSearchToolResult is a structured search result from a tool use.
type WebSearchToolResult struct {
	// ToolUseID is the ID of the tool use.
	ToolUseID string `json:"tool_use_id"`
	// Content contains the search hits.
	Content []WebSearchHit `json:"content"`
}

// WebSearchOutput is the output from the WebSearch tool.
type WebSearchOutput struct {
	// Query is the search query that was executed.
	Query string `json:"query"`
	// Results contains search results. Each element is either a
	// WebSearchToolResult or a string commentary.
	Results []interface{} `json:"results"`
	// DurationSeconds is the time taken to complete the search.
	DurationSeconds float64 `json:"durationSeconds"`
}
