import { describe, it, expect, beforeEach, vi } from 'vitest';
import { YahooFinanceProvider } from '../../src/providers/YahooFinanceProvider.js';

global.fetch = vi.fn();

describe('YahooFinanceProvider', () => {
  let provider;
  let mockLogger;

  beforeEach(() => {
    mockLogger = {
      log: vi.fn(),
      debug: vi.fn(),
      error: vi.fn(),
    };
    provider = new YahooFinanceProvider(mockLogger);
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

    it('should fetch and return market data', async() => {
      global.fetch.mockResolvedValue({
        ok: true,
        status: 200,
        statusText: 'OK',
        headers: new Map(),
        text: async() => JSON.stringify(mockYahooResponse),
      });

      const data = await provider.getMarketData('AAPL', 'D', 100);

      expect(data).toHaveLength(2);
      expect(data[0].open).toBe(100);
    });

    it('should return empty array on error', async() => {
      global.fetch.mockResolvedValue({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        text: async() => 'Not Found',
      });

      const data = await provider.getMarketData('INVALID', 'D', 100);
      expect(data).toEqual([]);
    });
  });
});
