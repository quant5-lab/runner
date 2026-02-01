package codegen

func RegisterPivotSignatures() []TAFunctionMetadata {
	signatures := make([]TAFunctionMetadata, 0)

	signatures = append(signatures, NewTAFunctionMetadata(
		"ta.pivothigh",
		"high",
		[]TAOverloadRule{
			NewSingleOverloadRule(2, []TAArgumentSpec{
				NewScalarIntArgument(0),
				NewScalarIntArgument(1),
			}),
			NewSingleOverloadRule(3, []TAArgumentSpec{
				NewSeriesArgument(0, ""),
				NewScalarIntArgument(1),
				NewScalarIntArgument(2),
			}),
		},
	))

	signatures = append(signatures, NewTAFunctionMetadata(
		"ta.pivotlow",
		"low",
		[]TAOverloadRule{
			NewSingleOverloadRule(2, []TAArgumentSpec{
				NewScalarIntArgument(0),
				NewScalarIntArgument(1),
			}),
			NewSingleOverloadRule(3, []TAArgumentSpec{
				NewSeriesArgument(0, ""),
				NewScalarIntArgument(1),
				NewScalarIntArgument(2),
			}),
		},
	))

	return signatures
}
