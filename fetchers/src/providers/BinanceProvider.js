import { TimeframeParser, SUPPORTED_TIMEFRAMES } from '../utils/timeframeParser.js';
import { TimeframeError } from '../errors/TimeframeError.js';

const BINANCE_API_URL = 'https://api.binance.com/api/v3';

class CacheManager {
  constructor(cacheDuration = 5 * 60 * 1000) {
    this.cache = new Map();
    this.cacheDuration = cacheDuration;
  }

  generateKey(params) {
    return Object.entries(params)
      .filter(([_, value]) => value !== undefined)
      .map(([key, value]) => `${key}:${value}`)
      .join('|');
  }

  get(params) {
    const key = this.generateKey(params);
    const cached = this.cache.get(key);

    if (!cached) return null;

    if (Date.now() - cached.timestamp > this.cacheDuration) {
      this.cache.delete(key);
      return null;
    }

    return cached.data;
  }

  set(params, data) {
    const key = this.generateKey(params);
    this.cache.set(key, {
      data,
      timestamp: Date.now(),
    });
  }

  clear() {
    this.cache.clear();
  }
}

class BinanceProvider {
  constructor(logger, statsCollector) {
    this.logger = logger;
    this.stats = statsCollector;
    this.cacheManager = new CacheManager();
    this.supportedTimeframes = SUPPORTED_TIMEFRAMES.BINANCE;
    this.timezone = 'UTC';
  }

  clearCache() {
    this.cacheManager.clear();
  }

  async getMarketData(symbol, timeframe, limit = 100, sDate, eDate) {
    try {
      const convertedTimeframe = TimeframeParser.toBinanceTimeframe(timeframe);

      if (limit > 1000) {
        return await this.getPaginatedData(symbol, convertedTimeframe, limit);
      }

      const cacheParams = { symbol, timeframe: convertedTimeframe, limit, sDate, eDate };
      const cachedData = this.cacheManager.get(cacheParams);
      if (cachedData) {
        return cachedData;
      }

      this.stats.recordRequest('Binance', timeframe);

      let url = `${BINANCE_API_URL}/klines?symbol=${symbol}&interval=${convertedTimeframe}`;
      if (limit) url += `&limit=${limit}`;
      if (sDate) url += `&startTime=${sDate}`;
      if (eDate) url += `&endTime=${eDate}`;

      const response = await fetch(url);
      if (!response.ok) {
        const errorText = await response.text();
        if (errorText.includes('Invalid symbol')) {
          this.logger.debug(`Binance: Invalid symbol ${symbol}`);
          return [];
        }
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result = await response.json();
      if (!result || result.length === 0) {
        this.logger.debug(`No data from Binance for: ${symbol}`);
        return [];
      }

      const data = result.map((item) => ({
        openTime: parseInt(item[0]),
        open: parseFloat(item[1]),
        high: parseFloat(item[2]),
        low: parseFloat(item[3]),
        close: parseFloat(item[4]),
        volume: parseFloat(item[5]),
        closeTime: parseInt(item[6]),
        quoteAssetVolume: parseFloat(item[7]),
        numberOfTrades: parseInt(item[8]),
        takerBuyBaseAssetVolume: parseFloat(item[9]),
        takerBuyQuoteAssetVolume: parseFloat(item[10]),
        ignore: item[11],
      }));

      this.cacheManager.set(cacheParams, data);
      return data;
    } catch (error) {
      const errorMsg = error.message || '';

      if (errorMsg.includes('Invalid symbol')) {
        this.logger.debug(`Binance: Invalid symbol ${symbol}`);
        return [];
      }

      if (errorMsg.includes('Invalid interval') || error instanceof TimeframeError) {
        throw new TimeframeError(timeframe, symbol, 'Binance', this.supportedTimeframes);
      }

      this.logger.debug(`Binance Provider error: ${error.message}`);
      return [];
    }
  }

  async getPaginatedData(symbol, convertedTimeframe, limit) {
    const allData = [];
    let oldestTime = null;

    while (allData.length < limit) {
      const batchSize = Math.min(1000, limit - allData.length);
      this.stats.recordRequest('Binance', convertedTimeframe);

      let url = `${BINANCE_API_URL}/klines?symbol=${symbol}&interval=${convertedTimeframe}&limit=${batchSize}`;
      if (oldestTime) url += `&endTime=${oldestTime - 1}`;

      const response = await fetch(url);
      if (!response.ok) break;

      const result = await response.json();
      if (!result || result.length === 0) break;

      const batch = result.map((item) => ({
        openTime: parseInt(item[0]),
        open: parseFloat(item[1]),
        high: parseFloat(item[2]),
        low: parseFloat(item[3]),
        close: parseFloat(item[4]),
        volume: parseFloat(item[5]),
        closeTime: parseInt(item[6]),
        quoteAssetVolume: parseFloat(item[7]),
        numberOfTrades: parseInt(item[8]),
        takerBuyBaseAssetVolume: parseFloat(item[9]),
        takerBuyQuoteAssetVolume: parseFloat(item[10]),
        ignore: item[11],
      }));

      allData.unshift(...batch);
      oldestTime = batch[0].openTime;

      if (batch.length < batchSize) break;
    }

    return allData.slice(-limit);
  }
}

export { BinanceProvider };
