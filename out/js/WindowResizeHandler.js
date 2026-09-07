export class WindowResizeHandler {
  #handler;

  constructor(onResize) {
    this.#handler = onResize;
    window.addEventListener('resize', this.#handler);
  }

  destroy() {
    window.removeEventListener('resize', this.#handler);
  }
}
