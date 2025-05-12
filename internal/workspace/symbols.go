package workspace

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"snakelsp/internal/messages"
	"snakelsp/internal/progress"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/google/uuid"
	"github.com/lithammer/fuzzysearch/fuzzy"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

// TODO: Now we have duplications of symbols in the flatSymbols, because BulkParseSymbols conflicts with ParseSymbols, and if you use both, you will have duplicates.
type Symbol struct {
	UUID       uuid.UUID
	Name       string
	Kind       messages.SymbolKind
	Parameters string
	ReturnType string
	FullName   string
	File       *PythonFile
	Range      messages.Range
	NameRange  messages.Range
	Children   []*Symbol
	Parent     *Symbol

	// Base classes for class
	// TODO: 1. Parse superclasses with attributes
	// 2. Parse imports and superclasses from related files
	SuperClasses      []*Symbol
	superClassesNames []string
}

var (
	WorkspaceSymbols sync.Map
	flatSymbols      *orderedmap.OrderedMap[uuid.UUID, *Symbol] = orderedmap.NewOrderedMap[uuid.UUID, *Symbol]()
)

func SearchSymbolByUUID(uuid uuid.UUID) (*Symbol, error) {
	symbol, exists := flatSymbols.Get(uuid)
	if !exists {
		return nil, fmt.Errorf("symbol not found")
	}
	return symbol, nil
}

func (f *PythonFile) parseSymbols() ([]*Symbol, error) {
	qc := tree_sitter.NewQueryCursor()
	defer qc.Close()
	language := tree_sitter.NewLanguage(tree_sitter_python.Language())
	query, err := tree_sitter.NewQuery(language, getTreeSitterQuery())
	if err != nil {
		slog.Error("Error creating query", "error", err)
		return nil, err
	}
	symbols := processSymbols(f, qc, query)
	return symbols, nil
}

func BulkParseSymbols(pr *progress.WorkDone) error {
	slog.Debug("Bulk parse symbols")
	pr.Start("Parsing symbols")
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
		if pythonFile.External {
			return true
		}
		symbols := processSymbols(pythonFile, qc, query)
		for _, symbol := range symbols {
			resolveExternalSuperclassSymbol(pythonFile, symbol)
		}
		WorkspaceSymbols.Store(pythonFile, symbols)
		for _, symbol := range symbols {
			flatSymbols.Set(symbol.UUID, symbol)
			for _, children := range symbol.Children {
				flatSymbols.Set(children.UUID, children)
				resolveExternalSuperclassSymbol(pythonFile, children)
			}
		}
		return true
	})
	slog.Debug("Bulk parse symbols done")
	pr.End("Symbols parsed")
	return nil
}

func (f *PythonFile) FileSymbols(query string) ([]*Symbol, error) {
	var symbols []*Symbol

	value, exists := WorkspaceSymbols.Load(f)
	if !exists {
		var err error
		symbols, err = f.parseSymbols()
		WorkspaceSymbols.Store(f, symbols)
		if !f.External {
			for _, symbol := range symbols {
				flatSymbols.Set(symbol.UUID, symbol)
				for _, children := range symbol.Children {
					flatSymbols.Set(children.UUID, children)
				}
			}
		}
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
	symbols := slices.Collect(flatSymbols.Values())
	if query == "" {
		return symbols, nil
	}
	symbols, err := filterSymbols(symbols, query)
	if err != nil {
		return nil, err
	}
	return symbols, nil
}

func FindSymbolByPosition(file *PythonFile, line, character uint32) (*Symbol, error) {
	for symbol := range flatSymbols.Values() {
		if symbol.File.Url == file.Url &&
			(symbol.NameRange.Start.Line <= line && symbol.NameRange.End.Line >= line) &&
			(symbol.NameRange.Start.Character <= character && symbol.NameRange.End.Character >= character) {
			return symbol, nil
		}
	}
	return nil, fmt.Errorf("symbol not found")
}

func (s *Symbol) SymbolNameWithParent() string {
	if s.Parent == nil {
		return s.Name
	}
	return fmt.Sprintf("%s.%s", s.Parent.Name, s.Name)
}

func filterSymbols(symbols []*Symbol, query string) ([]*Symbol, error) {
	var filteredSymbols []*Symbol

	// Collect all symbol names into a slice
	var names []string
	for _, symbol := range symbols {
		names = append(names, symbol.SymbolNameWithParent())
	}
	slog.Debug("Names", "names", names)
	// Perform fuzzy matching on names
	matchedNames := fuzzy.FindFold(query, names)

	// Collect the symbols that match the names
	for _, matchedName := range matchedNames {
		for _, symbol := range symbols {
			if symbol.SymbolNameWithParent() == matchedName {
				filteredSymbols = append(filteredSymbols, symbol)
				break
			}
		}
	}

	return filteredSymbols, nil
}

func getTreeSitterQuery() string {
	return `
    ;; Capture class definitions with their full body
	(class_definition
		name: (identifier) @class.name
		(argument_list
		(identifier) @class.superclass)? @class.superclasses
		body: (block) @class.body)

    ;; Capture methods definitions inside a class (ensuring no duplication with functions) without decorators
    (class_definition
        body: (block
            (function_definition
                name: (identifier) @method.name
                parameters: (parameters) @method.params
                return_type: (type)? @method.return_type
                body:(_) @method.body))) ; Capture full method range


	;; Method definitions (decorated)
	(class_definition
	  body: (block
		(decorated_definition
		  (decorator)* ; optionally capture individual decorators if needed
		  definition: (function_definition
			name: (identifier) @method.name
			parameters: (parameters) @method.params
			return_type: (type)? @method.return_type
			body: (_) @method.body))))

    ;; Capture standalone module functions (without duplication)

	;; Without decorators
    (module (function_definition
        name: (identifier) @function.name
        parameters: (parameters) @function.params
        return_type: (type)? @function.return_type
        body:(_) @function.body)) ; Capture function body for range
	;; With decorators
	(module
	  (decorated_definition
		(decorator)* ; optional, you can also write [(decorator)] as a short form
		definition: (function_definition
		  name: (identifier) @function.name
		  parameters: (parameters) @function.params
		  return_type: (type)? @function.return_type
		  body: (_) @function.body)))
    `
}

func createSymbol(
	name string,
	kind messages.SymbolKind,
	params string,
	returnType string,
	fullName string,
	file *PythonFile,
	startPos messages.Position,
	endPos messages.Position,
	nameStartPos messages.Position,
	nameEndPos messages.Position,
	superClassName string,
) *Symbol {
	symbol := &Symbol{
		UUID:       uuid.New(),
		Name:       name,
		Kind:       kind,
		Parameters: params,
		ReturnType: returnType,
		FullName:   fullName,
		File:       file,
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
	if superClassName != "" {
		symbol.superClassesNames = append(symbol.superClassesNames, superClassName)
	}
	return symbol
}

func resolveExternalSuperclassSymbol(f *PythonFile, symbol *Symbol) *Symbol {
	// Trey to find symbol inside the same file
	fileSymbols, err := f.FileSymbols("")
	if err != nil {
		slog.Warn("Unable to get symbols for resolving superclass", slog.String("file", f.Url), "error", err)
		return symbol
	}
	for _, fsymb := range fileSymbols {
		for _, superClassName := range symbol.superClassesNames {
			if fsymb.Name == superClassName {
				symbol.SuperClasses = append(symbol.SuperClasses, fsymb)
				return symbol
			}
		}
	}
	// Try to find symbol in the import
	imports, err := f.GetImports()
	if err != nil {
		slog.Warn("Unable to get import for resolving superclass", slog.String("file", f.Url), "error", err)
		return symbol
	}
	for _, imp := range imports {
		for _, superClassName := range symbol.superClassesNames {
			if imp.ImportedName == superClassName {
				symbol.SuperClasses = append(symbol.SuperClasses, imp.Symbol)
				return symbol
			}
		}
	}
	return symbol
}

func processSymbols(pythonFile *PythonFile, qc *tree_sitter.QueryCursor, query *tree_sitter.Query) []*Symbol {
	classSymbols := map[string]*Symbol{} // Store classes by name and name range
	moduleSymbols := []*Symbol{}         // Store standalone functions
	matches := qc.Matches(query, pythonFile.GetOrCreateAst(), []byte(pythonFile.Text))
	for match := matches.Next(); match != nil; match = matches.Next() {

		var name, params, returnType string
		var superClass string
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
			case "class.superclass":
				superClass = captureText
				// for _, classSymbol := range classSymbols {
				// 	if classSymbol.Name == captureText {
				// 		superClass = classSymbol
				// 	}
				// }
				// if superClass == nil {
				//
				// 	slog.Debug("Resolving external superclass", slog.String("name", captureText))
				// 	superClass, _ = resolveExternalSuperclassSymbol(pythonFile, captureText)
				// }
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
		if kind == messages.SymbolKindMethod {
			newSymbol := createSymbol(name, kind, params, returnType, fullName, pythonFile, startPos, endPos, nameStartPos, nameEndPos, "")
			for _, classSymbol := range classSymbols {
				if isChildOf(newSymbol, classSymbol) {
					newSymbol.Parent = classSymbol
					classSymbol.Children = append(classSymbol.Children, newSymbol)
					break
				}
			}
		} else if kind == messages.SymbolKindClass {
			key := fmt.Sprintf("%s:%d:%d", name, nameStartPos.Line, nameStartPos.Character)
			symbol, exists := classSymbols[key]
			// Create new symbol if it doesn't exist
			if !exists {
				newSymbol := createSymbol(name, kind, params, returnType, fullName, pythonFile, startPos, endPos, nameStartPos, nameEndPos, superClass)
				classSymbols[key] = newSymbol
			} else if superClass != "" {
				// Update existing symbol with superclass
				symbol.superClassesNames = append(symbol.superClassesNames, superClass)
			}
		} else {
			newSymbol := createSymbol(name, kind, params, returnType, fullName, pythonFile, startPos, endPos, nameStartPos, nameEndPos, "")
			moduleSymbols = append(moduleSymbols, newSymbol)
		}
	}
	var symbols []*Symbol
	// symbols = append(symbols, classSymbols...)
	// appends class symbols to the symbols slice
	for _, symbol := range classSymbols {
		symbols = append(symbols, symbol)
	}
	symbols = append(symbols, moduleSymbols...)
	return symbols
}

func isChildOf(symbol *Symbol, class *Symbol) bool {
	classStart := class.Range.Start.Line
	classEnd := class.Range.End.Line
	symbolStart := symbol.Range.Start.Line

	return symbolStart >= classStart && symbolStart <= classEnd
}
