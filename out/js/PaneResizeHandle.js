import { PaneResizeCalculator } from './PaneResizeCalculator.js';

/**
 * Call destroy() when the associated panes are torn down to remove the element
 * and deregister all document-level listeners.
 *
 * Pane shape: { container: HTMLElement, chart: IChartApi }
 */
export class PaneResizeHandle {
  #above;
  #below;
  #element;
  #drag = null;

  #onMouseMove = (e) => this.#move(e.clientY);
  #onTouchMove = (e) => { if (this.#drag) { e.preventDefault(); this.#move(e.touches[0].clientY); } };
  #onDragEnd   = ()  => { this.#drag = null; };

  constructor(above, below) {
    this.#above   = above;
    this.#below   = below;
    this.#element = this.#buildElement();

    document.addEventListener('mousemove', this.#onMouseMove);
    document.addEventListener('touchmove', this.#onTouchMove, { passive: false });
    document.addEventListener('mouseup',   this.#onDragEnd);
    document.addEventListener('touchend',  this.#onDragEnd);
  }

  get element() { return this.#element; }

  destroy() {
    this.#element.remove();
    document.removeEventListener('mousemove', this.#onMouseMove);
    document.removeEventListener('touchmove', this.#onTouchMove);
    document.removeEventListener('mouseup',   this.#onDragEnd);
    document.removeEventListener('touchend',  this.#onDragEnd);
  }

  #buildElement() {
    const el = document.createElement('div');
    el.className = 'pane-resize-handle';
    el.addEventListener('mousedown',  (e) => this.#startDrag(e.clientY));
    el.addEventListener('touchstart', (e) => { e.preventDefault(); this.#startDrag(e.touches[0].clientY); }, { passive: false });
    return el;
  }

  #startDrag(y) {
    this.#drag = {
      startY: y,
      aboveH: this.#above.container.clientHeight,
      belowH: this.#below.container.clientHeight,
    };
  }

  #move(y) {
    if (!this.#drag) return;
    const delta = y - this.#drag.startY;
    const { above, below } = PaneResizeCalculator.calculate(
      this.#drag.aboveH, this.#drag.belowH, delta,
    );
    this.#applyHeight(this.#above, above);
    this.#applyHeight(this.#below, below);
  }

  #applyHeight({ container, chart }, height) {
    container.style.height = `${height}px`;
    chart.resize(container.clientWidth, height);
  }
}
