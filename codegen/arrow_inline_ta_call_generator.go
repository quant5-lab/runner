package codegen

import (
	"fmt"
	"strconv"

	"github.com/quant5-lab/runner/ast"
)

type ArrowInlineTACallGenerator struct {
	accessorFactory   *ArrowAwareAccessorFactory
	iifeRegistry      *InlineTAIIFERegistry
	signatureResolver *ArrowTACallSignatureResolver
	signatureRegistry *TAFunctionSignatureRegistry
}

func NewArrowInlineTACallGenerator(
	factory *ArrowAwareAccessorFactory,
	registry *InlineTAIIFERegistry,
) *ArrowInlineTACallGenerator {
	signatureRegistry := NewTAFunctionSignatureRegistry()
	return &ArrowInlineTACallGenerator{
		accessorFactory:   factory,
		iifeRegistry:      registry,
		signatureRegistry: signatureRegistry,
		signatureResolver: NewArrowTACallSignatureResolver(signatureRegistry),
	}
}

func (g *ArrowInlineTACallGenerator) GenerateInlineTACall(call *ast.CallExpression) (string, bool, error) {
	funcName := extractCallFunctionName(call)

	if !g.iifeRegistry.IsSupported(funcName) {
		return "", false, nil
	}

	if len(call.Arguments) < 1 {
		return "", false, nil
	}

	resolved, err := g.signatureResolver.ResolveCall(funcName, call)
	if err != nil {
		return "", false, fmt.Errorf("failed to resolve TA call signature for %s: %w", funcName, err)
	}

	var sourceExpr ast.Expression
	if resolved.NeedsDefaultSource {
		sourceExpr = &ast.Identifier{Name: resolved.DefaultSourceName}
	} else {
		sourceExpr = resolved.SourceExpr
	}

	periodExpr := NewConstantPeriod(1)
	if resolved.LengthExpr != nil {
		extractedPeriod, err := g.extractPeriod(resolved.LengthExpr)
		if err != nil {
			return "", false, fmt.Errorf("failed to extract period from expression: %w", err)
		}
		if extractedPeriod == 0 {
			return "", false, nil
		}
		periodExpr = NewConstantPeriod(extractedPeriod)
	}

	accessor, err := g.accessorFactory.CreateAccessorForExpression(sourceExpr)
	if err != nil {
		return "", false, fmt.Errorf("failed to create accessor for source expression: %w", err)
	}

	hasher := &ExpressionHasher{}
	sourceHash := hasher.Hash(sourceExpr)

	iifeCode, exists := g.iifeRegistry.Generate(funcName, accessor, periodExpr, sourceHash)
	if !exists {
		return "", false, nil
	}

	preamble := ""
	if preambleAccessor, ok := accessor.(interface{ GetPreamble() string }); ok {
		preamble = preambleAccessor.GetPreamble()
	}

	if preamble != "" {
		return fmt.Sprintf("func() float64 { %s\nreturn %s }()", preamble, iifeCode), true, nil
	}

	return iifeCode, true, nil
}

func (g *ArrowInlineTACallGenerator) extractPeriod(expr ast.Expression) (int, error) {
	switch e := expr.(type) {
	case *ast.Literal:
		if intVal, ok := e.Value.(int); ok {
			return intVal, nil
		}
		if floatVal, ok := e.Value.(float64); ok {
			return int(floatVal), nil
		}
		if strVal, ok := e.Value.(string); ok {
			return strconv.Atoi(strVal)
		}
	case *ast.Identifier:
		// Period is runtime parameter - inline IIFE requires compile-time constant
		// Signal NOT HANDLED so caller delegates to runtime TA generation
		return 0, nil
	}
	return 0, fmt.Errorf("unsupported period expression type: %T", expr)
}
