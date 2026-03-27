package gomoku

import (
	"testing"

	"github.com/todd/earthmover/internal/board"
)

func makeTestPoints(absScore [2]int) [board.Length]Point {
	var points [board.Length]Point
	for i := range points {
		points[i].AbsScore = absScore
	}
	return points
}

func TestEvaluateRelativeScoreMove0(t *testing.T) {
	points := makeTestPoints([2]int{100, 100})

	EvaluateRelativeScore(&points, 0, nil)

	if points[112].Scr != 1 {
		t.Errorf("center score = %d, want 1", points[112].Scr)
	}
	if points[0].Scr != -1 {
		t.Errorf("corner score = %d, want -1", points[0].Scr)
	}
}

func TestEvaluateRelativeScoreFiltering(t *testing.T) {
	points := makeTestPoints([2]int{10, 0})
	points[0].AbsScore[0] = 800
	points[1].AbsScore[0] = 200
	points[2].AbsScore[0] = 50

	EvaluateRelativeScore(&points, 10, nil)

	if points[0].Scr != 800 {
		t.Errorf("high score point: score = %d, want 800", points[0].Scr)
	}
	if points[1].Scr != 200 {
		t.Errorf("medium score point: score = %d, want 200", points[1].Scr)
	}
	if points[2].Scr != -1 {
		t.Errorf("low score point: score = %d, want -1", points[2].Scr)
	}
}

func TestEvaluateRelativeScoreEarlyGameThreshold(t *testing.T) {
	points := makeTestPoints([2]int{130, 0})
	points[0].AbsScore[0] = 1000

	EvaluateRelativeScore(&points, 6, nil)

	if points[1].Scr != -1 {
		t.Errorf("early game low score: score = %d, want -1", points[1].Scr)
	}
}

func TestEvaluateRelativeScoreWithOpening(t *testing.T) {
	points := makeTestPoints([2]int{100, 100})

	classifier := func(pts *[board.Length]Point) int { return 5 }

	EvaluateRelativeScore(&points, 2, classifier)

	if points[5].Scr != 1 {
		t.Errorf("opening suggested point: score = %d, want 1", points[5].Scr)
	}
	if points[0].Scr != -1 {
		t.Errorf("non-opening point: score = %d, want -1", points[0].Scr)
	}
}
