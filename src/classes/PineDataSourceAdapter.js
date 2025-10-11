import TimeframeConverter from '../utils/timeframeConverter.js';

class PineDataSourceAdapter {
  constructor(providerManager) {
    this.providerManager = providerManager;
  }

  async getMarketData(symbol, timeframe, limit, sDate, eDate) {
    const ourTimeframe = TimeframeConverter.fromPineTS(timeframe);

    const { data } = await this.providerManager.fetchMarketData(
      symbol,
      ourTimeframe,
      limit,
    );

    return data;
  }
}

export default PineDataSourceAdapter;
