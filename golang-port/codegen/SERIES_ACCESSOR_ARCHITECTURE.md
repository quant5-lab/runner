# Series Expression Accessor Architecture

## Overview

Unified accessor architecture for converting AST expressions to series-aware Go code across both TA indicator context and arrow function context.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                   UNIFIED ARCHITECTURE                       │
└─────────────────────────────────────────────────────────────┘

TA Context                          Arrow Function Context
    │                                       │
    ├─ TAArgumentExtractor                 ├─ ArrowAwareAccessorFactory
    │  (complex expressions)                │  (binary/conditional)
    │                                       │
    └────────┬──────────────────────────────┘
             │
             ▼
    ┌────────────────────────┐
    │ SeriesExpressionAccessor│  ← UNIFIED COMPONENT
    │  - expr: Expression     │
    │  - symbolTable: Table   │
    └────────────────────────┘
             │
             │ delegates to
             ▼
    ┌────────────────────────┐
    │ SeriesAccessConverter   │  ← CORE ENGINE
    │  - symbolTable: Table   │
    │  - offset: string       │
    └────────────────────────┘
             │
             ├─ Converts: series var → varSeries.Get(offset)
             ├─ Preserves: scalar var → var
             └─ Transforms: builtin → ctx.Data[i-offset].Field
```

## Component Responsibilities (SRP)

### SeriesExpressionAccessor
**Single Responsibility:** Bridge between AST expressions and series-aware code generation

**Does:**
- Stores AST expression and symbol table
- Implements AccessGenerator interface
- Delegates conversion to SeriesAccessConverter

**Doesn't:**
- Perform AST traversal (delegates to converter)
- Make type decisions (uses symbol table)
- Generate complex code logic (simple delegation)

### SeriesAccessConverter
**Single Responsibility:** Transform AST to series-aware Go code

**Does:**
- Traverse AST expressions recursively
- Convert series identifiers to buffer access
- Preserve scalar identifiers unchanged
- Map builtin fields to ctx.Data access

**Doesn't:**
- Store state between calls (stateless converter)
- Make accessor decisions (focused on conversion)

### SymbolTable
**Single Responsibility:** Track variable type information

**Does:**
- Register variables with types (scalar/series)
- Lookup variable types
- Provide type predicates (IsSeries/IsScalar)

**Doesn't:**
- Generate code
- Traverse AST
- Perform conversions

## Design Principles Applied

### Single Responsibility Principle (SRP)
✓ Each component has ONE reason to change:
  - SeriesExpressionAccessor: Changes when accessor interface changes
  - SeriesAccessConverter: Changes when conversion rules change
  - SymbolTable: Changes when type system changes

### Don't Repeat Yourself (DRY)
✓ NO duplication:
  - BEFORE: ASTExpressionAccessor + SeriesConvertingExpressionAccessor (95% identical)
  - AFTER: Single SeriesExpressionAccessor (unified)

### Keep It Simple, Stupid (KISS)
✓ Simple delegation pattern:
  - Accessor stores context → delegates to converter
  - Converter performs transformation → returns code
  - No complex inheritance hierarchies
  - No over-engineered abstractions

## Usage Patterns

### TA Context (Complex Expressions)
```go
// ta.sma(close - open, 14)
sourceExpr := &ast.BinaryExpression{...}
accessor := NewSeriesExpressionAccessor(sourceExpr, symbolTable)

// Generate loop access: closeSeries.Get(j) - openSeries.Get(j)
code := accessor.GenerateLoopValueAccess("j")
```

### Arrow Function Context (Binary/Conditional)
```go
// (a, b) => condition ? a + b : a - b
condExpr := &ast.ConditionalExpression{...}
accessor := NewSeriesExpressionAccessor(condExpr, symbolTable)

// Generate series-aware conditional
code := accessor.GenerateLoopValueAccess("j")
```

## Consolidation Benefits

### Before (Redundant)
```
ast_expression_accessor.go           (1550 lines)
ast_expression_accessor_test.go      (3904 lines)
series_converting_expression_accessor.go  (2255 lines)
                                     ─────────────
TOTAL:                               7709 lines
```

### After (Unified)
```
series_expression_accessor.go        (70 lines)
series_access_converter.go           (280 lines)
series_access_converter_test.go      (200 lines)
                                     ─────────────
TOTAL:                               550 lines
```

**Reduction:** 7709 → 550 lines (93% reduction)

## Type Safety

### Series vs Scalar Distinction
```go
// Series variable: converted
sum → sumSeries.Get(offset)

// Scalar variable: preserved
period → period

// Builtin field: transformed
close → ctx.Data[i-offset].Close
```

## Extensibility

### Adding New Expression Types
1. Add case to SeriesAccessConverter.ConvertExpression()
2. Implement conversion logic
3. No changes needed to SeriesExpressionAccessor

### Adding New Contexts
1. Create new factory/extractor
2. Use SeriesExpressionAccessor
3. No new accessor classes needed

## Testing Strategy

### Unit Tests
- SeriesAccessConverter: Tests conversion logic
- SymbolTable: Tests type tracking
- SeriesExpressionAccessor: Tests delegation

### Integration Tests
- ta_complex_source_expression_test.go: Validates TA context
- arrow_function tests: Validates arrow context

## Migration Path

### Completed
✓ Renamed SeriesConvertingExpressionAccessor → SeriesExpressionAccessor
✓ Updated TAArgumentExtractor (TA context)
✓ Updated ArrowAwareAccessorFactory (arrow context)
✓ Updated ArrowFunctionTACallGenerator (nested TA in arrows)
✓ Removed redundant ASTExpressionAccessor
✓ Updated all tests
✓ Verified build passes
✓ Verified tests pass

### No Breaking Changes
- Public interface unchanged (AccessGenerator)
- Behavior unchanged (same conversion logic)
- Only internal implementation unified
