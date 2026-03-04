package security

import (
	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

type ATRStateManager struct {
	cacheKey        string
	period          int
	trCalculator    *TrueRangeCalculator
	rmaStateManager *RMAStateManager
	prevClose       float64
	computed        int
	hasHistory      bool
}

func NewATRStateManager(cacheKey string, period int, capacity int) *ATRStateManager {
	return &ATRStateManager{
		cacheKey:     cacheKey,
		period:       period,
		trCalculator: NewTrueRangeCalculator(),
		rmaStateManager: &RMAStateManager{
			cacheKey: cacheKey + "_rma_tr",
			period:   period,
			storage:  NewSeriesStorage(capacity),
			computed: 0,
		},
		computed:   0,
		hasHistory: false,
	}
}

func (s *ATRStateManager) ComputeAtBar(secCtx *context.Context, sourceID *ast.Identifier, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		if s.computed >= len(secCtx.Data) {
			break
		}

		isFirstBar := s.computed == 0 || !s.hasHistory
		trueRange := s.trCalculator.CalculateAtBar(secCtx.Data, s.computed, s.prevClose, isFirstBar, true)

		var atrValue float64
		if s.computed == 0 {
			atrValue = trueRange
		} else if s.computed < s.period {
			prevATR := s.rmaStateManager.storage.Get(s.computed - 1)
			atrValue = (prevATR*float64(s.computed) + trueRange) / float64(s.computed+1)
		} else {
			alpha := 1.0 / float64(s.period)
			prevATR := s.rmaStateManager.storage.Get(s.computed - 1)
			atrValue = alpha*trueRange + (1-alpha)*prevATR
		}

		s.rmaStateManager.storage.Set(s.computed, atrValue)
		s.prevClose = secCtx.Data[s.computed].Close
		s.hasHistory = true
		s.computed++
	}

	if barIdx < s.period-1 {
		return 0.0, nil
	}

	return s.rmaStateManager.storage.Get(barIdx), nil
}
