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
      prefetchSecurityData: vi.fn().mockResolvedValue(undefined),
      run: vi.fn(),
    };

    /* Mock PineTS constructor */
    const { PineTS } = await import('../../../PineTS/dist/pinets.dev.es.js');
    PineTS.mockImplementation(() => mockPineTS);
  });

  describe('executeTranspiledStrategy', () => {
    it('should create PineTS and execute wrapped code', async () => {
      const { PineTS } = await import('../../../PineTS/dist/pinets.dev.es.js');
      const jsCode = 'plot(close, "Close", { color: color.blue });';
      const data = [{ time: 1, open: 100, high: 105, low: 95, close: 102 }];
      mockPineTS.run.mockResolvedValue({});

      const result = await runner.executeTranspiledStrategy(jsCode, data);

      expect(PineTS).toHaveBeenCalledWith(data);
      expect(mockPineTS.run).toHaveBeenCalledTimes(1);
      expect(mockPineTS.run).toHaveBeenCalledWith(expect.stringContaining(jsCode));
      expect(result).toEqual({ plots: [] });
    });

    it('should wrap jsCode in arrow function string', async () => {
      const jsCode = 'const ema = ta.ema(close, 9);';
      const data = [{ time: 1, open: 100 }];
      mockPineTS.run.mockResolvedValue({});

      await runner.executeTranspiledStrategy(jsCode, data);

      const callArg = mockPineTS.run.mock.calls[0][0];
      expect(callArg).toContain('(context) => {');
      expect(callArg).toContain('const ta = context.ta;');
      expect(callArg).toContain(
        'const { plot: corePlot, color, na, nz, fixnan, time } = context.core;',
      );
      expect(callArg).toContain('const syminfo = context.syminfo;');
      expect(callArg).toContain('function indicator() {}');
      expect(callArg).toContain('const strategy = context.strategy;');
      expect(callArg).toContain(jsCode);
    });

    it('should provide indicator and strategy stubs', async () => {
      const jsCode = 'indicator("Test", { overlay: true });';
      const data = [{ time: 1, open: 100 }];
      mockPineTS.run.mockResolvedValue({});

      await runner.executeTranspiledStrategy(jsCode, data);

      const callArg = mockPineTS.run.mock.calls[0][0];
      expect(callArg).toContain('function indicator() {}');
      expect(callArg).toContain('const strategy = context.strategy;');
    });

    it('should return empty plots array', async () => {
      const jsCode = 'const x = 1 + 1;';
      const data = [{ time: 1, open: 100 }];
      mockPineTS.run.mockResolvedValue({});

      const result = await runner.executeTranspiledStrategy(jsCode, data);

      expect(result.plots).toEqual([]);
    });
  });
});
