package freestyle

import (
	"github.com/todd/earthmover/internal/board"
	"github.com/todd/earthmover/internal/gomoku"
)

// scoreTable is indexed by [length][life][level][attackOrDefense].
// attack=0, defense=1.
var scoreTable = [6][2][4][2]int{
	// length 0
	{{{0, 0}, {0, 0}, {0, 0}, {0, 0}},
		{{0, 0}, {0, 0}, {0, 0}, {0, 0}}},
	// length 1
	{{{2, 1}, {0, 0}, {0, 0}, {0, 0}},          // dead 1
		{{6, 2}, {9, 5}, {15, 12}, {20, 16}}},   // live 1
	// length 2
	{{{25, 14}, {0, 0}, {0, 0}, {0, 0}},         // dead 2
		{{64, 40}, {90, 62}, {110, 75}, {0, 0}}}, // live 2
	// length 3
	{{{95, 60}, {0, 0}, {0, 0}, {0, 0}},          // dead 3
		{{265, 135}, {320, 170}, {0, 0}, {0, 0}}}, // live 3
	// length 4
	{{{310, 190}, {0, 0}, {0, 0}, {0, 0}},              // dead 4
		{{10000, 800}, {0, 0}, {0, 0}, {0, 0}}},         // live 4
	// length 5
	{{{gomoku.ScoreWin, 500000}, {0, 0}, {0, 0}, {0, 0}}, // 5
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

// FreestyleEvaluator implements gomoku.Evaluator for the freestyle rule.
type FreestyleEvaluator struct{}

func (e *FreestyleEvaluator) Init() {
	InitTypeTree()
}

func (e *FreestyleEvaluator) EvaluateType(status []board.StoneStatus) gomoku.ChessType {
	return ClassifyType(status)
}

func (e *FreestyleEvaluator) CheckWinOrLose(score int) board.GameStatus {
	if score >= gomoku.ScoreWin {
		return board.Winning
	}
	return board.Nothing
}

func (e *FreestyleEvaluator) EvaluateScore(types [4]gomoku.ChessType, score *[2]int) {
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

	for pass := 0; pass < 2; pass++ {
		selfColor := pass          // 0=black first, then 1=white
		opponentColor := 1 - pass

		for d := 0; d < 4; d++ {
			length := types[d].Length(selfColor)
			life := types[d].Life(selfColor)
			level := types[d].Level(selfColor)

			if length < 0 || length > 5 || level < 0 || level > 3 {
				continue
			}

			count[selfColor][length][life]++

			score[selfColor] += scoreTable[length][life][level][attack]
			score[opponentColor] += scoreTable[length][life][level][defense]
		}
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
			score[selfColor] += scoreDouble4Defense
		} else if count[opponentColor][4][dead] > 0 && count[opponentColor][3][live] > 0 {
			score[selfColor] += scoreDead4Live3Defense
		} else if count[selfColor][3][live] >= 2 {
			score[selfColor] += scoreDoubleLive3Attack
		} else if count[opponentColor][3][live] >= 2 {
			score[selfColor] += scoreDoubleLive3Defense
		}
	}
}
