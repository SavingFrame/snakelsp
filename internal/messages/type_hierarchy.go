package messages

type CallHierarchyPrepareParams struct {
	TextDocumentPositionParams
}
type CallHierarchyItem struct {
	Name           string      `json:"name"`
	Kind           SymbolKind  `json:"kind"`
	URI            DocumentUri `json:"uri"`
	Range          *Range      `json:"range"`
	SelectionRange *Range      `json:"selectionRange"`
	Data           any         `json:"data,omitempty"`
}

type TypeHierarchyItem struct {
	Name           string      `json:"name"`
	Kind           SymbolKind  `json:"kind"`
	Tags           []SymbolTag `json:"tags,omitempty"`
	Detail         string      `json:"detail,omitempty"`
	URI            DocumentUri `json:"uri"`
	Range          *Range      `json:"range"`
	SelectionRange *Range      `json:"selectionRange"`
	Data           any         `json:"data,omitempty"`
}

type TypeHierarchySupertypesParams struct {
	Item TypeHierarchyItem `json:"item"`
}
