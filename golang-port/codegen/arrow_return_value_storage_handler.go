package codegen

import (
	"fmt"
	"strings"
)

/*
ReturnValueSeriesStorageHandler generates Series.Set() statements for arrow function return values.

Responsibilities:
- Converts scalar return values to Series storage
- Aligns with ForwardSeriesBuffer paradigm
- Maintains PineScript historical value semantics

Design:
- SRP: Single responsibility - generate Series storage code
- KISS: Simple template-based code generation
- DRY: Centralizes all return value storage logic

PineScript Semantics:

	[ADX, up, down] = adx(len)  // Returns become Series
	plot(ADX)                    // Accesses current Series value

Go Translation:

	ADX, up, down := adx(arrowCtx, len)  // Scalar float64 values
	ADXSeries.Set(ADX)                   // Store in Series
	upSeries.Set(up)
	downSeries.Set(down)
	// Later: ADXSeries.GetCurrent() retrieves value
*/
type ReturnValueSeriesStorageHandler struct {
	indentation string
}

func NewReturnValueSeriesStorageHandler(indent string) *ReturnValueSeriesStorageHandler {
	return &ReturnValueSeriesStorageHandler{
		indentation: indent,
	}
}

/* Generates Series.Set() statements for return values to maintain ForwardSeriesBuffer */
func (h *ReturnValueSeriesStorageHandler) GenerateStorageStatements(varNames []string) string {
	if len(varNames) == 0 {
		return ""
	}

	statements := make([]string, len(varNames))
	for i, varName := range varNames {
		statements[i] = h.indentation + h.generateSingleStorageStatement(varName)
	}

	return strings.Join(statements, "\n") + "\n"
}

func (h *ReturnValueSeriesStorageHandler) generateSingleStorageStatement(varName string) string {
	return fmt.Sprintf("%sSeries.Set(%s)", varName, varName)
}

/*
ValidateReturnValueNames checks variable names meet Go identifier rules.
Returns error if any name is invalid.
*/
func (h *ReturnValueSeriesStorageHandler) ValidateReturnValueNames(varNames []string) error {
	for _, name := range varNames {
		if name == "" {
			return fmt.Errorf("empty variable name")
		}
		if !isValidGoIdentifier(name) {
			return fmt.Errorf("invalid Go identifier: %q", name)
		}
	}
	return nil
}

func isValidGoIdentifier(name string) bool {
	if len(name) == 0 {
		return false
	}
	first := name[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}
	for i := 1; i < len(name); i++ {
		c := name[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}
