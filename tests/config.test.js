import { describe, it, expect, beforeEach, vi } from 'vitest';
import { PROVIDER_CHAIN, DEFAULTS } from '../config.js';

vi.mock('../PineTS/dist/pinets.dev.es.js', () => ({
  Provider: {
    Binance: { name: 'MockBinance' },
  },
}));

describe('config', () => {
  describe('PROVIDER_CHAIN', () => {
    it('should have 3 providers', () => {
      expect(PROVIDER_CHAIN).toHaveLength(3);
    });

    it('should have MOEX as first provider', () => {
      expect(PROVIDER_CHAIN[0].name).toBe('MOEX');
      expect(PROVIDER_CHAIN[0].instance).toBeDefined();
    });

    it('should have Binance as second provider', () => {
      expect(PROVIDER_CHAIN[1].name).toBe('Binance');
      expect(PROVIDER_CHAIN[1].instance).toBeDefined();
    });

    it('should have YahooFinance as third provider', () => {
      expect(PROVIDER_CHAIN[2].name).toBe('YahooFinance');
      expect(PROVIDER_CHAIN[2].instance).toBeDefined();
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
