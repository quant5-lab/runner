package codegen

import (
	"fmt"
	"regexp"
	"strings"
)

/*
InlineLoopExpressionAccessor transforms expressions for inline TA loops.

Transforms:
  - .GetCurrent() → .Get(loopVar) for historical series access
  - Local variables → varNameSeries.Get(loopVar) when series exists

Used for complex expressions in inline RMA/EMA/SMA where source is a ternary/binary
expression referencing series variables with accumulated history.
*/
type InlineLoopExpressionAccessor struct {
	expressionCode string
	resolver       *ArrowIdentifierResolver
}

func NewInlineLoopExpressionAccessor(expressionCode string) *InlineLoopExpressionAccessor {
	return &InlineLoopExpressionAccessor{
		expressionCode: expressionCode,
		resolver:       nil,
	}
}

func NewInlineLoopExpressionAccessorWithResolver(expressionCode string, resolver *ArrowIdentifierResolver) *InlineLoopExpressionAccessor {
	return &InlineLoopExpressionAccessor{
		expressionCode: expressionCode,
		resolver:       resolver,
	}
}

func (a *InlineLoopExpressionAccessor) GenerateLoopValueAccess(loopVar string) string {
	return a.transformForLoopIteration(a.expressionCode, loopVar)
}

func (a *InlineLoopExpressionAccessor) GenerateInitialValueAccess(period int) string {
	return a.transformForLoopIteration(a.expressionCode, "0")
}

func (a *InlineLoopExpressionAccessor) transformForLoopIteration(expression, offset string) string {
	result := strings.ReplaceAll(expression, ".GetCurrent()", fmt.Sprintf(".Get(%s)", offset))

	if a.resolver != nil {
		result = a.transformLocalVariablesToSeriesAccess(result, offset)
	}

	return result
}

func (a *InlineLoopExpressionAccessor) transformLocalVariablesToSeriesAccess(code, offset string) string {
	if a.resolver == nil {
		return code
	}

	identifierPattern := regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*)\b`)

	result := identifierPattern.ReplaceAllStringFunc(code, func(match string) string {
		if strings.HasSuffix(match, "Series") {
			return match
		}

		if isGoKeywordOrBuiltin(match) {
			return match
		}

		if a.resolver.IsLocalVariable(match) {
			return fmt.Sprintf("%sSeries.Get(%s)", match, offset)
		}

		return match
	})

	return result
}

func isGoKeywordOrBuiltin(id string) bool {
	keywords := map[string]bool{
		"return": true, "if": true, "else": true, "for": true,
		"func": true, "true": true, "false": true, "nil": true,
		"math": true, "float64": true, "int": true, "bool": true,
		"string": true, "Get": true, "Set": true,
	}
	return keywords[id]
}

func (a *InlineLoopExpressionAccessor) GetPreamble() string {
	return ""
}
