import { describe, test, expect, vi } from 'vitest';
import { createPlotAdapter } from '../../src/adapters/PinePlotAdapter.js';

describe('PinePlotAdapter', () => {
  describe('createPlotAdapter', () => {
    test('should pass string title directly to corePlot', () => {
      const corePlot = vi.fn();
      const plot = createPlotAdapter(corePlot);
      const series = [1, 2, 3];
      const options = { color: 'blue', linewidth: 2 };

      plot(series, 'My Title', options);

      expect(corePlot).toHaveBeenCalledWith(series, 'My Title', options);
      expect(corePlot).toHaveBeenCalledTimes(1);
    });

    test('should extract title from options object', () => {
      const corePlot = vi.fn();
      const plot = createPlotAdapter(corePlot);
      const series = [4, 5, 6];
      const options = { title: 'EMA 20', color: 'red', linewidth: 1 };

      plot(series, options);

      expect(corePlot).toHaveBeenCalledWith(series, 'EMA 20', {
        color: 'red',
        linewidth: 1,
      });
    });

    test('should handle options array (PyneScript array wrapper)', () => {
      const corePlot = vi.fn();
      const plot = createPlotAdapter(corePlot);
      const series = [7, 8, 9];
      const options = [{ title: 'Signal', color: 'green', style: 'line' }];

      plot(series, options);

      expect(corePlot).toHaveBeenCalledWith(series, 'Signal', {
        color: 'green',
        style: 'line',
      });
    });

    test('should handle missing options gracefully', () => {
      const corePlot = vi.fn();
      const plot = createPlotAdapter(corePlot);
      const series = [10, 11, 12];

      plot(series, null);

      expect(corePlot).toHaveBeenCalledWith(series, undefined, {});
    });

    test('should default to empty options when maybeOptions not provided', () => {
      const corePlot = vi.fn();
      const plot = createPlotAdapter(corePlot);
      const series = [13, 14, 15];

      plot(series, 'Title Only');

      expect(corePlot).toHaveBeenCalledWith(series, 'Title Only', {});
    });

    test('should pass through all plot-related options', () => {
      const corePlot = vi.fn();
      const plot = createPlotAdapter(corePlot);
      const series = [16, 17, 18];
      const options = {
        title: 'Custom',
        color: 'yellow',
        style: 'linebr',
        linewidth: 3,
        unrelatedProp: 'ignored',
      };

      plot(series, options);

      expect(corePlot).toHaveBeenCalledWith(series, 'Custom', {
        color: 'yellow',
        style: 'linebr',
        linewidth: 3,
        unrelatedProp: 'ignored',
      });
    });

    test('should pass transp parameter to corePlot', () => {
      const corePlot = vi.fn();
      const plot = createPlotAdapter(corePlot);
      const series = [19, 20, 21];
      const options = {
        title: 'Transparent Line',
        color: 'blue',
        linewidth: 2,
        transp: 50,
      };

      plot(series, options);

      expect(corePlot).toHaveBeenCalledWith(series, 'Transparent Line', {
        color: 'blue',
        linewidth: 2,
        transp: 50,
      });
    });

    test('should pass histbase and offset parameters to corePlot', () => {
      const corePlot = vi.fn();
      const plot = createPlotAdapter(corePlot);
      const series = [22, 23, 24];
      const options = {
        title: 'Histogram',
        color: 'green',
        style: 'histogram',
        histbase: 0,
        offset: 1,
      };

      plot(series, options);

      expect(corePlot).toHaveBeenCalledWith(series, 'Histogram', {
        color: 'green',
        style: 'histogram',
        histbase: 0,
        offset: 1,
      });
    });
  });
});
