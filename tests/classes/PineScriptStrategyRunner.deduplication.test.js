import { describe, it, expect } from 'vitest';
import { PineScriptStrategyRunner } from '../../src/classes/PineScriptStrategyRunner.js';

describe('PineScriptStrategyRunner - deduplicatePrefetchData', () => {
  const runner = new PineScriptStrategyRunner(null, null);

  it('should deduplicate identical prefetch requests', () => {
    const input = [
      { symbol: 'BTCUSDT', timeframe: 'D', limit: 3 },
      { symbol: 'BTCUSDT', timeframe: 'D', limit: 3 },
      { symbol: 'BTCUSDT', timeframe: 'D', limit: 3 },
    ];

    const result = runner.deduplicatePrefetchData(input);

    expect(result).toHaveLength(1);
    expect(result[0]).toEqual({ symbol: 'BTCUSDT', timeframe: 'D', limit: 3 });
  });

  it('should preserve unique requests', () => {
    const input = [
      { symbol: 'BTCUSDT', timeframe: 'D', limit: 3 },
      { symbol: 'BTCUSDT', timeframe: '1h', limit: 240 },
      { symbol: 'ETHUSDT', timeframe: 'D', limit: 3 },
    ];

    const result = runner.deduplicatePrefetchData(input);

    expect(result).toHaveLength(3);
  });

  it('should deduplicate when only symbol differs', () => {
    const input = [
      { symbol: 'BTCUSDT', timeframe: 'D', limit: 3 },
      { symbol: 'ETHUSDT', timeframe: 'D', limit: 3 },
      { symbol: 'BTCUSDT', timeframe: 'D', limit: 3 },
    ];

    const result = runner.deduplicatePrefetchData(input);

    expect(result).toHaveLength(2);
  });

  it('should deduplicate when only timeframe differs', () => {
    const input = [
      { symbol: 'BTCUSDT', timeframe: 'D', limit: 3 },
      { symbol: 'BTCUSDT', timeframe: '1h', limit: 3 },
      { symbol: 'BTCUSDT', timeframe: 'D', limit: 3 },
    ];

    const result = runner.deduplicatePrefetchData(input);

    expect(result).toHaveLength(2);
  });

  it('should deduplicate when only limit differs', () => {
    const input = [
      { symbol: 'BTCUSDT', timeframe: 'D', limit: 3 },
      { symbol: 'BTCUSDT', timeframe: 'D', limit: 5 },
      { symbol: 'BTCUSDT', timeframe: 'D', limit: 3 },
    ];

    const result = runner.deduplicatePrefetchData(input);

    expect(result).toHaveLength(2);
  });

  it('should handle empty array', () => {
    const result = runner.deduplicatePrefetchData([]);
    expect(result).toHaveLength(0);
  });

  it('should keep first occurrence when duplicates exist', () => {
    const input = [
      { symbol: 'BTCUSDT', timeframe: 'D', limit: 3, extra: 'first' },
      { symbol: 'BTCUSDT', timeframe: 'D', limit: 3, extra: 'second' },
    ];

    const result = runner.deduplicatePrefetchData(input);

    expect(result).toHaveLength(1);
    expect(result[0].extra).toBe('first');
  });
});
