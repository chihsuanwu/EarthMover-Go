package gomoku

import (
	"math/rand/v2"
	"sync"

	"github.com/todd/earthmover/internal/board"
	"github.com/todd/earthmover/internal/opening"
)

// Directions: 0=→ 1=↓ 2=↗ 3=↘
var dir = [4][2]int{{0, 1}, {1, 0}, {-1, 1}, {1, 1}}

// GomokuBoard implements the board.Board interface for Gomoku.
// E is the concrete evaluator type, enabling the compiler to inline
// EvaluateType/EvaluateScore calls instead of going through interface dispatch.
type GomokuBoard[E Evaluator] struct {
	Points    [board.Length]Point
	PlayNo    int
	StatusLen int // 8 for freestyle, 10 for renju
	Eval      E

	// statusBuf is a reusable buffer for fillDirStatus to avoid per-call allocation.
	statusBuf [MaxStatusLength]board.StoneStatus

	// pool is shared among all clones of the same board type for reuse.
	pool *sync.Pool
}

// NewBoard creates and initializes a new GomokuBoard with the given evaluator and status length.
func NewBoard[E Evaluator](eval E, statusLen int, pool *sync.Pool) *GomokuBoard[E] {
	b := &GomokuBoard[E]{
		StatusLen: statusLen,
		Eval:      eval,
		pool:      pool,
	}
	eval.Init()
	b.initNeighbors()
	b.initScores()
	return b
}

func (b *GomokuBoard[E]) initNeighbors() {
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

func (b *GomokuBoard[E]) initScores() {
	for i := 0; i < board.Length; i++ {
		for d := 0; d < 4; d++ {
			b.fillDirStatus(i, d)
			b.Points[i].Types[d] = b.Eval.EvaluateType(b.statusBuf[:b.StatusLen])
		}
		b.Eval.EvaluateScore(b.Points[i].Types, &b.Points[i].AbsScore)
	}
	b.evaluateRelativeScore()
}

// fillDirStatus reads the neighbor stone statuses along a direction into b.statusBuf.
func (b *GomokuBoard[E]) fillDirStatus(pointIdx, dir int) {
	for i := 0; i < b.StatusLen; i++ {
		idx := b.Points[pointIdx].DirIdx[dir][i]
		if idx < 0 {
			b.statusBuf[i] = board.Bound
		} else {
			b.statusBuf[i] = b.Points[idx].Stat
		}
	}
}

func (b *GomokuBoard[E]) evaluateRelativeScore() {
	EvaluateRelativeScore(&b.Points, b.PlayNo, openingClassifyPoints)
}

// openingClassifyPoints bridges Points array to the opening book.
func openingClassifyPoints(points *[board.Length]Point) int {
	opPts := make([]opening.PointStatus, board.Length)
	for i := range opPts {
		opPts[i] = &points[i]
	}
	return opening.Classify(opPts)
}

// --- board.Board interface implementation ---

func (b *GomokuBoard[E]) Play(index int) board.GameStatus {
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

func (b *GomokuBoard[E]) Undo(index int) {
	b.PlayNo--

	// index == 225 means previous move was a pass
	if index == board.Dimen*board.Dimen {
		return
	}

	b.Points[index].Stat = board.Empty

	// Re-evaluate the point itself
	for d := 0; d < 4; d++ {
		b.fillDirStatus(index, d)
		b.Points[index].Types[d] = b.Eval.EvaluateType(b.statusBuf[:b.StatusLen])
	}
	b.Eval.EvaluateScore(b.Points[index].Types, &b.Points[index].AbsScore)

	// Re-evaluate neighbors
	b.updateNeighbors(index)
	b.evaluateRelativeScore()
}

// updateNeighbors re-evaluates types and scores for neighboring empty points
// affected by a stone placement or removal at the given index.
func (b *GomokuBoard[E]) updateNeighbors(index int) {
	row := index / board.Dimen
	col := index % board.Dimen
	halfLen := b.StatusLen / 2

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

				b.fillDirStatus(checkIdx, d)
				b.Points[checkIdx].Types[d] = b.Eval.EvaluateType(b.statusBuf[:b.StatusLen])
				b.Eval.EvaluateScore(b.Points[checkIdx].Types, &b.Points[checkIdx].AbsScore)
			}
		}
	}
}

func (b *GomokuBoard[E]) Pass() int {
	b.PlayNo++
	return board.Length // 225
}

func (b *GomokuBoard[E]) GetScore(index int) int {
	return b.Points[index].Scr
}

func (b *GomokuBoard[E]) GetScoreSum() int {
	sum := 0
	for i := 0; i < board.Length; i++ {
		if s := b.Points[i].Scr; s > 0 {
			sum += s
		}
	}
	return sum
}

func (b *GomokuBoard[E]) GetHSI() int {
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

func (b *GomokuBoard[E]) GetHSIFiltered(ignore []bool) int {
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

func (b *GomokuBoard[E]) WhoTurn() bool {
	return b.PlayNo&1 == 1
}

func (b *GomokuBoard[E]) Length() int {
	return board.Length
}

func (b *GomokuBoard[E]) Clone() board.Board {
	clone := b.pool.Get().(*GomokuBoard[E])
	clone.PlayNo = b.PlayNo
	clone.StatusLen = b.StatusLen
	clone.Eval = b.Eval
	clone.Points = b.Points
	clone.pool = b.pool
	return clone
}

// Release returns a cloned board to its pool for reuse.
func Release(b board.Board) {
	type poolable interface {
		release()
	}
	if p, ok := b.(poolable); ok {
		p.release()
	}
}

func (b *GomokuBoard[E]) release() {
	b.pool.Put(b)
}

func (b *GomokuBoard[E]) Create() board.Board {
	return NewBoard(b.Eval, b.StatusLen, b.pool)
}
