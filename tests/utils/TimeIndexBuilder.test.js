import { describe, test, expect } from 'vitest';
import { TimeIndexBuilder } from '../../out/js/TimeIndexBuilder.js';

describe('TimeIndexBuilder', () => {
  describe('build', () => {
    test('should build correct time-to-index mapping', () => {
      const builder = new TimeIndexBuilder();
      const candlestickData = [
        { time: 1000, open: 100, high: 105, low: 95, close: 102 },
        { time: 2000, open: 102, high: 107, low: 100, close: 105 },
        { time: 3000, open: 105, high: 110, low: 103, close: 108 },
      ];

      const timeIndex = builder.build(candlestickData);

      expect(timeIndex.get(1000)).toBe(0);
      expect(timeIndex.get(2000)).toBe(1);
      expect(timeIndex.get(3000)).toBe(2);
      expect(timeIndex.get(4000)).toBeUndefined();
    });

    test('should handle single candlestick', () => {
      const builder = new TimeIndexBuilder();
      const candlestickData = [{ time: 1000, open: 100, high: 105, low: 95, close: 102 }];

      const timeIndex = builder.build(candlestickData);

      expect(timeIndex.size).toBe(1);
      expect(timeIndex.get(1000)).toBe(0);
    });

    test('should handle large datasets efficiently', () => {
      const builder = new TimeIndexBuilder();
      const largeDataset = Array.from({ length: 5000 }, (_, i) => ({
        time: 1000 + i * 1000,
        close: 100 + i,
      }));

      const timeIndex = builder.build(largeDataset);

      expect(timeIndex.size).toBe(5000);
      expect(timeIndex.get(1000)).toBe(0);
      expect(timeIndex.get(5000000)).toBe(4999);
    });

    test('should handle non-sequential timestamps', () => {
      const builder = new TimeIndexBuilder();
      const candlestickData = [
        { time: 5000, close: 100 },
        { time: 1000, close: 101 },
        { time: 3000, close: 102 },
      ];

      const timeIndex = builder.build(candlestickData);

      expect(timeIndex.get(5000)).toBe(0);
      expect(timeIndex.get(1000)).toBe(1);
      expect(timeIndex.get(3000)).toBe(2);
    });
  });

  describe('edge cases', () => {
    test('should handle empty candlestick data', () => {
      const builder = new TimeIndexBuilder();
      const timeIndex = builder.build([]);

      expect(timeIndex.size).toBe(0);
    });

    test('should handle null input gracefully', () => {
      const builder = new TimeIndexBuilder();
      expect(() => builder.build(null)).toThrow();
    });

    test('should handle undefined input gracefully', () => {
      const builder = new TimeIndexBuilder();
      expect(() => builder.build(undefined)).toThrow();
    });

    test('should handle zero timestamps', () => {
      const builder = new TimeIndexBuilder();
      const candlestickData = [
        { time: 0, close: 100 },
        { time: 1000, close: 101 },
      ];

      const timeIndex = builder.build(candlestickData);

      expect(timeIndex.get(0)).toBe(0);
      expect(timeIndex.get(1000)).toBe(1);
    });

    test('should handle negative timestamps', () => {
      const builder = new TimeIndexBuilder();
      const candlestickData = [
        { time: -1000, close: 100 },
        { time: 0, close: 101 },
        { time: 1000, close: 102 },
      ];

      const timeIndex = builder.build(candlestickData);

      expect(timeIndex.get(-1000)).toBe(0);
      expect(timeIndex.get(0)).toBe(1);
      expect(timeIndex.get(1000)).toBe(2);
    });
  });
});
