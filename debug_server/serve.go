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

func flattenSymbols(symbols []*workspace.Symbol, depth int) []RenderSymbol {
	var result []RenderSymbol
	for _, sym := range symbols {
		result = append(result, RenderSymbol{Symbol: sym, Depth: depth})
		result = append(result, flattenSymbols(sym.Children, depth+1)...) // recursion
	}
	return result
}
func loadTemplates() {

	var templateFuncs = template.FuncMap{
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

	workspace.ProjectFiles.Range(func(_, value interface{}) bool {
		if file, ok := value.(*workspace.PythonFile); ok {
			// if !file.External {
			// 	files = append(files, file)
			// }
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
	var flatSymbols []RenderSymbol
	if ok {
		symbols = wsSymbols.([]*workspace.Symbol)
		flatSymbols = flattenSymbols(symbols, 0) // Create flattened recursive display list
	}
	slog.Debug("Amount of symbols", slog.Int("amount", len(symbols)))

	slog.Debug("Amount of imports", slog.Int("amount", len(pythonFile.Imports)))
	templates.ExecuteTemplate(w, "file.html", struct {
		File        *workspace.PythonFile
		SymbolsFlat []RenderSymbol
		Imports     []workspace.Import
	}{
		File:        pythonFile,
		SymbolsFlat: flatSymbols,
		Imports:     pythonFile.Imports,
	})
}
