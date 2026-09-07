import { TradeRowData  } from './TradeRowData.js';
import { SortDirection } from './SortDirection.js';

export class TradeRowspanTransformer {
  constructor(formatter) {
    this.formatter = formatter;
  }

  /**
   * `isPrimary` on the first row drives rowspan-cell placement in the renderer.
   * `size` is populated on both rows so the renderer always finds it on the primary.
   */
  transformTrade(trade, tradeNumber, currentPrice, sortDirection = SortDirection.DESC) {
    const isOpen           = trade.status === 'open';
    const unrealizedProfit = this.formatter.calculateUnrealizedProfit(trade, currentPrice);
    const exitPrice        = isOpen ? (currentPrice ?? trade.entryPrice) : (trade.exitPrice ?? 0);
    const profitValue      = isOpen ? unrealizedProfit : trade.profit;
    const sizeFormatted    = trade.size.toFixed(2);

    const entryRow = new TradeRowData({
      tradeNumber,
      rowType:    'entry',
      isPrimary:  sortDirection === SortDirection.ASC,
      dateTime:   this.formatter.getTradeDate(trade, true),
      signal:     trade.entryComment || trade.entryId || '',
      price:      this.formatter.formatPrice(trade.entryPrice),
      size:       sizeFormatted,
      profitLoss: '',
      direction:  trade.direction,
      isOpen:     false,
      profitRaw:  0,
    });

    const exitRow = new TradeRowData({
      tradeNumber,
      rowType:    'exit',
      isPrimary:  sortDirection === SortDirection.DESC,
      dateTime:   isOpen ? 'Open' : this.formatter.getTradeDate(trade, false),
      signal:     isOpen ? '' : (trade.exitComment || trade.exitId || ''),
      price:      this.formatter.formatPrice(exitPrice),
      size:       sizeFormatted,
      profitLoss: this.formatter.formatProfit(profitValue),
      direction:  trade.direction,
      isOpen,
      profitRaw:  profitValue,
    });

    return sortDirection === SortDirection.ASC
      ? [entryRow, exitRow]
      : [exitRow,  entryRow];
  }

  transformTrades(trades, currentPrice, sortDirection = SortDirection.DESC) {
    return trades.flatMap((trade, index) =>
      this.transformTrade(trade, index + 1, currentPrice, sortDirection)
    );
  }
}
