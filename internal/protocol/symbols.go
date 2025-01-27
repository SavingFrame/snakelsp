package protocol

import (
	"encoding/json"
	"log/slog"
	"snakelsp/internal/messages"
	"snakelsp/internal/workspace"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

func HandleWorkspaceSymbol(c *Context) (interface{}, error) {
	response := []messages.WorkspaceSymbol{}
	var data messages.WorkspaceSymbolParams
	err := json.Unmarshal(c.Params, &data)
	if err != nil {
		c.Logger.Error("Unmarshalling error: %v", slog.Any("error", err))
		return nil, err
	}
	qc := tree_sitter.NewQueryCursor()
	defer qc.Close()
	qc.SetMatchLimit(100)
	language := tree_sitter.NewLanguage(tree_sitter_python.Language())
	query, err := tree_sitter.NewQuery(language, `
    (function_definition name: (identifier) @function)
    (class_definition name: (identifier) @class)
    `)
	slog.Debug("Im here")
	workspace.ProjectFiles.Range(func(key interface{}, value interface{}) bool {
		if pythonFile, ok := value.(*workspace.PythonFile); ok {
			matches := qc.Matches(query, pythonFile.GetOrCreateAst(), []byte(pythonFile.Text))
			for match := matches.Next(); match != nil; match = matches.Next() {
				for _, capture := range match.Captures {
					response = append(response, messages.WorkspaceSymbol{
						Name: capture.Node.Utf8Text([]byte(pythonFile.Text)),
						// FIXME: Use data from the response
						Kind: messages.SymbolKindClass,
						Tags: messages.SymbolTag(1),
						Location: messages.Location{
							URI: pythonFile.Url,
							Range: messages.Range{
								Start: messages.Position{
									Line:      messages.UInteger(capture.Node.StartPosition().Row),
									Character: messages.UInteger(capture.Node.StartPosition().Column),
								},
								End: messages.Position{
									Line:      messages.UInteger(capture.Node.StartPosition().Row),
									Character: messages.UInteger(capture.Node.EndPosition().Column),
								},
							},
						},
					})
				}
			}
		} else {
			slog.Error("Unexpected value type")
		}
		return len(response) < 100
	})
	return response, nil
}
