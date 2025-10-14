import { TimeframeError } from '../errors/TimeframeError.js';
import TimeframeConverter from '../utils/timeframeConverter.js';

class ProviderManager {
  constructor(providerChain, logger) {
    this.providerChain = providerChain;
    this.logger = logger;
    this.pending = new Map();
  }

  getCacheKey(symbol, timeframe, limit) {
    return `${symbol}|${timeframe}|${limit}`;
  }

  async getMarketData(symbol, timeframe, limit, sDate, eDate) {
    const ourTimeframe = TimeframeConverter.fromPineTS(timeframe);
    const cacheKey = this.getCacheKey(symbol, ourTimeframe, limit);

    if (this.pending.has(cacheKey)) {
      return (await this.pending.get(cacheKey)).data;
    }

    const fetchPromise = this.fetchMarketData(symbol, ourTimeframe, limit);
    this.pending.set(cacheKey, fetchPromise);

    try {
      const result = await fetchPromise;
      return result.data;
    } finally {
      this.pending.delete(cacheKey);
    }
  }

  validateDataFreshness(marketData, symbol, timeframe, providerName) {
    if (!marketData?.length) return;

    const mostRecentCandle = marketData[marketData.length - 1];

    const timeField = mostRecentCandle.time || mostRecentCandle.closeTime;

    const candleTime = new Date(timeField * (timeField > 1000000000000 ? 1 : 1000));

    const now = new Date();

    const ageInDays = (now - candleTime) / (24 * 60 * 60 * 1000);

    let maxAgeDays;
    if (timeframe.includes('m') && !timeframe.includes('mo')) {
      maxAgeDays = 1;
    } else if (timeframe.includes('h')) {
      maxAgeDays = 2;
    } else if (timeframe.includes('d') || timeframe === 'D') {
      maxAgeDays = 7;
    } else {
      maxAgeDays = 30;
    }

    if (ageInDays > maxAgeDays) {
      throw new Error(
        `${providerName} returned stale data for ${symbol} ${timeframe}: ` +
          `latest candle is ${Math.floor(ageInDays)} days old (${candleTime.toDateString()}). ` +
          `Expected data within ${maxAgeDays} days.`,
      );
    }
  }

  async fetchMarketData(symbol, timeframe, bars) {
    for (let i = 0; i < this.providerChain.length; i++) {
      const { name, instance } = this.providerChain[i];

      const providerStartTime = performance.now();
      this.logger.log(`Attempting:\t${name} > ${symbol}`);

      try {
        const marketData = await instance.getMarketData(symbol, timeframe, bars);

        if (marketData?.length > 0) {
          this.validateDataFreshness(marketData, symbol, timeframe, name);

          const providerDuration = (performance.now() - providerStartTime).toFixed(2);
          this.logger.log(
            `Found data:\t${name} (${marketData.length} candles, took ${providerDuration}ms)`,
          );
          return { provider: name, data: marketData, instance };
        }

        this.logger.log(`No data:\t${name} > ${symbol}`);
      } catch (error) {
        if (error.message.includes('returned stale data')) {
          throw error;
        }
        if (error instanceof TimeframeError) {
          throw error;
        }
        this.logger.log(`Failed:\t\t${name} > ${symbol}`);
        this.logger.debug(`Error from ${name} provider: ${error}`);
        continue;
      }
    }

    throw new Error(`All providers failed for symbol: ${symbol}`);
  }
}

export { ProviderManager };
