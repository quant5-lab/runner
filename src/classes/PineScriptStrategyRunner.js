import { PineTS } from '../../../PineTS/dist/pinets.dev.es.js';
import TimeframeConverter from '../utils/timeframeConverter.js';
import { TimeframeParser } from '../utils/timeframeParser.js';
import PineSecurityAdapter from './PineSecurityAdapter.js';
import { plotAdapterSource } from '../adapters/PinePlotAdapter.js';

class PineScriptStrategyRunner {
  constructor(providerManager) {
    this.providerManager = providerManager;
  }

  async executeTranspiledStrategy(jsCode, symbol, bars, timeframe) {
    const adapter = new PineSecurityAdapter(this.providerManager);

    const minutes = TimeframeParser.parseToMinutes(timeframe);
    const pineTSTimeframe = TimeframeConverter.toPineTS(minutes);
    const pineTS = new PineTS(
      adapter,
      symbol,
      pineTSTimeframe,
      bars,
      null,
      null,
    );

    const wrappedCode = `(context) => {
      const { close, open, high, low, volume } = context.data;
      const ta = context.ta;
      const request = context.request;
      const { plot: corePlot, color } = context.core;
      const tickerid = context.tickerId;
      
      ${plotAdapterSource}
      
      /* Pine Script version compatibility aliases
       * v3/v4: Used global functions sma(), security(), study()
       * v5: Uses namespaced functions ta.sma(), request.security(), indicator()
       * Reference: Pine Script Language Reference Manual v5 (tradingview.com/pine-script-reference/v5)
       * These function declarations bridge v3/v4 syntax to v5 PineTS runtime context */
      function security(sym, tf, expr) { return request.security(sym, tf, expr); }
      function sma(src, len) { return ta.sma(src, len); }
      function indicator() {}
      function strategy() {}
      function study() {}
      
      const yellow = 'yellow';
      const green = color.green;
      const red = color.red;
      const line = 'line';
      const linebr = 'linebr';
      
      ${jsCode}
    }`;

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
