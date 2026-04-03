package codegen

import "github.com/quant5-lab/runner/ast"

type BBWMultExtractor struct {
	exprGen ArrowExpressionGenerator
}

func NewBBWMultExtractor(exprGen ArrowExpressionGenerator) *BBWMultExtractor {
	return &BBWMultExtractor{exprGen: exprGen}
}

type BBWMultResult struct {
	IsLiteral  bool
	Literal    float64
	Expression string
}

func (e *BBWMultExtractor) Extract(call *ast.CallExpression) (*BBWMultResult, error) {
	const defaultMult = 2.0

	multArgIndex := e.determineMultIndex(call)
	if multArgIndex < 0 {
		return &BBWMultResult{
			IsLiteral: true,
			Literal:   defaultMult,
		}, nil
	}

	multArg := call.Arguments[multArgIndex]

	if lit, ok := multArg.(*ast.Literal); ok {
		if floatVal, ok := lit.Value.(float64); ok {
			return &BBWMultResult{
				IsLiteral: true,
				Literal:   floatVal,
			}, nil
		}
	}

	rendered, err := e.exprGen.Generate(multArg)
	if err != nil {
		return nil, err
	}

	return &BBWMultResult{
		IsLiteral:  false,
		Expression: rendered,
	}, nil
}

func (e *BBWMultExtractor) determineMultIndex(call *ast.CallExpression) int {
	argCount := len(call.Arguments)

	if argCount == 2 {
		return 1
	}

	if argCount >= 3 {
		return 2
	}

	return -1
}
