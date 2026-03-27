package mcts

import (
	"testing"

	"github.com/todd/earthmover/internal/freestyle"
)

func BenchmarkMCTS100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		brd := freestyle.NewBoard()
		tree := NewGameTree()
		tree.Init(brd)
		tree.MCTS(100)
	}
}

func BenchmarkMCTS1Cycle(b *testing.B) {
	brd := freestyle.NewBoard()
	tree := NewGameTree()
	tree.Init(brd)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.MCTS(1)
	}
}
