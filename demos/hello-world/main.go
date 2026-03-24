// Hello World — Claude Agent SDK for Go
//
// A simple demo that sends a greeting to Claude with a PreToolUse hook
// that restricts script file writes to the custom_scripts directory.
//
// Equivalent to: https://github.com/anthropics/claude-agent-sdk-demos/tree/main/hello-world
//
// Usage: go run ./demos/hello-world
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	claudeagent "github.com/anthropics/claude-agent-sdk-go"
)

func main() {
	cwd, _ := os.Getwd()
	agentDir := filepath.Join(cwd, "agent")

	q := claudeagent.NewQuery(claudeagent.QueryParams{
		Prompt: "Hello, Claude! Please introduce yourself in one sentence.",
		Options: &claudeagent.Options{
			MaxTurns: claudeagent.Int(100),
			Cwd:      &agentDir,
			Model:    claudeagent.String("sonnet"),
			AllowedTools: []string{
				"Bash", "Glob", "Grep", "Read", "Edit", "Write",
				"WebFetch", "WebSearch",
			},
			Hooks: map[claudeagent.HookEvent][]claudeagent.HookCallbackMatcher{
				claudeagent.HookEventPreToolUse: {
					{
						Matcher: claudeagent.String("Write|Edit"),
						Hooks: []claudeagent.HookCallback{
							restrictScriptWrites(agentDir),
						},
					},
				},
			},
		},
	})
	defer q.Close()

	for msg := range q.Messages() {
		switch m := msg.(type) {
		case *claudeagent.SDKAssistantMessage:
			var parsed struct {
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			}
			if err := json.Unmarshal(m.Message, &parsed); err == nil {
				for _, block := range parsed.Content {
					if block.Type == "text" {
						fmt.Println("Claude says:", block.Text)
					}
				}
			}

		case *claudeagent.SDKResultSuccess:
			fmt.Printf("\nDone! Cost: $%.4f, Turns: %d\n", m.TotalCostUSD, m.NumTurns)

		case *claudeagent.SDKResultError:
			fmt.Println("Error:", m.Errors)
		}
	}
}

// restrictScriptWrites returns a hook that blocks .js/.ts writes outside custom_scripts/.
func restrictScriptWrites(agentDir string) claudeagent.HookCallback {
	customScriptsPath := filepath.Join(agentDir, "custom_scripts")

	return func(ctx context.Context, input claudeagent.HookInput, toolUseID *string) (claudeagent.HookJSONOutput, error) {
		preInput, ok := input.(*claudeagent.PreToolUseHookInput)
		if !ok {
			return claudeagent.SyncHookJSONOutput{Continue: claudeagent.Bool(true)}, nil
		}

		toolName := preInput.ToolName
		if toolName != "Write" && toolName != "Edit" {
			return claudeagent.SyncHookJSONOutput{Continue: claudeagent.Bool(true)}, nil
		}

		// Extract file_path from tool input
		var toolInput struct {
			FilePath string `json:"file_path"`
		}
		if raw, err := json.Marshal(preInput.ToolInput); err == nil {
			json.Unmarshal(raw, &toolInput)
		}

		ext := strings.ToLower(filepath.Ext(toolInput.FilePath))
		if (ext == ".js" || ext == ".ts") && !strings.HasPrefix(toolInput.FilePath, customScriptsPath) {
			return claudeagent.SyncHookJSONOutput{
				Continue: claudeagent.Bool(false),
				Decision: claudeagent.String("block"),
				Reason: claudeagent.String(fmt.Sprintf(
					"Script files must be written to %s/%s",
					customScriptsPath, filepath.Base(toolInput.FilePath),
				)),
			}, nil
		}

		return claudeagent.SyncHookJSONOutput{Continue: claudeagent.Bool(true)}, nil
	}
}
