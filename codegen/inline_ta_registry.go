package codegen

import (
	"fmt"

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

	r.Register("ta.sma", &SMAIIFEGenerator{namingStrategy: windowNamer})
	r.Register("sma", &SMAIIFEGenerator{namingStrategy: windowNamer})
	r.Register("ta.wma", &WMAIIFEGenerator{namingStrategy: windowNamer})
	r.Register("wma", &WMAIIFEGenerator{namingStrategy: windowNamer})
	r.Register("ta.stdev", &STDEVIIFEGenerator{namingStrategy: windowNamer})
	r.Register("stdev", &STDEVIIFEGenerator{namingStrategy: windowNamer})
	r.Register("ta.highest", &HighestIIFEGenerator{namingStrategy: windowNamer})
	r.Register("highest", &HighestIIFEGenerator{namingStrategy: windowNamer})
	r.Register("ta.lowest", &LowestIIFEGenerator{namingStrategy: windowNamer})
	r.Register("lowest", &LowestIIFEGenerator{namingStrategy: windowNamer})
	r.Register("ta.change", &ChangeIIFEGenerator{namingStrategy: windowNamer})
	r.Register("change", &ChangeIIFEGenerator{namingStrategy: windowNamer})

	r.Register("ta.ema", &EMAIIFEGenerator{namingStrategy: statefulNamer})
	r.Register("ema", &EMAIIFEGenerator{namingStrategy: statefulNamer})
	r.Register("ta.rma", &RMAIIFEGenerator{namingStrategy: statefulNamer})
	r.Register("rma", &RMAIIFEGenerator{namingStrategy: statefulNamer})
	r.Register("ta.rsi", &RSIIIFEGenerator{namingStrategy: statefulNamer})
	r.Register("rsi", &RSIIIFEGenerator{namingStrategy: statefulNamer})

	r.RegisterDualPeriod("ta.pivothigh", &PivotHighIIFEGenerator{namingStrategy: windowNamer})
	r.RegisterDualPeriod("ta.pivotlow", &PivotLowIIFEGenerator{namingStrategy: windowNamer})
}

func (r *InlineTAIIFERegistry) Register(name string, generator InlineTAIIFEGenerator) {
	r.generators[name] = generator
}

func (r *InlineTAIIFERegistry) RegisterDualPeriod(name string, generator InlineTADualPeriodGenerator) {
	r.dualPeriodGenerators[name] = generator
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

type SMAIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type EMAIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type RMAIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type RSIIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type WMAIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type STDEVIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type HighestIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type LowestIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type ChangeIIFEGenerator struct{ namingStrategy series_naming.Strategy }

func (g *SMAIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, sourceHash string) string {
	body := fmt.Sprintf("sum := 0.0; for j := 0; j < %s; j++ { sum += %s }; ", period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += fmt.Sprintf("return sum / %s", period.AsFloat64Cast())

	/* Previous bar access requires additional warmup bar */
	warmupPeriod := period.AsInt() + accessor.GetBaseOffset()
	return NewIIFECodeBuilder().WithWarmupCheck(warmupPeriod).WithBody(body).Build()
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
	body := fmt.Sprintf("sum := 0.0; weightSum := 0.0; for j := 0; j < %s; j++ { weight := float64(%s - j); sum += weight * %s; weightSum += weight }; ", period.AsIntCast(), period.AsGoExpr(), accessor.GenerateLoopValueAccess("j"))
	body += "return sum / weightSum"

	warmupPeriod := period.AsInt() + accessor.GetBaseOffset()
	return NewIIFECodeBuilder().WithWarmupCheck(warmupPeriod).WithBody(body).Build()
}

func (g *STDEVIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, sourceHash string) string {
	body := fmt.Sprintf("sum := 0.0; for j := 0; j < %s; j++ { sum += %s }; ", period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += fmt.Sprintf("mean := sum / %s; ", period.AsFloat64Cast())
	body += fmt.Sprintf("variance := 0.0; for j := 0; j < %s; j++ { diff := %s - mean; variance += diff * diff }; ", period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += fmt.Sprintf("return math.Sqrt(variance / %s)", period.AsFloat64Cast())

	warmupPeriod := period.AsInt() + accessor.GetBaseOffset()
	return NewIIFECodeBuilder().WithWarmupCheck(warmupPeriod).WithBody(body).Build()
}

func (g *HighestIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, sourceHash string) string {
	periodInt := period.AsInt()
	body := fmt.Sprintf("highest := %s; ", accessor.GenerateInitialValueAccess(periodInt))
	body += fmt.Sprintf("for j := %d; j >= 0; j-- { v := %s; if v > highest { highest = v } }; ", periodInt-1, accessor.GenerateLoopValueAccess("j"))
	body += "return highest"

	warmupPeriod := periodInt + accessor.GetBaseOffset()
	return NewIIFECodeBuilder().WithWarmupCheck(warmupPeriod).WithBody(body).Build()
}

func (g *LowestIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, sourceHash string) string {
	periodInt := period.AsInt()
	body := fmt.Sprintf("lowest := %s; ", accessor.GenerateInitialValueAccess(periodInt))
	body += fmt.Sprintf("for j := %d; j >= 0; j-- { v := %s; if v < lowest { lowest = v } }; ", periodInt-1, accessor.GenerateLoopValueAccess("j"))
	body += "return lowest"

	warmupPeriod := periodInt + accessor.GetBaseOffset()
	return NewIIFECodeBuilder().WithWarmupCheck(warmupPeriod).WithBody(body).Build()
}

func (g *ChangeIIFEGenerator) Generate(accessor AccessGenerator, offset PeriodExpression, sourceHash string) string {
	offsetInt := offset.AsInt()
	if offsetInt <= 0 {
		offsetInt = 1
	}

	body := fmt.Sprintf("current := %s; ", accessor.GenerateLoopValueAccess("0"))
	body += fmt.Sprintf("previous := %s; ", accessor.GenerateLoopValueAccess(fmt.Sprintf("%d", offsetInt)))
	body += "return current - previous"

	warmupPeriod := offsetInt + 1 + accessor.GetBaseOffset()
	return NewIIFECodeBuilder().WithWarmupCheck(warmupPeriod).WithBody(body).Build()
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

	warmupPeriod := totalWindow + accessor.GetBaseOffset()
	return NewIIFECodeBuilder().WithWarmupCheck(warmupPeriod).WithBody(body).Build()
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
