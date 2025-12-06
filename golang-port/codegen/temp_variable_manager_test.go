package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestTempVariableManager_GetOrCreate tests basic temp var generation */
func TestTempVariableManager_GetOrCreate(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
	g.taRegistry = NewTAFunctionRegistry()
	mgr := NewTempVariableManager(g)

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 20},
		},
	}

	info := CallInfo{
		Call:     call,
		FuncName: "ta.sma",
		ArgHash:  "abc123",
	}

	varName := mgr.GetOrCreate(info)

	// Check format: ta_sma_20_abc123
	if !strings.HasPrefix(varName, "ta_sma_20_") {
		t.Errorf("Expected varName to start with 'ta_sma_20_', got %q", varName)
	}

	if !strings.Contains(varName, "abc123") {
		t.Errorf("Expected varName to contain hash 'abc123', got %q", varName)
	}

	// Check it was NOT added to g.variables (managed separately)
	if _, exists := g.variables[varName]; exists {
		t.Error("Temp var should NOT be in g.variables (managed by TempVariableManager)")
	}
}

/* TestTempVariableManager_Deduplication tests that same call returns same var */
func TestTempVariableManager_Deduplication(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
	g.taRegistry = NewTAFunctionRegistry()
	mgr := NewTempVariableManager(g)

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 50},
		},
	}

	info := CallInfo{
		Call:     call,
		FuncName: "ta.sma",
		ArgHash:  "def456",
	}

	varName1 := mgr.GetOrCreate(info)
	varName2 := mgr.GetOrCreate(info)

	if varName1 != varName2 {
		t.Errorf("Expected same varName for same call, got %q vs %q", varName1, varName2)
	}

	// Should only be declared once
	if len(mgr.declaredVars) != 1 {
		t.Errorf("Expected 1 declared var, got %d", len(mgr.declaredVars))
	}
}

/* TestTempVariableManager_DifferentCalls tests that different calls get different vars */
func TestTempVariableManager_DifferentCalls(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
	g.taRegistry = NewTAFunctionRegistry()
	mgr := NewTempVariableManager(g)

	call1 := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 50},
		},
	}

	call2 := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 200},
		},
	}

	info1 := CallInfo{Call: call1, FuncName: "ta.sma", ArgHash: "hash1"}
	info2 := CallInfo{Call: call2, FuncName: "ta.sma", ArgHash: "hash2"}

	varName1 := mgr.GetOrCreate(info1)
	varName2 := mgr.GetOrCreate(info2)

	if varName1 == varName2 {
		t.Errorf("Expected different varNames for different calls, both got %q", varName1)
	}

	// Should have 2 declared vars
	if len(mgr.declaredVars) != 2 {
		t.Errorf("Expected 2 declared vars, got %d", len(mgr.declaredVars))
	}
}

/* TestTempVariableManager_GenerateDeclarations tests declaration code generation */
func TestTempVariableManager_GenerateDeclarations(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
		indent:    1,
	}
	g.taRegistry = NewTAFunctionRegistry()
	mgr := NewTempVariableManager(g)

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 20},
		},
	}

	info := CallInfo{Call: call, FuncName: "ta.sma", ArgHash: "test123"}
	varName := mgr.GetOrCreate(info)

	decls := mgr.GenerateDeclarations()

	if !strings.Contains(decls, "// Temp variables for inline TA calls") {
		t.Error("Expected comment in declarations")
	}

	if !strings.Contains(decls, "var "+varName+"Series *series.Series") {
		t.Errorf("Expected declaration for %s, got:\n%s", varName, decls)
	}
}

/* TestTempVariableManager_GenerateInitializations tests initialization code generation */
func TestTempVariableManager_GenerateInitializations(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
		indent:    1,
	}
	g.taRegistry = NewTAFunctionRegistry()
	mgr := NewTempVariableManager(g)

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "ema"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 10},
		},
	}

	info := CallInfo{Call: call, FuncName: "ta.ema", ArgHash: "init789"}
	varName := mgr.GetOrCreate(info)

	inits := mgr.GenerateInitializations()

	if !strings.Contains(inits, varName+"Series = series.NewSeries(len(ctx.Data))") {
		t.Errorf("Expected initialization for %s, got:\n%s", varName, inits)
	}
}

/* TestTempVariableManager_GenerateNextCalls tests Next() call generation */
func TestTempVariableManager_GenerateNextCalls(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
		indent:    1,
	}
	g.taRegistry = NewTAFunctionRegistry()
	mgr := NewTempVariableManager(g)

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "rma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 14},
		},
	}

	info := CallInfo{Call: call, FuncName: "ta.rma", ArgHash: "next999"}
	varName := mgr.GetOrCreate(info)

	nextCalls := mgr.GenerateNextCalls()

	if !strings.Contains(nextCalls, "if i < barCount-1 { "+varName+"Series.Next() }") {
		t.Errorf("Expected Next() call for %s, got:\n%s", varName, nextCalls)
	}
}

/* TestTempVariableManager_EmptyManager tests behavior with no temp vars */
func TestTempVariableManager_EmptyManager(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
	g.taRegistry = NewTAFunctionRegistry()
	mgr := NewTempVariableManager(g)

	decls := mgr.GenerateDeclarations()
	inits := mgr.GenerateInitializations()
	nexts := mgr.GenerateNextCalls()

	if decls != "" {
		t.Errorf("Expected empty declarations, got: %q", decls)
	}
	if inits != "" {
		t.Errorf("Expected empty initializations, got: %q", inits)
	}
	if nexts != "" {
		t.Errorf("Expected empty next calls, got: %q", nexts)
	}
}

/* TestTempVariableManager_ExtractPeriod tests period extraction from arguments */
func TestTempVariableManager_ExtractPeriod(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
	g.taRegistry = NewTAFunctionRegistry()
	mgr := NewTempVariableManager(g)

	testCases := []struct {
		name           string
		secondArg      ast.Expression
		expectedPeriod int
	}{
		{
			name:           "int literal",
			secondArg:      &ast.Literal{Value: 20},
			expectedPeriod: 20,
		},
		{
			name:           "float literal",
			secondArg:      &ast.Literal{Value: 50.0},
			expectedPeriod: 50,
		},
		{
			name:           "identifier (non-literal)",
			secondArg:      &ast.Identifier{Name: "period"},
			expectedPeriod: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					tc.secondArg,
				},
			}

			period := mgr.extractPeriodFromCall(call)
			if period != tc.expectedPeriod {
				t.Errorf("Expected period %d, got %d", tc.expectedPeriod, period)
			}
		})
	}
}

/* TestTempVariableManager_Reset tests clearing state */
func TestTempVariableManager_Reset(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
	g.taRegistry = NewTAFunctionRegistry()
	mgr := NewTempVariableManager(g)

	// Add some temp vars
	call := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "sma"},
		Arguments: []ast.Expression{&ast.Literal{Value: 20}},
	}
	info := CallInfo{Call: call, FuncName: "ta.sma", ArgHash: "reset123"}
	mgr.GetOrCreate(info)

	if len(mgr.declaredVars) == 0 {
		t.Fatal("Expected declared vars before reset")
	}

	// Reset
	mgr.Reset()

	if len(mgr.declaredVars) != 0 {
		t.Errorf("Expected 0 declared vars after reset, got %d", len(mgr.declaredVars))
	}
	if len(mgr.callToVar) != 0 {
		t.Errorf("Expected 0 call mappings after reset, got %d", len(mgr.callToVar))
	}
	if len(mgr.varToCallInfo) != 0 {
		t.Errorf("Expected 0 var mappings after reset, got %d", len(mgr.varToCallInfo))
	}
}
