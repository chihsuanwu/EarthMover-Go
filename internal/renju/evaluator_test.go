package renju

import (
	"testing"

	"github.com/todd/earthmover/internal/gomoku"
)

func TestRenjuCheckWinOrLose(t *testing.T) {
	e := &RenjuEvaluator{}

	tests := []struct {
		score int
		want  int
	}{
		{0, 0},
		{100, 0},
		{gomoku.ScoreWin, 1},
		{gomoku.ScoreForbidden, -1}, // renju: forbidden → losing
		{-50, 0},
	}

	for _, tt := range tests {
		got := int(e.CheckWinOrLose(tt.score))
		if got != tt.want {
			t.Errorf("CheckWinOrLose(%d) = %d, want %d", tt.score, got, tt.want)
		}
	}
}

func TestRenjuForbiddenOverline(t *testing.T) {
	e := &RenjuEvaluator{}
	// Black has length=-1 (overline) in one direction → forbidden
	types := [4]gomoku.ChessType{
		gomoku.NewChessType(gomoku.SingleType{Length: -1, Life: 0, Level: 0}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{Length: 1, Life: 1, Level: 0}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{}),
	}
	var score [2]int
	e.EvaluateScore(types, &score)
	if score[0] != gomoku.ScoreForbidden {
		t.Errorf("black score = %d, want %d (forbidden)", score[0], gomoku.ScoreForbidden)
	}
}

func TestRenjuForbiddenDouble4(t *testing.T) {
	e := &RenjuEvaluator{}
	// Black has two dead fours → forbidden (double-4)
	types := [4]gomoku.ChessType{
		gomoku.NewChessType(gomoku.SingleType{Length: 4, Life: 0, Level: 0}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{Length: 4, Life: 0, Level: 0}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{}),
	}
	var score [2]int
	e.EvaluateScore(types, &score)
	if score[0] != gomoku.ScoreForbidden {
		t.Errorf("black score = %d, want %d (forbidden double-4)", score[0], gomoku.ScoreForbidden)
	}
}

func TestRenjuForbiddenDoubleLive3(t *testing.T) {
	e := &RenjuEvaluator{}
	// Black has two live threes → forbidden (double-live-3)
	types := [4]gomoku.ChessType{
		gomoku.NewChessType(gomoku.SingleType{Length: 3, Life: 1, Level: 0}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{Length: 3, Life: 1, Level: 0}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{}),
	}
	var score [2]int
	e.EvaluateScore(types, &score)
	if score[0] != gomoku.ScoreForbidden {
		t.Errorf("black score = %d, want %d (forbidden double-live-3)", score[0], gomoku.ScoreForbidden)
	}
}

func TestRenjuForbiddenOverriddenByWin(t *testing.T) {
	e := &RenjuEvaluator{}
	// Black has five-in-a-row AND an overline in another direction.
	// Win overrides forbidden.
	types := [4]gomoku.ChessType{
		gomoku.NewChessType(gomoku.SingleType{Length: 5, Life: 0, Level: 0}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{Length: -1, Life: 0, Level: 0}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{}),
	}
	var score [2]int
	e.EvaluateScore(types, &score)
	if score[0] == gomoku.ScoreForbidden {
		t.Error("five-in-row + overline: should NOT be forbidden (win overrides)")
	}
	if score[0] < gomoku.ScoreWin {
		t.Errorf("black score = %d, want >= %d (win)", score[0], gomoku.ScoreWin)
	}
}

func TestRenjuWhiteNotForbidden(t *testing.T) {
	e := &RenjuEvaluator{}
	// White has two dead fours — not forbidden (only black has forbidden rules)
	types := [4]gomoku.ChessType{
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{Length: 4, Life: 0, Level: 0}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{Length: 4, Life: 0, Level: 0}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{}),
	}
	var score [2]int
	e.EvaluateScore(types, &score)
	if score[1] == gomoku.ScoreForbidden {
		t.Error("white double-4 should NOT be forbidden")
	}
	// White should get double-4 attack bonus
	expected := 310*2 + 9000
	if score[1] != expected {
		t.Errorf("white score = %d, want %d", score[1], expected)
	}
}
