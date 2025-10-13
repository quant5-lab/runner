import { describe, it, expect, beforeEach, vi } from 'vitest';
import ApiStatsCollector from '../../src/utils/ApiStatsCollector.js';

describe('ApiStatsCollector', () => {
  let collector;
  let mockLogger;

  beforeEach(() => {
    /* Reset singleton instance before each test */
    ApiStatsCollector.instance = null;
    collector = new ApiStatsCollector();
    
    mockLogger = {
      info: vi.fn(),
      debug: vi.fn(),
      error: vi.fn(),
    };
    
    vi.clearAllMocks();
  });

  describe('constructor and singleton pattern', () => {
    it('should create singleton instance', () => {
      const collector1 = new ApiStatsCollector();
      const collector2 = new ApiStatsCollector();
      
      expect(collector1).toBe(collector2);
    });

    it('should initialize with empty stats', () => {
      const stats = collector.getSummary();
      
      expect(stats.totalRequests).toBe(0);
      expect(stats.cacheHits).toBe(0);
      expect(stats.cacheMisses).toBe(0);
      expect(stats.cacheHitRate).toBe('0%');
      expect(stats.byTimeframe).toEqual({});
      expect(stats.byProvider).toEqual({});
    });
  });

  describe('reset()', () => {
    it('should clear all stats', () => {
      collector.recordRequest('MOEX', '10m');
      collector.recordCacheHit();
      collector.recordCacheMiss();
      
      collector.reset();
      const stats = collector.getSummary();
      
      expect(stats.totalRequests).toBe(0);
      expect(stats.cacheHits).toBe(0);
      expect(stats.cacheMisses).toBe(0);
      expect(stats.byTimeframe).toEqual({});
      expect(stats.byProvider).toEqual({});
    });
  });

  describe('recordRequest()', () => {
    it('should increment total requests', () => {
      collector.recordRequest('MOEX', '10m');
      
      const stats = collector.getSummary();
      expect(stats.totalRequests).toBe(1);
    });

    it('should track requests by provider', () => {
      collector.recordRequest('MOEX', '10m');
      collector.recordRequest('MOEX', '1h');
      collector.recordRequest('Binance', '15m');
      
      const stats = collector.getSummary();
      expect(stats.byProvider.MOEX).toBe(2);
      expect(stats.byProvider.Binance).toBe(1);
    });

    it('should track requests by timeframe', () => {
      collector.recordRequest('MOEX', '10m');
      collector.recordRequest('Binance', '10m');
      collector.recordRequest('MOEX', '1h');
      
      const stats = collector.getSummary();
      expect(stats.byTimeframe['10m']).toBe(2);
      expect(stats.byTimeframe['1h']).toBe(1);
    });

    it('should handle multiple providers and timeframes', () => {
      collector.recordRequest('MOEX', 'D');
      collector.recordRequest('Binance', '1h');
      collector.recordRequest('YahooFinance', 'W');
      collector.recordRequest('MOEX', 'D');
      
      const stats = collector.getSummary();
      expect(stats.totalRequests).toBe(4);
      expect(stats.byProvider.MOEX).toBe(2);
      expect(stats.byProvider.Binance).toBe(1);
      expect(stats.byProvider.YahooFinance).toBe(1);
      expect(stats.byTimeframe.D).toBe(2);
      expect(stats.byTimeframe['1h']).toBe(1);
      expect(stats.byTimeframe.W).toBe(1);
    });
  });

  describe('recordCacheHit()', () => {
    it('should increment cache hits', () => {
      collector.recordCacheHit();
      
      const stats = collector.getSummary();
      expect(stats.cacheHits).toBe(1);
    });

    it('should handle multiple cache hits', () => {
      collector.recordCacheHit();
      collector.recordCacheHit();
      collector.recordCacheHit();
      
      const stats = collector.getSummary();
      expect(stats.cacheHits).toBe(3);
    });
  });

  describe('recordCacheMiss()', () => {
    it('should increment cache misses', () => {
      collector.recordCacheMiss();
      
      const stats = collector.getSummary();
      expect(stats.cacheMisses).toBe(1);
    });

    it('should handle multiple cache misses', () => {
      collector.recordCacheMiss();
      collector.recordCacheMiss();
      
      const stats = collector.getSummary();
      expect(stats.cacheMisses).toBe(2);
    });
  });

  describe('getSummary()', () => {
    it('should calculate cache hit rate correctly', () => {
      collector.recordCacheHit();
      collector.recordCacheHit();
      collector.recordCacheMiss();
      collector.recordCacheMiss();
      
      const stats = collector.getSummary();
      expect(stats.cacheHitRate).toBe('50.0%');
    });

    it('should calculate 100% cache hit rate', () => {
      collector.recordCacheHit();
      collector.recordCacheHit();
      collector.recordCacheHit();
      
      const stats = collector.getSummary();
      expect(stats.cacheHitRate).toBe('100.0%');
    });

    it('should calculate 0% cache hit rate with only misses', () => {
      collector.recordCacheMiss();
      collector.recordCacheMiss();
      
      const stats = collector.getSummary();
      expect(stats.cacheHitRate).toBe('0.0%');
    });

    it('should return 0% when no cache operations', () => {
      const stats = collector.getSummary();
      expect(stats.cacheHitRate).toBe('0%');
    });

    it('should return complete stats object', () => {
      collector.recordRequest('MOEX', '10m');
      collector.recordRequest('Binance', '1h');
      collector.recordCacheHit();
      collector.recordCacheMiss();
      
      const stats = collector.getSummary();
      
      expect(stats).toHaveProperty('totalRequests');
      expect(stats).toHaveProperty('cacheHits');
      expect(stats).toHaveProperty('cacheMisses');
      expect(stats).toHaveProperty('cacheHitRate');
      expect(stats).toHaveProperty('byTimeframe');
      expect(stats).toHaveProperty('byProvider');
      
      expect(stats.totalRequests).toBe(2);
      expect(stats.cacheHits).toBe(1);
      expect(stats.cacheMisses).toBe(1);
      expect(stats.cacheHitRate).toBe('50.0%');
      expect(stats.byTimeframe['10m']).toBe(1);
      expect(stats.byTimeframe['1h']).toBe(1);
      expect(stats.byProvider.MOEX).toBe(1);
      expect(stats.byProvider.Binance).toBe(1);
    });
  });

  describe('logSummary()', () => {
    it('should call logger.debug with stats summary', () => {
      collector.recordRequest('MOEX', '10m');
      collector.recordCacheHit();
      
      collector.logSummary(mockLogger);
      
      expect(mockLogger.debug).toHaveBeenCalledWith(expect.stringContaining('API Statistics:'));
      expect(mockLogger.debug).toHaveBeenCalledWith(expect.stringContaining('"totalRequests": 1'));
    });

    it('should log correct stats object', () => {
      collector.recordRequest('MOEX', 'D');
      collector.recordRequest('Binance', '1h');
      collector.recordCacheHit();
      collector.recordCacheMiss();
      
      collector.logSummary(mockLogger);
      
      const loggedMessage = mockLogger.debug.mock.calls[0][0];
      expect(loggedMessage).toContain('API Statistics:');
      expect(loggedMessage).toContain('"totalRequests": 2');
      expect(loggedMessage).toContain('"cacheHits": 1');
      expect(loggedMessage).toContain('"cacheMisses": 1');
      expect(loggedMessage).toContain('"cacheHitRate": "50.0%"');
    });

    it('should log empty stats when no operations recorded', () => {
      collector.logSummary(mockLogger);
      
      expect(mockLogger.debug).toHaveBeenCalled();
      const loggedMessage = mockLogger.debug.mock.calls[0][0];
      expect(loggedMessage).toContain('API Statistics:');
      expect(loggedMessage).toContain('"totalRequests": 0');
      expect(loggedMessage).toContain('"cacheHits": 0');
      expect(loggedMessage).toContain('"cacheMisses": 0');
    });
  });

  describe('integration scenario', () => {
    it('should track realistic strategy execution stats', () => {
      /* Simulate strategy with security() prefetch:
       * - Initial request for main symbol data (cache miss)
       * - Prefetch for security() calls (cache miss)
       * - security() calls hit cache
       */
      
      /* Main data request - MOEX provider */
      collector.recordRequest('MOEX', '10m');
      
      /* security() prefetch - daily data */
      collector.recordRequest('MOEX', 'D');
      
      /* Strategy execution - 25 security() calls hit cache */
      for (let i = 0; i < 25; i++) {
        collector.recordCacheHit();
      }
      
      const stats = collector.getSummary();
      
      expect(stats.totalRequests).toBe(2);
      expect(stats.cacheHits).toBe(25);
      expect(stats.cacheMisses).toBe(0);
      expect(stats.cacheHitRate).toBe('100.0%');
      expect(stats.byProvider.MOEX).toBe(2);
      expect(stats.byTimeframe['10m']).toBe(1);
      expect(stats.byTimeframe.D).toBe(1);
    });

    it('should track multi-provider scenario', () => {
      /* Try MOEX first (no data) */
      collector.recordRequest('MOEX', '1h');
      
      /* Fallback to Binance (success) */
      collector.recordRequest('Binance', '1h');
      
      /* Cache operations */
      collector.recordCacheMiss();
      collector.recordCacheHit();
      collector.recordCacheHit();
      
      const stats = collector.getSummary();
      
      expect(stats.totalRequests).toBe(2);
      expect(stats.byProvider.MOEX).toBe(1);
      expect(stats.byProvider.Binance).toBe(1);
      expect(stats.cacheHits).toBe(2);
      expect(stats.cacheMisses).toBe(1);
      expect(stats.cacheHitRate).toBe('66.7%');
    });
  });
});
