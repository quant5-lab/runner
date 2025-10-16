// Yahoo Finance Provider for PineTS - Real market data for US stocks
import { TimeframeParser, SUPPORTED_TIMEFRAMES } from '../utils/timeframeParser.js';
import { TimeframeError } from '../errors/TimeframeError.js';

export class YahooFinanceProvider {
  constructor(logger, statsCollector) {
    this.baseUrl = 'https://query1.finance.yahoo.com/v8/finance/chart';
    this.cache = new Map();
    this.cacheDuration = 5 * 60 * 1000; // 5 minutes
    this.headers = {
      'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
    };
    this.logger = logger;
    this.stats = statsCollector;
    this.supportedTimeframes = SUPPORTED_TIMEFRAMES.YAHOO;
  }

  /* Convert PineTS timeframe to Yahoo interval */
  convertTimeframe(timeframe) {
    return TimeframeParser.toYahooInterval(timeframe);
  }

  /* Generate cache key */
  getCacheKey(symbol, timeframe, range) {
    return `${symbol}_${timeframe}_${range}`;
  }

  /* Check cache */
  getFromCache(key) {
    const cached = this.cache.get(key);
    if (!cached) return null;

    if (Date.now() - cached.timestamp > this.cacheDuration) {
      this.cache.delete(key);
      return null;
    }

    return cached.data;
  }

  /* Set cache */
  setCache(key, data) {
    this.cache.set(key, {
      data,
      timestamp: Date.now(),
    });
  }

  /* Convert Yahoo Finance data to PineTS format */
  convertYahooCandle(timestamp, open, high, low, close, volume, timeframe) {
    const intervalMinutes = TimeframeParser.parseToMinutes(timeframe);
    const openTime = timestamp * 1000;
    const closeTime = openTime + intervalMinutes * 60 * 1000 - 1;
    return {
      openTime,
      open: parseFloat(open),
      high: parseFloat(high),
      low: parseFloat(low),
      close: parseFloat(close),
      volume: parseFloat(volume || 0),
      closeTime,
      quoteAssetVolume: parseFloat(close) * parseFloat(volume || 0),
      numberOfTrades: 0,
      takerBuyBaseAssetVolume: 0,
      takerBuyQuoteAssetVolume: 0,
      ignore: null,
    };
  }

  /* Get date range string for Yahoo API */
  getDateRange(limit, timeframe) {
    // Convert timeframe to minutes to determine appropriate range
    const minutes = TimeframeParser.parseToMinutes(timeframe);

    // Calculate ranges based on actual trading hours and requested limit
    // Markets trade ~6.5 hours/day, 5 days/week = ~32.5 hours/week

    // Dynamic range selection based on requested limit
    if (minutes <= 1) {
      // 1 minute: ~390 points per day
      if (limit <= 390) return '1d';
      if (limit <= 1950) return '5d'; // ~1950 points in 5 days
      if (limit <= 3900) return '10d'; // ~3900 points in 10 days
      return '1mo'; // For larger requests
    }

    if (minutes <= 5) {
      // 5 minutes: ~78 points per day (390/5)
      if (limit <= 78) return '1d';
      if (limit <= 390) return '5d';
      if (limit <= 780) return '10d';
      return '1mo';
    }

    if (minutes <= 15) {
      // 15 minutes: ~26 points per day (390/15)
      if (limit <= 26) return '1d';
      if (limit <= 130) return '5d';
      if (limit <= 260) return '10d';
      return '1mo';
    }

    if (minutes <= 30) {
      // 30 minutes: ~13 points per day (390/30)
      if (limit <= 65) return '5d';
      if (limit <= 130) return '10d';
      if (limit <= 260) return '1mo';
      return '3mo';
    }

    if (minutes <= 60) {
      // 1 hour: ~6.5 points per day
      if (limit <= 65) return '10d'; // ~65 hours in 10 trading days
      if (limit <= 130) return '1mo'; // ~130 hours in 20 trading days
      if (limit <= 195) return '3mo'; // ~195 hours in 30 trading days
      return '6mo'; // ~390 hours for larger requests
    }

    if (minutes <= 240) {
      // 4 hours: ~1.6 points per day
      if (limit <= 50) return '1mo';
      if (limit <= 100) return '3mo';
      if (limit <= 200) return '6mo';
      return '1y';
    }

    if (minutes <= 1440) {
      // Daily: ~1 point per day
      if (limit <= 30) return '1mo';
      if (limit <= 90) return '3mo';
      if (limit <= 180) return '6mo';
      if (limit <= 250) return '1y';
      return '2y';
    }

    if (minutes <= 10080) {
      // Weekly: ~52 points per year
      if (limit <= 52) return '1y';
      if (limit <= 104) return '2y';
      if (limit <= 260) return '5y';
      return '10y';
    }

    // Monthly: ~12 points per year
    if (limit <= 24) return '2y';
    if (limit <= 60) return '5y';
    return '10y';
  }

  /* Build Yahoo Finance API URL */
  buildUrl(symbol, interval, range) {
    const params = new URLSearchParams({
      interval,
      range,
    });

    return `${this.baseUrl}/${symbol}?${params.toString()}`;
  }

  /* Main method - get market data */
  async getMarketData(symbol, timeframe, limit = 100, sDate, eDate) {
    try {
      /* Try to convert timeframe - if fails, test with 1d to check if symbol exists */
      let interval, range, url;
      try {
        interval = this.convertTimeframe(timeframe);
        range = this.getDateRange(limit, timeframe);
        url = this.buildUrl(symbol, interval, range);
      } catch (error) {
        if (error instanceof TimeframeError) {
          /* Timeframe unsupported - test with 1d to check if symbol exists */
          this.logger.debug(`Yahoo: Timeframe ${timeframe} unsupported, testing ${symbol} with 1d`);

          const testInterval = TimeframeParser.toYahooInterval('1d');
          const testRange = '1d';
          const testUrl = this.buildUrl(symbol, testInterval, testRange);

          const testResponse = await fetch(testUrl, { headers: this.headers });

          if (testResponse.ok) {
            const testText = await testResponse.text();
            const testData = JSON.parse(testText);

            if (
              testData.chart?.result?.[0]?.timestamp &&
              testData.chart.result[0].timestamp.length > 0
            ) {
              /* Symbol EXISTS but timeframe INVALID */
              throw new TimeframeError(timeframe, symbol, 'YahooFinance', this.supportedTimeframes);
            }
          }

          /* Symbol NOT FOUND or test failed */
          return [];
        }
        /* Other errors - return [] to allow next provider */
        this.logger.debug(`Yahoo buildUrl error: ${error.message}`);
        return [];
      }

      const cacheKey = this.getCacheKey(symbol, interval, range);

      const cached = this.getFromCache(cacheKey);
      if (cached) {
        console.log('Yahoo Finance cache hit:', symbol, interval);
        return cached.slice(-limit);
      }

      console.log('Yahoo Finance API request:', url);
      console.log('Yahoo Finance headers:', JSON.stringify(this.headers));

      this.stats.recordRequest('YahooFinance', timeframe);
      const response = await fetch(url, { headers: this.headers });
      console.log('Yahoo Finance response status:', response.status, response.statusText);
      console.log(
        'Yahoo Finance response headers:',
        JSON.stringify(Object.fromEntries(response.headers.entries())),
      );

      if (!response.ok) {
        const errorText = await response.text();
        console.log('Yahoo Finance error response body:', errorText);
        throw new Error(`Yahoo Finance API error: ${response.status} ${response.statusText}`);
      }

      const responseText = await response.text();
      console.log('Yahoo Finance response body length:', responseText.length);
      console.log('Yahoo Finance response body start:', responseText.substring(0, 200));

      const data = JSON.parse(responseText);

      if (!data.chart || !data.chart.result || data.chart.result.length === 0) {
        console.warn('No chart data from Yahoo Finance for:', symbol);
        return [];
      }

      const result = data.chart.result[0];
      if (
        !result.timestamp ||
        !result.indicators ||
        !result.indicators.quote ||
        !result.indicators.quote[0]
      ) {
        console.warn('Invalid data structure from Yahoo Finance for:', symbol);
        return [];
      }

      const timestamps = result.timestamp;
      const quote = result.indicators.quote[0];
      const { open, high, low, close, volume } = quote;

      /* Empty response - disambiguate with 1d test */
      if (!timestamps || timestamps.length === 0) {
        console.warn('No timestamps in Yahoo Finance data for:', symbol);

        if (timeframe !== '1d') {
          /* Test with 1d to determine if symbol exists or timeframe invalid */
          const testInterval = TimeframeParser.toYahooInterval('1d');
          const testRange = '1d';
          const testUrl = this.buildUrl(symbol, testInterval, testRange);

          const testResponse = await fetch(testUrl, { headers: this.headers });

          if (testResponse.ok) {
            const testText = await testResponse.text();
            const testData = JSON.parse(testText);

            if (
              testData.chart?.result?.[0]?.timestamp &&
              testData.chart.result[0].timestamp.length > 0
            ) {
              /* Symbol EXISTS but original timeframe INVALID */
              throw new TimeframeError(timeframe, symbol, 'YahooFinance', this.supportedTimeframes);
            }
          }
        }

        /* Symbol NOT FOUND */
        return [];
      }

      const convertedData = [];
      for (let i = 0; i < timestamps.length; i++) {
        if (open[i] !== null && high[i] !== null && low[i] !== null && close[i] !== null) {
          convertedData.push(
            this.convertYahooCandle(timestamps[i], open[i], high[i], low[i], close[i], volume[i], timeframe),
          );
        }
      }

      // Sort by time ascending
      convertedData.sort((a, b) => a.openTime - b.openTime);

      this.setCache(cacheKey, convertedData);
      console.log(`Yahoo Finance data retrieved: ${convertedData.length} candles for ${symbol}`);

      // Apply limit
      return convertedData.slice(-limit);
    } catch (error) {
      if (error.name === 'TimeframeError') {
        throw error; // Re-throw TimeframeError to be handled by ProviderManager
      }
      this.logger.debug(`Yahoo Finance provider error for ${symbol}: ${error.message}`);
      if (error.stack) {
        this.logger.debug(`Yahoo Finance error stack: ${error.stack}`);
      }
      return [];
    }
  }
}
