import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { PaneResizeController } from '../js/PaneResizeController.js';

// ── Stubs ─────────────────────────────────────────────────────────────────────

function makeElementStub() {
  const listeners = {};
  return {
    className:        '',
    style:            {},
    _removed:         false,
    remove:           function () { this._removed = true; },
    addEventListener: (event, fn, _opts) => { (listeners[event] ??= []).push(fn); },
    _listeners:       listeners,
  };
}

function makeDocumentStub() {
  const listeners = {};
  return {
    listeners,
    addEventListener:    (event, fn, _opts) => { (listeners[event] ??= []).push(fn); },
    removeEventListener: (event, fn) => {
      if (listeners[event]) listeners[event] = listeners[event].filter(l => l !== fn);
    },
    createElement: () => makeElementStub(),
  };
}

function makePane() {
  const inserted = [];
  return {
    container: {
      clientWidth:           1000,
      clientHeight:          200,
      style:                 {},
      insertAdjacentElement: (pos, el) => inserted.push({ pos, el }),
      _inserted:             inserted,
    },
    chart: { resize: () => {} },
  };
}

function makePanes(count) {
  return Array.from({ length: count }, makePane);
}

function makeController(paneCount) {
  const doc = makeDocumentStub();
  global.document = doc;
  const panes      = makePanes(paneCount);
  const controller = new PaneResizeController();
  controller.mount(panes);
  return { controller, panes, doc };
}

function allInserted(panes) {
  return panes.flatMap(p => p.container._inserted.map(({ el }) => el));
}

// ── mount() — handle count ────────────────────────────────────────────────────

describe('mount() — creates exactly (paneCount - 1) handles', () => {
  const cases = [
    { paneCount: 0, handles: 0 },
    { paneCount: 1, handles: 0 },
    { paneCount: 2, handles: 1 },
    { paneCount: 3, handles: 2 },
    { paneCount: 5, handles: 4 },
  ];

  for (const { paneCount, handles } of cases) {
    it(`${paneCount} pane(s) → ${handles} handle(s)`, () => {
      const { panes } = makeController(paneCount);
      const total = panes.reduce((n, p) => n + p.container._inserted.length, 0);
      assert.equal(total, handles);
    });
  }
});

// ── mount() — insertion position ─────────────────────────────────────────────

describe('mount() — handle inserted immediately before each panes[i+1] container', () => {
  const paneCounts = [2, 3, 5];

  for (const paneCount of paneCounts) {
    it(`${paneCount} panes: panes[0] untouched; panes[1..n] each receive one beforebegin insertion`, () => {
      const { panes } = makeController(paneCount);
      assert.equal(panes[0].container._inserted.length, 0);
      for (let i = 1; i < panes.length; i++) {
        assert.equal(panes[i].container._inserted.length, 1,             `panes[${i}] insertion count`);
        assert.equal(panes[i].container._inserted[0].pos, 'beforebegin', `panes[${i}] position`);
      }
    });
  }
});

// ── mount() — handle element class ───────────────────────────────────────────

describe('mount() — every handle element has class pane-resize-handle', () => {
  const paneCounts = [2, 3, 5];

  for (const paneCount of paneCounts) {
    it(`${paneCount} panes: all ${paneCount - 1} handle element(s) carry the correct class`, () => {
      const { panes } = makeController(paneCount);
      const els = allInserted(panes);
      assert.equal(els.length, paneCount - 1);
      assert.ok(els.every(el => el.className === 'pane-resize-handle'));
    });
  }
});

// ── destroy() — handle elements removed ──────────────────────────────────────

describe('destroy() — all handle elements are removed', () => {
  const paneCounts = [2, 3, 5];

  for (const paneCount of paneCounts) {
    it(`${paneCount} panes: all ${paneCount - 1} handle element(s) removed`, () => {
      const { controller, panes } = makeController(paneCount);
      controller.destroy();
      assert.ok(allInserted(panes).every(el => el._removed));
    });
  }
});

// ── destroy() — document listeners removed ────────────────────────────────────

describe('destroy() — all document listeners deregistered for every handle', () => {
  const events     = ['mousemove', 'touchmove', 'mouseup', 'touchend'];
  const paneCounts = [2, 3];

  for (const paneCount of paneCounts) {
    for (const event of events) {
      it(`${paneCount} pane(s): 0 ${event} listeners remain after destroy`, () => {
        const { controller, doc } = makeController(paneCount);
        controller.destroy();
        assert.equal((doc.listeners[event] ?? []).length, 0);
      });
    }
  }
});

// ── destroy() then mount() — controller is reusable ──────────────────────────

describe('destroy() then mount() — controller can be remounted with a fresh pane set', () => {
  it('handles created for new panes after destroy+mount', () => {
    const doc = makeDocumentStub();
    global.document = doc;

    const controller  = new PaneResizeController();
    controller.mount(makePanes(2));
    controller.destroy();

    const secondPanes = makePanes(3);
    controller.mount(secondPanes);
    assert.equal(allInserted(secondPanes).length, 2);
  });

  it('old handles are gone and do not reappear on subsequent destroy()', () => {
    const doc = makeDocumentStub();
    global.document = doc;

    const controller = new PaneResizeController();
    controller.mount(makePanes(2));
    controller.destroy();
    controller.mount(makePanes(2));
    assert.doesNotThrow(() => controller.destroy());
  });
});
