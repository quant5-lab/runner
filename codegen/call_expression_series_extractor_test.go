package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestExtractCallExpression_ExecutionOrder(t *testing.T) {
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "ta.sma"},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 20.0},
		},
	}

	result := g.extractCallExpression(call)

	if result == "" {
		t.Error("Expected non-empty result from extractCallExpression")
	}
}

func TestExtractCallExpression_NilCall(t *testing.T) {
	g := newTestGenerator()

	result := g.extractCallExpression(nil)

	if result != "Series.GetCurrent()" {
		t.Errorf("Expected Series.GetCurrent() for nil call (fallback to default), got %q", result)
	}
}

func TestExtractTempVariable_ExistingTempVariable(t *testing.T) {
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "ta.sma"},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 20.0},
		},
	}

	info := CallInfo{
		Call:     call,
		FuncName: "ta.sma",
		ArgHash:  "test123",
	}
	g.tempVarMgr.GetOrCreate(info)

	result := g.extractTempVariable(call)

	if result == "" {
		t.Error("Expected non-empty result for existing temp variable")
	}
	if !contains(result, "Series.GetCurrent()") {
		t.Errorf("Expected Series.GetCurrent() in result, got %q", result)
	}
}

func TestExtractTempVariable_NoTempVariable(t *testing.T) {
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "ta.rsi"},
	}

	result := g.extractTempVariable(call)

	if result != "" {
		t.Errorf("Expected empty result when no temp variable, got %q", result)
	}
}

func TestExtractTempVariable_NilCall(t *testing.T) {
	g := newTestGenerator()

	result := g.extractTempVariable(nil)

	if result != "" {
		t.Errorf("Expected empty result for nil call, got %q", result)
	}
}

func TestExtractValueFunction_ValidValueFunction(t *testing.T) {
	g := newTestGenerator()
	g.valueHandler = NewValueHandler()

	call := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "na"},
		Arguments: []ast.Expression{},
	}

	result := g.extractValueFunction(call)

	if result == "" {
		t.Error("Expected non-empty result for value function")
	}
}

func TestExtractValueFunction_NotValueFunction(t *testing.T) {
	g := newTestGenerator()
	g.valueHandler = NewValueHandler()

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "customFunc"},
	}

	result := g.extractValueFunction(call)

	if result != "" {
		t.Errorf("Expected empty result for non-value function, got %q", result)
	}
}

func TestExtractValueFunction_NilCall(t *testing.T) {
	g := newTestGenerator()

	result := g.extractValueFunction(nil)

	if result != "" {
		t.Errorf("Expected empty result for nil call, got %q", result)
	}
}

func TestExtractMathFunction_ValidMathFunction(t *testing.T) {
	g := newTestGenerator()
	g.mathHandler = NewMathHandler()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "math"},
			Property: &ast.Identifier{Name: "max"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: 1.0},
			&ast.Literal{Value: 2.0},
		},
	}

	result := g.extractMathFunction(call)

	if result == "" {
		t.Error("Expected non-empty result for math function")
	}
}

func TestExtractMathFunction_NotMathFunction(t *testing.T) {
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "customFunc"},
	}

	result := g.extractMathFunction(call)

	if result != "" {
		t.Errorf("Expected empty result for non-math function, got %q", result)
	}
}

func TestExtractMathFunction_NilCall(t *testing.T) {
	g := newTestGenerator()

	result := g.extractMathFunction(nil)

	if result != "" {
		t.Errorf("Expected empty result for nil call, got %q", result)
	}
}

func TestExtractColorFunction_ValidColorFunction(t *testing.T) {
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "color"},
			Property: &ast.Identifier{Name: "new"},
		},
		Arguments: []ast.Expression{
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Identifier{Name: "red"},
			},
			&ast.Literal{Value: 50.0},
		},
	}

	result := g.extractColorFunction(call)

	if result == "" {
		t.Error("Expected non-empty result for color function")
	}
}

func TestExtractColorFunction_NotColorFunction(t *testing.T) {
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "customFunc"},
	}

	result := g.extractColorFunction(call)

	if result != "" {
		t.Errorf("Expected empty result for non-color function, got %q", result)
	}
}

func TestExtractColorFunction_NilCall(t *testing.T) {
	g := newTestGenerator()

	result := g.extractColorFunction(nil)

	if result != "" {
		t.Errorf("Expected empty result for nil call, got %q", result)
	}
}

func TestExtractUserDefinedFunction_ValidUserDefinedFunction(t *testing.T) {
	g := newTestGenerator()
	g.variables["customFunc"] = "function"

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "customFunc"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: 10.0},
		},
	}

	result := g.extractUserDefinedFunction(call)

	if result == "" {
		t.Error("Expected non-empty result for user-defined function")
	}
	if !contains(result, "customFunc") {
		t.Errorf("Expected customFunc in result, got %q", result)
	}
	if !contains(result, "arrowCtx") {
		t.Errorf("Expected arrowCtx in result, got %q", result)
	}
}

func TestExtractUserDefinedFunction_NotUserDefinedFunction(t *testing.T) {
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "ta.sma"},
	}

	result := g.extractUserDefinedFunction(call)

	if result != "" {
		t.Errorf("Expected empty result for non-user-defined function, got %q", result)
	}
}

func TestExtractUserDefinedFunction_NilCall(t *testing.T) {
	g := newTestGenerator()

	result := g.extractUserDefinedFunction(nil)

	if result != "" {
		t.Errorf("Expected empty result for nil call, got %q", result)
	}
}

func TestExtractDefaultSeries_ValidCall(t *testing.T) {
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "ta.sma"},
	}

	result := g.extractDefaultSeries(call)

	if result == "" {
		t.Error("Expected non-empty result for default series")
	}
	if !contains(result, "ta_smaSeries.GetCurrent()") {
		t.Errorf("Expected ta_smaSeries.GetCurrent() in result, got %q", result)
	}
}

func TestExtractDefaultSeries_NilCall(t *testing.T) {
	g := newTestGenerator()

	result := g.extractDefaultSeries(nil)

	if result != "Series.GetCurrent()" {
		t.Errorf("Expected Series.GetCurrent() for nil call, got %q", result)
	}
}

func TestMathHandler_CanHandle_ValidMathFunctions(t *testing.T) {
	mh := NewMathHandler()

	tests := []struct {
		name     string
		funcName string
		expected bool
	}{
		{"math.max", "math.max", true},
		{"math.min", "math.min", true},
		{"math.abs", "math.abs", true},
		{"math.pow", "math.pow", true},
		{"max builtin", "max", true},
		{"min builtin", "min", true},
		{"abs builtin", "abs", true},
		{"pow builtin", "pow", true},
		{"sqrt builtin", "sqrt", true},
		{"floor builtin", "floor", true},
		{"ceil builtin", "ceil", true},
		{"round builtin", "round", true},
		{"log builtin", "log", true},
		{"exp builtin", "exp", true},
		{"not math", "ta.sma", false},
		{"not math", "customFunc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mh.CanHandle(tt.funcName)
			if result != tt.expected {
				t.Errorf("MathHandler.CanHandle(%q) = %v, want %v", tt.funcName, result, tt.expected)
			}
		})
	}
}
