package security

import (
	"fmt"
	"math"
	"sort"

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
	volumeStateCache  map[string]*volumeIndicatorState
	barsSinceCache    map[*ast.CallExpression]*BarsSinceStateManager
	valuewhenCache    map[*ast.CallExpression]*ValuewhenStateManager
	fixnanEvaluator   *FixnanEvaluator
	varRegistry       *VariableRegistry
	secBarMapper      *BarIndexMapper
	varLookup         VarLookupFunc
	inputConstantsMap map[string]float64
}

func NewStreamingBarEvaluator() *StreamingBarEvaluator {
	return &StreamingBarEvaluator{
		taStateCache:     make(map[string]TAStateManager),
		volumeStateCache: make(map[string]*volumeIndicatorState),
		barsSinceCache:   make(map[*ast.CallExpression]*BarsSinceStateManager),
		valuewhenCache:   make(map[*ast.CallExpression]*ValuewhenStateManager),
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

	if id.Name == "bar_index" {
		return float64(barIdx), nil
	}

	if id.Name == "na" {
		return math.NaN(), nil
	}

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

func (e *StreamingBarEvaluator) evaluateSMAAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}

	cacheKey := buildTACacheKey("sma", expressionKey(sourceExpr), period)
	stateManager := e.getOrCreateTAState(cacheKey, period, secCtx)

	return stateManager.ComputeAtBar(secCtx, sourceExpr, barIdx)
}

func (e *StreamingBarEvaluator) evaluateEMAAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}

	cacheKey := buildTACacheKey("ema", expressionKey(sourceExpr), period)
	stateManager := e.getOrCreateTAState(cacheKey, period, secCtx)

	return stateManager.ComputeAtBar(secCtx, sourceExpr, barIdx)
}

func (e *StreamingBarEvaluator) evaluateRMAAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}

	cacheKey := buildTACacheKey("rma", expressionKey(sourceExpr), period)
	stateManager := e.getOrCreateTAState(cacheKey, period, secCtx)

	return stateManager.ComputeAtBar(secCtx, sourceExpr, barIdx)
}

func (e *StreamingBarEvaluator) evaluateRSIAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}

	cacheKey := buildTACacheKey("rsi", expressionKey(sourceExpr), period)
	stateManager := e.getOrCreateTAState(cacheKey, period, secCtx)

	return stateManager.ComputeAtBar(secCtx, sourceExpr, barIdx)
}

func (e *StreamingBarEvaluator) evaluateATRAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	period, err := extractPeriodArgument(call, "atr")
	if err != nil {
		return 0.0, err
	}

	cacheKey := buildTACacheKey("atr", "hlc", period)
	stateManager := e.getOrCreateTAState(cacheKey, period, secCtx)

	return stateManager.ComputeAtBar(secCtx, nil, barIdx)
}

func (e *StreamingBarEvaluator) evaluateSTDEVAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}

	cacheKey := buildTACacheKey("stdev", expressionKey(sourceExpr), period)
	stateManager := e.getOrCreateTAState(cacheKey, period, secCtx)

	return stateManager.ComputeAtBar(secCtx, sourceExpr, barIdx)
}

func (e *StreamingBarEvaluator) evaluateSWMAAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, err := extractSourceOnlyArgument(call, "swma")
	if err != nil {
		return 0.0, err
	}
	if barIdx < 3 {
		return math.NaN(), nil
	}
	weights := [4]float64{1.0 / 6.0, 2.0 / 6.0, 2.0 / 6.0, 1.0 / 6.0}
	result := 0.0
	for i := 0; i < 4; i++ {
		v, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx-3+i)
		if err != nil {
			return math.NaN(), err
		}
		result += v * weights[i]
	}
	return result, nil
}

func (e *StreamingBarEvaluator) evaluateCCIAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < period-1 {
		return math.NaN(), nil
	}
	sma := 0.0
	for i := 0; i < period; i++ {
		v, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx-period+1+i)
		if err != nil {
			return math.NaN(), err
		}
		sma += v
	}
	sma /= float64(period)
	dev := 0.0
	for i := 0; i < period; i++ {
		v, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx-period+1+i)
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
	cur, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	return (cur - sma) / (0.015 * dev), nil
}

func (e *StreamingBarEvaluator) evaluateBBWAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, period, err := extractTAArguments(call, e.inputConstantsMap)
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
		v, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx-period+1+i)
		if err != nil {
			return math.NaN(), err
		}
		sma += v
	}
	sma /= float64(period)
	sd := 0.0
	for i := 0; i < period; i++ {
		v, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx-period+1+i)
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
	sourceExpr, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < period-1 {
		return math.NaN(), nil
	}
	num, den := 0.0, 0.0
	for i := 0; i < period; i++ {
		v, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx-i)
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
	sourceExpr := call.Arguments[0]
	shortLength, err := extractNumberLiteral(call.Arguments[1])
	if err != nil {
		return 0.0, err
	}
	longLength, err := extractNumberLiteral(call.Arguments[2])
	if err != nil {
		return 0.0, err
	}

	cacheKey := fmt.Sprintf("tsi_%s_%d_%d", expressionKey(sourceExpr), int(shortLength), int(longLength))
	if state, exists := e.taStateCache[cacheKey]; exists {
		return state.ComputeAtBar(secCtx, sourceExpr, barIdx)
	}
	state := NewTSIStateManager(cacheKey, int(shortLength), int(longLength), len(secCtx.Data), e)
	e.taStateCache[cacheKey] = state
	return state.ComputeAtBar(secCtx, sourceExpr, barIdx)
}

func (e *StreamingBarEvaluator) getOrCreateTAState(cacheKey string, period int, secCtx *context.Context) TAStateManager {
	if state, exists := e.taStateCache[cacheKey]; exists {
		return state
	}

	state := NewTAStateManager(cacheKey, period, len(secCtx.Data), e)
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

	state, cached := e.valuewhenCache[call]
	if !cached {
		cacheKey := buildValuewhenCacheKey(conditionExpr, sourceExpr, occurrence)
		state = NewValuewhenStateManager(cacheKey, occurrence, conditionExpr, sourceExpr, len(secCtx.Data), e)
		e.valuewhenCache[call] = state
	}

	return state.ComputeAtBar(secCtx, nil, barIdx)
}

func (e *StreamingBarEvaluator) evaluateMemberExpressionAtBar(expr *ast.MemberExpression, secCtx *context.Context, barIdx int) (float64, error) {
	if propID, ok := expr.Property.(*ast.Identifier); ok {
		if objID, ok := expr.Object.(*ast.Identifier); ok {
			if objID.Name == "ta" {
				if propID.Name == "tr" {
					return e.evaluateTrueRangeAtBar(secCtx, barIdx)
				}
				if _, known := volumeIndicatorFactories[propID.Name]; known {
					return e.evaluateVolumeIndicatorAtBar(propID.Name, secCtx, barIdx)
				}
			}
			if val, ok := lookupMemberConstant(objID.Name, propID.Name); ok {
				return val, nil
			}
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
	case *ast.MemberExpression:
		if innerProp, ok := obj.Property.(*ast.Identifier); ok {
			if innerObj, ok := obj.Object.(*ast.Identifier); ok && innerObj.Name == "ta" {
				if _, known := volumeIndicatorFactories[innerProp.Name]; known {
					return e.evaluateVolumeIndicatorAtBar(innerProp.Name, secCtx, targetIdx)
				}
			}
		}
		return 0.0, newUnsupportedExpressionError(expr)
	default:
		return 0.0, newUnsupportedExpressionError(expr)
	}
}

func (e *StreamingBarEvaluator) evaluateVolumeIndicatorAtBar(propName string, secCtx *context.Context, barIdx int) (float64, error) {
	state, cached := e.volumeStateCache[propName]
	if !cached {
		var err error
		state, err = newVolumeState(propName, len(secCtx.Data))
		if err != nil {
			return 0.0, err
		}
		e.volumeStateCache[propName] = state
	}
	return state.computeAtBar(secCtx, barIdx)
}

func (e *StreamingBarEvaluator) evaluateTrueRangeAtBar(secCtx *context.Context, barIdx int) (float64, error) {
	return e.computeTrueRangeAtBar(secCtx, barIdx, false)
}

func (e *StreamingBarEvaluator) evaluateTRFuncAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	return e.computeTrueRangeAtBar(secCtx, barIdx, extractTRHandleNAArg(call))
}

func (e *StreamingBarEvaluator) computeTrueRangeAtBar(secCtx *context.Context, barIdx int, handleNA bool) (float64, error) {
	if barIdx < 0 || barIdx >= len(secCtx.Data) {
		return 0.0, newBarIndexOutOfRangeError(barIdx, len(secCtx.Data))
	}

	isFirstBar := barIdx == 0

	var prevClose float64
	if !isFirstBar {
		prevClose = secCtx.Data[barIdx-1].Close
	}

	return NewTrueRangeCalculator().CalculateAtBar(secCtx.Data, barIdx, prevClose, isFirstBar, handleNA), nil
}

func extractTRHandleNAArg(call *ast.CallExpression) bool {
	if len(call.Arguments) < 1 {
		return false
	}
	lit, ok := call.Arguments[0].(*ast.Literal)
	if !ok {
		return false
	}
	boolVal, ok := lit.Value.(bool)
	if !ok {
		return false
	}
	return boolVal
}

func (e *StreamingBarEvaluator) evaluateWMAAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < period-1 {
		return math.NaN(), nil
	}
	sum, weightSum := 0.0, 0.0
	for i := 0; i < period; i++ {
		v, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx-i)
		if err != nil {
			return math.NaN(), err
		}
		weight := float64(period - i)
		sum += weight * v
		weightSum += weight
	}
	return sum / weightSum, nil
}

func (e *StreamingBarEvaluator) evaluatePercentrankAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < period-1 {
		return math.NaN(), nil
	}
	current, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	count := 0
	for i := 0; i < period; i++ {
		v, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx-i)
		if err != nil {
			return math.NaN(), err
		}
		if v < current {
			count++
		}
	}
	return float64(count) / float64(period) * 100.0, nil
}

func (e *StreamingBarEvaluator) evaluatePercentileNearestRankAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	pct := 50.0
	if len(call.Arguments) >= 3 {
		if v, err2 := extractNumberLiteral(call.Arguments[2]); err2 == nil {
			pct = v
		}
	}
	if barIdx < period-1 {
		return math.NaN(), nil
	}
	window := make([]float64, period)
	for i := 0; i < period; i++ {
		v, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx-i)
		if err != nil {
			return math.NaN(), err
		}
		window[i] = v
	}
	sortedWindow := make([]float64, period)
	copy(sortedWindow, window)
	sort.Float64s(sortedWindow)
	idx := int(math.Ceil(pct/100.0*float64(period))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= period {
		idx = period - 1
	}
	return sortedWindow[idx], nil
}

func (e *StreamingBarEvaluator) evaluatePercentileLinearInterpolationAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	pct := 50.0
	if len(call.Arguments) >= 3 {
		if v, err2 := extractNumberLiteral(call.Arguments[2]); err2 == nil {
			pct = v
		}
	}
	if barIdx < period-1 {
		return math.NaN(), nil
	}
	window := make([]float64, period)
	for i := 0; i < period; i++ {
		v, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx-i)
		if err != nil {
			return math.NaN(), err
		}
		window[i] = v
	}
	sortedWindow := make([]float64, period)
	copy(sortedWindow, window)
	sort.Float64s(sortedWindow)
	rank := pct / 100.0 * float64(period-1)
	lower := int(math.Floor(rank))
	upper := lower + 1
	frac := rank - float64(lower)
	if upper >= period {
		return sortedWindow[period-1], nil
	}
	return sortedWindow[lower] + frac*(sortedWindow[upper]-sortedWindow[lower]), nil
}

func (e *StreamingBarEvaluator) evaluateCorrelationAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	if len(call.Arguments) < 3 {
		return 0.0, newInsufficientArgumentsError("correlation", 3, len(call.Arguments))
	}
	src1Expr := call.Arguments[0]
	src2Expr := call.Arguments[1]
	period, err := extractNumberLiteral(call.Arguments[2], e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	p := int(period)
	if barIdx < p-1 {
		return math.NaN(), nil
	}
	sum1, sum2 := 0.0, 0.0
	for i := 0; i < p; i++ {
		v1, err := e.EvaluateAtBar(src1Expr, secCtx, barIdx-i)
		if err != nil {
			return math.NaN(), err
		}
		v2, err := e.EvaluateAtBar(src2Expr, secCtx, barIdx-i)
		if err != nil {
			return math.NaN(), err
		}
		sum1 += v1
		sum2 += v2
	}
	m1, m2 := sum1/float64(p), sum2/float64(p)
	cov, var1, var2 := 0.0, 0.0, 0.0
	for i := 0; i < p; i++ {
		v1, _ := e.EvaluateAtBar(src1Expr, secCtx, barIdx-i)
		v2, _ := e.EvaluateAtBar(src2Expr, secCtx, barIdx-i)
		d1, d2 := v1-m1, v2-m2
		cov += d1 * d2
		var1 += d1 * d1
		var2 += d2 * d2
	}
	if var1 == 0 || var2 == 0 {
		return 0.0, nil
	}
	return cov / math.Sqrt(var1*var2), nil
}

func (e *StreamingBarEvaluator) evaluateALMAAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	offset := 0.85
	sigma := 6.0
	if len(call.Arguments) >= 3 {
		if v, err2 := extractNumberLiteral(call.Arguments[2]); err2 == nil {
			offset = v
		}
	}
	if len(call.Arguments) >= 4 {
		if v, err2 := extractNumberLiteral(call.Arguments[3]); err2 == nil {
			sigma = v
		}
	}
	if barIdx < period-1 {
		return math.NaN(), nil
	}
	m := offset * float64(period-1)
	s := float64(period) / sigma
	weights := make([]float64, period)
	wSum := 0.0
	for j := 0; j < period; j++ {
		d := float64(j) - m
		weights[j] = math.Exp(-(d * d) / (2 * s * s))
		wSum += weights[j]
	}
	val := 0.0
	for j := 0; j < period; j++ {
		v, err := e.EvaluateAtBar(sourceExpr, secCtx, barIdx-period+1+j)
		if err != nil {
			return math.NaN(), err
		}
		val += weights[j] * v
	}
	return val / wSum, nil
}

func (e *StreamingBarEvaluator) evaluateHMAAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	halfPeriod := period / 2
	sqrtPeriod := int(math.Round(math.Sqrt(float64(period))))
	totalWarmup := period + sqrtPeriod - 1
	if barIdx < totalWarmup-1 {
		return math.NaN(), nil
	}

	wmaAt := func(src ast.Expression, idx, p int) (float64, error) {
		if idx < p-1 {
			return math.NaN(), nil
		}
		s, ws := 0.0, 0.0
		for i := 0; i < p; i++ {
			v, err := e.EvaluateAtBar(src, secCtx, idx-i)
			if err != nil {
				return math.NaN(), err
			}
			w := float64(p - i)
			s += w * v
			ws += w
		}
		return s / ws, nil
	}

	diffAt := func(idx int) (float64, error) {
		w1, err := wmaAt(sourceExpr, idx, halfPeriod)
		if err != nil || math.IsNaN(w1) {
			return math.NaN(), err
		}
		w2, err := wmaAt(sourceExpr, idx, period)
		if err != nil || math.IsNaN(w2) {
			return math.NaN(), err
		}
		return 2*w1 - w2, nil
	}

	finalSum, finalWSum := 0.0, 0.0
	for i := 0; i < sqrtPeriod; i++ {
		d, err := diffAt(barIdx - i)
		if err != nil || math.IsNaN(d) {
			return math.NaN(), err
		}
		w := float64(sqrtPeriod - i)
		finalSum += w * d
		finalWSum += w
	}
	return finalSum / finalWSum, nil
}

func (e *StreamingBarEvaluator) evaluateKCWAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceExpr, period, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	mult := 1.5
	if len(call.Arguments) >= 3 {
		if v, err2 := extractNumberLiteral(call.Arguments[2]); err2 == nil {
			mult = v
		}
	}
	useTrueRange := extractKCWUseTrueRangeArg(call)

	emaCacheKey := buildTACacheKey("ema", expressionKey(sourceExpr), period)
	emaState := e.getOrCreateTAState(emaCacheKey, period, secCtx)
	emaVal, err := emaState.ComputeAtBar(secCtx, sourceExpr, barIdx)
	if err != nil || math.IsNaN(emaVal) || emaVal == 0 {
		return math.NaN(), err
	}

	var rangeVal float64
	if useTrueRange {
		atrCacheKey := buildTACacheKey("atr", "hlc", period)
		atrState := e.getOrCreateTAState(atrCacheKey, period, secCtx)
		rangeVal, err = atrState.ComputeAtBar(secCtx, nil, barIdx)
		if err != nil || math.IsNaN(rangeVal) {
			return math.NaN(), err
		}
	} else {
		rangeVal, err = e.computeHighLowSMAAtBar(secCtx, barIdx, period)
		if err != nil || math.IsNaN(rangeVal) {
			return math.NaN(), err
		}
	}

	return 2.0 * mult * rangeVal / emaVal, nil
}

func extractKCWUseTrueRangeArg(call *ast.CallExpression) bool {
	if len(call.Arguments) < 4 {
		return true
	}
	lit, ok := call.Arguments[3].(*ast.Literal)
	if !ok {
		return true
	}
	boolVal, ok := lit.Value.(bool)
	if !ok {
		return true
	}
	return boolVal
}

func (e *StreamingBarEvaluator) computeHighLowSMAAtBar(secCtx *context.Context, barIdx int, period int) (float64, error) {
	if barIdx < period-1 {
		return math.NaN(), nil
	}
	sum := 0.0
	for i := 0; i < period; i++ {
		bar := secCtx.Data[barIdx-i]
		sum += bar.High - bar.Low
	}
	return sum / float64(period), nil
}

func (e *StreamingBarEvaluator) evaluateSARAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	start := 0.02
	inc := 0.02
	maxAF := 0.2
	if len(call.Arguments) >= 1 {
		if v, err := extractNumberLiteral(call.Arguments[0]); err == nil {
			start = v
		}
	}
	if len(call.Arguments) >= 2 {
		if v, err := extractNumberLiteral(call.Arguments[1]); err == nil {
			inc = v
		}
	}
	if len(call.Arguments) >= 3 {
		if v, err := extractNumberLiteral(call.Arguments[2]); err == nil {
			maxAF = v
		}
	}

	cacheKey := fmt.Sprintf("sar_%.4f_%.4f_%.4f", start, inc, maxAF)
	if state, exists := e.taStateCache[cacheKey]; exists {
		return state.ComputeAtBar(secCtx, nil, barIdx)
	}
	state := NewSARStateManager(cacheKey, start, inc, maxAF, len(secCtx.Data))
	e.taStateCache[cacheKey] = state
	return state.ComputeAtBar(secCtx, nil, barIdx)
}

func (e *StreamingBarEvaluator) fixnanCallAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	return e.fixnanEvaluator.EvaluateAtBar(e, call, secCtx, barIdx)
}

func init() {
	registerCallHandler("ta.sma", (*StreamingBarEvaluator).evaluateSMAAtBar)
	registerCallHandler("ta.ema", (*StreamingBarEvaluator).evaluateEMAAtBar)
	registerCallHandler("ta.rma", (*StreamingBarEvaluator).evaluateRMAAtBar)
	registerCallHandler("ta.rsi", (*StreamingBarEvaluator).evaluateRSIAtBar)
	registerCallHandler("ta.atr", (*StreamingBarEvaluator).evaluateATRAtBar)
	registerCallHandler("ta.stdev", (*StreamingBarEvaluator).evaluateSTDEVAtBar)
	registerCallHandler("ta.swma", (*StreamingBarEvaluator).evaluateSWMAAtBar)
	registerCallHandler("ta.cci", (*StreamingBarEvaluator).evaluateCCIAtBar)
	registerCallHandler("ta.bbw", (*StreamingBarEvaluator).evaluateBBWAtBar)
	registerCallHandler("ta.cog", (*StreamingBarEvaluator).evaluateCOGAtBar)
	registerCallHandler("ta.tsi", (*StreamingBarEvaluator).evaluateTSIAtBar)
	registerCallHandler("ta.pivothigh", (*StreamingBarEvaluator).evaluatePivotHighAtBar)
	registerCallHandler("ta.pivotlow", (*StreamingBarEvaluator).evaluatePivotLowAtBar)
	registerCallHandlerAliases((*StreamingBarEvaluator).evaluateValuewhenAtBar, "ta.valuewhen", "valuewhen")
	registerCallHandlerAliases((*StreamingBarEvaluator).fixnanCallAtBar, "fixnan", "ta.fixnan")
	registerCallHandler("ta.percentrank", (*StreamingBarEvaluator).evaluatePercentrankAtBar)
	registerCallHandler("ta.percentile_nearest_rank", (*StreamingBarEvaluator).evaluatePercentileNearestRankAtBar)
	registerCallHandler("ta.percentile_linear_interpolation", (*StreamingBarEvaluator).evaluatePercentileLinearInterpolationAtBar)
	registerCallHandler("ta.correlation", (*StreamingBarEvaluator).evaluateCorrelationAtBar)
	registerCallHandler("ta.wma", (*StreamingBarEvaluator).evaluateWMAAtBar)
	registerCallHandler("ta.alma", (*StreamingBarEvaluator).evaluateALMAAtBar)
	registerCallHandler("ta.hma", (*StreamingBarEvaluator).evaluateHMAAtBar)
	registerCallHandler("ta.kcw", (*StreamingBarEvaluator).evaluateKCWAtBar)
	registerCallHandler("ta.sar", (*StreamingBarEvaluator).evaluateSARAtBar)
	registerCallHandlerAliases((*StreamingBarEvaluator).evaluateTRFuncAtBar, "ta.tr", "tr")
}
