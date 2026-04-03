package codegen

const (
	StrategyPositionAvgPriceSeriesName = "strategy_position_avg_priceSeries"
	StrategyPositionSizeSeriesName     = "strategy_position_sizeSeries"
	StrategyEquitySeriesName           = "strategy_equitySeries"
	StrategyNetProfitSeriesName        = "strategy_netprofitSeries"
	StrategyClosedTradesSeriesName     = "strategy_closedtradesSeries"
	StrategyInitialCapitalSeriesName   = "strategy_initial_capitalSeries"
	StrategyGrossProfitSeriesName      = "strategy_grossprofitSeries"
	StrategyGrossLossSeriesName        = "strategy_grosslossSeries"
	StrategyWinTradesSeriesName        = "strategy_wintradesSeries"
	StrategyLossTradesSeriesName       = "strategy_losstradesSeries"
	StrategyEvenTradesSeriesName       = "strategy_eventradesSeries"
	StrategyOpenProfitSeriesName       = "strategy_openprofitSeries"
	StrategyOpenTradesSeriesName       = "strategy_opentradesSeries"
	StrategyAvgTradeSeriesName         = "strategy_avg_tradeSeries"
	StrategyAvgWinningTradeSeriesName  = "strategy_avg_winning_tradeSeries"
	StrategyAvgLosingTradeSeriesName   = "strategy_avg_losing_tradeSeries"
	StrategyMaxDrawdownSeriesName      = "strategy_max_drawdownSeries"
	StrategyMaxRunupSeriesName         = "strategy_max_runupSeries"
	StrategyMaxDrawdownPctSeriesName   = "strategy_max_drawdown_percentSeries"
	StrategyMaxRunupPctSeriesName      = "strategy_max_runup_percentSeries"
)

type strategySeriesBinding struct {
	varName  string
	accessor string
}

func strategySeriesBindings() []strategySeriesBinding {
	return []strategySeriesBinding{
		{StrategyPositionAvgPriceSeriesName, "PositionAvgPriceSeries"},
		{StrategyPositionSizeSeriesName, "PositionSizeSeries"},
		{StrategyEquitySeriesName, "EquitySeries"},
		{StrategyNetProfitSeriesName, "NetProfitSeries"},
		{StrategyClosedTradesSeriesName, "ClosedTradesSeries"},
		{StrategyInitialCapitalSeriesName, "InitialCapitalSeries"},
		{StrategyGrossProfitSeriesName, "GrossProfitSeries"},
		{StrategyGrossLossSeriesName, "GrossLossSeries"},
		{StrategyWinTradesSeriesName, "WinTradesSeries"},
		{StrategyLossTradesSeriesName, "LossTradesSeries"},
		{StrategyEvenTradesSeriesName, "EvenTradesSeries"},
		{StrategyOpenProfitSeriesName, "OpenProfitSeries"},
		{StrategyOpenTradesSeriesName, "OpenTradesSeries"},
		{StrategyAvgTradeSeriesName, "AvgTradeSeries"},
		{StrategyAvgWinningTradeSeriesName, "AvgWinningTradeSeries"},
		{StrategyAvgLosingTradeSeriesName, "AvgLosingTradeSeries"},
		{StrategyMaxDrawdownSeriesName, "MaxDrawdownSeries"},
		{StrategyMaxRunupSeriesName, "MaxRunupSeries"},
		{StrategyMaxDrawdownPctSeriesName, "MaxDrawdownPctSeries"},
		{StrategyMaxRunupPctSeriesName, "MaxRunupPctSeries"},
	}
}
