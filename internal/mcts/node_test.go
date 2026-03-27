package mcts

import (
	"math"
	"testing"

	"github.com/todd/earthmover/internal/board"
)

func TestNewRoot(t *testing.T) {
	root := NewRoot()
	if root.Index != -1 {
		t.Errorf("root Index = %d, want -1", root.Index)
	}
	if root.Count != 0 || root.WinLoseCount != 0 {
		t.Error("root should have zero counts")
	}
	if !root.NotWinOrLose() {
		t.Error("root should be NotWinOrLose")
	}
}

func TestNewChildNode(t *testing.T) {
	parent := NewRoot()

	// Parent status = Nothing → child status = Nothing (negated)
	child := parent.NewChildNode(42, board.Nothing)
	if child.Index != 42 {
		t.Errorf("child Index = %d, want 42", child.Index)
	}
	if child.Parent != parent {
		t.Error("child Parent should be parent")
	}
	if parent.Child != child {
		t.Error("parent Child should be the new child")
	}

	// Add another child → prepended
	child2 := parent.NewChildNode(43, board.Nothing)
	if parent.Child != child2 {
		t.Error("parent Child should be the newest child")
	}
	if child2.Next != child {
		t.Error("child2.Next should be child")
	}
}

func TestChildFromWinningParent(t *testing.T) {
	parent := NewRoot()

	// Parent winning → child losing → sets parent to winning
	child := parent.NewChildNode(10, board.Winning)
	if child.Status != board.Losing {
		t.Errorf("child status = %d, want Losing", child.Status)
	}
	if parent.Status != board.Winning {
		t.Errorf("parent status = %d, want Winning", parent.Status)
	}
}

func TestFindChild(t *testing.T) {
	parent := NewRoot()
	parent.NewChildNode(10, board.Nothing)
	parent.NewChildNode(20, board.Nothing)
	parent.NewChildNode(30, board.Nothing)

	found := parent.FindChild(20)
	if found == nil || found.Index != 20 {
		t.Error("FindChild(20) should return the child with index 20")
	}

	notFound := parent.FindChild(99)
	if notFound != nil {
		t.Error("FindChild(99) should return nil")
	}
}

func TestWinRate(t *testing.T) {
	n := &Node{Count: 10, WinLoseCount: 4}
	// winRate = (10+4) / (10*2) = 14/20 = 0.7
	expected := 0.7
	if math.Abs(n.WinRate()-expected) > 1e-9 {
		t.Errorf("WinRate = %f, want %f", n.WinRate(), expected)
	}

	// All losses: winLoseCount = -10
	n2 := &Node{Count: 10, WinLoseCount: -10}
	// winRate = (10-10) / 20 = 0
	if n2.WinRate() != 0 {
		t.Errorf("all losses WinRate = %f, want 0", n2.WinRate())
	}
}

func TestUpdate(t *testing.T) {
	n := &Node{}

	n.Update(board.Win)
	if n.Count != 1 || n.WinLoseCount != 1 {
		t.Errorf("after Win: Count=%d WLC=%d", n.Count, n.WinLoseCount)
	}

	n.Update(board.Lose)
	if n.Count != 2 || n.WinLoseCount != 0 {
		t.Errorf("after Lose: Count=%d WLC=%d", n.Count, n.WinLoseCount)
	}

	n.Update(board.Tie)
	if n.Count != 3 || n.WinLoseCount != 0 {
		t.Errorf("after Tie: Count=%d WLC=%d", n.Count, n.WinLoseCount)
	}
}

func TestUCBValue(t *testing.T) {
	parent := &Node{Count: 100}
	child := &Node{Count: 10, WinLoseCount: 4}

	ucb := parent.UCBValue(child)
	if ucb <= 0 {
		t.Errorf("UCB value should be positive, got %f", ucb)
	}

	// Unexplored node should have higher UCB than explored child
	ucbUnexplored := parent.UCBValue(nil)
	if ucbUnexplored <= 0 {
		t.Errorf("unexplored UCB should be positive, got %f", ucbUnexplored)
	}
}

func TestMergeAndMinus(t *testing.T) {
	a := &Node{Count: 10, WinLoseCount: 3}
	b := &Node{Count: 5, WinLoseCount: 2}

	a.Merge(b)
	if a.Count != 15 || a.WinLoseCount != 5 {
		t.Errorf("after Merge: Count=%d WLC=%d, want 15, 5", a.Count, a.WinLoseCount)
	}

	a.Minus(b)
	if a.Count != 10 || a.WinLoseCount != 3 {
		t.Errorf("after Minus: Count=%d WLC=%d, want 10, 3", a.Count, a.WinLoseCount)
	}
}

func TestDeleteChildren(t *testing.T) {
	parent := NewRoot()
	parent.NewChildNode(1, board.Nothing)
	parent.NewChildNode(2, board.Nothing)

	parent.DeleteChildren()
	if parent.HasChild() {
		t.Error("after DeleteChildren: should have no children")
	}
}

func TestDeleteChildrenExcept(t *testing.T) {
	parent := NewRoot()
	c1 := parent.NewChildNode(1, board.Nothing)
	parent.NewChildNode(2, board.Nothing)
	parent.NewChildNode(3, board.Nothing)

	parent.DeleteChildrenExcept(c1)
	if parent.Child != c1 {
		t.Error("Child should be c1")
	}
	if c1.Next != nil {
		t.Error("c1.Next should be nil")
	}
}

func TestClear(t *testing.T) {
	n := &Node{Count: 100, WinLoseCount: 50, Status: board.Winning}
	n.Clear()
	if n.Count != 0 || n.WinLoseCount != 0 || n.Status != board.Nothing {
		t.Error("Clear should reset all stats")
	}
}
