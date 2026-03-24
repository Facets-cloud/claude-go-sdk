package tools

// NotebookEditInput is the input for the NotebookEdit tool.
type NotebookEditInput struct {
	// NotebookPath is the absolute path to the notebook file.
	NotebookPath string `json:"notebook_path"`
	// CellID is the ID of the cell to edit.
	CellID *string `json:"cell_id,omitempty"`
	// NewSource is the new source for the cell.
	NewSource string `json:"new_source"`
	// CellType is "code" or "markdown".
	CellType *string `json:"cell_type,omitempty"`
	// EditMode is "replace", "insert", or "delete".
	EditMode *string `json:"edit_mode,omitempty"`
}

// NotebookEditOutput is the output from the NotebookEdit tool.
type NotebookEditOutput struct {
	// NewSource is the new source code written to the cell.
	NewSource string `json:"new_source"`
	// CellID is the ID of the cell that was edited.
	CellID *string `json:"cell_id,omitempty"`
	// CellType is "code" or "markdown".
	CellType string `json:"cell_type"`
	// Language is the programming language of the notebook.
	Language string `json:"language"`
	// EditMode is the edit mode that was used.
	EditMode string `json:"edit_mode"`
	// Error is the error message if the operation failed.
	Error *string `json:"error,omitempty"`
	// NotebookPath is the path to the notebook file.
	NotebookPath string `json:"notebook_path"`
	// OriginalFile is the notebook content before modification.
	OriginalFile string `json:"original_file"`
	// UpdatedFile is the notebook content after modification.
	UpdatedFile string `json:"updated_file"`
}
