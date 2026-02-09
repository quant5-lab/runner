package codegen

func RegisterTupleSignatures() []TAFunctionMetadata {
	signatures := make([]TAFunctionMetadata, 0)

	/* macd(source, fastlen, slowlen, siglen) → [macdLine, signal, hist] */
	macdOverloads := []TAOverloadRule{
		{ArgCount: 4, Arguments: []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
			NewScalarIntArgument(2),
			NewScalarIntArgument(3),
		}},
	}
	signatures = appendTupleWithBareAlias(signatures, "ta.macd", "close", macdOverloads)

	/* stoch(source, high, low, length) */
	stochOverloads := []TAOverloadRule{
		{ArgCount: 4, Arguments: []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewSeriesArgument(1, ""),
			NewSeriesArgument(2, ""),
			NewScalarIntArgument(3),
		}},
	}
	signatures = appendTupleWithBareAlias(signatures, "ta.stoch", "", stochOverloads)

	/* dmi(diLength, adxSmoothing) → [plus, minus, adx] — implicit OHLC */
	dmiOverloads := []TAOverloadRule{
		{ArgCount: 2, Arguments: []TAArgumentSpec{
			NewImplicitOHLCArgument(),
			NewScalarIntArgument(0),
			NewScalarIntArgument(1),
		}},
	}
	signatures = appendTupleWithBareAlias(signatures, "ta.dmi", "", dmiOverloads)

	return signatures
}
