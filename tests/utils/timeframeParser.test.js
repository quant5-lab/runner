import { describe, test, expect } from 'vitest';
import { TimeframeParser } from '../../src/utils/timeframeParser.js';

describe('TimeframeParser', () => {
  describe('parseToMinutes', () => {
    test('should parse string timeframes correctly', () => {
      expect(TimeframeParser.parseToMinutes('1m')).toBe(1);
      expect(TimeframeParser.parseToMinutes('5m')).toBe(5);
      expect(TimeframeParser.parseToMinutes('15m')).toBe(15);
      expect(TimeframeParser.parseToMinutes('30m')).toBe(30);
      expect(TimeframeParser.parseToMinutes('1h')).toBe(60);
      expect(TimeframeParser.parseToMinutes('4h')).toBe(240);
      expect(TimeframeParser.parseToMinutes('1d')).toBe(1440);
      // Large timeframes use single letters only - no digit prefixes
      expect(TimeframeParser.parseToMinutes('W')).toBe(10080);
      expect(TimeframeParser.parseToMinutes('M')).toBe(43200);
    });

    test('should parse numeric timeframes correctly', () => {
      expect(TimeframeParser.parseToMinutes(1)).toBe(1);
      expect(TimeframeParser.parseToMinutes(5)).toBe(5);
      expect(TimeframeParser.parseToMinutes(15)).toBe(15);
      expect(TimeframeParser.parseToMinutes(30)).toBe(30);
      expect(TimeframeParser.parseToMinutes(60)).toBe(60);
      expect(TimeframeParser.parseToMinutes(240)).toBe(240);
      expect(TimeframeParser.parseToMinutes(1440)).toBe(1440);
    });

    test('should parse letter timeframes correctly', () => {
      expect(TimeframeParser.parseToMinutes('D')).toBe(1440);
      expect(TimeframeParser.parseToMinutes('W')).toBe(10080);
      expect(TimeframeParser.parseToMinutes('M')).toBe(43200);
    });

    test('should return 1440 (daily) for unparseable inputs', () => {
      expect(TimeframeParser.parseToMinutes('invalid')).toBe(1440);
      expect(TimeframeParser.parseToMinutes(null)).toBe(1440);
      expect(TimeframeParser.parseToMinutes(undefined)).toBe(1440);
      expect(TimeframeParser.parseToMinutes('')).toBe(1440);
      expect(TimeframeParser.parseToMinutes('xyz')).toBe(1440);
      /* Valid numeric+letter formats: 1w, 1M correctly parse */
      expect(TimeframeParser.parseToMinutes('1w')).toBe(10080); // Valid weekly
      expect(TimeframeParser.parseToMinutes('1W')).toBe(1440); // Invalid - capital W without digit
      expect(TimeframeParser.parseToMinutes('1M')).toBe(43200); // Valid monthly
    });
  });

  describe('toMoexInterval', () => {
    test('should convert string timeframes to MOEX intervals', () => {
      expect(TimeframeParser.toMoexInterval('1m')).toBe('1');
      expect(TimeframeParser.toMoexInterval('10m')).toBe('10');
      expect(TimeframeParser.toMoexInterval('1h')).toBe('60');
      expect(TimeframeParser.toMoexInterval('1d')).toBe('24');

      // Test unsupported timeframes throw TimeframeError
      expect(() => TimeframeParser.toMoexInterval('5m')).toThrow("Timeframe '5m' not supported");
      expect(() => TimeframeParser.toMoexInterval('15m')).toThrow("Timeframe '15m' not supported");
      expect(() => TimeframeParser.toMoexInterval('30m')).toThrow("Timeframe '30m' not supported");
      expect(() => TimeframeParser.toMoexInterval('4h')).toThrow("Timeframe '4h' not supported");
    });

    test('should convert numeric timeframes to MOEX intervals', () => {
      expect(TimeframeParser.toMoexInterval(1)).toBe('1');
      expect(TimeframeParser.toMoexInterval(10)).toBe('10');
      expect(TimeframeParser.toMoexInterval(60)).toBe('60');
      expect(TimeframeParser.toMoexInterval(1440)).toBe('24');

      // Test unsupported numeric timeframes throw TimeframeError
      expect(() => TimeframeParser.toMoexInterval(5)).toThrow("Timeframe '5' not supported");
      expect(() => TimeframeParser.toMoexInterval(15)).toThrow("Timeframe '15' not supported");
      expect(() => TimeframeParser.toMoexInterval(30)).toThrow("Timeframe '30' not supported");
      expect(() => TimeframeParser.toMoexInterval(240)).toThrow("Timeframe '240' not supported");
    });

    test('should convert letter timeframes to MOEX intervals', () => {
      expect(TimeframeParser.toMoexInterval('D')).toBe('24');
      expect(TimeframeParser.toMoexInterval('W')).toBe('7');
      expect(TimeframeParser.toMoexInterval('M')).toBe('31');
    });

    test('should fallback to daily for invalid timeframes', () => {
      expect(TimeframeParser.toMoexInterval('invalid')).toBe('24');
      expect(TimeframeParser.toMoexInterval(null)).toBe('24');
      expect(TimeframeParser.toMoexInterval(undefined)).toBe('24');
      expect(TimeframeParser.toMoexInterval('')).toBe('24');
    });
  });

  describe('toYahooInterval', () => {
    test('should convert string timeframes to Yahoo intervals', () => {
      expect(TimeframeParser.toYahooInterval('1m')).toBe('1m');
      expect(TimeframeParser.toYahooInterval('5m')).toBe('5m');
      expect(TimeframeParser.toYahooInterval('15m')).toBe('15m');
      expect(TimeframeParser.toYahooInterval('30m')).toBe('30m');
      expect(TimeframeParser.toYahooInterval('1h')).toBe('1h');
      expect(TimeframeParser.toYahooInterval('1d')).toBe('1d');

      // Test unsupported timeframes throw TimeframeError
      expect(() => TimeframeParser.toYahooInterval('4h')).toThrow("Timeframe '4h' not supported");
    });

    test('should convert numeric timeframes to Yahoo intervals', () => {
      expect(TimeframeParser.toYahooInterval(1)).toBe('1m');
      expect(TimeframeParser.toYahooInterval(5)).toBe('5m');
      expect(TimeframeParser.toYahooInterval(15)).toBe('15m');
      expect(TimeframeParser.toYahooInterval(30)).toBe('30m');
      expect(TimeframeParser.toYahooInterval(60)).toBe('1h');
      expect(TimeframeParser.toYahooInterval(1440)).toBe('1d');

      // Test unsupported numeric timeframes throw TimeframeError
      expect(() => TimeframeParser.toYahooInterval(240)).toThrow("Timeframe '240' not supported");
    });

    test('should convert letter timeframes to Yahoo intervals', () => {
      expect(TimeframeParser.toYahooInterval('D')).toBe('1d');
      expect(TimeframeParser.toYahooInterval('W')).toBe('1wk');
      expect(TimeframeParser.toYahooInterval('M')).toBe('1mo');
    });

    test('should fallback to daily for invalid timeframes', () => {
      expect(TimeframeParser.toYahooInterval('invalid')).toBe('1d');
      expect(TimeframeParser.toYahooInterval(null)).toBe('1d');
      expect(TimeframeParser.toYahooInterval(undefined)).toBe('1d');
      expect(TimeframeParser.toYahooInterval('')).toBe('1d');
    });
  });

  describe('regression tests for critical timeframe bug', () => {
    test('10m string should not fallback to daily - MOEX supported', () => {
      /* This test prevents the critical bug where supported timeframes were parsed as daily */
      expect(TimeframeParser.parseToMinutes('10m')).toBe(10);
      expect(TimeframeParser.toMoexInterval('10m')).toBe('10');

      /* These should NOT be daily fallbacks */
      expect(TimeframeParser.toMoexInterval('10m')).not.toBe('24');
    });

    test('1h string should not fallback to daily', () => {
      /* This test prevents the critical bug where "1h" was parsed as daily */
      expect(TimeframeParser.parseToMinutes('1h')).toBe(60);
      expect(TimeframeParser.toMoexInterval('1h')).toBe('60');
      expect(TimeframeParser.toYahooInterval('1h')).toBe('1h');

      /* These should NOT be daily fallbacks */
      expect(TimeframeParser.toMoexInterval('1h')).not.toBe('24');
      expect(TimeframeParser.toYahooInterval('1h')).not.toBe('1d');
    });

    test('supported timeframes should parse correctly', () => {
      /* MOEX supported timeframes */
      const moexSupported = ['1m', '10m', '1h', '1d', 'D', 'W', 'M'];

      for (const tf of moexSupported) {
        expect(TimeframeParser.parseToMinutes(tf)).toBeGreaterThan(0);
        expect(() => TimeframeParser.toMoexInterval(tf)).not.toThrow();
      }

      /* Yahoo supported timeframes */
      const yahooSupported = ['1m', '5m', '15m', '30m', '1h', '1d', 'D', 'W', 'M'];

      for (const tf of yahooSupported) {
        expect(TimeframeParser.parseToMinutes(tf)).toBeGreaterThan(0);
        expect(() => TimeframeParser.toYahooInterval(tf)).not.toThrow();
      }

      /* MOEX unsupported should throw TimeframeError */
      const moexUnsupported = ['5m', '15m', '30m', '4h'];

      for (const tf of moexUnsupported) {
        expect(() => TimeframeParser.toMoexInterval(tf)).toThrow("not supported");
      }
    });
  });

  describe('toBinanceTimeframe', () => {
    test('should convert string timeframes to Binance format', () => {
      expect(TimeframeParser.toBinanceTimeframe('1m')).toBe('1');
      expect(TimeframeParser.toBinanceTimeframe('3m')).toBe('3');
      expect(TimeframeParser.toBinanceTimeframe('5m')).toBe('5');
      expect(TimeframeParser.toBinanceTimeframe('15m')).toBe('15');
      expect(TimeframeParser.toBinanceTimeframe('30m')).toBe('30');
      expect(TimeframeParser.toBinanceTimeframe('1h')).toBe('60');
      expect(TimeframeParser.toBinanceTimeframe('2h')).toBe('120');
      expect(TimeframeParser.toBinanceTimeframe('4h')).toBe('240');
      expect(TimeframeParser.toBinanceTimeframe('6h')).toBe('360');
      expect(TimeframeParser.toBinanceTimeframe('8h')).toBe('480');
      expect(TimeframeParser.toBinanceTimeframe('12h')).toBe('720');
      expect(TimeframeParser.toBinanceTimeframe('1d')).toBe('D');
      expect(TimeframeParser.toBinanceTimeframe('D')).toBe('D');
      expect(TimeframeParser.toBinanceTimeframe('W')).toBe('W');
      expect(TimeframeParser.toBinanceTimeframe('M')).toBe('M');
    });

    test('should convert numeric timeframes to Binance format', () => {
      expect(TimeframeParser.toBinanceTimeframe(1)).toBe('1');
      expect(TimeframeParser.toBinanceTimeframe(3)).toBe('3');
      expect(TimeframeParser.toBinanceTimeframe(5)).toBe('5');
      expect(TimeframeParser.toBinanceTimeframe(15)).toBe('15');
      expect(TimeframeParser.toBinanceTimeframe(30)).toBe('30');
      expect(TimeframeParser.toBinanceTimeframe(60)).toBe('60');
      expect(TimeframeParser.toBinanceTimeframe(120)).toBe('120');
      expect(TimeframeParser.toBinanceTimeframe(240)).toBe('240');
      expect(TimeframeParser.toBinanceTimeframe(360)).toBe('360');
      expect(TimeframeParser.toBinanceTimeframe(480)).toBe('480');
      expect(TimeframeParser.toBinanceTimeframe(720)).toBe('720');
      expect(TimeframeParser.toBinanceTimeframe(1440)).toBe('D');
      expect(TimeframeParser.toBinanceTimeframe(10080)).toBe('W');
      expect(TimeframeParser.toBinanceTimeframe(43200)).toBe('M');
    });

    test('should default to D for unparseable timeframes', () => {
      expect(TimeframeParser.toBinanceTimeframe('invalid')).toBe('D'); // defaults to daily
      expect(TimeframeParser.toBinanceTimeframe(null)).toBe('D'); // defaults to daily
      expect(TimeframeParser.toBinanceTimeframe(undefined)).toBe('D'); // defaults to daily

      // However, specific numeric values that don't map should throw
      expect(() => TimeframeParser.toBinanceTimeframe(999)).toThrow("Timeframe '999' not supported");
    });

    test('should handle critical crypto timeframes correctly', () => {
      // The bug was specifically with 1h -> should convert to 60
      expect(TimeframeParser.toBinanceTimeframe('1h')).toBe('60');
      // Other common crypto timeframes
      expect(TimeframeParser.toBinanceTimeframe('4h')).toBe('240');
      expect(TimeframeParser.toBinanceTimeframe('1d')).toBe('D');
    });
  });
});
