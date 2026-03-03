package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/codegen/series_naming"
)

// ALMAIIFEGenerator generates inline IIFE code for ta.alma in arrow function context.
// Uses default offset=0.85 and sigma=6 for the default registration.
// Custom offset/sigma values are handled by the arrow_function_ta_call_generator special-casing.
type ALMAIIFEGenerator struct {
	namingStrategy series_naming.Strategy
	offset         float64
	sigma          float64
	useDefaults    bool
}

func (g *ALMAIIFEGenerator) effectiveOffset() float64 {
	if g.useDefaults || g.offset == 0 {
		return 0.85
	}
	return g.offset
}

func (g *ALMAIIFEGenerator) effectiveSigma() float64 {
	if g.useDefaults || g.sigma == 0 {
		return 6.0
	}
	return g.sigma
}

func (g *ALMAIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	if !period.IsConstant() {
		return "math.NaN()"
	}
	periodInt := period.AsInt()
	offset := g.effectiveOffset()
	sigma := g.effectiveSigma()

	m := offset * float64(periodInt-1)
	s := float64(periodInt) / sigma

	body := fmt.Sprintf("w := [%d]float64{}; wsum := 0.0; ", periodInt)
	body += fmt.Sprintf("for j := 0; j < %d; j++ { d := float64(j) - %g; w[j] = math.Exp(-(d*d)/(2*%g*%g)); wsum += w[j] }; ",
		periodInt, m, s, s)
	body += fmt.Sprintf("val := 0.0; for j := 0; j < %d; j++ { val += w[j] * %s }; ",
		periodInt, accessor.GenerateLoopValueAccess(fmt.Sprintf("%d-1-j", periodInt)))
	body += "return val / wsum"

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}
