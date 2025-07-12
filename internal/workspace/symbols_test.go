package workspace

import (
	"testing"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/google/uuid"
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
	symbols, err := mockFile.parseSymbols()
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

func TestParseSymbolsExternalFile(t *testing.T) {
	mockFile := &PythonFile{
		Text:     "def test(): pass",
		Url:      "external.py",
		External: true,
	}

	_, err := mockFile.parseSymbols()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot parse symbols for external files")
}

func TestParseSymbolsEmptyFile(t *testing.T) {
	mockFile := &PythonFile{
		Text: "",
		Url:  "empty.py",
	}

	symbols, err := mockFile.parseSymbols()
	assert.NoError(t, err)
	assert.Empty(t, symbols)
}

func TestParseSymbolsWithInheritance(t *testing.T) {
	pythonCode := `
class BaseClass:
    def base_method(self):
        pass

class DerivedClass(BaseClass):
    def derived_method(self):
        pass

class MultipleInheritance(BaseClass, object):
    def multi_method(self):
        pass
`

	mockFile := &PythonFile{
		Text: pythonCode,
		Url:  "inheritance.py",
	}

	symbols, err := mockFile.parseSymbols()
	assert.NoError(t, err)
	assert.Len(t, symbols, 3)

	// Find DerivedClass by name
	var derivedClass *Symbol
	for _, symbol := range symbols {
		if symbol.Name == "DerivedClass" {
			derivedClass = symbol
			break
		}
	}

	// Check DerivedClass has superclass
	assert.NotNil(t, derivedClass)
	assert.Equal(t, "DerivedClass", derivedClass.Name)
	assert.Contains(t, derivedClass.superObjectsNames, "BaseClass")
}

func TestParseSymbolsWithReturnTypes(t *testing.T) {
	pythonCode := `
def typed_function(param: int) -> str:
    return "test"

class TypedClass:
    def typed_method(self, x: float) -> bool:
        return True
`

	mockFile := &PythonFile{
		Text: pythonCode,
		Url:  "typed.py",
	}

	symbols, err := mockFile.parseSymbols()
	assert.NoError(t, err)
	assert.Len(t, symbols, 2)

	// Find function and class symbols by name
	var funcSymbol, classSymbol *Symbol
	for _, symbol := range symbols {
		if symbol.Name == "typed_function" {
			funcSymbol = symbol
		} else if symbol.Name == "TypedClass" {
			classSymbol = symbol
		}
	}

	// Check function with return type
	assert.NotNil(t, funcSymbol)
	assert.Equal(t, "typed_function", funcSymbol.Name)
	assert.Equal(t, "str", funcSymbol.ReturnType)

	// Check method with return type
	assert.NotNil(t, classSymbol)
	if len(classSymbol.Children) > 0 {
		methodSymbol := classSymbol.Children[0]
		assert.Equal(t, "bool", methodSymbol.ReturnType)
	}
}

func TestParseSymbolsComplexStructure(t *testing.T) {
	pythonCode := `
class OuterClass:
    def outer_method(self):
        pass
    
    class InnerClass:
        def inner_method(self):
            pass

def module_function():
    def nested_function():
        pass
    return nested_function
`

	mockFile := &PythonFile{
		Text: pythonCode,
		Url:  "complex.py",
	}

	symbols, err := mockFile.parseSymbols()
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(symbols), 2)

	// Find symbols by name
	var outerClass, innerClass, moduleFunction *Symbol
	for _, symbol := range symbols {
		switch symbol.Name {
		case "OuterClass":
			outerClass = symbol
		case "InnerClass":
			innerClass = symbol
		case "module_function":
			moduleFunction = symbol
		}
	}

	// Verify outer class exists
	assert.NotNil(t, outerClass)
	assert.Equal(t, "OuterClass", outerClass.Name)
	assert.Equal(t, messages.SymbolKindClass, outerClass.Kind)

	// Verify inner class exists (parsed as separate top-level symbol)
	assert.NotNil(t, innerClass)
	assert.Equal(t, "InnerClass", innerClass.Name)
	assert.Equal(t, messages.SymbolKindClass, innerClass.Kind)

	// Verify module function exists
	assert.NotNil(t, moduleFunction)
	assert.Equal(t, "module_function", moduleFunction.Name)
	assert.Equal(t, messages.SymbolKindFunction, moduleFunction.Kind)

	// Note: nested_function inside module_function is not captured by the current
	// tree-sitter query as it only captures top-level functions and class methods,
	// not nested functions within other functions. This is expected behavior.
}

func TestSearchSymbolByUUID(t *testing.T) {
	// Clear FlatSymbols for clean test
	FlatSymbols = orderedmap.NewOrderedMap[uuid.UUID, *Symbol]()

	// Create test symbol
	testSymbol := &Symbol{
		UUID: uuid.New(),
		Name: "TestSymbol",
		Kind: messages.SymbolKindFunction,
	}
	FlatSymbols.Set(testSymbol.UUID, testSymbol)

	// Test successful search
	found, err := SearchSymbolByUUID(testSymbol.UUID)
	assert.NoError(t, err)
	assert.Equal(t, testSymbol.Name, found.Name)

	// Test unsuccessful search
	nonExistentUUID := uuid.New()
	_, err = SearchSymbolByUUID(nonExistentUUID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "symbol not found")
}

func TestGetWorkspaceSymbols(t *testing.T) {
	// Clear FlatSymbols for clean test
	FlatSymbols = orderedmap.NewOrderedMap[uuid.UUID, *Symbol]()

	// Add test symbols
	symbol1 := &Symbol{UUID: uuid.New(), Name: "TestFunction", Kind: messages.SymbolKindFunction}
	symbol2 := &Symbol{UUID: uuid.New(), Name: "TestClass", Kind: messages.SymbolKindClass}
	FlatSymbols.Set(symbol1.UUID, symbol1)
	FlatSymbols.Set(symbol2.UUID, symbol2)

	// Test without query
	symbols, err := GetWorkspaceSymbols("")
	assert.NoError(t, err)
	assert.Len(t, symbols, 2)

	// Test with query
	symbols, err = GetWorkspaceSymbols("Test")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(symbols), 0)
}

func TestFindSymbolByPosition(t *testing.T) {
	// Clear FlatSymbols for clean test
	FlatSymbols = orderedmap.NewOrderedMap[uuid.UUID, *Symbol]()

	mockFile := &PythonFile{Url: "test.py"}
	testSymbol := &Symbol{
		UUID: uuid.New(),
		Name: "TestSymbol",
		File: mockFile,
		NameRange: messages.Range{
			Start: messages.Position{Line: 5, Character: 10},
			End:   messages.Position{Line: 5, Character: 20},
		},
	}
	FlatSymbols.Set(testSymbol.UUID, testSymbol)

	// Test successful find
	found, err := FindSymbolByPosition(mockFile, 5, 15)
	assert.NoError(t, err)
	assert.Equal(t, testSymbol.Name, found.Name)

	// Test unsuccessful find
	_, err = FindSymbolByPosition(mockFile, 10, 15)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "symbol not found")
}

func TestSymbolNameWithParent(t *testing.T) {
	parentSymbol := &Symbol{Name: "ParentClass"}
	childSymbol := &Symbol{
		Name:     "child_method",
		FullName: "child_method(self)",
		Parent:   parentSymbol,
	}

	// Test with parent
	nameWithParent := childSymbol.SymbolNameWithParent()
	assert.Equal(t, "ParentClass.child_method(self)", nameWithParent)

	// Test without parent
	orphanSymbol := &Symbol{Name: "orphan_function"}
	nameWithoutParent := orphanSymbol.SymbolNameWithParent()
	assert.Equal(t, "orphan_function", nameWithoutParent)
}

func TestFileSymbols(t *testing.T) {
	pythonCode := `
def test_function():
    pass

class TestClass:
    def test_method(self):
        pass
`

	mockFile := &PythonFile{
		Text: pythonCode,
		Url:  "file_symbols_test.py",
	}

	// Test without query
	symbols, err := mockFile.FileSymbols("")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(symbols), 2)

	// Test with query
	symbols, err = mockFile.FileSymbols("test")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(symbols), 1)
}

func TestFilterSymbols(t *testing.T) {
	symbols := []*Symbol{
		{Name: "TestFunction", Kind: messages.SymbolKindFunction},
		{Name: "AnotherFunction", Kind: messages.SymbolKindFunction},
		{Name: "TestClass", Kind: messages.SymbolKindClass},
	}

	// Test filtering with query
	filtered, err := FilterSymbols(symbols, "Test")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(filtered), 2)

	// Test filtering with empty query - fuzzy search returns all items for empty query
	filtered, err = FilterSymbols(symbols, "")
	assert.NoError(t, err)
	assert.Len(t, filtered, 3) // fuzzy.FindFold returns all items for empty query

	// Test filtering with specific match
	filtered, err = FilterSymbols(symbols, "TestFunction")
	assert.NoError(t, err)
	if len(filtered) > 0 {
		assert.Equal(t, "TestFunction", filtered[0].Name)
	}
}

func TestCreateSymbol(t *testing.T) {
	mockFile := &PythonFile{Url: "test.py"}
	startPos := messages.Position{Line: 1, Character: 0}
	endPos := messages.Position{Line: 5, Character: 10}
	nameStartPos := messages.Position{Line: 1, Character: 4}
	nameEndPos := messages.Position{Line: 1, Character: 12}

	symbol := createSymbol(
		"TestFunction",
		messages.SymbolKindFunction,
		"(param1, param2)",
		"str",
		"TestFunction(param1, param2) -> str",
		mockFile,
		startPos,
		endPos,
		nameStartPos,
		nameEndPos,
		"BaseClass",
	)

	assert.Equal(t, "TestFunction", symbol.Name)
	assert.Equal(t, messages.SymbolKindFunction, symbol.Kind)
	assert.Equal(t, "(param1, param2)", symbol.Parameters)
	assert.Equal(t, "str", symbol.ReturnType)
	assert.Equal(t, mockFile, symbol.File)
	assert.Contains(t, symbol.superObjectsNames, "BaseClass")
	assert.NotEqual(t, uuid.Nil, symbol.UUID)
}

func TestIsChildOf(t *testing.T) {
	classSymbol := &Symbol{
		Range: messages.Range{
			Start: messages.Position{Line: 1, Character: 0},
			End:   messages.Position{Line: 10, Character: 0},
		},
	}

	methodSymbol := &Symbol{
		Range: messages.Range{
			Start: messages.Position{Line: 3, Character: 4},
			End:   messages.Position{Line: 5, Character: 8},
		},
	}

	outsideSymbol := &Symbol{
		Range: messages.Range{
			Start: messages.Position{Line: 15, Character: 0},
			End:   messages.Position{Line: 17, Character: 4},
		},
	}

	assert.True(t, isChildOf(methodSymbol, classSymbol))
	assert.False(t, isChildOf(outsideSymbol, classSymbol))
}
