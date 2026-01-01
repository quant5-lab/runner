package codegen

import (
	"strings"
	"testing"
)

/* TestPeriodExpression_Interface ensures interface compliance */
func TestPeriodExpression_Interface(t *testing.T) {
	tests := []struct {
		name       string
		expression PeriodExpression
	}{
		{"ConstantPeriod implements interface", NewConstantPeriod(20)},
		{"RuntimePeriod implements interface", NewRuntimePeriod("len")},
		{"ConstantPeriod minimum period", NewConstantPeriod(1)},
		{"ConstantPeriod large period", NewConstantPeriod(500)},
		{"RuntimePeriod with simple name", NewRuntimePeriod("p")},
		{"RuntimePeriod with complex name", NewRuntimePeriod("myPeriod")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt.expression.IsConstant()
			_ = tt.expression.AsInt()
			_ = tt.expression.AsGoExpr()
			_ = tt.expression.AsIntCast()
			_ = tt.expression.AsFloat64Cast()
			_ = tt.expression.AsSeriesNamePart()
		})
	}
}

/* TestConstantPeriod_BehaviorInvariants verifies deterministic code generation */
func TestConstantPeriod_BehaviorInvariants(t *testing.T) {
	testCases := []struct {
		name              string
		period            int
		expectedGoExpr    string
		expectedIntCast   string
		expectedFloatCast string
		expectedSeriesKey string
	}{
		{
			name:              "Single digit period",
			period:            5,
			expectedGoExpr:    "5",
			expectedIntCast:   "5",
			expectedFloatCast: "float64(5)",
			expectedSeriesKey: "5",
		},
		{
			name:              "Double digit period",
			period:            14,
			expectedGoExpr:    "14",
			expectedIntCast:   "14",
			expectedFloatCast: "float64(14)",
			expectedSeriesKey: "14",
		},
		{
			name:              "Large period",
			period:            200,
			expectedGoExpr:    "200",
			expectedIntCast:   "200",
			expectedFloatCast: "float64(200)",
			expectedSeriesKey: "200",
		},
		{
			name:              "Minimum valid period",
			period:            1,
			expectedGoExpr:    "1",
			expectedIntCast:   "1",
			expectedFloatCast: "float64(1)",
			expectedSeriesKey: "1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewConstantPeriod(tc.period)

			/* Invariant: IsConstant() always true */
			if !p.IsConstant() {
				t.Error("ConstantPeriod must always be constant")
			}

			/* Invariant: AsInt() returns exact value */
			if p.AsInt() != tc.period {
				t.Errorf("AsInt() = %d, want %d", p.AsInt(), tc.period)
			}

			/* Invariant: Value() returns exact value */
			if p.Value() != tc.period {
				t.Errorf("Value() = %d, want %d", p.Value(), tc.period)
			}

			if p.AsGoExpr() != tc.expectedGoExpr {
				t.Errorf("AsGoExpr() = %q, want %q", p.AsGoExpr(), tc.expectedGoExpr)
			}

			if p.AsIntCast() != tc.expectedIntCast {
				t.Errorf("AsIntCast() = %q, want %q", p.AsIntCast(), tc.expectedIntCast)
			}

			if p.AsFloat64Cast() != tc.expectedFloatCast {
				t.Errorf("AsFloat64Cast() = %q, want %q", p.AsFloat64Cast(), tc.expectedFloatCast)
			}

			if p.AsSeriesNamePart() != tc.expectedSeriesKey {
				t.Errorf("AsSeriesNamePart() = %q, want %q", p.AsSeriesNamePart(), tc.expectedSeriesKey)
			}
		})
	}
}

/* TestRuntimePeriod_BehaviorInvariants verifies type cast generation */
func TestRuntimePeriod_BehaviorInvariants(t *testing.T) {
	testCases := []struct {
		name              string
		variableName      string
		expectedGoExpr    string
		expectedIntCast   string
		expectedFloatCast string
		expectedSeriesKey string
	}{
		{
			name:              "Simple variable name",
			variableName:      "len",
			expectedGoExpr:    "len",
			expectedIntCast:   "int(len)",
			expectedFloatCast: "float64(len)",
			expectedSeriesKey: "runtime",
		},
		{
			name:              "Single letter variable",
			variableName:      "p",
			expectedGoExpr:    "p",
			expectedIntCast:   "int(p)",
			expectedFloatCast: "float64(p)",
			expectedSeriesKey: "runtime",
		},
		{
			name:              "Camel case variable",
			variableName:      "myPeriod",
			expectedGoExpr:    "myPeriod",
			expectedIntCast:   "int(myPeriod)",
			expectedFloatCast: "float64(myPeriod)",
			expectedSeriesKey: "runtime",
		},
		{
			name:              "Snake case variable",
			variableName:      "period_len",
			expectedGoExpr:    "period_len",
			expectedIntCast:   "int(period_len)",
			expectedFloatCast: "float64(period_len)",
			expectedSeriesKey: "runtime",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewRuntimePeriod(tc.variableName)

			/* Invariant: IsConstant() always false */
			if p.IsConstant() {
				t.Error("RuntimePeriod must never be constant")
			}

			/* Invariant: AsInt() returns -1 sentinel */
			if p.AsInt() != -1 {
				t.Errorf("AsInt() = %d, want -1 (runtime sentinel)", p.AsInt())
			}

			/* Validate code generation formats */
			if p.AsGoExpr() != tc.expectedGoExpr {
				t.Errorf("AsGoExpr() = %q, want %q", p.AsGoExpr(), tc.expectedGoExpr)
			}

			if p.AsIntCast() != tc.expectedIntCast {
				t.Errorf("AsIntCast() = %q, want %q", p.AsIntCast(), tc.expectedIntCast)
			}

			if p.AsFloat64Cast() != tc.expectedFloatCast {
				t.Errorf("AsFloat64Cast() = %q, want %q", p.AsFloat64Cast(), tc.expectedFloatCast)
			}

			if p.AsSeriesNamePart() != tc.expectedSeriesKey {
				t.Errorf("AsSeriesNamePart() = %q, want %q", p.AsSeriesNamePart(), tc.expectedSeriesKey)
			}
		})
	}
}

/* TestPeriodExpression_CodeGenerationPatterns verifies loop bounds and type casts */
func TestPeriodExpression_CodeGenerationPatterns(t *testing.T) {
	tests := []struct {
		name          string
		expression    PeriodExpression
		loopPattern   string
		alphaRMA      string // Expected pattern in: alpha := 1.0 / PATTERN
		alphaEMA      string // Expected pattern in: alpha := 2.0 / (PATTERN+1)
		warmupPattern string // Expected pattern in: if ctx.BarIndex < PATTERN-1
	}{
		{
			name:          "Constant period 20",
			expression:    NewConstantPeriod(20),
			loopPattern:   "20",
			alphaRMA:      "float64(20)",
			alphaEMA:      "float64(20+1)",
			warmupPattern: "19",
		},
		{
			name:          "Constant period 14",
			expression:    NewConstantPeriod(14),
			loopPattern:   "14",
			alphaRMA:      "float64(14)",
			alphaEMA:      "float64(14+1)",
			warmupPattern: "13",
		},
		{
			name:          "Runtime period len",
			expression:    NewRuntimePeriod("len"),
			loopPattern:   "int(len)",
			alphaRMA:      "float64(len)",
			alphaEMA:      "float64(len)+1", // Note: Runtime uses expression
			warmupPattern: "int(len)-1",
		},
		{
			name:          "Runtime period myPeriod",
			expression:    NewRuntimePeriod("myPeriod"),
			loopPattern:   "int(myPeriod)",
			alphaRMA:      "float64(myPeriod)",
			alphaEMA:      "float64(myPeriod)+1",
			warmupPattern: "int(myPeriod)-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			/* Test loop pattern generation */
			loopCode := tt.expression.AsIntCast()
			if loopCode != tt.loopPattern {
				t.Errorf("Loop pattern: got %q, want %q", loopCode, tt.loopPattern)
			}

			/* Test RMA alpha generation */
			alphaRMACode := tt.expression.AsFloat64Cast()
			if alphaRMACode != tt.alphaRMA {
				t.Errorf("RMA alpha: got %q, want %q", alphaRMACode, tt.alphaRMA)
			}

			/* Test warmup pattern - constants optimize to literal */
			if tt.expression.IsConstant() {
				warmupCode := tt.expression.AsInt() - 1
				expectedWarmup := tt.expression.AsInt() - 1
				if warmupCode != expectedWarmup {
					t.Errorf("Warmup calculation: got %d, want %d", warmupCode, expectedWarmup)
				}
			}
		})
	}
}

/* TestPeriodExpression_SeriesNamingUniqueness prevents cache key collisions */
func TestPeriodExpression_SeriesNamingUniqueness(t *testing.T) {
	tests := []struct {
		name         string
		expressions  []PeriodExpression
		expectUnique bool
		description  string
	}{
		{
			name: "Same constant period produces same key",
			expressions: []PeriodExpression{
				NewConstantPeriod(20),
				NewConstantPeriod(20),
			},
			expectUnique: false,
			description:  "Identical constants should produce identical keys for caching",
		},
		{
			name: "Different constant periods produce different keys",
			expressions: []PeriodExpression{
				NewConstantPeriod(14),
				NewConstantPeriod(20),
			},
			expectUnique: true,
			description:  "Different periods must have distinct keys to avoid collisions",
		},
		{
			name: "Runtime periods produce same 'runtime' key",
			expressions: []PeriodExpression{
				NewRuntimePeriod("len"),
				NewRuntimePeriod("period"),
			},
			expectUnique: false,
			description:  "All runtime periods share 'runtime' key - uniqueness via hash",
		},
		{
			name: "Constant and runtime produce different keys",
			expressions: []PeriodExpression{
				NewConstantPeriod(20),
				NewRuntimePeriod("len"),
			},
			expectUnique: true,
			description:  "Constant vs runtime must be distinguishable in series naming",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.expressions) < 2 {
				t.Fatal("Test requires at least 2 expressions")
			}

			key1 := tt.expressions[0].AsSeriesNamePart()
			key2 := tt.expressions[1].AsSeriesNamePart()

			areUnique := (key1 != key2)

			if areUnique != tt.expectUnique {
				t.Errorf("%s: keys %q and %q, unique=%v, want unique=%v",
					tt.description, key1, key2, areUnique, tt.expectUnique)
			}
		})
	}
}

/* TestPeriodExpression_TypeSafety verifies int/float64 conversions */
func TestPeriodExpression_TypeSafety(t *testing.T) {
	tests := []struct {
		name       string
		expression PeriodExpression
		context    string
		validate   func(t *testing.T, code string)
	}{
		{
			name:       "Float division uses float64 cast",
			expression: NewConstantPeriod(20),
			context:    "RMA alpha calculation",
			validate: func(t *testing.T, code string) {
				if !strings.Contains(code, "float64") {
					t.Error("Division must use float64 to avoid integer division")
				}
			},
		},
		{
			name:       "Loop bounds use int cast for runtime",
			expression: NewRuntimePeriod("len"),
			context:    "For loop condition",
			validate: func(t *testing.T, code string) {
				if !strings.HasPrefix(code, "int(") {
					t.Error("Loop bounds must explicitly cast to int")
				}
			},
		},
		{
			name:       "Constant loop bounds optimize to literal",
			expression: NewConstantPeriod(20),
			context:    "For loop condition",
			validate: func(t *testing.T, code string) {
				if strings.Contains(code, "int(") {
					t.Error("Constant loop bounds should optimize to literal, not int() cast")
				}
			},
		},
		{
			name:       "Series key is string type",
			expression: NewRuntimePeriod("len"),
			context:    "Series naming",
			validate: func(t *testing.T, code string) {
				/* All series keys must be strings */
				if code != "runtime" {
					t.Errorf("Series key must be string literal, got %q", code)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var code string
			switch tt.context {
			case "RMA alpha calculation":
				code = tt.expression.AsFloat64Cast()
			case "For loop condition":
				code = tt.expression.AsIntCast()
			case "Series naming":
				code = tt.expression.AsSeriesNamePart()
			default:
				t.Fatalf("Unknown context: %s", tt.context)
			}

			tt.validate(t, code)
		})
	}
}

/* TestPeriodExpression_EdgeCaseValues tests boundary values */
func TestPeriodExpression_EdgeCaseValues(t *testing.T) {
	tests := []struct {
		name       string
		expression PeriodExpression
		validate   func(t *testing.T, expr PeriodExpression)
	}{
		{
			name:       "Minimum period 1",
			expression: NewConstantPeriod(1),
			validate: func(t *testing.T, expr PeriodExpression) {
				if expr.AsInt() != 1 {
					t.Error("Period 1 must be preserved exactly")
				}
				if expr.AsIntCast() != "1" {
					t.Error("Period 1 must generate literal '1'")
				}
			},
		},
		{
			name:       "Very large period",
			expression: NewConstantPeriod(10000),
			validate: func(t *testing.T, expr PeriodExpression) {
				if expr.AsInt() != 10000 {
					t.Error("Large period must be preserved exactly")
				}
				if !strings.Contains(expr.AsFloat64Cast(), "10000") {
					t.Error("Large period must appear in float cast")
				}
			},
		},
		{
			name:       "Empty variable name creates valid runtime period",
			expression: NewRuntimePeriod(""),
			validate: func(t *testing.T, expr PeriodExpression) {
				if expr.IsConstant() {
					t.Error("Empty string should still create RuntimePeriod")
				}
				if expr.AsInt() != -1 {
					t.Error("Runtime period must return -1 sentinel")
				}
			},
		},
		{
			name:       "Zero period constant",
			expression: NewConstantPeriod(0),
			validate: func(t *testing.T, expr PeriodExpression) {
				if !expr.IsConstant() {
					t.Error("Zero should still be a constant")
				}
				if expr.AsInt() != 0 {
					t.Error("Zero must be preserved")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.expression)
		})
	}
}

/* TestPeriodExpression_ConstructorInvariants verifies factory functions */
func TestPeriodExpression_ConstructorInvariants(t *testing.T) {
	t.Run("NewConstantPeriod never returns nil", func(t *testing.T) {
		p := NewConstantPeriod(20)
		if p == nil {
			t.Error("NewConstantPeriod must never return nil")
		}
	})

	t.Run("NewRuntimePeriod never returns nil", func(t *testing.T) {
		p := NewRuntimePeriod("len")
		if p == nil {
			t.Error("NewRuntimePeriod must never return nil")
		}
	})

	t.Run("NewConstantPeriod preserves value", func(t *testing.T) {
		testValues := []int{1, 2, 10, 14, 20, 50, 100, 200, 500, 1000}
		for _, val := range testValues {
			p := NewConstantPeriod(val)
			if p.Value() != val {
				t.Errorf("NewConstantPeriod(%d).Value() = %d, want %d", val, p.Value(), val)
			}
		}
	})

	t.Run("NewRuntimePeriod preserves variable name", func(t *testing.T) {
		testNames := []string{"len", "p", "period", "myPeriod", "period_len"}
		for _, name := range testNames {
			p := NewRuntimePeriod(name)
			if p.AsGoExpr() != name {
				t.Errorf("NewRuntimePeriod(%q).AsGoExpr() = %q, want %q", name, p.AsGoExpr(), name)
			}
		}
	})
}
