export class AlphaVantageProvider {
  constructor() {
    this.baseUrl = 'https://www.alphavantage.co/query';
    this.apiKey = process.env.ALPHA_VANTAGE_API_KEY || 'demo';
  }

  /* Convert Alpha Vantage timeframe to API interval */
  convertTimeframe(timeframe) {
    const timeframeMap = {
      1: '1min',
      5: '5min',
      15: '15min',
      30: '30min',
      60: '60min',
      D: 'daily',
      W: 'weekly',
      M: 'monthly',
    };

    return timeframeMap[timeframe] || 'daily';
  }

  /* Convert Alpha Vantage candle to PineTS format */
  convertAlphaVantageCandle(timestamp, data) {
    return {
      openTime: new Date(timestamp).getTime(),
      open: parseFloat(data['1. open']),
      high: parseFloat(data['2. high']),
      low: parseFloat(data['3. low']),
      close: parseFloat(data['4. close']),
      volume: parseFloat(data['5. volume']),
    };
  }

  /* Get appropriate API function based on timeframe */
  getApiFunction(timeframe) {
    const interval = this.convertTimeframe(timeframe);

    if (['1min', '5min', '15min', '30min', '60min'].includes(interval)) {
      return 'TIME_SERIES_INTRADAY';
    } else if (interval === 'daily') {
      return 'TIME_SERIES_DAILY';
    } else if (interval === 'weekly') {
      return 'TIME_SERIES_WEEKLY';
    } else if (interval === 'monthly') {
      return 'TIME_SERIES_MONTHLY';
    }

    return 'TIME_SERIES_DAILY';
  }

  /* Build API URL */
  buildUrl(symbol, timeframe) {
    const apiFunction = this.getApiFunction(timeframe);
    const interval = this.convertTimeframe(timeframe);

    let url = `${this.baseUrl}?function=${apiFunction}&symbol=${symbol}&apikey=${this.apiKey}`;

    if (apiFunction === 'TIME_SERIES_INTRADAY') {
      url += `&interval=${interval}`;
    }

    return url;
  }

  /* Get time series key based on API function */
  getTimeSeriesKey(apiFunction) {
    const keyMap = {
      TIME_SERIES_INTRADAY: 'Time Series',
      TIME_SERIES_DAILY: 'Time Series (Daily)',
      TIME_SERIES_WEEKLY: 'Weekly Time Series',
      TIME_SERIES_MONTHLY: 'Monthly Time Series',
    };

    return keyMap[apiFunction] || 'Time Series (Daily)';
  }

  async getMarketData(symbol, timeframe, bars) {
    const url = this.buildUrl(symbol, timeframe);
    const apiFunction = this.getApiFunction(timeframe);

    console.log(`Alpha Vantage API request: ${url}`);

    try {
      const response = await fetch(url);
      const data = await response.json();

      if (data['Error Message']) {
        throw new Error(`Alpha Vantage API error: ${data['Error Message']}`);
      }

      if (data.Note) {
        throw new Error(`Alpha Vantage API limit: ${data.Note}`);
      }

      if (data.Information) {
        throw new Error(`Alpha Vantage API info: ${data.Information}`);
      }

      const timeSeriesKey = this.getTimeSeriesKey(apiFunction);
      const timeSeries = data[timeSeriesKey];

      if (!timeSeries) {
        throw new Error(`No time series data found for ${symbol}`);
      }

      /* Convert to PineTS format and sort by time */
      const candles = Object.entries(timeSeries)
        .map(([timestamp, candleData]) => this.convertAlphaVantageCandle(timestamp, candleData))
        .sort((a, b) => a.openTime - b.openTime)
        .slice(-bars); // Take last N bars

      console.log(`Alpha Vantage data retrieved: ${candles.length} candles for ${symbol}`);

      return candles;
    } catch (error) {
      console.error(`Alpha Vantage provider error for ${symbol}:`, error.message);
      throw error;
    }
  }
}
