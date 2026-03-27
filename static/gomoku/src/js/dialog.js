class Dialog {
  constructor() {
    this.dialog = document.getElementById('dialog-new-game');

    document.getElementById('dl-ok').addEventListener('click', () => this.ok());
    document.getElementById('dl-cancel').addEventListener('click', () => this.toggle());
    document.getElementById('new-game').addEventListener('click', () => this.toggle());
  }

  ok() {
    const black = document.querySelector('input[name="black"]:checked').value;
    const white = document.querySelector('input[name="white"]:checked').value;

    const nameBlack = document.getElementById('dl-name-black');
    const nameWhite = document.getElementById('dl-name-white');
    if (nameBlack.value.trim() === '') nameBlack.value = 'you';
    if (nameWhite.value.trim() === '') nameWhite.value = 'you';

    player = { black, white };

    const rule = parseInt(document.querySelector('input[name="rule"]:checked').value, 10);
    const level = parseInt(document.getElementById('dl-select').value, 10);

    post({ rule, level }, 'start').catch(err => {
      alert('start failed');
      throw err;
    }).then(() => {
      if (black === 'human') {
        this.initPlayer('sel', 'black', true);
        this.initPlayer('opp', 'white', white === 'human');
      } else {
        if (white === 'human') {
          this.initPlayer('sel', 'white', true);
          this.initPlayer('opp', 'black', false);
        } else {
          this.initPlayer('sel', 'black', false);
          this.initPlayer('opp', 'white', false);
        }
      }

      board.init();

      if (timer.black != null) timer.black.stop();
      if (timer.white != null) timer.white.stop();
      timer = {
        black: new Timer(document.querySelector('.black .pi-timer')),
        white: new Timer(document.querySelector('.white .pi-timer'))
      };

      setDisabled('.control input', true);
      setDisabled('.ctrl-replay input', true);

      this.toggle();

      game.rule = rule;
      game.earthmover.level = level;
      game.black = black;
      game.white = white;
      game.startTime = Date.now();

      checkNextPlayer();
    });
  }

  initPlayer(playerPos, color, human) {
    const oppColor = color === 'black' ? 'white' : 'black';
    const el = document.querySelector(`.player-information.${playerPos}`);
    el.classList.remove(oppColor);
    el.classList.add(color);

    document.getElementById(`pi-chess-${playerPos}`).src = `gomoku/src/png/chess_${color}.png`;
    document.getElementById(`pi-icon-${playerPos}`).src = `gomoku/src/png/${human ? 'human' : 'icon'}.png`;
    document.getElementById(`pi-name-${playerPos}`).textContent =
      human ? document.getElementById(`dl-name-${color}`).value : 'EarthMover';
  }

  toggle() {
    const current = this.dialog.style.display;
    this.dialog.style.display = (current === 'none' || current === '') ? 'block' : 'none';
  }
}
