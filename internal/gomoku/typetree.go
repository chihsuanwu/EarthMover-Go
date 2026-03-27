package gomoku

import "github.com/todd/earthmover/internal/board"

// TypeNode is a node in the pattern recognition trie.
// The trie classifies a status array (neighboring stone colors along one direction)
// into a ChessType.
type TypeNode struct {
	// Children indexed by StoneStatus: 0=Black, 1=White, 2=Empty, 3=Bound
	Children [4]*TypeNode
	Type     ChessType
	Jump     bool
	Leaf     bool
}

// Classify walks the type tree to classify a status array into a ChessType.
// statusLen is the length of the status array (8 for freestyle, 10 for renju).
func Classify(root *TypeNode, status []board.StoneStatus, statusLen int) ChessType {
	node := root
	start := statusLen/2 - 1
	for move := -1; ; {
		for cp := start; ; cp += move {
			node = node.Children[status[cp]]
			if node.Leaf {
				return node.Type
			}
			if node.Jump {
				break
			}
		}
		// After scanning left side, jump to right side
		move = 1
		start = statusLen / 2
	}
}

// CutSameResultChild prunes subtrees where all children yield the same type.
// Returns the type pointer if the subtree can be collapsed, nil otherwise.
func CutSameResultChild(node *TypeNode) *ChessType {
	if node.Leaf {
		return &node.Type
	}

	var currentType *ChessType
	canCut := true

	for i := 0; i < 4; i++ {
		if node.Children[i] == nil {
			continue
		}
		returnType := CutSameResultChild(node.Children[i])
		if returnType == nil {
			canCut = false
		} else if currentType == nil {
			currentType = returnType
		} else if !currentType.Equal(*returnType) {
			canCut = false
		}
	}

	if !canCut {
		return nil
	}

	// Collapse: set this node's type and mark as leaf, free children
	node.Type = *currentType
	node.Leaf = true
	for i := 0; i < 4; i++ {
		node.Children[i] = nil // GC will collect the subtrees
	}

	return &node.Type
}
