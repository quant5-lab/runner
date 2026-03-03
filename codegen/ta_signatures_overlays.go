package codegen

func RegisterOverlaySignatures() []TAFunctionMetadata {
	signatures := make([]TAFunctionMetadata, 0)

	bbOverloads := []TAOverloadRule{
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewScalarIntArgument(0),
			NewScalarFloatArgument(1),
		}),
		NewSingleOverloadRule(3, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
			NewScalarFloatArgument(2),
		}),
	}
	signatures = appendTupleWithBareAlias(signatures, "ta.bb", "close", bbOverloads)

	supertrendOverloads := []TAOverloadRule{
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewImplicitOHLCArgument(),
			NewScalarFloatArgument(0),
			NewScalarIntArgument(1),
		}),
	}
	signatures = appendTupleWithBareAlias(signatures, "ta.supertrend", "", supertrendOverloads)

	kcOverloads := []TAOverloadRule{
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewScalarIntArgument(0),
			NewScalarFloatArgument(1),
		}),
		NewSingleOverloadRule(3, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
			NewScalarFloatArgument(2),
		}),
	}
	signatures = appendTupleWithBareAlias(signatures, "ta.kc", "close", kcOverloads)

	bbwOverloads := []TAOverloadRule{
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewScalarIntArgument(0),
			NewScalarFloatArgument(1),
		}),
		NewSingleOverloadRule(3, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
			NewScalarFloatArgument(2),
		}),
	}
	signatures = appendWithBareAlias(signatures, "ta.bbw", "close", bbwOverloads)

	cogOverloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{
			NewScalarIntArgument(0),
		}),
		NewSingleOverloadRule(2, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
		}),
	}
	signatures = appendWithBareAlias(signatures, "ta.cog", "close", cogOverloads)

	return signatures
}
