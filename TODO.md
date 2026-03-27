# Future Improvements

## AI Strength

### Better Simulation Strategy
Currently simulation is pure greedy (always picks highest-score point). This creates bias toward positions that score well locally but miss tactical sequences. Options:
- Add randomness with epsilon-greedy or softmax selection
- Use pattern-based rollout (recognize common tactical shapes during simulation)
- Reduce MAX_DEPTH from 50 when position evaluation is confident

### RAVE / AMAF (All-Moves-As-First)
Let less-visited nodes borrow statistics from other paths where the same move appeared. This accelerates convergence in early search when most nodes have few simulations. Blend RAVE value with UCB using a decay parameter.

### Progressive Widening
Currently selection considers all empty points. Progressive widening limits the number of children expanded per node, growing as visit count increases. This concentrates search resources on promising moves instead of spreading thin.

### Neural Network Integration
The original C++ project had unfinished CNN experiments (`neuralnetwork/`). A modern approach:
- Policy network to guide MCTS selection (replace UCB + score heuristic)
- Value network to replace simulation (direct position evaluation)
- Self-play training loop (AlphaZero-style)

### Evaluation Function Tuning
The score tables (`scoreTable[6][2][4][2]`) use hand-tuned magic numbers. Could be improved by:
- Self-play with parameter optimization (e.g. CMA-ES, Bayesian optimization)
- Training from game records to learn position values
- A/B testing via the `compare/` framework (reimplemented in Go)

## Engineering Quality

### Cross-Validation with C++ Version
Run the C++ type tree with instrumentation to dump all (status array → ChessType) mappings, then verify the Go type tree produces identical results. This is the most critical correctness check since typeAnalyze is the most complex algorithm.

### Frontend Modernization
Current frontend uses jQuery + D3.js (circa 2015). Consider:
- React or Vue rewrite
- Better mobile support
- Improved game replay controls

### WebSocket for Search Progress
Currently `/think` blocks until search completes. WebSocket would allow:
- Real-time search progress (simulation count, current best move, win rate)
- User can see the AI "thinking" and stop it early
- Server push instead of client polling for keepAlive

### Slab Allocator for MCTS Nodes
Currently MCTS nodes use standard Go allocation + `sync.Pool` for boards. A slab allocator (pre-allocated `[]Node` slice with free-list) would:
- Eliminate all GC scanning of the tree during search
- Match C++ memory pool performance
- Allow `debug.SetGCPercent(-1)` during search without memory growth

### Opening Book Expansion
Current book covers moves 1-5 only (from `opening.txt`). Could expand by:
- Importing professional game databases
- Generating openings from self-play at high search depth
- Supporting deeper opening lines (moves 6-10)

### Dashboard
The C++ version had `dashboard.html` with CPU/memory monitoring (via `proc_stat`). Could implement a Go equivalent using `runtime.MemStats` and `runtime/pprof` endpoints, or integrate with Prometheus/Grafana.
