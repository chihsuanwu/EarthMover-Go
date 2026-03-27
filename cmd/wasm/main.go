//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/todd/earthmover/internal/board"
	"github.com/todd/earthmover/internal/freestyle"
	"github.com/todd/earthmover/internal/mcts"
	"github.com/todd/earthmover/internal/renju"
)

// wasmAI holds the game state for the WASM build.
// Since there is no server, we only need a single instance.
type wasmAI struct {
	tree  *mcts.GameTree
	level int
}

// Single-thread level settings: [batch, minCount]
// No multi-threading in WASM — goroutines share one OS thread.
var wasmLevelConfig = [3][2]int{
	{9600, 7200},   // Level 0: Normal   (1.5x server)
	{9600, 12000},  // Level 1: Advanced (1.5x server)
	{9600, 19200},  // Level 2: Master   (1.5x server)
}

var ai wasmAI

func main() {
	ai.tree = mcts.NewGameTree()

	js.Global().Set("wasmStart", js.FuncOf(wasmStart))
	js.Global().Set("wasmPlay", js.FuncOf(wasmPlay))
	js.Global().Set("wasmThink", js.FuncOf(wasmThink))
	js.Global().Set("wasmPass", js.FuncOf(wasmPass))
	js.Global().Set("wasmUndo", js.FuncOf(wasmUndo))
	js.Global().Set("wasmVisualize", js.FuncOf(wasmVisualize))

	// Block forever so the Go runtime stays alive.
	select {}
}

// wasmStart(level, rule) — initialize a new game.
func wasmStart(_ js.Value, args []js.Value) any {
	level := args[0].Int()
	rule := board.Rule(args[1].Int())

	var brd board.Board
	switch rule {
	case board.RuleFreestyle:
		brd = freestyle.NewBoard()
	case board.RuleRenjuBasic:
		brd = renju.NewBoard()
	default:
		brd = freestyle.NewBoard()
	}

	ai.tree.Init(brd)
	ai.level = level
	return nil
}

// wasmPlay(row, col) — human places a stone. Returns JSON: {"winner": -1|0|1}
func wasmPlay(_ js.Value, args []js.Value) any {
	row := args[0].Int()
	col := args[1].Int()
	index := row*board.Dimen + col

	status := ai.tree.Play(index)
	winner := -1
	if status != board.Nothing {
		whoTurn := ai.tree.CurrentBoard.WhoTurn()
		if (status == board.Losing) != whoTurn {
			winner = 1
		} else {
			winner = 0
		}
	}

	return toJSObject(map[string]any{"winner": winner})
}

// wasmThink() — AI computes and plays its move. Returns JSON: {"row", "col", "winner"}
func wasmThink(_ js.Value, _ []js.Value) any {
	cfg := wasmLevelConfig[ai.level]
	ai.tree.MCTSBatch(cfg[0], cfg[1])
	index := ai.tree.MCTSResult()

	row, col := -1, -1
	winner := -1

	if index == -1 {
		ai.tree.Pass()
	} else {
		row = index / board.Dimen
		col = index % board.Dimen
		status := ai.tree.Play(index)
		if status != board.Nothing {
			whoTurn := ai.tree.CurrentBoard.WhoTurn()
			if (status == board.Losing) != whoTurn {
				winner = 1
			} else {
				winner = 0
			}
		}
	}

	return toJSObject(map[string]any{
		"row":    row,
		"col":    col,
		"winner": winner,
	})
}

// wasmPass() — current player passes.
func wasmPass(_ js.Value, _ []js.Value) any {
	ai.tree.Pass()
	return nil
}

// wasmUndo(times) — undo the given number of moves.
func wasmUndo(_ js.Value, args []js.Value) any {
	times := args[0].Int()
	for range times {
		ai.tree.Undo()
	}
	return nil
}

// wasmVisualize() — returns the MCTS tree as a JSON string for D3 visualization.
func wasmVisualize(_ js.Value, _ []js.Value) any {
	tj := ai.tree.GetTreeJSON()
	if tj == nil {
		return js.Null()
	}
	data, err := json.Marshal(tj)
	if err != nil {
		return js.Null()
	}
	return js.Global().Get("JSON").Call("parse", string(data))
}

// toJSObject converts a Go map to a JS object.
func toJSObject(m map[string]any) js.Value {
	obj := js.Global().Get("Object").New()
	for k, v := range m {
		obj.Set(k, v)
	}
	return obj
}
