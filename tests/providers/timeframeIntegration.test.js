import { describe, test, expect } from 'vitest';
import { MoexProvider } from '../../src/providers/MoexProvider.js';
import { YahooFinanceProvider } from '../../src/providers/YahooFinanceProvider.js';

describe('Provider Timeframe Integration Tests', () => {
  describe('MoexProvider timeframe conversion', () => {
    const provider = new MoexProvider(console);

    test('should convert string timeframes correctly', () => {
      expect(provider.convertTimeframe('10m')).toBe('10');
      expect(provider.convertTimeframe('1h')).toBe('60');
      expect(provider.convertTimeframe('1d')).toBe('24');
    });

    test('should convert numeric timeframes correctly', () => {
      expect(provider.convertTimeframe(10)).toBe('10');
      expect(provider.convertTimeframe(60)).toBe('60');
      expect(provider.convertTimeframe(1440)).toBe('24');
    });

    test('should convert letter timeframes correctly', () => {
      expect(provider.convertTimeframe('D')).toBe('24');
      expect(provider.convertTimeframe('W')).toBe('7');
      expect(provider.convertTimeframe('M')).toBe('31');
    });

    test('REGRESSION: critical timeframe bug prevention', () => {
      /* These were the failing cases that caused the bug */
      expect(() => provider.convertTimeframe('15m')).toThrow("Timeframe '15m' not supported");
      expect(provider.convertTimeframe('1h')).toBe('60'); // This is supported

      /* Verify unsupported timeframes throw errors instead of fallback to daily */
      expect(() => provider.convertTimeframe('15m')).toThrow("Timeframe '15m' not supported");
      expect(() => provider.convertTimeframe('5m')).toThrow("Timeframe '5m' not supported");
    });
  });

  describe('YahooFinanceProvider timeframe conversion', () => {
    const provider = new YahooFinanceProvider(console);

    test('should convert string timeframes correctly', () => {
      expect(provider.convertTimeframe('15m')).toBe('15m');
      expect(provider.convertTimeframe('1h')).toBe('1h');
      expect(provider.convertTimeframe('1d')).toBe('1d');
    });

    test('should convert numeric timeframes correctly', () => {
      expect(provider.convertTimeframe(15)).toBe('15m');
      expect(provider.convertTimeframe(60)).toBe('1h');
      expect(provider.convertTimeframe(1440)).toBe('1d');
    });

    test('should convert letter timeframes correctly', () => {
      expect(provider.convertTimeframe('D')).toBe('1d');
      expect(provider.convertTimeframe('W')).toBe('1wk');
      expect(provider.convertTimeframe('M')).toBe('1mo');
    });

    test('REGRESSION: critical timeframe bug prevention', () => {
      /* These were the failing cases that caused the bug */
      expect(provider.convertTimeframe('15m')).toBe('15m'); // NOT '1d'
      expect(provider.convertTimeframe('1h')).toBe('1h'); // NOT '1d'

      /* Verify they don't fallback to daily */
      expect(provider.convertTimeframe('15m')).not.toBe('1d');
      expect(provider.convertTimeframe('1h')).not.toBe('1d');
    });
  });

  describe('Cross-provider timeframe consistency', () => {
    const moexProvider = new MoexProvider(console);
    const yahooProvider = new YahooFinanceProvider(console);

    test('should handle common timeframes consistently', () => {
      const testCases = [
        { input: '1m', moexExpected: '1', yahooExpected: '1m' },
        { input: '1h', moexExpected: '60', yahooExpected: '1h' },
        { input: '1d', moexExpected: '24', yahooExpected: '1d' },
        { input: 1, moexExpected: '1', yahooExpected: '1m' },
        { input: 60, moexExpected: '60', yahooExpected: '1h' },
        { input: 'D', moexExpected: '24', yahooExpected: '1d' },
      ];

      for (const { input, moexExpected, yahooExpected } of testCases) {
        expect(moexProvider.convertTimeframe(input)).toBe(moexExpected);
        expect(yahooProvider.convertTimeframe(input)).toBe(yahooExpected);
      }
    });

    test('should not return daily fallback for valid timeframes', () => {
      const validTimeframes = ['1m', '1h', '1d']; // Only common supported timeframes

      for (const tf of validTimeframes) {
        const moexResult = moexProvider.convertTimeframe(tf);
        const yahooResult = yahooProvider.convertTimeframe(tf);

        /* For non-daily timeframes, should not fallback to daily */
        if (tf !== '1d' && tf !== 'D') {
          expect(moexResult).not.toBe('24');
          expect(yahooResult).not.toBe('1d');
        }
      }
    });
  });

  describe('Date range calculation integration', () => {
    const moexProvider = new MoexProvider(console);
    const yahooProvider = new YahooFinanceProvider(console);

    test('MOEX getTimeframeDays should calculate correct date ranges', () => {
      // Test the actual date range calculation logic with MOEX trading hours
      const testCases = [
        {
          timeframe: '1h',
          expectedCalc: 60 / 540 * 1.4, // 60 min ÷ 540 trading min/day × 1.4 weekend buffer
          description: '1 hour accounting for ~9 trading hours/day + weekend buffer',
        },
        {
          timeframe: '15m',
          expectedCalc: 15 / 540 * 1.4, // 15 min ÷ 540 trading min/day × 1.4 weekend buffer
          description: '15 minutes accounting for trading hours + weekend buffer',
        },
        {
          timeframe: '1d',
          expectedCalc: 1,
          description: '1 day = 1 calendar day (daily+ use calendar days)',
        },
        {
          timeframe: 'D',
          expectedCalc: 1,
          description: 'Daily = 1 calendar day',
        },
        {
          timeframe: 'W',
          expectedCalc: 7,
          description: 'Weekly = 7 calendar days',
        },
      ];

      testCases.forEach(({ timeframe, expectedCalc, description }) => {
        const actualDays = moexProvider.getTimeframeDays(timeframe);
        expect(actualDays).toBeCloseTo(expectedCalc, 3); // Use toBeCloseTo for floating point comparison
      });
    });

    test('Yahoo getDateRange should return appropriate ranges', () => {
      const testCases = [
        { timeframe: '1h', expectedRange: '1mo', description: 'Hourly data needs 1 month for ~130 points' },
        { timeframe: '15m', expectedRange: '5d', description: '15-minute data needs 5 days based on dynamic logic' },
        { timeframe: '1d', expectedRange: '6mo', description: 'Daily data needs 6 months for 100 points' },
        { timeframe: 'D', expectedRange: '6mo', description: 'Daily (letter) data needs 6 months' },
        { timeframe: 'W', expectedRange: '2y', description: 'Weekly data needs 2 years for 100 points' },
      ];

      testCases.forEach(({ timeframe, expectedRange, description }) => {
        const actualRange = yahooProvider.getDateRange(100, timeframe);
        expect(actualRange).toBe(expectedRange);
      });
    });

    test('REGRESSION: date range bug fix verification', () => {
      // Test that the original date range bug is fixed

      // MOEX: 1h timeframe should calculate for trading hours, not full days
      const moexHourlyDays = moexProvider.getTimeframeDays('1h');
      const expectedMoexDays = 60 / 540 * 1.4; // ~0.156 days accounting for trading hours + buffer
      expect(moexHourlyDays).toBeCloseTo(expectedMoexDays, 3);
      expect(moexHourlyDays).toBeGreaterThan(0.1); // Should be reasonable fraction
      expect(moexHourlyDays).toBeLessThan(1); // Should be less than full day

      // Yahoo: 1h timeframe should return '1mo', not '5d' insufficient range
      const yahooHourlyRange = yahooProvider.getDateRange(100, '1h');
      expect(yahooHourlyRange).toBe('1mo');
      expect(yahooHourlyRange).not.toBe('5d'); // Should NOT be the old insufficient range

      // Verify the calculation chain works end-to-end
      // For 100 bars of 1h data with MOEX trading hours:
      // 100 × (60/540 × 1.4) = 100 × ~0.156 = ~15.6 days back
      const expectedDaysBack = Math.ceil(100 * moexHourlyDays);
      expect(expectedDaysBack).toBeGreaterThan(10); // Should be ~15-16 days, not 5 days
      expect(expectedDaysBack).toBeLessThan(25); // Reasonable upper bound
    });

    test('TimeframeParser integration completeness', () => {
      // Verify that both providers handle timeframes according to their capabilities
      const moexSupportedTimeframes = [
        // MOEX supported formats based on evidence
        '1m', '10m', '1h', '1d',
        1, 10, 60, 1440,
        'D', 'W', 'M',
      ];

      const moexUnsupportedTimeframes = [
        '5m', '15m', '30m', '4h',
        5, 15, 30, 240,
      ];

      // MOEX should handle supported timeframes without errors
      moexSupportedTimeframes.forEach(tf => {
        expect(() => moexProvider.convertTimeframe(tf)).not.toThrow();
        expect(() => moexProvider.getTimeframeDays(tf)).not.toThrow();
        expect(moexProvider.convertTimeframe(tf)).toBeTruthy();
        expect(moexProvider.getTimeframeDays(tf)).toBeGreaterThan(0);
      });

      // MOEX should throw TimeframeError for unsupported timeframes
      moexUnsupportedTimeframes.forEach(tf => {
        expect(() => moexProvider.convertTimeframe(tf)).toThrow('not supported');
      });

      // Yahoo should handle its supported formats without errors
      const yahooSupportedTimeframes = [
        '1m', '2m', '5m', '15m', '30m', '1h', '90m', '1d',
        1, 2, 5, 15, 30, 60, 90, 1440,
        'D', 'W', 'M',
      ];

      yahooSupportedTimeframes.forEach(tf => {
        expect(() => yahooProvider.convertTimeframe(tf)).not.toThrow();
        expect(() => yahooProvider.getDateRange(100, tf)).not.toThrow();
        expect(yahooProvider.convertTimeframe(tf)).toBeTruthy();
        expect(yahooProvider.getDateRange(100, tf)).toBeTruthy();
      });
    });
  });
});
