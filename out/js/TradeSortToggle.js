import { SortDirection } from './SortDirection.js';

export class TradeSortToggle {
  #direction;
  #onChange;
  #button = null;

  constructor(onChange, initialDirection = SortDirection.DESC) {
    this.#direction = initialDirection;
    this.#onChange  = onChange;
  }

  get direction() {
    return this.#direction;
  }

  mount(headerEl, hasAnyTrades) {
    if (!this.#button) {
      this.#button = document.createElement('button');
      this.#button.className = 'sort-toggle-btn';
      this.#button.addEventListener('click', () => this.#toggle());
      headerEl.appendChild(this.#button);
    }
    this.#button.style.display = hasAnyTrades ? '' : 'none';
    this.#syncLabel();
  }

  #toggle() {
    this.#direction = this.#direction === SortDirection.DESC
      ? SortDirection.ASC
      : SortDirection.DESC;
    this.#syncLabel();
    this.#onChange(this.#direction);
  }

  #syncLabel() {
    if (!this.#button) return;
    const arrow = this.#direction === SortDirection.DESC ? '▼' : '▲';
    this.#button.textContent = `${arrow} ${this.#direction === SortDirection.DESC ? 'Newest first' : 'Oldest first'}`;
  }
}
