import { PineTS, Provider } from '../../PineTS/dist/pinets.dev.es.js';

class TechnicalAnalysisEngine {
    async createPineTSAdapter(provider, data, instance, symbol, timeframe, bars) {
        if (provider === 'Binance') {
            // Binance: Use PineTS built-in Binance provider - exact pattern from docs
            const pineTS = new PineTS(Provider.Binance, symbol, timeframe, bars);
            await pineTS.ready();
            return pineTS;
        } else if (provider === 'MOEX') {
            // MOEX: Use existing MOEX provider instance
            const pineTS = new PineTS(instance, symbol, timeframe, bars);
            await pineTS.ready();
            return pineTS;
        } else if (provider === 'YahooFinance') {
            // Yahoo: Use existing Yahoo provider instance
            const pineTS = new PineTS(instance, symbol, timeframe, bars);
            await pineTS.ready();
            return pineTS;
        } else {
            // Fallback: Pass data array directly
            const pineTS = new PineTS(data, symbol, timeframe, bars);
            await pineTS.ready();
            return pineTS;
        }
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
            BullSignal: { title: 'Bull Signal', type: 'signal' }
        };
    }
}

export { TechnicalAnalysisEngine };