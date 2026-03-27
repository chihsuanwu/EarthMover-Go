package renju

import (
	"testing"

	"github.com/todd/earthmover/internal/board"
)

func init() {
	InitTypeTree()
}

// parseStatus builds a status array from a string.
// For renju (statusLength=10):
//   - index 0: farthest left neighbor
//   - index 4: nearest left neighbor
//   - index 5: nearest right neighbor
//   - index 9: farthest right neighbor
func parseStatus(s string) []board.StoneStatus {
	result := make([]board.StoneStatus, len(s))
	for i, c := range s {
		switch c {
		case 'B':
			result[i] = board.Black
		case 'W':
			result[i] = board.White
		case '.':
			result[i] = board.Empty
		case '#':
			result[i] = board.Bound
		}
	}
	return result
}

func TestClassifyKnownPatterns(t *testing.T) {
	tests := []struct {
		name      string
		status    string // 10 chars for renju
		blackLen  int8
		blackLife int8
	}{
		// All empty → live 1
		{"all empty", "..........", 1, 1},

		// One black neighbor (nearest left)
		{"one black left", "....B.....", 2, 1},

		// Two blacks left, open
		{"two black left", "...BB.....", 3, 1},

		// Three blacks left, open → live 4
		{"three black left open", "..BBB.....", 4, 1},

		// Four blacks left → five
		{"four black left", ".BBBB.....", 5, 0},

		// Three blacks left, blocked by bound → dead 4
		{"three black left bound", "##BBB.....", 4, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := parseStatus(tt.status)
			ct := ClassifyType(status)
			if ct.Length(0) != tt.blackLen {
				t.Errorf("black length = %d, want %d", ct.Length(0), tt.blackLen)
			}
			if ct.Life(0) != tt.blackLife {
				t.Errorf("black life = %d, want %d", ct.Life(0), tt.blackLife)
			}
		})
	}
}

func TestClassifyOverlineBlack(t *testing.T) {
	// 4 black on each side → connect = 9 > 5 → overline → forbidden for black
	// The renju DFS allows at most 4 consecutive same-color per side.
	status := parseStatus(".BBBBBBBB.")
	ct := ClassifyType(status)
	if ct.Length(0) >= 5 {
		t.Errorf("black overline: length = %d, want < 5 (forbidden)", ct.Length(0))
	}
}

func TestClassifyOverlineWhite(t *testing.T) {
	// Overline for white is still counted as 5 (no forbidden rule)
	status := parseStatus(".WWWWWWWW.")
	ct := ClassifyType(status)
	if ct.Length(1) != 5 {
		t.Errorf("white overline: length = %d, want 5", ct.Length(1))
	}
}

func TestClassifySymmetry(t *testing.T) {
	statusL := parseStatus("...BB.....")
	statusR := parseStatus(".....BB...")
	ctL := ClassifyType(statusL)
	ctR := ClassifyType(statusR)
	if ctL.Length(0) != ctR.Length(0) || ctL.Life(0) != ctR.Life(0) {
		t.Errorf("symmetry broken: left=(%d,%d) right=(%d,%d)",
			ctL.Length(0), ctL.Life(0), ctR.Length(0), ctR.Life(0))
	}
}

func TestClassifyNoPanic(t *testing.T) {
	// Verify the tree handles various valid patterns without panicking
	patterns := []string{
		"BBB..BBB..",
		"BB...BB...",
		"B....B....",
		"..........",
		".WWWW.....",
		"#BBBB.....",
		"WBBBB.....",
		"..BBWW....",
		"####......",
	}
	for _, p := range patterns {
		t.Run(p, func(t *testing.T) {
			status := parseStatus(p)
			ct := ClassifyType(status)
			_ = ct
		})
	}
}

func TestTypeTreeInitIdempotent(t *testing.T) {
	InitTypeTree()
	InitTypeTree()
	status := parseStatus("..........")
	_ = ClassifyType(status)
}
