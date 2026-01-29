package codegen

import (
	"strings"
	"testing"
)

/*
TestArrowFunctionTACallGenerator_TupleIndicatorDetection validates routing
of multi-output indicators to tuple IIFE generation.

Edge cases:
- All registered tuple functions (ta.macd, ta.bb, ta.stoch, ta.dmi, ta.kc, ta.supertrend)
- With and without ta. prefix (macd vs ta.macd)
- Mixed with single-output indicators
- Tuple functions with various argument counts
*/
func TestArrowFunctionTACallGenerator_TupleIndicatorDetection(t *testing.T) {
	tests := []struct {
		name          string
		funcName      string
		isTuple       bool
		outputCount   int
		expectedRoute string
		desc          string
	}{
		{
			name:          "ta.macd_is_tuple",
			funcName:      "ta.macd",
			isTuple:       true,
			outputCount:   3,
			expectedRoute: "tuple IIFE",
			desc:          "ta.macd returns (macdLine, signal, histogram)",
		},
		{
			name:          "macd_without_prefix",
			funcName:      "macd",
			isTuple:       true,
			outputCount:   3,
			expectedRoute: "tuple IIFE",
			desc:          "macd alias should also route to tuple",
		},
		{
			name:          "ta.bb_is_tuple",
			funcName:      "ta.bb",
			isTuple:       true,
			outputCount:   3,
			expectedRoute: "tuple IIFE",
			desc:          "ta.bb returns (upper, basis, lower)",
		},
		{
			name:          "ta.stoch_is_tuple",
			funcName:      "ta.stoch",
			isTuple:       true,
			outputCount:   2,
			expectedRoute: "tuple IIFE",
			desc:          "ta.stoch returns (k, d)",
		},
		{
			name:          "ta.dmi_is_tuple",
			funcName:      "ta.dmi",
			isTuple:       true,
			outputCount:   3,
			expectedRoute: "tuple IIFE",
			desc:          "ta.dmi returns (plus, minus, adx)",
		},
		{
			name:          "ta.kc_is_tuple",
			funcName:      "ta.kc",
			isTuple:       true,
			outputCount:   3,
			expectedRoute: "tuple IIFE",
			desc:          "ta.kc returns (upper, basis, lower)",
		},
		{
			name:          "ta.supertrend_is_tuple",
			funcName:      "ta.supertrend",
			isTuple:       true,
			outputCount:   2,
			expectedRoute: "tuple IIFE",
			desc:          "ta.supertrend returns (supertrend, direction)",
		},
		{
			name:          "ta.sma_not_tuple",
			funcName:      "ta.sma",
			isTuple:       false,
			outputCount:   1,
			expectedRoute: "single IIFE",
			desc:          "Single-output indicator not routed to tuple",
		},
		{
			name:          "ta.ema_not_tuple",
			funcName:      "ta.ema",
			isTuple:       false,
			outputCount:   1,
			expectedRoute: "single IIFE",
			desc:          "Single-output indicator not routed to tuple",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewTupleIndicatorRegistry()
			isRegistered := registry.IsRegistered(tt.funcName)

			if tt.isTuple && !isRegistered {
				t.Errorf("%s: expected %s to be registered as tuple, but not found", tt.desc, tt.funcName)
			}

			if !tt.isTuple && isRegistered {
				t.Errorf("%s: %s should not be registered as tuple", tt.desc, tt.funcName)
			}

			if tt.isTuple {
				spec := registry.Lookup(tt.funcName)
				if spec == nil {
					t.Fatalf("%s: spec not found for %s", tt.desc, tt.funcName)
				}
				if spec.OutputCount != tt.outputCount {
					t.Errorf("%s: expected %d outputs, got %d", tt.desc, tt.outputCount, spec.OutputCount)
				}
			}
		})
	}
}

/*
TestArrowFunctionTACallGenerator_TupleIIFESignature validates return type
signatures for tuple indicators in arrow context.

Edge cases:
- 2-tuple return types (ta.stoch, ta.supertrend)
- 3-tuple return types (ta.macd, ta.bb, ta.dmi, ta.kc)
- Runtime function mapping correctness
*/
func TestArrowFunctionTACallGenerator_TupleIIFESignature(t *testing.T) {
	tests := []struct {
		name                string
		funcName            string
		expectedOutputCount int
		expectedRuntimeFunc string
		desc                string
	}{
		{
			name:                "ta.macd_triple_output",
			funcName:            "ta.macd",
			expectedOutputCount: 3,
			expectedRuntimeFunc: "ta.Macd",
			desc:                "MACD generates 3-tuple signature",
		},
		{
			name:                "ta.bb_triple_output",
			funcName:            "ta.bb",
			expectedOutputCount: 3,
			expectedRuntimeFunc: "ta.BBands",
			desc:                "Bollinger Bands generates 3-tuple signature",
		},
		{
			name:                "ta.stoch_double_output",
			funcName:            "ta.stoch",
			expectedOutputCount: 2,
			expectedRuntimeFunc: "ta.Stoch",
			desc:                "Stochastic generates 2-tuple signature",
		},
		{
			name:                "ta.dmi_triple_output",
			funcName:            "ta.dmi",
			expectedOutputCount: 3,
			expectedRuntimeFunc: "ta.Dmi",
			desc:                "DMI generates 3-tuple signature",
		},
		{
			name:                "ta.kc_triple_output",
			funcName:            "ta.kc",
			expectedOutputCount: 3,
			expectedRuntimeFunc: "ta.KeltnerChannels",
			desc:                "Keltner Channels generates 3-tuple signature",
		},
		{
			name:                "ta.supertrend_double_output",
			funcName:            "ta.supertrend",
			expectedOutputCount: 2,
			expectedRuntimeFunc: "ta.Supertrend",
			desc:                "Supertrend generates 2-tuple signature",
		},
		{
			name:                "macd_without_ta_prefix",
			funcName:            "macd",
			expectedOutputCount: 3,
			expectedRuntimeFunc: "ta.Macd",
			desc:                "macd without prefix maps to ta.Macd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewTupleIndicatorRegistry()
			spec := registry.Lookup(tt.funcName)

			if spec == nil {
				t.Fatalf("%s: spec not found for %s", tt.desc, tt.funcName)
			}

			if spec.OutputCount != tt.expectedOutputCount {
				t.Errorf("%s: expected %d outputs, got %d", tt.desc, tt.expectedOutputCount, spec.OutputCount)
			}

			if spec.RuntimeFunction != tt.expectedRuntimeFunc {
				t.Errorf("%s: expected runtime function %s, got %s", tt.desc, tt.expectedRuntimeFunc, spec.RuntimeFunction)
			}
		})
	}
}

/*
TestArrowFunctionTACallGenerator_TupleVsSingleOutputRouting validates correct
routing logic that distinguishes tuple indicators from single-output indicators.
*/
func TestArrowFunctionTACallGenerator_TupleVsSingleOutputRouting(t *testing.T) {
	tests := []struct {
		name          string
		funcName      string
		shouldBeTuple bool
		desc          string
	}{
		{
			name:          "ta.macd_routes_to_tuple",
			funcName:      "ta.macd",
			shouldBeTuple: true,
			desc:          "MACD should route to tuple IIFE, not single-output",
		},
		{
			name:          "ta.sma_routes_to_single",
			funcName:      "ta.sma",
			shouldBeTuple: false,
			desc:          "SMA should route to single IIFE, not tuple",
		},
		{
			name:          "ta.bb_routes_to_tuple",
			funcName:      "ta.bb",
			shouldBeTuple: true,
			desc:          "Bollinger Bands should route to tuple IIFE",
		},
		{
			name:          "ta.ema_routes_to_single",
			funcName:      "ta.ema",
			shouldBeTuple: false,
			desc:          "EMA should route to single IIFE, not tuple",
		},
		{
			name:          "ta.stoch_routes_to_tuple",
			funcName:      "ta.stoch",
			shouldBeTuple: true,
			desc:          "Stochastic should route to tuple IIFE",
		},
		{
			name:          "ta.rsi_routes_to_single",
			funcName:      "ta.rsi",
			shouldBeTuple: false,
			desc:          "RSI should route to single output",
		},
		{
			name:          "ta.dmi_routes_to_tuple",
			funcName:      "ta.dmi",
			shouldBeTuple: true,
			desc:          "DMI should route to tuple IIFE",
		},
		{
			name:          "ta.kc_routes_to_tuple",
			funcName:      "ta.kc",
			shouldBeTuple: true,
			desc:          "Keltner Channels should route to tuple IIFE",
		},
		{
			name:          "ta.supertrend_routes_to_tuple",
			funcName:      "ta.supertrend",
			shouldBeTuple: true,
			desc:          "Supertrend should route to tuple IIFE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewTupleIndicatorRegistry()
			isRegistered := registry.IsRegistered(tt.funcName)

			if tt.shouldBeTuple && !isRegistered {
				t.Errorf("%s: %s should be registered as tuple", tt.desc, tt.funcName)
			}

			if !tt.shouldBeTuple && isRegistered {
				t.Errorf("%s: %s incorrectly registered as tuple", tt.desc, tt.funcName)
			}
		})
	}
}

/*
TestArrowFunctionTACallGenerator_TupleRegistryCompleteness validates all
known multi-output PineScript functions are registered.
*/
func TestArrowFunctionTACallGenerator_TupleRegistryCompleteness(t *testing.T) {
	registry := NewTupleIndicatorRegistry()

	requiredFunctions := map[string]int{
		"ta.macd":       3,
		"macd":          3,
		"ta.bb":         3,
		"ta.stoch":      2,
		"ta.dmi":        3,
		"ta.kc":         3,
		"ta.supertrend": 2,
	}

	for funcName, expectedOutputs := range requiredFunctions {
		t.Run(funcName, func(t *testing.T) {
			if !registry.IsRegistered(funcName) {
				t.Errorf("Required function %s not registered", funcName)
				return
			}

			spec := registry.Lookup(funcName)
			if spec == nil {
				t.Fatalf("Spec for %s is nil after IsRegistered returned true", funcName)
			}

			if spec.OutputCount != expectedOutputs {
				t.Errorf("%s: expected %d outputs, got %d", funcName, expectedOutputs, spec.OutputCount)
			}

			if spec.RuntimeFunction == "" {
				t.Errorf("%s: RuntimeFunction is empty", funcName)
			}

			if !strings.HasPrefix(spec.RuntimeFunction, "ta.") {
				t.Errorf("%s: RuntimeFunction should start with 'ta.', got %s", funcName, spec.RuntimeFunction)
			}
		})
	}
}

/*
TestArrowFunctionTACallGenerator_TupleArgumentClassification validates argument
classification logic for tuple indicators: distinguishing series (source) arguments
from period (literal) arguments based on spec.SourceArgIndex.
*/
func TestArrowFunctionTACallGenerator_TupleArgumentClassification(t *testing.T) {
	tests := []struct {
		name             string
		spec             *TupleIndicatorSpec
		argIndex         int
		expectedIsSource bool
		desc             string
	}{
		{
			name: "macd_first_arg_is_source",
			spec: &TupleIndicatorSpec{
				FunctionName:   "ta.macd",
				SourceArgIndex: 0,
			},
			argIndex:         0,
			expectedIsSource: true,
			desc:             "First argument is source series",
		},
		{
			name: "macd_second_arg_is_period",
			spec: &TupleIndicatorSpec{
				FunctionName:   "ta.macd",
				SourceArgIndex: 0,
			},
			argIndex:         1,
			expectedIsSource: false,
			desc:             "Second argument is period, not source",
		},
		{
			name: "bb_first_arg_is_source",
			spec: &TupleIndicatorSpec{
				FunctionName:   "ta.bb",
				SourceArgIndex: 0,
			},
			argIndex:         0,
			expectedIsSource: true,
			desc:             "Bollinger Bands source at index 0",
		},
		{
			name: "dmi_no_source_arg",
			spec: &TupleIndicatorSpec{
				FunctionName:   "ta.dmi",
				SourceArgIndex: -1,
			},
			argIndex:         0,
			expectedIsSource: false,
			desc:             "DMI has no source, arg 0 is period",
		},
		{
			name: "stoch_no_explicit_source",
			spec: &TupleIndicatorSpec{
				FunctionName:   "ta.stoch",
				SourceArgIndex: -1,
			},
			argIndex:         0,
			expectedIsSource: false,
			desc:             "Stochastic uses implicit OHLC, not explicit source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGeneratorForTupleTests()
			accessorFactory := NewArrowAwareAccessorFactory(
				NewArrowIdentifierResolver(NewArrowSeriesAccessResolver()),
				&legacyArrowExpressionGenerator{gen: g},
				g,
				g.symbolTable,
			)

			builder := NewTupleArgumentBuilder(g, accessorFactory)
			isSource := builder.isSourceArgument(tt.argIndex, tt.spec)

			if isSource != tt.expectedIsSource {
				t.Errorf("%s: expected isSource=%v, got %v", tt.desc, tt.expectedIsSource, isSource)
			}
		})
	}
}
