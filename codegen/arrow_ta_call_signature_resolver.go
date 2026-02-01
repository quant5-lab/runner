package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

type ResolvedTACall struct {
	SourceExpr         ast.Expression
	LengthExpr         ast.Expression
	NeedsDefaultSource bool
	DefaultSourceName  string
}

type ArrowTACallSignatureResolver struct {
	legacyRegistry    *TAFunctionSignatureRegistry
	signatureRegistry *TASignatureRegistry
	callResolver      *TACallResolver
}

func NewArrowTACallSignatureResolver(registry *TAFunctionSignatureRegistry) *ArrowTACallSignatureResolver {
	signatureRegistry := NewTASignatureRegistry()
	return &ArrowTACallSignatureResolver{
		legacyRegistry:    registry,
		signatureRegistry: signatureRegistry,
		callResolver:      NewTACallResolver(signatureRegistry),
	}
}

func (r *ArrowTACallSignatureResolver) ResolveCall(functionName string, call *ast.CallExpression) (*ResolvedTACall, error) {
	newResolved, err := r.callResolver.Resolve(functionName, call)
	if err != nil {
		return nil, err
	}

	return r.adaptToLegacyFormat(newResolved), nil
}

func (r *ArrowTACallSignatureResolver) adaptToLegacyFormat(resolved *TAResolvedCall) *ResolvedTACall {
	var sourceExpr ast.Expression
	var lengthExpr ast.Expression

	if resolved.DefaultSourceApplied {
		if len(resolved.SeriesArguments) > 1 {
			sourceExpr = resolved.SeriesArguments[1]
		}
	} else {
		if len(resolved.SeriesArguments) > 0 {
			sourceExpr = resolved.SeriesArguments[0]
		}
	}

	if len(resolved.ScalarArguments) > 0 {
		lengthExpr = resolved.ScalarArguments[0]
	} else if (resolved.FunctionName == "ta.change" || resolved.FunctionName == "change") && len(resolved.SeriesArguments) > 0 {
		lengthExpr = &ast.Literal{Value: "1"}
	}

	return &ResolvedTACall{
		SourceExpr:         sourceExpr,
		LengthExpr:         lengthExpr,
		NeedsDefaultSource: resolved.DefaultSourceApplied,
		DefaultSourceName:  resolved.DefaultSourceName,
	}
}
