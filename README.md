# EarthMover-Go

Gomoku (Five in a Row) AI engine, ported from C++ to Go.

Original project: [Pin-Yen/EarthMover](https://github.com/Pin-Yen/EarthMover)

## Play Online

https://chihsuanwu.github.io/EarthMover-Go/

Runs entirely in your browser via WebAssembly — no server required.

## Features

- **MCTS-based AI** with 3 difficulty levels
- **Freestyle and Renju-basic** rule support
- **Opening book** with 8-orientation trie lookup
- **Analysis tree visualization** powered by D3.js
- **WebAssembly build** for serverless deployment

## Run Locally

```bash
# WebAssembly (browser)
./build_wasm.sh
cd static && python3 -m http.server 8080

# HTTP server
go run ./cmd/earthmover/ 8080

# Terminal interactive mode
go run ./cmd/cli/
```

## Changes from C++ Original

- Ported to idiomatic Go with full test coverage
- Fixed multiple bugs from C++ version (see CLAUDE.md for details)
- Added WebAssembly support for static hosting deployment
- Replaced raw threads with goroutines and `context.Context`
- Replaced custom memory pool with Go GC + `sync.Pool`
