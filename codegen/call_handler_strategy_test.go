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

// TestStrategyExit_CallShapes verifies ID, from_entry, stop, and limit extraction for all three
// Pine call shapes across a representative set of level combinations.
//
// Shape A: strategy.exit(id="X", from_entry="Y", stop=s, limit=l)  — single ObjectExpression
// Shape B: strategy.exit("X", stop=s, limit=l)                     — positional id, no from_entry
// Shape C: strategy.exit("X", "Y", stop=s, limit=l)                — all positional + named ObjExpr
func TestStrategyExit_CallShapes(t *testing.T) {
	namedObjExpr := func(props ...ast.Property) *ast.ObjectExpression {
		return &ast.ObjectExpression{Properties: props}
	}
	prop := func(key string, val ast.Expression) ast.Property {
		return ast.Property{Key: &ast.Identifier{Name: key}, Value: val}
	}
	lit := func(v float64) ast.Expression { return &ast.Literal{Value: v} }

	tests := []struct {
		name      string
		args      []ast.Expression
		wantID    string
		wantEntry string
		wantStop  string
		wantLimit string
		wantNoNaN bool
	}{
		{
			name: "shape A: id, from_entry, stop and limit",
			args: []ast.Expression{namedObjExpr(
				prop("id", &ast.Literal{Value: "X"}),
				prop("from_entry", &ast.Literal{Value: "Long"}),
				prop("stop", lit(90)),
				prop("limit", lit(110)),
			)},
			wantID: "X", wantEntry: "Long", wantStop: "90", wantLimit: "110", wantNoNaN: true,
		},
		{
			name: "shape A: stop only, limit absent",
			args: []ast.Expression{namedObjExpr(
				prop("id", &ast.Literal{Value: "StopOnly"}),
				prop("stop", lit(85)),
			)},
			wantID: "StopOnly", wantEntry: "", wantStop: "85", wantLimit: "math.NaN()",
		},
		{
			name: "shape A: limit only, stop absent",
			args: []ast.Expression{namedObjExpr(
				prop("id", &ast.Literal{Value: "LimitOnly"}),
				prop("limit", lit(120)),
			)},
			wantID: "LimitOnly", wantEntry: "", wantStop: "math.NaN()", wantLimit: "120",
		},
		{
			name: "shape B: positional id, no from_entry, stop and limit",
			args: []ast.Expression{
				&ast.Literal{Value: "TP/SL"},
				namedObjExpr(prop("stop", lit(230)), prop("limit", lit(260))),
			},
			wantID: "TP/SL", wantEntry: "", wantStop: "230", wantLimit: "260", wantNoNaN: true,
		},
		{
			name: "shape B: stop only, limit absent",
			args: []ast.Expression{
				&ast.Literal{Value: "SL"},
				namedObjExpr(prop("stop", lit(200))),
			},
			wantID: "SL", wantEntry: "", wantStop: "200", wantLimit: "math.NaN()",
		},
		{
			name: "shape B: limit only, stop absent",
			args: []ast.Expression{
				&ast.Literal{Value: "TP"},
				namedObjExpr(prop("limit", lit(300))),
			},
			wantID: "TP", wantEntry: "", wantStop: "math.NaN()", wantLimit: "300",
		},
		{
			name: "shape C: positional id and from_entry, stop and limit",
			args: []ast.Expression{
				&ast.Literal{Value: "Exit"},
				&ast.Literal{Value: "Long"},
				namedObjExpr(prop("stop", lit(95)), prop("limit", lit(110))),
			},
			wantID: "Exit", wantEntry: "Long", wantStop: "95", wantLimit: "110", wantNoNaN: true,
		},
		{
			name: "shape C: stop only, limit absent",
			args: []ast.Expression{
				&ast.Literal{Value: "Exit"},
				&ast.Literal{Value: "Short"},
				namedObjExpr(prop("stop", lit(50))),
			},
			wantID: "Exit", wantEntry: "Short", wantStop: "50", wantLimit: "math.NaN()",
		},
		{
			name: "shape C: limit only, stop absent",
			args: []ast.Expression{
				&ast.Literal{Value: "TakeProfit"},
				&ast.Literal{Value: "Long"},
				namedObjExpr(prop("limit", lit(180))),
			},
			wantID: "TakeProfit", wantEntry: "Long", wantStop: "math.NaN()", wantLimit: "180",
		},
	}

	handler := NewStrategyActionHandler()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			call := &ast.CallExpression{Arguments: tt.args}
			code, err := handler.generateExit(g, call)
			if err != nil {
				t.Fatalf("generateExit failed: %v", err)
			}
			if !strings.Contains(code, `"`+tt.wantID+`"`) {
				t.Errorf("id=%q not found in output:\n%s", tt.wantID, code)
			}
			if !strings.Contains(code, `"`+tt.wantEntry+`"`) {
				t.Errorf("from_entry=%q not found in output:\n%s", tt.wantEntry, code)
			}
			if !strings.Contains(code, tt.wantStop) {
				t.Errorf("stop=%q not found in output:\n%s", tt.wantStop, code)
			}
			if !strings.Contains(code, tt.wantLimit) {
				t.Errorf("limit=%q not found in output:\n%s", tt.wantLimit, code)
			}
			if tt.wantNoNaN && strings.Contains(code, "math.NaN()") {
				t.Errorf("unexpected math.NaN() when all levels provided:\n%s", code)
			}
		})
	}
}

// TestStrategyExit_VariableResolution verifies that identifier stop/limit arguments resolve
// to series accessor expressions in all three call shapes.
func TestStrategyExit_VariableResolution(t *testing.T) {
	namedObjExpr := func(props ...ast.Property) *ast.ObjectExpression {
		return &ast.ObjectExpression{Properties: props}
	}
	prop := func(key, varName string) ast.Property {
		return ast.Property{Key: &ast.Identifier{Name: key}, Value: &ast.Identifier{Name: varName}}
	}

	tests := []struct {
		name string
		args []ast.Expression
	}{
		{
			name: "shape A: variables in named ObjExpr",
			args: []ast.Expression{namedObjExpr(
				ast.Property{Key: &ast.Identifier{Name: "id"}, Value: &ast.Literal{Value: "X"}},
				prop("stop", "sl"),
				prop("limit", "tp"),
			)},
		},
		{
			name: "shape B: variables after positional id",
			args: []ast.Expression{
				&ast.Literal{Value: "X"},
				namedObjExpr(prop("stop", "sl"), prop("limit", "tp")),
			},
		},
		{
			name: "shape C: variables after positional id and from_entry",
			args: []ast.Expression{
				&ast.Literal{Value: "X"},
				&ast.Literal{Value: "Long"},
				namedObjExpr(prop("stop", "sl"), prop("limit", "tp")),
			},
		},
	}

	handler := NewStrategyActionHandler()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			g.variables["sl"] = "float64"
			g.variables["tp"] = "float64"
			call := &ast.CallExpression{Arguments: tt.args}
			code, err := handler.generateExit(g, call)
			if err != nil {
				t.Fatalf("generateExit failed: %v", err)
			}
			if !strings.Contains(code, "slSeries.GetCurrent()") {
				t.Errorf("expected slSeries.GetCurrent(); got:\n%s", code)
			}
			if !strings.Contains(code, "tpSeries.GetCurrent()") {
				t.Errorf("expected tpSeries.GetCurrent(); got:\n%s", code)
			}
			if strings.Contains(code, "math.NaN()") {
				t.Errorf("unexpected math.NaN() when variables supplied:\n%s", code)
			}
		})
	}
}

// TestStrategyExit_WhenCondition verifies that a when= named argument wraps the exit call in an
// if-block, and that the inner ExitWithLevels call is still emitted correctly.
func TestStrategyExit_WhenCondition(t *testing.T) {
	tests := []struct {
		name string
		args []ast.Expression
	}{
		{
			name: "shape B with when",
			args: []ast.Expression{
				&ast.Literal{Value: "Exit"},
				&ast.ObjectExpression{Properties: []ast.Property{
					{Key: &ast.Identifier{Name: "stop"}, Value: &ast.Literal{Value: 100.0}},
					{Key: &ast.Identifier{Name: "when"}, Value: &ast.Identifier{Name: "myCondition"}},
				}},
			},
		},
		{
			name: "shape C with when",
			args: []ast.Expression{
				&ast.Literal{Value: "Exit"},
				&ast.Literal{Value: "Long"},
				&ast.ObjectExpression{Properties: []ast.Property{
					{Key: &ast.Identifier{Name: "stop"}, Value: &ast.Literal{Value: 100.0}},
					{Key: &ast.Identifier{Name: "when"}, Value: &ast.Identifier{Name: "myCondition"}},
				}},
			},
		},
	}

	handler := NewStrategyActionHandler()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			g.variables["myCondition"] = "bool"
			call := &ast.CallExpression{Arguments: tt.args}
			code, err := handler.generateExit(g, call)
			if err != nil {
				t.Fatalf("generateExit failed: %v", err)
			}
			if !strings.Contains(code, "if ") {
				t.Errorf("expected if-wrapper for when= condition; got:\n%s", code)
			}
			if !strings.Contains(code, "ExitWithLevels") {
				t.Errorf("expected ExitWithLevels inside when wrapper; got:\n%s", code)
			}
		})
	}
}

// TestStrategyExit_InvalidArgs verifies graceful degradation when arguments are structurally invalid.
func TestStrategyExit_InvalidArgs(t *testing.T) {
	tests := []struct {
		name string
		args []ast.Expression
	}{
		{
			name: "zero arguments",
			args: []ast.Expression{},
		},
		{
			name: "empty string id",
			args: []ast.Expression{&ast.Literal{Value: ""}},
		},
	}

	handler := NewStrategyActionHandler()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			call := &ast.CallExpression{Arguments: tt.args}
			code, err := handler.generateExit(g, call)
			if err != nil {
				t.Fatalf("generateExit returned unexpected error: %v", err)
			}
			if !strings.Contains(code, "// strategy.exit()") {
				t.Errorf("expected comment stub for invalid args; got:\n%s", code)
			}
			if strings.Contains(code, "ExitWithLevels") {
				t.Errorf("must not emit ExitWithLevels for invalid args; got:\n%s", code)
			}
		})
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
		name                string
		defaultQtyType      string
		defaultQtyVal       float64
		wantDefaultQtyLogic bool   // entry emits EntryWithDefaultQty; order emits DefaultEntryQty block
		wantLiteralQty      string // for fixed types
		wantWarning         bool
	}
	qtyCases := []qtyCase{
		{
			name:                "strategy.cash defers qty computation to fill time",
			defaultQtyType:      "strategy.cash",
			defaultQtyVal:       600000.0,
			wantDefaultQtyLogic: true,
		},
		{
			name:                "cash unprefixed defers qty computation to fill time",
			defaultQtyType:      "cash",
			defaultQtyVal:       50000.0,
			wantDefaultQtyLogic: true,
		},
		{
			name:                "strategy.percent_of_equity defers qty computation to fill time",
			defaultQtyType:      "strategy.percent_of_equity",
			defaultQtyVal:       10.0,
			wantDefaultQtyLogic: true,
		},
		{
			name:                "percent_of_equity unprefixed defers qty computation to fill time",
			defaultQtyType:      "percent_of_equity",
			defaultQtyVal:       25.5,
			wantDefaultQtyLogic: true,
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

				switch {
				case qc.wantDefaultQtyLogic && fn.funcName == "entry":
					if !strings.Contains(code, "strat.EntryWithDefaultQty") {
						t.Errorf("Expected strat.EntryWithDefaultQty in output, got:\n%s", code)
					}
					if strings.Contains(code, fn.qtyVarName+" :=") {
						t.Errorf("Must NOT emit qty variable for entry+default, got:\n%s", code)
					}
				case qc.wantDefaultQtyLogic && fn.funcName == "order":
					want := fn.qtyVarName + " := strat.DefaultEntryQty(closeSeries.GetCurrent())"
					if !strings.Contains(code, want) {
						t.Errorf("Expected %q in output, got:\n%s", want, code)
					}
				default:
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

// TestStrategyEntry_QtyResolution tests that strategy.entry and strategy.order resolve
// qty from named arguments, including via constant-registry-backed identifiers.
// This covers the two call shapes Pine emits for named parameters:
//
//	mixed:  strategy.entry("id", long=strategy.long, qty=expr, ...)
//	all-named: strategy.entry(id="id", long=strategy.long, qty=expr, ...)
func TestStrategyEntry_QtyResolution(t *testing.T) {
	stratLong := &ast.MemberExpression{
		Object:   &ast.Identifier{Name: "strategy"},
		Property: &ast.Identifier{Name: "long"},
	}

	type callShape struct {
		name      string
		buildArgs func(qtyExpr ast.Expression) []ast.Expression
	}
	shapes := []callShape{
		{
			name: "mixed (positional id + named rest)",
			buildArgs: func(qtyExpr ast.Expression) []ast.Expression {
				return []ast.Expression{
					&ast.Literal{Value: "Buy"},
					&ast.ObjectExpression{Properties: []ast.Property{
						{Key: &ast.Identifier{Name: "long"}, Value: stratLong},
						{Key: &ast.Identifier{Name: "qty"}, Value: qtyExpr},
					}},
				}
			},
		},
	}

	type qtyCase struct {
		name           string
		qtyExpr        func() ast.Expression
		registerConst  func(g *generator)
		wantLiteralQty string // fragment expected in emitted code
	}
	qtyCases := []qtyCase{
		{
			name:           "named qty literal",
			qtyExpr:        func() ast.Expression { return &ast.Literal{Value: float64(300)} },
			registerConst:  func(g *generator) {},
			wantLiteralQty: "300,",
		},
		{
			name:    "named qty backed by constant registry identifier",
			qtyExpr: func() ast.Expression { return &ast.Identifier{Name: "myQty_"} },
			registerConst: func(g *generator) {
				g.constantRegistry.Register("myQty_", float64(5000))
			},
			wantLiteralQty: "5000,",
		},
	}

	stratFuncs := []struct {
		funcName string
	}{
		{"entry"},
		{"order"},
	}

	handler := NewStrategyActionHandler()

	for _, fn := range stratFuncs {
		for _, shape := range shapes {
			for _, qc := range qtyCases {
				testName := fn.funcName + "/" + shape.name + "/" + qc.name
				t.Run(testName, func(t *testing.T) {
					g := newTestGenerator()
					g.strategyConfig.DefaultQtyType = "fixed"
					g.strategyConfig.DefaultQtyValue = 1.0
					qc.registerConst(g)

					call := &ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "strategy"},
							Property: &ast.Identifier{Name: fn.funcName},
						},
						Arguments: shape.buildArgs(qc.qtyExpr()),
					}

					code, err := handler.GenerateCode(g, call)
					if err != nil {
						t.Fatalf("GenerateCode failed: %v", err)
					}
					if !strings.Contains(code, qc.wantLiteralQty) {
						t.Errorf("expected qty %q in output, got:\n%s", qc.wantLiteralQty, code)
					}
				})
			}
		}
	}
}
