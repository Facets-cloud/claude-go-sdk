// Resume Generator — Claude Agent SDK for Go
//
// Uses web search to research a person and generates a professional
// 1-page .docx resume. Demonstrates query() with system prompts,
// allowed tools, and message streaming.
//
// Equivalent to: https://github.com/anthropics/claude-agent-sdk-demos/tree/main/resume-generator
//
// Usage: go run ./demos/resume-generator "Person Name"
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	claudeagent "github.com/anthropics/claude-agent-sdk-go"
)

const systemPrompt = `You are a professional resume writer. Research a person and create a 1-page .docx resume.

WORKFLOW:
1. WebSearch for the person's background (LinkedIn, GitHub, company pages)
2. Create a .docx file using available tools

OUTPUT:
- Resume: resume.docx in the current directory

PAGE FIT (must be exactly 1 page):
- 0.5 inch margins, Name 24pt, Headers 12pt, Body 10pt
- 2-3 bullet points per job, ~80-100 chars each
- Max 3 job roles, 2-line summary, 2-line skills`

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./demos/resume-generator \"Person Name\"")
		fmt.Println("Example: go run ./demos/resume-generator \"Jane Doe\"")
		os.Exit(1)
	}

	personName := os.Args[1]
	fmt.Printf("\nGenerating resume for: %s\n", personName)
	fmt.Println(repeat("=", 50))

	prompt := fmt.Sprintf(
		`Research "%s" and create a professional 1-page resume as a .docx file. Search for their professional background, experience, education, and skills.`,
		personName,
	)

	fmt.Println("\nResearching and creating resume...")

	cwd, _ := os.Getwd()
	q := claudeagent.NewQuery(claudeagent.QueryParams{
		Prompt: prompt,
		Options: &claudeagent.Options{
			MaxTurns:     claudeagent.Int(30),
			Cwd:          &cwd,
			Model:        claudeagent.String("sonnet"),
			SystemPrompt: systemPrompt,
			AllowedTools: []string{
				"WebSearch", "WebFetch", "Bash", "Write", "Read", "Glob",
			},
			SettingSources: []claudeagent.SettingSource{
				claudeagent.SettingSourceProject,
			},
		},
	})
	defer q.Close()

	for msg := range q.Messages() {
		switch m := msg.(type) {
		case *claudeagent.SDKAssistantMessage:
			var parsed struct {
				Content []struct {
					Type  string                 `json:"type"`
					Text  string                 `json:"text"`
					Name  string                 `json:"name"`
					Input map[string]interface{} `json:"input"`
				} `json:"content"`
			}
			if err := json.Unmarshal(m.Message, &parsed); err == nil {
				for _, block := range parsed.Content {
					switch block.Type {
					case "text":
						fmt.Println(block.Text)
					case "tool_use":
						if block.Name == "WebSearch" {
							if q, ok := block.Input["query"].(string); ok {
								fmt.Printf("\nSearching: \"%s\"\n", q)
							}
						} else {
							fmt.Printf("\nUsing tool: %s\n", block.Name)
						}
					}
				}
			}

		case *claudeagent.SDKResultSuccess:
			expectedPath := filepath.Join(cwd, "resume.docx")
			if _, err := os.Stat(expectedPath); err == nil {
				fmt.Println("\n" + repeat("=", 50))
				fmt.Printf("Resume saved to: %s\n", expectedPath)
				fmt.Println(repeat("=", 50))
			} else {
				fmt.Println("\nResume file was not created. Check output above for errors.")
			}
			fmt.Printf("\nCost: $%.4f, Turns: %d\n", m.TotalCostUSD, m.NumTurns)

		case *claudeagent.SDKResultError:
			fmt.Println("\nError:", m.Errors)
		}
	}
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
