package codegen

import (
	"fmt"
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
		// action functions
		{"strategy.entry", true},
		{"strategy.close", true},
		{"strategy.close_all", true},
		{"strategy.exit", true},
		{"strategy.order", true},
		{"strategy.cancel", true},
		{"strategy.cancel_all", true},
		// risk management functions
		{"strategy.risk.allow_entry_in", true},
		{"strategy.risk.max_drawdown", true},
		{"strategy.risk.max_cons_loss_days", true},
		{"strategy.risk.max_intraday_filled_orders", true},
		{"strategy.risk.max_intraday_loss", true},
		{"strategy.risk.max_position_size", true},
		// non-strategy calls
		{"strategy", false},
		{"ta.entry", false},
		{"entry", false},
		{"strategy.risk", false},
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
			code, err := handler.GenerateCode(g, tt.call)
			if err != nil {
				t.Errorf("GenerateCode() unexpected error: %v", err)
			}
			_ = code
		})
	}
}

// TestStrategyExit_NamedArguments verifies named stop/limit args are extracted and not defaulted to NaN
func TestStrategyExit_NamedArguments(t *testing.T) {
	handler := NewStrategyActionHandler()
	g := newTestGenerator()

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

	if !strings.Contains(code, "95") {
		t.Errorf("Expected stop value 95 in generated code, got:\n%s", code)
	}
	if !strings.Contains(code, "110") {
		t.Errorf("Expected limit value 110 in generated code, got:\n%s", code)
	}
	if strings.Contains(code, "math.NaN()") {
		t.Errorf("Should not contain math.NaN() when named args provided, got:\n%s", code)
	}
}

// TestStrategyExit_NamedVariables verifies identifier stop/limit args resolve to series accessors
func TestStrategyExit_NamedVariables(t *testing.T) {
	handler := NewStrategyActionHandler()
	g := newTestGenerator()
	g.variables["stop_level"] = "float64"
	g.variables["limit_level"] = "float64"

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

	if !strings.Contains(code, "stop_levelSeries.GetCurrent()") {
		t.Errorf("Expected stop_levelSeries.GetCurrent() in code, got:\n%s", code)
	}
	if !strings.Contains(code, "limit_levelSeries.GetCurrent()") {
		t.Errorf("Expected limit_levelSeries.GetCurrent() in code, got:\n%s", code)
	}
}

// TestStrategyExit_OnlyStop verifies absent limit defaults to math.NaN()
func TestStrategyExit_OnlyStop(t *testing.T) {
	handler := NewStrategyActionHandler()
	g := newTestGenerator()

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

	if !strings.Contains(code, "95") {
		t.Errorf("Expected stop value 95, got:\n%s", code)
	}
	if !strings.Contains(code, "math.NaN()") {
		t.Errorf("Expected limit=math.NaN() when not provided, got:\n%s", code)
	}
}

// TestStrategyFunctionQuantityCalculation verifies qty calculation for strategy.entry and strategy.order
// across all default_qty_type values, asserting each emits only its own runtime call and variable name.
func TestStrategyFunctionQuantityCalculation(t *testing.T) {
	type stratFn struct {
		funcName    string
		emitMethod  string
		qtyVarName  string
		otherMethod string
	}
	funcs := []stratFn{
		{"entry", "strat.Entry", "entryQty", "strat.Order"},
		{"order", "strat.Order", "orderQty", "strat.Entry"},
	}

	type qtyCase struct {
		name           string
		defaultQtyType string
		defaultQtyVal  float64
		dynamicExpr    string // non-empty for cash/percent_of_equity types; uses %s for qtyVarName
		wantLiteralQty string // non-empty for fixed types; e.g. "100,"
		wantWarning    bool
	}
	qtyCases := []qtyCase{
		{
			name:           "strategy.cash generates runtime division",
			defaultQtyType: "strategy.cash",
			defaultQtyVal:  600000.0,
			dynamicExpr:    "%s := 600000 / closeSeries.GetCurrent()",
		},
		{
			name:           "cash unprefixed generates runtime division",
			defaultQtyType: "cash",
			defaultQtyVal:  50000.0,
			dynamicExpr:    "%s := 50000 / closeSeries.GetCurrent()",
		},
		{
			name:           "strategy.percent_of_equity generates equity percentage",
			defaultQtyType: "strategy.percent_of_equity",
			defaultQtyVal:  10.0,
			dynamicExpr:    "%s := (strat.Equity() * 10.00 / 100) / closeSeries.GetCurrent()",
		},
		{
			name:           "percent_of_equity unprefixed generates equity percentage",
			defaultQtyType: "percent_of_equity",
			defaultQtyVal:  25.5,
			dynamicExpr:    "%s := (strat.Equity() * 25.50 / 100) / closeSeries.GetCurrent()",
		},
		{
			name:           "strategy.fixed uses qty directly",
			defaultQtyType: "strategy.fixed",
			defaultQtyVal:  100.0,
			wantLiteralQty: "100,",
		},
		{
			name:           "fixed unprefixed uses qty directly",
			defaultQtyType: "fixed",
			defaultQtyVal:  50.0,
			wantLiteralQty: "50,",
		},
		{
			name:           "empty string defaults to fixed",
			defaultQtyType: "",
			defaultQtyVal:  75.0,
			wantLiteralQty: "75,",
		},
		{
			name:           "unknown type uses fixed with warning",
			defaultQtyType: "invalid_type",
			defaultQtyVal:  123.0,
			wantLiteralQty: "123,",
			wantWarning:    true,
		},
	}

	handler := NewStrategyActionHandler()

	for _, fn := range funcs {
		for _, qc := range qtyCases {
			t.Run(fn.funcName+"/"+qc.name, func(t *testing.T) {
				g := newTestGenerator()
				g.strategyConfig.DefaultQtyType = qc.defaultQtyType
				g.strategyConfig.DefaultQtyValue = qc.defaultQtyVal

				call := &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: fn.funcName},
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

				if !strings.Contains(code, fn.emitMethod) {
					t.Errorf("Expected %q in output, got:\n%s", fn.emitMethod, code)
				}
				if strings.Contains(code, fn.otherMethod) {
					t.Errorf("Must NOT contain %q in output, got:\n%s", fn.otherMethod, code)
				}

				if qc.dynamicExpr != "" {
					want := fmt.Sprintf(qc.dynamicExpr, fn.qtyVarName)
					if !strings.Contains(code, want) {
						t.Errorf("Expected dynamic expr %q, got:\n%s", want, code)
					}
					if strings.Contains(code, fn.qtyVarName+" :=") && !strings.Contains(code, want) {
						t.Errorf("Unexpected assignment form, got:\n%s", code)
					}
				} else {
					if strings.Contains(code, fn.qtyVarName+" :=") {
						t.Errorf("Must NOT emit qty variable assignment for fixed type, got:\n%s", code)
					}
					if !strings.Contains(code, qc.wantLiteralQty) {
						t.Errorf("Expected literal qty %q in output, got:\n%s", qc.wantLiteralQty, code)
					}
				}

				if qc.wantWarning {
					wantWarn := fmt.Sprintf("// WARNING: Unknown default_qty_type '%s'", qc.defaultQtyType)
					if !strings.Contains(code, wantWarn) {
						t.Errorf("Expected warning comment %q, got:\n%s", wantWarn, code)
					}
				}
			})
		}
	}
}

// TestStrategyOrder_EmitsOrderNotEntry verifies strategy.order emits strat.Order, not strat.Entry
func TestStrategyOrder_EmitsOrderNotEntry(t *testing.T) {
	handler := NewStrategyActionHandler()
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Identifier{Name: "order"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "Sell"},
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "short"},
			},
		},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Fatalf("GenerateCode() error: %v", err)
	}
	if !strings.Contains(code, "strat.Order") {
		t.Errorf("Expected strat.Order in generated code, got:\n%s", code)
	}
	if strings.Contains(code, "strat.Entry") {
		t.Errorf("strat.Entry must NOT appear in strategy.order output, got:\n%s", code)
	}
}

// TestStrategyOrder_InvalidArgs verifies fewer than 2 args emits a comment stub
func TestStrategyOrder_InvalidArgs(t *testing.T) {
	handler := NewStrategyActionHandler()
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Identifier{Name: "order"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "only_one_arg"},
		},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Fatalf("GenerateCode() error: %v", err)
	}
	if !strings.Contains(code, "// strategy.order()") {
		t.Errorf("Expected comment stub for invalid args, got:\n%s", code)
	}
}

// TestStrategyOrder_WhenCondition verifies when= wraps strategy.order in a conditional
func TestStrategyOrder_WhenCondition(t *testing.T) {
	handler := NewStrategyActionHandler()
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Identifier{Name: "order"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "Buy"},
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "long"},
			},
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{Key: &ast.Identifier{Name: "when"}, Value: &ast.Literal{Value: true}},
				},
			},
		},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Fatalf("GenerateCode() error: %v", err)
	}
	if !strings.Contains(code, "if ") {
		t.Errorf("Expected if-wrapper for when= condition, got:\n%s", code)
	}
	if !strings.Contains(code, "strat.Order") {
		t.Errorf("Expected strat.Order inside when wrapper, got:\n%s", code)
	}
}

// TestStrategyCancel_CodeGen verifies strategy.cancel emits strat.Cancel with the order ID
func TestStrategyCancel_CodeGen(t *testing.T) {
	handler := NewStrategyActionHandler()
	g := newTestGenerator()

	tests := []struct {
		name         string
		args         []ast.Expression
		wantContains []string
	}{
		{
			name:         "valid: emits strat.Cancel with ID",
			args:         []ast.Expression{&ast.Literal{Value: "myOrder"}},
			wantContains: []string{`strat.Cancel("myOrder")`},
		},
		{
			name:         "invalid: no args emits comment stub",
			args:         []ast.Expression{},
			wantContains: []string{"// strategy.cancel()"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "cancel"},
				},
				Arguments: tt.args,
			}
			code, err := handler.GenerateCode(g, call)
			if err != nil {
				t.Fatalf("GenerateCode() error: %v", err)
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("Expected %q in code, got:\n%s", want, code)
				}
			}
		})
	}
}

// TestStrategyCancel_WhenCondition verifies when= wraps strategy.cancel in a conditional
func TestStrategyCancel_WhenCondition(t *testing.T) {
	handler := NewStrategyActionHandler()
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Identifier{Name: "cancel"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "myOrder"},
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{Key: &ast.Identifier{Name: "when"}, Value: &ast.Literal{Value: true}},
				},
			},
		},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Fatalf("GenerateCode() error: %v", err)
	}
	if !strings.Contains(code, "if ") {
		t.Errorf("Expected if-wrapper for when= condition, got:\n%s", code)
	}
	if !strings.Contains(code, "strat.Cancel") {
		t.Errorf("Expected strat.Cancel inside when wrapper, got:\n%s", code)
	}
}

// TestStrategyCancelAll_CodeGen verifies strategy.cancel_all emits strat.CancelAll()
func TestStrategyCancelAll_CodeGen(t *testing.T) {
	handler := NewStrategyActionHandler()
	g := newTestGenerator()

	tests := []struct {
		name string
		args []ast.Expression
		want string
	}{
		{
			name: "no args emits strat.CancelAll()",
			args: []ast.Expression{},
			want: "strat.CancelAll()",
		},
		{
			name: "extra args are ignored",
			args: []ast.Expression{&ast.Literal{Value: "ignored"}},
			want: "strat.CancelAll()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "cancel_all"},
				},
				Arguments: tt.args,
			}
			code, err := handler.GenerateCode(g, call)
			if err != nil {
				t.Fatalf("GenerateCode() error: %v", err)
			}
			if !strings.Contains(code, tt.want) {
				t.Errorf("Expected %q in code, got:\n%s", tt.want, code)
			}
		})
	}
}

// TestStrategyCancelAll_WhenCondition verifies when= wraps strategy.cancel_all in a conditional
func TestStrategyCancelAll_WhenCondition(t *testing.T) {
	handler := NewStrategyActionHandler()
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Identifier{Name: "cancel_all"},
		},
		Arguments: []ast.Expression{
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{Key: &ast.Identifier{Name: "when"}, Value: &ast.Literal{Value: true}},
				},
			},
		},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Fatalf("GenerateCode() error: %v", err)
	}
	if !strings.Contains(code, "if ") {
		t.Errorf("Expected if-wrapper for when= condition, got:\n%s", code)
	}
	if !strings.Contains(code, "strat.CancelAll()") {
		t.Errorf("Expected strat.CancelAll() inside when wrapper, got:\n%s", code)
	}
}

// TestStrategyAllowEntryIn_CodeGen verifies strategy.risk.allow_entry_in emits strat.SetAllowedDirection
func TestStrategyAllowEntryIn_CodeGen(t *testing.T) {
	handler := NewStrategyActionHandler()
	g := newTestGenerator()

	tests := []struct {
		name         string
		args         []ast.Expression
		wantContains []string
	}{
		{
			name: "long direction emits SetAllowedDirection",
			args: []ast.Expression{
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "long"},
				},
			},
			wantContains: []string{"strat.SetAllowedDirection", "strategy.Long"},
		},
		{
			name: "short direction emits SetAllowedDirection",
			args: []ast.Expression{
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "short"},
				},
			},
			wantContains: []string{"strat.SetAllowedDirection", "strategy.Short"},
		},
		{
			name:         "no args emits comment stub",
			args:         []ast.Expression{},
			wantContains: []string{"// strategy.risk.allow_entry_in()"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "risk"},
					},
					Property: &ast.Identifier{Name: "allow_entry_in"},
				},
				Arguments: tt.args,
			}
			code, err := handler.GenerateCode(g, call)
			if err != nil {
				t.Fatalf("GenerateCode() error: %v", err)
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("Expected %q in code, got:\n%s", want, code)
				}
			}
		})
	}
}

// TestStrategyRiskNoOps verifies deprecated strategy.risk.* limits emit no Go code
func TestStrategyRiskNoOps(t *testing.T) {
	handler := NewStrategyActionHandler()
	g := newTestGenerator()

	noOpFuncs := []string{
		"strategy.risk.max_drawdown",
		"strategy.risk.max_cons_loss_days",
		"strategy.risk.max_intraday_filled_orders",
		"strategy.risk.max_intraday_loss",
		"strategy.risk.max_position_size",
	}

	for _, funcName := range noOpFuncs {
		t.Run(funcName, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "risk"},
					},
					Property: &ast.Identifier{Name: funcName[len("strategy.risk."):]},
				},
				Arguments: []ast.Expression{&ast.Literal{Value: 1000.0}},
			}
			code, err := handler.GenerateCode(g, call)
			if err != nil {
				t.Fatalf("%s: GenerateCode() error: %v", funcName, err)
			}
			if code != "" {
				t.Errorf("%s: expected empty code (no-op), got:\n%s", funcName, code)
			}
		})
	}
}
