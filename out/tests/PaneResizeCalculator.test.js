import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { PaneResizeCalculator, MIN_PANE_HEIGHT } from '../js/PaneResizeCalculator.js';

// ── Helper ────────────────────────────────────────────────────────────────────

function calc(above, below, delta) {
  return PaneResizeCalculator.calculate(above, below, delta);
}

// ── Zero delta ────────────────────────────────────────────────────────────────

describe('calculate() — zero delta: both heights returned unchanged', () => {
  it('above and below equal their inputs', () => {
    assert.deepEqual(calc(400, 200, 0), { above: 400, below: 200 });
  });
});

// ── Unclamped drag ────────────────────────────────────────────────────────────

describe('calculate() — drag within bounds: above and below change by equal and opposite amounts', () => {
  const cases = [
    { delta: 50,  expectedAbove: 450, expectedBelow: 150, label: 'drag down' },
    { delta: -50, expectedAbove: 350, expectedBelow: 250, label: 'drag up'   },
  ];

  for (const { delta, expectedAbove, expectedBelow, label } of cases) {
    it(`${label}: delta=${delta}`, () => {
      const { above, below } = calc(400, 200, delta);
      assert.equal(above, expectedAbove);
      assert.equal(below, expectedBelow);
    });
  }
});

// ── Total conservation ────────────────────────────────────────────────────────

describe('calculate() — total height (above + below) is always conserved', () => {
  const deltas = [-9999, -320, -50, 0, 50, 120, 9999];

  for (const delta of deltas) {
    it(`delta=${delta}: above + below equals 600`, () => {
      const { above, below } = calc(400, 200, delta);
      assert.equal(above + below, 600);
    });
  }
});

// ── Clamping ──────────────────────────────────────────────────────────────────

describe('calculate() — clamp: neither pane shrinks below MIN_PANE_HEIGHT', () => {
  const cases = [
    {
      delta:         9999,
      expectedAbove: 400 + 200 - MIN_PANE_HEIGHT,
      expectedBelow: MIN_PANE_HEIGHT,
      label:         'delta exceeds below headroom',
    },
    {
      delta:         -9999,
      expectedAbove: MIN_PANE_HEIGHT,
      expectedBelow: 400 + 200 - MIN_PANE_HEIGHT,
      label:         'delta exceeds above headroom',
    },
  ];

  for (const { delta, expectedAbove, expectedBelow, label } of cases) {
    it(`${label}: both panes land at boundary values`, () => {
      const { above, below } = calc(400, 200, delta);
      assert.equal(above, expectedAbove);
      assert.equal(below, expectedBelow);
    });
  }
});

// ── Already at minimum ────────────────────────────────────────────────────────

describe('calculate() — pane already at MIN_PANE_HEIGHT: further shrink is blocked', () => {
  it('above at min: upward drag produces no change', () => {
    assert.deepEqual(
      calc(MIN_PANE_HEIGHT, 400, -100),
      { above: MIN_PANE_HEIGHT, below: 400 },
    );
  });

  it('below at min: downward drag produces no change', () => {
    assert.deepEqual(
      calc(400, MIN_PANE_HEIGHT, 100),
      { above: 400, below: MIN_PANE_HEIGHT },
    );
  });
});

// ── Exact boundary ────────────────────────────────────────────────────────────

describe('calculate() — delta exactly at headroom boundary: pane lands on MIN_PANE_HEIGHT', () => {
  it('drag down by exactly (below - MIN_PANE_HEIGHT): below lands on MIN_PANE_HEIGHT', () => {
    const headroom = 200 - MIN_PANE_HEIGHT;
    assert.equal(calc(400, 200, headroom).below, MIN_PANE_HEIGHT);
  });

  it('drag up by exactly (above - MIN_PANE_HEIGHT): above lands on MIN_PANE_HEIGHT', () => {
    const headroom = 400 - MIN_PANE_HEIGHT;
    assert.equal(calc(400, 200, -headroom).above, MIN_PANE_HEIGHT);
  });
});
