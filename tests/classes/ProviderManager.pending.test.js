import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ProviderManager } from '../../src/classes/ProviderManager.js';

describe('ProviderManager - pending requests deduplication', () => {
  let providerManager;
  let mockLogger;
  let mockProviderChain;
  let callCount;

  beforeEach(() => {
    callCount = 0;
    mockLogger = {
      log: vi.fn(),
      debug: vi.fn(),
    };

    mockProviderChain = [
      {
        name: 'TestProvider',
        instance: {
          getMarketData: vi.fn(async () => {
            callCount++;
            await new Promise((resolve) => setTimeout(resolve, 10));
            return [{ openTime: 1, close: 100 }];
          }),
        },
      },
    ];

    providerManager = new ProviderManager(mockProviderChain, mockLogger);
  });

  it('should generate correct cache key', () => {
    const key1 = providerManager.getCacheKey('BTCUSDT', '1h', 240);
    const key2 = providerManager.getCacheKey('BTCUSDT', '1h', 240);
    const key3 = providerManager.getCacheKey('ETHUSDT', '1h', 240);

    expect(key1).toBe('BTCUSDT|1h|240');
    expect(key2).toBe(key1);
    expect(key3).not.toBe(key1);
  });

  it('should deduplicate simultaneous identical requests', async () => {
    const promise1 = providerManager.getMarketData('BTCUSDT', '60', 240);
    const promise2 = providerManager.getMarketData('BTCUSDT', '60', 240);
    const promise3 = providerManager.getMarketData('BTCUSDT', '60', 240);

    const [result1, result2, result3] = await Promise.all([promise1, promise2, promise3]);

    expect(callCount).toBe(1);
    expect(result1).toEqual(result2);
    expect(result2).toEqual(result3);
  });

  it('should allow sequential requests after first completes', async () => {
    const result1 = await providerManager.getMarketData('BTCUSDT', '60', 240);
    const result2 = await providerManager.getMarketData('BTCUSDT', '60', 240);

    expect(callCount).toBe(2);
    expect(result1).toEqual(result2);
  });

  it('should not deduplicate different symbols', async () => {
    const promise1 = providerManager.getMarketData('BTCUSDT', '60', 240);
    const promise2 = providerManager.getMarketData('ETHUSDT', '60', 240);

    await Promise.all([promise1, promise2]);

    expect(callCount).toBe(2);
  });

  it('should not deduplicate different timeframes', async () => {
    const promise1 = providerManager.getMarketData('BTCUSDT', '60', 240);
    const promise2 = providerManager.getMarketData('BTCUSDT', '1440', 240);

    await Promise.all([promise1, promise2]);

    expect(callCount).toBe(2);
  });

  it('should not deduplicate different limits', async () => {
    const promise1 = providerManager.getMarketData('BTCUSDT', '60', 240);
    const promise2 = providerManager.getMarketData('BTCUSDT', '60', 500);

    await Promise.all([promise1, promise2]);

    expect(callCount).toBe(2);
  });

  it('should clean up pending map after request completes', async () => {
    await providerManager.getMarketData('BTCUSDT', '60', 240);

    expect(providerManager.pending.size).toBe(0);
  });

  it('should clean up pending map even on error', async () => {
    mockProviderChain[0].instance.getMarketData.mockRejectedValueOnce(new Error('Test error'));

    try {
      await providerManager.getMarketData('BTCUSDT', '60', 240);
    } catch (error) {
      expect(error.message).toContain('All providers failed');
    }

    expect(providerManager.pending.size).toBe(0);
  });
});
