import { describe, it, expect, beforeEach, vi } from 'vitest';
import { TradingAnalysisRunner } from '../../src/classes/TradingAnalysisRunner.js';

describe('TradingAnalysisRunner', () => {
  let runner;
  let mockProviderManager;
  let mockPineScriptStrategyRunner;
  let mockCandlestickDataSanitizer;
  let mockConfigurationBuilder;
  let mockJsonFileWriter;
  let mockLogger;

  beforeEach(() => {
    mockProviderManager = {
      fetchMarketData: vi.fn(),
    };
    mockPineScriptStrategyRunner = {
      executeTranspiledStrategy: vi.fn(),
    };
    mockCandlestickDataSanitizer = {
      processCandlestickData: vi.fn(),
    };
    mockConfigurationBuilder = {
      createTradingConfig: vi.fn(),
      generateChartConfig: vi.fn(),
    };
    mockJsonFileWriter = {
      exportChartData: vi.fn(),
      exportConfiguration: vi.fn(),
    };
    mockLogger = {
      log: vi.fn(),
      error: vi.fn(),
      debug: vi.fn(),
    };

    runner = new TradingAnalysisRunner(
      mockProviderManager,
      mockPineScriptStrategyRunner,
      mockCandlestickDataSanitizer,
      mockConfigurationBuilder,
      mockJsonFileWriter,
      mockLogger,
    );
  });

  describe('constructor', () => {
    it('should store all dependencies', () => {
      expect(runner.providerManager).toBe(mockProviderManager);
      expect(runner.pineScriptStrategyRunner).toBe(mockPineScriptStrategyRunner);
      expect(runner.candlestickDataSanitizer).toBe(mockCandlestickDataSanitizer);
      expect(runner.configurationBuilder).toBe(mockConfigurationBuilder);
      expect(runner.jsonFileWriter).toBe(mockJsonFileWriter);
      expect(runner.logger).toBe(mockLogger);
    });
  });
});
