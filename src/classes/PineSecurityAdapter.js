import TimeframeConverter from '../utils/timeframeConverter.js';

let adapterInstanceCounter = 0;

class PineSecurityAdapter {
  constructor(providerManager, statsCollector) {
    this.providerManager = providerManager;
    this.marketDataCache = new Map();
    this.instanceId = ++adapterInstanceCounter;
    this.stats = statsCollector;
  }

  async getMarketData(symbol, timeframe, limit, sDate, eDate) {
    const ourTimeframe = TimeframeConverter.fromPineTS(timeframe);

    const cacheKey = `${symbol}|${ourTimeframe}|${limit || 'all'}`;
    
    if (this.marketDataCache.has(cacheKey)) {
      const cached = this.marketDataCache.get(cacheKey);
      this.stats.recordCacheHit();
      return cached;
    }
    
    this.stats.recordCacheMiss();

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
