class ConfigurationBuilder {
  constructor(defaultConfig) {
    this.defaultConfig = defaultConfig;
  }

  createTradingConfig(symbol, timeframe = 'D', bars = 100, strategyPath = 'Multi-Provider Strategy') {
    return {
      symbol: symbol.toUpperCase(),
      timeframe,
      bars,
      strategy: strategyPath,
    };
  }

  generateChartConfig(tradingConfig, indicatorMetadata) {
    return {
      ui: this.buildUIConfig(tradingConfig),
      dataSource: this.buildDataSourceConfig(),
      chartLayout: this.buildLayoutConfig(),
      seriesConfig: {
        candlestick: {
          upColor: '#26a69a',
          downColor: '#ef5350',
          borderVisible: false,
          wickUpColor: '#26a69a',
          wickDownColor: '#ef5350',
        },
        series: this.buildSeriesConfig(indicatorMetadata),
      },
    };
  }

  buildUIConfig(tradingConfig) {
    return {
      title: `${tradingConfig.strategy} - ${tradingConfig.symbol}`,
      symbol: tradingConfig.symbol,
      timeframe: this.formatTimeframe(tradingConfig.timeframe),
      strategy: tradingConfig.strategy,
    };
  }

  buildDataSourceConfig() {
    return {
      url: 'chart-data.json',
      candlestickPath: 'candlestick',
      plotsPath: 'plots',
      timestampPath: 'timestamp',
    };
  }

  buildLayoutConfig() {
    return {
      main: { height: 400 },
      indicator: { height: 200 },
    };
  }

  buildSeriesConfig(indicators) {
    const series = {};

    Object.entries(indicators).forEach(([key, config]) => {
      series[key] = {
        color: config.color,
        lineWidth: 2,
        title: key,
        chart: this.determineChartType(key),
      };
    });

    return series;
  }

  determineChartType(key) {
    return key.includes('EMA') || key.includes('SMA') || key.includes('MA') ? 'main' : 'indicator';
  }

  formatTimeframe(timeframe) {
    const timeframes = {
      1: '1 Minute',
      5: '5 Minutes',
      10: '10 Minutes',
      15: '15 Minutes',
      30: '30 Minutes',
      60: '1 Hour',
      240: '4 Hours',
      D: 'Daily',
      W: 'Weekly',
      M: 'Monthly',
    };
    return timeframes[timeframe] || timeframe;
  }
}

export { ConfigurationBuilder };
