import { describe, it, expect, beforeEach, vi } from 'vitest';
import { MoexProvider } from '../../src/providers/MoexProvider.js';

/* Mock global fetch */
global.fetch = vi.fn();

describe('MoexProvider Pagination', () => {
  let provider;
  let mockStatsCollector;
  let mockLogger;

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
  });

  describe('Single page response (≤500 candles)', () => {
    it('should fetch 100 candles in single request', async () => {
      const mockCandles = Array.from({ length: 100 }, (_, i) => [
        `${100 + i}`,
        `${105 + i}`,
        `${110 + i}`,
        `${90 + i}`,
        `${1000 + i * 10}`,
        `${1000 + i * 10}`,
        `2024-01-${String(i + 1).padStart(2, '0')} 00:00:00`,
        `2024-01-${String(i + 1).padStart(2, '0')} 23:59:59`,
      ]);

      global.fetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        json: async () => ({ candles: { data: mockCandles } }),
      });

      const result = await provider.getMarketData('SBER', '1d', 100);

      expect(global.fetch).toHaveBeenCalledTimes(1);
      expect(result).toHaveLength(100);
      expect(result[0]).toHaveProperty('openTime');
      expect(result[0]).toHaveProperty('open');
      expect(result[0]).toHaveProperty('high');
      expect(result[0]).toHaveProperty('low');
      expect(result[0]).toHaveProperty('close');
      expect(result[0]).toHaveProperty('volume');
    });

    it('should fetch exactly 500 candles in single request', async () => {
      const mockCandles = Array.from({ length: 500 }, (_, i) => [
        '100',
        '105',
        '110',
        '90',
        '1000',
        '1000',
        `2020-01-01 00:00:00`,
        `2020-01-01 23:59:59`,
      ]);

      global.fetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        json: async () => ({ candles: { data: mockCandles } }),
      });

      const result = await provider.getMarketData('SBER', '1d', 500);

      expect(global.fetch).toHaveBeenCalledTimes(1);
      expect(result).toHaveLength(500);
    });
  });

  describe('Multi-page response (>500 candles)', () => {
    it('should fetch 589 candles in 2 requests (500 + 89)', async () => {
      const firstBatch = Array.from({ length: 500 }, (_, i) => [
        `2024-01-01 00:00:00`,
        100 + i,
        110 + i,
        90 + i,
        105 + i,
        1000,
      ]);

      const secondBatch = Array.from({ length: 89 }, (_, i) => [
        `2020-01-01 00:00:00`,
        200 + i,
        210 + i,
        190 + i,
        205 + i,
        2000,
      ]);

      global.fetch = vi
        .fn()
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ candles: { data: firstBatch } }),
        })
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ candles: { data: secondBatch } }),
        });

      const result = await provider.getMarketData('BSPB', 'W', 700);

      expect(global.fetch).toHaveBeenCalledTimes(2);
      expect(result).toHaveLength(589);

      const firstCall = global.fetch.mock.calls[0][0];
      const secondCall = global.fetch.mock.calls[1][0];

      expect(firstCall).not.toContain('start=');
      expect(secondCall).toContain('start=500');
    });

    it('should fetch 2897 candles in 6 requests (500×5 + 397)', async () => {
      const batches = [
        { size: 500, start: 0 },
        { size: 500, start: 500 },
        { size: 500, start: 1000 },
        { size: 500, start: 1500 },
        { size: 500, start: 2000 },
        { size: 397, start: 2500 },
      ];

      global.fetch = vi.fn();

      batches.forEach((batch) => {
        const mockCandles = Array.from({ length: batch.size }, (_, i) => [
          `2020-01-01 00:00:00`,
          100,
          110,
          90,
          105,
          1000,
        ]);

        global.fetch.mockResolvedValueOnce({
          ok: true,
          json: async () => ({ candles: { data: mockCandles } }),
        });
      });

      const result = await provider.getMarketData('BSPB', '1d', 3000);

      expect(global.fetch).toHaveBeenCalledTimes(6);
      expect(result).toHaveLength(2897);

      batches.forEach((batch, index) => {
        const callUrl = global.fetch.mock.calls[index][0];
        if (batch.start === 0) {
          expect(callUrl).not.toContain('start=');
        } else {
          expect(callUrl).toContain(`start=${batch.start}`);
        }
      });
    });

    it('should respect limit parameter during pagination', async () => {
      const firstBatch = Array.from({ length: 500 }, (_, i) => [
        `2024-01-01 00:00:00`,
        100,
        110,
        90,
        105,
        1000,
      ]);

      const secondBatch = Array.from({ length: 200 }, (_, i) => [
        `2023-01-01 00:00:00`,
        200,
        210,
        190,
        205,
        2000,
      ]);

      global.fetch = vi
        .fn()
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ candles: { data: firstBatch } }),
        })
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ candles: { data: secondBatch } }),
        });

      const result = await provider.getMarketData('SBER', '1d', 600);

      expect(global.fetch).toHaveBeenCalledTimes(2);
      expect(result).toHaveLength(600);
    });
  });

  describe('Pagination edge cases', () => {
    it('should stop pagination when empty batch received', async () => {
      const firstBatch = Array.from({ length: 500 }, (_, i) => [
        `2024-01-01 00:00:00`,
        100,
        110,
        90,
        105,
        1000,
      ]);

      global.fetch = vi
        .fn()
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ candles: { data: firstBatch } }),
        })
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ candles: { data: [] } }),
        });

      const result = await provider.getMarketData('SBER', '1d', 1000);

      expect(global.fetch).toHaveBeenCalledTimes(2);
      expect(result).toHaveLength(500);
    });

    it('should stop pagination when batch size < 500', async () => {
      const firstBatch = Array.from({ length: 500 }, (_, i) => [
        `2024-01-01 00:00:00`,
        100,
        110,
        90,
        105,
        1000,
      ]);

      const secondBatch = Array.from({ length: 300 }, (_, i) => [
        `2023-01-01 00:00:00`,
        200,
        210,
        190,
        205,
        2000,
      ]);

      global.fetch = vi
        .fn()
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ candles: { data: firstBatch } }),
        })
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ candles: { data: secondBatch } }),
        });

      const result = await provider.getMarketData('SBER', '1d', 1000);

      expect(global.fetch).toHaveBeenCalledTimes(2);
      expect(result).toHaveLength(800);
    });

    it('should handle API error during pagination', async () => {
      const firstBatch = Array.from({ length: 500 }, (_, i) => [
        `2024-01-01 00:00:00`,
        100,
        110,
        90,
        105,
        1000,
      ]);

      global.fetch = vi
        .fn()
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ candles: { data: firstBatch } }),
        })
        .mockResolvedValueOnce({
          ok: false,
          status: 500,
          statusText: 'Internal Server Error',
        });

      const result = await provider.getMarketData('SBER', '1d', 1000);

      expect(result).toEqual([]);
    });

    it('should stop pagination when limit reached mid-batch', async () => {
      const firstBatch = Array.from({ length: 500 }, (_, i) => [
        `2024-01-01 00:00:00`,
        100,
        110,
        90,
        105,
        1000,
      ]);

      const secondBatch = Array.from({ length: 500 }, (_, i) => [
        `2023-01-01 00:00:00`,
        200,
        210,
        190,
        205,
        2000,
      ]);

      global.fetch = vi
        .fn()
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ candles: { data: firstBatch } }),
        })
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ candles: { data: secondBatch } }),
        });

      const result = await provider.getMarketData('SBER', '1d', 700);

      expect(global.fetch).toHaveBeenCalledTimes(2);
      expect(result).toHaveLength(700);
    });
  });

  describe('Pagination URL construction', () => {
    it('should not add start parameter for first request', async () => {
      const mockCandles = Array.from({ length: 100 }, (_, i) => [
        `2024-01-01 00:00:00`,
        100,
        110,
        90,
        105,
        1000,
      ]);

      global.fetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        json: async () => ({ candles: { data: mockCandles } }),
      });

      await provider.getMarketData('SBER', '1d', 100);

      const callUrl = global.fetch.mock.calls[0][0];
      expect(callUrl).not.toContain('start=');
      expect(callUrl).toContain('iss.reverse=true');
    });

    it('should add correct start parameter for subsequent requests', async () => {
      const batches = [500, 500, 300];

      global.fetch = vi.fn();

      batches.forEach((size) => {
        const mockCandles = Array.from({ length: size }, (_, i) => [
          `2024-01-01 00:00:00`,
          100,
          110,
          90,
          105,
          1000,
        ]);

        global.fetch.mockResolvedValueOnce({
          ok: true,
          json: async () => ({ candles: { data: mockCandles } }),
        });
      });

      await provider.getMarketData('SBER', '1d', 2000);

      expect(global.fetch.mock.calls[0][0]).not.toContain('start=');
      expect(global.fetch.mock.calls[1][0]).toContain('start=500');
      expect(global.fetch.mock.calls[2][0]).toContain('start=1000');
    });
  });

  describe('Pagination with caching', () => {
    it('should cache paginated results', async () => {
      const firstBatch = Array.from({ length: 500 }, (_, i) => [
        `2024-01-01 00:00:00`,
        100,
        110,
        90,
        105,
        1000,
      ]);

      const secondBatch = Array.from({ length: 300 }, (_, i) => [
        `2023-01-01 00:00:00`,
        200,
        210,
        190,
        205,
        2000,
      ]);

      global.fetch = vi
        .fn()
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ candles: { data: firstBatch } }),
        })
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ candles: { data: secondBatch } }),
        });

      const result1 = await provider.getMarketData('SBER', '1d', 800);
      expect(global.fetch).toHaveBeenCalledTimes(2);
      expect(result1).toHaveLength(800);

      const result2 = await provider.getMarketData('SBER', '1d', 800);
      expect(global.fetch).toHaveBeenCalledTimes(2);
      expect(result2).toHaveLength(800);
      expect(result2).toEqual(result1);
    });
  });

  describe('Real-world pagination scenarios', () => {
    it('should handle BSPB weekly data (589 candles)', async () => {
      const firstBatch = Array.from({ length: 500 }, (_, i) => {
        const weekDate = new Date(2024, 0, 1 + i * 7);
        return [
          '100',
          '105',
          '110',
          '90',
          '1000',
          '1000',
          weekDate.toISOString().slice(0, 19).replace('T', ' '),
          new Date(weekDate.getTime() + 6 * 24 * 60 * 60 * 1000).toISOString().slice(0, 19).replace('T', ' '),
        ];
      });

      const secondBatch = Array.from({ length: 89 }, (_, i) => {
        const weekDate = new Date(2016, 0, 1 + i * 7);
        return [
          '50',
          '55',
          '60',
          '40',
          '500',
          '500',
          weekDate.toISOString().slice(0, 19).replace('T', ' '),
          new Date(weekDate.getTime() + 6 * 24 * 60 * 60 * 1000).toISOString().slice(0, 19).replace('T', ' '),
        ];
      });

      global.fetch = vi
        .fn()
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ candles: { data: firstBatch } }),
        })
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ candles: { data: secondBatch } }),
        });

      const result = await provider.getMarketData('BSPB', 'W', 700);

      expect(global.fetch).toHaveBeenCalledTimes(2);
      expect(result).toHaveLength(589);
      expect(result[0].openTime).toBeLessThan(result[588].openTime);
    });

    it('should handle BSPB daily data (2897 candles)', async () => {
      const batches = [
        { size: 500, startDate: '2024-01-01' },
        { size: 500, startDate: '2022-01-01' },
        { size: 500, startDate: '2020-01-01' },
        { size: 500, startDate: '2018-01-01' },
        { size: 500, startDate: '2016-01-01' },
        { size: 397, startDate: '2014-01-01' },
      ];

      const mockCalls = batches.map((batch) => {
        const mockCandles = Array.from({ length: batch.size }, (_, i) => [
          `${batch.startDate} 00:00:00`,
          100,
          110,
          90,
          105,
          1000,
        ]);
        return {
          ok: true,
          json: async () => ({ candles: { data: mockCandles } }),
        };
      });

      global.fetch = vi.fn()
        .mockResolvedValueOnce(mockCalls[0])
        .mockResolvedValueOnce(mockCalls[1])
        .mockResolvedValueOnce(mockCalls[2])
        .mockResolvedValueOnce(mockCalls[3])
        .mockResolvedValueOnce(mockCalls[4])
        .mockResolvedValueOnce(mockCalls[5]);

      const result = await provider.getMarketData('BSPB', '1d', 3000);

      expect(global.fetch).toHaveBeenCalledTimes(6);
      expect(result).toHaveLength(2897);
    });
  });
});
