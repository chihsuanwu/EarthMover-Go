let timer = { black: null, white: null };

let player = { black: 'human', white: 'computer' };

const game = {
  black: 'N/A', white: 'N/A', rule: 'N/A', startTime: -1,
  earthmover: { level: -1, version: 0 }
};

const board = new Board();

const dialog = new Dialog();

const sessionManager = new SessionManager();

// --- Helpers ---

function setDisabled(selector, disabled) {
  document.querySelectorAll(selector).forEach(el => el.disabled = disabled);
}

function notifyWinner(winnerColor) {
  setTimeout(() => alert(winnerColor + ' wins !'), 0);
  board.enable = false;
  board.gameStarted = false;
  setDisabled('.ctrl-replay input', false);
  setDisabled('.ctrl-game input', true);
  setDisabled('.ctrl-analyze input', true);
}

async function post(params, path) {
  params = params || {};
  params.sessionId = sessionManager.getSessionID();

  const res = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params),
  });

  if (res.status === 204) return;
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

// Handles the next round of game.
function checkNextPlayer() {
  if (humanTurn()) {
    board.draw(board.mousePos);
    board.enable = true;
    setDisabled('.ctrl-game input', false);
    setDisabled('.ctrl-analyze input', false);
    if (board.playNo === 0 || (board.playNo === 1 && player.black === 'computer'))
      document.getElementById('ctrl-undo').disabled = true;

  } else {
    post(null, 'think').then(response => {
      putAiPoint(response.row, response.col);
      if (checkWinner(response)) return;
      checkNextPlayer();
    }).catch(() => alert('think failed'));
  }

  timer[board.whoTurn()].start();
}

function checkWinner(response) {
  if (response.winner !== -1) {
    notifyWinner(response.winner === 0 ? 'Black' : 'White');
    return true;
  }
  return false;
}

function putAiPoint(row, col) {
  if (row === -1) {
    board.pass();
    alert('computer pass');
  } else {
    board.play([col, row]);
  }

  if (col === board.mousePos[0] && row === board.mousePos[1])
    board.mousePos = [-1, -1];
}

function humanTurn() {
  return player[board.whoTurn()] === 'human';
}

function btnPlayNumberClick() {
  board.changePlayNumber();
  document.getElementById('play-number-check').classList.toggle('hidden');
}

function btnCoordinateClick() {
  board.changeCoordinate();
  document.getElementById('coordinate-check').classList.toggle('hidden');
}

function changeDisplayNo(changeAmount) {
  board.changeDisplayNo(changeAmount);
}

function undo() {
  setDisabled('.ctrl-game input', true);
  setDisabled('.ctrl-analyze input', true);
  const times = (player[board.whoTurn(board.playNo + 1)] === 'human' ? 1 : 2);
  board.undo(times);

  post({ times }, 'undo').then(() => {
    checkNextPlayer();
  }).catch(() => alert('undo failed'));
}

function pass() {
  setDisabled('.ctrl-game input', true);
  setDisabled('.ctrl-analyze input', true);
  board.pass();
  post(null, 'pass').then(() => {
    checkNextPlayer();
  }).catch(() => alert('pass failed'));
}

function resign() {
  setDisabled('.ctrl-game input', true);
  setDisabled('.ctrl-analyze input', true);
  timer[board.whoTurn()].stop();
  post(null, 'resign').then(() => {
    notifyWinner(board.whoTurn(board.playNo + 1));
  }).catch(() => alert('resign failed'));
}

function hint() {
  setDisabled('.ctrl-game input', true);
  setDisabled('.ctrl-analyze input', true);
  board.enable = false;

  post(null, 'think').then(response => {
    putAiPoint(response.row, response.col);
    if (checkWinner(response)) return;
    checkNextPlayer();
  }).catch(() => alert('think failed'));
}

let analyzing = false;

function analyzeClick() {
  analyzing = !analyzing;

  document.getElementById('analyze-check').classList.toggle('hidden');

  document.querySelectorAll('.player-information').forEach(el => {
    el.style.display = analyzing ? 'none' : '';
  });

  const treeViz = document.getElementById('tree-visualize');
  treeViz.style.display = analyzing ? 'block' : 'none';

  if (analyzing) {
    document.querySelectorAll('.board').forEach(el => el.classList.add('aside'));
    D3.requestTree();
  } else {
    document.querySelectorAll('.board').forEach(el => el.classList.remove('aside'));
    D3.removeTree();
  }
}

function SessionManager() {
  let sessionID;

  function getSessionID() {
    return sessionID || generateSessionID();
  }

  function generateSessionID() {
    const charset = 'abcdefghijklmnopqrstuvwxyz0123456789';
    sessionID = '';
    for (let i = 0; i < 10; ++i)
      sessionID += charset.charAt(Math.floor(Math.random() * charset.length));
    return sessionID;
  }

  return { getSessionID };
}

function refresh() {
  D3.requestTree();
}

// Keep alive when game is in progress.
function keepAlive(prevResponse) {
  if (board.gameStarted) {
    post(null, 'keepAlive').then(() => {
      setTimeout(keepAlive, 15000, true);
    }).catch(() => {
      if (prevResponse) {
        alert(
          'Sorry, there are some issues with your game.\n' +
          '1. Check your internet connection.\n' +
          '2. Your game may be terminated by the server due to long period of inactivity.'
        );
      }
      setTimeout(keepAlive, 15000, false);
    });
  }
}

setTimeout(keepAlive, 15000, true);

// Quit game when user leaves.
window.addEventListener('unload', () => post(null, '/quit'));

window.addEventListener('beforeunload', (e) => {
  e.preventDefault();
  e.returnValue = '';
});
