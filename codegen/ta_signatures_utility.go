package codegen

func RegisterUtilitySignatures() []TAFunctionMetadata {
	signatures := make([]TAFunctionMetadata, 0)

	twoSeriesOverloads := []TAOverloadRule{
		{ArgCount: 2, Arguments: []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewSeriesArgument(1, ""),
		}},
	}
	signatures = appendWithBareAlias(signatures, "ta.crossover", "", twoSeriesOverloads)
	signatures = appendWithBareAlias(signatures, "ta.crossunder", "", twoSeriesOverloads)

	singleSeriesOverloads := []TAOverloadRule{
		{ArgCount: 1, Arguments: []TAArgumentSpec{
			NewSeriesArgument(0, ""),
		}},
	}
	signatures = appendWithBareAlias(signatures, "ta.fixnan", "", singleSeriesOverloads)
	signatures = appendWithBareAlias(signatures, "ta.barssince", "", singleSeriesOverloads)
	signatures = appendWithBareAlias(signatures, "ta.cum", "", singleSeriesOverloads)

	valuewhenOverloads := []TAOverloadRule{
		{ArgCount: 3, Arguments: []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewSeriesArgument(1, ""),
			NewScalarIntArgument(2),
		}},
	}
	signatures = appendWithBareAlias(signatures, "ta.valuewhen", "", valuewhenOverloads)

	seriesAndLengthOverloads := []TAOverloadRule{
		{ArgCount: 2, Arguments: []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
		}},
	}
	signatures = appendWithBareAlias(signatures, "ta.falling", "", seriesAndLengthOverloads)
	signatures = appendWithBareAlias(signatures, "ta.rising", "", seriesAndLengthOverloads)

	return signatures
}
