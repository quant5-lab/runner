package codegen

func RegisterMovingAverageSignatures() []TAFunctionMetadata {
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

	signatures = appendWithBareAlias(signatures, "ta.sma", "close", sourceAndLengthOverloads)
	signatures = appendWithBareAlias(signatures, "ta.ema", "close", sourceAndLengthOverloads)
	signatures = appendWithBareAlias(signatures, "ta.rma", "close", sourceAndLengthOverloads)
	signatures = appendWithBareAlias(signatures, "ta.wma", "close", sourceAndLengthOverloads)
	signatures = appendWithBareAlias(signatures, "ta.vwma", "close", sourceAndLengthOverloads)
	signatures = appendWithBareAlias(signatures, "ta.hma", "close", sourceAndLengthOverloads)

	swmaOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
		}),
	}
	swma := NewTAFunctionMetadata("ta.swma", "", swmaOverloads)
	swma.SourceOnlyLookback = true
	swmaBare := NewTAFunctionMetadata("swma", "", swmaOverloads)
	swmaBare.SourceOnlyLookback = true
	signatures = append(signatures, swma, swmaBare)

	almaOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewScalarIntArgument(0),
		}),
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
		}),
		NewSingleOverloadRule(3, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
			NewScalarFloatArgument(2),
		}),
		NewSingleOverloadRule(4, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
			NewScalarFloatArgument(2),
			NewScalarFloatArgument(3),
		}),
	}
	signatures = appendWithBareAlias(signatures, "ta.alma", "close", almaOverloads)

	return signatures
}
