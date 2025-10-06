import { PineTS } from '../../../PineTS/dist/pinets.dev.es.js';

class PineScriptStrategyRunner {
  async createPineTSAdapter(provider, data, instance, symbol, timeframe, bars) {
    /* All providers now return consistent data arrays - create PineTS from data */
    const pineTS = new PineTS(data, symbol, timeframe, bars);
    await pineTS.ready();
    return pineTS;
  }

  async runEMAStrategy(pineTS) {
    /* Add signal calculation using proper PineTS syntax */
    const { plots } = await pineTS.run((context) => {
      const { close } = context.data;
      const { plot } = context.core;
      const ta = context.ta;

      // Basic EMA calculation
      const ema9 = ta.ema(close, 9);
      const ema18 = ta.ema(close, 18);

      // Bull signal using Pine Script style comparison - current values
      const bullSignal = ema9 > ema18 ? 1 : 0;

      // Plot calls - exact pattern from docs
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

  executeTranspiledStrategy(jsCode, marketData) {
    /* Create execution context with market data arrays and ta library stubs */
    const plots = [];

    const context = {
      data: {
        open: marketData.open || [],
        high: marketData.high || [],
        low: marketData.low || [],
        close: marketData.close || [],
        volume: marketData.volume || [],
      },
      ta: {
        ema: (src, len) => src,
        sma: (src, len) => src,
        rsi: (src, len) => src,
        stdev: (src, len) => 0,
        crossover: (a, b) => false,
        crossunder: (a, b) => false,
      },
      core: {
        plot: (series, title, options) => {
          plots.push({ title, series, options });
        },
      },
    };

    /* Provide global functions for Pine Script API */
    const globalScope = {
      context,
      plots,
      /* Pine Script declaration functions */
      indicator: (title, options) => {
        /* Indicator declaration - no-op in execution */
        return { title, ...options };
      },
      strategy: (title, options) => {
        /* Strategy declaration - no-op in execution */
        return { title, ...options };
      },
      /* Pine Script built-in variables */
      close: context.data.close,
      open: context.data.open,
      high: context.data.high,
      low: context.data.low,
      volume: context.data.volume,
      /* Pine Script plot function */
      plot: (series, title = '', options = {}) => {
        plots.push({ title, series, options });
      },
      plotshape: (series, title, options) => {
        plots.push({ title, series, options, type: 'shape' });
      },
      plotchar: (series, title, options) => {
        plots.push({ title, series, options, type: 'char' });
      },
    };

    /* Execute transpiled code with Function constructor */
    try {
      const paramNames = Object.keys(globalScope);
      const paramValues = Object.values(globalScope);
      // eslint-disable-next-line no-new-func
      const strategyFunc = new Function(...paramNames, jsCode + '\nreturn plots;');
      const result = strategyFunc(...paramValues);
      return { plots: result || plots };
    } catch (error) {
      throw new Error(`Strategy execution failed: ${error.message}`);
    }
  }
}

export { PineScriptStrategyRunner };
