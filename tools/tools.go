// Package tools defines the input and output types for all Claude Code CLI tools.
// These types correspond to the JSON schemas in the TypeScript SDK's sdk-tools.d.ts.
//
// Each tool has an Input struct (parameters sent to the tool) and an Output struct
// (the result returned). Union-type outputs (e.g., AgentOutput, FileReadOutput)
// use Go interfaces with concrete variant structs.
package tools
