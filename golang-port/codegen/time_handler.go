package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

type TimeHandler struct {
	parser    *SessionArgumentParser
	generator *TimeCodeGenerator
}

func NewTimeHandler(indentation string) *TimeHandler {
	return &TimeHandler{
		parser:    NewSessionArgumentParser(),
		generator: NewTimeCodeGenerator(indentation),
	}
}

func (h *TimeHandler) HandleVariableInit(varName string, call *ast.CallExpression) string {
	argCount := len(call.Arguments)

	if argCount == 0 {
		return h.generator.GenerateNoArguments(varName)
	}

	if argCount == 1 {
		return h.generator.GenerateSingleArgument(varName)
	}

	sessionArg := call.Arguments[1]
	session := h.parser.Parse(sessionArg)

	return h.generator.GenerateWithSession(varName, session)
}

func (h *TimeHandler) HandleInlineExpression(args []ast.Expression) string {
	if len(args) < 2 {
		return "float64(ctx.Data[ctx.BarIndex].Time)"
	}

	sessionArg := args[1]
	session := h.parser.Parse(sessionArg)

	if !session.IsValid() {
		return "math.NaN()"
	}

	if session.IsLiteral() {
		return h.generateInlineLiteral(session.Value)
	}

	return h.generateInlineVariable(session.Value)
}

func (h *TimeHandler) generateInlineLiteral(sessionValue string) string {
	return "session.TimeFunc(ctx.Data[ctx.BarIndex].Time*1000, ctx.Timeframe, \"" + sessionValue + "\", ctx.Timezone)"
}

func (h *TimeHandler) generateInlineVariable(sessionValue string) string {
	return "session.TimeFunc(ctx.Data[ctx.BarIndex].Time*1000, ctx.Timeframe, " + sessionValue + ", ctx.Timezone)"
}
