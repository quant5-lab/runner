package codegen

type TAFunctionMetadata struct {
	FunctionName  string
	Overloads     []TAOverloadRule
	DefaultSource string
}

func NewTAFunctionMetadata(name string, defaultSource string, overloads []TAOverloadRule) TAFunctionMetadata {
	return TAFunctionMetadata{
		FunctionName:  name,
		DefaultSource: defaultSource,
		Overloads:     overloads,
	}
}

func (m TAFunctionMetadata) FindOverload(argCount int) (TAOverloadRule, bool) {
	for _, overload := range m.Overloads {
		if overload.Matches(argCount) {
			return overload, true
		}
	}
	return TAOverloadRule{}, false
}

func (m TAFunctionMetadata) SupportsArgCount(argCount int) bool {
	_, found := m.FindOverload(argCount)
	return found
}

func (m TAFunctionMetadata) MinArgCount() int {
	if len(m.Overloads) == 0 {
		return 0
	}
	min := m.Overloads[0].ArgCount
	for _, overload := range m.Overloads {
		if overload.ArgCount < min {
			min = overload.ArgCount
		}
	}
	return min
}

func (m TAFunctionMetadata) MaxArgCount() int {
	if len(m.Overloads) == 0 {
		return 0
	}
	max := m.Overloads[0].ArgCount
	for _, overload := range m.Overloads {
		if overload.ArgCount > max {
			max = overload.ArgCount
		}
	}
	return max
}
