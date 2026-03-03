package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// CorrelationHandler generates code for ta.correlation(source1, source2, length):
// Pearson r over rolling window of size length.
type CorrelationHandler struct{}

func (h *CorrelationHandler) CanHandle(funcName string) bool {
	return funcName == "ta.correlation" || funcName == "correlation"
}

func (h *CorrelationHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 3 {
		return "", fmt.Errorf("ta.correlation requires 3 arguments: source1, source2, length")
	}

	classifier := NewSeriesSourceClassifier()
	extractor := NewTAArgumentExtractor(g)

	src1Info := classifier.ClassifyAST(call.Arguments[0])
	src2Info := classifier.ClassifyAST(call.Arguments[1])
	src1AccessGen := CreateAccessGenerator(src1Info)
	src2AccessGen := CreateAccessGenerator(src2Info)

	periodResult := extractor.extractPeriodResult(call.Arguments[2], "ta.correlation")
	if periodResult.IsFailed() {
		return "", fmt.Errorf("ta.correlation: %s", periodResult.FailureReason)
	}
	if periodResult.IsRuntimeDynamic() {
		return "", fmt.Errorf("ta.correlation does not support runtime dynamic period")
	}

	period := periodResult.StaticValue
	baseOffset := src1AccessGen.GetBaseOffset()
	if src2AccessGen.GetBaseOffset() > baseOffset {
		baseOffset = src2AccessGen.GetBaseOffset()
	}
	warmup := period + baseOffset

	s1 := fmt.Sprintf("_%s_x", varName)
	s2 := fmt.Sprintf("_%s_y", varName)

	code := g.ind() + fmt.Sprintf("if ctx.BarIndex < %d {\n", warmup)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%s_sum1, %s_sum2 := 0.0, 0.0\n", varName, varName)
	code += g.ind() + fmt.Sprintf("for j := 0; j < %d; j++ { %s_sum1 += %s; %s_sum2 += %s }\n",
		period,
		varName, src1AccessGen.GenerateLoopValueAccess("j"),
		varName, src2AccessGen.GenerateLoopValueAccess("j"))
	code += g.ind() + fmt.Sprintf("%s := %s_sum1 / %d.0\n", s1, varName, period)
	code += g.ind() + fmt.Sprintf("%s := %s_sum2 / %d.0\n", s2, varName, period)
	code += g.ind() + fmt.Sprintf("%s_cov, %s_var1, %s_var2 := 0.0, 0.0, 0.0\n", varName, varName, varName)
	code += g.ind() + fmt.Sprintf("for j := 0; j < %d; j++ {\n", period)
	g.indent++
	code += g.ind() + fmt.Sprintf("%s_d1 := %s - %s\n", varName, src1AccessGen.GenerateLoopValueAccess("j"), s1)
	code += g.ind() + fmt.Sprintf("%s_d2 := %s - %s\n", varName, src2AccessGen.GenerateLoopValueAccess("j"), s2)
	code += g.ind() + fmt.Sprintf("%s_cov += %s_d1 * %s_d2\n", varName, varName, varName)
	code += g.ind() + fmt.Sprintf("%s_var1 += %s_d1 * %s_d1\n", varName, varName, varName)
	code += g.ind() + fmt.Sprintf("%s_var2 += %s_d2 * %s_d2\n", varName, varName, varName)
	g.indent--
	code += g.ind() + "}\n"
	code += g.ind() + fmt.Sprintf("if %s_var1 == 0 || %s_var2 == 0 {\n", varName, varName)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(0.0)\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(%s_cov / math.Sqrt(%s_var1*%s_var2))\n",
		varName, varName, varName, varName)
	g.indent--
	code += g.ind() + "}\n"
	g.indent--
	code += g.ind() + "}\n"

	return code, nil
}
