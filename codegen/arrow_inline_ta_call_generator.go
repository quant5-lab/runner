package codegen

import (
	"fmt"
	"strconv"

	"github.com/quant5-lab/runner/ast"
)

/*
ArrowInlineTACallGenerator generates inline TA function calls with arrow-context awareness.

Responsibility (SRP):
  - Single purpose: generate arrow-aware inline TA calls (rma, sma, ema, etc.)
  - Uses ArrowAwareAccessorFactory to create proper accessors
  - Delegates IIFE generation to InlineTAIIFERegistry
  - No knowledge of expression evaluation or identifier resolution

Design:
  - Composition: uses factory and registry for separation of concerns
  - DRY: reuses existing IIFE generators, only provides arrow-aware accessors
  - KISS: simple delegation pattern, no complex logic
*/
type ArrowInlineTACallGenerator struct {
	accessorFactory *ArrowAwareAccessorFactory
	iifeRegistry    *InlineTAIIFERegistry
}

func NewArrowInlineTACallGenerator(
	factory *ArrowAwareAccessorFactory,
	registry *InlineTAIIFERegistry,
) *ArrowInlineTACallGenerator {
	return &ArrowInlineTACallGenerator{
		accessorFactory: factory,
		iifeRegistry:    registry,
	}
}

/* Generates arrow-aware inline TA function with proper accessor for source expression */
func (g *ArrowInlineTACallGenerator) GenerateInlineTACall(call *ast.CallExpression) (string, bool, error) {
	funcName := extractCallFunctionName(call)

	if !g.iifeRegistry.IsSupported(funcName) {
		return "", false, nil
	}

	if len(call.Arguments) < 1 {
		return "", false, fmt.Errorf("inline TA function '%s' requires at least 1 argument", funcName)
	}

	sourceExpr := call.Arguments[0]

	periodExpr := NewConstantPeriod(1)
	if len(call.Arguments) >= 2 {
		periodArg := call.Arguments[1]
		extractedPeriod, err := g.extractPeriod(periodArg)
		if err != nil {
			return "", false, fmt.Errorf("failed to extract period for '%s': %w", funcName, err)
		}
		if extractedPeriod == 0 {
			return "", false, nil
		}
		periodExpr = NewConstantPeriod(extractedPeriod)
	}

	accessor, err := g.accessorFactory.CreateAccessorForExpression(sourceExpr)
	if err != nil {
		return "", false, fmt.Errorf("failed to create accessor for '%s': %w", funcName, err)
	}

	hasher := &ExpressionHasher{}
	sourceHash := hasher.Hash(sourceExpr)

	iifeCode, exists := g.iifeRegistry.Generate(funcName, accessor, periodExpr, sourceHash)
	if !exists {
		return "", false, fmt.Errorf("IIFE generator not found for '%s'", funcName)
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
