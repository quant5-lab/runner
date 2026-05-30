package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

/* PineScript na: returned when no branch executes in expression context */
const pineNaExpression = "math.NaN()"

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
	if c.baseGenerator.loopContextStack != nil {
		c.baseGenerator.loopContextStack.Push(counterVar)
		defer c.baseGenerator.loopContextStack.Pop()
	}

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

	if err := c.generateLoopBodyWithResult(&builder, forStmt.Body); err != nil {
		return "", err
	}

	return builder.String(), nil
}

func (c *ControlFlowExpressionGenerator) GenerateForInExpressionAsIIFE(forIn *ast.ForInStatement) (string, error) {
	var builder strings.Builder

	builder.WriteString("(func() float64 {\n")
	c.baseGenerator.indent++

	builder.WriteString(c.baseGenerator.ind())
	builder.WriteString("var __result float64\n")

	collCode, err := c.baseGenerator.generateExpression(forIn.Collection)
	if err != nil {
		return "", fmt.Errorf("generating for-in collection expression: %w", err)
	}
	collCode = strings.TrimSpace(collCode)

	indexVar := "_"
	if forIn.IndexVar != "" {
		indexVar = forIn.IndexVar
	}

	builder.WriteString(c.baseGenerator.ind())
	builder.WriteString(fmt.Sprintf("for %s, %s := range %s {\n", indexVar, forIn.ElementVar, collCode))
	c.baseGenerator.indent++

	if err := c.generateLoopBodyWithResult(&builder, forIn.Body); err != nil {
		return "", err
	}

	return builder.String(), nil
}

func (c *ControlFlowExpressionGenerator) GenerateWhileExpressionAsIIFE(whileStmt *ast.WhileStatement) (string, error) {
	var builder strings.Builder

	builder.WriteString("(func() float64 {\n")
	c.baseGenerator.indent++

	builder.WriteString(c.baseGenerator.ind())
	builder.WriteString("var __result float64\n")

	condCode, err := c.baseGenerator.generateConditionExpression(whileStmt.Condition)
	if err != nil {
		return "", fmt.Errorf("generating while-expression condition: %w", err)
	}
	condCode = strings.TrimSpace(condCode)
	condCode = c.baseGenerator.addBoolConversionIfNeeded(whileStmt.Condition, condCode)

	guard := NewLoopIterationGuard()

	builder.WriteString(guard.InitCode(c.baseGenerator.ind()))
	builder.WriteString(c.baseGenerator.ind())
	builder.WriteString(fmt.Sprintf("for %s {\n", condCode))
	c.baseGenerator.indent++

	builder.WriteString(guard.CheckCode(c.baseGenerator.ind()))

	if err := c.generateLoopBodyWithResult(&builder, whileStmt.Body); err != nil {
		return "", err
	}

	return builder.String(), nil
}

/* Shared by for, for-in, and while IIFE: last ExpressionStatement becomes __result assignment */
func (c *ControlFlowExpressionGenerator) generateLoopBodyWithResult(builder *strings.Builder, body []ast.Node) error {
	lastStatementIsAssignment := false
	for i, bodyNode := range body {
		isLastStatement := i == len(body)-1

		if isLastStatement {
			if exprStmt, ok := bodyNode.(*ast.ExpressionStatement); ok {
				exprCode, err := c.baseGenerator.generateArrowFunctionExpression(exprStmt.Expression)
				if err != nil {
					return fmt.Errorf("generating loop result expression: %w", err)
				}
				exprCode = strings.TrimSpace(exprCode)
				builder.WriteString(c.baseGenerator.ind())
				builder.WriteString(fmt.Sprintf("__result = float64(%s)\n", exprCode))
				lastStatementIsAssignment = true
				continue
			}
		}

		code, err := c.baseGenerator.generateStatement(bodyNode)
		if err != nil {
			return fmt.Errorf("generating loop expression body: %w", err)
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

	return nil
}

// GenerateIfExpressionAsTypedIIFE is like GenerateIfExpressionAsIIFE but allows
// specifying the return type (e.g. "string") for non-float64 branch results.
func (c *ControlFlowExpressionGenerator) GenerateIfExpressionAsTypedIIFE(ifStmt *ast.IfStatement, returnType string) (string, error) {
	if returnType == "" || returnType == "float64" {
		return c.GenerateIfExpressionAsIIFE(ifStmt)
	}
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("(func() %s {\n", returnType))
	c.baseGenerator.indent++
	if err := c.generateIfBranchWithReturnTyped(&builder, ifStmt, c.baseGenerator.ind(), returnType); err != nil {
		return "", err
	}
	c.baseGenerator.indent--
	builder.WriteString(c.baseGenerator.ind())
	builder.WriteString("}())")
	return builder.String(), nil
}

func (c *ControlFlowExpressionGenerator) GenerateIfExpressionAsIIFE(ifStmt *ast.IfStatement) (string, error) {
	var builder strings.Builder

	builder.WriteString("(func() float64 {\n")
	c.baseGenerator.indent++

	if err := c.generateIfBranchWithReturn(&builder, ifStmt, c.baseGenerator.ind()); err != nil {
		return "", err
	}

	c.baseGenerator.indent--
	builder.WriteString(c.baseGenerator.ind())
	builder.WriteString("}())")

	return builder.String(), nil
}

func (c *ControlFlowExpressionGenerator) generateIfBranchWithReturnTyped(builder *strings.Builder, ifStmt *ast.IfStatement, prefix string, returnType string) error {
	condCode, err := c.baseGenerator.generateConditionExpression(ifStmt.Test)
	if err != nil {
		return fmt.Errorf("generating if-expression condition: %w", err)
	}
	condCode = c.baseGenerator.addBoolConversionIfNeeded(ifStmt.Test, condCode)

	builder.WriteString(prefix)
	builder.WriteString(fmt.Sprintf("if %s {\n", condCode))
	c.baseGenerator.indent++

	if err := c.generateBodyWithReturn(builder, ifStmt.Consequent); err != nil {
		return err
	}

	c.baseGenerator.indent--

	return c.generateIIFEAlternateTyped(builder, ifStmt.Alternate, returnType)
}

func (c *ControlFlowExpressionGenerator) generateIfBranchWithReturn(builder *strings.Builder, ifStmt *ast.IfStatement, prefix string) error {
	return c.generateIfBranchWithReturnTyped(builder, ifStmt, prefix, "float64")
}

func (c *ControlFlowExpressionGenerator) iifeFallback(returnType string) string {
	if returnType == "string" {
		return `""`
	}
	return pineNaExpression
}

func (c *ControlFlowExpressionGenerator) generateIIFEAlternateTyped(builder *strings.Builder, alternate []ast.Node, returnType string) error {
	if len(alternate) == 0 {
		builder.WriteString(c.baseGenerator.ind())
		builder.WriteString("}\n")
		builder.WriteString(c.baseGenerator.ind())
		builder.WriteString("return " + c.iifeFallback(returnType) + "\n")
		return nil
	}

	if nestedIf, isChain := extractElseIfChain(alternate); isChain {
		prefix := c.baseGenerator.ind() + "} else "
		return c.generateIfBranchWithReturnTyped(builder, nestedIf, prefix, returnType)
	}

	builder.WriteString(c.baseGenerator.ind())
	builder.WriteString("} else {\n")
	c.baseGenerator.indent++

	if err := c.generateBodyWithReturn(builder, alternate); err != nil {
		return err
	}

	c.baseGenerator.indent--
	builder.WriteString(c.baseGenerator.ind())
	builder.WriteString("}\n")
	return nil
}

func (c *ControlFlowExpressionGenerator) generateIIFEAlternate(builder *strings.Builder, alternate []ast.Node) error {
	return c.generateIIFEAlternateTyped(builder, alternate, "float64")
}

func (c *ControlFlowExpressionGenerator) generateBodyWithReturn(builder *strings.Builder, body []ast.Node) error {
	/* IIFE bodies need arrow-function-style expression handling (e.g. binary expressions as values) */
	wasInArrow := c.baseGenerator.inArrowFunctionBody
	c.baseGenerator.inArrowFunctionBody = true
	defer func() { c.baseGenerator.inArrowFunctionBody = wasInArrow }()

	/* Snapshot outer scope to detect IIFE-local variables */
	outerVars := make(map[string]bool, len(c.baseGenerator.variables))
	for k := range c.baseGenerator.variables {
		outerVars[k] = true
	}

	/* Activate IIFE-local var tracking, nested IIFEs get a fresh scope */
	prevIifeLocalVars := c.baseGenerator.blockLocalVars
	c.baseGenerator.blockLocalVars = make(map[string]bool)
	defer func() {
		/* Remove IIFE-local vars from g.variables so they don't pollute outer scope */
		for k := range c.baseGenerator.blockLocalVars {
			if !outerVars[k] {
				delete(c.baseGenerator.variables, k)
			}
		}
		c.baseGenerator.blockLocalVars = prevIifeLocalVars
	}()

	/* Last ExpressionStatement = return value; all preceding nodes emit normally */
	lastIdx := len(body) - 1
	for i, node := range body {
		if i == lastIdx {
			if _, isExpr := node.(*ast.ExpressionStatement); isExpr {
				break
			}
		}

		/* Variable declarations for vars not in outer scope become Go locals */
		if code, ok := c.tryGenerateIIFELocalDeclaration(node, outerVars); ok {
			builder.WriteString(code)
			continue
		}

		code, err := c.baseGenerator.generateStatement(node)
		if err != nil {
			return fmt.Errorf("generating expression body node: %w", err)
		}
		builder.WriteString(code)
	}

	lastExprCode, err := c.extractLastExpressionFromBlock(body)
	if err != nil {
		return fmt.Errorf("extracting return value: %w", err)
	}

	builder.WriteString(c.baseGenerator.ind())
	builder.WriteString(fmt.Sprintf("return %s\n", lastExprCode))
	return nil
}

/* tryGenerateIIFELocalDeclaration emits a Go-local `:=` declaration for variables first
 * introduced inside an IIFE body. Returns ("", false) if the node is not an IIFE-local decl. */
func (c *ControlFlowExpressionGenerator) tryGenerateIIFELocalDeclaration(node ast.Node, outerVars map[string]bool) (string, bool) {
	varDecl, ok := node.(*ast.VariableDeclaration)
	if !ok || len(varDecl.Declarations) == 0 {
		return "", false
	}
	decl := varDecl.Declarations[0]
	ident, ok := decl.ID.(*ast.Identifier)
	if !ok || outerVars[ident.Name] {
		return "", false
	}
	if decl.Init == nil {
		return "", false
	}

	exprCode := c.baseGenerator.extractSeriesExpression(decl.Init)
	c.baseGenerator.blockLocalVars[ident.Name] = true
	/* Register with special type so identifier resolvers return plain name */
	c.baseGenerator.variables[ident.Name] = "iife_local"

	return c.baseGenerator.ind() + fmt.Sprintf("%s := %s\n", ident.Name, exprCode), true
}

func (c *ControlFlowExpressionGenerator) extractLastExpressionFromBlock(body []ast.Node) (string, error) {
	if len(body) == 0 {
		return pineNaExpression, nil
	}

	lastNode := body[len(body)-1]

	if exprStmt, ok := lastNode.(*ast.ExpressionStatement); ok {
		return c.baseGenerator.generateArrowFunctionExpression(exprStmt.Expression)
	}

	if varDecl, ok := lastNode.(*ast.VariableDeclaration); ok {
		if len(varDecl.Declarations) > 0 {
			if ident, ok := varDecl.Declarations[0].ID.(*ast.Identifier); ok {
				return ident.Name, nil
			}
		}
	}

	return pineNaExpression, nil
}
