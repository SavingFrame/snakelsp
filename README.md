# 🐍 SnakeLSP

**SnakeLSP** is a Language Server (LSP) for Python written in Go. It is designed to provide fast and efficient language features by pre-parsing the entire project at startup, improving performance for large-scale Python codebases.

## 🚀 Features

- **Performance-oriented**: Parses the full project once at startup, avoiding redundant re-parsing on requests.
- **Basic LSP support**:
  - **(Planned) Go-to Definition, Auto-completion, Diagnostics**
  - **File open, change, and close events**
  - **Document and workspace symbol retrieval**
- **Written in Go**: Designed for speed and efficiency compared to Python-based LSP implementations.

## 📜 Supported LSP Handlers

| Method                          | Handler                  | Description |
|---------------------------------|--------------------------|-------------|
| `initialize`                    | `HandleInitialize`       | Initializes the LSP server |
| `initialized`                   | `HandleInitialized`      | Called after initialization |
| `textDocument/didOpen`          | `HandleDidOpen`          | Handles opening a new document |
| `textDocument/didChange`        | `HandleDidChange`        | Tracks document changes |
| `textDocument/didClose`         | `HandleDidClose`         | Handles document close events |
| `shutdown`                      | `HandleShutdown`         | Gracefully shuts down the server |
| `textDocument/definition`       | _Planned_ `HandleGotoDefinition`   | Jumps to the definition of a symbol(Not implemented yet) |
| `workspace/symbol`              | `HandleWorkspaceSymbol`  | Retrieves all symbols in the workspace |
| `textDocument/documentSymbol`   | `HandleDocumentSymbol`   | Retrieves document-level symbols |
| `$/cancelRequest`               | _Dummy_ `HandleCancelRequest`    | Handles request cancellations (Placeholder, does nothing currently)|

## 🏗️ Installation

### Prerequisites

- **Go 1.18+** installed
- **A Language Server Protocol (LSP) client** (VS Code, Neovim, etc.)

### Build & Install

```sh
git clone https://github.com/your_username/snakelsp.git
cd snakelsp
go build -o snakelsp
```

### Neovim Configuration

#### LspConfig

```lua
local lspconfig = require("lspconfig")
local configs = require("lspconfig.configs")

if not configs.snakelsp then
  local root_files = {
    "pyproject.toml",
    "requirements.txt",
    "Pipfile",
    "pyrightconfig.json",
    ".git",
  }
  configs.snakelsp = {
    default_config = {
      cmd = { "snakelsp" },
      root_dir = function(fname)
        return lspconfig.util.root_pattern(unpack(root_files))(fname)
      end,
      single_file_support = true,
      filetypes = { "python" },
      settings = {
      },
    },
  }
end
```

## 📅 Roadmap

- [ ] **Implement LSP Notifications**
  - [ ] `window/showMessage` → Show messages (errors, warnings, info) to the user
  - [ ] `window/logMessage` → Log messages for debugging inside the editor
  - [ ] `$/progress` → Support reporting progress (useful for indexing phase)
  - [ ] `workspace/didChangeWatchedFiles` → Handle file changes from outside the editor (e.g., Git updates)
- [ ] Implement **Go-to Definition** (`textDocument/definition`)
- [ ] Add **Hover support** (`textDocument/hover`)
- [ ] **Class Hierarchy Navigation** (like PyCharm)
  - [ ] **Find subclasses (inheritors)** (`textDocument/typeHierarchy/subtypes`)
  - [ ] **Find parent classes** (`textDocument/typeHierarchy/supertypes`)
- [ ] Optimize performance for large projects
- [ ] Multi-threaded parsing

## 🤝 Contributing

Contributions are welcome! If you have ideas for improving **SnakeLSP**, feel free to open an issue or submit a pull request.

## 📄 License

MIT License. See [`LICENSE`](./LICENSE) for details.

---

🚀 Built for speed, with ❤️ in Go.
