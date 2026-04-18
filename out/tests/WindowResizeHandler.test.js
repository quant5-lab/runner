import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { WindowResizeHandler } from '../js/WindowResizeHandler.js';

// ── Stubs ─────────────────────────────────────────────────────────────────────

function makeWindowStub() {
  const listeners = {};
  return {
    listeners,
    addEventListener(event, fn)    { (listeners[event] ??= []).push(fn); },
    removeEventListener(event, fn) {
      if (listeners[event]) listeners[event] = listeners[event].filter(h => h !== fn);
    },
    _fire(event) { (listeners[event] ?? []).slice().forEach(fn => fn()); },
  };
}

function makeFixture() {
  const win   = makeWindowStub();
  const calls = { n: 0 };
  global.window = win;
  const handler = new WindowResizeHandler(() => ++calls.n);
  return { handler, calls, win };
}

// ── Constructor — immediate listener registration ─────────────────────────────
//
// The handler must register on window at construction time — no lazy init,
// no mount() step.  Exactly one listener must be present; no extras.

describe('constructor — registers exactly one resize listener at construction time', () => {
  it('zero listeners present on a fresh window before construction', () => {
    const win = makeWindowStub();
    assert.equal((win.listeners['resize'] ?? []).length, 0);
  });

  it('exactly one resize listener present immediately after construction', () => {
    const { win } = makeFixture();
    assert.equal(win.listeners['resize'].length, 1);
  });

  it('construction does not register listeners for any other event', () => {
    const { win } = makeFixture();
    const nonResizeEvents = Object.keys(win.listeners).filter(e => e !== 'resize');
    assert.equal(nonResizeEvents.length, 0);
  });
});

// ── onResize — call semantics ─────────────────────────────────────────────────
//
// Construction alone must not invoke onResize.  Every subsequent resize
// event must fire it exactly once — no skipping, no batching.

describe('onResize — not called before the first resize event', () => {
  it('construction alone: 0 onResize calls', () => {
    const { calls } = makeFixture();
    assert.equal(calls.n, 0);
  });
});

describe('onResize — called exactly once per resize event', () => {
  const eventCounts = [1, 2, 5, 10];

  for (const n of eventCounts) {
    it(`${n} resize event(s) → onResize called exactly ${n} time(s)`, () => {
      const { win, calls } = makeFixture();
      for (let i = 0; i < n; i++) win._fire('resize');
      assert.equal(calls.n, n);
    });
  }
});

// ── destroy() — listener removal ─────────────────────────────────────────────
//
// destroy() must remove the registered listener so that subsequent resize
// events produce no more onResize calls.  Calls accumulated before destroy()
// are unaffected.

describe('destroy() — removes the window resize listener', () => {
  it('zero resize listeners remain on window after destroy()', () => {
    const { handler, win } = makeFixture();
    handler.destroy();
    assert.equal((win.listeners['resize'] ?? []).length, 0);
  });

  it('onResize is not invoked for resize events fired after destroy()', () => {
    const { handler, win, calls } = makeFixture();
    handler.destroy();
    win._fire('resize');
    assert.equal(calls.n, 0);
  });

  it('onResize calls accumulated before destroy() are preserved', () => {
    const { handler, win, calls } = makeFixture();
    win._fire('resize');
    win._fire('resize');
    handler.destroy();
    win._fire('resize');
    assert.equal(calls.n, 2);
  });
});

describe('destroy() — idempotent: safe to call multiple times', () => {
  it('calling destroy() twice does not throw', () => {
    const { handler } = makeFixture();
    handler.destroy();
    assert.doesNotThrow(() => handler.destroy());
  });

  it('no onResize calls after double destroy()', () => {
    const { handler, win, calls } = makeFixture();
    handler.destroy();
    handler.destroy();
    win._fire('resize');
    assert.equal(calls.n, 0);
  });
});

// ── destroy() — handler identity ─────────────────────────────────────────────
//
// destroy() must surgically remove only its own handler.  Any unrelated
// resize listeners on window must continue to fire.

describe('destroy() — removes only its own listener; unrelated listeners survive', () => {
  it('a bystander resize listener fires after the handler is destroyed', () => {
    const { handler, win } = makeFixture();
    let bystander = 0;
    win.addEventListener('resize', () => ++bystander);
    handler.destroy();
    win._fire('resize');
    assert.equal(bystander, 1);
  });

  it('bystander listener count is unchanged by destroy()', () => {
    const { handler, win } = makeFixture();
    win.addEventListener('resize', () => {});
    handler.destroy();
    assert.equal((win.listeners['resize'] ?? []).length, 1);
  });
});

// ── Multiple instances — independence ─────────────────────────────────────────
//
// Each WindowResizeHandler instance owns its own listener slot.  Constructing
// N instances adds N listeners.  Destroying one must not affect the others.

describe('multiple instances — each registers its own listener independently', () => {
  const instanceCounts = [2, 3, 5];

  for (const n of instanceCounts) {
    it(`${n} instances → exactly ${n} resize listeners on window`, () => {
      const win = makeWindowStub();
      global.window = win;
      for (let i = 0; i < n; i++) new WindowResizeHandler(() => {});
      assert.equal(win.listeners['resize'].length, n);
    });
  }
});

describe('multiple instances — each onResize fires independently on resize', () => {
  it('two instances: both onResize callbacks receive every resize event', () => {
    const win = makeWindowStub();
    global.window = win;
    const a = { n: 0 }, b = { n: 0 };
    new WindowResizeHandler(() => ++a.n);
    new WindowResizeHandler(() => ++b.n);
    win._fire('resize');
    win._fire('resize');
    assert.equal(a.n, 2);
    assert.equal(b.n, 2);
  });
});

describe('multiple instances — destroying one does not affect the others', () => {
  it('destroy() on instance A: instance B still receives resize events', () => {
    const win = makeWindowStub();
    global.window = win;
    const a = { n: 0 }, b = { n: 0 };
    const handlerA = new WindowResizeHandler(() => ++a.n);
    new WindowResizeHandler(() => ++b.n);
    handlerA.destroy();
    win._fire('resize');
    assert.equal(a.n, 0);
    assert.equal(b.n, 1);
  });

  it('destroy() on instance A: its own listener is removed, others remain', () => {
    const win = makeWindowStub();
    global.window = win;
    const handlerA = new WindowResizeHandler(() => {});
    new WindowResizeHandler(() => {});
    new WindowResizeHandler(() => {});
    handlerA.destroy();
    assert.equal(win.listeners['resize'].length, 2);
  });
});
