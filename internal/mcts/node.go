package mcts

import (
	"math"

	"github.com/todd/earthmover/internal/board"
)

// Node is a node in the MCTS game tree.
// Children are stored as a singly-linked list (Child → Next → Next → ...).
type Node struct {
	Parent, Child, Next *Node
	Count               int
	WinLoseCount        int
	Index               int
	Status              board.GameStatus
}

// NewRoot creates a root node.
func NewRoot() *Node {
	return &Node{Index: -1}
}

// NewChildNode creates a child node and prepends it to this node's child list.
func (n *Node) NewChildNode(index int, parentStatus board.GameStatus) *Node {
	child := &Node{
		Parent: n,
		Index:  index,
		Status: board.GameStatus(-int(parentStatus)),
	}
	// If child is losing, parent is winning
	if child.Status == board.Losing {
		n.Status = board.Winning
	}
	child.Next = n.Child
	n.Child = child
	return child
}

// NewChildFromSource creates a child node by copying data from a source node.
func (n *Node) NewChildFromSource(source *Node) *Node {
	child := &Node{
		Parent:       n,
		Index:        source.Index,
		Status:       source.Status,
		Count:        source.Count,
		WinLoseCount: source.WinLoseCount,
	}
	child.Next = n.Child
	n.Child = child
	return child
}

// FindChild returns the child with the given index, or nil if not found.
func (n *Node) FindChild(index int) *Node {
	for child := n.Child; child != nil; child = child.Next {
		if child.Index == index {
			return child
		}
	}
	return nil
}

// HasChild returns true if this node has at least one child.
func (n *Node) HasChild() bool { return n.Child != nil }

// WinRate returns the win rate from the parent's perspective.
func (n *Node) WinRate() float64 {
	if n.Count == 0 {
		return 0
	}
	return float64(n.Count+n.WinLoseCount) / float64(n.Count*2)
}

// Update records a simulation result.
func (n *Node) Update(result board.SearchStatus) {
	n.Count++
	switch result {
	case board.Win:
		n.WinLoseCount++
	case board.Lose:
		n.WinLoseCount--
	}
}

// Clear resets this node's statistics.
func (n *Node) Clear() {
	n.Count = 0
	n.WinLoseCount = 0
	n.Status = board.Nothing
}

// Winning returns true if this node is a winning position.
func (n *Node) Winning() bool { return n.Status == board.Winning }

// Losing returns true if this node is a losing position.
func (n *Node) Losing() bool { return n.Status == board.Losing }

// NotWinOrLose returns true if this node is neither winning nor losing.
func (n *Node) NotWinOrLose() bool { return n.Status == board.Nothing }

// UCBValue computes the Upper Confidence Bound value for a child node.
// If child is nil, returns the UCB value for an unexplored node.
func (n *Node) UCBValue(child *Node) float64 {
	if n.Count == 0 {
		return 0
	}
	if child != nil {
		return child.WinRate() + math.Sqrt(0.5*math.Log(float64(n.Count))/float64(1+child.Count))
	}
	return math.Sqrt(0.5 * math.Log(float64(n.Count)))
}

// Selection selects the best child node using UCB + board score.
// Returns the search status and the selected node.
func (n *Node) Selection(brd board.Board) (board.SearchStatus, *Node) {
	if n.Winning() {
		return board.Lose, n // Parent loses
	}
	if n.Losing() {
		return board.Win, n // Parent wins
	}

	maxVal := -1.0
	childWinning := false
	scoreSum := float64(brd.GetScoreSum())

	var checked [board.Length]bool
	var bestNode *Node

	// Evaluate existing children
	for child := n.Child; child != nil; child = child.Next {
		i := child.Index
		checked[i] = true

		if child.Winning() {
			childWinning = true
			continue
		}

		// 100% win rate → select immediately
		if child.WinRate() == 1 {
			brd.Play(i)
			return board.Unknown, child
		}

		val := float64(brd.GetScore(i))/scoreSum + n.UCBValue(child)
		if val > maxVal {
			maxVal = val
			bestNode = child
		}
	}

	// Check unexplored moves
	uncheckIdx := brd.GetHSIFiltered(checked[:])
	if uncheckIdx != -1 {
		val := float64(brd.GetScore(uncheckIdx))/scoreSum + n.UCBValue(nil)
		if val > maxVal {
			status := brd.Play(uncheckIdx)
			child := n.NewChildNode(uncheckIdx, status)
			if status == board.Winning {
				return board.Lose, n // Parent loses
			}
			return board.Leaf, child
		}
	}

	// No point selected
	if maxVal == -1 {
		if childWinning {
			n.Status = board.Losing
			if n.Parent != nil {
				n.Parent.Status = board.Winning
			}
			return board.Win, n // Parent wins
		}
		return board.Tie, n
	}

	brd.Play(bestNode.Index)
	return board.Unknown, bestNode
}

// Merge adds another node's statistics into this one.
func (n *Node) Merge(other *Node) {
	n.Count += other.Count
	n.WinLoseCount += other.WinLoseCount
	if n.NotWinOrLose() && !other.NotWinOrLose() {
		n.Status = other.Status
	}
}

// Minus subtracts another node's statistics from this one.
func (n *Node) Minus(other *Node) {
	n.Count -= other.Count
	n.WinLoseCount -= other.WinLoseCount
}

// DeleteChildren removes all children.
func (n *Node) DeleteChildren() {
	n.Child = nil // GC handles the rest
}

// DeleteChildrenExcept removes all children except the specified one.
func (n *Node) DeleteChildrenExcept(keep *Node) {
	n.Child = keep
	keep.Next = nil
}
