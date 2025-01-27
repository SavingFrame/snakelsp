package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

var OpenFiles sync.Map

var ProjectFiles sync.Map

type PythonFile struct {
	url     string
	Text    string
	astTree *tree_sitter.Tree
	astRoot *tree_sitter.Node
}

func ParseProject(projectPath string, envPath string) error {
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
		pythonFile := NewPythonFile(url, string(file))
		pythonFiles = append(pythonFiles, pythonFile)
		ProjectFiles.Store(url, pythonFile)
		return nil
	})
	bulkParseAst(pythonFiles)
	return nil
}

func GetPythonFile(url string) (*PythonFile, error) {
	file, ok := OpenFiles.Load(url)
	if !ok {
		return nil, fmt.Errorf("file not found")
	}
	return file.(*PythonFile), nil
}

func NewPythonFile(url string, text string) *PythonFile {
	pythonFile := &PythonFile{
		url:  url,
		Text: text,
	}
	OpenFiles.Store(url, pythonFile)
	return pythonFile
}

func (p *PythonFile) parseAst() *tree_sitter.Node {
	parser := tree_sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_python.Language()))
	tree := parser.Parse([]byte(p.Text), nil)
	root := tree.RootNode()
	p.astRoot = root
	// slog.Debug(root.ToSexp())
	return p.astRoot
}

func bulkParseAst(files []*PythonFile) {
	parser := tree_sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_python.Language()))
	for _, file := range files {
		tree := parser.Parse([]byte(file.Text), nil)
		root := tree.RootNode()
		file.astRoot = root
	}
}

func (p *PythonFile) GetOrCreateAst() *tree_sitter.Node {
	if p.astRoot == nil {
		return p.parseAst()
	} else {
		return p.astRoot
	}
}

func (p *PythonFile) ExtractTextFromNode(node *tree_sitter.Node) string {
	startByte := node.StartByte()
	endByte := node.EndByte()

	return string(p.Text[startByte:endByte])
}

func (p *PythonFile) CloseFile() {
	p.astTree.Close()
	OpenFiles.Delete(p.url)
}
