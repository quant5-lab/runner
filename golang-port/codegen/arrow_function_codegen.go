package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

type ArrowFunctionCodegen struct {
	gen *generator
}

func NewArrowFunctionCodegen(gen *generator) *ArrowFunctionCodegen {
	return &ArrowFunctionCodegen{gen: gen}
}

func (a *ArrowFunctionCodegen) Generate(funcName string, arrowFunc *ast.ArrowFunctionExpression) (string, error) {
	signature, returnType, err := a.analyzeAndGenerateSignature(funcName, arrowFunc)
	if err != nil {
		return "", err
	}

	body, err := a.generateFunctionBody(arrowFunc)
	if err != nil {
		return "", err
	}

	code := a.gen.ind() + signature + " " + returnType + " {\n"
	a.gen.indent++
	code += body
	a.gen.indent--
	code += a.gen.ind() + "}\n\n"

	return code, nil
}

func (a *ArrowFunctionCodegen) analyzeAndGenerateSignature(funcName string, arrowFunc *ast.ArrowFunctionExpression) (string, string, error) {
	params := a.buildParameterList(arrowFunc.Params)
	returnType, err := a.inferReturnType(arrowFunc)
	if err != nil {
		return "", "", err
	}

	signature := fmt.Sprintf("func %s(ctx *Context%s)", funcName, params)
	return signature, returnType, nil
}

func (a *ArrowFunctionCodegen) buildParameterList(params []ast.Identifier) string {
	if len(params) == 0 {
		return ""
	}

	var parts []string
	for _, param := range params {
		parts = append(parts, fmt.Sprintf("%s float64", param.Name))
	}

	return ", " + strings.Join(parts, ", ")
}

func (a *ArrowFunctionCodegen) inferReturnType(arrowFunc *ast.ArrowFunctionExpression) (string, error) {
	if len(arrowFunc.Body) == 0 {
		return "", fmt.Errorf("arrow function has empty body")
	}

	lastStmt := arrowFunc.Body[len(arrowFunc.Body)-1]

	switch stmt := lastStmt.(type) {
	case *ast.VariableDeclaration:
		if len(stmt.Declarations) > 0 {
			if arrayPattern, ok := stmt.Declarations[0].ID.(*ast.ArrayPattern); ok {
				return a.buildTupleReturnType(len(arrayPattern.Elements)), nil
			}
		}
		return "float64", nil

	case *ast.ExpressionStatement:
		if literal, ok := stmt.Expression.(*ast.Literal); ok {
			if elemSlice, ok := literal.Value.([]ast.Expression); ok {
				return a.buildTupleReturnType(len(elemSlice)), nil
			}
		}
		return "float64", nil

	default:
		return "float64", nil
	}
}

func (a *ArrowFunctionCodegen) buildTupleReturnType(count int) string {
	if count == 1 {
		return "float64"
	}

	parts := make([]string, count)
	for i := range parts {
		parts[i] = "float64"
	}

	return "(" + strings.Join(parts, ", ") + ")"
}

func (a *ArrowFunctionCodegen) generateFunctionBody(arrowFunc *ast.ArrowFunctionExpression) (string, error) {
	if len(arrowFunc.Body) == 0 {
		return "", fmt.Errorf("arrow function has empty body")
	}

	// Register function parameters as variables (runtime values, not constants)
	savedVariables := make(map[string]string)
	for _, param := range arrowFunc.Params {
		if existingType, exists := a.gen.variables[param.Name]; exists {
			savedVariables[param.Name] = existingType
		}
		a.gen.variables[param.Name] = "float"
	}

	// Mark that we're inside arrow function body (affects TA generation)
	wasInArrowFunction := a.gen.inArrowFunctionBody
	a.gen.inArrowFunctionBody = true

	// Restore state after body generation
	defer func() {
		a.gen.inArrowFunctionBody = wasInArrowFunction
		for _, param := range arrowFunc.Params {
			if savedType, wasSaved := savedVariables[param.Name]; wasSaved {
				a.gen.variables[param.Name] = savedType
			} else {
				delete(a.gen.variables, param.Name)
			}
		}
	}()

	lastIdx := len(arrowFunc.Body) - 1
	var bodyCode string

	for i, stmt := range arrowFunc.Body {
		if i == lastIdx {
			returnCode, err := a.generateFinalReturnStatement(stmt)
			if err != nil {
				return "", err
			}
			bodyCode += returnCode
			break
		}

		stmtCode, err := a.gen.generateStatement(stmt)
		if err != nil {
			return "", fmt.Errorf("failed to generate statement: %w", err)
		}
		bodyCode += stmtCode
	}

	return bodyCode, nil
}

func (a *ArrowFunctionCodegen) generateFinalReturnStatement(lastStmt ast.Node) (string, error) {
	switch stmt := lastStmt.(type) {
	case *ast.VariableDeclaration:
		return a.generateVariableReturnStatement(stmt)

	case *ast.ExpressionStatement:
		return a.generateExpressionReturnStatement(stmt)

	default:
		return "", fmt.Errorf("unsupported last statement type in arrow function: %T", lastStmt)
	}
}

func (a *ArrowFunctionCodegen) generateVariableReturnStatement(varDecl *ast.VariableDeclaration) (string, error) {
	if len(varDecl.Declarations) == 0 {
		return "", fmt.Errorf("empty variable declaration")
	}

	decl := varDecl.Declarations[0]

	if arrayPattern, ok := decl.ID.(*ast.ArrayPattern); ok {
		return a.generateTupleReturn(arrayPattern, decl.Init)
	}

	if id, ok := decl.ID.(*ast.Identifier); ok {
		stmtCode, err := a.gen.generateStatement(varDecl)
		if err != nil {
			return "", err
		}
		return stmtCode + a.gen.ind() + "return " + id.Name + "\n", nil
	}

	return "", fmt.Errorf("unsupported variable declarator pattern: %T", decl.ID)
}

func (a *ArrowFunctionCodegen) generateTupleReturn(arrayPattern *ast.ArrayPattern, init ast.Expression) (string, error) {
	if len(arrayPattern.Elements) == 0 {
		return "", fmt.Errorf("empty tuple pattern")
	}

	var returnVars []string
	for _, elem := range arrayPattern.Elements {
		returnVars = append(returnVars, elem.Name)
	}

	initCode, err := a.generateTupleInitExpression(init, returnVars)
	if err != nil {
		return "", err
	}

	code := initCode
	code += a.gen.ind() + "return " + strings.Join(returnVars, ", ") + "\n"

	return code, nil
}

func (a *ArrowFunctionCodegen) generateTupleInitExpression(expr ast.Expression, varNames []string) (string, error) {
	exprCode, err := a.generateExpression(expr)
	if err != nil {
		return "", err
	}

	return a.gen.ind() + strings.Join(varNames, ", ") + " := " + exprCode + "\n", nil
}

func (a *ArrowFunctionCodegen) generateExpressionReturnStatement(exprStmt *ast.ExpressionStatement) (string, error) {
	if literal, ok := exprStmt.Expression.(*ast.Literal); ok {
		if elemSlice, ok := literal.Value.([]ast.Expression); ok {
			return a.generateTupleReturnFromLiteral(elemSlice)
		}
	}

	exprCode, err := a.generateExpression(exprStmt.Expression)
	if err != nil {
		return "", err
	}

	return a.gen.ind() + "return " + exprCode + "\n", nil
}

func (a *ArrowFunctionCodegen) generateTupleReturnFromLiteral(elements []ast.Expression) (string, error) {
	var varNames []string
	for _, elem := range elements {
		elemCode, err := a.generateExpression(elem)
		if err != nil {
			return "", err
		}
		varNames = append(varNames, elemCode)
	}
	return a.gen.ind() + "return " + strings.Join(varNames, ", ") + "\n", nil
}

func (a *ArrowFunctionCodegen) generateExpression(expr ast.Expression) (string, error) {
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Name, nil

	case *ast.Literal:
		return fmt.Sprintf("%v", e.Value), nil

	case *ast.CallExpression:
		return a.gen.generateCallExpression(e)

	case *ast.BinaryExpression:
		return a.generateBinaryExpression(e)

	case *ast.MemberExpression:
		return a.gen.generateMemberExpression(e)

	case *ast.ConditionalExpression:
		return a.generateConditionalExpression(e)

	case *ast.UnaryExpression:
		return a.generateUnaryExpression(e)

	default:
		return "", fmt.Errorf("unsupported expression type: %T", expr)
	}
}

func (a *ArrowFunctionCodegen) generateBinaryExpression(binExpr *ast.BinaryExpression) (string, error) {
	left, err := a.generateExpression(binExpr.Left)
	if err != nil {
		return "", err
	}

	right, err := a.generateExpression(binExpr.Right)
	if err != nil {
		return "", err
	}

	op := a.mapOperator(binExpr.Operator)
	return fmt.Sprintf("(%s %s %s)", left, op, right), nil
}

func (a *ArrowFunctionCodegen) generateConditionalExpression(condExpr *ast.ConditionalExpression) (string, error) {
	testCode, err := a.generateExpression(condExpr.Test)
	if err != nil {
		return "", err
	}

	consCode, err := a.generateExpression(condExpr.Consequent)
	if err != nil {
		return "", err
	}

	altCode, err := a.generateExpression(condExpr.Alternate)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("func() float64 { if %s { return %s } else { return %s } }()", testCode, consCode, altCode), nil
}

func (a *ArrowFunctionCodegen) generateUnaryExpression(unaryExpr *ast.UnaryExpression) (string, error) {
	argCode, err := a.generateExpression(unaryExpr.Argument)
	if err != nil {
		return "", err
	}

	op := a.mapOperator(unaryExpr.Operator)
	return fmt.Sprintf("(%s%s)", op, argCode), nil
}

func (a *ArrowFunctionCodegen) mapOperator(op string) string {
	switch op {
	case "and":
		return "&&"
	case "or":
		return "||"
	case "not":
		return "!"
	default:
		return op
	}
}
