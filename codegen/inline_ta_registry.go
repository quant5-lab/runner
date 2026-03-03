package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/codegen/series_naming"
)

type InlineTAIIFEGenerator interface {
	Generate(accessor AccessGenerator, period PeriodExpression, sourceHash string) string
}

type InlineTADualPeriodGenerator interface {
	GenerateDualPeriod(accessor AccessGenerator, leftPeriod, rightPeriod PeriodExpression, sourceHash string) string
}

type InlineTAIIFERegistry struct {
	generators           map[string]InlineTAIIFEGenerator
	dualPeriodGenerators map[string]InlineTADualPeriodGenerator
}

func NewInlineTAIIFERegistry() *InlineTAIIFERegistry {
	r := &InlineTAIIFERegistry{
		generators:           make(map[string]InlineTAIIFEGenerator),
		dualPeriodGenerators: make(map[string]InlineTADualPeriodGenerator),
	}
	r.registerDefaults()
	return r
}

func (r *InlineTAIIFERegistry) registerDefaults() {
	windowNamer := series_naming.NewWindowBasedNamer()
	statefulNamer := series_naming.NewStatefulIndicatorNamer()

	r.RegisterWithBareAlias("ta.sma", &SMAIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.wma", &WMAIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.vwma", &VWMAIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.stdev", &STDEVIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.highest", &HighestIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.lowest", &LowestIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.change", &ChangeIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.linreg", &LinregIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.swma", &SWMAIIFEGenerator{namingStrategy: windowNamer})

	r.RegisterWithBareAlias("ta.ema", &EMAIIFEGenerator{namingStrategy: statefulNamer})
	r.RegisterWithBareAlias("ta.rma", &RMAIIFEGenerator{namingStrategy: statefulNamer})
	r.RegisterWithBareAlias("ta.rsi", &RSIIIFEGenerator{namingStrategy: statefulNamer})
	r.RegisterWithBareAlias("ta.atr", &ATRIIFEGenerator{namingStrategy: statefulNamer})

	sumGen := &SumIIFEGenerator{namingStrategy: windowNamer}
	r.RegisterWithBareAlias("ta.sum", sumGen)
	r.Register("math.sum", sumGen)

	r.RegisterDualPeriodWithBareAlias("ta.pivothigh", &PivotHighIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterDualPeriodWithBareAlias("ta.pivotlow", &PivotLowIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterDualPeriodWithBareAlias("ta.tsi", &TSIIIFEGenerator{namingStrategy: statefulNamer})

	r.RegisterWithBareAlias("ta.max", &MaxIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.min", &MinIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.range", &RangeIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.variance", &VarianceIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.dev", &DevIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.median", &MedianIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.mode", &ModeIIFEGenerator{namingStrategy: windowNamer})

	r.RegisterWithBareAlias("ta.rising", &RisingIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.falling", &FallingIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.highestbars", &HighestbarsIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.lowestbars", &LowestbarsIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.mom", &MomIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.roc", &RocIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.cmo", &CmoIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.wpr", &WprIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.cci", &CCIIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.bbw", &BBWIIFEGenerator{
		namingStrategy: windowNamer,
		multLiteral:    2.0,
		useLiteralMult: true,
	})
	r.RegisterWithBareAlias("ta.cog", &COGIIFEGenerator{namingStrategy: windowNamer})

	r.RegisterWithBareAlias("ta.percentrank", &PercentrankIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.percentile_nearest_rank", &PercentileNearestRankIIFEGenerator{
		namingStrategy: windowNamer,
		pct:            50.0,
	})
	r.RegisterWithBareAlias("ta.percentile_linear_interpolation", &PercentileLinearInterpolationIIFEGenerator{
		namingStrategy: windowNamer,
		pct:            50.0,
	})
	r.RegisterWithBareAlias("ta.alma", &ALMAIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterWithBareAlias("ta.hma", &HMAIIFEGenerator{})
	r.RegisterWithBareAlias("ta.kcw", &KCWIIFEGenerator{multExpr: "1.5"})
}

func (r *InlineTAIIFERegistry) Register(name string, generator InlineTAIIFEGenerator) {
	r.generators[name] = generator
}

func (r *InlineTAIIFERegistry) RegisterWithBareAlias(namespacedName string, generator InlineTAIIFEGenerator) {
	r.Register(namespacedName, generator)
	if i := strings.LastIndex(namespacedName, "."); i >= 0 {
		r.Register(namespacedName[i+1:], generator)
	}
}

func (r *InlineTAIIFERegistry) RegisterDualPeriod(name string, generator InlineTADualPeriodGenerator) {
	r.dualPeriodGenerators[name] = generator
}

func (r *InlineTAIIFERegistry) RegisterDualPeriodWithBareAlias(namespacedName string, generator InlineTADualPeriodGenerator) {
	r.RegisterDualPeriod(namespacedName, generator)
	if i := strings.LastIndex(namespacedName, "."); i >= 0 {
		r.RegisterDualPeriod(namespacedName[i+1:], generator)
	}
}

func (r *InlineTAIIFERegistry) IsRegisteredDualPeriod(funcName string) bool {
	_, ok := r.dualPeriodGenerators[funcName]
	return ok
}

func (r *InlineTAIIFERegistry) IsSupported(funcName string) bool {
	_, ok := r.generators[funcName]
	if ok {
		return true
	}
	_, ok = r.dualPeriodGenerators[funcName]
	return ok
}

func (r *InlineTAIIFERegistry) Generate(funcName string, accessor AccessGenerator, period PeriodExpression, sourceHash string) (string, bool) {
	gen, ok := r.generators[funcName]
	if !ok {
		return "", false
	}
	return gen.Generate(accessor, period, sourceHash), true
}

func (r *InlineTAIIFERegistry) GenerateDualPeriod(funcName string, accessor AccessGenerator, leftPeriod, rightPeriod PeriodExpression, sourceHash string) (string, bool) {
	gen, ok := r.dualPeriodGenerators[funcName]
	if !ok {
		return "", false
	}
	return gen.GenerateDualPeriod(accessor, leftPeriod, rightPeriod, sourceHash), true
}

func (r *InlineTAIIFERegistry) GenerateLinreg(accessor AccessGenerator, period PeriodExpression, offset int, sourceHash string) (string, bool) {
	gen, ok := r.generators["ta.linreg"]
	if !ok {
		return "", false
	}
	if linregGen, ok := gen.(*LinregIIFEGenerator); ok {
		return linregGen.GenerateWithOffset(accessor, period, offset, sourceHash), true
	}
	return "", false
}

type SMAIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type EMAIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type RMAIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type RSIIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type WMAIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type STDEVIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type HighestIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type LowestIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type ChangeIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type LinregIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type SumIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type SWMAIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type VWMAIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type ATRIIFEGenerator struct{ namingStrategy series_naming.Strategy }

func (g *SMAIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, sourceHash string) string {
	body := fmt.Sprintf("sum := 0.0; for j := 0; j < %s; j++ { sum += %s }; ", period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += fmt.Sprintf("return sum / %s", period.AsFloat64Cast())

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *SumIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, sourceHash string) string {
	body := fmt.Sprintf("sum := 0.0; for j := 0; j < %s; j++ { sum += %s }; ", period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += "return sum"

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *SWMAIIFEGenerator) Generate(accessor AccessGenerator, _ PeriodExpression, sourceHash string) string {
	/* Fixed period=4, weights [1/6, 2/6, 2/6, 1/6] */
	fixedPeriod := NewConstantPeriod(4)
	body := fmt.Sprintf("return %s*(1.0/6.0) + %s*(2.0/6.0) + %s*(2.0/6.0) + %s*(1.0/6.0)",
		accessor.GenerateLoopValueAccess("3"),
		accessor.GenerateLoopValueAccess("2"),
		accessor.GenerateLoopValueAccess("1"),
		accessor.GenerateLoopValueAccess("0"))

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(fixedPeriod, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *VWMAIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, sourceHash string) string {
	body := fmt.Sprintf("weightedSum := 0.0; volumeSum := 0.0; for j := 0; j < %s; j++ { val := %s; if !math.IsNaN(val) { weightedSum += val * ctx.Data[ctx.BarIndex-j].Volume; volumeSum += ctx.Data[ctx.BarIndex-j].Volume } }; ", period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += "return weightedSum / volumeSum"

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *ATRIIFEGenerator) Generate(_ AccessGenerator, period PeriodExpression, sourceHash string) string {
	/* RMA(TrueRange, period) — ignores passed accessor, uses OHLC directly */
	context := NewArrowFunctionIndicatorContext()
	varName := g.namingStrategy.GenerateName("atr", period.AsSeriesNamePart(), sourceHash)
	trAccessor := NewTrueRangeAccessGenerator()

	builder := NewStatefulIndicatorBuilder("ta.atr", varName, period, trAccessor, false, context)
	statefulCode := builder.BuildRMA()
	seriesAccess := fmt.Sprintf("arrowCtx.GetOrCreateSeries(%q).Get(0)", varName)

	return fmt.Sprintf("func() float64 {\n\t%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}

func (g *EMAIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, sourceHash string) string {
	context := NewArrowFunctionIndicatorContext()
	varName := g.namingStrategy.GenerateName("ema", period.AsSeriesNamePart(), sourceHash)

	builder := NewStatefulIndicatorBuilder("ta.ema", varName, period, accessor, false, context)
	statefulCode := builder.BuildEMA()
	seriesAccess := fmt.Sprintf("arrowCtx.GetOrCreateSeries(%q).Get(0)", varName)

	return fmt.Sprintf("func() float64 {\n\t%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}

func (g *RMAIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, sourceHash string) string {
	context := NewArrowFunctionIndicatorContext()
	varName := g.namingStrategy.GenerateName("rma", period.AsSeriesNamePart(), sourceHash)

	builder := NewStatefulIndicatorBuilder("ta.rma", varName, period, accessor, false, context)
	statefulCode := builder.BuildRMA()
	seriesAccess := fmt.Sprintf("arrowCtx.GetOrCreateSeries(%q).Get(0)", varName)

	return fmt.Sprintf("func() float64 {\n\t%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}

func (g *RSIIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, sourceHash string) string {
	context := NewArrowFunctionIndicatorContext()
	varName := g.namingStrategy.GenerateName("rsi", period.AsSeriesNamePart(), sourceHash)

	builder := NewRSIIndicatorBuilder(varName, period, accessor, false, context)
	statefulCode := builder.Build()
	seriesAccess := fmt.Sprintf("arrowCtx.GetOrCreateSeries(%q).Get(0)", varName)

	return fmt.Sprintf("func() float64 {\n\t%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}

func (g *WMAIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, sourceHash string) string {
	body := fmt.Sprintf("sum := 0.0; weightSum := 0.0; for j := 0; j < %s; j++ { weight := %s - float64(j); sum += weight * %s; weightSum += weight }; ", period.AsIntCast(), period.AsFloat64Cast(), accessor.GenerateLoopValueAccess("j"))
	body += "return sum / weightSum"

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *STDEVIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, sourceHash string) string {
	body := fmt.Sprintf("sum := 0.0; for j := 0; j < %s; j++ { sum += %s }; ", period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += fmt.Sprintf("mean := sum / %s; ", period.AsFloat64Cast())
	body += fmt.Sprintf("variance := 0.0; for j := 0; j < %s; j++ { diff := %s - mean; variance += diff * diff }; ", period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += fmt.Sprintf("return math.Sqrt(variance / %s)", period.AsFloat64Cast())

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *HighestIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, sourceHash string) string {
	if period.IsConstant() {
		periodInt := period.AsInt()
		body := fmt.Sprintf("highest := %s; ", accessor.GenerateInitialValueAccess(periodInt))
		body += fmt.Sprintf("for j := %d; j >= 0; j-- { v := %s; if v > highest { highest = v } }; ", periodInt-1, accessor.GenerateLoopValueAccess("j"))
		body += "return highest"

		return NewIIFECodeBuilder().
			WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
			WithBody(body).
			Build()
	}

	body := fmt.Sprintf("periodVal := %s; ", period.AsIntCast())
	body += fmt.Sprintf("highest := %s; ", accessor.GenerateLoopValueAccess("periodVal - 1"))
	body += fmt.Sprintf("for j := periodVal - 1; j >= 0; j-- { v := %s; if v > highest { highest = v } }; ", accessor.GenerateLoopValueAccess("j"))
	body += "return highest"

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *LowestIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, sourceHash string) string {
	if period.IsConstant() {
		periodInt := period.AsInt()
		body := fmt.Sprintf("lowest := %s; ", accessor.GenerateInitialValueAccess(periodInt))
		body += fmt.Sprintf("for j := %d; j >= 0; j-- { v := %s; if v < lowest { lowest = v } }; ", periodInt-1, accessor.GenerateLoopValueAccess("j"))
		body += "return lowest"

		return NewIIFECodeBuilder().
			WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
			WithBody(body).
			Build()
	}

	body := fmt.Sprintf("periodVal := %s; ", period.AsIntCast())
	body += fmt.Sprintf("lowest := %s; ", accessor.GenerateLoopValueAccess("periodVal - 1"))
	body += fmt.Sprintf("for j := periodVal - 1; j >= 0; j-- { v := %s; if v < lowest { lowest = v } }; ", accessor.GenerateLoopValueAccess("j"))
	body += "return lowest"

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *ChangeIIFEGenerator) Generate(accessor AccessGenerator, offset PeriodExpression, sourceHash string) string {
	if offset.IsConstant() {
		offsetInt := offset.AsInt()
		if offsetInt <= 0 {
			offsetInt = 1
		}

		body := fmt.Sprintf("current := %s; ", accessor.GenerateLoopValueAccess("0"))
		body += fmt.Sprintf("previous := %s; ", accessor.GenerateLoopValueAccess(fmt.Sprintf("%d", offsetInt)))
		body += "return current - previous"

		offsetPeriod := NewConstantPeriod(offsetInt + 1)
		return NewIIFECodeBuilder().
			WithWarmupCheckPeriodExpression(offsetPeriod, accessor.GetBaseOffset()).
			WithBody(body).
			Build()
	}

	body := fmt.Sprintf("offsetVal := %s; ", offset.AsIntCast())
	body += "if offsetVal <= 0 { offsetVal = 1 }; "
	body += fmt.Sprintf("current := %s; ", accessor.GenerateLoopValueAccess("0"))
	body += fmt.Sprintf("previous := %s; ", accessor.GenerateLoopValueAccess("offsetVal"))
	body += "return current - previous"

	runtimeWarmup := &runtimeOffsetPlusOne{base: offset}
	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(runtimeWarmup, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

type runtimeOffsetPlusOne struct {
	base PeriodExpression
}

func (r *runtimeOffsetPlusOne) IsConstant() bool { return false }
func (r *runtimeOffsetPlusOne) AsInt() int       { return -1 }
func (r *runtimeOffsetPlusOne) AsGoExpr() string { return r.base.AsGoExpr() + "+1" }
func (r *runtimeOffsetPlusOne) AsIntCast() string {
	return fmt.Sprintf("(%s+1)", r.base.AsIntCast())
}
func (r *runtimeOffsetPlusOne) AsFloat64Cast() string {
	return fmt.Sprintf("float64(%s+1)", r.base.AsIntCast())
}
func (r *runtimeOffsetPlusOne) AsSeriesNamePart() string { return "runtime" }

func (g *LinregIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, sourceHash string) string {
	return g.GenerateWithOffset(accessor, period, 0, sourceHash)
}

func (g *LinregIIFEGenerator) GenerateWithOffset(accessor AccessGenerator, period PeriodExpression, offset int, sourceHash string) string {
	if period.IsConstant() {
		periodInt := period.AsInt()
		formulaMultiplier := periodInt - 1 - offset

		body := "n := " + period.AsFloat64Cast() + "; "
		body += "sumX := 0.0; sumY := 0.0; sumXY := 0.0; sumX2 := 0.0; "
		body += fmt.Sprintf("for j := 0; j < %s; j++ { ", period.AsIntCast())
		body += fmt.Sprintf("x := float64(j); y := %s; ", accessor.GenerateLoopValueAccess(fmt.Sprintf("%d - j - 1", periodInt)))
		body += "sumX += x; sumY += y; sumXY += x * y; sumX2 += x * x"
		body += " }; "
		body += "slope := (n * sumXY - sumX * sumY) / (n * sumX2 - sumX * sumX); "
		body += "intercept := (sumY - slope * sumX) / n; "
		body += fmt.Sprintf("return intercept + slope * float64(%d)", formulaMultiplier)

		return NewIIFECodeBuilder().
			WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
			WithBody(body).
			Build()
	}

	body := fmt.Sprintf("periodVal := %s; ", period.AsIntCast())
	body += "n := " + period.AsFloat64Cast() + "; "
	body += fmt.Sprintf("formulaMultiplier := periodVal - 1 - %d; ", offset)
	body += "sumX := 0.0; sumY := 0.0; sumXY := 0.0; sumX2 := 0.0; "
	body += "for j := 0; j < periodVal; j++ { "
	body += fmt.Sprintf("x := float64(j); y := %s; ", accessor.GenerateLoopValueAccess("periodVal - j - 1"))
	body += "sumX += x; sumY += y; sumXY += x * y; sumX2 += x * x"
	body += " }; "
	body += "slope := (n * sumXY - sumX * sumY) / (n * sumX2 - sumX * sumX); "
	body += "intercept := (sumY - slope * sumX) / n; "
	body += "return intercept + slope * float64(formulaMultiplier)"

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

type TSIIIFEGenerator struct{ namingStrategy series_naming.Strategy }

func (g *TSIIIFEGenerator) GenerateDualPeriod(accessor AccessGenerator, leftPeriod, rightPeriod PeriodExpression, sourceHash string) string {
	context := NewArrowFunctionIndicatorContext()
	periodPart := leftPeriod.AsSeriesNamePart() + "_" + rightPeriod.AsSeriesNamePart()
	varName := g.namingStrategy.GenerateName("tsi", periodPart, sourceHash)

	builder := NewTSIIndicatorBuilder(varName, leftPeriod, rightPeriod, accessor, context)
	statefulCode := builder.Build()
	seriesAccess := fmt.Sprintf("arrowCtx.GetOrCreateSeries(%q).Get(0)", varName)

	return fmt.Sprintf("func() float64 {\n%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}

type PivotHighIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type PivotLowIIFEGenerator struct{ namingStrategy series_naming.Strategy }

func (g *PivotHighIIFEGenerator) GenerateDualPeriod(accessor AccessGenerator, leftPeriod, rightPeriod PeriodExpression, sourceHash string) string {
	return generatePivotIIFE(accessor, leftPeriod, rightPeriod, ">=")
}

func (g *PivotLowIIFEGenerator) GenerateDualPeriod(accessor AccessGenerator, leftPeriod, rightPeriod PeriodExpression, sourceHash string) string {
	return generatePivotIIFE(accessor, leftPeriod, rightPeriod, "<=")
}

func generatePivotIIFE(accessor AccessGenerator, leftPeriod, rightPeriod PeriodExpression, comparisonOp string) string {
	if leftPeriod.IsConstant() && rightPeriod.IsConstant() {
		return generatePivotUnrolled(accessor, leftPeriod.AsInt(), rightPeriod.AsInt(), comparisonOp)
	}
	return generatePivotRuntimeLoop(accessor, leftPeriod, rightPeriod, comparisonOp)
}

func generatePivotUnrolled(accessor AccessGenerator, leftInt, rightInt int, comparisonOp string) string {
	totalWindow := leftInt + rightInt + 1

	body := fmt.Sprintf("centerValue := %s; ", accessor.GenerateLoopValueAccess(fmt.Sprintf("%d", rightInt)))
	body += "if math.IsNaN(centerValue) { return math.NaN() }; "
	body += "isPivot := true; "

	for j := 0; j < leftInt; j++ {
		offset := totalWindow - 1 - j
		body += fmt.Sprintf("if leftVal := %s; !math.IsNaN(leftVal) && leftVal %s centerValue { isPivot = false }; ", accessor.GenerateLoopValueAccess(fmt.Sprintf("%d", offset)), comparisonOp)
	}

	for j := 1; j <= rightInt; j++ {
		offset := rightInt - j
		body += fmt.Sprintf("if rightVal := %s; !math.IsNaN(rightVal) && rightVal %s centerValue { isPivot = false }; ", accessor.GenerateLoopValueAccess(fmt.Sprintf("%d", offset)), comparisonOp)
	}

	body += "if isPivot { return centerValue }; "
	body += "return math.NaN()"

	pivotPeriod := NewConstantPeriod(totalWindow)
	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(pivotPeriod, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func generatePivotRuntimeLoop(accessor AccessGenerator, leftPeriod, rightPeriod PeriodExpression, comparisonOp string) string {
	leftExpr := leftPeriod.AsIntCast()
	rightExpr := rightPeriod.AsIntCast()

	code := "func() float64 { "
	code += fmt.Sprintf("leftBars := %s; rightBars := %s; ", leftExpr, rightExpr)
	code += "totalWindow := leftBars + rightBars + 1; "
	code += fmt.Sprintf("if ctx.BarIndex < totalWindow - 1 + %d { return math.NaN() }; ", accessor.GetBaseOffset())
	code += fmt.Sprintf("centerValue := %s; ", accessor.GenerateLoopValueAccess("rightBars"))
	code += "if math.IsNaN(centerValue) { return math.NaN() }; "
	code += "isPivot := true; "
	code += fmt.Sprintf("for j := 0; j < leftBars && isPivot; j++ { if leftVal := %s; !math.IsNaN(leftVal) && leftVal %s centerValue { isPivot = false } }; ", accessor.GenerateLoopValueAccess("totalWindow - 1 - j"), comparisonOp)
	code += fmt.Sprintf("for j := 1; j <= rightBars && isPivot; j++ { if rightVal := %s; !math.IsNaN(rightVal) && rightVal %s centerValue { isPivot = false } }; ", accessor.GenerateLoopValueAccess("rightBars - j"), comparisonOp)
	code += "if isPivot { return centerValue }; "
	code += "return math.NaN() }()"

	return code
}
