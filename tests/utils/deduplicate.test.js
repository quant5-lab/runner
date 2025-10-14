import { describe, it, expect } from 'vitest';
import { deduplicate } from '../../src/utils/deduplicate.js';

describe('deduplicate', () => {
  it('should remove duplicate objects by key', () => {
    const items = [
      { id: 1, name: 'Alice' },
      { id: 2, name: 'Bob' },
      { id: 1, name: 'Alice Duplicate' },
    ];

    const result = deduplicate(items, item => item.id);

    expect(result).toHaveLength(2);
    expect(result[0]).toEqual({ id: 1, name: 'Alice' });
    expect(result[1]).toEqual({ id: 2, name: 'Bob' });
  });

  it('should handle composite keys', () => {
    const items = [
      { symbol: 'BTC', timeframe: '1h', limit: 100 },
      { symbol: 'BTC', timeframe: '1h', limit: 100 },
      { symbol: 'BTC', timeframe: '1d', limit: 100 },
      { symbol: 'ETH', timeframe: '1h', limit: 100 },
    ];

    const result = deduplicate(items, item => `${item.symbol}:${item.timeframe}:${item.limit}`);

    expect(result).toHaveLength(3);
    expect(result.map(r => `${r.symbol}:${r.timeframe}`)).toEqual(['BTC:1h', 'BTC:1d', 'ETH:1h']);
  });

  it('should keep first occurrence when duplicates exist', () => {
    const items = [
      { id: 1, value: 'first' },
      { id: 1, value: 'second' },
      { id: 1, value: 'third' },
    ];

    const result = deduplicate(items, item => item.id);

    expect(result).toHaveLength(1);
    expect(result[0].value).toBe('first');
  });

  it('should handle empty array', () => {
    const result = deduplicate([], item => item.id);

    expect(result).toEqual([]);
  });

  it('should handle array with no duplicates', () => {
    const items = [
      { id: 1 },
      { id: 2 },
      { id: 3 },
    ];

    const result = deduplicate(items, item => item.id);

    expect(result).toHaveLength(3);
    expect(result).toEqual(items);
  });

  it('should handle complex key getters', () => {
    const items = [
      { user: { id: 1 }, action: 'login' },
      { user: { id: 1 }, action: 'logout' },
      { user: { id: 2 }, action: 'login' },
    ];

    const result = deduplicate(items, item => `${item.user.id}:${item.action}`);

    expect(result).toHaveLength(3);
  });

  it('should handle primitive values', () => {
    const items = [1, 2, 3, 2, 1, 4];

    const result = deduplicate(items, item => item);

    expect(result).toEqual([1, 2, 3, 4]);
  });

  it('should handle string arrays', () => {
    const items = ['apple', 'banana', 'apple', 'cherry', 'banana'];

    const result = deduplicate(items, item => item);

    expect(result).toEqual(['apple', 'banana', 'cherry']);
  });

  it('should preserve object references', () => {
    const obj1 = { id: 1, name: 'Alice' };
    const obj2 = { id: 2, name: 'Bob' };
    const obj3 = { id: 1, name: 'Alice Duplicate' };
    const items = [obj1, obj2, obj3];

    const result = deduplicate(items, item => item.id);

    expect(result[0]).toBe(obj1);
    expect(result[1]).toBe(obj2);
  });

  it('should handle null/undefined keys gracefully', () => {
    const items = [
      { id: null, name: 'A' },
      { id: null, name: 'B' },
      { id: undefined, name: 'C' },
      { id: 1, name: 'D' },
    ];

    const result = deduplicate(items, item => item.id);

    expect(result).toHaveLength(3);
    expect(result.map(r => r.name)).toEqual(['A', 'C', 'D']);
  });
});
