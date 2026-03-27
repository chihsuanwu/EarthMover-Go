package renju

import (
	"sync"

	"github.com/todd/earthmover/internal/board"
	"github.com/todd/earthmover/internal/gomoku"
	"github.com/todd/earthmover/internal/opening"
)

func init() {
	opening.Init()
}

var renjuPool = sync.Pool{
	New: func() any { return &gomoku.GomokuBoard[RenjuEvaluator]{} },
}

// NewBoard creates a new Gomoku board with Renju-Basic rules.
func NewBoard() board.Board {
	return gomoku.NewBoard(RenjuEvaluator{}, statusLength, &renjuPool)
}
