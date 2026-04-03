package codegen

import "strings"

type TAFunctionMetadata struct {
	FunctionName       string
	Overloads          []TAOverloadRule
	DefaultSource      string
	SourceOnlyLookback bool
	IsTuple            bool
}

func NewTAFunctionMetadata(name string, defaultSource string, overloads []TAOverloadRule) TAFunctionMetadata {
	return TAFunctionMetadata{
		FunctionName:  name,
		DefaultSource: defaultSource,
		Overloads:     overloads,
	}
}

/* Registers both namespaced ("ta.sma") and bare ("sma") forms from a single declaration */
func appendWithBareAlias(signatures []TAFunctionMetadata, namespacedName, defaultSource string, overloads []TAOverloadRule) []TAFunctionMetadata {
	signatures = append(signatures, NewTAFunctionMetadata(namespacedName, defaultSource, overloads))
	if i := strings.LastIndex(namespacedName, "."); i >= 0 {
		signatures = append(signatures, NewTAFunctionMetadata(namespacedName[i+1:], defaultSource, overloads))
	}
	return signatures
}

/* Registers tuple function in both namespaced and bare forms with IsTuple flag set */
func appendTupleWithBareAlias(signatures []TAFunctionMetadata, namespacedName, defaultSource string, overloads []TAOverloadRule) []TAFunctionMetadata {
	meta := NewTAFunctionMetadata(namespacedName, defaultSource, overloads)
	meta.IsTuple = true
	signatures = append(signatures, meta)
	if i := strings.LastIndex(namespacedName, "."); i >= 0 {
		bare := NewTAFunctionMetadata(namespacedName[i+1:], defaultSource, overloads)
		bare.IsTuple = true
		signatures = append(signatures, bare)
	}
	return signatures
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

/*
HasOverloadWithSourceAt returns true when there is an overload for the given arg count
whose first non-OHLC argument is a series. Used by the resolver to distinguish
"caller provided explicit source" from "use default source".
*/
func (m TAFunctionMetadata) HasOverloadWithSourceAt(argCount int) bool {
	for _, overload := range m.Overloads {
		if !overload.Matches(argCount) {
			continue
		}
		for _, arg := range overload.Arguments {
			if arg.Classification == TAArgImplicitOHLC {
				continue
			}
			return arg.Classification.IsSeries()
		}
	}
	return false
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
