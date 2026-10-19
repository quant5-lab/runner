import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { PaneManager } from '../js/PaneManager.js';

// ── Stubs ─────────────────────────────────────────────────────────────────────

function makeChartStub() {
  return {
    resize:    () => {},
    timeScale: () => ({ subscribeVisibleLogicalRangeChange: () => {} }),
  };
}

function makeLightweightCharts() {
  global.LightweightCharts = { createChart: () => makeChartStub() };
}

function makeContainer(clientWidth = 1000) {
  return { clientWidth, style: {} };
}

function makeDocumentStub() {
  const appended    = [];
  const chartHolder = { appendChild: el => appended.push(el) };
  global.document   = {
    createElement: () => ({ id: '', style: {} }),
    querySelector: () => chartHolder,
  };
}

// ── createMainPane() — style.height ──────────────────────────────────────────

describe('createMainPane() — stamps style.height from config', () => {
  const heights = [1, 200, 300, 400, 500];

  for (const height of heights) {
    it(`config.height=${height}: container.style.height is '${height}px'`, () => {
      makeLightweightCharts();
      const container = makeContainer();
      new PaneManager({}).createMainPane(container, { height });
      assert.equal(container.style.height, `${height}px`);
    });
  }
});

// ── createDynamicPane() — style.height ───────────────────────────────────────

describe('createDynamicPane() — stamps style.height from config', () => {
  const heights = [150, 200, 300];

  for (const height of heights) {
    it(`config.height=${height}: created container.style.height is '${height}px'`, () => {
      makeLightweightCharts();
      makeDocumentStub();
      const manager = new PaneManager({});
      manager.createMainPane(makeContainer(), { height: 400 });
      const { container } = manager.createDynamicPane('pane', { height });
      assert.equal(container.style.height, `${height}px`);
    });
  }
});

// ── style.height — stable after creation ─────────────────────────────────────

describe('style.height — not affected by subsequent mutations to clientHeight', () => {
  it('main pane: style.height remains at stamped value after clientHeight is overwritten', () => {
    makeLightweightCharts();
    const container = makeContainer();
    new PaneManager({}).createMainPane(container, { height: 400 });
    container.clientHeight = 700;
    assert.equal(container.style.height, '400px');
  });

  it('dynamic pane: style.height remains at stamped value after clientHeight is overwritten', () => {
    makeLightweightCharts();
    makeDocumentStub();
    const manager = new PaneManager({});
    manager.createMainPane(makeContainer(), { height: 400 });
    const { container } = manager.createDynamicPane('vol', { height: 150 });
    container.clientHeight = 700;
    assert.equal(container.style.height, '150px');
  });

  it('each pane carries its own style.height independently', () => {
    makeLightweightCharts();
    makeDocumentStub();
    const manager  = new PaneManager({});
    const mainCont = makeContainer();
    manager.createMainPane(mainCont, { height: 400 });
    const { container: dynCont } = manager.createDynamicPane('rsi', { height: 200 });
    assert.equal(mainCont.style.height, '400px');
    assert.equal(dynCont.style.height,  '200px');
    assert.notEqual(mainCont.style.height, dynCont.style.height);
  });
});

// ── getAllPanes() — shape and order ───────────────────────────────────────────

describe('getAllPanes() — returns [{container, chart}, ...] in creation order', () => {
  it('main-only: length is 1', () => {
    makeLightweightCharts();
    const manager = new PaneManager({});
    manager.createMainPane(makeContainer(), { height: 400 });
    assert.equal(manager.getAllPanes().length, 1);
  });

  it('main + 1 dynamic: length is 2', () => {
    makeLightweightCharts();
    makeDocumentStub();
    const manager = new PaneManager({});
    manager.createMainPane(makeContainer(), { height: 400 });
    manager.createDynamicPane('rsi', { height: 200 });
    assert.equal(manager.getAllPanes().length, 2);
  });

  it('main + 3 dynamic: length is 4', () => {
    makeLightweightCharts();
    makeDocumentStub();
    const manager = new PaneManager({});
    manager.createMainPane(makeContainer(), { height: 400 });
    manager.createDynamicPane('a', { height: 200 });
    manager.createDynamicPane('b', { height: 200 });
    manager.createDynamicPane('c', { height: 200 });
    assert.equal(manager.getAllPanes().length, 4);
  });

  it('main pane is always index 0', () => {
    makeLightweightCharts();
    const manager  = new PaneManager({});
    const mainCont = makeContainer();
    manager.createMainPane(mainCont, { height: 400 });
    assert.equal(manager.getAllPanes()[0].container, mainCont);
  });

  it('dynamic panes appear after main in insertion order', () => {
    makeLightweightCharts();
    makeDocumentStub();
    const manager = new PaneManager({});
    manager.createMainPane(makeContainer(), { height: 400 });
    const { container: c1 } = manager.createDynamicPane('first',  { height: 200 });
    const { container: c2 } = manager.createDynamicPane('second', { height: 200 });
    const { container: c3 } = manager.createDynamicPane('third',  { height: 200 });
    const panes = manager.getAllPanes();
    assert.equal(panes[1].container, c1);
    assert.equal(panes[2].container, c2);
    assert.equal(panes[3].container, c3);
  });

  it('every entry exposes both container and chart', () => {
    makeLightweightCharts();
    const manager = new PaneManager({});
    manager.createMainPane(makeContainer(), { height: 400 });
    const [pane] = manager.getAllPanes();
    assert.ok('container' in pane);
    assert.ok('chart'     in pane);
  });

  it('is consistent with getAllContainers() and getAllCharts()', () => {
    makeLightweightCharts();
    makeDocumentStub();
    const manager = new PaneManager({});
    manager.createMainPane(makeContainer(), { height: 400 });
    manager.createDynamicPane('macd', { height: 200 });
    manager.createDynamicPane('rsi',  { height: 200 });
    const panes      = manager.getAllPanes();
    const containers = manager.getAllContainers();
    const charts     = manager.getAllCharts();
    panes.forEach((p, i) => {
      assert.equal(p.container, containers[i], `container[${i}]`);
      assert.equal(p.chart,     charts[i],     `chart[${i}]`);
    });
  });
});
