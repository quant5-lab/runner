import { describe, it, expect, beforeEach } from 'vitest';
import { TradeDataFormatter } from '../../out/js/TradeTable.js';

describe('TradeDataFormatter', () => {
  let formatter;
  let candlestickData;

  beforeEach(() => {
    candlestickData = [
      { time: 1500000000, open: 100, high: 110, low: 95, close: 105 },
      { time: 1500003600, open: 105, high: 115, low: 100, close: 110 },
      { time: 1500007200, open: 110, high: 120, low: 105, close: 115 },
    ];
    formatter = new TradeDataFormatter(candlestickData);
  });

  describe('Timestamp Conversion', () => {
    it('should convert Unix seconds to milliseconds for date creation', () => {
      const unixSeconds = 1500000000;
      const result = formatter.formatDate(unixSeconds * 1000);
      
      expect(result).toContain('2017');
      expect(result).not.toContain('1970');
    });

    it('should handle epoch timestamp (0) correctly', () => {
      const result = formatter.formatDate(0);
      
      expect(result).toContain('1970');
      expect(result).toContain('Jan');
    });

    it('should handle future timestamps correctly', () => {
      const futureTimestamp = 2000000000 * 1000;
      const result = formatter.formatDate(futureTimestamp);
      
      expect(result).toContain('2033');
    });

    it('should handle very large timestamps without crashing', () => {
      const largeTimestamp = 9999999999 * 1000;
      
      expect(() => formatter.formatDate(largeTimestamp)).not.toThrow();
    });
  });

  describe('Date Formatting', () => {
    it('should format dates with consistent structure', () => {
      const timestamp = 1500000000 * 1000;
      const result = formatter.formatDate(timestamp);
      
      expect(result).toMatch(/^[A-Z][a-z]{2} \d{1,2}, \d{4}, \d{2}:\d{2} [AP]M$/);
    });

    it('should format midnight times correctly', () => {
      const midnight = new Date('2023-01-15T00:00:00Z').getTime();
      const result = formatter.formatDate(midnight);
      
      expect(result).toContain('2023');
      expect(result).toMatch(/\d{2}:\d{2} [AP]M/);
    });

    it('should format noon times correctly', () => {
      const noon = new Date('2023-01-15T12:00:00Z').getTime();
      const result = formatter.formatDate(noon);
      
      expect(result).toContain('2023');
      expect(result).toMatch(/\d{2}:\d{2} [AP]M/);
    });
  });

  describe('Price Formatting', () => {
    it('should format whole numbers with two decimals', () => {
      expect(formatter.formatPrice(100)).toBe('$100.00');
    });

    it('should format decimals with two decimal places', () => {
      expect(formatter.formatPrice(123.456)).toBe('$123.46');
    });

    it('should format zero correctly', () => {
      expect(formatter.formatPrice(0)).toBe('$0.00');
    });

    it('should format very small numbers correctly', () => {
      expect(formatter.formatPrice(0.001)).toBe('$0.00');
    });

    it('should format very large numbers correctly', () => {
      expect(formatter.formatPrice(1234567.89)).toBe('$1234567.89');
    });

    it('should handle negative prices', () => {
      expect(formatter.formatPrice(-100.50)).toBe('$-100.50');
    });
  });

  describe('Profit Formatting', () => {
    it('should format positive profit with plus sign', () => {
      expect(formatter.formatProfit(50.75)).toBe('+$50.75');
    });

    it('should format negative profit with minus sign', () => {
      expect(formatter.formatProfit(-25.50)).toBe('-$25.50');
    });

    it('should format zero profit with plus sign', () => {
      expect(formatter.formatProfit(0)).toBe('+$0.00');
    });

    it('should handle very small profits', () => {
      expect(formatter.formatProfit(0.01)).toBe('+$0.01');
    });

    it('should handle very large profits', () => {
      expect(formatter.formatProfit(10000.99)).toBe('+$10000.99');
    });
  });

  describe('Trade Date Extraction - Entry', () => {
    it('should extract entry date from entryTime field', () => {
      const trade = {
        entryTime: 1500000000,
        entryBar: 0,
      };

      const result = formatter.getTradeDate(trade, true);

      expect(result).toContain('2017');
      expect(result).not.toContain('N/A');
    });

    it('should extract entry date from candlestick data when entryTime missing', () => {
      const trade = {
        entryBar: 1,
      };

      const result = formatter.getTradeDate(trade, true);

      expect(result).toContain('2017');
      expect(result).not.toContain('N/A');
    });

    it('should return N/A when both entryTime and entryBar are unavailable', () => {
      const trade = {
        entryBar: 999,
      };

      const result = formatter.getTradeDate(trade, true);

      expect(result).toBe('N/A');
    });

    it('should handle entryBar at boundary (0)', () => {
      const trade = {
        entryBar: 0,
      };

      const result = formatter.getTradeDate(trade, true);

      expect(result).not.toBe('N/A');
      expect(result).toContain('2017');
    });

    it('should handle entryBar at boundary (last index)', () => {
      const trade = {
        entryBar: candlestickData.length - 1,
      };

      const result = formatter.getTradeDate(trade, true);

      expect(result).not.toBe('N/A');
    });

    it('should prefer entryTime over entryBar when both available', () => {
      const trade = {
        entryTime: 1600000000,
        entryBar: 0,
      };

      const result = formatter.getTradeDate(trade, true);

      expect(result).toContain('2020');
      expect(result).not.toContain('2017');
    });
  });

  describe('Trade Date Extraction - Exit', () => {
    it('should extract exit date from exitTime field', () => {
      const trade = {
        exitTime: 1500010000,
        exitBar: 2,
      };

      const result = formatter.getTradeDate(trade, false);

      expect(result).toContain('2017');
      expect(result).not.toContain('N/A');
    });

    it('should extract exit date from candlestick data when exitTime missing', () => {
      const trade = {
        exitBar: 2,
      };

      const result = formatter.getTradeDate(trade, false);

      expect(result).toContain('2017');
    });

    it('should return N/A when exit data unavailable', () => {
      const trade = {
        exitBar: 999,
      };

      const result = formatter.getTradeDate(trade, false);

      expect(result).toBe('N/A');
    });

    it('should handle same bar entry and exit', () => {
      const trade = {
        entryBar: 1,
        exitBar: 1,
      };

      const entryDate = formatter.getTradeDate(trade, true);
      const exitDate = formatter.getTradeDate(trade, false);

      expect(entryDate).toBe(exitDate);
    });
  });

  describe('Unrealized Profit Calculation', () => {
    it('should calculate unrealized profit for long position', () => {
      const trade = {
        status: 'open',
        direction: 'long',
        entryPrice: 100,
        size: 2,
      };
      const currentPrice = 110;

      const profit = formatter.calculateUnrealizedProfit(trade, currentPrice);

      expect(profit).toBe(20);
    });

    it('should calculate unrealized loss for long position', () => {
      const trade = {
        status: 'open',
        direction: 'long',
        entryPrice: 100,
        size: 2,
      };
      const currentPrice = 95;

      const profit = formatter.calculateUnrealizedProfit(trade, currentPrice);

      expect(profit).toBe(-10);
    });

    it('should calculate unrealized profit for short position', () => {
      const trade = {
        status: 'open',
        direction: 'short',
        entryPrice: 100,
        size: 2,
      };
      const currentPrice = 90;

      const profit = formatter.calculateUnrealizedProfit(trade, currentPrice);

      expect(profit).toBe(20);
    });

    it('should calculate unrealized loss for short position', () => {
      const trade = {
        status: 'open',
        direction: 'short',
        entryPrice: 100,
        size: 2,
      };
      const currentPrice = 110;

      const profit = formatter.calculateUnrealizedProfit(trade, currentPrice);

      expect(profit).toBe(-20);
    });

    it('should return zero for closed trade', () => {
      const trade = {
        status: 'closed',
        direction: 'long',
        entryPrice: 100,
        size: 2,
      };
      const currentPrice = 110;

      const profit = formatter.calculateUnrealizedProfit(trade, currentPrice);

      expect(profit).toBe(0);
    });

    it('should return zero when currentPrice is null', () => {
      const trade = {
        status: 'open',
        direction: 'long',
        entryPrice: 100,
        size: 2,
      };

      const profit = formatter.calculateUnrealizedProfit(trade, null);

      expect(profit).toBe(0);
    });

    it('should handle fractional position sizes', () => {
      const trade = {
        status: 'open',
        direction: 'long',
        entryPrice: 100,
        size: 0.5,
      };
      const currentPrice = 120;

      const profit = formatter.calculateUnrealizedProfit(trade, currentPrice);

      expect(profit).toBe(10);
    });

    it('should handle zero price movement', () => {
      const trade = {
        status: 'open',
        direction: 'long',
        entryPrice: 100,
        size: 2,
      };
      const currentPrice = 100;

      const profit = formatter.calculateUnrealizedProfit(trade, currentPrice);

      expect(profit).toBe(0);
    });
  });

  describe('Complete Trade Formatting - Closed Trades', () => {
    it('should format closed long trade with all fields', () => {
      const trade = {
        entryId: 'LONG_001',
        entryPrice: 100.50,
        entryBar: 0,
        entryTime: 1500000000,
        exitPrice: 110.75,
        exitBar: 2,
        exitTime: 1500007200,
        size: 2.5,
        profit: 25.625,
        direction: 'long',
      };

      const formatted = formatter.formatTrade(trade, 0, null);

      expect(formatted.number).toBe(1);
      expect(formatted.entryDate).toContain('2017');
      expect(formatted.entryBar).toBe(0);
      expect(formatted.exitDate).toContain('2017');
      expect(formatted.exitBar).toBe(2);
      expect(formatted.direction).toBe('long');
      expect(formatted.entryPrice).toBe('$100.50');
      expect(formatted.exitPrice).toBe('$110.75');
      expect(formatted.size).toBe('2.50');
      expect(formatted.profit).toBe('+$25.63');
      expect(formatted.entryId).toBe('LONG_001');
      expect(formatted.isOpen).toBe(false);
    });

    it('should format closed short trade with loss', () => {
      const trade = {
        entryId: 'SHORT_001',
        entryPrice: 100,
        entryBar: 1,
        entryTime: 1500003600,
        exitPrice: 110,
        exitBar: 2,
        exitTime: 1500007200,
        size: 1,
        profit: -10,
        direction: 'short',
      };

      const formatted = formatter.formatTrade(trade, 5, null);

      expect(formatted.number).toBe(6);
      expect(formatted.profit).toBe('-$10.00');
      expect(formatted.profitRaw).toBe(-10);
      expect(formatted.direction).toBe('short');
    });

    it('should handle missing entryId field', () => {
      const trade = {
        entryPrice: 100,
        entryBar: 0,
        exitPrice: 105,
        exitBar: 1,
        size: 1,
        profit: 5,
        direction: 'long',
      };

      const formatted = formatter.formatTrade(trade, 0, null);

      expect(formatted.entryId).toBe('N/A');
    });

    it('should handle alternative entryID capitalization', () => {
      const trade = {
        entryID: 'TEST_ID',
        entryPrice: 100,
        entryBar: 0,
        exitPrice: 105,
        exitBar: 1,
        size: 1,
        profit: 5,
        direction: 'long',
      };

      const formatted = formatter.formatTrade(trade, 0, null);

      expect(formatted.entryId).toBe('TEST_ID');
    });

    it('should handle undefined bar numbers', () => {
      const trade = {
        entryPrice: 100,
        entryTime: 1500000000,
        exitPrice: 105,
        exitTime: 1500007200,
        size: 1,
        profit: 5,
        direction: 'long',
      };

      const formatted = formatter.formatTrade(trade, 0, null);

      expect(formatted.entryBar).toBe('N/A');
      expect(formatted.exitBar).toBe('N/A');
    });

    it('should handle zero profit trades', () => {
      const trade = {
        entryPrice: 100,
        entryBar: 0,
        exitPrice: 100,
        exitBar: 1,
        size: 1,
        profit: 0,
        direction: 'long',
      };

      const formatted = formatter.formatTrade(trade, 0, null);

      expect(formatted.profit).toBe('+$0.00');
      expect(formatted.profitRaw).toBe(0);
    });
  });

  describe('Complete Trade Formatting - Open Trades', () => {
    it('should format open long trade with unrealized profit', () => {
      const trade = {
        status: 'open',
        entryId: 'OPEN_LONG',
        entryPrice: 100,
        entryBar: 0,
        entryTime: 1500000000,
        size: 2,
        direction: 'long',
      };
      const currentPrice = 110;

      const formatted = formatter.formatTrade(trade, 0, currentPrice);

      expect(formatted.exitDate).toBe('Open');
      expect(formatted.exitBar).toBe('-');
      expect(formatted.exitPrice).toBe('$110.00');
      expect(formatted.profit).toBe('+$20.00');
      expect(formatted.profitRaw).toBe(20);
      expect(formatted.isOpen).toBe(true);
    });

    it('should format open short trade with unrealized loss', () => {
      const trade = {
        status: 'open',
        entryId: 'OPEN_SHORT',
        entryPrice: 100,
        entryBar: 1,
        entryTime: 1500003600,
        size: 1,
        direction: 'short',
      };
      const currentPrice = 110;

      const formatted = formatter.formatTrade(trade, 2, currentPrice);

      expect(formatted.number).toBe(3);
      expect(formatted.exitDate).toBe('Open');
      expect(formatted.profit).toBe('-$10.00');
      expect(formatted.profitRaw).toBe(-10);
    });

    it('should handle open trade without current price gracefully', () => {
      const trade = {
        status: 'open',
        entryPrice: 100,
        entryBar: 0,
        size: 1,
        direction: 'long',
      };

      const formatted = formatter.formatTrade(trade, 0, null);

      expect(formatted.profit).toBe('+$0.00');
      expect(formatted.profitRaw).toBe(0);
      expect(formatted.exitPrice).toBe('$100.00');
    });

    it('should handle open trade with zero bar number', () => {
      const trade = {
        status: 'open',
        entryBar: 0,
        entryPrice: 100,
        size: 1,
        direction: 'long',
      };

      const formatted = formatter.formatTrade(trade, 0, 105);

      expect(formatted.entryBar).toBe(0);
      expect(formatted.exitBar).toBe('-');
    });
  });

  describe('Edge Cases and Boundary Conditions', () => {
    it('should handle empty candlestick data', () => {
      const emptyFormatter = new TradeDataFormatter([]);
      const trade = {
        entryBar: 0,
        direction: 'long',
        entryPrice: 100,
        size: 1,
      };

      const result = emptyFormatter.getTradeDate(trade, true);

      expect(result).toBe('N/A');
    });

    it('should handle null candlestick data', () => {
      const nullFormatter = new TradeDataFormatter(null);
      const trade = {
        entryBar: 0,
        direction: 'long',
        entryPrice: 100,
        size: 1,
      };

      expect(() => nullFormatter.getTradeDate(trade, true)).not.toThrow();
    });

    it('should handle undefined candlestick data', () => {
      const undefinedFormatter = new TradeDataFormatter(undefined);
      const trade = {
        entryBar: 0,
        direction: 'long',
        entryPrice: 100,
        size: 1,
      };

      expect(() => undefinedFormatter.getTradeDate(trade, true)).not.toThrow();
    });

    it('should handle negative bar numbers gracefully', () => {
      const trade = {
        entryBar: -1,
        direction: 'long',
        entryPrice: 100,
        size: 1,
      };

      const result = formatter.getTradeDate(trade, true);

      expect(result).toBe('N/A');
    });

    it('should handle fractional bar numbers by accessing array element', () => {
      const trade = {
        entryBar: 1.5,
        direction: 'long',
        entryPrice: 100,
        size: 1,
      };

      const result = formatter.getTradeDate(trade, true);

      expect(result).toBe('N/A');
    });

    it('should handle very long entry IDs', () => {
      const trade = {
        entryId: 'A'.repeat(1000),
        entryPrice: 100,
        entryBar: 0,
        exitPrice: 105,
        exitBar: 1,
        size: 1,
        profit: 5,
        direction: 'long',
      };

      const formatted = formatter.formatTrade(trade, 0, null);

      expect(formatted.entryId).toHaveLength(1000);
    });

    it('should handle special characters in entry ID', () => {
      const trade = {
        entryId: 'ID<>&"\'',
        entryPrice: 100,
        entryBar: 0,
        exitPrice: 105,
        exitBar: 1,
        size: 1,
        profit: 5,
        direction: 'long',
      };

      const formatted = formatter.formatTrade(trade, 0, null);

      expect(formatted.entryId).toBe('ID<>&"\'');
    });

    it('should handle extreme profit values', () => {
      const trade = {
        entryPrice: 1,
        entryBar: 0,
        exitPrice: 1000000,
        exitBar: 1,
        size: 100,
        profit: 99999900,
        direction: 'long',
      };

      const formatted = formatter.formatTrade(trade, 0, null);

      expect(formatted.profit).toContain('99999900');
    });

    it('should maintain precision for small price differences', () => {
      const trade = {
        entryPrice: 100.001,
        entryBar: 0,
        exitPrice: 100.002,
        exitBar: 1,
        size: 1000,
        profit: 1,
        direction: 'long',
      };

      const formatted = formatter.formatTrade(trade, 0, null);

      expect(formatted.profit).toBe('+$1.00');
    });
  });
});
