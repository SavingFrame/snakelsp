package messages

type InitializationOptionsParams struct {
	VirtualEnvPath string `json:"virtualenv_path"`
}

type InitializeParams struct {
	WorkDoneProgressParams

	/**
	 * The process Id of the parent process that started the server. Is null if
	 * the process has not been started by another process. If the parent
	 * process is not alive then the server should exit (see exit notification)
	 * its process.
	 */
	ProcessID *Integer `json:"processId"`

	/**
	 * Information about the client
	 *
	 * @since 3.15.0
	 */
	ClientInfo *struct {
		/**
		 * The name of the client as defined by the client.
		 */
		Name string `json:"name"`

		/**
		 * The client's version as defined by the client.
		 */
		Version *string `json:"version,omitempty"`
	} `json:"clientInfo,omitempty"`

	/**
	 * The locale the client is currently showing the user interface
	 * in. This must not necessarily be the locale of the operating
	 * system.
	 *
	 * Uses IETF language tags as the value's syntax
	 * (See https://en.wikipedia.org/wiki/IETF_language_tag)
	 *
	 * @since 3.16.0
	 */
	Locale *string `json:"locale,omitempty"`

	/**
	 * The rootPath of the workspace. Is null
	 * if no folder is open.
	 *
	 * @deprecated in favour of `rootUri`.
	 */
	RootPath string `json:"rootPath,omitempty"`

	/**
	 * The rootUri of the workspace. Is null if no
	 * folder is open. If both `rootPath` and `rootUri` are set
	 * `rootUri` wins.
	 *
	 * @deprecated in favour of `workspaceFolders`
	 */
	RootURI *DocumentUri `json:"rootUri"`

	/**
	 * User provided initialization options.
	 */
	InitializationOptions *InitializationOptionsParams `json:"initializationOptions,omitempty"`

	/**
	 * The capabilities provided by the client (editor or tool)
	 */
	// TODO:

	Capabilities ClientCapabilities `json:"capabilities"`

	/**
	 * The initial trace setting. If omitted trace is disabled ('off').
	 */
	Trace *TraceValue `json:"trace,omitempty"`

	/**
	 * The workspace folders configured in the client when the server starts.
	 * This property is only available if the client supports workspace folders.
	 * It can be `null` if the client supports workspace folders but none are
	 * configured.
	 *
	 * @since 3.6.0
	 */
	WorkspaceFolders []WorkspaceFolder `json:"workspaceFolders,omitempty"`
}

type ClientCapabilities struct {
	/**
	 * Workspace specific client capabilities.
	 */
	Workspace *struct {
		/**
		 * The client supports applying batch edits
		 * to the workspace by supporting the request
		 * 'workspace/applyEdit'
		 */
		ApplyEdit *bool `json:"applyEdit,omitempty"`

		/**
		 * Capabilities specific to `WorkspaceEdit`s
		 */
		WorkspaceEdit *WorkspaceEditClientCapabilities `json:"workspaceEdit,omitempty"`

		/**
		 * Capabilities specific to the `workspace/didChangeConfiguration`
		 * notification.
		 */
		DidChangeConfiguration *DidChangeConfigurationClientCapabilities `json:"didChangeConfiguration,omitempty"`

		/**
		 * Capabilities specific to the `workspace/didChangeWatchedFiles`
		 * notification.
		 */
		DidChangeWatchedFiles *DidChangeWatchedFilesClientCapabilities `json:"didChangeWatchedFiles,omitempty"`

		/**
		 * Capabilities specific to the `workspace/symbol` request.
		 */
		Symbol *WorkspaceSymbolClientCapabilities `json:"symbol,omitempty"`

		/**
		 * Capabilities specific to the `workspace/executeCommand` request.
		 */
		ExecuteCommand *ExecuteCommandClientCapabilities `json:"executeCommand,omitempty"`

		/**
		 * The client has support for workspace folders.
		 *
		 * @since 3.6.0
		 */
		WorkspaceFolders *bool `json:"workspaceFolders,omitempty"`

		/**
		 * The client supports `workspace/configuration` requests.
		 *
		 * @since 3.6.0
		 */
		Configuration *bool `json:"configuration,omitempty"`

		/**
		 * Capabilities specific to the semantic token requests scoped to the
		 * workspace.
		 *
		 * @since 3.16.0
		 */
		SemanticTokens *SemanticTokensWorkspaceClientCapabilities `json:"semanticTokens,omitempty"`

		/**
		 * Capabilities specific to the code lens requests scoped to the
		 * workspace.
		 *
		 * @since 3.16.0
		 */
		CodeLens *CodeLensWorkspaceClientCapabilities `json:"codeLens,omitempty"`

		/**
		 * The client has support for file requests/notifications.
		 *
		 * @since 3.16.0
		 */
		FileOperations *struct {
			/**
			 * Whether the client supports dynamic registration for file
			 * requests/notifications.
			 */
			DynamicRegistration *bool `json:"dynamicRegistration,omitempty"`

			/**
			 * The client has support for sending didCreateFiles notifications.
			 */
			DidCreate *bool `json:"didCreate,omitempty"`

			/**
			 * The client has support for sending willCreateFiles requests.
			 */
			WillCreate *bool `json:"willCreate,omitempty"`

			/**
			 * The client has support for sending didRenameFiles notifications.
			 */
			DidRename *bool `json:"didRename,omitempty"`

			/**
			 * The client has support for sending willRenameFiles requests.
			 */
			WillRename *bool `json:"willRename,omitempty"`

			/**
			 * The client has support for sending didDeleteFiles notifications.
			 */
			DidDelete *bool `json:"didDelete,omitempty"`

			/**
			 * The client has support for sending willDeleteFiles requests.
			 */
			WillDelete *bool `json:"willDelete,omitempty"`
		} `json:"fileOperations,omitempty"`
	} `json:"workspace,omitempty"`

	/**
	 * Text document specific client capabilities.
	 */
	TextDocument *TextDocumentClientCapabilities `json:"textDocument,omitempty"`

	/**
	 * Window specific client capabilities.
	 */
	Window *struct {
		/**
		 * Whether client supports handling progress notifications. If set
		 * servers are allowed to report in `workDoneProgress` property in the
		 * request specific server capabilities.
		 *
		 * @since 3.15.0
		 */
		WorkDoneProgress *bool `json:"workDoneProgress,omitempty"`

		/**
		 * Capabilities specific to the showMessage request
		 *
		 * @since 3.16.0
		 */
		// ShowMessage *ShowMessageRequestClientCapabilities `json:"showMessage,omitempty"`

		/**
		 * Client capabilities for the show document request.
		 *
		 * @since 3.16.0
		 */
		// ShowDocument *ShowDocumentClientCapabilities `json:"showDocument,omitempty"`
	} `json:"window,omitempty"`

	/**
	 * General client capabilities.
	 *
	 * @since 3.16.0
	 */
	General *struct {
		/**
		 * Client capabilities specific to regular expressions.
		 *
		 * @since 3.16.0
		 */
		// RegularExpressions *RegularExpressionsClientCapabilities `json:"regularExpressions,omitempty"`

		/**
		 * Client capabilities specific to the client's markdown parser.
		 *
		 * @since 3.16.0
		 */
		// Markdown *MarkdownClientCapabilities `json:"markdown,omitempty"`
	} `json:"general,omitempty"`

	/**
	 * Experimental client capabilities.
	 */
	Experimental any `json:"experimental,omitempty"`
}

// https://microsoft.github.io/language-server-protocol/specifications/specification-3-16#textDocument_documentSymbol

type DocumentSymbolClientCapabilities struct {
	/**
	 * Whether document symbol supports dynamic registration.
	 */
	DynamicRegistration *bool `json:"dynamicRegistration,omitempty"`

	/**
	 * Specific capabilities for the `SymbolKind` in the
	 * `textDocument/documentSymbol` request.
	 */
	SymbolKind *struct {
		/**
		 * The symbol kind values the client supports. When this
		 * property exists the client also guarantees that it will
		 * handle values outside its set gracefully and falls back
		 * to a default value when unknown.
		 *
		 * If this property is not present the client only supports
		 * the symbol kinds from `File` to `Array` as defined in
		 * the initial version of the protocol.
		 */
		ValueSet []SymbolKind `json:"valueSet,omitempty"`
	} `json:"symbolKind,omitempty"`

	/**
	 * The client supports hierarchical document symbols.
	 */
	HierarchicalDocumentSymbolSupport *bool `json:"hierarchicalDocumentSymbolSupport,omitempty"`

	/**
	 * The client supports tags on `SymbolInformation`. Tags are supported on
	 * `DocumentSymbol` if `hierarchicalDocumentSymbolSupport` is set to true.
	 * Clients supporting tags have to handle unknown tags gracefully.
	 *
	 * @since 3.16.0
	 */
	TagSupport *struct {
		/**
		 * The tags supported by the client.
		 */
		ValueSet []SymbolTag `json:"valueSet"`
	} `json:"tagSupport,omitempty"`

	/**
	 * The client supports an additional label presented in the UI when
	 * registering a document symbol provider.
	 *
	 * @since 3.16.0
	 */
	LabelSupport *bool `json:"labelSupport,omitempty"`
}

// https://microsoft.github.io/language-server-protocol/specifications/specification-3-16#textDocument_prepareCallHierarchy

type CallHierarchyClientCapabilities struct {
	/**
	 * Whether implementation supports dynamic registration. If this is set to
	 * `true` the client supports the new `(TextDocumentRegistrationOptions &
	 * StaticRegistrationOptions)` return value for the corresponding server
	 * capability as well.
	 */
	DynamicRegistration *bool `json:"dynamicRegistration,omitempty"`
}

// https://microsoft.github.io/language-server-protocol/specifications/specification-3-16#textDocument_declaration

type DeclarationClientCapabilities struct {
	/**
	 * Whether declaration supports dynamic registration. If this is set to
	 * `true` the client supports the new `DeclarationRegistrationOptions`
	 * return value for the corresponding server capability as well.
	 */
	DynamicRegistration *bool `json:"dynamicRegistration,omitempty"`

	/**
	 * The client supports additional metadata in the form of declaration links.
	 */
	LinkSupport *bool `json:"linkSupport,omitempty"`
}

/**
 * Text document specific client capabilities.
 */
type TextDocumentClientCapabilities struct {
	// Synchronization *TextDocumentSyncClientCapabilities `json:"synchronization,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/completion` request.
	 */
	// Completion *CompletionClientCapabilities `json:"completion,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/hover` request.
	 */
	// Hover *HoverClientCapabilities `json:"hover,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/signatureHelp` request.
	 */
	// SignatureHelp *SignatureHelpClientCapabilities `json:"signatureHelp,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/declaration` request.
	 *
	 * @since 3.14.0
	 */
	Declaration *DeclarationClientCapabilities `json:"declaration,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/definition` request.
	 */
	// Definition *DefinitionClientCapabilities `json:"definition,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/typeDefinition` request.
	 *
	 * @since 3.6.0
	 */
	// TypeDefinition *TypeDefinitionClientCapabilities `json:"typeDefinition,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/implementation` request.
	 *
	 * @since 3.6.0
	 */
	// Implementation *ImplementationClientCapabilities `json:"implementation,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/references` request.
	 */
	// References *ReferenceClientCapabilities `json:"references,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/documentHighlight` request.
	 */
	// DocumentHighlight *DocumentHighlightClientCapabilities `json:"documentHighlight,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/documentSymbol` request.
	 */
	DocumentSymbol *DocumentSymbolClientCapabilities `json:"documentSymbol,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/codeAction` request.
	 */
	// CodeAction *CodeActionClientCapabilities `json:"codeAction,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/codeLens` request.
	 */
	// CodeLens *CodeLensClientCapabilities `json:"codeLens,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/documentLink` request.
	 */
	// DocumentLink *DocumentLinkClientCapabilities `json:"documentLink,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/documentColor` and the
	 * `textDocument/colorPresentation` request.
	 *
	 * @since 3.6.0
	 */
	// ColorProvider *DocumentColorClientCapabilities `json:"colorProvider,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/formatting` request.
	 */
	// Formatting *DocumentFormattingClientCapabilities `json:"formatting,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/rangeFormatting` request.
	 */
	// RangeFormatting *DocumentRangeFormattingClientCapabilities `json:"rangeFormatting,omitempty"`

	/** request.
	 * Capabilities specific to the `textDocument/onTypeFormatting` request.
	 */
	// OnTypeFormatting *DocumentOnTypeFormattingClientCapabilities `json:"onTypeFormatting,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/rename` request.
	 */
	// Rename *RenameClientCapabilities `json:"rename,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/publishDiagnostics`
	 * notification.
	 */
	// PublishDiagnostics *PublishDiagnosticsClientCapabilities `json:"publishDiagnostics,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/foldingRange` request.
	 *
	 * @since 3.10.0
	 */
	// FoldingRange *FoldingRangeClientCapabilities `json:"foldingRange,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/selectionRange` request.
	 *
	 * @since 3.15.0
	 */
	// SelectionRange *SelectionRangeClientCapabilities `json:"selectionRange,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/linkedEditingRange` request.
	 *
	 * @since 3.16.0
	 */
	// LinkedEditingRange *LinkedEditingRangeClientCapabilities `json:"linkedEditingRange,omitempty"`

	/**
	 * Capabilities specific to the various call hierarchy requests.
	 *
	 * @since 3.16.0
	 */
	CallHierarchy *CallHierarchyClientCapabilities `json:"callHierarchy,omitempty"`

	/**
	 * Capabilities specific to the various semantic token requests.
	 *
	 * @since 3.16.0
	 */
	// SemanticTokens *SemanticTokensClientCapabilities `json:"semanticTokens,omitempty"`

	/**
	 * Capabilities specific to the `textDocument/moniker` request.
	 *
	 * @since 3.16.0
	 */
	// Moniker *MonikerClientCapabilities `json:"moniker,omitempty"`
}

// https://microsoft.github.io/language-server-protocol/specifications/specification-3-16#codeLens_refresh

type CodeLensWorkspaceClientCapabilities struct {
	/**
	 * Whether the client implementation supports a refresh request sent from the
	 * server to the client.
	 *
	 * Note that this event is global and will force the client to refresh all
	 * code lenses currently shown. It should be used with absolute care and is
	 * useful for situation where a server for example detect a project wide
	 * change that requires such a calculation.
	 */
	RefreshSupport *bool `json:"refreshSupport,omitempty"`
}
type SemanticTokensWorkspaceClientCapabilities struct {
	/**
	 * Whether the client implementation supports a refresh request sent from
	 * the server to the client.
	 *
	 * Note that this event is global and will force the client to refresh all
	 * semantic tokens currently shown. It should be used with absolute care
	 * and is useful for situation where a server for example detect a project
	 * wide change that requires such a calculation.
	 */
	RefreshSupport *bool `json:"refreshSupport,omitempty"`
}

// https://microsoft.github.io/language-server-protocol/specifications/specification-3-16#workspace_executeCommand

type ExecuteCommandClientCapabilities struct {
	/**
	 * Execute command supports dynamic registration.
	 */
	DynamicRegistration *bool `json:"dynamicRegistration,omitempty"`
}

// https://microsoft.github.io/language-server-protocol/specifications/specification-3-16#workspace_symbol

type WorkspaceSymbolClientCapabilities struct {
	/**
	 * Symbol request supports dynamic registration.
	 */
	DynamicRegistration *bool `json:"dynamicRegistration,omitempty"`

	/**
	 * Specific capabilities for the `SymbolKind` in the `workspace/symbol`
	 * request.
	 */
	SymbolKind *struct {
		/**
		 * The symbol kind values the client supports. When this
		 * property exists the client also guarantees that it will
		 * handle values outside its set gracefully and falls back
		 * to a default value when unknown.
		 *
		 * If this property is not present the client only supports
		 * the symbol kinds from `File` to `Array` as defined in
		 * the initial version of the protocol.
		 */
		ValueSet []SymbolKind `json:"valueSet,omitempty"`
	} `json:"symbolKind,omitempty"`

	/**
	 * The client supports tags on `SymbolInformation`.
	 * Clients supporting tags have to handle unknown tags gracefully.
	 *
	 * @since 3.16.0
	 */
	TagSupport *struct {
		/**
		 * The tags supported by the client.
		 */
		ValueSet []SymbolTag `json:"valueSet"`
	} `json:"tagSupport,omitempty"`
}

// https://microsoft.github.io/language-server-protocol/specifications/specification-3-16#workspace_didChangeWatchedFiles

type DidChangeWatchedFilesClientCapabilities struct {
	/**
	 * Did change watched files notification supports dynamic registration.
	 * Please note that the current protocol doesn't support static
	 * configuration for file changes from the server side.
	 */
	DynamicRegistration *bool `json:"dynamicRegistration,omitempty"`
}

// https://microsoft.github.io/language-server-protocol/specifications/specification-3-16#workspace_didChangeConfiguration

type DidChangeConfigurationClientCapabilities struct {
	/**
	 * Did change configuration notification supports dynamic registration.
	 */
	DynamicRegistration *bool `json:"dynamicRegistration,omitempty"`
}

// https://microsoft.github.io/language-server-protocol/specifications/specification-3-16#workspaceEditClientCapabilities

type WorkspaceEditClientCapabilities struct {
	/**
	 * The client supports versioned document changes in `WorkspaceEdit`s
	 */
	DocumentChanges *bool `json:"documentChanges,omitempty"`

	/**
	 * The resource operations the client supports. Clients should at least
	 * support 'create', 'rename' and 'delete' files and folders.
	 *
	 * @since 3.13.0
	 */
	ResourceOperations []ResourceOperationKind `json:"resourceOperations,omitempty"`

	/**
	 * The failure handling strategy of a client if applying the workspace edit
	 * fails.
	 *
	 * @since 3.13.0
	 */
	FailureHandling *FailureHandlingKind `json:"failureHandling,omitempty"`

	/**
	 * Whether the client normalizes line endings to the client specific
	 * setting.
	 * If set to `true` the client will normalize line ending characters
	 * in a workspace edit to the client specific new line character(s).
	 *
	 * @since 3.16.0
	 */
	NormalizesLineEndings *bool `json:"normalizesLineEndings,omitempty"`

	/**
	 * Whether the client in general supports change annotations on text edits,
	 * create file, rename file and delete file changes.
	 *
	 * @since 3.16.0
	 */
	ChangeAnnotationSupport struct {
		/**
		 * Whether the client groups edits with equal labels into tree nodes,
		 * for instance all edits labelled with "Changes in Strings" would
		 * be a tree node.
		 */
		GroupsOnLabel *bool `json:"groupsOnLabel,omitempty"`
	} `json:"changeAnnotationSupport,omitempty"`
}
