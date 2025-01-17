package messages

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
	RootPath *string `json:"rootPath,omitempty"`

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
	InitializationOptions any `json:"initializationOptions,omitempty"`

	/**
	 * The capabilities provided by the client (editor or tool)
	 */
	// TODO:

	// Capabilities ClientCapabilities `json:"capabilities"`

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
