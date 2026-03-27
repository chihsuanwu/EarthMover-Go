package gomoku

import (
	"testing"

	"github.com/todd/earthmover/internal/board"
)

// stubEvaluator is a minimal evaluator for board-level tests.
// It uses the freestyle type tree via a test helper.
type stubEvaluator struct {
	initCalled bool
}

func (e *stubEvaluator) Init() { e.initCalled = true }
func (e *stubEvaluator) EvaluateType(status []board.StoneStatus) ChessType {
	return ChessType{} // zero type — sufficient for structural tests
}
func (e *stubEvaluator) EvaluateScore(types [4]ChessType, score *[2]int) {
	score[0] = 0
	score[1] = 0
}
func (e *stubEvaluator) CheckWinOrLose(score int) board.GameStatus {
	if score >= ScoreWin {
		return board.Winning
	}
	return board.Nothing
}

func newTestBoard() *GomokuBoard {
	return NewBoard(&stubEvaluator{}, 8)
}

func TestNewBoard(t *testing.T) {
	b := newTestBoard()
	if b.PlayNo != 0 {
		t.Errorf("PlayNo = %d, want 0", b.PlayNo)
	}
	if b.StatusLen != 8 {
		t.Errorf("StatusLen = %d, want 8", b.StatusLen)
	}
	// All points should be empty
	for i := 0; i < board.Length; i++ {
		if b.Points[i].Stat != board.Empty {
			t.Errorf("Points[%d].Stat = %d, want Empty", i, b.Points[i].Stat)
			break
		}
	}
}

func TestNeighborSetup(t *testing.T) {
	b := newTestBoard()

	// Center point (7,7) index 112
	// Direction 0 (→): neighbors at offsets -4..+4 (skip 0)
	// offset -1 → (7,6) = 111, offset +1 → (7,8) = 113
	center := 7*board.Dimen + 7

	// nearest left neighbor (index 3 in DirIdx[0]) = (7,6) = 111
	if idx := b.Points[center].DirIdx[0][3]; idx != 111 {
		t.Errorf("center dir0[3] = %d, want 111", idx)
	}

	// nearest right neighbor (index 4 in DirIdx[0]) = (7,8) = 113
	if idx := b.Points[center].DirIdx[0][4]; idx != 113 {
		t.Errorf("center dir0[4] = %d, want 113", idx)
	}

	// Corner point (0,0) index 0
	// Direction 0 (→): offsets -4..-1 should be out of bounds (-1)
	for i := 0; i < 4; i++ {
		if idx := b.Points[0].DirIdx[0][i]; idx != -1 {
			t.Errorf("corner(0,0) dir0[%d] = %d, want -1 (OOB)", i, idx)
		}
	}
}

func TestPlayAndUndo(t *testing.T) {
	b := newTestBoard()
	center := 7*board.Dimen + 7

	// Play black at center
	status := b.Play(center)
	if status != board.Nothing {
		t.Errorf("Play(center) = %d, want Nothing", status)
	}
	if b.Points[center].Stat != board.Black {
		t.Errorf("after Play: Stat = %d, want Black", b.Points[center].Stat)
	}
	if b.PlayNo != 1 {
		t.Errorf("after Play: PlayNo = %d, want 1", b.PlayNo)
	}
	if !b.WhoTurn() {
		t.Error("after Play: WhoTurn should be true (white's turn)")
	}

	// Undo
	b.Undo(center)
	if b.Points[center].Stat != board.Empty {
		t.Errorf("after Undo: Stat = %d, want Empty", b.Points[center].Stat)
	}
	if b.PlayNo != 0 {
		t.Errorf("after Undo: PlayNo = %d, want 0", b.PlayNo)
	}
}

func TestPlayAlternatesColors(t *testing.T) {
	b := newTestBoard()

	b.Play(112) // black
	if b.Points[112].Stat != board.Black {
		t.Errorf("first move: Stat = %d, want Black", b.Points[112].Stat)
	}

	b.Play(113) // white
	if b.Points[113].Stat != board.White {
		t.Errorf("second move: Stat = %d, want White", b.Points[113].Stat)
	}
}

func TestPass(t *testing.T) {
	b := newTestBoard()
	idx := b.Pass()
	if idx != board.Length {
		t.Errorf("Pass() = %d, want %d", idx, board.Length)
	}
	if b.PlayNo != 1 {
		t.Errorf("after Pass: PlayNo = %d, want 1", b.PlayNo)
	}
}

func TestClone(t *testing.T) {
	b := newTestBoard()
	b.Play(112)

	clone := b.Clone()
	gb := clone.(*GomokuBoard)

	if gb.PlayNo != b.PlayNo {
		t.Errorf("clone PlayNo = %d, want %d", gb.PlayNo, b.PlayNo)
	}
	if gb.Points[112].Stat != board.Black {
		t.Errorf("clone Points[112].Stat = %d, want Black", gb.Points[112].Stat)
	}

	// Mutating clone should not affect original
	clone.Play(113)
	if b.PlayNo != 1 {
		t.Errorf("original PlayNo changed to %d after clone mutation", b.PlayNo)
	}
	if b.Points[113].Stat != board.Empty {
		t.Error("original Points[113] changed after clone mutation")
	}
}

func TestGetScoreSum(t *testing.T) {
	b := newTestBoard()
	sum := b.GetScoreSum()
	// With stub evaluator, only center gets score 1 from relative score
	if sum < 0 {
		t.Errorf("GetScoreSum = %d, should be >= 0", sum)
	}
}

func TestGetHSIOnEmptyBoard(t *testing.T) {
	b := newTestBoard()
	// On empty board with stub evaluator, center (112) should be the only
	// point with score > 0 (from evaluateRelativeScore move 0 logic)
	idx := b.GetHSI()
	if idx != board.Length/2 {
		t.Errorf("GetHSI on empty board = %d, want %d (center)", idx, board.Length/2)
	}
}

func TestGetHSIFiltered(t *testing.T) {
	b := newTestBoard()
	ignore := make([]bool, board.Length)
	// Ignore the center point
	ignore[board.Length/2] = true

	idx := b.GetHSIFiltered(ignore)
	// With only center having score 1 and it's ignored, should return -1
	if idx != -1 {
		t.Errorf("GetHSIFiltered(ignore center) = %d, want -1", idx)
	}
}

func TestLength(t *testing.T) {
	b := newTestBoard()
	if b.Length() != 225 {
		t.Errorf("Length() = %d, want 225", b.Length())
	}
}
