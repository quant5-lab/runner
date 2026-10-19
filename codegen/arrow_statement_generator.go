package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* ArrowStatementGenerator generates variable declarations with dual scalar+series pattern */
type ArrowStatementGenerator struct {
	gen           *generator
	localStorage  *ArrowLocalVariableStorage
	exprGenerator *ArrowExpressionGeneratorImpl
	symbolTable   SymbolTable
	coercer       *NumericExpressionCoercer
	tupleSecGen   *ArrowTupleSecurityGenerator
}

func NewArrowStatementGenerator(
	gen *generator,
	localStorage *ArrowLocalVariableStorage,
	exprGen *ArrowExpressionGeneratorImpl,
	symbolTable SymbolTable,
) *ArrowStatementGenerator {
	return &ArrowStatementGenerator{
		gen:           gen,
		localStorage:  localStorage,
		exprGenerator: exprGen,
		symbolTable:   symbolTable,
		coercer:       NewNumericExpressionCoercer(gen.boolConverter),
	}
}

func (s *ArrowStatementGenerator) WithTupleSecurityGenerator(g *ArrowTupleSecurityGenerator) *ArrowStatementGenerator {
	s.tupleSecGen = g
	return s
}

/* GenerateStatement generates arrow-aware statement code with Series.Set() for variables */
func (s *ArrowStatementGenerator) GenerateStatement(stmt ast.Node) (string, error) {
	switch st := stmt.(type) {
	case *ast.VariableDeclaration:
		return s.generateVariableDeclaration(st)

	case *ast.ForStatement:
		return s.generateForStatement(st)

	case *ast.ForInStatement:
		return s.generateForInStatement(st)

	case *ast.WhileStatement:
		return s.generateWhileStatement(st)

	case *ast.IfStatement:
		return s.generateIfStatement(st)

	default:
		return s.gen.generateStatement(stmt)
	}
}

func (s *ArrowStatementGenerator) generateVariableDeclaration(varDecl *ast.VariableDeclaration) (string, error) {
	if len(varDecl.Declarations) == 0 {
		return "", fmt.Errorf("empty variable declaration")
	}

	decl := varDecl.Declarations[0]

	if arrayPattern, ok := decl.ID.(*ast.ArrayPattern); ok {
		return s.generateTupleDeclaration(arrayPattern, decl.Init)
	}

	if id, ok := decl.ID.(*ast.Identifier); ok {
		operationType := OperationTypeFromASTKind(varDecl.Kind)
		return s.generateSingleVariableDeclaration(id.Name, decl.Init, operationType)
	}

	return "", fmt.Errorf("unsupported variable declarator pattern: %T", decl.ID)
}

func (s *ArrowStatementGenerator) generateSingleVariableDeclaration(
	varName string,
	initExpr ast.Expression,
	operationType VariableOperationType,
) (string, error) {
	if s.symbolTable != nil {
		s.symbolTable.Register(varName, VariableTypeSeries)
	}

	exprCode, err := s.exprGenerator.Generate(initExpr)
	if err != nil {
		return "", fmt.Errorf("failed to generate init expression for '%s': %w", varName, err)
	}

	exprCode = s.coercer.CoerceToFloat64(initExpr, exprCode)

	return s.localStorage.GenerateScalarAndSeriesStorage(varName, exprCode, operationType), nil
}

func (s *ArrowStatementGenerator) generateTupleDeclaration(arrayPattern *ast.ArrayPattern, initExpr ast.Expression) (string, error) {
	varNames := make([]string, len(arrayPattern.Elements))
	for i, elem := range arrayPattern.Elements {
		varNames[i] = elem.Name
		if s.symbolTable != nil {
			s.symbolTable.Register(varNames[i], VariableTypeSeries)
		}
	}

	if call, ok := initExpr.(*ast.CallExpression); ok {
		if isSecurityCallExpression(call) && s.tupleSecGen != nil {
			return s.tupleSecGen.Generate(varNames, call)
		}

		funcName := extractCallFunctionName(call)
		detector := NewUserDefinedFunctionDetector(s.gen.variables)
		if detector.IsUserDefinedFunction(funcName) {
			return s.gen.generateUserDefinedFunctionTupleCall(varNames, funcName, call)
		}
	}

	exprCode, err := s.exprGenerator.Generate(initExpr)
	if err != nil {
		return "", fmt.Errorf("failed to generate tuple init expression: %w", err)
	}

	return s.localStorage.GenerateTupleDualStorage(varNames, exprCode), nil
}

/* generateForStatement generates arrow-aware for-loop using ArrowExpressionGeneratorImpl */
func (s *ArrowStatementGenerator) generateForStatement(forStmt *ast.ForStatement) (string, error) {
	counterVar := forStmt.Counter

	s.gen.loopContextStack.Push(counterVar)
	defer s.gen.loopContextStack.Pop()

	/* Use arrow-aware expression generator for bounds (respects parameters vs locals) */
	fromCode, err := s.exprGenerator.Generate(forStmt.From)
	if err != nil {
		return "", fmt.Errorf("for-loop from expression failed: %w", err)
	}

	toCode, err := s.exprGenerator.Generate(forStmt.To)
	if err != nil {
		return "", fmt.Errorf("for-loop to expression failed: %w", err)
	}

	stepCode := "1"
	if forStmt.Step != nil {
		stepCode, err = s.exprGenerator.Generate(forStmt.Step)
		if err != nil {
			return "", fmt.Errorf("for-loop step expression failed: %w", err)
		}
	}

	code := s.gen.ind() + "{\n"
	s.gen.indent++
	code += s.gen.ind() + fmt.Sprintf("%s := int(%s)\n", counterVar, fromCode)
	code += s.gen.ind() + fmt.Sprintf("_to := int(%s)\n", toCode)
	code += s.gen.ind() + fmt.Sprintf("_step := int(%s)\n", stepCode)

	code += s.gen.ind() + "if _step == 0 {\n"
	s.gen.indent++
	code += s.gen.ind() + "panic(\"for loop step cannot be zero\")\n"
	s.gen.indent--
	code += s.gen.ind() + "}\n"

	code += s.gen.ind() + "_ascending := _step > 0\n"
	code += s.gen.ind() + fmt.Sprintf("for ; (_ascending && %s <= _to) || (!_ascending && %s >= _to); %s += _step {\n", counterVar, counterVar, counterVar)
	s.gen.indent++

	for _, stmt := range forStmt.Body {
		stmtCode, err := s.GenerateStatement(stmt)
		if err != nil {
			return "", fmt.Errorf("for-loop body statement failed: %w", err)
		}
		if stmtCode != "" {
			code += stmtCode
		}
	}

	s.gen.indent--
	code += s.gen.ind() + "}\n"
	s.gen.indent--
	code += s.gen.ind() + "}\n"

	return code, nil
}

func (s *ArrowStatementGenerator) generateForInStatement(forIn *ast.ForInStatement) (string, error) {
	collCode, err := s.exprGenerator.Generate(forIn.Collection)
	if err != nil {
		return "", fmt.Errorf("for-in collection expression failed: %w", err)
	}

	indexVar := "_"
	if forIn.IndexVar != "" {
		indexVar = forIn.IndexVar
	}
	elementVar := forIn.ElementVar

	code := s.gen.ind() + fmt.Sprintf("for %s, %s := range %s {\n", indexVar, elementVar, collCode)
	s.gen.indent++

	for _, stmt := range forIn.Body {
		stmtCode, err := s.GenerateStatement(stmt)
		if err != nil {
			return "", fmt.Errorf("for-in body statement failed: %w", err)
		}
		if stmtCode != "" {
			code += stmtCode
		}
	}

	s.gen.indent--
	code += s.gen.ind() + "}\n"

	return code, nil
}

func (s *ArrowStatementGenerator) generateWhileStatement(whileStmt *ast.WhileStatement) (string, error) {
	s.gen.loopContextStack.Push("")
	defer s.gen.loopContextStack.Pop()

	condCode, err := s.exprGenerator.Generate(whileStmt.Condition)
	if err != nil {
		return "", fmt.Errorf("while-loop condition expression failed: %w", err)
	}

	condCode = s.gen.addBoolConversionIfNeeded(whileStmt.Condition, condCode)

	guard := NewLoopIterationGuard()

	code := s.gen.ind() + "{\n"
	s.gen.indent++

	code += guard.InitCode(s.gen.ind())
	code += s.gen.ind() + fmt.Sprintf("for %s {\n", condCode)
	s.gen.indent++

	code += guard.CheckCode(s.gen.ind())

	for _, stmt := range whileStmt.Body {
		stmtCode, err := s.GenerateStatement(stmt)
		if err != nil {
			return "", fmt.Errorf("while-loop body statement failed: %w", err)
		}
		if stmtCode != "" {
			code += stmtCode
		}
	}

	s.gen.indent--
	code += s.gen.ind() + "}\n"
	s.gen.indent--
	code += s.gen.ind() + "}\n"

	return code, nil
}
