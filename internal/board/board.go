// Package board defines the core constants, enums, and Board interface
// for the Gomoku AI engine.
package board

// StoneStatus represents the state of a board intersection.
type StoneStatus int8

const (
	Black StoneStatus = 0
	White StoneStatus = 1
	Empty StoneStatus = 2
	Bound StoneStatus = 3
)

// Opponent returns the opposing color. Only valid for Black and White.
func (s StoneStatus) Opponent() StoneStatus {
	return 1 - s
}

// GameStatus represents the result of a move (from the current player's perspective).
type GameStatus int

const (
	Losing  GameStatus = -1
	Nothing GameStatus = 0
	Winning GameStatus = 1
)

// SearchStatus represents the result of an MCTS selection or simulation step.
type SearchStatus int

const (
	Lose    SearchStatus = -1
	Tie     SearchStatus = 0
	Win     SearchStatus = 1
	Unknown SearchStatus = 2
	Leaf    SearchStatus = 3
)

// Reverse flips Win <-> Lose, leaves Tie unchanged.
func (s SearchStatus) Reverse() SearchStatus {
	return SearchStatus(-int(s))
}

// Rule identifies which Gomoku ruleset to use.
type Rule int

const (
	RuleFreestyle  Rule = 0
	RuleRenjuBasic Rule = 1
)

// Board dimensions.
const (
	Dimen  = 15
	Length = Dimen * Dimen // 225
)

// Board is the interface that all board implementations must satisfy.
// It is used by the MCTS engine to evaluate positions.
type Board interface {
	// Play places a stone at the given index.
	// Returns the game status after the move.
	Play(index int) GameStatus

	// Undo removes the stone at the given index.
	Undo(index int)

	// Pass skips the current player's turn.
	// Returns the index representing the pass move, or -1 if pass is not allowed.
	Pass() int

	// GetScore returns the evaluation score at the given index.
	GetScore(index int) int

	// GetScoreSum returns the sum of all positive point scores.
	GetScoreSum() int

	// GetHSI returns the index of the highest-scoring empty point.
	// Returns -1 if no useful point exists.
	GetHSI() int

	// GetHSIFiltered returns the highest-scoring index among points
	// where ignore[index] is false.
	// Returns -1 if no useful point exists.
	GetHSIFiltered(ignore []bool) int

	// WhoTurn returns the color of the current player.
	// false (0) = Black, true (1) = White.
	WhoTurn() bool

	// Length returns the total number of intersections (225).
	Length() int

	// Clone creates a deep copy of the board.
	Clone() Board

	// Create creates a new empty board with the same rule variant.
	Create() Board
}
