import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { TradeRowspanTransformer } from '../js/TradeRowspanTransformer.js';

// ── Stub formatter ─────────────────────────────────────────────────────────────

function makeFormatter(overrides = {}) {
  return {
    calculateUnrealizedProfit: (_trade, _price) => 0,
    getTradeDate:              (_trade, isEntry) => isEntry ? 'entry-date' : 'exit-date',
    formatPrice:               (price)           => `$${(price ?? 0).toFixed(2)}`,
    formatProfit:              (value)            => `profit:${value}`,
    ...overrides,
  };
}

// ── Trade fixtures ─────────────────────────────────────────────────────────────

function makeClosedTrade(overrides = {}) {
  return {
    direction:  'long',
    status:     'closed',
    entryPrice: 100,
    exitPrice:  110,
    size:       1,
    profit:     10,
    entryId:    'eid',
    exitId:     'xid',
    ...overrides,
  };
}

function makeOpenTrade(overrides = {}) {
  return {
    direction:  'long',
    status:     'open',
    entryPrice: 100,
    size:       1,
    entryId:    'eid',
    ...overrides,
  };
}

// ── Helpers ────────────────────────────────────────────────────────────────────

function transform(trade, { formatter, currentPrice = null, tradeNumber = 1 } = {}) {
  return new TradeRowspanTransformer(formatter ?? makeFormatter())
    .transformTrade(trade, tradeNumber, currentPrice);
}

function transformAll(trades, { formatter, currentPrice = null } = {}) {
  return new TradeRowspanTransformer(formatter ?? makeFormatter())
    .transformTrades(trades, currentPrice);
}

function assertSignalFallback(rowExtractor, signalCases) {
  for (const { label, overrides, expected } of signalCases) {
    it(label, () => {
      const rows = transform(makeClosedTrade(overrides));
      assert.equal(rowExtractor(rows).signal, expected);
    });
  }
}

const ENTRY_ROW = ([entry])  => entry;
const EXIT_ROW  = ([, exit]) => exit;

const SIGNAL_CASES = (commentKey, idKey) => [
  { label: `${commentKey} takes precedence over ${idKey}`, overrides: { [commentKey]: 'COMMENT', [idKey]: 'ID'        }, expected: 'COMMENT' },
  { label: `${idKey} used when ${commentKey} absent`,      overrides: {                          [idKey]: 'ID'        }, expected: 'ID'      },
  { label: 'empty string when neither present',             overrides: { [commentKey]: undefined, [idKey]: undefined   }, expected: ''        },
];

// ── Row count and types ────────────────────────────────────────────────────────

describe('transformTrade() — row count and types', () => {
  it('produces exactly 2 rows per trade regardless of status', () => {
    assert.equal(transform(makeClosedTrade()).length, 2);
    assert.equal(transform(makeOpenTrade()).length,   2);
  });

  it('first row is always entry, second is always exit', () => {
    for (const trade of [makeClosedTrade(), makeOpenTrade()]) {
      const [entry, exit] = transform(trade);
      assert.ok(entry.isEntryRow(), 'first row must be entry');
      assert.ok(exit.isExitRow(),   'second row must be exit');
    }
  });

  it('tradeNumber propagates to both rows', () => {
    const [entry, exit] = transform(makeClosedTrade(), { tradeNumber: 7 });
    assert.equal(entry.tradeNumber, 7);
    assert.equal(exit.tradeNumber,  7);
  });
});

// ── Entry row fields ───────────────────────────────────────────────────────────

describe('transformTrade() — entry row fields', () => {
  it('direction propagates to entry row for both trade directions', () => {
    for (const direction of ['long', 'short']) {
      const [entry] = transform(makeClosedTrade({ direction }));
      assert.equal(entry.direction, direction);
    }
  });

  it('direction propagates to exit row for both trade directions', () => {
    for (const direction of ['long', 'short']) {
      const [, exit] = transform(makeClosedTrade({ direction }));
      assert.equal(exit.direction, direction);
    }
  });

  it('entry row dateTime delegates to formatter.getTradeDate(trade, true)', () => {
    const formatter = makeFormatter({ getTradeDate: (_t, isEntry) => isEntry ? 'ENTRY-TS' : 'EXIT-TS' });
    const [entry] = transform(makeClosedTrade(), { formatter });
    assert.equal(entry.dateTime, 'ENTRY-TS');
  });

  it('entry row price delegates to formatter.formatPrice(entryPrice)', () => {
    const formatter = makeFormatter({ formatPrice: (p) => `FMT:${p}` });
    const [entry] = transform(makeClosedTrade({ entryPrice: 999 }), { formatter });
    assert.equal(entry.price, 'FMT:999');
  });

  it('entry row size is trade.size.toFixed(2)', () => {
    for (const [size, expected] of [[2.5, '2.50'], [1, '1.00'], [0, '0.00'], [10.123, '10.12']]) {
      const [entry] = transform(makeClosedTrade({ size }));
      assert.equal(entry.size, expected, `size=${size}`);
    }
  });

  it('entry row isOpen is always false regardless of trade status', () => {
    const [closedEntry] = transform(makeClosedTrade());
    const [openEntry]   = transform(makeOpenTrade());
    assert.equal(closedEntry.isOpen, false);
    assert.equal(openEntry.isOpen,   false);
  });

  it('entry row P/L fields are always zeroed regardless of trade profit', () => {
    const [entry] = transform(makeClosedTrade({ profit: 9999 }));
    assert.equal(entry.profitLoss, '');
    assert.equal(entry.profitRaw,  0);
  });
});

// ── Entry signal fallback chain ────────────────────────────────────────────────

describe('transformTrade() — entry signal fallback chain (entryComment > entryId > "")', () => {
  assertSignalFallback(ENTRY_ROW, SIGNAL_CASES('entryComment', 'entryId'));
});

// ── Closed exit row fields ─────────────────────────────────────────────────────

describe('transformTrade() — closed exit row fields', () => {
  it('exit row isOpen is false', () => {
    const [, exit] = transform(makeClosedTrade());
    assert.equal(exit.isOpen, false);
  });

  it('exit row dateTime delegates to formatter.getTradeDate(trade, false)', () => {
    const formatter = makeFormatter({ getTradeDate: (_t, isEntry) => isEntry ? 'ENTRY-TS' : 'EXIT-TS' });
    const [, exit] = transform(makeClosedTrade(), { formatter });
    assert.equal(exit.dateTime, 'EXIT-TS');
  });

  it('exit row profitRaw equals trade.profit', () => {
    for (const profit of [42, -42, 0, 0.001, -0.001]) {
      const [, exit] = transform(makeClosedTrade({ profit }));
      assert.equal(exit.profitRaw, profit, `profit=${profit}`);
    }
  });

  it('exit row profitLoss delegates to formatter.formatProfit(trade.profit)', () => {
    const formatter = makeFormatter({ formatProfit: (v) => `FORMATTED:${v}` });
    const [, exit] = transform(makeClosedTrade({ profit: 55 }), { formatter });
    assert.equal(exit.profitLoss, 'FORMATTED:55');
  });

  it('exit row price delegates to formatter.formatPrice(exitPrice)', () => {
    const formatter = makeFormatter({ formatPrice: (p) => `FMT:${p}` });
    const [, exit] = transform(makeClosedTrade({ exitPrice: 222 }), { formatter });
    assert.equal(exit.price, 'FMT:222');
  });

  it('exit row price falls back to 0 when exitPrice is null', () => {
    const formatter = makeFormatter({ formatPrice: (p) => `FMT:${p}` });
    const [, exit] = transform(makeClosedTrade({ exitPrice: null }), { formatter });
    assert.equal(exit.price, 'FMT:0');
  });
});

// ── Exit signal fallback chain ─────────────────────────────────────────────────

describe('transformTrade() — exit signal fallback chain (exitComment > exitId > "")', () => {
  assertSignalFallback(EXIT_ROW, SIGNAL_CASES('exitComment', 'exitId'));
});

// ── Open trade exit row ────────────────────────────────────────────────────────

describe('transformTrade() — open trade exit row', () => {
  it('exit row isOpen is true', () => {
    const [, exit] = transform(makeOpenTrade());
    assert.equal(exit.isOpen, true);
  });

  it('exit row dateTime is literal "Open"', () => {
    const [, exit] = transform(makeOpenTrade());
    assert.equal(exit.dateTime, 'Open');
  });

  it('exit row signal is always empty string regardless of exitComment/exitId presence', () => {
    const [, exit] = transform(makeOpenTrade({ exitComment: 'x', exitId: 'y' }));
    assert.equal(exit.signal, '');
  });

  it('exit row profitRaw delegates to formatter.calculateUnrealizedProfit(trade, currentPrice)', () => {
    let capturedArgs = null;
    const formatter = makeFormatter({
      calculateUnrealizedProfit: (trade, price) => { capturedArgs = { trade, price }; return 77; },
    });
    const trade = makeOpenTrade();
    const [, exit] = transform(trade, { formatter, currentPrice: 200 });
    assert.equal(exit.profitRaw,       77);
    assert.equal(capturedArgs.price,   200);
    assert.equal(capturedArgs.trade,   trade);
  });

  it('exit row price uses currentPrice when provided', () => {
    const formatter = makeFormatter({ formatPrice: (p) => `FMT:${p}` });
    const [, exit] = transform(makeOpenTrade({ entryPrice: 50 }), { formatter, currentPrice: 120 });
    assert.equal(exit.price, 'FMT:120');
  });

  it('exit row price falls back to entryPrice when currentPrice is null', () => {
    const formatter = makeFormatter({ formatPrice: (p) => `FMT:${p}` });
    const [, exit] = transform(makeOpenTrade({ entryPrice: 50 }), { formatter, currentPrice: null });
    assert.equal(exit.price, 'FMT:50');
  });

  it('exit row price uses 0 as currentPrice (not entryPrice) when currentPrice is 0', () => {
    const formatter = makeFormatter({ formatPrice: (p) => `FMT:${p}` });
    const [, exit] = transform(makeOpenTrade({ entryPrice: 50 }), { formatter, currentPrice: 0 });
    assert.equal(exit.price, 'FMT:0');
  });
});

// ── transformTrades() — multi-trade composition ────────────────────────────────

describe('transformTrades() — ordering and row count', () => {
  it('produces 2N rows for N trades', () => {
    for (const n of [0, 1, 3, 5]) {
      const trades = Array.from({ length: n }, () => makeClosedTrade());
      assert.equal(transformAll(trades).length, n * 2, `expected ${n * 2} rows for ${n} trades`);
    }
  });

  it('row order is entry,exit,entry,exit,... preserving input trade order', () => {
    const trades = [makeClosedTrade({ entryId: 'A' }), makeClosedTrade({ entryId: 'B' })];
    const rows = transformAll(trades);
    assert.ok(rows[0].isEntryRow());
    assert.equal(rows[0].signal, 'A');
    assert.ok(rows[1].isExitRow());
    assert.ok(rows[2].isEntryRow());
    assert.equal(rows[2].signal, 'B');
    assert.ok(rows[3].isExitRow());
  });

  it('tradeNumber is 1-based and increments per trade across both its rows', () => {
    const trades = [makeClosedTrade(), makeClosedTrade(), makeClosedTrade()];
    const rows = transformAll(trades);
    assert.equal(rows[0].tradeNumber, 1);
    assert.equal(rows[1].tradeNumber, 1);
    assert.equal(rows[2].tradeNumber, 2);
    assert.equal(rows[3].tradeNumber, 2);
    assert.equal(rows[4].tradeNumber, 3);
    assert.equal(rows[5].tradeNumber, 3);
  });

  it('isOpen on exit rows reflects each trade status in mixed input', () => {
    const trades = [makeClosedTrade(), makeOpenTrade(), makeClosedTrade()];
    const rows = transformAll(trades);
    assert.equal(rows[1].isOpen, false);
    assert.equal(rows[3].isOpen, true);
    assert.equal(rows[5].isOpen, false);
  });
});
