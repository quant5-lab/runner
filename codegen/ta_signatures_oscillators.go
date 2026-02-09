package codegen

func RegisterOscillatorSignatures() []TAFunctionMetadata {
	signatures := make([]TAFunctionMetadata, 0)

	rsiOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewScalarIntArgument(0),
		}),
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
		}),
	}
	signatures = appendWithBareAlias(signatures, "ta.rsi", "close", rsiOverloads)

	cciOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewScalarIntArgument(0),
		}),
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
		}),
	}
	signatures = appendWithBareAlias(signatures, "ta.cci", "close", cciOverloads)

	mfiOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewImplicitOHLCArgument(),
			NewScalarIntArgument(0),
		}),
	}
	signatures = appendWithBareAlias(signatures, "ta.mfi", "", mfiOverloads)

	tsiOverloads := []TAOverloadRule{
		NewSingleOverloadRule(3, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
			NewScalarIntArgument(2),
		}),
	}
	signatures = appendWithBareAlias(signatures, "ta.tsi", "", tsiOverloads)

	vwapOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
		}),
	}
	signatures = appendWithBareAlias(signatures, "ta.vwap", "", vwapOverloads)

	return signatures
}
