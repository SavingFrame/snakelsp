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

func loadTemplates() {
	once.Do(func() {
		templates = template.Must(template.ParseFS(templateFiles, "*.html"))
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

	workspace.ProjectFiles.Range(func(_, value interface{}) bool {
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

	var symbols []*workspace.Symbol
	wsSymbols, ok := workspace.WorkspaceSymbols.Load(pythonFile)
	if ok {
		symbols = wsSymbols.([]*workspace.Symbol)
	}

	templates.ExecuteTemplate(w, "file.html", struct {
		File    *workspace.PythonFile
		Symbols []*workspace.Symbol
		Imports []workspace.Import
	}{
		File:    pythonFile,
		Symbols: symbols,
		Imports: pythonFile.Imports,
	})
}
