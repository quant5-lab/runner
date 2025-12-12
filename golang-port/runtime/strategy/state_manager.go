package strategy

import (
	"math"

	"github.com/quant5-lab/runner/runtime/series"
)

// StateManager samples strategy runtime state into Series buffers per bar
type StateManager struct {
	positionAvgPriceSeries *series.Series
	positionSizeSeries     *series.Series
	equitySeries           *series.Series
	netProfitSeries        *series.Series
	closedTradesSeries     *series.Series
}

// NewStateManager creates manager with Series buffers for given bar count
func NewStateManager(barCount int) *StateManager {
	return &StateManager{
		positionAvgPriceSeries: series.NewSeries(barCount),
		positionSizeSeries:     series.NewSeries(barCount),
		equitySeries:           series.NewSeries(barCount),
		netProfitSeries:        series.NewSeries(barCount),
		closedTradesSeries:     series.NewSeries(barCount),
	}
}

// SampleCurrentBar captures current strategy state into all Series at cursor position
func (sm *StateManager) SampleCurrentBar(strat *Strategy, currentPrice float64) {
	avgPrice := strat.GetPositionAvgPrice()
	if avgPrice == 0 {
		avgPrice = math.NaN()
	}

	sm.positionAvgPriceSeries.Set(avgPrice)
	sm.positionSizeSeries.Set(strat.GetPositionSize())
	sm.equitySeries.Set(strat.GetEquity(currentPrice))
	sm.netProfitSeries.Set(strat.GetNetProfit())
	sm.closedTradesSeries.Set(float64(len(strat.GetTradeHistory().GetClosedTrades())))
}

// AdvanceCursors moves all Series forward to next bar
func (sm *StateManager) AdvanceCursors() {
	sm.positionAvgPriceSeries.Next()
	sm.positionSizeSeries.Next()
	sm.equitySeries.Next()
	sm.netProfitSeries.Next()
	sm.closedTradesSeries.Next()
}

// PositionAvgPriceSeries returns Series for strategy.position_avg_price access
func (sm *StateManager) PositionAvgPriceSeries() *series.Series {
	return sm.positionAvgPriceSeries
}

// PositionSizeSeries returns Series for strategy.position_size access
func (sm *StateManager) PositionSizeSeries() *series.Series {
	return sm.positionSizeSeries
}

// EquitySeries returns Series for strategy.equity access
func (sm *StateManager) EquitySeries() *series.Series {
	return sm.equitySeries
}

// NetProfitSeries returns Series for strategy.netprofit access
func (sm *StateManager) NetProfitSeries() *series.Series {
	return sm.netProfitSeries
}

// ClosedTradesSeries returns Series for strategy.closedtrades access
func (sm *StateManager) ClosedTradesSeries() *series.Series {
	return sm.closedTradesSeries
}
