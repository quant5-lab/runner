import { ConfigLoader            } from './ConfigLoader.js';
import { PaneAssigner            } from './PaneAssigner.js';
import { PaneManager             } from './PaneManager.js';
import { SeriesRouter            } from './SeriesRouter.js';
import { ChartManager            } from './ChartManager.js';
import { TradeDataFormatter,
         formatCurrency          } from './TradeTable.js';
import { TradeRowspanTransformer } from './TradeRowspanTransformer.js';
import { TradeRowspanRenderer    } from './TradeRowspanRenderer.js';
import { TradeMarkerBuilder      } from './TradeMarkerBuilder.js';
import { TimeIndexBuilder        } from './TimeIndexBuilder.js';
import { PlotOffsetTransformer   } from './PlotOffsetTransformer.js';
import { SeriesDataMapper        } from './SeriesDataMapper.js';
import { LineStyleConverter      } from './LineStyleConverter.js';
import { TradeSortToggle         } from './TradeSortToggle.js';
import { FullscreenToggle        } from './FullscreenToggle.js';
import { SortDirection           } from './SortDirection.js';
import { PaneResizeController   } from './PaneResizeController.js';
import { WindowResizeHandler    } from './WindowResizeHandler.js';

export class ChartApplication {
  constructor(chartOptions) {
    this.chartOptions          = chartOptions;
    this.paneManager           = null;
    this.seriesMap             = {};
    this.timeIndexBuilder      = new TimeIndexBuilder();
    this.plotOffsetTransformer = new PlotOffsetTransformer(this.timeIndexBuilder);
    this.seriesDataMapper      = new SeriesDataMapper();
    this.tradeMarkerBuilder    = new TradeMarkerBuilder();

    this.sortToggle = new TradeSortToggle(
      (direction) => this._rerenderTrades(direction),
      SortDirection.DESC,
    );

    this.fullscreenToggle    = new FullscreenToggle(
      document.querySelector('.chart-container'),
      () => this._onResize(),
    );

    this.resizeController    = new PaneResizeController();
    this.windowResizeHandler = new WindowResizeHandler(() => this._onResize());
  }

  async initialize() {
    const data = await ConfigLoader.loadChartData();

    const indicatorsWithPanes = new PaneAssigner(data.candlestick).assignAllPanes(data.indicators);

    this.updateMetadataDisplay(data.metadata);

    const paneConfig = this._buildPaneConfig(indicatorsWithPanes, data.ui?.panes);
    this.paneManager = new PaneManager(this.chartOptions);
    this._createCharts(paneConfig);
    this.resizeController.mount(this.paneManager.getAllPanes());

    const seriesRouter   = new SeriesRouter(this.paneManager, this.seriesMap);
    const candlestickData = this._routeAndLoadSeries(indicatorsWithPanes, data, seriesRouter);

    this._loadTradeMarkers(data.strategy, candlestickData);
    this._loadTrades(data.strategy, data.candlestick);
    this.updateTimestamp(data.metadata);
    this.paneManager.synchronizeTimeScales();

    setTimeout(() => {
      ChartManager.fitContent(this.paneManager.getAllCharts());
      ChartManager.handleResize(
        this.paneManager.getAllCharts(),
        this.paneManager.getAllContainers(),
      );
    }, 50);
  }

  async refresh() {
    this.resizeController.destroy();
    this.paneManager.getAllCharts().forEach(chart => chart.remove());
    this.paneManager.dynamicPanes.forEach(({ container }) => container.remove());
    this.paneManager.mainPane.container.innerHTML = '';

    this.seriesMap   = {};
    this.paneManager = null;

    await this.initialize();
  }

  updateMetadataDisplay(metadata) {
    if (!metadata) return;
    document.getElementById('chart-title').textContent      = metadata.title     || 'Financial Chart';
    document.getElementById('symbol-display').textContent   = metadata.symbol    || 'Unknown';
    document.getElementById('timeframe-display').textContent = metadata.timeframe || 'Unknown';
    document.getElementById('strategy-display').textContent = metadata.strategy  || 'Unknown';
  }

  updateTimestamp(metadata) {
    if (!metadata?.timestamp) return;
    document.getElementById('timestamp').textContent =
      'Last updated: ' + new Date(metadata.timestamp).toLocaleString();
  }

  _buildPaneConfig(indicatorsWithPanes, uiPanes) {
    const config = { main: { height: 400, fixed: true } };

    const dynamicPanes = new Set(
      Object.values(indicatorsWithPanes)
        .map(ind => ind.pane)
        .filter(pane => pane && pane !== 'main')
    );

    dynamicPanes.forEach(paneName => {
      config[paneName] = uiPanes?.[paneName] || { height: 200, fixed: false };
    });

    return config;
  }

  _createCharts(paneConfig) {
    const mainContainer = document.getElementById('main-chart');
    this.paneManager.createMainPane(mainContainer, paneConfig.main);

    this.fullscreenToggle.mount();

    for (const [paneName, config] of Object.entries(paneConfig)) {
      if (paneName !== 'main') {
        this.paneManager.createDynamicPane(paneName, config);
      }
    }
  }

  _routeAndLoadSeries(indicatorsWithPanes, data, seriesRouter) {
    const mainChart = this.paneManager.mainPane.chart;

    this.seriesMap.candlestick = ChartManager.addCandlestickSeries(mainChart, {
      upColor:       '#26a69a',
      downColor:     '#ef5350',
      borderVisible: false,
      wickUpColor:   '#26a69a',
      wickDownColor: '#ef5350',
    });

    const candlestickData = data.candlestick
      .sort((a, b) => a.time - b.time)
      .map(({ time, open, high, low, close }) => ({ time, open, high, low, close }));

    this.seriesMap.candlestick.setData(candlestickData);

    for (const [key, indicator] of Object.entries(indicatorsWithPanes)) {
      const color       = indicator.style?.color || '#2196F3';
      const seriesConfig = {
        color,
        lineWidth:             indicator.style?.lineWidth || 2,
        title:                 indicator.title || key,
        chart:                 indicator.pane  || 'main',
        style:                 indicator.style?.plotStyle || 'line',
        priceLineVisible:      false,
        lastValueVisible:      true,
        crosshairMarkerVisible: true,
        lineStyle:             LineStyleConverter.toNumeric(indicator.style?.lineStyle),
      };

      const series = seriesRouter.routeSeries(key, seriesConfig, ChartManager);
      if (!series) {
        console.error(`Failed to create series for '${key}'`);
        continue;
      }

      const offset    = indicator.offset || 0;
      const shifted   = this.plotOffsetTransformer.transform(indicator.data, offset, candlestickData);
      const colored   = this.seriesDataMapper.applyColorToData(shifted, color);
      const processed = window.adaptLineSeriesData(colored);

      if (processed.length > 0) {
        series.setData(processed);
        this._applySignalZoom(key, processed);
      }
    }

    return candlestickData;
  }

  _applySignalZoom(key, processedData) {
    if (key !== 'Buy Potential' && key !== 'Sell Potential') return;
    const valid = processedData.filter(p => p.value != null && !isNaN(p.value));
    if (!valid.length) return;
    const contextBars  = 50;
    const barInterval  = 3600;
    this.paneManager.mainPane.chart.timeScale().setVisibleRange({
      from: valid[0].time                    - contextBars * barInterval,
      to:   valid[valid.length - 1].time     + contextBars * barInterval,
    });
  }

  _loadTradeMarkers(strategy, candlestickData) {
    const markers = this.tradeMarkerBuilder.build(strategy, candlestickData);
    if (markers.length > 0) {
      this.seriesMap.candlestick.setMarkers(markers);
    }
  }

  _loadTrades(strategy, candlestickData) {
    if (!strategy) return;

    const openTrades = strategy.openTrades || [];
    const allTrades  = [
      ...(strategy.trades || []),
      ...openTrades.map(t => ({ ...t, status: 'open' })),
    ];

    allTrades.sort((a, b) => (b.entryTime || 0) - (a.entryTime || 0));

    this._tradeState = {
      allTrades,
      openTrades,
      netProfit: strategy.netProfit || 0,
      candlestickData,
    };

    const headerEl = document.querySelector('.trades-header');
    this.sortToggle.mount(headerEl, allTrades.length > 0);

    this._rerenderTrades(this.sortToggle.direction);
  }

  _rerenderTrades(sortDirection) {
    if (!this._tradeState) return;
    const { allTrades, openTrades, netProfit, candlestickData } = this._tradeState;

    const tbody   = document.getElementById('trades-tbody');
    const summary = document.getElementById('trades-summary');

    if (allTrades.length === 0) {
      tbody.innerHTML = '<tr><td colspan="7" class="no-trades">No trades to display</td></tr>';
      summary.textContent = 'No trades';
      return;
    }

    const sorted = sortDirection === SortDirection.ASC
      ? [...allTrades].reverse()
      : allTrades;

    const currentPrice = candlestickData?.length > 0
      ? candlestickData[candlestickData.length - 1].close
      : null;

    const formatter = new TradeDataFormatter(candlestickData);
    const rows      = new TradeRowspanTransformer(formatter).transformTrades(sorted, currentPrice, sortDirection);
    tbody.innerHTML = new TradeRowspanRenderer().renderRows(rows);

    const unrealizedProfit = currentPrice
      ? openTrades.reduce((sum, t) => {
          const multiplier = t.direction === 'long' ? 1 : -1;
          return sum + (currentPrice - t.entryPrice) * t.size * multiplier;
        }, 0)
      : 0;
    const totalProfit = netProfit + unrealizedProfit;
    const profitClass = totalProfit >= 0 ? 'trade-profit-positive' : 'trade-profit-negative';
    summary.innerHTML =
      `${allTrades.length} trades | Net P/L: <span class="${profitClass}">${formatCurrency(totalProfit)}</span>`;
  }

  _onResize() {
    if (!this.paneManager) return;
    ChartManager.handleResize(
      this.paneManager.getAllCharts(),
      this.paneManager.getAllContainers(),
    );
  }
}
