#!/bin/bash
set -e

echo "Building WASM..."
GOOS=js GOARCH=wasm go build -o static/earthmover.wasm ./cmd/wasm/

echo "Copying wasm_exec.js..."
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" static/wasm_exec.js

WASM_SIZE=$(ls -lh static/earthmover.wasm | awk '{print $5}')
echo "Done! WASM binary: static/earthmover.wasm (${WASM_SIZE})"
echo "Open static/index.html with a web server to play."
echo ""
echo "Quick start:"
echo "  cd static && python3 -m http.server 8080"
