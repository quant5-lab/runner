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
	initialCapitalSeries   *series.Series
	grossProfitSeries      *series.Series
	grossLossSeries        *series.Series
	winTradesSeries        *series.Series
	lossTradesSeries       *series.Series
	evenTradesSeries       *series.Series
	openProfitSeries       *series.Series
	openTradesSeries       *series.Series
	avgTradeSeries         *series.Series
	avgWinningTradeSeries  *series.Series
	avgLosingTradeSeries   *series.Series
	maxDrawdownSeries      *series.Series
	maxRunupSeries         *series.Series
	maxDrawdownPctSeries   *series.Series
	maxRunupPctSeries      *series.Series

	intrabarCalc *IntrabarEquityCalculator
	ddRunupTrack *DrawdownRunupTracker
}

// NewStateManager creates manager with Series buffers for given bar count
func NewStateManager(barCount int) *StateManager {
	return &StateManager{
		positionAvgPriceSeries: series.NewSeries(barCount),
		positionSizeSeries:     series.NewSeries(barCount),
		equitySeries:           series.NewSeries(barCount),
		netProfitSeries:        series.NewSeries(barCount),
		closedTradesSeries:     series.NewSeries(barCount),
		initialCapitalSeries:   series.NewSeries(barCount),
		grossProfitSeries:      series.NewSeries(barCount),
		grossLossSeries:        series.NewSeries(barCount),
		winTradesSeries:        series.NewSeries(barCount),
		lossTradesSeries:       series.NewSeries(barCount),
		evenTradesSeries:       series.NewSeries(barCount),
		openProfitSeries:       series.NewSeries(barCount),
		openTradesSeries:       series.NewSeries(barCount),
		avgTradeSeries:         series.NewSeries(barCount),
		avgWinningTradeSeries:  series.NewSeries(barCount),
		avgLosingTradeSeries:   series.NewSeries(barCount),
		maxDrawdownSeries:      series.NewSeries(barCount),
		maxRunupSeries:         series.NewSeries(barCount),
		maxDrawdownPctSeries:   series.NewSeries(barCount),
		maxRunupPctSeries:      series.NewSeries(barCount),
		intrabarCalc:           NewIntrabarEquityCalculator(),
		ddRunupTrack:           NewDrawdownRunupTracker(),
	}
}

// SampleCurrentBar captures current strategy state into all Series at cursor position
func (sm *StateManager) SampleCurrentBar(strat *Strategy, closePrice, highPrice, lowPrice float64) {
	avgPrice := strat.GetPositionAvgPrice()
	if avgPrice == 0 {
		avgPrice = math.NaN()
	}

	equityClose := strat.GetEquity(closePrice)

	openTrades := strat.GetTradeHistory().GetOpenTrades()
	realizedProfit := strat.GetNetProfit()
	initialCapital := strat.GetInitialCapital()

	equityAdverse := sm.intrabarCalc.CalculateAdverseEquity(
		openTrades,
		realizedProfit,
		initialCapital,
		highPrice,
		lowPrice,
	)

	equityFavorable := sm.intrabarCalc.CalculateFavorableEquity(
		openTrades,
		realizedProfit,
		initialCapital,
		highPrice,
		lowPrice,
	)

	maxDrawdown, maxRunup, maxDrawdownPct, maxRunupPct := sm.ddRunupTrack.UpdateWithIntrabar(
		equityClose,
		equityAdverse,
		equityFavorable,
	)

	closedTrades := strat.GetTradeHistory().GetClosedTrades()

	sm.positionAvgPriceSeries.Set(avgPrice)
	sm.positionSizeSeries.Set(strat.GetPositionSize())
	sm.equitySeries.Set(equityClose)
	sm.netProfitSeries.Set(realizedProfit)
	sm.closedTradesSeries.Set(float64(len(closedTrades)))
	sm.initialCapitalSeries.Set(strat.GetInitialCapital())
	sm.grossProfitSeries.Set(AggregateGrossProfit(closedTrades))
	sm.grossLossSeries.Set(AggregateGrossLoss(closedTrades))
	sm.winTradesSeries.Set(float64(CountWinningTrades(closedTrades)))
	sm.lossTradesSeries.Set(float64(CountLosingTrades(closedTrades)))
	sm.evenTradesSeries.Set(float64(CountEvenTrades(closedTrades)))
	sm.openProfitSeries.Set(strat.GetOpenProfit(closePrice))
	sm.openTradesSeries.Set(float64(strat.GetOpenTradesCount()))
	sm.avgTradeSeries.Set(AvgTrade(closedTrades))
	sm.avgWinningTradeSeries.Set(AvgWinningTrade(closedTrades))
	sm.avgLosingTradeSeries.Set(AvgLosingTrade(closedTrades))
	sm.maxDrawdownSeries.Set(maxDrawdown)
	sm.maxRunupSeries.Set(maxRunup)
	sm.maxDrawdownPctSeries.Set(maxDrawdownPct)
	sm.maxRunupPctSeries.Set(maxRunupPct)
}

// AdvanceCursors moves all Series forward to next bar
func (sm *StateManager) AdvanceCursors() {
	sm.positionAvgPriceSeries.Next()
	sm.positionSizeSeries.Next()
	sm.equitySeries.Next()
	sm.netProfitSeries.Next()
	sm.closedTradesSeries.Next()
	sm.initialCapitalSeries.Next()
	sm.grossProfitSeries.Next()
	sm.grossLossSeries.Next()
	sm.winTradesSeries.Next()
	sm.lossTradesSeries.Next()
	sm.evenTradesSeries.Next()
	sm.openProfitSeries.Next()
	sm.openTradesSeries.Next()
	sm.avgTradeSeries.Next()
	sm.avgWinningTradeSeries.Next()
	sm.avgLosingTradeSeries.Next()
	sm.maxDrawdownSeries.Next()
	sm.maxRunupSeries.Next()
	sm.maxDrawdownPctSeries.Next()
	sm.maxRunupPctSeries.Next()
}

func (sm *StateManager) PositionAvgPriceSeries() *series.Series { return sm.positionAvgPriceSeries }
func (sm *StateManager) PositionSizeSeries() *series.Series     { return sm.positionSizeSeries }
func (sm *StateManager) EquitySeries() *series.Series           { return sm.equitySeries }
func (sm *StateManager) NetProfitSeries() *series.Series        { return sm.netProfitSeries }
func (sm *StateManager) ClosedTradesSeries() *series.Series     { return sm.closedTradesSeries }
func (sm *StateManager) InitialCapitalSeries() *series.Series   { return sm.initialCapitalSeries }
func (sm *StateManager) GrossProfitSeries() *series.Series      { return sm.grossProfitSeries }
func (sm *StateManager) GrossLossSeries() *series.Series        { return sm.grossLossSeries }
func (sm *StateManager) WinTradesSeries() *series.Series        { return sm.winTradesSeries }
func (sm *StateManager) LossTradesSeries() *series.Series       { return sm.lossTradesSeries }
func (sm *StateManager) EvenTradesSeries() *series.Series       { return sm.evenTradesSeries }
func (sm *StateManager) OpenProfitSeries() *series.Series       { return sm.openProfitSeries }
func (sm *StateManager) OpenTradesSeries() *series.Series       { return sm.openTradesSeries }
func (sm *StateManager) AvgTradeSeries() *series.Series         { return sm.avgTradeSeries }
func (sm *StateManager) AvgWinningTradeSeries() *series.Series  { return sm.avgWinningTradeSeries }
func (sm *StateManager) AvgLosingTradeSeries() *series.Series   { return sm.avgLosingTradeSeries }
func (sm *StateManager) MaxDrawdownSeries() *series.Series      { return sm.maxDrawdownSeries }
func (sm *StateManager) MaxRunupSeries() *series.Series         { return sm.maxRunupSeries }
func (sm *StateManager) MaxDrawdownPctSeries() *series.Series   { return sm.maxDrawdownPctSeries }
func (sm *StateManager) MaxRunupPctSeries() *series.Series      { return sm.maxRunupPctSeries }
