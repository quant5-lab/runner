import { describe, test, expect } from 'vitest';
import { PlotOffsetTransformer } from '../../out/js/PlotOffsetTransformer.js';
import { TimeIndexBuilder } from '../../out/js/TimeIndexBuilder.js';

describe('PlotOffsetTransformer', () => {
  const timeIndexBuilder = new TimeIndexBuilder();
  const transformer = new PlotOffsetTransformer(timeIndexBuilder);

  const candlestickData = [
    { time: 1000, close: 100 },
    { time: 2000, close: 101 },
    { time: 3000, close: 102 },
    { time: 4000, close: 103 },
    { time: 5000, close: 104 },
  ];

  describe('negative offset (shift left)', () => {
    test('should shift all points left by offset amount', () => {
      const indicatorData = [
        { time: 3000, value: 200 },
        { time: 4000, value: 201 },
        { time: 5000, value: 202 },
      ];

      const result = transformer.transform(indicatorData, -2, candlestickData);

      expect(result).toEqual([
        { time: 1000, value: 200 },
        { time: 2000, value: 201 },
        { time: 3000, value: 202 },
      ]);
    });

    test('should filter points that shift beyond first bar', () => {
      const indicatorData = [
        { time: 1000, value: 200 },
        { time: 2000, value: 201 },
        { time: 3000, value: 202 },
      ];

      const result = transformer.transform(indicatorData, -2, candlestickData);

      expect(result).toEqual([{ time: 1000, value: 202 }]);
    });

    test('should filter all points when offset exceeds available bars', () => {
      const indicatorData = [
        { time: 1000, value: 200 },
        { time: 2000, value: 201 },
      ];

      const result = transformer.transform(indicatorData, -5, candlestickData);

      expect(result).toEqual([]);
    });

    test('should handle single point with large negative offset', () => {
      const indicatorData = [{ time: 5000, value: 200 }];

      const result = transformer.transform(indicatorData, -10, candlestickData);

      expect(result).toEqual([]);
    });
  });

  describe('positive offset (shift right)', () => {
    test('should shift all points right by offset amount', () => {
      const indicatorData = [
        { time: 1000, value: 200 },
        { time: 2000, value: 201 },
      ];

      const result = transformer.transform(indicatorData, 2, candlestickData);

      expect(result).toEqual([
        { time: 3000, value: 200 },
        { time: 4000, value: 201 },
      ]);
    });

    test('should filter points that shift beyond last bar', () => {
      const indicatorData = [
        { time: 3000, value: 200 },
        { time: 4000, value: 201 },
        { time: 5000, value: 202 },
      ];

      const result = transformer.transform(indicatorData, 2, candlestickData);

      expect(result).toEqual([{ time: 5000, value: 200 }]);
    });

    test('should filter all points when offset exceeds available bars', () => {
      const indicatorData = [
        { time: 1000, value: 200 },
        { time: 2000, value: 201 },
      ];

      const result = transformer.transform(indicatorData, 10, candlestickData);

      expect(result).toEqual([]);
    });

    test('should handle single point with large positive offset', () => {
      const indicatorData = [{ time: 1000, value: 200 }];

      const result = transformer.transform(indicatorData, 5, candlestickData);

      expect(result).toEqual([]);
    });
  });

  describe('zero and no-op cases', () => {
    test('should return original data when offset is zero', () => {
      const indicatorData = [
        { time: 2000, value: 200 },
        { time: 3000, value: 201 },
      ];

      const result = transformer.transform(indicatorData, 0, candlestickData);

      expect(result).toBe(indicatorData);
    });

    test('should return original data when candlestick data is empty', () => {
      const indicatorData = [{ time: 1000, value: 200 }];

      const result = transformer.transform(indicatorData, -2, []);

      expect(result).toBe(indicatorData);
    });

    test('should return original data when candlestick data is null', () => {
      const indicatorData = [{ time: 1000, value: 200 }];

      const result = transformer.transform(indicatorData, -2, null);

      expect(result).toBe(indicatorData);
    });

    test('should return original data when candlestick data is undefined', () => {
      const indicatorData = [{ time: 1000, value: 200 }];

      const result = transformer.transform(indicatorData, -2, undefined);

      expect(result).toBe(indicatorData);
    });
  });

  describe('data integrity', () => {
    test('should preserve all point properties during transformation', () => {
      const indicatorData = [
        { time: 3000, value: 200, options: { color: 'red' }, custom: 'data', nested: { prop: 'value' } },
      ];

      const result = transformer.transform(indicatorData, -2, candlestickData);

      expect(result[0]).toEqual({
        time: 1000,
        value: 200,
        options: { color: 'red' },
        custom: 'data',
        nested: { prop: 'value' },
      });
    });

    test('should filter points with timestamps not in candlestick data', () => {
      const indicatorData = [
        { time: 9999, value: 200 },
        { time: 3000, value: 201 },
        { time: 8888, value: 202 },
      ];

      const result = transformer.transform(indicatorData, -1, candlestickData);

      expect(result).toEqual([{ time: 2000, value: 201 }]);
    });

    test('should handle empty indicator data', () => {
      const result = transformer.transform([], -2, candlestickData);

      expect(result).toEqual([]);
    });

    test('should handle single indicator point', () => {
      const indicatorData = [{ time: 3000, value: 200 }];

      const result = transformer.transform(indicatorData, -1, candlestickData);

      expect(result).toEqual([{ time: 2000, value: 200 }]);
    });
  });

  describe('boundary conditions', () => {
    test('should handle offset equal to negative data length', () => {
      const indicatorData = [
        { time: 5000, value: 200 },
      ];

      const result = transformer.transform(indicatorData, -4, candlestickData);

      expect(result).toEqual([{ time: 1000, value: 200 }]);
    });

    test('should handle offset equal to positive remaining space', () => {
      const indicatorData = [
        { time: 1000, value: 200 },
      ];

      const result = transformer.transform(indicatorData, 4, candlestickData);

      expect(result).toEqual([{ time: 5000, value: 200 }]);
    });

    test('should handle minimum viable dataset (1 candle, 1 point)', () => {
      const minimalCandlesticks = [{ time: 1000, close: 100 }];
      const indicatorData = [{ time: 1000, value: 200 }];

      const resultZero = transformer.transform(indicatorData, 0, minimalCandlesticks);
      const resultNeg = transformer.transform(indicatorData, -1, minimalCandlesticks);
      const resultPos = transformer.transform(indicatorData, 1, minimalCandlesticks);

      expect(resultZero).toBe(indicatorData);
      expect(resultNeg).toEqual([]);
      expect(resultPos).toEqual([]);
    });

    test('should handle very large offset values', () => {
      const indicatorData = [{ time: 3000, value: 200 }];

      const resultLargeNeg = transformer.transform(indicatorData, -1000, candlestickData);
      const resultLargePos = transformer.transform(indicatorData, 1000, candlestickData);

      expect(resultLargeNeg).toEqual([]);
      expect(resultLargePos).toEqual([]);
    });
  });

  describe('real-world scenarios', () => {
    test('should handle typical support/resistance offset pattern', () => {
      const resistanceData = Array.from({ length: 20 }, (_, i) => ({
        time: 1000 + (i + 16) * 1000,
        value: 100 + i,
      }));
      const fullCandlesticks = Array.from({ length: 40 }, (_, i) => ({
        time: 1000 + i * 1000,
        close: 100,
      }));

      const result = transformer.transform(resistanceData, -16, fullCandlesticks);

      expect(result.length).toBe(20);
      expect(result[0].time).toBe(1000);
      expect(result[0].value).toBe(100);
      expect(result[19].time).toBe(20000);
    });

    test('should handle mixed in-bounds and out-of-bounds points', () => {
      const indicatorData = [
        { time: 1000, value: 200 },
        { time: 2000, value: 201 },
        { time: 3000, value: 202 },
        { time: 4000, value: 203 },
        { time: 5000, value: 204 },
      ];

      const result = transformer.transform(indicatorData, -3, candlestickData);

      expect(result).toEqual([
        { time: 1000, value: 203 },
        { time: 2000, value: 204 },
      ]);
    });

    test('should handle sparse indicator data with offset', () => {
      const sparseData = [
        { time: 1000, value: 200 },
        { time: 5000, value: 204 },
      ];

      const result = transformer.transform(sparseData, 2, candlestickData);

      expect(result).toEqual([{ time: 3000, value: 200 }]);
    });
  });
});
