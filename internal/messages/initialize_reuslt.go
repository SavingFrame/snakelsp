package messages

type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
type TextDocumentSyncKind Integer

/**
 * Defines how the host (editor) should sync document changes to the language
 * server.
 */
const (
	/**
	 * Documents should not be synced at all.
	 */
	TextDocumentSyncKindNone = TextDocumentSyncKind(0)

	/**
	 * Documents are synced by always sending the full content
	 * of the document.
	 */
	TextDocumentSyncKindFull = TextDocumentSyncKind(1)

	/**
	 * Documents are synced by sending the full content on open.
	 * After that only incremental updates to the document are
	 * send.
	 */
	TextDocumentSyncKindIncremental = TextDocumentSyncKind(2)
)

type textDocumentSyncOptions struct {
	OpenClose bool                 `json:"openClose"`
	Change    TextDocumentSyncKind `json:"change"`
}

type serverCapabilities struct {
	TextDocumentSync        *textDocumentSyncOptions `json:"textDocumentSync"`
	DefinitionProvider      bool                     `json:"definitionProvider"`
	WorkspaceSymbolProvider bool                     `json:"workspaceSymbolProvider"`
	DocumentSymbolProvider  bool                     `json:"documentSymbolProvider"`
}

type InitializeResult struct {
	ServerInfo         *serverInfo         `json:"serverInfo"`
	ServerCapabilities *serverCapabilities `json:"capabilities"`
}

func NewInitializeResult() *InitializeResult {
	return &InitializeResult{
		ServerCapabilities: &serverCapabilities{
			TextDocumentSync: &textDocumentSyncOptions{
				OpenClose: true,
				Change:    TextDocumentSyncKindIncremental,
			},
			DefinitionProvider:      true,
			WorkspaceSymbolProvider: true,
			DocumentSymbolProvider:  true,
		},
		ServerInfo: &serverInfo{
			Name:    "SnakeLSP",
			Version: "0.0.1",
		},
	}
}
