export function formatCurrency(value) {
  if (value == null || !isFinite(value)) return '$0.00';
  if (value === 0) return '$0.00';
  const abs = value < 0 ? -value : value;
  const formatted = abs.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
  return value > 0 ? '+$' + formatted : '-$' + formatted;
}

export class TradeDataFormatter {
  constructor(candlestickData) {
    this.candlestickData = candlestickData || [];
  }

  formatDate(timestamp) {
    return new Date(timestamp).toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  formatPrice(price) {
    if (price == null) return '$0.00';
    return '$' + price.toFixed(2);
  }

  formatProfit(profit) {
    return formatCurrency(profit);
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
      if (bar?.time !== undefined) {
        return this.formatDate(bar.time * 1000);
      }
    }

    return 'N/A';
  }

  calculateUnrealizedProfit(trade, currentPrice) {
    if (trade.status !== 'open' || !currentPrice) return 0;
    const multiplier = trade.direction === 'long' ? 1 : -1;
    return (currentPrice - trade.entryPrice) * trade.size * multiplier;
  }
}
