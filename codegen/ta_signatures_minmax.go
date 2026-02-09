package codegen

func RegisterMinMaxSignatures() []TAFunctionMetadata {
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

	signatures = appendWithBareAlias(signatures, "ta.highest", "high", sourceAndLengthOverloads)
	signatures = appendWithBareAlias(signatures, "ta.lowest", "low", sourceAndLengthOverloads)
	signatures = appendWithBareAlias(signatures, "ta.highestbars", "high", sourceAndLengthOverloads)
	signatures = appendWithBareAlias(signatures, "ta.lowestbars", "low", sourceAndLengthOverloads)

	return signatures
}
