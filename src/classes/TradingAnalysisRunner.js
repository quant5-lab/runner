import { CHART_COLORS } from '../config.js';

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

  async runPineScriptStrategy(symbol, timeframe, bars, jsCode, strategyPath, settings = null) {
    const runStartTime = performance.now();
    this.logger.log(`Configuration:\tSymbol=${symbol}, Timeframe=${timeframe}, Bars=${bars}`);

    const tradingConfig = this.configurationBuilder.createTradingConfig(
      symbol,
      timeframe,
      bars,
      strategyPath,
    );

    const fetchStartTime = performance.now();
    this.logger.log(`Fetching data:\t${symbol} (${timeframe})`);

    const { provider, data } = await this.providerManager.fetchMarketData(symbol, timeframe, bars);

    const fetchDuration = (performance.now() - fetchStartTime).toFixed(2);
    this.logger.log(`Data source:\t${provider} (took ${fetchDuration}ms)`);

    const execStartTime = performance.now();

    this.logger.debug('=== TRANSPILED JAVASCRIPT CODE START ===');
    this.logger.debug(jsCode);
    this.logger.debug('=== TRANSPILED JAVASCRIPT CODE END ===');

    const executionResult = await this.pineScriptStrategyRunner.executeTranspiledStrategy(
      jsCode,
      symbol,
      bars,
      timeframe,
      settings,
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

    return executionResult;
  }

  extractIndicatorMetadata(plots) {
    const metadata = {};

    Object.keys(plots).forEach((plotKey) => {
      const color = this.extractPlotColor(plots[plotKey]);
      const style = this.extractPlotStyle(plots[plotKey]);
      const linewidth = this.extractPlotLineWidth(plots[plotKey]);
      const transp = this.extractPlotTransp(plots[plotKey]);

      metadata[plotKey] = {
        color,
        style,
        linewidth,
        transp,
        title: plotKey,
        type: 'indicator',
        chartPane: this.determineChartPane(plotKey),
      };
    });

    return metadata;
  }

  determineChartPane(plotKey) {
    const mainChartPlots = ['Avg Price', 'Stop Level', 'Take Profit Level', 'Support', 'Resistance'];

    if (mainChartPlots.includes(plotKey)) {
      return 'main';
    }

    if (plotKey.includes('CAGR')) {
      return 'indicator';
    }

    return plotKey.includes('EMA') || plotKey.includes('SMA') || plotKey.includes('MA') ? 'main' : 'indicator';
  }

  extractPlotColor(plotData) {
    if (!plotData?.data || !Array.isArray(plotData.data)) {
      return CHART_COLORS.DEFAULT_PLOT;
    }

    const firstPointWithColor = plotData.data.find((point) => point?.options?.color);
    const rawColor = firstPointWithColor?.options?.color || CHART_COLORS.DEFAULT_PLOT;
    return this.normalizeRgbaAlpha(rawColor);
  }

  normalizeRgbaAlpha(color) {
    // PineTS outputs rgba with alpha 0-100, lightweight-charts needs 0-1
    const rgbaMatch = color.match(/^rgba\((\d+),\s*(\d+),\s*(\d+),\s*(\d+)\)$/);
    if (rgbaMatch) {
      const [, r, g, b, a] = rgbaMatch;
      const alphaValue = parseInt(a);
      if (alphaValue > 1) {
        // Convert from 0-100 to 0-1
        return `rgba(${r}, ${g}, ${b}, ${alphaValue / 100})`;
      }
    }
    return color;
  }

  extractPlotStyle(plotData) {
    if (!plotData?.data || !Array.isArray(plotData.data)) {
      return 'line';
    }

    const firstPointWithStyle = plotData.data.find((point) => point?.options?.style);
    return firstPointWithStyle?.options?.style || 'line';
  }

  extractPlotLineWidth(plotData) {
    if (!plotData?.data || !Array.isArray(plotData.data)) {
      return 2;
    }

    const firstPointWithWidth = plotData.data.find((point) => point?.options?.linewidth);
    return firstPointWithWidth?.options?.linewidth || 2;
  }

  extractPlotTransp(plotData) {
    if (!plotData?.data || !Array.isArray(plotData.data)) {
      return 0;
    }

    const firstPointWithTransp = plotData.data.find((point) => point?.options?.transp !== undefined);
    return firstPointWithTransp?.options?.transp ?? 0;
  }
}

export { TradingAnalysisRunner };
