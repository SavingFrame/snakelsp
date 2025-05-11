package workspace

import (
	"errors"
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

func (f *PythonFile) ParseImports() ([]Import, error) {
	slog.Debug("Parse Imports")
	return f.parseImports(true)
}

func (f *PythonFile) parseImports(withResolvedSymbols bool) ([]Import, error) {
	qc := tree_sitter.NewQueryCursor()
	defer qc.Close()
	language := tree_sitter.NewLanguage(tree_sitter_python.Language())
	query, err := tree_sitter.NewQuery(language, getTreeSitterImportQuery())
	if err != nil {
		return nil, err
	}
	imports := processImports(f, qc, query, withResolvedSymbols)
	return imports, nil
}

func (f *PythonFile) GetImports() ([]Import, error) {
	if f.Imports == nil {
		imports, err := f.ParseImports()
		if err != nil {
			return nil, err
		}
		f.Imports = imports
	}
	return f.Imports, nil
}

func BulkParseImports(pr *progress.WorkDone) error {
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
		if file.External {
			return true
		}
		imports := processImports(file, qc, query, true)
		file.Imports = imports
		return true
	})
	slog.Debug("Imports parsed")
	return nil
}

func resolveImportSymbol(file *PythonFile, imp *Import) (*Symbol, error) {
	// slog.Debug("Resolve import symbol for file", slog.String("fileUrl", file.Url), slog.String("importedName", imp.ImportedName), slog.String("sourceModule", imp.SourceModule))
	module := strings.ReplaceAll(imp.SourceModule, ".", string(filepath.Separator))
	var moduleFile string
	for _, workspaceRoot := range ClientSettings.ModulesPath() {
		path := filepath.Join(workspaceRoot, module)

		// Try as a module: foo/bar.py
		filePath := path + ".py"
		if _, err := os.Stat(filePath); err == nil {
			moduleFile = filePath
			break
		}

		// Try as a package: foo/bar/__init__.py
		filePath = filepath.Join(path, "__init__.py")
		if _, err := os.Stat(filePath); err == nil {
			moduleFile = filePath
			break
		}
	}
	if moduleFile == "" {
		slog.Warn("File for module not found", slog.String("module", imp.SourceModule))
		return nil, errors.New("module file not found")
	}

	// Get or create destination pythonFile
	fileUrl := "file://" + moduleFile
	dstFile, err := GetPythonFile(fileUrl)
	if err != nil {
		dstFile, err = ImportPythonFileFromFile(moduleFile, true)
		if err != nil {
			slog.Warn("Error importing file", slog.String("fileUrl", fileUrl), slog.Any("error", err))
			return nil, err
		}
	}

	// Get symbols from the destination file
	fileSymbols, err := dstFile.FileSymbols("")
	if err != nil {
		slog.Warn("Error getting file symbols", slog.String("fileUrl", fileUrl), slog.Any("error", err))
		return nil, err
	}

	// Find the specific symbol
	for _, symbol := range fileSymbols {
		if symbol.Name == imp.ImportedName {
			return symbol, nil
		}
	}

	var imports []Import
	if dstFile.Imports == nil {
		imports, err = dstFile.parseImports(false)
		if err != nil {
			slog.Warn("Error parsing nested imports", slog.String("fileUrl", fileUrl), slog.Any("error", err))
		}
	} else {
		imports = dstFile.Imports
	}
	for _, nestedImport := range imports {
		if nestedImport.ImportedName == imp.ImportedName {
			slog.Debug("Found nested import", slog.String("importedName", nestedImport.ImportedName), slog.String("sourceModule", nestedImport.SourceModule))
			return resolveImportSymbol(file, &nestedImport)
		}
	}

	return nil, errors.New("symbol not found")
}

func processImports(pythonFile *PythonFile, qc *tree_sitter.QueryCursor, query *tree_sitter.Query, withResolvedSymbols bool) []Import {
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
			if withResolvedSymbols {
				symbol, err := resolveImportSymbol(pythonFile, &i)
				if err != nil {
					slog.Warn("Error finding import symbol", slog.String("sourceModule", sourceModule), slog.Any("error", err))
				} else {
					i.Symbol = symbol
					i.PythonFile = symbol.File
				}
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
