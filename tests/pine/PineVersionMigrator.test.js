import { describe, it, expect } from 'vitest';
import PineVersionMigrator from '../../src/pine/PineVersionMigrator.js';

describe('PineVersionMigrator', () => {
  describe('needsMigration', () => {
    it('should return false for version 5', () => {
      const pineCode = '//@version=5\nindicator("Test")';
      const result = PineVersionMigrator.needsMigration(pineCode, 5);
      expect(result).toBe(false);
    });

    it('should return true for version 4', () => {
      const pineCode = '//@version=4\nstudy("Test")';
      const result = PineVersionMigrator.needsMigration(pineCode, 4);
      expect(result).toBe(true);
    });

    it('should return true for version 3', () => {
      const pineCode = '//@version=3\nstudy("Test")';
      const result = PineVersionMigrator.needsMigration(pineCode, 3);
      expect(result).toBe(true);
    });

    it('should return true for null version', () => {
      const pineCode = 'study("Test")';
      const result = PineVersionMigrator.needsMigration(pineCode, null);
      expect(result).toBe(true);
    });

    it('should return true for version 2', () => {
      const pineCode = '//@version=2\nstudy("Test")';
      const result = PineVersionMigrator.needsMigration(pineCode, 2);
      expect(result).toBe(true);
    });

    it('should return true for version 1', () => {
      const pineCode = '//@version=1\nstudy("Test")';
      const result = PineVersionMigrator.needsMigration(pineCode, 1);
      expect(result).toBe(true);
    });
  });

  describe('migrate - study/indicator', () => {
    it('should migrate study to indicator', () => {
      const pineCode = '//@version=3\nstudy("Test Strategy")';
      const result = PineVersionMigrator.migrate(pineCode, 3);
      expect(result).toContain('indicator("Test Strategy")');
    });

    it('should not change version 5 code', () => {
      const pineCode = '//@version=5\nindicator("Test")\nma = ta.sma(close, 20)';
      const result = PineVersionMigrator.migrate(pineCode, 5);
      expect(result).toBe(pineCode);
    });
  });

  describe('migrate - ta.* functions', () => {
    it('should migrate sma to ta.sma', () => {
      const pineCode = '//@version=3\nma = sma(close, 20)';
      const result = PineVersionMigrator.migrate(pineCode, 3);
      expect(result).toContain('ma = ta.sma(close, 20)');
    });

    it('should migrate ema to ta.ema', () => {
      const pineCode = '//@version=4\nema20 = ema(close, 20)';
      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('ta.ema(close, 20)');
    });

    it('should migrate rsi to ta.rsi', () => {
      const pineCode = '//@version=4\nrsiValue = rsi(close, 14)';
      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('ta.rsi(close, 14)');
    });

    it('should migrate multiple ta functions', () => {
      const pineCode = '//@version=3\nma20 = sma(close, 20)\nma50 = ema(close, 50)\nrsi14 = rsi(close, 14)';
      const result = PineVersionMigrator.migrate(pineCode, 3);
      expect(result).toContain('ta.sma(close, 20)');
      expect(result).toContain('ta.ema(close, 50)');
      expect(result).toContain('ta.rsi(close, 14)');
    });

    it('should migrate crossover to ta.crossover', () => {
      const pineCode = '//@version=4\nbullish = crossover(fast, slow)';
      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('ta.crossover(fast, slow)');
    });

    it('should migrate highest to ta.highest', () => {
      const pineCode = '//@version=4\nhi = highest(high, 10)';
      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('ta.highest(high, 10)');
    });
  });

  describe('migrate - request.* functions', () => {
    it('should migrate security to request.security', () => {
      const pineCode = '//@version=4\ndailyClose = security(tickerid, "D", close)';
      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('dailyClose = request.security(syminfo.tickerid, "D", close)');
    });

    it('should migrate financial to request.financial', () => {
      const pineCode = '//@version=4\nearnings = financial(tickerid, "EARNINGS")';
      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('request.financial(syminfo.tickerid, "EARNINGS")');
    });

    it('should migrate splits to request.splits', () => {
      const pineCode = '//@version=4\nsplitData = splits(tickerid)';
      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('request.splits(syminfo.tickerid)');
    });
  });

  describe('migrate - math.* functions', () => {
    it('should migrate abs to math.abs', () => {
      const pineCode = '//@version=4\nabsValue = abs(-5)';
      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('math.abs(-5)');
    });

    it('should migrate max to math.max', () => {
      const pineCode = '//@version=4\nmaxVal = max(a, b)';
      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('math.max(a, b)');
    });

    it('should migrate sqrt to math.sqrt', () => {
      const pineCode = '//@version=4\nsqrtVal = sqrt(16)';
      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('math.sqrt(16)');
    });

    it('should migrate multiple math functions', () => {
      const pineCode = '//@version=4\nval1 = abs(a)\nval2 = max(b, c)\nval3 = sqrt(d)';
      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('math.abs(a)');
      expect(result).toContain('math.max(b, c)');
      expect(result).toContain('math.sqrt(d)');
    });
  });

  describe('migrate - ticker.* functions', () => {
    it('should migrate heikinashi to ticker.heikinashi', () => {
      const pineCode = '//@version=4\nhaData = heikinashi(tickerid)';
      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('ticker.heikinashi(syminfo.tickerid)');
    });

    it('should migrate renko to ticker.renko', () => {
      const pineCode = '//@version=4\nrenkoData = renko(tickerid)';
      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('ticker.renko(syminfo.tickerid)');
    });

    it('should migrate tickerid() to ticker.new()', () => {
      const pineCode = '//@version=4\ntid = tickerid()';
      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('ticker.new()');
    });
  });

  describe('migrate - str.* functions', () => {
    it('should migrate tostring to str.tostring', () => {
      const pineCode = '//@version=4\ntext = tostring(value)';
      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('str.tostring(value)');
    });

    it('should migrate tonumber to str.tonumber', () => {
      const pineCode = '//@version=4\nnum = tonumber(text)';
      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('str.tonumber(text)');
    });
  });

  describe('migrate - complex scenarios', () => {
    it('should migrate complete v3 strategy', () => {
      const pineCode = `//@version=3
study("V3 Strategy", overlay=true)
ma20 = sma(close, 20)
ma50 = ema(close, 50)
bullish = crossover(ma20, ma50)
plot(ma20, color=yellow)`;

      const result = PineVersionMigrator.migrate(pineCode, 3);
      expect(result).toContain('indicator("V3 Strategy"');
      expect(result).toContain('ta.sma(close, 20)');
      expect(result).toContain('ta.ema(close, 50)');
      expect(result).toContain('ta.crossover(ma20, ma50)');
    });

    it('should migrate v4 with security and ta functions', () => {
      const pineCode = `//@version=4
study("V4 Security")
dailyMA = security(tickerid, 'D', sma(close, 20))
rsiVal = rsi(close, 14)`;

      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('indicator("V4 Security")');
      expect(result).toContain('request.security(syminfo.tickerid');
      expect(result).toContain('ta.sma(close, 20)');
      expect(result).toContain('ta.rsi(close, 14)');
    });

    it('should handle nested function calls', () => {
      const pineCode = '//@version=4\nval = abs(max(sma(close, 20), ema(close, 50)))';
      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('math.abs(math.max(ta.sma(close, 20), ta.ema(close, 50)))');
    });

    it('should not migrate identifiers without function calls', () => {
      const pineCode = '//@version=4\nvar sma_value = 100';
      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('var sma_value = 100');
      expect(result).not.toContain('ta.sma_value');
    });

    it('should handle multiple occurrences', () => {
      const pineCode = `//@version=4
ma1 = sma(close, 20)
ma2 = sma(open, 20)
ma3 = sma(high, 20)`;

      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('ma1 = ta.sma(close, 20)');
      expect(result).toContain('ma2 = ta.sma(open, 20)');
      expect(result).toContain('ma3 = ta.sma(high, 20)');
    });
  });

  describe('migrate - edge cases', () => {
    it('should handle empty code', () => {
      const pineCode = '';
      const result = PineVersionMigrator.migrate(pineCode, 3);
      expect(result).toBe('');
    });

    it('should handle code with only version comment', () => {
      const pineCode = '//@version=3';
      const result = PineVersionMigrator.migrate(pineCode, 3);
      expect(result).toBe('//@version=3');
    });

    it('should handle code with comments containing function names', () => {
      const pineCode = `//@version=4
// This uses sma function
ma = sma(close, 20)`;

      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('// This uses sma function');
      expect(result).toContain('ta.sma(close, 20)');
    });

    it('should preserve whitespace and formatting', () => {
      const pineCode = `//@version=4
ma20    =    sma(close,    20)`;

      const result = PineVersionMigrator.migrate(pineCode, 4);
      expect(result).toContain('ta.sma(close,    20)');
    });
  });

  describe('escapeRegex', () => {
    it('should escape special regex characters', () => {
      const escaped = PineVersionMigrator.escapeRegex('test(value)');
      expect(escaped).toBe('test\\(value\\)');
    });

    it('should escape multiple special characters', () => {
      const escaped = PineVersionMigrator.escapeRegex('a.b[c]d*e+f?g^h$i{j}k|l');
      expect(escaped).toContain('\\.');
      expect(escaped).toContain('\\[');
      expect(escaped).toContain('\\]');
      expect(escaped).toContain('\\*');
    });
  });
});
