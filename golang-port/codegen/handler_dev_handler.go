package codegen

import "github.com/quant5-lab/runner/ast"

/* DEVHandler generates inline code for Mean Absolute Deviation calculations */
type DEVHandler struct{}

func (h *DEVHandler) CanHandle(funcName string) bool {
	return funcName == "ta.dev" || funcName == "dev"
}

func (h *DEVHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	sourceASTExpr, period, err := extractTAArgumentsAST(g, call, "ta.dev")
	if err != nil {
		return "", err
	}

	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.ClassifyAST(sourceASTExpr)
	accessGen := CreateAccessGenerator(sourceInfo)
	needsNaN := sourceInfo.IsSeriesVariable()

	builder := NewTAIndicatorBuilder("ta.dev", varName, period, accessGen, needsNaN)
	return g.indentCode(builder.BuildDEV()), nil
}
