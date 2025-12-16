import { describe, test, expect, beforeAll, vi } from 'vitest';
import { SeriesDataMapper } from '../../out/js/SeriesDataMapper.js';

describe('Support/Resistance Level Gaps - SeriesDataMapper Integration', () => {
  let mapper;

  beforeAll(() => {
    mapper = new SeriesDataMapper();
  });

  test('should preserve null colors marking level transitions', () => {
    const plotDataWithGaps = [
      { time: 1000, value: 100, options: { color: '#FF0000', linewidth: 1 } },
      { time: 2000, value: 100, options: { color: '#FF0000', linewidth: 1 } },
      { time: 3000, value: 100, options: { color: '#FF0000', linewidth: 1 } },
      { time: 4000, value: 105, options: { color: null, linewidth: 1 } },
      { time: 5000, value: 105, options: { color: '#FF0000', linewidth: 1 } },
      { time: 6000, value: 105, options: { color: '#FF0000', linewidth: 1 } },
      { time: 7000, value: 110, options: { color: null, linewidth: 1 } },
      { time: 8000, value: 110, options: { color: '#FF0000', linewidth: 1 } },
    ];

    const result = mapper.applyColorToData(plotDataWithGaps, '#00FF00');

    expect(result[0].options.color).toBe('#FF0000');
    expect(result[1].options.color).toBe('#FF0000');
    expect(result[2].options.color).toBe('#FF0000');
    expect(result[3].options.color).toBeNull();
    expect(result[4].options.color).toBe('#FF0000');
    expect(result[5].options.color).toBe('#FF0000');
    expect(result[6].options.color).toBeNull();
    expect(result[7].options.color).toBe('#FF0000');
  });

  test('should apply series color to points without explicit color', () => {
    const plotDataMixedColors = [
      { time: 1000, value: 100, options: {} },
      { time: 2000, value: 100, options: { color: '#FF0000' } },
      { time: 3000, value: 105, options: { color: null } },
      { time: 4000, value: 105, options: {} },
    ];

    const result = mapper.applyColorToData(plotDataMixedColors, '#0000FF');

    expect(result[0].options.color).toBe('#0000FF');
    expect(result[1].options.color).toBe('#FF0000');
    expect(result[2].options.color).toBeNull();
    expect(result[3].options.color).toBe('#0000FF');
  });

  test('should handle support/resistance pattern from PineScript change() conditional', () => {
    const resistanceLevelData = [
      { time: 1000, value: 90160.27, options: { color: '#FF0000', title: 'Resistance' } },
      { time: 2000, value: 90160.27, options: { color: '#FF0000', title: 'Resistance' } },
      { time: 3000, value: 92265.69, options: { color: null, title: 'Resistance' } },
      { time: 4000, value: 92265.69, options: { color: '#FF0000', title: 'Resistance' } },
      { time: 5000, value: 92265.69, options: { color: '#FF0000', title: 'Resistance' } },
    ];

    const result = mapper.applyColorToData(resistanceLevelData, '#FF0000');

    expect(result[0].options.color).toBe('#FF0000');
    expect(result[1].options.color).toBe('#FF0000');
    expect(result[2].options.color).toBeNull();
    expect(result[3].options.color).toBe('#FF0000');
    expect(result[4].options.color).toBe('#FF0000');

    const gapIndices = result
      .map((point, i) => (point.options.color === null ? i : -1))
      .filter((i) => i >= 0);

    expect(gapIndices).toEqual([2]);
  });
});
