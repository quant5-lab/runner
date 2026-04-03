package codegen

import (
	"testing"
)

/* Argument classification: series (source) vs period (literal) by spec.SourceArgIndex */
func TestTupleArgumentBuilder_SourceArgClassification(t *testing.T) {
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
