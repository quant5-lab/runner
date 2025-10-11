import { Provider } from '../../../PineTS/dist/pinets.dev.es.js';
import { TimeframeParser, SUPPORTED_TIMEFRAMES } from '../utils/timeframeParser.js';
import { TimeframeError } from '../errors/TimeframeError.js';

class BinanceProvider {
  constructor(logger) {
    this.logger = logger;
    this.binanceProvider = Provider.Binance;
    this.supportedTimeframes = SUPPORTED_TIMEFRAMES.BINANCE;
  }

  async getMarketData(symbol, timeframe, limit = 100, sDate, eDate) {
    try {
      /* Convert timeframe to Binance format */
      const convertedTimeframe = TimeframeParser.toBinanceTimeframe(timeframe);

      const result = await this.binanceProvider.getMarketData(symbol, convertedTimeframe, limit, sDate, eDate);

      /* Symbol not found or no data - return [] to allow next provider to try */
      if (!result || result.length === 0) {
        this.logger.debug(`No data from Binance for: ${symbol}`);
        return [];
      }

      return result;
    } catch (error) {
      /* Parse Binance API error messages */
      const errorMsg = error.message || '';

      /* Invalid symbol - return [] to continue chain */
      if (errorMsg.includes('Invalid symbol')) {
        this.logger.debug(`Binance: Invalid symbol ${symbol}`);
        return [];
      }

      /* Invalid interval - throw TimeframeError to stop chain */
      if (errorMsg.includes('Invalid interval') || error instanceof TimeframeError) {
        throw new TimeframeError(timeframe, symbol, 'Binance', this.supportedTimeframes);
      }

      /* Other errors - return [] to allow next provider to try */
      this.logger.debug(`Binance Provider error: ${error.message}`);
      return [];
    }
  }
}

export { BinanceProvider };
