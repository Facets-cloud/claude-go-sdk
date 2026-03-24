// Hello World V2 — Claude Agent SDK for Go (V2 Session API)
//
// Demonstrates the V2 session-based API with separate Send()/Stream(),
// multi-turn conversations, one-shot prompts, and session resume.
//
// Equivalent to: https://github.com/anthropics/claude-agent-sdk-demos/tree/main/hello-world-v2
//
// Usage:
//
//	go run ./demos/hello-world-v2 basic
//	go run ./demos/hello-world-v2 multi-turn
//	go run ./demos/hello-world-v2 one-shot
//	go run ./demos/hello-world-v2 resume
package main

import (
	"encoding/json"
	"fmt"
	"os"

	claudeagent "github.com/anthropics/claude-agent-sdk-go"
)

func main() {
	example := "basic"
	if len(os.Args) > 1 {
		example = os.Args[1]
	}

	switch example {
	case "basic":
		basicSession()
	case "multi-turn":
		multiTurn()
	case "one-shot":
		oneShot()
	case "resume":
		sessionResume()
	default:
		fmt.Println("Usage: go run ./demos/hello-world-v2 [basic|multi-turn|one-shot|resume]")
	}
}

// extractText pulls the first text block from a raw assistant message.
func extractText(raw json.RawMessage) string {
	var parsed struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &parsed); err == nil {
		for _, block := range parsed.Content {
			if block.Type == "text" {
				return block.Text
			}
		}
	}
	return ""
}

// basicSession demonstrates a simple send/stream pattern.
func basicSession() {
	fmt.Println("=== Basic Session ===")

	sess, err := claudeagent.CreateSession(claudeagent.SDKSessionOptions{
		Model: "sonnet",
	})
	if err != nil {
		fmt.Println("Error creating session:", err)
		return
	}
	defer sess.Close()

	sess.Send("Hello! Introduce yourself in one sentence.")

	for msg := range sess.Stream() {
		if m, ok := msg.(*claudeagent.SDKAssistantMessage); ok {
			if text := extractText(m.Message); text != "" {
				fmt.Println("Claude:", text)
			}
		}
	}
}

// multiTurn shows V2's key advantage: multi-turn conversations.
func multiTurn() {
	fmt.Println("=== Multi-Turn Conversation ===")

	sess, err := claudeagent.CreateSession(claudeagent.SDKSessionOptions{
		Model: "sonnet",
	})
	if err != nil {
		fmt.Println("Error creating session:", err)
		return
	}
	defer sess.Close()

	// Turn 1
	sess.Send("What is 5 + 3? Just the number.")
	for msg := range sess.Stream() {
		if m, ok := msg.(*claudeagent.SDKAssistantMessage); ok {
			if text := extractText(m.Message); text != "" {
				fmt.Println("Turn 1:", text)
			}
		}
	}

	// Turn 2 — Claude remembers context
	sess.Send("Multiply that by 2. Just the number.")
	for msg := range sess.Stream() {
		if m, ok := msg.(*claudeagent.SDKAssistantMessage); ok {
			if text := extractText(m.Message); text != "" {
				fmt.Println("Turn 2:", text)
			}
		}
	}
}

// oneShot uses the Prompt() convenience function for single-turn queries.
func oneShot() {
	fmt.Println("=== One-Shot Prompt ===")

	result, err := claudeagent.Prompt("What is the capital of France? One word.", claudeagent.SDKSessionOptions{
		Model: "sonnet",
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if r, ok := result.(*claudeagent.SDKResultSuccess); ok {
		fmt.Printf("Answer: %s\n", r.Result)
		fmt.Printf("Cost: $%.4f\n", r.TotalCostUSD)
	}
}

// sessionResume demonstrates persisting context across sessions.
func sessionResume() {
	fmt.Println("=== Session Resume ===")

	var sessionID string

	// First session — establish a memory
	{
		sess, err := claudeagent.CreateSession(claudeagent.SDKSessionOptions{
			Model: "sonnet",
		})
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Println("[Session 1] Telling Claude my favorite color...")
		sess.Send("My favorite color is blue. Remember this!")

		for msg := range sess.Stream() {
			switch m := msg.(type) {
			case *claudeagent.SDKSystemMessage:
				sessionID = m.SessionID
				fmt.Printf("[Session 1] ID: %s\n", sessionID)
			case *claudeagent.SDKAssistantMessage:
				if text := extractText(m.Message); text != "" {
					fmt.Printf("[Session 1] Claude: %s\n\n", text)
				}
			}
		}
		sess.Close()
	}

	fmt.Println("--- Session closed. Time passes... ---")

	// Resume and verify Claude remembers
	{
		sess, err := claudeagent.ResumeSession(sessionID, claudeagent.SDKSessionOptions{
			Model: "sonnet",
		})
		if err != nil {
			fmt.Println("Error resuming:", err)
			return
		}
		defer sess.Close()

		fmt.Println("[Session 2] Resuming and asking Claude...")
		sess.Send("What is my favorite color?")

		for msg := range sess.Stream() {
			if m, ok := msg.(*claudeagent.SDKAssistantMessage); ok {
				if text := extractText(m.Message); text != "" {
					fmt.Printf("[Session 2] Claude: %s\n", text)
				}
			}
		}
	}
}
