import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { TradeSortToggle } from '../js/TradeSortToggle.js';
import { SortDirection }   from '../js/SortDirection.js';

// ── DOM stubs ─────────────────────────────────────────────────────────────────

function makeButtonStub() {
  const handlers = {};
  const attrs    = {};
  return {
    className:   '',
    textContent: '',
    style:       { display: '' },
    addEventListener(event, fn) { (handlers[event] ??= []).push(fn); },
    setAttribute(name, value)   { attrs[name] = value; },
    getAttribute(name)          { return attrs[name] ?? null; },
    _fire: (event) => (handlers[event] ?? []).forEach(fn => fn()),
  };
}

function makeHeaderStub() {
  const children = [];
  return {
    children,
    appendChild: el => children.push(el),
  };
}

// ── Toggle fixture factory ────────────────────────────────────────────────────

function makeToggle({
  initialDirection = SortDirection.DESC,
  hasAnyTrades     = true,
} = {}) {
  const changes = [];
  const header  = makeHeaderStub();

  global.document = { createElement: () => makeButtonStub() };

  const toggle = new TradeSortToggle(dir => changes.push(dir), initialDirection);
  toggle.mount(header, hasAnyTrades);

  return {
    toggle,
    header,
    changes,
    button: () => header.children[0],
  };
}

// ── Direction semantics ───────────────────────────────────────────────────────
//
// DESC is the default/off direction (aria-pressed=false, ▼, "Newest first").
// ASC  is the non-default/active direction (aria-pressed=true, ▲, "Oldest first").
// This table drives both label and aria-pressed assertions in one place so that
// any future direction addition is covered by adding a single row.

const directionCases = [
  {
    direction:    SortDirection.DESC,
    arrow:        '▼',
    label:        'Newest first',
    ariaPressedInitial: 'false',
    ariaPressedAfterClick: 'true',
  },
  {
    direction:    SortDirection.ASC,
    arrow:        '▲',
    label:        'Oldest first',
    ariaPressedInitial: 'true',
    ariaPressedAfterClick: 'false',
  },
];

// ── mount() — button lifecycle ────────────────────────────────────────────────

describe('mount() — button is created and appended on first call', () => {
  it('header has exactly one button child after mount', () => {
    const { header } = makeToggle();
    assert.equal(header.children.length, 1);
  });

  it('button has class sort-toggle-btn', () => {
    const { button } = makeToggle();
    assert.equal(button().className, 'sort-toggle-btn');
  });

  it('second mount() call does not append another button', () => {
    const { toggle, header } = makeToggle();
    toggle.mount(header, true);
    assert.equal(header.children.length, 1);
  });
});

describe('mount() — visibility follows hasAnyTrades on every call', () => {
  it('hasAnyTrades=true → display is empty string (visible)', () => {
    const { button } = makeToggle({ hasAnyTrades: true });
    assert.equal(button().style.display, '');
  });

  it('hasAnyTrades=false → display is "none"', () => {
    const { button } = makeToggle({ hasAnyTrades: false });
    assert.equal(button().style.display, 'none');
  });

  it('re-mount with hasAnyTrades=false hides a previously visible button', () => {
    const { toggle, header, button } = makeToggle({ hasAnyTrades: true });
    assert.equal(button().style.display, '');
    toggle.mount(header, false);
    assert.equal(button().style.display, 'none');
  });

  it('re-mount with hasAnyTrades=true shows a previously hidden button', () => {
    const { toggle, header, button } = makeToggle({ hasAnyTrades: false });
    assert.equal(button().style.display, 'none');
    toggle.mount(header, true);
    assert.equal(button().style.display, '');
  });
});

// ── direction getter ──────────────────────────────────────────────────────────

describe('direction getter — reflects current direction', () => {
  for (const { direction } of directionCases) {
    it(`initialDirection=${direction} → getter returns ${direction}`, () => {
      const { toggle } = makeToggle({ initialDirection: direction });
      assert.equal(toggle.direction, direction);
    });
  }

  it('getter tracks state through multiple clicks', () => {
    const { toggle, button } = makeToggle({ initialDirection: SortDirection.DESC });
    button()._fire('click');
    assert.equal(toggle.direction, SortDirection.ASC);
    button()._fire('click');
    assert.equal(toggle.direction, SortDirection.DESC);
    button()._fire('click');
    assert.equal(toggle.direction, SortDirection.ASC);
  });
});

// ── onChange callback ─────────────────────────────────────────────────────────

describe('onChange — not called on mount()', () => {
  it('onChange receives no calls when toggle is created and mounted', () => {
    const { changes } = makeToggle();
    assert.deepEqual(changes, []);
  });
});

describe('onChange — called with the new direction on each click', () => {
  it('first click from DESC fires onChange with ASC', () => {
    const { button, changes } = makeToggle({ initialDirection: SortDirection.DESC });
    button()._fire('click');
    assert.deepEqual(changes, [SortDirection.ASC]);
  });

  it('first click from ASC fires onChange with DESC', () => {
    const { button, changes } = makeToggle({ initialDirection: SortDirection.ASC });
    button()._fire('click');
    assert.deepEqual(changes, [SortDirection.DESC]);
  });

  it('alternating clicks produce alternating directions in order', () => {
    const { button, changes } = makeToggle({ initialDirection: SortDirection.DESC });
    button()._fire('click');
    button()._fire('click');
    button()._fire('click');
    assert.deepEqual(changes, [SortDirection.ASC, SortDirection.DESC, SortDirection.ASC]);
  });
});

// ── label — arrow + text reflect direction ────────────────────────────────────

describe('label — initial text and arrow match initialDirection', () => {
  for (const { direction, arrow, label } of directionCases) {
    it(`initialDirection=${direction} → shows '${arrow} ${label}'`, () => {
      const { button } = makeToggle({ initialDirection: direction });
      assert.ok(button().textContent.includes(arrow),
        `arrow ${arrow} missing: ${button().textContent}`);
      assert.ok(button().textContent.includes(label),
        `label '${label}' missing: ${button().textContent}`);
    });
  }
});

describe('label — updates arrow and text after click', () => {
  for (const { direction, ariaPressedAfterClick } of directionCases) {
    const opposite = directionCases.find(d => d.direction !== direction);
    it(`initialDirection=${direction} → click → shows '${opposite.arrow} ${opposite.label}'`, () => {
      const { button } = makeToggle({ initialDirection: direction });
      button()._fire('click');
      assert.ok(button().textContent.includes(opposite.arrow),
        `arrow ${opposite.arrow} missing after click: ${button().textContent}`);
      assert.ok(button().textContent.includes(opposite.label),
        `label '${opposite.label}' missing after click: ${button().textContent}`);
    });
  }
});

// ── aria-pressed — assistive technology announcement ─────────────────────────
//
// aria-pressed=false is the default/off (DESC) state — matching initial-load
// where neither toggle reports an active/pressed state.
// aria-pressed=true  is the non-default/active (ASC) state.

describe('aria-pressed — initial state matches direction semantics', () => {
  for (const { direction, ariaPressedInitial } of directionCases) {
    it(`initialDirection=${direction} → aria-pressed="${ariaPressedInitial}"`, () => {
      const { button } = makeToggle({ initialDirection: direction });
      assert.equal(button().getAttribute('aria-pressed'), ariaPressedInitial);
    });
  }
});

describe('aria-pressed — updates after click', () => {
  for (const { direction, ariaPressedAfterClick } of directionCases) {
    it(`initialDirection=${direction} → click → aria-pressed="${ariaPressedAfterClick}"`, () => {
      const { button } = makeToggle({ initialDirection: direction });
      button()._fire('click');
      assert.equal(button().getAttribute('aria-pressed'), ariaPressedAfterClick);
    });
  }

  it('returns to aria-pressed="false" when toggled back to default (DESC)', () => {
    const { button } = makeToggle({ initialDirection: SortDirection.DESC });
    button()._fire('click');
    assert.equal(button().getAttribute('aria-pressed'), 'true');
    button()._fire('click');
    assert.equal(button().getAttribute('aria-pressed'), 'false');
  });

  it('aria-pressed tracks direction through multiple alternating clicks', () => {
    const { button } = makeToggle({ initialDirection: SortDirection.DESC });
    const expected = ['true', 'false', 'true', 'false'];
    for (const want of expected) {
      button()._fire('click');
      assert.equal(button().getAttribute('aria-pressed'), want,
        `expected aria-pressed="${want}" after click`);
    }
  });
});
