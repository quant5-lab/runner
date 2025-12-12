# Grammar Alignment with pynescript ASDL

## Current State (grammar.go)
Our participle-based grammar implements:
- ✅ Script with Statements
- ✅ Assignment (maps to Assign)
- ✅ IfStatement (maps to If)
- ✅ ExpressionStmt (maps to Expr)
- ✅ CallExpr (maps to Call)
- ✅ MemberAccess (maps to Attribute/Subscript)
- ✅ Comparison (maps to Compare)
- ✅ Boolean/Number/String literals (maps to Constant)

## pynescript ASDL Reference

### Statement Types
```asdl
stmt = FunctionDef | TypeDef | Assign | ReAssign | AugAssign 
     | Import | Expr | Break | Continue
```

### Expression Types
```asdl
expr = BoolOp | BinOp | UnaryOp | Conditional | Compare | Call | Constant
     | Attribute | Subscript | Name | Tuple
     | ForTo | ForIn | While | If | Switch
     | Qualify | Specialize
```

### Key Findings

1. **Subscript Handling (Critical)**
   - pynescript: `Subscript(expr value, expr? slice, expr_context ctx)`
   - Our grammar: MemberExpression with computed property
   - Python parser treats `close[0]` as Subscript, we parse as MemberExpression
   - ✅ Our converter correctly maps Subscript → MemberExpression

2. **Series Access Pattern**
   - Built-ins (close, open, high, low, volume) are arrays in PineTS
   - Subscript access `close[1]` accesses history (previous bars)
   - Our crossover implementation correctly uses ctx.Data[i-1].Close for prev

3. **Missing Features**
   - ⚠️ AugAssign (+=, -=, *=, /=)
   - ⚠️ ForTo, ForIn loops
   - ⚠️ While loops
   - ⚠️ Switch statements
   - ⚠️ Type qualifiers (const, input, simple, series)
   - ⚠️ Tuple assignments
   - ⚠️ UnaryOp (not, +, -)
   - ⚠️ BoolOp (and/or with multiple operands)

4. **Operator Mapping Consistency**
   ```
   pynescript → JavaScript → Go
   And → && → &&
   Or → || → ||
   Eq → === → ==
   NotEq → !== → !=
   ```
   ✅ Our generateConditionExpression correctly maps these

5. **Indentation-Based Parsing**
   - pynescript uses Python-style indentation
   - Our grammar uses statement terminators
   - Known issue: Sequential if statements nest incorrectly
   - Solution: Need indentation-aware lexer or explicit block delimiters

## Recommendations

### Priority 1: Current Implementation Stability
- ✅ Crossover/crossunder with builtin series works
- ⚠️ Need series history tracking for user variables
- ⚠️ Need proper if block parsing

### Priority 2: Core Language Features
- Implement AugAssign for += operator (common in strategies)
- Add UnaryOp for negation (-x, not condition)
- Add BoolOp for multi-operand logical expressions

### Priority 3: Advanced Features
- ForTo/ForIn loops (required for indicator calculations)
- Type qualifiers (needed for var/varip distinction)
- Tuple assignments (for ta.macd return values)

## Implementation Strategy

1. **Grammar Extensions (Phase 1)**
   - Add AugAssign to Statement alternatives
   - Add UnaryOp to Expression
   - Extend BinOp to handle all arithmetic operators

2. **Converter Updates (Phase 2)**
   - Map AugAssign to AssignmentExpression with operator
   - Handle UnaryOp in expression conversion
   - Support tuple destructuring for multi-return functions

3. **Codegen Enhancements (Phase 3)**
   - Generate += style operators
   - Handle unary expressions
   - Implement series history storage for crossover with calcu
lated series

## Cross-Reference

- pynescript ASDL: `/opt/homebrew/lib/python3.13/site-packages/pynescript/ast/grammar/asdl/resource/Pinescript.asdl`
- Python parser: `services/pine-parser/parser.py`
- Our grammar: `golang-port/parser/grammar.go`
- Our converter: `golang-port/parser/converter.go`
