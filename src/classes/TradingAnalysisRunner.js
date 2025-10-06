class TradingAnalysisRunner {
  constructor(
    providerManager,
    pineScriptStrategyRunner,
    candlestickDataSanitizer,
    configurationBuilder,
    jsonFileWriter,
    logger,
  ) {
    this.providerManager = providerManager;
    this.pineScriptStrategyRunner = pineScriptStrategyRunner;
    this.candlestickDataSanitizer = candlestickDataSanitizer;
    this.configurationBuilder = configurationBuilder;
    this.jsonFileWriter = jsonFileWriter;
    this.logger = logger;
  }

  async run(symbol, timeframe, bars) {
    this.logger.log(`📊 Configuration: Symbol=${symbol}, Timeframe=${timeframe}, Bars=${bars}`);

    const tradingConfig = this.configurationBuilder.createTradingConfig(symbol, timeframe, bars);

    this.logger.log(
      `🎯 Attempting to fetch ${symbol} (${timeframe}) with dynamic provider fallback`,
    );

    const { provider, data, instance } = await this.providerManager.fetchMarketData(
      symbol,
      timeframe,
      bars,
    );

    this.logger.log(`📊 Using ${provider} provider for ${symbol}`);

    const pineTS = await this.pineScriptStrategyRunner.createPineTSAdapter(
      provider,
      data,
      instance,
      symbol,
      timeframe,
      bars,
    );

    const { result, plots } = await this.pineScriptStrategyRunner.runEMAStrategy(pineTS);
    const indicatorMetadata = this.pineScriptStrategyRunner.getIndicatorMetadata();

    // Process indicator plots - handle both custom providers and real PineTS
    let processedPlots = plots || {};

    if (result && Object.keys(processedPlots).length === 0) {
      processedPlots = this.processIndicatorPlots(result, data);
    }

    if (!data?.length) {
      throw new Error(`No valid market data available for ${symbol}`);
    }

    const candlestickData = this.candlestickDataSanitizer.processCandlestickData(data);

    this.jsonFileWriter.exportChartData(candlestickData, processedPlots);

    const chartConfig = this.configurationBuilder.generateChartConfig(
      tradingConfig,
      indicatorMetadata,
    );
    this.jsonFileWriter.exportConfiguration(chartConfig);

    this.logger.log(
      `Successfully processed ${candlestickData.length} candles for ${tradingConfig.symbol}`,
    );
  }

  processIndicatorPlots(result, data) {
    const processedPlots = {};

    // Ensure consistent time base for all indicators
    const createTimestamp = (i, length) => {
      return data[i]?.openTime
        ? Math.floor(data[i].openTime / 1000)
        : Math.floor((Date.now() - (length - i - 1) * 86400000) / 1000);
    };

    // For real PineTS, extract plot data from result
    if (result.ema9 && Array.isArray(result.ema9)) {
      processedPlots.EMA9 = {
        data: result.ema9.map((value, i) => ({
          time: createTimestamp(i, result.ema9.length),
          value: typeof value === 'number' ? value : 0,
        })),
      };
    }

    if (result.ema18 && Array.isArray(result.ema18)) {
      processedPlots.EMA18 = {
        data: result.ema18.map((value, i) => ({
          time: createTimestamp(i, result.ema18.length),
          value: typeof value === 'number' ? value : 0,
        })),
      };
    }

    if (result.bullSignal !== undefined) {
      // Handle both single value and array
      const bullValues = Array.isArray(result.bullSignal)
        ? result.bullSignal
        : data.map((_, i) => (i === data.length - 1 ? (result.bullSignal ? 1 : 0) : 0));

      processedPlots.BullSignal = {
        data: bullValues.map((value, i) => ({
          time: createTimestamp(i, bullValues.length),
          value:
            typeof value === 'boolean' ? (value ? 1 : 0) : typeof value === 'number' ? value : 0,
        })),
      };
    }

    return processedPlots;
  }
}

export { TradingAnalysisRunner };
