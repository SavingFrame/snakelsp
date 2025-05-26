package protocol

import (
	"encoding/json"
	"log/slog"

	"snakelsp/internal/messages"
	"snakelsp/internal/request"
	"snakelsp/internal/workspace"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

func findDefinition(astRoot *tree_sitter.Node, symbol string, cursorNode *tree_sitter.Node, sourceCode []byte, logger *slog.Logger) *tree_sitter.Node {
	// TODO: Rework to go recursive through the tree from cursorNode to parent, scan all in that parent, and then go up to parent's parent and so on.
	logger.Debug("Starting definition search", slog.String("symbol", symbol))

	// Step 1: Search parent nodes upward.
	current := cursorNode
	for current != nil {
		currentKind := current.Kind() // Get the type of the current node

		switch currentKind {
		case "function_definition", "class_definition":
			nameNode := current.ChildByFieldName("name")
			if nameNode != nil {
				nameText := string(sourceCode[nameNode.StartByte():nameNode.EndByte()])
				if nameText == symbol {
					return nameNode
				}
			}
		case "assignment":
			leftNode := current.Child(0)

			if leftNode != nil && leftNode.Kind() == "identifier" {
				nameText := string(sourceCode[leftNode.StartByte():leftNode.EndByte()])
				if nameText == symbol {
					return leftNode
				}
			}
		}
		current = current.Parent()
	}

	logger.Debug("No definition found in parent nodes; falling back to BFS")

	// Step 2: If not found, perform a BFS from the AST root.
	var queue []*tree_sitter.Node
	queue = append(queue, astRoot)

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		nodeKind := node.Kind()
		switch nodeKind {
		case "function_definition", "class_definition":
			nameNode := node.ChildByFieldName("name")
			if nameNode != nil {
				nameText := string(sourceCode[nameNode.StartByte():nameNode.EndByte()])
				if nameText == symbol {
					logger.Debug("Definition found during BFS", slog.String("name", nameText))
					return nameNode
				}
			}
		case "assignment":
			leftNode := node.Child(0)

			if leftNode != nil && leftNode.Kind() == "identifier" {
				nameText := string(sourceCode[leftNode.StartByte():leftNode.EndByte()])
				if nameText == symbol {
					return leftNode
				}
			}

		}

		// Add all named children to queue
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(uint(i))
			if child != nil {
				queue = append(queue, child)
			}
		}
	}

	logger.Debug("Definition not found anywhere")
	return nil
}

func extractTextFromNode(p *workspace.PythonFile, node *tree_sitter.Node) string {
	startByte := node.StartByte()
	endByte := node.EndByte()

	return string(p.Text[startByte:endByte])
}

func HandleGotoDefinition(r *request.Request) (interface{}, error) {
	var data messages.DefinitionParams
	err := json.Unmarshal(r.Params, &data)
	if err != nil {
		r.Logger.Error("Unmarshalling error: %v", slog.Any("error", err))
		return nil, err
	}
	pythonFile, err := workspace.GetPythonFile(data.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	astRoot := pythonFile.GetOrCreateAst()
	foundedNode := astRoot.NamedDescendantForPointRange(
		tree_sitter.Point{Row: uint(data.Position.Line), Column: uint(data.Position.Character)},
		tree_sitter.Point{Row: uint(data.Position.Line), Column: uint(data.Position.Character)},
	)
	nodeText := extractTextFromNode(pythonFile, foundedNode)
	definitionNode := findDefinition(astRoot, nodeText, foundedNode, []byte(pythonFile.Text), r.Logger)
	if definitionNode == nil {
		return nil, nil
	}
	r.Logger.Debug("Definition found", slog.String("name", definitionNode.ToSexp()))
	definitionRange := &messages.Range{
		Start: messages.Position{
			Line:      uint32(definitionNode.StartPosition().Row),
			Character: uint32(definitionNode.StartPosition().Column),
		},
		End: messages.Position{
			Line:      uint32(definitionNode.EndPosition().Row),
			Character: uint32(definitionNode.EndPosition().Column),
		},
	}
	return &messages.LocationLink{
		TargetURI:            data.TextDocument.URI,
		TargetRange:          *definitionRange,
		TargetSelectionRange: *definitionRange,
	}, nil
}
