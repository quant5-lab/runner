import { PineTS } from '../../../PineTS/dist/pinets.dev.es.js';
import TimeframeConverter from '../utils/timeframeConverter.js';
import { TimeframeParser } from '../utils/timeframeParser.js';
import PineSecurityAdapter from './PineSecurityAdapter.js';
import { plotAdapterSource } from '../adapters/PinePlotAdapter.js';
import PineVersionMigrator from '../pine/PineVersionMigrator.js';

class PineScriptStrategyRunner {
  constructor(providerManager) {
    this.providerManager = providerManager;
  }

  parseSecurityCalls(jsCode) {
    /* Extract security() calls: request.security(symbol, timeframe, expression) */
    const regex = /request\.security\(\s*([^,]+)\s*,\s*['"]([^'"]+)['"]/g;
    const calls = [];
    let match;

    while ((match = regex.exec(jsCode)) !== null) {
      calls.push({
        symbolExpr: match[1].trim(),
        timeframe: match[2],
      });
    }

    console.log('!!! PARSED SECURITY CALLS:', calls);
    return calls;
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

    /* Parse and prefetch security() data */
    const securityCalls = this.parseSecurityCalls(jsCode);
    const sourceDurationMinutes = bars * minutes;
    
    const prefetchData = securityCalls.map(call => {
      const resolvedSymbol = call.symbolExpr === 'syminfo.tickerid' 
        ? symbol 
        : call.symbolExpr;
      
      /* Calculate correct limit based on duration */
      const targetMinutes = TimeframeParser.parseToMinutes(call.timeframe);
      const targetLimit = Math.ceil(sourceDurationMinutes / targetMinutes);
      
      console.log(`!!! DURATION CALC: ${bars} bars × ${minutes}m = ${sourceDurationMinutes}m → ${call.timeframe} (${targetMinutes}m) = ${targetLimit} candles`);
      
      return {
        symbol: resolvedSymbol,
        timeframe: call.timeframe,
        limit: targetLimit,
      };
    });

    console.log('!!! PREFETCH DATA:', prefetchData);
    if (prefetchData.length > 0) {
      await pineTS.prefetchSecurityData(prefetchData);
      console.log('!!! PREFETCH COMPLETE');
    }

    const wrappedCode = `(context) => {
      const { close, open, high, low, volume } = context.data;
      const { plot: corePlot, color } = context.core;
      const ta = context.ta;
      const request = context.request;
      const syminfo = context.syminfo;
      
      ${plotAdapterSource}
      
      function indicator() {}
      function strategy() {}
      
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
