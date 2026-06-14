package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

func (s *ArrowStatementGenerator) generateIfStatement(ifStmt *ast.IfStatement) (string, error) {
	condCode, err := s.exprGenerator.Generate(ifStmt.Test)
	if err != nil {
		return "", fmt.Errorf("if-statement condition: %w", err)
	}
	condCode = s.gen.addBoolConversionIfNeeded(ifStmt.Test, condCode)

	code := s.gen.ind() + fmt.Sprintf("if %s {\n", condCode)
	s.gen.indent++

	bodyCode, err := s.generateIfBody(ifStmt.Consequent)
	if err != nil {
		return "", err
	}
	code += bodyCode

	s.gen.indent--

	altCode, err := s.generateIfAlternate(ifStmt.Alternate)
	if err != nil {
		return "", err
	}
	return code + altCode, nil
}

// generateIfBody skips pure-value expression statements — Pine's implicit branch-return
// syntax has no side effect in statement position.
func (s *ArrowStatementGenerator) generateIfBody(body []ast.Node) (string, error) {
	var code string
	for _, stmt := range body {
		if shouldSkipIfBodyStatement(stmt) {
			continue
		}
		stmtCode, err := s.GenerateStatement(stmt)
		if err != nil {
			return "", fmt.Errorf("if-body statement: %w", err)
		}
		code += stmtCode
	}
	return code, nil
}

func (s *ArrowStatementGenerator) generateIfAlternate(alternate []ast.Node) (string, error) {
	if len(alternate) == 0 {
		return s.gen.ind() + "}\n", nil
	}
	if elseIf, ok := extractElseIfChain(alternate); ok {
		return s.generateElseIfChain(elseIf)
	}
	return s.generateElseBlock(alternate)
}

func (s *ArrowStatementGenerator) generateElseIfChain(ifStmt *ast.IfStatement) (string, error) {
	condCode, err := s.exprGenerator.Generate(ifStmt.Test)
	if err != nil {
		return "", fmt.Errorf("else-if condition: %w", err)
	}
	condCode = s.gen.addBoolConversionIfNeeded(ifStmt.Test, condCode)

	code := s.gen.ind() + fmt.Sprintf("} else if %s {\n", condCode)
	s.gen.indent++

	bodyCode, err := s.generateIfBody(ifStmt.Consequent)
	if err != nil {
		return "", err
	}
	code += bodyCode

	s.gen.indent--

	altCode, err := s.generateIfAlternate(ifStmt.Alternate)
	if err != nil {
		return "", err
	}
	return code + altCode, nil
}

func (s *ArrowStatementGenerator) generateElseBlock(nodes []ast.Node) (string, error) {
	code := s.gen.ind() + "} else {\n"
	s.gen.indent++

	for _, stmt := range nodes {
		if shouldSkipIfBodyStatement(stmt) {
			continue
		}
		stmtCode, err := s.GenerateStatement(stmt)
		if err != nil {
			return "", fmt.Errorf("else-block statement: %w", err)
		}
		code += stmtCode
	}

	s.gen.indent--
	code += s.gen.ind() + "}\n"
	return code, nil
}
