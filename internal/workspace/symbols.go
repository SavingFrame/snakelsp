package workspace

import (
	"errors"
	"log/slog"
	"snakelsp/internal/messages"
	"sync"

	"github.com/lithammer/fuzzysearch/fuzzy"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

type Symbol struct {
	Name  string
	Kind  messages.SymbolKind
	File  *PythonFile
	Range messages.Range
}

var WorkspaceSymbols sync.Map

func (f *PythonFile) ParseSymbols() ([]*Symbol, error) {
	qc := tree_sitter.NewQueryCursor()
	defer qc.Close()
	language := tree_sitter.NewLanguage(tree_sitter_python.Language())
	query, err := tree_sitter.NewQuery(language, `
        (class_definition
            body: (block (function_definition name: (identifier) @method)))
        (module
            (function_definition name: (identifier) @function))
        (class_definition name: (identifier) @class)
    `)
	if err != nil {
		return nil, err
	}
	symbols := findSymbols(f, qc, query)
	WorkspaceSymbols.Store(f, symbols)
	return symbols, nil
}

func BulkParseSymbols() error {
	slog.Debug("Bulk parse symbols")
	qc := tree_sitter.NewQueryCursor()
	defer qc.Close()
	language := tree_sitter.NewLanguage(tree_sitter_python.Language())
	query, err := tree_sitter.NewQuery(language, `
        (class_definition
            body: (block (function_definition name: (identifier) @method)))
        (module
            (function_definition name: (identifier) @function))
        (class_definition name: (identifier) @class)
    `)
	if err != nil {
		return err
	}
	ProjectFiles.Range(func(key interface{}, value interface{}) bool {
		pythonFile, ok := value.(*PythonFile)
		if !ok {
			slog.Error("Unexpected value type")
			return true
		}
		symbols := findSymbols(pythonFile, qc, query)
		WorkspaceSymbols.Store(pythonFile, symbols)
		return true
	})
	return nil
}

func newSymbol(name string, kind messages.SymbolKind, file *PythonFile, symbolRange messages.Range) *Symbol {
	return &Symbol{
		Name:  name,
		Kind:  kind,
		File:  file,
		Range: symbolRange,
	}
}

func findSymbols(pythonFile *PythonFile, qc *tree_sitter.QueryCursor, query *tree_sitter.Query) []*Symbol {
	captures := []*Symbol{}
	matches := qc.Matches(query, pythonFile.GetOrCreateAst(), []byte(pythonFile.Text))
	for match := matches.Next(); match != nil; match = matches.Next() {
		for _, capture := range match.Captures {
			captureName := query.CaptureNames()[capture.Index]
			kind, exists := messages.GetSymbolKind(captureName)
			if !exists {
				kind = messages.SymbolKindObject
			}
			symbolName := capture.Node.Utf8Text([]byte(pythonFile.Text))
			captures = append(captures, newSymbol(symbolName, kind, pythonFile, messages.Range{
				Start: messages.Position{
					Line:      messages.UInteger(capture.Node.StartPosition().Row),
					Character: messages.UInteger(capture.Node.StartPosition().Column),
				},
				End: messages.Position{
					Line:      messages.UInteger(capture.Node.EndPosition().Row),
					Character: messages.UInteger(capture.Node.EndPosition().Column),
				},
			}))
		}
	}
	return captures
}

func (f *PythonFile) FileSymbols(query string) ([]*Symbol, error) {
	var symbols []*Symbol

	value, exists := WorkspaceSymbols.Load(f)
	if !exists {
		var err error
		symbols, err = f.ParseSymbols()
		if err != nil {
			return nil, err
		}
	} else {
		var ok bool
		symbols, ok = value.([]*Symbol)
		if !ok {
			return nil, errors.New("unexpected value type")
		}
	}
	if query == "" {
		return symbols, nil
	}
	symbols, err := filterSymbols(symbols, query)
	if err != nil {
		return nil, err
	}
	return symbols, nil
}

func GetWorkspaceSymbols(query string) ([]*Symbol, error) {
	var symbols []*Symbol

	ProjectFiles.Range(func(key interface{}, value interface{}) bool {
		pythonFile, ok := value.(*PythonFile)
		if !ok {
			slog.Error("Unexpected value type")
			return true
		}
		value, exists := WorkspaceSymbols.Load(pythonFile)
		if !exists {
			return true
		}
		fileSymbols, exists := value.([]*Symbol)
		if !exists {
			// TODO: Generate file symbols for file?
			return true
		}
		symbols = append(symbols, fileSymbols...)
		return true
	})
	if query == "" {
		return symbols, nil
	}
	symbols, err := filterSymbols(symbols, query)
	if err != nil {
		return nil, err
	}
	return symbols, nil
}

func filterSymbols(symbols []*Symbol, query string) ([]*Symbol, error) {
	var filteredSymbols []*Symbol

	// Collect all symbol names into a slice
	var names []string
	for _, symbol := range symbols {
		names = append(names, symbol.Name)
	}

	// Perform fuzzy matching on names
	matchedNames := fuzzy.FindFold(query, names)

	// Collect the symbols that match the names
	for _, matchedName := range matchedNames {
		for _, symbol := range symbols {
			if symbol.Name == matchedName {
				filteredSymbols = append(filteredSymbols, symbol)
				break
			}
		}
	}

	return filteredSymbols, nil
}
