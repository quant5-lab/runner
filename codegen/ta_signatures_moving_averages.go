package codegen

func RegisterMovingAverageSignatures() []TAFunctionMetadata {
	signatures := make([]TAFunctionMetadata, 0)

	smaOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewScalarIntArgument(0),
		}),
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
		}),
	}
	signatures = append(signatures, NewTAFunctionMetadata("ta.sma", "close", smaOverloads))
	signatures = append(signatures, NewTAFunctionMetadata("sma", "close", smaOverloads))

	emaOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewScalarIntArgument(0),
		}),
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
		}),
	}
	signatures = append(signatures, NewTAFunctionMetadata("ta.ema", "close", emaOverloads))
	signatures = append(signatures, NewTAFunctionMetadata("ema", "close", emaOverloads))

	rmaOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewScalarIntArgument(0),
		}),
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
		}),
	}
	signatures = append(signatures, NewTAFunctionMetadata("ta.rma", "close", rmaOverloads))
	signatures = append(signatures, NewTAFunctionMetadata("rma", "close", rmaOverloads))

	wmaOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewScalarIntArgument(0),
		}),
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
		}),
	}
	signatures = append(signatures, NewTAFunctionMetadata("ta.wma", "close", wmaOverloads))
	signatures = append(signatures, NewTAFunctionMetadata("wma", "close", wmaOverloads))

	vwmaOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewScalarIntArgument(0),
		}),
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
		}),
	}
	signatures = append(signatures, NewTAFunctionMetadata("ta.vwma", "close", vwmaOverloads))
	signatures = append(signatures, NewTAFunctionMetadata("vwma", "close", vwmaOverloads))

	swmaOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
		}),
	}
	signatures = append(signatures, NewTAFunctionMetadata("ta.swma", "", swmaOverloads))
	signatures = append(signatures, NewTAFunctionMetadata("swma", "", swmaOverloads))

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
	signatures = append(signatures, NewTAFunctionMetadata("ta.alma", "close", almaOverloads))
	signatures = append(signatures, NewTAFunctionMetadata("alma", "close", almaOverloads))

	return signatures
}
