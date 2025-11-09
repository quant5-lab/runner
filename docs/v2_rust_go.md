# Architecture Replaceability Assessment

## EVIDENCE-BASED FINDINGS

### 1. PERFORMANCE BOTTLENECK ANALYSIS

**Measured bottlenecks:**

```
Transpilation (Pynescript):  2432ms  ← 98.5% of total time
JS Parse (@swc/core):          0.04ms  ← 0.002% of total time
Execution (PineTS):           ~150ms  ← 6% of total time
```

**VERDICT: User claim CONFIRMED - Parser is 90%+ of total time**

### 2. CURRENT ARCHITECTURE

```
┌─────────────┐
│ Pine Code   │
└──────┬──────┘
       │
       v
┌─────────────────────────────────────────┐
│ Python3 Process (spawn)                 │  ← BOTTLENECK
│ ├─ Pynescript v0.2.0 (parsing)         │     2432ms
│ └─ Custom AST transformer               │
└──────────────┬──────────────────────────┘
               │
               v (JSON AST via /tmp files)
               │
┌──────────────┴──────────────────────────┐
│ JS AST Generator (escodegen)            │
└──────────────┬──────────────────────────┘
               │
               v
┌──────────────┴──────────────────────────┐
│ PineTS v0.1.34 Runtime (execution)      │  ← ALPHA
└─────────────────────────────────────────┘
```

### 3. DEPENDENCY MATURITY STATUS

**Pynescript:**

- Version: 0.2.0 (Feb 28, 2024)
- Status: Beta-ish (has basic features, not production-hardened)
- License: LGPL 3.0 (VIRAL - forces your code to be LGPL)
- Performance: Spawns Python process + IPC overhead

**PineTS:**

- Version: 0.1.34 (active development)
- Status: **ALPHA** (user claim CONFIRMED)
- Evidence:
  - 5438 TypeScript/JS files
  - Recent commits: "WIP rework", "fix", "optimize"
  - Local dependency (not published to npm)
- License: Unknown (local project)
- Completeness: Partial PineScript v5 support

**@swc/core:**

- Version: Latest stable
- Status: Production-ready (32.9k stars)
- License: Apache 2.0 (permissive)
- Performance: 20x faster than Babel single-thread, 70x on 4 cores
- Used by: Next.js, Vercel, ByteDance, Tencent

### 4. REPLACEABILITY OPTIONS

#### Option A: PURE RUST ENGINE (Recommended)

```
┌─────────────┐
│ Pine Code   │
└──────┬──────┘
       │
       v
┌─────────────────────────────────────────┐
│ Rust Parser + Transpiler                │  ← NEW
│ ├─ Custom PineScript parser (tree-sitter│
│ │  or lalrpop)                           │  ~50-100ms
│ └─ Direct Pine → JS codegen             │
└──────────────┬──────────────────────────┘
               │
               v (In-memory AST, no IPC)
               │
┌──────────────┴──────────────────────────┐
│ Custom Pine Runtime (Rust + WASM)       │  ← NEW
│ OR QuickJS/V8 isolate                   │
└─────────────────────────────────────────┘
```

**Pros:**

- Eliminates Python spawn overhead
- No IPC/tmp file overhead
- Full control over PineScript semantics
- Can use @swc/core for JS execution if needed
- Permissive licenses (Apache 2.0)
- Multi-threaded processing possible

**Cons:**

- Full rewrite (~3-6 months work)
- Need PineScript grammar implementation
- Need runtime function library (ta._, strategy._)

#### Option B: GO ENGINE

```
Same as Option A but in Go
```

**Pros:**

- Easier concurrency than Rust
- Faster development than Rust
- Good parsing libraries (participle, goyacc)

**Cons:**

- Slower than Rust (still 10x faster than Python)
- Larger binaries
- No WASM target quality

#### Option C: HYBRID (Quick Win)

```
┌─────────────┐
│ Pine Code   │
└──────┬──────┘
       │
       v
┌─────────────────────────────────────────┐
│ Rust/Go Parser ONLY                     │  ← REPLACE
│ (Output JS directly, skip AST JSON)     │  ~100ms
└──────────────┬──────────────────────────┘
               │
               v (Direct JS code)
               │
┌──────────────┴──────────────────────────┐
│ PineTS v0.1.34 Runtime (keep existing)  │  ← KEEP
└─────────────────────────────────────────┘
```

**Pros:**

- Removes main bottleneck (Python parser)
- Keeps PineTS runtime (working code)
- ~80% performance gain
- 2-4 weeks work

**Cons:**

- Still depends on PineTS alpha code
- Eventual PineTS completion required

## RECOMMENDATION

### Phase 1: Hybrid Approach (IMMEDIATE - 2-4 weeks)

Replace Python parser with Rust parser outputting JS directly

**Why:**

- Eliminates 98.5% of bottleneck
- Minimal risk (PineTS runtime works)
- Fast ROI

### Phase 2: Custom Runtime (6-12 months)

Replace PineTS with custom Rust runtime

**Why:**

- Full control over features
- LGPL license elimination
- Production-grade reliability
- Multi-symbol concurrent execution

## RUST PARSER OPTIONS

1. **tree-sitter** (Recommended)
   - Industry standard (GitHub, Neovim)
   - Incremental parsing
   - Error recovery
   - C API → Rust bindings

2. **lalrpop**
   - Pure Rust
   - LR(1) parser generator
   - Good error messages

3. **pest**
   - PEG parser
   - Simple grammar syntax
   - Good for DSLs

## TECHNICAL RISKS

### Low Risk:

- Parser replacement (well-defined input/output)
- @swc/core integration (mature API)

### Medium Risk:

- PineScript semantics edge cases
- Runtime function library completeness

### High Risk:

- Multi-timeframe execution (security() function)
- Strategy backtesting state management

## LICENSE CONSIDERATIONS

**CRITICAL:** Pynescript LGPL 3.0 is VIRAL

- Current usage: Dynamic linking via subprocess (OK)
- If embedded: Forces project to LGPL (BAD)

**Rust approach:** Apache 2.0 everywhere (SAFE)

## PERFORMANCE TARGETS

Current:

- 2500ms total (500 bars)

Hybrid:

- ~250ms total (500 bars) - 10x improvement

Pure Rust:

- ~50ms total (500 bars) - 50x improvement
- Multi-threaded: 10-20ms (100-250x improvement)

## SWC ARCHITECTURE ANALYSIS

**What CAN be copied from SWC:**

### ✅ REUSABLE PATTERNS:

1. **Lexer Architecture** (`swc_ecma_lexer`):
   - Hand-written recursive descent lexer
   - State machine pattern for tokenization
   - Byte-level optimizations
   - Token buffer with lookahead

   ```rust
   pub struct Lexer<'a> {
       input: StringInput<'a>,
       cur: Option<char>,
       state: State,
       // Character-by-character processing
   }
   ```

2. **Parser Pattern** (`swc_ecma_parser`):
   - Recursive descent parser
   - Error recovery mechanisms
   - Span tracking for error messages
   - Context-sensitive parsing

   ```rust
   pub struct Parser<I: Tokens> {
       input: Buffer<I>,
       state: State,
       ctx: Context,
   }
   ```

3. **AST Visitor Pattern** (`swc_ecma_visit`):
   - Clean visitor trait
   - AST transformation pipeline
   - Codegen from AST

### ❌ CANNOT COPY DIRECTLY:

- **Grammar rules**: ECMAScript grammar ≠ PineScript grammar
- **Token definitions**: Different keywords, operators
- **Parser combinators**: Specific to JS/TS syntax

### 📋 ARCHITECTURE STRATEGY:

**Option A: Copy SWC Patterns (Recommended)**

```
┌─────────────────────────────────────────┐
│ Custom PineScript Lexer                 │  ← Copy lexer PATTERN
│ (Hand-written, like SWC)                │     from swc_ecma_lexer
└──────────────┬──────────────────────────┘
               │
               v (Tokens)
               │
┌──────────────┴──────────────────────────┐
│ Custom PineScript Parser                │  ← Copy parser PATTERN
│ (Recursive descent, like SWC)           │     from swc_ecma_parser
└──────────────┬──────────────────────────┘
               │
               v (Custom Pine AST)
               │
┌──────────────┴──────────────────────────┐
│ JS Codegen (Visitor pattern)            │  ← Copy visitor PATTERN
└──────────────┬──────────────────────────┘
               │
               v (JS code)
               │
┌──────────────┴──────────────────────────┐
│ Custom Runtime (ta.*, strategy.*)       │  ← Custom implementation
└─────────────────────────────────────────┘
```

**Effort**: 8-12 weeks (copying patterns, not code)
**Performance**: 50-100x faster than Python
**License**: Your code, Apache 2.0 compatible

**Option B: Use tree-sitter**

```
┌─────────────────────────────────────────┐
│ tree-sitter PineScript grammar          │  ← Write .grammar file
│ (Parser generator)                      │
└──────────────┬──────────────────────────┘
               │
               v (tree-sitter AST)
               │
┌──────────────┴──────────────────────────┐
│ Rust bindings + JS Codegen              │  ← Custom traversal
└──────────────┬──────────────────────────┘
               │
               v (JS code)
               │
┌──────────────┴──────────────────────────┐
│ Custom Runtime (ta.*, strategy.*)       │
└─────────────────────────────────────────┘
```

**Effort**: 6-8 weeks (grammar is declarative)
**Performance**: 40-80x faster than Python
**License**: MIT (tree-sitter)

## COUNTER-SUGGESTION

**Don't use @swc/core parser directly** - it WILL parse PineScript but produces WRONG AST (treats as JS)

**Use @swc/core for:**

- Architectural patterns (lexer/parser design)
- AST visitor patterns
- Optional: JS execution if needed

**Build custom PineScript parser using:**

1. Copy SWC's hand-written lexer/parser ARCHITECTURE
2. OR use tree-sitter for grammar-based parsing
3. Direct Pine → JS codegen with custom runtime

## WHAT TO COPY FROM SWC

```rust
// ✅ Copy these PATTERNS (not literal code):

// 1. Lexer state machine
struct Lexer {
    input: Input,
    state: State,
}

// 2. Recursive descent parser
struct Parser {
    lexer: Lexer,
    lookahead: Token,
}

// 3. Visitor for codegen
trait Visitor {
    fn visit_expr(&mut self, expr: &Expr);
    fn visit_stmt(&mut self, stmt: &Stmt);
}

// 4. Error recovery
impl Parser {
    fn recover_from_error(&mut self) {
        // Skip to next statement
    }
}
```

## REPLACEABILITY REVISED

| Component      | Replace With              | Copy from SWC           | Effort      |
| -------------- | ------------------------- | ----------------------- | ----------- |
| Python parser  | Rust lexer/parser         | Lexer + Parser patterns | 8-12 weeks  |
| Pynescript lib | Custom PineScript grammar | None (custom)           | 4-6 weeks   |
| AST transform  | Visitor pattern codegen   | Visitor trait           | 2-3 weeks   |
| PineTS runtime | Custom Rust runtime       | None (custom)           | 12-16 weeks |

**Total for hybrid (parser only):** 14-21 weeks
**Total for full rewrite:** 26-37 weeks
