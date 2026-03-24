package tools

// McpInput is the input for MCP tool execution. It accepts arbitrary JSON.
type McpInput map[string]interface{}

// McpOutput is the output from MCP tool execution (a string).
type McpOutput = string

// ListMcpResourcesInput is the input for listing MCP resources.
type ListMcpResourcesInput struct {
	// Server optionally filters resources by server name.
	Server *string `json:"server,omitempty"`
}

// McpResource represents a single MCP resource entry.
type McpResource struct {
	// URI is the resource URI.
	URI string `json:"uri"`
	// Name is the resource name.
	Name string `json:"name"`
	// MimeType is the MIME type of the resource.
	MimeType *string `json:"mimeType,omitempty"`
	// Description is the resource description.
	Description *string `json:"description,omitempty"`
	// Server is the server that provides this resource.
	Server string `json:"server"`
}

// ListMcpResourcesOutput is a list of MCP resources.
type ListMcpResourcesOutput []McpResource

// ReadMcpResourceInput is the input for reading an MCP resource.
type ReadMcpResourceInput struct {
	// Server is the MCP server name.
	Server string `json:"server"`
	// URI is the resource URI to read.
	URI string `json:"uri"`
}

// McpResourceContent represents content from a read MCP resource.
type McpResourceContent struct {
	// URI is the resource URI.
	URI string `json:"uri"`
	// MimeType is the MIME type of the content.
	MimeType *string `json:"mimeType,omitempty"`
	// Text is the text content.
	Text *string `json:"text,omitempty"`
	// BlobSavedTo is the path where binary blob content was saved.
	BlobSavedTo *string `json:"blobSavedTo,omitempty"`
}

// ReadMcpResourceOutput is the output from reading an MCP resource.
type ReadMcpResourceOutput struct {
	Contents []McpResourceContent `json:"contents"`
}

// SubscribeMcpResourceInput is the input for subscribing to an MCP resource.
type SubscribeMcpResourceInput struct {
	// Server is the MCP server name.
	Server string `json:"server"`
	// URI is the resource URI to subscribe to.
	URI string `json:"uri"`
	// Reason is an optional reason for subscribing.
	Reason *string `json:"reason,omitempty"`
}

// SubscribeMcpResourceOutput is the output from subscribing to an MCP resource.
type SubscribeMcpResourceOutput struct {
	// Subscribed indicates whether the subscription was successful.
	Subscribed bool `json:"subscribed"`
	// SubscriptionID is the unique identifier for this subscription.
	SubscriptionID string `json:"subscriptionId"`
}

// UnsubscribeMcpResourceInput is the input for unsubscribing from an MCP resource.
type UnsubscribeMcpResourceInput struct {
	// Server is the MCP server name.
	Server *string `json:"server,omitempty"`
	// URI is the resource URI to unsubscribe from.
	URI *string `json:"uri,omitempty"`
	// SubscriptionID is the subscription ID to unsubscribe.
	SubscriptionID *string `json:"subscriptionId,omitempty"`
}

// UnsubscribeMcpResourceOutput is the output from unsubscribing from an MCP resource.
type UnsubscribeMcpResourceOutput struct {
	// Unsubscribed indicates whether the unsubscription was successful.
	Unsubscribed bool `json:"unsubscribed"`
}

// SubscribePollingInput is the input for subscribing to polling.
type SubscribePollingInput struct {
	// Type is "tool" to poll a tool, "resource" to poll a resource URI.
	Type string `json:"type"`
	// Server is the MCP server name.
	Server string `json:"server"`
	// ToolName is the tool to call periodically (required when type is "tool").
	ToolName *string `json:"toolName,omitempty"`
	// Arguments are passed to the tool on each call.
	Arguments map[string]interface{} `json:"arguments,omitempty"`
	// URI is the resource URI to poll (required when type is "resource").
	URI *string `json:"uri,omitempty"`
	// IntervalMs is the polling interval in milliseconds.
	IntervalMs int `json:"intervalMs"`
	// Reason is an optional reason for subscribing.
	Reason *string `json:"reason,omitempty"`
}

// SubscribePollingOutput is the output from subscribing to polling.
type SubscribePollingOutput struct {
	// Subscribed indicates whether the subscription was successful.
	Subscribed bool `json:"subscribed"`
	// SubscriptionID is the unique identifier for this subscription.
	SubscriptionID string `json:"subscriptionId"`
}

// UnsubscribePollingInput is the input for unsubscribing from polling.
type UnsubscribePollingInput struct {
	// SubscriptionID is the subscription ID to unsubscribe.
	SubscriptionID *string `json:"subscriptionId,omitempty"`
	// Server is the MCP server name.
	Server *string `json:"server,omitempty"`
	// Target is the tool name or URI to unsubscribe.
	Target *string `json:"target,omitempty"`
}

// UnsubscribePollingOutput is the output from unsubscribing from polling.
type UnsubscribePollingOutput struct {
	// Unsubscribed indicates whether the unsubscription was successful.
	Unsubscribed bool `json:"unsubscribed"`
}
