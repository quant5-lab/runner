# Parser Dead Code Removal: Evidence-Based Minimal Plan

## Runtime Evidence

```
✅ USED:
- _param_ prefix: 255 test assertions (parameter shadowing)
- _param_rename_stack: Runtime parameter mapping
- _is_shadowing_parameter: Required by BB Strategy 7
- ScopeChain: Tested (196 LOC tests)

❌ DEAD CODE:
- $.let.glb1_ wrapping: Zero references (PineTS/runner/tests)
- _rename_identifiers_in_ast: Only self-recursive, zero calls
- Global wrapping in visit_Name: Lines 329-342
```

## Current State

```
parser.py: 659 LOC
├─ Node classes: 163 LOC (lines 11-173)
├─ estree_node: 5 LOC (lines 167-171)  
├─ PyneToJsAstConverter: 491 LOC
│  ├─ 20 visit_* methods
│  ├─ visit_Name: 27 LOC (19 LOC dead code)
│  ├─ _rename_identifiers_in_ast: 18 LOC (DEAD)
│  └─ Working: _scope_chain, _param_rename_stack
```

## COUNTER-SUGGESTION: Skip Full Refactoring

```
Original Plan: 18h effort, 8 phases
Risk: HIGH (visitor pattern changes, strategy extraction)
Benefit: Marginal (complexity 2/10 → slightly better)

Evidence Against Full Refactoring:
1. Node classes: Self-contained (163 LOC, 37 classes)
2. estree_node: Trivial utility (5 LOC)
3. Scope chain: Already extracted
4. Parameter shadowing: Working (255 tests pass)
5. Zero functional gain

Minimal Plan: 2h effort, remove dead code only
Risk: LOW (surgical removal)
Benefit: YAGNI compliance + 19 LOC reduction
```

## Surgical Removal Plan (2h Total)

### Phase 1: Remove Global Wrapping (30min)

```python
# BEFORE (visit_Name lines 318-345):
def visit_Name(self, node):
    var_name = node.id
    
    if self._param_rename_stack:
        current_mapping = self._param_rename_stack[-1]
        if var_name in current_mapping:
            return estree_node('Identifier', name=current_mapping[var_name])
    
    # DEAD CODE (19 LOC) ⬇
    if self._scope_chain.depth() > 0:
        if not self._scope_chain.is_declared_in_current_scope(var_name):
            if self._scope_chain.is_global(var_name):
                return estree_node('MemberExpression',
                    object=estree_node('MemberExpression',
                        object=estree_node('Identifier', name='$'),
                        property=estree_node('Identifier', name='let'),
                        computed=False
                    ),
                    property=estree_node('Identifier', name=f'glb1_{var_name}'),
                    computed=False
                )
    # DEAD CODE ⬆
    
    return estree_node('Identifier', name=var_name)

# AFTER (8 LOC):
def visit_Name(self, node):
    var_name = node.id
    
    if self._param_rename_stack:
        current_mapping = self._param_rename_stack[-1]
        if var_name in current_mapping:
            return estree_node('Identifier', name=current_mapping[var_name])
    
    return estree_node('Identifier', name=var_name)
```

Commands:
```bash
# Edit parser.py lines 318-345
# Run validation
docker compose exec runner python3 services/pine-parser/test_parameter_shadowing.py
# Expect: All 11 tests pass ✅
```

### Phase 2: Remove AST Rewriter (30min)

```python
# REMOVE lines 186-203:
def _rename_identifiers_in_ast(self, node, param_mapping):
    """Recursively rename identifiers in AST based on param_mapping"""
    if not param_mapping or not node:
        return node
    
    if isinstance(node, dict):
        if node.get('type') == 'Identifier' and node.get('name') in param_mapping:
            node['name'] = param_mapping[node['name']]
        
        for key, value in node.items():
            if isinstance(value, (dict, list)):
                self._rename_identifiers_in_ast(value, param_mapping)
    
    elif isinstance(node, list):
        for item in node:
            self._rename_identifiers_in_ast(item, param_mapping)
    
    return node

# Evidence: grep -r "_rename_identifiers_in_ast" shows only self-recursion
```

Commands:
```bash
# Remove method
# Run validation
docker compose exec runner python3 services/pine-parser/parser.py < strategies/test.pine
# Expect: Output identical ✅
```

### Phase 3: Validation (1h)

```bash
# Unit tests
pnpm test
# Expect: 515/515 pass ✅

# E2E tests
pnpm e2e
# Expect: 7/7 pass ✅

# Real strategy
docker compose exec runner node src/index.js CHMF M 72 strategies/rolling-cagr-5-10yr.pine
# Expect: 12 CAGR values ✅
```

## Impact Analysis

```
LOC Reduction:
- visit_Name: 27 → 8 LOC (-19)
- _rename_identifiers_in_ast: 18 → 0 LOC (-18)
- Total: 659 → 622 LOC (-37, 5.6%)

Complexity:
- Before: 2/10 (manageable)
- After:  4/10 (simpler)
- vs Full Refactoring: 7/10 (overcomplicated)

Effort:
- Minimal: 2h
- Full: 18h
- Saved: 16h (89%)

Risk:
- Minimal: LOW (surgical)
- Full: HIGH (architectural)
```

## Decision Matrix

```
         | Effort | Risk | Benefit | ROI
---------|--------|------|---------|-----
Minimal  |   2h   | LOW  | YAGNI   | ✅ HIGH
Full     |  18h   | HIGH | Marginal| ❌ LOW
```

## When to Revisit Full Refactoring

```
Triggers:
1. Adding 3rd identifier transformation mode
2. ScopeChain requires changes
3. Parser becomes performance bottleneck
4. Team grows to 3+ developers

Current Reality:
- 1 developer
- 2 transformation modes (stable)
- ScopeChain working
- Parser fast enough
```

## Updated TODO.md Entry

```markdown
- [ ] **TECH DEBT: Remove parser dead code**
  - **Status**: NOT STARTED
  - **Goal**: Remove $.let.glb1_ wrapping and _rename_identifiers_in_ast
  - **Effort**: 2h
  - **Plan**: docs/REFACTOR_PARSER_MINIMAL.md
```
