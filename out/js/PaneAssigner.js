/* Pane assignment logic based on value range analysis (SRP) */
export class PaneAssigner {
  constructor(candlestickData) {
    this.candlestickRange = this.calculateCandlestickRange(candlestickData);
  }

  calculateCandlestickRange(candlestickData) {
    if (!candlestickData || candlestickData.length === 0) {
      return { min: 0, max: 0 };
    }

    let min = Infinity;
    let max = -Infinity;

    candlestickData.forEach((candle) => {
      if (candle.low < min) min = candle.low;
      if (candle.high > max) max = candle.high;
    });

    return { min, max };
  }

  calculateIndicatorRange(indicatorData) {
    if (!indicatorData || indicatorData.length === 0) {
      return { min: 0, max: 0 };
    }

    let min = Infinity;
    let max = -Infinity;
    let validCount = 0;

    indicatorData.forEach((point) => {
      if (point.value !== null && point.value !== undefined && !isNaN(point.value) && point.value !== 0) {
        if (point.value < min) min = point.value;
        if (point.value > max) max = point.value;
        validCount++;
      }
    });

    if (validCount === 0) {
      return { min: 0, max: 0 };
    }

    return { min, max };
  }

  rangesOverlap(range1, range2, overlapThreshold = 0.3) {
    const range1Span = range1.max - range1.min;
    const range2Span = range2.max - range2.min;

    if (range1Span === 0 || range2Span === 0) return false;

    const overlapMin = Math.max(range1.min, range2.min);
    const overlapMax = Math.min(range1.max, range2.max);
    const overlapSpan = Math.max(0, overlapMax - overlapMin);

    const overlapRatio = overlapSpan / Math.min(range1Span, range2Span);

    return overlapRatio >= overlapThreshold;
  }

  assignPane(indicatorKey, indicator, configOverride = null) {
    if (configOverride && configOverride[indicatorKey]) {
      return configOverride[indicatorKey];
    }

    if (indicator.pane && indicator.pane !== '') {
      return indicator.pane;
    }

    const indicatorRange = this.calculateIndicatorRange(indicator.data);

    if (this.rangesOverlap(this.candlestickRange, indicatorRange)) {
      return 'main';
    }

    return 'indicator';
  }

  assignAllPanes(indicators, configOverride = null) {
    const result = {};

    Object.entries(indicators).forEach(([key, indicator]) => {
      result[key] = {
        ...indicator,
        pane: this.assignPane(key, indicator, configOverride),
      };
    });

    return result;
  }
}
