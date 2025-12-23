import { describe, it, expect, beforeEach } from 'vitest';
import { ConfigurationBuilder } from '../../src/classes/ConfigurationBuilder.js';
import { DEFAULTS } from '../../src/config.js';

describe('ConfigurationBuilder - Dynamic Pane Configuration', () => {
  let builder;

  beforeEach(() => {
    builder = new ConfigurationBuilder(DEFAULTS);
  });

  describe('buildLayoutConfig() - Dynamic Pane Extraction', () => {
    it('should extract unique pane names from indicator metadata', () => {
      const metadata = {
        'ADX #1': { chartPane: 'adx_primary' },
        'DI+ #1': { chartPane: 'adx_primary' },
        'ADX #2': { chartPane: 'adx_secondary' },
        'Buy Signal': { chartPane: 'signals' },
      };

      const config = builder.buildLayoutConfig(metadata);

      expect(config).toHaveProperty('main');
      expect(config).toHaveProperty('adx_primary');
      expect(config).toHaveProperty('adx_secondary');
      expect(config).toHaveProperty('signals');
      expect(Object.keys(config)).toHaveLength(4);
    });

    it('should deduplicate repeated pane names', () => {
      const metadata = {
        'Indicator A': { chartPane: 'custom_pane' },
        'Indicator B': { chartPane: 'custom_pane' },
        'Indicator C': { chartPane: 'custom_pane' },
        'Indicator D': { chartPane: 'another_pane' },
      };

      const config = builder.buildLayoutConfig(metadata);

      expect(config).toEqual({
        main: { height: 400, fixed: true },
        custom_pane: { height: 200 },
        another_pane: { height: 200 },
      });
    });

    it('should ignore main pane in custom pane extraction', () => {
      const metadata = {
        'SMA 20': { chartPane: 'main' },
        'EMA 50': { chartPane: 'main' },
        'RSI': { chartPane: 'oscillators' },
      };

      const config = builder.buildLayoutConfig(metadata);

      expect(config).toEqual({
        main: { height: 400, fixed: true },
        oscillators: { height: 200 },
      });
    });

    it('should handle indicators with null or undefined chartPane', () => {
      const metadata = {
        'Indicator A': { chartPane: null },
        'Indicator B': { chartPane: undefined },
        'Indicator C': {},
        'Indicator D': { chartPane: 'custom' },
      };

      const config = builder.buildLayoutConfig(metadata);

      expect(config).toEqual({
        main: { height: 400, fixed: true },
        custom: { height: 200 },
      });
    });

    it('should handle empty string chartPane', () => {
      const metadata = {
        'Indicator A': { chartPane: '' },
        'Indicator B': { chartPane: 'valid_pane' },
      };

      const config = builder.buildLayoutConfig(metadata);

      expect(config).toEqual({
        main: { height: 400, fixed: true },
        valid_pane: { height: 200 },
      });
    });

    it('should handle indicators with only main chartPane', () => {
      const metadata = {
        'SMA 20': { chartPane: 'main' },
        'BB Upper': { chartPane: 'main' },
        'BB Lower': { chartPane: 'main' },
      };

      const config = builder.buildLayoutConfig(metadata);

      expect(config).toEqual({
        main: { height: 400, fixed: true },
        indicator: { height: 200 },
      });
    });

    it('should ensure backward compatibility with default indicator pane', () => {
      const config = builder.buildLayoutConfig();

      expect(config).toEqual({
        main: { height: 400, fixed: true },
        indicator: { height: 200 },
      });
    });

    it('should handle empty metadata object', () => {
      const config = builder.buildLayoutConfig({});

      expect(config).toEqual({
        main: { height: 400, fixed: true },
        indicator: { height: 200 },
      });
    });

    it('should handle large number of unique panes', () => {
      const metadata = {};
      for (let i = 1; i <= 10; i++) {
        metadata[`Indicator ${i}`] = { chartPane: `pane_${i}` };
      }

      const config = builder.buildLayoutConfig(metadata);

      expect(Object.keys(config)).toHaveLength(11); // main + 10 custom panes
      expect(config).toHaveProperty('main');
      for (let i = 1; i <= 10; i++) {
        expect(config).toHaveProperty(`pane_${i}`);
        expect(config[`pane_${i}`]).toEqual({ height: 200 });
      }
    });

    it('should preserve main pane properties when custom panes exist', () => {
      const metadata = {
        'Custom Indicator': { chartPane: 'custom' },
      };

      const config = builder.buildLayoutConfig(metadata);

      expect(config.main).toEqual({
        height: 400,
        fixed: true,
      });
    });

    it('should handle mixed case pane names', () => {
      const metadata = {
        'Indicator A': { chartPane: 'CustomPane' },
        'Indicator B': { chartPane: 'UPPERCASE' },
        'Indicator C': { chartPane: 'lowercase' },
        'Indicator D': { chartPane: 'Mixed_Case_123' },
      };

      const config = builder.buildLayoutConfig(metadata);

      expect(config).toHaveProperty('CustomPane');
      expect(config).toHaveProperty('UPPERCASE');
      expect(config).toHaveProperty('lowercase');
      expect(config).toHaveProperty('Mixed_Case_123');
    });

    it('should handle special characters in pane names', () => {
      const metadata = {
        'Indicator A': { chartPane: 'pane-with-dashes' },
        'Indicator B': { chartPane: 'pane_with_underscores' },
        'Indicator C': { chartPane: 'pane.with.dots' },
      };

      const config = builder.buildLayoutConfig(metadata);

      expect(config).toHaveProperty('pane-with-dashes');
      expect(config).toHaveProperty('pane_with_underscores');
      expect(config).toHaveProperty('pane.with.dots');
    });
  });

  describe('buildLayoutConfig() - Integration with generateChartConfig()', () => {
    it('should generate complete chart config with dynamic panes', () => {
      const tradingConfig = {
        symbol: 'BTCUSDT',
        timeframe: '1h',
        bars: 100,
        strategy: 'test-strategy',
      };

      const indicatorMetadata = {
        'ADX': { chartPane: 'adx_pane', color: '#FF9800', linewidth: 2 },
        'RSI': { chartPane: 'oscillators', color: '#2196F3', linewidth: 1 },
        'Volume': { chartPane: 'volume', color: '#4CAF50', linewidth: 2 },
      };

      const chartConfig = builder.generateChartConfig(tradingConfig, indicatorMetadata);

      expect(chartConfig.chartLayout).toHaveProperty('main');
      expect(chartConfig.chartLayout).toHaveProperty('adx_pane');
      expect(chartConfig.chartLayout).toHaveProperty('oscillators');
      expect(chartConfig.chartLayout).toHaveProperty('volume');

      expect(chartConfig.seriesConfig.series['ADX'].chart).toBe('adx_pane');
      expect(chartConfig.seriesConfig.series['RSI'].chart).toBe('oscillators');
      expect(chartConfig.seriesConfig.series['Volume'].chart).toBe('volume');
    });

    it('should handle metadata with no chartPane property', () => {
      const tradingConfig = {
        symbol: 'TEST',
        timeframe: 'D',
        bars: 50,
        strategy: 'test',
      };

      const indicatorMetadata = {
        'SMA 20': { color: '#2196F3', linewidth: 2 },
        'EMA 50': { color: '#FF9800', linewidth: 2 },
      };

      const chartConfig = builder.generateChartConfig(tradingConfig, indicatorMetadata);

      expect(chartConfig.chartLayout).toEqual({
        main: { height: 400, fixed: true },
        indicator: { height: 200 },
      });
    });
  });

  describe('buildLayoutConfig() - Edge Cases', () => {
    it('should handle null metadata', () => {
      const config = builder.buildLayoutConfig(null);

      expect(config).toEqual({
        main: { height: 400, fixed: true },
        indicator: { height: 200 },
      });
    });

    it('should handle undefined metadata', () => {
      const config = builder.buildLayoutConfig(undefined);

      expect(config).toEqual({
        main: { height: 400, fixed: true },
        indicator: { height: 200 },
      });
    });

    it('should handle metadata with non-object values', () => {
      const metadata = {
        'Valid Indicator': { chartPane: 'custom' },
        'Invalid 1': null,
        'Invalid 2': undefined,
        'Invalid 3': 'string',
        'Invalid 4': 123,
      };

      const config = builder.buildLayoutConfig(metadata);

      expect(config).toEqual({
        main: { height: 400, fixed: true },
        custom: { height: 200 },
      });
    });

    it('should handle very long pane names', () => {
      const longPaneName = 'a'.repeat(100);
      const metadata = {
        'Indicator': { chartPane: longPaneName },
      };

      const config = builder.buildLayoutConfig(metadata);

      expect(config).toHaveProperty(longPaneName);
      expect(config[longPaneName]).toEqual({ height: 200 });
    });

    it('should handle pane name with whitespace', () => {
      const metadata = {
        'Indicator A': { chartPane: 'pane with spaces' },
        'Indicator B': { chartPane: '  leading_trailing  ' },
      };

      const config = builder.buildLayoutConfig(metadata);

      expect(config).toHaveProperty('pane with spaces');
      expect(config).toHaveProperty('  leading_trailing  ');
    });
  });
});
