class Board {
  constructor() {
    this.enable = false;
    this.gameStarted = false;
    this.display = { playNumber: false, coordinate: false };
    this.mousePos = [-1, -1];
    this.lastPlay = [-1, -1];

    this.canvas = document.getElementById('cvs');
    this.context = this.canvas.getContext('2d');
    this.canvas.width = 565;
    this.canvas.height = 565;

    this.canvas.addEventListener('click', this.click.bind(this));
    this.canvas.addEventListener('mousemove', this.mouseMoveOrOver.bind(this));
    this.canvas.addEventListener('mouseover', this.mouseMoveOrOver.bind(this));
    this.canvas.addEventListener('mouseout', this.mouseOut.bind(this));

    this.playNo = 0;
    this.displayNo = 0;

    // board status array
    this.status = Array.from({ length: 15 }, () => new Array(15).fill(0));

    const createImage = (src) => {
      const image = new Image();
      image.src = `gomoku/src/png/chess_${src}.png`;
      return image;
    };

    this.image = {
      black: {
        normal: createImage('black'),
        transparent: createImage('black_transparent'),
        marked: createImage('black_marked')
      },
      white: {
        normal: createImage('white'),
        transparent: createImage('white_transparent'),
        marked: createImage('white_marked')
      }
    };
  }

  init() {
    this.playNo = 0;
    this.displayNo = 0;

    for (let row = 0; row < 15; row++)
      for (let col = 0; col < 15; col++)
        this.status[row][col] = 0;

    this.mousePos = [-1, -1];
    this.lastPlay = [-1, -1];
    this.context.clearRect(20, 20, 525, 525);
    this.enable = false;
    this.gameStarted = true;
  }

  mouseMoveOrOver(event) {
    const rect = this.canvas.getBoundingClientRect();
    const scaling = this.canvas.scrollWidth / 565;

    let x = Math.floor((event.clientX - rect.left - 20 * scaling) / (35 * scaling));
    let y = Math.floor((event.clientY - rect.top - 20 * scaling) / (35 * scaling));

    if (x < 0 || x > 14 || y < 0 || y > 14) {
      x = -1; y = -1;
    }

    const pos = [x, y];

    if (this.mousePos[0] === pos[0] && this.mousePos[1] === pos[1]) return;

    if (pos[0] !== -1 && !this.status[pos[0]][pos[1]]) {
      if (this.enable) {
        this.clear(this.mousePos);
        this.draw(pos);
      }
      this.mousePos = pos.slice();
    } else {
      if (this.enable) this.clear(this.mousePos);
      this.mousePos = [-1, -1];
    }
  }

  mouseOut() {
    if (this.mousePos[0] !== -1) {
      this.clear(this.mousePos);
      this.mousePos = [-1, -1];
    }
  }

  click() {
    if (!this.gameStarted) {
      dialog.toggle();
      return;
    }

    if (!this.enable) return;
    if (this.mousePos[0] === -1) return;

    if (!this.status[this.mousePos[0]][this.mousePos[1]]) {
      this.play(this.mousePos);
      post({ row: this.mousePos[1], col: this.mousePos[0] }, 'play').then(
        response => {
          if (checkWinner(response)) return;
          checkNextPlayer();
        }).catch(() => alert('play failed'));

      this.mousePos = [-1, -1];
    }
  }

  draw(pos) {
    if (pos[0] === -1) return;

    const status = this.status[pos[0]][pos[1]];
    let image;
    if (status === 0) {
      image = this.image[this.whoTurn()].transparent;
    } else if (status === this.displayNo) {
      image = this.display.playNumber
        ? this.image[this.whoTurn(status - 1)].normal
        : this.image[this.whoTurn(status - 1)].marked;
    } else {
      image = this.image[this.whoTurn(status - 1)].normal;
    }

    this.drawChess(image, pos[0], pos[1]);

    if (this.display.playNumber && status > 0) {
      const color = status === this.displayNo ? 'red' : ((status - 1) & 1) ? 'black' : 'white';
      this.drawNumber(status, color, pos[0], pos[1]);
    }
  }

  drawChess(image, x, y) {
    this.context.drawImage(image, x * 35 + 21, y * 35 + 21, 33, 33);
  }

  drawNumber(number, color, x, y) {
    this.context.fillStyle = color;
    this.context.font = '29px Ubuntu';
    this.context.textAlign = 'center';
    this.context.fillText(number, x * 35 + 37, y * 35 + 47, 27);
  }

  clear(pos) {
    if (pos[0] === -1) return;
    this.context.clearRect(pos[0] * 35 + 21, pos[1] * 35 + 21, 33, 33);
  }

  play(pos) {
    this.enable = false;

    setDisabled('.ctrl-game input', true);
    setDisabled('.ctrl-analyze input', true);

    timer[this.whoTurn()].stop();

    ++this.playNo;
    ++this.displayNo;
    this.status[pos[0]][pos[1]] = this.playNo;

    this.draw(pos);

    this.clear(this.lastPlay);
    this.draw(this.lastPlay);

    this.lastPlay = pos.slice(0);
  }

  pass() {
    this.enable = false;
    timer[this.whoTurn()].stop();
    ++this.playNo;
    ++this.displayNo;
  }

  undo(times) {
    this.enable = false;
    timer[this.whoTurn()].stop();
    this.playNo -= times;
    this.displayNo = this.playNo;

    for (let row = this.status.length - 1; row >= 0; row--)
      for (let col = this.status[row].length - 1; col >= 0; col--) {
        if (this.status[row][col] > this.playNo) {
          this.status[row][col] = 0;
          this.clear([row, col]);
        } else if (this.playNo > 0 && this.status[row][col] === this.playNo) {
          this.lastPlay = [row, col];
          this.clear(this.lastPlay);
          this.draw(this.lastPlay);
        }
      }
  }

  whoTurn(param) {
    const num = param === undefined ? this.playNo : param;
    return (num & 1) ? 'white' : 'black';
  }

  changePlayNumber() {
    this.display.playNumber = !this.display.playNumber;
    this.drawAll();
  }

  changeCoordinate() {
    this.display.coordinate = !this.display.coordinate;

    if (this.display.coordinate) {
      this.context.fillStyle = '#444';
      this.context.font = '12px Ubuntu';
      this.context.textAlign = 'center';
      for (let i = 1; i <= 15; ++i) {
        const text = String.fromCharCode(64 + i);
        this.context.fillText(text, i * 35 + 2, 15);
        this.context.fillText(text, i * 35 + 2, 558);
        this.context.fillText(i, 10, i * 35 + 7);
        this.context.fillText(i, 552, i * 35 + 7);
      }
    } else {
      this.context.clearRect(0, 20, 20, 525);
      this.context.clearRect(545, 20, 20, 525);
      this.context.clearRect(20, 0, 525, 20);
      this.context.clearRect(20, 545, 525, 20);
    }
  }

  changeDisplayNo(changeAmount) {
    this.displayNo += changeAmount;
    if (this.displayNo < 0) this.displayNo = 0;
    if (this.displayNo > this.playNo) this.displayNo = this.playNo;
    this.drawAll();
  }

  drawAll() {
    for (let row = this.status.length - 1; row >= 0; row--)
      for (let col = this.status[row].length - 1; col >= 0; col--) {
        this.clear([row, col]);
        if (this.status[row][col] > 0 && this.status[row][col] <= this.displayNo)
          this.draw([row, col]);
      }
  }

  drawEvolve(evolve) {
    const playNumber = this.display.playNumber;
    if (playNumber) {
      this.display.playNumber = false;
      this.drawAll();
    }

    for (let i = evolve.length - 1; i >= 0; i--) {
      this.drawChess(this.image[evolve[i].color].normal, evolve[i].col, evolve[i].row);
      this.drawNumber(i + 1, evolve[i].color === 'black' ? 'white' : 'black', evolve[i].col, evolve[i].row);
    }

    if (playNumber) this.display.playNumber = true;
  }
}
