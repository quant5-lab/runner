import { CHART_COLORS } from '../config.js';

class ConfigurationBuilder {
  constructor(defaultConfig) {
    this.defaultConfig = defaultConfig;
  }

  createTradingConfig(
    symbol,
    timeframe = 'D',
    bars = 100,
    strategyPath = 'Multi-Provider Strategy',
  ) {
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
      chartLayout: this.buildLayoutConfig(indicatorMetadata),
      seriesConfig: {
        candlestick: {
          upColor: CHART_COLORS.CANDLESTICK_UP,
          downColor: CHART_COLORS.CANDLESTICK_DOWN,
          borderVisible: false,
          wickUpColor: CHART_COLORS.CANDLESTICK_UP,
          wickDownColor: CHART_COLORS.CANDLESTICK_DOWN,
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

  buildLayoutConfig(indicatorMetadata) {
    const panes = { main: { height: 400, fixed: true } };

    if (indicatorMetadata) {
      const uniquePanes = new Set();
      Object.values(indicatorMetadata).forEach((metadata) => {
        const pane = metadata.chartPane;
        if (pane && pane !== 'main') {
          uniquePanes.add(pane);
        }
      });

      uniquePanes.forEach((paneName) => {
        panes[paneName] = { height: 200 };
      });
    }

    /* Backward compatibility: ensure 'indicator' pane exists if no dynamic panes */
    if (Object.keys(panes).length === 1) {
      panes.indicator = { height: 200 };
    }

    return panes;
  }

  buildSeriesConfig(indicators) {
    const series = {};

    Object.entries(indicators).forEach(([key, config]) => {
      const chartType = config.chartPane || 'indicator';
      const isMainChart = chartType === 'main';

      const finalColor = config.transp && config.transp > 0
        ? this.applyTransparency(config.color, config.transp)
        : config.color;

      series[key] = {
        color: finalColor,
        style: config.style || 'line',
        lineWidth: config.linewidth || 2,
        title: key,
        chart: chartType,
        lastValueVisible: !isMainChart,
        priceLineVisible: !isMainChart,
      };
    });

    return series;
  }

  determineChartType(key) {
    const mainChartPlots = ['Avg Price', 'Stop Level', 'Take Profit Level'];

    if (mainChartPlots.includes(key)) {
      return 'main';
    }

    if (key.includes('CAGR')) {
      return 'indicator';
    }

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

  applyTransparency(color, transp) {
    if (!transp || transp === 0) {
      return color;
    }

    const hexMatch = color.match(/^#([0-9A-Fa-f]{2})([0-9A-Fa-f]{2})([0-9A-Fa-f]{2})$/);
    if (hexMatch) {
      const r = parseInt(hexMatch[1], 16);
      const g = parseInt(hexMatch[2], 16);
      const b = parseInt(hexMatch[3], 16);
      const alpha = 1 - (transp / 100);
      return `rgba(${r}, ${g}, ${b}, ${alpha})`;
    }

    return color;
  }
}

export { ConfigurationBuilder };
