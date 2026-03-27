package mcts

import (
	"testing"

	"github.com/todd/earthmover/internal/board"
	"github.com/todd/earthmover/internal/freestyle"
)

func newTestTree() *GameTree {
	brd := freestyle.NewBoard()
	tree := NewGameTree()
	tree.Init(brd)
	return tree
}

func TestInit(t *testing.T) {
	tree := newTestTree()
	if tree.Root == nil {
		t.Fatal("Root should not be nil")
	}
	if tree.CurrentNode != tree.Root {
		t.Error("CurrentNode should be Root")
	}
	if tree.CurrentBoard == nil {
		t.Fatal("CurrentBoard should not be nil")
	}
}

func TestMCTSSingleCycle(t *testing.T) {
	tree := newTestTree()

	ok := tree.MCTS(1)
	if !ok {
		t.Error("MCTS(1) should return true (not prematurely ended)")
	}
	if tree.CurrentNode.Count == 0 {
		t.Error("after 1 cycle, currentNode should have count > 0")
	}
}

func TestMCTSMultipleCycles(t *testing.T) {
	tree := newTestTree()

	tree.MCTS(100)
	if tree.CurrentNode.Count < 100 {
		t.Errorf("after 100 cycles, count = %d, want >= 100", tree.CurrentNode.Count)
	}
	if !tree.CurrentNode.HasChild() {
		t.Error("after 100 cycles, should have children")
	}
}

func TestMCTSBatch(t *testing.T) {
	tree := newTestTree()
	tree.MCTSBatch(100, 50)

	// Best child should have at least minCount simulations
	maxCount := 0
	for child := tree.CurrentNode.Child; child != nil; child = child.Next {
		if child.Count > maxCount {
			maxCount = child.Count
		}
	}
	if maxCount < 50 {
		t.Errorf("best child count = %d, want >= 50", maxCount)
	}
}

func TestMCTSResult(t *testing.T) {
	tree := newTestTree()
	tree.MCTSBatch(200, 50)

	idx := tree.MCTSResult()
	if idx < 0 || idx >= board.Length {
		t.Errorf("MCTSResult = %d, want valid index", idx)
	}
}

func TestPlay(t *testing.T) {
	tree := newTestTree()

	// Play center
	center := board.Length / 2
	status := tree.Play(center)
	if status != board.Nothing {
		t.Errorf("Play(center) = %d, want Nothing", status)
	}
	if tree.CurrentNode.Index != center {
		t.Errorf("CurrentNode.Index = %d, want %d", tree.CurrentNode.Index, center)
	}
}

func TestPlayAndUndo(t *testing.T) {
	tree := newTestTree()

	center := board.Length / 2
	tree.Play(center)
	tree.Undo()

	if tree.CurrentNode != tree.Root {
		t.Error("after Undo, CurrentNode should be Root")
	}
}

func TestMCTSMulti(t *testing.T) {
	tree := newTestTree()
	tree.MCTSMulti(2, 100, 50)

	// After multi-threaded search, should have results
	if tree.CurrentNode.Count < 50 {
		t.Errorf("after MCTSMulti, count = %d, want >= 50", tree.CurrentNode.Count)
	}

	idx := tree.MCTSResult()
	if idx < 0 || idx >= board.Length {
		t.Errorf("MCTSResult = %d, want valid index", idx)
	}
}

func TestCopyAndMerge(t *testing.T) {
	tree := newTestTree()
	tree.MCTS(50)

	// Copy
	treeCopy := &GameTree{}
	treeCopy.copyFrom(tree)

	if treeCopy.CurrentNode.Count != tree.CurrentNode.Count {
		t.Errorf("copy count = %d, want %d", treeCopy.CurrentNode.Count, tree.CurrentNode.Count)
	}

	// Run more search on copy
	treeCopy.MCTS(50)

	// Snapshot for minus
	origin := &GameTree{}
	origin.copyFrom(tree)

	// Merge delta
	minusTree(treeCopy, origin)
	tree.mergeTree(treeCopy)

	// Tree should have more simulations now
	if tree.CurrentNode.Count < 50 {
		t.Errorf("after merge, count = %d, want >= 50", tree.CurrentNode.Count)
	}
}
