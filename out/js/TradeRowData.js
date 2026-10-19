/**
 * `isPrimary` marks the first row of the pair; only the primary row carries
 * Direction[rowspan=2] and Size[rowspan=2] cells.
 * P/L is always on the exit row regardless of pair position.
 */
export class TradeRowData {
  constructor(config) {
    this.tradeNumber = config.tradeNumber;
    this.rowType     = config.rowType;
    this.isPrimary   = config.isPrimary;
    this.dateTime    = config.dateTime;
    this.signal      = config.signal;
    this.price       = config.price;
    this.size        = config.size;
    this.profitLoss  = config.profitLoss;
    this.direction   = config.direction;
    this.isOpen      = config.isOpen;
    this.profitRaw   = config.profitRaw;
  }

  isEntryRow() { return this.rowType === 'entry'; }
  isExitRow()  { return this.rowType === 'exit';  }
}
