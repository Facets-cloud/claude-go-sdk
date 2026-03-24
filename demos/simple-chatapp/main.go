// Simple Chat App — Claude Agent SDK for Go
//
// A terminal-based multi-turn chat application that demonstrates
// streaming input via channels. Type messages, see Claude respond
// in real-time, and maintain conversation context across turns.
//
// Equivalent to the server-side logic of:
// https://github.com/anthropics/claude-agent-sdk-demos/tree/main/simple-chatapp
//
// Usage: go run ./demos/simple-chatapp
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	claudeagent "github.com/anthropics/claude-agent-sdk-go"
)

func main() {
	fmt.Println("Claude Chat (Go SDK)")
	fmt.Println("Type your messages. Type /quit to exit.")
	fmt.Println(strings.Repeat("-", 40))

	// Create a streaming input channel for multi-turn conversation.
	input := make(chan claudeagent.SDKUserMessage, 1)

	cwd, _ := os.Getwd()
	permMode := claudeagent.PermissionModeDontAsk
	q := claudeagent.NewQuery(claudeagent.QueryParams{
		Prompt: input, // Channel-based streaming input
		Options: &claudeagent.Options{
			Cwd:            &cwd,
			Model:          claudeagent.String("sonnet"),
			SystemPrompt:   "You are in a conversational chat. Be concise and helpful.",
			PermissionMode: &permMode,
		},
	})
	defer q.Close()

	// Read messages from the channel in a goroutine.
	go func() {
		for msg := range q.Messages() {
			switch m := msg.(type) {
			case *claudeagent.SDKSystemMessage:
				fmt.Printf("[connected] model=%s, session=%s\n\n", m.Model, m.SessionID)

			case *claudeagent.SDKAssistantMessage:
				var parsed struct {
					Content []struct {
						Type string `json:"type"`
						Text string `json:"text"`
						Name string `json:"name"`
					} `json:"content"`
				}
				if err := json.Unmarshal(m.Message, &parsed); err == nil {
					for _, block := range parsed.Content {
						switch block.Type {
						case "text":
							fmt.Printf("\nClaude: %s\n\n", block.Text)
						case "tool_use":
							fmt.Printf("[tool: %s]\n", block.Name)
						}
					}
				}

			case *claudeagent.SDKResultSuccess:
				fmt.Printf("[turn complete - $%.4f]\n", m.TotalCostUSD)
				fmt.Print("You: ")

			case *claudeagent.SDKResultError:
				fmt.Printf("\n[error: %v]\n", m.Errors)
				fmt.Print("You: ")

			case *claudeagent.SDKToolProgressMessage:
				fmt.Printf("[%s running %.0fs...]\n", m.ToolName, m.ElapsedTimeSeconds)
			}
		}
	}()

	// Read user input from stdin and send to the query.
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("You: ")
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			fmt.Print("You: ")
			continue
		}
		if text == "/quit" || text == "/exit" {
			fmt.Println("Goodbye!")
			close(input)
			return
		}

		msgJSON, _ := json.Marshal(map[string]interface{}{
			"role":    "user",
			"content": text,
		})

		input <- claudeagent.SDKUserMessage{
			Type:      "user",
			Message:   msgJSON,
			SessionID: "chat",
		}
	}

	close(input)
}
