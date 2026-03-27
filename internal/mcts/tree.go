package mcts

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/todd/earthmover/internal/board"
	"github.com/todd/earthmover/internal/gomoku"
)

// TreeJSON is the JSON representation of a game tree node for the frontend.
type TreeJSON struct {
	Index      int         `json:"i"`
	TotalCount int         `json:"tc"`
	WinRate    float64     `json:"wr"`
	WinOrLose  int         `json:"wol"`
	WhoTurn    bool        `json:"wt"`
	Children   []*TreeJSON `json:"ch"`
}

const maxSimulationDepth = 50

// GameTree is the MCTS game tree.
type GameTree struct {
	Root         *Node
	CurrentNode  *Node
	CurrentBoard board.Board
}

// NewGameTree creates a new game tree.
func NewGameTree() *GameTree {
	return &GameTree{}
}

// Init initializes the tree with a board.
func (t *GameTree) Init(brd board.Board) {
	t.Root = NewRoot()
	t.CurrentNode = t.Root
	t.CurrentBoard = brd.Create()
}

// MCTS runs Monte-Carlo Tree Search for a fixed number of cycles.
// Returns false if the search ends prematurely (current node is win/lose).
func (t *GameTree) MCTS(cycles int) bool {
	for i := 0; i < cycles; i++ {
		if !t.CurrentNode.NotWinOrLose() {
			return false
		}

		clone := t.CurrentBoard.Clone()
		status, node := t.selection(clone)

		if status == board.Leaf {
			status = t.simulation(clone)
		}

		t.backProp(node, status)
		gomoku.Release(clone)
	}
	return true
}

// MCTSBatch runs MCTS in batches until the best move has at least minCount simulations.
func (t *GameTree) MCTSBatch(batch, minCount int) {
	for {
		if !t.MCTS(batch) {
			return
		}

		mostTimes := 0
		for child := t.CurrentNode.Child; child != nil; child = child.Next {
			if child.Count > mostTimes {
				mostTimes = child.Count
			}
		}

		// mostTimes == 0 means no useful point (can pass)
		if mostTimes >= minCount || mostTimes == 0 {
			return
		}
	}
}

// MCTSWithContext runs MCTS until the context is cancelled.
func (t *GameTree) MCTSWithContext(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if !t.MCTS(1000) {
			return
		}
	}
}

// MCTSMulti runs parallel MCTS with multiple goroutines.
// threadCount includes the main goroutine.
func (t *GameTree) MCTSMulti(threadCount, batch, minCount int) {
	if threadCount <= 1 {
		t.MCTSBatch(batch, minCount)
		return
	}

	extraCount := threadCount - 1

	// Copy trees for extra goroutines
	trees := make([]*GameTree, extraCount)
	for i := range trees {
		trees[i] = &GameTree{}
		trees[i].copyFrom(t)
	}

	// Snapshot the original tree before parallel search
	originTree := &GameTree{}
	originTree.copyFrom(t)

	// Launch extra goroutines
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	for i := 0; i < extraCount; i++ {
		wg.Add(1)
		go func(tree *GameTree) {
			defer wg.Done()
			tree.MCTSWithContext(ctx)
		}(trees[i])
	}

	// Main goroutine thinking
	t.MCTSBatch(batch, minCount)

	// Stop extra goroutines
	cancel()
	wg.Wait()

	// Merge results: tree_delta = tree_copy - origin; main += tree_delta
	for i := 0; i < extraCount; i++ {
		minusTree(trees[i], originTree)
		t.mergeTree(trees[i])
	}
}

// MCTSMultiWithContext runs parallel MCTS until the context is cancelled.
func (t *GameTree) MCTSMultiWithContext(threadCount int, ctx context.Context) {
	if threadCount <= 1 {
		t.MCTSWithContext(ctx)
		return
	}

	extraCount := threadCount - 1

	trees := make([]*GameTree, extraCount)
	for i := range trees {
		trees[i] = &GameTree{}
		trees[i].copyFrom(t)
	}

	originTree := &GameTree{}
	originTree.copyFrom(t)

	var wg sync.WaitGroup
	for i := 0; i < extraCount; i++ {
		wg.Add(1)
		go func(tree *GameTree) {
			defer wg.Done()
			tree.MCTSWithContext(ctx)
		}(trees[i])
	}

	t.MCTSWithContext(ctx)
	wg.Wait()

	for i := 0; i < extraCount; i++ {
		minusTree(trees[i], originTree)
		t.mergeTree(trees[i])
	}
}

// MCTSResult returns the index of the best move.
func (t *GameTree) MCTSResult() int {
	var index int

	if t.CurrentNode.Winning() {
		// Select the losing child (opponent loses = we win)
		for child := t.CurrentNode.Child; child != nil; child = child.Next {
			if child.Losing() {
				index = child.Index
			}
		}
	} else {
		// Select child with most playouts, break ties by score
		mostTimes := -1
		score := 0

		for child := t.CurrentNode.Child; child != nil; child = child.Next {
			if child.Count > mostTimes {
				index = child.Index
				mostTimes = child.Count
				score = t.CurrentBoard.GetScore(index)
			} else if child.Count == mostTimes {
				childScore := t.CurrentBoard.GetScore(child.Index)
				if childScore > score {
					index = child.Index
					mostTimes = child.Count
					score = childScore
				}
			}
		}

		if mostTimes == -1 {
			return -1 // pass
		}
	}

	child := t.CurrentNode.FindChild(index)
	fmt.Printf("total sim: %d  best: %c%d  sim: %d  WinR: %.3f  W/L: %d\n",
		t.CurrentNode.Count,
		rune('A'+index%15), index/15+1,
		child.Count,
		child.WinRate(),
		int(t.CurrentNode.Status))

	return index
}

// Play records a real move on the board and advances the tree.
func (t *GameTree) Play(index int) board.GameStatus {
	status := t.CurrentBoard.Play(index)

	child := t.CurrentNode.FindChild(index)
	if child == nil {
		child = t.CurrentNode.NewChildNode(index, status)
	}

	t.CurrentNode.DeleteChildrenExcept(child)
	t.CurrentNode.Clear()
	t.CurrentNode = child

	return status
}

// Pass skips the current player's turn.
func (t *GameTree) Pass() {
	index := t.CurrentBoard.Pass()
	if index == -1 {
		return
	}

	child := t.CurrentNode.FindChild(index)
	if child == nil {
		child = t.CurrentNode.NewChildNode(index, board.Nothing)
	}

	t.CurrentNode.DeleteChildrenExcept(child)
	t.CurrentNode.Clear()
	t.CurrentNode = child
}

// Undo reverts the last move.
func (t *GameTree) Undo() {
	t.CurrentBoard.Undo(t.CurrentNode.Index)
	t.CurrentNode = t.CurrentNode.Parent
	t.CurrentNode.Clear()
	t.CurrentNode.DeleteChildren()
}

// GetTreeJSON returns the game tree as a JSON-serializable structure for the frontend.
func (t *GameTree) GetTreeJSON() *TreeJSON {
	if t.CurrentNode == nil || t.CurrentNode.Parent == nil {
		return nil
	}
	threshold := t.CurrentNode.Count / 1000
	return t.getSubTreeJSON(t.CurrentNode, t.CurrentBoard.WhoTurn(), threshold)
}

func (t *GameTree) getSubTreeJSON(node *Node, whoTurn bool, threshold int) *TreeJSON {
	wr := node.WinRate()
	if !whoTurn {
		wr = 1 - wr
	}
	wol := int(node.Status)
	if !whoTurn {
		wol = -wol
	}

	tj := &TreeJSON{
		Index:      node.Index,
		TotalCount: node.Count,
		WinRate:    math.Round(wr*100) / 100,
		WinOrLose:  wol,
		WhoTurn:    whoTurn,
		Children:   []*TreeJSON{},
	}

	for child := node.Child; child != nil; child = child.Next {
		if child.Count >= 8 && child.Count >= threshold {
			tj.Children = append(tj.Children, t.getSubTreeJSON(child, !whoTurn, threshold))
		}
	}

	return tj
}

// --- Internal MCTS methods ---

func (t *GameTree) selection(brd board.Board) (board.SearchStatus, *Node) {
	node := t.CurrentNode
	for {
		status, selected := node.Selection(brd)
		if status != board.Unknown {
			return status, selected
		}
		node = selected
	}
}

func (t *GameTree) simulation(brd board.Board) board.SearchStatus {
	for d := 0; d < maxSimulationDepth; d++ {
		index := brd.GetHSI()
		if index == -1 {
			return board.Tie
		}
		if brd.Play(index) != board.Nothing {
			if d&1 == 0 {
				return board.Lose
			}
			return board.Win
		}
	}
	return board.Tie
}

func (t *GameTree) backProp(node *Node, result board.SearchStatus) {
	for node != t.CurrentNode {
		node.Update(result)
		node = node.Parent
		result = result.Reverse()
	}
	node.Update(result)
}

// --- Tree copy/merge ---

func (t *GameTree) copyFrom(source *GameTree) {
	t.Root = &Node{
		Parent:       source.CurrentNode.Parent,
		Index:        source.CurrentNode.Index,
		Status:       source.CurrentNode.Status,
		Count:        source.CurrentNode.Count,
		WinLoseCount: source.CurrentNode.WinLoseCount,
	}
	t.CurrentNode = t.Root
	t.CurrentBoard = source.CurrentBoard.Clone()
	copyAllChildren(source.CurrentNode, t.CurrentNode)
}

func copyAllChildren(src, dest *Node) {
	for child := src.Child; child != nil; child = child.Next {
		destChild := dest.NewChildFromSource(child)
		if child.HasChild() {
			copyAllChildren(child, destChild)
		}
	}
}

func minusTree(beMinusTree, minusSource *GameTree) {
	beMinusTree.Root.Minus(minusSource.Root)
	minusAllChildren(beMinusTree.Root, minusSource.Root)
}

func minusAllChildren(beMinusNode, minusNode *Node) {
	for minusChild := minusNode.Child; minusChild != nil; minusChild = minusChild.Next {
		beMinusChild := beMinusNode.FindChild(minusChild.Index)
		if beMinusChild == nil {
			continue
		}
		beMinusChild.Minus(minusChild)
		if minusChild.HasChild() {
			minusAllChildren(beMinusChild, minusChild)
		}
	}
}

func (t *GameTree) mergeTree(other *GameTree) {
	t.CurrentNode.Merge(other.Root)
	mergeAllChildren(t.CurrentNode, other.Root)
}

func mergeAllChildren(mergingNode, mergedNode *Node) {
	for mergedChild := mergedNode.Child; mergedChild != nil; mergedChild = mergedChild.Next {
		mergingChild := mergingNode.FindChild(mergedChild.Index)
		if mergingChild != nil {
			mergingChild.Merge(mergedChild)
		} else {
			mergingChild = mergingNode.NewChildFromSource(mergedChild)
		}
		if mergedChild.HasChild() {
			mergeAllChildren(mergingChild, mergedChild)
		}
	}
}
