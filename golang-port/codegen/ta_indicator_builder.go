package codegen

import "fmt"

// TAIndicatorBuilder constructs technical analysis indicator code using the Builder pattern.
//
// This builder provides a fluent interface for generating inline TA indicator calculations
// (SMA, EMA, STDEV, etc.) with proper warmup period handling, NaN propagation, and
// indentation management.
//
// Usage:
//
//	// Create accessor for data source
//	accessor := CreateAccessGenerator("close")
//
//	// Build SMA indicator
//	builder := NewTAIndicatorBuilder("SMA", "sma20", 20, accessor, false)
//	builder.WithAccumulator(NewSumAccumulator())
//	code := builder.Build()
//
//	// Build STDEV indicator (requires two passes)
//	// Pass 1: Calculate mean
//	meanBuilder := NewTAIndicatorBuilder("STDEV", "stdev20", 20, accessor, false)
//	meanBuilder.WithAccumulator(NewSumAccumulator())
//	meanCode := meanBuilder.Build()
//
//	// Pass 2: Calculate variance
//	varianceBuilder := NewTAIndicatorBuilder("STDEV", "stdev20", 20, accessor, false)
//	varianceBuilder.WithAccumulator(NewVarianceAccumulator("mean"))
//	varianceCode := varianceBuilder.Build()
//
// Design:
//   - Builder Pattern: Step-by-step construction of complex indicator code
//   - Strategy Pattern: Pluggable accumulation strategies (Sum, Variance, EMA)
//   - Single Responsibility: Each component handles one concern
//   - Open/Closed: Easy to extend with new indicator types
type TAIndicatorBuilder struct {
	indicatorName string              // Name of the indicator (SMA, EMA, STDEV)
	varName       string              // Variable name for the Series
	period        int                 // Lookback period
	warmupChecker *WarmupChecker      // Handles warmup period validation
	loopGen       *LoopGenerator      // Generates for loops with NaN handling
	accumulator   AccumulatorStrategy // Accumulation logic (sum, variance, ema)
	indenter      CodeIndenter        // Manages code indentation
}

// NewTAIndicatorBuilder creates a new builder for generating TA indicator code.
//
// Parameters:
//   - name: Indicator name (e.g., "SMA", "EMA", "STDEV")
//   - varName: Variable name for the output Series (e.g., "sma20")
//   - period: Lookback period for the indicator
//   - accessor: AccessGenerator for retrieving data values (Series or OHLCV field)
//   - needsNaN: Whether to add NaN checking in the accumulation loop
//
// Returns a builder that must be configured with an accumulator before calling Build().
func NewTAIndicatorBuilder(name, varName string, period int, accessor AccessGenerator, needsNaN bool) *TAIndicatorBuilder {
	return &TAIndicatorBuilder{
		indicatorName: name,
		varName:       varName,
		period:        period,
		warmupChecker: NewWarmupChecker(period),
		loopGen:       NewLoopGenerator(period, accessor, needsNaN),
		indenter:      NewCodeIndenter(),
	}
}

// WithAccumulator sets the accumulation strategy for this indicator.
//
// Common strategies:
//   - NewSumAccumulator(): For SMA calculations
//   - NewEMAAccumulator(alpha): For EMA calculations
//   - NewVarianceAccumulator(meanVar): For STDEV variance calculation
//
// Returns the builder for method chaining.
func (b *TAIndicatorBuilder) WithAccumulator(acc AccumulatorStrategy) *TAIndicatorBuilder {
	b.accumulator = acc
	return b
}

// BuildHeader generates the comment header for the indicator code.
func (b *TAIndicatorBuilder) BuildHeader() string {
	return b.indenter.Line(fmt.Sprintf("/* Inline %s(%d) */", b.indicatorName, b.period))
}

// BuildWarmupCheck generates the warmup period check that sets NaN during warmup.
func (b *TAIndicatorBuilder) BuildWarmupCheck() string {
	return b.warmupChecker.GenerateCheck(b.varName, &b.indenter)
}

// BuildInitialization generates variable initialization code for the accumulator.
func (b *TAIndicatorBuilder) BuildInitialization() string {
	if b.accumulator == nil {
		return ""
	}

	code := ""
	initCode := b.accumulator.Initialize()
	if initCode != "" {
		code += b.indenter.Line(initCode)
	}
	return code
}

func (b *TAIndicatorBuilder) BuildLoop(loopBody func(valueExpr string) string) string {
	code := b.loopGen.GenerateForwardLoop(&b.indenter)

	b.indenter.IncreaseIndent()
	valueAccess := b.loopGen.GenerateValueAccess()

	if b.loopGen.RequiresNaNCheck() && b.accumulator.NeedsNaNGuard() {
		code += b.indenter.Line(fmt.Sprintf("val := %s", valueAccess))
		code += b.indenter.Line("if math.IsNaN(val) {")
		b.indenter.IncreaseIndent()
		code += b.indenter.Line("hasNaN = true")
		code += b.indenter.Line("break")
		b.indenter.DecreaseIndent()
		code += b.indenter.Line("}")
		code += loopBody("val")
	} else {
		code += loopBody(valueAccess)
	}

	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")

	return code
}

func (b *TAIndicatorBuilder) BuildFinalization(resultExpr string) string {
	code := ""

	if b.accumulator.NeedsNaNGuard() {
		code += b.indenter.Line("if hasNaN {")
		b.indenter.IncreaseIndent()
		code += b.indenter.Line(fmt.Sprintf("%sSeries.Set(math.NaN())", b.varName))
		b.indenter.DecreaseIndent()
		code += b.indenter.Line("} else {")
		b.indenter.IncreaseIndent()
		code += b.indenter.Line(fmt.Sprintf("%sSeries.Set(%s)", b.varName, resultExpr))
		b.indenter.DecreaseIndent()
		code += b.indenter.Line("}")
	} else {
		code += b.indenter.Line(fmt.Sprintf("%sSeries.Set(%s)", b.varName, resultExpr))
	}

	return code
}

func (b *TAIndicatorBuilder) CloseBlock() string {
	b.indenter.DecreaseIndent()
	return b.indenter.Line("}")
}

// Build generates complete TA indicator code
func (b *TAIndicatorBuilder) Build() string {
	b.indenter.IncreaseIndent() // Start at indent level 1

	code := b.BuildHeader()
	code += b.BuildWarmupCheck()

	b.indenter.IncreaseIndent()
	code += b.BuildInitialization()

	if b.accumulator != nil {
		code += b.BuildLoop(func(val string) string {
			return b.indenter.Line(b.accumulator.Accumulate(val))
		})

		finalizeCode := b.accumulator.Finalize(b.period)
		code += b.BuildFinalization(finalizeCode)
	}

	code += b.CloseBlock()

	return code
}

// BuildEMA generates EMA-specific code with backward loop and initial value handling
func (b *TAIndicatorBuilder) BuildEMA() string {
	b.indenter.IncreaseIndent() // Start at indent level 1

	code := b.BuildHeader()
	code += b.BuildWarmupCheck()

	b.indenter.IncreaseIndent()

	// Calculate alpha and initialize EMA with oldest value
	code += b.indenter.Line(fmt.Sprintf("alpha := 2.0 / float64(%d+1)", b.period))
	initialAccess := b.loopGen.accessor.GenerateInitialValueAccess(b.period)
	code += b.indenter.Line(fmt.Sprintf("ema := %s", initialAccess))

	// Check if initial value is NaN
	code += b.indenter.Line("if math.IsNaN(ema) {")
	b.indenter.IncreaseIndent()
	code += b.indenter.Line(fmt.Sprintf("%sSeries.Set(math.NaN())", b.varName))
	b.indenter.DecreaseIndent()
	code += b.indenter.Line("} else {")
	b.indenter.IncreaseIndent()

	// Loop backwards from period-2 to 0
	code += b.loopGen.GenerateBackwardLoop(&b.indenter)
	b.indenter.IncreaseIndent()

	valueAccess := b.loopGen.GenerateValueAccess()

	if b.loopGen.RequiresNaNCheck() {
		code += b.indenter.Line(fmt.Sprintf("val := %s", valueAccess))
		code += b.indenter.Line("if math.IsNaN(val) {")
		b.indenter.IncreaseIndent()
		code += b.indenter.Line("ema = math.NaN()")
		code += b.indenter.Line("break")
		b.indenter.DecreaseIndent()
		code += b.indenter.Line("}")
		code += b.indenter.Line("ema = alpha*val + (1-alpha)*ema")
	} else {
		code += b.indenter.Line(fmt.Sprintf("ema = alpha*%s + (1-alpha)*ema", valueAccess))
	}

	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")

	// Set final result
	code += b.indenter.Line(fmt.Sprintf("%sSeries.Set(ema)", b.varName))

	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}") // end else (initial value check)

	code += b.CloseBlock()

	return code
}

// CodeIndenter implements Indenter interface
type CodeIndenter struct {
	level int
	tab   string
}

func NewCodeIndenter() CodeIndenter {
	return CodeIndenter{level: 0, tab: "\t"}
}

func (c *CodeIndenter) Line(code string) string {
	indent := ""
	for i := 0; i < c.level; i++ {
		indent += c.tab
	}
	return indent + code + "\n"
}

func (c *CodeIndenter) Indent(fn func() string) string {
	c.level++
	result := fn()
	c.level--
	return result
}

func (c *CodeIndenter) CurrentLevel() int {
	return c.level
}

func (c *CodeIndenter) IncreaseIndent() {
	c.level++
}

func (c *CodeIndenter) DecreaseIndent() {
	if c.level > 0 {
		c.level--
	}
}

// BuildSTDEV generates STDEV-specific code with two-pass algorithm (mean then variance)
func (b *TAIndicatorBuilder) BuildSTDEV() string {
	b.indenter.IncreaseIndent() // Start at indent level 1

	code := b.BuildHeader()
	code += b.BuildWarmupCheck()

	b.indenter.IncreaseIndent()

	// Pass 1: Calculate mean
	code += b.indenter.Line("sum := 0.0")
	if b.loopGen.RequiresNaNCheck() {
		code += b.indenter.Line("hasNaN := false")
	}

	// Forward loop for sum
	code += b.loopGen.GenerateForwardLoop(&b.indenter)
	b.indenter.IncreaseIndent()

	valueAccess := b.loopGen.GenerateValueAccess()
	if b.loopGen.RequiresNaNCheck() {
		code += b.indenter.Line(fmt.Sprintf("val := %s", valueAccess))
		code += b.indenter.Line("if math.IsNaN(val) {")
		b.indenter.IncreaseIndent()
		code += b.indenter.Line("hasNaN = true")
		code += b.indenter.Line("break")
		b.indenter.DecreaseIndent()
		code += b.indenter.Line("}")
		code += b.indenter.Line("sum += val")
	} else {
		code += b.indenter.Line(fmt.Sprintf("sum += %s", valueAccess))
	}

	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")

	// Check for NaN and calculate mean
	if b.loopGen.RequiresNaNCheck() {
		code += b.indenter.Line("if hasNaN {")
		b.indenter.IncreaseIndent()
		code += b.indenter.Line(fmt.Sprintf("%sSeries.Set(math.NaN())", b.varName))
		b.indenter.DecreaseIndent()
		code += b.indenter.Line("} else {")
		b.indenter.IncreaseIndent()
	}

	code += b.indenter.Line(fmt.Sprintf("mean := sum / %d.0", b.period))

	// Pass 2: Calculate variance
	code += b.indenter.Line("variance := 0.0")
	code += b.loopGen.GenerateForwardLoop(&b.indenter)
	b.indenter.IncreaseIndent()

	code += b.indenter.Line(fmt.Sprintf("diff := %s - mean", valueAccess))
	code += b.indenter.Line("variance += diff * diff")

	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")

	code += b.indenter.Line(fmt.Sprintf("variance /= %d.0", b.period))
	code += b.indenter.Line(fmt.Sprintf("%sSeries.Set(math.Sqrt(variance))", b.varName))

	if b.loopGen.RequiresNaNCheck() {
		b.indenter.DecreaseIndent()
		code += b.indenter.Line("}")
	}

	code += b.CloseBlock()

	return code
}

// BuildDEV generates DEV-specific code with two-pass algorithm (mean then absolute deviation)
func (b *TAIndicatorBuilder) BuildDEV() string {
	b.indenter.IncreaseIndent() // Start at indent level 1

	code := b.BuildHeader()
	code += b.BuildWarmupCheck()

	b.indenter.IncreaseIndent()

	// Pass 1: Calculate mean
	code += b.indenter.Line("sum := 0.0")
	if b.loopGen.RequiresNaNCheck() {
		code += b.indenter.Line("hasNaN := false")
	}

	// Forward loop for sum
	code += b.loopGen.GenerateForwardLoop(&b.indenter)
	b.indenter.IncreaseIndent()

	valueAccess := b.loopGen.GenerateValueAccess()
	if b.loopGen.RequiresNaNCheck() {
		code += b.indenter.Line(fmt.Sprintf("val := %s", valueAccess))
		code += b.indenter.Line("if math.IsNaN(val) {")
		b.indenter.IncreaseIndent()
		code += b.indenter.Line("hasNaN = true")
		code += b.indenter.Line("break")
		b.indenter.DecreaseIndent()
		code += b.indenter.Line("}")
		code += b.indenter.Line("sum += val")
	} else {
		code += b.indenter.Line(fmt.Sprintf("sum += %s", valueAccess))
	}

	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")

	// Check for NaN and calculate mean
	if b.loopGen.RequiresNaNCheck() {
		code += b.indenter.Line("if hasNaN {")
		b.indenter.IncreaseIndent()
		code += b.indenter.Line(fmt.Sprintf("%sSeries.Set(math.NaN())", b.varName))
		b.indenter.DecreaseIndent()
		code += b.indenter.Line("} else {")
		b.indenter.IncreaseIndent()
	}

	code += b.indenter.Line(fmt.Sprintf("mean := sum / %d.0", b.period))

	// Pass 2: Calculate absolute deviation
	code += b.indenter.Line("deviation := 0.0")
	code += b.loopGen.GenerateForwardLoop(&b.indenter)
	b.indenter.IncreaseIndent()

	code += b.indenter.Line(fmt.Sprintf("diff := %s - mean", valueAccess))
	code += b.indenter.Line("if diff < 0 { diff = -diff }")
	code += b.indenter.Line("deviation += diff")

	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")

	code += b.indenter.Line(fmt.Sprintf("%sSeries.Set(deviation / %d.0)", b.varName, b.period))

	if b.loopGen.RequiresNaNCheck() {
		b.indenter.DecreaseIndent()
		code += b.indenter.Line("}")
	}

	code += b.CloseBlock()

	return code
}
