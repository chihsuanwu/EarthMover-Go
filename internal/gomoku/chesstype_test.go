package gomoku

import "testing"

func TestSingleTypeComparison(t *testing.T) {
	tests := []struct {
		name    string
		a, b    SingleType
		less    bool
		greater bool
		equal   bool
	}{
		{
			name:  "equal types",
			a:     SingleType{3, 1, 0},
			b:     SingleType{3, 1, 0},
			equal: true,
		},
		{
			name: "shorter length is less",
			a:    SingleType{2, 1, 0},
			b:    SingleType{3, 1, 0},
			less: true,
		},
		{
			name:    "longer length is greater",
			a:       SingleType{4, 0, 0},
			b:       SingleType{3, 1, 0},
			greater: true,
		},
		{
			name: "same length, dead < live",
			a:    SingleType{3, 0, 0},
			b:    SingleType{3, 1, 0},
			less: true,
		},
		{
			name:    "same length, live > dead",
			a:       SingleType{3, 1, 0},
			b:       SingleType{3, 0, 0},
			greater: true,
		},
		{
			name:  "level does not affect ordering",
			a:     SingleType{3, 1, 2},
			b:     SingleType{3, 1, 0},
			equal: true, // Equal in ordering (Less and Greater both false)
		},
		{
			name: "zero type less than any positive",
			a:    SingleType{0, 0, 0},
			b:    SingleType{1, 0, 0},
			less: true,
		},
		{
			name:    "five is greatest",
			a:       SingleType{5, 0, 0},
			b:       SingleType{4, 1, 0},
			greater: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.equal {
				// For "equal" in ordering, Less and Greater should both be false.
				// But Equal checks all three fields including Level.
				if tt.a.Level == tt.b.Level {
					if !tt.a.Equal(tt.b) {
						t.Error("expected Equal")
					}
				}
				if tt.a.Less(tt.b) {
					t.Error("expected not Less")
				}
				if tt.a.Greater(tt.b) {
					t.Error("expected not Greater")
				}
				if !tt.a.LessOrEqual(tt.b) {
					t.Error("expected LessOrEqual")
				}
				if !tt.a.GreaterOrEqual(tt.b) {
					t.Error("expected GreaterOrEqual")
				}
			}
			if tt.less {
				if !tt.a.Less(tt.b) {
					t.Error("expected Less")
				}
				if tt.a.Greater(tt.b) {
					t.Error("expected not Greater")
				}
				if !tt.a.LessOrEqual(tt.b) {
					t.Error("expected LessOrEqual")
				}
				if tt.a.GreaterOrEqual(tt.b) {
					t.Error("expected not GreaterOrEqual")
				}
			}
			if tt.greater {
				if !tt.a.Greater(tt.b) {
					t.Error("expected Greater")
				}
				if tt.a.Less(tt.b) {
					t.Error("expected not Less")
				}
				if !tt.a.GreaterOrEqual(tt.b) {
					t.Error("expected GreaterOrEqual")
				}
				if tt.a.LessOrEqual(tt.b) {
					t.Error("expected not LessOrEqual")
				}
			}
		})
	}
}

func TestSingleTypeEqual(t *testing.T) {
	a := SingleType{3, 1, 2}
	b := SingleType{3, 1, 2}
	c := SingleType{3, 1, 0}

	if !a.Equal(b) {
		t.Error("identical types should be equal")
	}
	if a.Equal(c) {
		t.Error("different level should not be equal")
	}
}

func TestChessType(t *testing.T) {
	black := SingleType{3, 1, 0}
	white := SingleType{2, 0, 0}

	ct := NewChessType(black, white)

	if ct.Length(0) != 3 {
		t.Errorf("black length = %d, want 3", ct.Length(0))
	}
	if ct.Life(0) != 1 {
		t.Errorf("black life = %d, want 1", ct.Life(0))
	}
	if ct.Length(1) != 2 {
		t.Errorf("white length = %d, want 2", ct.Length(1))
	}
	if ct.Life(1) != 0 {
		t.Errorf("white life = %d, want 0", ct.Life(1))
	}
}

func TestChessTypeEqual(t *testing.T) {
	a := NewChessType(SingleType{3, 1, 0}, SingleType{2, 0, 1})
	b := NewChessType(SingleType{3, 1, 0}, SingleType{2, 0, 1})
	c := NewChessType(SingleType{3, 1, 0}, SingleType{2, 1, 0})

	if !a.Equal(b) {
		t.Error("identical ChessTypes should be equal")
	}
	if a.Equal(c) {
		t.Error("different ChessTypes should not be equal")
	}
}
