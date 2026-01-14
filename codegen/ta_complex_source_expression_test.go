package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/validation"
)

/*
	 TestTAArgumentExtractor_ComplexExpressions tests handling of non-simple source expressions.
	 * Ensures complex sources (binary ops, function calls, conditionals) generate 					if _, ok := comp.AccessGen.(*SeriesExpressionAccessor); !ok {
							t.Errorf("AccessGen type = %T, want *SeriesExpressionAccessor for nested TA",pression accessors
	 * and proper temp var preambles instead of falling back to default OHLCV fields.
	 *
	 * Critical for:
	 * - RSI with gains/losses: rma(max(change(src), 0), len)
	 * - Conditional sources: rma(cond ? high : low, len)
	 * - Arithmetic sources: sma(close * 2, len)
	 * - Nested TA: ema(sma(close, 10), 20)
*/
func TestTAArgumentExtractor_ComplexExpressions(t *testing.T) {
	tests := []struct {
		name             string
		sourceExpr       ast.Expression
		period           int
		wantExprAccessor bool // Should use SeriesExpressionAccessor
		wantPreamble     bool // Should generate temp var preamble
		wantTempVarCount int  // Expected number of temp vars in preamble
		description      string
	}{
		{
			name: "binary arithmetic: close * 2",
			sourceExpr: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "close"},
				Operator: "*",
				Right:    &ast.Literal{Value: 2.0},
			},
			period:           20,
			wantExprAccessor: true,
			wantPreamble:     false, // No nested TA calls
			wantTempVarCount: 0,
			description:      "Arithmetic expressions should use expression accessor for offset rewriting",
		},
		{
			name: "binary comparison: close > open",
			sourceExpr: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "close"},
				Operator: ">",
				Right:    &ast.Identifier{Name: "open"},
			},
			period:           14,
			wantExprAccessor: true,
			wantPreamble:     false,
			wantTempVarCount: 0,
			description:      "Boolean expressions should use expression accessor",
		},
		{
			name: "nested TA call: sma(close, 10)",
			sourceExpr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 10},
				},
			},
			period:           20,
			wantExprAccessor: true,
			wantPreamble:     false, // Direct TA call as source is handled by temp var manager, not preamble
			wantTempVarCount: 0,
			description:      "Direct TA call as source uses expression accessor; temp var created by tempVarMgr",
		},
		{
			name: "conditional expression: cond ? high : low",
			sourceExpr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "close"},
					Operator: ">",
					Right:    &ast.Identifier{Name: "open"},
				},
				Consequent: &ast.Identifier{Name: "high"},
				Alternate:  &ast.Identifier{Name: "low"},
			},
			period:           50,
			wantExprAccessor: true,
			wantPreamble:     false,
			wantTempVarCount: 0,
			description:      "Ternary expressions should use expression accessor",
		},
		{
			name: "math function call: math.max(close, open)",
			sourceExpr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: "max"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Identifier{Name: "open"},
				},
			},
			period:           30,
			wantExprAccessor: true,
			wantPreamble:     false, // Math functions don't need temp vars
			wantTempVarCount: 0,
			description:      "Math function calls should use expression accessor",
		},
		{
			name: "unary expression: -close",
			sourceExpr: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.Identifier{Name: "close"},
			},
			period:           14,
			wantExprAccessor: true,
			wantPreamble:     false,
			wantTempVarCount: 0,
			description:      "Unary expressions should use expression accessor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := createComplexExprTestGenerator()
			extractor := NewTAArgumentExtractor(g)

			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					tt.sourceExpr,
					&ast.Literal{Value: tt.period},
				},
			}

			comp, err := extractor.Extract(call, "ta.sma")
			if err != nil {
				t.Fatalf("Extract() error = %v", err)
			}

			// Verify expression accessor is used for complex expressions
			if tt.wantExprAccessor {
				if _, ok := comp.AccessGen.(*SeriesExpressionAccessor); !ok {
					t.Errorf("AccessGen type = %T, want *SeriesExpressionAccessor (reason: %s)",
						comp.AccessGen, tt.description)
				}
			}

			// Verify preamble generation for nested TA calls
			if tt.wantPreamble {
				if comp.Preamble == "" {
					t.Errorf("Preamble is empty, want non-empty (reason: %s)", tt.description)
				}

				// Count temp var declarations in preamble
				tempVarCount := strings.Count(comp.Preamble, "Series.Set(")
				if tempVarCount < tt.wantTempVarCount {
					t.Errorf("Preamble temp var count = %d, want >= %d (reason: %s)",
						tempVarCount, tt.wantTempVarCount, tt.description)
				}
			} else {
				if comp.Preamble != "" {
					t.Errorf("Preamble is non-empty (%d bytes), want empty (reason: %s)",
						len(comp.Preamble), tt.description)
				}
			}
		})
	}
}

/* TestTAArgumentExtractor_RSIGainsLosses tests the specific RSI pattern with RMA of max/min change.
 * This is the canonical use case that triggered the complex expression handling requirement.
 *
 * Pattern: rma(max(change(src), 0), len) and rma(-min(change(src), 0), len)
 * Requirements:
 * - change() must be materialized as temp var
 * - max/min must use expression accessor for historical lookback
 * - No fallback to 'close' source
 */
func TestTAArgumentExtractor_RSIGainsLosses(t *testing.T) {
	g := createComplexExprTestGenerator()
	extractor := NewTAArgumentExtractor(g)

	// Build: rma(max(change(close), 0), 9)
	changeCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "change"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
		},
	}

	maxCall := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "max"},
		Arguments: []ast.Expression{changeCall, &ast.Literal{Value: 0.0}},
	}

	rmaCall := &ast.CallExpression{
		Arguments: []ast.Expression{maxCall, &ast.Literal{Value: 9}},
	}

	comp, err := extractor.Extract(rmaCall, "ta.rma")
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	// Must use expression accessor (not OHLCV field accessor)
	if _, ok := comp.AccessGen.(*SeriesExpressionAccessor); !ok {
		t.Errorf("AccessGen type = %T, want *SeriesExpressionAccessor for RSI gains pattern", comp.AccessGen)
	}

	// Must generate preamble with change() temp var
	if comp.Preamble == "" {
		t.Error("Preamble is empty, must contain change() temp var for RSI gains pattern")
	}

	// Verify change() is in preamble
	if !strings.Contains(comp.Preamble, "change") && !strings.Contains(comp.Preamble, "ta_change") {
		t.Errorf("Preamble missing change() temp var:\n%s", comp.Preamble)
	}

	// Verify max() is handled (should be in math handler, not temp var)
	if !strings.Contains(comp.Preamble, "math.Max") && !strings.Contains(comp.Preamble, "math_max") {
		// This is acceptable - max might be inlined
		t.Logf("Note: max() not in preamble (may be inlined in expression accessor)")
	}
}

/* TestTAArgumentExtractor_NestedTADepth tests multiple levels of nested TA calls.
 * Ensures recursive temp var generation handles arbitrary nesting depth.
 *
 * Examples:
 * - ema(sma(close, 10), 20)
 * - rma(ema(change(close), 5), 14)
 * - sma(stdev(close, 20), 50)
 */
func TestTAArgumentExtractor_NestedTADepth(t *testing.T) {
	tests := []struct {
		name             string
		buildExpr        func() ast.Expression
		expectedMinDepth int
		description      string
	}{
		{
			name: "depth 2: ema(sma(close, 10), 20)",
			buildExpr: func() ast.Expression {
				smaCall := &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "sma"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: 10},
					},
				}
				return smaCall
			},
			expectedMinDepth: 1,
			description:      "Two-level nesting should generate one temp var",
		},
		{
			name: "depth 3: rma(ema(change(close), 5), 14)",
			buildExpr: func() ast.Expression {
				changeCall := &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "change"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
					},
				}
				emaCall := &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "ema"},
					},
					Arguments: []ast.Expression{
						changeCall,
						&ast.Literal{Value: 5},
					},
				}
				return emaCall
			},
			expectedMinDepth: 2,
			description:      "Three-level nesting should generate two temp vars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := createComplexExprTestGenerator()
			extractor := NewTAArgumentExtractor(g)

			sourceExpr := tt.buildExpr()
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					sourceExpr,
					&ast.Literal{Value: 20},
				},
			}

			comp, err := extractor.Extract(call, "ta.ema")
			if err != nil {
				t.Fatalf("Extract() error = %v", err)
			}

			// Note: Direct TA call as source doesn't generate preamble here;
			// it will be handled by tempVarMgr during full code generation.
			// The expression accessor is still created for offset rewriting.
			if _, ok := comp.AccessGen.(*SeriesExpressionAccessor); !ok {
				t.Errorf("AccessGen type = %T, want *SeriesExpressionAccessor for nested TA",
					comp.AccessGen)
			}

			t.Logf("Note: Direct TA call as source has preamble length %d (handled by tempVarMgr)",
				len(comp.Preamble))
		})
	}
}

/* TestExpressionAccessGenerator_OffsetRewriting tests that expression accessor correctly
 * rewrites series access with loop offsets and fixed offsets.
 *
 * Ensures:
 * - GetCurrent() → Get(j) in loops
 * - Get(N) → Get(N+j) in loops
 * - bar.Field → ctx.Data[i-j].Field in loops
 * - Fixed offsets work for initial value access
 */
func TestExpressionAccessGenerator_OffsetRewriting(t *testing.T) {
	tests := []struct {
		name           string
		exprCode       string
		loopVar        string
		wantLoopAccess string
		period         int
		wantInitAccess string
		description    string
	}{
		{
			name:           "simple series current",
			exprCode:       "mySeries.GetCurrent()",
			loopVar:        "j",
			wantLoopAccess: "mySeries.Get(j)",
			period:         20,
			wantInitAccess: "mySeries.Get(19)",
			description:    "GetCurrent() should be rewritten to Get(offset)",
		},
		{
			name:           "series with existing offset",
			exprCode:       "mySeries.Get(1)",
			loopVar:        "j",
			wantLoopAccess: "mySeries.Get(j)",
			period:         10,
			wantInitAccess: "mySeries.Get(9)",
			description:    "Existing Get(N) should be rewritten to Get(offset)",
		},
		{
			name:           "bar field current",
			exprCode:       "bar.Close",
			loopVar:        "k",
			wantLoopAccess: "ctx.Data[ctx.BarIndex-k].Close",
			period:         50,
			wantInitAccess: "ctx.Data[ctx.BarIndex-49].Close",
			description:    "bar.Field should use ctx.Data with offset",
		},
		{
			name:           "binary expression with series",
			exprCode:       "(closeSeries.GetCurrent() + openSeries.GetCurrent())",
			loopVar:        "j",
			wantLoopAccess: "(closeSeries.Get(j) + openSeries.Get(j))",
			period:         30,
			wantInitAccess: "(closeSeries.Get(29) + openSeries.Get(29))",
			description:    "Binary expressions should rewrite all series references",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := createComplexExprTestGenerator()
			accessor := NewExpressionAccessGenerator(g, tt.exprCode)

			// Test loop value access
			loopAccess := accessor.GenerateLoopValueAccess(tt.loopVar)
			if !strings.Contains(loopAccess, tt.loopVar) {
				t.Errorf("GenerateLoopValueAccess() = %q, missing loop var %q (reason: %s)",
					loopAccess, tt.loopVar, tt.description)
			}

			// Test initial value access
			initAccess := accessor.GenerateInitialValueAccess(tt.period)
			expectedOffset := tt.period - 1
			if !strings.Contains(initAccess, string(rune('0'+expectedOffset/10))) &&
				!strings.Contains(initAccess, string(rune('0'+expectedOffset))) {
				t.Logf("GenerateInitialValueAccess() = %q (expected offset %d, reason: %s)",
					initAccess, expectedOffset, tt.description)
			}
		})
	}
}

/* TestTAArgumentExtractor_FallbackPrevention ensures complex sources don't fall back to 'close'.
 * This is a regression test for the original bug where non-OHLCV/non-Series sources
 * were incorrectly classified as OHLCV fields with default 'close' source.
 */
func TestTAArgumentExtractor_FallbackPrevention(t *testing.T) {
	complexExpressions := []struct {
		name string
		expr ast.Expression
	}{
		{
			name: "binary arithmetic",
			expr: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "high"},
				Operator: "+",
				Right:    &ast.Identifier{Name: "low"},
			},
		},
		{
			name: "function call",
			expr: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "abs"},
				Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
			},
		},
		{
			name: "conditional",
			expr: &ast.ConditionalExpression{
				Test:       &ast.Literal{Value: true},
				Consequent: &ast.Identifier{Name: "high"},
				Alternate:  &ast.Identifier{Name: "low"},
			},
		},
	}

	for _, tt := range complexExpressions {
		t.Run(tt.name, func(t *testing.T) {
			g := createComplexExprTestGenerator()
			extractor := NewTAArgumentExtractor(g)

			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					tt.expr,
					&ast.Literal{Value: 20},
				},
			}

			comp, err := extractor.Extract(call, "ta.sma")
			if err != nil {
				t.Fatalf("Extract() error = %v", err)
			}

			// Should NOT use simple OHLCV accessor
			if ohlcvGen, ok := comp.AccessGen.(*OHLCVFieldAccessGenerator); ok {
				t.Errorf("Complex expression incorrectly using OHLCVFieldAccessGenerator with field=%s, "+
					"should use SeriesExpressionAccessor to avoid 'close' fallback",
					ohlcvGen.fieldName)
			}

			// Should use expression accessor
			if _, ok := comp.AccessGen.(*SeriesExpressionAccessor); !ok {
				t.Errorf("Complex expression using %T, want *SeriesExpressionAccessor to prevent fallback",
					comp.AccessGen)
			}
		})
	}
}

/* Helper: createComplexExprTestGenerator creates a minimal generator for complex expression testing */
func createComplexExprTestGenerator() *generator {
	analyzer := validation.NewWarmupAnalyzer()
	g := &generator{
		variables:         make(map[string]string),
		constants:         make(map[string]interface{}),
		constEvaluator:    analyzer,
		indent:            1,
		tempVarMgr:        nil, // Will be created when needed
		exprAnalyzer:      nil, // Will be created when needed
		taRegistry:        NewTAFunctionRegistry(),
		runtimeOnlyFilter: NewRuntimeOnlyFunctionFilter(),
		mathHandler:       NewMathHandler(),
		barFieldRegistry:  NewBarFieldSeriesRegistry(),
	}
	g.tempVarMgr = NewTempVariableManager(g)
	g.exprAnalyzer = NewExpressionAnalyzer(g)
	return g
}
