package codegen

import (
	"strings"

	"github.com/quant5-lab/runner/ast"
)

/*
ArgumentExtractor extracts named and positional arguments from Pine Script function calls.

Single Responsibility: Parse AST arguments into named key-value pairs
Design: Stateless extractor - pure function transformation
*/
type ArgumentExtractor struct {
	generator *generator
}

/*
ExtractNamedArgument extracts value from named argument (stop=48000).
Returns generated code string and success boolean.

Pine named args are converted to ObjectExpression with Properties:

	strategy.exit("Exit", "Long", stop=48000, limit=58000)
	→  Arguments: ["Exit", "Long", ObjectExpression{Properties: [{Key: "stop", Value: 48000}, {Key: "limit", Value: 58000}]}]
*/
func (e *ArgumentExtractor) ExtractNamedArgument(args []ast.Expression, argName string) (string, bool) {
	// Check if last argument is ObjectExpression (named args container)
	if len(args) == 0 {
		return "", false
	}

	lastArg := args[len(args)-1]
	objExpr, isObject := lastArg.(*ast.ObjectExpression)
	if !isObject {
		return "", false
	}

	// Search for named argument in Properties
	for _, prop := range objExpr.Properties {
		if prop.Key == nil {
			continue
		}
		keyIdent, isIdent := prop.Key.(*ast.Identifier)
		if !isIdent || keyIdent.Name != argName {
			continue
		}

		// Use extractSeriesExpression for proper identifier/series resolution
		code := e.generator.extractSeriesExpression(prop.Value)
		return strings.TrimRight(code, "\n"), true
	}
	return "", false
}

/*
ExtractPositionalArgument extracts value at specific index.
Returns generated code string and success boolean.
*/
func (e *ArgumentExtractor) ExtractPositionalArgument(args []ast.Expression, index int) (string, bool) {
	if index < 0 || index >= len(args) {
		return "", false
	}

	// Use extractSeriesExpression for proper identifier/series resolution
	code := e.generator.extractSeriesExpression(args[index])
	return strings.TrimRight(code, "\n"), true
}

/*
ExtractNamedOrPositional tries named extraction first, falls back to positional.
Returns generated code string or default value if not found.
*/
func (e *ArgumentExtractor) ExtractNamedOrPositional(args []ast.Expression, argName string, positionalIndex int, defaultValue string) string {
	if value, found := e.ExtractNamedArgument(args, argName); found {
		return value
	}
	if value, found := e.ExtractPositionalArgument(args, positionalIndex); found {
		return value
	}
	return defaultValue
}

/*
ExtractCommentArgument extracts comment parameter as quoted string literal or identifier.
Used for strategy.entry/close/exit comment parameters.
Returns Go code string (quoted literal or identifier) or default value.
*/
func (e *ArgumentExtractor) ExtractCommentArgument(args []ast.Expression, argName string, positionalIndex int, defaultValue string) string {
	expr, ok := newStrategyCallArguments(args).value(argName, positionalIndex)
	if !ok {
		return defaultValue
	}
	return e.extractCommentValue(expr)
}

/*
ExtractConditionArgument extracts a boolean condition parameter (when=buySignal) as a Go boolean
expression, coercing a numeric operand and leaving an already-boolean operand bare through
BooleanConverter.EnsureGoBool.
*/
func (e *ArgumentExtractor) ExtractConditionArgument(args []ast.Expression, argName string) (string, bool) {
	if len(args) == 0 {
		return "", false
	}

	lastArg := args[len(args)-1]
	objExpr, isObject := lastArg.(*ast.ObjectExpression)
	if !isObject {
		return "", false
	}

	for _, prop := range objExpr.Properties {
		if prop.Key == nil {
			continue
		}
		keyIdent, isIdent := prop.Key.(*ast.Identifier)
		if !isIdent || keyIdent.Name != argName {
			continue
		}

		operandCode := e.generator.extractSeriesExpression(prop.Value)
		conditionCode := e.generator.boolConverter.EnsureGoBool(prop.Value, operandCode)
		return strings.TrimRight(conditionCode, "\n"), true
	}
	return "", false
}

/*
extractCommentValue converts AST expression to Go string literal or identifier.
*/
func (e *ArgumentExtractor) extractCommentValue(expr ast.Expression) string {
	code, _ := e.commentExpression(expr)
	return code
}

/*
ExtractWhenCondition extracts the when= parameter of a strategy call as a Go boolean expression.
*/
func (e *ArgumentExtractor) ExtractWhenCondition(args []ast.Expression) (string, bool) {
	return e.ExtractConditionArgument(args, "when")
}
