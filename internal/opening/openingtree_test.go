package opening

import (
	"testing"

	"github.com/todd/earthmover/internal/board"
)

func init() {
	Init()
}

// mockPoint implements PointStatus for testing.
type mockPoint struct {
	status board.StoneStatus
}

func (p *mockPoint) Status() board.StoneStatus { return p.status }

func emptyBoard() []PointStatus {
	points := make([]PointStatus, board.Length)
	for i := range points {
		points[i] = &mockPoint{status: board.Empty}
	}
	return points
}

func setStone(points []PointStatus, row, col int, color board.StoneStatus) {
	points[row*board.Dimen+col] = &mockPoint{status: color}
}

func TestInitIdempotent(t *testing.T) {
	Init()
	Init()
	if root == nil {
		t.Fatal("root should not be nil after init")
	}
}

func TestClassifyEmptyBoard(t *testing.T) {
	// Empty board should not match any opening
	points := emptyBoard()
	result := Classify(points)
	if result != -1 {
		t.Errorf("empty board: classify = %d, want -1", result)
	}
}

func TestClassifyAfterFirstMove(t *testing.T) {
	// After black plays center (H8 = row 7, col 7), opening 2.1 should match.
	// Opening 2.1:
	//   -----
	//   -PP--
	//   --X--
	//   -----
	//   -----
	// X is at [2][2] in the 5x5 grid. With 8 rotations, one of them should
	// match black at center.
	points := emptyBoard()
	setStone(points, 7, 7, board.Black) // center

	result := Classify(points)
	// Should suggest a move (not -1) since opening 2.1 covers the first move
	if result == -1 {
		t.Error("after first move at center: expected opening suggestion, got -1")
	}
	// The suggested move should be a valid board index
	if result < 0 || result >= board.Length {
		t.Errorf("invalid result index: %d", result)
	}
}

func TestClassifyPatternTooLarge(t *testing.T) {
	// Place stones that span more than 5 rows/cols → should return -1
	points := emptyBoard()
	setStone(points, 0, 0, board.Black)
	setStone(points, 5, 5, board.White) // 6 rows/cols span

	result := Classify(points)
	if result != -1 {
		t.Errorf("too large pattern: classify = %d, want -1", result)
	}
}

func TestRotate(t *testing.T) {
	// Verify 4 rotations return to original
	original := [5][5]byte{
		{'X', '-', '-', '-', '-'},
		{'-', 'O', '-', '-', '-'},
		{'-', '-', '-', '-', '-'},
		{'-', '-', '-', '-', '-'},
		{'-', '-', '-', '-', '-'},
	}

	table := original
	for i := 0; i < 4; i++ {
		table = rotate(table)
	}
	if table != original {
		t.Error("4 rotations should return to original")
	}
}

func TestMirror(t *testing.T) {
	// Verify mirror is its own inverse (transposing twice returns to original)
	original := [5][5]byte{
		{'X', '-', '-', '-', '-'},
		{'-', 'O', '-', '-', '-'},
		{'-', '-', 'P', '-', '-'},
		{'-', '-', '-', '-', '-'},
		{'-', '-', '-', '-', '-'},
	}

	table := mirror(mirror(original))
	if table != original {
		t.Error("double mirror should return to original")
	}
}

func TestClassifySuggestsWithinBounds(t *testing.T) {
	// Run classify multiple times to check that results are always within bounds
	points := emptyBoard()
	setStone(points, 7, 7, board.Black) // center

	for i := 0; i < 100; i++ {
		result := Classify(points)
		if result != -1 && (result < 0 || result >= board.Length) {
			t.Fatalf("iteration %d: result %d out of bounds", i, result)
		}
	}
}
