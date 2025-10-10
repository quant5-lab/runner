import TimeframeConverter from '../utils/timeframeConverter.js';

class PineDataSourceAdapter {
  constructor(providerManager) {
    this.providerManager = providerManager;
  }

  async getMarketData(symbol, timeframe, limit, sDate, eDate) {
    console.log(`!!! ADAPTER RECEIVED: symbol=${symbol}, timeframe=${timeframe}, limit=${limit}`);
    const ourTimeframe = TimeframeConverter.fromPineTS(timeframe);
    console.log(`!!! ADAPTER CONVERTED: ${timeframe} -> ${ourTimeframe}`);

    const { data } = await this.providerManager.fetchMarketData(
      symbol,
      ourTimeframe,
      limit
    );

    console.log(`!!! ADAPTER RETURNING: ${data.length} candles`);
    return data;
  }
}

export default PineDataSourceAdapter;
