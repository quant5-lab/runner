package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

type ArrowFunctionCodegen struct {
	gen            *generator
	accessResolver *ArrowSeriesAccessResolver
	seriesVarGen   *ArrowSeriesVariableGenerator
	statementGen   *ArrowStatementGenerator
}

func NewArrowFunctionCodegen(gen *generator) *ArrowFunctionCodegen {
	return &ArrowFunctionCodegen{
		gen:            gen,
		accessResolver: NewArrowSeriesAccessResolver(),
		seriesVarGen:   nil, // Initialized in Generate with proper indentation
	}
}

func (a *ArrowFunctionCodegen) Generate(funcName string, arrowFunc *ast.ArrowFunctionExpression) (string, error) {
	analyzer := NewParameterUsageAnalyzer()
	paramUsage := analyzer.AnalyzeArrowFunction(arrowFunc)

	a.gen.signatureRegistrar.RegisterArrowFunction(funcName, arrowFunc.Params, paramUsage, "float64")

	// Register all parameters in access resolver
	for _, param := range arrowFunc.Params {
		a.accessResolver.RegisterParameter(param.Name)
	}

	// Register all local variables in access resolver
	for _, stmt := range arrowFunc.Body {
		if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
			for _, declarator := range varDecl.Declarations {
				if id, ok := declarator.ID.(*ast.Identifier); ok {
					a.accessResolver.RegisterLocalVariable(id.Name)
				} else if arrayPattern, ok := declarator.ID.(*ast.ArrayPattern); ok {
					for _, elem := range arrayPattern.Elements {
						a.accessResolver.RegisterLocalVariable(elem.Name)
					}
				}
			}
		}
	}

	signature, returnType, err := a.analyzeAndGenerateSignature(funcName, arrowFunc, paramUsage)
	if err != nil {
		return "", err
	}

	// Initialize Series variable generator with proper indentation context
	exprGen := NewArrowExpressionGeneratorImpl(a.gen, a.accessResolver)
	a.seriesVarGen = NewArrowSeriesVariableGenerator(a.gen.ind(), exprGen)
	a.statementGen = NewArrowStatementGenerator(a.gen, a.seriesVarGen, exprGen)

	body, err := a.generateFunctionBody(arrowFunc)
	if err != nil {
		return "", err
	}

	code := a.gen.ind() + signature + " " + returnType + " {\n"
	a.gen.indent++

	code += a.gen.ind() + "ctx := arrowCtx.Context\n\n"

	// Generate Series declarations for ALL local variables (universal ForwardSeriesBuffer)
	seriesDecls := a.generateAllSeriesDeclarations(arrowFunc)
	if seriesDecls != "" {
		code += seriesDecls + "\n"
	}

	code += body
	a.gen.indent--
	code += a.gen.ind() + "}\n\n"

	return code, nil
}

func (a *ArrowFunctionCodegen) analyzeAndGenerateSignature(funcName string, arrowFunc *ast.ArrowFunctionExpression, paramTypes map[string]ParameterUsageType) (string, string, error) {
	params := a.buildParameterList(arrowFunc.Params, paramTypes)
	returnType, err := a.inferReturnType(arrowFunc)
	if err != nil {
		return "", "", err
	}

	signature := fmt.Sprintf("func %s(arrowCtx *context.ArrowContext%s)", funcName, params)
	return signature, returnType, nil
}

func (a *ArrowFunctionCodegen) buildParameterList(params []ast.Identifier, paramTypes map[string]ParameterUsageType) string {
	if len(params) == 0 {
		return ""
	}

	var parts []string
	for _, param := range params {
		paramType := paramTypes[param.Name]
		if paramType == ParameterUsageSeries {
			parts = append(parts, fmt.Sprintf("%sSeries *series.Series", param.Name))
		} else {
			parts = append(parts, fmt.Sprintf("%s float64", param.Name))
		}
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

/*
generateAllSeriesDeclarations creates Series storage for ALL local variables.

Universal ForwardSeriesBuffer paradigm: every variable gets Series storage.
This ensures historical access and TA function compatibility.
*/
func (a *ArrowFunctionCodegen) generateAllSeriesDeclarations(arrowFunc *ast.ArrowFunctionExpression) string {
	var code string

	for _, stmt := range arrowFunc.Body {
		if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
			for _, declarator := range varDecl.Declarations {
				if id, ok := declarator.ID.(*ast.Identifier); ok {
					code += a.seriesVarGen.GenerateDeclaration(id.Name)
				} else if arrayPattern, ok := declarator.ID.(*ast.ArrayPattern); ok {
					for _, elem := range arrayPattern.Elements {
						code += a.seriesVarGen.GenerateDeclaration(elem.Name)
					}
				}
			}
		}
	}

	return code
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

		stmtCode, err := a.statementGen.GenerateStatement(stmt)
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
		// Return Series.GetCurrent() since all variables use Series storage
		return stmtCode + a.gen.ind() + "return " + id.Name + "Series.GetCurrent()\n", nil
	}

	return "", fmt.Errorf("unsupported variable declarator pattern: %T", decl.ID)
}

func (a *ArrowFunctionCodegen) generateTupleReturn(arrayPattern *ast.ArrayPattern, init ast.Expression) (string, error) {
	if len(arrayPattern.Elements) == 0 {
		return "", fmt.Errorf("empty tuple pattern")
	}

	var returnVars []string
	for _, elem := range arrayPattern.Elements {
		// Use Series.GetCurrent() for all tuple elements
		returnVars = append(returnVars, elem.Name+"Series.GetCurrent()")
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

	baseVarNames := make([]string, len(varNames))
	tempVarNames := make([]string, len(varNames))
	for i, varName := range varNames {
		baseName := strings.TrimSuffix(varName, "Series.GetCurrent()")
		baseVarNames[i] = baseName
		tempVarNames[i] = "temp_" + baseName
	}

	code := a.gen.ind() + strings.Join(tempVarNames, ", ") + " := " + exprCode + "\n"

	for i, baseName := range baseVarNames {
		code += a.gen.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", baseName, tempVarNames[i])
	}

	return code, nil
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
	// Delegate ALL expression generation to Series-aware generator
	// This ensures proper identifier resolution (parameters vs local variables)
	// AND proper inline TA generation with arrow-aware accessors
	exprGen := NewArrowExpressionGeneratorImpl(a.gen, a.accessResolver)
	return exprGen.Generate(expr)
}
