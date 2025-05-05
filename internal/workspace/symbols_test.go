package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"snakelsp/internal/messages"
)

func TestParseSymbols(t *testing.T) {
	// Mock Python file content
	pythonCode := `
class MyClass:
    def method_one(self, param1):
        pass

    @decorator
    def decorated_method(self, param2):
        pass

    def method_three(self, param3):
        pass

def standalone_function(param2):
    return param2

@decorator
def decorated_function(param3):
	pass

@decorator
@decorator2
def multiple_decoration_function(param3):
	pass
`

	// Create a mock PythonFile
	mockFile := &PythonFile{
		Text: pythonCode,
		Url:  "mock_file.py",
	}

	// Parse symbols
	symbols, err := mockFile.ParseSymbols()
	assert.NoError(t, err)

	// Verify parsed symbols
	assert.Len(t, symbols, 4) // Expecting one class and one function

	// Verify class symbol
	classSymbol := symbols[0]
	assert.Equal(t, "MyClass", classSymbol.Name)
	assert.Equal(t, messages.SymbolKindClass, classSymbol.Kind)

	assert.Equal(t, 3, len(classSymbol.Children))

	// Verify standalone function symbol
	funcSymbol := symbols[1]
	assert.Equal(t, "standalone_function", funcSymbol.Name)
	assert.Equal(t, messages.SymbolKindFunction, funcSymbol.Kind)
}
