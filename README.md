# 🐍 SnakeLSP

**SnakeLSP** is a high-performance Language Server (LSP) for Python written in Go, designed to add additional features that basedpyright/pyright doesn't support while providing significant performance improvements for large-scale Python codebases.

## 🚀 Key Advantages Over Pyright/BasedPyright

- **⚡ Lightning-fast workspace symbols**: Uses cached values instead of real-time parsing, making workspace symbol requests that take ages in basedpyright nearly instantaneous
- **🔍 PyCharm-like navigation**: Full support for class/method implementations and definitions, similar to PyCharm's advanced code navigation
- **🏗️ Smart caching architecture**: Pre-parses the entire project at startup and maintains intelligent caches, eliminating redundant re-parsing on subsequent requests
- **🚀 Go-powered performance**: Written in Go for superior speed and efficiency compared to Python-based LSP implementations

## 🎯 Features

- **Advanced symbol navigation**:
  - **Workspace symbols** with instant cached lookups
  - **Document symbols** with hierarchical structure
  - **Go-to implementation** for classes and methods
  - **Go-to declaration** with precise location tracking
- **Performance optimizations**:
  - **Single startup parse** of entire project
  - **Intelligent caching** for all symbol requests
  - **Multi-threaded processing** (planned)
- **Standard LSP support**:
  - **File lifecycle management** (open, change, close events)
  - **Progress reporting** for long-running operations

## 📜 Supported LSP Handlers

<details>
<summary>Click to expand supported LSP methods</summary>

| Method                          | Handler                  | Description |
|---------------------------------|--------------------------|-------------|
| `initialize`                    | `HandleInitialize`       | Initializes the LSP server |
| `initialized`                   | `HandleInitialized`      | Called after initialization |
| `textDocument/didOpen`          | `HandleDidOpen`          | Handles opening a new document |
| `textDocument/didChange`        | `HandleDidChange`        | Tracks document changes |
| `textDocument/didClose`         | `HandleDidClose`         | Handles document close events |
| `shutdown`                      | `HandleShutdown`         | Gracefully shuts down the server |
| `textDocument/definition`       | _Planned_ `HandleGotoDefinition`   | Jumps to the definition of a symbol (Not implemented yet) |
| `textDocument/declaration`      | `HandleSymbolDeclaration`          | Jumps to the declaration of a symbol |
| `textDocument/implementation`   | `HandleSymbolImplementation`       | Jumps to the implementation of a symbol |
| `workspace/symbol`              | `HandleWorkspaceSymbol`            | Retrieves all symbols in the workspace |
| `textDocument/documentSymbol`   | `HandleDocumentSymbol`             | Retrieves document-level symbols |
| `$/cancelRequest`               | _Dummy_ `HandleCancelRequest`    | Handles request cancellations (Placeholder, does nothing currently)|
| `window/workDoneProgress/create`               | `progress/progress.go`    | Generate notifications for ongoing progress|
| `$/progress`               | `progress/progress.go`    | Update notifications for ongoing progress|

</details>

## 🏗️ Installation

### Prerequisites

- **A Language Server Protocol (LSP) client** (VS Code, Neovim, etc.)

### Download Latest Release (Recommended)

Download the latest pre-built binary from GitHub releases:

**[📥 Download Latest Release](https://github.com/SavingFrame/snakelsp/releases/latest/)**

1. Download the appropriate binary for your platform
2. Make it executable: `chmod +x snakelsp`
3. Move to your PATH: `sudo mv snakelsp /usr/local/bin/`

### Build from Source (Alternative)

If you prefer to build from source:

```sh
# Prerequisites: Go 1.18+ installed
git clone https://github.com/SavingFrame/snakelsp.git
cd snakelsp
go build -o snakelsp
```

### Neovim Configuration

#### Modern Neovim 0.11+ (Recommended)

For Neovim 0.11+, use the built-in LSP configuration:

```lua
-- Configure SnakeLSP
vim.lsp.config.snakelsp = {
  cmd = { 'snakelsp' },
  filetypes = { 'python' },
  capabilities = {
    textDocument = {
      -- Disable documentSymbol to use basedpyright's implementation
      documentSymbol = vim.NIL,
    },
  },
  root_markers = {
    'pyproject.toml',
    'setup.py',
    'setup.cfg',
    'requirements.txt',
    'Pipfile',
    'pyrightconfig.json',
    '.git',
    '.venv',
  },
  init_options = {
    virtualenv_path = os.getenv('VIRTUAL_ENV'),
  },
}

-- Disable conflicting capabilities in basedpyright/pyright
vim.api.nvim_create_autocmd('LspAttach', {
  group = vim.api.nvim_create_augroup('snakelsp-setup', { clear = true }),
  callback = function(event)
    local client = vim.lsp.get_client_by_id(event.data.client_id)
    if client and client.name == 'basedpyright' then
      -- Disable capabilities that SnakeLSP handles better
      client.server_capabilities.workspaceSymbolProvider = false
      client.server_capabilities.declarationProvider = false
    end
  end,
})

vim.lsp.enable 'snakelsp'
```

#### Legacy LspConfig (Neovim < 0.11)

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
  - [x] `$/progress` → Support reporting progress (useful for indexing phase)
  - [ ] `workspace/didChangeWatchedFiles` → Handle file changes from outside the editor (e.g., Git updates)
- [ ] Implement **Go-to Definition** (`textDocument/definition`)
- [ ] Add **Hover support** (`textDocument/hover`)
- [x] **Class Hierarchy Navigation** (like PyCharm)
  - [x] **Find subclasses (inheritors)** (`textDocument/typeHierarchy/subtypes`)
  - [x] **Find parent classes** (`textDocument/typeHierarchy/supertypes`)
- [ ] Optimize performance for large projects
- [ ] Multi-threaded parsing
- [ ] **Add Testing**
  - [ ] **Unit tests** for core LSP handlers (`initialize`, `didOpen`, `didChange`, etc.)  
  - [ ] **Integration tests** for testing full LSP interactions with editors  

## 🤝 Contributing

Contributions are welcome! If you have ideas for improving **SnakeLSP**, feel free to open an issue or submit a pull request.

## 📄 License

MIT License. See [`LICENSE`](./LICENSE) for details.

---

🚀 Built for speed, with ❤️ in Go.
