package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type CalendarHandler struct {
	functions map[string]string
}

func NewCalendarHandler() *CalendarHandler {
	return &CalendarHandler{
		functions: map[string]string{
			"year":       "calendar.Year",
			"month":      "calendar.Month",
			"dayofweek":  "calendar.DayOfWeek",
			"dayofmonth": "calendar.DayOfMonth",
			"hour":       "calendar.Hour",
			"minute":     "calendar.Minute",
			"second":     "calendar.Second",
			"weekofyear": "calendar.WeekOfYear",
		},
	}
}

func (h *CalendarHandler) CanHandle(funcName string) bool {
	if _, ok := h.functions[funcName]; ok {
		return true
	}
	return funcName == "timestamp"
}

func (h *CalendarHandler) GenerateCalendarCall(funcName string, args []ast.Expression, g *generator) (string, error) {
	if funcName == "timestamp" {
		return h.generateTimestamp(args, g)
	}

	goFunc, ok := h.functions[funcName]
	if !ok {
		return "", fmt.Errorf("unsupported calendar function: %s", funcName)
	}

	return h.generateExtraction(goFunc, args, g)
}

func (h *CalendarHandler) generateExtraction(goFunc string, args []ast.Expression, g *generator) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("%s() requires at least 1 argument (timestamp)", goFunc)
	}

	tsExpr := g.extractSeriesExpression(args[0])
	tz := "ctx.Timezone"

	if len(args) >= 2 {
		tzArg, err := g.generateConditionExpression(args[1])
		if err != nil {
			return "", fmt.Errorf("timezone argument: %w", err)
		}
		tz = tzArg
	}

	return fmt.Sprintf("%s(%s, %s)", goFunc, tsExpr, tz), nil
}

func (h *CalendarHandler) generateTimestamp(args []ast.Expression, g *generator) (string, error) {
	switch len(args) {
	case 1:
		return h.generateTimestampFromString(args, g)
	case 5:
		return h.generateTimestampFromComponents(args, g, "ctx.Timezone")
	case 6:
		if h.isTimezoneArg(args[0]) {
			tzExpr, err := g.generateConditionExpression(args[0])
			if err != nil {
				return "", fmt.Errorf("timestamp timezone arg: %w", err)
			}
			return h.generateTimestampFromComponents(args[1:], g, tzExpr)
		}
		return h.generateTimestampFromComponents(args, g, "ctx.Timezone")
	case 7:
		tzExpr, err := g.generateConditionExpression(args[0])
		if err != nil {
			return "", fmt.Errorf("timestamp timezone arg: %w", err)
		}
		return h.generateTimestampFromComponents(args[1:], g, tzExpr)
	default:
		return "", fmt.Errorf("timestamp() requires 1, 5, 6, or 7 arguments, got %d", len(args))
	}
}

/* Literal string args route to TimestampFromString; variable args pass through as-is */
func (h *CalendarHandler) generateTimestampFromString(args []ast.Expression, g *generator) (string, error) {
	argCode, err := g.generateConditionExpression(args[0])
	if err != nil {
		return "", fmt.Errorf("timestamp string arg: %w", err)
	}

	if _, ok := args[0].(*ast.Literal); ok {
		return fmt.Sprintf("calendar.TimestampFromString(%s, ctx.Timezone)", argCode), nil
	}

	return argCode, nil
}

func (h *CalendarHandler) generateTimestampFromComponents(args []ast.Expression, g *generator, tz string) (string, error) {
	if len(args) < 5 || len(args) > 6 {
		return "", fmt.Errorf("timestamp components require 5 or 6 arguments, got %d", len(args))
	}

	parts := make([]string, len(args))
	for i, arg := range args {
		code, err := g.generateConditionExpression(arg)
		if err != nil {
			return "", fmt.Errorf("timestamp arg %d: %w", i, err)
		}
		parts[i] = code
	}

	second := "0"
	if len(args) == 6 {
		second = parts[5]
	}

	return fmt.Sprintf("calendar.Timestamp(%s, %s, %s, %s, %s, %s, %s)",
		parts[0], parts[1], parts[2], parts[3], parts[4], second, tz), nil
}

/* 6-arg ambiguity: string literal or syminfo.timezone → timezone, otherwise → year */
func (h *CalendarHandler) isTimezoneArg(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.Literal:
		_, isStr := e.Value.(string)
		return isStr
	case *ast.MemberExpression:
		obj, okObj := e.Object.(*ast.Identifier)
		prop, okProp := e.Property.(*ast.Identifier)
		return okObj && okProp && obj.Name == "syminfo" && prop.Name == "timezone"
	default:
		return false
	}
}
