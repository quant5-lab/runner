package codegen

import (
	"regexp"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

func isGoBoolLiteral(code string) bool {
	return code == "true" || code == "false"
}

func isLogicalNegation(operator string) bool {
	return operator == "not" || operator == "!"
}

/*
Matches code that is nothing but a Series accessor, the lone shape the series extractor emits for a
call it hoisted into a float64 binding; either accessor form counts, because an offset does not
change the Go type. IsFloat64SeriesAccess is the wider predicate: it also matches a series access
nested inside a larger expression, as in the Go bool na lowering of na(rsi14).
*/
var hoistedSeriesCodePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*Series\.Get(Current|\d+)\(\)$`)

func isHoistedSeriesCode(code string) bool {
	return hoistedSeriesCodePattern.MatchString(code)
}

/*
GoTypeResolver answers the Go type of a namespace member's lowering. The coercion decision reads
that type instead of going through the Pine type system, where barstate.islast has no bool type at
all and ta.crossover is bool although its emitted binding is a float64 series.
*/
type GoTypeResolver interface {
	ResolveMemberExpressionGoType(expr *ast.MemberExpression) (GoValueType, bool)
}

type BooleanConverter struct {
	typeSystem            *TypeInferenceEngine
	goTypes               GoTypeResolver
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

/*
SetGoTypeResolver attaches the namespace type oracle consulted for member operands. Without it a
member keeps its value.IsTrue coercion.
*/
func (bc *BooleanConverter) SetGoTypeResolver(resolver GoTypeResolver) {
	bc.goTypes = resolver
}

/*
EnsureBooleanOperand returns code as a Go boolean expression for a condition position reached from
the condition generators: an if statement, a logical operand or a variable initialiser. A numeric
operand is coerced with value.IsTrue, an operand that already evaluates to a Go bool is passed
through unchanged.
*/
func (bc *BooleanConverter) EnsureBooleanOperand(expr ast.Expression, generatedCode string) string {
	if expr == nil {
		return generatedCode
	}

	if bc.isGoBoolOperand(expr, generatedCode) {
		return generatedCode
	}

	if _, ok := expr.(*ast.Literal); ok {
		return generatedCode
	}

	if unary, ok := expr.(*ast.UnaryExpression); ok {
		if _, isLit := unary.Argument.(*ast.Literal); isLit {
			return generatedCode
		}
	}

	if bc.seriesAccessRule.ShouldConvert(expr, generatedCode) {
		return bc.parenthesesTransform.Transform(
			bc.notEqualZeroTransform.Transform(generatedCode),
		)
	}

	if bc.typeSystem.IsBoolVariable(expr) {
		return generatedCode
	}

	if _, ok := expr.(*ast.Identifier); ok && !strings.Contains(generatedCode, "Series") {
		return generatedCode
	}

	return bc.parenthesesTransform.Transform(
		bc.notEqualZeroTransform.Transform(generatedCode),
	)
}

func (bc *BooleanConverter) IsAlreadyBoolean(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.BinaryExpression:
		return bc.IsComparisonOperator(e.Operator)
	case *ast.LogicalExpression:
		return true
	case *ast.UnaryExpression:
		return isLogicalNegation(e.Operator)
	case *ast.CallExpression:
		return bc.IsBooleanFunction(e)
	case *ast.Literal:
		_, isBool := e.Value.(bool)
		return isBool
	default:
		return false
	}
}

/*
IsAlreadyGoBool reports whether code, the series-extractor lowering of expr, is already a Go bool,
so a condition position consumes it without a value.IsTrue coercion.

It is deliberately narrower than IsAlreadyBoolean, because the Pine-level type does not decide the
Go type: the series extractor lowers a logical expression to a float64 IIFE and ta.crossover to a
float64 series binding, while the value handler lowers na to a Go bool and the namespace resolver
types barstate.islast as GoBool.
*/
func (bc *BooleanConverter) IsAlreadyGoBool(expr ast.Expression, code string) bool {
	if isGoBoolLiteral(code) {
		return true
	}

	switch e := expr.(type) {
	case *ast.BinaryExpression:
		return bc.IsComparisonOperator(e.Operator)
	case *ast.UnaryExpression:
		/* The series extractor coerces the operand before negating it, so a negation is a Go bool. */
		return isLogicalNegation(e.Operator)
	case *ast.CallExpression:
		/* A hoisted boolean call carries a float64 series binding instead. */
		return bc.IsBooleanFunction(e) && !isHoistedSeriesCode(code)
	case *ast.MemberExpression:
		return bc.isGoBoolMember(e)
	default:
		return false
	}
}

func (bc *BooleanConverter) isGoBoolMember(member *ast.MemberExpression) bool {
	if bc.goTypes == nil {
		return false
	}
	goType, resolved := bc.goTypes.ResolveMemberExpressionGoType(member)
	return resolved && goType == GoBool
}

/*
isGoBoolOperand reports whether code, the condition lowering of expr, is already a Go bool. It is
IsAlreadyGoBool plus the logical case: a condition lowering emits a logical expression as
(left op right) over coerced operands, while the series extractor emits a float64 IIFE for the
same expression.
*/
func (bc *BooleanConverter) isGoBoolOperand(expr ast.Expression, code string) bool {
	if _, ok := expr.(*ast.LogicalExpression); ok {
		return true
	}
	return bc.IsAlreadyGoBool(expr, code)
}

/*
EnsureGoBool returns code as a Go boolean expression for a condition position reached from the
series extractor: a strategy argument or a logical operand of a series expression. A numeric
operand is coerced with value.IsTrue, an operand that already lowers to a Go bool is passed through
unchanged.
*/
func (bc *BooleanConverter) EnsureGoBool(expr ast.Expression, code string) string {
	if bc.IsAlreadyGoBool(expr, code) {
		return code
	}
	return "value.IsTrue(" + code + ")"
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
	if isGoBoolLiteral(generatedCode) {
		return generatedCode
	}

	if unary, ok := expr.(*ast.UnaryExpression); ok {
		if isLogicalNegation(unary.Operator) {
			return generatedCode
		}
	}

	// LogicalExpression is already boolean
	if _, ok := expr.(*ast.LogicalExpression); ok {
		return generatedCode
	}

	if _, ok := expr.(*ast.ConditionalExpression); ok {
		return bc.notEqualZeroTransform.Transform(generatedCode)
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
