class ProviderManager {
    constructor(providerChain) {
        this.providerChain = providerChain;
    }

    async fetchMarketData(symbol, timeframe, bars) {
        for (let i = 0; i < this.providerChain.length; i++) {
            const { name, instance } = this.providerChain[i];
            
            try {
                const marketData = await instance.getMarketData(symbol, timeframe, bars);
                
                if (marketData?.length > 0) {
                    return { provider: name, data: marketData, instance };
                }
            } catch (error) {
                continue;
            }
        }
        
        throw new Error(`All providers failed for symbol: ${symbol}`);
    }
}

export { ProviderManager };