import { describe, it, expect, beforeEach, vi } from 'vitest';
import { MoexProvider } from '../../src/providers/MoexProvider.js';

/* Mock global fetch */
global.fetch = vi.fn();

describe('MoexProvider', () => {
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
    provider = new MoexProvider(mockLogger, mockStatsCollector);
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
    it('should convert supported numeric minute timeframes', () => {
      expect(provider.convertTimeframe(1)).toBe('1');
      expect(provider.convertTimeframe(10)).toBe('10');
      expect(provider.convertTimeframe(60)).toBe('60');
    });

    it('should throw TimeframeError for unsupported numeric timeframes', () => {
      expect(() => provider.convertTimeframe(5)).toThrow("Timeframe '5' not supported");
      expect(() => provider.convertTimeframe(15)).toThrow("Timeframe '15' not supported");
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

  describe('getTimeframeDays()', () => {
    it('should convert string timeframes to correct day fractions', () => {
      // For 1m and 5m: (minutes / 540 * 1.4) + 2 days delay buffer
      expect(provider.getTimeframeDays('1m')).toBeCloseTo((1 / 540 * 1.4) + 2, 5); // 1 min with delay buffer
      expect(provider.getTimeframeDays('15m')).toBeCloseTo(15 / 540 * 1.4, 5); // 15 min in trading days with buffer
      expect(provider.getTimeframeDays('1h')).toBeCloseTo(60 / 540 * 1.4, 5); // 1 hour in trading days with buffer
      expect(provider.getTimeframeDays('4h')).toBeCloseTo(240 / 540 * 1.4, 5); // 4 hours in trading days with buffer
      // Daily and above use calendar days
      expect(provider.getTimeframeDays('1d')).toBe(1440 / 1440); // 1 day = 1 calendar day
    });

    it('should convert letter timeframes to correct day fractions', () => {
      expect(provider.getTimeframeDays('D')).toBe(1440 / 1440); // Daily = 1 calendar day
      expect(provider.getTimeframeDays('W')).toBe(10080 / 1440); // Weekly = 7 calendar days
      expect(provider.getTimeframeDays('M')).toBe(43200 / 1440); // Monthly = 30 calendar days
    });

    it('should convert numeric timeframes to correct day fractions', () => {
      // For 1m: (1 / 540 * 1.4) + 2 days delay buffer
      expect(provider.getTimeframeDays(1)).toBeCloseTo((1 / 540 * 1.4) + 2, 5); // 1 minute with delay buffer
      expect(provider.getTimeframeDays(60)).toBeCloseTo(60 / 540 * 1.4, 5); // 60 minutes = 1 hour
      // Daily timeframes use calendar days
      expect(provider.getTimeframeDays(1440)).toBe(1440 / 1440); // 1440 minutes = 1 calendar day
    });

    it('should handle invalid timeframes with fallback', () => {
      // Invalid timeframes should fallback to daily (1440 minutes = 1 calendar day)
      expect(provider.getTimeframeDays('invalid')).toBe(1440 / 1440); // 1 calendar day
      expect(provider.getTimeframeDays(null)).toBe(1440 / 1440); // 1 calendar day
      expect(provider.getTimeframeDays(undefined)).toBe(1440 / 1440); // 1 calendar day
    });

    it('REGRESSION: should fix the original date range bug', () => {
      // This test prevents the bug where "1h" was treated as 1 day instead of trading hours
      const hourlyDays = provider.getTimeframeDays('1h');
      const expectedHourlyDays = 60 / 540 * 1.4; // ~0.156 days for 1 hour with trading hours + buffer

      expect(hourlyDays).toBeCloseTo(expectedHourlyDays, 3);
      expect(hourlyDays).toBeGreaterThan(0.1); // Should be reasonable fraction of day
      expect(hourlyDays).toBeLessThan(1); // Should be less than full day

      // Verify 15m also uses trading hours calculation
      const fifteenMinDays = provider.getTimeframeDays('15m');
      expect(fifteenMinDays).toBeCloseTo(15 / 540 * 1.4, 3);
      expect(fifteenMinDays).toBeLessThan(hourlyDays); // 15min should be less than 1h
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

    it('should include iss.reverse=true parameter to get newest candles', () => {
      const url = provider.buildUrl('SBER', 'D', 100, null, null);
      expect(url).toContain('iss.reverse=true');
    });

    it('should include reverse parameter for all timeframes when using limit', () => {
      const timeframes = ['1', '10', '60', 'D', 'W', 'M'];

      timeframes.forEach((timeframe) => {
        const url = provider.buildUrl('SBER', timeframe, 100, null, null);
        expect(url).toContain('iss.reverse=true');
      });
    });

    it('should NOT include reverse parameter when custom dates provided', () => {
      const sDate = new Date('2024-01-01').getTime();
      const eDate = new Date('2024-01-31').getTime();
      const url = provider.buildUrl('SBER', 'D', null, sDate, eDate);

      expect(url).not.toContain('iss.reverse=true');
      expect(url).toContain('from=2024-01-01');
      expect(url).toContain('till=2024-01-31');
    });

    it('should include reverse parameter when limit provided without custom dates', () => {
      const url = provider.buildUrl('SBER', '1h', 50, null, null);

      expect(url).toContain('iss.reverse=true');
      expect(url).toContain('from=');
      expect(url).toContain('till=');
    });

    describe('Enhanced date calculation with trading period multipliers', () => {
      it('should apply 1.4x multiplier for daily+ timeframes to account for weekends/holidays', () => {
        const url = provider.buildUrl('SBER', '1d', 100, null, null);
        const urlParams = new URLSearchParams(url.split('?')[1]);
        const fromDate = new Date(urlParams.get('from'));
        const tillDate = new Date(urlParams.get('till'));

        const daysDiff = Math.ceil((tillDate - fromDate) / (24 * 60 * 60 * 1000));

        // For 100 daily candles: 100 * 1 day * 1.4 = 140 days back
        expect(daysDiff).toBeGreaterThanOrEqual(135);
        expect(daysDiff).toBeLessThanOrEqual(145);
      });

      it('should apply 2.4x multiplier for hourly timeframes to account for trading hours', () => {
        const url = provider.buildUrl('SBER', '1h', 200, null, null);
        const urlParams = new URLSearchParams(url.split('?')[1]);
        const fromDate = new Date(urlParams.get('from'));
        const tillDate = new Date(urlParams.get('till'));

        const daysDiff = Math.ceil((tillDate - fromDate) / (24 * 60 * 60 * 1000));

        // For 200 hourly candles:
        // getTimeframeDays: 60/540 * 1.4 = 0.155 days per candle
        // daysBack: Math.ceil(200 * 0.155) = Math.ceil(31.1) = 32
        // With 2.4x multiplier: Math.ceil(32 * 2.4) = Math.ceil(76.8) = 77
        expect(daysDiff).toBeGreaterThanOrEqual(75);
        expect(daysDiff).toBeLessThanOrEqual(82);
      });

      it('should apply 2.2x multiplier for 10m timeframes', () => {
        const url = provider.buildUrl('SBER', '10m', 300, null, null);
        const urlParams = new URLSearchParams(url.split('?')[1]);
        const fromDate = new Date(urlParams.get('from'));
        const tillDate = new Date(urlParams.get('till'));

        const daysDiff = Math.ceil((tillDate - fromDate) / (24 * 60 * 60 * 1000));

        // For 300 10m candles:
        // getTimeframeDays: 10/540 * 1.4 = 0.0259 days per candle
        // daysBack: Math.ceil(300 * 0.0259) = Math.ceil(7.77) = 8
        // With 2.2x multiplier: Math.ceil(8 * 2.2) = Math.ceil(17.6) = 18
        expect(daysDiff).toBeGreaterThanOrEqual(17);
        expect(daysDiff).toBeLessThanOrEqual(22);
      });

      it('should apply 2.0x multiplier for 1m timeframes with delay buffer', () => {
        const url = provider.buildUrl('SBER', '1m', 100, null, null);
        const urlParams = new URLSearchParams(url.split('?')[1]);
        const fromDate = new Date(urlParams.get('from'));
        const tillDate = new Date(urlParams.get('till'));

        const daysDiff = Math.ceil((tillDate - fromDate) / (24 * 60 * 60 * 1000));

        // For 100 1m candles:
        // getTimeframeDays: (1/540 * 1.4) + 2 = 2.0026 days per candle (includes delay buffer)
        // daysBack: Math.ceil(100 * 2.0026) = Math.ceil(200.26) = 201
        // With 2.0x multiplier: Math.ceil(201 * 2.0) = 402
        expect(daysDiff).toBeGreaterThanOrEqual(400);
        expect(daysDiff).toBeLessThanOrEqual(410);
      });

      it('should extend end date to tomorrow for intraday timeframes', () => {
        const now = new Date();
        const tomorrow = new Date(now.getTime() + 24 * 60 * 60 * 1000);

        const url = provider.buildUrl('SBER', '1h', 100, null, null);
        const urlParams = new URLSearchParams(url.split('?')[1]);
        const tillDate = new Date(urlParams.get('till'));

        // Till date should be tomorrow for intraday
        expect(tillDate.getDate()).toBe(tomorrow.getDate());
      });

      it('should use today as end date for daily+ timeframes', () => {
        const now = new Date();

        const url = provider.buildUrl('SBER', '1d', 100, null, null);
        const urlParams = new URLSearchParams(url.split('?')[1]);
        const tillDate = new Date(urlParams.get('till'));

        // Till date should be today for daily+
        expect(tillDate.getDate()).toBe(now.getDate());
      });

      it('should handle weekly timeframes with appropriate multiplier', () => {
        const url = provider.buildUrl('SBER', 'W', 50, null, null);
        const urlParams = new URLSearchParams(url.split('?')[1]);
        const fromDate = new Date(urlParams.get('from'));
        const tillDate = new Date(urlParams.get('till'));

        const daysDiff = Math.ceil((tillDate - fromDate) / (24 * 60 * 60 * 1000));

        // For 50 weekly candles: 50 * 7 days * 1.4 = 490 days back
        expect(daysDiff).toBeGreaterThanOrEqual(480);
        expect(daysDiff).toBeLessThanOrEqual(500);
      });
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

    it('should fetch and return market data', async() => {
      global.fetch.mockResolvedValue({
        ok: true,
        json: async() => mockMoexResponse,
      });

      const data = await provider.getMarketData('SBER', 'D', 100);

      expect(data).toHaveLength(2);
      expect(data[0].open).toBe(100);
      expect(data[1].open).toBe(102);
    });

    it('should return cached data on second call', async() => {
      global.fetch.mockResolvedValue({
        ok: true,
        json: async() => mockMoexResponse,
      });

      await provider.getMarketData('SBER', 'D', 100);
      const data = await provider.getMarketData('SBER', 'D', 100);

      expect(global.fetch).toHaveBeenCalledTimes(1);
      expect(data).toHaveLength(2);
    });

    it('should sort data by time ascending', async() => {
      global.fetch.mockResolvedValue({
        ok: true,
        json: async() => ({
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

    it('should apply limit to data', async() => {
      global.fetch.mockResolvedValue({
        ok: true,
        json: async() => ({
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

    it('should return empty array on API error', async() => {
      global.fetch.mockResolvedValue({
        ok: false,
        status: 404,
        statusText: 'Not Found',
      });

      const data = await provider.getMarketData('INVALID', 'D', 100);

      expect(data).toEqual([]);
    });

    it('should return empty array when no candle data', async() => {
      global.fetch.mockResolvedValue({
        ok: true,
        json: async() => ({ candles: { data: null } }),
      });

      const data = await provider.getMarketData('SBER', 'D', 100);

      expect(data).toEqual([]);
    });

    it('should handle fetch rejection', async() => {
      global.fetch.mockRejectedValue(new Error('Network error'));

      const data = await provider.getMarketData('SBER', 'D', 100);

      expect(data).toEqual([]);
    });

    describe('1d test probe disambiguation', () => {
      it('should throw TimeframeError when empty response and 1d test returns data', async() => {
        /* 15m throws TimeframeError during buildUrl, then 1d test fetch returns data */
        global.fetch.mockResolvedValueOnce({
          ok: true,
          json: async() => mockMoexResponse,
        });

        await expect(provider.getMarketData('CHMF', '15m', 100))
          .rejects
          .toThrow("Timeframe '15m' not supported for symbol 'CHMF' by provider MOEX");

        expect(global.fetch).toHaveBeenCalledTimes(1); // Only 1d probe fetch
      });

      it('should return [] when empty response and 1d test returns empty', async() => {
        /* 15m throws TimeframeError, 1d test fetch returns empty - symbol not found */
        global.fetch.mockResolvedValueOnce({
          ok: true,
          json: async() => ({ candles: { data: [] } }),
        });

        const data = await provider.getMarketData('INVALID_SYMBOL', '15m', 100);

        expect(data).toEqual([]);
        expect(global.fetch).toHaveBeenCalledTimes(1); // Only 1d probe fetch
      });

      it('should return [] when empty response and timeframe is 1d', async() => {
        /* Empty response for 1d - no test needed */
        global.fetch.mockResolvedValue({
          ok: true,
          json: async() => ({ candles: { data: [] } }),
        });

        const data = await provider.getMarketData('INVALID_SYMBOL', '1d', 100);

        expect(data).toEqual([]);
        expect(global.fetch).toHaveBeenCalledTimes(1);
      });

      it('should handle 1d test failure gracefully', async() => {
        /* First call returns empty, 1d test fails with API error */
        global.fetch
          .mockResolvedValueOnce({
            ok: true,
            json: async() => ({ candles: { data: [] } }),
          })
          .mockResolvedValueOnce({
            ok: false,
            status: 500,
          });

        const data = await provider.getMarketData('SBER', '15m', 100);

        expect(data).toEqual([]);
      });
    });
  });
});
