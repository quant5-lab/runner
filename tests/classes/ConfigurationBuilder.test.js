import { describe, it, expect, beforeEach } from 'vitest';
import { ConfigurationBuilder } from '../../src/classes/ConfigurationBuilder.js';

describe('ConfigurationBuilder', () => {
  let builder;
  const defaultConfig = {
    indicators: {
      EMA20: { period: 20, color: '#2196F3' },
      RSI: { period: 14, color: '#FF9800' },
    },
  };

  beforeEach(() => {
    builder = new ConfigurationBuilder(defaultConfig);
  });

  describe('constructor', () => {
    it('should store default configuration', () => {
      expect(builder.defaultConfig).toEqual(defaultConfig);
    });
  });

  describe('createTradingConfig()', () => {
    it('should create trading config with required parameters', () => {
      const config = builder.createTradingConfig('BTCUSDT');
      expect(config).toEqual({
        symbol: 'BTCUSDT',
        timeframe: 'D',
        bars: 100,
        strategy: 'Multi-Provider Strategy',
      });
    });

    it('should convert symbol to uppercase', () => {
      const config = builder.createTradingConfig('btcusdt');
      expect(config.symbol).toBe('BTCUSDT');
    });

    it('should use custom timeframe', () => {
      const config = builder.createTradingConfig('AAPL', 'W');
      expect(config.timeframe).toBe('W');
    });

    it('should use custom bars', () => {
      const config = builder.createTradingConfig('AAPL', 'D', 200);
      expect(config.bars).toBe(200);
    });

    it('should handle numeric timeframes', () => {
      const config = builder.createTradingConfig('AAPL', 60, 50);
      expect(config.timeframe).toBe(60);
      expect(config.bars).toBe(50);
    });
  });

  describe('formatTimeframe()', () => {
    it('should format numeric timeframes', () => {
      expect(builder.formatTimeframe(1)).toBe('1 Minute');
      expect(builder.formatTimeframe(5)).toBe('5 Minutes');
      expect(builder.formatTimeframe(15)).toBe('15 Minutes');
      expect(builder.formatTimeframe(60)).toBe('1 Hour');
      expect(builder.formatTimeframe(240)).toBe('4 Hours');
    });

    it('should format letter timeframes', () => {
      expect(builder.formatTimeframe('D')).toBe('Daily');
      expect(builder.formatTimeframe('W')).toBe('Weekly');
      expect(builder.formatTimeframe('M')).toBe('Monthly');
    });

    it('should return original value for unknown timeframe', () => {
      expect(builder.formatTimeframe('X')).toBe('X');
      expect(builder.formatTimeframe(999)).toBe(999);
    });
  });

  describe('determineChartType()', () => {
    it('should return main for moving averages', () => {
      expect(builder.determineChartType('EMA20')).toBe('main');
      expect(builder.determineChartType('SMA50')).toBe('main');
      expect(builder.determineChartType('MA100')).toBe('main');
    });

    it('should return indicator for non-moving averages', () => {
      expect(builder.determineChartType('RSI')).toBe('indicator');
      expect(builder.determineChartType('MACD')).toBe('main');
      expect(builder.determineChartType('Volume')).toBe('indicator');
    });
  });

  describe('buildUIConfig()', () => {
    it('should build UI configuration', () => {
      const tradingConfig = {
        strategy: 'Test Strategy',
        symbol: 'BTCUSDT',
        timeframe: 'D',
      };
      const uiConfig = builder.buildUIConfig(tradingConfig);
      expect(uiConfig).toEqual({
        title: 'Test Strategy - BTCUSDT',
        symbol: 'BTCUSDT',
        timeframe: 'Daily',
        strategy: 'Test Strategy',
      });
    });

    it('should format timeframe in UI config', () => {
      const tradingConfig = { strategy: 'Test', symbol: 'AAPL', timeframe: 'W' };
      const uiConfig = builder.buildUIConfig(tradingConfig);
      expect(uiConfig.timeframe).toBe('Weekly');
    });
  });

  describe('buildDataSourceConfig()', () => {
    it('should return data source configuration', () => {
      const config = builder.buildDataSourceConfig();
      expect(config).toEqual({
        url: 'chart-data.json',
        candlestickPath: 'candlestick',
        plotsPath: 'plots',
        timestampPath: 'timestamp',
      });
    });
  });

  describe('buildLayoutConfig()', () => {
    it('should return layout configuration', () => {
      const config = builder.buildLayoutConfig();
      expect(config).toEqual({
        main: { height: 400 },
        indicator: { height: 200 },
      });
    });
  });

  describe('buildSeriesConfig()', () => {
    it('should build series config from indicators', () => {
      const indicators = {
        EMA20: { color: '#2196F3' },
        RSI: { color: '#FF9800' },
      };
      const series = builder.buildSeriesConfig(indicators);
      expect(series).toEqual({
        EMA20: {
          color: '#2196F3',
          style: 'line',
          lineWidth: 2,
          title: 'EMA20',
          chart: 'main',
        },
        RSI: {
          color: '#FF9800',
          style: 'line',
          lineWidth: 2,
          title: 'RSI',
          chart: 'indicator',
        },
      });
    });

    it('should handle empty indicators', () => {
      const series = builder.buildSeriesConfig({});
      expect(series).toEqual({});
    });
  });

  describe('generateChartConfig()', () => {
    it('should generate complete chart configuration', () => {
      const tradingConfig = {
        strategy: 'Test Strategy',
        symbol: 'BTCUSDT',
        timeframe: 'D',
      };
      const indicators = {
        EMA20: { color: '#2196F3' },
      };

      const chartConfig = builder.generateChartConfig(tradingConfig, indicators);

      expect(chartConfig).toHaveProperty('ui');
      expect(chartConfig).toHaveProperty('dataSource');
      expect(chartConfig).toHaveProperty('chartLayout');
      expect(chartConfig).toHaveProperty('seriesConfig');

      expect(chartConfig.ui.symbol).toBe('BTCUSDT');
      expect(chartConfig.dataSource.url).toBe('chart-data.json');
      expect(chartConfig.chartLayout.main.height).toBe(400);
      expect(chartConfig.seriesConfig.candlestick.upColor).toBe('#26a69a');
    });

    it('should include candlestick configuration', () => {
      const config = builder.generateChartConfig(
        { strategy: 'Test', symbol: 'AAPL', timeframe: 'D' },
        {},
      );
      expect(config.seriesConfig.candlestick).toEqual({
        upColor: '#26a69a',
        downColor: '#ef5350',
        borderVisible: false,
        wickUpColor: '#26a69a',
        wickDownColor: '#ef5350',
      });
    });
  });
});
