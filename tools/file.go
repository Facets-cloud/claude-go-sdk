package tools

import "encoding/json"

// FileReadInput is the input for the file Read tool.
type FileReadInput struct {
	// FilePath is the absolute path to the file to read.
	FilePath string `json:"file_path"`
	// Offset is the line number to start reading from.
	Offset *int `json:"offset,omitempty"`
	// Limit is the number of lines to read.
	Limit *int `json:"limit,omitempty"`
	// Pages is the page range for PDF files (e.g., "1-5").
	Pages *string `json:"pages,omitempty"`
}

// FileReadOutput is a union type for file read results.
// Variants: FileReadText, FileReadImage, FileReadNotebook, FileReadPDF, FileReadParts.
type FileReadOutput interface {
	fileReadOutput()
}

// FileReadText is returned when reading a text file.
type FileReadText struct {
	Type string       `json:"type"` // "text"
	File FileTextData `json:"file"`
}

func (FileReadText) fileReadOutput() {}

// FileTextData holds text file content.
type FileTextData struct {
	FilePath   string `json:"filePath"`
	Content    string `json:"content"`
	NumLines   int    `json:"numLines"`
	StartLine  int    `json:"startLine"`
	TotalLines int    `json:"totalLines"`
}

// FileReadImage is returned when reading an image file.
type FileReadImage struct {
	Type string        `json:"type"` // "image"
	File FileImageData `json:"file"`
}

func (FileReadImage) fileReadOutput() {}

// FileImageData holds image file content.
type FileImageData struct {
	Base64       string           `json:"base64"`
	Type         string           `json:"type"` // MIME type
	OriginalSize int              `json:"originalSize"`
	Dimensions   *ImageDimensions `json:"dimensions,omitempty"`
}

// ImageDimensions holds image dimension information.
type ImageDimensions struct {
	OriginalWidth  *int `json:"originalWidth,omitempty"`
	OriginalHeight *int `json:"originalHeight,omitempty"`
	DisplayWidth   *int `json:"displayWidth,omitempty"`
	DisplayHeight  *int `json:"displayHeight,omitempty"`
}

// FileReadNotebook is returned when reading a Jupyter notebook.
type FileReadNotebook struct {
	Type string           `json:"type"` // "notebook"
	File FileNotebookData `json:"file"`
}

func (FileReadNotebook) fileReadOutput() {}

// FileNotebookData holds notebook file content.
type FileNotebookData struct {
	FilePath string        `json:"filePath"`
	Cells    []interface{} `json:"cells"`
}

// FileReadPDF is returned when reading a PDF as base64.
type FileReadPDF struct {
	Type string      `json:"type"` // "pdf"
	File FilePDFData `json:"file"`
}

func (FileReadPDF) fileReadOutput() {}

// FilePDFData holds PDF file content.
type FilePDFData struct {
	FilePath     string `json:"filePath"`
	Base64       string `json:"base64"`
	OriginalSize int    `json:"originalSize"`
}

// FileReadParts is returned when reading a PDF as extracted page images.
type FileReadParts struct {
	Type string        `json:"type"` // "parts"
	File FilePartsData `json:"file"`
}

func (FileReadParts) fileReadOutput() {}

// FilePartsData holds extracted PDF page data.
type FilePartsData struct {
	FilePath     string `json:"filePath"`
	OriginalSize int    `json:"originalSize"`
	Count        int    `json:"count"`
	OutputDir    string `json:"outputDir"`
}

// UnmarshalFileReadOutput unmarshals JSON into the correct FileReadOutput variant.
func UnmarshalFileReadOutput(data []byte) (FileReadOutput, error) {
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return nil, err
	}
	switch probe.Type {
	case "image":
		var out FileReadImage
		if err := json.Unmarshal(data, &out); err != nil {
			return nil, err
		}
		return out, nil
	case "notebook":
		var out FileReadNotebook
		if err := json.Unmarshal(data, &out); err != nil {
			return nil, err
		}
		return out, nil
	case "pdf":
		var out FileReadPDF
		if err := json.Unmarshal(data, &out); err != nil {
			return nil, err
		}
		return out, nil
	case "parts":
		var out FileReadParts
		if err := json.Unmarshal(data, &out); err != nil {
			return nil, err
		}
		return out, nil
	default: // "text"
		var out FileReadText
		if err := json.Unmarshal(data, &out); err != nil {
			return nil, err
		}
		return out, nil
	}
}

// FileEditInput is the input for the file Edit tool.
type FileEditInput struct {
	// FilePath is the absolute path to the file to modify.
	FilePath string `json:"file_path"`
	// OldString is the text to replace.
	OldString string `json:"old_string"`
	// NewString is the replacement text.
	NewString string `json:"new_string"`
	// ReplaceAll replaces all occurrences when true.
	ReplaceAll *bool `json:"replace_all,omitempty"`
}

// StructuredPatchHunk represents a hunk in a structured diff patch.
type StructuredPatchHunk struct {
	OldStart int      `json:"oldStart"`
	OldLines int      `json:"oldLines"`
	NewStart int      `json:"newStart"`
	NewLines int      `json:"newLines"`
	Lines    []string `json:"lines"`
}

// GitDiff contains git diff information for a file change.
type GitDiff struct {
	Filename   string  `json:"filename"`
	Status     string  `json:"status"` // "modified" or "added"
	Additions  int     `json:"additions"`
	Deletions  int     `json:"deletions"`
	Changes    int     `json:"changes"`
	Patch      string  `json:"patch"`
	Repository *string `json:"repository,omitempty"`
}

// FileEditOutput is the output from the file Edit tool.
type FileEditOutput struct {
	// FilePath is the file that was edited.
	FilePath string `json:"filePath"`
	// OldString is the original string that was replaced.
	OldString string `json:"oldString"`
	// NewString is the replacement string.
	NewString string `json:"newString"`
	// OriginalFile is the file contents before editing.
	OriginalFile string `json:"originalFile"`
	// StructuredPatch shows the changes as diff hunks.
	StructuredPatch []StructuredPatchHunk `json:"structuredPatch"`
	// UserModified indicates whether the user modified the proposed changes.
	UserModified bool `json:"userModified"`
	// ReplaceAll indicates whether all occurrences were replaced.
	ReplaceAll bool `json:"replaceAll"`
	// GitDiff contains git diff information.
	GitDiff *GitDiff `json:"gitDiff,omitempty"`
}

// FileWriteInput is the input for the file Write tool.
type FileWriteInput struct {
	// FilePath is the absolute path to the file to write.
	FilePath string `json:"file_path"`
	// Content is the content to write.
	Content string `json:"content"`
}

// FileWriteOutput is the output from the file Write tool.
type FileWriteOutput struct {
	// Type is "create" for new files or "update" for existing files.
	Type string `json:"type"`
	// FilePath is the path to the file that was written.
	FilePath string `json:"filePath"`
	// Content is the content that was written.
	Content string `json:"content"`
	// StructuredPatch shows the changes as diff hunks.
	StructuredPatch []StructuredPatchHunk `json:"structuredPatch"`
	// OriginalFile is the file content before the write (nil for new files).
	OriginalFile *string `json:"originalFile"`
	// GitDiff contains git diff information.
	GitDiff *GitDiff `json:"gitDiff,omitempty"`
}
