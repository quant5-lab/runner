import { describe, test, expect } from 'vitest';
import { adaptLineSeriesData } from '../../out/lineSeriesAdapter.js';

describe('lineSeriesAdapter', () => {
  describe('adaptLineSeriesData', () => {
    test('should filter out null values', () => {
      const input = [
        { time: 1000, value: 10 },
        { time: 2000, value: null },
        { time: 3000, value: 30 },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(2);
      expect(result[0]).toEqual({ time: 1, value: 10, color: 'transparent' });
      expect(result[1]).toEqual({ time: 3, value: 30 });
    });

    test('should filter out undefined values', () => {
      const input = [
        { time: 1000, value: 10 },
        { time: 2000, value: undefined },
        { time: 3000, value: 30 },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(2);
      expect(result[0].value).toBe(10);
      expect(result[1].value).toBe(30);
    });

    test('should filter out NaN values', () => {
      const input = [
        { time: 1000, value: 10 },
        { time: 2000, value: NaN },
        { time: 3000, value: 30 },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(2);
      expect(result[0].value).toBe(10);
      expect(result[1].value).toBe(30);
    });

    test('should mark last point before gap as transparent', () => {
      const input = [
        { time: 1000, value: 10 },
        { time: 2000, value: 20 },
        { time: 3000, value: null },
        { time: 4000, value: 40 },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(3);
      expect(result[0]).toEqual({ time: 1, value: 10 });
      expect(result[1]).toEqual({ time: 2, value: 20, color: 'transparent' });
      expect(result[2]).toEqual({ time: 4, value: 40 });
    });

    test('should not mark last point as transparent if followed by valid value', () => {
      const input = [
        { time: 1000, value: 10 },
        { time: 2000, value: 20 },
        { time: 3000, value: 30 },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(3);
      expect(result[0]).toEqual({ time: 1, value: 10 });
      expect(result[1]).toEqual({ time: 2, value: 20 });
      expect(result[2]).toEqual({ time: 3, value: 30 });
    });

    test('should handle multiple consecutive gaps', () => {
      const input = [
        { time: 1000, value: 10 },
        { time: 2000, value: null },
        { time: 3000, value: null },
        { time: 4000, value: 40 },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(2);
      expect(result[0]).toEqual({ time: 1, value: 10, color: 'transparent' });
      expect(result[1]).toEqual({ time: 4, value: 40 });
    });

    test('should handle multiple gaps with transitions', () => {
      const input = [
        { time: 1000, value: 10 },
        { time: 2000, value: 20 },
        { time: 3000, value: null },
        { time: 4000, value: 40 },
        { time: 5000, value: 50 },
        { time: 6000, value: null },
        { time: 7000, value: 70 },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(5);
      expect(result[0]).toEqual({ time: 1, value: 10 });
      expect(result[1]).toEqual({ time: 2, value: 20, color: 'transparent' });
      expect(result[2]).toEqual({ time: 4, value: 40 });
      expect(result[3]).toEqual({ time: 5, value: 50, color: 'transparent' });
      expect(result[4]).toEqual({ time: 7, value: 70 });
    });

    test('should convert millisecond timestamps to seconds', () => {
      const input = [{ time: 1609459200000, value: 100 }];

      const result = adaptLineSeriesData(input);

      expect(result[0].time).toBe(1609459200);
    });

    test('should handle empty array', () => {
      const result = adaptLineSeriesData([]);

      expect(result).toEqual([]);
    });

    test('should handle non-array input', () => {
      expect(adaptLineSeriesData(null)).toEqual([]);
      expect(adaptLineSeriesData(undefined)).toEqual([]);
      expect(adaptLineSeriesData('invalid')).toEqual([]);
      expect(adaptLineSeriesData(123)).toEqual([]);
    });

    test('should handle all null values', () => {
      const input = [
        { time: 1000, value: null },
        { time: 2000, value: null },
        { time: 3000, value: null },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toEqual([]);
    });

    test('should handle single valid value', () => {
      const input = [{ time: 1000, value: 42 }];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(1);
      expect(result[0]).toEqual({ time: 1, value: 42 });
    });

    test('should handle gap at the beginning', () => {
      const input = [
        { time: 1000, value: null },
        { time: 2000, value: 20 },
        { time: 3000, value: 30 },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(2);
      expect(result[0]).toEqual({ time: 2, value: 20 });
      expect(result[1]).toEqual({ time: 3, value: 30 });
    });

    test('should handle gap at the end', () => {
      const input = [
        { time: 1000, value: 10 },
        { time: 2000, value: 20 },
        { time: 3000, value: null },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(2);
      expect(result[0]).toEqual({ time: 1, value: 10 });
      expect(result[1]).toEqual({ time: 2, value: 20, color: 'transparent' });
    });

    test('should preserve zero values as valid data', () => {
      const input = [
        { time: 1000, value: 10 },
        { time: 2000, value: 0 },
        { time: 3000, value: 10 },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(3);
      expect(result[0]).toEqual({ time: 1, value: 10 });
      expect(result[1]).toEqual({ time: 2, value: 0 });
      expect(result[2]).toEqual({ time: 3, value: 10 });
    });

    test('should preserve negative values as valid data', () => {
      const input = [
        { time: 1000, value: 10 },
        { time: 2000, value: -5 },
        { time: 3000, value: 10 },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(3);
      expect(result[0]).toEqual({ time: 1, value: 10 });
      expect(result[1]).toEqual({ time: 2, value: -5 });
      expect(result[2]).toEqual({ time: 3, value: 10 });
    });
  });
});
