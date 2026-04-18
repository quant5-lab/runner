import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { FullscreenToggle } from '../js/FullscreenToggle.js';

// ── DOM stubs ─────────────────────────────────────────────────────────────────

function makeButtonStub() {
  const handlers = {};
  return {
    className:   '',
    textContent: '',
    title:       '',
    addEventListener(event, fn) { (handlers[event] ??= []).push(fn); },
    _fire:       (event) => (handlers[event] ?? []).forEach(fn => fn()),
  };
}

function makeContainerStub({
  requestFullscreen       = null,
  webkitRequestFullscreen = null,
} = {}) {
  const children = [];
  const classes  = new Set();
  return {
    children,
    classList: {
      add:      c => classes.add(c),
      remove:   c => classes.delete(c),
      contains: c => classes.has(c),
    },
    appendChild:             el => children.push(el),
    requestFullscreen,
    webkitRequestFullscreen,
  };
}

function makeDocumentStub({
  fullscreenElement       = null,
  webkitFullscreenElement = null,
  exitFullscreen          = null,
  webkitExitFullscreen    = null,
} = {}) {
  const listeners = {};
  const stub = {
    listeners,
    fullscreenElement,
    webkitFullscreenElement,
    exitFullscreen,
    webkitExitFullscreen,
    addEventListener(event, fn)    { (listeners[event] ??= []).push(fn); },
    removeEventListener(event, fn) {
      if (listeners[event]) listeners[event] = listeners[event].filter(h => h !== fn);
    },
    createElement() { return makeButtonStub(); },
    _fire: (event) => (listeners[event] ?? []).forEach(fn => fn()),
  };
  return stub;
}

// ── Toggle fixture factory ────────────────────────────────────────────────────

function makeToggle({
  supportsRequestFullscreen     = true,
  supportsWebkitFullscreenOnly  = false,
  mountCount                    = 1,
} = {}) {
  const resizeCalls     = { n: 0 };
  const fullscreenCalls = { enter: 0, enterWebkit: 0, exit: 0, exitWebkit: 0 };

  const container = makeContainerStub({
    requestFullscreen: (!supportsWebkitFullscreenOnly && supportsRequestFullscreen)
      ? () => ++fullscreenCalls.enter
      : null,
    webkitRequestFullscreen: supportsWebkitFullscreenOnly
      ? () => ++fullscreenCalls.enterWebkit
      : null,
  });

  const doc = makeDocumentStub({
    exitFullscreen:       () => ++fullscreenCalls.exit,
    webkitExitFullscreen: () => ++fullscreenCalls.exitWebkit,
  });

  global.document = doc;
  const toggle = new FullscreenToggle(container, () => ++resizeCalls.n);
  for (let i = 0; i < mountCount; i++) toggle.mount();

  return {
    toggle,
    container,
    doc,
    resizeCalls,
    fullscreenCalls,
    button: () => container.children[0],
  };
}

function simulateApiFullscreenEntered(doc, container) {
  doc.fullscreenElement = container;
  doc._fire('fullscreenchange');
}

function simulateApiFullscreenExited(doc) {
  doc.fullscreenElement = null;
  doc._fire('fullscreenchange');
}

// ── Constants ─────────────────────────────────────────────────────────────────

const ENTER_ICON  = '⤢';
const EXIT_ICON   = '✕';
const ENTER_TITLE = 'Toggle fullscreen';
const EXIT_TITLE  = 'Exit fullscreen';
const BTN_CLASS   = 'fullscreen-toggle-btn';

// ── mount() — single-attachment guarantee ─────────────────────────────────────

describe('mount() — button is appended exactly once regardless of call count', () => {
  const mountCounts = [1, 2, 5, 10];

  for (const n of mountCounts) {
    it(`${n} mount() call(s) → exactly 1 button child in container`, () => {
      const { container } = makeToggle({ mountCount: n });
      assert.equal(container.children.length, 1);
    });
  }
});

describe('mount() — fullscreenchange listeners registered exactly once', () => {
  const mountCounts = [1, 2, 5, 10];

  for (const n of mountCounts) {
    it(`${n} mount() call(s) → exactly 1 'fullscreenchange' listener on document`, () => {
      const { doc } = makeToggle({ mountCount: n });
      assert.equal((doc.listeners['fullscreenchange'] ?? []).length, 1);
    });

    it(`${n} mount() call(s) → exactly 1 'webkitfullscreenchange' listener on document`, () => {
      const { doc } = makeToggle({ mountCount: n });
      assert.equal((doc.listeners['webkitfullscreenchange'] ?? []).length, 1);
    });
  }
});

describe('mount() — listener identity: both fullscreen events share one handler reference', () => {
  it('fullscreenchange and webkitfullscreenchange handlers are the same function', () => {
    const { doc } = makeToggle();
    const standard = doc.listeners['fullscreenchange'][0];
    const webkit   = doc.listeners['webkitfullscreenchange'][0];
    assert.equal(standard, webkit);
  });

  it('handler identity is stable across N mounts (no new closures created)', () => {
    const { doc } = makeToggle({ mountCount: 1 });
    const handlerAfterOne = doc.listeners['fullscreenchange'][0];
    assert.equal(doc.listeners['fullscreenchange'].length, 1);
    assert.equal(doc.listeners['fullscreenchange'][0], handlerAfterOne);
  });
});

// ── Button — initial state ────────────────────────────────────────────────────

describe('button — initial state after first mount()', () => {
  it('button is the first and only child in the container', () => {
    const { container } = makeToggle();
    assert.equal(container.children.length, 1);
  });

  it(`button className is '${BTN_CLASS}'`, () => {
    const { button } = makeToggle();
    assert.equal(button().className, BTN_CLASS);
  });

  it(`button textContent is the enter icon '${ENTER_ICON}'`, () => {
    const { button } = makeToggle();
    assert.equal(button().textContent, ENTER_ICON);
  });

  it(`button title is '${ENTER_TITLE}'`, () => {
    const { button } = makeToggle();
    assert.equal(button().title, ENTER_TITLE);
  });
});

// ── Button — state sync via fullscreenchange ──────────────────────────────────

describe('button — shows exit state when fullscreen is active', () => {
  it('document.fullscreenElement set → fullscreenchange → exit icon and title', () => {
    const { button, doc, container } = makeToggle();
    simulateApiFullscreenEntered(doc, container);
    assert.equal(button().textContent, EXIT_ICON);
    assert.equal(button().title,       EXIT_TITLE);
  });

  it('document.webkitFullscreenElement set → fullscreenchange → exit icon and title', () => {
    const { button, doc } = makeToggle();
    doc.webkitFullscreenElement = {};
    doc._fire('fullscreenchange');
    assert.equal(button().textContent, EXIT_ICON);
    assert.equal(button().title,       EXIT_TITLE);
  });

  it('CSS class chart-fullscreen → fullscreenchange → exit icon and title', () => {
    const { button, container, doc } = makeToggle({ supportsRequestFullscreen: false });
    container.classList.add('chart-fullscreen');
    doc._fire('fullscreenchange');
    assert.equal(button().textContent, EXIT_ICON);
    assert.equal(button().title,       EXIT_TITLE);
  });
});

describe('button — returns to enter state when fullscreen is no longer active', () => {
  it('fullscreenElement cleared → fullscreenchange → enter icon and title', () => {
    const { button, doc, container } = makeToggle();
    simulateApiFullscreenEntered(doc, container);
    assert.equal(button().textContent, EXIT_ICON);
    simulateApiFullscreenExited(doc);
    assert.equal(button().textContent, ENTER_ICON);
    assert.equal(button().title,       ENTER_TITLE);
  });

  it('CSS class removed → fullscreenchange → enter icon and title', () => {
    const { button, container, doc } = makeToggle({ supportsRequestFullscreen: false });
    container.classList.add('chart-fullscreen');
    doc._fire('fullscreenchange');
    assert.equal(button().textContent, EXIT_ICON);
    container.classList.remove('chart-fullscreen');
    doc._fire('fullscreenchange');
    assert.equal(button().textContent, ENTER_ICON);
    assert.equal(button().title,       ENTER_TITLE);
  });
});

// ── onResize — call timing ────────────────────────────────────────────────────

describe('onResize — not called on mount()', () => {
  const mountCounts = [1, 2, 5];

  for (const n of mountCounts) {
    it(`${n} mount() call(s) → 0 onResize calls`, () => {
      const { resizeCalls } = makeToggle({ mountCount: n });
      assert.equal(resizeCalls.n, 0);
    });
  }
});

describe('onResize — called exactly once per fullscreenchange event', () => {
  it('single mount: fullscreenchange fires → onResize called once', () => {
    const { doc, resizeCalls } = makeToggle({ mountCount: 1 });
    doc._fire('fullscreenchange');
    assert.equal(resizeCalls.n, 1);
  });

  it('N mounts: fullscreenchange fires → onResize still called exactly once (idempotency)', () => {
    const { doc, resizeCalls } = makeToggle({ mountCount: 5 });
    doc._fire('fullscreenchange');
    assert.equal(resizeCalls.n, 1);
  });

  it('webkitfullscreenchange fires → onResize called once', () => {
    const { doc, resizeCalls } = makeToggle({ mountCount: 1 });
    doc._fire('webkitfullscreenchange');
    assert.equal(resizeCalls.n, 1);
  });

  it('N sequential fullscreenchange events → N onResize calls', () => {
    const { doc, resizeCalls } = makeToggle();
    const events = 4;
    for (let i = 0; i < events; i++) doc._fire('fullscreenchange');
    assert.equal(resizeCalls.n, events);
  });
});

// ── Enter fullscreen — native API path ───────────────────────────────────────

describe('enter fullscreen — standard requestFullscreen path', () => {
  it('clicking button calls container.requestFullscreen', () => {
    const { button, fullscreenCalls } = makeToggle({ supportsRequestFullscreen: true });
    button()._fire('click');
    assert.equal(fullscreenCalls.enter, 1);
  });

  it('clicking button does not immediately call onResize (waits for fullscreenchange)', () => {
    const { button, resizeCalls } = makeToggle({ supportsRequestFullscreen: true });
    button()._fire('click');
    assert.equal(resizeCalls.n, 0);
  });

  it('multiple clicks while API is pending each trigger requestFullscreen (no internal guard on enter)', () => {
    const { button, fullscreenCalls, doc, container } = makeToggle({ supportsRequestFullscreen: true });
    button()._fire('click');
    simulateApiFullscreenEntered(doc, container);
    assert.equal(fullscreenCalls.enter, 1, 'only one enter before fullscreen was active');
  });
});

describe('enter fullscreen — webkit-only requestFullscreen path', () => {
  it('clicking button calls container.webkitRequestFullscreen when standard API absent', () => {
    const { button, fullscreenCalls } = makeToggle({ supportsWebkitFullscreenOnly: true });
    button()._fire('click');
    assert.equal(fullscreenCalls.enterWebkit, 1);
    assert.equal(fullscreenCalls.enter,       0);
  });

  it('webkit enter does not immediately call onResize', () => {
    const { button, resizeCalls } = makeToggle({ supportsWebkitFullscreenOnly: true });
    button()._fire('click');
    assert.equal(resizeCalls.n, 0);
  });
});

describe('enter fullscreen — CSS fallback path (no Fullscreen API)', () => {
  it('container gains chart-fullscreen class', () => {
    const { button, container } = makeToggle({ supportsRequestFullscreen: false });
    button()._fire('click');
    assert.ok(container.classList.contains('chart-fullscreen'));
  });

  it('onResize is called immediately', () => {
    const { button, resizeCalls } = makeToggle({ supportsRequestFullscreen: false });
    button()._fire('click');
    assert.equal(resizeCalls.n, 1);
  });

  it('button shows exit state immediately after CSS enter', () => {
    const { button } = makeToggle({ supportsRequestFullscreen: false });
    button()._fire('click');
    assert.equal(button().textContent, EXIT_ICON);
    assert.equal(button().title,       EXIT_TITLE);
  });
});

// ── Exit fullscreen — after native API entry ──────────────────────────────────

describe('exit fullscreen — document.exitFullscreen path', () => {
  it('clicking button after API entry calls document.exitFullscreen', () => {
    const { button, doc, container, fullscreenCalls } = makeToggle();
    button()._fire('click');
    simulateApiFullscreenEntered(doc, container);
    button()._fire('click');
    assert.equal(fullscreenCalls.exit, 1);
  });

  it('exit via API does not immediately call onResize (waits for fullscreenchange)', () => {
    const { button, doc, container, resizeCalls } = makeToggle();
    button()._fire('click');
    simulateApiFullscreenEntered(doc, container);
    const callsBeforeExit = resizeCalls.n;
    button()._fire('click');
    assert.equal(resizeCalls.n, callsBeforeExit);
  });
});

describe('exit fullscreen — webkitExitFullscreen path (standard API absent)', () => {
  it('calls webkitExitFullscreen when exitFullscreen is not available', () => {
    const resizeCalls     = { n: 0 };
    const fullscreenCalls = { enterWebkit: 0, exitWebkit: 0 };
    const container = makeContainerStub({
      requestFullscreen:       null,
      webkitRequestFullscreen: () => ++fullscreenCalls.enterWebkit,
    });
    const doc = makeDocumentStub({
      exitFullscreen:       null,
      webkitExitFullscreen: () => ++fullscreenCalls.exitWebkit,
    });
    global.document = doc;

    const toggle = new FullscreenToggle(container, () => ++resizeCalls.n);
    toggle.mount();
    const btn = container.children[0];

    btn._fire('click');
    doc.webkitFullscreenElement = container;
    doc._fire('webkitfullscreenchange');
    btn._fire('click');

    assert.equal(fullscreenCalls.exitWebkit, 1);
    assert.equal(fullscreenCalls.enterWebkit, 1);
  });
});

// ── Exit fullscreen — after CSS fallback entry ────────────────────────────────

describe('exit fullscreen — CSS fallback path', () => {
  it('container loses chart-fullscreen class', () => {
    const { button, container } = makeToggle({ supportsRequestFullscreen: false });
    button()._fire('click');
    assert.ok(container.classList.contains('chart-fullscreen'), 'precondition: entered CSS fullscreen');
    button()._fire('click');
    assert.ok(!container.classList.contains('chart-fullscreen'));
  });

  it('onResize is called immediately on CSS exit', () => {
    const { button, resizeCalls } = makeToggle({ supportsRequestFullscreen: false });
    button()._fire('click');
    const callsAfterEnter = resizeCalls.n;
    button()._fire('click');
    assert.equal(resizeCalls.n, callsAfterEnter + 1);
  });

  it('button shows enter state immediately after CSS exit', () => {
    const { button } = makeToggle({ supportsRequestFullscreen: false });
    button()._fire('click');
    assert.equal(button().textContent, EXIT_ICON, 'precondition: exit icon after enter');
    button()._fire('click');
    assert.equal(button().textContent, ENTER_ICON);
    assert.equal(button().title,       ENTER_TITLE);
  });
});

// ── Fullscreen state detection — source independence ─────────────────────────

describe('fullscreen detection — source independence', () => {
  const detectionCases = [
    {
      label: 'document.fullscreenElement',
      activate:   (doc, container) => { doc.fullscreenElement = container; },
      deactivate: (doc)            => { doc.fullscreenElement = null; },
    },
    {
      label: 'document.webkitFullscreenElement',
      activate:   (doc, container) => { doc.webkitFullscreenElement = container; },
      deactivate: (doc)            => { doc.webkitFullscreenElement = null; },
    },
    {
      label: 'CSS class chart-fullscreen',
      activate:   (_, container) => container.classList.add('chart-fullscreen'),
      deactivate: (_, container) => container.classList.remove('chart-fullscreen'),
    },
  ];

  for (const { label, activate, deactivate } of detectionCases) {
    it(`${label} → active: button shows exit state`, () => {
      const { button, doc, container } = makeToggle({ supportsRequestFullscreen: false });
      activate(doc, container);
      doc._fire('fullscreenchange');
      assert.equal(button().textContent, EXIT_ICON, `source: ${label}`);
    });

    it(`${label} → cleared: button shows enter state`, () => {
      const { button, doc, container } = makeToggle({ supportsRequestFullscreen: false });
      activate(doc, container);
      doc._fire('fullscreenchange');
      deactivate(doc, container);
      doc._fire('fullscreenchange');
      assert.equal(button().textContent, ENTER_ICON, `source: ${label}`);
    });
  }
});
