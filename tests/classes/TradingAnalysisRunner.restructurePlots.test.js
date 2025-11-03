import { describe, it, expect, beforeEach } from 'vitest';
import { TradingAnalysisRunner } from '../../src/classes/TradingAnalysisRunner.js';
import { Logger } from '../../src/classes/Logger.js';

describe('TradingAnalysisRunner.restructurePlots', () => {
  let runner;

  beforeEach(() => {
    const logger = new Logger(false);
    runner = new TradingAnalysisRunner(null, null, null, null, null, logger);
  });

  describe('Edge Cases', () => {
    it('should return empty object for null input', () => {
      const result = runner.restructurePlots(null);
      expect(result).toEqual({});
    });

    it('should return empty object for undefined input', () => {
      const result = runner.restructurePlots(undefined);
      expect(result).toEqual({});
    });

    it('should return empty object for non-object input', () => {
      const result = runner.restructurePlots('invalid');
      expect(result).toEqual({});
    });

    it('should normalize timestamps for plots with multiple named keys', () => {
      const input = {
        SMA20: { data: [{ time: 1000000, value: 100 }] },
        EMA50: { data: [{ time: 1000000, value: 105 }] },
      };
      const result = runner.restructurePlots(input);
      expect(result).toEqual({
        SMA20: { data: [{ time: 1000, value: 100, options: undefined }] },
        EMA50: { data: [{ time: 1000, value: 105, options: undefined }] },
      });
    });

    it('should normalize timestamps when Plot key does not exist', () => {
      const input = {
        CustomPlot: { data: [{ time: 2000000, value: 100 }] },
      };
      const result = runner.restructurePlots(input);
      expect(result).toEqual({
        CustomPlot: { data: [{ time: 2000, value: 100, options: undefined }] },
      });
    });

    it('should return empty object when Plot.data is not array', () => {
      const input = {
        Plot: { data: 'invalid' },
      };
      const result = runner.restructurePlots(input);
      expect(result).toEqual({});
    });

    it('should return empty object when Plot.data is empty', () => {
      const input = {
        Plot: { data: [] },
      };
      const result = runner.restructurePlots(input);
      expect(result).toEqual({});
    });
  });

  describe('Single Plot per Candle', () => {
    it('should handle single plot with unique timestamps', () => {
      const input = {
        Plot: {
          data: [
            { time: 1000000, value: 100, options: { color: '#FF5252', linewidth: 1 } },
            { time: 2000000, value: 101, options: { color: '#FF5252', linewidth: 1 } },
            { time: 3000000, value: 102, options: { color: '#FF5252', linewidth: 1 } },
          ],
        },
      };

      const result = runner.restructurePlots(input);

      expect(Object.keys(result)).toHaveLength(1);
      expect(result['Red Plot 1']).toBeDefined();
      expect(result['Red Plot 1'].data).toHaveLength(3);
      expect(result['Red Plot 1'].data[0].time).toBe(1000);
      expect(result['Red Plot 1'].data[1].time).toBe(2000);
      expect(result['Red Plot 1'].data[2].time).toBe(3000);
    });
  });

  describe('Multiple Plots per Candle', () => {
    it('should separate 2 plots with different colors', () => {
      const input = {
        Plot: {
          data: [
            { time: 1000000, value: 100, options: { color: '#FF5252', linewidth: 1 } },
            { time: 1000000, value: 200, options: { color: '#00E676', linewidth: 1 } },
            { time: 2000000, value: 101, options: { color: '#FF5252', linewidth: 1 } },
            { time: 2000000, value: 201, options: { color: '#00E676', linewidth: 1 } },
          ],
        },
      };

      const result = runner.restructurePlots(input);

      expect(Object.keys(result)).toHaveLength(2);
      expect(result['Red Plot 1']).toBeDefined();
      expect(result['Lime Plot 2']).toBeDefined();
      expect(result['Red Plot 1'].data).toHaveLength(2);
      expect(result['Lime Plot 2'].data).toHaveLength(2);
    });

    it('should separate 7 plots matching BB strategy pattern', () => {
      const input = {
        Plot: {
          data: [
            /* Timestamp 1000 */
            { time: 1000000, value: 100, options: { linewidth: 1, color: '#FF5252', transp: 0 } },
            { time: 1000000, value: 101, options: { linewidth: 1, color: '#363A45', transp: 0 } },
            { time: 1000000, value: 102, options: { linewidth: 1, color: '#00E676', transp: 0 } },
            { time: 1000000, value: null, options: { color: '#787B86', style: 'linebr' } },
            { time: 1000000, value: null, options: { color: '#787B86', style: 'linebr' } },
            { time: 1000000, value: 150, options: { color: '#FFFFFF', style: 'linebr', linewidth: 2 } },
            { time: 1000000, value: 90, options: { color: '#FFFFFF', style: 'linebr', linewidth: 2 } },
            /* Timestamp 2000 */
            { time: 2000000, value: 105, options: { linewidth: 1, color: '#FF5252', transp: 0 } },
            { time: 2000000, value: 106, options: { linewidth: 1, color: '#363A45', transp: 0 } },
            { time: 2000000, value: 107, options: { linewidth: 1, color: '#00E676', transp: 0 } },
            { time: 2000000, value: null, options: { color: '#787B86', style: 'linebr' } },
            { time: 2000000, value: null, options: { color: '#787B86', style: 'linebr' } },
            { time: 2000000, value: 155, options: { color: '#FFFFFF', style: 'linebr', linewidth: 2 } },
            { time: 2000000, value: 95, options: { color: '#FFFFFF', style: 'linebr', linewidth: 2 } },
          ],
        },
      };

      const result = runner.restructurePlots(input);

      expect(Object.keys(result)).toHaveLength(7);
      expect(result['Red Plot 1']).toBeDefined();
      expect(result['Black Plot 2']).toBeDefined();
      expect(result['Lime Plot 3']).toBeDefined();
      expect(result['Gray Line 4']).toBeDefined();
      expect(result['Gray Line 5']).toBeDefined();
      expect(result['White Level 6']).toBeDefined();
      expect(result['White Level 7']).toBeDefined();

      /* Verify each plot has correct number of points */
      Object.values(result).forEach((plot) => {
        expect(plot.data).toHaveLength(2);
      });

      /* Verify timestamps are in seconds */
      expect(result['Red Plot 1'].data[0].time).toBe(1000);
      expect(result['Red Plot 1'].data[1].time).toBe(2000);
    });

    it('should handle plots with identical colors by position', () => {
      const input = {
        Plot: {
          data: [
            { time: 1000000, value: 100, options: { color: '#FF5252', linewidth: 1 } },
            { time: 1000000, value: 200, options: { color: '#FF5252', linewidth: 1 } },
            { time: 2000000, value: 101, options: { color: '#FF5252', linewidth: 1 } },
            { time: 2000000, value: 201, options: { color: '#FF5252', linewidth: 1 } },
          ],
        },
      };

      const result = runner.restructurePlots(input);

      expect(Object.keys(result)).toHaveLength(2);
      expect(result['Red Plot 1']).toBeDefined();
      expect(result['Red Plot 2']).toBeDefined();

      /* First position should have values 100, 101 */
      expect(result['Red Plot 1'].data[0].value).toBe(100);
      expect(result['Red Plot 1'].data[1].value).toBe(101);

      /* Second position should have values 200, 201 */
      expect(result['Red Plot 2'].data[0].value).toBe(200);
      expect(result['Red Plot 2'].data[1].value).toBe(201);
    });
  });

  describe('Timestamp Conversion', () => {
    it('should convert milliseconds to seconds', () => {
      const input = {
        Plot: {
          data: [
            { time: 1609459200000, value: 100, options: { color: '#FF5252' } },
            { time: 1609545600000, value: 101, options: { color: '#FF5252' } },
          ],
        },
      };

      const result = runner.restructurePlots(input);

      expect(result['Red Plot 1'].data[0].time).toBe(1609459200);
      expect(result['Red Plot 1'].data[1].time).toBe(1609545600);
    });

    it('should handle fractional milliseconds', () => {
      const input = {
        Plot: {
          data: [{ time: 1609459200999, value: 100, options: { color: '#FF5252' } }],
        },
      };

      const result = runner.restructurePlots(input);

      expect(result['Red Plot 1'].data[0].time).toBe(1609459200);
    });
  });

  describe('Plot Naming', () => {
    it('should use counter suffix for unique names', () => {
      const result = runner.generatePlotName({ color: '#FF5252', linewidth: 1 }, 5);
      expect(result).toBe('Red Plot 5');
    });

    it('should name linebr style with linewidth 2 as Level', () => {
      const result = runner.generatePlotName(
        { color: '#FFFFFF', style: 'linebr', linewidth: 2 },
        3,
      );
      expect(result).toBe('White Level 3');
    });

    it('should name linebr style without linewidth 2 as Line', () => {
      const result = runner.generatePlotName({ color: '#787B86', style: 'linebr' }, 4);
      expect(result).toBe('Gray Line 4');
    });

    it('should handle unmapped colors', () => {
      const result = runner.generatePlotName({ color: '#123456', linewidth: 1 }, 8);
      expect(result).toBe('Color8 Plot 8');
    });

    it('should handle missing color with default', () => {
      const result = runner.generatePlotName({}, 1);
      expect(result).toBe('Color1 Plot 1');
    });
  });

  describe('Options Preservation', () => {
    it('should preserve all original options', () => {
      const input = {
        Plot: {
          data: [
            {
              time: 1000000,
              value: 100,
              options: { color: '#FF5252', linewidth: 2, transp: 50, style: 'line' },
            },
          ],
        },
      };

      const result = runner.restructurePlots(input);

      const plotData = result['Red Plot 1'].data[0];
      expect(plotData.options).toEqual({
        color: '#FF5252',
        linewidth: 2,
        transp: 50,
        style: 'line',
      });
    });

    it('should handle missing options gracefully', () => {
      const input = {
        Plot: {
          data: [{ time: 1000000, value: 100 }],
        },
      };

      const result = runner.restructurePlots(input);

      expect(Object.keys(result)).toHaveLength(1);
      expect(result['Color1 Plot 1']).toBeDefined();
    });
  });

  describe('Value Preservation', () => {
    it('should preserve null values', () => {
      const input = {
        Plot: {
          data: [
            { time: 1000000, value: null, options: { color: '#FF5252' } },
            { time: 2000000, value: 100, options: { color: '#FF5252' } },
          ],
        },
      };

      const result = runner.restructurePlots(input);

      expect(result['Red Plot 1'].data[0].value).toBeNull();
      expect(result['Red Plot 1'].data[1].value).toBe(100);
    });

    it('should preserve zero values', () => {
      const input = {
        Plot: {
          data: [{ time: 1000000, value: 0, options: { color: '#FF5252' } }],
        },
      };

      const result = runner.restructurePlots(input);

      expect(result['Red Plot 1'].data[0].value).toBe(0);
    });

    it('should preserve negative values', () => {
      const input = {
        Plot: {
          data: [{ time: 1000000, value: -42.5, options: { color: '#FF5252' } }],
        },
      };

      const result = runner.restructurePlots(input);

      expect(result['Red Plot 1'].data[0].value).toBe(-42.5);
    });
  });
});
