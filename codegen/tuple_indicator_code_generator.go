package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

/* TupleIndicatorCodeGenerator produces runtime-delegated code for tuple indicators */
type TupleIndicatorCodeGenerator struct {
	windowExtractor   *SeriesWindowExtractor
	argumentExtractor *TupleIndicatorArgumentExtractor
}

func NewTupleIndicatorCodeGenerator() *TupleIndicatorCodeGenerator {
	return &TupleIndicatorCodeGenerator{
		windowExtractor:   NewSeriesWindowExtractor(),
		argumentExtractor: NewTupleIndicatorArgumentExtractor(),
	}
}

func (g *TupleIndicatorCodeGenerator) Generate(
	spec *TupleIndicatorSpec,
	outputVars []string,
	callExpr *ast.CallExpression,
	ctx CodeGenContext,
	sourceExprExtractor func(ast.Expression) string,
	constants map[string]interface{},
) (string, error) {
	if len(outputVars) != spec.OutputCount {
		return "", errOutputCountMismatch(spec.FunctionName, spec.OutputCount, len(outputVars))
	}

	params, err := g.argumentExtractor.Extract(callExpr, sourceExprExtractor, constants)
	if err != nil {
		return "", err
	}

	out := &strings.Builder{}
	ind := ctx.Indenter

	warmupPeriod := g.calculateWarmupPeriod(params)

	out.WriteString(ind() + g.generateComment(spec, params))
	out.WriteString(ind() + fmt.Sprintf("if i < %d {\n", warmupPeriod))
	ctx.IncreaseIndent()
	out.WriteString(g.generateWarmupCode(outputVars, ind))
	ctx.DecreaseIndent()
	out.WriteString(ind() + "} else {\n")
	ctx.IncreaseIndent()

	out.WriteString(g.generateWindowExtraction(params.SourceExpr, ind))
	out.WriteString(g.generateRuntimeCall(spec, params, ind))
	out.WriteString(g.generateResultStorage(spec, outputVars, ind))

	ctx.DecreaseIndent()
	out.WriteString(ind() + "}\n")

	return out.String(), nil
}

func (g *TupleIndicatorCodeGenerator) calculateWarmupPeriod(params *TupleIndicatorArguments) int {
	maxPeriod := 0
	for _, p := range params.Periods {
		if p > maxPeriod {
			maxPeriod = p
		}
	}
	return maxPeriod - 1
}

func (g *TupleIndicatorCodeGenerator) generateComment(spec *TupleIndicatorSpec, params *TupleIndicatorArguments) string {
	periodStr := ""
	for i, p := range params.Periods {
		if i > 0 {
			periodStr += ","
		}
		periodStr += fmt.Sprintf("%d", p)
	}
	return fmt.Sprintf("/* Runtime %s(%s) */\n", spec.FunctionName, periodStr)
}

func (g *TupleIndicatorCodeGenerator) generateWarmupCode(outputVars []string, indenter func() string) string {
	out := &strings.Builder{}
	for _, varName := range outputVars {
		out.WriteString(indenter() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName))
	}
	return out.String()
}

func (g *TupleIndicatorCodeGenerator) generateWindowExtraction(sourceExpr string, indenter func() string) string {
	return g.windowExtractor.GenerateExtractionCode(sourceExpr, "i+1", indenter)
}

func (g *TupleIndicatorCodeGenerator) generateRuntimeCall(
	spec *TupleIndicatorSpec,
	params *TupleIndicatorArguments,
	indenter func() string,
) string {
	periodArgs := ""
	for _, p := range params.Periods {
		periodArgs += fmt.Sprintf(", %d", p)
	}

	outputVarList := g.buildOutputVarList(spec.OutputCount)

	return indenter() + fmt.Sprintf(
		"%s := %s(sourceWindow%s)\n\n",
		outputVarList,
		spec.RuntimeFunction,
		periodArgs,
	)
}

func (g *TupleIndicatorCodeGenerator) buildOutputVarList(count int) string {
	names := []string{}
	suffixes := []string{"Arr", "Arr2", "Arr3", "Arr4", "Arr5"}
	for i := 0; i < count; i++ {
		if i < len(suffixes) {
			names = append(names, "result"+suffixes[i])
		} else {
			names = append(names, fmt.Sprintf("result%dArr", i+1))
		}
	}
	return strings.Join(names, ", ")
}

func (g *TupleIndicatorCodeGenerator) generateResultStorage(
	spec *TupleIndicatorSpec,
	outputVars []string,
	indenter func() string,
) string {
	out := &strings.Builder{}
	suffixes := []string{"Arr", "Arr2", "Arr3", "Arr4", "Arr5"}

	for i, varName := range outputVars {
		arrName := "resultArr"
		if i < len(suffixes) {
			arrName = "result" + suffixes[i]
		} else {
			arrName = fmt.Sprintf("result%dArr", i+1)
		}
		out.WriteString(indenter() + fmt.Sprintf(
			"%sSeries.Set(%s[len(%s)-1])\n",
			varName,
			arrName,
			arrName,
		))
	}

	return out.String()
}
