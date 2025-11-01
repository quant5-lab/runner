import { PineTS } from '../../../PineTS/dist/pinets.dev.es.js';
import TimeframeConverter from '../utils/timeframeConverter.js';
import { TimeframeParser } from '../utils/timeframeParser.js';
import { plotAdapterSource } from '../adapters/PinePlotAdapter.js';

class PineScriptStrategyRunner {
  constructor(providerManager, statsCollector) {
    this.providerManager = providerManager;
    this.statsCollector = statsCollector;
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
      const { close, open, high, low, volume } = context.data;
      const { plot: corePlot, color, na, nz, fixnan, time } = context.core;
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
      
      ${plotAdapterSource}
      
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

    await pineTS.prefetchSecurityData(wrappedCode);

    const result = await pineTS.run(wrappedCode);
    return { plots: result?.plots || [] };
  }
}

export { PineScriptStrategyRunner };
