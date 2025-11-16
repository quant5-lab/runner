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

  calculateUnrealizedProfit(trade, currentPrice) {
    if (trade.status !== 'open' || !currentPrice) return 0;
    const multiplier = trade.direction === 'long' ? 1 : -1;
    return (currentPrice - trade.entryPrice) * trade.size * multiplier;
  }

  formatTrade(trade, index, currentPrice) {
    const isOpen = trade.status === 'open';
    const unrealizedProfit = this.calculateUnrealizedProfit(trade, currentPrice);
    
    return {
      number: index + 1,
      date: this.getTradeDate(trade),
      direction: trade.direction,
      entryPrice: this.formatPrice(trade.entryPrice),
      exitPrice: isOpen ? this.formatPrice(currentPrice) : this.formatPrice(trade.exitPrice),
      size: trade.size.toFixed(2),
      profit: isOpen ? this.formatProfit(unrealizedProfit) : this.formatProfit(trade.profit),
      profitRaw: isOpen ? unrealizedProfit : trade.profit,
      isOpen: isOpen,
    };
  }
}

/* Trade table HTML renderer (SRP, KISS) */
export class TradeTableRenderer {
  constructor(formatter) {
    this.formatter = formatter;
  }

  renderRows(trades, currentPrice) {
    return trades
      .map((trade, index) => {
        const formatted = this.formatter.formatTrade(trade, index, currentPrice);
        const directionClass =
          formatted.direction === 'long' ? 'trade-long' : 'trade-short';
        const profitClass = formatted.isOpen
          ? formatted.profitRaw >= 0
            ? 'trade-profit-positive'
            : 'trade-profit-negative'
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
