package codegen

import (
	"fmt"
	"strings"
)

type ArrowLocalVariableStorage struct {
	indentation string
	normalizer  *ExpressionNormalizer
}

func NewArrowLocalVariableStorage(indent string) *ArrowLocalVariableStorage {
	return &ArrowLocalVariableStorage{
		indentation: indent,
		normalizer:  NewExpressionNormalizer(),
	}
}

func (s *ArrowLocalVariableStorage) GenerateScalarOperation(
	varName string,
	exprCode string,
	operationType VariableOperationType,
) string {
	normalized := s.normalizer.NormalizeForSeriesStorage(exprCode)
	operator := operationType.GoAssignmentOperator()
	return s.indentation + fmt.Sprintf("%s %s %s\n", varName, operator, normalized)
}

func (s *ArrowLocalVariableStorage) GenerateSeriesStorage(varName string) string {
	return s.indentation + fmt.Sprintf("%sSeries.Set(%s)\n", varName, varName)
}

func (s *ArrowLocalVariableStorage) GenerateScalarAndSeriesStorage(
	varName string,
	exprCode string,
	operationType VariableOperationType,
) string {
	return s.GenerateScalarOperation(varName, exprCode, operationType) +
		s.GenerateSeriesStorage(varName)
}

func (s *ArrowLocalVariableStorage) GenerateScalarDeclaration(varName, exprCode string) string {
	return s.GenerateScalarOperation(varName, exprCode, VariableDeclaration)
}

func (s *ArrowLocalVariableStorage) GenerateScalarReassignment(varName, exprCode string) string {
	return s.GenerateScalarOperation(varName, exprCode, VariableReassignment)
}

func (s *ArrowLocalVariableStorage) GenerateDualStorage(varName, exprCode string) string {
	return s.GenerateScalarAndSeriesStorage(varName, exprCode, VariableDeclaration)
}

func (s *ArrowLocalVariableStorage) GenerateDualReassignment(varName, exprCode string) string {
	return s.GenerateScalarAndSeriesStorage(varName, exprCode, VariableReassignment)
}

func isSimpleInteger(expr string) bool {
	return isSimpleIntegerLiteral(expr)
}

/* GenerateTupleDualStorage generates dual storage for tuple destructuring */
func (s *ArrowLocalVariableStorage) GenerateTupleDualStorage(varNames []string, exprCode string) string {
	tempVars := make([]string, len(varNames))
	for i, name := range varNames {
		tempVars[i] = "temp_" + name
	}

	code := s.indentation + fmt.Sprintf("%s := %s\n", strings.Join(tempVars, ", "), exprCode)

	for i, varName := range varNames {
		code += s.GenerateScalarDeclaration(varName, tempVars[i])
		code += s.GenerateSeriesStorage(varName)
	}

	return code
}
