/* Build timestamp→bar index mapping for candlestick data */
export class TimeIndexBuilder {
  build(candlestickData) {
    const timeIndex = new Map();
    candlestickData.forEach((candle, idx) => {
      timeIndex.set(candle.time, idx);
    });
    return timeIndex;
  }
}
