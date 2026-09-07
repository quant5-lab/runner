import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { TradeRowspanTransformer } from '../js/TradeRowspanTransformer.js';
import { SortDirection           } from '../js/SortDirection.js';

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

// ── Call helpers ───────────────────────────────────────────────────────────────

function transform(trade, { formatter, currentPrice = null, tradeNumber = 1, sortDirection = SortDirection.DESC } = {}) {
  return new TradeRowspanTransformer(formatter ?? makeFormatter())
    .transformTrade(trade, tradeNumber, currentPrice, sortDirection);
}

function transformAll(trades, { formatter, currentPrice = null, sortDirection = SortDirection.DESC } = {}) {
  return new TradeRowspanTransformer(formatter ?? makeFormatter())
    .transformTrades(trades, currentPrice, sortDirection);
}

// ── Row accessors — type-based, position-independent ──────────────────────────

function entryRowOf(rows) { return rows.find(r => r.isEntryRow()); }
function exitRowOf(rows)  { return rows.find(r => r.isExitRow());  }

// ── Signal fallback case table ─────────────────────────────────────────────────

const SIGNAL_CASES = (commentKey, idKey) => [
  { label: `${commentKey} takes precedence over ${idKey}`, overrides: { [commentKey]: 'COMMENT', [idKey]: 'ID'      }, expected: 'COMMENT' },
  { label: `${idKey} used when ${commentKey} absent`,      overrides: {                          [idKey]: 'ID'      }, expected: 'ID'      },
  { label: 'empty string when neither present',             overrides: { [commentKey]: undefined, [idKey]: undefined }, expected: ''        },
];

// ── Row count ─────────────────────────────────────────────────────────────────

describe('transformTrade() — row count', () => {
  it('produces exactly 2 rows for a closed trade', () => {
    assert.equal(transform(makeClosedTrade()).length, 2);
  });

  it('produces exactly 2 rows for an open trade', () => {
    assert.equal(transform(makeOpenTrade()).length, 2);
  });

  it('always produces one entry row and one exit row regardless of trade status or sort direction', () => {
    for (const dir of [SortDirection.ASC, SortDirection.DESC]) {
      for (const trade of [makeClosedTrade(), makeOpenTrade()]) {
        const rows = transform(trade, { sortDirection: dir });
        assert.equal(rows.filter(r => r.isEntryRow()).length, 1, `dir=${dir} entryRow count`);
        assert.equal(rows.filter(r => r.isExitRow()).length,  1, `dir=${dir} exitRow count`);
      }
    }
  });
});

// ── Pair order and isPrimary ───────────────────────────────────────────────────

describe('transformTrade() — pair order and isPrimary by sort direction', () => {
  it('ASC: first row is entry (isPrimary=true), second is exit (isPrimary=false)', () => {
    const [first, second] = transform(makeClosedTrade(), { sortDirection: SortDirection.ASC });
    assert.ok(first.isEntryRow(),  'first must be entry');
    assert.ok(second.isExitRow(),  'second must be exit');
    assert.equal(first.isPrimary,  true,  'first is primary');
    assert.equal(second.isPrimary, false, 'second is not primary');
  });

  it('DESC: first row is exit (isPrimary=true), second is entry (isPrimary=false)', () => {
    const [first, second] = transform(makeClosedTrade(), { sortDirection: SortDirection.DESC });
    assert.ok(first.isExitRow(),   'first must be exit');
    assert.ok(second.isEntryRow(), 'second must be entry');
    assert.equal(first.isPrimary,  true,  'first is primary');
    assert.equal(second.isPrimary, false, 'second is not primary');
  });

  it('exactly one primary and one secondary row per pair for both directions', () => {
    for (const dir of [SortDirection.ASC, SortDirection.DESC]) {
      const rows = transform(makeClosedTrade(), { sortDirection: dir });
      assert.equal(rows.filter(r =>  r.isPrimary).length, 1, `dir=${dir} primary count`);
      assert.equal(rows.filter(r => !r.isPrimary).length, 1, `dir=${dir} secondary count`);
    }
  });
});

// ── tradeNumber propagation ────────────────────────────────────────────────────

describe('transformTrade() — tradeNumber propagation', () => {
  it('propagates to both rows for both sort directions', () => {
    for (const dir of [SortDirection.ASC, SortDirection.DESC]) {
      const rows = transform(makeClosedTrade(), { tradeNumber: 7, sortDirection: dir });
      assert.equal(entryRowOf(rows).tradeNumber, 7);
      assert.equal(exitRowOf(rows).tradeNumber,  7);
    }
  });
});

// ── Entry row fields ───────────────────────────────────────────────────────────

describe('transformTrade() — entry row fields', () => {
  it('direction propagates to entry row for both trade directions', () => {
    for (const direction of ['long', 'short']) {
      assert.equal(entryRowOf(transform(makeClosedTrade({ direction }))).direction, direction);
    }
  });

  it('direction propagates to exit row for both trade directions', () => {
    for (const direction of ['long', 'short']) {
      assert.equal(exitRowOf(transform(makeClosedTrade({ direction }))).direction, direction);
    }
  });

  it('entry dateTime delegates to formatter.getTradeDate(trade, true)', () => {
    const formatter = makeFormatter({ getTradeDate: (_t, isEntry) => isEntry ? 'ENTRY-TS' : 'EXIT-TS' });
    assert.equal(entryRowOf(transform(makeClosedTrade(), { formatter })).dateTime, 'ENTRY-TS');
  });

  it('entry price delegates to formatter.formatPrice(entryPrice)', () => {
    const formatter = makeFormatter({ formatPrice: (p) => `FMT:${p}` });
    assert.equal(entryRowOf(transform(makeClosedTrade({ entryPrice: 999 }), { formatter })).price, 'FMT:999');
  });

  it('entry size is trade.size.toFixed(2) for representative values', () => {
    for (const [size, expected] of [[2.5, '2.50'], [1, '1.00'], [0, '0.00'], [10.123, '10.12']]) {
      assert.equal(entryRowOf(transform(makeClosedTrade({ size }))).size, expected, `size=${size}`);
    }
  });

  it('entry isOpen is always false regardless of trade status', () => {
    assert.equal(entryRowOf(transform(makeClosedTrade())).isOpen, false);
    assert.equal(entryRowOf(transform(makeOpenTrade())).isOpen,   false);
  });

  it('entry P/L fields are always zeroed regardless of trade profit', () => {
    const row = entryRowOf(transform(makeClosedTrade({ profit: 9999 })));
    assert.equal(row.profitLoss, '');
    assert.equal(row.profitRaw,  0);
  });
});

// ── Size on both rows ──────────────────────────────────────────────────────────

describe('transformTrade() — size populated on both rows', () => {
  it('closed trade: both rows carry the same formatted size for both sort directions', () => {
    for (const dir of [SortDirection.ASC, SortDirection.DESC]) {
      const rows = transform(makeClosedTrade({ size: 2.5 }), { sortDirection: dir });
      assert.equal(entryRowOf(rows).size, '2.50', `dir=${dir} entry size`);
      assert.equal(exitRowOf(rows).size,  '2.50', `dir=${dir} exit size`);
    }
  });

  it('open trade: both rows carry the same formatted size for both sort directions', () => {
    for (const dir of [SortDirection.ASC, SortDirection.DESC]) {
      const rows = transform(makeOpenTrade({ size: 3 }), { sortDirection: dir });
      assert.equal(entryRowOf(rows).size, '3.00', `dir=${dir} entry size`);
      assert.equal(exitRowOf(rows).size,  '3.00', `dir=${dir} exit size`);
    }
  });
});

// ── Entry signal fallback ──────────────────────────────────────────────────────

describe('transformTrade() — entry signal fallback (entryComment > entryId > "")', () => {
  for (const { label, overrides, expected } of SIGNAL_CASES('entryComment', 'entryId')) {
    it(label, () => {
      assert.equal(entryRowOf(transform(makeClosedTrade(overrides))).signal, expected);
    });
  }
});

// ── Signal fields are sort-direction invariant ─────────────────────────────────

describe('transformTrade() — signal fields are sort-direction invariant', () => {
  it('entry signal is identical for ASC and DESC', () => {
    for (const overrides of [
      { entryComment: 'COMMENT', entryId: 'ID' },
      {                          entryId: 'ID' },
      { entryComment: undefined, entryId: undefined },
    ]) {
      const asc  = entryRowOf(transform(makeClosedTrade(overrides), { sortDirection: SortDirection.ASC  }));
      const desc = entryRowOf(transform(makeClosedTrade(overrides), { sortDirection: SortDirection.DESC }));
      assert.equal(asc.signal, desc.signal, `overrides=${JSON.stringify(overrides)}`);
    }
  });

  it('exit signal is identical for ASC and DESC for closed trades', () => {
    for (const overrides of [
      { exitComment: 'COMMENT', exitId: 'ID' },
      {                         exitId: 'ID' },
      { exitComment: undefined, exitId: undefined },
    ]) {
      const asc  = exitRowOf(transform(makeClosedTrade(overrides), { sortDirection: SortDirection.ASC  }));
      const desc = exitRowOf(transform(makeClosedTrade(overrides), { sortDirection: SortDirection.DESC }));
      assert.equal(asc.signal, desc.signal, `overrides=${JSON.stringify(overrides)}`);
    }
  });

  it('open trade exit signal is always empty string for both directions', () => {
    for (const dir of [SortDirection.ASC, SortDirection.DESC]) {
      assert.equal(exitRowOf(transform(makeOpenTrade({ exitComment: 'x', exitId: 'y' }), { sortDirection: dir })).signal, '');
    }
  });
});

// ── Closed exit row fields ─────────────────────────────────────────────────────

describe('transformTrade() — closed exit row fields', () => {
  it('exit isOpen is false', () => {
    assert.equal(exitRowOf(transform(makeClosedTrade())).isOpen, false);
  });

  it('exit dateTime delegates to formatter.getTradeDate(trade, false)', () => {
    const formatter = makeFormatter({ getTradeDate: (_t, isEntry) => isEntry ? 'ENTRY-TS' : 'EXIT-TS' });
    assert.equal(exitRowOf(transform(makeClosedTrade(), { formatter })).dateTime, 'EXIT-TS');
  });

  it('exit profitRaw equals trade.profit for representative values', () => {
    for (const profit of [42, -42, 0, 0.001, -0.001]) {
      assert.equal(exitRowOf(transform(makeClosedTrade({ profit }))).profitRaw, profit, `profit=${profit}`);
    }
  });

  it('exit profitLoss delegates to formatter.formatProfit(trade.profit)', () => {
    const formatter = makeFormatter({ formatProfit: (v) => `FORMATTED:${v}` });
    assert.equal(exitRowOf(transform(makeClosedTrade({ profit: 55 }), { formatter })).profitLoss, 'FORMATTED:55');
  });

  it('exit price delegates to formatter.formatPrice(exitPrice)', () => {
    const formatter = makeFormatter({ formatPrice: (p) => `FMT:${p}` });
    assert.equal(exitRowOf(transform(makeClosedTrade({ exitPrice: 222 }), { formatter })).price, 'FMT:222');
  });

  it('exit price falls back to 0 when exitPrice is null', () => {
    const formatter = makeFormatter({ formatPrice: (p) => `FMT:${p}` });
    assert.equal(exitRowOf(transform(makeClosedTrade({ exitPrice: null }), { formatter })).price, 'FMT:0');
  });
});

// ── Exit signal fallback ───────────────────────────────────────────────────────

describe('transformTrade() — exit signal fallback (exitComment > exitId > "")', () => {
  for (const { label, overrides, expected } of SIGNAL_CASES('exitComment', 'exitId')) {
    it(label, () => {
      assert.equal(exitRowOf(transform(makeClosedTrade(overrides))).signal, expected);
    });
  }
});

// ── Open trade exit row ────────────────────────────────────────────────────────

describe('transformTrade() — open trade exit row', () => {
  it('exit isOpen is true', () => {
    assert.equal(exitRowOf(transform(makeOpenTrade())).isOpen, true);
  });

  it('exit dateTime is literal "Open"', () => {
    assert.equal(exitRowOf(transform(makeOpenTrade())).dateTime, 'Open');
  });

  it('exit signal is always empty string regardless of exitComment/exitId', () => {
    assert.equal(exitRowOf(transform(makeOpenTrade({ exitComment: 'x', exitId: 'y' }))).signal, '');
  });

  it('exit profitRaw delegates to formatter.calculateUnrealizedProfit with correct arguments', () => {
    let captured = null;
    const formatter = makeFormatter({
      calculateUnrealizedProfit: (trade, price) => { captured = { trade, price }; return 77; },
    });
    const trade = makeOpenTrade();
    assert.equal(exitRowOf(transform(trade, { formatter, currentPrice: 200 })).profitRaw, 77);
    assert.equal(captured.price, 200);
    assert.equal(captured.trade, trade);
  });

  it('exit price uses currentPrice when provided', () => {
    const formatter = makeFormatter({ formatPrice: (p) => `FMT:${p}` });
    assert.equal(
      exitRowOf(transform(makeOpenTrade({ entryPrice: 50 }), { formatter, currentPrice: 120 })).price,
      'FMT:120',
    );
  });

  it('exit price falls back to entryPrice when currentPrice is null', () => {
    const formatter = makeFormatter({ formatPrice: (p) => `FMT:${p}` });
    assert.equal(
      exitRowOf(transform(makeOpenTrade({ entryPrice: 50 }), { formatter, currentPrice: null })).price,
      'FMT:50',
    );
  });

  it('exit price uses 0 (not entryPrice) when currentPrice is exactly 0', () => {
    const formatter = makeFormatter({ formatPrice: (p) => `FMT:${p}` });
    assert.equal(
      exitRowOf(transform(makeOpenTrade({ entryPrice: 50 }), { formatter, currentPrice: 0 })).price,
      'FMT:0',
    );
  });
});

// ── transformTrades() — row count ─────────────────────────────────────────────

describe('transformTrades() — row count', () => {
  it('produces 2N rows for N trades for both sort directions', () => {
    for (const dir of [SortDirection.ASC, SortDirection.DESC]) {
      for (const n of [0, 1, 3, 5]) {
        const trades = Array.from({ length: n }, () => makeClosedTrade());
        assert.equal(transformAll(trades, { sortDirection: dir }).length, n * 2, `dir=${dir} n=${n}`);
      }
    }
  });
});

// ── transformTrades() — pair order and per-trade isPrimary consistency ─────────

describe('transformTrades() — trade order preserved across sort directions', () => {
  it('ASC: pairs appear in input order, each pair entry-first', () => {
    const trades = [makeClosedTrade({ entryId: 'A' }), makeClosedTrade({ entryId: 'B' })];
    const rows   = transformAll(trades, { sortDirection: SortDirection.ASC });
    assert.ok(rows[0].isEntryRow());  assert.equal(rows[0].signal, 'A');
    assert.ok(rows[1].isExitRow());
    assert.ok(rows[2].isEntryRow());  assert.equal(rows[2].signal, 'B');
    assert.ok(rows[3].isExitRow());
  });

  it('DESC: pairs appear in input order, each pair exit-first', () => {
    const trades = [makeClosedTrade({ entryId: 'A' }), makeClosedTrade({ entryId: 'B' })];
    const rows   = transformAll(trades, { sortDirection: SortDirection.DESC });
    assert.ok(rows[0].isExitRow());
    assert.ok(rows[1].isEntryRow());  assert.equal(rows[1].signal, 'A');
    assert.ok(rows[2].isExitRow());
    assert.ok(rows[3].isEntryRow());  assert.equal(rows[3].signal, 'B');
  });
});

describe('transformTrades() — isPrimary consistency across N trades', () => {
  it('every pair has exactly one primary row regardless of N and sort direction', () => {
    for (const dir of [SortDirection.ASC, SortDirection.DESC]) {
      for (const n of [1, 2, 5]) {
        const trades = Array.from({ length: n }, () => makeClosedTrade());
        const rows   = transformAll(trades, { sortDirection: dir });
        for (let i = 0; i < n; i++) {
          const pair = rows.slice(i * 2, i * 2 + 2);
          assert.equal(
            pair.filter(r =>  r.isPrimary).length, 1,
            `dir=${dir} trade=${i} primary count`,
          );
          assert.equal(
            pair.filter(r => !r.isPrimary).length, 1,
            `dir=${dir} trade=${i} secondary count`,
          );
        }
      }
    }
  });
});

// ── transformTrades() — tradeNumber ───────────────────────────────────────────

describe('transformTrades() — tradeNumber is 1-based and position-derived', () => {
  it('tradeNumber increments per input trade and spans both its rows, for both directions', () => {
    for (const dir of [SortDirection.ASC, SortDirection.DESC]) {
      const trades  = [makeClosedTrade(), makeClosedTrade(), makeClosedTrade()];
      const numbers = transformAll(trades, { sortDirection: dir }).map(r => r.tradeNumber);
      assert.deepEqual(numbers, [1, 1, 2, 2, 3, 3], `dir=${dir}`);
    }
  });
});

// ── transformTrades() — isOpen propagation ────────────────────────────────────

describe('transformTrades() — isOpen on exit rows reflects per-trade status', () => {
  it('each exit row reflects its own trade status in mixed closed/open input', () => {
    const trades = [makeClosedTrade(), makeOpenTrade(), makeClosedTrade()];
    for (const dir of [SortDirection.ASC, SortDirection.DESC]) {
      const exitRows = transformAll(trades, { sortDirection: dir }).filter(r => r.isExitRow());
      assert.equal(exitRows[0].isOpen, false, `dir=${dir} trade0`);
      assert.equal(exitRows[1].isOpen, true,  `dir=${dir} trade1`);
      assert.equal(exitRows[2].isOpen, false, `dir=${dir} trade2`);
    }
  });
});
