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

// EvaluateRelativeScore computes the relative (filtered) score for all points.
// Operates directly on the Points array to avoid interface slice allocations.
// openingClassify may be nil if the opening book is not available.
func EvaluateRelativeScore(points *[board.Length]Point, playNo int, openingClassify func(*[board.Length]Point) int) {
	if playNo == 0 {
		for i := 0; i < board.Length; i++ {
			points[i].Scr = -1
		}
		points[board.Length/2].Scr = 1
		return
	}

	// Try opening book for early moves
	if playNo <= 4 && openingClassify != nil {
		index := openingClassify(points)
		if index != -1 {
			for i := 0; i < board.Length; i++ {
				points[i].Scr = -1
			}
			points[index].Scr = 1
			return
		}
	}

	// Use absolute score filtering
	whoTurn := playNo & 1 // 0=black, 1=white

	// Find highest score for current player
	highestScore := -1
	for i := 0; i < board.Length; i++ {
		if s := points[i].AbsScore[whoTurn]; s > highestScore {
			highestScore = s
		}
	}

	for i := 0; i < board.Length; i++ {
		score := points[i].AbsScore[whoTurn]
		if score*8 <= highestScore || (playNo < 10 && score < 140) {
			points[i].Scr = -1
		} else {
			points[i].Scr = score
		}
	}
}
