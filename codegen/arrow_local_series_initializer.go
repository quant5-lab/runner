package codegen

import (
	"fmt"
	"sort"

	"github.com/quant5-lab/runner/ast"
)

/* ArrowLocalSeriesInitializer generates Series initialization code for arrow function local variables */
type ArrowLocalSeriesInitializer struct {
	analyzer    *LocalSeriesAnalyzer
	indentation string
}

func NewArrowLocalSeriesInitializer(indent string) *ArrowLocalSeriesInitializer {
	return &ArrowLocalSeriesInitializer{
		analyzer:    NewLocalSeriesAnalyzer(),
		indentation: indent,
	}
}

func (i *ArrowLocalSeriesInitializer) GenerateInitializations(arrowFunc *ast.ArrowFunctionExpression) string {
	needsSeries := i.analyzer.Analyze(arrowFunc)

	if len(needsSeries) == 0 {
		return ""
	}

	varNames := i.sortVariableNames(needsSeries)

	code := ""
	for _, varName := range varNames {
		code += i.indentation + fmt.Sprintf("%sSeries := arrowCtx.GetOrCreateSeries(%q)\n", varName, varName)
	}
	code += "\n"

	return code
}

func (i *ArrowLocalSeriesInitializer) sortVariableNames(varMap map[string]bool) []string {
	names := make([]string, 0, len(varMap))
	for name := range varMap {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
