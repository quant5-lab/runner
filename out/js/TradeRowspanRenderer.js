/**
 * TradeRowspanRenderer — stateless HTML generator for the rowspan trade table.
 *
 * Layout contract (7 columns: Type | Entry/Exit | DateTime | Signal | Price | Size | P/L):
 *
 *   Primary row   (isPrimary=true):
 *     Type[rowspan=2] | label | DateTime | Signal | Price | Size[rowspan=2] | P/L (exit only)
 *
 *   Secondary row (isPrimary=false):
 *     label | DateTime | Signal | Price | P/L (exit only)
 *
 * P/L is always and only on exit rows, regardless of pair position.
 * `.trade-open` class is applied to the <tr> and the P/L <td> of open exit rows.
 */
export class TradeRowspanRenderer {
  #directionClass(row) {
    return row.direction === 'long' ? 'trade-long' : 'trade-short';
  }

  #profitClass(row) {
    if (row.isOpen) return 'trade-open';
    return row.profitRaw >= 0 ? 'trade-profit-positive' : 'trade-profit-negative';
  }

  #label(row) {
    return row.isEntryRow() ? 'Entry' : 'Exit';
  }

  #plCell(row) {
    if (!row.isExitRow()) return '';
    return `<td class="${this.#profitClass(row)}">${row.profitLoss}</td>`;
  }

  #primaryRowHtml(row) {
    return [
      `<td rowspan="2" class="${this.#directionClass(row)}">${row.direction.toUpperCase()}</td>`,
      `<td>${this.#label(row)}</td>`,
      `<td>${row.dateTime}</td>`,
      `<td>${row.signal}</td>`,
      `<td>${row.price}</td>`,
      `<td rowspan="2">${row.size}</td>`,
      this.#plCell(row),
    ].join('');
  }

  #secondaryRowHtml(row) {
    return [
      `<td>${this.#label(row)}</td>`,
      `<td>${row.dateTime}</td>`,
      `<td>${row.signal}</td>`,
      `<td>${row.price}</td>`,
      this.#plCell(row),
    ].join('');
  }

  #trClass(row) {
    return row.isExitRow() && row.isOpen ? ' class="trade-open"' : '';
  }

  renderRow(row) {
    const cells = row.isPrimary
      ? this.#primaryRowHtml(row)
      : this.#secondaryRowHtml(row);
    return `<tr${this.#trClass(row)}>${cells}</tr>`;
  }

  renderRows(rows) {
    return rows.map(row => this.renderRow(row)).join('\n');
  }
}
