/**
 * Prefers the native Fullscreen API; falls back to a `chart-fullscreen` CSS class
 * for browsers that do not support the API on non-video elements (e.g. iPhone Safari).
 *
 * `onResize` is called on every viewport change so the chart library can reflow.
 *
 * `mount()` is idempotent — safe to call on every render cycle; attaches exactly
 * one button and one pair of `fullscreenchange` listeners for the lifetime of the instance.
 */
export class FullscreenToggle {
  #container;
  #button;
  #onResize;
  #usingApiFullscreen = false;
  #mounted            = false;
  #boundHandler       = () => this.#handleFullscreenChange();

  constructor(container, onResize) {
    this.#container = container;
    this.#onResize  = onResize;
    this.#button    = this.#createButton();
  }

  mount() {
    if (this.#mounted) return;
    this.#mounted = true;
    this.#container.appendChild(this.#button);
    document.addEventListener('fullscreenchange',       this.#boundHandler);
    document.addEventListener('webkitfullscreenchange', this.#boundHandler);
  }

  #createButton() {
    const btn = document.createElement('button');
    btn.className        = 'fullscreen-toggle-btn';
    btn.textContent      = '⤢';
    btn.title            = 'Toggle fullscreen';
    btn.setAttribute('aria-pressed', 'false');
    btn.addEventListener('click', () => this.#toggle());
    return btn;
  }

  #supportsFullscreenApi() {
    return typeof this.#container.requestFullscreen === 'function'
      || typeof this.#container.webkitRequestFullscreen === 'function';
  }

  #isFullscreen() {
    return !!(document.fullscreenElement || document.webkitFullscreenElement)
      || this.#container.classList.contains('chart-fullscreen');
  }

  #toggle() {
    if (this.#isFullscreen()) {
      this.#exit();
    } else {
      this.#enter();
    }
  }

  #enter() {
    if (this.#supportsFullscreenApi()) {
      this.#usingApiFullscreen = true;
      (this.#container.requestFullscreen || this.#container.webkitRequestFullscreen)
        .call(this.#container);
    } else {
      this.#container.classList.add('chart-fullscreen');
      this.#syncButton();
      this.#onResize();
    }
  }

  #exit() {
    if (this.#usingApiFullscreen) {
      this.#usingApiFullscreen = false;
      (document.exitFullscreen || document.webkitExitFullscreen).call(document);
    } else {
      this.#container.classList.remove('chart-fullscreen');
      this.#syncButton();
      this.#onResize();
    }
  }

  #handleFullscreenChange() {
    this.#syncButton();
    this.#onResize();
  }

  #syncButton() {
    const active = this.#isFullscreen();
    this.#button.textContent = active ? '✕' : '⤢';
    this.#button.title       = active ? 'Exit fullscreen' : 'Toggle fullscreen';
    this.#button.setAttribute('aria-pressed', String(active));
  }
}
