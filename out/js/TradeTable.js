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
    if (price === null || price === undefined) return '$0.00';
    return `$${price.toFixed(2)}`;
  }

  formatProfit(profit) {
    const formatted = `$${Math.abs(profit).toFixed(2)}`;
    return profit >= 0 ? `+${formatted}` : `-${formatted}`;
  }

  getTradeDate(trade, isEntry = true) {
    const timeField = isEntry ? 'entryTime' : 'exitTime';
    const barField = isEntry ? 'entryBar' : 'exitBar';
    
    if (trade[timeField]) {
      return this.formatDate(trade[timeField] * 1000);
    }
    
    const barIndex = trade[barField];
    if (barIndex !== undefined && barIndex >= 0 && barIndex < this.candlestickData.length) {
      const bar = this.candlestickData[barIndex];
      if (bar && bar.time !== undefined) {
        const timestamp = bar.time * 1000;
        return this.formatDate(timestamp);
      }
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
    
    const exitPrice = isOpen 
      ? (currentPrice !== null && currentPrice !== undefined ? currentPrice : trade.entryPrice)
      : (trade.exitPrice !== null && trade.exitPrice !== undefined ? trade.exitPrice : 0);
    
    return {
      number: index + 1,
      entryDate: this.getTradeDate(trade, true),
      entryBar: trade.entryBar !== undefined ? trade.entryBar : 'N/A',
      exitDate: isOpen ? 'Open' : this.getTradeDate(trade, false),
      exitBar: isOpen ? '-' : (trade.exitBar !== undefined ? trade.exitBar : 'N/A'),
      direction: trade.direction,
      entryPrice: this.formatPrice(trade.entryPrice),
      exitPrice: this.formatPrice(exitPrice),
      size: trade.size.toFixed(2),
      profit: isOpen ? this.formatProfit(unrealizedProfit) : this.formatProfit(trade.profit),
      profitRaw: isOpen ? unrealizedProfit : trade.profit,
      entryId: trade.entryId || trade.entryID || 'N/A',
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
            <td>${formatted.entryDate}</td>
            <td>${formatted.entryBar}</td>
            <td class="${directionClass}">${formatted.direction.toUpperCase()}</td>
            <td>${formatted.entryPrice}</td>
            <td>${formatted.exitDate}</td>
            <td>${formatted.exitBar}</td>
            <td>${formatted.exitPrice}</td>
            <td>${formatted.size}</td>
            <td class="${profitClass}">${formatted.profit}</td>
            <td>${formatted.entryId}</td>
          </tr>
        `;
      })
      .join('');
  }
}
