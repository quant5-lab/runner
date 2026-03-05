package security

import (
	"math"
	"sort"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func collectOHLCVWindow(sourceID *ast.Identifier, secCtx *context.Context, barIdx, length int) ([]float64, error) {
	vals := make([]float64, length)
	for i := 0; i < length; i++ {
		v, err := evaluateOHLCVAtBar(sourceID, secCtx, barIdx-length+1+i)
		if err != nil {
			return nil, err
		}
		vals[i] = v
	}
	return vals, nil
}

func windowSum(vals []float64) float64 {
	s := 0.0
	for _, v := range vals {
		s += v
	}
	return s
}

func windowMean(vals []float64) float64 {
	return windowSum(vals) / float64(len(vals))
}

func (e *StreamingBarEvaluator) evaluateHighestAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, length, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < length-1 {
		return math.NaN(), nil
	}
	vals, err := collectOHLCVWindow(sourceID, secCtx, barIdx, length)
	if err != nil {
		return math.NaN(), err
	}
	hi := math.Inf(-1)
	for _, v := range vals {
		if v > hi {
			hi = v
		}
	}
	return hi, nil
}

func (e *StreamingBarEvaluator) evaluateLowestAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, length, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < length-1 {
		return math.NaN(), nil
	}
	vals, err := collectOHLCVWindow(sourceID, secCtx, barIdx, length)
	if err != nil {
		return math.NaN(), err
	}
	lo := math.Inf(1)
	for _, v := range vals {
		if v < lo {
			lo = v
		}
	}
	return lo, nil
}

func (e *StreamingBarEvaluator) evaluateSumAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, length, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < length-1 {
		return math.NaN(), nil
	}
	vals, err := collectOHLCVWindow(sourceID, secCtx, barIdx, length)
	if err != nil {
		return math.NaN(), err
	}
	return windowSum(vals), nil
}

func (e *StreamingBarEvaluator) evaluateRangeAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, length, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < length-1 {
		return math.NaN(), nil
	}
	vals, err := collectOHLCVWindow(sourceID, secCtx, barIdx, length)
	if err != nil {
		return math.NaN(), err
	}
	hi, lo := math.Inf(-1), math.Inf(1)
	for _, v := range vals {
		if v > hi {
			hi = v
		}
		if v < lo {
			lo = v
		}
	}
	return hi - lo, nil
}

func (e *StreamingBarEvaluator) evaluateDevAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, length, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < length-1 {
		return math.NaN(), nil
	}
	vals, err := collectOHLCVWindow(sourceID, secCtx, barIdx, length)
	if err != nil {
		return math.NaN(), err
	}
	mean := windowMean(vals)
	mad := 0.0
	for _, v := range vals {
		d := v - mean
		if d < 0 {
			d = -d
		}
		mad += d
	}
	return mad / float64(length), nil
}

func (e *StreamingBarEvaluator) evaluateVarianceAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, length, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < length-1 {
		return math.NaN(), nil
	}

	biased := true
	if len(call.Arguments) >= 3 {
		if lit, ok := call.Arguments[2].(*ast.Literal); ok {
			if b, ok2 := lit.Value.(bool); ok2 {
				biased = b
			}
		}
	}

	vals, err := collectOHLCVWindow(sourceID, secCtx, barIdx, length)
	if err != nil {
		return math.NaN(), err
	}
	mean := windowMean(vals)
	variance := 0.0
	for _, v := range vals {
		d := v - mean
		variance += d * d
	}
	divisor := float64(length)
	if !biased {
		if length <= 1 {
			return 0.0, nil
		}
		divisor = float64(length - 1)
	}
	return variance / divisor, nil
}

func (e *StreamingBarEvaluator) evaluateMedianAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, length, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < length-1 {
		return math.NaN(), nil
	}
	vals, err := collectOHLCVWindow(sourceID, secCtx, barIdx, length)
	if err != nil {
		return math.NaN(), err
	}
	sorted := make([]float64, length)
	copy(sorted, vals)
	sort.Float64s(sorted)
	mid := length / 2
	if length%2 == 1 {
		return sorted[mid], nil
	}
	return (sorted[mid-1] + sorted[mid]) / 2.0, nil
}

func (e *StreamingBarEvaluator) evaluateModeAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, length, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < length-1 {
		return math.NaN(), nil
	}
	vals, err := collectOHLCVWindow(sourceID, secCtx, barIdx, length)
	if err != nil {
		return math.NaN(), err
	}
	counts := make(map[float64]int, length)
	for _, v := range vals {
		counts[v]++
	}
	maxCount := 0
	mode := math.NaN()
	for v, c := range counts {
		if c > maxCount || (c == maxCount && (math.IsNaN(mode) || v < mode)) {
			maxCount = c
			mode = v
		}
	}
	return mode, nil
}

func (e *StreamingBarEvaluator) evaluateCMOAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, length, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < length {
		return math.NaN(), nil
	}
	sumGain, sumLoss := 0.0, 0.0
	for i := 0; i < length; i++ {
		curr, err := evaluateOHLCVAtBar(sourceID, secCtx, barIdx-i)
		if err != nil {
			return math.NaN(), err
		}
		prev, err := evaluateOHLCVAtBar(sourceID, secCtx, barIdx-i-1)
		if err != nil {
			return math.NaN(), err
		}
		change := curr - prev
		if change > 0 {
			sumGain += change
		} else {
			sumLoss -= change
		}
	}
	total := sumGain + sumLoss
	if total == 0.0 {
		return 0.0, nil
	}
	return 100.0 * (sumGain - sumLoss) / total, nil
}

func (e *StreamingBarEvaluator) evaluateWPRAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	length, err := extractPeriodArgument(call, "wpr")
	if err != nil {
		return 0.0, err
	}
	if barIdx < length-1 {
		return math.NaN(), nil
	}
	hi, lo := math.Inf(-1), math.Inf(1)
	for i := 0; i < length; i++ {
		bar := secCtx.Data[barIdx-i]
		if bar.High > hi {
			hi = bar.High
		}
		if bar.Low < lo {
			lo = bar.Low
		}
	}
	closeVal := secCtx.Data[barIdx].Close
	if hi == lo {
		return 0.0, nil
	}
	return (hi - closeVal) / (hi - lo) * -100.0, nil
}

func (e *StreamingBarEvaluator) evaluateMFIAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, length, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < length {
		return math.NaN(), nil
	}
	posFlow, negFlow := 0.0, 0.0
	for i := 0; i < length; i++ {
		idx := barIdx - i
		curr, err := evaluateOHLCVAtBar(sourceID, secCtx, idx)
		if err != nil {
			return math.NaN(), err
		}
		prev, err := evaluateOHLCVAtBar(sourceID, secCtx, idx-1)
		if err != nil {
			return math.NaN(), err
		}
		flow := curr * secCtx.Data[idx].Volume
		if curr > prev {
			posFlow += flow
		} else {
			negFlow += flow
		}
	}
	if negFlow == 0.0 {
		return 100.0, nil
	}
	return 100.0 - (100.0 / (1.0 + posFlow/negFlow)), nil
}

func (e *StreamingBarEvaluator) evaluateVWMAAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, length, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < length-1 {
		return math.NaN(), nil
	}
	weightedSum, volumeSum := 0.0, 0.0
	for i := 0; i < length; i++ {
		idx := barIdx - i
		val, err := evaluateOHLCVAtBar(sourceID, secCtx, idx)
		if err != nil {
			return math.NaN(), err
		}
		vol := secCtx.Data[idx].Volume
		weightedSum += val * vol
		volumeSum += vol
	}
	if volumeSum == 0.0 {
		return math.NaN(), nil
	}
	return weightedSum / volumeSum, nil
}

func (e *StreamingBarEvaluator) evaluateLinregAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, length, offset, err := extractLinregArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < length-1 {
		return math.NaN(), nil
	}
	vals, err := collectOHLCVWindow(sourceID, secCtx, barIdx, length)
	if err != nil {
		return math.NaN(), err
	}
	n := float64(length)
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0
	for i, y := range vals {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	denom := n*sumX2 - sumX*sumX
	if denom == 0.0 {
		return windowMean(vals), nil
	}
	slope := (n*sumXY - sumX*sumY) / denom
	intercept := (sumY - slope*sumX) / n
	return slope*float64(length-1-offset) + intercept, nil
}

func (e *StreamingBarEvaluator) evaluateHighestBarsAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, length, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < length-1 {
		return math.NaN(), nil
	}
	vals, err := collectOHLCVWindow(sourceID, secCtx, barIdx, length)
	if err != nil {
		return math.NaN(), err
	}
	hi := math.Inf(-1)
	hiOffset := 0
	for i, v := range vals {
		if v > hi {
			hi = v
			hiOffset = length - 1 - i
		}
	}
	return float64(-hiOffset), nil
}

func (e *StreamingBarEvaluator) evaluateLowestBarsAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	sourceID, length, err := extractTAArguments(call, e.inputConstantsMap)
	if err != nil {
		return 0.0, err
	}
	if barIdx < length-1 {
		return math.NaN(), nil
	}
	vals, err := collectOHLCVWindow(sourceID, secCtx, barIdx, length)
	if err != nil {
		return math.NaN(), err
	}
	lo := math.Inf(1)
	loOffset := 0
	for i, v := range vals {
		if v < lo {
			lo = v
			loOffset = length - 1 - i
		}
	}
	return float64(-loOffset), nil
}
