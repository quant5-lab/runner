package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* MockDMIHandler demonstrates generic composite indicator architecture.
 *
 * Purpose: Proof that adding new composite indicators requires ZERO changes to generator.go
 * DMI (Directional Movement Index) needs 5 internal series:
 * - Plus directional movement (+DM)
 * - Minus directional movement (-DM)
 * - True range (TR)
 * - Smoothed +DM
 * - Smoothed -DM
 *
 * To add DMI support:
 * 1. Create DMIHandler (this file)
 * 2. Implement CompositeIndicatorMetadata interface
 * 3. Register in generator initialization: registry.Register("ta.dmi", &DMIHandler{})
 * 4. DONE - No generator.go modifications needed
 */
type MockDMIHandler struct{}

func (h *MockDMIHandler) CanHandle(funcName string) bool {
	return funcName == "ta.dmi" || funcName == "dmi"
}

func (h *MockDMIHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	return "// Mock DMI generation would go here\n", nil
}

/* GetInternalSeriesNames implements CompositeIndicatorMetadata interface.
 * DMI requires 5 internal series for directional movement calculations.
 */
func (h *MockDMIHandler) GetInternalSeriesNames(varName string, call *ast.CallExpression) ([]string, error) {
	return []string{
		fmt.Sprintf("_%s_plus_dm", varName),
		fmt.Sprintf("_%s_minus_dm", varName),
		fmt.Sprintf("_%s_true_range", varName),
		fmt.Sprintf("_%s_smoothed_plus_dm", varName),
		fmt.Sprintf("_%s_smoothed_minus_dm", varName),
	}, nil
}
