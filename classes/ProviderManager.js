class ProviderManager {
  constructor(providerChain, logger) {
    this.providerChain = providerChain;
    this.logger = logger;
  }

  async fetchMarketData(symbol, timeframe, bars) {
    for (let i = 0; i < this.providerChain.length; i++) {
      const { name, instance } = this.providerChain[i];

      this.logger.log(`🔍 Trying ${name} provider for ${symbol}`);

      try {
        const marketData = await instance.getMarketData(symbol, timeframe, bars);

        if (marketData?.length > 0) {
          this.logger.log(`✅ ${name} provider succeeded for ${symbol}`);
          return { provider: name, data: marketData, instance };
        }

        this.logger.log(`➡️  Symbol ${symbol} not found in ${name} provider`);
      } catch (error) {
        this.logger.log(`➡️  Symbol ${symbol} not found in ${name} provider`);
        this.logger.debug(`Error from ${name} provider: ${error}`);
        continue;
      }
    }

    throw new Error(`All providers failed for symbol: ${symbol}`);
  }
}

export { ProviderManager };
