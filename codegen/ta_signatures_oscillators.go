package codegen

func RegisterOscillatorSignatures() []TAFunctionMetadata {
	signatures := make([]TAFunctionMetadata, 0)

	signatures = append(signatures, NewTAFunctionMetadata(
		"ta.rsi",
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
		"ta.cci",
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
		"ta.mfi",
		"",
		[]TAOverloadRule{
			NewSingleOverloadRule(1, []TAArgumentSpec{
				NewImplicitOHLCArgument(),
				NewScalarIntArgument(0),
			}),
		},
	))

	return signatures
}
