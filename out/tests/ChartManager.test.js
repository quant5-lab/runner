import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { ChartManager } from '../js/ChartManager.js';

// ── Stubs ──────────────────────────────────────────────────────────────────────

function makeChartSpy() {
  const resizeCalls = [];
  const fitCalls    = [];
  return {
    resize:       (w, h) => resizeCalls.push({ w, h }),
    timeScale:    ()     => ({ fitContent: () => fitCalls.push(1) }),
    _resizeCalls: resizeCalls,
    _fitCalls:    fitCalls,
  };
}

function makeContainer(clientWidth, clientHeight) {
  return { clientWidth, clientHeight };
}

function stubLightweightCharts() {
  const calls     = [];
  const chartStub = {};
  global.LightweightCharts = {
    createChart: (container, options) => { calls.push({ container, options }); return chartStub; },
  };
  return { calls, chartStub };
}

// ── ChartManager.createChart — option composition ─────────────────────────────
//
// createChart spreads chartOptions then unconditionally overrides height and
// width from config and container respectively.  Non-conflicting chartOptions
// keys must pass through intact.

describe('createChart() — height comes from config, never from chartOptions', () => {
  it('config.height is used as the height option', () => {
    const { calls } = stubLightweightCharts();
    ChartManager.createChart({ clientWidth: 800 }, { height: 400 }, {});
    assert.equal(calls[0].options.height, 400);
  });

  it('config.height overrides a conflicting height key in chartOptions', () => {
    const { calls } = stubLightweightCharts();
    ChartManager.createChart({ clientWidth: 800 }, { height: 400 }, { height: 9999 });
    assert.equal(calls[0].options.height, 400);
  });
});

describe('createChart() — width comes from container.clientWidth, never from chartOptions', () => {
  it('container.clientWidth is used as the width option', () => {
    const { calls } = stubLightweightCharts();
    ChartManager.createChart({ clientWidth: 1200 }, { height: 400 }, {});
    assert.equal(calls[0].options.width, 1200);
  });

  it('container.clientWidth overrides a conflicting width key in chartOptions', () => {
    const { calls } = stubLightweightCharts();
    ChartManager.createChart({ clientWidth: 1200 }, { height: 400 }, { width: 1 });
    assert.equal(calls[0].options.width, 1200);
  });
});

describe('createChart() — chartOptions pass-through and LightweightCharts delegation', () => {
  it('non-conflicting chartOptions properties reach LightweightCharts.createChart', () => {
    const { calls } = stubLightweightCharts();
    ChartManager.createChart({ clientWidth: 800 }, { height: 400 }, { layout: { background: '#000' } });
    assert.deepEqual(calls[0].options.layout, { background: '#000' });
  });

  it('passes the container element as the first argument to LightweightCharts.createChart', () => {
    const { calls } = stubLightweightCharts();
    const container = { clientWidth: 800 };
    ChartManager.createChart(container, { height: 400 }, {});
    assert.equal(calls[0].container, container);
  });

  it('returns the chart instance from LightweightCharts.createChart', () => {
    const { chartStub } = stubLightweightCharts();
    const result = ChartManager.createChart({ clientWidth: 800 }, { height: 400 }, {});
    assert.equal(result, chartStub);
  });
});

// ── ChartManager.fitContent — timescale fit delegation ────────────────────────
//
// fitContent calls timeScale().fitContent() on every chart exactly once per
// invocation.  The implementation must iterate all charts, not only the first.

describe('fitContent() — zero charts: no error', () => {
  it('empty chart list produces no error', () => {
    assert.doesNotThrow(() => ChartManager.fitContent([]));
  });
});

describe('fitContent() — timeScale().fitContent() called exactly once per chart', () => {
  const chartCounts = [1, 2, 5];

  for (const n of chartCounts) {
    it(`${n} chart(s): each receives exactly one fitContent call`, () => {
      const charts = Array.from({ length: n }, makeChartSpy);
      ChartManager.fitContent(charts);
      charts.forEach((chart, i) => {
        assert.equal(chart._fitCalls.length, 1, `chart[${i}] fitContent call count`);
      });
    });
  }
});

// ── ChartManager.handleResize — per-container resize delegation ───────────────
//
// handleResize reads clientWidth and clientHeight from each chart's own
// container and calls chart.resize(width, height) exactly once per chart per
// invocation.  Width and height must come from the same container with no
// cross-contamination between panes.

describe('handleResize() — zero charts: no error', () => {
  it('empty chart and container lists produce no error', () => {
    assert.doesNotThrow(() => ChartManager.handleResize([], []));
  });
});

describe('handleResize() — each chart receives its own container dimensions', () => {
  const resizeCases = [
    {
      label:      '1 pane (main only)',
      containers: [makeContainer(800, 400)],
    },
    {
      label:      '2 panes (main + 1 indicator)',
      containers: [makeContainer(1024, 600), makeContainer(1024, 200)],
    },
    {
      label:      '5 panes — each with a distinct width and height',
      containers: [
        makeContainer(320,  568),
        makeContainer(768, 1024),
        makeContainer(1440, 900),
        makeContainer(2560, 1440),
        makeContainer(3840, 2160),
      ],
    },
  ];

  for (const { label, containers } of resizeCases) {
    it(`${label}: resize(clientWidth, clientHeight) called once per chart with matching dimensions`, () => {
      const charts = Array.from({ length: containers.length }, makeChartSpy);
      ChartManager.handleResize(charts, containers);
      charts.forEach((chart, i) => {
        assert.deepEqual(
          chart._resizeCalls,
          [{ w: containers[i].clientWidth, h: containers[i].clientHeight }],
          `chart[${i}]`,
        );
      });
    });
  }
});

describe('handleResize() — reads container state at call time', () => {
  it('successive calls each propagate the current container dimensions', () => {
    const chart     = makeChartSpy();
    const container = makeContainer(800, 400);

    ChartManager.handleResize([chart], [container]);

    container.clientHeight = 700;
    ChartManager.handleResize([chart], [container]);

    assert.deepEqual(chart._resizeCalls, [
      { w: 800, h: 400 },
      { w: 800, h: 700 },
    ]);
  });
});

describe('handleResize() — boundary dimensions', () => {
  it('zero width and height are propagated faithfully (collapsed or hidden pane)', () => {
    const chart = makeChartSpy();
    ChartManager.handleResize([chart], [makeContainer(0, 0)]);
    assert.deepEqual(chart._resizeCalls, [{ w: 0, h: 0 }]);
  });

  it('fractional sub-pixel HiDPI dimensions are propagated without rounding', () => {
    const chart = makeChartSpy();
    ChartManager.handleResize([chart], [makeContainer(1920.5, 1080.5)]);
    assert.deepEqual(chart._resizeCalls, [{ w: 1920.5, h: 1080.5 }]);
  });
});
