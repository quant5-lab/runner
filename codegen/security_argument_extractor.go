package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

type SecurityArgumentExtractor struct {
	gen *generator
}

type ExtractionResult struct {
	Code      string
	IsRuntime bool
}

func NewSecurityArgumentExtractor(gen *generator) *SecurityArgumentExtractor {
	return &SecurityArgumentExtractor{gen: gen}
}

func (e *SecurityArgumentExtractor) ExtractSymbol(expr ast.Expression) (*ExtractionResult, error) {
	if expr == nil {
		return nil, fmt.Errorf("symbol expression is nil")
	}

	switch exp := expr.(type) {
	case *ast.Identifier:
		if exp.Name == "tickerid" {
			return &ExtractionResult{Code: "ctx.Symbol", IsRuntime: true}, nil
		}

		if e.gen != nil {
			if constVal, exists := e.gen.constants[exp.Name]; exists {
				if strVal, ok := constVal.(string); ok {
					return &ExtractionResult{Code: fmt.Sprintf("%q", strVal), IsRuntime: false}, nil
				}
			}
			if varType, exists := e.gen.variables[exp.Name]; exists && varType == "string" {
				return &ExtractionResult{Code: exp.Name, IsRuntime: true}, nil
			}
		}

		return &ExtractionResult{Code: fmt.Sprintf("%q", exp.Name), IsRuntime: false}, nil

	case *ast.CallExpression:
		if e.gen != nil {
			code, err := e.gen.generateExpression(exp)
			if err != nil {
				return nil, fmt.Errorf("failed to generate symbol call expression: %w", err)
			}
			return &ExtractionResult{Code: code, IsRuntime: true}, nil
		}
		return nil, fmt.Errorf("symbol CallExpression requires generator context")

	case *ast.MemberExpression:
		if exp.Object == nil || exp.Property == nil {
			return nil, fmt.Errorf("member expression has nil object or property")
		}

		obj, objOk := exp.Object.(*ast.Identifier)
		prop, propOk := exp.Property.(*ast.Identifier)

		if !objOk || !propOk {
			return nil, fmt.Errorf("member expression object or property is not an identifier")
		}

		if obj.Name == "syminfo" && (prop.Name == "tickerid" || prop.Name == "ticker") {
			return &ExtractionResult{Code: "ctx.Symbol", IsRuntime: true}, nil
		}
		return nil, fmt.Errorf("unsupported member expression for symbol: %s.%s", obj.Name, prop.Name)

	case *ast.Literal:
		if s, ok := exp.Value.(string); ok {
			return &ExtractionResult{Code: fmt.Sprintf("%q", s), IsRuntime: false}, nil
		}
		return nil, fmt.Errorf("invalid symbol literal type: %T", exp.Value)

	default:
		return nil, fmt.Errorf("unsupported symbol expression type: %T", expr)
	}
}

func (e *SecurityArgumentExtractor) ExtractTimeframe(expr ast.Expression) (*ExtractionResult, error) {
	if expr == nil {
		return nil, fmt.Errorf("timeframe expression is nil")
	}

	switch exp := expr.(type) {
	case *ast.Literal:
		if s, ok := exp.Value.(string); ok {
			normalized := e.normalizeTimeframe(strings.Trim(s, "'\""))
			return &ExtractionResult{Code: fmt.Sprintf("%q", normalized), IsRuntime: false}, nil
		}
		return nil, fmt.Errorf("invalid timeframe literal type: %T", exp.Value)

	case *ast.Identifier:
		if e.gen != nil {
			if constVal, exists := e.gen.constants[exp.Name]; exists {
				if strVal, ok := constVal.(string); ok {
					normalized := e.normalizeTimeframe(strVal)
					return &ExtractionResult{Code: fmt.Sprintf("%q", normalized), IsRuntime: false}, nil
				}
			}
			if varType, exists := e.gen.variables[exp.Name]; exists && varType == "string" {
				return &ExtractionResult{Code: exp.Name, IsRuntime: true}, nil
			}
		}
		return nil, fmt.Errorf("unsupported timeframe identifier: %s", exp.Name)

	case *ast.CallExpression:
		if e.gen != nil {
			code, err := e.gen.generateExpression(exp)
			if err != nil {
				return nil, fmt.Errorf("failed to generate timeframe call expression: %w", err)
			}
			return &ExtractionResult{Code: code, IsRuntime: true}, nil
		}
		return nil, fmt.Errorf("timeframe CallExpression requires generator context")

	case *ast.MemberExpression:
		if exp.Object == nil || exp.Property == nil {
			return nil, fmt.Errorf("member expression has nil object or property")
		}

		obj, objOk := exp.Object.(*ast.Identifier)
		prop, propOk := exp.Property.(*ast.Identifier)

		if !objOk || !propOk {
			return nil, fmt.Errorf("member expression object or property is not an identifier")
		}

		if obj.Name == "timeframe" && prop.Name == "period" {
			return &ExtractionResult{Code: "ctx.Timeframe", IsRuntime: true}, nil
		}
		return nil, fmt.Errorf("unsupported member expression for timeframe: %s.%s", obj.Name, prop.Name)

	default:
		return nil, fmt.Errorf("unsupported timeframe expression type: %T", expr)
	}
}

func (e *SecurityArgumentExtractor) normalizeTimeframe(tf string) string {
	switch tf {
	case "D":
		return "1D"
	case "W":
		return "1W"
	case "M":
		return "1M"
	default:
		return tf
	}
}
