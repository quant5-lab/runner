import { describe, it, expect, beforeEach, vi } from 'vitest';
import { MoexProvider } from '../../src/providers/MoexProvider.js';

/* TEST 1: Mock provider - verify request parameters don't overlap */
describe('MoexProvider Pagination Overlap Prevention - Mock Provider', () => {
  let provider;
  let mockStatsCollector;
  let mockLogger;
  let capturedRequests;

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
    capturedRequests = [];
    vi.clearAllMocks();
  });

  /* Extract start parameter from URL */
  const extractStartParam = (url) => {
    const match = url.match(/start=(\d+)/);
    return match ? parseInt(match[1], 10) : 0;
  };

  /* Mock fetch that captures all request parameters */
  const setupMockFetch = (batches) => {
    global.fetch = vi.fn((url) => {
      const start = extractStartParam(url);
      capturedRequests.push({ url, start });

      const batchIndex = start / 500;
      const batch = batches[batchIndex];

      if (!batch) {
        return Promise.resolve({
          ok: true,
          json: async () => ({ candles: { data: [] } }),
        });
      }

      const mockCandles = Array.from({ length: batch }, (_, i) => [
        '100',
        '105',
        '110',
        '90',
        '1000',
        '1000',
        `2024-01-${String(start + i + 1).padStart(2, '0')} 00:00:00`,
        `2024-01-${String(start + i + 1).padStart(2, '0')} 23:59:59`,
      ]);

      return Promise.resolve({
        ok: true,
        json: async () => ({ candles: { data: mockCandles } }),
      });
    });
  };

  /* Verify no overlapping start parameters */
  const assertNoOverlap = () => {
    for (let i = 1; i < capturedRequests.length; i++) {
      const prevStart = capturedRequests[i - 1].start;
      const currStart = capturedRequests[i].start;

      /* Current start must be exactly prevStart + 500 */
      expect(currStart).toBe(prevStart + 500);

      /* No gap or overlap */
      expect(currStart - prevStart).toBe(500);
    }
  };

  describe('Sequential pagination - 50 test cases', () => {
    it('Case 1: 2 pages (500 + 500)', async () => {
      setupMockFetch([500, 500]);
      await provider.getMarketData('TEST', '1d', 1000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(2);
      expect(capturedRequests[0].start).toBe(0);
      expect(capturedRequests[1].start).toBe(500);
    });

    it('Case 2: 3 pages (500 + 500 + 500)', async () => {
      setupMockFetch([500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 1500);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(3);
    });

    it('Case 3: 5 pages (500×5)', async () => {
      setupMockFetch([500, 500, 500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 2500);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(5);
    });

    it('Case 4: 10 pages (500×10)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500, 500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 5000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(10);
    });

    it('Case 5: 2 pages with partial last (500 + 200)', async () => {
      setupMockFetch([500, 200]);
      await provider.getMarketData('TEST', '1d', 1000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(2);
    });

    it('Case 6: 3 pages with partial last (500 + 500 + 89)', async () => {
      setupMockFetch([500, 500, 89]);
      await provider.getMarketData('TEST', '1d', 2000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(3);
    });

    it('Case 7: 6 pages with partial last (500×5 + 397)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 397]);
      await provider.getMarketData('TEST', '1d', 3000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(6);
    });

    it('Case 8: 2 pages with limit mid-batch (500 + partial)', async () => {
      setupMockFetch([500, 500]);
      await provider.getMarketData('TEST', '1d', 700);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(2);
    });

    it('Case 9: 4 pages (500×4)', async () => {
      setupMockFetch([500, 500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 2000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(4);
    });

    it('Case 10: 7 pages (500×7)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 3500);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(7);
    });

    it('Case 11: 2 pages with very small last (500 + 1)', async () => {
      setupMockFetch([500, 1]);
      await provider.getMarketData('TEST', '1d', 1000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(2);
    });

    it('Case 12: 2 pages with exact boundary (500 + 499)', async () => {
      setupMockFetch([500, 499]);
      await provider.getMarketData('TEST', '1d', 1000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(2);
    });

    it('Case 13: 8 pages (500×8)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 4000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(8);
    });

    it('Case 14: 3 pages with limit at boundary (500 + 500 + partial)', async () => {
      setupMockFetch([500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 1000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(2);
    });

    it('Case 15: 9 pages (500×9)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 4500);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(9);
    });

    it('Case 16: 2 pages with limit=501', async () => {
      setupMockFetch([500, 500]);
      await provider.getMarketData('TEST', '1d', 501);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(2);
    });

    it('Case 17: 2 pages with limit=999', async () => {
      setupMockFetch([500, 500]);
      await provider.getMarketData('TEST', '1d', 999);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(2);
    });

    it('Case 18: 3 pages with limit=1001', async () => {
      setupMockFetch([500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 1001);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(3);
    });

    it('Case 19: 11 pages (500×11)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 5500);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(11);
    });

    it('Case 20: 4 pages with partial last (500×3 + 250)', async () => {
      setupMockFetch([500, 500, 500, 250]);
      await provider.getMarketData('TEST', '1d', 2000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(4);
    });

    it('Case 21: 12 pages (500×12)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 6000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(12);
    });

    it('Case 22: 5 pages with limit mid-batch (500×4 + partial)', async () => {
      setupMockFetch([500, 500, 500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 2200);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(5);
    });

    it('Case 23: 6 pages (500×6)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 3000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(6);
    });

    it('Case 24: 2 pages with limit=750', async () => {
      setupMockFetch([500, 500]);
      await provider.getMarketData('TEST', '1d', 750);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(2);
    });

    it('Case 25: 3 pages with limit=1250', async () => {
      setupMockFetch([500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 1250);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(3);
    });

    it('Case 26: 13 pages (500×13)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 6500);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(13);
    });

    it('Case 27: 7 pages with partial last (500×6 + 100)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 100]);
      await provider.getMarketData('TEST', '1d', 4000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(7);
    });

    it('Case 28: 14 pages (500×14)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 7000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(14);
    });

    it('Case 29: 15 pages (500×15)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 7500);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(15);
    });

    it('Case 30: 5 pages with limit=2001', async () => {
      setupMockFetch([500, 500, 500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 2001);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(5);
    });

    it('Case 31: 16 pages (500×16)', async () => {
      setupMockFetch([
        500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500,
      ]);
      await provider.getMarketData('TEST', '1d', 8000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(16);
    });

    it('Case 32: 8 pages with partial last (500×7 + 450)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500, 450]);
      await provider.getMarketData('TEST', '1d', 4500);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(8);
    });

    it('Case 33: 17 pages (500×17)', async () => {
      setupMockFetch([
        500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500,
      ]);
      await provider.getMarketData('TEST', '1d', 8500);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(17);
    });

    it('Case 34: 18 pages (500×18)', async () => {
      setupMockFetch([
        500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500,
      ]);
      await provider.getMarketData('TEST', '1d', 9000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(18);
    });

    it('Case 35: 19 pages (500×19)', async () => {
      setupMockFetch([
        500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500,
        500,
      ]);
      await provider.getMarketData('TEST', '1d', 9500);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(19);
    });

    it('Case 36: 20 pages (500×20)', async () => {
      setupMockFetch([
        500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500,
        500, 500,
      ]);
      await provider.getMarketData('TEST', '1d', 10000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(20);
    });

    it('Case 37: 4 pages with empty 5th triggers stop', async () => {
      setupMockFetch([500, 500, 500, 500, 0]);
      await provider.getMarketData('TEST', '1d', 3000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(5);
    });

    it('Case 38: 9 pages with partial last (500×8 + 321)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500, 500, 321]);
      await provider.getMarketData('TEST', '1d', 5000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(9);
    });

    it('Case 39: 10 pages with partial last (500×9 + 150)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500, 500, 500, 150]);
      await provider.getMarketData('TEST', '1d', 6000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(10);
    });

    it('Case 40: 6 pages with limit=2750', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 2750);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(6);
    });

    it('Case 41: 11 pages with partial last (500×10 + 275)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 275]);
      await provider.getMarketData('TEST', '1d', 6000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(11);
    });

    it('Case 42: 7 pages with limit=3333', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 3333);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(7);
    });

    it('Case 43: 12 pages with partial last (500×11 + 88)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 88]);
      await provider.getMarketData('TEST', '1d', 7000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(12);
    });

    it('Case 44: 8 pages with limit=3777', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 3777);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(8);
    });

    it('Case 45: 13 pages with partial last (500×12 + 444)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 444]);
      await provider.getMarketData('TEST', '1d', 7000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(13);
    });

    it('Case 46: 9 pages with limit=4321', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 4321);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(9);
    });

    it('Case 47: 14 pages with partial last (500×13 + 199)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 199]);
      await provider.getMarketData('TEST', '1d', 8000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(14);
    });

    it('Case 48: 10 pages with limit=4888', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500, 500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 4888);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(10);
    });

    it('Case 49: 15 pages with partial last (500×14 + 333)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 333]);
      await provider.getMarketData('TEST', '1d', 8000);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(15);
    });

    it('Case 50: 12 pages with limit=5555 (fetches until 6000)', async () => {
      setupMockFetch([500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500]);
      await provider.getMarketData('TEST', '1d', 5555);
      assertNoOverlap();
      expect(capturedRequests).toHaveLength(12);
    });
  });

  describe('Overlap edge cases - deduplication', () => {
    /* Helper to create overlapping pages */
    const setupOverlapFetch = (pageConfigs) => {
      global.fetch = vi.fn((url) => {
        const start = extractStartParam(url);
        capturedRequests.push({ url, start });

        const batchIndex = start / 500;
        const config = pageConfigs[batchIndex];

        if (!config) {
          return Promise.resolve({
            ok: true,
            json: async () => ({ candles: { data: [] } }),
          });
        }

        const mockCandles = Array.from({ length: config.size }, (_, i) => {
          const candleIndex = config.startIndex + i;
          const dayStr = String((candleIndex % 28) + 1).padStart(2, '0');
          return [
            '100',
            '105',
            '110',
            '90',
            '1000',
            '1000',
            `2024-01-${dayStr} 00:00:00`,
            `2024-01-${dayStr} 23:59:59`,
          ];
        });

        return Promise.resolve({
          ok: true,
          json: async () => ({ candles: { data: mockCandles } }),
        });
      });
    };

    /* Verify timeline consistency */
    const assertTimelineConsistency = (result) => {
      for (let i = 1; i < result.length; i++) {
        expect(result[i].openTime).toBeGreaterThan(result[i - 1].openTime);
      }
    };

    it('Overlap Case 1: Last 50 candles of page 1 duplicated in page 2', async () => {
      setupOverlapFetch([
        { size: 500, startIndex: 0 },
        { size: 500, startIndex: 450 },
      ]);
      const result = await provider.getMarketData('TEST', '1d', 1000);
      assertTimelineConsistency(result);
      expect(result).toHaveLength(28); /* 28 unique days (1-28) after deduplication */
    });

    it('Overlap Case 2: Last 1 candle of page 1 duplicated in page 2', async () => {
      setupOverlapFetch([
        { size: 500, startIndex: 0 },
        { size: 500, startIndex: 499 },
      ]);
      const result = await provider.getMarketData('TEST', '1d', 1000);
      assertTimelineConsistency(result);
      expect(result).toHaveLength(28); /* 28 unique days after deduplication */
    });

    it('Overlap Case 3: Last 200 candles of page 1 duplicated in page 2', async () => {
      setupOverlapFetch([
        { size: 500, startIndex: 0 },
        { size: 500, startIndex: 300 },
      ]);
      const result = await provider.getMarketData('TEST', '1d', 1000);
      assertTimelineConsistency(result);
      expect(result).toHaveLength(28); /* 28 unique days after deduplication */
    });

    it('Overlap Case 4: Page 2 completely contained in page 1', async () => {
      setupOverlapFetch([
        { size: 500, startIndex: 0 },
        { size: 300, startIndex: 100 },
      ]);
      const result = await provider.getMarketData('TEST', '1d', 1000);
      assertTimelineConsistency(result);
      expect(result).toHaveLength(28); /* 28 unique days after deduplication */
    });

    it('Overlap Case 5: Multiple overlaps across 3 pages', async () => {
      setupOverlapFetch([
        { size: 500, startIndex: 0 },
        { size: 500, startIndex: 450 },
        { size: 500, startIndex: 900 },
      ]);
      const result = await provider.getMarketData('TEST', '1d', 1500);
      assertTimelineConsistency(result);
      expect(result).toHaveLength(28); /* 28 unique days after deduplication */
    });

    it('Overlap Case 6: Overlaps at every page boundary (4 pages)', async () => {
      setupOverlapFetch([
        { size: 500, startIndex: 0 },
        { size: 500, startIndex: 480 },
        { size: 500, startIndex: 960 },
        { size: 500, startIndex: 1440 },
      ]);
      const result = await provider.getMarketData('TEST', '1d', 2000);
      assertTimelineConsistency(result);
      expect(result).toHaveLength(28); /* 28 unique days after deduplication */
    });

    it('Overlap Case 7: Random overlaps with varying sizes', async () => {
      setupOverlapFetch([
        { size: 500, startIndex: 0 },
        { size: 400, startIndex: 470 },
        { size: 350, startIndex: 820 },
      ]);
      const result = await provider.getMarketData('TEST', '1d', 1500);
      assertTimelineConsistency(result);
      expect(result).toHaveLength(28); /* 28 unique days after deduplication */
    });

    it('Overlap Case 8: Backward overlap (page 2 starts before page 1 ends)', async () => {
      setupOverlapFetch([
        { size: 500, startIndex: 0 },
        { size: 500, startIndex: 250 },
      ]);
      const result = await provider.getMarketData('TEST', '1d', 1000);
      assertTimelineConsistency(result);
      expect(result).toHaveLength(28); /* 28 unique days after deduplication */
    });

    it('Overlap Case 9: Exact duplicate - page 2 identical to page 1', async () => {
      setupOverlapFetch([
        { size: 500, startIndex: 0 },
        { size: 500, startIndex: 0 },
      ]);
      const result = await provider.getMarketData('TEST', '1d', 1000);
      assertTimelineConsistency(result);
      expect(result).toHaveLength(28); /* 28 unique days after deduplication */
    });

    it('Overlap Case 10: Interleaved duplicates across 5 pages', async () => {
      setupOverlapFetch([
        { size: 500, startIndex: 0 },
        { size: 500, startIndex: 490 },
        { size: 500, startIndex: 980 },
        { size: 500, startIndex: 1470 },
        { size: 500, startIndex: 1960 },
      ]);
      const result = await provider.getMarketData('TEST', '1d', 2500);
      assertTimelineConsistency(result);
      expect(result).toHaveLength(28); /* 28 unique days after deduplication */
    });
  });
});
