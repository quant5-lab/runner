import { describe, it, expect, beforeEach, vi } from 'vitest';
import { YahooFinanceProvider } from '../../src/providers/YahooFinanceProvider.js';

global.fetch = vi.fn();

describe('YahooFinanceProvider', () => {
  let provider;
  let mockLogger;
  let mockStatsCollector;

  beforeEach(() => {
    mockLogger = {
      log: vi.fn(),
      debug: vi.fn(),
      error: vi.fn(),
    };
    mockStatsCollector = {
      recordRequest: vi.fn(),
      recordCacheHit: vi.fn(),
      recordCacheMiss: vi.fn(),
    };
    provider = new YahooFinanceProvider(mockLogger, mockStatsCollector);
    vi.clearAllMocks();
    provider.cache.clear();
  });

  describe('convertTimeframe()', () => {
    it('should convert numeric timeframes', () => {
      expect(provider.convertTimeframe(1)).toBe('1m');
      expect(provider.convertTimeframe(60)).toBe('1h');
      expect(provider.convertTimeframe('D')).toBe('1d');
      expect(provider.convertTimeframe('W')).toBe('1wk');
    });
  });

  describe('getDateRange()', () => {
    it('should return appropriate ranges for minute timeframes', () => {
      expect(provider.getDateRange(100, '1m')).toBe('1d'); // 1 minute intervals
      expect(provider.getDateRange(100, '5m')).toBe('5d'); // 5 minute intervals (100 > 78 for 1d, 100 < 390 for 5d)
      expect(provider.getDateRange(100, '15m')).toBe('5d'); // 15 minute intervals (100 > 26 for 1d, 100 < 130 for 5d)
      expect(provider.getDateRange(100, '30m')).toBe('10d'); // 30 minute intervals (100 > 65 for 5d, 100 < 130 for 10d)
    });

    it('should return appropriate ranges for hour timeframes', () => {
      expect(provider.getDateRange(100, '1h')).toBe('1mo'); // 1 hour intervals - need 1 month for ~130 candles
      expect(provider.getDateRange(100, '4h')).toBe('3mo'); // 4 hour intervals
    });

    it('should return appropriate ranges for day/week/month timeframes', () => {
      expect(provider.getDateRange(100, '1d')).toBe('6mo'); // Daily intervals (100 > 90 for 3mo, 100 < 180 for 6mo)
      expect(provider.getDateRange(100, 'D')).toBe('6mo'); // Daily intervals (letter format)
      expect(provider.getDateRange(100, 'W')).toBe('2y'); // Weekly intervals (100 > 52 for 1y, 100 < 104 for 2y)
      expect(provider.getDateRange(100, 'M')).toBe('10y'); // Monthly intervals (100 > 60 for 5y, so returns default 10y)
    });

    it('should handle numeric timeframe inputs', () => {
      expect(provider.getDateRange(100, 1)).toBe('1d'); // 1 minute
      expect(provider.getDateRange(100, 15)).toBe('5d'); // 15 minutes (same as string '15m')
      expect(provider.getDateRange(100, 60)).toBe('1mo'); // 60 minutes = 1 hour
      expect(provider.getDateRange(100, 240)).toBe('3mo'); // 240 minutes = 4 hours
      expect(provider.getDateRange(100, 1440)).toBe('6mo'); // 1440 minutes = 1 day (same as string '1d')
    });

    it('should handle invalid timeframes with fallback', () => {
      // Invalid timeframes should fallback to daily (1440 minutes) → '6mo' range for 100 candles
      expect(provider.getDateRange(100, 'invalid')).toBe('6mo'); // Fallback to daily → 6mo
      expect(provider.getDateRange(100, null)).toBe('6mo'); // Fallback to daily → 6mo
      expect(provider.getDateRange(100, undefined)).toBe('6mo'); // Fallback to daily → 6mo
    });

    it('REGRESSION: should fix the original date range selection bug', () => {
      // This test prevents the bug where "1h" was not found in mapping and defaulted to '6mo'
      expect(provider.getDateRange(100, '1h')).toBe('1mo'); // Should be 1mo for hourly to get ~130 candles
      expect(provider.getDateRange(100, '15m')).toBe('5d'); // Should be 5d for 15min based on dynamic logic

      // Verify these are NOT the old insufficient values
      expect(provider.getDateRange(100, '1h')).not.toBe('5d'); // 5d only gives ~33 candles
      expect(provider.getDateRange(100, '15m')).not.toBe('1d'); // 1d insufficient for 100 candles
    });

    it('should use TimeframeParser logic for all timeframe formats', () => {
      // Test that TimeframeParser integration works for string, numeric, and letter formats
      const stringFormats = ['1m', '5m', '15m', '30m', '1h', '4h', '1d'];
      const numericFormats = [1, 5, 15, 30, 60, 240, 1440];
      const letterFormats = ['D', 'W', 'M'];

      // All should return valid range strings
      [...stringFormats, ...numericFormats, ...letterFormats].forEach((tf) => {
        const range = provider.getDateRange(100, tf);
        expect(range).toBeTruthy();
        expect(typeof range).toBe('string');
        expect(['1d', '5d', '10d', '1mo', '3mo', '6mo', '1y', '2y', '5y', '10y']).toContain(range);
      });
    });
  });

  describe('getMarketData()', () => {
    const mockYahooResponse = {
      chart: {
        result: [
          {
            timestamp: [1609459200, 1609545600],
            indicators: {
              quote: [
                {
                  open: [100, 102],
                  high: [105, 108],
                  low: [95, 100],
                  close: [102, 107],
                  volume: [1000, 1200],
                },
              ],
            },
          },
        ],
      },
    };

    it('should fetch and return market data', async () => {
      global.fetch.mockResolvedValue({
        ok: true,
        status: 200,
        statusText: 'OK',
        headers: new Map(),
        text: async () => JSON.stringify(mockYahooResponse),
      });

      const data = await provider.getMarketData('AAPL', 'D', 100);

      expect(data).toHaveLength(2);
      expect(data[0].open).toBe(100);
    });

    it('should return empty array on error', async () => {
      global.fetch.mockResolvedValue({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        text: async () => 'Not Found',
      });

      const data = await provider.getMarketData('INVALID', 'D', 100);
      expect(data).toEqual([]);
    });
  });
});
