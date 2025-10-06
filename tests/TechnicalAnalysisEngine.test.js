import { describe, it, expect, beforeEach, vi } from 'vitest';
import { TechnicalAnalysisEngine } from '../classes/TechnicalAnalysisEngine.js';

/* Mock PineTS module */
vi.mock('../../PineTS/dist/pinets.dev.es.js', () => ({
  PineTS: vi.fn(),
}));

describe('TechnicalAnalysisEngine', () => {
  let engine;
  let mockPineTS;

  beforeEach(async () => {
    engine = new TechnicalAnalysisEngine();

    /* Create mock PineTS instance */
    mockPineTS = {
      ready: vi.fn().mockResolvedValue(undefined),
      run: vi.fn(),
    };

    /* Mock PineTS constructor */
    const { PineTS } = await import('../../PineTS/dist/pinets.dev.es.js');
    PineTS.mockImplementation(() => mockPineTS);
  });

  describe('createPineTSAdapter()', () => {
    it('should create PineTS instance with market data', async () => {
      const { PineTS } = await import('../../PineTS/dist/pinets.dev.es.js');
      const data = [{ time: 1, open: 100, high: 105, low: 95, close: 102 }];

      const result = await engine.createPineTSAdapter('BINANCE', data, {}, 'BTCUSDT', 'D', 100);

      expect(PineTS).toHaveBeenCalledWith(data, 'BTCUSDT', 'D', 100);
      expect(mockPineTS.ready).toHaveBeenCalled();
      expect(result).toBe(mockPineTS);
    });

    it('should pass correct parameters to PineTS', async () => {
      const { PineTS } = await import('../../PineTS/dist/pinets.dev.es.js');
      const data = [{ time: 1, open: 100 }];

      await engine.createPineTSAdapter('YAHOO', data, {}, 'AAPL', 'W', 200);

      expect(PineTS).toHaveBeenCalledWith(data, 'AAPL', 'W', 200);
    });

    it('should wait for PineTS ready()', async () => {
      const data = [{ time: 1 }];
      await engine.createPineTSAdapter('TEST', data, {}, 'TEST', 'D', 100);

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

      const result = await engine.runEMAStrategy(mockPineTS);

      expect(mockPineTS.run).toHaveBeenCalledTimes(1);
      expect(result).toEqual({
        result: mockPlots,
        plots: mockPlots,
      });
    });

    it('should call PineTS.run with strategy function', async () => {
      mockPineTS.run.mockResolvedValue({ plots: {} });

      await engine.runEMAStrategy(mockPineTS);

      expect(mockPineTS.run).toHaveBeenCalledWith(expect.any(Function));
    });

    it('should handle empty plots', async () => {
      mockPineTS.run.mockResolvedValue({ plots: null });

      const result = await engine.runEMAStrategy(mockPineTS);

      expect(result.plots).toEqual({});
    });

    it('should handle undefined plots', async () => {
      mockPineTS.run.mockResolvedValue({});

      const result = await engine.runEMAStrategy(mockPineTS);

      expect(result.plots).toEqual({});
    });
  });

  describe('getIndicatorMetadata()', () => {
    it('should return indicator metadata', () => {
      const metadata = engine.getIndicatorMetadata();

      expect(metadata).toEqual({
        EMA9: { title: 'EMA 9', type: 'moving_average' },
        EMA18: { title: 'EMA 18', type: 'moving_average' },
        BullSignal: { title: 'Bull Signal', type: 'signal' },
      });
    });

    it('should return consistent metadata on multiple calls', () => {
      const metadata1 = engine.getIndicatorMetadata();
      const metadata2 = engine.getIndicatorMetadata();

      expect(metadata1).toEqual(metadata2);
    });

    it('should have correct indicator types', () => {
      const metadata = engine.getIndicatorMetadata();

      expect(metadata.EMA9.type).toBe('moving_average');
      expect(metadata.EMA18.type).toBe('moving_average');
      expect(metadata.BullSignal.type).toBe('signal');
    });
  });
});
