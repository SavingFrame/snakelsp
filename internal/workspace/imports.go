package workspace

import (
	"log/slog"

	"snakelsp/internal/progress"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

type Import struct {
	Alias        string // Name in the file (e.g., "foo" or "alias")
	SourceModule string // Full import path (like "foo" or "foo.bar")
	ImportedName string // Specific name if "from foo import Bar" (it's "Bar"), else empty
}

func getTreeSitterImportQuery() string {
	return `
        (import_statement
            name: (dotted_name) @import.module
            alias: (aliased_import 
                name: (identifier) @import.alias)?
        )

        (import_from_statement 
            module_name: (dotted_name) @import.from_module
            (import_list (aliased_import 
                name: (identifier) @import.name 
                alias: (identifier) @import.alias) 
            )?
            (import_list (
                name: (identifier) @import.name_without_alias
            ))?
        )
    `
}

func (f *PythonFile) ParseImports() ([]Import, error) {
	qc := tree_sitter.NewQueryCursor()
	defer qc.Close()
	language := tree_sitter.NewLanguage(tree_sitter_python.Language())
	query, err := tree_sitter.NewQuery(language, getTreeSitterImportQuery())
	if err != nil {
		return nil, err
	}
	imports := processImports(f, qc, query)
	return imports, nil
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
		imports := processImports(file, qc, query)
		file.Imports = imports
		return true
	})
	slog.Debug("Imports parsed")
	return nil
}

func processImports(pythonFile *PythonFile, qc *tree_sitter.QueryCursor, query *tree_sitter.Query) []Import {
	imports := []Import{}
	matches := qc.Matches(query, pythonFile.GetOrCreateAst(), []byte(pythonFile.Text))
	for match := matches.Next(); match != nil; match = matches.Next() {
		var sourceModule string
		var aliasName string
		var importedName string

		for _, capture := range match.Captures {
			captureName := query.CaptureNames()[capture.Index]
			captureText := capture.Node.Utf8Text([]byte(pythonFile.Text))
			slog.Debug("Import found", "capture", captureName, "text", string(captureText))

			switch captureName {
			case "import.module":
				sourceModule = captureText
			case "import.alias":
				aliasName = captureText
			case "import.from_module":
				sourceModule = captureText
			case "import.name":
				importedName = captureText
			case "import.name_without_alias":
				importedName = captureText
			}
		}

		if sourceModule != "" {
			imports = append(imports, Import{
				Alias:        aliasName,
				SourceModule: sourceModule,
				ImportedName: importedName,
			})
		}
	}
	return imports
}
