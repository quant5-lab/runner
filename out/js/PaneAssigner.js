export class PaneAssigner {
  constructor(candlestickData) {
    this.candlestickRange = this._candlestickRange(candlestickData);
  }

  _candlestickRange(candlestickData) {
    if (!candlestickData?.length) return { min: 0, max: 0 };
    let min = Infinity, max = -Infinity;
    for (const { low, high } of candlestickData) {
      if (low < min) min = low;
      if (high > max) max = high;
    }
    return { min, max };
  }

  _indicatorRange(indicatorData) {
    if (!indicatorData?.length) return { min: 0, max: 0 };
    let min = Infinity, max = -Infinity, count = 0;
    for (const { value } of indicatorData) {
      if (value != null && !isNaN(value) && value !== 0) {
        if (value < min) min = value;
        if (value > max) max = value;
        count++;
      }
    }
    return count === 0 ? { min: 0, max: 0 } : { min, max };
  }

  _rangesOverlap(r1, r2, threshold = 0.3) {
    const span1 = r1.max - r1.min;
    const span2 = r2.max - r2.min;
    if (span1 === 0 || span2 === 0) return false;
    const overlapSpan = Math.max(0, Math.min(r1.max, r2.max) - Math.max(r1.min, r2.min));
    return overlapSpan / Math.min(span1, span2) >= threshold;
  }

  _assignPane(indicator) {
    if (indicator.pane) return indicator.pane;
    const range = this._indicatorRange(indicator.data);
    return this._rangesOverlap(this.candlestickRange, range) ? 'main' : 'indicator';
  }

  assignAllPanes(indicators) {
    const result = {};
    for (const [key, indicator] of Object.entries(indicators)) {
      result[key] = { ...indicator, pane: this._assignPane(indicator) };
    }
    return result;
  }
}
