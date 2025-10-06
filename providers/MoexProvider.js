// MOEX (Moscow Exchange) Provider for PineTS
export class MoexProvider {
  constructor(logger) {
    this.baseUrl = 'https://iss.moex.com/iss';
    this.cache = new Map();
    this.cacheDuration = 5 * 60 * 1000; // 5 minutes
    this.logger = logger;
  }

  /* MOEX timeframe mapping */
  static timeframeMap = {
    1: '1', // 1 minute
    5: '5', // 5 minutes
    10: '10', // 10 minutes
    15: '15', // 15 minutes
    30: '30', // 30 minutes
    60: '60', // 1 hour
    240: '240', // 4 hours
    D: '24', // Daily
    W: '7', // Weekly (7 days)
    M: '31', // Monthly (31 days)
  };

  /* Convert PineTS timeframe to MOEX interval */
  convertTimeframe(timeframe) {
    return MoexProvider.timeframeMap[timeframe] || '24';
  }

  /* Generate cache key */
  getCacheKey(tickerId, timeframe, limit, sDate, eDate) {
    return `${tickerId}_${timeframe}_${limit}_${sDate}_${eDate}`;
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

  /* Convert MOEX candle to PineTS format */
  convertMoexCandle(moexCandle) {
    const [open, close, high, low, value, volume, begin, end] = moexCandle;

    return {
      openTime: new Date(begin).getTime(),
      open: parseFloat(open),
      high: parseFloat(high),
      low: parseFloat(low),
      close: parseFloat(close),
      volume: parseFloat(volume),
      closeTime: new Date(end).getTime(),
      quoteAssetVolume: parseFloat(value),
      numberOfTrades: 0,
      takerBuyBaseAssetVolume: 0,
      takerBuyQuoteAssetVolume: 0,
      ignore: null,
    };
  }

  /* Format dates for MOEX API */
  formatDate(timestamp) {
    if (!timestamp) return '';
    return new Date(timestamp).toISOString().split('T')[0];
  }

  /* Build MOEX API URL */
  buildUrl(tickerId, timeframe, limit, sDate, eDate) {
    const interval = this.convertTimeframe(timeframe);
    const url = `${this.baseUrl}/engines/stock/markets/shares/boards/TQBR/securities/${tickerId}/candles.json`;

    const params = new URLSearchParams();

    params.append('interval', interval);

    if (sDate) {
      params.append('from', this.formatDate(sDate));
    }

    if (eDate) {
      params.append('till', this.formatDate(eDate));
    }

    if (limit && !sDate && !eDate) {
      // Calculate date range based on limit
      const now = new Date();
      const daysBack = Math.ceil(limit * this.getTimeframeDays(timeframe));
      const startDate = new Date(now.getTime() - daysBack * 24 * 60 * 60 * 1000);
      params.append('from', this.formatDate(startDate.getTime()));
      params.append('till', this.formatDate(now.getTime()));
    }

    return url + (params.toString() ? '?' + params.toString() : '');
  }

  /* Get timeframe in days for limit calculation */
  getTimeframeDays(timeframe) {
    const days = {
      1: 1 / 1440, // 1 minute
      5: 1 / 288, // 5 minutes
      10: 1 / 144, // 10 minutes
      15: 1 / 96, // 15 minutes
      30: 1 / 48, // 30 minutes
      60: 1 / 24, // 1 hour
      240: 1 / 6, // 4 hours
      D: 1, // Daily
      W: 7, // Weekly
      M: 30, // Monthly
    };
    return days[timeframe] || 1;
  }

  /* Main method - get market data */
  async getMarketData(tickerId, timeframe, limit, sDate, eDate) {
    try {
      const cacheKey = this.getCacheKey(tickerId, timeframe, limit, sDate, eDate);
      const cached = this.getFromCache(cacheKey);

      if (cached) {
        console.log('MOEX cache hit:', tickerId, timeframe);
        return cached;
      }

      const url = this.buildUrl(tickerId, timeframe, limit, sDate, eDate);
      console.log('MOEX API request:', url);

      const response = await fetch(url);
      if (!response.ok) {
        throw new Error(`MOEX API error: ${response.status} ${response.statusText}`);
      }

      const data = await response.json();

      if (!data.candles || !data.candles.data) {
        this.logger.debug(`No candle data from MOEX for: ${tickerId}`);
        return [];
      }

      const convertedData = data.candles.data
        .map((candle) => this.convertMoexCandle(candle))
        .sort((a, b) => a.openTime - b.openTime); // Sort by time ascending

      // Apply limit if specified
      const limitedData = limit ? convertedData.slice(-limit) : convertedData;

      this.setCache(cacheKey, limitedData);
      console.log(`MOEX data retrieved: ${limitedData.length} candles for ${tickerId}`);

      return limitedData;
    } catch (error) {
      this.logger.debug(`MOEX Provider error: ${error.message}`);
      return [];
    }
  }
}
