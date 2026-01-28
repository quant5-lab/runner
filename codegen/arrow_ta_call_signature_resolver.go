package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type ResolvedTACall struct {
	SourceExpr         ast.Expression
	LengthExpr         ast.Expression
	NeedsDefaultSource bool
	DefaultSourceName  string
}

type ArrowTACallSignatureResolver struct {
	registry *TAFunctionSignatureRegistry
}

func NewArrowTACallSignatureResolver(registry *TAFunctionSignatureRegistry) *ArrowTACallSignatureResolver {
	return &ArrowTACallSignatureResolver{
		registry: registry,
	}
}

func (r *ArrowTACallSignatureResolver) ResolveCall(functionName string, call *ast.CallExpression) (*ResolvedTACall, error) {
	signature, exists := r.registry.GetSignature(functionName)
	if !exists {
		return nil, fmt.Errorf("unknown TA function: %s", functionName)
	}

	argCount := len(call.Arguments)

	switch signature.ArgumentPattern {
	case TAPatternSingleArgIsLength:
		return r.resolveSingleArgIsLength(signature, call, argCount)
	case TAPatternSingleArgIsSource:
		return r.resolveSingleArgIsSource(call, argCount)
	case TAPatternExplicitSourceAndLength:
		return r.resolveExplicitSourceAndLength(call, argCount)
	default:
		return nil, fmt.Errorf("unsupported argument pattern: %v", signature.ArgumentPattern)
	}
}

func (r *ArrowTACallSignatureResolver) resolveSingleArgIsLength(signature TAFunctionSignature, call *ast.CallExpression, argCount int) (*ResolvedTACall, error) {
	if argCount == 1 {
		return &ResolvedTACall{
			SourceExpr:         nil,
			LengthExpr:         call.Arguments[0],
			NeedsDefaultSource: true,
			DefaultSourceName:  signature.DefaultSource,
		}, nil
	}

	if argCount == 2 {
		return &ResolvedTACall{
			SourceExpr:         call.Arguments[0],
			LengthExpr:         call.Arguments[1],
			NeedsDefaultSource: false,
			DefaultSourceName:  "",
		}, nil
	}

	return nil, fmt.Errorf("expected 1 or 2 arguments, got %d", argCount)
}

func (r *ArrowTACallSignatureResolver) resolveSingleArgIsSource(call *ast.CallExpression, argCount int) (*ResolvedTACall, error) {
	if argCount == 1 {
		return &ResolvedTACall{
			SourceExpr:         call.Arguments[0],
			LengthExpr:         &ast.Literal{Value: "1"},
			NeedsDefaultSource: false,
			DefaultSourceName:  "",
		}, nil
	}

	if argCount == 2 {
		return &ResolvedTACall{
			SourceExpr:         call.Arguments[0],
			LengthExpr:         call.Arguments[1],
			NeedsDefaultSource: false,
			DefaultSourceName:  "",
		}, nil
	}

	return nil, fmt.Errorf("expected 1 or 2 arguments, got %d", argCount)
}

func (r *ArrowTACallSignatureResolver) resolveExplicitSourceAndLength(call *ast.CallExpression, argCount int) (*ResolvedTACall, error) {
	if argCount != 2 {
		return nil, fmt.Errorf("expected exactly 2 arguments, got %d", argCount)
	}

	return &ResolvedTACall{
		SourceExpr:         call.Arguments[0],
		LengthExpr:         call.Arguments[1],
		NeedsDefaultSource: false,
		DefaultSourceName:  "",
	}, nil
}
