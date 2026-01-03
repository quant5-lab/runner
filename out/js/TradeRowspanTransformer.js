import { TradeRowData } from './TradeRowData.js';

/**
 * TradeRowspanTransformer - Transforms Trade objects into Entry/Exit row pairs
 * 
 * SRP: Single responsibility - convert domain Trade to presentation TradeRowData pairs
 * DRY: Reuses TradeDataFormatter for date/price formatting
 * KISS: Simple transformation logic, no business logic
 */
export class TradeRowspanTransformer {
  constructor(formatter) {
    this.formatter = formatter;
  }

  /**
   * Transform single trade into [entryRow, exitRow] pair
   */
  transformTrade(trade, tradeNumber, currentPrice) {
    const isOpen = trade.status === 'open';
    const unrealizedProfit = this.formatter.calculateUnrealizedProfit(trade, currentPrice);
    
    const exitPrice = isOpen 
      ? (currentPrice !== null && currentPrice !== undefined ? currentPrice : trade.entryPrice)
      : (trade.exitPrice !== null && trade.exitPrice !== undefined ? trade.exitPrice : 0);

    const profitValue = isOpen ? unrealizedProfit : trade.profit;
    const formattedProfit = this.formatter.formatProfit(profitValue);

    // Entry row
    const entryRow = new TradeRowData({
      tradeNumber: tradeNumber,
      rowType: 'entry',
      dateTime: this.formatter.getTradeDate(trade, true),
      signal: trade.entryComment || trade.EntryComment || '',
      price: this.formatter.formatPrice(trade.entryPrice),
      size: trade.size.toFixed(2),
      profitLoss: '',
      direction: trade.direction,
      isOpen: false,
      profitRaw: 0,
    });

    // Exit row
    const exitRow = new TradeRowData({
      tradeNumber: tradeNumber,
      rowType: 'exit',
      dateTime: isOpen ? 'Open' : this.formatter.getTradeDate(trade, false),
      signal: isOpen ? '' : (trade.exitComment || trade.ExitComment || ''),
      price: this.formatter.formatPrice(exitPrice),
      size: '',
      profitLoss: formattedProfit,
      direction: trade.direction,
      isOpen: isOpen,
      profitRaw: profitValue,
    });

    return [entryRow, exitRow];
  }

  /**
   * Transform array of trades into flat array of TradeRowData
   * [trade1, trade2] → [trade1_entry, trade1_exit, trade2_entry, trade2_exit]
   */
  transformTrades(trades, currentPrice) {
    const rows = [];
    trades.forEach((trade, index) => {
      const [entryRow, exitRow] = this.transformTrade(trade, index + 1, currentPrice);
      rows.push(entryRow, exitRow);
    });
    return rows;
  }
}
