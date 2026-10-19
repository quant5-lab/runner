package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

// strategyCommentAnnotation is a comment parameter resolved from a strategy
// command: the Go string expression that carries it, and whether the Pine value
// was substituted because it is unsupported.
type strategyCommentAnnotation struct {
	parameter   string
	code        string
	substituted bool
}

// strategyAnnotationSlot is one comment parameter and the Pine expression bound
// to it, resolved by name first and by Pine position second so every consumer
// reads the same slot for the same call.
type strategyAnnotationSlot struct {
	parameter  string
	expression ast.Expression
}

// strategyAnnotationSlots resolves every present comment parameter of a strategy
// command in Pine parameter order. A named parameter wins over the same
// parameter written positionally.
func strategyAnnotationSlots(functionName string, args []ast.Expression) []strategyAnnotationSlot {
	call := newStrategyCallArguments(args)
	slots := make([]strategyAnnotationSlot, 0, len(strategyCommentParameters))
	for _, parameter := range strategyCommentParameters {
		expression, present := call.value(parameter, strategyAnnotationIndex(functionName, parameter))
		if !present {
			continue
		}
		slots = append(slots, strategyAnnotationSlot{parameter: parameter, expression: expression})
	}
	return slots
}

// commentAnnotations converts every present comment parameter into the Go string
// expression the call generator emits and reports whether the value was
// substituted.
func (e *ArgumentExtractor) commentAnnotations(functionName string, args []ast.Expression) []strategyCommentAnnotation {
	slots := strategyAnnotationSlots(functionName, args)
	annotations := make([]strategyCommentAnnotation, 0, len(slots))
	for _, slot := range slots {
		code, substituted := e.commentExpression(slot.expression)
		annotations = append(annotations, strategyCommentAnnotation{
			parameter:   slot.parameter,
			code:        code,
			substituted: substituted,
		})
	}
	return annotations
}

func primaryComment(annotations []strategyCommentAnnotation, def string) string {
	for _, annotation := range annotations {
		if annotation.parameter == "comment" {
			return annotation.code
		}
	}
	return def
}

// unsupportedCommentDiagnostics records every comment variant the strategy
// runtime cannot carry when its value is a feature-gap sentinel, so an
// unsupported annotation still surfaces instead of being dropped.
func unsupportedCommentDiagnostics(g *generator, annotations []strategyCommentAnnotation) string {
	var out strings.Builder
	for _, annotation := range annotations {
		if annotation.parameter == "comment" || !annotation.substituted {
			continue
		}
		out.WriteString(g.ind())
		out.WriteString("_ = ")
		out.WriteString(annotation.code)
		out.WriteString("\n")
	}
	return out.String()
}

// commentExpression converts a Pine comment value to a Go string expression and
// reports whether the value was substituted by a feature-gap sentinel.
func (e *ArgumentExtractor) commentExpression(expr ast.Expression) (string, bool) {
	switch v := expr.(type) {
	case *ast.Literal:
		if str, ok := v.Value.(string); ok {
			return fmt.Sprintf("%q", str), false
		}
	case *ast.Identifier:
		if e.generator != nil {
			name := SanitizeGoIdentifier(v.Name)
			if varType, ok := e.generator.variables[name]; ok {
				if varType == "string" {
					return name, false
				}
				return `""`, false
			}
		}
		return fmt.Sprintf("%sSeries.GetCurrent()", v.Name), false
	case *ast.ConditionalExpression:
		condition := strings.TrimRight(e.generator.extractSeriesExpression(v.Test), "\n")
		consequent, consequentSubstituted := e.commentExpression(v.Consequent)
		alternate, alternateSubstituted := e.commentExpression(v.Alternate)
		code := fmt.Sprintf("func() string { if (%s != 0) { return %s } else { return %s } }()", condition, consequent, alternate)
		return code, consequentSubstituted || alternateSubstituted
	case *ast.CallExpression:
		if e.generator != nil {
			dispatcher := NewExpressionPositionDispatcher(e.generator.callRouter)
			code, substituted, err := dispatcher.DispatchStringAnnotation(e.generator, v, "generator.comment_unknown")
			if err == nil {
				return code, substituted
			}
		}
	}
	return `""`, false
}
