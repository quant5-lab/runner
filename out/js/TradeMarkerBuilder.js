const DIRECTION_LONG_COLOR  = '#26a69a';
const DIRECTION_SHORT_COLOR = '#ef5350';
const OUTCOME_PROFIT_COLOR  = '#26a69a';
const OUTCOME_LOSS_COLOR    = '#ef5350';

export class TradeMarkerBuilder {
  build(strategy, candlestickData) {
    if (!strategy || !candlestickData?.length) return [];

    const markers = [];

    for (const trade of strategy.trades || []) {
      const entryTime = this._resolveTime(trade.entryBar, trade.entryTime, candlestickData);
      const exitTime  = this._resolveTime(trade.exitBar,  trade.exitTime,  candlestickData);
      if (entryTime !== null) markers.push(this._entryMarker(trade, entryTime));
      if (exitTime  !== null) markers.push(this._exitMarker(trade, exitTime));
    }

    for (const trade of strategy.openTrades || []) {
      const entryTime = this._resolveTime(trade.entryBar, trade.entryTime, candlestickData);
      if (entryTime !== null) markers.push(this._entryMarker(trade, entryTime));
    }

    markers.sort((a, b) => a.time - b.time);
    return markers;
  }

  _resolveTime(barIndex, unixTime, candlestickData) {
    if (barIndex >= 0 && barIndex < candlestickData.length) {
      return candlestickData[barIndex].time;
    }
    return unixTime || null;
  }

  _entryMarker(trade, time) {
    const isLong = trade.direction === 'long';
    return {
      time,
      position: isLong ? 'belowBar' : 'aboveBar',
      color:    this._entryColor(trade),
      shape:    isLong ? 'arrowUp' : 'arrowDown',
    };
  }

  _exitMarker(trade, time) {
    const isLong = trade.direction === 'long';
    return {
      time,
      position: isLong ? 'aboveBar' : 'belowBar',
      color:    this._exitColor(trade),
      shape:    'circle',
    };
  }

  _entryColor(trade) {
    return trade.direction === 'long' ? DIRECTION_LONG_COLOR : DIRECTION_SHORT_COLOR;
  }

  _exitColor(trade) {
    return trade.profit >= 0 ? OUTCOME_PROFIT_COLOR : OUTCOME_LOSS_COLOR;
  }
}
