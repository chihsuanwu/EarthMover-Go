package gomoku

import "testing"

// mockPoint implements PointScorer for testing evaluateRelativeScore.
type mockPoint struct {
	absScore [2]int
	score    int
}

func (p *mockPoint) GetAbsScore(color int) int { return p.absScore[color] }
func (p *mockPoint) SetScore(score int)        { p.score = score }

func TestEvaluateRelativeScoreMove0(t *testing.T) {
	// Move 0: only center gets score 1, all others -1
	points := make([]PointScorer, 225)
	for i := range points {
		points[i] = &mockPoint{absScore: [2]int{100, 100}}
	}

	EvaluateRelativeScore(points, 0, nil)

	center := points[112].(*mockPoint)
	if center.score != 1 {
		t.Errorf("center score = %d, want 1", center.score)
	}

	corner := points[0].(*mockPoint)
	if corner.score != -1 {
		t.Errorf("corner score = %d, want -1", corner.score)
	}
}

func TestEvaluateRelativeScoreFiltering(t *testing.T) {
	// playNo >= 5: filter by absolute score
	// Highest score for black (whoTurn=1 for playNo=5, actually playNo&1=1=white)
	// Let's use playNo=10 (even, so whoTurn=0=black)
	points := make([]PointScorer, 225)
	for i := range points {
		points[i] = &mockPoint{absScore: [2]int{10, 0}} // low score
	}
	// Set a few high-score points
	points[0].(*mockPoint).absScore[0] = 800
	points[1].(*mockPoint).absScore[0] = 200
	points[2].(*mockPoint).absScore[0] = 50

	EvaluateRelativeScore(points, 10, nil)

	// Highest = 800. Threshold: score*8 > 800 → score > 100
	p0 := points[0].(*mockPoint)
	if p0.score != 800 {
		t.Errorf("high score point: score = %d, want 800", p0.score)
	}

	p1 := points[1].(*mockPoint)
	if p1.score != 200 {
		t.Errorf("medium score point: score = %d, want 200", p1.score)
	}

	p2 := points[2].(*mockPoint)
	if p2.score != -1 {
		t.Errorf("low score point: score = %d, want -1", p2.score)
	}
}

func TestEvaluateRelativeScoreEarlyGameThreshold(t *testing.T) {
	// playNo < 10: additional filter score < 140
	points := make([]PointScorer, 225)
	for i := range points {
		points[i] = &mockPoint{absScore: [2]int{130, 0}} // passes ratio test but < 140
	}
	points[0].(*mockPoint).absScore[0] = 1000

	EvaluateRelativeScore(points, 6, nil) // even playNo → black's turn

	// score=130: 130*8=1040 > 1000 ✓ but 130 < 140 in early game → filtered
	p := points[1].(*mockPoint)
	if p.score != -1 {
		t.Errorf("early game low score: score = %d, want -1", p.score)
	}
}

func TestEvaluateRelativeScoreWithOpening(t *testing.T) {
	// With opening classifier that returns index 5
	points := make([]PointScorer, 225)
	for i := range points {
		points[i] = &mockPoint{absScore: [2]int{100, 100}}
	}

	classifier := func(pts []PointScorer) int { return 5 }

	EvaluateRelativeScore(points, 2, classifier)

	p5 := points[5].(*mockPoint)
	if p5.score != 1 {
		t.Errorf("opening suggested point: score = %d, want 1", p5.score)
	}

	p0 := points[0].(*mockPoint)
	if p0.score != -1 {
		t.Errorf("non-opening point: score = %d, want -1", p0.score)
	}
}
