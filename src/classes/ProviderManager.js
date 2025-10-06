class ProviderManager {
  constructor(providerChain, logger) {
    this.providerChain = providerChain;
    this.logger = logger;
  }

  async fetchMarketData(symbol, timeframe, bars) {
    for (let i = 0; i < this.providerChain.length; i++) {
      const { name, instance } = this.providerChain[i];

      const providerStartTime = performance.now();
      this.logger.log(`Attempting:\t${name} > ${symbol}`);

      try {
        const marketData = await instance.getMarketData(symbol, timeframe, bars);

        if (marketData?.length > 0) {
          const providerDuration = (performance.now() - providerStartTime).toFixed(2);
          this.logger.log(
            `Found data:\t${name} (${marketData.length} candles, took ${providerDuration}ms)`,
          );
          return { provider: name, data: marketData, instance };
        }

        this.logger.log(`No data:\t${name} > ${symbol}`);
      } catch (error) {
        this.logger.log(`Failed:\t\t${name} > ${symbol}`);
        this.logger.debug(`Error from ${name} provider: ${error}`);
        continue;
      }
    }

    throw new Error(`All providers failed for symbol: ${symbol}`);
  }
}

export { ProviderManager };
