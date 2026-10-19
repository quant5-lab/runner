# Design System

## File Structure

```
out/css/tokens.css   — CSS custom properties (single source of truth for all tokens)
out/css/styles.css   — component styles (@import ./tokens.css; zero hardcoded values)
out/index.html       — <link rel="stylesheet" href="css/styles.css" />
```

Dependency graph: `index.html → styles.css → tokens.css`

## Tokens (`out/css/tokens.css`)

### Surfaces
| Token | Value | Usage |
|---|---|---|
| `--surface-base` | `#0f172a` | Deepest bg: body gradient start, table `<th>` |
| `--surface-card` | `#1a2035` | Section bg: `.trades-section`, `.info` |
| `--surface-raised` | `#1e293b` | Card bg: `#container`, `.chart-container` |
| `--surface-raised-alpha` | `rgba(30,41,59,0.85)` | Fullscreen toggle button bg |
| `--surface-hover` | `rgba(255,255,255,0.06)` | Row hover — visible on any dark surface |

### Borders
| Token | Value | Usage |
|---|---|---|
| `--border-subtle` | `#2d3748` | Inner dividers, row borders, section internals |
| `--border-default` | `#334155` | Outer borders: container, chart, buttons |

### Text
| Token | Value | Usage |
|---|---|---|
| `--text-primary` | `#cbd5e1` | Body text, card text |
| `--text-secondary` | `#94a3b8` | `<th>`, summary, button default label |
| `--text-muted` | `#6b7280` | Timestamp, `.no-trades` placeholder |

### Accent
| Token | Value | Usage |
|---|---|---|
| `--color-accent` | `#5eead4` | Section headings, hover state, focus ring |
| `--color-accent-alt` | `#2dd4bf` | h1 gradient end |
| `--color-accent-subtle` | `rgba(94,234,212,0.25)` | Pane resize handle hover fill |

### Actions
| Token | Value | Usage |
|---|---|---|
| `--color-action` | `#2563eb` | `.refresh-btn` bg |
| `--color-action-hover` | `#1d4ed8` | `.refresh-btn:hover` bg |

### Semantic
| Token | Value | Usage |
|---|---|---|
| `--color-long` | `#10b981` | Long direction, positive P/L |
| `--color-short` | `#ef4444` | Short direction, negative P/L |
| `--color-open` | `#3b82f6` | Open / in-progress trades |

### Trade Table
| Token | Value | Usage |
|---|---|---|
| `--stripe-tint` | `rgba(255,255,255,0.03)` | Alternating trade-pair row tint |

### Typography
| Token | Value | Usage |
|---|---|---|
| `--font-ui` | `'Segoe UI', system-ui, sans-serif` | All text |

### Border Radius
| Token | Value | Usage |
|---|---|---|
| `--radius-md` | `6px` | Most elements |
| `--radius-lg` | `8px` | Outer `#container` only |

### Shadows
| Token | Value | Usage |
|---|---|---|
| `--shadow-container` | `0 4px 6px -1px rgba(0,0,0,0.2)` | `#container` only |
| `--shadow-text` | `0 2px 4px rgba(0,0,0,0.3)` | `h1` text-shadow |

### Transitions
| Token | Value | Usage |
|---|---|---|
| `--transition-duration` | `0.2s` | All interactive elements |

## Spacing Scale
Rem-based: `0.5rem`, `0.75rem`, `1rem`, `1.25rem`, `1.5rem`, `2rem`, `2.5rem`
Pixel exceptions: `8px` (absolute inset positioning only)

## Font Scale
| Size | Usage |
|---|---|
| `2.5rem` | `h1` page title |
| `1.5rem` | `h2` section title |
| `1rem` | Body, action button |
| `0.875rem` | Table, timestamp, secondary labels, compact buttons |

Font weight `600`: profit values, table headers.

## Component Patterns

### Filled button (`.refresh-btn`)
- `background-color: var(--color-action)` → hover: `var(--color-action-hover)`
- `padding: 0.75rem 1.5rem`, `border-radius: var(--radius-md)`, `font-size: 1rem`
- `:focus-visible`: `outline: 2px solid var(--color-accent); outline-offset: 2px`
- `:active`: `transform: scale(0.97)`

### Ghost/outline button (`.sort-toggle-btn`, `.fullscreen-toggle-btn`)
- `background: transparent`, `border: 1px solid var(--border-default)`, `color: var(--text-secondary)`
- Hover: `border-color: var(--color-accent); color: var(--color-accent)`
- `:focus-visible`: same accent outline ring as filled button
- `:active`: `transform: scale(0.97)`

### Section card (`.trades-section`, `.info`)
- `background-color: var(--surface-card)`, `border: 1px solid var(--border-subtle)`, `border-radius: var(--radius-md)`, `padding: 1rem`

### Section header (`.trades-header`)
- `display: flex`, `justify-content: space-between`, `align-items: center`
- `margin-bottom: 1rem`, `padding-bottom: 0.5rem`, `border-bottom: 1px solid var(--border-subtle)`
- Title: `color: var(--color-accent)`

## Chart-specific (JS, not CSS)
Chart colors are LightweightCharts API options passed in `index.html` inline `<script>`.
They are intentionally separate from the CSS token system (different rendering context).

| Value | Usage |
|---|---|
| `#1e1e3c` | Chart layout background |
| `#DDD` | Chart text color |
| `#2B2B43` | Grid lines, scale borders |
| `#26a69a` / `#ef5350` | Candlestick up/down (TradingView convention) |
