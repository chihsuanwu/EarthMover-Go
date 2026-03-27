# EarthMover-Go

Gomoku (五子棋) AI engine, refactored from C++ to Go.

Original project: `/Users/todd/EarthMover`

## How to Run

```bash
# Terminal interactive mode
go run ./cmd/cli/

# HTTP server (open browser at http://localhost:8080)
go run ./cmd/earthmover/ 8080

# WebAssembly (no backend needed — runs entirely in browser)
./build_wasm.sh
cd static && python3 -m http.server 8080

# Run all tests
go test ./...

# Run tests with race detector
go test -race ./...
```

## Project Structure

```
cmd/
  earthmover/main.go         # HTTP server entry point
  cli/main.go                # Terminal interactive mode
  wasm/main.go               # WebAssembly entry point (syscall/js bridge)
internal/
  board/board.go              # Board interface + constants/enums (StoneStatus, GameStatus, etc.)
  gomoku/
    chesstype.go              # SingleType, ChessType (pattern classification types)
    point.go                  # Point struct (neighbor indices, types, scores)
    typetree.go               # Shared TypeNode, Classify, CutSameResultChild
    evaluator.go              # Evaluator interface, EvaluateRelativeScore (shared logic)
    board.go                  # GomokuBoard (play/undo/scoring, implements board.Board)
  freestyle/
    typetree.go               # Freestyle DFS tree build + typeAnalyze (statusLen=8)
    evaluator.go              # Freestyle scoring table + combo bonuses
    board.go                  # NewBoard() factory
  renju/
    typetree.go               # Renju DFS tree build + typeAnalyze (statusLen=10, overline/forbidden)
    evaluator.go              # Renju scoring + forbidden move detection
    board.go                  # NewBoard() factory
  opening/
    openingtree.go            # Opening book trie (go:embed opening.txt, 8 orientations)
    opening.txt               # Opening data file (copied from C++)
  mcts/
    node.go                   # Node struct (linked list children, UCB, Selection, Merge/Minus)
    tree.go                   # GameTree (MCTS loop, simulation, backprop, multi-goroutine, tree JSON)
  ai/
    ai.go                     # AI controller (3 levels, background thinking, context cancellation)
  server/
    server.go                 # HTTP server (net/http, session management, static file serving)
build_wasm.sh                   # Build script for WASM target
static/
  index.html                  # Web frontend
  gomoku/src/                 # CSS/JS/PNG resources
  gomoku/src/js/worker.js     # Web Worker that loads WASM and bridges AI calls
```

## Excluded from Refactoring

| Module | Reason |
|--------|--------|
| `neuralnetwork/` | Experimental CNN/ResCNN, never integrated into engine |
| `go/` | Abandoned Go (board game) implementation |
| `test/` | Manual test harness — replaced by Go tests |
| `proc_stat.c/.h` | Linux-specific `/proc` parsing — Go has `runtime.MemStats` |
| `dashboard.html` | Depends on `proc_stat` — `/usage` endpoint provides basic data instead |
| `objectcounter.cpp/.h` | C++ debug utility — Go has race detector and pprof |
| `log.cpp/.h` | Trivial logger — Go has `log/slog` |
| `makefile.py` / `Makefile` | Build scripts — Go uses `go build` |
| `lib/json.h` | Third-party JSON — Go has `encoding/json` |
| `compare/` | Python AI comparison framework — can reimplement later via CLI |

## Bugs Fixed from C++ Original

| Bug | C++ Location | Description |
|-----|-------------|-------------|
| Assignment vs comparison | `evaluatorrenjubasic.cpp:118` | `if (selfColor = BLACK)` used `=` instead of `==`. Always entered else branch, corrupting both passes of the scoring loop. Go implements correct asymmetric logic: black gets defense bonus for opponent's double-live-3, white subtracts inflated score (because black's double-live-3 is a forbidden move). |
| Non-atomic thread control | `gametree.cpp:112`, `ai.cpp:64` | `bool*` shared between main thread and search threads without atomic or memory barrier — undefined behavior in C++. Go uses `context.Context`. |
| Thread-unsafe PRNG | `virtualboardgomoku.h:192` | Global `rand()` called from multiple MCTS threads simultaneously. Go uses `math/rand/v2` per-goroutine PRNG. |
| Score table out-of-bounds | `evaluatorrenjubasic.cpp` | `typeAnalyze` can produce `level < 0` (formula: `level - (3 - (length-1))`), used as array index without bounds check. Go adds guard: `if length < 0 \|\| length > 5 \|\| level < 0 \|\| level > 3 { continue }`. |
| Session map data race | `server/httpserver.h:72` | `session2instance_` unordered_map accessed from multiple request handlers without locking. Go uses `sync.Mutex`. |

## Performance Design Decisions

### Preserved from C++
- **Node linked list children**: 16 bytes/node vs sparse array's 1800 bytes/node. Hot path iterates sequentially; random access only on cold paths.
- **Type tree trie lookup**: Built once via `sync.Once`, then read-only. `CutSameResultChild` pruning preserved.
- **UCB + score/scoreSum selection**: Formula identical: `winRate + sqrt(0.5 * ln(parentCount) / (1 + childCount))`.
- **Tree copy-minus-merge parallel search**: Same pattern as C++ but with goroutines + context instead of threads + bool*.
- **Greedy simulation**: `GetHSI()` picks highest-score point, MAX_DEPTH=50.
- **Relative score filtering**: `score * 8 <= highest` and `playNo < 10 && score < 140` thresholds identical.
- **Opening book**: 8-orientation trie with rotate/mirror, boundary check (4-10), random selection among matches.

### Changed from C++
- **Memory pool → GC**: C++ pre-allocates 800K nodes in free-list pool. Go uses standard allocator + GC. Can add slab allocator later if profiling shows GC pressure.
- **Raw pointers → index array**: C++ stores `const StoneStatus*` to neighbor fields. Go stores `[4][10]int16` indices with -1 sentinel. ~5-15% overhead from bounds checking, but clone is a simple value copy (no pointer rewiring).
- **Template inlining → interface dispatch**: C++ `template <class Eva>` fully inlined at compile time. Go uses `Evaluator` interface (~2-5ns/call, <1% of total search time).
- **std::thread → goroutine**: ~2KB stack vs ~1MB, faster startup. `context.Context` for race-free cancellation.
- **Global rand() → per-P PRNG**: `math/rand/v2` is thread-safe with zero contention.
- **Board clone**: C++ manually rewires 225×4×10 neighbor pointers. Go does `clone.Points = b.Points` (value copy, indices remain valid).

## C++ Source Reference

| Go Package | C++ Source |
|------------|-----------|
| `internal/board/` | `const.h`, `virtualboard.h` |
| `internal/gomoku/` | `gomoku/virtualboardgomoku.h`, `gomoku/point.h`, `gomoku/chesstype.h` |
| `internal/freestyle/` | `gomoku/freestyle/typetreefreestyle.{h,cpp}`, `gomoku/freestyle/evaluatorfreestyle.{h,cpp}`, `gomoku/freestyle/virtualboardfreestyle.{h,cpp}` |
| `internal/renju/` | `gomoku/renju_basic/typetreerenjubasic.{h,cpp}`, `gomoku/renju_basic/evaluatorrenjubasic.{h,cpp}`, `gomoku/renju_basic/virtualboardrenjubasic.{h,cpp}` |
| `internal/opening/` | `gomoku/openingtree.h`, `gomoku/opening.txt` |
| `internal/mcts/` | `gametree.{h,cpp}`, `node.{h,cpp}`, `memorypool.{h,cpp}` |
| `internal/ai/` | `ai.{h,cpp}` |
| `internal/server/` | `server/httpserver.{h,cpp}`, `server/httprequest.{h,cpp}`, `server/httpresponse.{h,cpp}` |
| `cmd/earthmover/` | `networkmain.cpp` |
| `cmd/cli/` | `main.cpp`, `gomoku/displayboard.{h,cpp}` |
