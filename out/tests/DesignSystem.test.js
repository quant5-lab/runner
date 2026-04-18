import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

// ── File readers ──────────────────────────────────────────────────────────────
//
// Files are read once at module scope; all test bodies reference the cached
// strings.  A missing file throws at load time, which fails the suite clearly.

const tokensCSS = readFileSync(new URL('../css/tokens.css', import.meta.url), 'utf8');
const stylesCSS = readFileSync(new URL('../css/styles.css', import.meta.url), 'utf8');
const indexHTML = readFileSync(new URL('../index.html',     import.meta.url), 'utf8');

// ── CSS parsers ───────────────────────────────────────────────────────────────
//
// Pure functions over CSS text.  Each extracts one kind of information so
// tests can compose them without re-implementing the same regex in every it().

function stripComments(css) {
  return css.replace(/\/\*[\s\S]*?\*\//g, '');
}

function declaredTokensIn(css) {
  return new Set(stripComments(css).match(/--[\w-]+(?=\s*:)/g) ?? []);
}

function referencedTokensIn(css) {
  return new Set(
    [...stripComments(css).matchAll(/var\(\s*(--[\w-]+)/g)].map(m => m[1]),
  );
}

function hexColorsIn(css) {
  return stripComments(css).match(/#[0-9a-fA-F]{3,8}\b/g) ?? [];
}

function rgbLiteralsIn(css) {
  return stripComments(css).match(/rgba?\s*\(/g) ?? [];
}

function importPathsIn(css) {
  return [...css.matchAll(/@import\s+['"]([^'"]+)['"]/g)].map(m => m[1]);
}

const declaredTokens   = declaredTokensIn(tokensCSS);
const referencedTokens = referencedTokensIn(stylesCSS);

// ── tokens.css — file structure ───────────────────────────────────────────────
//
// Tokens file must be non-empty and contain the :root declaration block that
// makes the custom properties available to the entire document.

describe('tokens.css — file is non-empty', () => {
  it('file has content', () => {
    assert.ok(tokensCSS.trim().length > 0);
  });
});

describe('tokens.css — :root block', () => {
  it(':root block is present', () => {
    assert.ok(stripComments(tokensCSS).includes(':root'));
  });

  it(':root block has an opening brace', () => {
    assert.match(stripComments(tokensCSS), /:root\s*\{/);
  });
});

describe('tokens.css — token naming convention', () => {
  it('at least one custom property is declared', () => {
    assert.ok(declaredTokens.size > 0);
  });

  it('every declared name starts with -- (CSS custom property)', () => {
    for (const token of declaredTokens) {
      assert.ok(token.startsWith('--'), `'${token}' does not start with '--'`);
    }
  });
});

// ── styles.css — file structure ───────────────────────────────────────────────
//
// styles.css must be non-empty, must explicitly @import tokens.css so the
// dependency is visible in code, and must not re-declare the :root block
// (tokens live only in tokens.css).

describe('styles.css — file is non-empty', () => {
  it('file has content', () => {
    assert.ok(stylesCSS.trim().length > 0);
  });
});

describe('styles.css — @import dependency declaration', () => {
  it("@imports exactly one file", () => {
    assert.equal(importPathsIn(stylesCSS).length, 1);
  });

  it("@imports './tokens.css'", () => {
    assert.ok(
      importPathsIn(stylesCSS).includes('./tokens.css'),
      `imports found: ${importPathsIn(stylesCSS).join(', ')}`,
    );
  });

  it('@import is the first non-whitespace statement', () => {
    assert.match(stylesCSS.trimStart(), /^@import/);
  });
});

describe('styles.css — no :root block (tokens live only in tokens.css)', () => {
  it(':root is absent from the component stylesheet', () => {
    assert.ok(!stripComments(stylesCSS).includes(':root'));
  });
});

// ── styles.css — single source of truth for color values ─────────────────────
//
// Every color must be defined in tokens.css and consumed via var().  A
// hardcoded color literal in styles.css means the token system has been
// bypassed and a future design change requires editing both files.

describe('styles.css — no hardcoded hex color literals', () => {
  it('zero hex color patterns after comment stripping', () => {
    const found = hexColorsIn(stylesCSS);
    assert.deepEqual(
      found,
      [],
      `hardcoded hex colors found: ${found.join(', ')}`,
    );
  });
});

describe('styles.css — no hardcoded rgb() / rgba() literals', () => {
  it('zero rgb/rgba patterns after comment stripping', () => {
    const found = rgbLiteralsIn(stylesCSS);
    assert.deepEqual(
      found,
      [],
      `hardcoded rgb/rgba literals found: ${found.join(', ')}`,
    );
  });
});

// ── styles.css → tokens.css — variable resolution ────────────────────────────
//
// Every var(--x) reference in styles.css must have a matching declaration
// in tokens.css.  An unresolved reference silently falls back to the
// browser default (typically empty string), producing invisible breakage.

describe('styles.css — every var() reference resolves to a declared token', () => {
  for (const token of referencedTokens) {
    it(`var(${token}) is declared in tokens.css`, () => {
      assert.ok(
        declaredTokens.has(token),
        `'${token}' is referenced in styles.css but not declared in tokens.css`,
      );
    });
  }
});

// ── tokens.css — no dead tokens ───────────────────────────────────────────────
//
// Every declared token must be consumed by styles.css.  Dead tokens add noise
// to the token surface, mislead future editors, and never survive rotation of
// the design.

describe('tokens.css — every declared token is consumed by styles.css', () => {
  for (const token of declaredTokens) {
    it(`${token} is referenced in styles.css`, () => {
      assert.ok(
        referencedTokens.has(token),
        `'${token}' is declared in tokens.css but never referenced in styles.css`,
      );
    });
  }
});

// ── index.html — stylesheet integration ──────────────────────────────────────
//
// The page must load styles through the external stylesheet, never through an
// inline <style> block.  An inline block bypasses the token system and creates
// a parallel styling path that silently diverges from the design system.

describe('index.html — no inline <style> block', () => {
  it('<style> tag is absent', () => {
    assert.ok(!indexHTML.includes('<style'), 'inline <style> block found in index.html');
  });
});

describe('index.html — links the external stylesheet', () => {
  it('<link rel="stylesheet"> pointing to css/styles.css is present', () => {
    assert.match(indexHTML, /<link[^>]+href="css\/styles\.css"/);
  });

  it('css/tokens.css is not linked directly (styles.css @imports it)', () => {
    assert.ok(
      !indexHTML.includes('href="css/tokens.css"'),
      'index.html should not link tokens.css directly; styles.css @imports it',
    );
  });
});
