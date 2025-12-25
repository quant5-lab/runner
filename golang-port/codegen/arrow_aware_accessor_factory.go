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
	exprGenerator      *ArrowExpressionGeneratorImpl
}

func NewArrowAwareAccessorFactory(
	resolver *ArrowIdentifierResolver,
	exprGen *ArrowExpressionGeneratorImpl,
) *ArrowAwareAccessorFactory {
	return &ArrowAwareAccessorFactory{
		identifierResolver: resolver,
		exprGenerator:      exprGen,
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

	// Try other builtin resolution (high, low, close, etc.)
	code, resolved := f.exprGenerator.gen.builtinHandler.TryResolveIdentifier(id, false)
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

	return nil, fmt.Errorf("identifier '%s' not registered in arrow context", id.Name)
}

func (f *ArrowAwareAccessorFactory) createBinaryAccessor(binExpr *ast.BinaryExpression) (AccessGenerator, error) {
	binaryCode, err := f.exprGenerator.Generate(binExpr)
	if err != nil {
		return nil, fmt.Errorf("failed to generate binary expression for accessor: %w", err)
	}

	return NewInlineLoopExpressionAccessorWithResolver(binaryCode, f.identifierResolver), nil
}

func (f *ArrowAwareAccessorFactory) createCallAccessor(call *ast.CallExpression) (AccessGenerator, error) {
	callCode, err := f.exprGenerator.Generate(call)
	if err != nil {
		return nil, fmt.Errorf("failed to generate call expression for accessor: %w", err)
	}

	return NewInlineLoopExpressionAccessorWithResolver(callCode, f.identifierResolver), nil
}

func (f *ArrowAwareAccessorFactory) createConditionalAccessor(cond *ast.ConditionalExpression) (AccessGenerator, error) {
	condCode, err := f.exprGenerator.Generate(cond)
	if err != nil {
		return nil, fmt.Errorf("failed to generate conditional expression for accessor: %w", err)
	}

	return NewInlineLoopExpressionAccessorWithResolver(condCode, f.identifierResolver), nil
}
