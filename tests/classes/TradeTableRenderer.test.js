import { describe, it, expect, beforeEach } from 'vitest';
import { TradeDataFormatter, TradeTableRenderer } from '../../out/js/TradeTable.js';

describe('TradeTableRenderer', () => {
  let formatter;
  let renderer;
  let candlestickData;

  beforeEach(() => {
    candlestickData = [
      { time: 1500000000, open: 100, high: 110, low: 95, close: 105 },
      { time: 1500003600, open: 105, high: 115, low: 100, close: 110 },
      { time: 1500007200, open: 110, high: 120, low: 105, close: 115 },
    ];
    formatter = new TradeDataFormatter(candlestickData);
    renderer = new TradeTableRenderer(formatter);
  });

  describe('HTML Structure Generation', () => {
    it('should generate valid HTML table rows', () => {
      const trades = [
        {
          entryId: 'TEST_001',
          entryPrice: 100,
          entryBar: 0,
          entryTime: 1500000000,
          exitPrice: 105,
          exitBar: 1,
          exitTime: 1500003600,
          size: 1,
          profit: 5,
          direction: 'long',
        },
      ];

      const html = renderer.renderRows(trades, null);

      expect(html).toContain('<tr>');
      expect(html).toContain('</tr>');
      expect(html).toContain('<td>');
      expect(html).toContain('</td>');
    });

    it('should generate correct number of columns (11)', () => {
      const trades = [
        {
          entryPrice: 100,
          entryBar: 0,
          exitPrice: 105,
          exitBar: 1,
          size: 1,
          profit: 5,
          direction: 'long',
        },
      ];

      const html = renderer.renderRows(trades, null);
      const tdCount = (html.match(/<td/g) || []).length;

      expect(tdCount).toBe(11);
    });

    it('should generate one row per trade', () => {
      const trades = [
        {
          entryPrice: 100,
          entryBar: 0,
          exitPrice: 105,
          exitBar: 1,
          size: 1,
          profit: 5,
          direction: 'long',
        },
        {
          entryPrice: 110,
          entryBar: 1,
          exitPrice: 115,
          exitBar: 2,
          size: 1,
          profit: 5,
          direction: 'long',
        },
      ];

      const html = renderer.renderRows(trades, null);
      const rowCount = (html.match(/<tr>/g) || []).length;

      expect(rowCount).toBe(2);
    });

    it('should handle empty trade array', () => {
      const html = renderer.renderRows([], null);

      expect(html).toBe('');
    });

    it('should generate sequential trade numbers', () => {
      const trades = Array.from({ length: 5 }, (_, i) => ({
        entryPrice: 100 + i,
        entryBar: i,
        exitPrice: 105 + i,
        exitBar: i + 1,
        size: 1,
        profit: 5,
        direction: 'long',
      }));

      const html = renderer.renderRows(trades, null);

      expect(html).toContain('>1<');
      expect(html).toContain('>2<');
      expect(html).toContain('>3<');
      expect(html).toContain('>4<');
      expect(html).toContain('>5<');
    });
  });

  describe('CSS Class Application', () => {
    it('should apply trade-long class for long trades', () => {
      const trades = [
        {
          entryPrice: 100,
          entryBar: 0,
          exitPrice: 105,
          exitBar: 1,
          size: 1,
          profit: 5,
          direction: 'long',
        },
      ];

      const html = renderer.renderRows(trades, null);

      expect(html).toContain('class="trade-long"');
    });

    it('should apply trade-short class for short trades', () => {
      const trades = [
        {
          entryPrice: 100,
          entryBar: 0,
          exitPrice: 95,
          exitBar: 1,
          size: 1,
          profit: 5,
          direction: 'short',
        },
      ];

      const html = renderer.renderRows(trades, null);

      expect(html).toContain('class="trade-short"');
    });

    it('should apply trade-profit-positive class for profitable trades', () => {
      const trades = [
        {
          entryPrice: 100,
          entryBar: 0,
          exitPrice: 110,
          exitBar: 1,
          size: 1,
          profit: 10,
          direction: 'long',
        },
      ];

      const html = renderer.renderRows(trades, null);

      expect(html).toContain('class="trade-profit-positive"');
    });

    it('should apply trade-profit-negative class for losing trades', () => {
      const trades = [
        {
          entryPrice: 100,
          entryBar: 0,
          exitPrice: 90,
          exitBar: 1,
          size: 1,
          profit: -10,
          direction: 'long',
        },
      ];

      const html = renderer.renderRows(trades, null);

      expect(html).toContain('class="trade-profit-negative"');
    });

    it('should apply trade-profit-positive class for zero profit', () => {
      const trades = [
        {
          entryPrice: 100,
          entryBar: 0,
          exitPrice: 100,
          exitBar: 1,
          size: 1,
          profit: 0,
          direction: 'long',
        },
      ];

      const html = renderer.renderRows(trades, null);

      expect(html).toContain('class="trade-profit-positive"');
    });

    it('should apply correct classes for open trades with unrealized profit', () => {
      const trades = [
        {
          status: 'open',
          entryPrice: 100,
          entryBar: 0,
          size: 1,
          direction: 'long',
        },
      ];
      const currentPrice = 110;

      const html = renderer.renderRows(trades, currentPrice);

      expect(html).toContain('class="trade-long"');
      expect(html).toContain('class="trade-profit-positive"');
    });

    it('should apply correct classes for open trades with unrealized loss', () => {
      const trades = [
        {
          status: 'open',
          entryPrice: 100,
          entryBar: 0,
          size: 1,
          direction: 'long',
        },
      ];
      const currentPrice = 90;

      const html = renderer.renderRows(trades, currentPrice);

      expect(html).toContain('class="trade-profit-negative"');
    });
  });

  describe('Column Content Rendering', () => {
    it('should render all required columns in correct order', () => {
      const trades = [
        {
          entryId: 'TEST_ID',
          entryPrice: 100.50,
          entryBar: 5,
          entryTime: 1500000000,
          exitPrice: 110.75,
          exitBar: 10,
          exitTime: 1500003600,
          size: 2.5,
          profit: 25.625,
          direction: 'long',
        },
      ];

      const html = renderer.renderRows(trades, null);

      expect(html).toContain('>1<');
      expect(html).toContain('2017');
      expect(html).toContain('>5<');
      expect(html).toContain('LONG');
      expect(html).toContain('$100.50');
      expect(html).toContain('>10<');
      expect(html).toContain('$110.75');
      expect(html).toContain('2.50');
      expect(html).toContain('+$25.63');
      expect(html).toContain('TEST_ID');
    });

    it('should render open trades with "Open" exit label', () => {
      const trades = [
        {
          status: 'open',
          entryPrice: 100,
          entryBar: 0,
          size: 1,
          direction: 'long',
        },
      ];

      const html = renderer.renderRows(trades, 105);

      expect(html).toContain('>Open<');
      expect(html).toContain('>-<');
    });

    it('should uppercase direction values', () => {
      const trades = [
        {
          entryPrice: 100,
          entryBar: 0,
          exitPrice: 105,
          exitBar: 1,
          size: 1,
          profit: 5,
          direction: 'long',
        },
        {
          entryPrice: 100,
          entryBar: 0,
          exitPrice: 95,
          exitBar: 1,
          size: 1,
          profit: 5,
          direction: 'short',
        },
      ];

      const html = renderer.renderRows(trades, null);

      expect(html).toContain('>LONG<');
      expect(html).toContain('>SHORT<');
      expect(html).not.toContain('>long<');
      expect(html).not.toContain('>short<');
    });

    it('should handle N/A values in rendered output', () => {
      const trades = [
        {
          entryPrice: 100,
          exitPrice: 105,
          size: 1,
          profit: 5,
          direction: 'long',
        },
      ];

      const html = renderer.renderRows(trades, null);

      expect(html).toContain('N/A');
    });
  });

  describe('Mixed Trade Types', () => {
    it('should render both open and closed trades correctly', () => {
      const trades = [
        {
          entryId: 'CLOSED_001',
          entryPrice: 100,
          entryBar: 0,
          exitPrice: 110,
          exitBar: 1,
          size: 1,
          profit: 10,
          direction: 'long',
        },
        {
          status: 'open',
          entryId: 'OPEN_001',
          entryPrice: 105,
          entryBar: 1,
          size: 1,
          direction: 'long',
        },
      ];

      const html = renderer.renderRows(trades, 115);

      expect(html).toContain('CLOSED_001');
      expect(html).toContain('OPEN_001');
      expect(html).toContain('$110.00');
      expect(html).toContain('>Open<');
    });

    it('should render both long and short trades in same list', () => {
      const trades = [
        {
          entryPrice: 100,
          entryBar: 0,
          exitPrice: 110,
          exitBar: 1,
          size: 1,
          profit: 10,
          direction: 'long',
        },
        {
          entryPrice: 100,
          entryBar: 0,
          exitPrice: 90,
          exitBar: 1,
          size: 1,
          profit: 10,
          direction: 'short',
        },
      ];

      const html = renderer.renderRows(trades, null);

      expect(html).toContain('trade-long');
      expect(html).toContain('trade-short');
    });

    it('should render profitable and losing trades with correct classes', () => {
      const trades = [
        {
          entryPrice: 100,
          entryBar: 0,
          exitPrice: 110,
          exitBar: 1,
          size: 1,
          profit: 10,
          direction: 'long',
        },
        {
          entryPrice: 100,
          entryBar: 0,
          exitPrice: 90,
          exitBar: 1,
          size: 1,
          profit: -10,
          direction: 'long',
        },
      ];

      const html = renderer.renderRows(trades, null);

      const positiveMatch = html.match(/trade-profit-positive/g);
      const negativeMatch = html.match(/trade-profit-negative/g);

      expect(positiveMatch).toHaveLength(1);
      expect(negativeMatch).toHaveLength(1);
    });
  });

  describe('Large Data Sets', () => {
    it('should handle 100 trades without errors', () => {
      const trades = Array.from({ length: 100 }, (_, i) => ({
        entryId: `TRADE_${i}`,
        entryPrice: 100 + i,
        entryBar: i,
        exitPrice: 105 + i,
        exitBar: i + 1,
        size: 1,
        profit: 5,
        direction: i % 2 === 0 ? 'long' : 'short',
      }));

      expect(() => renderer.renderRows(trades, null)).not.toThrow();
    });

    it('should generate correct row count for large data sets', () => {
      const trades = Array.from({ length: 50 }, (_, i) => ({
        entryPrice: 100,
        entryBar: i,
        exitPrice: 105,
        exitBar: i + 1,
        size: 1,
        profit: 5,
        direction: 'long',
      }));

      const html = renderer.renderRows(trades, null);
      const rowCount = (html.match(/<tr>/g) || []).length;

      expect(rowCount).toBe(50);
    });

    it('should maintain column count consistency across many rows', () => {
      const trades = Array.from({ length: 20 }, () => ({
        entryPrice: 100,
        entryBar: 0,
        exitPrice: 105,
        exitBar: 1,
        size: 1,
        profit: 5,
        direction: 'long',
      }));

      const html = renderer.renderRows(trades, null);
      const rows = html.split('</tr>').filter((row) => row.includes('<tr>'));

      rows.forEach((row) => {
        const tdCount = (row.match(/<td/g) || []).length;
        expect(tdCount).toBe(11);
      });
    });
  });

  describe('Edge Cases and Error Handling', () => {
    it('should handle trades with missing optional fields', () => {
      const trades = [
        {
          entryPrice: 100,
          exitPrice: 105,
          size: 1,
          direction: 'long',
        },
      ];

      expect(() => renderer.renderRows(trades, null)).not.toThrow();
    });

    it('should handle null currentPrice for closed trades', () => {
      const trades = [
        {
          entryPrice: 100,
          entryBar: 0,
          exitPrice: 105,
          exitBar: 1,
          size: 1,
          profit: 5,
          direction: 'long',
        },
      ];

      expect(() => renderer.renderRows(trades, null)).not.toThrow();
    });

    it('should render HTML special characters in entry IDs as-is', () => {
      const trades = [
        {
          entryId: '<script>alert("xss")</script>',
          entryPrice: 100,
          entryBar: 0,
          exitPrice: 105,
          exitBar: 1,
          size: 1,
          profit: 5,
          direction: 'long',
        },
      ];

      const html = renderer.renderRows(trades, null);

      expect(html).toContain('<script>alert("xss")</script>');
    });

    it('should handle very long HTML output without truncation', () => {
      const trades = Array.from({ length: 1000 }, (_, i) => ({
        entryId: `TRADE_${i}`,
        entryPrice: 100,
        entryBar: i,
        exitPrice: 105,
        exitBar: i + 1,
        size: 1,
        profit: 5,
        direction: 'long',
      }));

      const html = renderer.renderRows(trades, null);
      const rowCount = (html.match(/<tr>/g) || []).length;

      expect(rowCount).toBe(1000);
    });

    it('should maintain formatter reference after instantiation', () => {
      expect(renderer.formatter).toBe(formatter);
    });

    it('should handle trades with undefined status as closed', () => {
      const trades = [
        {
          entryPrice: 100,
          entryBar: 0,
          exitPrice: 105,
          exitBar: 1,
          size: 1,
          profit: 5,
          direction: 'long',
        },
      ];

      const html = renderer.renderRows(trades, 110);

      expect(html).not.toContain('>Open<');
      expect(html).toContain('$105.00');
    });
  });

  describe('Column Order Validation', () => {
    it('should maintain correct column order: #, Entry Date, Entry Bar, Direction, Entry Price, Exit Date, Exit Bar, Exit Price, Size, P/L, Entry ID', () => {
      const trades = [
        {
          entryId: 'ORDER_TEST',
          entryPrice: 100,
          entryBar: 5,
          entryTime: 1500000000,
          exitPrice: 110,
          exitBar: 10,
          exitTime: 1500003600,
          size: 2,
          profit: 20,
          direction: 'long',
        },
      ];

      const html = renderer.renderRows(trades, null);
      const cells = html.match(/<td[^>]*>([^<]*)<\/td>/g) || [];

      expect(cells[0]).toContain('>1<');
      expect(cells[1]).toContain('2017');
      expect(cells[2]).toContain('>5<');
      expect(cells[3]).toContain('LONG');
      expect(cells[4]).toContain('$100.00');
      expect(cells[5]).toContain('2017');
      expect(cells[6]).toContain('>10<');
      expect(cells[7]).toContain('$110.00');
      expect(cells[8]).toContain('2.00');
      expect(cells[9]).toContain('+$20.00');
      expect(cells[10]).toContain('ORDER_TEST');
    });
  });
});
