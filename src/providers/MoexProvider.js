import { TimeframeParser, SUPPORTED_TIMEFRAMES } from '../utils/timeframeParser.js';
import { TimeframeError } from '../errors/TimeframeError.js';

class MoexProvider {
  constructor(logger, statsCollector) {
    this.logger = logger;
    this.stats = statsCollector;
    this.baseUrl = 'https://iss.moex.com/iss';
    this.cache = new Map();
    this.cacheDuration = 5 * 60 * 1000;
    this.supportedTimeframes = SUPPORTED_TIMEFRAMES.MOEX;
  }

  /* Convert timeframe - throws TimeframeError if invalid */
  convertTimeframe(timeframe) {
    return TimeframeParser.toMoexInterval(timeframe);
  }

  getCacheKey(tickerId, timeframe, limit, sDate, eDate) {
    return `${tickerId}_${timeframe}_${limit}_${sDate}_${eDate}`;
  }

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
    const date = new Date(timestamp);
    return date.toISOString().split('T')[0];
  }

  /* Format end date for MOEX API - set to end of day */
  formatEndDate(timestamp) {
    if (!timestamp) return '';
    const date = new Date(timestamp);
    date.setHours(23, 59, 59, 999);
    return date.toISOString().replace('T', ' ').replace('Z', '');
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
      params.append('till', this.formatEndDate(eDate));
    }

    if (limit && !sDate && !eDate) {
      // Calculate date range based on limit
      const now = new Date();
      const minutes = TimeframeParser.parseToMinutes(timeframe);

      // Calculate proper days back based on limit and timeframe
      let daysBack = Math.ceil(limit * this.getTimeframeDays(timeframe));

      // Apply multipliers to account for non-trading periods
      if (minutes >= 1440) {
        // For daily+ timeframes, account for weekends and holidays
        // Request ~1.4x more calendar days to ensure we get enough trading days
        daysBack = Math.ceil(daysBack * 1.4);
      } else if (minutes >= 60) {
        // For hourly+ intraday timeframes, account for non-trading hours
        // MOEX trades ~10 hours/day, so need ~2.4x more calendar days (24/10)
        daysBack = Math.ceil(daysBack * 2.4);
      } else if (minutes >= 10) {
        // For 10m+ timeframes, similar logic but slightly less buffer
        daysBack = Math.ceil(daysBack * 2.2);
      } else {
        // For very short timeframes (1m), use conservative buffer
        daysBack = Math.ceil(daysBack * 2.0);
      }

      const startDate = new Date(now.getTime() - daysBack * 24 * 60 * 60 * 1000);

      // For intraday timeframes, extend end date to ensure current day data is included
      const endDate =
        minutes < 1440
          ? new Date(now.getTime() + 24 * 60 * 60 * 1000) // Tomorrow for intraday
          : now; // Today for daily+

      params.append('from', this.formatDate(startDate.getTime()));
      params.append('till', this.formatEndDate(endDate.getTime()));

      // MOEX returns oldest 500 candles by default - use reverse to get newest
      params.append('iss.reverse', 'true');
    }

    return url + (params.toString() ? '?' + params.toString() : '');
  }

  /* Get timeframe in days for limit calculation */
  getTimeframeDays(timeframe) {
    const minutes = TimeframeParser.parseToMinutes(timeframe);

    // MOEX trading hours: ~9 hours per trading day (10:00-18:50 Moscow time)
    // Need to account for actual trading time, not full calendar days
    const tradingHoursPerDay = 9;
    const minutesPerTradingDay = tradingHoursPerDay * 60; // 540 minutes per trading day

    // For timeframes >= 1 day, use calendar days
    if (minutes >= 1440) {
      return minutes / 1440; // Calendar days for daily+ timeframes
    }

    // CRITICAL: 1m and 5m data has processing delays on MOEX
    // Current day data not immediately available for very short timeframes
    if (minutes <= 5) {
      // For 1m and 5m, go back extra days to ensure sufficient data
      const tradingDays = minutes / minutesPerTradingDay;
      const calendarDaysWithWeekends = tradingDays * 1.4; // ~40% extra for weekends
      const delayBuffer = 2; // Extra 2 days for processing delays
      return calendarDaysWithWeekends + delayBuffer;
    }

    // For other intraday timeframes, account for trading hours and add buffer for weekends
    const tradingDays = minutes / minutesPerTradingDay;
    const calendarDaysWithWeekends = tradingDays * 1.4; // ~40% extra for weekends

    return calendarDaysWithWeekends;
  }

  /* Main method - get market data */
  async getMarketData(tickerId, timeframe, limit, sDate, eDate) {
    try {
      const cacheKey = this.getCacheKey(tickerId, timeframe, limit, sDate, eDate);
      const cached = this.getFromCache(cacheKey);

      if (cached) {
        this.stats.recordCacheHit();
        console.log('MOEX cache hit:', tickerId, timeframe);
        return cached;
      }

      this.stats.recordCacheMiss();

      /* Try to convert timeframe - if fails, test with 1d to check if symbol exists */
      let url;
      try {
        url = this.buildUrl(tickerId, timeframe, limit, sDate, eDate);
      } catch (error) {
        if (error instanceof TimeframeError) {
          /* Timeframe unsupported - test with 1d to check if symbol exists */
          this.logger.debug(
            `MOEX: Timeframe ${timeframe} unsupported, testing ${tickerId} with 1d`,
          );

          const testUrl = this.buildUrl(tickerId, '1d', 1, sDate, eDate);
          const testResponse = await fetch(testUrl);

          if (testResponse.ok) {
            const testData = await testResponse.json();

            if (testData.candles?.data?.length > 0) {
              /* Symbol EXISTS but timeframe INVALID */
              throw new TimeframeError(timeframe, tickerId, 'MOEX', this.supportedTimeframes);
            }
          }

          /* Symbol NOT FOUND or test failed */
          return [];
        }
        /* Other errors - return [] to allow next provider */
        this.logger.debug(`MOEX buildUrl error: ${error.message}`);
        return [];
      }

      console.log('MOEX API request:', url);

      this.stats.recordRequest('MOEX', timeframe);
      const response = await fetch(url);

      /* HTTP error - return [] to allow next provider to try */
      if (!response.ok) {
        this.logger.debug(
          `MOEX API error: ${response.status} ${response.statusText} for ${tickerId}`,
        );
        return [];
      }

      const data = await response.json();

      /* Data found - success */
      if (data.candles?.data?.length > 0) {
        const convertedData = data.candles.data
          .map((candle) => this.convertMoexCandle(candle))
          .sort((a, b) => a.openTime - b.openTime);

        const limitedData = limit ? convertedData.slice(-limit) : convertedData;

        this.setCache(cacheKey, limitedData);
        console.log(`MOEX data retrieved: ${limitedData.length} candles for ${tickerId}`);

        return limitedData;
      }

      /* Empty response - disambiguate: symbol not found OR invalid timeframe */
      if (timeframe !== '1d') {
        this.logger.debug(`MOEX: Empty response for ${tickerId} ${timeframe}, testing with 1d`);

        const testUrl = this.buildUrl(tickerId, '1d', 1, sDate, eDate);
        const testResponse = await fetch(testUrl);

        if (testResponse.ok) {
          const testData = await testResponse.json();
          if (testData.candles?.data?.length > 0) {
            /* Symbol exists but requested timeframe invalid */
            throw new TimeframeError(timeframe, tickerId, 'MOEX', this.supportedTimeframes);
          }
        }
      }

      /* Symbol not found - return [] to allow next provider to try */
      this.logger.debug(`MOEX: Symbol not found: ${tickerId}`);
      return [];
    } catch (error) {
      /* TimeframeError - symbol exists but timeframe invalid - STOP chain */
      if (error instanceof TimeframeError) {
        throw error;
      }
      /* Other errors (network, etc) - return [] to allow next provider to try */
      this.logger.debug(`MOEX Provider error: ${error.message}`);
      return [];
    }
  }
}

export { MoexProvider };
