package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestUserDefinedFunctionHandler_CanHandle validates handler recognizes user functions */
func TestUserDefinedFunctionHandler_CanHandle(t *testing.T) {
	handler := &UserDefinedFunctionHandler{}

	// Handler determines dynamically by checking g.variables, so CanHandle always returns false
	if handler.CanHandle("myFunc") {
		t.Error("CanHandle should return false - dynamic check happens in GenerateCode")
	}
}

/* TestUserDefinedFunctionHandler_GenerateCode validates function call generation */
func TestUserDefinedFunctionHandler_GenerateCode(t *testing.T) {
	gen := newTestGenerator()

	// Register a user-defined function
	gen.variables["double"] = "function"

	handler := &UserDefinedFunctionHandler{}

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "double"},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
		},
	}

	code, err := handler.GenerateCode(gen, call)
	if err != nil {
		t.Fatalf("GenerateCode() error: %v", err)
	}

	if code != "double(arrowCtx_double_1, closeSeries.Get(0))" {
		t.Errorf("Expected 'double(arrowCtx_double_1, closeSeries.Get(0))', got %q", code)
	}
}

/* TestUserDefinedFunctionHandler_NotUserDefined validates passthrough for non-user functions */
func TestUserDefinedFunctionHandler_GenerateCode_NotUserDefined(t *testing.T) {
	gen := newTestGenerator()

	handler := &UserDefinedFunctionHandler{}

	call := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "unknownFunc"},
		Arguments: []ast.Expression{},
	}

	code, err := handler.GenerateCode(gen, call)
	if err != nil {
		t.Fatalf("GenerateCode() error: %v", err)
	}

	if code != "" {
		t.Errorf("Expected empty string for non-user function, got %q", code)
	}
}
