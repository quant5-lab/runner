package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

type ControlFlowExpressionGenerator struct {
	baseGenerator *generator
}

func NewControlFlowExpressionGenerator(g *generator) *ControlFlowExpressionGenerator {
	return &ControlFlowExpressionGenerator{
		baseGenerator: g,
	}
}

func (c *ControlFlowExpressionGenerator) GenerateForExpressionAsIIFE(forStmt *ast.ForStatement) (string, error) {
	var builder strings.Builder

	builder.WriteString("(func() float64 {\n")
	c.baseGenerator.indent++

	builder.WriteString(c.baseGenerator.ind())
	builder.WriteString("var __result float64\n")

	counterVar := forStmt.Counter
	fromCode, err := c.baseGenerator.generateExpression(forStmt.From)
	if err != nil {
		return "", fmt.Errorf("generating for-expression from bound: %w", err)
	}
	fromCode = strings.TrimSpace(fromCode)

	toCode, err := c.baseGenerator.generateExpression(forStmt.To)
	if err != nil {
		return "", fmt.Errorf("generating for-expression to bound: %w", err)
	}
	toCode = strings.TrimSpace(toCode)

	stepCode := "1"
	if forStmt.Step != nil {
		stepCode, err = c.baseGenerator.generateExpression(forStmt.Step)
		if err != nil {
			return "", fmt.Errorf("generating for-expression step: %w", err)
		}
		stepCode = strings.TrimSpace(stepCode)
	}

	builder.WriteString(c.baseGenerator.ind())
	builder.WriteString(fmt.Sprintf("for %s := int(%s); %s <= int(%s); %s += int(%s) {\n",
		counterVar, fromCode, counterVar, toCode, counterVar, stepCode))

	c.baseGenerator.indent++

	lastStatementIsAssignment := false
	for i, bodyNode := range forStmt.Body {
		isLastStatement := i == len(forStmt.Body)-1

		if isLastStatement {
			if exprStmt, ok := bodyNode.(*ast.ExpressionStatement); ok {
				if ident, ok := exprStmt.Expression.(*ast.Identifier); ok {
					builder.WriteString(c.baseGenerator.ind())
					builder.WriteString(fmt.Sprintf("__result = float64(%s)\n", ident.Name))
					lastStatementIsAssignment = true
					continue
				}
			}
		}

		code, err := c.baseGenerator.generateStatement(bodyNode)
		if err != nil {
			return "", fmt.Errorf("generating for-expression body node: %w", err)
		}

		builder.WriteString(code)
	}

	if !lastStatementIsAssignment {
		builder.WriteString(c.baseGenerator.ind())
		builder.WriteString("__result = 0.0\n")
	}

	c.baseGenerator.indent--
	builder.WriteString(c.baseGenerator.ind())
	builder.WriteString("}\n")

	builder.WriteString(c.baseGenerator.ind())
	builder.WriteString("return __result\n")

	c.baseGenerator.indent--
	builder.WriteString(c.baseGenerator.ind())
	builder.WriteString("}())")

	return builder.String(), nil
}

func (c *ControlFlowExpressionGenerator) GenerateIfExpressionAsIIFE(ifStmt *ast.IfStatement) (string, error) {
	var builder strings.Builder

	builder.WriteString("(func() float64 {\n")
	c.baseGenerator.indent++

	condCode, err := c.baseGenerator.generateConditionExpression(ifStmt.Test)
	if err != nil {
		return "", fmt.Errorf("generating if-expression condition: %w", err)
	}

	builder.WriteString(c.baseGenerator.ind())
	builder.WriteString(fmt.Sprintf("if %s {\n", condCode))

	c.baseGenerator.indent++

	lastExprCode, err := c.extractLastExpressionFromBlock(ifStmt.Consequent)
	if err != nil {
		return "", fmt.Errorf("extracting if-expression consequent value: %w", err)
	}

	for _, node := range ifStmt.Consequent {
		code, err := c.baseGenerator.generateStatement(node)
		if err != nil {
			return "", fmt.Errorf("generating if-expression consequent node: %w", err)
		}
		builder.WriteString(code)
	}

	builder.WriteString(c.baseGenerator.ind())
	builder.WriteString(fmt.Sprintf("return %s\n", lastExprCode))

	c.baseGenerator.indent--
	builder.WriteString(c.baseGenerator.ind())
	builder.WriteString("}\n")

	if len(ifStmt.Alternate) > 0 {
		builder.WriteString(c.baseGenerator.ind())
		builder.WriteString("else {\n")

		c.baseGenerator.indent++

		altLastExprCode, err := c.extractLastExpressionFromBlock(ifStmt.Alternate)
		if err != nil {
			return "", fmt.Errorf("extracting if-expression alternate value: %w", err)
		}

		for _, node := range ifStmt.Alternate {
			code, err := c.baseGenerator.generateStatement(node)
			if err != nil {
				return "", fmt.Errorf("generating if-expression alternate node: %w", err)
			}
			builder.WriteString(code)
		}

		builder.WriteString(c.baseGenerator.ind())
		builder.WriteString(fmt.Sprintf("return %s\n", altLastExprCode))

		c.baseGenerator.indent--
		builder.WriteString(c.baseGenerator.ind())
		builder.WriteString("}\n")
	} else {
		builder.WriteString(c.baseGenerator.ind())
		builder.WriteString("return 0.0\n")
	}

	c.baseGenerator.indent--
	builder.WriteString(c.baseGenerator.ind())
	builder.WriteString("}())")

	return builder.String(), nil
}

func (c *ControlFlowExpressionGenerator) extractLastExpressionFromBlock(body []ast.Node) (string, error) {
	if len(body) == 0 {
		return "0.0", nil
	}

	lastNode := body[len(body)-1]

	if exprStmt, ok := lastNode.(*ast.ExpressionStatement); ok {
		return c.baseGenerator.generateExpression(exprStmt.Expression)
	}

	if varDecl, ok := lastNode.(*ast.VariableDeclaration); ok {
		if len(varDecl.Declarations) > 0 {
			if ident, ok := varDecl.Declarations[0].ID.(*ast.Identifier); ok {
				return ident.Name, nil
			}
		}
	}

	return "0.0", nil
}
