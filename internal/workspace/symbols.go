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
	SuperObjects      []*Symbol
	superObjectsNames []string
}

var (
	WorkspaceSymbols sync.Map
	FlatSymbols      *orderedmap.OrderedMap[uuid.UUID, *Symbol] = orderedmap.NewOrderedMap[uuid.UUID, *Symbol]()
)

func SearchSymbolByUUID(uuid uuid.UUID) (*Symbol, error) {
	symbol, exists := FlatSymbols.Get(uuid)
	if !exists {
		return nil, fmt.Errorf("symbol not found")
	}
	return symbol, nil
}

// It just parse symbols from the file and doesn't store them in the WorkspaceSymbols
func (f *PythonFile) parseFileSymbols() ([]*Symbol, error) {
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

		// Get existing symbols if they exist
		existingSymbols, hasExisting := WorkspaceSymbols.Load(pythonFile)
		newSymbols := processSymbols(pythonFile, qc, query)

		if hasExisting {
			// Update existing symbols in place to preserve references
			updateSymbolsInPlace(existingSymbols.([]*Symbol), newSymbols)
		} else {
			// Store new symbols
			WorkspaceSymbols.Store(pythonFile, newSymbols)
			for _, symbol := range newSymbols {
				FlatSymbols.Set(symbol.UUID, symbol)
				for _, children := range symbol.Children {
					FlatSymbols.Set(children.UUID, children)
				}
			}
		}
		return true
	})
	for symbol := range FlatSymbols.Values() {
		resolveExternalSuperclassSymbol(symbol.File, symbol)
		if symbol.Kind == messages.SymbolKindMethod {
			resolveExternalSuperMethodSymbol(symbol.File, symbol)
		}
	}

	slog.Debug("Bulk parse symbols done")
	pr.End("Symbols parsed")
	return nil
}

func (f *PythonFile) parseSymbols() ([]*Symbol, error) {
	if f.External {
		return nil, errors.New("cannot parse symbols for external files")
	}
	qc := tree_sitter.NewQueryCursor()
	defer qc.Close()
	language := tree_sitter.NewLanguage(tree_sitter_python.Language())
	query, err := tree_sitter.NewQuery(language, getTreeSitterQuery())
	if err != nil {
		slog.Error("Error creating query", "error", err)
		return nil, err
	}
	symbols := processSymbols(f, qc, query)
	WorkspaceSymbols.Store(f, symbols)
	slog.Debug("Symbols for file parsed from the parseSymbols func", slog.String("file", f.Url), slog.Int("symbols", len(symbols)))
	for _, symbol := range symbols {
		resolveExternalSuperclassSymbol(f, symbol)
	}
	for _, symbol := range symbols {
		FlatSymbols.Set(symbol.UUID, symbol)
		for _, children := range symbol.Children {
			FlatSymbols.Set(children.UUID, children)
			resolveExternalSuperclassSymbol(f, children)
			if children.Kind == messages.SymbolKindMethod {
				resolveExternalSuperMethodSymbol(f, children)
			}
		}
	}
	return symbols, nil
}

func (f *PythonFile) FileSymbols(query string) ([]*Symbol, error) {
	var symbols []*Symbol

	value, exists := WorkspaceSymbols.Load(f)
	if !exists {
		var err error
		symbols, err = f.parseFileSymbols()
		WorkspaceSymbols.Store(f, symbols)
		if !f.External {
			for _, symbol := range symbols {
				FlatSymbols.Set(symbol.UUID, symbol)
				for _, children := range symbol.Children {
					FlatSymbols.Set(children.UUID, children)
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
	symbols, err := FilterSymbols(symbols, query)
	if err != nil {
		return nil, err
	}
	return symbols, nil
}

func GetWorkspaceSymbols(query string) ([]*Symbol, error) {
	symbols := slices.Collect(FlatSymbols.Values())
	if query == "" {
		return symbols, nil
	}
	symbols, err := FilterSymbols(symbols, query)
	if err != nil {
		return nil, err
	}
	return symbols, nil
}

func FindSymbolByPosition(file *PythonFile, line, character uint32) (*Symbol, error) {
	for symbol := range FlatSymbols.Values() {
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
	return fmt.Sprintf("%s.%s", s.Parent.Name, s.FullName)
}

func FilterSymbols(symbols []*Symbol, query string) ([]*Symbol, error) {
	var filteredSymbols []*Symbol

	// Collect all symbol names into a slice
	var names []string
	for _, symbol := range symbols {
		names = append(names, symbol.SymbolNameWithParent())
	}
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
		symbol.superObjectsNames = append(symbol.superObjectsNames, superClassName)
	}
	return symbol
}

func resolveExternalSuperMethodSymbol(f *PythonFile, symbol *Symbol) *Symbol {
	classSymbol := symbol.Parent
	if classSymbol == nil {
		return nil
	}
	var superObject *Symbol
	if len(classSymbol.SuperObjects) > 0 {
		for _, superClassMethod := range classSymbol.SuperObjects[0].Children {
			if superClassMethod.Name == symbol.Name {
				symbol.SuperObjects = append(symbol.SuperObjects, superClassMethod)
				symbol.superObjectsNames = append(symbol.superObjectsNames, superClassMethod.Name)
				superObject = superClassMethod

			}
		}
	}
	return superObject
}

func resolveExternalSuperclassSymbol(f *PythonFile, symbol *Symbol) *Symbol {
	// Trey to find symbol inside the same file
	fileSymbols, err := f.FileSymbols("")
	if err != nil {
		slog.Warn("Unable to get symbols for resolving superclass", slog.String("file", f.Url), "error", err)
		return symbol
	}
	for _, fsymb := range fileSymbols {
		for _, superClassName := range symbol.superObjectsNames {
			if fsymb.Name == superClassName {
				symbol.SuperObjects = append(symbol.SuperObjects, fsymb)
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
		for _, superClassName := range symbol.superObjectsNames {
			if imp.ImportedName == superClassName && imp.Symbol != nil {
				symbol.SuperObjects = append(symbol.SuperObjects, imp.Symbol)
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
				symbol.superObjectsNames = append(symbol.superObjectsNames, superClass)
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

func updateSymbolsInPlace(existingSymbols []*Symbol, newSymbols []*Symbol) {
	// Create a map of new symbols by name and position for quick lookup
	newSymbolMap := make(map[string]*Symbol)
	for _, newSymbol := range newSymbols {
		key := fmt.Sprintf("%s:%d:%d", newSymbol.Name, newSymbol.NameRange.Start.Line, newSymbol.NameRange.Start.Character)
		newSymbolMap[key] = newSymbol
	}

	// Update existing symbols with new data
	for _, existingSymbol := range existingSymbols {
		key := fmt.Sprintf("%s:%d:%d", existingSymbol.Name, existingSymbol.NameRange.Start.Line, existingSymbol.NameRange.Start.Character)
		if newSymbol, found := newSymbolMap[key]; found {
			// Update all fields except UUID to preserve references
			existingSymbol.Name = newSymbol.Name
			existingSymbol.Kind = newSymbol.Kind
			existingSymbol.Parameters = newSymbol.Parameters
			existingSymbol.ReturnType = newSymbol.ReturnType
			existingSymbol.FullName = newSymbol.FullName
			existingSymbol.Range = newSymbol.Range
			existingSymbol.NameRange = newSymbol.NameRange
			existingSymbol.superObjectsNames = newSymbol.superObjectsNames

			// Update children recursively
			updateSymbolsInPlace(existingSymbol.Children, newSymbol.Children)

			// Update parent relationships for children
			for _, child := range existingSymbol.Children {
				child.Parent = existingSymbol
			}
		}
	}
}

func isChildOf(symbol *Symbol, class *Symbol) bool {
	classStart := class.Range.Start.Line
	classEnd := class.Range.End.Line
	symbolStart := symbol.Range.Start.Line

	return symbolStart >= classStart && symbolStart <= classEnd
}
