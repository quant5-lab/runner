package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/*
ArrowAwareAccessorFactory creates AccessGenerator instances that are arrow-context-aware.

Responsibility (SRP):
  - Single purpose: create appropriate accessor for expression types in arrow context
  - Factory pattern: encapsulates complex accessor creation logic
  - No code generation logic (delegates to created accessors)

Design Rationale:
  - Centralizes accessor creation decisions
  - Maintains separation between accessor types and their creation
  - Extensible: easy to add new expression types
*/
type ArrowAwareAccessorFactory struct {
	identifierResolver *ArrowIdentifierResolver
	exprGenerator      ArrowExpressionGenerator
	gen                *generator
	symbolTable        SymbolTable
}

func NewArrowAwareAccessorFactory(
	resolver *ArrowIdentifierResolver,
	exprGen ArrowExpressionGenerator,
	gen *generator,
	symbolTable SymbolTable,
) *ArrowAwareAccessorFactory {
	return &ArrowAwareAccessorFactory{
		identifierResolver: resolver,
		exprGenerator:      exprGen,
		gen:                gen,
		symbolTable:        symbolTable,
	}
}

/*
CreateAccessorForExpression creates an arrow-aware accessor for any expression.
Supports identifiers, binary expressions, call expressions, and conditionals.
*/
func (f *ArrowAwareAccessorFactory) CreateAccessorForExpression(expr ast.Expression) (AccessGenerator, error) {

	switch e := expr.(type) {
	case *ast.Identifier:
		return f.createIdentifierAccessor(e)

	case *ast.MemberExpression:
		return f.createMemberAccessor(e)

	case *ast.BinaryExpression:
		return f.createBinaryAccessor(e)

	case *ast.CallExpression:
		return f.createCallAccessor(e)

	case *ast.ConditionalExpression:
		return f.createConditionalAccessor(e)

	default:
		return nil, fmt.Errorf("unsupported expression type for arrow accessor: %T", expr)
	}
}

func (f *ArrowAwareAccessorFactory) createIdentifierAccessor(id *ast.Identifier) (AccessGenerator, error) {
	// Check builtins FIRST - they take precedence over local variables
	// This prevents builtins like 'tr' from being mistakenly treated as Series variables
	if id.Name == "tr" {
		return NewBuiltinTrueRangeAccessor(), nil
	}

	// Check OHLCV builtins - use arrow-specific accessor (ctx.Data pattern)
	// Arrow functions can't access main scope Series variables like highSeries
	if isOHLCVBuiltin(id.Name) {
		return NewArrowOHLCVFieldAccessGenerator(capitalizeFirstLetter(id.Name)), nil
	}

	// Try other builtin resolution (high, low, close, etc.)
	code, resolved := f.gen.builtinHandler.TryResolveIdentifier(id, false)
	if resolved {
		return NewBuiltinIdentifierAccessor(code), nil
	}

	// Check local variables and parameters
	if f.identifierResolver.IsLocalVariable(id.Name) {
		return NewArrowAwareSeriesAccessor(id.Name), nil
	}

	if f.identifierResolver.IsParameter(id.Name) {
		return NewArrowFunctionParameterAccessor(id.Name), nil
	}

	// Fallback: treat as series variable
	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.ClassifyAST(id)
	return CreateAccessGenerator(sourceInfo), nil
}

func (f *ArrowAwareAccessorFactory) createMemberAccessor(member *ast.MemberExpression) (AccessGenerator, error) {
	obj, okObj := member.Object.(*ast.Identifier)
	if !okObj {
		return f.createFallbackMemberAccessor(member)
	}

	prop, okProp := member.Property.(*ast.Identifier)

	if okProp && obj.Name == "ta" && prop.Name == "tr" {
		return NewBuiltinTrueRangeAccessor(), nil
	}

	if okProp && obj.Name == "ctx" {
		fieldName := capitalizeFirstLetter(prop.Name)
		// Use arrow-specific accessor - arrow functions can't access main scope Series
		return NewArrowOHLCVFieldAccessGenerator(fieldName), nil
	}

	code, resolved := f.gen.builtinHandler.TryResolveMemberExpression(member, false)
	if resolved {
		return NewBuiltinIdentifierAccessor(code), nil
	}

	return f.createFallbackMemberAccessor(member)
}

func (f *ArrowAwareAccessorFactory) createFallbackMemberAccessor(member *ast.MemberExpression) (AccessGenerator, error) {
	if f.symbolTable != nil {
		return NewSeriesExpressionAccessor(member, f.symbolTable, nil), nil
	}

	// Without symbolTable, we can't safely handle arbitrary member expressions
	return nil, fmt.Errorf("unsupported member expression in TA call: %T", member)
}

func capitalizeFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-32) + s[1:]
	}
	return s
}

/* isOHLCVBuiltin checks if identifier is an OHLCV builtin (open, high, low, close, volume) */
func isOHLCVBuiltin(name string) bool {
	switch name {
	case "open", "high", "low", "close", "volume":
		return true
	default:
		return false
	}
}

func (f *ArrowAwareAccessorFactory) createBinaryAccessor(binExpr *ast.BinaryExpression) (AccessGenerator, error) {
	if f.symbolTable != nil {
		return NewSeriesExpressionAccessor(binExpr, f.symbolTable, nil), nil
	}

	tempVarName := "binary_source_temp"

	binaryCode, err := f.exprGenerator.Generate(binExpr)
	if err != nil {
		return nil, fmt.Errorf("failed to generate binary expression for accessor: %w", err)
	}

	return &FixnanCallExpressionAccessor{
		tempVarName: tempVarName,
		tempVarCode: fmt.Sprintf("%s := %s", tempVarName, binaryCode),
		exprCode:    binaryCode,
	}, nil
}

func (f *ArrowAwareAccessorFactory) createCallAccessor(call *ast.CallExpression) (AccessGenerator, error) {
	callCode, err := f.exprGenerator.Generate(call)
	if err != nil {
		return nil, fmt.Errorf("failed to generate call expression for accessor: %w", err)
	}

	/* TA calls used as sources need series-backed storage for historical lookback */
	funcName := extractCallFunctionName(call)
	if isTAFunction(funcName) {
		hasher := &ExpressionHasher{}
		hash := hasher.Hash(call)
		seriesName := fmt.Sprintf("_ta_src_%s", hash)
		return NewArrowCtxSeriesAccessor(seriesName, callCode), nil
	}

	tempVarName := "call_source_temp"
	return &FixnanCallExpressionAccessor{
		tempVarName: tempVarName,
		tempVarCode: fmt.Sprintf("%s := %s", tempVarName, callCode),
		exprCode:    callCode,
	}, nil
}

func (f *ArrowAwareAccessorFactory) createConditionalAccessor(cond *ast.ConditionalExpression) (AccessGenerator, error) {
	if f.symbolTable != nil {
		return NewSeriesExpressionAccessor(cond, f.symbolTable, nil), nil
	}

	tempVarName := "ternary_source_temp"

	condCode, err := f.exprGenerator.Generate(cond)
	if err != nil {
		return nil, fmt.Errorf("failed to generate conditional expression for accessor: %w", err)
	}

	return &FixnanCallExpressionAccessor{
		tempVarName: tempVarName,
		tempVarCode: fmt.Sprintf("%s := %s", tempVarName, condCode),
		exprCode:    condCode,
	}, nil
}
