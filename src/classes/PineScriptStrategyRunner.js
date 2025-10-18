import { PineTS } from '../../../PineTS/dist/pinets.dev.es.js';
import TimeframeConverter from '../utils/timeframeConverter.js';
import { TimeframeParser } from '../utils/timeframeParser.js';
import { plotAdapterSource } from '../adapters/PinePlotAdapter.js';

class PineScriptStrategyRunner {
  constructor(providerManager, statsCollector) {
    this.providerManager = providerManager;
    this.statsCollector = statsCollector;
  }

  async executeTranspiledStrategy(jsCode, symbol, bars, timeframe) {
    const minutes = TimeframeParser.parseToMinutes(timeframe);
    const pineTSTimeframe = TimeframeConverter.toPineTS(minutes);
    const pineTS = new PineTS(this.providerManager, symbol, pineTSTimeframe, bars, null, null);

    const wrappedCode = `(context) => {
      const { close, open, high, low, volume } = context.data;
      const { plot: corePlot, color, na, nz } = context.core;
      const ta = context.ta;
      const math = context.math;
      const request = context.request;
      const syminfo = context.syminfo;
      const dayofweek = context.dayofweek;
      
      ${plotAdapterSource}
      
      function indicator() {}
      function strategy() {}
      
      ${jsCode}
    }`;

    await pineTS.prefetchSecurityData(wrappedCode);

    const result = await pineTS.run(wrappedCode);
    return { plots: result?.plots || [] };
  }

  async runEMAStrategy(data) {
    const pineTS = new PineTS(data);

    const { plots } = await pineTS.run((context) => {
      const { close } = context.data;
      const { plot } = context.core;
      const ta = context.ta;

      const ema9 = ta.ema(close, 9);
      const ema18 = ta.ema(close, 18);

      const bullSignal = ema9 > ema18 ? 1 : 0;

      plot(ema9, 'EMA9', { style: 'line', linewidth: 2, color: 'blue' });
      plot(ema18, 'EMA18', { style: 'line', linewidth: 2, color: 'red' });
      plot(bullSignal, 'BullSignal', { style: 'line', linewidth: 1, color: 'green' });
    });

    return { result: plots, plots: plots || {} };
  }

  getIndicatorMetadata() {
    return {
      EMA9: { title: 'EMA 9', type: 'moving_average' },
      EMA18: { title: 'EMA 18', type: 'moving_average' },
      BullSignal: { title: 'Bull Signal', type: 'signal' },
    };
  }
}

export { PineScriptStrategyRunner };
