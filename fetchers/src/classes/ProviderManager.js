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
      maxAgeDays = 4;
    } else if (timeframe.includes('d') || timeframe === 'D') {
      maxAgeDays = 10;
    } else {
      maxAgeDays = 45;
    }

    if (ageInDays > maxAgeDays) {
      this.logger.log(
        `⚠️  ${providerName} data age warning for ${symbol} ${timeframe}: ` +
          `latest candle is ${Math.floor(ageInDays)} days old (${candleTime.toDateString()}). ` +
          `Expected within ${maxAgeDays} days. Continuing anyway...`,
      );
    }
  }

  checkDiskCache(symbol, timeframe, outputFile) {
    const fs = require('fs');
    const path = require('path');
    
    const maxAge = parseInt(process.env.TEST_DATA_MAX_AGE_SECONDS || '300');
    const cacheFile = outputFile;
    
    if (!fs.existsSync(cacheFile)) {
      return null;
    }
    
    const stats = fs.statSync(cacheFile);
    const ageSeconds = (Date.now() - stats.mtimeMs) / 1000;
    
    if (ageSeconds < maxAge) {
      this.logger.log(`✓ Using cached data (age: ${Math.floor(ageSeconds)}s)`);
      const data = JSON.parse(fs.readFileSync(cacheFile, 'utf8'));
      return { provider: 'disk-cache', data, timezone: 'UTC', message: `Using cached ${symbol} ${timeframe}` };
    }
    
    this.logger.log(`⚠ Cache expired (age: ${Math.floor(ageSeconds)}s > ${maxAge}s)`);
    return null;
  }

  async fetchMarketData(symbol, timeframe, bars, outputFile) {
    if (outputFile) {
      const cached = this.checkDiskCache(symbol, timeframe, outputFile);
      if (cached) {
        return cached;
      }
    }

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
          
          const result = {
            provider: name,
            data: marketData,
            instance,
            timezone: instance.timezone || 'UTC',
            message: `Fetched ${marketData.length} bars from ${name}`,
          };
          
          if (outputFile) {
            const fs = require('fs');
            const path = require('path');
            const dir = path.dirname(outputFile);
            if (!fs.existsSync(dir)) {
              fs.mkdirSync(dir, { recursive: true });
            }
            fs.writeFileSync(outputFile, JSON.stringify(marketData, null, 2));
            this.logger.log(`✓ Saved: ${outputFile}`);
          }
          
          return result;
        }

        this.logger.log(`No data:\t${name} > ${symbol}`);
      } catch (error) {
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
