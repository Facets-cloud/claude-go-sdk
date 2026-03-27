// WebSocket Server Demo — tests SDK from a server context
//
// Tests the exact scenario the user reports: SDK used from a WebSocket server
// where the process environment, working directory, and pipe handling differ
// from a terminal.
//
// Usage: go run ./demos/ws-server
// Then: wscat -c ws://localhost:8765/ws
// Type a message and press enter to chat.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	claudeagent "github.com/anthropics/claude-agent-sdk-go"
	"golang.org/x/net/websocket"
)

func main() {
	http.Handle("/ws", websocket.Handler(handleWS))
	http.HandleFunc("/test", handleTest)

	fmt.Println("WebSocket server starting on :8765")
	fmt.Println("  WS endpoint: ws://localhost:8765/ws")
	fmt.Println("  HTTP test:   http://localhost:8765/test")
	log.Fatal(http.ListenAndServe(":8765", nil))
}

// handleTest runs a simple SDK query without WebSocket (HTTP GET)
func handleTest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, "Starting SDK query from HTTP handler...")

	q := claudeagent.NewQuery(claudeagent.QueryParams{
		Prompt: "Say PONG and nothing else",
		Options: &claudeagent.Options{
			MaxTurns: claudeagent.Int(1),
			Model:    claudeagent.String("sonnet"),
		},
	})
	defer q.Close()

	count := 0
	for msg := range q.Messages() {
		count++
		fmt.Fprintf(w, "Message %d: type=%s\n", count, msg.MessageType())
		if m, ok := msg.(*claudeagent.SDKResultSuccess); ok {
			fmt.Fprintf(w, "Result: %q (cost=$%.4f)\n", m.Result, m.TotalCostUSD)
		}
	}
	fmt.Fprintf(w, "\nTotal: %d messages\n", count)
	if count == 0 {
		fmt.Fprintln(w, "BUG: 0 messages received!")
	}
}

// handleWS handles WebSocket connections for chat
func handleWS(ws *websocket.Conn) {
	defer ws.Close()
	log.Printf("[WS] Client connected: %s", ws.Request().RemoteAddr)

	var mu sync.Mutex
	sendJSON := func(v interface{}) {
		mu.Lock()
		defer mu.Unlock()
		data, _ := json.Marshal(v)
		websocket.Message.Send(ws, string(data))
	}

	for {
		var msgText string
		if err := websocket.Message.Receive(ws, &msgText); err != nil {
			log.Printf("[WS] Client disconnected: %v", err)
			return
		}

		log.Printf("[WS] Received: %s", msgText)
		sendJSON(map[string]string{"type": "status", "message": "Starting query..."})

		// Create SDK query from WebSocket context
		q := claudeagent.NewQuery(claudeagent.QueryParams{
			Prompt: msgText,
			Options: &claudeagent.Options{
				MaxTurns: claudeagent.Int(3),
				Model:    claudeagent.String("sonnet"),
			},
		})

		count := 0
		for msg := range q.Messages() {
			count++
			log.Printf("[WS] SDK message %d: type=%s", count, msg.MessageType())

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
							sendJSON(map[string]string{"type": "text", "content": block.Text})
						}
					}
				}
			case *claudeagent.SDKResultSuccess:
				sendJSON(map[string]interface{}{
					"type":   "result",
					"result": m.Result,
					"cost":   m.TotalCostUSD,
					"turns":  m.NumTurns,
				})
			case *claudeagent.SDKResultError:
				sendJSON(map[string]interface{}{
					"type":   "error",
					"errors": m.Errors,
				})
			}
		}

		q.Close()
		log.Printf("[WS] Query complete: %d messages", count)

		if count == 0 {
			sendJSON(map[string]string{"type": "error", "message": "BUG: 0 messages from SDK"})
		}
	}
}
