import { describe, it, expect, beforeEach, vi } from 'vitest';
import { TradingAnalysisRunner } from '../../src/classes/TradingAnalysisRunner.js';
import { CHART_COLORS } from '../../src/constants/chartColors.js';

describe('TradingAnalysisRunner - Metadata Extraction', () => {
  let runner;
  let mockProviderManager;
  let mockPineScriptStrategyRunner;
  let mockCandlestickDataSanitizer;
  let mockConfigurationBuilder;
  let mockJsonFileWriter;
  let mockLogger;

  beforeEach(() => {
    mockProviderManager = { fetchMarketData: vi.fn() };
    mockPineScriptStrategyRunner = {
      runEMAStrategy: vi.fn(),
      getIndicatorMetadata: vi.fn(),
      executeTranspiledStrategy: vi.fn(),
    };
    mockCandlestickDataSanitizer = { processCandlestickData: vi.fn() };
    mockConfigurationBuilder = {
      createTradingConfig: vi.fn(),
      generateChartConfig: vi.fn(),
    };
    mockJsonFileWriter = {
      exportChartData: vi.fn(),
      exportConfiguration: vi.fn(),
    };
    mockLogger = { log: vi.fn(), error: vi.fn(), debug: vi.fn() };

    runner = new TradingAnalysisRunner(
      mockProviderManager,
      mockPineScriptStrategyRunner,
      mockCandlestickDataSanitizer,
      mockConfigurationBuilder,
      mockJsonFileWriter,
      mockLogger,
    );
  });

  describe('extractIndicatorMetadata', () => {
    it('should extract metadata from plots with colors', () => {
      const plots = {
        'EMA 9': {
          data: [
            { time: 1000, value: 100, options: { color: 'blue', linewidth: 2 } },
            { time: 2000, value: 101, options: { color: 'blue', linewidth: 2 } },
          ],
        },
        'EMA 18': {
          data: [{ time: 1000, value: 99, options: { color: 'red', linewidth: 2 } }],
        },
      };

      const metadata = runner.extractIndicatorMetadata(plots);

      expect(metadata).toEqual({
        'EMA 9': {
          color: 'blue',
          style: 'line',
          title: 'EMA 9',
          type: 'indicator',
        },
        'EMA 18': {
          color: 'red',
          style: 'line',
          title: 'EMA 18',
          type: 'indicator',
        },
      });
    });

    it('should use default color when no color in plot data', () => {
      const plots = {
        'Custom Indicator': {
          data: [{ time: 1000, value: 100 }],
        },
      };

      const metadata = runner.extractIndicatorMetadata(plots);

      expect(metadata['Custom Indicator'].color).toBe(CHART_COLORS.DEFAULT_PLOT);
    });

    it('should handle empty plots object', () => {
      const plots = {};

      const metadata = runner.extractIndicatorMetadata(plots);

      expect(metadata).toEqual({});
    });

    it('should handle plots without data array', () => {
      const plots = {
        'Broken Plot': {},
      };

      const metadata = runner.extractIndicatorMetadata(plots);

      expect(metadata['Broken Plot'].color).toBe(CHART_COLORS.DEFAULT_PLOT);
      expect(metadata['Broken Plot'].title).toBe('Broken Plot');
      expect(metadata['Broken Plot'].type).toBe('indicator');
    });

    it('should handle multiple plots with mixed color availability', () => {
      const plots = {
        'Plot With Color': {
          data: [{ time: 1000, value: 50, options: { color: 'green' } }],
        },
        'Plot Without Color': {
          data: [{ time: 1000, value: 60 }],
        },
      };

      const metadata = runner.extractIndicatorMetadata(plots);

      expect(metadata['Plot With Color'].color).toBe('green');
      expect(metadata['Plot Without Color'].color).toBe(CHART_COLORS.DEFAULT_PLOT);
    });
  });

  describe('extractPlotColor', () => {
    it('should extract color from first data point with color', () => {
      const plotData = {
        data: [
          { time: 1000, value: 100, options: { color: 'green' } },
          { time: 2000, value: 101, options: { color: 'blue' } },
        ],
      };

      const color = runner.extractPlotColor(plotData);

      expect(color).toBe('green');
    });

    it('should return default color when no data points have color', () => {
      const plotData = {
        data: [
          { time: 1000, value: 100 },
          { time: 2000, value: 101 },
        ],
      };

      const color = runner.extractPlotColor(plotData);

      expect(color).toBe(CHART_COLORS.DEFAULT_PLOT);
    });

    it('should skip data points without color options', () => {
      const plotData = {
        data: [
          { time: 1000, value: 100 },
          { time: 2000, value: 101, options: {} },
          { time: 3000, value: 102, options: { color: 'purple' } },
        ],
      };

      const color = runner.extractPlotColor(plotData);

      expect(color).toBe('purple');
    });

    it('should return default color when data is not an array', () => {
      const plotData = {
        data: 'invalid',
      };

      const color = runner.extractPlotColor(plotData);

      expect(color).toBe(CHART_COLORS.DEFAULT_PLOT);
    });

    it('should return default color when plotData is null', () => {
      const color = runner.extractPlotColor(null);

      expect(color).toBe(CHART_COLORS.DEFAULT_PLOT);
    });

    it('should return default color when plotData is undefined', () => {
      const color = runner.extractPlotColor(undefined);

      expect(color).toBe(CHART_COLORS.DEFAULT_PLOT);
    });

    it('should return default color when plotData has no data property', () => {
      const plotData = {};

      const color = runner.extractPlotColor(plotData);

      expect(color).toBe(CHART_COLORS.DEFAULT_PLOT);
    });

    it('should return default color when data array is empty', () => {
      const plotData = {
        data: [],
      };

      const color = runner.extractPlotColor(plotData);

      expect(color).toBe(CHART_COLORS.DEFAULT_PLOT);
    });
  });
});
