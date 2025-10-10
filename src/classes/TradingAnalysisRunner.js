import { CHART_COLORS } from '../constants/chartColors.js';

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

  async runPineScriptStrategy(symbol, timeframe, bars, jsCode, strategyPath) {
    const runStartTime = performance.now();
    this.logger.log(`Configuration:\tSymbol=${symbol}, Timeframe=${timeframe}, Bars=${bars}`);

    const tradingConfig = this.configurationBuilder.createTradingConfig(symbol, timeframe, bars, strategyPath);

    const fetchStartTime = performance.now();
    this.logger.log(`Fetching data:\t${symbol} (${timeframe})`);

    const { provider, data, instance } = await this.providerManager.fetchMarketData(
      symbol,
      timeframe,
      bars,
    );

    const fetchDuration = (performance.now() - fetchStartTime).toFixed(2);
    this.logger.log(`Data source:\t${provider} (took ${fetchDuration}ms)`);

    const execStartTime = performance.now();
    
    this.logger.debug('=== TRANSPILED JAVASCRIPT CODE START ===');
    this.logger.debug(jsCode);
    this.logger.debug('=== TRANSPILED JAVASCRIPT CODE END ===');

    const executionResult = await this.pineScriptStrategyRunner.executeTranspiledStrategy(
      jsCode,
      data,
      symbol,
      timeframe,
    );
    const execDuration = (performance.now() - execStartTime).toFixed(2);
    this.logger.log(`Execution:\ttook ${execDuration}ms`);
    
    const plots = executionResult.plots || {};
    const indicatorMetadata = this.extractIndicatorMetadata(plots);

    if (!data?.length) {
      throw new Error(`No valid market data available for ${symbol}`);
    }

    const candlestickData = this.candlestickDataSanitizer.processCandlestickData(data);
    this.jsonFileWriter.exportChartData(candlestickData, plots);

    const chartConfig = this.configurationBuilder.generateChartConfig(
      tradingConfig,
      indicatorMetadata,
    );
    this.jsonFileWriter.exportConfiguration(chartConfig);

    const runDuration = (performance.now() - runStartTime).toFixed(2);
    this.logger.log(`Processing:\t${candlestickData.length} candles (took ${runDuration}ms)`);
  }

  async runDefaultStrategy(symbol, timeframe, bars) {
    const runStartTime = performance.now();
    this.logger.log(`Configuration:\tSymbol=${symbol}, Timeframe=${timeframe}, Bars=${bars}`);

    const tradingConfig = this.configurationBuilder.createTradingConfig(symbol, timeframe, bars, 'Multi-Provider Strategy');

    const fetchStartTime = performance.now();
    this.logger.log(`Fetching data:\t${symbol} (${timeframe})`);

    const { provider, data, instance } = await this.providerManager.fetchMarketData(
      symbol,
      timeframe,
      bars,
    );

    const fetchDuration = (performance.now() - fetchStartTime).toFixed(2);
    this.logger.log(`Data source:\t${provider} (took ${fetchDuration}ms)`);

    const emaResult = await this.pineScriptStrategyRunner.runEMAStrategy(data);
    const result = emaResult.result;
    const plots = emaResult.plots;
    const indicatorMetadata = this.pineScriptStrategyRunner.getIndicatorMetadata();

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

    const runDuration = (performance.now() - runStartTime).toFixed(2);
    this.logger.log(`Processing:\t${candlestickData.length} candles (took ${runDuration}ms)`);
  }

  processIndicatorPlots(result, data) {
    const processedPlots = {};

    const createTimestamp = (i, length) => {
      return data[i]?.openTime
        ? Math.floor(data[i].openTime / 1000)
        : Math.floor((Date.now() - (length - i - 1) * 86400000) / 1000);
    };

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

  extractIndicatorMetadata(plots) {
    const metadata = {};
    
    Object.keys(plots).forEach((plotKey) => {
      const color = this.extractPlotColor(plots[plotKey]);
      
      metadata[plotKey] = {
        color,
        title: plotKey,
        type: 'indicator',
      };
    });
    
    return metadata;
  }

  extractPlotColor(plotData) {
    if (!plotData?.data || !Array.isArray(plotData.data)) {
      return CHART_COLORS.DEFAULT_PLOT;
    }
    
    const firstPointWithColor = plotData.data.find(point => point?.options?.color);
    return firstPointWithColor?.options?.color || CHART_COLORS.DEFAULT_PLOT;
  }
}

export { TradingAnalysisRunner };
