package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
	// "snakelsp/internal/messages"
)

func TestParseImports(t *testing.T) {
	// Mock Python file content
	pythonCode := `
from tmp2 import ClassA as ClassB
import pandas
import pandas as pd
from pandas import (
	DataFrame, 
	Series as se
	)

`

	// Create a mock PythonFile
	mockFile := &PythonFile{
		Text: pythonCode,
		Url:  "mock_file.py",
	}

	// Parse imports
	imports, err := mockFile.ParseImports()
	assert.NoError(t, err)

	assert.Len(t, imports, 5)
}
