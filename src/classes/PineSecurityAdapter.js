import TimeframeConverter from '../utils/timeframeConverter.js';

let adapterInstanceCounter = 0;

class PineSecurityAdapter {
  constructor(providerManager) {
    this.providerManager = providerManager;
    this.marketDataCache = new Map();
    this.instanceId = ++adapterInstanceCounter;
    console.log('!!! NEW PINE SECURITY ADAPTER CREATED - instanceId:', this.instanceId);
  }

  async getMarketData(symbol, timeframe, limit, sDate, eDate) {
    const ourTimeframe = TimeframeConverter.fromPineTS(timeframe);

    const cacheKey = `${symbol}|${ourTimeframe}|${limit || 'all'}`;
    console.log('!!! ADAPTER', this.instanceId, 'CACHE KEY:', cacheKey);
    console.log('!!! ADAPTER', this.instanceId, 'CACHE SIZE:', this.marketDataCache.size, 'keys:', Array.from(this.marketDataCache.keys()));
    
    if (this.marketDataCache.has(cacheKey)) {
      const cached = this.marketDataCache.get(cacheKey);
      console.log('!!! ADAPTER', this.instanceId, 'CACHE HIT:', cacheKey, `${cached.length} candles`);
      return cached;
    }
    
    console.log('!!! ADAPTER', this.instanceId, 'CACHE MISS:', cacheKey);

    const { data } = await this.providerManager.fetchMarketData(
      symbol,
      ourTimeframe,
      limit,
    );

    console.log('!!! ADAPTER', this.instanceId, 'SETTING CACHE:', cacheKey, `(${data.length} candles)`);
    this.marketDataCache.set(cacheKey, data);
    console.log('!!! ADAPTER', this.instanceId, 'CACHE SIZE AFTER SET:', this.marketDataCache.size, 'keys:', Array.from(this.marketDataCache.keys()));
    return data;
  }
}

export default PineSecurityAdapter;
