import { describe, it, expect, beforeEach, vi } from 'vitest';
import { TradingAnalysisRunner } from '../../src/classes/TradingAnalysisRunner.js';
import { CHART_COLORS } from '../../src/config.js';

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
            { time: 1000, value: 100, options: { color: '#2196F3', linewidth: 2 } },
            { time: 2000, value: 101, options: { color: '#2196F3', linewidth: 2 } },
          ],
        },
        'EMA 18': {
          data: [{ time: 1000, value: 99, options: { color: '#F23645', linewidth: 2 } }],
        },
      };

      const metadata = runner.extractIndicatorMetadata(plots);

      expect(metadata).toEqual({
        'EMA 9': {
          color: '#2196F3',
          style: 'line',
          linewidth: 2,
          transp: 0,
          title: 'EMA 9',
          type: 'indicator',
          chartPane: 'main',
        },
        'EMA 18': {
          color: '#F23645',
          style: 'line',
          linewidth: 2,
          transp: 0,
          title: 'EMA 18',
          type: 'indicator',
          chartPane: 'main',
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
      expect(metadata['Custom Indicator'].linewidth).toBe(2);
      expect(metadata['Custom Indicator'].transp).toBe(0);
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
      expect(metadata['Broken Plot'].linewidth).toBe(2);
      expect(metadata['Broken Plot'].transp).toBe(0);
      expect(metadata['Broken Plot'].title).toBe('Broken Plot');
      expect(metadata['Broken Plot'].type).toBe('indicator');
    });

    it('should handle multiple plots with mixed color availability', () => {
      const plots = {
        'Plot With Color': {
          data: [{ time: 1000, value: 50, options: { color: '#4CAF50' } }],
        },
        'Plot Without Color': {
          data: [{ time: 1000, value: 60 }],
        },
      };

      const metadata = runner.extractIndicatorMetadata(plots);

      expect(metadata['Plot With Color'].color).toBe('#4CAF50');
      expect(metadata['Plot With Color'].linewidth).toBe(2);
      expect(metadata['Plot With Color'].transp).toBe(0);
      expect(metadata['Plot Without Color'].color).toBe(CHART_COLORS.DEFAULT_PLOT);
      expect(metadata['Plot Without Color'].linewidth).toBe(2);
      expect(metadata['Plot Without Color'].transp).toBe(0);
    });
  });

  describe('extractPlotLineWidth', () => {
    it('should extract linewidth from first data point with linewidth', () => {
      const plotData = {
        data: [
          { time: 1000, value: 100, options: { linewidth: 3 } },
          { time: 2000, value: 101, options: { linewidth: 2 } },
        ],
      };

      const linewidth = runner.extractPlotLineWidth(plotData);

      expect(linewidth).toBe(3);
    });

    it('should return default linewidth when no data points have linewidth', () => {
      const plotData = {
        data: [
          { time: 1000, value: 100 },
          { time: 2000, value: 101 },
        ],
      };

      const linewidth = runner.extractPlotLineWidth(plotData);

      expect(linewidth).toBe(2);
    });

    it('should skip data points without linewidth options', () => {
      const plotData = {
        data: [
          { time: 1000, value: 100 },
          { time: 2000, value: 101, options: {} },
          { time: 3000, value: 102, options: { linewidth: 5 } },
        ],
      };

      const linewidth = runner.extractPlotLineWidth(plotData);

      expect(linewidth).toBe(5);
    });

    it('should return default when data is not an array', () => {
      const plotData = {
        data: 'invalid',
      };

      const linewidth = runner.extractPlotLineWidth(plotData);

      expect(linewidth).toBe(2);
    });

    it('should return default when plotData is null', () => {
      const linewidth = runner.extractPlotLineWidth(null);

      expect(linewidth).toBe(2);
    });
  });

  describe('extractPlotTransp', () => {
    it('should extract transp from first data point with transp', () => {
      const plotData = {
        data: [
          { time: 1000, value: 100, options: { transp: 50 } },
          { time: 2000, value: 101, options: { transp: 75 } },
        ],
      };

      const transp = runner.extractPlotTransp(plotData);

      expect(transp).toBe(50);
    });

    it('should return 0 when no data points have transp', () => {
      const plotData = {
        data: [
          { time: 1000, value: 100 },
          { time: 2000, value: 101 },
        ],
      };

      const transp = runner.extractPlotTransp(plotData);

      expect(transp).toBe(0);
    });

    it('should handle transp=0 explicitly', () => {
      const plotData = {
        data: [
          { time: 1000, value: 100, options: { transp: 0 } },
        ],
      };

      const transp = runner.extractPlotTransp(plotData);

      expect(transp).toBe(0);
    });

    it('should return default when data is not an array', () => {
      const plotData = {
        data: 'invalid',
      };

      const transp = runner.extractPlotTransp(plotData);

      expect(transp).toBe(0);
    });

    it('should return default when plotData is null', () => {
      const transp = runner.extractPlotTransp(null);

      expect(transp).toBe(0);
    });
  });

  describe('extractPlotColor', () => {
    it('should extract color from first data point with color', () => {
      const plotData = {
        data: [
          { time: 1000, value: 100, options: { color: '#4CAF50' } },
          { time: 2000, value: 101, options: { color: '#2196F3' } },
        ],
      };

      const color = runner.extractPlotColor(plotData);

      expect(color).toBe('#4CAF50');
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
          { time: 3000, value: 102, options: { color: '#9C27B0' } },
        ],
      };

      const color = runner.extractPlotColor(plotData);

      expect(color).toBe('#9C27B0');
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
