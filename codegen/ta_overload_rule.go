package codegen

type TAOverloadRule struct {
	ArgCount  int
	Arguments []TAArgumentSpec
}

func NewSingleOverloadRule(argCount int, args []TAArgumentSpec) TAOverloadRule {
	return TAOverloadRule{
		ArgCount:  argCount,
		Arguments: args,
	}
}

func (r TAOverloadRule) Matches(providedArgCount int) bool {
	return r.ArgCount == providedArgCount
}

func (r TAOverloadRule) GetArgumentSpec(position int) (TAArgumentSpec, bool) {
	for _, arg := range r.Arguments {
		if arg.Position == position {
			return arg, true
		}
	}
	return TAArgumentSpec{}, false
}

func (r TAOverloadRule) HasImplicitOHLC() bool {
	for _, arg := range r.Arguments {
		if arg.Classification == TAArgImplicitOHLC {
			return true
		}
	}
	return false
}

func (r TAOverloadRule) CountSeriesArguments() int {
	count := 0
	for _, arg := range r.Arguments {
		if arg.Classification.IsSeries() {
			count++
		}
	}
	return count
}

func (r TAOverloadRule) CountScalarArguments() int {
	count := 0
	for _, arg := range r.Arguments {
		if arg.Classification.IsScalar() {
			count++
		}
	}
	return count
}
