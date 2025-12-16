import { describe, test, expect } from 'vitest';
import { SeriesDataMapper } from '../../out/js/SeriesDataMapper.js';

describe('SeriesDataMapper', () => {
  const mapper = new SeriesDataMapper();

  describe('applyColorToData', () => {
    test('should add color to all data points', () => {
      const data = [
        { time: 1000, value: 100 },
        { time: 2000, value: 101 },
        { time: 3000, value: 102 },
      ];

      const result = mapper.applyColorToData(data, '#FF0000');

      expect(result).toEqual([
        { time: 1000, value: 100, options: { color: '#FF0000' } },
        { time: 2000, value: 101, options: { color: '#FF0000' } },
        { time: 3000, value: 102, options: { color: '#FF0000' } },
      ]);
    });

    test('should handle single data point', () => {
      const data = [{ time: 1000, value: 100 }];

      const result = mapper.applyColorToData(data, '#00FF00');

      expect(result).toEqual([{ time: 1000, value: 100, options: { color: '#00FF00' } }]);
    });

    test('should preserve existing point properties', () => {
      const data = [{ time: 1000, value: 100, custom: 'property', metadata: { key: 'value' } }];

      const result = mapper.applyColorToData(data, '#00FF00');

      expect(result[0]).toEqual({
        time: 1000,
        value: 100,
        custom: 'property',
        metadata: { key: 'value' },
        options: { color: '#00FF00' },
      });
    });

    test('should overwrite existing options.color', () => {
      const data = [{ time: 1000, value: 100, options: { color: '#FF0000', style: 'line', width: 2 } }];

      const result = mapper.applyColorToData(data, '#00FF00');

      expect(result[0].options).toEqual({ color: '#FF0000', style: 'line', width: 2 });
    });

    test('should preserve existing non-null color', () => {
      const data = [{ time: 1000, value: 100, options: { color: '#FF0000', style: 'line', width: 2 } }];

      const result = mapper.applyColorToData(data, '#00FF00');

      expect(result[0].options.color).toBe('#FF0000');
    });

    test('should preserve null color as gap marker', () => {
      const data = [
        { time: 1000, value: 100, options: { color: '#FF0000' } },
        { time: 2000, value: 100, options: { color: null } },
        { time: 3000, value: 100, options: { color: '#FF0000' } },
      ];

      const result = mapper.applyColorToData(data, '#00FF00');

      expect(result[0].options.color).toBe('#FF0000');
      expect(result[1].options.color).toBeNull();
      expect(result[2].options.color).toBe('#FF0000');
    });

    test('should create options object when not present', () => {
      const data = [{ time: 1000, value: 100 }];

      const result = mapper.applyColorToData(data, '#0000FF');

      expect(result[0]).toHaveProperty('options');
      expect(result[0].options).toEqual({ color: '#0000FF' });
    });
  });

  describe('edge cases', () => {
    test('should handle empty data array', () => {
      const result = mapper.applyColorToData([], '#FF0000');

      expect(result).toEqual([]);
    });

    test('should handle null color', () => {
      const data = [{ time: 1000, value: 100 }];

      const result = mapper.applyColorToData(data, null);

      expect(result[0].options.color).toBeNull();
    });

    test('should apply series color when point has no color', () => {
      const data = [{ time: 1000, value: 100, options: {} }];

      const result = mapper.applyColorToData(data, '#FF0000');

      expect(result[0].options.color).toBe('#FF0000');
    });

    test('should handle undefined color', () => {
      const data = [{ time: 1000, value: 100 }];

      const result = mapper.applyColorToData(data, undefined);

      expect(result[0].options.color).toBeUndefined();
    });

    test('should handle empty string color', () => {
      const data = [{ time: 1000, value: 100 }];

      const result = mapper.applyColorToData(data, '');

      expect(result[0].options.color).toBe('');
    });

    test('should handle various color formats', () => {
      const data = [{ time: 1000, value: 100 }];

      const hexResult = mapper.applyColorToData(data, '#FF0000');
      expect(hexResult[0].options.color).toBe('#FF0000');

      const rgbResult = mapper.applyColorToData(data, 'rgb(255, 0, 0)');
      expect(rgbResult[0].options.color).toBe('rgb(255, 0, 0)');

      const namedResult = mapper.applyColorToData(data, 'red');
      expect(namedResult[0].options.color).toBe('red');
    });

    test('should handle large datasets', () => {
      const largeData = Array.from({ length: 5000 }, (_, i) => ({
        time: 1000 + i * 1000,
        value: 100 + i,
      }));

      const result = mapper.applyColorToData(largeData, '#FF0000');

      expect(result).toHaveLength(5000);
      expect(result[0].options.color).toBe('#FF0000');
      expect(result[4999].options.color).toBe('#FF0000');
    });

    test('should not mutate original data array', () => {
      const data = [{ time: 1000, value: 100 }];
      const originalData = JSON.parse(JSON.stringify(data));

      mapper.applyColorToData(data, '#FF0000');

      expect(data).toEqual(originalData);
    });

    test('should preserve options object properties', () => {
      const data = [
        {
          time: 1000,
          value: 100,
          options: {
            lineStyle: 'solid',
            lineWidth: 3,
            visible: true,
          },
        },
      ];

      const result = mapper.applyColorToData(data, '#0000FF');

      expect(result[0].options).toEqual({
        lineStyle: 'solid',
        lineWidth: 3,
        visible: true,
        color: '#0000FF',
      });
    });
  });
});
