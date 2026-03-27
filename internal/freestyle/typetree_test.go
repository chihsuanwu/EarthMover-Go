package freestyle

import (
	"testing"

	"github.com/todd/earthmover/internal/board"
)

func init() {
	InitTypeTree()
}

// parseStatus builds a status array from a string.
// Characters: 'B'=Black, 'W'=White, '.'=Empty, '#'=Bound
//
// The status array represents neighbors around the analyze point (which is NOT in the array).
// For freestyle (statusLength=8):
//   - index 0: farthest left neighbor
//   - index 3: nearest left neighbor
//   - index 4: nearest right neighbor
//   - index 7: farthest right neighbor
//
// Classify scans left (3→2→1→0) then right (4→5→6→7).
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
	// The typeAnalyze function works by:
	// 1. Counting the "center group" (CG): consecutive same-color stones adjacent to center
	// 2. Setting center = color, then recursively analyzing boundary points
	// 3. The returned length reflects "what pattern would this point complete"
	//
	// Key insight: if center has N same-color neighbors on one side,
	// connect = N+1 (including center). The recursive analysis then determines
	// if playing here would create a 4, 5, or shorter pattern.

	tests := []struct {
		name      string
		status    string // 8 chars for freestyle
		blackLen  int8
		blackLife int8
	}{
		// All empty: center point with open surroundings → live 1
		{"all empty", "........", 1, 1},

		// One black neighbor (nearest left), rest empty
		// Board: . . . B [*] . . . .
		// connect=2, recursive → live 2
		{"one black left", "...B....", 2, 1},

		// One black each side
		// Board: . . . B [*] B . . .
		{"one black each side", "...BB...", 3, 1},

		// Two blacks left, open
		// Board: . . B B [*] . . . .
		// connect=3
		{"two black left open", "..BB....", 3, 1},

		// Three blacks left, open
		// Board: . B B B [*] . . . .
		// connect=4, both sides reach 5 → live 4
		{"three black left open", ".BBB....", 4, 1},

		// Three blacks left, blocked by white
		// Board: W B B B [*] . . . .
		// connect=4, left blocked → dead 4
		{"three black left blocked", "WBBB....", 4, 0},

		// Three blacks left, blocked by bound
		// Board: # B B B [*] . . . .
		{"three black left bound", "#BBB....", 4, 0},

		// Four blacks (both sides of center) → five in a row
		// Board: . . . B [*] B B B .
		// But wait, that's connect=1+1+3=5 → five
		{"four black both sides", "...BBBB.", 5, 0},

		// Full five: 4 left + right
		{"five across", "BBBBBBB.", 5, 0},
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

func TestClassifySymmetry(t *testing.T) {
	// Mirrored patterns should give the same classification.
	// Left: ..BB....  (2 blacks on left)
	// Right: ....BB.. (2 blacks on right)
	statusL := parseStatus("..BB....")
	statusR := parseStatus("....BB..")
	ctL := ClassifyType(statusL)
	ctR := ClassifyType(statusR)
	if ctL.Length(0) != ctR.Length(0) || ctL.Life(0) != ctR.Life(0) {
		t.Errorf("symmetry broken: left=(%d,%d) right=(%d,%d)",
			ctL.Length(0), ctL.Life(0), ctR.Length(0), ctR.Life(0))
	}
}

func TestClassifyBothColors(t *testing.T) {
	// Mixed colors: blacks on left, whites on right
	// Board: . . B B [*] W W . .
	// status = [..BB WW..]
	status := parseStatus("..BBWW..")
	ct := ClassifyType(status)
	// Both colors should have some pattern
	if ct.Length(0) == 0 && ct.Length(1) == 0 {
		t.Error("expected at least one color to have nonzero length")
	}
}

func TestTypeTreeInitIdempotent(t *testing.T) {
	InitTypeTree()
	InitTypeTree()
	status := parseStatus("........")
	ct := ClassifyType(status)
	_ = ct
}
