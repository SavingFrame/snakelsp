package workspace

import (
	"fmt"
	"log/slog"
	"sync"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

var OpenFiles sync.Map

type PythonFile struct {
	url     string
	text    string
	astTree *tree_sitter.Tree
	astRoot *tree_sitter.Node
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
		text: text,
	}
	OpenFiles.Store(url, pythonFile)
	return pythonFile
}

func (p *PythonFile) parseAst() *tree_sitter.Node {
	parser := tree_sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_python.Language()))
	tree := parser.Parse([]byte(p.text), nil)
	root := tree.RootNode()
	p.astRoot = root
	slog.Debug(root.ToSexp())
	return p.astRoot
}

func (p *PythonFile) GetOrCreateAst(node *tree_sitter.Node) *tree_sitter.Node {
	if p.astRoot == nil {
		return p.parseAst()
	} else {
		return p.astRoot
	}
}

func (p *PythonFile) CloseFile() {
	p.astTree.Close()
	OpenFiles.Delete(p.url)
}
