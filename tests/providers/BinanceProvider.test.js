import { describe, it, expect, vi, beforeEach } from 'vitest';
import { BinanceProvider } from '../../src/providers/BinanceProvider.js';
import { TimeframeParser } from '../../src/utils/timeframeParser.js';

// Mock the PineTS Provider
vi.mock('../../../PineTS/dist/pinets.dev.es.js', () => ({
  Provider: {
    Binance: {
      getMarketData: vi.fn(),
    },
  },
}));

// Mock TimeframeParser
vi.mock('../../src/utils/timeframeParser.js', () => ({
  TimeframeParser: {
    toBinanceTimeframe: vi.fn(),
  },
  SUPPORTED_TIMEFRAMES: {
    BINANCE: ['1m', '3m', '5m', '15m', '30m', '1h', '2h', '4h', '6h', '8h', '12h', '1d', '3d', '1w', '1M'],
  },
}));

describe('BinanceProvider', () => {
  let provider;
  let mockLogger;
  let mockBinanceProvider;
  let mockStatsCollector;

  beforeEach(async() => {
    mockLogger = {
      log: vi.fn(),
      error: vi.fn(),
      debug: vi.fn(),
    };
    mockStatsCollector = {
      recordRequest: vi.fn(),
      recordCacheHit: vi.fn(),
      recordCacheMiss: vi.fn(),
    };

    // Get the mocked Binance provider
    const { Provider } = await import('../../../PineTS/dist/pinets.dev.es.js');
    mockBinanceProvider = Provider.Binance;

    provider = new BinanceProvider(mockLogger, mockStatsCollector);

    // Reset all mocks
    vi.clearAllMocks();
  });

  it('should create BinanceProvider with logger', () => {
    expect(provider.logger).toBe(mockLogger);
    expect(provider.binanceProvider).toBe(mockBinanceProvider);
  });

  it('should convert timeframe and call underlying Binance provider', async() => {
    // Setup mocks
    const mockTimeframe = '1h';
    const convertedTimeframe = '60';
    const mockSymbol = 'BTCUSDT';
    const mockLimit = 100;
    const mockData = [{ open: 100, high: 110, low: 90, close: 105 }];

    TimeframeParser.toBinanceTimeframe.mockReturnValue(convertedTimeframe);
    mockBinanceProvider.getMarketData.mockResolvedValue(mockData);

    // Execute
    const result = await provider.getMarketData(mockSymbol, mockTimeframe, mockLimit);

    // Verify timeframe conversion
    expect(TimeframeParser.toBinanceTimeframe).toHaveBeenCalledWith(mockTimeframe);

    // Verify underlying provider call with converted timeframe
    expect(mockBinanceProvider.getMarketData).toHaveBeenCalledWith(
      mockSymbol,
      convertedTimeframe,
      mockLimit,
      undefined,
      undefined,
    );

    // Verify result
    expect(result).toBe(mockData);
  });

  it('should pass sDate and eDate to underlying provider', async() => {
    const mockSymbol = 'ETHUSDT';
    const mockTimeframe = '15m';
    const convertedTimeframe = '15';
    const mockLimit = 50;
    const mockSDate = '2024-01-01';
    const mockEDate = '2024-01-31';
    const mockData = [];

    TimeframeParser.toBinanceTimeframe.mockReturnValue(convertedTimeframe);
    mockBinanceProvider.getMarketData.mockResolvedValue(mockData);

    await provider.getMarketData(mockSymbol, mockTimeframe, mockLimit, mockSDate, mockEDate);

    expect(mockBinanceProvider.getMarketData).toHaveBeenCalledWith(
      mockSymbol,
      convertedTimeframe,
      mockLimit,
      mockSDate,
      mockEDate,
    );
  });

  it('should handle various timeframe formats', async() => {
    const testCases = [
      { input: '1h', expected: '60' },
      { input: '15m', expected: '15' },
      { input: '5m', expected: '5' },
      { input: 'D', expected: 'D' },
    ];

    for (const testCase of testCases) {
      TimeframeParser.toBinanceTimeframe.mockReturnValue(testCase.expected);
      mockBinanceProvider.getMarketData.mockResolvedValue([]);

      await provider.getMarketData('BTCUSDT', testCase.input, 100);

      expect(TimeframeParser.toBinanceTimeframe).toHaveBeenCalledWith(testCase.input);
      expect(mockBinanceProvider.getMarketData).toHaveBeenCalledWith(
        'BTCUSDT',
        testCase.expected,
        100,
        undefined,
        undefined,
      );
    }
  });

  it('should return empty array for provider errors', async() => {
    const error = new Error('Binance API error');
    TimeframeParser.toBinanceTimeframe.mockReturnValue('60');
    mockBinanceProvider.getMarketData.mockRejectedValue(error);

    const result = await provider.getMarketData('BTCUSDT', '1h', 100);
    expect(result).toEqual([]);
  });
});
