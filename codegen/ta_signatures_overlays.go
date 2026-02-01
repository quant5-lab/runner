package codegen

func RegisterOverlaySignatures() []TAFunctionMetadata {
	signatures := make([]TAFunctionMetadata, 0)

	signatures = append(signatures, NewTAFunctionMetadata(
		"ta.bb",
		"close",
		[]TAOverloadRule{
			NewSingleOverloadRule(2, []TAArgumentSpec{
				NewScalarIntArgument(0),
				NewScalarFloatArgument(1),
			}),
			NewSingleOverloadRule(3, []TAArgumentSpec{
				NewSeriesArgument(0, ""),
				NewScalarIntArgument(1),
				NewScalarFloatArgument(2),
			}),
		},
	))

	signatures = append(signatures, NewTAFunctionMetadata(
		"ta.supertrend",
		"",
		[]TAOverloadRule{
			NewSingleOverloadRule(2, []TAArgumentSpec{
				NewImplicitOHLCArgument(),
				NewScalarFloatArgument(0),
				NewScalarIntArgument(1),
			}),
		},
	))

	signatures = append(signatures, NewTAFunctionMetadata(
		"ta.kc",
		"close",
		[]TAOverloadRule{
			NewSingleOverloadRule(2, []TAArgumentSpec{
				NewScalarIntArgument(0),
				NewScalarFloatArgument(1),
			}),
			NewSingleOverloadRule(3, []TAArgumentSpec{
				NewSeriesArgument(0, ""),
				NewScalarIntArgument(1),
				NewScalarFloatArgument(2),
			}),
		},
	))

	return signatures
}
