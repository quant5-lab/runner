package security

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

type FixnanState struct {
	lastValidValue float64
}

func NewFixnanState() *FixnanState {
	return &FixnanState{
		lastValidValue: math.NaN(),
	}
}

func (s *FixnanState) ForwardFill(value float64) float64 {
	if !math.IsNaN(value) {
		s.lastValidValue = value
		return value
	}
	return s.lastValidValue
}

func computeExpressionHash(expr ast.Expression) string {
	data, _ := json.Marshal(expr)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash[:4])
}

func (e *StreamingBarEvaluator) evaluateFixnanAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	if len(call.Arguments) < 1 {
		return 0.0, newInsufficientArgumentsError("fixnan", 1, len(call.Arguments))
	}

	cacheKey := "fixnan_" + computeExpressionHash(call.Arguments[0])

	state, exists := e.fixnanStateCache[cacheKey]
	if !exists {
		state = NewFixnanState()
		e.fixnanStateCache[cacheKey] = state
	}

	value, err := e.EvaluateAtBar(call.Arguments[0], secCtx, barIdx)
	if err != nil {
		return 0.0, err
	}

	return state.ForwardFill(value), nil
}
