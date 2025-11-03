import { CHART_COLORS, PLOT_COLOR_NAMES } from '../config.js';

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
    const restructuredPlots = this.restructurePlots(plots);
    
    /* Debug: Check plot timestamps */
    const indicatorMetadata = this.extractIndicatorMetadata(restructuredPlots);

    if (!data?.length) {
      throw new Error(`No valid market data available for ${symbol}`);
    }

    const candlestickData = this.candlestickDataSanitizer.processCandlestickData(data);
    this.jsonFileWriter.exportChartData(candlestickData, restructuredPlots);

    const chartConfig = this.configurationBuilder.generateChartConfig(
      tradingConfig,
      indicatorMetadata,
    );
    this.jsonFileWriter.exportConfiguration(chartConfig);

    const runDuration = (performance.now() - runStartTime).toFixed(2);
    this.logger.log(`Processing:\t${candlestickData.length} candles (took ${runDuration}ms)`);

    return executionResult;
  }

  /* Restructure PineTS plot output from single "Plot" array to named plots */
  restructurePlots(plots) {
    if (!plots || typeof plots !== 'object') {
      return {};
    }

    /* If already structured with multiple named plots, normalize timestamps */
    if (Object.keys(plots).length > 1 || !plots.Plot) {
      const normalized = {};
      Object.keys(plots).forEach((plotKey) => {
        normalized[plotKey] = {
          data: plots[plotKey].data?.map((point) => ({
            time: Math.floor(point.time / 1000),
            value: point.value,
            options: point.options,
          })) || [],
        };
      });
      return normalized;
    }

    const plotData = plots.Plot?.data;
    if (!Array.isArray(plotData) || plotData.length === 0) {
      return {};
    }

    /* Group by timestamp to find how many plots per candle */
    const timeMap = new Map();
    plotData.forEach((point) => {
      const timeKey = point.time;
      if (!timeMap.has(timeKey)) {
        timeMap.set(timeKey, []);
      }
      timeMap.get(timeKey).push(point);
    });

    /* Detect plot count per candle */
    const plotsPerCandle = timeMap.values().next().value?.length || 0;
    
    /* Create plot groups by position index (0, 1, 2, ...) */
    const plotGroups = [];
    for (let i = 0; i < plotsPerCandle; i++) {
      plotGroups.push({
        name: null,
        data: [],
        options: null,
      });
    }

    /* Assign data points to correct plot group by position */
    timeMap.forEach((pointsAtTime, timeKey) => {
      pointsAtTime.forEach((point, index) => {
        if (index < plotGroups.length) {
          plotGroups[index].data.push({
            time: Math.floor(timeKey / 1000),
            value: point.value,
            options: point.options,
          });
          
          /* Capture first non-null options for naming */
          if (!plotGroups[index].options && point.options) {
            plotGroups[index].options = point.options;
          }
        }
      });
    });

    /* Generate names based on options */
    const restructured = {};
    plotGroups.forEach((group, index) => {
      const plotName = this.generatePlotName(group.options || {}, index + 1);
      restructured[plotName] = {
        data: group.data,
      };
    });

    return restructured;
  }

  /* Generate plot name from options */
  generatePlotName(options, counter) {
    const color = options.color || '#000000';
    const style = options.style || 'line';
    const linewidth = options.linewidth || 1;
    
    const colorName = PLOT_COLOR_NAMES[color] || `Color${counter}`;
    
    /* Always include counter for uniqueness when no title */
    if (style === 'linebr' && linewidth === 2) {
      return `${colorName} Level ${counter}`;
    }
    
    if (style === 'linebr') {
      return `${colorName} Line ${counter}`;
    }
    
    return `${colorName} Plot ${counter}`;
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
