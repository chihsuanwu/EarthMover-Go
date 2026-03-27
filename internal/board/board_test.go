package board

import "testing"

func TestStoneStatusOpponent(t *testing.T) {
	if Black.Opponent() != White {
		t.Errorf("Black.Opponent() = %d, want %d", Black.Opponent(), White)
	}
	if White.Opponent() != Black {
		t.Errorf("White.Opponent() = %d, want %d", White.Opponent(), Black)
	}
}

func TestStoneStatusValues(t *testing.T) {
	// These values must match the C++ enum exactly, as they are used
	// as array indices throughout the codebase.
	if Black != 0 {
		t.Errorf("Black = %d, want 0", Black)
	}
	if White != 1 {
		t.Errorf("White = %d, want 1", White)
	}
	if Empty != 2 {
		t.Errorf("Empty = %d, want 2", Empty)
	}
	if Bound != 3 {
		t.Errorf("Bound = %d, want 3", Bound)
	}
}

func TestGameStatusValues(t *testing.T) {
	if Losing != -1 {
		t.Errorf("Losing = %d, want -1", Losing)
	}
	if Nothing != 0 {
		t.Errorf("Nothing = %d, want 0", Nothing)
	}
	if Winning != 1 {
		t.Errorf("Winning = %d, want 1", Winning)
	}
}

func TestSearchStatusReverse(t *testing.T) {
	tests := []struct {
		input SearchStatus
		want  SearchStatus
	}{
		{Win, Lose},
		{Lose, Win},
		{Tie, Tie},
	}
	for _, tt := range tests {
		got := tt.input.Reverse()
		if got != tt.want {
			t.Errorf("SearchStatus(%d).Reverse() = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestBoardConstants(t *testing.T) {
	if Dimen != 15 {
		t.Errorf("Dimen = %d, want 15", Dimen)
	}
	if Length != 225 {
		t.Errorf("Length = %d, want 225", Length)
	}
}
