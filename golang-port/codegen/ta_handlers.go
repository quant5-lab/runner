package codegen

import (
	"fmt"
	"math"
	"strings"

	"github.com/quant5-lab/runner/ast"
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

	builder := NewTAIndicatorBuilder("ta.ema", varName, period, accessGen, needsNaN)
	return g.indentCode(builder.BuildEMA()), nil
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

	builder := NewTAIndicatorBuilder("ta.stdev", varName, period, accessGen, needsNaN)
	if builder.loopGen == nil {
		return "", fmt.Errorf("FATAL: loopGen is nil after NewTAIndicatorBuilder (period=%d, accessGen=%+v)", period, accessGen)
	}
	return g.indentCode(builder.BuildSTDEV()), nil
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
	return funcName == "ta.change" || funcName == "change"
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

// FixnanHandler generates inline code for forward-filling NaN values
type FixnanHandler struct{}

func (h *FixnanHandler) CanHandle(funcName string) bool {
	return funcName == "fixnan"
}

func (h *FixnanHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 {
		return "", fmt.Errorf("fixnan requires 1 argument")
	}

	var code string
	argExpr := call.Arguments[0]

	/* Handle nested function: fixnan(pivothigh()[1]) */
	if memberExpr, ok := argExpr.(*ast.MemberExpression); ok {
		if nestedCall, isCall := memberExpr.Object.(*ast.CallExpression); isCall {
			/* Generate intermediate variable for nested function */
			nestedFuncName := g.extractFunctionName(nestedCall.Callee)
			// Use funcName-based naming to match extractSeriesExpression
			tempVarName := strings.ReplaceAll(nestedFuncName, ".", "_")

			/* Generate nested function code */
			nestedCode, err := g.generateVariableFromCall(tempVarName, nestedCall)
			if err != nil {
				return "", fmt.Errorf("failed to generate nested function in fixnan: %w", err)
			}
			code += nestedCode
		}
	}

	sourceExpr := g.extractSeriesExpression(argExpr)
	stateVar := "fixnanState_" + varName

	code += g.ind() + fmt.Sprintf("if !math.IsNaN(%s) {\n", sourceExpr)
	g.indent++
	code += g.ind() + fmt.Sprintf("%s = %s\n", stateVar, sourceExpr)
	g.indent--
	code += g.ind() + "}\n"
	code += g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, stateVar)

	return code, nil
}

// Helper functions

// extractTAArguments extracts source and period from standard TA function arguments
// Supports: literals (14), variables (sr_len), expressions (round(sr_n / 2))
func extractTAArguments(g *generator, call *ast.CallExpression, funcName string) (string, int, error) {
	if len(call.Arguments) < 2 {
		return "", 0, fmt.Errorf("%s requires at least 2 arguments", funcName)
	}

	sourceExpr := g.extractSeriesExpression(call.Arguments[0])
	periodArg := call.Arguments[1]

	// Try literal period first (fast path)
	if periodLit, ok := periodArg.(*ast.Literal); ok {
		period, err := extractPeriod(periodLit)
		if err != nil {
			return "", 0, fmt.Errorf("%s: %w", funcName, err)
		}
		return sourceExpr, period, nil
	}

	// Try compile-time constant evaluation (handles variables + expressions)
	periodValue := g.constEvaluator.EvaluateConstant(periodArg)
	if !math.IsNaN(periodValue) && periodValue > 0 {
		return sourceExpr, int(periodValue), nil
	}

	return "", 0, fmt.Errorf("%s period must be compile-time constant (got %T that evaluates to NaN)", funcName, periodArg)
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

// WMAHandler generates inline code for Weighted Moving Average calculations
type WMAHandler struct{}

func (h *WMAHandler) CanHandle(funcName string) bool {
	return funcName == "ta.wma" || funcName == "wma"
}

func (h *WMAHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	sourceExpr, period, err := extractTAArguments(g, call, "ta.wma")
	if err != nil {
		return "", err
	}

	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify(sourceExpr)
	accessGen := CreateAccessGenerator(sourceInfo)
	needsNaN := sourceInfo.IsSeriesVariable()

	builder := NewTAIndicatorBuilder("ta.wma", varName, period, accessGen, needsNaN)
	builder.WithAccumulator(NewWeightedSumAccumulator(period))
	return g.indentCode(builder.Build()), nil
}

// DEVHandler generates inline code for Mean Absolute Deviation calculations
type DEVHandler struct{}

func (h *DEVHandler) CanHandle(funcName string) bool {
	return funcName == "ta.dev" || funcName == "dev"
}

func (h *DEVHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	sourceExpr, period, err := extractTAArguments(g, call, "ta.dev")
	if err != nil {
		return "", err
	}

	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify(sourceExpr)
	accessGen := CreateAccessGenerator(sourceInfo)
	needsNaN := sourceInfo.IsSeriesVariable()

	builder := NewTAIndicatorBuilder("ta.dev", varName, period, accessGen, needsNaN)
	return g.indentCode(builder.BuildDEV()), nil
}

type SumHandler struct{}

func (h *SumHandler) CanHandle(funcName string) bool {
	return funcName == "sum" || funcName == "math.sum"
}

func (h *SumHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 2 {
		return "", fmt.Errorf("sum requires 2 arguments")
	}

	// Check if source is ConditionalExpression - needs temp var for accessor pattern
	var code string
	sourceArg := call.Arguments[0]
	var sourceExpr string

	if condExpr, ok := sourceArg.(*ast.ConditionalExpression); ok {
		// Create temp var for ternary expression
		tempVarName := g.tempVarMgr.GetOrCreate(CallInfo{
			FuncName: "ternary",
			Call:     call,
			ArgHash:  fmt.Sprintf("%p", condExpr),
		})

		// Generate ternary as temp var
		condCode, err := g.generateConditionExpression(condExpr.Test)
		if err != nil {
			return "", err
		}
		condCode = g.addBoolConversionIfNeeded(condExpr.Test, condCode)

		consequentCode, err := g.generateNumericExpression(condExpr.Consequent)
		if err != nil {
			return "", err
		}
		alternateCode, err := g.generateNumericExpression(condExpr.Alternate)
		if err != nil {
			return "", err
		}

		code += g.ind() + fmt.Sprintf("%sSeries.Set(func() float64 { if %s { return %s } else { return %s } }())\n",
			tempVarName, condCode, consequentCode, alternateCode)

		sourceExpr = tempVarName + "Series.GetCurrent()"
	} else {
		extracted, _, extractErr := extractTAArguments(g, call, "sum")
		if extractErr != nil {
			return "", extractErr
		}
		sourceExpr = extracted
	}

	// Extract period
	periodArg := call.Arguments[1]
	var period int
	if periodLit, ok := periodArg.(*ast.Literal); ok {
		p, err := extractPeriod(periodLit)
		if err != nil {
			return "", fmt.Errorf("sum: %w", err)
		}
		period = p
	} else {
		periodValue := g.constEvaluator.EvaluateConstant(periodArg)
		if math.IsNaN(periodValue) || periodValue <= 0 {
			return "", fmt.Errorf("sum period must be compile-time constant")
		}
		period = int(periodValue)
	}

	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify(sourceExpr)
	accessGen := CreateAccessGenerator(sourceInfo)
	needsNaN := sourceInfo.IsSeriesVariable()

	builder := NewTAIndicatorBuilder("sum", varName, period, accessGen, needsNaN)
	builder.WithAccumulator(NewSumAccumulator())
	sumCode := g.indentCode(builder.Build())

	return code + sumCode, nil
}

type ValuewhenHandler struct{}

func (h *ValuewhenHandler) CanHandle(funcName string) bool {
	return funcName == "ta.valuewhen" || funcName == "valuewhen"
}

func (h *ValuewhenHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 3 {
		return "", fmt.Errorf("valuewhen requires 3 arguments (condition, source, occurrence)")
	}

	conditionExpr := g.extractSeriesExpression(call.Arguments[0])
	sourceExpr := g.extractSeriesExpression(call.Arguments[1])

	occurrenceArg, ok := call.Arguments[2].(*ast.Literal)
	if !ok {
		return "", fmt.Errorf("valuewhen occurrence must be literal")
	}

	occurrence, err := extractPeriod(occurrenceArg)
	if err != nil {
		return "", fmt.Errorf("valuewhen: %w", err)
	}

	return g.generateValuewhen(varName, conditionExpr, sourceExpr, occurrence)
}
