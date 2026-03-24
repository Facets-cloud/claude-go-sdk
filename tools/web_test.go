package tools

import (
	"encoding/json"
	"testing"
)

func TestWebFetchInput_JSON(t *testing.T) {
	input := WebFetchInput{
		URL:    "https://example.com",
		Prompt: "Summarize this page",
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got WebFetchInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.URL != "https://example.com" {
		t.Errorf("URL = %q", got.URL)
	}
}

func TestWebFetchOutput_JSON(t *testing.T) {
	raw := `{
		"bytes": 1024,
		"code": 200,
		"codeText": "OK",
		"result": "Summary of page",
		"durationMs": 500.5,
		"url": "https://example.com"
	}`
	var out WebFetchOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.Code != 200 {
		t.Errorf("Code = %d", out.Code)
	}
	if out.DurationMs != 500.5 {
		t.Errorf("DurationMs = %f", out.DurationMs)
	}
}

func TestWebSearchInput_JSON(t *testing.T) {
	input := WebSearchInput{
		Query:          "golang best practices",
		AllowedDomains: []string{"go.dev", "github.com"},
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got WebSearchInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.AllowedDomains) != 2 {
		t.Errorf("AllowedDomains len = %d", len(got.AllowedDomains))
	}
}

func TestWebSearchOutput_JSON(t *testing.T) {
	raw := `{
		"query": "test query",
		"results": [
			{"tool_use_id": "tu-1", "content": [{"title": "Result 1", "url": "https://example.com"}]},
			"Some commentary text"
		],
		"durationSeconds": 1.5
	}`
	var out WebSearchOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.Query != "test query" {
		t.Errorf("Query = %q", out.Query)
	}
	if len(out.Results) != 2 {
		t.Errorf("Results len = %d", len(out.Results))
	}
	if out.DurationSeconds != 1.5 {
		t.Errorf("DurationSeconds = %f", out.DurationSeconds)
	}
}
