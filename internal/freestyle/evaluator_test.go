package freestyle

import (
	"testing"

	"github.com/todd/earthmover/internal/gomoku"
)

func TestFreestyleCheckWinOrLose(t *testing.T) {
	e := &FreestyleEvaluator{}

	tests := []struct {
		score int
		want  int // -1, 0, 1
	}{
		{0, 0},
		{100, 0},
		{gomoku.ScoreWin, 1},
		{gomoku.ScoreWin + 1, 1},
		{-100, 0}, // freestyle has no forbidden
	}

	for _, tt := range tests {
		got := int(e.CheckWinOrLose(tt.score))
		if got != tt.want {
			t.Errorf("CheckWinOrLose(%d) = %d, want %d", tt.score, got, tt.want)
		}
	}
}

func TestFreestyleEvaluateScoreAllZero(t *testing.T) {
	e := &FreestyleEvaluator{}
	types := [4]gomoku.ChessType{}
	var score [2]int
	e.EvaluateScore(types, &score)
	if score[0] != 0 || score[1] != 0 {
		t.Errorf("all zero types: score = [%d, %d], want [0, 0]", score[0], score[1])
	}
}

func TestFreestyleEvaluateScoreFive(t *testing.T) {
	e := &FreestyleEvaluator{}
	// One direction has five-in-a-row for black
	types := [4]gomoku.ChessType{
		gomoku.NewChessType(gomoku.SingleType{Length: 5, Life: 0, Level: 0}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{}),
	}
	var score [2]int
	e.EvaluateScore(types, &score)
	if score[0] < gomoku.ScoreWin {
		t.Errorf("five-in-row black score = %d, want >= %d", score[0], gomoku.ScoreWin)
	}
}

func TestFreestyleEvaluateScoreDouble4(t *testing.T) {
	e := &FreestyleEvaluator{}
	// Two dead fours for black → double-4 bonus
	types := [4]gomoku.ChessType{
		gomoku.NewChessType(gomoku.SingleType{Length: 4, Life: 0, Level: 0}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{Length: 4, Life: 0, Level: 0}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{}),
	}
	var score [2]int
	e.EvaluateScore(types, &score)
	// Base score for 2 dead fours = 310*2 = 620, plus double-4 bonus = 9000
	expected := 310*2 + 9000
	if score[0] != expected {
		t.Errorf("double-4 black score = %d, want %d", score[0], expected)
	}
}

func TestFreestyleEvaluateScoreDead4Live3(t *testing.T) {
	e := &FreestyleEvaluator{}
	// One dead four + one live three for black
	types := [4]gomoku.ChessType{
		gomoku.NewChessType(gomoku.SingleType{Length: 4, Life: 0, Level: 0}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{Length: 3, Life: 1, Level: 0}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{}),
	}
	var score [2]int
	e.EvaluateScore(types, &score)
	// Base: 310 + 265 = 575, plus d4l3 bonus = 2400
	expected := 310 + 265 + 2400
	if score[0] != expected {
		t.Errorf("d4l3 black score = %d, want %d", score[0], expected)
	}
}

func TestFreestyleEvaluateScoreDefense(t *testing.T) {
	e := &FreestyleEvaluator{}
	// White has a dead four → white gets attack score, black gets defense score
	types := [4]gomoku.ChessType{
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{Length: 4, Life: 0, Level: 0}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{}),
		gomoku.NewChessType(gomoku.SingleType{}, gomoku.SingleType{}),
	}
	var score [2]int
	e.EvaluateScore(types, &score)
	if score[1] != 310 {
		t.Errorf("white attack score = %d, want 310", score[1])
	}
	if score[0] != 190 {
		t.Errorf("black defense score = %d, want 190", score[0])
	}
}
