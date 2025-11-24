import { ConfigLoader } from './ConfigLoader.js';
import { PaneAssigner } from './PaneAssigner.js';
import { PaneManager } from './PaneManager.js';
import { SeriesRouter } from './SeriesRouter.js';
import { ChartManager } from './ChartManager.js';
import { TradeDataFormatter, TradeTableRenderer } from './TradeTable.js';

/* Main application orchestrator (SRP, DIP) */
export class ChartApplication {
  constructor(chartOptions) {
    this.chartOptions = chartOptions;
    this.paneManager = null;
    this.seriesMap = {};
  }

  async initialize() {
    const data = await ConfigLoader.loadChartData();
    const configOverride = await ConfigLoader.loadStrategyConfig(
      data.metadata?.strategy || 'strategy'
    );

    const paneAssigner = new PaneAssigner(data.candlestick);
    const indicatorsWithPanes = paneAssigner.assignAllPanes(
      data.indicators,
      configOverride
    );

    // Merge config style/color overrides into indicators
    if (configOverride) {
      Object.entries(indicatorsWithPanes).forEach(([key, indicator]) => {
        const override = configOverride[key];
        if (override && typeof override === 'object') {
          if (override.style) indicator.style = { ...indicator.style, ...override };
        }
      });
    }

    this.updateMetadataDisplay(data.metadata);

    const paneConfig = this.buildPaneConfig(indicatorsWithPanes, data.ui?.panes);
    
    this.paneManager = new PaneManager(this.chartOptions);
    this.createCharts(paneConfig);

    const seriesRouter = new SeriesRouter(this.paneManager, this.seriesMap);
    this.routeAndLoadSeries(indicatorsWithPanes, data, seriesRouter, configOverride);

    this.loadTrades(data.strategy, data.candlestick);
    this.updateTimestamp(data.metadata);

    this.setupEventListeners();
    this.paneManager.synchronizeTimeScales();

    setTimeout(() => {
      ChartManager.fitContent(this.paneManager.getAllCharts());
    }, 50);
  }

  buildPaneConfig(indicatorsWithPanes, uiPanes) {
    const config = {
      main: { height: 400, fixed: true },
    };

    const hasIndicatorPane = Object.values(indicatorsWithPanes).some(
      (ind) => ind.pane === 'indicator'
    );

    if (hasIndicatorPane) {
      config.indicator = uiPanes?.indicator || { height: 200, fixed: false };
    }

    return config;
  }

  createCharts(paneConfig) {
    const mainContainer = document.getElementById('main-chart');
    this.paneManager.createMainPane(mainContainer, paneConfig.main);

    Object.entries(paneConfig).forEach(([paneName, config]) => {
      if (paneName !== 'main') {
        this.paneManager.createDynamicPane(paneName, config);
      }
    });
  }

  routeAndLoadSeries(indicatorsWithPanes, data, seriesRouter, configOverride) {
    const mainChart = this.paneManager.mainPane.chart;

    this.seriesMap.candlestick = ChartManager.addCandlestickSeries(mainChart, {
      upColor: '#26a69a',
      downColor: '#ef5350',
      borderVisible: false,
      wickUpColor: '#26a69a',
      wickDownColor: '#ef5350',
    });

    const candlestickData = data.candlestick
      .sort((a, b) => a.time - b.time)
      .map((c) => ({
        time: c.time,
        open: c.open,
        high: c.high,
        low: c.low,
        close: c.close,
      }));

    this.seriesMap.candlestick.setData(candlestickData);

    Object.entries(indicatorsWithPanes).forEach(([key, indicator]) => {
      // Extract style from config override
      const styleType = configOverride?.[key]?.style || 'line';
      const color = indicator.style?.color || configOverride?.[key]?.color || '#2196F3';
      
      const seriesConfig = {
        color: color,
        lineWidth: indicator.style?.lineWidth || 2,
        title: indicator.title || key,
        chart: indicator.pane || 'main',
        style: styleType,
      };

      const series = seriesRouter.routeSeries(key, seriesConfig, ChartManager);
      
      if (!series) {
        console.error(`Failed to create series for '${key}'`);
        return;
      }

      const dataWithColor = indicator.data.map((point) => ({
        ...point,
        options: { color: color },
      }));

      const processedData = window.adaptLineSeriesData(dataWithColor);
      if (processedData.length > 0) {
        series.setData(processedData);
      }
    });
  }

  loadTrades(strategy, candlestickData) {
    if (!strategy) return;

    const allTrades = [
      ...(strategy.trades || []),
      ...(strategy.openTrades || []).map((t) => ({ ...t, status: 'open' })),
    ];

    const tbody = document.getElementById('trades-tbody');
    const summary = document.getElementById('trades-summary');

    if (allTrades.length === 0) {
      tbody.innerHTML =
        '<tr><td colspan="7" class="no-trades">No trades to display</td></tr>';
      summary.textContent = 'No trades';
      return;
    }

    const currentPrice = candlestickData?.length > 0 
      ? candlestickData[candlestickData.length - 1].close 
      : null;

    const formatter = new TradeDataFormatter(candlestickData);
    const renderer = new TradeTableRenderer(formatter);
    tbody.innerHTML = renderer.renderRows(allTrades, currentPrice);

    const realizedProfit = strategy.netProfit || 0;
    const unrealizedProfit = currentPrice
      ? (strategy.openTrades || []).reduce((sum, trade) => {
          const multiplier = trade.direction === 'long' ? 1 : -1;
          return sum + (currentPrice - trade.entryPrice) * trade.size * multiplier;
        }, 0)
      : 0;
    const totalProfit = realizedProfit + unrealizedProfit;
    
    const profitClass =
      totalProfit >= 0 ? 'trade-profit-positive' : 'trade-profit-negative';
    summary.innerHTML = `${allTrades.length} trades | Net P/L: <span class="${profitClass}">$${totalProfit.toFixed(
      2
    )}</span>`;
  }

  updateMetadataDisplay(metadata) {
    if (!metadata) return;

    document.getElementById('chart-title').textContent =
      metadata.title || 'Financial Chart';
    document.getElementById('symbol-display').textContent =
      metadata.symbol || 'Unknown';
    document.getElementById('timeframe-display').textContent =
      metadata.timeframe || 'Unknown';
    document.getElementById('strategy-display').textContent =
      metadata.strategy || 'Unknown';
  }

  updateTimestamp(metadata) {
    if (!metadata?.timestamp) return;

    document.getElementById('timestamp').textContent =
      'Last updated: ' + new Date(metadata.timestamp).toLocaleString();
  }

  setupEventListeners() {
    window.addEventListener('resize', () => {
      const containers = this.paneManager.getAllContainers();
      const charts = this.paneManager.getAllCharts();
      ChartManager.handleResize(charts, containers);
    });
  }

  async refresh() {
    // Clear all charts and containers
    const charts = this.paneManager.getAllCharts();
    charts.forEach(chart => chart.remove());
    
    const containers = this.paneManager.getAllContainers();
    containers.forEach((container) => {
      container.innerHTML = '';
    });

    this.seriesMap = {};
    this.paneManager = null;
    
    await this.initialize();
  }
}
