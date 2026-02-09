package codegen

func RegisterPivotSignatures() []TAFunctionMetadata {
	signatures := make([]TAFunctionMetadata, 0)

	dualPeriodOverloads := []TAOverloadRule{
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewScalarIntArgument(0),
			NewScalarIntArgument(1),
		}),
		NewSingleOverloadRule(3, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
			NewScalarIntArgument(2),
		}),
	}

	signatures = appendWithBareAlias(signatures, "ta.pivothigh", "high", dualPeriodOverloads)
	signatures = appendWithBareAlias(signatures, "ta.pivotlow", "low", dualPeriodOverloads)

	return signatures
}
