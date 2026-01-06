package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// TestStrategyActionHandler_CanHandle verifies strategy action recognition
func TestStrategyActionHandler_CanHandle(t *testing.T) {
	handler := &StrategyActionHandler{}

	tests := []struct {
		funcName string
		want     bool
	}{
		{"strategy.entry", true},
		{"strategy.close", true},
		{"strategy.close_all", true},
		{"strategy.exit", true},
		{"strategy", false},
		{"ta.entry", false},
		{"entry", false},
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

// TestStrategyActionHandler_EntryValidCases verifies correct entry code generation
func TestStrategyActionHandler_EntryValidCases(t *testing.T) {
	handler := &StrategyActionHandler{}
	g := newTestGenerator()

	tests := []struct {
		name         string
		call         *ast.CallExpression
		wantContains []string
	}{
		{
			name: "entry with 2 args",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Buy"},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "long"},
					},
				},
			},
			wantContains: []string{"strat.Entry", `"Buy"`, "strategy.Long"},
		},
		{
			name: "entry with 3 args (quantity)",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Sell"},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "short"},
					},
					&ast.Literal{Value: 2.0},
				},
			},
			wantContains: []string{"strat.Entry", `"Sell"`, "strategy.Short", "2"},
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
					t.Errorf("GenerateCode() = %q, want to contain %q", code, want)
				}
			}
		})
	}
}

// TestStrategyActionHandler_EntryInvalidArgs verifies graceful handling of invalid entry args
func TestStrategyActionHandler_EntryInvalidArgs(t *testing.T) {
	handler := &StrategyActionHandler{}
	g := newTestGenerator()

	tests := []struct {
		name         string
		call         *ast.CallExpression
		wantContains string
	}{
		{
			name: "no arguments",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
				Arguments: []ast.Expression{},
			},
			wantContains: "// strategy.entry() - invalid arguments",
		},
		{
			name: "one argument",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Buy"},
				},
			},
			wantContains: "// strategy.entry() - invalid arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := handler.GenerateCode(g, tt.call)
			if err != nil {
				t.Errorf("GenerateCode() unexpected error: %v", err)
			}

			if !strings.Contains(code, tt.wantContains) {
				t.Errorf("GenerateCode() = %q, want to contain %q", code, tt.wantContains)
			}
		})
	}
}

// TestStrategyActionHandler_CloseValidCases verifies correct close code generation
func TestStrategyActionHandler_CloseValidCases(t *testing.T) {
	handler := &StrategyActionHandler{}
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Identifier{Name: "close"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "Buy"},
		},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Errorf("GenerateCode() unexpected error: %v", err)
	}

	wantContains := []string{"strat.Close", `"Buy"`, "bar.Close", "bar.Time"}
	for _, want := range wantContains {
		if !strings.Contains(code, want) {
			t.Errorf("GenerateCode() = %q, want to contain %q", code, want)
		}
	}
}

// TestStrategyActionHandler_CloseInvalidArgs verifies graceful handling of invalid close args
func TestStrategyActionHandler_CloseInvalidArgs(t *testing.T) {
	handler := &StrategyActionHandler{}
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Identifier{Name: "close"},
		},
		Arguments: []ast.Expression{},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Errorf("GenerateCode() unexpected error: %v", err)
	}

	if !strings.Contains(code, "// strategy.close() - invalid arguments") {
		t.Errorf("GenerateCode() = %q, want TODO comment for invalid args", code)
	}
}

// TestStrategyActionHandler_CloseAll verifies close_all code generation
func TestStrategyActionHandler_CloseAll(t *testing.T) {
	handler := &StrategyActionHandler{}
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Identifier{Name: "close_all"},
		},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Errorf("GenerateCode() unexpected error: %v", err)
	}

	wantContains := []string{"strat.CloseAll", "bar.Close", "bar.Time"}
	for _, want := range wantContains {
		if !strings.Contains(code, want) {
			t.Errorf("GenerateCode() = %q, want to contain %q", code, want)
		}
	}
}

// TestStrategyActionHandler_IntegrationWithGenerator tests strategy actions in full pipeline
func TestStrategyActionHandler_IntegrationWithGenerator(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "signal"},
						Init: &ast.Literal{Value: 1.0},
					},
				},
			},
			&ast.IfStatement{
				Test: &ast.Identifier{Name: "signal"},
				Consequent: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "strategy"},
								Property: &ast.Identifier{Name: "entry"},
							},
							Arguments: []ast.Expression{
								&ast.Literal{Value: "Long"},
								&ast.MemberExpression{
									Object:   &ast.Identifier{Name: "strategy"},
									Property: &ast.Identifier{Name: "long"},
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

	// Should generate strat.Entry call inside if statement
	if !strings.Contains(code.FunctionBody, "strat.Entry") {
		t.Error("Expected strat.Entry call in generated code")
	}

	if !strings.Contains(code.FunctionBody, "if value.IsTrue(signalSeries.GetCurrent())") {
		t.Errorf("Expected if statement in generated code. Got:\n%s", code.FunctionBody)
	}
}

// TestStrategyActionHandler_EdgeCases tests unusual but valid scenarios
func TestStrategyActionHandler_EdgeCases(t *testing.T) {
	handler := &StrategyActionHandler{}
	g := newTestGenerator()

	tests := []struct {
		name string
		call *ast.CallExpression
	}{
		{
			name: "entry with expression as ID",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
				Arguments: []ast.Expression{
					&ast.BinaryExpression{
						Left:     &ast.Literal{Value: "Buy"},
						Operator: "+",
						Right:    &ast.Literal{Value: "1"},
					},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "long"},
					},
				},
			},
		},
		{
			name: "close_all with extra arguments (ignored)",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "close_all"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "ignored"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			code, err := handler.GenerateCode(g, tt.call)
			if err != nil {
				t.Errorf("GenerateCode() unexpected error: %v", err)
			}
			_ = code
		})
	}
}

/* Test strategy.exit() with named arguments */
func TestStrategyExit_NamedArguments(t *testing.T) {
	handler := NewStrategyActionHandler()
	g := newTestGenerator()

	/* strategy.exit("Exit", "Long", stop=95.0, limit=110.0) */
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Literal{Value: "Exit"},
			&ast.Literal{Value: "Long"},
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{Key: &ast.Identifier{Name: "stop"}, Value: &ast.Literal{Value: 95.0}},
					{Key: &ast.Identifier{Name: "limit"}, Value: &ast.Literal{Value: 110.0}},
				},
			},
		},
	}

	code, err := handler.generateExit(g, call)
	if err != nil {
		t.Fatalf("generateExit failed: %v", err)
	}

	/* Verify stop and limit extracted correctly (not NaN) */
	if !strings.Contains(code, "95.00") {
		t.Errorf("Expected stop value 95.00 in generated code, got:\n%s", code)
	}
	if !strings.Contains(code, "110.00") {
		t.Errorf("Expected limit value 110.00 in generated code, got:\n%s", code)
	}
	if strings.Contains(code, "math.NaN()") {
		t.Errorf("Should not contain math.NaN() when named args provided, got:\n%s", code)
	}
}

/* Test with identifier variables */
func TestStrategyExit_NamedVariables(t *testing.T) {
	handler := NewStrategyActionHandler()
	g := newTestGenerator()
	g.variables["stop_level"] = "float64"
	g.variables["limit_level"] = "float64"

	/* strategy.exit("Exit", "Long", stop=stop_level, limit=limit_level) */
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Literal{Value: "Exit"},
			&ast.Literal{Value: "Long"},
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{Key: &ast.Identifier{Name: "stop"}, Value: &ast.Identifier{Name: "stop_level"}},
					{Key: &ast.Identifier{Name: "limit"}, Value: &ast.Identifier{Name: "limit_level"}},
				},
			},
		},
	}

	code, err := handler.generateExit(g, call)
	if err != nil {
		t.Fatalf("generateExit failed: %v", err)
	}

	/* Verify series access generated */
	if !strings.Contains(code, "stop_levelSeries.GetCurrent()") {
		t.Errorf("Expected stop_levelSeries.GetCurrent() in code, got:\n%s", code)
	}
	if !strings.Contains(code, "limit_levelSeries.GetCurrent()") {
		t.Errorf("Expected limit_levelSeries.GetCurrent() in code, got:\n%s", code)
	}
}

/* Test with only stop (no limit) */
func TestStrategyExit_OnlyStop(t *testing.T) {
	handler := NewStrategyActionHandler()
	g := newTestGenerator()

	/* strategy.exit("Exit", "Long", stop=95.0) */
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Literal{Value: "Exit"},
			&ast.Literal{Value: "Long"},
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{Key: &ast.Identifier{Name: "stop"}, Value: &ast.Literal{Value: 95.0}},
				},
			},
		},
	}

	code, err := handler.generateExit(g, call)
	if err != nil {
		t.Fatalf("generateExit failed: %v", err)
	}

	/* stop=95.0, limit=NaN */
	if !strings.Contains(code, "95.00") {
		t.Errorf("Expected stop value 95.00, got:\n%s", code)
	}
	/* Limit should be NaN (not provided) */
	if !strings.Contains(code, "math.NaN()") {
		t.Errorf("Expected limit=math.NaN() when not provided, got:\n%s", code)
	}
}

/* TestStrategyEntry_QuantityCalculation verifies runtime qty calculation based on default_qty_type */
func TestStrategyEntry_QuantityCalculation(t *testing.T) {
	handler := NewStrategyActionHandler()

	tests := []struct {
		name           string
		defaultQtyType string
		defaultQtyVal  float64
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:           "strategy.cash generates runtime division",
			defaultQtyType: "strategy.cash",
			defaultQtyVal:  600000.0,
			wantContains: []string{
				"entryQty := 600000 / closeSeries.GetCurrent()",
				"strat.Entry",
				"entryQty",
			},
			wantNotContain: []string{
				"600000,",
			},
		},
		{
			name:           "cash unprefixed generates runtime division",
			defaultQtyType: "cash",
			defaultQtyVal:  50000.0,
			wantContains: []string{
				"entryQty := 50000 / closeSeries.GetCurrent()",
				"strat.Entry",
				"entryQty",
			},
			wantNotContain: []string{
				"50000,",
			},
		},
		{
			name:           "strategy.percent_of_equity generates equity percentage",
			defaultQtyType: "strategy.percent_of_equity",
			defaultQtyVal:  10.0,
			wantContains: []string{
				"entryQty := (strat.Equity() * 10.00 / 100) / closeSeries.GetCurrent()",
				"strat.Entry",
				"entryQty",
			},
			wantNotContain: []string{
				"10,",
			},
		},
		{
			name:           "percent_of_equity unprefixed generates equity percentage",
			defaultQtyType: "percent_of_equity",
			defaultQtyVal:  25.5,
			wantContains: []string{
				"entryQty := (strat.Equity() * 25.50 / 100) / closeSeries.GetCurrent()",
				"strat.Entry",
				"entryQty",
			},
			wantNotContain: []string{
				"25.50,",
			},
		},
		{
			name:           "strategy.fixed uses qty directly",
			defaultQtyType: "strategy.fixed",
			defaultQtyVal:  100.0,
			wantContains: []string{
				"strat.Entry",
				"100,",
			},
			wantNotContain: []string{
				"entryQty :=",
				"GetCurrent()",
			},
		},
		{
			name:           "fixed unprefixed uses qty directly",
			defaultQtyType: "fixed",
			defaultQtyVal:  50.0,
			wantContains: []string{
				"strat.Entry",
				"50,",
			},
			wantNotContain: []string{
				"entryQty :=",
			},
		},
		{
			name:           "empty string defaults to fixed",
			defaultQtyType: "",
			defaultQtyVal:  75.0,
			wantContains: []string{
				"strat.Entry",
				"75,",
			},
			wantNotContain: []string{
				"entryQty :=",
			},
		},
		{
			name:           "unknown type uses fixed with warning",
			defaultQtyType: "invalid_type",
			defaultQtyVal:  123.0,
			wantContains: []string{
				"// WARNING: Unknown default_qty_type 'invalid_type'",
				"strat.Entry",
				"123,",
			},
			wantNotContain: []string{
				"entryQty :=",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			g.strategyConfig.DefaultQtyType = tt.defaultQtyType
			g.strategyConfig.DefaultQtyValue = tt.defaultQtyVal

			// strategy.entry("Buy", strategy.long)
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Buy"},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "long"},
					},
				},
			}

			code, err := handler.GenerateCode(g, call)
			if err != nil {
				t.Fatalf("GenerateCode failed: %v", err)
			}

			// Check expected strings are present
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("Expected code to contain %q, got:\n%s", want, code)
				}
			}

			// Check unwanted strings are absent
			for _, unwant := range tt.wantNotContain {
				if strings.Contains(code, unwant) {
					t.Errorf("Expected code NOT to contain %q, but it does:\n%s", unwant, code)
				}
			}
		})
	}
}

