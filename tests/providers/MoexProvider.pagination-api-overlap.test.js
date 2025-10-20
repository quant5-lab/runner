import { describe, it, expect, beforeEach, vi, beforeAll, afterAll } from 'vitest';
import { MoexProvider } from '../../src/providers/MoexProvider.js';
import { createServer } from 'http';

/* Mock fetch globally to prevent any real API calls */
const originalFetch = global.fetch;
const mockFetch = vi.fn();

/* TEST 2: Real provider + fake API - verify API parameters don't overlap */
describe('MoexProvider Pagination Overlap Prevention - Fake API', () => {
  let provider;
  let mockStatsCollector;
  let mockLogger;
  let fakeServer;
  let capturedApiRequests;
  let serverPort;

  beforeAll(async () => {
    /* Replace global fetch with mock that only allows localhost */
    global.fetch = mockFetch.mockImplementation(async (url, options) => {
      if (!url.toString().includes('localhost')) {
        throw new Error(`SECURITY VIOLATION: Test attempted to fetch non-localhost URL: ${url}`);
      }
      return originalFetch(url, options);
    });

    /* Create HTTP server ONCE for all tests - SINGLETON */
    await new Promise((resolve) => {
      fakeServer = createServer((req, res) => {
        const url = new URL(req.url, `http://localhost:${serverPort}`);
        const start = parseInt(url.searchParams.get('start') || '0', 10);
        const interval = url.searchParams.get('interval');

        capturedApiRequests.push({
          url: req.url,
          start,
          interval,
          from: url.searchParams.get('from'),
          till: url.searchParams.get('till'),
        });

        /* Determine batch size based on request index */
        const requestIndex = capturedApiRequests.length - 1;
        let batchSize = 500;

        /* Simulate various batch patterns */
        if (url.pathname.includes('partial-last')) {
          batchSize = requestIndex === 0 ? 500 : 200;
        } else if (url.pathname.includes('empty-trigger')) {
          batchSize = requestIndex >= 4 ? 0 : 500;
        } else if (url.pathname.includes('small-last')) {
          batchSize = requestIndex === 1 ? 89 : 500;
        }

        const mockCandles = Array.from({ length: batchSize }, (_, i) => [
          '100',
          '105',
          '110',
          '90',
          '1000',
          '1000',
          `2024-01-${String(start + i + 1).padStart(2, '0')} 00:00:00`,
          `2024-01-${String(start + i + 1).padStart(2, '0')} 23:59:59`,
        ]);

        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ candles: { data: mockCandles } }));
      });

      fakeServer.listen(0, () => {
        serverPort = fakeServer.address().port;
        resolve();
      });
    });
  });

  afterAll(() => {
    /* Restore original fetch */
    global.fetch = originalFetch;

    /* Close server once after all tests */
    return new Promise((resolve) => {
      if (fakeServer) {
        fakeServer.close(() => resolve());
      } else {
        resolve();
      }
    });
  });

  beforeEach(() => {
    mockFetch.mockClear();
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
    capturedApiRequests = [];

    provider = new MoexProvider(mockLogger, mockStatsCollector);
    provider.baseUrl = `http://localhost:${serverPort}`;
    vi.clearAllMocks();
  });

  /* Verify no overlapping API start parameters */
  const assertNoApiOverlap = () => {
    for (let i = 1; i < capturedApiRequests.length; i++) {
      const prevStart = capturedApiRequests[i - 1].start;
      const currStart = capturedApiRequests[i].start;

      /* Current start must be exactly prevStart + 500 */
      expect(currStart).toBe(prevStart + 500);

      /* No gap or overlap */
      expect(currStart - prevStart).toBe(500);
    }
  };

  describe('API pagination - 50 test cases', () => {
    it('Case 1: 2 API requests (500 + 500)', async () => {
      await provider.getMarketData('TEST', '1d', 1000);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(2);
      expect(capturedApiRequests[0].start).toBe(0);
      expect(capturedApiRequests[1].start).toBe(500);
    });

    it('Case 2: 3 API requests (500 + 500 + 500)', async () => {
      await provider.getMarketData('TEST', '1d', 1500);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(3);
    });

    it('Case 3: 5 API requests (500×5)', async () => {
      await provider.getMarketData('TEST', '1d', 2500);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(5);
    });

    it('Case 4: 10 API requests (500×10)', async () => {
      await provider.getMarketData('TEST', '1d', 5000);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(10);
    });

    it('Case 5: 2 API requests with partial last (500 + 200)', async () => {
      provider.baseUrl = `http://localhost:${serverPort}/partial-last`;
      await provider.getMarketData('TEST', '1d', 1000);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(2);
    });

    it('Case 6: 2 API requests with partial last (500 + 200 stops)', async () => {
      provider.baseUrl = `http://localhost:${serverPort}/partial-last`;
      await provider.getMarketData('TEST', '1d', 2000);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(2);
    });

    it('Case 7: 6 API requests (500×6)', async () => {
      await provider.getMarketData('TEST', '1d', 3000);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(6);
    });

    it('Case 8: 2 API requests with limit mid-batch (500 + partial)', async () => {
      await provider.getMarketData('TEST', '1d', 700);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(2);
    });

    it('Case 9: 4 API requests (500×4)', async () => {
      await provider.getMarketData('TEST', '1d', 2000);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(4);
    });

    it('Case 10: 7 API requests (500×7)', async () => {
      await provider.getMarketData('TEST', '1d', 3500);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(7);
    });

    it('Case 11: 8 API requests (500×8)', async () => {
      await provider.getMarketData('TEST', '1d', 4000);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(8);
    });

    it('Case 12: 9 API requests (500×9)', async () => {
      await provider.getMarketData('TEST', '1d', 4500);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(9);
    });

    it('Case 13: 11 API requests (500×11)', async () => {
      await provider.getMarketData('TEST', '1d', 5500);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(11);
    });

    it('Case 14: 12 API requests (500×12)', async () => {
      await provider.getMarketData('TEST', '1d', 6000);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(12);
    });

    it('Case 15: 13 API requests (500×13)', async () => {
      await provider.getMarketData('TEST', '1d', 6500);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(13);
    });

    it('Case 16: 14 API requests (500×14)', async () => {
      await provider.getMarketData('TEST', '1d', 7000);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(14);
    });

    it('Case 17: 15 API requests (500×15)', async () => {
      await provider.getMarketData('TEST', '1d', 7500);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(15);
    });

    it('Case 18: 16 API requests (500×16)', async () => {
      await provider.getMarketData('TEST', '1d', 8000);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(16);
    });

    it('Case 19: 17 API requests (500×17)', async () => {
      await provider.getMarketData('TEST', '1d', 8500);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(17);
    });

    it('Case 20: 18 API requests (500×18)', async () => {
      await provider.getMarketData('TEST', '1d', 9000);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(18);
    });

    it('Case 21: 19 API requests (500×19)', async () => {
      await provider.getMarketData('TEST', '1d', 9500);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(19);
    });

    it('Case 22: 20 API requests (500×20)', async () => {
      await provider.getMarketData('TEST', '1d', 10000);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(20);
    });

    it('Case 23: API parameters include interval', async () => {
      await provider.getMarketData('TEST', '1d', 1000);
      assertNoApiOverlap();
      capturedApiRequests.forEach((req) => {
        expect(req.interval).toBe('24');
      });
    });

    it('Case 24: API parameters include from/till', async () => {
      await provider.getMarketData('TEST', '1d', 1000);
      assertNoApiOverlap();
      capturedApiRequests.forEach((req) => {
        expect(req.from).toBeTruthy();
        expect(req.till).toBeTruthy();
      });
    });

    it('Case 25: 5 API requests with empty 5th triggers stop', async () => {
      provider.baseUrl = `http://localhost:${serverPort}/empty-trigger`;
      await provider.getMarketData('TEST', '1d', 3000);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(5);
    });

    it('Case 26: 3 API requests with limit=1250', async () => {
      await provider.getMarketData('TEST', '1d', 1250);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(3);
    });

    it('Case 27: 6 API requests with limit=2750', async () => {
      await provider.getMarketData('TEST', '1d', 2750);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(6);
    });

    it('Case 28: 7 API requests with limit=3333', async () => {
      await provider.getMarketData('TEST', '1d', 3333);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(7);
    });

    it('Case 29: 8 API requests with limit=3777', async () => {
      await provider.getMarketData('TEST', '1d', 3777);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(8);
    });

    it('Case 30: 9 API requests with limit=4321', async () => {
      await provider.getMarketData('TEST', '1d', 4321);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(9);
    });

    it('Case 31: 10 API requests with limit=4888', async () => {
      await provider.getMarketData('TEST', '1d', 4888);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(10);
    });

    it('Case 32: 12 API requests with limit=5555 (fetches until 6000)', async () => {
      await provider.getMarketData('TEST', '1d', 5555);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(12);
    });

    it('Case 33: 2 API requests with limit=501', async () => {
      await provider.getMarketData('TEST', '1d', 501);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(2);
    });

    it('Case 34: 2 API requests with limit=999', async () => {
      await provider.getMarketData('TEST', '1d', 999);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(2);
    });

    it('Case 35: 3 API requests with limit=1001', async () => {
      await provider.getMarketData('TEST', '1d', 1001);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(3);
    });

    it('Case 36: 5 API requests with limit=2001', async () => {
      await provider.getMarketData('TEST', '1d', 2001);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(5);
    });

    it('Case 37: 2 API requests with limit=750', async () => {
      await provider.getMarketData('TEST', '1d', 750);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(2);
    });

    it('Case 38: First API request has no start parameter', async () => {
      await provider.getMarketData('TEST', '1d', 1000);
      expect(capturedApiRequests[0].start).toBe(0);
      expect(capturedApiRequests[0].url).not.toContain('start=');
    });

    it('Case 39: Second API request has start=500', async () => {
      await provider.getMarketData('TEST', '1d', 1000);
      expect(capturedApiRequests[1].start).toBe(500);
      expect(capturedApiRequests[1].url).toContain('start=500');
    });

    it('Case 40: Third API request has start=1000', async () => {
      await provider.getMarketData('TEST', '1d', 1500);
      expect(capturedApiRequests[2].start).toBe(1000);
      expect(capturedApiRequests[2].url).toContain('start=1000');
    });

    it('Case 41: Fourth API request has start=1500', async () => {
      await provider.getMarketData('TEST', '1d', 2000);
      expect(capturedApiRequests[3].start).toBe(1500);
      expect(capturedApiRequests[3].url).toContain('start=1500');
    });

    it('Case 42: Fifth API request has start=2000', async () => {
      await provider.getMarketData('TEST', '1d', 2500);
      expect(capturedApiRequests[4].start).toBe(2000);
      expect(capturedApiRequests[4].url).toContain('start=2000');
    });

    it('Case 43: Tenth API request has start=4500', async () => {
      await provider.getMarketData('TEST', '1d', 5000);
      expect(capturedApiRequests[9].start).toBe(4500);
      expect(capturedApiRequests[9].url).toContain('start=4500');
    });

    it('Case 44: Twentieth API request has start=9500', async () => {
      await provider.getMarketData('TEST', '1d', 10000);
      expect(capturedApiRequests[19].start).toBe(9500);
      expect(capturedApiRequests[19].url).toContain('start=9500');
    });

    it('Case 45: All API requests have consistent interval', async () => {
      await provider.getMarketData('TEST', '1h', 2000);
      assertNoApiOverlap();
      const intervals = capturedApiRequests.map((req) => req.interval);
      expect(new Set(intervals).size).toBe(1);
    });

    it('Case 46: All API requests have consistent from/till', async () => {
      await provider.getMarketData('TEST', '1d', 2000);
      assertNoApiOverlap();
      const fromDates = capturedApiRequests.map((req) => req.from);
      const tillDates = capturedApiRequests.map((req) => req.till);
      expect(new Set(fromDates).size).toBe(1);
      expect(new Set(tillDates).size).toBe(1);
    });

    it('Case 47: 12 API requests with various limits', async () => {
      await provider.getMarketData('TEST', '1d', 6000);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(12);
    });

    it('Case 48: 15 API requests with various limits', async () => {
      await provider.getMarketData('TEST', '1d', 7500);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(15);
    });

    it('Case 49: 18 API requests with various limits', async () => {
      await provider.getMarketData('TEST', '1d', 9000);
      assertNoApiOverlap();
      expect(capturedApiRequests).toHaveLength(18);
    });

    it('Case 50: Sequential start values across all requests', async () => {
      await provider.getMarketData('TEST', '1d', 5000);
      assertNoApiOverlap();
      const starts = capturedApiRequests.map((req) => req.start);
      const expected = [0, 500, 1000, 1500, 2000, 2500, 3000, 3500, 4000, 4500];
      expect(starts).toEqual(expected);
    });
  });
});
