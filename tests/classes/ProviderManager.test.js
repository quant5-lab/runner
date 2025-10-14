import { describe, it, expect, beforeEach, vi } from 'vitest';
import { ProviderManager } from '../../src/classes/ProviderManager.js';
import { TimeframeError } from '../../src/errors/TimeframeError.js';

describe('ProviderManager', () => {
  let manager;
  let mockProvider1;
  let mockProvider2;
  let mockProvider3;
  let mockLogger;

  beforeEach(() => {
    mockProvider1 = {
      getMarketData: vi.fn(),
    };
    mockProvider2 = {
      getMarketData: vi.fn(),
    };
    mockProvider3 = {
      getMarketData: vi.fn(),
    };
    mockLogger = {
      log: vi.fn(),
      debug: vi.fn(),
      error: vi.fn(),
    };
  });

  describe('constructor', () => {
    it('should store provider chain', () => {
      const chain = [
        { name: 'Provider1', instance: mockProvider1 },
        { name: 'Provider2', instance: mockProvider2 },
      ];
      manager = new ProviderManager(chain, mockLogger);
      expect(manager.providerChain).toEqual(chain);
    });
  });

  describe('fetchMarketData()', () => {
    it('should return data from first successful provider', async () => {
      const currentTime = Date.now();
      const marketData = [
        {
          openTime: currentTime,
          closeTime: currentTime,
          open: 100,
          high: 105,
          low: 95,
          close: 102,
        },
      ];
      mockProvider1.getMarketData.mockResolvedValue(marketData);

      const chain = [{ name: 'Provider1', instance: mockProvider1 }];
      manager = new ProviderManager(chain, mockLogger);

      const result = await manager.fetchMarketData('BTCUSDT', 'D', 100);

      expect(result).toEqual({
        provider: 'Provider1',
        data: marketData,
        instance: mockProvider1,
      });
      expect(mockProvider1.getMarketData).toHaveBeenCalledWith('BTCUSDT', 'D', 100);
    });

    it('should fallback to second provider when first fails', async () => {
      const currentTime = Date.now();
      const marketData = [
        {
          openTime: currentTime,
          closeTime: currentTime,
          open: 100,
          high: 105,
          low: 95,
          close: 102,
        },
      ];
      mockProvider1.getMarketData.mockRejectedValue(new Error('Provider1 failed'));
      mockProvider2.getMarketData.mockResolvedValue(marketData);

      const chain = [
        { name: 'Provider1', instance: mockProvider1 },
        { name: 'Provider2', instance: mockProvider2 },
      ];
      manager = new ProviderManager(chain, mockLogger);

      const result = await manager.fetchMarketData('BTCUSDT', 'D', 100);

      expect(result.provider).toBe('Provider2');
      expect(result.data).toEqual(marketData);
      expect(mockProvider1.getMarketData).toHaveBeenCalled();
      expect(mockProvider2.getMarketData).toHaveBeenCalled();
    });

    it('should fallback through all providers in chain', async () => {
      const currentTime = Date.now();
      const marketData = [
        {
          openTime: currentTime,
          closeTime: currentTime,
          open: 100,
          high: 105,
          low: 95,
          close: 102,
        },
      ];
      mockProvider1.getMarketData.mockRejectedValue(new Error('Fail'));
      mockProvider2.getMarketData.mockRejectedValue(new Error('Fail'));
      mockProvider3.getMarketData.mockResolvedValue(marketData);

      const chain = [
        { name: 'Provider1', instance: mockProvider1 },
        { name: 'Provider2', instance: mockProvider2 },
        { name: 'Provider3', instance: mockProvider3 },
      ];
      manager = new ProviderManager(chain, mockLogger);

      const result = await manager.fetchMarketData('BTCUSDT', 'D', 100);

      expect(result.provider).toBe('Provider3');
      expect(mockProvider1.getMarketData).toHaveBeenCalled();
      expect(mockProvider2.getMarketData).toHaveBeenCalled();
      expect(mockProvider3.getMarketData).toHaveBeenCalled();
    });

    it('should throw error when all providers fail', async () => {
      mockProvider1.getMarketData.mockRejectedValue(new Error('Fail1'));
      mockProvider2.getMarketData.mockRejectedValue(new Error('Fail2'));

      const chain = [
        { name: 'Provider1', instance: mockProvider1 },
        { name: 'Provider2', instance: mockProvider2 },
      ];
      manager = new ProviderManager(chain, mockLogger);

      await expect(manager.fetchMarketData('BTCUSDT', 'D', 100)).rejects.toThrow(
        'All providers failed for symbol: BTCUSDT',
      );
    });

    it('should skip provider returning empty array', async () => {
      const currentTime = Date.now();
      const marketData = [
        {
          openTime: currentTime,
          closeTime: currentTime,
          open: 100,
          high: 105,
          low: 95,
          close: 102,
        },
      ];
      mockProvider1.getMarketData.mockResolvedValue([]);
      mockProvider2.getMarketData.mockResolvedValue(marketData);

      const chain = [
        { name: 'Provider1', instance: mockProvider1 },
        { name: 'Provider2', instance: mockProvider2 },
      ];
      manager = new ProviderManager(chain, mockLogger);

      const result = await manager.fetchMarketData('BTCUSDT', 'D', 100);

      expect(result.provider).toBe('Provider2');
      expect(result.data).toEqual(marketData);
    });

    it('should skip provider returning null', async () => {
      const currentTime = Date.now();
      const marketData = [
        {
          openTime: currentTime,
          closeTime: currentTime,
          open: 100,
          high: 105,
          low: 95,
          close: 102,
        },
      ];
      mockProvider1.getMarketData.mockResolvedValue(null);
      mockProvider2.getMarketData.mockResolvedValue(marketData);

      const chain = [
        { name: 'Provider1', instance: mockProvider1 },
        { name: 'Provider2', instance: mockProvider2 },
      ];
      manager = new ProviderManager(chain, mockLogger);

      const result = await manager.fetchMarketData('BTCUSDT', 'D', 100);

      expect(result.provider).toBe('Provider2');
    });

    it('should skip provider returning undefined', async () => {
      const currentTime = Date.now();
      const marketData = [
        {
          openTime: currentTime,
          closeTime: currentTime,
          open: 100,
          high: 105,
          low: 95,
          close: 102,
        },
      ];
      mockProvider1.getMarketData.mockResolvedValue(undefined);
      mockProvider2.getMarketData.mockResolvedValue(marketData);

      const chain = [
        { name: 'Provider1', instance: mockProvider1 },
        { name: 'Provider2', instance: mockProvider2 },
      ];
      manager = new ProviderManager(chain, mockLogger);

      const result = await manager.fetchMarketData('BTCUSDT', 'D', 100);

      expect(result.provider).toBe('Provider2');
    });

    it('should pass symbol, timeframe, and bars to provider', async () => {
      const currentTime = Date.now();
      mockProvider1.getMarketData.mockResolvedValue([
        { openTime: currentTime, closeTime: currentTime },
      ]);

      const chain = [{ name: 'Provider1', instance: mockProvider1 }];
      manager = new ProviderManager(chain, mockLogger);

      await manager.fetchMarketData('AAPL', 'W', 200);

      expect(mockProvider1.getMarketData).toHaveBeenCalledWith('AAPL', 'W', 200);
    });

    it('should return provider instance in result', async () => {
      const currentTime = Date.now();
      const marketData = [
        {
          openTime: currentTime,
          closeTime: currentTime,
          open: 100,
          high: 105,
          low: 95,
          close: 102,
        },
      ];
      mockProvider1.getMarketData.mockResolvedValue(marketData);

      const chain = [{ name: 'Provider1', instance: mockProvider1 }];
      manager = new ProviderManager(chain, mockLogger);

      const result = await manager.fetchMarketData('BTCUSDT', 'D', 100);

      expect(result.instance).toBe(mockProvider1);
    });
  });

  describe('validateDataFreshness() - closeTime fallback', () => {
    it('should use time field when present for freshness validation', async () => {
      const currentTime = Date.now();
      const marketData = [
        {
          time: Math.floor(currentTime / 1000),
          closeTime: Math.floor((currentTime - 10 * 24 * 60 * 60 * 1000) / 1000), // 10 days old
          open: 100,
          high: 105,
          low: 95,
          close: 102,
        },
      ];
      mockProvider1.getMarketData.mockResolvedValue(marketData);

      const chain = [{ name: 'Provider1', instance: mockProvider1 }];
      manager = new ProviderManager(chain, mockLogger);

      const result = await manager.fetchMarketData('BTCUSDT', 'D', 100);

      expect(result.provider).toBe('Provider1');
      expect(result.data).toEqual(marketData);
    });

    it('should fallback to closeTime when time field is missing', async () => {
      const currentTime = Date.now();
      const marketData = [
        {
          closeTime: Math.floor(currentTime / 1000),
          open: 100,
          high: 105,
          low: 95,
          close: 102,
        },
      ];
      mockProvider1.getMarketData.mockResolvedValue(marketData);

      const chain = [{ name: 'Provider1', instance: mockProvider1 }];
      manager = new ProviderManager(chain, mockLogger);

      const result = await manager.fetchMarketData('BTCUSDT', 'D', 100);

      expect(result.provider).toBe('Provider1');
      expect(result.data).toEqual(marketData);
    });

    it('should reject stale data using closeTime field', async () => {
      const tenDaysAgo = Date.now() - 10 * 24 * 60 * 60 * 1000;
      const marketData = [
        {
          closeTime: Math.floor(tenDaysAgo / 1000),
          open: 100,
          high: 105,
          low: 95,
          close: 102,
        },
      ];
      mockProvider1.getMarketData.mockResolvedValue(marketData);

      const chain = [{ name: 'Provider1', instance: mockProvider1 }];
      manager = new ProviderManager(chain, mockLogger);

      await expect(manager.fetchMarketData('BTCUSDT', 'D', 100)).rejects.toThrow(
        /Provider1 returned stale data/,
      );
    });

    it('should handle millisecond timestamps for closeTime', async () => {
      const currentTime = Date.now();
      const marketData = [
        {
          closeTime: currentTime, // milliseconds
          open: 100,
          high: 105,
          low: 95,
          close: 102,
        },
      ];
      mockProvider1.getMarketData.mockResolvedValue(marketData);

      const chain = [{ name: 'Provider1', instance: mockProvider1 }];
      manager = new ProviderManager(chain, mockLogger);

      const result = await manager.fetchMarketData('BTCUSDT', 'D', 100);

      expect(result.provider).toBe('Provider1');
    });

    it('should handle second timestamps for closeTime', async () => {
      const currentTime = Math.floor(Date.now() / 1000);
      const marketData = [
        {
          closeTime: currentTime, // seconds
          open: 100,
          high: 105,
          low: 95,
          close: 102,
        },
      ];
      mockProvider1.getMarketData.mockResolvedValue(marketData);

      const chain = [{ name: 'Provider1', instance: mockProvider1 }];
      manager = new ProviderManager(chain, mockLogger);

      const result = await manager.fetchMarketData('BTCUSDT', 'D', 100);

      expect(result.provider).toBe('Provider1');
    });
  });

  describe('TimeframeError handling with 3 mocked providers', () => {
    it('should stop chain when first provider throws TimeframeError', async () => {
      const supportedTimeframes = ['1m', '10m', '1h', '1d'];
      const timeframeError = new TimeframeError('5s', 'SBER', 'MockProvider1', supportedTimeframes);

      mockProvider1.getMarketData.mockRejectedValue(timeframeError);
      mockProvider2.getMarketData.mockResolvedValue([
        { openTime: Date.now(), closeTime: Date.now() },
      ]);
      mockProvider3.getMarketData.mockResolvedValue([
        { openTime: Date.now(), closeTime: Date.now() },
      ]);

      const chain = [
        { name: 'MockProvider1', instance: mockProvider1 },
        { name: 'MockProvider2', instance: mockProvider2 },
        { name: 'MockProvider3', instance: mockProvider3 },
      ];
      manager = new ProviderManager(chain, mockLogger);

      await expect(manager.fetchMarketData('SBER', '5s', 100)).rejects.toThrow(
        "Timeframe '5s' not supported for symbol 'SBER'",
      );

      expect(mockProvider1.getMarketData).toHaveBeenCalledWith('SBER', '5s', 100);
      expect(mockProvider2.getMarketData).not.toHaveBeenCalled();
      expect(mockProvider3.getMarketData).not.toHaveBeenCalled();
    });

    it('should include supported timeframes list in error message', async () => {
      const supportedTimeframes = ['1m', '10m', '1h', '1d', '1w', '1M'];
      const timeframeError = new TimeframeError('5s', 'CHMF', 'MOEX', supportedTimeframes);

      mockProvider1.getMarketData.mockRejectedValue(timeframeError);

      const chain = [{ name: 'MOEX', instance: mockProvider1 }];
      manager = new ProviderManager(chain, mockLogger);

      await expect(manager.fetchMarketData('CHMF', '5s', 100)).rejects.toThrow(
        'Supported timeframes: 1m, 10m, 1h, 1d, 1w, 1M',
      );
    });

    it('should continue chain when provider returns empty array', async () => {
      const currentTime = Date.now();
      const marketData = [{ openTime: currentTime, closeTime: currentTime, open: 100, close: 102 }];

      mockProvider1.getMarketData.mockResolvedValue([]);
      mockProvider2.getMarketData.mockResolvedValue([]);
      mockProvider3.getMarketData.mockResolvedValue(marketData);

      const chain = [
        { name: 'MockProvider1', instance: mockProvider1 },
        { name: 'MockProvider2', instance: mockProvider2 },
        { name: 'MockProvider3', instance: mockProvider3 },
      ];
      manager = new ProviderManager(chain, mockLogger);

      const result = await manager.fetchMarketData('BTCUSDT', '15m', 100);

      expect(result.provider).toBe('MockProvider3');
      expect(result.data).toEqual(marketData);
      expect(mockProvider1.getMarketData).toHaveBeenCalled();
      expect(mockProvider2.getMarketData).toHaveBeenCalled();
      expect(mockProvider3.getMarketData).toHaveBeenCalled();
    });

    it('should stop chain when middle provider throws TimeframeError', async () => {
      const supportedTimeframes = ['1m', '3m', '5m', '15m', '1h'];
      const timeframeError = new TimeframeError('5s', 'BTCUSDT', 'Binance', supportedTimeframes);

      mockProvider1.getMarketData.mockResolvedValue([]);
      mockProvider2.getMarketData.mockRejectedValue(timeframeError);
      mockProvider3.getMarketData.mockResolvedValue([
        { openTime: Date.now(), closeTime: Date.now() },
      ]);

      const chain = [
        { name: 'MOEX', instance: mockProvider1 },
        { name: 'Binance', instance: mockProvider2 },
        { name: 'Yahoo', instance: mockProvider3 },
      ];
      manager = new ProviderManager(chain, mockLogger);

      await expect(manager.fetchMarketData('BTCUSDT', '5s', 100)).rejects.toThrow(
        "Timeframe '5s' not supported for symbol 'BTCUSDT'",
      );

      expect(mockProvider1.getMarketData).toHaveBeenCalled();
      expect(mockProvider2.getMarketData).toHaveBeenCalled();
      expect(mockProvider3.getMarketData).not.toHaveBeenCalled();
    });

    it('should continue chain on non-TimeframeError exceptions', async () => {
      const currentTime = Date.now();
      const marketData = [{ openTime: currentTime, closeTime: currentTime, open: 100, close: 102 }];

      mockProvider1.getMarketData.mockRejectedValue(new Error('Network timeout'));
      mockProvider2.getMarketData.mockRejectedValue(new Error('API rate limit'));
      mockProvider3.getMarketData.mockResolvedValue(marketData);

      const chain = [
        { name: 'Provider1', instance: mockProvider1 },
        { name: 'Provider2', instance: mockProvider2 },
        { name: 'Provider3', instance: mockProvider3 },
      ];
      manager = new ProviderManager(chain, mockLogger);

      const result = await manager.fetchMarketData('AAPL', '1h', 100);

      expect(result.provider).toBe('Provider3');
      expect(result.data).toEqual(marketData);
      expect(mockProvider1.getMarketData).toHaveBeenCalled();
      expect(mockProvider2.getMarketData).toHaveBeenCalled();
      expect(mockProvider3.getMarketData).toHaveBeenCalled();
    });

    it('should preserve original TimeframeError properties', async () => {
      const supportedTimeframes = ['1m', '2m', '5m', '15m', '1h', '1d'];
      const timeframeError = new TimeframeError('7m', 'AAPL', 'Yahoo', supportedTimeframes);

      mockProvider1.getMarketData.mockRejectedValue(timeframeError);

      const chain = [{ name: 'Yahoo', instance: mockProvider1 }];
      manager = new ProviderManager(chain, mockLogger);

      try {
        await manager.fetchMarketData('AAPL', '7m', 100);
        expect.fail('Should have thrown error');
      } catch (error) {
        expect(error.message).toContain("Timeframe '7m' not supported for symbol 'AAPL'");
        expect(error.message).toContain('Supported timeframes: 1m, 2m, 5m, 15m, 1h, 1d');
      }
    });

    it('should re-throw stale data error without continuing chain', async () => {
      const staleError = new Error(
        'Provider1 returned stale data for BTCUSDT 1h: latest candle is 10 days old',
      );

      mockProvider1.getMarketData.mockRejectedValue(staleError);
      mockProvider2.getMarketData.mockResolvedValue([
        { openTime: Date.now(), closeTime: Date.now() },
      ]);

      const chain = [
        { name: 'Provider1', instance: mockProvider1 },
        { name: 'Provider2', instance: mockProvider2 },
      ];
      manager = new ProviderManager(chain, mockLogger);

      await expect(manager.fetchMarketData('BTCUSDT', '1h', 100)).rejects.toThrow(
        'returned stale data',
      );

      expect(mockProvider1.getMarketData).toHaveBeenCalled();
      expect(mockProvider2.getMarketData).not.toHaveBeenCalled();
    });
  });
});
