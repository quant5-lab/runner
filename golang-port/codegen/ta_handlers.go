package codegen

import (
	"fmt"

	"github.com/borisquantlab/pinescript-go/ast"
)

// SMAHandler generates inline code for Simple Moving Average calculations
type SMAHandler struct{}

func (h *SMAHandler) CanHandle(funcName string) bool {
	return funcName == "ta.sma" || funcName == "sma"
}

func (h *SMAHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	sourceExpr, period, err := extractTAArguments(g, call, "ta.sma")
	if err != nil {
		return "", err
	}

	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify(sourceExpr)
	accessGen := CreateAccessGenerator(sourceInfo)
	needsNaN := sourceInfo.IsSeriesVariable()

	builder := NewTAIndicatorBuilder("ta.sma", varName, period, accessGen, needsNaN)
	builder.WithAccumulator(NewSumAccumulator())
	return g.indentCode(builder.Build()), nil
}

// EMAHandler generates inline code for Exponential Moving Average calculations
type EMAHandler struct{}

func (h *EMAHandler) CanHandle(funcName string) bool {
	return funcName == "ta.ema" || funcName == "ema"
}

func (h *EMAHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	sourceExpr, period, err := extractTAArguments(g, call, "ta.ema")
	if err != nil {
		return "", err
	}

	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify(sourceExpr)
	accessGen := CreateAccessGenerator(sourceInfo)
	needsNaN := sourceInfo.IsSeriesVariable()

	return g.generateEMA(varName, period, accessGen, needsNaN)
}

// STDEVHandler generates inline code for Standard Deviation calculations
type STDEVHandler struct{}

func (h *STDEVHandler) CanHandle(funcName string) bool {
	return funcName == "ta.stdev" || funcName == "stdev"
}

func (h *STDEVHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	sourceExpr, period, err := extractTAArguments(g, call, "ta.stdev")
	if err != nil {
		return "", err
	}

	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify(sourceExpr)
	accessGen := CreateAccessGenerator(sourceInfo)
	needsNaN := sourceInfo.IsSeriesVariable()

	return g.generateSTDEV(varName, period, accessGen, needsNaN)
}

// ATRHandler generates inline code for Average True Range calculations
type ATRHandler struct{}

func (h *ATRHandler) CanHandle(funcName string) bool {
	return funcName == "ta.atr" || funcName == "atr"
}

func (h *ATRHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 {
		return "", fmt.Errorf("ta.atr requires 1 argument (period)")
	}

	periodArg, ok := call.Arguments[0].(*ast.Literal)
	if !ok {
		return "", fmt.Errorf("ta.atr period must be literal")
	}

	period, err := extractPeriod(periodArg)
	if err != nil {
		return "", fmt.Errorf("ta.atr: %w", err)
	}

	return g.generateInlineATR(varName, period)
}

// RMAHandler generates inline code for RMA (Relative Moving Average) calculations
type RMAHandler struct{}

func (h *RMAHandler) CanHandle(funcName string) bool {
	return funcName == "ta.rma" || funcName == "rma"
}

func (h *RMAHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	// RMA is an exponentially weighted moving average with alpha = 1/period
	// Same as EMA but with different smoothing factor
	sourceExpr, period, err := extractTAArguments(g, call, "ta.rma")
	if err != nil {
		return "", err
	}

	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify(sourceExpr)
	accessGen := CreateAccessGenerator(sourceInfo)
	needsNaN := sourceInfo.IsSeriesVariable()

	// For RMA, use inline generation similar to EMA but with alpha = 1/period
	return g.generateRMA(varName, period, accessGen, needsNaN)
}

// RSIHandler generates inline code for Relative Strength Index calculations
type RSIHandler struct{}

func (h *RSIHandler) CanHandle(funcName string) bool {
	return funcName == "ta.rsi" || funcName == "rsi"
}

func (h *RSIHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	sourceExpr, period, err := extractTAArguments(g, call, "ta.rsi")
	if err != nil {
		return "", err
	}

	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify(sourceExpr)
	accessGen := CreateAccessGenerator(sourceInfo)
	needsNaN := sourceInfo.IsSeriesVariable()

	return g.generateRSI(varName, period, accessGen, needsNaN)
}

// ChangeHandler generates inline code for change calculations
type ChangeHandler struct{}

func (h *ChangeHandler) CanHandle(funcName string) bool {
	return funcName == "ta.change"
}

func (h *ChangeHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 {
		return "", fmt.Errorf("ta.change requires at least 1 argument")
	}

	sourceExpr := g.extractSeriesExpression(call.Arguments[0])

	// Default offset is 1 if not specified
	offset := 1
	if len(call.Arguments) >= 2 {
		offsetArg, ok := call.Arguments[1].(*ast.Literal)
		if !ok {
			return "", fmt.Errorf("ta.change offset must be literal")
		}
		var err error
		offset, err = extractPeriod(offsetArg)
		if err != nil {
			return "", fmt.Errorf("ta.change: %w", err)
		}
	}

	return g.generateChange(varName, sourceExpr, offset)
}

// PivotHighHandler generates inline code for pivot high detection
type PivotHighHandler struct{}

func (h *PivotHighHandler) CanHandle(funcName string) bool {
	return funcName == "ta.pivothigh"
}

func (h *PivotHighHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	return g.generatePivot(varName, call, true)
}

// PivotLowHandler generates inline code for pivot low detection
type PivotLowHandler struct{}

func (h *PivotLowHandler) CanHandle(funcName string) bool {
	return funcName == "ta.pivotlow"
}

func (h *PivotLowHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	return g.generatePivot(varName, call, false)
}

// CrossoverHandler generates inline code for crossover detection (series1 crosses above series2)
type CrossoverHandler struct{}

func (h *CrossoverHandler) CanHandle(funcName string) bool {
	return funcName == "ta.crossover"
}

func (h *CrossoverHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	return generateCrossDetection(g, varName, call, false)
}

// CrossunderHandler generates inline code for crossunder detection (series1 crosses below series2)
type CrossunderHandler struct{}

func (h *CrossunderHandler) CanHandle(funcName string) bool {
	return funcName == "ta.crossunder"
}

func (h *CrossunderHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	return generateCrossDetection(g, varName, call, true)
}

// Helper functions

// extractTAArguments extracts source and period from standard TA function arguments
func extractTAArguments(g *generator, call *ast.CallExpression, funcName string) (string, int, error) {
	if len(call.Arguments) < 2 {
		return "", 0, fmt.Errorf("%s requires at least 2 arguments", funcName)
	}

	sourceExpr := g.extractSeriesExpression(call.Arguments[0])

	periodArg, ok := call.Arguments[1].(*ast.Literal)
	if !ok {
		return "", 0, fmt.Errorf("%s period must be literal", funcName)
	}

	period, err := extractPeriod(periodArg)
	if err != nil {
		return "", 0, fmt.Errorf("%s: %w", funcName, err)
	}

	return sourceExpr, period, nil
}

// extractPeriod converts a literal to an integer period value
func extractPeriod(lit *ast.Literal) (int, error) {
	switch v := lit.Value.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	default:
		return 0, fmt.Errorf("period must be numeric, got %T", v)
	}
}

// generateCrossDetection generates code for crossover/crossunder detection
func generateCrossDetection(g *generator, varName string, call *ast.CallExpression, isCrossunder bool) (string, error) {
	if len(call.Arguments) < 2 {
		funcName := "ta.crossover"
		if isCrossunder {
			funcName = "ta.crossunder"
		}
		return "", fmt.Errorf("%s requires 2 arguments", funcName)
	}

	series1 := g.extractSeriesExpression(call.Arguments[0])
	series2 := g.extractSeriesExpression(call.Arguments[1])

	prev1Var := varName + "_prev1"
	prev2Var := varName + "_prev2"

	var code string
	var description string
	var condition string

	if isCrossunder {
		description = fmt.Sprintf("// Crossunder: %s crosses below %s\n", series1, series2)
		condition = fmt.Sprintf("if %s < %s && %s >= %s { return 1.0 } else { return 0.0 }", series1, series2, prev1Var, prev2Var)
	} else {
		description = fmt.Sprintf("// Crossover: %s crosses above %s\n", series1, series2)
		condition = fmt.Sprintf("if %s > %s && %s <= %s { return 1.0 } else { return 0.0 }", series1, series2, prev1Var, prev2Var)
	}

	code += g.ind() + description
	code += g.ind() + "if i > 0 {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%s := %s\n", prev1Var, g.convertSeriesAccessToPrev(series1))
	code += g.ind() + fmt.Sprintf("%s := %s\n", prev2Var, g.convertSeriesAccessToPrev(series2))
	code += g.ind() + fmt.Sprintf("%sSeries.Set(func() float64 { %s }())\n", varName, condition)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(0.0)\n", varName)
	g.indent--
	code += g.ind() + "}\n"

	return code, nil
}
