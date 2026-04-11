export class TradeRowspanRenderer {
  #directionClass(row) {
    return row.direction === 'long' ? 'trade-long' : 'trade-short';
  }

  #profitClass(row) {
    if (row.isOpen) return 'trade-open';
    return row.profitRaw >= 0 ? 'trade-profit-positive' : 'trade-profit-negative';
  }

  #openExitRowAttr(row) {
    return row.isExitRow() && row.isOpen ? ' class="trade-open"' : '';
  }

  #entryRowHtml(row) {
    return [
      `<td rowspan="2" class="${this.#directionClass(row)}">${row.direction.toUpperCase()}</td>`,
      `<td>Entry</td>`,
      `<td>${row.dateTime}</td>`,
      `<td>${row.signal}</td>`,
      `<td>${row.price}</td>`,
      `<td rowspan="2">${row.size}</td>`,
    ].join('');
  }

  #exitRowHtml(row) {
    return [
      `<td>Exit</td>`,
      `<td>${row.dateTime}</td>`,
      `<td>${row.signal}</td>`,
      `<td>${row.price}</td>`,
      `<td class="${this.#profitClass(row)}">${row.profitLoss}</td>`,
    ].join('');
  }

  renderRow(row) {
    const cells = row.isEntryRow() ? this.#entryRowHtml(row) : this.#exitRowHtml(row);
    return `<tr${this.#openExitRowAttr(row)}>${cells}</tr>`;
  }

  renderRows(rows) {
    return rows.map(row => this.renderRow(row)).join('\n');
  }
}
