import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { TradeMarkerBuilder } from '../js/TradeMarkerBuilder.js';

const EXPECTED_PROFIT_COLOR = '#26a69a';
const EXPECTED_LOSS_COLOR   = '#ef5350';

// ── Fixtures ──────────────────────────────────────────────────────────────────

function makeCandlesticks(count) {
  return Array.from({ length: count }, (_, i) => ({ time: (i + 1) * 1000 }));
}

function makeClosedTrade(overrides = {}) {
  return {
    direction: 'long',
    profit:    100,
    entryBar:  0,
    entryTime: 1000,
    exitBar:   1,
    exitTime:  2000,
    ...overrides,
  };
}

function makeOpenTrade(overrides = {}) {
  return {
    direction: 'long',
    entryBar:  0,
    entryTime: 1000,
    ...overrides,
  };
}

function buildMarkers(trades = [], openTrades = [], candlesticks = makeCandlesticks(10)) {
  return new TradeMarkerBuilder().build({ trades, openTrades }, candlesticks);
}

// ── Guard conditions ──────────────────────────────────────────────────────────

describe('build() — guard conditions', () => {
  const guardCases = [
    { label: 'null strategy',           fn: () => new TradeMarkerBuilder().build(null,      makeCandlesticks(5)) },
    { label: 'undefined strategy',      fn: () => new TradeMarkerBuilder().build(undefined, makeCandlesticks(5)) },
    { label: 'null candlestick data',   fn: () => new TradeMarkerBuilder().build({ trades: [], openTrades: [] }, null) },
    { label: 'empty candlestick array', fn: () => new TradeMarkerBuilder().build({ trades: [], openTrades: [] }, []) },
    { label: 'empty trade lists',       fn: () => buildMarkers([], []) },
    { label: 'strategy with no fields', fn: () => new TradeMarkerBuilder().build({}, makeCandlesticks(5)) },
  ];

  for (const { label, fn } of guardCases) {
    it(`returns empty array for ${label}`, () => {
      assert.deepEqual(fn(), []);
    });
  }
});

// ── Sort order ────────────────────────────────────────────────────────────────

describe('build() — marker sort order', () => {
  it('sorts markers ascending by time regardless of trade input order', () => {
    const cd = makeCandlesticks(6);
    const trades = [
      makeClosedTrade({ entryBar: 4, exitBar: 5 }),
      makeClosedTrade({ entryBar: 0, exitBar: 1 }),
      makeClosedTrade({ entryBar: 2, exitBar: 3 }),
    ];
    const markers = buildMarkers(trades, [], cd);
    for (let i = 1; i < markers.length; i++) {
      assert.ok(
        markers[i].time >= markers[i - 1].time,
        `out-of-order at index ${i}: ${markers[i - 1].time} > ${markers[i].time}`,
      );
    }
  });

  it('emits all markers when multiple trades share the same bar', () => {
    const cd = makeCandlesticks(3);
    const trades = [
      makeClosedTrade({ entryBar: 0, exitBar: 1, profit:  50 }),
      makeClosedTrade({ entryBar: 0, exitBar: 1, profit: -30 }),
    ];
    assert.equal(buildMarkers(trades, [], cd).length, 4);
  });

  it('emits entry+exit for each closed trade', () => {
    const cd  = makeCandlesticks(10);
    const n   = 5;
    const trades = Array.from({ length: n }, (_, i) =>
      makeClosedTrade({ entryBar: i * 2, exitBar: i * 2 + 1 }),
    );
    assert.equal(buildMarkers(trades, [], cd).length, n * 2);
  });
});

// ── Entry marker — shape and position ─────────────────────────────────────────

describe('entry marker — shape and position', () => {
  const cases = [
    { direction: 'long',  shape: 'arrowUp',   position: 'belowBar' },
    { direction: 'short', shape: 'arrowDown',  position: 'aboveBar' },
  ];

  for (const { direction, shape, position } of cases) {
    it(`${direction} entry → shape=${shape}, position=${position}`, () => {
      const cd = makeCandlesticks(3);
      const markers = buildMarkers(
        [makeClosedTrade({ direction, entryBar: 0, exitBar: 1 })], [], cd,
      );
      const entry = markers.find(m => m.shape === shape);
      assert.ok(entry, `no marker with shape=${shape}`);
      assert.equal(entry.position, position);
    });
  }
});

// ── Entry marker — color follows direction ────────────────────────────────────

describe('entry marker — color axis is direction', () => {
  it('long entry → profit color regardless of trade P/L', () => {
    const cd = makeCandlesticks(3);
    for (const profit of [500, -500, 0]) {
      const entry = buildMarkers(
        [makeClosedTrade({ direction: 'long', profit, entryBar: 0, exitBar: 1 })], [], cd,
      ).find(m => m.shape === 'arrowUp');
      assert.equal(entry.color, EXPECTED_PROFIT_COLOR, `failed for profit=${profit}`);
    }
  });

  it('short entry → loss color regardless of trade P/L', () => {
    const cd = makeCandlesticks(3);
    for (const profit of [500, -500, 0]) {
      const entry = buildMarkers(
        [makeClosedTrade({ direction: 'short', profit, entryBar: 0, exitBar: 1 })], [], cd,
      ).find(m => m.shape === 'arrowDown');
      assert.equal(entry.color, EXPECTED_LOSS_COLOR, `failed for profit=${profit}`);
    }
  });
});

// ── Exit marker — shape and position ─────────────────────────────────────────

describe('exit marker — shape and position', () => {
  const cases = [
    { direction: 'long',  position: 'aboveBar' },
    { direction: 'short', position: 'belowBar' },
  ];

  for (const { direction, position } of cases) {
    it(`${direction} exit → shape=circle, position=${position}`, () => {
      const cd = makeCandlesticks(3);
      const exit = buildMarkers(
        [makeClosedTrade({ direction, entryBar: 0, exitBar: 1 })], [], cd,
      ).find(m => m.shape === 'circle');
      assert.ok(exit, 'no circle marker found');
      assert.equal(exit.shape,    'circle');
      assert.equal(exit.position, position);
    });
  }
});

// ── Exit marker — color axis is P/L ──────────────────────────────────────────

describe('exit marker — color axis is realized P/L', () => {
  const colorCases = [
    { profit:  100,   expected: EXPECTED_PROFIT_COLOR, label: 'positive profit' },
    { profit: -100,   expected: EXPECTED_LOSS_COLOR,   label: 'negative profit' },
    { profit:    0,   expected: EXPECTED_PROFIT_COLOR, label: 'zero profit (breakeven = not a loss)' },
    { profit:  0.01,  expected: EXPECTED_PROFIT_COLOR, label: 'fractional positive' },
    { profit: -0.01,  expected: EXPECTED_LOSS_COLOR,   label: 'fractional negative' },
    { profit:  1e9,   expected: EXPECTED_PROFIT_COLOR, label: 'very large profit' },
    { profit: -1e9,   expected: EXPECTED_LOSS_COLOR,   label: 'very large loss' },
  ];

  for (const { profit, expected, label } of colorCases) {
    it(`${label} (profit=${profit}) → correct color`, () => {
      const cd = makeCandlesticks(3);
      const exit = buildMarkers(
        [makeClosedTrade({ profit, entryBar: 0, exitBar: 1 })], [], cd,
      ).find(m => m.shape === 'circle');
      assert.equal(exit.color, expected);
    });
  }
});

// ── Exit marker — color independent of direction ──────────────────────────────

describe('exit marker — direction does not affect color', () => {
  const directionCases = [
    { direction: 'long',  profit:  100, expected: EXPECTED_PROFIT_COLOR, label: 'long + profit' },
    { direction: 'long',  profit: -100, expected: EXPECTED_LOSS_COLOR,   label: 'long + loss' },
    { direction: 'long',  profit:    0, expected: EXPECTED_PROFIT_COLOR, label: 'long + zero' },
    { direction: 'short', profit:  100, expected: EXPECTED_PROFIT_COLOR, label: 'short + profit' },
    { direction: 'short', profit: -100, expected: EXPECTED_LOSS_COLOR,   label: 'short + loss' },
    { direction: 'short', profit:    0, expected: EXPECTED_PROFIT_COLOR, label: 'short + zero' },
  ];

  for (const { direction, profit, expected, label } of directionCases) {
    it(label, () => {
      const cd = makeCandlesticks(3);
      const exit = buildMarkers(
        [makeClosedTrade({ direction, profit, entryBar: 0, exitBar: 1 })], [], cd,
      ).find(m => m.shape === 'circle');
      assert.equal(exit.color, expected);
    });
  }
});

// ── Open trades ───────────────────────────────────────────────────────────────

describe('open trade markers', () => {
  it('produces exactly one marker per open trade (entry only, no exit)', () => {
    const cd = makeCandlesticks(5);
    assert.equal(buildMarkers([], [makeOpenTrade({ entryBar: 2 })], cd).length, 1);
    assert.equal(buildMarkers([], [makeOpenTrade({ entryBar: 2 }), makeOpenTrade({ entryBar: 3 })], cd).length, 2);
  });

  const openShapeCases = [
    { direction: 'long',  shape: 'arrowUp',  position: 'belowBar' },
    { direction: 'short', shape: 'arrowDown', position: 'aboveBar' },
  ];

  for (const { direction, shape, position } of openShapeCases) {
    it(`open ${direction} entry → shape=${shape}, position=${position}`, () => {
      const cd = makeCandlesticks(5);
      const [marker] = buildMarkers([], [makeOpenTrade({ direction, entryBar: 2 })], cd);
      assert.equal(marker.shape,    shape);
      assert.equal(marker.position, position);
    });
  }

  it('open trade entry color is identical to closed trade entry color for same direction', () => {
    const cd = makeCandlesticks(5);
    for (const direction of ['long', 'short']) {
      const shape       = direction === 'long' ? 'arrowUp' : 'arrowDown';
      const openColor   = buildMarkers([], [makeOpenTrade({ direction, entryBar: 2 })], cd)[0].color;
      const closedEntry = buildMarkers(
        [makeClosedTrade({ direction, entryBar: 2, exitBar: 3 })], [], cd,
      ).find(m => m.shape === shape);
      assert.equal(openColor, closedEntry.color, `color mismatch for direction=${direction}`);
    }
  });
});

// ── Time resolution ───────────────────────────────────────────────────────────

describe('time resolution — barIndex to candlestick time', () => {
  const cd = makeCandlesticks(5);

  const boundCases = [
    { barIndex: 0,           expected: 1000, label: 'first bar (barIndex=0)' },
    { barIndex: 4,           expected: 5000, label: 'last bar (barIndex=length-1)' },
    { barIndex: 2,           expected: 3000, label: 'mid-range barIndex' },
  ];

  for (const { barIndex, expected, label } of boundCases) {
    it(`${label} → uses candlestick time`, () => {
      const [marker] = buildMarkers([], [makeOpenTrade({ entryBar: barIndex })], cd);
      assert.equal(marker.time, expected);
    });
  }

  it('negative barIndex falls back to unixTime', () => {
    const [marker] = buildMarkers([], [makeOpenTrade({ entryBar: -1, entryTime: 9999 })], cd);
    assert.equal(marker.time, 9999);
  });

  it('barIndex >= candlestick length falls back to unixTime', () => {
    const [marker] = buildMarkers([], [makeOpenTrade({ entryBar: 5, entryTime: 8888 })], cd);
    assert.equal(marker.time, 8888);
  });

  it('OOB barIndex with falsy unixTime produces no marker', () => {
    assert.equal(buildMarkers([], [makeOpenTrade({ entryBar: 99, entryTime: 0 })], cd).length, 0);
  });
});

// ── Color axis invariants ─────────────────────────────────────────────────────

describe('color axis invariants — profit and direction axes share palette', () => {
  it('exit profit color equals long entry color', () => {
    const cd = makeCandlesticks(3);
    const longEntry  = buildMarkers([makeClosedTrade({ direction: 'long',  entryBar: 0, exitBar: 1 })], [], cd).find(m => m.shape === 'arrowUp');
    const profitExit = buildMarkers([makeClosedTrade({ direction: 'short', profit: 1, entryBar: 0, exitBar: 1 })], [], cd).find(m => m.shape === 'circle');
    assert.equal(profitExit.color, longEntry.color);
  });

  it('exit loss color equals short entry color', () => {
    const cd = makeCandlesticks(3);
    const shortEntry = buildMarkers([makeClosedTrade({ direction: 'short', entryBar: 0, exitBar: 1 })], [], cd).find(m => m.shape === 'arrowDown');
    const lossExit   = buildMarkers([makeClosedTrade({ direction: 'long',  profit: -1, entryBar: 0, exitBar: 1 })], [], cd).find(m => m.shape === 'circle');
    assert.equal(lossExit.color, shortEntry.color);
  });

  it('profit color and loss color are visually distinct', () => {
    const cd         = makeCandlesticks(3);
    const profitExit = buildMarkers([makeClosedTrade({ profit:  1, entryBar: 0, exitBar: 1 })], [], cd).find(m => m.shape === 'circle');
    const lossExit   = buildMarkers([makeClosedTrade({ profit: -1, entryBar: 0, exitBar: 1 })], [], cd).find(m => m.shape === 'circle');
    assert.notEqual(profitExit.color, lossExit.color);
  });
});
