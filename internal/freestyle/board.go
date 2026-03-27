package freestyle

import (
	"sync"

	"github.com/todd/earthmover/internal/board"
	"github.com/todd/earthmover/internal/gomoku"
	"github.com/todd/earthmover/internal/opening"
)

func init() {
	opening.Init()
}

var freestylePool = sync.Pool{
	New: func() any { return &gomoku.GomokuBoard[FreestyleEvaluator]{} },
}

// NewBoard creates a new Gomoku board with Freestyle rules.
func NewBoard() board.Board {
	return gomoku.NewBoard(FreestyleEvaluator{}, statusLength, &freestylePool)
}
