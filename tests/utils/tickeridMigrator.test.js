import { describe, it, expect } from 'vitest';
import TickeridMigrator from '../../src/utils/tickeridMigrator.js';

describe('TickeridMigrator', () => {
  describe('standalone tickerid variable', () => {
    it('migrates tickerid in security() call', () => {
      const input = 'ma20 = security(tickerid, "D", sma(close, 20))';
      const expected = 'ma20 = security(syminfo.tickerid, "D", sma(close, 20))';
      expect(TickeridMigrator.migrate(input)).toBe(expected);
    });

    it('migrates multiple tickerid occurrences', () => {
      const input = `ma20 = security(tickerid, 'D', sma(close, 20))
ma50 = security(tickerid, 'D', sma(close, 50))
ma200 = security(tickerid, 'D', sma(close, 200))`;
      const expected = `ma20 = security(syminfo.tickerid, 'D', sma(close, 20))
ma50 = security(syminfo.tickerid, 'D', sma(close, 50))
ma200 = security(syminfo.tickerid, 'D', sma(close, 200))`;
      expect(TickeridMigrator.migrate(input)).toBe(expected);
    });

    it('migrates tickerid in assignment', () => {
      const input = 'symbol = tickerid';
      const expected = 'symbol = syminfo.tickerid';
      expect(TickeridMigrator.migrate(input)).toBe(expected);
    });

    it('migrates tickerid with spaces', () => {
      const input = 'security( tickerid , "D", close)';
      const expected = 'security( syminfo.tickerid , "D", close)';
      expect(TickeridMigrator.migrate(input)).toBe(expected);
    });

    it('migrates tickerid at start of line', () => {
      const input = 'tickerid';
      const expected = 'syminfo.tickerid';
      expect(TickeridMigrator.migrate(input)).toBe(expected);
    });

    it('migrates tickerid at end of line', () => {
      const input = 'symbol = tickerid';
      const expected = 'symbol = syminfo.tickerid';
      expect(TickeridMigrator.migrate(input)).toBe(expected);
    });
  });

  describe('tickerId camelCase variant', () => {
    it('migrates tickerId to syminfo.tickerid', () => {
      const input = 'ma20 = security(tickerId, "D", sma(close, 20))';
      const expected = 'ma20 = security(syminfo.tickerid, "D", sma(close, 20))';
      expect(TickeridMigrator.migrate(input)).toBe(expected);
    });
  });

  describe('tickerid() function call', () => {
    it('migrates tickerid() to ticker.new()', () => {
      const input = 'symbol = tickerid()';
      const expected = 'symbol = ticker.new()';
      expect(TickeridMigrator.migrate(input)).toBe(expected);
    });

    it('migrates tickerid() with spaces', () => {
      const input = 'symbol = tickerid( )';
      const expected = 'symbol = ticker.new( )';
      expect(TickeridMigrator.migrate(input)).toBe(expected);
    });
  });

  describe('should NOT migrate', () => {
    it('does not migrate syminfo.tickerid', () => {
      const input = 'ma20 = security(syminfo.tickerid, "D", sma(close, 20))';
      expect(TickeridMigrator.migrate(input)).toBe(input);
    });

    it('does not migrate when part of identifier', () => {
      const input = 'mytickeridfunc()';
      expect(TickeridMigrator.migrate(input)).toBe(input);
    });

    it('does not migrate tickerid_custom', () => {
      const input = 'tickerid_custom = "BTCUSDT"';
      expect(TickeridMigrator.migrate(input)).toBe(input);
    });

    it('does not migrate custom_tickerid', () => {
      const input = 'custom_tickerid = "BTCUSDT"';
      expect(TickeridMigrator.migrate(input)).toBe(input);
    });
  });

  describe('real-world examples', () => {
    it('migrates daily-lines.pine strategy', () => {
      const input = `study(title="20-50-100-200 SMA Daily", shorttitle="Daily Lines", overlay=true)
ma20 = security(tickerid, 'D', sma(close, 20))
ma50 = security(tickerid, 'D', sma(close, 50))
ma200 = security(tickerid, 'D', sma(close, 200))`;
      const expected = `study(title="20-50-100-200 SMA Daily", shorttitle="Daily Lines", overlay=true)
ma20 = security(syminfo.tickerid, 'D', sma(close, 20))
ma50 = security(syminfo.tickerid, 'D', sma(close, 50))
ma200 = security(syminfo.tickerid, 'D', sma(close, 200))`;
      expect(TickeridMigrator.migrate(input)).toBe(expected);
    });

    it('handles mixed tickerid usage', () => {
      const input = `symbol = tickerid
price = security(tickerid, "D", close)
newSymbol = tickerid()`;
      const expected = `symbol = syminfo.tickerid
price = security(syminfo.tickerid, "D", close)
newSymbol = ticker.new()`;
      expect(TickeridMigrator.migrate(input)).toBe(expected);
    });
  });
});
