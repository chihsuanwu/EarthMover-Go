package gomoku

import "github.com/todd/earthmover/internal/board"

// MaxStatusLength is the maximum neighbor status array length across all rule variants.
// Freestyle uses 8, Renju uses 10.
const MaxStatusLength = 10

// Point represents a single intersection on the Gomoku board.
type Point struct {
	// Types holds the chess type classification for each of the 4 directions.
	// Directions: 0=→ 1=↓ 2=↗ 3=↘
	Types [4]ChessType

	// DirIdx stores neighbor indices for each direction.
	// DirIdx[dir][i] is the index into the Board's points array,
	// or -1 if out of bounds.
	DirIdx [4][MaxStatusLength]int16

	// Stat is the stone status at this point (Black, White, Empty, Bound).
	Stat board.StoneStatus

	// AbsScore holds absolute scores: [0]=black, [1]=white.
	AbsScore [2]int

	// Scr holds the relative (filtered) score used by MCTS.
	Scr int
}

// Status implements opening.PointStatus.
func (p *Point) Status() board.StoneStatus { return p.Stat }

// GetAbsScore implements gomoku.PointScorer.
func (p *Point) GetAbsScore(color int) int { return p.AbsScore[color] }

// SetScore implements gomoku.PointScorer.
func (p *Point) SetScore(score int) { p.Scr = score }
