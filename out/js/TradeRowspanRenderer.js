/**
 * TradeRowspanRenderer - Generates HTML for rowspan table structure
 * 
 * SRP: Single responsibility - HTML generation for rowspan cells
 * KISS: Simple row generation logic
 * 
 * Rowspan Structure:
 * Row 1 (Entry): Type[rowspan=2] | Entry Label | Entry Date | Entry Signal | Entry Price | Size[rowspan=2] | empty
 * Row 2 (Exit):                  | Exit Label  | Exit Date  | Exit Signal  | Exit Price  |                 | P/L[rowspan=2]
 */
export class TradeRowspanRenderer {
  constructor() {}

  /**
   * Render single TradeRowData as HTML row with rowspan cells
   */
  renderRow(row) {
    const directionClass = row.direction === 'long' ? 'trade-long' : 'trade-short';
    const profitClass = row.isOpen
      ? (row.profitRaw >= 0 ? 'trade-profit-positive' : 'trade-profit-negative')
      : (row.profitRaw >= 0 ? 'trade-profit-positive' : 'trade-profit-negative');

    let html = '<tr>';

    if (row.isEntryRow()) {
      // Entry row: Type (rowspan=2), Entry label, date, signal, price, Size (rowspan=2)
      html += `<td rowspan="2" class="${directionClass}">${row.direction.toUpperCase()}</td>`;
      html += `<td>Entry</td>`;
      html += `<td>${row.dateTime}</td>`;
      html += `<td>${row.signal}</td>`;
      html += `<td>${row.price}</td>`;
      html += `<td rowspan="2">${row.size}</td>`;
    } else {
      // Exit row: Exit label, date, signal, price, P/L
      html += `<td>Exit</td>`;
      html += `<td>${row.dateTime}</td>`;
      html += `<td>${row.signal}</td>`;
      html += `<td>${row.price}</td>`;
      html += `<td class="${profitClass}">${row.profitLoss}</td>`;
    }

    html += '</tr>';
    return html;
  }

  /**
   * Render array of TradeRowData as complete HTML
   */
  renderRows(rows) {
    return rows.map(row => this.renderRow(row)).join('\n');
  }
}
