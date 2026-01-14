package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestTempVariableManager_GenerateCalculations_SingleVariable tests calculation generation for one temp var */
func TestTempVariableManager_GenerateCalculations_SingleVariable(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
		indent:    2,
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
		ArgHash:  "test123",
	}

	varName := mgr.GetOrCreate(info)

	code, err := mgr.GenerateCalculations()
	if err != nil {
		t.Fatalf("GenerateCalculations() error = %v", err)
	}

	// Should contain inline SMA calculation
	if !strings.Contains(code, "Inline ta.sma(20)") {
		t.Errorf("Expected SMA comment not found in:\n%s", code)
	}

	// Should set the temp variable Series
	if !strings.Contains(code, varName+"Series.Set(") {
		t.Errorf("Expected Series.Set() for %s not found in:\n%s", varName, code)
	}

	// Should have warmup check
	if !strings.Contains(code, "if ctx.BarIndex <") {
		t.Errorf("Expected warmup check not found in:\n%s", code)
	}

	// Should have accumulation loop
	if !strings.Contains(code, "for j := 0; j < 20; j++") {
		t.Errorf("Expected accumulation loop not found in:\n%s", code)
	}
}

/* TestTempVariableManager_GenerateCalculations_MultipleVariables tests multiple temp vars */
func TestTempVariableManager_GenerateCalculations_MultipleVariables(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
		indent:    2,
	}
	g.taRegistry = NewTAFunctionRegistry()

	mgr := NewTempVariableManager(g)

	// Register SMA(20)
	call1 := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 20},
		},
	}

	info1 := CallInfo{Call: call1, FuncName: "ta.sma", ArgHash: "hash1"}
	varName1 := mgr.GetOrCreate(info1)

	// Register EMA(14)
	call2 := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "ema"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 14},
		},
	}

	info2 := CallInfo{Call: call2, FuncName: "ta.ema", ArgHash: "hash2"}
	varName2 := mgr.GetOrCreate(info2)

	code, err := mgr.GenerateCalculations()
	if err != nil {
		t.Fatalf("GenerateCalculations() error = %v", err)
	}

	// Should contain both calculations
	if !strings.Contains(code, varName1+"Series.Set(") {
		t.Errorf("Expected calculation for %s not found", varName1)
	}

	if !strings.Contains(code, varName2+"Series.Set(") {
		t.Errorf("Expected calculation for %s not found", varName2)
	}

	// Should have both indicator comments
	if !strings.Contains(code, "ta.sma(20)") {
		t.Error("Expected SMA calculation not found")
	}

	if !strings.Contains(code, "ta.ema(14)") {
		t.Error("Expected EMA calculation not found")
	}
}

/* TestTempVariableManager_GenerateCalculations_DifferentSources tests various source types */
func TestTempVariableManager_GenerateCalculations_DifferentSources(t *testing.T) {
	tests := []struct {
		name       string
		sourceExpr ast.Expression
		funcName   string
		period     int
		wantAccess string // Expected access pattern in generated code
	}{
		{
			name:       "SMA of close",
			sourceExpr: &ast.Identifier{Name: "close"},
			funcName:   "ta.sma",
			period:     50,
			wantAccess: "ctx.Data[ctx.BarIndex-j].Close",
		},
		{
			name:       "SMA of high",
			sourceExpr: &ast.Identifier{Name: "high"},
			funcName:   "ta.sma",
			period:     20,
			wantAccess: "ctx.Data[ctx.BarIndex-j].High",
		},
		{
			name: "SMA of close[4]",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 4},
				Computed: true,
			},
			funcName:   "ta.sma",
			period:     200,
			wantAccess: "ctx.Data[ctx.BarIndex-(j+4)].Close",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &generator{
				variables: make(map[string]string),
				constants: make(map[string]interface{}),
				indent:    2,
			}
			g.taRegistry = NewTAFunctionRegistry()

			mgr := NewTempVariableManager(g)

			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: tt.funcName[3:]}, // Strip "ta."
				},
				Arguments: []ast.Expression{
					tt.sourceExpr,
					&ast.Literal{Value: tt.period},
				},
			}

			info := CallInfo{Call: call, FuncName: tt.funcName, ArgHash: "test"}
			mgr.GetOrCreate(info)

			code, err := mgr.GenerateCalculations()
			if err != nil {
				t.Fatalf("GenerateCalculations() error = %v", err)
			}

			if !strings.Contains(code, tt.wantAccess) {
				t.Errorf("Expected access pattern %q not found in:\n%s", tt.wantAccess, code)
			}
		})
	}
}

/* TestTempVariableManager_GenerateCalculations_EmptyManager tests behavior with no registered vars */
func TestTempVariableManager_GenerateCalculations_EmptyManager(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
	g.taRegistry = NewTAFunctionRegistry()

	mgr := NewTempVariableManager(g)

	code, err := mgr.GenerateCalculations()
	if err != nil {
		t.Errorf("GenerateCalculations() unexpected error = %v", err)
	}

	if code != "" {
		t.Errorf("Expected empty code, got: %q", code)
	}
}

/* TestTempVariableManager_GenerateCalculations_ATRFunction tests ATR-specific calculation */
func TestTempVariableManager_GenerateCalculations_ATRFunction(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
		indent:    2,
	}
	g.taRegistry = NewTAFunctionRegistry()

	mgr := NewTempVariableManager(g)

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "atr"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: 2}, // ATR period
		},
	}

	info := CallInfo{Call: call, FuncName: "ta.atr", ArgHash: "atr_test"}
	varName := mgr.GetOrCreate(info)

	code, err := mgr.GenerateCalculations()
	if err != nil {
		t.Fatalf("GenerateCalculations() error = %v", err)
	}

	// ATR-specific checks
	if !strings.Contains(code, "Inline ATR(2)") {
		t.Errorf("Expected ATR comment not found in:\n%s", code)
	}

	// Should calculate True Range
	if !strings.Contains(code, "hl := ctx.Data[ctx.BarIndex].High - ctx.Data[ctx.BarIndex].Low") {
		t.Error("Expected True Range calculation not found")
	}

	// Should use RMA smoothing
	if !strings.Contains(code, "alpha := 1.0 / 2.0") {
		t.Error("Expected RMA alpha calculation not found")
	}

	// Should set temp variable
	if !strings.Contains(code, varName+"Series.Set(") {
		t.Errorf("Expected Series.Set() for %s not found", varName)
	}
}

/* TestTempVariableManager_GenerateCalculations_VariousPeriods tests different period values */
func TestTempVariableManager_GenerateCalculations_VariousPeriods(t *testing.T) {
	periods := []int{2, 5, 10, 14, 20, 50, 100, 200}

	for _, period := range periods {
		t.Run(string(rune(period)), func(t *testing.T) {
			g := &generator{
				variables: make(map[string]string),
				constants: make(map[string]interface{}),
				indent:    2,
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
					&ast.Literal{Value: period},
				},
			}

			info := CallInfo{Call: call, FuncName: "ta.sma", ArgHash: "test"}
			mgr.GetOrCreate(info)

			code, err := mgr.GenerateCalculations()
			if err != nil {
				t.Fatalf("Period %d: GenerateCalculations() error = %v", period, err)
			}

			// Should have correct warmup threshold
			expectedWarmup := strings.Contains(code, "if ctx.BarIndex <")
			if !expectedWarmup {
				t.Errorf("Period %d: Expected warmup check not found", period)
			}

			// Should have correct loop bound
			expectedLoop := strings.Contains(code, "for j := 0; j <")
			if !expectedLoop {
				t.Errorf("Period %d: Expected accumulation loop not found", period)
			}
		})
	}
}

/* TestTempVariableManager_GenerateCalculations_Deduplication ensures same call generates once */
func TestTempVariableManager_GenerateCalculations_Deduplication(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
		indent:    2,
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

	info := CallInfo{Call: call, FuncName: "ta.sma", ArgHash: "same"}

	// Register same call twice
	varName1 := mgr.GetOrCreate(info)
	varName2 := mgr.GetOrCreate(info)

	if varName1 != varName2 {
		t.Errorf("Expected same variable name, got %q and %q", varName1, varName2)
	}

	code, err := mgr.GenerateCalculations()
	if err != nil {
		t.Fatalf("GenerateCalculations() error = %v", err)
	}

	// Count occurrences of the calculation
	commentCount := strings.Count(code, "Inline ta.sma(20)")
	if commentCount != 1 {
		t.Errorf("Expected 1 calculation, found %d", commentCount)
	}
}

/* TestTempVariableManager_FullLifecycle tests complete temp var lifecycle */
func TestTempVariableManager_FullLifecycle(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
		indent:    2,
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
			&ast.Literal{Value: 21},
		},
	}

	info := CallInfo{Call: call, FuncName: "ta.ema", ArgHash: "lifecycle"}

	// Phase 1: Registration
	varName := mgr.GetOrCreate(info)
	if varName == "" {
		t.Fatal("GetOrCreate() returned empty variable name")
	}

	// Phase 2: Declaration
	decls := mgr.GenerateDeclarations()
	if !strings.Contains(decls, "var "+varName+"Series *series.Series") {
		t.Errorf("Declaration not found for %s in:\n%s", varName, decls)
	}

	// Phase 3: Initialization
	inits := mgr.GenerateInitializations()
	if !strings.Contains(inits, varName+"Series = series.NewSeries(len(ctx.Data))") {
		t.Errorf("Initialization not found for %s in:\n%s", varName, inits)
	}

	// Phase 4: Calculation
	calcs, err := mgr.GenerateCalculations()
	if err != nil {
		t.Fatalf("GenerateCalculations() error = %v", err)
	}
	if !strings.Contains(calcs, varName+"Series.Set(") {
		t.Errorf("Calculation not found for %s in:\n%s", varName, calcs)
	}

	// Phase 5: Advancement
	nexts := mgr.GenerateNextCalls()
	if !strings.Contains(nexts, varName+"Series.Next()") {
		t.Errorf("Next() call not found for %s in:\n%s", varName, nexts)
	}
}

/* BenchmarkGenerateCalculations measures calculation generation performance */
func BenchmarkGenerateCalculations(b *testing.B) {
	b.Run("SingleSMA", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			g := &generator{
				variables: make(map[string]string),
				constants: make(map[string]interface{}),
				indent:    2,
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

			info := CallInfo{Call: call, FuncName: "ta.sma", ArgHash: "bench"}
			mgr.GetOrCreate(info)

			_, err := mgr.GenerateCalculations()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("MultipleTAFunctions", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			g := &generator{
				variables: make(map[string]string),
				constants: make(map[string]interface{}),
				indent:    2,
			}
			g.taRegistry = NewTAFunctionRegistry()

			mgr := NewTempVariableManager(g)

			// Register 5 different TA functions
			functions := []string{"sma", "ema", "rma", "stdev", "wma"}
			for idx, fn := range functions {
				call := &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: fn},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: 20},
					},
				}

				info := CallInfo{Call: call, FuncName: "ta." + fn, ArgHash: string(rune(idx))}
				mgr.GetOrCreate(info)
			}

			_, err := mgr.GenerateCalculations()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
