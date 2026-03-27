// Package ai provides the high-level AI controller for the Gomoku engine.
package ai

import (
	"context"
	"encoding/json"
	"time"

	"github.com/todd/earthmover/internal/board"
	"github.com/todd/earthmover/internal/freestyle"
	"github.com/todd/earthmover/internal/mcts"
	"github.com/todd/earthmover/internal/renju"
)

// Level settings: [threads, batch, minCount]
var levelConfig = [3][3]int{
	{2, 6400, 4800},  // Level 0: Normal
	{3, 6400, 8000},  // Level 1: Advanced
	{4, 6400, 12800}, // Level 2: Master
}

const (
	maxBackgroundCycle = 100000
	maxKeepAliveSec    = 70
)

// AI is the high-level controller for the Gomoku AI.
type AI struct {
	tree          *mcts.GameTree
	level         int
	lastAliveTime time.Time

	// bgCancel cancels the background thinking goroutine.
	bgCancel context.CancelFunc
	bgDone   chan struct{}
}

// New creates a new AI instance.
func New() *AI {
	return &AI{
		tree:          mcts.NewGameTree(),
		lastAliveTime: time.Now(),
	}
}

// Reset initializes the AI for a new game with the given level and rule.
func (a *AI) Reset(level int, rule board.Rule) {
	var brd board.Board
	switch rule {
	case board.RuleFreestyle:
		brd = freestyle.NewBoard()
	case board.RuleRenjuBasic:
		brd = renju.NewBoard()
	}

	a.tree.Init(brd)
	a.level = level
}

// Think runs MCTS and returns the best move index. Returns -1 for pass.
func (a *AI) Think() int {
	cfg := levelConfig[a.level]
	a.tree.MCTSMulti(cfg[0], cfg[1], cfg[2])
	return a.tree.MCTSResult()
}

// Play records an opponent's move. Returns the winner:
//
//	-1 = no winner, 0 = black wins, 1 = white wins.
func (a *AI) Play(index int) int {
	status := a.tree.Play(index)
	if status == board.Nothing {
		return -1
	}
	// Map (status, whoTurn) to winner color
	whoTurn := a.tree.CurrentBoard.WhoTurn()
	if (status == board.Losing) != whoTurn {
		return 1
	}
	return 0
}

// Undo reverts the given number of moves.
func (a *AI) Undo(times int) {
	for i := 0; i < times; i++ {
		a.tree.Undo()
	}
}

// Pass records an opponent's pass.
func (a *AI) Pass() {
	a.tree.Pass()
}

// WhoTurn returns false for black, true for white.
func (a *AI) WhoTurn() bool {
	return a.tree.CurrentBoard.WhoTurn()
}

// ThinkInBackground starts background MCTS thinking.
// Call StopBackground to stop it.
func (a *AI) ThinkInBackground() {
	a.StopBackground()

	ctx, cancel := context.WithCancel(context.Background())
	a.bgCancel = cancel
	a.bgDone = make(chan struct{})

	go func() {
		defer close(a.bgDone)
		a.tree.MCTSMultiWithContext(4, ctx, maxBackgroundCycle)
	}()
}

// StopBackground stops background thinking and waits for it to finish.
func (a *AI) StopBackground() {
	if a.bgCancel != nil {
		a.bgCancel()
		<-a.bgDone // wait for goroutine to finish
		a.bgCancel = nil
		a.bgDone = nil
	}
}

// RenewLiveTime updates the last alive timestamp.
func (a *AI) RenewLiveTime() {
	a.lastAliveTime = time.Now()
}

// IsAlive returns true if the AI has been active within the keep-alive window.
func (a *AI) IsAlive() bool {
	return time.Since(a.lastAliveTime).Seconds() <= maxKeepAliveSec
}

// GetTreeJSON returns the game tree as a JSON string for the frontend visualization.
func (a *AI) GetTreeJSON() string {
	tj := a.tree.GetTreeJSON()
	if tj == nil {
		return ""
	}
	data, err := json.Marshal(tj)
	if err != nil {
		return ""
	}
	return string(data)
}

// LastAliveTime returns the last alive timestamp.
func (a *AI) LastAliveTime() time.Time {
	return a.lastAliveTime
}
