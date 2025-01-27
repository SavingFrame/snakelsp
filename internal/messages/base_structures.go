package messages

import "encoding/json"

type (
	URI         = string
	Integer     = int32
	DocumentUri = string
	UInteger    = uint32
)

type IntegerOrString struct {
	Value any // Integer | string
}

// ([json.Marshaler] interface)
func (t *IntegerOrString) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Value)
}

// ([json.Unmarshaler] interface)
func (t *IntegerOrString) UnmarshalJSON(data []byte) error {
	var value Integer
	if err := json.Unmarshal(data, &value); err == nil {
		t.Value = value
		return nil
	} else {
		var value string
		if err := json.Unmarshal(data, &value); err == nil {
			t.Value = value
			return nil
		} else {
			return err
		}
	}
}

type Position struct {
	/**
	 * Line position in a document (zero-based).
	 */
	Line UInteger `json:"line"`

	/**
	 * Character offset on a line in a document (zero-based). Assuming that
	 * the line is represented as a string, the `character` value represents
	 * the gap between the `character` and `character + 1`.
	 *
	 * If the character value is greater than the line length it defaults back
	 * to the line length.
	 */
	Character UInteger `json:"character"`
}
type Range struct {
	/**
	 * The range's start position.
	 */
	Start Position `json:"start"`

	/**
	 * The range's end position.
	 */
	End Position `json:"end"`
}
type ProgressToken = IntegerOrString

type WorkDoneProgressParams struct {
	/**
	 * An optional token that a server can use to report work done progress.
	 */
	WorkDoneToken *ProgressToken `json:"workDoneToken,omitempty"`
}
type TraceValue string

const (
	TraceValueOff     = TraceValue("off")
	TraceValueMessage = TraceValue("message") // The spec clearly says "message", but some implementations use "messages" instead
	TraceValueVerbose = TraceValue("verbose")
)

// https://microsoft.github.io/language-server-protocol/specifications/specification-3-16#textDocumentPositionParams

type TextDocumentPositionParams struct {
	/**
	 * The text document.
	 */
	TextDocument TextDocumentIdentifier `json:"textDocument"`

	/**
	 * The position inside the text document.
	 */
	Position Position `json:"position"`
}
type PartialResultParams struct {
	/**
	 * An optional token that a server can use to report partial results (e.g.
	 * streaming) to the client.
	 */
	PartialResultToken *ProgressToken `json:"partialResultToken,omitempty"`
}

// https://microsoft.github.io/language-server-protocol/specifications/specification-3-16#location

type Location struct {
	URI   DocumentUri `json:"uri"`
	Range Range       `json:"range"`
}
