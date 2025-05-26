package workspace

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"snakelsp/internal/messages"
	"snakelsp/internal/progress"
	"snakelsp/pkg/debounce"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

var ProjectFiles sync.Map // Projects files, also 3rd libraries files

type PythonFile struct {
	Url      string
	Text     string
	astTree  *tree_sitter.Tree
	astRoot  *tree_sitter.Node
	External bool
	isOpened bool

	Imports []Import

	debouncer debounce.Debouncer
}

func ParseProjectFiles(projectPath string, envPath string, progress *progress.WorkDone) error {
	progress.Start("Parsing project files")
	excludedFolders := []string{".git", ".venv", ".mypy_cache"}
	pythonFiles := []*PythonFile{}
	filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && (envPath == path || slices.Contains(excludedFolders, info.Name())) {
			return filepath.SkipDir
		}
		if filepath.Ext(path) != ".py" {
			return nil
		}
		url := "file://" + path
		file, rfErr := os.ReadFile(path)
		if rfErr != nil {
			return err
		}
		pythonFile := NewPythonFile(url, string(file), false, false)
		pythonFiles = append(pythonFiles, pythonFile)
		return nil
	})
	bulkParseAst(pythonFiles, progress)
	return nil
}

func GetPythonFile(url string) (*PythonFile, error) {
	file, ok := ProjectFiles.Load(url)
	if !ok {
		return nil, fmt.Errorf("file in the ProjectFiles map not found")
	}
	return file.(*PythonFile), nil
}

func NewPythonFile(url string, text string, external, isOpen bool) *PythonFile {
	pythonFile := &PythonFile{
		Url:       url,
		Text:      text,
		External:  external,
		isOpened:  isOpen,
		debouncer: debounce.NewDebounce(2 * time.Second),
	}
	ProjectFiles.LoadOrStore(url, pythonFile)
	return pythonFile
}

func ImportPythonFileFromFile(path string, external bool) (*PythonFile, error) {
	url := "file://" + path
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewPythonFile(url, string(content), external, false), nil
}

func (p *PythonFile) parseAst() *tree_sitter.Node {
	parser := tree_sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_python.Language()))
	tree := parser.Parse([]byte(p.Text), nil)
	root := tree.RootNode()
	p.astRoot = root
	p.astTree = tree
	return p.astRoot
}

func bulkParseAst(files []*PythonFile, pr *progress.WorkDone) {
	parser := tree_sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_python.Language()))
	totalFiles := len(files)
	for i, file := range files {
		pr.Report(fmt.Sprintf("Processing file %d of %d", i+1, totalFiles), uint16(float64(i+1)/float64(totalFiles)*100))
		tree := parser.Parse([]byte(file.Text), nil)
		root := tree.RootNode()
		file.astRoot = root
	}
	pr.End("Finished parsing project files")
}

func (p *PythonFile) GetOrCreateAst() *tree_sitter.Node {
	if p.astRoot == nil {
		return p.parseAst()
	} else {
		return p.astRoot
	}
}

func (p *PythonFile) CloseFile() {
	p.isOpened = false
	p.astTree.Close()
}

func (p *PythonFile) parseOnUpdate() {
	slog.Debug("Parsing file on update", slog.String("file", p.Url))
	p.parseAst()
	p.ParseImports()
	p.parseSymbols()
}

func (f *PythonFile) ApplyChange(contentChanges []messages.TextDocumentContentChangeEvent) {
	slog.Debug("Applying changes to file", slog.String("file", f.Url))
	content := f.Text
	for _, change := range contentChanges {
		content = fullContentFromChange(change.Range, content, change.Text)
	}
	f.Text = content
	slog.Debug("Updated file content", slog.String("content", f.Text))
	f.debouncer.Debounce(f.parseOnUpdate)
}

func fullContentFromChange(r *messages.Range, content, newText string) string {
	lines := strings.Split(content, "\n")

	startLine := r.Start.Line
	endLine := r.End.Line
	startCharacter := r.Start.Character
	endCharacter := r.End.Character
	// Ensure the start and end line indices are within bounds
	if int(startLine) >= len(lines) || int(endLine) >= len(lines) {
		slog.Warn("Invalid line numbers")
		return content
	}

	// Extract the lines where the range starts and ends
	startTargetLine := lines[startLine]
	endTargetLine := lines[endLine]

	// Ensure that character positions are valid within the respective lines
	if int(startCharacter) > len(startTargetLine) || int(endCharacter) > len(endTargetLine) {
		slog.Warn("Invalid character positions")
		return content
	}

	// Handle different cases based on start and end indices
	if startLine == endLine {
		// Case 1: Change occurs within a single line
		// Replace the range directly in the same line
		updatedLine := startTargetLine[:startCharacter] + newText + startTargetLine[endCharacter:]
		lines[startLine] = updatedLine
	} else {
		// Case 2: Change spans multiple lines
		// Compose new content from fragments:
		// - Start of the first line up to `startCharacter`
		startFragment := startTargetLine[:startCharacter]

		// - End of the last line from `endCharacter`
		endFragment := endTargetLine[endCharacter:]

		// Replace the lines in between with the new text
		updatedLine := startFragment + newText + endFragment
		lines = append(lines[:startLine], append([]string{updatedLine}, lines[endLine+1:]...)...)
	}

	// Reassemble the lines back into the full content
	return strings.Join(lines, "\n")
}
