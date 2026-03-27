package gomoku

import (
	"math/rand/v2"

	"github.com/todd/earthmover/internal/board"
	"github.com/todd/earthmover/internal/opening"
)

// Directions: 0=→ 1=↓ 2=↗ 3=↘
var dir = [4][2]int{{0, 1}, {1, 0}, {-1, 1}, {1, 1}}

// GomokuBoard implements the board.Board interface for Gomoku.
type GomokuBoard struct {
	Points    [board.Length]Point
	PlayNo    int
	StatusLen int // 8 for freestyle, 10 for renju
	Eval      Evaluator
}

// NewBoard creates and initializes a new GomokuBoard with the given evaluator and status length.
func NewBoard(eval Evaluator, statusLen int) *GomokuBoard {
	b := &GomokuBoard{
		StatusLen: statusLen,
		Eval:      eval,
	}
	eval.Init()
	b.initNeighbors()
	b.initScores()
	return b
}

func (b *GomokuBoard) initNeighbors() {
	// Initialize all points to Empty (zero-value is Black=0, not Empty=2)
	for i := 0; i < board.Length; i++ {
		b.Points[i].Stat = board.Empty
	}

	halfLen := b.StatusLen / 2

	for r := 0; r < board.Dimen; r++ {
		for c := 0; c < board.Dimen; c++ {
			i := r*board.Dimen + c
			for d := 0; d < 4; d++ {
				idx := 0
				for offset := -halfLen; offset <= halfLen; offset++ {
					if offset == 0 {
						continue
					}
					checkRow := r + dir[d][0]*offset
					checkCol := c + dir[d][1]*offset
					if checkRow < 0 || checkRow >= board.Dimen ||
						checkCol < 0 || checkCol >= board.Dimen {
						b.Points[i].DirIdx[d][idx] = -1
					} else {
						b.Points[i].DirIdx[d][idx] = int16(checkRow*board.Dimen + checkCol)
					}
					idx++
				}
			}
		}
	}
}

func (b *GomokuBoard) initScores() {
	statusBuf := make([]board.StoneStatus, b.StatusLen)

	for i := 0; i < board.Length; i++ {
		for d := 0; d < 4; d++ {
			b.getDirStatus(i, d, statusBuf)
			b.Points[i].Types[d] = b.Eval.EvaluateType(statusBuf)
		}
		b.Eval.EvaluateScore(b.Points[i].Types, &b.Points[i].AbsScore)
	}
	b.evaluateRelativeScore()
}

// getDirStatus reads the neighbor stone statuses along a direction into buf.
func (b *GomokuBoard) getDirStatus(pointIdx, dir int, buf []board.StoneStatus) {
	for i := 0; i < b.StatusLen; i++ {
		idx := b.Points[pointIdx].DirIdx[dir][i]
		if idx < 0 {
			buf[i] = board.Bound
		} else {
			buf[i] = b.Points[idx].Stat
		}
	}
}

func (b *GomokuBoard) evaluateRelativeScore() {
	// Bridge Point to PointScorer interface
	scorers := make([]PointScorer, board.Length)
	for i := range scorers {
		scorers[i] = &b.Points[i]
	}

	// Bridge to opening.PointStatus
	openingClassify := func(pts []PointScorer) int {
		opPts := make([]opening.PointStatus, len(pts))
		for i := range pts {
			opPts[i] = &b.Points[i]
		}
		return opening.Classify(opPts)
	}

	EvaluateRelativeScore(scorers, b.PlayNo, openingClassify)
}

// --- board.Board interface implementation ---

func (b *GomokuBoard) Play(index int) board.GameStatus {
	whoTurn := b.PlayNo & 1
	status := b.Eval.CheckWinOrLose(b.Points[index].AbsScore[whoTurn])
	if status != board.Nothing {
		return status
	}

	b.PlayNo++

	color := board.Black
	if b.PlayNo&1 == 0 {
		color = board.White
	}

	b.Points[index].Stat = color
	b.Points[index].AbsScore = [2]int{-1, -1}

	b.updateNeighbors(index)
	b.evaluateRelativeScore()

	return board.Nothing
}

func (b *GomokuBoard) Undo(index int) {
	b.PlayNo--

	// index == 225 means previous move was a pass
	if index == board.Dimen*board.Dimen {
		return
	}

	b.Points[index].Stat = board.Empty

	// Re-evaluate the point itself
	statusBuf := make([]board.StoneStatus, b.StatusLen)
	for d := 0; d < 4; d++ {
		b.getDirStatus(index, d, statusBuf)
		b.Points[index].Types[d] = b.Eval.EvaluateType(statusBuf)
	}
	b.Eval.EvaluateScore(b.Points[index].Types, &b.Points[index].AbsScore)

	// Re-evaluate neighbors
	b.updateNeighbors(index)
	b.evaluateRelativeScore()
}

// updateNeighbors re-evaluates types and scores for neighboring empty points
// affected by a stone placement or removal at the given index.
func (b *GomokuBoard) updateNeighbors(index int) {
	row := index / board.Dimen
	col := index % board.Dimen
	halfLen := b.StatusLen / 2
	statusBuf := make([]board.StoneStatus, b.StatusLen)

	for d := 0; d < 4; d++ {
		for move := -1; move <= 1; move += 2 {
			var block [2]bool

			for offset := 1; offset <= halfLen+1; offset++ {
				checkRow := row + dir[d][0]*move*offset
				checkCol := col + dir[d][1]*move*offset

				if checkRow < 0 || checkRow >= board.Dimen ||
					checkCol < 0 || checkCol >= board.Dimen {
					break
				}

				checkIdx := checkRow*board.Dimen + checkCol

				if b.Points[checkIdx].Stat != board.Empty {
					block[b.Points[checkIdx].Stat] = true
					if block[0] && block[1] {
						break
					}
					continue
				}

				b.getDirStatus(checkIdx, d, statusBuf)
				b.Points[checkIdx].Types[d] = b.Eval.EvaluateType(statusBuf)
				b.Eval.EvaluateScore(b.Points[checkIdx].Types, &b.Points[checkIdx].AbsScore)
			}
		}
	}
}

func (b *GomokuBoard) Pass() int {
	b.PlayNo++
	return board.Length // 225
}

func (b *GomokuBoard) GetScore(index int) int {
	return b.Points[index].Scr
}

func (b *GomokuBoard) GetScoreSum() int {
	sum := 0
	for i := 0; i < board.Length; i++ {
		if s := b.Points[i].Scr; s > 0 {
			sum += s
		}
	}
	return sum
}

func (b *GomokuBoard) GetHSI() int {
	max := 0
	same := 0
	index := -1

	for i := 0; i < board.Length; i++ {
		if b.Points[i].Stat != board.Empty {
			continue
		}

		score := b.Points[i].Scr
		if score > max {
			same = 1
			max = score
			index = i
		} else if score == max {
			same++
			if rand.Float64() <= 1.0/float64(same) {
				index = i
			}
		}
	}
	return index
}

func (b *GomokuBoard) GetHSIFiltered(ignore []bool) int {
	max := 0
	index := -1

	for i := 0; i < board.Length; i++ {
		if ignore[i] {
			continue
		}
		if score := b.Points[i].Scr; score > max {
			max = score
			index = i
		}
	}
	return index
}

func (b *GomokuBoard) WhoTurn() bool {
	return b.PlayNo&1 == 1
}

func (b *GomokuBoard) Length() int {
	return board.Length
}

func (b *GomokuBoard) Clone() board.Board {
	clone := &GomokuBoard{
		PlayNo:    b.PlayNo,
		StatusLen: b.StatusLen,
		Eval:      b.Eval,
	}

	// Copy points (value types, no pointer sharing)
	clone.Points = b.Points

	// Re-wire neighbor indices (they index into clone.Points, same layout)
	// Since DirIdx stores indices (not pointers), no rewiring needed.

	return clone
}

func (b *GomokuBoard) Create() board.Board {
	return NewBoard(b.Eval, b.StatusLen)
}
