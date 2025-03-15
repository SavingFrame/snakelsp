package messages

/**
 * A symbol kind.
 */
type SymbolKind Integer

const (
	SymbolKindFile          = SymbolKind(1)
	SymbolKindModule        = SymbolKind(2)
	SymbolKindNamespace     = SymbolKind(3)
	SymbolKindPackage       = SymbolKind(4)
	SymbolKindClass         = SymbolKind(5)
	SymbolKindMethod        = SymbolKind(6)
	SymbolKindProperty      = SymbolKind(7)
	SymbolKindField         = SymbolKind(8)
	SymbolKindConstructor   = SymbolKind(9)
	SymbolKindEnum          = SymbolKind(10)
	SymbolKindInterface     = SymbolKind(11)
	SymbolKindFunction      = SymbolKind(12)
	SymbolKindVariable      = SymbolKind(13)
	SymbolKindConstant      = SymbolKind(14)
	SymbolKindString        = SymbolKind(15)
	SymbolKindNumber        = SymbolKind(16)
	SymbolKindBoolean       = SymbolKind(17)
	SymbolKindArray         = SymbolKind(18)
	SymbolKindObject        = SymbolKind(19)
	SymbolKindKey           = SymbolKind(20)
	SymbolKindNull          = SymbolKind(21)
	SymbolKindEnumMember    = SymbolKind(22)
	SymbolKindStruct        = SymbolKind(23)
	SymbolKindEvent         = SymbolKind(24)
	SymbolKindOperator      = SymbolKind(25)
	SymbolKindTypeParameter = SymbolKind(26)
)

var SymbolKindMap = map[string]SymbolKind{
	"file":          SymbolKindFile,
	"module":        SymbolKindModule,
	"namespace":     SymbolKindNamespace,
	"package":       SymbolKindPackage,
	"class":         SymbolKindClass,
	"method":        SymbolKindMethod,
	"property":      SymbolKindProperty,
	"field":         SymbolKindField,
	"constructor":   SymbolKindConstructor,
	"enum":          SymbolKindEnum,
	"interface":     SymbolKindInterface,
	"function":      SymbolKindFunction,
	"variable":      SymbolKindVariable,
	"constant":      SymbolKindConstant,
	"string":        SymbolKindString,
	"number":        SymbolKindNumber,
	"boolean":       SymbolKindBoolean,
	"array":         SymbolKindArray,
	"object":        SymbolKindObject,
	"key":           SymbolKindKey,
	"null":          SymbolKindNull,
	"enumMember":    SymbolKindEnumMember,
	"struct":        SymbolKindStruct,
	"event":         SymbolKindEvent,
	"operator":      SymbolKindOperator,
	"typeParameter": SymbolKindTypeParameter,
}

func GetSymbolKind(name string) (SymbolKind, bool) {
	kind, exists := SymbolKindMap[name]
	return kind, exists
}

/**
 * Symbol tags are extra annotations that tweak the rendering of a symbol.
 *
 * @since 3.16.0
 */
type SymbolTag Integer

type WorkspaceSymbolParams struct {
	/**
	 * A query string to filter symbols by. Clients may send an empty
	 * string here to request all symbols.
	 */
	Query string `json:"query"`
}

type DocumentSymbolParams struct {
	/**
	 * The text document.
	 */
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

/**
 * A special workspace symbol that supports locations without a range
 *
 * @since 3.17.0
 */
type WorkspaceSymbol struct {
	/**
	 * The name of this symbol.
	 */
	Name string `json:"name"`

	/**
	 * The kind of this symbol.
	 */
	Kind SymbolKind `json:"kind"`

	/**
	 * Tags for this completion item.
	 */
	Tags SymbolTag `json:"tags,omitempty`

	/**
	 * The name of the symbol containing this symbol. This information is for
	 * user interface purposes (e.g. to render a qualifier in the user interface
	 * if necessary). It can't be used to re-infer a hierarchy for the document
	 * symbols.
	 */
	ContainerName string `json:"containerName,omitempty"`

	/**
	 * The location of this symbol. Whether a server is allowed to
	 * return a location without a range depends on the client
	 * capability `workspace.symbol.resolveSupport`.
	 *
	 * See also `SymbolInformation.location`.
	 */
	Location Location `json:"location"`

	/**
	 * A data entry field that is preserved on a workspace symbol between a
	 * workspace symbol request and a workspace symbol resolve request.
	 */
	data any `json:"data,omitempty"`
}

type DocumentSymbol struct {
	/**
	 * The name of this symbol.
	 */
	Name string `json:"name"`

	/**
	 * More detail for this symbol, e.g the signature of a function.
	 */
	Detail string `json:"detail,omitempty"`

	/**
	 * The kind of this symbol.
	 */
	Kind SymbolKind `json:"kind"`

	/**
	 * Tags for this completion item.
	 */
	Tags []SymbolTag `json:"tags,omitempty"`

	/**
	 * Indicates if this symbol is deprecated.
	 *
	 * @deprecated Use tags instead
	 */
	Deprecated bool `json:"deprecated,omitempty"`

	/**
	 * The range enclosing this symbol not including leading/trailing whitespace
	 * but everything else like comments. This information is typically used to
	 * determine if the clients cursor is inside the symbol to reveal in the
	 * symbol in the UI.
	 */
	Range Range `json:"range"`

	/**
	 * The range that should be selected and revealed when this symbol is being
	 * picked, e.g. the name of a function. Must be contained by the `range`.
	 */
	SelectionRange Range `json:"selectionRange"`

	/**
	 * Children of this symbol, e.g. properties of a class.
	 */
	Children []DocumentSymbol `json:"children,omitempty"`
}
