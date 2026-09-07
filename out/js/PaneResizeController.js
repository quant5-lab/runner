import { PaneResizeHandle } from './PaneResizeHandle.js';

/**
 * mount()   — inserts N-1 handles between N panes; safe to call after destroy().
 * destroy() — removes all handles and their document listeners; safe to call
 *             before new panes are created (e.g. on refresh).
 *
 * Pane shape: { container: HTMLElement, chart: IChartApi }
 */
export class PaneResizeController {
  #handles = [];

  mount(panes) {
    for (let i = 0; i < panes.length - 1; i++) {
      const handle = new PaneResizeHandle(panes[i], panes[i + 1]);
      panes[i + 1].container.insertAdjacentElement('beforebegin', handle.element);
      this.#handles.push(handle);
    }
  }

  destroy() {
    this.#handles.forEach(h => h.destroy());
    this.#handles = [];
  }
}
