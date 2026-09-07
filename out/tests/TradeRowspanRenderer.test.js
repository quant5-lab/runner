import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { TradeRowspanRenderer } from '../js/TradeRowspanRenderer.js';
import { TradeRowData          } from '../js/TradeRowData.js';

// ── Row factory ────────────────────────────────────────────────────────────────
//
// Constructs TradeRowData directly — no transformer dependency.
// Defaults represent a valid primary entry row; override as needed.

function makeRow(overrides = {}) {
  return new TradeRowData({
    tradeNumber: 1,
    rowType:     'entry',
    isPrimary:   true,
    dateTime:    'Jan 1, 2024, 12:00 PM',
    signal:      'signal-a',
    price:       '$100.00',
    size:        '1.00',
    profitLoss:  '',
    direction:   'long',
    isOpen:      false,
    profitRaw:   0,
    ...overrides,
  });
}

// Canonical row configurations — 4 positions in the 2×2 (rowType × isPrimary) matrix
const primaryEntryRow    = (o = {}) => makeRow({ rowType: 'entry', isPrimary: true,  ...o });
const secondaryEntryRow  = (o = {}) => makeRow({ rowType: 'entry', isPrimary: false, ...o });
const primaryExitRow     = (o = {}) => makeRow({ rowType: 'exit',  isPrimary: true,  profitLoss: '+$10.00', profitRaw: 10, ...o });
const secondaryExitRow   = (o = {}) => makeRow({ rowType: 'exit',  isPrimary: false, profitLoss: '+$10.00', profitRaw: 10, ...o });

const renderer = new TradeRowspanRenderer();

// ── Primary row — rowspan structure ────────────────────────────────────────────

describe('renderRow() — primary row: rowspan cells', () => {
  for (const [label, row] of [['entry primary', primaryEntryRow()], ['exit primary', primaryExitRow()]]) {
    it(`${label}: exactly 2 rowspan="2" cells`, () => {
      assert.equal((renderer.renderRow(row).match(/rowspan="2"/g) || []).length, 2);
    });

    it(`${label}: Type cell has rowspan="2"`, () => {
      assert.ok(renderer.renderRow(row).includes('<td rowspan="2"'));
    });

    it(`${label}: Size cell has rowspan="2" with correct value`, () => {
      const r = makeRow({ ...row, size: '3.75', rowType: row.rowType, isPrimary: true });
      assert.ok(renderer.renderRow(r).includes('rowspan="2">3.75<'));
    });
  }
});

// ── Secondary row — no rowspan ─────────────────────────────────────────────────

describe('renderRow() — secondary row: no rowspan cells', () => {
  it('secondary entry row has no rowspan attribute', () => {
    assert.ok(!renderer.renderRow(secondaryEntryRow()).includes('rowspan'));
  });

  it('secondary exit row has no rowspan attribute', () => {
    assert.ok(!renderer.renderRow(secondaryExitRow()).includes('rowspan'));
  });
});

// ── Label routing ──────────────────────────────────────────────────────────────

describe('renderRow() — label routing', () => {
  it('primary entry row emits "Entry" label', () => {
    assert.ok(renderer.renderRow(primaryEntryRow()).includes('<td>Entry</td>'));
  });

  it('secondary entry row emits "Entry" label', () => {
    assert.ok(renderer.renderRow(secondaryEntryRow()).includes('<td>Entry</td>'));
  });

  it('primary exit row emits "Exit" label', () => {
    assert.ok(renderer.renderRow(primaryExitRow()).includes('<td>Exit</td>'));
  });

  it('secondary exit row emits "Exit" label', () => {
    assert.ok(renderer.renderRow(secondaryExitRow()).includes('<td>Exit</td>'));
  });
});

// ── Type cell — direction value and class ──────────────────────────────────────
//
// Type cell (direction text + CSS class) appears only on the primary row.

describe('renderRow() — Type cell on primary rows', () => {
  const dirCases = [
    { direction: 'long',  cssClass: 'trade-long',  text: 'LONG'  },
    { direction: 'short', cssClass: 'trade-short', text: 'SHORT' },
  ];

  for (const { direction, cssClass, text } of dirCases) {
    it(`${direction} primary entry row → Type cell class="${cssClass}" text="${text}"`, () => {
      const html = renderer.renderRow(primaryEntryRow({ direction }));
      assert.ok(html.includes(`class="${cssClass}"`), 'missing direction class');
      assert.ok(html.includes(`>${text}</td>`),       'missing direction text');
    });

    it(`${direction} primary exit row → Type cell class="${cssClass}" text="${text}"`, () => {
      const html = renderer.renderRow(primaryExitRow({ direction }));
      assert.ok(html.includes(`class="${cssClass}"`), 'missing direction class');
      assert.ok(html.includes(`>${text}</td>`),       'missing direction text');
    });
  }

  it('secondary entry row has no direction class', () => {
    const html = renderer.renderRow(secondaryEntryRow({ direction: 'long' }));
    assert.ok(!html.includes('trade-long'),  'secondary must not carry trade-long');
    assert.ok(!html.includes('trade-short'), 'secondary must not carry trade-short');
  });

  it('secondary exit row has no direction class', () => {
    const html = renderer.renderRow(secondaryExitRow({ direction: 'short' }));
    assert.ok(!html.includes('trade-long'),  'secondary must not carry trade-long');
    assert.ok(!html.includes('trade-short'), 'secondary must not carry trade-short');
  });
});

// ── P/L cell placement — exit rows only ───────────────────────────────────────

describe('renderRow() — P/L cell: exit rows only, both positions', () => {
  it('primary exit row has a P/L cell', () => {
    assert.ok(renderer.renderRow(primaryExitRow({ profitLoss: '+$10.00' })).includes('+$10.00'));
  });

  it('secondary exit row has a P/L cell', () => {
    assert.ok(renderer.renderRow(secondaryExitRow({ profitLoss: '-$5.00' })).includes('-$5.00'));
  });

  it('primary entry row has no P/L cell', () => {
    assert.ok(!renderer.renderRow(primaryEntryRow({ profitLoss: '+$99.00' })).includes('+$99.00'));
  });

  it('secondary entry row has no P/L cell', () => {
    assert.ok(!renderer.renderRow(secondaryEntryRow({ profitLoss: '+$99.00' })).includes('+$99.00'));
  });
});

// ── Closed exit P/L class — profitRaw-based, position-invariant ────────────────

describe('renderRow() — closed exit P/L cell class', () => {
  const profitCases = [
    { label: 'positive profit',   profitRaw:  10,   expected: 'trade-profit-positive' },
    { label: 'zero profit',       profitRaw:  0,    expected: 'trade-profit-positive' },
    { label: 'fractional profit', profitRaw:  0.01, expected: 'trade-profit-positive' },
    { label: 'very large profit', profitRaw:  1e9,  expected: 'trade-profit-positive' },
    { label: 'negative profit',   profitRaw: -0.01, expected: 'trade-profit-negative' },
    { label: 'very large loss',   profitRaw: -1e9,  expected: 'trade-profit-negative' },
  ];

  for (const { label, profitRaw, expected } of profitCases) {
    it(`${label} (profitRaw=${profitRaw}) → class="${expected}" in both pair positions`, () => {
      for (const exitRow of [primaryExitRow({ profitRaw }), secondaryExitRow({ profitRaw })]) {
        assert.ok(
          renderer.renderRow(exitRow).includes(`<td class="${expected}">`),
          `isPrimary=${exitRow.isPrimary}`,
        );
      }
    });
  }
});

// ── Open exit row — trade-open on <tr> and P/L <td> ──────────────────────────

describe('renderRow() — open exit row styling', () => {
  it('primary open exit: <tr> carries class="trade-open"', () => {
    assert.ok(renderer.renderRow(primaryExitRow({ isOpen: true })).startsWith('<tr class="trade-open">'));
  });

  it('secondary open exit: <tr> carries class="trade-open"', () => {
    assert.ok(renderer.renderRow(secondaryExitRow({ isOpen: true })).startsWith('<tr class="trade-open">'));
  });

  it('open exit P/L cell carries class="trade-open" regardless of profitRaw sign, in both positions', () => {
    for (const profitRaw of [100, -100, 0]) {
      for (const exitRow of [primaryExitRow({ isOpen: true, profitRaw }), secondaryExitRow({ isOpen: true, profitRaw })]) {
        const html = renderer.renderRow(exitRow);
        assert.ok(html.includes('<td class="trade-open">'),   `profitRaw=${profitRaw} isPrimary=${exitRow.isPrimary}: missing trade-open`);
        assert.ok(!html.includes('trade-profit-positive'),    `profitRaw=${profitRaw} isPrimary=${exitRow.isPrimary}: unexpected profit class`);
        assert.ok(!html.includes('trade-profit-negative'),    `profitRaw=${profitRaw} isPrimary=${exitRow.isPrimary}: unexpected loss class`);
      }
    }
  });

  it('open exit trade-open styling is invariant across trade directions', () => {
    for (const direction of ['long', 'short']) {
      for (const exitRow of [primaryExitRow({ isOpen: true, direction }), secondaryExitRow({ isOpen: true, direction })]) {
        const html = renderer.renderRow(exitRow);
        assert.ok(html.startsWith('<tr class="trade-open">'), `direction=${direction} isPrimary=${exitRow.isPrimary} tr`);
        assert.ok(html.includes('<td class="trade-open">'),   `direction=${direction} isPrimary=${exitRow.isPrimary} td`);
      }
    }
  });
});

// ── Entry rows — never carry trade-open ────────────────────────────────────────

describe('renderRow() — entry rows never carry trade-open class', () => {
  it('primary entry row with isOpen=true has no class on <tr>', () => {
    assert.ok(renderer.renderRow(primaryEntryRow({ isOpen: true })).startsWith('<tr>'));
  });

  it('secondary entry row with isOpen=true has no class on <tr>', () => {
    assert.ok(renderer.renderRow(secondaryEntryRow({ isOpen: true })).startsWith('<tr>'));
  });
});

// ── Cell value passthrough ─────────────────────────────────────────────────────

describe('renderRow() — cell value passthrough', () => {
  it('entry row: dateTime appears in output', () => {
    assert.ok(renderer.renderRow(primaryEntryRow({ dateTime: 'Mar 15, 2025' })).includes('Mar 15, 2025'));
  });

  it('exit row: dateTime appears in output', () => {
    assert.ok(renderer.renderRow(primaryExitRow({ dateTime: 'Mar 16, 2025' })).includes('Mar 16, 2025'));
  });

  it('entry row: non-empty signal appears in output', () => {
    assert.ok(renderer.renderRow(primaryEntryRow({ signal: 'entry-sig' })).includes('entry-sig'));
  });

  it('exit row: non-empty signal appears in output', () => {
    assert.ok(renderer.renderRow(primaryExitRow({ signal: 'exit-sig' })).includes('exit-sig'));
  });

  it('entry row: empty signal renders as empty cell (not omitted)', () => {
    assert.ok(renderer.renderRow(primaryEntryRow({ signal: '' })).includes('<td></td>'));
  });

  it('exit row: empty signal renders as empty cell (not omitted)', () => {
    assert.ok(renderer.renderRow(primaryExitRow({ signal: '' })).includes('<td></td>'));
  });

  it('entry row: price appears in output', () => {
    assert.ok(renderer.renderRow(primaryEntryRow({ price: '$123.45' })).includes('$123.45'));
  });

  it('exit row: price appears in output', () => {
    assert.ok(renderer.renderRow(primaryExitRow({ price: '$999.00' })).includes('$999.00'));
  });

  it('exit row: profitLoss value appears in P/L cell', () => {
    assert.ok(renderer.renderRow(primaryExitRow({ profitLoss: '+$42.00' })).includes('+$42.00'));
  });

  it('secondary exit row: profitLoss value appears in P/L cell', () => {
    assert.ok(renderer.renderRow(secondaryExitRow({ profitLoss: '-$7.50' })).includes('-$7.50'));
  });
});

// ── renderRows() — batch output ────────────────────────────────────────────────

describe('renderRows()', () => {
  it('empty array produces empty string', () => {
    assert.equal(renderer.renderRows([]), '');
  });

  it('single row produces no newline separator', () => {
    const html = renderer.renderRows([primaryEntryRow()]);
    assert.equal((html.match(/\n/g) || []).length, 0);
    assert.equal((html.match(/<tr/g)  || []).length, 1);
  });

  it('N rows produces N <tr> elements with N-1 newline separators', () => {
    const rows = [primaryEntryRow(), secondaryExitRow(), primaryEntryRow(), secondaryExitRow()];
    const html = renderer.renderRows(rows);
    assert.equal((html.match(/<tr/g) || []).length, 4);
    assert.equal((html.match(/\n/g)  || []).length, 3);
  });

  it('output order matches input array order', () => {
    const first  = primaryEntryRow({ dateTime: 'FIRST'  });
    const second = secondaryExitRow({ dateTime: 'SECOND' });
    const html   = renderer.renderRows([first, second]);
    assert.ok(html.indexOf('FIRST') < html.indexOf('SECOND'));
  });
});
