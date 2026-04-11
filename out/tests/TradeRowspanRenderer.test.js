import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { TradeRowspanRenderer } from '../js/TradeRowspanRenderer.js';
import { TradeRowData } from '../js/TradeRowData.js';

// ── Row fixtures ───────────────────────────────────────────────────────────────

function makeEntryRow(overrides = {}) {
  return new TradeRowData({
    tradeNumber: 1,
    rowType:     'entry',
    dateTime:    '2024-01-01',
    signal:      'buy signal',
    price:       '$100.00',
    size:        '1.00',
    profitLoss:  '',
    direction:   'long',
    isOpen:      false,
    profitRaw:   0,
    ...overrides,
  });
}

function makeExitRow(overrides = {}) {
  return new TradeRowData({
    tradeNumber: 1,
    rowType:     'exit',
    dateTime:    '2024-01-02',
    signal:      'sell signal',
    price:       '$110.00',
    size:        '',
    profitLoss:  '+$10.00',
    direction:   'long',
    isOpen:      false,
    profitRaw:   10,
    ...overrides,
  });
}

// ── Helper ─────────────────────────────────────────────────────────────────────

const renderer = new TradeRowspanRenderer();

function render(row) {
  return renderer.renderRow(row);
}

// ── Entry row — structural invariants ─────────────────────────────────────────

describe('renderRow() — entry row structure', () => {
  it('row element has no class attribute', () => {
    assert.ok(render(makeEntryRow()).startsWith('<tr>'));
  });

  it('open entry row element has no class attribute', () => {
    assert.ok(render(makeEntryRow({ isOpen: true })).startsWith('<tr>'));
  });

  it('Type cell has rowspan="2"', () => {
    assert.ok(render(makeEntryRow()).includes('<td rowspan="2"'));
  });

  it('Size cell has rowspan="2"', () => {
    const html = render(makeEntryRow({ size: '2.50' }));
    assert.ok(html.includes('rowspan="2">2.50<'));
  });

  it('row contains exactly 2 rowspan="2" cells', () => {
    const html = render(makeEntryRow());
    assert.equal((html.match(/rowspan="2"/g) || []).length, 2);
  });

  it('contains "Entry" label cell', () => {
    assert.ok(render(makeEntryRow()).includes('<td>Entry</td>'));
  });
});

// ── Entry row — cell value passthrough ────────────────────────────────────────

describe('renderRow() — entry row cell value passthrough', () => {
  it('dateTime value appears in output', () => {
    assert.ok(render(makeEntryRow({ dateTime: 'Jan 1, 2024' })).includes('Jan 1, 2024'));
  });

  it('signal value appears in output', () => {
    assert.ok(render(makeEntryRow({ signal: 'my-entry-signal' })).includes('my-entry-signal'));
  });

  it('price value appears in output', () => {
    assert.ok(render(makeEntryRow({ price: '$123.45' })).includes('$123.45'));
  });

  it('size value appears in output', () => {
    assert.ok(render(makeEntryRow({ size: '9.99' })).includes('9.99'));
  });

  it('direction value (uppercased) appears in Type cell', () => {
    for (const [direction, expected] of [['long', 'LONG'], ['short', 'SHORT']]) {
      assert.ok(render(makeEntryRow({ direction })).includes(expected), `direction=${direction}`);
    }
  });
});

// ── Entry row — direction class ────────────────────────────────────────────────

describe('renderRow() — direction class on Type cell', () => {
  const cases = [
    { direction: 'long',  expected: 'trade-long'  },
    { direction: 'short', expected: 'trade-short' },
  ];

  for (const { direction, expected } of cases) {
    it(`${direction} → Type cell class="${expected}"`, () => {
      assert.ok(render(makeEntryRow({ direction })).includes(`class="${expected}"`));
    });
  }
});

// ── Exit row — structural invariants ──────────────────────────────────────────

describe('renderRow() — exit row structure', () => {
  it('closed exit row element has no class attribute', () => {
    assert.ok(render(makeExitRow({ isOpen: false })).startsWith('<tr>'));
  });

  it('exit row has no rowspan attribute', () => {
    assert.ok(!render(makeExitRow()).includes('rowspan'));
  });

  it('contains "Exit" label cell', () => {
    assert.ok(render(makeExitRow()).includes('<td>Exit</td>'));
  });
});

// ── Exit row — cell value passthrough ─────────────────────────────────────────

describe('renderRow() — exit row cell value passthrough', () => {
  it('dateTime value appears in output', () => {
    assert.ok(render(makeExitRow({ dateTime: 'Jan 2, 2024' })).includes('Jan 2, 2024'));
  });

  it('signal value appears in output', () => {
    assert.ok(render(makeExitRow({ signal: 'my-exit-signal' })).includes('my-exit-signal'));
  });

  it('price value appears in output', () => {
    assert.ok(render(makeExitRow({ price: '$999.00' })).includes('$999.00'));
  });

  it('profitLoss value appears in P/L cell', () => {
    assert.ok(render(makeExitRow({ profitLoss: '+$42.00' })).includes('+$42.00'));
  });
});

// ── Closed exit row — P/L cell class ──────────────────────────────────────────

describe('renderRow() — closed exit row P/L cell class', () => {
  const cases = [
    { label: 'positive profit',   profitRaw:  10,   expected: 'trade-profit-positive' },
    { label: 'zero profit',       profitRaw:  0,    expected: 'trade-profit-positive' },
    { label: 'fractional profit', profitRaw:  0.01, expected: 'trade-profit-positive' },
    { label: 'very large profit', profitRaw:  1e9,  expected: 'trade-profit-positive' },
    { label: 'negative profit',   profitRaw: -0.01, expected: 'trade-profit-negative' },
    { label: 'fractional loss',   profitRaw: -10,   expected: 'trade-profit-negative' },
    { label: 'very large loss',   profitRaw: -1e9,  expected: 'trade-profit-negative' },
  ];

  for (const { label, profitRaw, expected } of cases) {
    it(`${label} (profitRaw=${profitRaw}) → P/L cell class="${expected}"`, () => {
      assert.ok(render(makeExitRow({ profitRaw, isOpen: false })).includes(`<td class="${expected}">`));
    });
  }
});

// ── Open exit row ─────────────────────────────────────────────────────────────

describe('renderRow() — open exit row', () => {
  const openProfitValues = [100, -100, 0];

  it('row element carries class="trade-open"', () => {
    assert.ok(render(makeExitRow({ isOpen: true })).startsWith('<tr class="trade-open">'));
  });

  it('P/L cell carries class="trade-open" regardless of profitRaw sign', () => {
    for (const profitRaw of openProfitValues) {
      assert.ok(
        render(makeExitRow({ isOpen: true, profitRaw })).includes('<td class="trade-open">'),
        `profitRaw=${profitRaw}`,
      );
    }
  });

  it('P/L cell does not carry profit/loss class regardless of profitRaw sign', () => {
    for (const profitRaw of openProfitValues) {
      const html = render(makeExitRow({ isOpen: true, profitRaw }));
      assert.ok(!html.includes('trade-profit-positive'), `profitRaw=${profitRaw}`);
      assert.ok(!html.includes('trade-profit-negative'), `profitRaw=${profitRaw}`);
    }
  });

  it('open exit styling is invariant across trade directions', () => {
    for (const direction of ['long', 'short']) {
      const html = render(makeExitRow({ isOpen: true, direction }));
      assert.ok(html.startsWith('<tr class="trade-open">'),    `direction=${direction} tr`);
      assert.ok(html.includes('<td class="trade-open">'),      `direction=${direction} td`);
    }
  });
});

// ── renderRows() ───────────────────────────────────────────────────────────────

describe('renderRows() — batch output', () => {
  it('empty array produces empty string', () => {
    assert.equal(renderer.renderRows([]), '');
  });

  it('N rows produces N <tr> elements joined by newlines', () => {
    const rows = [makeEntryRow(), makeExitRow(), makeEntryRow({ tradeNumber: 2 }), makeExitRow({ tradeNumber: 2 })];
    const html = renderer.renderRows(rows);
    assert.equal((html.match(/<tr/g) || []).length, 4);
    assert.equal((html.match(/\n/g)  || []).length, 3);
  });

  it('output order matches input array order', () => {
    const entryHtml = render(makeEntryRow());
    const exitHtml  = render(makeExitRow());
    const html      = renderer.renderRows([makeEntryRow(), makeExitRow()]);
    assert.ok(html.indexOf(entryHtml) < html.indexOf(exitHtml));
  });
});
