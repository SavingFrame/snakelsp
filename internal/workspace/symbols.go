package workspace

import (
	"errors"
	"fmt"
	"log/slog"
	"snakelsp/internal/messages"
	"sync"

	"github.com/lithammer/fuzzysearch/fuzzy"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

type Symbol struct {
	Name       string
	Kind       messages.SymbolKind
	Parameters string
	ReturnType string
	FullName   string
	File       *PythonFile
	Range      messages.Range
	NameRange  messages.Range
	Children   []*Symbol
}

var (
	WorkspaceSymbols sync.Map
	flatSymbols      []*Symbol
)

func isChildOf(symbol *Symbol, class *Symbol) bool {
	classStart := class.Range.Start.Line
	classEnd := class.Range.End.Line
	symbolStart := symbol.Range.Start.Line

	return symbolStart >= classStart && symbolStart <= classEnd
}

func (f *PythonFile) ParseSymbols() ([]*Symbol, error) {
	qc := tree_sitter.NewQueryCursor()
	defer qc.Close()
	language := tree_sitter.NewLanguage(tree_sitter_python.Language())
	query, err := tree_sitter.NewQuery(language, getTreeSitterQuery())
	if err != nil {
		slog.Error("Error creating query", "error", err)
		return nil, err
	}
	symbols := processSymbols(f, qc, query)
	// WARNING: I dont think that we need it, cuz we will duplicate symbols with bulkParseSymbol
	// WorkspaceSymbols.Store(f, symbols)
	return symbols, nil
}

func getTreeSitterQuery() string {
	return `
    ;; Capture class definitions with their full body
    (class_definition
        name: (identifier) @class.name
        body: (block) @class.body) ; Capturing the full class body for accurate range

    ;; Capture methods inside a class (ensuring no duplication with functions)
    (class_definition
        body: (block
            (function_definition
                name: (identifier) @method.name
                parameters: (parameters) @method.params
                return_type: (type)? @method.return_type
                body:(_) @method.body))) ; Capture full method range

    ;; Capture standalone module functions (without duplication)
    (module (function_definition
        name: (identifier) @function.name
        parameters: (parameters) @function.params
        return_type: (type)? @function.return_type
        body:(_) @function.body)) ; Capture function body for range
    `
}

func BulkParseSymbols() error {
	slog.Debug("Bulk parse symbols")
	qc := tree_sitter.NewQueryCursor()
	defer qc.Close()
	language := tree_sitter.NewLanguage(tree_sitter_python.Language())
	query, err := tree_sitter.NewQuery(language, getTreeSitterQuery())
	if err != nil {
		slog.Error("Error creating query", "error", err)
		return err
	}
	ProjectFiles.Range(func(key interface{}, value interface{}) bool {
		pythonFile, ok := value.(*PythonFile)
		if !ok {
			slog.Error("Unexpected value type")
			return true
		}
		symbols := processSymbols(pythonFile, qc, query)
		WorkspaceSymbols.Store(pythonFile, symbols)
		for _, symbol := range symbols {
			flatSymbols = append(flatSymbols, symbol)
			flatSymbols = append(flatSymbols, symbol.Children...)
		}
		return true
	})
	slog.Debug("Bulk parse symbols done")
	return nil
}

func processSymbols(pythonFile *PythonFile, qc *tree_sitter.QueryCursor, query *tree_sitter.Query) []*Symbol {
	classSymbols := []*Symbol{}    // Store classes
	moduleSymbols := []*Symbol{}   // Store standalone functions
	childrenSymbols := []*Symbol{} // Store methods
	matches := qc.Matches(query, pythonFile.GetOrCreateAst(), []byte(pythonFile.Text))
	for match := matches.Next(); match != nil; match = matches.Next() {
		var newSymbol *Symbol

		var name, params, returnType string
		var kind messages.SymbolKind
		var startPos, endPos, nameStartPos, nameEndPos messages.Position
		for _, capture := range match.Captures {
			captureName := query.CaptureNames()[capture.Index]
			captureText := capture.Node.Utf8Text([]byte(pythonFile.Text))

			switch captureName {
			case "class.name", "method.name", "function.name":
				name = captureText
				nameStartPos = messages.Position{
					Line:      messages.UInteger(capture.Node.StartPosition().Row),
					Character: messages.UInteger(capture.Node.StartPosition().Column),
				}
				nameEndPos = messages.Position{
					Line:      messages.UInteger(capture.Node.EndPosition().Row),
					Character: messages.UInteger(capture.Node.EndPosition().Column),
				}
				if captureName == "class.name" {
					kind = messages.SymbolKindClass
				} else if captureName == "method.name" {
					kind = messages.SymbolKindMethod
				} else {
					kind = messages.SymbolKindFunction
				}
			case "function.params", "method.params":
				params = captureText
			case "function.return_type", "method.return_type":
				returnType = captureText
			case "function.body", "class.body", "method.body":
				startPos = messages.Position{
					Line:      messages.UInteger(capture.Node.StartPosition().Row),
					Character: messages.UInteger(capture.Node.StartPosition().Column),
				}
				endPos = messages.Position{
					Line:      messages.UInteger(capture.Node.EndPosition().Row),
					Character: messages.UInteger(capture.Node.EndPosition().Column),
				}
			}
		}
		if name == "" {
			return nil
		}
		fullName := fmt.Sprintf("%s%s", name, params)
		if returnType != "" {
			fullName += fmt.Sprintf(" -> %s", returnType)
		}
		newSymbol = &Symbol{
			Name:       name,
			Kind:       kind,
			Parameters: params,
			ReturnType: returnType,
			FullName:   fullName,
			File:       pythonFile,
			Range: messages.Range{
				Start: startPos,
				End:   endPos,
			},
			NameRange: messages.Range{
				Start: nameStartPos,
				End:   nameEndPos,
			},
			Children: []*Symbol{},
		}
		if kind == messages.SymbolKindMethod {
			for _, classSymbol := range classSymbols {
				if isChildOf(newSymbol, classSymbol) {
					classSymbol.Children = append(classSymbol.Children, newSymbol)
					childrenSymbols = append(childrenSymbols, newSymbol)
					break
				}
			}
		} else if kind == messages.SymbolKindClass {
			classSymbols = append(classSymbols, newSymbol)
		} else {
			moduleSymbols = append(moduleSymbols, newSymbol)
		}
		slog.Debug("Symbol parsing", "name", name, "kind", kind, "params", params, "returnType", returnType, "fullName", fullName, "startPos", startPos, "endPos", endPos)
	}
	var symbols []*Symbol
	symbols = append(symbols, classSymbols...)
	symbols = append(symbols, moduleSymbols...)
	return symbols
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
	if query == "" {
		return flatSymbols, nil
	}
	symbols, err := filterSymbols(flatSymbols, query)
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
