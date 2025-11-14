import { PineTS } from '../../../PineTS/dist/pinets.dev.es.js';
import TimeframeConverter from '../utils/timeframeConverter.js';
import { TimeframeParser } from '../utils/timeframeParser.js';

class PineScriptStrategyRunner {
  constructor(providerManager, statsCollector, logger) {
    this.providerManager = providerManager;
    this.statsCollector = statsCollector;
    this.logger = logger;
  }

  async executeTranspiledStrategy(jsCode, symbol, bars, timeframe, settings = null) {
    const minutes = TimeframeParser.parseToMinutes(timeframe);
    const pineTSTimeframe = TimeframeConverter.toPineTS(minutes);
    const constructorOptions = settings ? { inputOverrides: settings } : undefined;
    const pineTS = new PineTS(
      this.providerManager,
      symbol,
      pineTSTimeframe,
      bars,
      null,
      null,
      constructorOptions,
    );

    const wrappedCode = `(context) => {
      const { close, open, high, low, volume, bar_index, last_bar_index } = context.data;
      const { plot, color, na, nz, fixnan, time } = context.core;
      const ta = context.ta;
      const math = context.math;
      const request = context.request;
      const input = context.input;
      const strategy = context.strategy;
      const syminfo = context.syminfo;
      const barmerge = context.barmerge;
      const format = context.format;
      const scale = context.scale;
      const timeframe = context.timeframe;
      const barstate = context.barstate;
      const dayofweek = context.dayofweek;
      
      plot.style_line = 'line';
      plot.style_histogram = 'histogram';
      plot.style_cross = 'cross';
      plot.style_area = 'area';
      plot.style_columns = 'columns';
      plot.style_circles = 'circles';
      plot.style_linebr = 'linebr';
      plot.style_stepline = 'stepline';
      
      function indicator() {}
      
      ${jsCode}
    }`;

    this.logger.debug('=== WRAPPED CODE FOR PINETS START ===');
    this.logger.debug(wrappedCode);
    this.logger.debug('=== WRAPPED CODE FOR PINETS END ===');

    await pineTS.prefetchSecurityData(wrappedCode);

    const result = await pineTS.run(wrappedCode);

    /* Extract strategy data if available */
    const strategyData = {};
    if (result?.strategy) {
      strategyData.trades = result.strategy.tradeHistory?.getClosedTrades() || [];
      strategyData.openTrades = result.strategy.tradeHistory?.getOpenTrades() || [];
      strategyData.equity = result.strategy.equityCalculator?.getEquity() || 0;
      strategyData.netProfit = result.strategy.equityCalculator?.getNetProfit() || 0;
    }

    return {
      plots: result?.plots || [],
      strategy: strategyData,
    };
  }
}

export { PineScriptStrategyRunner };
