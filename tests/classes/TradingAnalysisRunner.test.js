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
      createPineTSAdapter: vi.fn(),
      runEMAStrategy: vi.fn(),
      getIndicatorMetadata: vi.fn(),
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

  describe('runTradingAnalysis()', () => {
    const mockMarketData = [
      { openTime: 1000000, open: 100, high: 105, low: 95, close: 102 },
      { openTime: 2000000, open: 102, high: 108, low: 100, close: 107 },
    ];

    const mockProcessedData = [
      { time: 1, open: 100, high: 105, low: 95, close: 102 },
      { time: 2, open: 102, high: 108, low: 100, close: 107 },
    ];

    const mockPlots = {
      EMA9: [100, 101],
      EMA18: [99, 100],
      BullSignal: [1, 1],
    };

    const mockTradingConfig = {
      symbol: 'BTCUSDT',
      timeframe: 'D',
      bars: 100,
    };

    const mockChartConfig = {
      ui: { title: 'Test' },
    };

    const mockIndicatorMetadata = {
      EMA9: { title: 'EMA 9' },
    };

    beforeEach(() => {
      mockProviderManager.fetchMarketData.mockResolvedValue({
        provider: 'BINANCE',
        data: mockMarketData,
        instance: {},
      });
      mockPineScriptStrategyRunner.createPineTSAdapter.mockResolvedValue({});
      mockPineScriptStrategyRunner.runEMAStrategy.mockResolvedValue({
        result: {},
        plots: mockPlots,
      });
      mockPineScriptStrategyRunner.getIndicatorMetadata.mockReturnValue(mockIndicatorMetadata);
      mockCandlestickDataSanitizer.processCandlestickData.mockReturnValue(mockProcessedData);
      mockConfigurationBuilder.createTradingConfig.mockReturnValue(mockTradingConfig);
      mockConfigurationBuilder.generateChartConfig.mockReturnValue(mockChartConfig);
    });

    it('should execute full trading analysis workflow', async() => {
      await runner.run('BTCUSDT', 'D', 100);

      expect(mockLogger.log).toHaveBeenCalled();
      expect(mockProviderManager.fetchMarketData).toHaveBeenCalledWith('BTCUSDT', 'D', 100);
      expect(mockPineScriptStrategyRunner.createPineTSAdapter).toHaveBeenCalled();
      expect(mockPineScriptStrategyRunner.runEMAStrategy).toHaveBeenCalled();
      expect(mockCandlestickDataSanitizer.processCandlestickData).toHaveBeenCalled();
      expect(mockJsonFileWriter.exportChartData).toHaveBeenCalled();
      expect(mockJsonFileWriter.exportConfiguration).toHaveBeenCalled();
    });

    it('should log configuration at start', async() => {
      await runner.run('BTCUSDT', 'D', 100);

      expect(mockLogger.log).toHaveBeenCalledWith(
        'Configuration:\tSymbol=BTCUSDT, Timeframe=D, Bars=100',
      );
    });

    it('should create trading config with correct parameters', async() => {
      await runner.run('AAPL', 'W', 200);

      expect(mockConfigurationBuilder.createTradingConfig).toHaveBeenCalledWith('AAPL', 'W', 200, 'Multi-Provider Strategy');
    });

    it('should fetch market data from provider manager', async() => {
      await runner.run('BTCUSDT', 'D', 100);

      expect(mockProviderManager.fetchMarketData).toHaveBeenCalledWith('BTCUSDT', 'D', 100);
    });

    it('should log provider used', async() => {
      await runner.run('BTCUSDT', 'D', 100);

      expect(mockLogger.log).toHaveBeenCalledWith(expect.stringContaining('Data source:\tBINANCE'));
    });

    it('should create PineTS adapter with market data', async() => {
      await runner.run('BTCUSDT', 'D', 100);

      expect(mockPineScriptStrategyRunner.createPineTSAdapter).toHaveBeenCalledWith(
        'BINANCE',
        mockMarketData,
        {},
        'BTCUSDT',
        'D',
        100,
      );
    });

    it('should run EMA strategy', async() => {
      const mockPineTS = { ready: vi.fn() };
      mockPineScriptStrategyRunner.createPineTSAdapter.mockResolvedValue(mockPineTS);

      await runner.run('BTCUSDT', 'D', 100);

      expect(mockPineScriptStrategyRunner.runEMAStrategy).toHaveBeenCalledWith(mockPineTS);
    });

    it('should process candlestick data', async() => {
      await runner.run('BTCUSDT', 'D', 100);

      expect(mockCandlestickDataSanitizer.processCandlestickData).toHaveBeenCalledWith(
        mockMarketData,
      );
    });

    it('should export chart data with processed candles and plots', async() => {
      await runner.run('BTCUSDT', 'D', 100);

      expect(mockJsonFileWriter.exportChartData).toHaveBeenCalledWith(mockProcessedData, mockPlots);
    });

    it('should generate and export chart configuration', async() => {
      await runner.run('BTCUSDT', 'D', 100);

      expect(mockConfigurationBuilder.generateChartConfig).toHaveBeenCalledWith(
        mockTradingConfig,
        mockIndicatorMetadata,
      );
      expect(mockJsonFileWriter.exportConfiguration).toHaveBeenCalledWith(mockChartConfig);
    });

    it('should log success message with candle count', async() => {
      await runner.run('BTCUSDT', 'D', 100);

      expect(mockLogger.log).toHaveBeenCalledWith(
        expect.stringContaining('Processing:\t2 candles'),
      );
    });

    it('should throw error when no market data available', async() => {
      mockProviderManager.fetchMarketData.mockResolvedValue({
        provider: 'BINANCE',
        data: [],
        instance: {},
      });

      await expect(runner.run('BTCUSDT', 'D', 100)).rejects.toThrow(
        'No valid market data available for BTCUSDT',
      );
    });

    it('should throw error when market data is null', async() => {
      mockProviderManager.fetchMarketData.mockResolvedValue({
        provider: 'BINANCE',
        data: null,
        instance: {},
      });

      await expect(runner.run('BTCUSDT', 'D', 100)).rejects.toThrow(
        'No valid market data available',
      );
    });
  });

  describe('processIndicatorPlots()', () => {
    const mockData = [
      { openTime: 1000000, open: 100 },
      { openTime: 2000000, open: 102 },
    ];

    it('should process EMA9 plot data', () => {
      const result = { ema9: [100, 101] };
      const processed = runner.processIndicatorPlots(result, mockData);

      expect(processed.EMA9).toBeDefined();
      expect(processed.EMA9.data).toHaveLength(2);
      expect(processed.EMA9.data[0].value).toBe(100);
      expect(processed.EMA9.data[1].value).toBe(101);
    });

    it('should process EMA18 plot data', () => {
      const result = { ema18: [99, 100] };
      const processed = runner.processIndicatorPlots(result, mockData);

      expect(processed.EMA18).toBeDefined();
      expect(processed.EMA18.data).toHaveLength(2);
    });

    it('should process bullSignal as single value', () => {
      const result = { bullSignal: true };
      const processed = runner.processIndicatorPlots(result, mockData);

      expect(processed.BullSignal).toBeDefined();
      expect(processed.BullSignal.data).toHaveLength(2);
      expect(processed.BullSignal.data[1].value).toBe(1);
    });

    it('should process bullSignal as array', () => {
      const result = { bullSignal: [true, false] };
      const processed = runner.processIndicatorPlots(result, mockData);

      expect(processed.BullSignal.data[0].value).toBe(1);
      expect(processed.BullSignal.data[1].value).toBe(0);
    });

    it('should use timestamp from market data', () => {
      const result = { ema9: [100] };
      const processed = runner.processIndicatorPlots(result, mockData);

      expect(processed.EMA9.data[0].time).toBe(1000);
    });

    it('should handle missing indicator data gracefully', () => {
      const result = {};
      const processed = runner.processIndicatorPlots(result, mockData);

      expect(Object.keys(processed)).toHaveLength(0);
    });
  });
});
