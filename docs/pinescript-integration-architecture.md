# Pine Script File Import Architecture Plan

## Objective

Enable direct `.pine` file import and execution using **pynescript → PineTS transpilation bridge**.

## Architecture Decision

### Primary Approach: Python/Docker with pynescript

**Rationale**:

- ✅ pynescript provides **native `.pine` file parsing** (PineTS does not)
- ✅ Proven transpilation path exists: `.pine` → pynescript AST → JavaScript → PineTS
- ✅ Reference implementation: [arose26/pinestuff](https://github.com/arose26/pinestuff)
- ⚠️ **Docker REQUIRED** for consistent Python 3.10+ environment

### Why Docker is Mandatory

- Python 3.10+ dependency for pynescript
- escodegen (Node.js) + pynescript (Python) multi-language stack
- Consistent environment across development and production
- Isolated dependency management (no system pollution)

## TODO: Implementation Plan

### Phase 1: Docker Environment Setup

- [ ] Create `Dockerfile` with Node.js 18 + Python 3.10+
- [ ] Install system dependencies: `python3-dev`, `build-base`
- [ ] Create `docker-compose.yml` for development workflow
- [ ] Configure volume mounts for live code reload
- [ ] Set up health checks for container services

### Phase 2: Python Parser Service

- [ ] Create `services/pine-parser/requirements.txt` with pynescript>=0.2.0
- [ ] Create `services/pine-parser/setup.sh` for pip install
- [ ] Install escodegen: `pnpm add escodegen`
- [ ] Create `services/pine-parser/parser.py` based on arose26/pinestuff
- [ ] Implement Pine Script AST → JavaScript AST converter class
- [ ] Handle Pine Script v4/v5 version detection
- [ ] Add error handling for malformed `.pine` files

### Phase 3: Transpilation Bridge

- [ ] Create `src/pine/PineScriptTranspiler.js` - Node.js wrapper
- [ ] Implement Python subprocess spawning for parser service
- [ ] Handle stdin/stdout communication between Node.js ↔ Python
- [ ] Parse pynescript AST output into JavaScript code
- [ ] Integrate escodegen for JavaScript AST formatting
- [ ] Add caching layer for transpiled strategies

### Phase 4: Strategy Loader

- [ ] Create `strategies/` directory for `.pine` files
- [ ] Implement `src/pine/PineScriptLoader.js` for file reading
- [ ] Add file validation (syntax, version, annotations)
- [ ] Implement strategy metadata extraction (title, parameters)
- [ ] Create file watcher for development hot-reload
- [ ] Add strategy registry for multiple `.pine` files

### Phase 5: Execution Integration

- [ ] Create `src/pine/StrategyExecutor.js` for transpiled code
- [ ] Map transpiled JavaScript to PineTS context.run()
- [ ] Handle Pine Script `input.*` parameters
- [ ] Convert Pine Script `plot()` calls to PineTS plots
- [ ] Implement Pine Script `strategy.*` to trading signals
- [ ] Handle Pine Script `alert()` conditions

### Phase 6: Testing & Validation

- [ ] Create test suite for pynescript transpilation
- [ ] Add example `.pine` strategies (EMA, RSI, MACD)
- [ ] Validate transpiled code against original behavior
- [ ] Benchmark performance: `.pine` vs inline JavaScript
- [ ] Test edge cases: complex indicators, nested conditionals
- [ ] Verify all Pine Script v5 technical analysis functions

### Phase 7: Production Deployment

- [ ] Optimize Docker image size (multi-stage build)
- [ ] Add transpilation result caching (Redis/filesystem)
- [ ] Implement error recovery and fallback strategies
- [ ] Create monitoring for parser service health
- [ ] Add strategy execution timeouts
- [ ] Document deployment procedures

## Architecture Components

### Dockerfile

```dockerfile
FROM node:18-alpine

# Install Python and build dependencies
RUN apk add --no-cache python3 py3-pip python3-dev build-base

WORKDIR /app

# Copy PineTS sibling dependency
COPY PineTS /PineTS

# Install Node.js dependencies
COPY runner/package.json runner/pnpm-lock.yaml ./
RUN npm install -g pnpm@10 && pnpm install --frozen-lockfile

# Install Python dependencies
COPY runner/services/pine-parser/requirements.txt ./services/pine-parser/
RUN pip3 install -r services/pine-parser/requirements.txt

# Copy application code
COPY runner ./

EXPOSE 8080

CMD ["pnpm", "start"]
```

### requirements.txt

```
pynescript>=0.2.0
```

### PineScriptTranspiler (Node.js Bridge)

```javascript
import { spawn } from 'child_process';
import escodegen from 'escodegen';

class PineScriptTranspiler {
  async transpile(pineScriptCode) {
    /* Spawn Python parser subprocess */
    const python = spawn('python3', ['services/pine-parser/parser.py']);

    /* Send Pine Script code via stdin */
    python.stdin.write(pineScriptCode);
    python.stdin.end();

    /* Collect pynescript AST output */
    const astOutput = await this.collectOutput(python.stdout);

    /* Parse AST and convert to JavaScript */
    const jsAst = JSON.parse(astOutput);

    /* Generate formatted JavaScript code */
    return escodegen.generate(jsAst);
  }

  collectOutput(stream) {
    return new Promise((resolve, reject) => {
      let data = '';
      stream.on('data', (chunk) => (data += chunk));
      stream.on('end', () => resolve(data));
      stream.on('error', reject);
    });
  }
}
```

### PineScriptLoader

```javascript
import fs from 'fs/promises';
import path from 'path';

class PineScriptLoader {
  constructor(transpiler) {
    this.transpiler = transpiler;
    this.strategiesDir = './strategies';
  }

  async loadStrategy(filename) {
    /* Read .pine file */
    const filePath = path.join(this.strategiesDir, filename);
    const pineCode = await fs.readFile(filePath, 'utf-8');

    /* Transpile to JavaScript */
    const jsCode = await this.transpiler.transpile(pineCode);

    /* Return executable function */
    return this.wrapStrategy(jsCode);
  }

  wrapStrategy(jsCode) {
    /* Wrap transpiled code in PineTS-compatible function */
    return new Function(
      'context',
      `
      const { data, ta, plot } = context
      const { open, high, low, close, volume } = data
      ${jsCode}
    `,
    );
  }
}
```

### StrategyExecutor

```javascript
class StrategyExecutor {
  async execute(strategyFn, marketData, symbol, timeframe, bars) {
    /* Create PineTS instance with market data */
    const pineTS = new PineTS(marketData, symbol, timeframe, bars);
    await pineTS.ready();

    /* Execute transpiled strategy */
    const { plots } = await pineTS.run(strategyFn);

    /* Return signals and plots */
    return { plots, data: marketData };
  }
}
```

## File Structure (Target)

```
runner/
├── src/
│   ├── classes/
│   │   ├── CandlestickDataSanitizer.js
│   │   ├── ConfigurationBuilder.js
│   │   ├── JsonFileWriter.js
│   │   ├── Logger.js
│   │   ├── PineScriptStrategyRunner.js
│   │   ├── ProviderManager.js
│   │   └── TradingAnalysisRunner.js
│   ├── providers/
│   │   ├── AlphaVantageProvider.js
│   │   ├── MoexProvider.js
│   │   └── YahooFinanceProvider.js
│   ├── pine/
│   │   ├── PineScriptLoader.js    # NEW: Load .pine files
│   │   ├── PineScriptTranspiler.js # NEW: Node.js ↔ Python bridge
│   │   └── StrategyExecutor.js    # NEW: Execute transpiled strategies
│   ├── config.js
│   ├── container.js
│   └── index.js
├── strategies/                 # NEW: Pine Script files
│   ├── ema_cross.pine
│   ├── rsi_divergence.pine
│   ├── macd_strategy.pine
│   └── custom/
├── services/
│   └── pine-parser/
│       ├── parser.py          # NEW: pynescript → AST bridge
│       ├── requirements.txt
│       └── setup.sh
├── tests/
├── out/
│   └── index.html
├── docs/
├── Dockerfile                  # NEW: Multi-language container
├── docker-compose.yml          # NEW: Development orchestration
├── package.json
├── pnpm-lock.yaml
└── vitest.config.js
```

## Technical Details: arose26/pinestuff Implementation

### Transpilation Workflow

```
.pine file → pynescript.parse() → Python AST → dump() → eval()
  → PyneToJsAstConverter.visit() → ESTree AST → escodegen.generate()
  → JavaScript code → PineTS execution
```

### Key Components from pinestuff

#### 1. PyneToJsAstConverter Class

Visitor pattern converting pynescript AST nodes to ESTree JavaScript AST:

- **Script** → Program with body array
- **Assign** → VariableDeclaration with VariableDeclarator
- **ReAssign** → AssignmentExpression
- **BinOp** → BinaryExpression (Add→'+', Sub→'-', Mult→'\*', Div→'/', Gt→'>', etc.)
- **Call** → CallExpression with callee and arguments
- **FunctionDef** → FunctionDeclaration
- **If** → IfStatement with test, consequent, alternate
- **While** → WhileStatement
- **ForTo** → ForStatement with init, test, update

#### 2. Operator Mappings

```python
# pynescript → JavaScript operators
Add → '+'
Sub → '-'
Mult → '*'
Div → '/'
Gt → '>'
GtE → '>='
Lt → '<'
LtE → '<='
Eq → '==='
NotEq → '!=='
And → '&&'
Or → '||'
Not → '!'
```

#### 3. Pine Script Specific Handling

- **input.int/float/bool**: Detected and converted to JavaScript variables
- **ta.\* functions**: Mapped to PineTS technical analysis context
- **Series access**: `close[1]` → array indexing `close[context.bar - 1]`
- **plot()**: Converted to context.plot() calls

#### 4. Performance Note

From arose26/pinestuff documentation:

> "Pynescript dump is very slow and probably not useful for real-time parsing"

**Mitigation Strategy**: Implement aggressive caching of transpiled JavaScript code.

## Technical Challenges & Solutions

### Challenge 1: Slow pynescript.dump()

**Evidence**: arose26/pinestuff notes "Pynescript dump is very slow"

- [ ] Solution: Implement caching layer for parsed AST
- [ ] Solution: Cache transpiled JavaScript by .pine file hash
- [ ] Solution: Use file watcher to invalidate cache on changes

### Challenge 2: Pine Script v4/v5 Syntax Differences

- [ ] Detect version from `//@version=X` annotation
- [ ] Map v4 `security()` to v5 `request.security()`
- [ ] Handle deprecated functions with compatibility layer

### Challenge 3: Context Mapping

- [ ] Map `strategy.*` declarations to trading signals
- [ ] Convert `input.*` parameters to JavaScript variables
- [ ] Handle `alert()` conditions as event emitters
- [ ] Support `plotshape()` and `plotchar()` annotations

### Challenge 4: Series vs Simple Types

- [ ] Detect series operations (e.g., `close[1]`)
- [ ] Map to PineTS array indexing
- [ ] Handle `ta.*` function series outputs

### Challenge 5: IPC Performance

**Evidence**: Python ↔ Node.js subprocess adds ~100ms overhead

- [ ] Solution: Batch multiple .pine files in single subprocess call
- [ ] Solution: Keep parser subprocess alive (persistent process)
- [ ] Solution: Use Unix domain sockets instead of stdio pipes

## Performance Optimization Strategy

### Baseline Metrics

- Current system: <1s for 100 candles (pure JavaScript)
- Target with .pine import: <2s for 100 candles (includes transpilation)

### Optimization TODO

- [ ] Implement in-memory AST cache (avoid re-parsing)
- [ ] Pre-transpile strategies at container startup
- [ ] Use persistent Python process pool
- [ ] Measure and profile each pipeline stage
- [ ] Add performance monitoring to StrategyExecutor

## Development Workflow

### Container Development

```bash
# Build and start containers
docker-compose up --build

# Hot reload enabled for both Node.js and Python
# - Node.js: --watch flag
# - .pine files: file watcher triggers re-transpilation

# Access chart visualization
open http://localhost:8080

# Execute with specific .pine strategy
STRATEGY=ema_cross.pine SYMBOL=BTCUSDT docker-compose exec app pnpm start
```

### Testing Transpilation

```bash
# Test Python parser directly
docker-compose exec app python3 services/pine-parser/parser.py strategies/ema_cross.pine

# Test full transpilation pipeline
docker-compose exec app node -e "
  import('./src/pine/PineScriptTranspiler.js').then(({ PineScriptTranspiler }) => {
    const t = new PineScriptTranspiler();
    t.transpile('indicator(\"Test\", overlay=true)\nplot(close)').then(console.log);
  })
"
```

## Migration Path

### From Current JavaScript Implementation

- ✅ Keep existing inline strategy support
- ✅ Add .pine file loader as alternative input method
- ✅ Use same ProviderManager, PineScriptStrategyRunner, TradingAnalysisRunner
- ✅ Extend IoC container with PineScriptLoader, PineScriptTranspiler
- ✅ Add Python setup to package.json preinstall script

### Coexistence Strategy

```javascript
// Option 1: Inline JavaScript (current)
SYMBOL=BTCUSDT pnpm start

// Option 2: .pine file import (new)
STRATEGY=strategies/ema_cross.pine SYMBOL=BTCUSDT pnpm start
```

## Risk Assessment

### High Risk: Transpilation Correctness

- **Mitigation**: Comprehensive test suite comparing .pine vs JavaScript output
- **Mitigation**: Validate against TradingView Pine Script reference behavior

### Medium Risk: Performance Degradation

- **Mitigation**: Aggressive caching of transpiled code
- **Mitigation**: Performance benchmarking in CI/CD pipeline

### Medium Risk: Python Dependency Management

- **Mitigation**: Docker ensures consistent Python 3.10+ environment
- **Mitigation**: Pin exact pynescript version in requirements.txt

### Low Risk: Docker Complexity

- **Mitigation**: docker-compose provides simple `up` command
- **Mitigation**: Document common Docker operations

## Success Criteria

- [ ] Successfully transpile arose26/pinestuff demo1.pine example
- [ ] Execute transpiled strategy over BTCUSDT data
- [ ] Generate chart output matching JavaScript inline strategy
- [ ] Transpilation time <500ms per .pine file (with caching)
- [ ] End-to-end execution time <2s for 100 candles
- [ ] Support Pine Script v5 technical analysis functions (ta.\*)
- [ ] Handle complex strategies with nested conditions
- [ ] Zero runtime crashes from transpiled code

## Next Steps (Priority Order)

1. **Immediate**: Set up Docker environment with Python + Node.js
2. **Week 1**: Implement Python parser service based on arose26/pinestuff
3. **Week 2**: Create Node.js transpilation bridge with subprocess IPC
4. **Week 3**: Integrate with existing PineScriptStrategyRunner
5. **Week 4**: Test with real .pine strategies and optimize performance
