import { describe, it, expect, beforeEach, vi } from 'vitest';
import { ProviderManager } from '../../src/classes/ProviderManager.js';

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
    it('should return data from first successful provider', async() => {
      const marketData = [{ openTime: 1000, open: 100, high: 105, low: 95, close: 102 }];
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

    it('should fallback to second provider when first fails', async() => {
      const marketData = [{ openTime: 1000, open: 100, high: 105, low: 95, close: 102 }];
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

    it('should fallback through all providers in chain', async() => {
      const marketData = [{ openTime: 1000, open: 100, high: 105, low: 95, close: 102 }];
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

    it('should throw error when all providers fail', async() => {
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

    it('should skip provider returning empty array', async() => {
      const marketData = [{ openTime: 1000, open: 100, high: 105, low: 95, close: 102 }];
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

    it('should skip provider returning null', async() => {
      const marketData = [{ openTime: 1000, open: 100, high: 105, low: 95, close: 102 }];
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

    it('should skip provider returning undefined', async() => {
      const marketData = [{ openTime: 1000, open: 100, high: 105, low: 95, close: 102 }];
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

    it('should pass symbol, timeframe, and bars to provider', async() => {
      mockProvider1.getMarketData.mockResolvedValue([{ openTime: 1000 }]);

      const chain = [{ name: 'Provider1', instance: mockProvider1 }];
      manager = new ProviderManager(chain, mockLogger);

      await manager.fetchMarketData('AAPL', 'W', 200);

      expect(mockProvider1.getMarketData).toHaveBeenCalledWith('AAPL', 'W', 200);
    });

    it('should return provider instance in result', async() => {
      const marketData = [{ openTime: 1000, open: 100, high: 105, low: 95, close: 102 }];
      mockProvider1.getMarketData.mockResolvedValue(marketData);

      const chain = [{ name: 'Provider1', instance: mockProvider1 }];
      manager = new ProviderManager(chain, mockLogger);

      const result = await manager.fetchMarketData('BTCUSDT', 'D', 100);

      expect(result.instance).toBe(mockProvider1);
    });
  });
});
