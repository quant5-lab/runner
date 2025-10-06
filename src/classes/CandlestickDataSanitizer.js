class CandlestickDataSanitizer {
  isValidCandle(candle) {
    const { open, high, low, close } = candle;
    const values = [open, high, low, close].map(parseFloat);

    return (
      values.every((val) => !isNaN(val) && val > 0) &&
      Math.max(...values) === parseFloat(high) &&
      Math.min(...values) === parseFloat(low)
    );
  }

  normalizeCandle(candle) {
    const open = parseFloat(candle.open);
    const high = parseFloat(candle.high);
    const low = parseFloat(candle.low);
    const close = parseFloat(candle.close);
    const volume = parseFloat(candle.volume) || 1000;

    return {
      time: Math.floor(candle.openTime / 1000),
      open,
      high: Math.max(open, high, low, close),
      low: Math.min(open, high, low, close),
      close,
      volume,
    };
  }

  processCandlestickData(rawData) {
    if (!rawData?.length) return [];

    return rawData.filter(this.isValidCandle).map(this.normalizeCandle);
  }
}

export { CandlestickDataSanitizer };
