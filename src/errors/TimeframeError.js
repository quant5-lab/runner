/* Custom error for invalid timeframe on FOUND symbol
 * Signals: symbol exists in provider but timeframe not supported
 * Provider chain behavior: STOP execution, do NOT continue to next provider */
export class TimeframeError extends Error {
  constructor(timeframe, symbol, providerName) {
    const message = `Timeframe '${timeframe}' not supported for symbol '${symbol}' by provider ${providerName}`;
    super(message);
    this.name = 'TimeframeError';
    this.timeframe = timeframe;
    this.symbol = symbol;
    this.providerName = providerName;
  }
}
