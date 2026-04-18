import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { PaneResizeHandle } from '../js/PaneResizeHandle.js';
import { MIN_PANE_HEIGHT  } from '../js/PaneResizeCalculator.js';

// ── DOM stubs ─────────────────────────────────────────────────────────────────

function makeElementStub() {
  const listeners = {};
  return {
    className:        '',
    style:            {},
    _removed:         false,
    remove:           function () { this._removed = true; },
    addEventListener: (event, fn, _opts) => { (listeners[event] ??= []).push(fn); },
    _fire:            (event, arg) => [...(listeners[event] ?? [])].forEach(fn => fn(arg)),
    _listeners:       listeners,
  };
}

function makeDocumentStub() {
  const listeners = {};
  const created   = [];
  return {
    listeners,
    created,
    addEventListener:    (event, fn, _opts) => { (listeners[event] ??= []).push(fn); },
    removeEventListener: (event, fn) => {
      if (listeners[event]) listeners[event] = listeners[event].filter(l => l !== fn);
    },
    createElement: () => { const el = makeElementStub(); created.push(el); return el; },
    _fire:         (event, arg) => [...(listeners[event] ?? [])].forEach(fn => fn(arg)),
  };
}

// ── Pane stub ─────────────────────────────────────────────────────────────────

function makePane(clientWidth, clientHeight) {
  const resizeCalls = [];
  return {
    container: { clientWidth, clientHeight, style: {} },
    chart:     { resize: (w, h) => resizeCalls.push({ w, h }), _calls: resizeCalls },
  };
}

// ── Event arg factories ───────────────────────────────────────────────────────

function mouseEvent(clientY) { return { clientY }; }
function touchEvent(clientY) { return { touches: [{ clientY }], preventDefault: () => {} }; }

// ── Fixture factory ───────────────────────────────────────────────────────────

function makeHandle(aboveWidth = 1000, aboveHeight = 400, belowWidth = 1000, belowHeight = 200) {
  const doc = makeDocumentStub();
  global.document = doc;
  const above  = makePane(aboveWidth, aboveHeight);
  const below  = makePane(belowWidth, belowHeight);
  const handle = new PaneResizeHandle(above, below);
  const el     = doc.created[0];
  return { handle, above, below, doc, el };
}

// ── Drag protocol descriptors ─────────────────────────────────────────────────

const protocols = [
  {
    label:      'mouse',
    startEvent: 'mousedown',
    moveEvent:  'mousemove',
    endEvent:   'mouseup',
    startArg:   (y) => mouseEvent(y),
    moveArg:    (y) => mouseEvent(y),
  },
  {
    label:      'touch',
    startEvent: 'touchstart',
    moveEvent:  'touchmove',
    endEvent:   'touchend',
    startArg:   (y) => touchEvent(y),
    moveArg:    (y) => touchEvent(y),
  },
];

// ── Element — initial state ───────────────────────────────────────────────────

describe('element — className assigned at construction', () => {
  it('class is pane-resize-handle', () => {
    const { el } = makeHandle();
    assert.equal(el.className, 'pane-resize-handle');
  });
});

// ── Document listeners — registered exactly once at construction ──────────────

describe('document listeners — attached once per event type at construction', () => {
  const events = ['mousemove', 'touchmove', 'mouseup', 'touchend'];

  for (const event of events) {
    it(`exactly one ${event} listener registered`, () => {
      const { doc } = makeHandle();
      assert.equal(doc.listeners[event].length, 1);
    });
  }
});

// ── Drag — resize args delivered to both panes ────────────────────────────────

describe('drag — chart.resize receives correct height and width for both panes', () => {
  const dragCases = [
    { startY: 100, endY: 150, expectedAbove: 450, expectedBelow: 150, label: 'drag down: above grows, below shrinks' },
    { startY: 200, endY: 150, expectedAbove: 350, expectedBelow: 250, label: 'drag up: above shrinks, below grows'   },
  ];

  for (const { startY, endY, expectedAbove, expectedBelow, label } of dragCases) {
    for (const { label: proto, startEvent, moveEvent, startArg, moveArg } of protocols) {
      it(`${proto} — ${label}`, () => {
        const { above, below, doc, el } = makeHandle();
        el._fire(startEvent,  startArg(startY));
        doc._fire(moveEvent,  moveArg(endY));
        assert.deepEqual(above.chart._calls, [{ w: 1000, h: expectedAbove }]);
        assert.deepEqual(below.chart._calls, [{ w: 1000, h: expectedBelow }]);
      });
    }
  }
});

// ── Width independence ────────────────────────────────────────────────────────

describe('drag — resize width comes from each pane container independently', () => {
  it('above gets its own clientWidth; below gets its own clientWidth', () => {
    const { above, below, doc, el } = makeHandle(800, 400, 1200, 200);
    el._fire('mousedown', mouseEvent(0));
    doc._fire('mousemove', mouseEvent(10));
    assert.equal(above.chart._calls[0].w, 800);
    assert.equal(below.chart._calls[0].w, 1200);
  });
});

// ── Style.height ──────────────────────────────────────────────────────────────

describe('drag — container.style.height updated in sync with chart.resize height', () => {
  it('above and below containers receive the new height string', () => {
    const { above, below, doc, el } = makeHandle();
    el._fire('mousedown', mouseEvent(100));
    doc._fire('mousemove', mouseEvent(130));
    assert.equal(above.container.style.height, '430px');
    assert.equal(below.container.style.height, '170px');
  });
});

// ── Multiple moves — delta always from startY ─────────────────────────────────

describe('drag — each mousemove computes delta from the startY of the current drag', () => {
  it('second move at y=120 produces delta=20 regardless of first move', () => {
    const { above, doc, el } = makeHandle();
    el._fire('mousedown', mouseEvent(100));
    doc._fire('mousemove', mouseEvent(110));
    doc._fire('mousemove', mouseEvent(120));
    assert.equal(above.chart._calls.length, 2);
    assert.equal(above.chart._calls[1].h, 420);
  });
});

// ── Drag end — subsequent moves ignored ──────────────────────────────────────

describe('drag end — subsequent move events do not call resize', () => {
  for (const { label, startEvent, moveEvent, endEvent, startArg, moveArg } of protocols) {
    it(`${label}: move after ${endEvent} does not call resize`, () => {
      const { above, below, doc, el } = makeHandle();
      el._fire(startEvent, startArg(100));
      doc._fire(endEvent);
      doc._fire(moveEvent, moveArg(200));
      assert.equal(above.chart._calls.length, 0);
      assert.equal(below.chart._calls.length, 0);
    });

    it(`${label}: new drag starts correctly after ${endEvent}`, () => {
      const { above, doc, el } = makeHandle();
      el._fire(startEvent, startArg(100));
      doc._fire(moveEvent,  moveArg(110));
      doc._fire(endEvent);
      el._fire(startEvent,  startArg(0));
      doc._fire(moveEvent,  moveArg(20));
      assert.equal(above.chart._calls.length, 2);
      assert.equal(above.chart._calls[1].h, 420);
    });
  }
});

// ── No active drag — move events ignored ─────────────────────────────────────

describe('no active drag — move events before drag start do not call resize', () => {
  for (const { label, moveEvent, moveArg } of protocols) {
    it(`${label}: ${moveEvent} before drag start does not call resize`, () => {
      const { above, below, doc } = makeHandle();
      doc._fire(moveEvent, moveArg(500));
      assert.equal(above.chart._calls.length, 0);
      assert.equal(below.chart._calls.length, 0);
    });
  }
});

// ── Clamping ──────────────────────────────────────────────────────────────────

describe('clamping — neither pane shrinks below MIN_PANE_HEIGHT during drag', () => {
  const clampCases = [
    {
      delta:         9999,
      expectedAbove: 400 + 200 - MIN_PANE_HEIGHT,
      expectedBelow: MIN_PANE_HEIGHT,
      label:         'drag past below headroom',
    },
    {
      delta:         -9999,
      expectedAbove: MIN_PANE_HEIGHT,
      expectedBelow: 400 + 200 - MIN_PANE_HEIGHT,
      label:         'drag past above headroom',
    },
  ];

  for (const { delta, expectedAbove, expectedBelow, label } of clampCases) {
    it(`${label}: both panes land at boundary values`, () => {
      const { above, below, doc, el } = makeHandle();
      el._fire('mousedown', mouseEvent(0));
      doc._fire('mousemove', mouseEvent(delta));
      assert.deepEqual(above.chart._calls, [{ w: 1000, h: expectedAbove }]);
      assert.deepEqual(below.chart._calls, [{ w: 1000, h: expectedBelow }]);
    });
  }
});

// ── Destroy — element removed ─────────────────────────────────────────────────

describe('destroy() — element removed from DOM', () => {
  it('element.remove() is called', () => {
    const { handle, el } = makeHandle();
    handle.destroy();
    assert.ok(el._removed);
  });
});

// ── Destroy — document listeners deregistered ─────────────────────────────────

describe('destroy() — all document listeners deregistered', () => {
  const events = ['mousemove', 'touchmove', 'mouseup', 'touchend'];

  for (const event of events) {
    it(`${event} listener count drops to 0`, () => {
      const { handle, doc } = makeHandle();
      handle.destroy();
      assert.equal((doc.listeners[event] ?? []).length, 0);
    });
  }
});

describe('destroy() — move events after destroy do not call resize', () => {
  for (const { label, moveEvent, moveArg } of protocols) {
    it(`${label}: ${moveEvent} after destroy is a no-op`, () => {
      const { handle, above, below, doc } = makeHandle();
      handle.destroy();
      doc._fire(moveEvent, moveArg(9999));
      assert.equal(above.chart._calls.length, 0);
      assert.equal(below.chart._calls.length, 0);
    });
  }
});
