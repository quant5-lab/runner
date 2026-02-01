package codegen

func RegisterMinMaxSignatures() []TAFunctionMetadata {
	signatures := make([]TAFunctionMetadata, 0)

	highestOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewScalarIntArgument(0),
		}),
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
		}),
	}
	signatures = append(signatures, NewTAFunctionMetadata("ta.highest", "high", highestOverloads))
	signatures = append(signatures, NewTAFunctionMetadata("highest", "high", highestOverloads))

	lowestOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewScalarIntArgument(0),
		}),
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
		}),
	}
	signatures = append(signatures, NewTAFunctionMetadata("ta.lowest", "low", lowestOverloads))
	signatures = append(signatures, NewTAFunctionMetadata("lowest", "low", lowestOverloads))

	highestbarsOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewScalarIntArgument(0),
		}),
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
		}),
	}
	signatures = append(signatures, NewTAFunctionMetadata("ta.highestbars", "high", highestbarsOverloads))
	signatures = append(signatures, NewTAFunctionMetadata("highestbars", "high", highestbarsOverloads))

	lowestbarsOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewScalarIntArgument(0),
		}),
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
		}),
	}
	signatures = append(signatures, NewTAFunctionMetadata("ta.lowestbars", "low", lowestbarsOverloads))
	signatures = append(signatures, NewTAFunctionMetadata("lowestbars", "low", lowestbarsOverloads))

	return signatures
}
