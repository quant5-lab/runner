package codegen

func RegisterVolatilitySignatures() []TAFunctionMetadata {
	signatures := make([]TAFunctionMetadata, 0)

	atrOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewImplicitOHLCArgument(),
			NewScalarIntArgument(0),
		}),
	}
	signatures = appendWithBareAlias(signatures, "ta.atr", "", atrOverloads)

	trOverloads := []TAOverloadRule{
		NewSingleOverloadRule(0, []TAArgumentSpec{
			NewImplicitOHLCArgument(),
		}),
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewImplicitOHLCArgument(),
			NewScalarBoolArgument(0),
		}),
	}
	signatures = appendWithBareAlias(signatures, "ta.tr", "", trOverloads)

	stdevOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewScalarIntArgument(0),
		}),
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
		}),
	}
	signatures = appendWithBareAlias(signatures, "ta.stdev", "close", stdevOverloads)

	devOverloads := []TAOverloadRule{
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
		}),
	}
	signatures = appendWithBareAlias(signatures, "ta.dev", "", devOverloads)

	return signatures
}
