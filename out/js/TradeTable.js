/* Trade data formatting (SRP, DRY) */
export class TradeDataFormatter {
  constructor(candlestickData) {
    this.candlestickData = candlestickData || [];
  }

  formatDate(timestamp) {
    const date = new Date(timestamp);
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  formatPrice(price) {
    return `$${price.toFixed(2)}`;
  }

  formatProfit(profit) {
    const formatted = `$${Math.abs(profit).toFixed(2)}`;
    return profit >= 0 ? `+${formatted}` : `-${formatted}`;
  }

  getTradeDate(trade) {
    if (trade.entryTime) {
      return this.formatDate(trade.entryTime);
    }
    if (trade.entryBar < this.candlestickData.length) {
      const timestamp = this.candlestickData[trade.entryBar].time * 1000;
      return this.formatDate(timestamp);
    }
    return 'N/A';
  }

  formatTrade(trade, index) {
    const isOpen = trade.status === 'open';
    return {
      number: index + 1,
      date: this.getTradeDate(trade),
      direction: trade.direction,
      entryPrice: this.formatPrice(trade.entryPrice),
      exitPrice: isOpen ? 'OPEN' : this.formatPrice(trade.exitPrice),
      size: trade.size.toFixed(2),
      profit: isOpen ? 'OPEN' : this.formatProfit(trade.profit),
      profitRaw: isOpen ? 0 : trade.profit,
      isOpen: isOpen,
    };
  }
}

/* Trade table HTML renderer (SRP, KISS) */
export class TradeTableRenderer {
  constructor(formatter) {
    this.formatter = formatter;
  }

  renderRows(trades) {
    return trades
      .map((trade, index) => {
        const formatted = this.formatter.formatTrade(trade, index);
        const directionClass =
          formatted.direction === 'long' ? 'trade-long' : 'trade-short';
        const profitClass = formatted.isOpen
          ? 'trade-open'
          : formatted.profitRaw >= 0
          ? 'trade-profit-positive'
          : 'trade-profit-negative';

        return `
          <tr>
            <td>${formatted.number}</td>
            <td>${formatted.date}</td>
            <td class="${directionClass}">${formatted.direction.toUpperCase()}</td>
            <td>${formatted.entryPrice}</td>
            <td>${formatted.exitPrice}</td>
            <td>${formatted.size}</td>
            <td class="${profitClass}">${formatted.profit}</td>
          </tr>
        `;
      })
      .join('');
  }
}
