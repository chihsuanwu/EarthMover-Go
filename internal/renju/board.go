package renju

import (
	"github.com/todd/earthmover/internal/board"
	"github.com/todd/earthmover/internal/gomoku"
	"github.com/todd/earthmover/internal/opening"
)

func init() {
	opening.Init()
}

// NewBoard creates a new Gomoku board with Renju-Basic rules.
func NewBoard() board.Board {
	return gomoku.NewBoard(&RenjuEvaluator{}, statusLength)
}
