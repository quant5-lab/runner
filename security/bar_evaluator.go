package security

import (
	"fmt"
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/series"
)

type BarEvaluator interface {
	EvaluateAtBar(expr ast.Expression, secCtx *context.Context, barIdx int) (float64, error)
}

/*
VarLookupFunc resolves a variable name to its Series from the main context.

	Returns (series, mainBarIndex, true) if variable exists, or (nil, -1, false) if not found.
	The mainBarIndex maps the security bar index to the corresponding main context bar index.
*/
type VarLookupFunc func(varName string, secBarIdx int) (*series.Series, int, bool)

type StreamingBarEvaluator struct {
	taStateCache      map[string]TAStateManager
	fixnanEvaluator   *FixnanEvaluator
	varRegistry       *VariableRegistry
	secBarMapper      *BarIndexMapper
	varLookup         VarLookupFunc
	inputConstantsMap map[string]float64 // input() constants for extractNumberLiteral
}

func NewStreamingBarEvaluator() *StreamingBarEvaluator {
	return &StreamingBarEvaluator{
		taStateCache: make(map[string]TAStateManager),
		fixnanEvaluator: NewFixnanEvaluator(
			NewMapStateStorage(),
			NewSequentialWarmupStrategy(),
			NewHashExpressionIdentifier(),
		),
		varRegistry:       NewVariableRegistry(),
		secBarMapper:      nil,
		varLookup:         nil,
		inputConstantsMap: nil,
	}
}

func (e *StreamingBarEvaluator) SetVariableRegistry(registry *VariableRegistry) {
	e.varRegistry = registry
}

func (e *StreamingBarEvaluator) SetBarIndexMapper(mapper *BarIndexMapper) {
	e.secBarMapper = mapper
}

func (e *StreamingBarEvaluator) SetVarLookup(lookup VarLookupFunc) {
	e.varLookup = lookup
}

func (e *StreamingBarEvaluator) SetInputConstantsMap(inputConstants map[string]float64) {
	e.inputConstantsMap = inputConstants
}

func (e *StreamingBarEvaluator) UpdateBarMapping(secBarIdx, mainBarIdx int) {
	if e.secBarMapper != nil {
		e.secBarMapper.SetMapping(secBarIdx, mainBarIdx)
	}
}

func (e *StreamingBarEvaluator) EvaluateAtBar(expr ast.Expression, secCtx *context.Context, barIdx int) (float64, error) {
	switch exp := expr.(type) {
	case *ast.Identifier:
		return e.evaluateIdentifierAtBar(exp, secCtx, barIdx)
	case *ast.CallExpression:
		return e.evaluateTACallAtBar(exp, secCtx, barIdx)
	case *ast.BinaryExpression:
		return e.evaluateBinaryExpressionAtBar(exp, secCtx, barIdx)
	case *ast.ConditionalExpression:
		return e.evaluateConditionalExpressionAtBar(exp, secCtx, barIdx)
	case *ast.Literal:
		if val, ok := exp.Value.(float64); ok {
			return val, nil
		}
		return 0.0, newUnsupportedExpressionError(exp)
	case *ast.MemberExpression:
		return e.evaluateMemberExpressionAtBar(exp, secCtx, barIdx)
	default:
		return 0.0, newUnsupportedExpressionError(exp)
	}
}

func (e *StreamingBarEvaluator) evaluateIdentifierAtBar(id *ast.Identifier, secCtx *context.Context, barIdx int) (float64, error) {
	if val, err := evaluateOHLCVAtBar(id, secCtx, barIdx); err == nil || !isUnknownIdentifierError(err) {
		return val, err
	}

	/* Handle bar_index builtin - returns security context bar index */
	if id.Name == "bar_index" {
		return float64(barIdx), nil
	}

	/* Check input constants first (compile-time constants from input()) */
	if e.inputConstantsMap != nil {
		if val, ok := e.inputConstantsMap[id.Name]; ok {
			return val, nil
		}
	}

	if secCtx != nil {
		result := secCtx.ResolveVariable(id.Name)
		if result.Found {
			if result.SourceBarIdx < 0 {
				return math.NaN(), nil
			}
			offset := result.Series.Position() - result.SourceBarIdx
			if offset >= 0 && offset < result.Series.Capacity() {
				return result.Series.Get(offset), nil
			}
		}
	}

	/* Try variable registry first (for security-context variables) */
	if e.varRegistry != nil {
		if varSeries, ok := e.varRegistry.Get(id.Name); ok {
			if e.secBarMapper != nil {
				mainIdx := e.secBarMapper.GetMainBarIndexForSecurityBar(barIdx)
				if mainIdx >= 0 {
					offset := varSeries.Position() - mainIdx
					if offset >= 0 && offset < varSeries.Capacity() {
						return varSeries.Get(offset), nil
					}
				}
				/* Warmup period: security bar has no corresponding main bar yet */
				if mainIdx < 0 {
					return math.NaN(), nil
				}
			}
		}
	}

	/* Fallback to main context lookup (PineScript lexical scoping) */
	if e.varLookup != nil {
		if varSeries, mainIdx, ok := e.varLookup(id.Name, barIdx); ok {
			if varSeries == nil {
				return 0.0, newUnknownIdentifierError(id.Name)
			}
			if mainIdx >= 0 {
				offset := varSeries.Position() - mainIdx
				if offset >= 0 && offset < varSeries.Capacity() {
					return varSeries.Get(offset), nil
				}
			}
			/* Warmup period: security bar has no corresponding main bar yet */
			if mainIdx < 0 {
				return math.NaN(), nil
			}
		}
	}

	return 0.0, newUnknownIdentifierError(id.Name)
}

func evaluateOHLCVAtBar(id *ast.Identifier, secCtx *context.Context, barIdx int) (float64, error) {
	if barIdx < 0 || barIdx >= len(secCtx.Data) {
		return 0.0, newBarIndexOutOfRangeError(barIdx, len(secCtx.Data))
	}

	bar := secCtx.Data[barIdx]

	switch id.Name {
	case "close":
		return bar.Close, nil
	case "open":
		return bar.Open, nil
	case "high":
		return bar.High, nil
	case "low":
		return bar.Low, nil
	case "volume":
		return bar.Volume, nil
	case "ohlc4":
		return (bar.Open + bar.High + bar.Low + bar.Close) / 4, nil
	case "hlc3":
		return (bar.High + bar.Low + bar.Close) / 3, nil
	case "hl2":
		return (bar.High + bar.Low) / 2, nil
	case "hlcc4":
		return (bar.High + bar.Low + bar.Close + bar.Close) / 4, nil
	default:
		return 0.0, newUnknownIdentifierError(id.Name)
	}
}

func (e *StreamingBarEvaluator) evaluateTACallAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	funcName := extractCallFunctionName(call.Callee)

	switch funcName {
	case "ta.sma":
		return e.evaluateSMAAtBar(call, secCtx, barIdx)
	case "ta.ema":
		return e.evaluateEMAAtBar(call, secCtx, barIdx)
	case "ta.rma":
		return e.evaluateRMAAtBar(call, secCtx, barIdx)
	case "ta.rsi":
		return e.evaluateRSIAtBar(call, secCtx, barIdx)
	case "ta.atr":
		return e.evaluateATRAtBar(call, secCtx, barIdx)
	case "ta.stdev":
		return e.evaluateSTDEVAtBar(call, secCtx, barIdx)
	case "ta.swma":
		return e.evaluateSWMAAtBar(call, secCtx, barIdx)
	case "ta.cci":
		return e.evaluateCCIAtBar(call, secCtx, barIdx)
	case "ta.bbw":
		return e.evaluateBBWAtBar(call, secCtx, barIdx)
	case "ta.cog":
		return e.evaluateCOGAtBar(call, secCtx, barIdx)
	case "ta.tsi":
		return e.evaluateTSIAtBar(call, secCtx, barIdx)
	case "ta.pivothigh":
		return e.evaluatePivotHighAtBar(call, secCtx, barIdx)
	case "ta.pivotlow":
		return e.evaluatePivotLowAtBar(call, secCtx, barIdx)
	case "ta.valuewhen", "valuewhen":
		return e.evaluateValuewhenAtBar(call, secCtx, barIdx)
	case "fixnan", "ta.fixnan":
		return e.fixnanEvaluator.EvaluateAtBar(e, call, secCtx, barIdx)
	default:
		return 0.0, newUnsupportedFunctionError(funcName)
	}
}

func (e *StreamingBarEvaluator) evaluateSMAAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}

	cacheKey := buildTACacheKey("sma", sourceID.Name, period)
	stateManager := e.getOrCreateTAState(cacheKey, period, secCtx)

	return stateManager.ComputeAtBar(secCtx, sourceID, barIdx)
}

func (e *StreamingBarEvaluator) evaluateEMAAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}

	cacheKey := buildTACacheKey("ema", sourceID.Name, period)
	stateManager := e.getOrCreateTAState(cacheKey, period, secCtx)

	return stateManager.ComputeAtBar(secCtx, sourceID, barIdx)
}

func (e *StreamingBarEvaluator) evaluateRMAAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}

	cacheKey := buildTACacheKey("rma", sourceID.Name, period)
	stateManager := e.getOrCreateTAState(cacheKey, period, secCtx)

	return stateManager.ComputeAtBar(secCtx, sourceID, barIdx)
}

func (e *StreamingBarEvaluator) evaluateRSIAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}

	cacheKey := buildTACacheKey("rsi", sourceID.Name, period)
	stateManager := e.getOrCreateTAState(cacheKey, period, secCtx)

	return stateManager.ComputeAtBar(secCtx, sourceID, barIdx)
}

func (e *StreamingBarEvaluator) evaluateATRAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	period, err := extractPeriodArgument(call, "atr")
	if err != nil {
		return 0.0, err
	}

	cacheKey := buildTACacheKey("atr", "hlc", period)
	stateManager := e.getOrCreateTAState(cacheKey, period, secCtx)

	dummyID := &ast.Identifier{Name: "close"}
	return stateManager.ComputeAtBar(secCtx, dummyID, barIdx)
}

func (e *StreamingBarEvaluator) evaluateSTDEVAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}

	cacheKey := buildTACacheKey("stdev", sourceID.Name, period)
	stateManager := e.getOrCreateTAState(cacheKey, period, secCtx)

	return stateManager.ComputeAtBar(secCtx, sourceID, barIdx)
}

func (e *StreamingBarEvaluator) evaluateSWMAAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, err := extractSourceOnlyArgument(call, "swma")
	if err != nil {
		return 0.0, err
	}
	if barIdx < 3 {
		return math.NaN(), nil
	}
	weights := [4]float64{1.0 / 6.0, 2.0 / 6.0, 2.0 / 6.0, 1.0 / 6.0}
	result := 0.0
	for i := 0; i < 4; i++ {
		v, err := evaluateOHLCVAtBar(sourceID, secCtx, barIdx-3+i)
		if err != nil {
			return math.NaN(), err
		}
		result += v * weights[i]
	}
	return result, nil
}

func (e *StreamingBarEvaluator) evaluateCCIAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < period-1 {
		return math.NaN(), nil
	}
	sma := 0.0
	for i := 0; i < period; i++ {
		v, err := evaluateOHLCVAtBar(sourceID, secCtx, barIdx-period+1+i)
		if err != nil {
			return math.NaN(), err
		}
		sma += v
	}
	sma /= float64(period)
	dev := 0.0
	for i := 0; i < period; i++ {
		v, err := evaluateOHLCVAtBar(sourceID, secCtx, barIdx-period+1+i)
		if err != nil {
			return math.NaN(), err
		}
		d := v - sma
		if d < 0 {
			d = -d
		}
		dev += d
	}
	dev /= float64(period)
	if dev == 0.0 {
		return 0.0, nil
	}
	cur, err := evaluateOHLCVAtBar(sourceID, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	return (cur - sma) / (0.015 * dev), nil
}

func (e *StreamingBarEvaluator) evaluateBBWAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	mult := 2.0
	if len(call.Arguments) >= 3 {
		if v, err2 := extractNumberLiteral(call.Arguments[2]); err2 == nil {
			mult = v
		}
	}
	if barIdx < period-1 {
		return math.NaN(), nil
	}
	sma := 0.0
	for i := 0; i < period; i++ {
		v, err := evaluateOHLCVAtBar(sourceID, secCtx, barIdx-period+1+i)
		if err != nil {
			return math.NaN(), err
		}
		sma += v
	}
	sma /= float64(period)
	sd := 0.0
	for i := 0; i < period; i++ {
		v, err := evaluateOHLCVAtBar(sourceID, secCtx, barIdx-period+1+i)
		if err != nil {
			return math.NaN(), err
		}
		d := v - sma
		sd += d * d
	}
	sd = math.Sqrt(sd / float64(period))
	if sma == 0.0 {
		return 0.0, nil
	}
	return 2.0 * mult * sd / sma, nil
}

func (e *StreamingBarEvaluator) evaluateCOGAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < period-1 {
		return math.NaN(), nil
	}
	num, den := 0.0, 0.0
	for i := 0; i < period; i++ {
		v, err := evaluateOHLCVAtBar(sourceID, secCtx, barIdx-i)
		if err != nil {
			return math.NaN(), err
		}
		num += v * float64(i+1)
		den += v
	}
	if den == 0.0 {
		return 0.0, nil
	}
	return -num / den, nil
}

func (e *StreamingBarEvaluator) evaluateTSIAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	if len(call.Arguments) < 3 {
		return 0.0, newInsufficientArgumentsError("tsi", 3, len(call.Arguments))
	}
	sourceID, ok := call.Arguments[0].(*ast.Identifier)
	if !ok {
		return 0.0, newInvalidArgumentTypeError("tsi", 0, "identifier")
	}
	shortLength, err := extractNumberLiteral(call.Arguments[1])
	if err != nil {
		return 0.0, err
	}
	longLength, err := extractNumberLiteral(call.Arguments[2])
	if err != nil {
		return 0.0, err
	}

	cacheKey := fmt.Sprintf("tsi_%s_%d_%d", sourceID.Name, int(shortLength), int(longLength))
	if state, exists := e.taStateCache[cacheKey]; exists {
		return state.ComputeAtBar(secCtx, sourceID, barIdx)
	}
	state := NewTSIStateManager(cacheKey, int(shortLength), int(longLength))
	e.taStateCache[cacheKey] = state
	return state.ComputeAtBar(secCtx, sourceID, barIdx)
}

func (e *StreamingBarEvaluator) getOrCreateTAState(cacheKey string, period int, secCtx *context.Context) TAStateManager {
	if state, exists := e.taStateCache[cacheKey]; exists {
		return state
	}

	state := NewTAStateManager(cacheKey, period, len(secCtx.Data))
	e.taStateCache[cacheKey] = state
	return state
}

func (e *StreamingBarEvaluator) evaluateBinaryExpressionAtBar(expr *ast.BinaryExpression, secCtx *context.Context, barIdx int) (float64, error) {
	leftValue, err := e.EvaluateAtBar(expr.Left, secCtx, barIdx)
	if err != nil {
		return 0.0, err
	}

	rightValue, err := e.EvaluateAtBar(expr.Right, secCtx, barIdx)
	if err != nil {
		return 0.0, err
	}

	return applyBinaryOperator(expr.Operator, leftValue, rightValue)
}

func (e *StreamingBarEvaluator) evaluateConditionalExpressionAtBar(expr *ast.ConditionalExpression, secCtx *context.Context, barIdx int) (float64, error) {
	testValue, err := e.EvaluateAtBar(expr.Test, secCtx, barIdx)
	if err != nil {
		return 0.0, err
	}

	if testValue != 0.0 {
		return e.EvaluateAtBar(expr.Consequent, secCtx, barIdx)
	}
	return e.EvaluateAtBar(expr.Alternate, secCtx, barIdx)
}

func (e *StreamingBarEvaluator) evaluatePivotHighAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, leftBars, rightBars, err := extractPivotArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}

	evaluator := NewDelayedPivotHighEvaluator(leftBars, rightBars)
	return evaluator.EvaluateAtBar(secCtx.Data, sourceID.Name, barIdx), nil
}

func (e *StreamingBarEvaluator) evaluatePivotLowAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, leftBars, rightBars, err := extractPivotArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}

	evaluator := NewDelayedPivotLowEvaluator(leftBars, rightBars)
	return evaluator.EvaluateAtBar(secCtx.Data, sourceID.Name, barIdx), nil
}

func (e *StreamingBarEvaluator) evaluateValuewhenAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	conditionExpr, sourceExpr, occurrence, err := extractValuewhenArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}

	occurrenceCount := 0
	for lookbackOffset := 0; lookbackOffset <= barIdx; lookbackOffset++ {
		lookbackBarIdx := barIdx - lookbackOffset

		conditionValue, err := e.EvaluateAtBar(conditionExpr, secCtx, lookbackBarIdx)
		if err != nil {
			return 0.0, err
		}

		if conditionValue != 0.0 {
			if occurrenceCount == occurrence {
				return e.EvaluateAtBar(sourceExpr, secCtx, lookbackBarIdx)
			}
			occurrenceCount++
		}
	}

	return math.NaN(), nil
}

func (e *StreamingBarEvaluator) evaluateMemberExpressionAtBar(expr *ast.MemberExpression, secCtx *context.Context, barIdx int) (float64, error) {
	if propID, ok := expr.Property.(*ast.Identifier); ok {
		if objID, ok := expr.Object.(*ast.Identifier); ok && objID.Name == "ta" && propID.Name == "tr" {
			return e.evaluateTrueRangeAtBar(secCtx, barIdx)
		}
		return 0.0, newUnsupportedExpressionError(expr)
	}

	propertyLit, ok := expr.Property.(*ast.Literal)
	if !ok {
		return 0.0, newUnsupportedExpressionError(expr)
	}

	offset, ok := propertyLit.Value.(float64)
	if !ok {
		return 0.0, newUnsupportedExpressionError(expr)
	}

	targetIdx := barIdx - int(offset)
	if targetIdx < 0 || targetIdx >= len(secCtx.Data) {
		return 0.0, newBarIndexOutOfRangeError(targetIdx, len(secCtx.Data))
	}

	switch obj := expr.Object.(type) {
	case *ast.Identifier:
		return evaluateOHLCVAtBar(obj, secCtx, targetIdx)
	case *ast.CallExpression:
		return e.evaluateTACallAtBar(obj, secCtx, targetIdx)
	default:
		return 0.0, newUnsupportedExpressionError(expr)
	}
}

func (e *StreamingBarEvaluator) evaluateTrueRangeAtBar(secCtx *context.Context, barIdx int) (float64, error) {
	if barIdx < 0 || barIdx >= len(secCtx.Data) {
		return 0.0, newBarIndexOutOfRangeError(barIdx, len(secCtx.Data))
	}

	isFirstBar := barIdx == 0

	var prevClose float64
	if !isFirstBar {
		prevClose = secCtx.Data[barIdx-1].Close
	}

	trCalculator := NewTrueRangeCalculator()
	return trCalculator.CalculateAtBar(secCtx.Data, barIdx, prevClose, isFirstBar), nil
}
