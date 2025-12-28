package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

type BooleanConverter struct {
	typeSystem            *TypeInferenceEngine
	skipComparisonRule    ConversionRule
	seriesAccessRule      ConversionRule
	typeBasedRule         ConversionRule
	notEqualZeroTransform CodeTransformer
	parenthesesTransform  CodeTransformer
}

func NewBooleanConverter(typeSystem *TypeInferenceEngine) *BooleanConverter {
	comparisonMatcher := NewComparisonPattern()
	seriesMatcher := NewSeriesAccessPattern()

	return &BooleanConverter{
		typeSystem:            typeSystem,
		skipComparisonRule:    NewSkipComparisonRule(comparisonMatcher),
		seriesAccessRule:      NewConvertSeriesAccessRule(seriesMatcher),
		typeBasedRule:         NewTypeBasedRule(typeSystem),
		notEqualZeroTransform: NewAddNotEqualZeroTransformer(),
		parenthesesTransform:  NewAddParenthesesTransformer(),
	}
}

func (bc *BooleanConverter) EnsureBooleanOperand(expr ast.Expression, generatedCode string) string {
	if bc.IsAlreadyBoolean(expr) {
		return generatedCode
	}

	if bc.seriesAccessRule.ShouldConvert(expr, generatedCode) {
		return bc.parenthesesTransform.Transform(
			bc.notEqualZeroTransform.Transform(generatedCode),
		)
	}

	if bc.typeSystem.IsBoolVariable(expr) {
		return bc.parenthesesTransform.Transform(
			bc.notEqualZeroTransform.Transform(generatedCode),
		)
	}

	return generatedCode
}

func (bc *BooleanConverter) IsAlreadyBoolean(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.BinaryExpression:
		return bc.IsComparisonOperator(e.Operator)
	case *ast.LogicalExpression:
		return true
	case *ast.UnaryExpression:
		return e.Operator == "not" || e.Operator == "!"
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

	if ident, ok := call.Callee.(*ast.Identifier); ok {
		return ident.Name == "na"
	}

	return false
}

func (bc *BooleanConverter) IsFloat64SeriesAccess(code string) bool {
	return bc.seriesAccessRule.ShouldConvert(nil, code)
}

func (bc *BooleanConverter) ConvertBoolSeriesForIfStatement(expr ast.Expression, generatedCode string) string {
	// UnaryExpression with 'not' is already boolean
	if unary, ok := expr.(*ast.UnaryExpression); ok {
		if unary.Operator == "not" || unary.Operator == "!" {
			return generatedCode
		}
	}

	if call, isCall := expr.(*ast.CallExpression); isCall {
		if bc.IsBooleanFunction(call) {
			return generatedCode
		}
		return bc.notEqualZeroTransform.Transform(generatedCode)
	}

	if !bc.skipComparisonRule.ShouldConvert(expr, generatedCode) {
		return generatedCode
	}

	if bc.seriesAccessRule.ShouldConvert(expr, generatedCode) {
		return bc.notEqualZeroTransform.Transform(generatedCode)
	}

	if bc.typeBasedRule.ShouldConvert(expr, generatedCode) {
		return bc.notEqualZeroTransform.Transform(generatedCode)
	}

	return generatedCode
}
