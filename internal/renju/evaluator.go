package renju

import (
	"github.com/todd/earthmover/internal/board"
	"github.com/todd/earthmover/internal/gomoku"
)

// scoreTable is indexed by [length][life][level][attackOrDefense].
// Same values as freestyle — the difference is in the combo/forbidden logic.
var scoreTable = [6][2][4][2]int{
	// length 0
	{{{0, 0}, {0, 0}, {0, 0}, {0, 0}},
		{{0, 0}, {0, 0}, {0, 0}, {0, 0}}},
	// length 1
	{{{2, 1}, {0, 0}, {0, 0}, {0, 0}},
		{{6, 2}, {9, 5}, {15, 12}, {20, 16}}},
	// length 2
	{{{25, 14}, {0, 0}, {0, 0}, {0, 0}},
		{{64, 40}, {90, 62}, {110, 75}, {0, 0}}},
	// length 3
	{{{95, 60}, {0, 0}, {0, 0}, {0, 0}},
		{{265, 135}, {320, 170}, {0, 0}, {0, 0}}},
	// length 4
	{{{310, 190}, {0, 0}, {0, 0}, {0, 0}},
		{{10000, 800}, {0, 0}, {0, 0}, {0, 0}}},
	// length 5
	{{{gomoku.ScoreWin, 500000}, {0, 0}, {0, 0}, {0, 0}},
		{{0, 0}, {0, 0}, {0, 0}, {0, 0}}},
}

const (
	scoreDouble4Attack     = 9000
	scoreDouble4Defense    = 500
	scoreDead4Live3Attack  = 2400
	scoreDead4Live3Defense = 400
	scoreDoubleLive3Attack  = 320
	scoreDoubleLive3Defense = 160
)

// RenjuEvaluator implements gomoku.Evaluator for the renju-basic rule.
type RenjuEvaluator struct{}

func (e *RenjuEvaluator) Init() {
	InitTypeTree()
}

func (e *RenjuEvaluator) EvaluateType(status []board.StoneStatus) gomoku.ChessType {
	return ClassifyType(status)
}

func (e *RenjuEvaluator) CheckWinOrLose(score int) board.GameStatus {
	if score >= gomoku.ScoreWin {
		return board.Winning
	}
	if score == gomoku.ScoreForbidden {
		return board.Losing
	}
	return board.Nothing
}

func (e *RenjuEvaluator) EvaluateScore(types [4]gomoku.ChessType, score *[2]int) {
	const (
		attack  = 0
		defense = 1
		black   = 0
		white   = 1
		live    = 1
		dead    = 0
	)

	score[black] = 0
	score[white] = 0

	// count[color][length][lifeOrDead]
	var count [2][6][2]int
	forbidden := false
	win := false

	for pass := 0; pass < 2; pass++ {
		selfColor := pass
		opponentColor := 1 - pass

		for d := 0; d < 4; d++ {
			length := types[d].Length(selfColor)
			life := types[d].Life(selfColor)
			level := types[d].Level(selfColor)

			// Renju-specific: track forbidden and win for black
			if selfColor == black {
				if length == -1 {
					forbidden = true
					continue
				} else if length == 5 {
					win = true
				}
			}

			// Guard against negative or out-of-range indices
			if length < 0 || length > 5 || level < 0 || level > 3 {
				continue
			}

			count[selfColor][length][life]++

			score[selfColor] += scoreTable[length][life][level][attack]
			score[opponentColor] += scoreTable[length][life][level][defense]
		}
	}

	// Renju forbidden: black double-4 or double-live-3
	if count[black][4][live]+count[black][4][dead] >= 2 ||
		count[black][3][live] >= 2 {
		forbidden = true
	}

	// Combo bonuses
	for pass := 0; pass < 2; pass++ {
		selfColor := pass
		opponentColor := 1 - pass

		if count[selfColor][5][dead] > 0 ||
			count[opponentColor][5][dead] > 0 ||
			count[selfColor][4][live] > 0 {
			continue
		}

		if count[selfColor][4][dead] >= 2 {
			score[selfColor] += scoreDouble4Attack
		} else if count[selfColor][4][dead] > 0 && count[selfColor][3][live] > 0 {
			score[selfColor] += scoreDead4Live3Attack
		} else if count[opponentColor][4][dead] >= 2 {
			// Renju: only black gets double-4 defense bonus
			if selfColor == black {
				score[selfColor] += scoreDouble4Defense
			}
		} else if count[opponentColor][4][dead] > 0 && count[opponentColor][3][live] > 0 {
			score[selfColor] += scoreDead4Live3Defense
		} else if count[selfColor][3][live] >= 2 {
			score[selfColor] += scoreDoubleLive3Attack
		} else if count[opponentColor][3][live] >= 2 {
			// In Renju, black's double-live-3 is a forbidden move, so the
			// two cases are asymmetric:
			// - Opponent is white (has double-live-3): real threat → black gets defense bonus
			// - Opponent is black (has double-live-3): forbidden move → not a real threat,
			//   subtract the inflated live-3 defense score from white
			// Note: C++ original has `if (selfColor = BLACK)` (assignment bug),
			// which breaks both cases. We implement the correct intended behavior.
			if selfColor == black {
				score[selfColor] += scoreDoubleLive3Defense
			} else {
				score[selfColor] -= scoreTable[3][live][0][attack]
			}
		}
	}

	if forbidden && !win {
		score[black] = gomoku.ScoreForbidden
	}
}
