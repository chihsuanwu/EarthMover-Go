package ai

import (
	"testing"
	"time"

	"github.com/todd/earthmover/internal/board"
)

func TestNewAndReset(t *testing.T) {
	a := New()
	a.Reset(0, board.RuleFreestyle)

	if a.WhoTurn() {
		t.Error("first move should be black (WhoTurn=false)")
	}
}

func TestThink(t *testing.T) {
	a := New()
	a.Reset(0, board.RuleFreestyle)

	idx := a.Think()
	if idx < 0 || idx >= board.Length {
		t.Errorf("Think() = %d, want valid index", idx)
	}
}

func TestPlayAndWin(t *testing.T) {
	a := New()
	a.Reset(0, board.RuleFreestyle)

	// Play a game where black wins with 5 in a row
	moves := []int{
		7*15 + 3, // black
		0*15 + 0, // white
		7*15 + 4,
		0*15 + 1,
		7*15 + 5,
		0*15 + 2,
		7*15 + 6,
		0*15 + 3,
		7*15 + 7, // black 5th in a row
	}

	for i, m := range moves {
		winner := a.Play(m)
		if i < len(moves)-1 {
			if winner != -1 {
				t.Fatalf("move %d: unexpected winner %d", i, winner)
			}
		} else {
			if winner != 0 {
				t.Errorf("last move: winner = %d, want 0 (black)", winner)
			}
		}
	}
}

func TestUndo(t *testing.T) {
	a := New()
	a.Reset(0, board.RuleFreestyle)

	a.Play(112) // black
	a.Play(113) // white
	a.Undo(2)

	if a.WhoTurn() {
		t.Error("after undo 2 moves, should be black's turn")
	}
}

func TestThinkInBackground(t *testing.T) {
	a := New()
	a.Reset(0, board.RuleFreestyle)

	a.ThinkInBackground()
	time.Sleep(100 * time.Millisecond)
	a.StopBackground()

	// Should have accumulated some simulations
	if a.tree.CurrentNode.Count == 0 {
		t.Error("background thinking should have produced simulations")
	}
}

func TestIsAlive(t *testing.T) {
	a := New()
	if !a.IsAlive() {
		t.Error("newly created AI should be alive")
	}

	a.lastAliveTime = time.Now().Add(-80 * time.Second)
	if a.IsAlive() {
		t.Error("AI inactive for 80s should not be alive (limit=70s)")
	}

	a.RenewLiveTime()
	if !a.IsAlive() {
		t.Error("after RenewLiveTime, should be alive")
	}
}

func TestResetRenju(t *testing.T) {
	a := New()
	a.Reset(1, board.RuleRenjuBasic)

	idx := a.Think()
	if idx < 0 || idx >= board.Length {
		t.Errorf("Renju Think() = %d, want valid index", idx)
	}
}
