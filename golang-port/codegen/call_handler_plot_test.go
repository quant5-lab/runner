package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// TestPlotFunctionHandler_CanHandle verifies plot function recognition
func TestPlotFunctionHandler_CanHandle(t *testing.T) {
	handler := &PlotFunctionHandler{}

	tests := []struct {
		funcName string
		want     bool
	}{
		{"plot", true},
		{"Plot", false}, // Case-sensitive
		{"ta.plot", false},
		{"plotshape", false},
		{"", false},
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

// TestPlotFunctionHandler_GenerateCode verifies collector.Add generation
func TestPlotFunctionHandler_GenerateCode(t *testing.T) {
	handler := &PlotFunctionHandler{}
	g := newTestGenerator()

	tests := []struct {
		name         string
		call         *ast.CallExpression
		wantContains []string
	}{
		{
			name: "simple variable plot",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "plot"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
				},
			},
			wantContains: []string{}, // Empty - added to plotCollector, not immediate code
		},
		{
			name: "plot with title",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "plot"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "sma20"},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "title"},
								Value: &ast.Literal{Value: "SMA 20"},
							},
						},
					},
				},
			},
			wantContains: []string{},
		},
		{
			name: "plot with builtin series",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "plot"},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "equity"},
					},
				},
			},
			wantContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := handler.GenerateCode(g, tt.call)
			if err != nil {
				t.Errorf("GenerateCode() unexpected error: %v", err)
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("GenerateCode() code = %q, want to contain %q", code, want)
				}
			}
		})
	}
}

// TestPlotFunctionHandler_EmptyArguments tests edge case with no arguments
func TestPlotFunctionHandler_EmptyArguments(t *testing.T) {
	handler := &PlotFunctionHandler{}
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "plot"},
		Arguments: []ast.Expression{},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Errorf("GenerateCode() unexpected error: %v", err)
	}

	// Should handle gracefully - no plot expression, no code
	if code != "" {
		t.Errorf("GenerateCode() with no arguments should return empty, got: %q", code)
	}
}

// TestPlotFunctionHandler_BuiltinResolution verifies strategy.equity is resolved
func TestPlotFunctionHandler_BuiltinResolution(t *testing.T) {
	// Integration test: plot(strategy.equity) should resolve via builtin handler
	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "strategy"},
					Arguments: []ast.Expression{
						&ast.Literal{Value: "Test"},
					},
				},
			},
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "plot"},
					Arguments: []ast.Expression{
						&ast.MemberExpression{
							Object:   &ast.Identifier{Name: "strategy"},
							Property: &ast.Identifier{Name: "equity"},
						},
						&ast.ObjectExpression{
							Properties: []ast.Property{
								{
									Key:   &ast.Identifier{Name: "title"},
									Value: &ast.Literal{Value: "Equity"},
								},
							},
						},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("GenerateStrategyCodeFromAST() error: %v", err)
	}

	// Should generate strategy_equitySeries.Get(0), not strategySeries.Get(0)
	if strings.Contains(code.FunctionBody, "strategySeries.Get(0)") {
		t.Error("Plot should resolve strategy.equity to strategy_equitySeries, not strategySeries")
	}

	if !strings.Contains(code.FunctionBody, "strategy_equitySeries") {
		t.Error("Plot should generate strategy_equitySeries reference")
	}
}

// TestPlotFunctionHandler_ComplexExpressions tests plot with calculations
func TestPlotFunctionHandler_ComplexExpressions(t *testing.T) {
	handler := &PlotFunctionHandler{}
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "plot"},
		Arguments: []ast.Expression{
			&ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "close"},
				Operator: "*",
				Right:    &ast.Literal{Value: 1.1},
			},
		},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Errorf("GenerateCode() unexpected error: %v", err)
	}

	// Should handle expression (delegated to plotCollector)
	_ = code // No immediate code, added to plotCollector
}

// TestPlotFunctionHandler_UniqueTitleGeneration verifies unique titles for untitled plots
func TestPlotFunctionHandler_UniqueTitleGeneration(t *testing.T) {
	handler := &PlotFunctionHandler{}
	g := newTestGenerator()

	// Plots with variable names should use variable name as title
	call1 := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "plot"},
		Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
	}

	_, err := handler.GenerateCode(g, call1)
	if err != nil {
		t.Fatalf("GenerateCode() error on first call: %v", err)
	}

	call2 := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "plot"},
		Arguments: []ast.Expression{&ast.Identifier{Name: "open"}},
	}

	_, err = handler.GenerateCode(g, call2)
	if err != nil {
		t.Fatalf("GenerateCode() error on second call: %v", err)
	}

	plots := g.plotCollector.GetPlots()
	if len(plots) != 2 {
		t.Fatalf("Expected 2 plots, got %d", len(plots))
	}

	// Variable names used as titles
	if !strings.Contains(plots[0].code, `"close"`) {
		t.Errorf("Expected first plot code to contain 'close', got %q", plots[0].code)
	}
	if !strings.Contains(plots[1].code, `"open"`) {
		t.Errorf("Expected second plot code to contain 'open', got %q", plots[1].code)
	}
}

// TestPlotFunctionHandler_GeneratedTitleForComplexExpr verifies generated titles for complex expressions
func TestPlotFunctionHandler_GeneratedTitleForComplexExpr(t *testing.T) {
	handler := &PlotFunctionHandler{}
	g := newTestGenerator()

	// Complex expressions should generate "Plot N" since extractPlotVariable returns ""
	call1 := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "plot"},
		Arguments: []ast.Expression{
			&ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "close"},
				Operator: "+",
				Right:    &ast.Literal{Value: 10.0},
			},
		},
	}

	_, err := handler.GenerateCode(g, call1)
	if err != nil {
		t.Fatalf("GenerateCode() error: %v", err)
	}

	call2 := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "plot"},
		Arguments: []ast.Expression{
			&ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "high"},
				Operator: "-",
				Right:    &ast.Identifier{Name: "low"},
			},
		},
	}

	_, err = handler.GenerateCode(g, call2)
	if err != nil {
		t.Fatalf("GenerateCode() error: %v", err)
	}

	plots := g.plotCollector.GetPlots()
	if len(plots) != 2 {
		t.Fatalf("Expected 2 plots, got %d", len(plots))
	}

	// Generated titles for complex expressions
	if !strings.Contains(plots[0].code, `"Plot 1"`) {
		t.Errorf("Expected first plot code to contain 'Plot 1', got %q", plots[0].code)
	}
	if !strings.Contains(plots[1].code, `"Plot 2"`) {
		t.Errorf("Expected second plot code to contain 'Plot 2', got %q", plots[1].code)
	}
}

// TestPlotFunctionHandler_ExplicitTitlePreserved verifies explicit titles not overwritten
func TestPlotFunctionHandler_ExplicitTitlePreserved(t *testing.T) {
	handler := &PlotFunctionHandler{}
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "plot"},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "sma20"},
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key:   &ast.Identifier{Name: "title"},
						Value: &ast.Literal{Value: "SMA 20"},
					},
				},
			},
		},
	}

	_, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Fatalf("GenerateCode() error: %v", err)
	}

	plots := g.plotCollector.GetPlots()
	if len(plots) != 1 {
		t.Fatalf("Expected 1 plot, got %d", len(plots))
	}

	if !strings.Contains(plots[0].code, `"SMA 20"`) {
		t.Errorf("Expected code to contain 'SMA 20', got %q", plots[0].code)
	}
}

// TestPlotFunctionHandler_MixedTitles verifies mixed explicit and generated titles
func TestPlotFunctionHandler_MixedTitles(t *testing.T) {
	handler := &PlotFunctionHandler{}
	g := newTestGenerator()

	calls := []*ast.CallExpression{
		// Explicit title
		{
			Callee: &ast.Identifier{Name: "plot"},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "sma20"},
				&ast.ObjectExpression{
					Properties: []ast.Property{
						{Key: &ast.Identifier{Name: "title"}, Value: &ast.Literal{Value: "SMA 20"}},
					},
				},
			},
		},
		// No title, simple variable - should use variable name "close"
		{
			Callee:    &ast.Identifier{Name: "plot"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		},
		// Explicit title
		{
			Callee: &ast.Identifier{Name: "plot"},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "ema50"},
				&ast.ObjectExpression{
					Properties: []ast.Property{
						{Key: &ast.Identifier{Name: "title"}, Value: &ast.Literal{Value: "EMA 50"}},
					},
				},
			},
		},
		// No title, complex expression - should generate "Plot 4" (total plot count)
		{
			Callee: &ast.Identifier{Name: "plot"},
			Arguments: []ast.Expression{
				&ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "high"},
					Operator: "-",
					Right:    &ast.Identifier{Name: "low"},
				},
			},
		},
	}

	for _, call := range calls {
		_, err := handler.GenerateCode(g, call)
		if err != nil {
			t.Fatalf("GenerateCode() error: %v", err)
		}
	}

	plots := g.plotCollector.GetPlots()
	if len(plots) != 4 {
		t.Fatalf("Expected 4 plots, got %d", len(plots))
	}

	expectedTitles := []string{`"SMA 20"`, `"close"`, `"EMA 50"`, `"Plot 4"`}
	for i, expectedTitle := range expectedTitles {
		if !strings.Contains(plots[i].code, expectedTitle) {
			t.Errorf("Plot %d: expected code to contain %s, got %q", i+1, expectedTitle, plots[i].code)
		}
	}
}
