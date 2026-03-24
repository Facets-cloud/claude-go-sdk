// Package claudeagent provides a Go SDK for building AI agents with Claude Code's
// capabilities. It enables programmatic interaction with Claude to build autonomous
// agents that can understand codebases, edit files, run commands, and execute
// complex workflows.
//
// The SDK works by spawning the Claude Code CLI as a subprocess and communicating
// via JSON over stdin/stdout. Messages are streamed to the caller via Go channels.
//
// Basic usage:
//
//	q := claudeagent.Query(claudeagent.QueryParams{
//	    Prompt: "Explain the main function in this project",
//	})
//	for msg := range q.Messages() {
//	    fmt.Println(msg)
//	}
package claudeagent

// Version is the SDK version, tracking parity with the TypeScript SDK.
const Version = "0.2.81"

// ClaudeCodeVersion is the minimum compatible Claude Code CLI version.
const ClaudeCodeVersion = "2.1.81"
