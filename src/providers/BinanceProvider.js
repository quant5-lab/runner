import { Provider } from '../../../PineTS/dist/pinets.dev.es.js';
import { TimeframeParser } from '../utils/timeframeParser.js';

class BinanceProvider {
  constructor(logger) {
    this.logger = logger;
    this.binanceProvider = Provider.Binance;
  }

  async getMarketData(symbol, timeframe, limit = 100, sDate, eDate) {
    const convertedTimeframe = TimeframeParser.toBinanceTimeframe(timeframe);

    const result = await this.binanceProvider.getMarketData(symbol, convertedTimeframe, limit, sDate, eDate);

    return result;
  }
}

export { BinanceProvider };
