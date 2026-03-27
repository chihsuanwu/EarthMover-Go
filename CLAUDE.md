# EarthMover-Go

Gomoku (五子棋) AI engine, refactored from C++ to Go.

Original project: `/Users/todd/EarthMover`

## Project Structure

```
earthmover-go/
  cmd/
    earthmover/              # HTTP server entry point
    cli/                     # Terminal interactive mode
  internal/
    board/
      board.go               # Board interface + constants/enums
    gomoku/
      point.go               # Point struct
      chesstype.go           # SingleType, ChessType
      board.go               # GomokuBoard (play/undo/scoring)
      evaluator.go           # Shared relative score logic
    freestyle/
      evaluator.go           # Freestyle scoring table + combo bonuses
      typetree.go            # Freestyle pattern recognition tree
    renju/
      evaluator.go           # Renju scoring + forbidden move detection
      typetree.go            # Renju pattern recognition tree
    opening/
      openingtree.go         # Opening book (go:embed opening.txt)
      opening.txt            # Opening data file (copied from C++)
    mcts/
      node.go                # Node struct, UCB, merge
      tree.go                # MCTS main loop: selection/simulation/backprop
    ai/
      ai.go                  # AI controller (levels, thinking, background)
    server/
      server.go              # HTTP handlers, session management
  static/                    # Frontend files (copied from C++ project as-is)
```

## Implementation Phases (Dependency Order)

| Phase | Content | Depends On |
|-------|---------|------------|
| 1 | Constants & core types (`StoneStatus`, `GameStatus`, `SingleType`, `ChessType`) | None |
| 2 | Board interface | Phase 1 |
| 3 | Type trees (pattern recognition) — DFS tree build + `typeAnalyze` recursion. Freestyle (statusLen=8) and Renju (statusLen=10) | Phase 1 |
| 4 | Evaluators — scoring tables, combo detection, forbidden move check. Define `Evaluator` interface | Phases 1, 3 |
| 5 | Opening book — parse `opening.txt`, insert in 8 orientations | Phase 1 |
| 6 | GomokuBoard — neighbor index init, play/undo, score update | Phases 2, 4, 5 |
| 7 | MCTS engine — Node + GameTree, UCB selection, simulation, backprop, multi-goroutine parallel search + tree merge | Phase 6 |
| 8 | AI controller — 3 levels, background thinking, `context.Context` cancellation | Phase 7 |
| 9 | HTTP server — `net/http`, `sync.RWMutex` session map, static file serving | Phase 8 |
| 10 | CLI mode — terminal interactive, lower priority | Phase 8 |

## Excluded from Refactoring

These modules from the C++ project are NOT being ported:

| Module | Reason |
|--------|--------|
| `neuralnetwork/` | Experimental CNN/ResCNN, never integrated into engine |
| `go/` | Abandoned Go (board game) implementation |
| `test/` | Manual test harness with no assertions — replaced by Go tests |
| `proc_stat.c/.h` | Linux-specific `/proc` parsing — Go has `runtime` and `pprof` |
| `dashboard.html` | Depends on `proc_stat` — replace with `pprof` endpoints |
| `objectcounter.cpp/.h` | C++ debug utility — Go has race detector and heap profiling |
| `log.cpp/.h` | Trivial logger — Go has `log/slog` |
| `makefile.py` / `Makefile` | Build scripts — Go uses `go build` |
| `lib/json.h` | Third-party JSON — Go has `encoding/json` |
| `compare/` | Python AI comparison framework — can reimplement later via CLI |

The web frontend (`index.html`, `gomoku/src/`) is copied as-is — it communicates via HTTP/JSON.

## Performance Optimization Strategy

### 1. Memory Pool (Slab Allocator)

C++ uses a free-list pool pre-allocating 800,000 nodes in contiguous memory.

Go strategy: Pre-allocate `[]Node` slice as slab allocator with free-list index management.
- `allocate()` → O(1), return next free index
- `deallocate()` → O(1), push index to free list head
- Optionally disable GC during search with `debug.SetGCPercent(-1)` and re-enable after

### 2. Neighbor Status Access (Index Array)

C++ stores raw pointers (`const StoneStatus*`) to neighbor points' `status_` fields for O(1) lookup.

Go strategy: Store neighbor indices as `[4][10]int16` per Point. Read neighbors via:
```go
func (b *Board) getDirStatus(p, dir int, buf []StoneStatus) {
    for i := 0; i < b.statusLen; i++ {
        idx := b.points[p].DirIdx[dir][i]
        if idx < 0 { buf[i] = Bound } else { buf[i] = b.points[idx].Status }
    }
}
```
Expected overhead: ~5-15% vs raw pointers. Go compiler's bounds check elimination helps.

### 3. Template Polymorphism → Interface

C++ uses `template <class Eva>` for compile-time evaluator injection (fully inlined).

Go strategy: Use `Evaluator` interface initially. The hot path (`evaluateType` + `evaluateScore`) is called ~2000 times per simulation. Interface dispatch cost (~2-5ns/call) is negligible relative to total search time. If profiling shows it matters, switch to concrete types or generics.

### 4. Node Children: Linked List (Preserve C++ Design)

C++ uses singly-linked list (`child_`, `next_` pointers) for Node children.

**Keep this design in Go.** Rationale:
- Linked list: 2 pointers = 16 bytes/node. 800K nodes = 12.8 MB
- Sparse array `[225]*Node`: 1800 bytes/node. 800K nodes = 1.44 GB — unacceptable
- Hot path (MCTS selection) iterates all children sequentially
- `child(index)` random access only on cold paths (play, merge after search)

### 5. Multi-threaded MCTS → Goroutines

C++ uses `std::thread` + `bool*` controller (has data race on the bool flag).

Go strategy: `goroutine` + `context.Context` for cancellation.
- Goroutine startup cost is 1-2 orders of magnitude lower than `std::thread`
- `context.Context` is race-free (unlike the C++ `bool*` which lacks atomic)
- Tree-copy-then-merge pattern stays the same

### 6. Type Tree (Trie Lookup) — No Change Needed

Trie-based pattern classification works identically in Go. Pointer dereference per trie step is the same cost. Tree is built once at init via `sync.Once`, then read-only.

### 7. Random Number Generation

C++ uses `rand()` (global, not thread-safe — data race bug).

Go strategy: `math/rand/v2` top-level functions are per-goroutine (per-P) in Go 1.22+. No locking, no manual `*rand.Rand` management needed.

## Known C++ Bugs to Fix During Refactoring

1. **Renju evaluator** (`evaluatorrenjubasic.cpp:118`): `if (selfColor = BLACK)` uses assignment `=` instead of comparison `==`. Investigate and fix.
2. **`bool*` controller**: Non-atomic shared bool between threads — undefined behavior in C++. Fixed by using `context.Context` in Go.
3. **`rand()` in multi-threaded simulation**: Global `rand()` is not thread-safe. Fixed by Go's per-P PRNG.
4. **HTTP server session map**: Accessed without locks in multi-request scenarios. Fixed by `sync.RWMutex` in Go.

## Testing Strategy

- Every package has unit tests, table-driven
- Type tree: Cross-validate against C++ by extracting all (status → ChessType) mappings
- Evaluator: Known type combinations → verify scores
- Board play/undo: Roundtrip tests ensuring state fully restored
- MCTS: Single cycle verifying selection → simulation → backprop correctness
- Server: `httptest.Server` for full game integration tests
- All tests run with `-race`
- Benchmarks: TypeTree classify, Board play, MCTS sims/sec — compare against C++

## C++ Source Reference

Key files in the original C++ project (`/Users/todd/EarthMover`):

| Go Package | C++ Source |
|------------|-----------|
| `internal/board/` | `const.h`, `virtualboard.h` |
| `internal/gomoku/` | `gomoku/virtualboardgomoku.h`, `gomoku/point.h`, `gomoku/chesstype.h` |
| `internal/freestyle/` | `gomoku/freestyle/typetreefreestyle.cpp`, `gomoku/freestyle/evaluatorfreestyle.cpp` |
| `internal/renju/` | `gomoku/renju_basic/typetreerenjubasic.cpp`, `gomoku/renju_basic/evaluatorrenjubasic.cpp` |
| `internal/opening/` | `gomoku/openingtree.h`, `gomoku/opening.txt` |
| `internal/mcts/` | `gametree.h`, `gametree.cpp`, `node.h`, `node.cpp`, `memorypool.h`, `memorypool.cpp` |
| `internal/ai/` | `ai.h`, `ai.cpp` |
| `internal/server/` | `server/httpserver.h`, `server/httpserver.cpp`, `server/httprequest.cpp`, `server/httpresponse.cpp` |
