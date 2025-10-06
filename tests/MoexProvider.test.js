import { describe, it, expect, beforeEach, vi } from 'vitest';
import { MoexProvider } from '../providers/MoexProvider.js';

/* Mock global fetch */
global.fetch = vi.fn();

describe('MoexProvider', () => {
  let provider;

  beforeEach(() => {
    provider = new MoexProvider();
    vi.clearAllMocks();
    provider.cache.clear();
  });

  describe('constructor', () => {
    it('should initialize with base URL', () => {
      expect(provider.baseUrl).toBe('https://iss.moex.com/iss');
    });

    it('should initialize empty cache', () => {
      expect(provider.cache.size).toBe(0);
    });

    it('should set cache duration to 5 minutes', () => {
      expect(provider.cacheDuration).toBe(5 * 60 * 1000);
    });
  });

  describe('convertTimeframe()', () => {
    it('should convert numeric minute timeframes', () => {
      expect(provider.convertTimeframe(1)).toBe('1');
      expect(provider.convertTimeframe(5)).toBe('5');
      expect(provider.convertTimeframe(15)).toBe('15');
      expect(provider.convertTimeframe(60)).toBe('60');
    });

    it('should convert letter timeframes', () => {
      expect(provider.convertTimeframe('D')).toBe('24');
      expect(provider.convertTimeframe('W')).toBe('7');
      expect(provider.convertTimeframe('M')).toBe('31');
    });

    it('should return default for unknown timeframe', () => {
      expect(provider.convertTimeframe('X')).toBe('24');
    });
  });

  describe('convertMoexCandle()', () => {
    it('should convert MOEX candle to standard format', () => {
      const moexCandle = [
        '100',
        '102',
        '105',
        '95',
        '50000',
        '1000',
        '2024-01-01 09:00:00',
        '2024-01-01 10:00:00',
      ];

      const converted = provider.convertMoexCandle(moexCandle);

      expect(converted.open).toBe(100);
      expect(converted.close).toBe(102);
      expect(converted.high).toBe(105);
      expect(converted.low).toBe(95);
      expect(converted.volume).toBe(1000);
      expect(typeof converted.openTime).toBe('number');
      expect(typeof converted.closeTime).toBe('number');
    });

    it('should parse string values to floats', () => {
      const moexCandle = [
        '100.5',
        '102.3',
        '105.7',
        '95.2',
        '50000',
        '1000',
        '2024-01-01',
        '2024-01-01',
      ];

      const converted = provider.convertMoexCandle(moexCandle);

      expect(converted.open).toBe(100.5);
      expect(converted.close).toBe(102.3);
    });
  });

  describe('formatDate()', () => {
    it('should format timestamp to YYYY-MM-DD', () => {
      const timestamp = new Date('2024-01-15T10:30:00Z').getTime();
      const formatted = provider.formatDate(timestamp);
      expect(formatted).toBe('2024-01-15');
    });

    it('should return empty string for null timestamp', () => {
      expect(provider.formatDate(null)).toBe('');
    });

    it('should return empty string for undefined timestamp', () => {
      expect(provider.formatDate(undefined)).toBe('');
    });
  });

  describe('getCacheKey()', () => {
    it('should generate cache key from parameters', () => {
      const key = provider.getCacheKey('SBER', 'D', 100, '2024-01-01', '2024-01-31');
      expect(key).toBe('SBER_D_100_2024-01-01_2024-01-31');
    });
  });

  describe('cache operations', () => {
    it('should set and get from cache', () => {
      const data = [{ openTime: 1000 }];
      provider.setCache('test_key', data);

      const cached = provider.getFromCache('test_key');
      expect(cached).toEqual(data);
    });

    it('should return null for non-existent cache key', () => {
      expect(provider.getFromCache('nonexistent')).toBeNull();
    });

    it('should expire cache after duration', () => {
      const data = [{ openTime: 1000 }];
      provider.setCache('test_key', data);

      /* Manipulate timestamp to simulate expiry */
      const cached = provider.cache.get('test_key');
      cached.timestamp = Date.now() - provider.cacheDuration - 1000;

      expect(provider.getFromCache('test_key')).toBeNull();
    });
  });

  describe('buildUrl()', () => {
    it('should build URL with interval parameter', () => {
      const url = provider.buildUrl('SBER', 'D', null, null, null);
      expect(url).toContain('interval=24');
    });

    it('should include ticker in URL path', () => {
      const url = provider.buildUrl('GAZP', 'D', null, null, null);
      expect(url).toContain('/securities/GAZP/candles.json');
    });

    it('should add from and till dates when provided', () => {
      const sDate = new Date('2024-01-01').getTime();
      const eDate = new Date('2024-01-31').getTime();
      const url = provider.buildUrl('SBER', 'D', null, sDate, eDate);

      expect(url).toContain('from=2024-01-01');
      expect(url).toContain('till=2024-01-31');
    });

    it('should calculate date range from limit when dates not provided', () => {
      const url = provider.buildUrl('SBER', 'D', 100, null, null);
      expect(url).toContain('from=');
      expect(url).toContain('till=');
    });
  });

  describe('getMarketData()', () => {
    const mockMoexResponse = {
      candles: {
        data: [
          ['100', '102', '105', '95', '50000', '1000', '2024-01-01', '2024-01-01'],
          ['102', '107', '108', '100', '60000', '1200', '2024-01-02', '2024-01-02'],
        ],
      },
    };

    it('should fetch and return market data', async () => {
      global.fetch.mockResolvedValue({
        ok: true,
        json: async () => mockMoexResponse,
      });

      const data = await provider.getMarketData('SBER', 'D', 100);

      expect(data).toHaveLength(2);
      expect(data[0].open).toBe(100);
      expect(data[1].open).toBe(102);
    });

    it('should return cached data on second call', async () => {
      global.fetch.mockResolvedValue({
        ok: true,
        json: async () => mockMoexResponse,
      });

      await provider.getMarketData('SBER', 'D', 100);
      const data = await provider.getMarketData('SBER', 'D', 100);

      expect(global.fetch).toHaveBeenCalledTimes(1);
      expect(data).toHaveLength(2);
    });

    it('should sort data by time ascending', async () => {
      global.fetch.mockResolvedValue({
        ok: true,
        json: async () => ({
          candles: {
            data: [
              ['102', '107', '108', '100', '60000', '1200', '2024-01-02', '2024-01-02'],
              ['100', '102', '105', '95', '50000', '1000', '2024-01-01', '2024-01-01'],
            ],
          },
        }),
      });

      const data = await provider.getMarketData('SBER', 'D', 100);

      expect(data[0].open).toBe(100);
      expect(data[1].open).toBe(102);
    });

    it('should apply limit to data', async () => {
      global.fetch.mockResolvedValue({
        ok: true,
        json: async () => ({
          candles: {
            data: Array(10)
              .fill(null)
              .map((_, i) => [
                '100',
                '102',
                '105',
                '95',
                '50000',
                '1000',
                '2024-01-01',
                '2024-01-01',
              ]),
          },
        }),
      });

      const data = await provider.getMarketData('SBER', 'D', 5);

      expect(data).toHaveLength(5);
    });

    it('should return empty array on API error', async () => {
      global.fetch.mockResolvedValue({
        ok: false,
        status: 404,
        statusText: 'Not Found',
      });

      const data = await provider.getMarketData('INVALID', 'D', 100);

      expect(data).toEqual([]);
    });

    it('should return empty array when no candle data', async () => {
      global.fetch.mockResolvedValue({
        ok: true,
        json: async () => ({ candles: { data: null } }),
      });

      const data = await provider.getMarketData('SBER', 'D', 100);

      expect(data).toEqual([]);
    });

    it('should handle fetch rejection', async () => {
      global.fetch.mockRejectedValue(new Error('Network error'));

      const data = await provider.getMarketData('SBER', 'D', 100);

      expect(data).toEqual([]);
    });
  });
});
