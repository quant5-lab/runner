package codegen

import "fmt"

// TAFunctionGenerator generates code for technical analysis functions
type TAFunctionGenerator interface {
	// GenerateCode generates the implementation code for the TA function
	GenerateCode(accessor SeriesAccessor, varName string, period int, indenter Indenter) string
}

// Indenter handles code indentation
type Indenter interface {
	Indent() string
	IncreaseIndent()
	DecreaseIndent()
}

// SimpleIndenter basic indentation implementation
type SimpleIndenter struct {
	level int
}

func NewSimpleIndenter(initialLevel int) *SimpleIndenter {
	return &SimpleIndenter{level: initialLevel}
}

func (i *SimpleIndenter) Indent() string {
	result := ""
	for j := 0; j < i.level; j++ {
		result += "\t"
	}
	return result
}

func (i *SimpleIndenter) IncreaseIndent() {
	i.level++
}

func (i *SimpleIndenter) DecreaseIndent() {
	i.level--
}

// SMAGenerator generates Simple Moving Average code
type SMAGenerator struct{}

func (g *SMAGenerator) GenerateCode(accessor SeriesAccessor, varName string, period int, indenter Indenter) string {
	code := ""
	ind := indenter.Indent()

	code += ind + fmt.Sprintf("/* Inline SMA(%d) */\n", period)
	code += ind + fmt.Sprintf("if ctx.BarIndex < %d-1 {\n", period)
	indenter.IncreaseIndent()
	code += indenter.Indent() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	indenter.DecreaseIndent()
	code += ind + "} else {\n"
	indenter.IncreaseIndent()

	code += indenter.Indent() + "sum := 0.0\n"

	if accessor.RequiresNaNCheck() {
		code += indenter.Indent() + "hasNaN := false\n"
	}

	code += indenter.Indent() + fmt.Sprintf("for j := 0; j < %d; j++ {\n", period)
	indenter.IncreaseIndent()

	if accessor.RequiresNaNCheck() {
		code += indenter.Indent() + fmt.Sprintf("val := %s\n", accessor.GetAccessExpression("j"))
		code += indenter.Indent() + "if math.IsNaN(val) {\n"
		indenter.IncreaseIndent()
		code += indenter.Indent() + "hasNaN = true\n"
		code += indenter.Indent() + "break\n"
		indenter.DecreaseIndent()
		code += indenter.Indent() + "}\n"
		code += indenter.Indent() + "sum += val\n"
	} else {
		code += indenter.Indent() + fmt.Sprintf("sum += %s\n", accessor.GetAccessExpression("j"))
	}

	indenter.DecreaseIndent()
	code += indenter.Indent() + "}\n"

	if accessor.RequiresNaNCheck() {
		code += indenter.Indent() + "if hasNaN {\n"
		indenter.IncreaseIndent()
		code += indenter.Indent() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
		indenter.DecreaseIndent()
		code += indenter.Indent() + "} else {\n"
		indenter.IncreaseIndent()
		code += indenter.Indent() + fmt.Sprintf("%sSeries.Set(sum / %d.0)\n", varName, period)
		indenter.DecreaseIndent()
		code += indenter.Indent() + "}\n"
	} else {
		code += indenter.Indent() + fmt.Sprintf("%sSeries.Set(sum / %d.0)\n", varName, period)
	}

	indenter.DecreaseIndent()
	code += ind + "}\n"

	return code
}

// EMAGenerator generates Exponential Moving Average code
type EMAGenerator struct{}

func (g *EMAGenerator) GenerateCode(accessor SeriesAccessor, varName string, period int, indenter Indenter) string {
	code := ""
	ind := indenter.Indent()

	code += ind + fmt.Sprintf("/* Inline EMA(%d) */\n", period)
	code += ind + fmt.Sprintf("if ctx.BarIndex < %d-1 {\n", period)
	indenter.IncreaseIndent()
	code += indenter.Indent() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	indenter.DecreaseIndent()
	code += ind + "} else {\n"
	indenter.IncreaseIndent()

	code += indenter.Indent() + fmt.Sprintf("alpha := 2.0 / float64(%d+1)\n", period)
	code += indenter.Indent() + fmt.Sprintf("ema := %s\n", accessor.GetAccessExpression(fmt.Sprintf("%d-1", period)))

	if accessor.RequiresNaNCheck() {
		code += indenter.Indent() + "if math.IsNaN(ema) {\n"
		indenter.IncreaseIndent()
		code += indenter.Indent() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
		indenter.DecreaseIndent()
		code += indenter.Indent() + "} else {\n"
		indenter.IncreaseIndent()
	}

	code += indenter.Indent() + fmt.Sprintf("for j := %d-2; j >= 0; j-- {\n", period)
	indenter.IncreaseIndent()

	if accessor.RequiresNaNCheck() {
		code += indenter.Indent() + fmt.Sprintf("val := %s\n", accessor.GetAccessExpression("j"))
		code += indenter.Indent() + "if math.IsNaN(val) {\n"
		indenter.IncreaseIndent()
		code += indenter.Indent() + "ema = math.NaN()\n"
		code += indenter.Indent() + "break\n"
		indenter.DecreaseIndent()
		code += indenter.Indent() + "}\n"
		code += indenter.Indent() + "ema = alpha*val + (1-alpha)*ema\n"
	} else {
		code += indenter.Indent() + fmt.Sprintf("ema = alpha*%s + (1-alpha)*ema\n", accessor.GetAccessExpression("j"))
	}

	indenter.DecreaseIndent()
	code += indenter.Indent() + "}\n"
	code += indenter.Indent() + fmt.Sprintf("%sSeries.Set(ema)\n", varName)

	if accessor.RequiresNaNCheck() {
		indenter.DecreaseIndent()
		code += indenter.Indent() + "}\n"
	}

	indenter.DecreaseIndent()
	code += ind + "}\n"

	return code
}

// STDEVGenerator generates Standard Deviation code
type STDEVGenerator struct{}

func (g *STDEVGenerator) GenerateCode(accessor SeriesAccessor, varName string, period int, indenter Indenter) string {
	code := ""
	ind := indenter.Indent()

	code += ind + fmt.Sprintf("/* Inline STDEV(%d) */\n", period)
	code += ind + fmt.Sprintf("if ctx.BarIndex < %d-1 {\n", period)
	indenter.IncreaseIndent()
	code += indenter.Indent() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	indenter.DecreaseIndent()
	code += ind + "} else {\n"
	indenter.IncreaseIndent()

	code += indenter.Indent() + "sum := 0.0\n"

	if accessor.RequiresNaNCheck() {
		code += indenter.Indent() + "hasNaN := false\n"
	}

	code += indenter.Indent() + fmt.Sprintf("for j := 0; j < %d; j++ {\n", period)
	indenter.IncreaseIndent()

	if accessor.RequiresNaNCheck() {
		code += indenter.Indent() + fmt.Sprintf("val := %s\n", accessor.GetAccessExpression("j"))
		code += indenter.Indent() + "if math.IsNaN(val) {\n"
		indenter.IncreaseIndent()
		code += indenter.Indent() + "hasNaN = true\n"
		code += indenter.Indent() + "break\n"
		indenter.DecreaseIndent()
		code += indenter.Indent() + "}\n"
		code += indenter.Indent() + "sum += val\n"
	} else {
		code += indenter.Indent() + fmt.Sprintf("sum += %s\n", accessor.GetAccessExpression("j"))
	}

	indenter.DecreaseIndent()
	code += indenter.Indent() + "}\n"

	if accessor.RequiresNaNCheck() {
		code += indenter.Indent() + "if hasNaN {\n"
		indenter.IncreaseIndent()
		code += indenter.Indent() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
		indenter.DecreaseIndent()
		code += indenter.Indent() + "} else {\n"
		indenter.IncreaseIndent()
	}

	code += indenter.Indent() + fmt.Sprintf("mean := sum / %d.0\n", period)
	code += indenter.Indent() + "variance := 0.0\n"
	code += indenter.Indent() + fmt.Sprintf("for j := 0; j < %d; j++ {\n", period)
	indenter.IncreaseIndent()
	code += indenter.Indent() + fmt.Sprintf("diff := %s - mean\n", accessor.GetAccessExpression("j"))
	code += indenter.Indent() + "variance += diff * diff\n"
	indenter.DecreaseIndent()
	code += indenter.Indent() + "}\n"
	code += indenter.Indent() + fmt.Sprintf("variance /= %d.0\n", period)
	code += indenter.Indent() + fmt.Sprintf("%sSeries.Set(math.Sqrt(variance))\n", varName)

	if accessor.RequiresNaNCheck() {
		indenter.DecreaseIndent()
		code += indenter.Indent() + "}\n"
	}

	indenter.DecreaseIndent()
	code += ind + "}\n"

	return code
}

// TAFunctionRegistry maps function names to their generators
type TAFunctionRegistry struct {
	generators map[string]TAFunctionGenerator
}

func NewTAFunctionRegistry() *TAFunctionRegistry {
	return &TAFunctionRegistry{
		generators: map[string]TAFunctionGenerator{
			"ta.sma":   &SMAGenerator{},
			"ta.ema":   &EMAGenerator{},
			"ta.stdev": &STDEVGenerator{},
		},
	}
}

func (r *TAFunctionRegistry) GetGenerator(funcName string) (TAFunctionGenerator, bool) {
	gen, exists := r.generators[funcName]
	return gen, exists
}
