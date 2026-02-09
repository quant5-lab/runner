package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// TestTAIndicatorCallHandler_CanHandle verifies TA indicator recognition
func TestTAIndicatorCallHandler_CanHandle(t *testing.T) {
	handler := &TAIndicatorCallHandler{}

	tests := []struct {
		funcName string
		want     bool
	}{
		// Standard TA functions
		{"ta.sma", true},
		{"ta.ema", true},
		{"ta.stdev", true},
		{"ta.rma", true},
		{"ta.wma", true},

		// Crossover functions
		{"ta.crossover", true},
		{"ta.crossunder", true},

		// Other TA functions
		{"ta.change", true},
		{"ta.pivothigh", true},
		{"ta.pivotlow", true},

		// Utility functions
		{"fixnan", true},
		{"valuewhen", true},

		// Should not handle
		{"strategy.entry", false},
		{"plot", false},
		{"", false},

		// Now handled via unified registry
		{"ta.highest", true},
		{"sma", true},

		/* Tuple indicators excluded — owned by TupleIndicatorHandler */
		{"ta.macd", false},
		{"macd", false},
		{"ta.dmi", false},
		{"dmi", false},
		{"ta.bb", false},
		{"bb", false},
		{"ta.stoch", false},
		{"stoch", false},
		{"ta.supertrend", false},
		{"supertrend", false},
		{"ta.kc", false},
		{"kc", false},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			got := handler.CanHandle(tt.funcName)
			if got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

// TestTAIndicatorCallHandler_GenerateCode verifies no immediate code for TA calls
func TestTAIndicatorCallHandler_GenerateCode(t *testing.T) {
	handler := &TAIndicatorCallHandler{}
	g := newTestGenerator()

	tests := []struct {
		name string
		call *ast.CallExpression
	}{
		{
			name: "ta.sma call",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 20.0},
				},
			},
		},
		{
			name: "ta.crossover call",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "crossover"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "fast"},
					&ast.Identifier{Name: "slow"},
				},
			},
		},
		{
			name: "valuewhen call",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "valuewhen"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "condition"},
					&ast.Identifier{Name: "source"},
					&ast.Literal{Value: 0.0},
				},
			},
		},
		{
			name: "fixnan call",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "fixnan"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "value"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := handler.GenerateCode(g, tt.call)
			if err != nil {
				t.Errorf("GenerateCode() unexpected error: %v", err)
			}

			// TA indicators are handled in variable declarations, not as statements
			if code != "" {
				t.Errorf("GenerateCode() should return empty string for TA indicators, got: %q", code)
			}
		})
	}
}

// TestTAIndicatorCallHandler_EdgeCases tests boundary conditions
func TestTAIndicatorCallHandler_EdgeCases(t *testing.T) {
	handler := &TAIndicatorCallHandler{}

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{"only ta.", "ta.", false},
		{"ta with space", "ta. sma", false},
		{"uppercase", "TA.SMA", false},
		{"partial match", "ta.sm", false},
		{"extra prefix", "x.ta.sma", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.CanHandle(tt.funcName)
			if got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

// TestTAIndicatorCallHandler_NoSideEffects verifies handler doesn't modify generator state
func TestTAIndicatorCallHandler_NoSideEffects(t *testing.T) {
	handler := &TAIndicatorCallHandler{}
	g := newTestGenerator()

	initialVarCount := len(g.variables)

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 20.0},
		},
	}

	_, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Fatalf("GenerateCode() error: %v", err)
	}

	// Handler should not modify generator state during call expression handling
	if len(g.variables) != initialVarCount {
		t.Error("Handler should not modify generator variables during call expression handling")
	}
}
