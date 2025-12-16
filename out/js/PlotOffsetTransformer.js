/* Apply PineScript plot offset to indicator timestamps
 * Offset semantics: negative shifts left (earlier), positive shifts right (later)
 */
export class PlotOffsetTransformer {
  constructor(timeIndexBuilder) {
    this.timeIndexBuilder = timeIndexBuilder;
  }

  transform(indicatorData, offset, candlestickData) {
    if (!this.shouldApplyOffset(offset, candlestickData)) {
      return indicatorData;
    }

    const timeIndex = this.timeIndexBuilder.build(candlestickData);
    return this.applyOffsetShift(indicatorData, offset, candlestickData, timeIndex);
  }

  shouldApplyOffset(offset, candlestickData) {
    return offset !== 0 && candlestickData?.length > 0;
  }

  applyOffsetShift(indicatorData, offset, candlestickData, timeIndex) {
    return indicatorData
      .map((point) => this.shiftPoint(point, offset, candlestickData, timeIndex))
      .filter((point) => point !== null);
  }

  shiftPoint(point, offset, candlestickData, timeIndex) {
    const currentBarIdx = timeIndex.get(point.time);
    if (currentBarIdx === undefined) {
      return null;
    }

    const targetBarIdx = currentBarIdx + offset;
    if (!this.isValidBarIndex(targetBarIdx, candlestickData.length)) {
      return null;
    }

    return {
      ...point,
      time: candlestickData[targetBarIdx].time,
    };
  }

  isValidBarIndex(barIdx, candlestickLength) {
    return barIdx >= 0 && barIdx < candlestickLength;
  }
}
