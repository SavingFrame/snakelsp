package workspace

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"snakelsp/internal/progress"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

type Import struct {
	Alias        string // Name in the file (e.g., "foo" or "alias")
	SourceModule string // Full import path (like "foo" or "foo.bar")
	ImportedName string // Specific name if "from foo import Bar" (it's "Bar"), else empty
	PythonFile   *PythonFile
	Symbol       *Symbol
}

func (f *PythonFile) ParseImports(cs *ClientSettingsType) ([]Import, error) {
	qc := tree_sitter.NewQueryCursor()
	defer qc.Close()
	language := tree_sitter.NewLanguage(tree_sitter_python.Language())
	query, err := tree_sitter.NewQuery(language, getTreeSitterImportQuery())
	if err != nil {
		return nil, err
	}
	imports := processImports(f, qc, query, cs)
	return imports, nil
}

func BulkParseImports(pr *progress.WorkDone, cs *ClientSettingsType) error {
	slog.Debug("Parsing imports")
	pr.Start("Parsing imports")
	defer pr.End("Parsing import end")
	qc := tree_sitter.NewQueryCursor()
	defer qc.Close()

	language := tree_sitter.NewLanguage(tree_sitter_python.Language())
	query, err := tree_sitter.NewQuery(language, getTreeSitterImportQuery())
	if err != nil {
		return err
	}
	ProjectFiles.Range(func(key, value any) bool {
		file := value.(*PythonFile)
		imports := processImports(file, qc, query, cs)
		file.Imports = imports
		return true
	})
	slog.Debug("Imports parsed")
	return nil
}

func findImportSymbol(cs *ClientSettingsType, i *Import) (*Symbol, error) {
	module := strings.ReplaceAll(i.SourceModule, ".", string(filepath.Separator))
	var moduleFile string
	for _, workspaceRoot := range cs.ModulesPath() {
		path := filepath.Join(workspaceRoot, module)

		// Try as a module: foo/bar.py
		filePath := path + ".py"
		slog.Debug("Checking for module", slog.String("filePath", filePath))
		if _, err := os.Stat(filePath); err == nil {
			slog.Debug("Module file found", slog.String("filePath", filePath))
			moduleFile = filePath
			break
		}

		// Try as a package: foo/bar/__init__.py
		filePath = filepath.Join(path, "__init__.py")
		slog.Debug("Checking for package", slog.String("filePath", filePath))
		if _, err := os.Stat(filePath); err == nil {
			moduleFile = filePath
			break
		}
	}
	if moduleFile == "" {
		slog.Warn("File for module not found", slog.String("module", i.SourceModule))
		return nil, nil
	}
	fileUrl := "file://" + moduleFile
	dstFile, err := GetPythonFile(fileUrl)
	if err != nil {
		dstFile, err = ImportPythonFileFromFile(moduleFile, true)
		if err != nil {
			slog.Warn("Error importing file", slog.String("fileUrl", fileUrl), slog.Any("error", err))
			return nil, err
		}
	}
	i.PythonFile = dstFile
	fileSymbols, err := dstFile.FileSymbols("")
	if err != nil {
		slog.Warn("Error getting file symbols", slog.String("fileUrl", fileUrl), slog.Any("error", err))
		return nil, err
	}
	for _, symbol := range fileSymbols {
		if symbol.Name == i.ImportedName {
			return symbol, nil
		}
	}
	slog.Debug("Symbol not found", slog.String("importedName", i.ImportedName), slog.String("sourceModule", i.SourceModule))
	return nil, nil
}

func processImports(pythonFile *PythonFile, qc *tree_sitter.QueryCursor, query *tree_sitter.Query, cs *ClientSettingsType) []Import {
	imports := []Import{}
	matches := qc.Matches(query, pythonFile.GetOrCreateAst(), []byte(pythonFile.Text))
	for match := matches.Next(); match != nil; match = matches.Next() {
		var sourceModule string
		var aliasName string
		var importedName string

		for _, capture := range match.Captures {
			captureName := query.CaptureNames()[capture.Index]
			captureText := capture.Node.Utf8Text([]byte(pythonFile.Text))

			switch captureName {
			case "module":
				sourceModule = captureText
			case "alias":
				aliasName = captureText
			case "imported_name":
				importedName = captureText
			}
		}
		if sourceModule != "" {
			i := Import{
				Alias:        aliasName,
				SourceModule: sourceModule,
				ImportedName: importedName,
			}
			symbol, err := findImportSymbol(cs, &i)
			if err != nil {
				slog.Warn("Error finding import symbol", slog.String("sourceModule", sourceModule), slog.Any("error", err))
			} else {
				i.Symbol = symbol
			}

			imports = append(imports, i)
		}
	}
	return imports
}
func getTreeSitterImportQuery() string {
	return `
;; import pandas
(import_statement
  name: (dotted_name) @module
  (#set! "type" "import"))

;; import pandas as pd
(import_statement
  name: (aliased_import
    name: (dotted_name) @module
    alias: (identifier) @alias)
  (#set! "type" "import"))

;; from module import single_name
(import_from_statement
  module_name: (dotted_name) @module
  name: (dotted_name) @imported_name
  (#set! "type" "from_import"))

;; from module import name as alias
(import_from_statement
  module_name: (dotted_name) @module
  name: (aliased_import
    name: (dotted_name) @imported_name
    alias: (identifier) @alias)
  (#set! "type" "from_import"))

    `
}
