/* Converts between our timeframe format and PineTS format */
class TimeframeConverter {
  static toPineTS(ourTimeframe) {
    const map = {
      '1m': '1',
      '3m': '3',
      '5m': '5',
      '15m': '15',
      '30m': '30',
      '45m': '45',
      '1h': '60',
      '2h': '120',
      '3h': '180',
      '4h': '240',
      '1d': 'D',
      '1w': 'W',
      '1M': 'M',
    };
    return map[ourTimeframe] || ourTimeframe;
  }

  static fromPineTS(pineTF) {
    const map = {
      '1': '1m',
      '3': '3m',
      '5': '5m',
      '15': '15m',
      '30': '30m',
      '45': '45m',
      '60': '1h',
      '120': '2h',
      '180': '3h',
      '240': '4h',
      'D': '1d',
      'W': '1w',
      'M': '1M',
    };
    return map[pineTF] || pineTF;
  }
}

export default TimeframeConverter;
