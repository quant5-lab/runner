import { describe, it, expect } from 'vitest';
import { TradeDataFormatter, TradeTableRenderer } from '../../out/js/TradeTable.js';

describe('TradeTable Integration Tests', () => {
  describe('Complete Trade Lifecycle Workflow', () => {
    it('should format and render complete trade history from raw data', () => {
      const candlestickData = [
        { time: 1500000000, open: 100, high: 110, low: 95, close: 105 },
        { time: 1500003600, open: 105, high: 115, low: 100, close: 110 },
        { time: 1500007200, open: 110, high: 120, low: 105, close: 115 },
        { time: 1500010800, open: 115, high: 125, low: 110, close: 120 },
      ];

      const trades = [
        {
          entryId: 'TRADE_001',
          entryPrice: 100,
          entryBar: 0,
          entryTime: 1500000000,
          exitPrice: 110,
          exitBar: 1,
          exitTime: 1500003600,
          size: 1,
          profit: 10,
          direction: 'long',
        },
        {
          entryId: 'TRADE_002',
          entryPrice: 110,
          entryBar: 1,
          entryTime: 1500003600,
          exitPrice: 105,
          exitBar: 2,
          exitTime: 1500007200,
          size: 1,
          profit: -5,
          direction: 'short',
        },
        {
          status: 'open',
          entryId: 'TRADE_003',
          entryPrice: 115,
          entryBar: 3,
          entryTime: 1500010800,
          size: 1,
          direction: 'long',
        },
      ];

      const formatter = new TradeDataFormatter(candlestickData);
      const renderer = new TradeTableRenderer(formatter);
      const html = renderer.renderRows(trades, 120);

      expect(html).toContain('TRADE_001');
      expect(html).toContain('TRADE_002');
      expect(html).toContain('TRADE_003');
      expect(html).toContain('trade-long');
      expect(html).toContain('trade-short');
      expect(html).toContain('trade-profit-positive');
      expect(html).toContain('trade-profit-negative');
      expect(html).toContain('>Open<');

      const rowCount = (html.match(/<tr>/g) || []).length;
      expect(rowCount).toBe(3);
    });

    it('should correctly calculate and display unrealized P/L across multiple open positions', () => {
      const candlestickData = [
        { time: 1500000000, open: 100, high: 110, low: 95, close: 105 },
        { time: 1500003600, open: 105, high: 115, low: 100, close: 110 },
      ];

      const openTrades = [
        {
          status: 'open',
          entryPrice: 100,
          entryBar: 0,
          entryTime: 1500000000,
          size: 2,
          direction: 'long',
        },
        {
          status: 'open',
          entryPrice: 110,
          entryBar: 1,
          entryTime: 1500003600,
          size: 1,
          direction: 'short',
        },
      ];

      const currentPrice = 105;
      const formatter = new TradeDataFormatter(candlestickData);
      const renderer = new TradeTableRenderer(formatter);
      const html = renderer.renderRows(openTrades, currentPrice);

      expect(html).toContain('+$10.00');
      expect(html).toContain('+$5.00');
    });
  });

  describe('Real-World Scenario Simulations', () => {
    it('should handle day trading scenario with multiple intraday trades', () => {
      const candlestickData = Array.from({ length: 390 }, (_, i) => ({
        time: 1500000000 + i * 60,
        open: 100 + Math.sin(i / 10) * 5,
        high: 105 + Math.sin(i / 10) * 5,
        low: 95 + Math.sin(i / 10) * 5,
        close: 100 + Math.sin(i / 10) * 5,
      }));

      const trades = Array.from({ length: 20 }, (_, i) => ({
        entryId: `DAY_${i}`,
        entryPrice: 100 + i * 0.5,
        entryBar: i * 10,
        entryTime: 1500000000 + i * 600,
        exitPrice: 100 + i * 0.5 + (i % 2 === 0 ? 2 : -1),
        exitBar: i * 10 + 5,
        exitTime: 1500000000 + i * 600 + 300,
        size: 1,
        profit: i % 2 === 0 ? 2 : -1,
        direction: 'long',
      }));

      const formatter = new TradeDataFormatter(candlestickData);
      const renderer = new TradeTableRenderer(formatter);

      expect(() => renderer.renderRows(trades, null)).not.toThrow();

      const html = renderer.renderRows(trades, null);
      const rowCount = (html.match(/<tr>/g) || []).length;

      expect(rowCount).toBe(20);
    });

    it('should handle swing trading scenario with mixed timeframes', () => {
      const candlestickData = Array.from({ length: 100 }, (_, i) => ({
        time: 1500000000 + i * 86400,
        open: 100 + i * 2,
        high: 110 + i * 2,
        low: 95 + i * 2,
        close: 105 + i * 2,
      }));

      const trades = [
        {
          entryId: 'SWING_LONG',
          entryPrice: 100,
          entryBar: 0,
          entryTime: 1500000000,
          exitPrice: 200,
          exitBar: 50,
          exitTime: 1500000000 + 50 * 86400,
          size: 10,
          profit: 1000,
          direction: 'long',
        },
        {
          entryId: 'SWING_SHORT',
          entryPrice: 200,
          entryBar: 50,
          entryTime: 1500000000 + 50 * 86400,
          exitPrice: 150,
          exitBar: 75,
          exitTime: 1500000000 + 75 * 86400,
          size: 5,
          profit: 250,
          direction: 'short',
        },
      ];

      const formatter = new TradeDataFormatter(candlestickData);
      const renderer = new TradeTableRenderer(formatter);
      const html = renderer.renderRows(trades, null);

      expect(html).toContain('+$1000.00');
      expect(html).toContain('+$250.00');
    });

    it('should handle scalping scenario with rapid entries and exits', () => {
      const candlestickData = Array.from({ length: 1000 }, (_, i) => ({
        time: 1500000000 + i * 5,
        open: 100,
        high: 100.1,
        low: 99.9,
        close: 100,
      }));

      const trades = Array.from({ length: 100 }, (_, i) => ({
        entryId: `SCALP_${i}`,
        entryPrice: 100,
        entryBar: i * 10,
        entryTime: 1500000000 + i * 50,
        exitPrice: 100 + (i % 2 === 0 ? 0.05 : -0.03),
        exitBar: i * 10 + 1,
        exitTime: 1500000000 + i * 50 + 5,
        size: 100,
        profit: i % 2 === 0 ? 5 : -3,
        direction: 'long',
      }));

      const formatter = new TradeDataFormatter(candlestickData);
      const renderer = new TradeTableRenderer(formatter);

      expect(() => renderer.renderRows(trades, null)).not.toThrow();
    });
  });

  describe('Mixed Portfolio Scenarios', () => {
    it('should handle portfolio with both winners and losers', () => {
      const candlestickData = Array.from({ length: 20 }, (_, i) => ({
        time: 1500000000 + i * 3600,
        open: 100,
        high: 110,
        low: 95,
        close: 105,
      }));

      const winners = Array.from({ length: 7 }, (_, i) => ({
        entryId: `WIN_${i}`,
        entryPrice: 100,
        entryBar: i * 2,
        exitPrice: 110,
        exitBar: i * 2 + 1,
        size: 1,
        profit: 10,
        direction: 'long',
      }));

      const losers = Array.from({ length: 3 }, (_, i) => ({
        entryId: `LOSS_${i}`,
        entryPrice: 100,
        entryBar: (i + 7) * 2,
        exitPrice: 95,
        exitBar: (i + 7) * 2 + 1,
        size: 1,
        profit: -5,
        direction: 'long',
      }));

      const trades = [...winners, ...losers];
      const formatter = new TradeDataFormatter(candlestickData);
      const renderer = new TradeTableRenderer(formatter);
      const html = renderer.renderRows(trades, null);

      const positiveMatches = (html.match(/trade-profit-positive/g) || []).length;
      const negativeMatches = (html.match(/trade-profit-negative/g) || []).length;

      expect(positiveMatches).toBe(7);
      expect(negativeMatches).toBe(3);
    });

    it('should handle mixed open and closed positions in portfolio', () => {
      const candlestickData = Array.from({ length: 10 }, (_, i) => ({
        time: 1500000000 + i * 3600,
        open: 100 + i,
        high: 110 + i,
        low: 95 + i,
        close: 105 + i,
      }));

      const closedTrades = Array.from({ length: 5 }, (_, i) => ({
        entryId: `CLOSED_${i}`,
        entryPrice: 100 + i,
        entryBar: i,
        exitPrice: 110 + i,
        exitBar: i + 1,
        size: 1,
        profit: 10,
        direction: 'long',
      }));

      const openTrades = Array.from({ length: 3 }, (_, i) => ({
        status: 'open',
        entryId: `OPEN_${i}`,
        entryPrice: 105 + i,
        entryBar: i + 5,
        size: 1,
        direction: 'long',
      }));

      const trades = [...closedTrades, ...openTrades];
      const formatter = new TradeDataFormatter(candlestickData);
      const renderer = new TradeTableRenderer(formatter);
      const html = renderer.renderRows(trades, 115);

      const openLabels = (html.match(/>Open</g) || []).length;
      expect(openLabels).toBe(3);

      const rowCount = (html.match(/<tr>/g) || []).length;
      expect(rowCount).toBe(8);
    });
  });

  describe('Data Quality and Consistency', () => {
    it('should maintain data integrity across formatting and rendering pipeline', () => {
      const candlestickData = [
        { time: 1500000000, open: 100, high: 110, low: 95, close: 105 },
      ];

      const originalTrade = {
        entryId: 'INTEGRITY_TEST',
        entryPrice: 123.456,
        entryBar: 0,
        entryTime: 1500000000,
        exitPrice: 234.567,
        exitBar: 0,
        exitTime: 1500000000,
        size: 7.89,
        profit: 876.54,
        direction: 'long',
      };

      const formatter = new TradeDataFormatter(candlestickData);
      const formatted = formatter.formatTrade(originalTrade, 5, null);

      expect(formatted.number).toBe(6);
      expect(formatted.entryPrice).toBe('$123.46');
      expect(formatted.exitPrice).toBe('$234.57');
      expect(formatted.size).toBe('7.89');
      expect(formatted.profit).toBe('+$876.54');
      expect(formatted.entryId).toBe('INTEGRITY_TEST');
      expect(formatted.direction).toBe('long');
    });

    it('should produce consistent HTML output for identical trades', () => {
      const candlestickData = [
        { time: 1500000000, open: 100, high: 110, low: 95, close: 105 },
      ];

      const trade = {
        entryPrice: 100,
        entryBar: 0,
        exitPrice: 105,
        exitBar: 0,
        size: 1,
        profit: 5,
        direction: 'long',
      };

      const formatter = new TradeDataFormatter(candlestickData);
      const renderer = new TradeTableRenderer(formatter);

      const html1 = renderer.renderRows([trade], null);
      const html2 = renderer.renderRows([trade], null);

      expect(html1).toBe(html2);
    });

    it('should handle trades with timestamps across timezone boundaries', () => {
      const candlestickData = [
        { time: 1609459200, open: 100, high: 110, low: 95, close: 105 },
        { time: 1609545600, open: 105, high: 115, low: 100, close: 110 },
      ];

      const trades = [
        {
          entryTime: 1609459200,
          entryBar: 0,
          exitTime: 1609545600,
          exitBar: 1,
          entryPrice: 100,
          exitPrice: 110,
          size: 1,
          profit: 10,
          direction: 'long',
        },
      ];

      const formatter = new TradeDataFormatter(candlestickData);
      const renderer = new TradeTableRenderer(formatter);

      expect(() => renderer.renderRows(trades, null)).not.toThrow();

      const html = renderer.renderRows(trades, null);
      expect(html).toContain('2021');
    });
  });

  describe('Performance and Scalability', () => {
    it('should handle 1000 trades without performance degradation', () => {
      const candlestickData = Array.from({ length: 2000 }, (_, i) => ({
        time: 1500000000 + i * 60,
        open: 100,
        high: 110,
        low: 95,
        close: 105,
      }));

      const trades = Array.from({ length: 1000 }, (_, i) => ({
        entryId: `PERF_${i}`,
        entryPrice: 100,
        entryBar: i * 2,
        entryTime: 1500000000 + i * 120,
        exitPrice: 105,
        exitBar: i * 2 + 1,
        exitTime: 1500000000 + i * 120 + 60,
        size: 1,
        profit: 5,
        direction: 'long',
      }));

      const formatter = new TradeDataFormatter(candlestickData);
      const renderer = new TradeTableRenderer(formatter);

      const startTime = Date.now();
      const html = renderer.renderRows(trades, null);
      const endTime = Date.now();

      expect(endTime - startTime).toBeLessThan(1000);
      expect(html.length).toBeGreaterThan(100000);
    });

    it('should efficiently handle large candlestick datasets', () => {
      const candlestickData = Array.from({ length: 10000 }, (_, i) => ({
        time: 1500000000 + i * 60,
        open: 100 + Math.sin(i / 100) * 10,
        high: 110 + Math.sin(i / 100) * 10,
        low: 95 + Math.sin(i / 100) * 10,
        close: 105 + Math.sin(i / 100) * 10,
      }));

      const trades = Array.from({ length: 50 }, (_, i) => ({
        entryBar: i * 200,
        exitBar: i * 200 + 100,
        entryPrice: 100,
        exitPrice: 105,
        size: 1,
        profit: 5,
        direction: 'long',
      }));

      const formatter = new TradeDataFormatter(candlestickData);
      const renderer = new TradeTableRenderer(formatter);

      const startTime = Date.now();
      const html = renderer.renderRows(trades, null);
      const endTime = Date.now();

      expect(endTime - startTime).toBeLessThan(500);
      expect((html.match(/<tr>/g) || []).length).toBe(50);
    });
  });

  describe('Backward Compatibility', () => {
    it('should handle legacy trade format without new fields', () => {
      const candlestickData = [
        { time: 1500000000, open: 100, high: 110, low: 95, close: 105 },
      ];

      const legacyTrade = {
        entryPrice: 100,
        exitPrice: 105,
        size: 1,
        profit: 5,
        direction: 'long',
      };

      const formatter = new TradeDataFormatter(candlestickData);
      const renderer = new TradeTableRenderer(formatter);

      expect(() => renderer.renderRows([legacyTrade], null)).not.toThrow();

      const html = renderer.renderRows([legacyTrade], null);
      expect(html).toContain('N/A');
    });

    it('should handle trades with entryBar but no entryTime', () => {
      const candlestickData = [
        { time: 1500000000, open: 100, high: 110, low: 95, close: 105 },
        { time: 1500003600, open: 105, high: 115, low: 100, close: 110 },
      ];

      const trade = {
        entryBar: 0,
        exitBar: 1,
        entryPrice: 100,
        exitPrice: 110,
        size: 1,
        profit: 10,
        direction: 'long',
      };

      const formatter = new TradeDataFormatter(candlestickData);
      const renderer = new TradeTableRenderer(formatter);
      const html = renderer.renderRows([trade], null);

      expect(html).toContain('2017');
      expect(html).toContain('N/A');
    });

    it('should handle trades with entryTime but no entryBar', () => {
      const candlestickData = [
        { time: 1500000000, open: 100, high: 110, low: 95, close: 105 },
      ];

      const trade = {
        entryTime: 1500000000,
        exitTime: 1500003600,
        entryPrice: 100,
        exitPrice: 110,
        size: 1,
        profit: 10,
        direction: 'long',
      };

      const formatter = new TradeDataFormatter(candlestickData);
      const renderer = new TradeTableRenderer(formatter);
      const html = renderer.renderRows([trade], null);

      expect(html).toContain('2017');
      expect(html).toContain('N/A');
    });
  });
});
