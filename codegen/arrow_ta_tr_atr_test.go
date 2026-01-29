package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestArrowTACall_TrAsAccessor validates ta.tr as AccessGenerator source in arrow functions */
func TestArrowTACall_TrAsAccessor(t *testing.T) {
	tests := []struct {
		name            string
		taFunc          string
		sourceExpr      ast.Expression
		lengthExpr      ast.Expression
		expectAccessor  string
		expectInline    bool
		expectSeriesGet bool
	}{
		{
			name:   "ta.sma(ta.tr, 20) windowed function",
			taFunc: "ta.sma",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "tr"},
			},
			lengthExpr:      &ast.Literal{Value: float64(20)},
			expectAccessor:  "BuiltinTrueRangeAccessor",
			expectInline:    true,
			expectSeriesGet: false,
		},
		{
			name:   "ta.ema(ta.tr, 14) stateful function",
			taFunc: "ta.ema",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "tr"},
			},
			lengthExpr:      &ast.Literal{Value: float64(14)},
			expectAccessor:  "BuiltinTrueRangeAccessor",
			expectInline:    true,
			expectSeriesGet: false,
		},
		{
			name:   "ta.rma(ta.tr, 14) for custom ATR",
			taFunc: "ta.rma",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "tr"},
			},
			lengthExpr:      &ast.Literal{Value: float64(14)},
			expectAccessor:  "BuiltinTrueRangeAccessor",
			expectInline:    true,
			expectSeriesGet: false,
		},
		{
			name:   "ta.stdev(ta.tr, 10) volatility indicator",
			taFunc: "ta.stdev",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "tr"},
			},
			lengthExpr:      &ast.Literal{Value: float64(10)},
			expectAccessor:  "BuiltinTrueRangeAccessor",
			expectInline:    true,
			expectSeriesGet: false,
		},
		{
			name:   "ta.highest(ta.tr, 5) range extremes",
			taFunc: "ta.highest",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "tr"},
			},
			lengthExpr:      &ast.Literal{Value: float64(5)},
			expectAccessor:  "BuiltinTrueRangeAccessor",
			expectInline:    true,
			expectSeriesGet: false,
		},
		{
			name:            "ta.sma(tr, 20) bare tr identifier",
			taFunc:          "ta.sma",
			sourceExpr:      &ast.Identifier{Name: "tr"},
			lengthExpr:      &ast.Literal{Value: float64(20)},
			expectAccessor:  "BuiltinTrueRangeAccessor",
			expectInline:    true,
			expectSeriesGet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			accessResolver := NewArrowSeriesAccessResolver()
			identifierResolver := NewArrowIdentifierResolver(accessResolver)
			exprGen := &legacyArrowExpressionGenerator{gen: gen}
			factory := NewArrowAwareAccessorFactory(identifierResolver, exprGen, gen, gen.symbolTable)

			accessor, err := factory.CreateAccessorForExpression(tt.sourceExpr)
			if err != nil {
				t.Fatalf("CreateAccessorForExpression failed: %v", err)
			}

			if accessor == nil {
				t.Fatal("CreateAccessorForExpression returned nil")
			}

			accessorType := strings.Contains(getTypeName(accessor), tt.expectAccessor)
			if !accessorType {
				t.Errorf("Expected accessor type containing %q, got %T", tt.expectAccessor, accessor)
			}

			loopCode := accessor.GenerateLoopValueAccess("j")
			if loopCode == "" {
				t.Fatal("GenerateLoopValueAccess returned empty string")
			}

			if tt.expectInline {
				if !strings.Contains(loopCode, "math.Max") {
					t.Errorf("Expected inline TR calculation with math.Max, got: %s", loopCode)
				}
				if !strings.Contains(loopCode, "ctx.BarIndex") {
					t.Errorf("Expected ctx.BarIndex in inline calculation, got: %s", loopCode)
				}
			}

			if !tt.expectSeriesGet && strings.Contains(loopCode, "Series.Get(") {
				t.Errorf("Should not use Series.Get() for builtin, got: %s", loopCode)
			}

			currentCode := accessor.GenerateCurrentValueAccess()
			if currentCode == "" {
				t.Error("GenerateCurrentValueAccess returned empty string")
			}

			if tt.expectInline && !strings.Contains(currentCode, "math.Max") {
				t.Errorf("Current value should contain inline calculation, got: %s", currentCode)
			}
		})
	}
}

/* TestArrowTACall_TrHistoricalOffset validates ta.tr[N] subscript access via builtin handler */
func TestArrowTACall_TrHistoricalOffset(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name         string
		offset       int
		expectValid  bool
		expectOffset bool
	}{
		{"ta.tr historical offset 0", 0, true, true},
		{"ta.tr historical offset 1", 1, true, true},
		{"ta.tr historical offset 2", 2, true, true},
		{"ta.tr historical offset 5", 5, true, true},
		{"ta.tr historical offset 10", 10, true, true},
		{"ta.tr historical offset 100", 100, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nestedExpr := &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "tr"},
					Computed: false,
				},
				Property: &ast.Literal{Value: tt.offset},
				Computed: true,
			}

			code, resolved := handler.TryResolveMemberExpression(nestedExpr, false)
			if !tt.expectValid {
				if resolved {
					t.Error("Expected unresolved for invalid offset, got resolved")
				}
				return
			}

			if !resolved {
				t.Fatalf("TryResolveMemberExpression should resolve ta.tr[%d]", tt.offset)
			}

			if tt.expectOffset && !strings.Contains(code, fmt.Sprintf("i-%d", tt.offset)) {
				t.Errorf("Expected offset i-%d in code, got: %s", tt.offset, code)
			}

			if !strings.Contains(code, "math.Max") {
				t.Errorf("Expected inline TR calculation, got: %s", code)
			}

			if !strings.Contains(code, "barIdx") {
				t.Errorf("Expected barIdx variable in historical access, got: %s", code)
			}
		})
	}
}

/* TestArrowTACall_TrVsParameter validates builtin precedence over user parameters */
func TestArrowTACall_TrVsParameter(t *testing.T) {
	t.Run("tr builtin takes precedence over parameter", func(t *testing.T) {
		gen := newTestGenerator()
		gen.variables = map[string]string{"my_custom_tr": "float"}

		accessResolver := NewArrowSeriesAccessResolver()
		accessResolver.parameters = map[string]bool{"tr_param": true}

		identifierResolver := NewArrowIdentifierResolver(accessResolver)
		exprGen := &legacyArrowExpressionGenerator{gen: gen}
		factory := NewArrowAwareAccessorFactory(identifierResolver, exprGen, gen, gen.symbolTable)

		trBuiltin := &ast.Identifier{Name: "tr"}
		accessor, err := factory.CreateAccessorForExpression(trBuiltin)
		if err != nil {
			t.Fatalf("CreateAccessorForExpression(tr) failed: %v", err)
		}

		if _, ok := accessor.(*BuiltinTrueRangeAccessor); !ok {
			t.Errorf("tr should resolve to BuiltinTrueRangeAccessor, got %T", accessor)
		}

		/* Test ta.tr MemberExpression - always builtin */
		taTr := &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "tr"},
		}
		accessor2, err := factory.CreateAccessorForExpression(taTr)
		if err != nil {
			t.Fatalf("CreateAccessorForExpression(ta.tr) failed: %v", err)
		}

		if _, ok := accessor2.(*BuiltinTrueRangeAccessor); !ok {
			t.Errorf("ta.tr should resolve to BuiltinTrueRangeAccessor, got %T", accessor2)
		}
	})

	t.Run("user parameter with different name", func(t *testing.T) {
		gen := newTestGenerator()
		accessResolver := NewArrowSeriesAccessResolver()
		accessResolver.parameters = map[string]bool{"my_tr_series": true}

		identifierResolver := NewArrowIdentifierResolver(accessResolver)
		exprGen := &legacyArrowExpressionGenerator{gen: gen}
		factory := NewArrowAwareAccessorFactory(identifierResolver, exprGen, gen, gen.symbolTable)

		myTr := &ast.Identifier{Name: "my_tr_series"}
		accessor, err := factory.CreateAccessorForExpression(myTr)
		if err != nil {
			t.Fatalf("CreateAccessorForExpression(my_tr_series) failed: %v", err)
		}

		if _, ok := accessor.(*BuiltinTrueRangeAccessor); ok {
			t.Error("User parameter should not resolve to BuiltinTrueRangeAccessor")
		}
	})
}

/* TestArrowTACall_TrFirstBarBehavior validates first bar TR calculation with no previous close */
func TestArrowTACall_TrFirstBarBehavior(t *testing.T) {
	gen := newTestGenerator()
	accessResolver := NewArrowSeriesAccessResolver()
	identifierResolver := NewArrowIdentifierResolver(accessResolver)
	exprGen := &legacyArrowExpressionGenerator{gen: gen}
	factory := NewArrowAwareAccessorFactory(identifierResolver, exprGen, gen, gen.symbolTable)

	taTr := &ast.MemberExpression{
		Object:   &ast.Identifier{Name: "ta"},
		Property: &ast.Identifier{Name: "tr"},
	}

	accessor, err := factory.CreateAccessorForExpression(taTr)
	if err != nil {
		t.Fatalf("CreateAccessorForExpression failed: %v", err)
	}

	trAccessor, ok := accessor.(*BuiltinTrueRangeAccessor)
	if !ok {
		t.Fatalf("Expected BuiltinTrueRangeAccessor, got %T", accessor)
	}

	currentCode := trAccessor.GenerateCurrentValueAccess()

	if !strings.Contains(currentCode, "ctx.BarIndex < 1") {
		t.Error("Expected first bar check 'ctx.BarIndex < 1' in generated code")
	}

	if !strings.Contains(currentCode, ".High") && !strings.Contains(currentCode, ".Low") {
		t.Error("Expected High-Low fallback calculation for first bar")
	}

	loopCode := trAccessor.GenerateLoopValueAccess("j")
	if !strings.Contains(loopCode, "barIdx < 1") {
		t.Error("Expected first bar check in loop access generation")
	}
}

/* TestArrowTACall_TrWarmupConsistency validates TR base offset for warmup calculations */
func TestArrowTACall_TrWarmupConsistency(t *testing.T) {
	tests := []struct {
		name        string
		period      int
		expectValid bool
	}{
		{"period 1 immediate", 1, true},
		{"period 10 small", 10, true},
		{"period 14 standard ATR", 14, true},
		{"period 50 medium", 50, true},
		{"period 200 large", 200, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			accessResolver := NewArrowSeriesAccessResolver()
			identifierResolver := NewArrowIdentifierResolver(accessResolver)
			exprGen := &legacyArrowExpressionGenerator{gen: gen}
			factory := NewArrowAwareAccessorFactory(identifierResolver, exprGen, gen, gen.symbolTable)

			taTr := &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "tr"},
			}

			accessor, err := factory.CreateAccessorForExpression(taTr)
			if err != nil {
				t.Fatalf("CreateAccessorForExpression failed: %v", err)
			}

			baseOffset := accessor.GetBaseOffset()
			if baseOffset < 0 {
				t.Errorf("Base offset should be non-negative, got %d", baseOffset)
			}

			if baseOffset != 0 {
				t.Errorf("TR base offset should be 0 (current bar), got %d", baseOffset)
			}
		})
	}
}

func getTypeName(v interface{}) string {
	if v == nil {
		return "nil"
	}
	return fmt.Sprintf("%T", v)
}
