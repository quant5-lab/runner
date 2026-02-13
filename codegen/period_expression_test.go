package codegen

import (
	"testing"
)

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
		{
			name:              "Zero period boundary",
			period:            0,
			expectedGoExpr:    "0",
			expectedIntCast:   "0",
			expectedFloatCast: "float64(0)",
			expectedSeriesKey: "0",
		},
		{
			name:              "Very large period",
			period:            10000,
			expectedGoExpr:    "10000",
			expectedIntCast:   "10000",
			expectedFloatCast: "float64(10000)",
			expectedSeriesKey: "10000",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewConstantPeriod(tc.period)

			if !p.IsConstant() {
				t.Error("ConstantPeriod must always be constant")
			}

			if p.AsInt() != tc.period {
				t.Errorf("AsInt() = %d, want %d", p.AsInt(), tc.period)
			}

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
		{
			name:              "Empty variable name boundary",
			variableName:      "",
			expectedGoExpr:    "",
			expectedIntCast:   "int()",
			expectedFloatCast: "float64()",
			expectedSeriesKey: "runtime",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewRuntimePeriod(tc.variableName)

			if p.IsConstant() {
				t.Error("RuntimePeriod must never be constant")
			}

			if p.AsInt() != -1 {
				t.Errorf("AsInt() = %d, want -1 (runtime sentinel)", p.AsInt())
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

/* TestComputedPeriod_BehaviorInvariants verifies pre-rendered Go expression wrapping */
func TestComputedPeriod_BehaviorInvariants(t *testing.T) {
	testCases := []struct {
		name              string
		goExpression      string
		expectedGoExpr    string
		expectedIntCast   string
		expectedFloatCast string
		expectedSeriesKey string
	}{
		{
			name:              "Binary arithmetic expression",
			goExpression:      "(lengthSeries.GetCurrent() / 2)",
			expectedGoExpr:    "(lengthSeries.GetCurrent() / 2)",
			expectedIntCast:   "int((lengthSeries.GetCurrent() / 2))",
			expectedFloatCast: "float64((lengthSeries.GetCurrent() / 2))",
			expectedSeriesKey: "computed",
		},
		{
			name:              "Nested function call",
			goExpression:      "math.Round(math.Sqrt(nSeries.GetCurrent()))",
			expectedGoExpr:    "math.Round(math.Sqrt(nSeries.GetCurrent()))",
			expectedIntCast:   "int(math.Round(math.Sqrt(nSeries.GetCurrent())))",
			expectedFloatCast: "float64(math.Round(math.Sqrt(nSeries.GetCurrent())))",
			expectedSeriesKey: "computed",
		},
		{
			name:              "Simple series access",
			goExpression:      "pSeries.GetCurrent()",
			expectedGoExpr:    "pSeries.GetCurrent()",
			expectedIntCast:   "int(pSeries.GetCurrent())",
			expectedFloatCast: "float64(pSeries.GetCurrent())",
			expectedSeriesKey: "computed",
		},
		{
			name:              "Compound arithmetic with multiple operators",
			goExpression:      "(aSeries.GetCurrent() * 2 + 1)",
			expectedGoExpr:    "(aSeries.GetCurrent() * 2 + 1)",
			expectedIntCast:   "int((aSeries.GetCurrent() * 2 + 1))",
			expectedFloatCast: "float64((aSeries.GetCurrent() * 2 + 1))",
			expectedSeriesKey: "computed",
		},
		{
			name:              "Empty expression boundary",
			goExpression:      "",
			expectedGoExpr:    "",
			expectedIntCast:   "int()",
			expectedFloatCast: "float64()",
			expectedSeriesKey: "computed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewComputedPeriod(tc.goExpression)

			if p.IsConstant() {
				t.Error("ComputedPeriod must never be constant")
			}

			if p.AsInt() != -1 {
				t.Errorf("AsInt() = %d, want -1 (computed sentinel)", p.AsInt())
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
		{
			name: "Computed periods share computed key",
			expressions: []PeriodExpression{
				NewComputedPeriod("(a / 2)"),
				NewComputedPeriod("math.Sqrt(b)"),
			},
			expectUnique: false,
			description:  "All computed periods share 'computed' key - uniqueness via hash",
		},
		{
			name: "Constant and computed produce different keys",
			expressions: []PeriodExpression{
				NewConstantPeriod(20),
				NewComputedPeriod("(n / 2)"),
			},
			expectUnique: true,
			description:  "Constant vs computed must be distinguishable in series naming",
		},
		{
			name: "Runtime and computed produce different keys",
			expressions: []PeriodExpression{
				NewRuntimePeriod("len"),
				NewComputedPeriod("(len / 2)"),
			},
			expectUnique: true,
			description:  "Runtime vs computed must be distinguishable in series naming",
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
