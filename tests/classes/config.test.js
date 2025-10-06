import { describe, it, expect, beforeEach, vi } from 'vitest';
import { createProviderChain, DEFAULTS } from '../../src/config.js';

vi.mock('../PineTS/dist/pinets.dev.es.js', () => ({
  Provider: {
    Binance: { name: 'MockBinance' },
  },
}));

describe('config', () => {
  describe('createProviderChain', () => {
    it('should return 3 providers', () => {
      const mockLogger = { debug: vi.fn(), log: vi.fn() };
      const chain = createProviderChain(mockLogger);
      expect(chain).toHaveLength(3);
    });

    it('should have MOEX as first provider', () => {
      const mockLogger = { debug: vi.fn(), log: vi.fn() };
      const chain = createProviderChain(mockLogger);
      expect(chain[0].name).toBe('MOEX');
      expect(chain[0].instance).toBeDefined();
    });

    it('should have Binance as second provider', () => {
      const mockLogger = { debug: vi.fn(), log: vi.fn() };
      const chain = createProviderChain(mockLogger);
      expect(chain[1].name).toBe('Binance');
      expect(chain[1].instance).toBeDefined();
    });

    it('should have YahooFinance as third provider', () => {
      const mockLogger = { debug: vi.fn(), log: vi.fn() };
      const chain = createProviderChain(mockLogger);
      expect(chain[2].name).toBe('YahooFinance');
      expect(chain[2].instance).toBeDefined();
    });
  });

  describe('DEFAULTS', () => {
    it('should have symbol from env or default BTCUSDT', () => {
      expect(DEFAULTS.symbol).toBe(process.env.SYMBOL || 'BTCUSDT');
    });

    it('should have timeframe from env or default D', () => {
      expect(DEFAULTS.timeframe).toBe(process.env.TIMEFRAME || 'D');
    });

    it('should have bars from env or default 100', () => {
      expect(DEFAULTS.bars).toBe(parseInt(process.env.BARS) || 100);
    });

    it('should have strategy name', () => {
      expect(DEFAULTS.strategy).toBe('EMA Crossover Strategy');
    });
  });
});
