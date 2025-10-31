import { describe, test, expect } from 'vitest';
import { adaptLineSeriesData } from '../../out/lineSeriesAdapter.js';

describe('lineSeriesAdapter', () => {
  describe('adaptLineSeriesData', () => {
    test('should filter out null values', () => {
      const input = [
        { time: 1000, value: 10, options: { color: 'blue' } },
        { time: 2000, value: null, options: { color: 'blue' } },
        { time: 3000, value: 30, options: { color: 'blue' } },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(3);
      expect(result[0]).toEqual({ time: 1, value: 10, color: 'transparent' });
      expect(result[1]).toEqual({ time: 2, value: NaN, color: 'transparent' });
      expect(result[2]).toEqual({ time: 3, value: 30 });
    });

    test('should filter out undefined values', () => {
      const input = [
        { time: 1000, value: 10, options: { color: 'blue' } },
        { time: 2000, value: undefined, options: { color: 'blue' } },
        { time: 3000, value: 30, options: { color: 'blue' } },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(3);
      expect(result[0]).toEqual({ time: 1, value: 10, color: 'transparent' });
      expect(result[1]).toEqual({ time: 2, value: NaN, color: 'transparent' });
      expect(result[2]).toEqual({ time: 3, value: 30 });
    });

    test('should filter out NaN values', () => {
      const input = [
        { time: 1000, value: 10, options: { color: 'blue' } },
        { time: 2000, value: NaN, options: { color: 'blue' } },
        { time: 3000, value: 30, options: { color: 'blue' } },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(3);
      expect(result[0]).toEqual({ time: 1, value: 10, color: 'transparent' });
      expect(result[1]).toEqual({ time: 2, value: NaN, color: 'transparent' });
      expect(result[2]).toEqual({ time: 3, value: 30 });
    });

    test('should filter out NaN values', () => {
      const input = [
        { time: 1000, value: 10, options: { color: 'blue' } },
        { time: 2000, value: NaN, options: { color: 'blue' } },
        { time: 3000, value: 30, options: { color: 'blue' } },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(3);
      expect(result[0]).toEqual({ time: 1, value: 10, color: 'transparent' });
      expect(result[1]).toEqual({ time: 2, value: NaN, color: 'transparent' });
      expect(result[2]).toEqual({ time: 3, value: 30 });
    });

    test('should mark last point before gap as transparent', () => {
    });

    test('should mark last point before gap as transparent', () => {
      const input = [
        { time: 1000, value: 10, options: { color: 'blue' } },
        { time: 2000, value: 20, options: { color: 'blue' } },
        { time: 3000, value: null, options: { color: 'blue' } },
        { time: 4000, value: 40, options: { color: 'blue' } },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(4);
      expect(result[0]).toEqual({ time: 1, value: 10 });
      expect(result[1]).toEqual({ time: 2, value: 20, color: 'transparent' });
      expect(result[2]).toEqual({ time: 3, value: NaN, color: 'transparent' });
      expect(result[3]).toEqual({ time: 4, value: 40 });
    });

    test('should not mark last point as transparent if followed by valid value', () => {
      const input = [
        { time: 1000, value: 10, options: { color: 'blue' } },
        { time: 2000, value: 20, options: { color: 'blue' } },
        { time: 3000, value: 30, options: { color: 'blue' } },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(3);
      expect(result[0]).toEqual({ time: 1, value: 10 });
      expect(result[1]).toEqual({ time: 2, value: 20 });
      expect(result[2]).toEqual({ time: 3, value: 30 });
    });

    test('should filter out NaN values', () => {
      const input = [
        { time: 1000, value: 10, options: { color: 'blue' } },
        { time: 2000, value: NaN, options: { color: 'blue' } },
        { time: 3000, value: 30, options: { color: 'blue' } },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(3);
      expect(result[0]).toEqual({ time: 1, value: 10, color: 'transparent' });
      expect(result[1]).toEqual({ time: 2, value: NaN, color: 'transparent' });
      expect(result[2]).toEqual({ time: 3, value: 30 });
    });

    test('should mark last point before gap as transparent', () => {
      const input = [
        { time: 1000, value: 10, options: { color: 'blue' } },
        { time: 2000, value: 20, options: { color: 'blue' } },
        { time: 3000, value: null, options: { color: 'blue' } },
        { time: 4000, value: 40, options: { color: 'blue' } },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(4);
      expect(result[0]).toEqual({ time: 1, value: 10 });
      expect(result[1]).toEqual({ time: 2, value: 20, color: 'transparent' });
      expect(result[2]).toEqual({ time: 3, value: NaN, color: 'transparent' });
      expect(result[3]).toEqual({ time: 4, value: 40 });
    });

    test('should not mark last point as transparent if followed by valid value', () => {
      const input = [
        { time: 1000, value: 10, options: { color: 'blue' } },
        { time: 2000, value: 20, options: { color: 'blue' } },
        { time: 3000, value: 30, options: { color: 'blue' } },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(3);
      expect(result[0]).toEqual({ time: 1, value: 10 });
      expect(result[1]).toEqual({ time: 2, value: 20 });
      expect(result[2]).toEqual({ time: 3, value: 30 });
    });

    test('should handle multiple consecutive gaps', () => {
      const input = [
        { time: 1000, value: 10, options: { color: 'blue', options: { color: 'blue' } } },
        { time: 2000, value: null, options: { color: 'blue', options: { color: 'blue' } } },
        { time: 3000, value: null, options: { color: 'blue', options: { color: 'blue' } } },
        { time: 4000, value: 40, options: { color: 'blue', options: { color: 'blue' } } },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(3);
      expect(result[0]).toEqual({ time: 1, value: 10, color: 'transparent' });
      expect(result[1]).toEqual({ time: 2, value: NaN, color: 'transparent' });
      expect(result[2]).toEqual({ time: 4, value: 40 });
    });

    test('should handle multiple gaps with transitions', () => {
      const input = [
        { time: 1000, value: 10, options: { color: 'blue', options: { color: 'blue' } } },
        { time: 2000, value: 20, options: { color: 'blue', options: { color: 'blue' } } },
        { time: 3000, value: null, options: { color: 'blue', options: { color: 'blue' } } },
        { time: 4000, value: 40, options: { color: 'blue', options: { color: 'blue' } } },
        { time: 5000, value: 50, options: { color: 'blue', options: { color: 'blue' } } },
        { time: 6000, value: null, options: { color: 'blue', options: { color: 'blue' } } },
        { time: 7000, value: 70, options: { color: 'blue', options: { color: 'blue' } } },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(7);
      expect(result[0]).toEqual({ time: 1, value: 10 });
      expect(result[1]).toEqual({ time: 2, value: 20, color: 'transparent' });
      expect(result[2]).toEqual({ time: 3, value: NaN, color: 'transparent' });
      expect(result[3]).toEqual({ time: 4, value: 40 });
      expect(result[4]).toEqual({ time: 5, value: 50, color: 'transparent' });
      expect(result[5]).toEqual({ time: 6, value: NaN, color: 'transparent' });
      expect(result[6]).toEqual({ time: 7, value: 70 });
    });

    test('should convert millisecond timestamps to seconds', () => {
      const input = [{ time: 1609459200000, value: 100, options: { color: 'blue', options: { color: 'blue' } } }];

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
        { time: 1000, value: null, options: { color: 'blue', options: { color: 'blue' } } },
        { time: 2000, value: null, options: { color: 'blue', options: { color: 'blue' } } },
        { time: 3000, value: null, options: { color: 'blue', options: { color: 'blue' } } },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toEqual([]);
    });

    test('should handle single valid value', () => {
      const input = [{ time: 1000, value: 42, options: { color: 'blue', options: { color: 'blue' } } }];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(1);
      expect(result[0]).toEqual({ time: 1, value: 42 });
    });

    test('should handle gap at the beginning', () => {
      const input = [
        { time: 1000, value: null, options: { color: 'blue', options: { color: 'blue' } } },
        { time: 2000, value: 20, options: { color: 'blue', options: { color: 'blue' } } },
        { time: 3000, value: 30, options: { color: 'blue', options: { color: 'blue' } } },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(3);
      expect(result[0]).toEqual({ time: 1, value: NaN, color: 'transparent' });
      expect(result[1]).toEqual({ time: 2, value: 20 });
      expect(result[2]).toEqual({ time: 3, value: 30 });
    });

    test('should handle gap at the end', () => {
      const input = [
        { time: 1000, value: 10, options: { color: 'blue', options: { color: 'blue' } } },
        { time: 2000, value: 20, options: { color: 'blue', options: { color: 'blue' } } },
        { time: 3000, value: null, options: { color: 'blue', options: { color: 'blue' } } },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(3);
      expect(result[0]).toEqual({ time: 1, value: 10 });
      expect(result[1]).toEqual({ time: 2, value: 20, color: 'transparent' });
      expect(result[2]).toEqual({ time: 3, value: NaN, color: 'transparent' });
    });

    test('should preserve zero values as valid data', () => {
      const input = [
        { time: 1000, value: 10, options: { color: 'blue', options: { color: 'blue' } } },
        { time: 2000, value: 0, options: { color: 'blue', options: { color: 'blue' } } },
        { time: 3000, value: 10, options: { color: 'blue', options: { color: 'blue' } } },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(3);
      expect(result[0]).toEqual({ time: 1, value: 10 });
      expect(result[1]).toEqual({ time: 2, value: 0 });
      expect(result[2]).toEqual({ time: 3, value: 10 });
    });

    test('should preserve negative values as valid data', () => {
      const input = [
        { time: 1000, value: 10, options: { color: 'blue', options: { color: 'blue' } } },
        { time: 2000, value: -5, options: { color: 'blue', options: { color: 'blue' } } },
        { time: 3000, value: 10, options: { color: 'blue', options: { color: 'blue' } } },
      ];

      const result = adaptLineSeriesData(input);

      expect(result).toHaveLength(3);
      expect(result[0]).toEqual({ time: 1, value: 10 });
      expect(result[1]).toEqual({ time: 2, value: -5 });
      expect(result[2]).toEqual({ time: 3, value: 10 });
    });
  });
});
