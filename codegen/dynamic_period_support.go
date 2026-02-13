package codegen

import "strings"

func SupportsDynamicPeriod(funcName string) bool {
	if _, ok := dynamicPeriodDispatch[funcName]; ok {
		return true
	}
	if !strings.Contains(funcName, ".") {
		_, ok := dynamicPeriodDispatch["ta."+funcName]
		return ok
	}
	return false
}
