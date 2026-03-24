package tools

import (
	"encoding/json"
	"testing"
)

func TestFileReadInput_JSON(t *testing.T) {
	input := FileReadInput{
		FilePath: "/path/to/file.go",
		Offset:   intPtr(10),
		Limit:    intPtr(50),
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got FileReadInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.FilePath != "/path/to/file.go" {
		t.Errorf("FilePath = %q", got.FilePath)
	}
	if got.Offset == nil || *got.Offset != 10 {
		t.Errorf("Offset = %v", got.Offset)
	}
}

func TestUnmarshalFileReadOutput_Text(t *testing.T) {
	raw := `{
		"type": "text",
		"file": {
			"filePath": "/tmp/test.go",
			"content": "package main",
			"numLines": 1,
			"startLine": 1,
			"totalLines": 10
		}
	}`
	out, err := UnmarshalFileReadOutput([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	text, ok := out.(FileReadText)
	if !ok {
		t.Fatalf("expected FileReadText, got %T", out)
	}
	if text.File.FilePath != "/tmp/test.go" {
		t.Errorf("FilePath = %q", text.File.FilePath)
	}
	if text.File.NumLines != 1 {
		t.Errorf("NumLines = %d", text.File.NumLines)
	}
}

func TestUnmarshalFileReadOutput_Image(t *testing.T) {
	raw := `{
		"type": "image",
		"file": {
			"base64": "iVBORw0KGgo=",
			"type": "image/png",
			"originalSize": 1024,
			"dimensions": {"originalWidth": 800, "originalHeight": 600}
		}
	}`
	out, err := UnmarshalFileReadOutput([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	img, ok := out.(FileReadImage)
	if !ok {
		t.Fatalf("expected FileReadImage, got %T", out)
	}
	if img.File.Type != "image/png" {
		t.Errorf("Type = %q", img.File.Type)
	}
	if img.File.Dimensions == nil || *img.File.Dimensions.OriginalWidth != 800 {
		t.Error("expected dimensions with width 800")
	}
}

func TestUnmarshalFileReadOutput_Notebook(t *testing.T) {
	raw := `{
		"type": "notebook",
		"file": {"filePath": "/tmp/test.ipynb", "cells": []}
	}`
	out, err := UnmarshalFileReadOutput([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := out.(FileReadNotebook); !ok {
		t.Fatalf("expected FileReadNotebook, got %T", out)
	}
}

func TestUnmarshalFileReadOutput_PDF(t *testing.T) {
	raw := `{
		"type": "pdf",
		"file": {"filePath": "/tmp/doc.pdf", "base64": "JVBER", "originalSize": 2048}
	}`
	out, err := UnmarshalFileReadOutput([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	pdf, ok := out.(FileReadPDF)
	if !ok {
		t.Fatalf("expected FileReadPDF, got %T", out)
	}
	if pdf.File.OriginalSize != 2048 {
		t.Errorf("OriginalSize = %d", pdf.File.OriginalSize)
	}
}

func TestUnmarshalFileReadOutput_Parts(t *testing.T) {
	raw := `{
		"type": "parts",
		"file": {"filePath": "/tmp/doc.pdf", "originalSize": 4096, "count": 5, "outputDir": "/tmp/pages"}
	}`
	out, err := UnmarshalFileReadOutput([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	parts, ok := out.(FileReadParts)
	if !ok {
		t.Fatalf("expected FileReadParts, got %T", out)
	}
	if parts.File.Count != 5 {
		t.Errorf("Count = %d", parts.File.Count)
	}
}

func TestFileEditInput_JSON(t *testing.T) {
	input := FileEditInput{
		FilePath:   "/tmp/file.go",
		OldString:  "old",
		NewString:  "new",
		ReplaceAll: boolPtr(true),
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got FileEditInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.FilePath != "/tmp/file.go" {
		t.Errorf("FilePath = %q", got.FilePath)
	}
	if got.ReplaceAll == nil || !*got.ReplaceAll {
		t.Error("expected ReplaceAll=true")
	}
}

func TestFileEditOutput_JSON(t *testing.T) {
	raw := `{
		"filePath": "/tmp/file.go",
		"oldString": "foo",
		"newString": "bar",
		"originalFile": "package main\nfoo\n",
		"structuredPatch": [{"oldStart": 2, "oldLines": 1, "newStart": 2, "newLines": 1, "lines": ["-foo", "+bar"]}],
		"userModified": false,
		"replaceAll": false
	}`
	var out FileEditOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.StructuredPatch) != 1 {
		t.Errorf("expected 1 patch hunk, got %d", len(out.StructuredPatch))
	}
	if out.StructuredPatch[0].OldStart != 2 {
		t.Errorf("OldStart = %d", out.StructuredPatch[0].OldStart)
	}
}

func TestFileWriteOutput_JSON(t *testing.T) {
	raw := `{
		"type": "create",
		"filePath": "/tmp/new.go",
		"content": "package main\n",
		"structuredPatch": [],
		"originalFile": null
	}`
	var out FileWriteOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.Type != "create" {
		t.Errorf("Type = %q, want %q", out.Type, "create")
	}
	if out.OriginalFile != nil {
		t.Errorf("OriginalFile should be nil for new files")
	}
}
