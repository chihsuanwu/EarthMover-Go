// Web Worker that loads the Go WASM binary and exposes AI functions.
// The main thread communicates via postMessage to avoid UI freezing.

importScripts('/wasm_exec.js');

const go = new Go();

let ready = false;
const pending = [];

WebAssembly.instantiateStreaming(fetch('/earthmover.wasm'), go.importObject).then(result => {
  go.run(result.instance);
  ready = true;
  // Process any messages that arrived before WASM was ready.
  for (const msg of pending) {
    handleMessage(msg);
  }
  pending.length = 0;
  postMessage({ type: 'ready' });
});

onmessage = function(e) {
  if (!ready) {
    pending.push(e.data);
    return;
  }
  handleMessage(e.data);
};

function handleMessage(data) {
  const { id, action, params } = data;

  try {
    let result;
    switch (action) {
      case 'start':
        wasmStart(params.level, params.rule);
        result = undefined;
        break;
      case 'play':
        result = wasmPlay(params.row, params.col);
        break;
      case 'think':
        result = wasmThink();
        break;
      case 'pass':
        wasmPass();
        result = undefined;
        break;
      case 'undo':
        wasmUndo(params.times);
        result = undefined;
        break;
      case 'visualize':
        result = wasmVisualize();
        break;
      case 'resign':
        // No backend state to update — client handles winner display.
        result = undefined;
        break;
      default:
        throw new Error('unknown action: ' + action);
    }
    postMessage({ id, result });
  } catch (err) {
    postMessage({ id, error: err.message || String(err) });
  }
}
