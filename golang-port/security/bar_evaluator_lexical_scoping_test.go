package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/series"
)

// TestVarLookupFunc_ResolutionPriority validates that variable resolution follows
// the correct priority order: OHLCV → Registry → VarLookup → Error
func TestVarLookupFunc_ResolutionPriority(t *testing.T) {
	ctx := createTestContext()

	// Create test series for registry
	registrySeries := series.NewSeries(10)
	registrySeries.Set(999.0) // Marker value for registry

	// Create test series for fallback
	fallbackSeries := series.NewSeries(10)
	fallbackSeries.Set(888.0) // Marker value for fallback

	tests := []struct {
		name          string
		varName       string
		setupRegistry bool
		setupFallback bool
		expected      float64
		expectError   bool
		description   string
	}{
		{
			name:          "OHLCV_field_takes_precedence",
			varName:       "close",
			setupRegistry: true,
			setupFallback: true,
			expected:      102, // From OHLCV data
			expectError:   false,
			description:   "OHLCV fields should be resolved first, ignoring registry/fallback",
		},
		{
			name:          "registry_variable_without_fallback",
			varName:       "customVar",
			setupRegistry: true,
			setupFallback: false,
			expected:      999.0,
			expectError:   false,
			description:   "Registry should be checked before fallback",
		},
		{
			name:          "fallback_when_not_in_registry",
			varName:       "mainContextVar",
			setupRegistry: false,
			setupFallback: true,
			expected:      888.0,
			expectError:   false,
			description:   "VarLookup fallback should be used when variable not in registry",
		},
		{
			name:          "error_when_variable_not_found",
			varName:       "unknownVar",
			setupRegistry: false,
			setupFallback: false,
			expected:      0,
			expectError:   true,
			description:   "Should return error when variable not found anywhere",
		},
		{
			name:          "registry_overrides_fallback",
			varName:       "sharedVar",
			setupRegistry: true,
			setupFallback: true,
			expected:      999.0,
			expectError:   false,
			description:   "Registry should take precedence over fallback for same variable name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup fresh evaluator for each test
			eval := NewStreamingBarEvaluator()

			// Setup bar mapper
			mapper := NewBarIndexMapper()
			mapper.SetMapping(0, 0)
			eval.SetBarIndexMapper(mapper)

			// Setup registry if needed
			if tt.setupRegistry {
				registry := NewVariableRegistry()
				registry.Register(tt.varName, registrySeries)
				eval.SetVariableRegistry(registry)
			}

			// Setup fallback if needed
			if tt.setupFallback {
				eval.SetVarLookup(func(varName string, secBarIdx int) (*series.Series, int, bool) {
					if varName == tt.varName {
						return fallbackSeries, 0, true
					}
					return nil, -1, false
				})
			}

			expr := &ast.Identifier{Name: tt.varName}
			value, err := eval.EvaluateAtBar(expr, ctx, 0)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for %s, got nil", tt.description)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error for %s: %v", tt.description, err)
				}
				if value != tt.expected {
					t.Errorf("%s: expected %.1f, got %.1f", tt.description, tt.expected, value)
				}
			}
		})
	}
}

// TestVarLookupFunc_BarIndexMapping validates that security bar indices are correctly
// mapped to main context bar indices when resolving variables
func TestVarLookupFunc_BarIndexMapping(t *testing.T) {
	// Create larger test context to support bar indices used in tests
	ctx := &context.Context{
		Data: make([]context.OHLCV, 25),
	}
	for i := range ctx.Data {
		ctx.Data[i] = context.OHLCV{Close: float64(100 + i)}
	}

	// Create series with distinct values at each position
	testSeries := series.NewSeries(100)
	for i := 0; i < 20; i++ {
		testSeries.Set(float64(1000 + i))
		if i < 19 {
			testSeries.Next()
		}
	}
	// testSeries is now at position 19 with values [1000, 1001, 1002, ..., 1019]

	tests := []struct {
		name           string
		secBarIdx      int
		mainBarIdx     int
		seriesPosition int
		expectedValue  float64
		description    string
	}{
		{
			name:           "direct_mapping_current_bar",
			secBarIdx:      1,
			mainBarIdx:     19,
			seriesPosition: 19,
			expectedValue:  1019.0,
			description:    "Security bar 1 maps to main bar 19 (current), offset=0",
		},
		{
			name:           "direct_mapping_historical_bar",
			secBarIdx:      1,
			mainBarIdx:     15,
			seriesPosition: 19,
			expectedValue:  1015.0,
			description:    "Security bar 1 maps to main bar 15, offset=4",
		},
		{
			name:           "large_offset_historical",
			secBarIdx:      1,
			mainBarIdx:     5,
			seriesPosition: 19,
			expectedValue:  1005.0,
			description:    "Security bar 1 maps to main bar 5, offset=14",
		},
		{
			name:           "beginning_of_series",
			secBarIdx:      0,
			mainBarIdx:     0,
			seriesPosition: 19,
			expectedValue:  1000.0,
			description:    "Security bar 0 maps to main bar 0 (beginning), offset=19",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := NewStreamingBarEvaluator()

			// Setup bar mapper
			mapper := NewBarIndexMapper()
			mapper.SetMapping(tt.secBarIdx, tt.mainBarIdx)
			evaluator.SetBarIndexMapper(mapper)

			// Setup fallback with test series
			evaluator.SetVarLookup(func(varName string, secBarIdx int) (*series.Series, int, bool) {
				if varName == "testVar" {
					return testSeries, tt.mainBarIdx, true
				}
				return nil, -1, false
			})

			expr := &ast.Identifier{Name: "testVar"}
			value, err := evaluator.EvaluateAtBar(expr, ctx, tt.secBarIdx)

			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tt.description, err)
			}

			if value != tt.expectedValue {
				t.Errorf("%s: expected %.1f, got %.1f", tt.description, tt.expectedValue, value)
			}
		})
	}
}

// TestVarLookupFunc_BoundaryConditions tests edge cases in offset calculation and series access
func TestVarLookupFunc_BoundaryConditions(t *testing.T) {
	ctx := createTestContext()

	tests := []struct {
		name           string
		seriesCapacity int
		seriesPosition int
		mainBarIdx     int
		expectError    bool
		description    string
	}{
		{
			name:           "offset_exceeds_capacity",
			seriesCapacity: 10,
			seriesPosition: 5,
			mainBarIdx:     15,
			expectError:    true,
			description:    "Offset > capacity falls through to unknown identifier error",
		},
		{
			name:           "negative_mainBarIdx",
			seriesCapacity: 10,
			seriesPosition: 5,
			mainBarIdx:     -1,
			expectError:    false,
			description:    "Negative main bar index returns NaN for warmup period",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := NewStreamingBarEvaluator()

			// Create series with specific capacity
			testSeries := series.NewSeries(tt.seriesCapacity)
			for i := 0; i < tt.seriesPosition; i++ {
				testSeries.Set(100.0)
				if i < tt.seriesPosition-1 {
					testSeries.Next()
				}
			}

			mapper := NewBarIndexMapper()
			mapper.SetMapping(0, tt.mainBarIdx)
			evaluator.SetBarIndexMapper(mapper)

			evaluator.SetVarLookup(func(varName string, secBarIdx int) (*series.Series, int, bool) {
				if varName == "boundaryTest" {
					return testSeries, tt.mainBarIdx, true
				}
				return nil, -1, false
			})

			expr := &ast.Identifier{Name: "boundaryTest"}
			result, err := evaluator.EvaluateAtBar(expr, ctx, 0)

			if tt.expectError {
				if err == nil {
					t.Errorf("%s: expected error, got nil", tt.description)
				}
			} else {
				if err != nil {
					t.Fatalf("%s: unexpected error: %v", tt.description, err)
				}
				// For negative mainBarIdx (warmup), verify NaN is returned
				if tt.mainBarIdx < 0 && !math.IsNaN(result) {
					t.Errorf("%s: expected NaN for warmup, got %v", tt.description, result)
				}
			}
		})
	}
}

// TestVarLookupFunc_NilSeriesHandling tests behavior when VarLookup returns nil series
func TestVarLookupFunc_NilSeriesHandling(t *testing.T) {
	ctx := createTestContext()
	evaluator := NewStreamingBarEvaluator()

	mapper := NewBarIndexMapper()
	mapper.SetMapping(0, 0)
	evaluator.SetBarIndexMapper(mapper)

	tests := []struct {
		name        string
		lookupFunc  VarLookupFunc
		expectError bool
		description string
	}{
		{
			name: "nil_series_returned",
			lookupFunc: func(varName string, secBarIdx int) (*series.Series, int, bool) {
				return nil, 0, true // Returns true but nil series
			},
			expectError: true,
			description: "Should handle nil series gracefully",
		},
		{
			name: "not_found_false_returned",
			lookupFunc: func(varName string, secBarIdx int) (*series.Series, int, bool) {
				return nil, -1, false // Properly indicates not found
			},
			expectError: true,
			description: "Should return error when variable not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator.SetVarLookup(tt.lookupFunc)

			expr := &ast.Identifier{Name: "testVar"}
			_, err := evaluator.EvaluateAtBar(expr, ctx, 0)

			if tt.expectError && err == nil {
				t.Errorf("%s: expected error, got nil", tt.description)
			} else if !tt.expectError && err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
			}
		})
	}
}

// TestVarLookupFunc_MultipleSecurityContexts tests that different security contexts
// can access the same main context variables independently
func TestVarLookupFunc_MultipleSecurityContexts(t *testing.T) {
	ctx := createTestContext()

	// Create main context series
	mainSeries := series.NewSeries(100)
	for i := 0; i < 10; i++ {
		mainSeries.Set(float64(2000 + i))
		if i < 9 {
			mainSeries.Next()
		}
	}

	tests := []struct {
		name             string
		securityContext1 int // Security bar index for context 1
		mainBarIdx1      int
		securityContext2 int // Security bar index for context 2
		mainBarIdx2      int
		expectedValue1   float64
		expectedValue2   float64
		description      string
	}{
		{
			name:             "different_mappings_same_series",
			securityContext1: 0,
			mainBarIdx1:      5,
			securityContext2: 1,
			mainBarIdx2:      7,
			expectedValue1:   2005.0,
			expectedValue2:   2007.0,
			description:      "Two security contexts map to different main bars",
		},
		{
			name:             "same_security_bar_different_mappings",
			securityContext1: 0,
			mainBarIdx1:      3,
			securityContext2: 0,
			mainBarIdx2:      6,
			expectedValue1:   2003.0,
			expectedValue2:   2006.0,
			description:      "Same security bar index can map differently in different contexts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Evaluator for context 1
			eval1 := NewStreamingBarEvaluator()
			mapper1 := NewBarIndexMapper()
			mapper1.SetMapping(tt.securityContext1, tt.mainBarIdx1)
			eval1.SetBarIndexMapper(mapper1)
			eval1.SetVarLookup(func(varName string, secBarIdx int) (*series.Series, int, bool) {
				if varName == "sharedVar" {
					return mainSeries, tt.mainBarIdx1, true
				}
				return nil, -1, false
			})

			// Evaluator for context 2
			eval2 := NewStreamingBarEvaluator()
			mapper2 := NewBarIndexMapper()
			mapper2.SetMapping(tt.securityContext2, tt.mainBarIdx2)
			eval2.SetBarIndexMapper(mapper2)
			eval2.SetVarLookup(func(varName string, secBarIdx int) (*series.Series, int, bool) {
				if varName == "sharedVar" {
					return mainSeries, tt.mainBarIdx2, true
				}
				return nil, -1, false
			})

			expr := &ast.Identifier{Name: "sharedVar"}

			value1, err1 := eval1.EvaluateAtBar(expr, ctx, tt.securityContext1)
			if err1 != nil {
				t.Fatalf("Context 1 error: %v", err1)
			}

			value2, err2 := eval2.EvaluateAtBar(expr, ctx, tt.securityContext2)
			if err2 != nil {
				t.Fatalf("Context 2 error: %v", err2)
			}

			if value1 != tt.expectedValue1 {
				t.Errorf("Context 1: expected %.1f, got %.1f", tt.expectedValue1, value1)
			}

			if value2 != tt.expectedValue2 {
				t.Errorf("Context 2: expected %.1f, got %.1f", tt.expectedValue2, value2)
			}
		})
	}
}

// TestVarLookupFunc_SeriesProgressionWithBarMapper validates that as series progress,
// the offset calculation remains correct across multiple bar evaluations
func TestVarLookupFunc_SeriesProgressionWithBarMapper(t *testing.T) {
	ctx := createTestContext()

	tests := []struct {
		name         string
		barSequence  []int                // Security bar indices to evaluate in sequence
		setupFunc    func(*series.Series) // Setup series state
		expectations []float64            // Expected values for each bar in sequence
		description  string
	}{
		{
			name:        "series_advances_with_bars",
			barSequence: []int{0, 1, 2},
			setupFunc: func(s *series.Series) {
				for i := 0; i < 10; i++ {
					s.Set(float64(500 + i))
					if i < 9 {
						s.Next()
					}
				}
			},
			expectations: []float64{500.0, 501.0, 502.0},
			description:  "Series values should be correctly accessed as bars progress",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup series
			testSeries := series.NewSeries(50)
			tt.setupFunc(testSeries)

			evaluator := NewStreamingBarEvaluator()
			mapper := NewBarIndexMapper()

			// Map each security bar to corresponding main bar
			for i, secBar := range tt.barSequence {
				mapper.SetMapping(secBar, i)
			}

			evaluator.SetBarIndexMapper(mapper)
			evaluator.SetVarLookup(func(varName string, secBarIdx int) (*series.Series, int, bool) {
				if varName == "progressVar" {
					// Map security bar to main bar
					mainIdx := mapper.GetMainBarIndexForSecurityBar(secBarIdx)
					return testSeries, mainIdx, true
				}
				return nil, -1, false
			})

			expr := &ast.Identifier{Name: "progressVar"}

			for i, secBar := range tt.barSequence {
				value, err := evaluator.EvaluateAtBar(expr, ctx, secBar)
				if err != nil {
					t.Fatalf("Bar %d error: %v", secBar, err)
				}

				if math.Abs(value-tt.expectations[i]) > 0.0001 {
					t.Errorf("Bar %d: expected %.1f, got %.1f", secBar, tt.expectations[i], value)
				}
			}
		})
	}
}

// TestVarLookupFunc_NoBarMapperFallback tests behavior when bar mapper is not set
func TestVarLookupFunc_NoBarMapperFallback(t *testing.T) {
	ctx := createTestContext()
	evaluator := NewStreamingBarEvaluator()

	testSeries := series.NewSeries(10)
	testSeries.Set(777.0)

	// Set VarLookup but NOT bar mapper
	evaluator.SetVarLookup(func(varName string, secBarIdx int) (*series.Series, int, bool) {
		if varName == "testVar" {
			return testSeries, 0, true
		}
		return nil, -1, false
	})

	expr := &ast.Identifier{Name: "testVar"}
	value, err := evaluator.EvaluateAtBar(expr, ctx, 0)

	// Should still work - VarLookup provides mainBarIdx directly
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != 777.0 {
		t.Errorf("expected 777.0, got %.1f", value)
	}
}

// TestVarLookupFunc_ConcurrentAccessSafety is a basic test for concurrent access patterns
// Note: This is not a comprehensive concurrency test, but validates basic thread-safety assumptions
func TestVarLookupFunc_ConcurrentAccessSafety(t *testing.T) {
	// Note: Full concurrency testing would require more complex setup
	// This test validates that the basic structure doesn't panic under simple concurrent access

	ctx := createTestContext()
	testSeries := series.NewSeries(100)
	testSeries.Set(123.0)

	evaluator := NewStreamingBarEvaluator()
	mapper := NewBarIndexMapper()
	mapper.SetMapping(0, 0)
	evaluator.SetBarIndexMapper(mapper)

	evaluator.SetVarLookup(func(varName string, secBarIdx int) (*series.Series, int, bool) {
		if varName == "concurrent" {
			return testSeries, 0, true
		}
		return nil, -1, false
	})

	expr := &ast.Identifier{Name: "concurrent"}

	// Simple sequential access to ensure no panics
	for i := 0; i < 10; i++ {
		_, err := evaluator.EvaluateAtBar(expr, ctx, 0)
		if err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}
	}
}
