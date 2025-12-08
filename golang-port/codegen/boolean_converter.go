package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

// BooleanConverter handles type conversions between Go bool and Pine float64 boolean model.
// Pine boolean model: 1.0 = true, 0.0 = false (float64)
// Go control flow: true/false (bool)
type BooleanConverter struct {
	typeSystem *TypeInferenceEngine
}

func NewBooleanConverter(typeSystem *TypeInferenceEngine) *BooleanConverter {
	return &BooleanConverter{
		typeSystem: typeSystem,
	}
}

func (bc *BooleanConverter) EnsureBooleanOperand(expr ast.Expression, generatedCode string) string {
	if bc.IsAlreadyBoolean(expr) {
		return generatedCode
	}

	if bc.IsFloat64SeriesAccess(generatedCode) {
		return fmt.Sprintf("(%s != 0)", generatedCode)
	}

	if bc.typeSystem.IsBoolVariable(expr) {
		return fmt.Sprintf("(%s != 0)", generatedCode)
	}

	return generatedCode
}

func (bc *BooleanConverter) IsAlreadyBoolean(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.BinaryExpression:
		return bc.IsComparisonOperator(e.Operator)
	case *ast.LogicalExpression:
		return true
	case *ast.CallExpression:
		return bc.IsBooleanFunction(e)
	default:
		return false
	}
}

func (bc *BooleanConverter) IsComparisonOperator(op string) bool {
	return op == ">" || op == "<" || op == ">=" || op == "<=" || op == "==" || op == "!="
}

func (bc *BooleanConverter) IsBooleanFunction(call *ast.CallExpression) bool {
	if member, ok := call.Callee.(*ast.MemberExpression); ok {
		if obj, ok := member.Object.(*ast.Identifier); ok {
			if prop, ok := member.Property.(*ast.Identifier); ok {
				funcName := obj.Name + "." + prop.Name
				return funcName == "ta.crossover" || funcName == "ta.crossunder"
			}
		}
	}
	return false
}

func (bc *BooleanConverter) IsFloat64SeriesAccess(code string) bool {
	return strings.Contains(code, ".GetCurrent()")
}

func (bc *BooleanConverter) ConvertBoolSeriesForIfStatement(expr ast.Expression, generatedCode string) string {
	needsConversion := false

	if ident, ok := expr.(*ast.Identifier); ok {
		if bc.typeSystem.IsBoolVariableByName(ident.Name) {
			needsConversion = true
		}
	}

	if member, ok := expr.(*ast.MemberExpression); ok {
		if ident, ok := member.Object.(*ast.Identifier); ok {
			if bc.typeSystem.IsBoolVariableByName(ident.Name) {
				needsConversion = true
			}
		}
	}

	if needsConversion {
		return fmt.Sprintf("%s != 0", generatedCode)
	}
	return generatedCode
}
