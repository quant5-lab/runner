import { PineTS } from '../../../PineTS/dist/pinets.dev.es.js';

class PineScriptStrategyRunner {
  async createPineTSAdapter(provider, data, instance, symbol, timeframe, bars) {
    const pineTS = new PineTS(data, symbol, timeframe, bars);
    await pineTS.ready();
    return pineTS;
  }

  executeTranspiledStrategy(jsCode, marketData, pineTS) {
    /* Execute transpiled code with static context - don't use pineTS.run() */
    const plots = [];
    
    /* Static color object from PineTS Core class */
    const color = {
      red: 'red',
      green: 'green',
      blue: 'blue',
      yellow: 'yellow',
      white: 'white',
      black: 'black',
      gray: 'gray',
      lime: 'lime',
      maroon: 'maroon',
      orange: 'orange',
      purple: 'purple',
    };
    
    /* Stub ta object */
    const ta = {
      ema: (src, len) => src,
      sma: (src, len) => src,
      rsi: (src, len) => src,
      stdev: (src, len) => 0,
      crossover: (a, b) => false,
      crossunder: (a, b) => false,
    };
    
    /* Execute with Function constructor */
    const func = new Function(
      'ta',
      'color',
      'close',
      'open',
      'high',
      'low',
      'volume',
      'plot',
      'indicator',
      'strategy',
      jsCode
    );

    func(
      ta,
      color,
      marketData.close,
      marketData.open,
      marketData.high,
      marketData.low,
      marketData.volume,
      (series, title = '', options = {}) => plots.push({ title, series, options }),
      (title, options) => ({ title, ...options }),
      (title, options) => ({ title, ...options })
    );

    return { plots };
  }

  async runEMAStrategy(pineTS) {
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
