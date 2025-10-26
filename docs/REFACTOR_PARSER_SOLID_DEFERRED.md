# Parser RefactoriViolations:
- SRP: Converter does visiting + scoping + renaming
- OCP: Adding identifier mode requires modifying visit_Name()
- DRY: Parameter renaming duplicated in visit_FunctionDef()
- KISS: 3 transformation modes in single method
- YAGNI: Global wrapping dead code (zero usage in PineTS/runner)
```n: SOLID DRY KISS YAGNI

## Current State Assessment

```
parser.py: 660 LOC
├─ PyneToJsAstConverter (main class)
│  ├─ Scope management (_scope_chain)
│  ├─ Parameter renaming (_param_rename_stack)
│  ├─ AST visiting (25+ visit_* methods)
│  ├─ Identifier transformation (4 modes)
│  └─ Recursive AST mutation (_rename_identifiers_in_ast)
├─ Node classes (Script, Assign, BinOp, etc.)
└─ estree_node() helper

scope_chain.py: 82 LOC
└─ ScopeChain (single responsibility)

Violations:
- SRP: Converter does visiting + scoping + renaming + wrapping
- OCP: Adding new identifier mode requires modifying visit_Name()
- DRY: Identifier wrapping logic duplicated in visit_Name()
- KISS: 4 transformation modes in single method
```

## Target Architecture

```
services/pine-parser/
├─ parser.py                    # Entry point (50 LOC)
│  └─ parse_pine_to_js()
│
├─ ast/
│  ├─ nodes.py                  # AST node classes (100 LOC)
│  │  └─ Node, Script, Assign, BinOp, etc.
│  └─ estree.py                 # ESTree helpers (30 LOC)
│     └─ estree_node()
│
├─ converter/
│  ├─ base_converter.py         # Base visitor pattern (80 LOC)
│  │  ├─ visit()
│  │  └─ generic_visit()
│  │
│  ├─ js_converter.py           # Main conversion logic (200 LOC)
│  │  └─ visit_* methods
│  │
│  └─ identifier_strategy.py   # Strategy pattern (80 LOC)
│     ├─ IdentifierStrategy (interface)
│     ├─ BareIdentifier
│     └─ RenamedParameterIdentifier
│
├─ scope/
│  ├─ scope_chain.py            # Existing (82 LOC)
│  └─ scope_manager.py          # Higher-level API (60 LOC)
│     ├─ ScopeManager
│     ├─ enter_scope()
│     ├─ exit_scope()
│     └─ resolve_identifier()
│
├─ transform/
│  └─ parameter_transformer.py  # Shadowing detection (80 LOC)
│     ├─ ParameterTransformer
│     ├─ detect_shadowing()
│     └─ build_rename_map()
│
└─ input_function_transformer.py  # Existing

Tests:
├─ test_scope_chain.py          # Existing (196 LOC)
├─ test_parameter_shadowing.py  # Existing (284 LOC)
├─ test_identifier_strategy.py  # New (100 LOC)
├─ test_scope_manager.py        # New (80 LOC)
└─ test_js_converter.py         # New (150 LOC)
```

## Refactoring Steps

### Phase 1: Extract Node Classes (2h)
```
1. Create ast/nodes.py
2. Move all Node classes from parser.py
3. Update imports in parser.py
4. Run: python3 services/pine-parser/parser.py
5. Verify: No functional changes
```

### Phase 2: Extract ESTree Helper (1h)
```
1. Create ast/estree.py
2. Move estree_node() function
3. Update imports
4. Run: python3 services/pine-parser/parser.py
5. Verify: Output identical
```

### Phase 3: Extract Base Converter (3h)
```
1. Create converter/base_converter.py
2. Implement generic visitor pattern
3. Create converter/js_converter.py
4. Move visit_* methods from PyneToJsAstConverter
5. Keep scope/rename logic in js_converter temporarily
6. Run: python3 services/pine-parser/parser.py
7. Verify: E2E tests pass
```

### Phase 4: Implement Identifier Strategy (3h)
```
1. Create converter/identifier_strategy.py
2. Define IdentifierStrategy interface:
   - can_handle(var_name, scope_context)
   - transform(var_name, scope_context)
3. Implement 2 strategies:
   - BareIdentifier: Local vars, global scope access
   - RenamedParameterIdentifier: _param_<var>
4. Replace visit_Name() logic with strategy selection
5. Remove dead global wrapping code
6. Run: python3 services/pine-parser/test_parameter_shadowing.py
7. Verify: All 284 lines of tests pass
```

### Phase 5: Extract Scope Manager (3h)
```
1. Create scope/scope_manager.py
2. Wrap ScopeChain with higher-level API:
   - enter_scope(scope_type)
   - exit_scope()
   - resolve_identifier(name) → strategy
3. Move scope logic from js_converter to scope_manager
4. Update js_converter to use scope_manager
5. Run: python3 services/pine-parser/test_scope_chain.py
6. Verify: 196 lines of tests pass
```

### Phase 6: Extract Parameter Transformer (3h)
```
1. Create transform/parameter_transformer.py
2. Move shadowing detection logic:
   - _is_shadowing_parameter()
   - Parameter rename mapping
3. Move from js_converter to parameter_transformer
4. Update visit_FunctionDef() to use transformer
5. Run: python3 services/pine-parser/test_parameter_shadowing.py
6. Verify: All shadowing tests pass
```

### Phase 7: Remove Dead Code (1h)
```
1. Remove _rename_identifiers_in_ast() method (dead code)
2. Remove global wrapping logic from visit_Name()
3. Run: Full E2E test suite
4. Verify: All tests pass
```

### Phase 8: Integration & Cleanup (2h)
```
1. Update parser.py entry point
2. Remove old PyneToJsAstConverter class
3. Wire new components together
4. Run: Full test suite (E2E + unit)
5. Verify: Zero regressions
6. Remove remaining dead code
7. Update documentation
```

## SOLID Principles Applied

### Single Responsibility
```
Before: PyneToJsAstConverter does 3 jobs
After:  
- BaseConverter: Visitor pattern
- JsConverter: AST transformation
- IdentifierStrategy: Name resolution (2 modes)
- ScopeManager: Scope tracking
- ParameterTransformer: Shadowing detection
```

### Open/Closed
```
Before: Adding identifier mode requires modifying visit_Name()
After:  Add new IdentifierStrategy implementation
```

### Liskov Substitution
```
All IdentifierStrategy implementations:
- Accept same inputs (var_name, scope_context)
- Return same output (ESTree node)
- Substitutable without breaking behavior
```

### Interface Segregation
```
Before: Monolithic converter with all methods
After:  
- IdentifierStrategy: transform() only
- ScopeManager: scope operations only
- ParameterTransformer: shadowing detection only
```

### Dependency Inversion
```
Before: JsConverter depends on concrete ScopeChain
After:  JsConverter depends on ScopeManager abstraction
        ScopeManager wraps ScopeChain implementation
```

## DRY Improvements

```
Before: Parameter renaming in visit_FunctionDef()
        Identifier resolution in visit_Name()
After:  Single source of truth per concern
```

## KISS Improvements

```
Before: visit_Name() has 3 transformation modes + dead global wrapping
After:  Strategy pattern selects between 2 modes
        Each strategy: 1 mode, 1 responsibility
```

## YAGNI Validation

```
Keep:
- Scope chain (needed for variable tracking)
- Parameter shadowing (required by Pine Script semantics)

YAGNI VIOLATION DETECTED:
- Global wrapping ($.let.glb1_<var>): UNUSED by PineTS/runner
  Evidence: Zero references in PineTS bundle, runner, or tests
  Action: REMOVE Phase 4 GlobalWrappedIdentifier strategy
  
Remove:
- _param_rename_stack (replaced by ParameterTransformer state)
- Complex nested conditionals (replaced by strategies)
- Global wrapping logic in visit_Name() (dead code)
```

## Success Metrics

```
Maintainability:
- LOC per file: <200 (was 660)
- Cyclomatic complexity per method: <10
- Test coverage: >90%

Performance:
- Parse time: No regression (±5%)
- Memory: No growth

Quality:
- Zero test regressions
- All E2E tests pass
- Linting: Zero violations
```

## Rollback Plan

```
If refactoring fails:
1. git revert <refactor-commits>
2. Return to 24162c4
3. Document failure reason
4. Reassess approach
```

## Estimated Effort

```
Phase 1: 2h  (Extract nodes)
Phase 2: 1h  (Extract estree)
Phase 3: 3h  (Base converter)
Phase 4: 3h  (Identifier strategy - 2 strategies only)
Phase 5: 3h  (Scope manager)
Phase 6: 3h  (Parameter transformer)
Phase 7: 1h  (Remove dead code)
Phase 8: 2h  (Integration)
-------
Total:  18h
```
