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
	// Ensure integer literals are float64 for Series.Set() compatibility
	// Only convert simple integer literals without decimal points
	if isSimpleInteger(exprCode) {
		exprCode = "float64(" + exprCode + ")"
	}
	return s.indentation + fmt.Sprintf("%s := %s\n", varName, exprCode)
}

/* isSimpleInteger checks if expression is a simple integer literal (no decimal point) */
func isSimpleInteger(expr string) bool {
	// Skip if already has decimal point or function call
	if strings.Contains(expr, ".") || strings.Contains(expr, "(") {
		return false
	}
	// Check if it's a simple numeric literal
	if len(expr) == 0 {
		return false
	}
	for i, ch := range expr {
		// Allow leading minus sign
		if i == 0 && ch == '-' {
			continue
		}
		// Must be digit
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
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
