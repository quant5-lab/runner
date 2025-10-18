import { Provider } from '../../../PineTS/dist/pinets.dev.es.js';
import { TimeframeParser, SUPPORTED_TIMEFRAMES } from '../utils/timeframeParser.js';
import { TimeframeError } from '../errors/TimeframeError.js';

class BinanceProvider {
  constructor(logger, statsCollector) {
    this.logger = logger;
    this.stats = statsCollector;
    this.binanceProvider = Provider.Binance;
    this.supportedTimeframes = SUPPORTED_TIMEFRAMES.BINANCE;
  }

  async getMarketData(symbol, timeframe, limit = 100, sDate, eDate) {
    try {
      /* Convert timeframe to Binance format */
      const convertedTimeframe = TimeframeParser.toBinanceTimeframe(timeframe);

      /* Binance API hard limit: 1000 candles per request - use pagination for more */
      if (limit > 1000) {
        return await this.getPaginatedData(symbol, convertedTimeframe, limit);
      }

      this.stats.recordRequest('Binance', timeframe);
      const result = await this.binanceProvider.getMarketData(
        symbol,
        convertedTimeframe,
        limit,
        sDate,
        eDate,
      );

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

  async getPaginatedData(symbol, convertedTimeframe, limit) {
    const allData = [];
    let oldestTime = null;

    while (allData.length < limit) {
      const batchSize = Math.min(1000, limit - allData.length);
      this.stats.recordRequest('Binance', convertedTimeframe);

      const batch = await this.binanceProvider.getMarketData(
        symbol,
        convertedTimeframe,
        batchSize,
        null,
        oldestTime ? oldestTime - 1 : null,
      );

      if (!batch || batch.length === 0) break;

      allData.unshift(...batch);
      oldestTime = batch[0].openTime;

      if (batch.length < batchSize) break;
    }

    return allData.slice(-limit);
  }
}

export { BinanceProvider };
