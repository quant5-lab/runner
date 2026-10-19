package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/validation"
)

func TestGenerateStringExpression_ColorConstants(t *testing.T) {
	// Test code generation for string constants (colors and strategy constants)
	tests := []struct {
		name     string
		expr     ast.Expression
		expected string
	}{
		{
			name: "color.red constant",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Identifier{Name: "red"},
			},
			expected: `"#FF5252"`,
		},
		{
			name: "color.blue constant",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Identifier{Name: "blue"},
			},
			expected: `"#2962FF"`,
		},
		{
			name: "color.silver constant",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Identifier{Name: "silver"},
			},
			expected: `"#B2B5BE"`,
		},
		{
			name: "strategy.long constant",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "long"},
			},
			expected: "strategy.Long",
		},
		{
			name: "strategy.short constant",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "short"},
			},
			expected: "strategy.Short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createStringExpressionTestGenerator()
			result, err := gen.generateStringExpression(tt.expr)

			if err != nil {
				t.Fatalf("generateStringExpression() error = %v", err)
			}

			if result != tt.expected {
				t.Errorf("generateStringExpression() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGenerateStringExpression_SimpleConditional(t *testing.T) {
	tests := []struct {
		name            string
		expr            *ast.ConditionalExpression
		expectContains  []string
		expectNotExists []string
	}{
		{
			name: "simple conditional with color branches",
			expr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Operator: ">",
					Left:     &ast.Identifier{Name: "close"},
					Right:    &ast.Identifier{Name: "open"},
				},
				Consequent: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "lime"},
				},
				Alternate: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "red"},
				},
			},
			expectContains: []string{
				"func() string {",
				`"#00E676"`,
				`"#FF5252"`,
			},
		},
		{
			name: "conditional with strategy constants",
			expr: &ast.ConditionalExpression{
				Test: &ast.Identifier{Name: "enabled"},
				Consequent: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "long"},
				},
				Alternate: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "short"},
				},
			},
			expectContains: []string{
				"func() string {",
				"strategy.Long",
				"strategy.Short",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createStringExpressionTestGenerator()
			result, err := gen.generateStringExpression(tt.expr)

			if err != nil {
				t.Fatalf("generateStringExpression() error = %v", err)
			}

			for _, expected := range tt.expectContains {
				if !strings.Contains(result, expected) {
					t.Errorf("generateStringExpression() result missing %q\nGot: %s", expected, result)
				}
			}

			for _, notExpected := range tt.expectNotExists {
				if strings.Contains(result, notExpected) {
					t.Errorf("generateStringExpression() result should not contain %q\nGot: %s", notExpected, result)
				}
			}
		})
	}
}

func TestGenerateStringExpression_NestedConditional(t *testing.T) {
	tests := []struct {
		name           string
		expr           *ast.ConditionalExpression
		expectContains []string
	}{
		{
			name: "nested conditional with color branches",
			expr: &ast.ConditionalExpression{
				Test: &ast.Identifier{Name: "cond1"},
				Consequent: &ast.ConditionalExpression{
					Test: &ast.Identifier{Name: "cond2"},
					Consequent: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "color"},
						Property: &ast.Identifier{Name: "lime"},
					},
					Alternate: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "color"},
						Property: &ast.Identifier{Name: "maroon"},
					},
				},
				Alternate: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "red"},
				},
			},
			expectContains: []string{
				"func() string {",
				`"#00E676"`,
				`"#880E4F"`,
				`"#FF5252"`,
			},
		},
		{
			name: "deeply nested conditional with multiple colors",
			expr: &ast.ConditionalExpression{
				Test: &ast.Identifier{Name: "cond1"},
				Consequent: &ast.ConditionalExpression{
					Test: &ast.Identifier{Name: "cond2"},
					Consequent: &ast.ConditionalExpression{
						Test: &ast.Identifier{Name: "cond3"},
						Consequent: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "color"},
							Property: &ast.Identifier{Name: "lime"},
						},
						Alternate: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "color"},
							Property: &ast.Identifier{Name: "blue"},
						},
					},
					Alternate: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "color"},
						Property: &ast.Identifier{Name: "maroon"},
					},
				},
				Alternate: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "red"},
				},
			},
			expectContains: []string{
				"func() string {",
				`"#00E676"`,
				`"#2962FF"`,
				`"#880E4F"`,
				`"#FF5252"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createStringExpressionTestGenerator()
			result, err := gen.generateStringExpression(tt.expr)

			if err != nil {
				t.Fatalf("generateStringExpression() error = %v", err)
			}

			for _, expected := range tt.expectContains {
				if !strings.Contains(result, expected) {
					t.Errorf("generateStringExpression() result missing %q\nGot: %s", expected, result)
				}
			}
		})
	}
}

func TestGenerateStringExpression_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		expr        ast.Expression
		expectError bool
	}{
		{
			name:        "unsupported identifier",
			expr:        &ast.Identifier{Name: "unknownVar"},
			expectError: true,
		},
		{
			name:        "unsupported literal",
			expr:        &ast.Literal{Value: 123.45},
			expectError: true,
		},
		{
			name: "unsupported call expression",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
			},
			expectError: true,
		},
		{
			name: "color.new call expression succeeds",
			expr: &ast.CallExpression{
				Callee: MemberExpr("color", "new"),
				Arguments: []ast.Expression{
					MemberExpr("color", "red"),
					&ast.Literal{Value: float64(50)},
				},
			},
			expectError: false,
		},
		{
			name: "color.rgb call expression succeeds",
			expr: &ast.CallExpression{
				Callee: MemberExpr("color", "rgb"),
				Arguments: []ast.Expression{
					&ast.Literal{Value: float64(255)},
					&ast.Literal{Value: float64(0)},
					&ast.Literal{Value: float64(0)},
				},
			},
			expectError: false,
		},
		{
			name:        "syminfo.tickerid — string-typed builtin member, not an error",
			expr:        MemberExpr("syminfo", "tickerid"),
			expectError: false,
		},
		{
			name: "ticker.heikinashi call — ticker constructor, not an error",
			expr: &ast.CallExpression{
				Callee:    MemberExpr("ticker", "heikinashi"),
				Arguments: []ast.Expression{MemberExpr("syminfo", "tickerid")},
			},
			expectError: false,
		},
		{
			name: "unsupported ta member expression",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "sma"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createStringExpressionTestGenerator()
			_, err := gen.generateStringExpression(tt.expr)

			if tt.expectError && err == nil {
				t.Error("generateStringExpression() expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("generateStringExpression() unexpected error = %v", err)
			}
		})
	}
}

func TestGenerateStringExpression_ColorFunctionCalls(t *testing.T) {
	tests := []struct {
		name           string
		expr           ast.Expression
		expectContains []string
	}{
		{
			name: "color.new produces runtime call",
			expr: &ast.CallExpression{
				Callee: MemberExpr("color", "new"),
				Arguments: []ast.Expression{
					MemberExpr("color", "red"),
					&ast.Literal{Value: float64(50)},
				},
			},
			expectContains: []string{"visual.PineColorNew", "#FF5252"},
		},
		{
			name: "color.rgb produces runtime call",
			expr: &ast.CallExpression{
				Callee: MemberExpr("color", "rgb"),
				Arguments: []ast.Expression{
					&ast.Literal{Value: float64(255)},
					&ast.Literal{Value: float64(128)},
					&ast.Literal{Value: float64(0)},
					&ast.Literal{Value: float64(20)},
				},
			},
			expectContains: []string{"visual.PineColorRGB", "255", "128", "20"},
		},
		{
			name: "color.from_gradient produces runtime call",
			expr: &ast.CallExpression{
				Callee: MemberExpr("color", "from_gradient"),
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "rsi"},
					&ast.Literal{Value: float64(0)},
					&ast.Literal{Value: float64(100)},
					MemberExpr("color", "red"),
					MemberExpr("color", "green"),
				},
			},
			expectContains: []string{"visual.PineColorFromGradient", "#FF5252", "#4CAF50"},
		},
		{
			name: "color.new in ternary branches",
			expr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Operator: ">",
					Left:     &ast.Identifier{Name: "close"},
					Right:    &ast.Identifier{Name: "open"},
				},
				Consequent: &ast.CallExpression{
					Callee: MemberExpr("color", "new"),
					Arguments: []ast.Expression{
						MemberExpr("color", "lime"),
						&ast.Literal{Value: float64(0)},
					},
				},
				Alternate: &ast.CallExpression{
					Callee: MemberExpr("color", "new"),
					Arguments: []ast.Expression{
						MemberExpr("color", "red"),
						&ast.Literal{Value: float64(50)},
					},
				},
			},
			expectContains: []string{
				"func() string {",
				"visual.PineColorNew",
				"#00E676",
				"#FF5252",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createStringExpressionTestGenerator()
			result, err := gen.generateStringExpression(tt.expr)

			if err != nil {
				t.Fatalf("generateStringExpression() error = %v", err)
			}

			for _, expected := range tt.expectContains {
				if !strings.Contains(result, expected) {
					t.Errorf("result missing %q\nGot: %s", expected, result)
				}
			}
		})
	}
}

func TestGenerateStringVariableInit_ColorFunctionCalls(t *testing.T) {
	tests := []struct {
		name           string
		expr           ast.Expression
		expectContains []string
	}{
		{
			name: "color.new assignment",
			expr: &ast.CallExpression{
				Callee: MemberExpr("color", "new"),
				Arguments: []ast.Expression{
					MemberExpr("color", "blue"),
					&ast.Literal{Value: float64(30)},
				},
			},
			expectContains: []string{"myColor = visual.PineColorNew", "#2962FF"},
		},
		{
			name: "color.rgb assignment",
			expr: &ast.CallExpression{
				Callee: MemberExpr("color", "rgb"),
				Arguments: []ast.Expression{
					&ast.Literal{Value: float64(100)},
					&ast.Literal{Value: float64(200)},
					&ast.Literal{Value: float64(50)},
				},
			},
			expectContains: []string{"myColor = visual.PineColorRGB"},
		},
		{
			name: "color.from_gradient assignment",
			expr: &ast.CallExpression{
				Callee: MemberExpr("color", "from_gradient"),
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "rsi"},
					&ast.Literal{Value: float64(0)},
					&ast.Literal{Value: float64(100)},
					MemberExpr("color", "red"),
					MemberExpr("color", "green"),
				},
			},
			expectContains: []string{"myColor = visual.PineColorFromGradient"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createStringExpressionTestGenerator()
			result, err := gen.generateStringVariableInit("myColor", tt.expr)

			if err != nil {
				t.Fatalf("generateStringVariableInit() error = %v", err)
			}

			for _, expected := range tt.expectContains {
				if !strings.Contains(result, expected) {
					t.Errorf("result missing %q\nGot: %s", expected, result)
				}
			}
		})
	}
}

func TestGenerateStringVariableInit_UnsupportedCallExpression(t *testing.T) {
	gen := createStringExpressionTestGenerator()

	_, err := gen.generateStringVariableInit("x", &ast.CallExpression{
		Callee: MemberExpr("ta", "sma"),
	})

	if err == nil {
		t.Error("expected error for non-color call expression in string variable init")
	}
}

func TestGenerateStringExpression_BuiltinStringMembers(t *testing.T) {
	tests := []struct {
		name        string
		expr        ast.Expression
		wantResult  string
		expectError bool
	}{
		// ── syminfo string properties ─────────────────────────────────────────
		{"syminfo.tickerid", MemberExpr("syminfo", "tickerid"), "syminfo_tickerid", false},
		{"syminfo.ticker", MemberExpr("syminfo", "ticker"), "syminfo_tickerid", false},
		{"syminfo.timezone", MemberExpr("syminfo", "timezone"), "ctx.Timezone", false},
		{"syminfo.type", MemberExpr("syminfo", "type"), `"stock"`, false},
		{"syminfo.currency", MemberExpr("syminfo", "currency"), `"USD"`, false},
		// ── timeframe string properties ───────────────────────────────────────
		{"timeframe.period", MemberExpr("timeframe", "period"), "ctx.Timeframe", false},
		// ── numeric namespace member is not a valid string expression ─────────
		{"syminfo.mintick — numeric, must error", MemberExpr("syminfo", "mintick"), "", true},
		{"ta.sma — not a string member, must error", MemberExpr("ta", "sma"), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createStringExpressionTestGenerator()
			result, err := gen.generateStringExpression(tt.expr)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got result %q", result)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.wantResult {
				t.Errorf("got %q, want %q", result, tt.wantResult)
			}
		})
	}
}

func TestGenerateStringExpression_TickerConstructorCalls(t *testing.T) {
	syminfo := func(prop string) ast.Expression { return MemberExpr("syminfo", prop) }
	call := func(ns, fn string, args ...ast.Expression) ast.Expression {
		return &ast.CallExpression{Callee: MemberExpr(ns, fn), Arguments: args}
	}

	tests := []struct {
		name         string
		expr         ast.Expression
		wantContains string
		expectError  bool
	}{
		{
			name:         "ticker.heikinashi with syminfo.tickerid",
			expr:         call("ticker", "heikinashi", syminfo("tickerid")),
			wantContains: "ticker.Heikinashi(ctx.Symbol)",
		},
		{
			name:         "bare v4 heikinashi form (preprocessor not yet applied)",
			expr:         &ast.CallExpression{Callee: &ast.Identifier{Name: "heikinashi"}, Arguments: []ast.Expression{syminfo("tickerid")}},
			wantContains: "ticker.Heikinashi(ctx.Symbol)",
		},
		{
			name:         "ticker.standard with syminfo.tickerid strips modifier",
			expr:         call("ticker", "standard", syminfo("tickerid")),
			wantContains: "ticker.Standard(ctx.Symbol)",
		},
		{
			name:         "ticker.standard with no args returns ctx.Symbol",
			expr:         &ast.CallExpression{Callee: MemberExpr("ticker", "standard")},
			wantContains: "ctx.Symbol",
		},
		{
			name:        "ta.sma is not a ticker constructor — must error",
			expr:        &ast.CallExpression{Callee: MemberExpr("ta", "sma")},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createStringExpressionTestGenerator()
			result, err := gen.generateStringExpression(tt.expr)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got result %q", result)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(result, tt.wantContains) {
				t.Errorf("got %q, want substring %q", result, tt.wantContains)
			}
		})
	}
}

func TestGenerateStringVariableInit_BuiltinStringMembers(t *testing.T) {
	tests := []struct {
		name        string
		varName     string
		expr        ast.Expression
		wantContain string
	}{
		{
			name:        "syminfo.tickerid assignment",
			varName:     "t",
			expr:        MemberExpr("syminfo", "tickerid"),
			wantContain: "t = syminfo_tickerid",
		},
		{
			name:        "syminfo.timezone assignment",
			varName:     "tz",
			expr:        MemberExpr("syminfo", "timezone"),
			wantContain: "tz = ctx.Timezone",
		},
		{
			name:        "timeframe.period assignment",
			varName:     "tf",
			expr:        MemberExpr("timeframe", "period"),
			wantContain: "tf = ctx.Timeframe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createStringExpressionTestGenerator()
			result, err := gen.generateStringVariableInit(tt.varName, tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(result, tt.wantContain) {
				t.Errorf("got %q, want substring %q", result, tt.wantContain)
			}
		})
	}
}

func TestGenerateStringVariableInit_TickerConstructorCalls(t *testing.T) {
	syminfo := func(prop string) ast.Expression { return MemberExpr("syminfo", prop) }

	tests := []struct {
		name        string
		varName     string
		expr        ast.Expression
		wantContain string
	}{
		{
			name:    "ticker.heikinashi assignment",
			varName: "_ha",
			expr: &ast.CallExpression{
				Callee:    MemberExpr("ticker", "heikinashi"),
				Arguments: []ast.Expression{syminfo("tickerid")},
			},
			wantContain: "_ha = ticker.Heikinashi(ctx.Symbol)",
		},
		{
			name:    "ticker.standard no-arg assignment returns ctx.Symbol",
			varName: "_std",
			expr: &ast.CallExpression{
				Callee: MemberExpr("ticker", "standard"),
			},
			wantContain: "_std = ctx.Symbol",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createStringExpressionTestGenerator()
			result, err := gen.generateStringVariableInit(tt.varName, tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(result, tt.wantContain) {
				t.Errorf("got %q, want substring %q", result, tt.wantContain)
			}
		})
	}
}

func createStringExpressionTestGenerator() *generator {
	typeSystem := NewTypeInferenceEngine()
	return &generator{
		imports:        make(map[string]bool),
		variables:      make(map[string]string),
		strategyConfig: NewStrategyConfig(),
		taRegistry:     NewTAFunctionRegistry(),
		constEvaluator: validation.NewWarmupAnalyzer(),
		boolConverter:  NewBooleanConverter(typeSystem),
		typeSystem:     typeSystem,
		colorHandler:   NewColorHandler(),
		builtinHandler: NewBuiltinIdentifierHandler(),
	}
}
