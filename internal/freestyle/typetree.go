// Package freestyle implements the Freestyle Gomoku rule variant.
package freestyle

import (
	"sync"

	"github.com/todd/earthmover/internal/board"
	"github.com/todd/earthmover/internal/gomoku"
)

const (
	statusLength  = 8
	analyzeLength = statusLength + 1 // 9
)

var (
	typeTreeRoot *gomoku.TypeNode
	typeTreeOnce sync.Once
)

// InitTypeTree initializes the freestyle type tree. Safe for concurrent use.
func InitTypeTree() {
	typeTreeOnce.Do(func() {
		typeTreeRoot = &gomoku.TypeNode{}
		plantTree()
		gomoku.CutSameResultChild(typeTreeRoot)
	})
}

// ClassifyType classifies a status array into a ChessType using the freestyle type tree.
func ClassifyType(status []board.StoneStatus) gomoku.ChessType {
	return gomoku.Classify(typeTreeRoot, status, statusLength)
}

func plantTree() {
	status := make([]board.StoneStatus, analyzeLength)
	for i := range status {
		status[i] = board.Empty
	}
	dfs(typeTreeRoot, status, analyzeLength/2, -1, false, false)
}

func dfs(node *gomoku.TypeNode, status []board.StoneStatus, location, move int,
	blackBlock, whiteBlock bool) {

	switch status[location] {
	case board.Black:
		blackBlock = true
	case board.White:
		whiteBlock = true
	}

	if (blackBlock && whiteBlock) || status[location] == board.Bound ||
		location <= 0 || location >= analyzeLength-1 {
		if move == 1 {
			// Reached leaf
			bType := typeAnalyze(status, board.Black, true)
			wType := typeAnalyze(status, board.White, true)
			node.Type = gomoku.NewChessType(bType, wType)
			node.Leaf = true
			return
		}
		// Jump to right side
		node.Jump = true
		move += 2
		location = analyzeLength / 2
		blackBlock = false
		whiteBlock = false
	}

	location += move

	stones := [4]board.StoneStatus{board.Black, board.White, board.Empty, board.Bound}
	for i := 0; i < 4; i++ {
		node.Children[i] = &gomoku.TypeNode{}
		status[location] = stones[i]
		dfs(node.Children[i], status, location, move, blackBlock, whiteBlock)
	}

	status[location] = board.Empty
}

func typeAnalyze(status []board.StoneStatus, color board.StoneStatus, checkLevel bool) gomoku.SingleType {
	connect := 1
	// Count the center group (CG) length around the analyze point
	for move, start := -1, analyzeLength/2-1; move <= 1; move, start = move+2, start+2 {
		for i, cp := 0, start; i < 4; i, cp = i+1, cp+move {
			if status[cp] != color {
				break
			}
			connect++
		}
	}

	if connect >= 5 {
		return gomoku.SingleType{Length: 5, Life: 0, Level: 0}
	}

	// CG length < 5: play at the analyze point and examine both sides
	status[analyzeLength/2] = color

	var lType, rType gomoku.SingleType
	lInit, rInit := false, false
	level := int8(0)

	for move, start := -1, analyzeLength/2-1; move <= 1; move, start = move+2, start+2 {
		for count, cp := 0, start; count < 4; count, cp = count+1, cp+move {
			if status[cp] == color {
				continue
			}

			// Reached CG boundary
			var typ gomoku.SingleType
			blocked := false

			if status[cp] == board.Empty {
				// Create transformed status array centered on this empty point
				newStatus := make([]board.StoneStatus, analyzeLength)
				for i := 0; i < analyzeLength; i++ {
					ti := i - (analyzeLength/2 - cp)
					if ti < 0 || ti >= analyzeLength {
						newStatus[i] = board.Bound
					} else {
						newStatus[i] = status[ti]
					}
				}
				typ = typeAnalyze(newStatus, color, false)
			} else {
				blocked = true
				typ = gomoku.SingleType{}
			}

			if move == -1 {
				// Left side
				if !lInit {
					lInit = true
					lType = typ
					if !checkLevel || blocked {
						break
					}
					if lType.Life == 0 || lType.Length > 3 {
						break
					}
				} else {
					if lType.Equal(typ) {
						level++
					} else {
						break
					}
				}
			} else {
				// Right side
				if !rInit {
					rInit = true
					rType = typ
					if !checkLevel || blocked {
						break
					}
					if rType.Life == 0 || rType.Length > 3 {
						break
					}
					if lType.Equal(rType) {
						level++
					} else if lType.Less(rType) {
						level = 0
					} else {
						break
					}
				} else {
					if rType.Equal(typ) && rType.GreaterOrEqual(lType) {
						level++
					} else {
						break
					}
				}
			}
		}
	}

	// Restore analyze point
	status[analyzeLength/2] = board.Empty

	// Keep lType >= rType
	if lType.Less(rType) {
		lType, rType = rType, lType
	}

	if lType.Length == 5 && rType.Length == 5 {
		// Both sides produce five → live four
		return gomoku.SingleType{Length: 4, Life: 1, Level: 0}
	} else if lType.Length == 5 {
		// Only one side produces five → dead four
		return gomoku.SingleType{Length: 4, Life: 0, Level: 0}
	} else if lType.Length == 0 {
		// Useless point
		return gomoku.SingleType{Length: 0, Life: 0, Level: 0}
	} else {
		// Length < 4: current result = lower recursion result - 1
		if checkLevel {
			if lType.Life == 0 || lType.Length > 3 {
				return gomoku.SingleType{Length: lType.Length - 1, Life: lType.Life, Level: 0}
			}
			return gomoku.SingleType{
				Length: lType.Length - 1,
				Life:   lType.Life,
				Level:  level - (3 - (lType.Length - 1)),
			}
		}
		return gomoku.SingleType{Length: lType.Length - 1, Life: lType.Life, Level: 0}
	}
}
