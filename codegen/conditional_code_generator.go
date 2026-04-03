package codegen

import (
	"fmt"
	"github.com/quant5-lab/runner/ast"
)

type ConditionalCodeGenerator struct {
	gen                    *generator
	conditionalArgAnalyzer *ConditionalArgumentAnalyzer
	tempVarMgr             *TempVariableManager
}

func NewConditionalCodeGenerator(
	g *generator,
	analyzer *ConditionalArgumentAnalyzer,
	tempVarMgr *TempVariableManager,
) *ConditionalCodeGenerator {
	return &ConditionalCodeGenerator{
		gen:                    g,
		conditionalArgAnalyzer: analyzer,
		tempVarMgr:             tempVarMgr,
	}
}

func (c *ConditionalCodeGenerator) GenerateSetCallsForExpression(expr ast.Expression, indent func() string) string {
	conditionals := c.conditionalArgAnalyzer.FindInExpression(expr)
	return c.generateSetCalls(conditionals, indent)
}

func (c *ConditionalCodeGenerator) GenerateSetCallsForStatement(stmt ast.Node, indent func() string) string {
	var expr ast.Expression

	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		expr = s.Expression
	case *ast.IfStatement:
		if s.Test != nil {
			expr = s.Test
		}
	default:
		return ""
	}

	if expr == nil {
		return ""
	}

	return c.GenerateSetCallsForExpression(expr, indent)
}

func (c *ConditionalCodeGenerator) generateSetCalls(conditionals []ConditionalArgumentInfo, indent func() string) string {
	var code string

	for _, condInfo := range conditionals {
		tempVarName := c.tempVarMgr.GetConditionalVarName(condInfo.ContentHash)
		if tempVarName == "" {
			continue
		}

		conditionalCode, err := c.gen.generateConditionalExpression(condInfo.Conditional)
		if err != nil {
			continue
		}

		code += fmt.Sprintf("%s%sSeries.Set(%s)\n", indent(), tempVarName, conditionalCode)
	}

	return code
}

func (c *ConditionalCodeGenerator) GetTempVarReference(cond *ast.ConditionalExpression, mode AccessMode) (string, bool) {
	hasher := &ExpressionHasher{}
	hash := hasher.Hash(cond)
	if hash != "" && len(hash) > 8 {
		hash = hash[:8]
	}

	tempVarName := c.tempVarMgr.GetConditionalVarName(hash)
	if tempVarName == "" {
		return "", false
	}

	switch mode {
	case AccessModeSeries:
		return tempVarName + "Series", true
	case AccessModeValue:
		return tempVarName + "Series.Get(0)", true
	case AccessModeCurrent:
		return tempVarName + "Series.GetCurrent()", true
	default:
		return tempVarName + "Series.Get(0)", true
	}
}

type AccessMode int

const (
	AccessModeValue AccessMode = iota
	AccessModeSeries
	AccessModeCurrent
)
