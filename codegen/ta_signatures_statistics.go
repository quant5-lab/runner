package codegen

func RegisterStatisticsSignatures() []TAFunctionMetadata {
	signatures := make([]TAFunctionMetadata, 0)

	changeOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
		}),
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
		}),
	}
	signatures = append(signatures, NewTAFunctionMetadata("ta.change", "", changeOverloads))
	signatures = append(signatures, NewTAFunctionMetadata("change", "", changeOverloads))

	signatures = append(signatures, NewTAFunctionMetadata(
		"ta.correlation",
		"",
		[]TAOverloadRule{
			NewSingleOverloadRule(3, []TAArgumentSpec{
				NewSeriesArgument(0, ""),
				NewSeriesArgument(1, ""),
				NewScalarIntArgument(2),
			}),
		},
	))

	signatures = append(signatures, NewTAFunctionMetadata(
		"ta.variance",
		"close",
		[]TAOverloadRule{
			NewSingleOverloadRule(1, []TAArgumentSpec{
				NewScalarIntArgument(0),
			}),
			NewSingleOverloadRule(2, []TAArgumentSpec{
				NewSeriesArgument(0, ""),
				NewScalarIntArgument(1),
			}),
		},
	))

	signatures = append(signatures, NewTAFunctionMetadata(
		"ta.median",
		"close",
		[]TAOverloadRule{
			NewSingleOverloadRule(1, []TAArgumentSpec{
				NewScalarIntArgument(0),
			}),
			NewSingleOverloadRule(2, []TAArgumentSpec{
				NewSeriesArgument(0, ""),
				NewScalarIntArgument(1),
			}),
		},
	))

	return signatures
}
