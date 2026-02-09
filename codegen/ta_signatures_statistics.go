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
	signatures = appendWithBareAlias(signatures, "ta.change", "", changeOverloads)

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

	sumOverloads := []TAOverloadRule{
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
		}),
	}
	signatures = appendWithBareAlias(signatures, "ta.sum", "close", sumOverloads)
	signatures = append(signatures, NewTAFunctionMetadata("math.sum", "close", sumOverloads))

	linregOverloads := []TAOverloadRule{
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
		}),
		NewSingleOverloadRule(3, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
			NewScalarIntArgument(2),
		}),
	}
	signatures = appendWithBareAlias(signatures, "ta.linreg", "", linregOverloads)

	return signatures
}
