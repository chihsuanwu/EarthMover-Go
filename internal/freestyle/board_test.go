package freestyle

import (
	"testing"

	"github.com/todd/earthmover/internal/board"
	"github.com/todd/earthmover/internal/gomoku"
)

func TestNewFreestyleBoard(t *testing.T) {
	b := NewBoard()
	if b.Length() != 225 {
		t.Errorf("Length() = %d, want 225", b.Length())
	}
	if b.WhoTurn() {
		t.Error("first move should be black (WhoTurn=false)")
	}
}

func TestFreestyleBoardPlayAndScores(t *testing.T) {
	b := NewBoard()

	// First move: center should have score > 0
	center := 7*board.Dimen + 7
	if s := b.GetScore(center); s <= 0 {
		t.Errorf("center score = %d, want > 0", s)
	}

	// Play center
	status := b.Play(center)
	if status != board.Nothing {
		t.Errorf("Play(center) = %d, want Nothing", status)
	}

	// Score sum should be positive after first move
	sum := b.GetScoreSum()
	if sum <= 0 {
		t.Errorf("GetScoreSum after first move = %d, want > 0", sum)
	}
}

func TestFreestyleBoardGetHSI(t *testing.T) {
	b := NewBoard()

	// On empty board, GetHSI should return center
	idx := b.GetHSI()
	if idx != board.Length/2 {
		t.Errorf("GetHSI on empty board = %d, want %d", idx, board.Length/2)
	}
}

func TestFreestyleBoardPlayUndoRoundtrip(t *testing.T) {
	b := NewBoard()
	gb := b.(*gomoku.GomokuBoard)

	// Record initial state
	center := 7*board.Dimen + 7
	origScore := gb.Points[center].AbsScore

	// Play and undo
	b.Play(center)
	b.Undo(center)

	// Verify restored
	if gb.Points[center].Stat != board.Empty {
		t.Errorf("after undo: Stat = %d, want Empty", gb.Points[center].Stat)
	}
	if gb.PlayNo != 0 {
		t.Errorf("after undo: PlayNo = %d, want 0", gb.PlayNo)
	}
	if gb.Points[center].AbsScore != origScore {
		t.Errorf("after undo: AbsScore = %v, want %v", gb.Points[center].AbsScore, origScore)
	}
}

func TestFreestyleBoardDetectsWin(t *testing.T) {
	b := NewBoard()

	// Play 5 black stones in a row (row 7, cols 3-7)
	// Black plays at col 3, 4, 5, 6; white plays elsewhere
	// Then black plays col 7 for the win
	moves := [][2]int{
		{7, 3}, // black
		{0, 0}, // white
		{7, 4}, // black
		{0, 1}, // white
		{7, 5}, // black
		{0, 2}, // white
		{7, 6}, // black
		{0, 3}, // white
	}

	for _, m := range moves {
		status := b.Play(m[0]*board.Dimen + m[1])
		if status != board.Nothing {
			t.Fatalf("unexpected status %d at move (%d,%d)", status, m[0], m[1])
		}
	}

	// Black plays 5th in a row
	status := b.Play(7*board.Dimen + 7)
	if status != board.Winning {
		t.Errorf("five in a row: status = %d, want Winning", status)
	}
}

func TestFreestyleBoardClone(t *testing.T) {
	b := NewBoard()
	b.Play(112) // black at center

	clone := b.Clone()

	// Clone should have same state
	if clone.GetScore(112) != b.GetScore(112) {
		t.Error("clone score doesn't match original")
	}

	// Mutating clone shouldn't affect original
	clone.Play(113)
	gb := b.(*gomoku.GomokuBoard)
	if gb.Points[113].Stat != board.Empty {
		t.Error("original changed after clone mutation")
	}
}
