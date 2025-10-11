import TimeframeConverter from '../utils/timeframeConverter.js';

class PineSecurityAdapter {
  constructor(providerManager) {
    this.providerManager = providerManager;
    this.marketDataCache = new Map();
  }

  async getMarketData(symbol, timeframe, limit, sDate, eDate) {
    const ourTimeframe = TimeframeConverter.fromPineTS(timeframe);

    const cacheKey = `${symbol}|${ourTimeframe}|${limit || 'all'}`;
    if (this.marketDataCache.has(cacheKey)) {
      return this.marketDataCache.get(cacheKey);
    }

    const { data } = await this.providerManager.fetchMarketData(
      symbol,
      ourTimeframe,
      limit,
    );

    this.marketDataCache.set(cacheKey, data);
    return data;
  }
}

export default PineSecurityAdapter;
