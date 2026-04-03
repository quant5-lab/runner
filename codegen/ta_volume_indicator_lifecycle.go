package codegen

import "fmt"

// VolumeIndicatorLifecycle manages the full ForwardSeriesBuffer lifecycle for the set
// of ta.* volume built-in variables that appear in a given script.
//
// It implements SeriesLifecycle and plugs into CompositeSeriesLifecycle exactly like
// TimeSeriesLifecycle or SessionSeriesLifecycle — zero generator coupling.
type VolumeIndicatorLifecycle struct {
	active []*VolumeIndicatorSpec
}

// NewVolumeIndicatorLifecycle filters the global spec table against the detected
// member keys map (format "ta.X" → bool) produced by BuiltinUsageDetector.
func NewVolumeIndicatorLifecycle(detected map[string]bool) *VolumeIndicatorLifecycle {
	var active []*VolumeIndicatorSpec
	for _, spec := range allVolumeIndicatorSpecs {
		if detected[spec.MemberKey] {
			active = append(active, spec)
		}
	}
	return &VolumeIndicatorLifecycle{active: active}
}

func (l *VolumeIndicatorLifecycle) HasUsage() bool {
	return len(l.active) > 0
}

func (l *VolumeIndicatorLifecycle) NeedsTimezone() bool {
	return false
}

func (l *VolumeIndicatorLifecycle) GenerateDeclarations(indent string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	for _, spec := range l.active {
		code += fmt.Sprintf("%svar %s *series.Series\n", indent, spec.SeriesName)
	}
	return code
}

func (l *VolumeIndicatorLifecycle) GenerateInitializations(indent string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	for _, spec := range l.active {
		code += fmt.Sprintf("%s%s = series.NewSeries(len(ctx.Data))\n", indent, spec.SeriesName)
	}
	return code
}

func (l *VolumeIndicatorLifecycle) GenerateBarPopulation(indent, iterVar string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	for _, spec := range l.active {
		code += spec.PopulateBarCode(indent, iterVar)
	}
	return code
}

func (l *VolumeIndicatorLifecycle) GenerateAdvancement(indent, iterVar string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	for _, spec := range l.active {
		code += fmt.Sprintf("%sif %s < barCount-1 { %s.Next() }\n", indent, iterVar, spec.SeriesName)
	}
	return code
}

func (l *VolumeIndicatorLifecycle) GenerateRegistrations(indent string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	for _, spec := range l.active {
		code += fmt.Sprintf("%sctx.RegisterSeries(%q, %s)\n", indent, spec.SeriesName, spec.SeriesName)
	}
	return code
}

func (l *VolumeIndicatorLifecycle) GenerateSuppressUnused(indent string) string {
	if !l.HasUsage() {
		return ""
	}
	code := ""
	for _, spec := range l.active {
		code += fmt.Sprintf("%s_ = %s\n", indent, spec.SeriesName)
	}
	return code
}

// GenerateSymbolTableRegistrations is intentionally a no-op for volume indicators.
// They are dispatched via evaluateMemberExpressionAtBar (MemberExpression path), not
// via the identifier var-lookup switch that the symbol table drives. Registering
// "ta.obv" would cause SecurityExpressionHandler to emit `varSeries = ta.obvSeries`
// which is invalid Go (package reference, not a variable).
func (l *VolumeIndicatorLifecycle) GenerateSymbolTableRegistrations(_ SymbolTable) {}
