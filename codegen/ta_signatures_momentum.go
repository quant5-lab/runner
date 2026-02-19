package codegen

func RegisterMomentumSignatures() []TAFunctionMetadata {
	signatures := make([]TAFunctionMetadata, 0)

	sourceAndLengthOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewScalarIntArgument(0),
		}),
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
		}),
	}

	signatures = appendWithBareAlias(signatures, "ta.mom", "close", sourceAndLengthOverloads)
	signatures = appendWithBareAlias(signatures, "ta.roc", "close", sourceAndLengthOverloads)
	signatures = appendWithBareAlias(signatures, "ta.cmo", "close", sourceAndLengthOverloads)
	signatures = appendWithBareAlias(signatures, "ta.rising", "close", sourceAndLengthOverloads)
	signatures = appendWithBareAlias(signatures, "ta.falling", "close", sourceAndLengthOverloads)
	signatures = appendWithBareAlias(signatures, "ta.highestbars", "high", sourceAndLengthOverloads)
	signatures = appendWithBareAlias(signatures, "ta.lowestbars", "low", sourceAndLengthOverloads)

	wprOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewScalarIntArgument(0),
		}),
	}
	signatures = appendWithBareAlias(signatures, "ta.wpr", "", wprOverloads)

	crossOverloads := []TAOverloadRule{
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewSeriesArgument(1, ""),
		}),
	}
	signatures = appendWithBareAlias(signatures, "ta.cross", "", crossOverloads)

	return signatures
}
