package tools

import (
	"encoding/json"
	"testing"
)

func TestGrepInput_JSON(t *testing.T) {
	input := GrepInput{
		Pattern:    "func Test",
		Path:       strPtr("/src"),
		OutputMode: strPtr("content"),
		HeadLimit:  intPtr(10),
		Multiline:  boolPtr(false),
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got GrepInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Pattern != "func Test" {
		t.Errorf("Pattern = %q", got.Pattern)
	}
	if got.OutputMode == nil || *got.OutputMode != "content" {
		t.Errorf("OutputMode = %v", got.OutputMode)
	}
}

func TestGrepOutput_JSON(t *testing.T) {
	raw := `{
		"mode": "files_with_matches",
		"numFiles": 2,
		"filenames": ["a.go", "b.go"],
		"numMatches": 5
	}`
	var out GrepOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.NumFiles != 2 {
		t.Errorf("NumFiles = %d", out.NumFiles)
	}
	if out.NumMatches == nil || *out.NumMatches != 5 {
		t.Errorf("NumMatches = %v", out.NumMatches)
	}
}
