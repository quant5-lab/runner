package codegen

import (
	"fmt"
	"sort"
)

type CalendarSeriesLifecycle struct {
	usedBuiltins []CalendarBuiltinInfo
}

func NewCalendarSeriesLifecycle(usedBuiltins []CalendarBuiltinInfo) *CalendarSeriesLifecycle {
	sorted := make([]CalendarBuiltinInfo, len(usedBuiltins))
	copy(sorted, usedBuiltins)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].PineName < sorted[j].PineName
	})
	return &CalendarSeriesLifecycle{usedBuiltins: sorted}
}

func (l *CalendarSeriesLifecycle) HasUsage() bool {
	return l != nil && len(l.usedBuiltins) > 0
}

func (l *CalendarSeriesLifecycle) NeedsTimezone() bool {
	return l.HasUsage()
}

func (l *CalendarSeriesLifecycle) GenerateDeclarations(indent string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	for _, info := range l.usedBuiltins {
		code += indent + fmt.Sprintf("var %s *series.Series\n", info.SeriesName)
	}
	return code
}

func (l *CalendarSeriesLifecycle) GenerateInitializations(indent string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	for _, info := range l.usedBuiltins {
		code += indent + fmt.Sprintf("%s = series.NewSeries(len(ctx.Data))\n", info.SeriesName)
	}
	return code
}

func (l *CalendarSeriesLifecycle) GenerateBarPopulation(indent, _ string) string {
	if !l.HasUsage() {
		return ""
	}
	code := indent + "barCal := context.DecomposeBarTime(bar.Time, exchangeLoc)\n"
	for _, info := range l.usedBuiltins {
		code += indent + fmt.Sprintf("%s.Set(barCal.%s)\n", info.SeriesName, info.StructField)
	}
	return code
}

func (l *CalendarSeriesLifecycle) GenerateAdvancement(indent, iterVar string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	for _, info := range l.usedBuiltins {
		code += indent + fmt.Sprintf("if %s < barCount-1 { %s.Next() }\n", iterVar, info.SeriesName)
	}
	return code
}

func (l *CalendarSeriesLifecycle) GenerateRegistrations(indent string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	for _, info := range l.usedBuiltins {
		code += indent + fmt.Sprintf("ctx.RegisterSeries(%q, %s)\n", info.SeriesName, info.SeriesName)
	}
	return code
}

func (l *CalendarSeriesLifecycle) GenerateSuppressUnused(indent string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	for _, info := range l.usedBuiltins {
		code += indent + fmt.Sprintf("_ = %s\n", info.SeriesName)
	}
	return code
}

func (l *CalendarSeriesLifecycle) GenerateSymbolTableRegistrations(symbolTable SymbolTable) {
	if l == nil {
		return
	}
	for _, info := range l.usedBuiltins {
		if symbolTable != nil {
			symbolTable.Register(info.PineName, VariableTypeSeries)
		}
	}
}
