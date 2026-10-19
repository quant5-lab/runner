package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type StochSingleValueHandler struct{}

func (h *StochSingleValueHandler) CanHandle(funcName string) bool {
	return funcName == "ta.stoch" || funcName == "stoch"
}

func (h *StochSingleValueHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 4 {
		return "", fmt.Errorf("stoch requires 4 arguments: source, high, low, length")
	}

	period, err := extractPeriodFromArgAt(g, call.Arguments[3], "stoch")
	if err != nil {
		return "", err
	}

	sourceExpr := g.extractSeriesExpression(call.Arguments[0])
	highAcc := newStochWindowAccessor(call.Arguments[1], "High")
	lowAcc := newStochWindowAccessor(call.Arguments[2], "Low")

	hh := fmt.Sprintf("_%s_hh", varName)
	ll := fmt.Sprintf("_%s_ll", varName)
	denom := fmt.Sprintf("_%s_denom", varName)

	code := g.ind() + fmt.Sprintf("if ctx.BarIndex < %d {\n", period-1)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++

	code += g.ind() + fmt.Sprintf("%s := %s\n", hh, highAcc.at("0"))
	code += g.ind() + fmt.Sprintf("%s := %s\n", ll, lowAcc.at("0"))
	code += g.ind() + fmt.Sprintf("for j := 1; j < %d; j++ {\n", period)
	g.indent++
	code += g.ind() + fmt.Sprintf("if %s > %s { %s = %s }\n", highAcc.at("j"), hh, hh, highAcc.at("j"))
	code += g.ind() + fmt.Sprintf("if %s < %s { %s = %s }\n", lowAcc.at("j"), ll, ll, lowAcc.at("j"))
	g.indent--
	code += g.ind() + "}\n"

	code += g.ind() + fmt.Sprintf("%s := %s - %s\n", denom, hh, ll)
	code += g.ind() + fmt.Sprintf("if %s == 0.0 {\n", denom)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(0.0)\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set((%s - %s) / %s * 100.0)\n", varName, sourceExpr, ll, denom)
	g.indent--
	code += g.ind() + "}\n"

	g.indent--
	code += g.ind() + "}\n"

	return code, nil
}

type stochWindowAccessor struct {
	template string
}

func newStochWindowAccessor(arg ast.Expression, defaultField string) stochWindowAccessor {
	if ident, ok := arg.(*ast.Identifier); ok {
		if field, isOHLCV := OHLCVFieldName(ident.Name); isOHLCV {
			return stochWindowAccessor{template: "ctx.Data[ctx.BarIndex-%s]." + field}
		}
		return stochWindowAccessor{template: ident.Name + "Series.Get(%s)"}
	}
	return stochWindowAccessor{template: "ctx.Data[ctx.BarIndex-%s]." + defaultField}
}

func (a stochWindowAccessor) at(j string) string {
	return fmt.Sprintf(a.template, j)
}

func extractPeriodFromArgAt(g *generator, arg ast.Expression, funcName string) (int, error) {
	synthetic := &ast.CallExpression{Arguments: []ast.Expression{arg}}
	return extractSinglePeriodArgument(g, synthetic, funcName)
}
