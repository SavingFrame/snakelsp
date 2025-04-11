package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"snakelsp/internal/progress"
	"sync"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

var OpenFiles sync.Map

var ProjectFiles sync.Map

type PythonFile struct {
	Url     string
	Text    string
	astTree *tree_sitter.Tree
	astRoot *tree_sitter.Node
}

func ParseProject(projectPath string, envPath string, progress *progress.WorkDone) error {
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
		pythonFile := NewPythonFile(url, string(file))
		pythonFiles = append(pythonFiles, pythonFile)
		ProjectFiles.Store(url, pythonFile)
		return nil
	})
	bulkParseAst(pythonFiles, progress)
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
		Url:  url,
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

func (p *PythonFile) ExtractTextFromNode(node *tree_sitter.Node) string {
	startByte := node.StartByte()
	endByte := node.EndByte()

	return string(p.Text[startByte:endByte])
}

func (p *PythonFile) CloseFile() {
	p.astTree.Close()
	OpenFiles.Delete(p.Url)
}
