package codegen

import (
	"fmt"
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
	// Check if last argument is ObjectExpression (named args container)
	if len(args) == 0 {
		return defaultValue
	}

	lastArg := args[len(args)-1]
	objExpr, isObject := lastArg.(*ast.ObjectExpression)

	// Try named argument first
	if isObject {
		for _, prop := range objExpr.Properties {
			if prop.Key == nil {
				continue
			}
			keyIdent, isIdent := prop.Key.(*ast.Identifier)
			if !isIdent || keyIdent.Name != argName {
				continue
			}

			// Extract string literal or identifier
			return e.extractCommentValue(prop.Value)
		}
	}

	// Try positional argument
	if positionalIndex >= 0 && positionalIndex < len(args) {
		arg := args[positionalIndex]
		// Skip ObjectExpression (named args)
		if _, isObj := arg.(*ast.ObjectExpression); !isObj {
			return e.extractCommentValue(arg)
		}
	}

	return defaultValue
}

/*
extractCommentValue converts AST expression to Go string literal or identifier.
*/
func (e *ArgumentExtractor) extractCommentValue(expr ast.Expression) string {
	switch v := expr.(type) {
	case *ast.Literal:
		// String literal: "Buy signal" → "Buy signal"
		if str, ok := v.Value.(string); ok {
			return fmt.Sprintf("%q", str)
		}
	case *ast.Identifier:
		// Variable reference: signal_msg → signal_msgSeries.GetCurrent()
		// In PineScript, string variables are stored as Series<string>
		return fmt.Sprintf("%sSeries.GetCurrent()", v.Name)
	case *ast.ConditionalExpression:
		// Ternary: condition ? "true_str" : "false_str"
		// Generate: func() string { if condition { return "true_str" } else { return "false_str" } }()
		condition := e.generator.extractSeriesExpression(v.Test)
		trueValue := e.extractCommentValue(v.Consequent)
		falseValue := e.extractCommentValue(v.Alternate)
		return fmt.Sprintf("func() string { if (%s != 0) { return %s } else { return %s } }()",
			strings.TrimRight(condition, "\n"), trueValue, falseValue)
	}
	return `""`
}
