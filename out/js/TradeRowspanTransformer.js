import { TradeRowData } from './TradeRowData.js';

export class TradeRowspanTransformer {
  constructor(formatter) {
    this.formatter = formatter;
  }

  transformTrade(trade, tradeNumber, currentPrice) {
    const isOpen = trade.status === 'open';
    const unrealizedProfit = this.formatter.calculateUnrealizedProfit(trade, currentPrice);

    const exitPrice = isOpen
      ? (currentPrice ?? trade.entryPrice)
      : (trade.exitPrice ?? 0);

    const profitValue = isOpen ? unrealizedProfit : trade.profit;

    const entryRow = new TradeRowData({
      tradeNumber,
      rowType: 'entry',
      dateTime: this.formatter.getTradeDate(trade, true),
      signal: trade.entryComment || trade.entryId || '',
      price: this.formatter.formatPrice(trade.entryPrice),
      size: trade.size.toFixed(2),
      profitLoss: '',
      direction: trade.direction,
      isOpen: false,
      profitRaw: 0,
    });

    const exitRow = new TradeRowData({
      tradeNumber,
      rowType: 'exit',
      dateTime: isOpen ? 'Open' : this.formatter.getTradeDate(trade, false),
      signal: isOpen ? '' : (trade.exitComment || trade.exitId || ''),
      price: this.formatter.formatPrice(exitPrice),
      size: '',
      profitLoss: this.formatter.formatProfit(profitValue),
      direction: trade.direction,
      isOpen,
      profitRaw: profitValue,
    });

    return [entryRow, exitRow];
  }

  transformTrades(trades, currentPrice) {
    const rows = [];
    trades.forEach((trade, index) => {
      const [entryRow, exitRow] = this.transformTrade(trade, index + 1, currentPrice);
      rows.push(entryRow, exitRow);
    });
    return rows;
  }
}
