package gomoku

import "github.com/todd/earthmover/internal/board"

const (
	ScoreWin      = 10000000
	ScoreForbidden = -100
)

// Evaluator defines the interface that each rule variant must implement.
type Evaluator interface {
	// Init initializes the evaluator (type tree, opening book, etc.)
	Init()

	// EvaluateType classifies a status array into a ChessType.
	EvaluateType(status []board.StoneStatus) ChessType

	// EvaluateScore computes absolute scores for both colors
	// given the 4-direction chess types at a point.
	// score[0] = black's score, score[1] = white's score.
	EvaluateScore(types [4]ChessType, score *[2]int)

	// CheckWinOrLose checks if a score indicates a win or loss.
	CheckWinOrLose(score int) board.GameStatus
}

// PointScorer provides access to a point's score and absScore for evaluateRelativeScore.
// This avoids a circular dependency with the Point struct.
type PointScorer interface {
	GetAbsScore(color int) int
	SetScore(score int)
}

// OpeningClassifier is an optional function that returns a suggested move index
// from the opening book, or -1 if no match.
type OpeningClassifier func(points []PointScorer) int

// EvaluateRelativeScore computes the relative (filtered) score for all points.
// This is shared logic between freestyle and renju.
// openingClassify may be nil if the opening book is not available.
func EvaluateRelativeScore(points []PointScorer, playNo int, openingClassify OpeningClassifier) {
	length := len(points)

	if playNo == 0 {
		for i := 0; i < length; i++ {
			points[i].SetScore(-1)
		}
		points[length/2].SetScore(1)
		return
	}

	// Try opening book for early moves
	if playNo <= 4 && openingClassify != nil {
		index := openingClassify(points)
		if index != -1 {
			for i := 0; i < length; i++ {
				points[i].SetScore(-1)
			}
			points[index].SetScore(1)
			return
		}
	}

	// Use absolute score filtering
	whoTurn := playNo & 1 // 0=black, 1=white

	// Find highest score for current player
	highestScore := -1
	for i := 0; i < length; i++ {
		if s := points[i].GetAbsScore(whoTurn); s > highestScore {
			highestScore = s
		}
	}

	for i := 0; i < length; i++ {
		score := points[i].GetAbsScore(whoTurn)
		if score*8 <= highestScore || (playNo < 10 && score < 140) {
			points[i].SetScore(-1)
		} else {
			points[i].SetScore(score)
		}
	}
}
