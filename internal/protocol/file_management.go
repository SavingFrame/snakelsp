package protocol

import (
	"encoding/json"
	"log/slog"
	"snakelsp/internal/messages"
	"snakelsp/internal/request"
	"snakelsp/internal/workspace"
	"strings"
)

func HandleDidOpen(r *request.Request) (interface{}, error) {
	var data messages.DidOpenTextDocumentParams
	var external bool
	err := json.Unmarshal(r.Params, &data)
	if err != nil {
		r.Logger.Error("Unmarshalling error: %v", slog.Any("error", err))
		return nil, err
	}
	if data.TextDocument.LanguageID != "python" {
		return nil, nil
	}
	if strings.Contains(data.TextDocument.URI, workspace.ClientSettings.WorkspaceRoot) {
		external = false
	} else {
		external = true
	}
	workspace.NewPythonFile(data.TextDocument.URI, data.TextDocument.Text, external, true)

	return interface{}(nil), nil
}

func applyChange(content string, startLine, startCharacter, endLine, endCharacter uint32, text string, logger *slog.Logger) string {
	lines := strings.Split(content, "\n")

	// Ensure the start and end line indices are within bounds
	if int(startLine) >= len(lines) || int(endLine) >= len(lines) {
		logger.Warn("Invalid line numbers")
		return content
	}

	// Extract the lines where the range starts and ends
	startTargetLine := lines[startLine]
	endTargetLine := lines[endLine]

	// Ensure that character positions are valid within the respective lines
	if int(startCharacter) > len(startTargetLine) || int(endCharacter) > len(endTargetLine) {
		logger.Warn("Invalid character positions")
		return content
	}

	// Handle different cases based on start and end indices
	if startLine == endLine {
		// Case 1: Change occurs within a single line
		// Replace the range directly in the same line
		updatedLine := startTargetLine[:startCharacter] + text + startTargetLine[endCharacter:]
		lines[startLine] = updatedLine
	} else {
		// Case 2: Change spans multiple lines
		// Compose new content from fragments:
		// - Start of the first line up to `startCharacter`
		startFragment := startTargetLine[:startCharacter]

		// - End of the last line from `endCharacter`
		endFragment := endTargetLine[endCharacter:]

		// Replace the lines in between with the new text
		updatedLine := startFragment + text + endFragment
		lines = append(lines[:startLine], append([]string{updatedLine}, lines[endLine+1:]...)...)
	}

	// Reassemble the lines back into the full content
	return strings.Join(lines, "\n")
}

func HandleDidChange(r *request.Request) (interface{}, error) {
	var data messages.DidChangeTextDocumentParams
	err := json.Unmarshal(r.Params, &data)
	if err != nil {
		r.Logger.Error("Unmarshalling error: %v", slog.Any("error", err))
		return nil, err
	}
	pythonFile, err := workspace.GetPythonFile(data.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	content := pythonFile.Text
	for _, change := range data.ContentChanges {
		content = applyChange(
			content,
			change.Range.Start.Line,
			change.Range.Start.Character,
			change.Range.End.Line,
			change.Range.End.Character,
			change.Text,
			r.Logger,
		)
	}
	pythonFile.Text = content
	workspace.OpenFiles.Store(data.TextDocument.URI, pythonFile)
	r.Logger.Debug("Content after change")
	r.Logger.Debug(pythonFile.Text)

	return nil, nil
}

func HandleDidClose(r *request.Request) (interface{}, error) {
	var data messages.DidCloseTextDocumentParams
	err := json.Unmarshal(r.Params, &data)
	if err != nil {
		r.Logger.Error("Unmarshalling error: %v", slog.Any("error", err))
		return nil, err
	}
	file, err := workspace.GetPythonFile(data.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	file.CloseFile()
	return nil, nil
}
