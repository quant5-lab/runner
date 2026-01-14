package codegen

import (
	"fmt"
	"strings"
)

/* ArrowLocalVariableStorage manages dual scalar+series storage for arrow function local variables */
type ArrowLocalVariableStorage struct {
	indentation string
}

func NewArrowLocalVariableStorage(indent string) *ArrowLocalVariableStorage {
	return &ArrowLocalVariableStorage{
		indentation: indent,
	}
}

/* GenerateScalarDeclaration generates scalar variable declaration */
func (s *ArrowLocalVariableStorage) GenerateScalarDeclaration(varName, exprCode string) string {
	return s.indentation + fmt.Sprintf("%s := %s\n", varName, exprCode)
}

/* GenerateSeriesStorage generates Series.Set() call to persist scalar value for history */
func (s *ArrowLocalVariableStorage) GenerateSeriesStorage(varName string) string {
	return s.indentation + fmt.Sprintf("%sSeries.Set(%s)\n", varName, varName)
}

/* GenerateDualStorage generates both scalar declaration and series storage */
func (s *ArrowLocalVariableStorage) GenerateDualStorage(varName, exprCode string) string {
	return s.GenerateScalarDeclaration(varName, exprCode) +
		s.GenerateSeriesStorage(varName)
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
