package debug_server

import (
	"embed"
	"html/template"
	"log/slog"
	"net/http"
	"sync"

	"snakelsp/internal/workspace"
)

//go:embed *.html
var templateFiles embed.FS

var (
	templates *template.Template
	once      sync.Once
)

type RenderSymbol struct {
	Symbol *workspace.Symbol
	Depth  int
}

func loadTemplates() {
	templateFuncs := template.FuncMap{
		"multiply": func(a, b int) int { return a * b },
	}
	once.Do(func() {
		templates = template.Must(template.New("").Funcs(templateFuncs).ParseFS(templateFiles, "*.html"))
	})
}

func StartHTTPServer(addr string) {
	loadTemplates()

	// Routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/file", handleFile)

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		slog.Warn("Debug server failed to start", "error", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	var files []*workspace.PythonFile

	workspace.ProjectFiles.Range(func(_, value any) bool {
		if file, ok := value.(*workspace.PythonFile); ok {
			files = append(files, file)
		}
		return true
	})

	templates.ExecuteTemplate(w, "index.html", struct {
		Files []*workspace.PythonFile
	}{
		Files: files,
	})
}

func handleFile(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "missing 'url' param", http.StatusBadRequest)
		return
	}

	var pythonFile *workspace.PythonFile
	workspace.ProjectFiles.Range(func(_, value interface{}) bool {
		file := value.(*workspace.PythonFile)
		if file.Url == url {
			pythonFile = file
			return false
		}
		return true
	})

	if pythonFile == nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	var flatSymbols []RenderSymbol
	for _, s := range workspace.FlatSymbols.AllFromBack() {
		if s.File.Url == pythonFile.Url {
			flatSymbols = append(flatSymbols, RenderSymbol{
				Symbol: s,
				Depth:  0,
			})
		}
	}

	var projectSymbols []RenderSymbol
	fileSymbols, exists := workspace.WorkspaceSymbols.Load(pythonFile)
	if exists {

		fs, ok := fileSymbols.([]*workspace.Symbol)
		if ok {
			for _, s := range fs {
				projectSymbols = append(projectSymbols, RenderSymbol{
					Symbol: s,
					Depth:  0,
				})
			}
		}
	}

	templates.ExecuteTemplate(w, "file.html", struct {
		File           *workspace.PythonFile
		SymbolsFlat    []RenderSymbol
		ProjectSymbols []RenderSymbol
		Imports        []workspace.Import
	}{
		File:           pythonFile,
		SymbolsFlat:    flatSymbols,
		ProjectSymbols: projectSymbols,
		Imports:        pythonFile.Imports,
	})
}
