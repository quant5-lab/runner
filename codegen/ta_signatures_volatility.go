package codegen

func RegisterVolatilitySignatures() []TAFunctionMetadata {
	signatures := make([]TAFunctionMetadata, 0)

	signatures = append(signatures, NewTAFunctionMetadata(
		"ta.atr",
		"",
		[]TAOverloadRule{
			NewSingleOverloadRule(1, []TAArgumentSpec{
				NewImplicitOHLCArgument(),
				NewScalarIntArgument(0),
			}),
		},
	))

	signatures = append(signatures, NewTAFunctionMetadata(
		"ta.tr",
		"",
		[]TAOverloadRule{
			NewSingleOverloadRule(0, []TAArgumentSpec{
				NewImplicitOHLCArgument(),
			}),
		},
	))

	signatures = append(signatures, NewTAFunctionMetadata(
		"ta.stdev",
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
