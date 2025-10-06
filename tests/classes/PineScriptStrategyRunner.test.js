import { describe, it, expect, beforeEach, vi } from 'vitest';
import { PineScriptStrategyRunner } from '../../src/classes/PineScriptStrategyRunner.js';

/* Mock PineTS module */
vi.mock('../../../PineTS/dist/pinets.dev.es.js', () => ({
  PineTS: vi.fn(),
}));

describe('PineScriptStrategyRunner', () => {
  let runner;
  let mockPineTS;

  beforeEach(async () => {
    runner = new PineScriptStrategyRunner();

    /* Create mock PineTS instance */
    mockPineTS = {
      ready: vi.fn().mockResolvedValue(undefined),
      run: vi.fn(),
    };

    /* Mock PineTS constructor */
    const { PineTS } = await import('../../../PineTS/dist/pinets.dev.es.js');
    PineTS.mockImplementation(() => mockPineTS);
  });

  describe('createPineTSAdapter()', () => {
    it('should create PineTS instance with market data', async () => {
      const { PineTS } = await import('../../../PineTS/dist/pinets.dev.es.js');
      const data = [{ time: 1, open: 100, high: 105, low: 95, close: 102 }];

      const result = await runner.createPineTSAdapter('BINANCE', data, {}, 'BTCUSDT', 'D', 100);

      expect(PineTS).toHaveBeenCalledWith(data, 'BTCUSDT', 'D', 100);
      expect(mockPineTS.ready).toHaveBeenCalled();
      expect(result).toBe(mockPineTS);
    });

    it('should pass correct parameters to PineTS', async () => {
      const { PineTS } = await import('../../../PineTS/dist/pinets.dev.es.js');
      const data = [{ time: 1, open: 100 }];

      await runner.createPineTSAdapter('YAHOO', data, {}, 'AAPL', 'W', 200);

      expect(PineTS).toHaveBeenCalledWith(data, 'AAPL', 'W', 200);
    });

    it('should wait for PineTS ready()', async () => {
      const data = [{ time: 1 }];
      await runner.createPineTSAdapter('TEST', data, {}, 'TEST', 'D', 100);

      expect(mockPineTS.ready).toHaveBeenCalledTimes(1);
    });
  });

  describe('runEMAStrategy()', () => {
    it('should run PineTS strategy and return plots', async () => {
      const mockPlots = {
        EMA9: [100, 101, 102],
        EMA18: [99, 100, 101],
        BullSignal: [1, 1, 1],
      };
      mockPineTS.run.mockResolvedValue({ plots: mockPlots });

      const result = await runner.runEMAStrategy(mockPineTS);

      expect(mockPineTS.run).toHaveBeenCalledTimes(1);
      expect(result).toEqual({
        result: mockPlots,
        plots: mockPlots,
      });
    });

    it('should call PineTS.run with strategy function', async () => {
      mockPineTS.run.mockResolvedValue({ plots: {} });

      await runner.runEMAStrategy(mockPineTS);

      expect(mockPineTS.run).toHaveBeenCalledWith(expect.any(Function));
    });

    it('should handle empty plots', async () => {
      mockPineTS.run.mockResolvedValue({ plots: null });

      const result = await runner.runEMAStrategy(mockPineTS);

      expect(result.plots).toEqual({});
    });

    it('should handle undefined plots', async () => {
      mockPineTS.run.mockResolvedValue({});

      const result = await runner.runEMAStrategy(mockPineTS);

      expect(result.plots).toEqual({});
    });
  });

  describe('getIndicatorMetadata()', () => {
    it('should return indicator metadata', () => {
      const metadata = runner.getIndicatorMetadata();

      expect(metadata).toEqual({
        EMA9: { title: 'EMA 9', type: 'moving_average' },
        EMA18: { title: 'EMA 18', type: 'moving_average' },
        BullSignal: { title: 'Bull Signal', type: 'signal' },
      });
    });

    it('should return consistent metadata on multiple calls', () => {
      const metadata1 = runner.getIndicatorMetadata();
      const metadata2 = runner.getIndicatorMetadata();

      expect(metadata1).toEqual(metadata2);
    });

    it('should have correct indicator types', () => {
      const metadata = runner.getIndicatorMetadata();

      expect(metadata.EMA9.type).toBe('moving_average');
      expect(metadata.EMA18.type).toBe('moving_average');
      expect(metadata.BullSignal.type).toBe('signal');
    });
  });

  describe('executeTranspiledStrategy', () => {
    it('should execute simple JavaScript code and return plots', () => {
      const jsCode = `
        context.core.plot([1, 2, 3], 'Test Plot', { color: 'blue' });
      `;
      const marketData = {
        open: [100, 101, 102],
        high: [103, 104, 105],
        low: [99, 100, 101],
        close: [102, 103, 104],
        volume: [1000, 1100, 1200],
      };

      const result = runner.executeTranspiledStrategy(jsCode, marketData);

      expect(result).toHaveProperty('plots');
      expect(Array.isArray(result.plots)).toBe(true);
      expect(result.plots.length).toBe(1);
      expect(result.plots[0].title).toBe('Test Plot');
      expect(result.plots[0].series).toEqual([1, 2, 3]);
    });

    it('should provide market data arrays in context', () => {
      const jsCode = `
        context.core.plot(context.data.close, 'Close', {});
      `;
      const marketData = {
        open: [100],
        high: [103],
        low: [99],
        close: [102],
        volume: [1000],
      };

      const result = runner.executeTranspiledStrategy(jsCode, marketData);

      expect(result.plots[0].series).toEqual([102]);
    });

    it('should provide ta library stubs in context', () => {
      const jsCode = `
        const ema = context.ta.ema(context.data.close, 9);
        context.core.plot(ema, 'EMA', {});
      `;
      const marketData = {
        open: [100, 101, 102],
        high: [103, 104, 105],
        low: [99, 100, 101],
        close: [102, 103, 104],
        volume: [1000, 1100, 1200],
      };

      const result = runner.executeTranspiledStrategy(jsCode, marketData);

      expect(result.plots.length).toBe(1);
      expect(result.plots[0].title).toBe('EMA');
    });

    it('should throw error when executing invalid code', () => {
      const jsCode = 'throw new Error("Test error");';
      const marketData = {
        open: [100],
        high: [103],
        low: [99],
        close: [102],
        volume: [1000],
      };

      expect(() => {
        runner.executeTranspiledStrategy(jsCode, marketData);
      }).toThrow('Strategy execution failed: Test error');
    });

    it('should handle syntax errors in transpiled code', () => {
      const jsCode = 'const x = {invalid syntax';
      const marketData = {
        open: [100],
        high: [103],
        low: [99],
        close: [102],
        volume: [1000],
      };

      expect(() => {
        runner.executeTranspiledStrategy(jsCode, marketData);
      }).toThrow(/Strategy execution failed/);
    });

    it('should return empty plots array when no plots generated', () => {
      const jsCode = 'const x = 1 + 1;';
      const marketData = {
        open: [100],
        high: [103],
        low: [99],
        close: [102],
        volume: [1000],
      };

      const result = runner.executeTranspiledStrategy(jsCode, marketData);

      expect(result.plots).toEqual([]);
    });
  });
});
