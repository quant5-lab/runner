import { PineTS } from '../../PineTS/dist/pinets.dev.es.js';

class ProviderManager {
    constructor(providerChain) {
        this.providerChain = providerChain;
    }

    async fetchMarketData(symbol, timeframe, bars) {
        for (let i = 0; i < this.providerChain.length; i++) {
            const { name, instance } = this.providerChain[i];
            
            try {
                const marketData = await this.executeProvider(name, instance, symbol, timeframe, bars);
                
                if (marketData?.length > 0) {
                    return { provider: name, data: marketData, instance };
                }
            } catch (error) {
                continue;
            }
        }
        
        throw new Error(`All providers failed for symbol: ${symbol}`);
    }

    async executeProvider(name, instance, symbol, timeframe, bars) {
        if (name === 'Binance') {
            const pineTS = new PineTS(instance, symbol, timeframe, bars);
            await pineTS.ready();
            return pineTS.data;
        }
        
        return await instance.getMarketData(symbol, timeframe, bars);
    }
}

export { ProviderManager };