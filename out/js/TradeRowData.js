/**
 * TradeRowData - Domain model for rowspan table rows
 * 
 * SRP: Represents a single visual row (Entry or Exit) in the rowspan table
 * Each trade produces TWO rows: one Entry row + one Exit row
 */
export class TradeRowData {
  constructor(config) {
    this.tradeNumber = config.tradeNumber;
    this.rowType = config.rowType; // 'entry' | 'exit'
    this.dateTime = config.dateTime;
    this.signal = config.signal;
    this.price = config.price;
    this.size = config.size;
    this.profitLoss = config.profitLoss;
    this.direction = config.direction;
    this.isOpen = config.isOpen;
    this.profitRaw = config.profitRaw;
  }

  isEntryRow() {
    return this.rowType === 'entry';
  }

  isExitRow() {
    return this.rowType === 'exit';
  }
}
