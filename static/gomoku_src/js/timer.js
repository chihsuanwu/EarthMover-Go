class Timer {
  constructor(span) {
    this.span = span;
    this.min = { ten: 0, one: 0 };
    this.sec = { ten: 0, one: 0 };
    this.counting = null;
    for (let i = 0; i < 4; ++i) {
      const wrapper = span.children[i];
      wrapper.querySelectorAll('.flip-top, .flip-bottom').forEach(el => el.textContent = '0');
      wrapper.querySelectorAll('.flip-next, .flip-back').forEach(el => el.textContent = '1');
      wrapper.querySelector('.flip-top').classList.remove('top-ani');
      wrapper.querySelector('.flip-back').classList.remove('back-ani');
    }
  }

  set(min, sec) {
    if (this.sec.one !== sec % 10) {
      this.sec.one = sec % 10;
      this.animate(this.span.children[3], this.sec.one, this.sec.one === 9 ? 0 : this.sec.one + 1);
    }
    if (this.sec.ten !== Math.floor(sec / 10)) {
      this.sec.ten = Math.floor(sec / 10);
      this.animate(this.span.children[2], this.sec.ten, this.sec.ten === 5 ? 0 : this.sec.ten + 1);
    }
    if (this.min.one !== min % 10) {
      this.min.one = min % 10;
      this.animate(this.span.children[1], this.min.one, this.min.one === 9 ? 0 : this.min.one + 1);
    }
    if (this.min.ten !== Math.floor(min / 10)) {
      this.min.ten = Math.floor(min / 10);
      this.animate(this.span.children[0], this.min.ten, this.min.ten === 5 ? 0 : this.min.ten + 1);
    }
  }

  stop() {
    clearInterval(this.counting);
  }

  start() {
    this.counting = setInterval(() => this.add(), 1000);
  }

  add() {
    ++this.sec.one;
    if (this.sec.one === 10) {
      this.sec.one = 0;
      ++this.sec.ten;
      if (this.sec.ten === 6) {
        this.sec.ten = 0;
        ++this.min.one;
        if (this.min.one === 10) {
          this.min.one = 0;
          ++this.min.ten;
          this.animate(this.span.children[0], this.min.ten, this.min.ten === 5 ? 0 : this.min.ten + 1);
        }
        this.animate(this.span.children[1], this.min.one, this.min.one === 9 ? 0 : this.min.one + 1);
      }
      this.animate(this.span.children[2], this.sec.ten, this.sec.ten === 5 ? 0 : this.sec.ten + 1);
    }
    this.animate(this.span.children[3], this.sec.one, this.sec.one === 9 ? 0 : this.sec.one + 1);
  }

  animate(wrapper, num, next) {
    const flipTop = wrapper.querySelector('.flip-top');
    const flipBack = wrapper.querySelector('.flip-back');

    flipTop.classList.add('top-ani');
    flipBack.classList.add('back-ani');

    flipBack.addEventListener('animationend', function handler() {
      flipBack.removeEventListener('animationend', handler);
      wrapper.querySelectorAll('.flip-top, .flip-bottom').forEach(el => el.textContent = num);

      requestAnimationFrame(() => {
        flipTop.classList.remove('top-ani');
        flipBack.classList.remove('back-ani');

        requestAnimationFrame(() => {
          wrapper.querySelectorAll('.flip-next, .flip-back').forEach(el => el.textContent = next);
        });
      });
    });
  }
}
