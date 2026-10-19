package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/*
TimeHandler routes Pine's time() function to TimeCodeGenerator. The Pine
version is propagated into the emitted session.TimeFuncWithVersion call so
v4 strategies receive the Mon-Fri default DAYS mask while v5 strategies
receive the all-7-days default — matching the v4→v5 sessions breaking
change documented at:

	https://www.tradingview.com/pine-script-docs/v5/concepts/sessions/
*/
type TimeHandler struct {
	parser      *SessionArgumentParser
	generator   *TimeCodeGenerator
	tfExtractor *TimeframeArgExtractor
	pineVersion int
}

func NewTimeHandler(indentation string) *TimeHandler {
	// Back-compat: defaults to v5 semantics for direct unit-test callers
	// that do not have a generator at hand.
	return NewTimeHandlerWithVersion(indentation, 5)
}

func NewTimeHandlerWithVersion(indentation string, pineVersion int) *TimeHandler {
	return &TimeHandler{
		parser:      NewSessionArgumentParser(),
		generator:   NewTimeCodeGeneratorWithVersion(indentation, pineVersion),
		tfExtractor: NewTimeframeArgExtractor(),
		pineVersion: pineVersion,
	}
}

func (th *TimeHandler) CanHandle(funcName string) bool {
	return funcName == "time"
}

/*
GenerateInline implements InlineConditionHandler interface. Uses the
generator's pineVersion when invoked through the inline-condition path
so the registry-level instance picks up the script's version.
*/
func (th *TimeHandler) GenerateInline(expr *ast.CallExpression, g *generator) (string, error) {
	v := th.pineVersion
	if g != nil && g.pineVersion != 0 {
		v = g.pineVersion
	}
	return th.handleInlineExpressionWithVersion(expr.Arguments, v), nil
}

func (h *TimeHandler) HandleVariableInit(varName string, call *ast.CallExpression) string {
	switch len(call.Arguments) {
	case 0:
		return h.generator.GenerateNoArguments(varName)
	case 1:
		tfGoExpr := h.tfExtractor.GoExpr(call.Arguments[0])
		return h.generator.GenerateSingleArgument(varName, tfGoExpr)
	default:
		sessionArg := call.Arguments[1]
		session := h.parser.Parse(sessionArg)
		return h.generator.GenerateWithSession(varName, session)
	}
}

func (h *TimeHandler) HandleInlineExpression(args []ast.Expression) string {
	return h.handleInlineExpressionWithVersion(args, h.pineVersion)
}

func (h *TimeHandler) handleInlineExpressionWithVersion(args []ast.Expression, pineVersion int) string {
	switch len(args) {
	case 0:
		return barTimestampMsExpr
	case 1:
		return alignedBarTimeMsExpr(h.tfExtractor.GoExpr(args[0]))
	default:
		return h.generateInlineWithSession(args[1], pineVersion)
	}
}

func (h *TimeHandler) generateInlineWithSession(sessionArg ast.Expression, pineVersion int) string {
	session := h.parser.Parse(sessionArg)

	if !session.IsValid() {
		return "math.NaN()"
	}

	if session.IsLiteral() {
		return h.generateInlineLiteral(session.Value, pineVersion)
	}

	return h.generateInlineVariable(session.Value, pineVersion)
}

func (h *TimeHandler) generateInlineLiteral(sessionValue string, pineVersion int) string {
	return fmt.Sprintf(
		"session.TimeFuncWithVersion(ctx.Data[ctx.BarIndex].Time*1000, ctx.Timeframe, %q, ctx.Timezone, %d)",
		sessionValue, pineVersion)
}

func (h *TimeHandler) generateInlineVariable(sessionValue string, pineVersion int) string {
	return fmt.Sprintf(
		"session.TimeFuncWithVersion(ctx.Data[ctx.BarIndex].Time*1000, ctx.Timeframe, %s, ctx.Timezone, %d)",
		sessionValue, pineVersion)
}
