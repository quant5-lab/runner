package strategy

import "math"

type TradeAccessor struct {
	history *TradeHistory
}

func NewTradeAccessor(history *TradeHistory) *TradeAccessor {
	return &TradeAccessor{history: history}
}

func (ta *TradeAccessor) GetClosedTrade(index int) *Trade {
	closedTrades := ta.history.GetClosedTrades()
	if index < 0 || index >= len(closedTrades) {
		return nil
	}
	return &closedTrades[index]
}

func (ta *TradeAccessor) GetOpenTrade(index int) *Trade {
	openTrades := ta.history.GetOpenTrades()
	if index < 0 || index >= len(openTrades) {
		return nil
	}
	return &openTrades[index]
}

func (ta *TradeAccessor) ClosedTradeCommission(index int) float64 {
	trade := ta.GetClosedTrade(index)
	if trade == nil {
		return math.NaN()
	}
	return trade.Commission
}

func (ta *TradeAccessor) ClosedTradeEntryBarIndex(index int) float64 {
	trade := ta.GetClosedTrade(index)
	if trade == nil {
		return math.NaN()
	}
	return float64(trade.EntryBar)
}

func (ta *TradeAccessor) ClosedTradeEntryComment(index int) string {
	trade := ta.GetClosedTrade(index)
	if trade == nil {
		return ""
	}
	return trade.EntryComment
}

func (ta *TradeAccessor) ClosedTradeEntryID(index int) string {
	trade := ta.GetClosedTrade(index)
	if trade == nil {
		return ""
	}
	return trade.EntryID
}

func (ta *TradeAccessor) ClosedTradeEntryPrice(index int) float64 {
	trade := ta.GetClosedTrade(index)
	if trade == nil {
		return math.NaN()
	}
	return trade.EntryPrice
}

func (ta *TradeAccessor) ClosedTradeEntryTime(index int) float64 {
	trade := ta.GetClosedTrade(index)
	if trade == nil {
		return math.NaN()
	}
	return float64(trade.EntryTime)
}

func (ta *TradeAccessor) ClosedTradeExitBarIndex(index int) float64 {
	trade := ta.GetClosedTrade(index)
	if trade == nil {
		return math.NaN()
	}
	return float64(trade.ExitBar)
}

func (ta *TradeAccessor) ClosedTradeExitComment(index int) string {
	trade := ta.GetClosedTrade(index)
	if trade == nil {
		return ""
	}
	return trade.ExitComment
}

func (ta *TradeAccessor) ClosedTradeExitID(index int) string {
	trade := ta.GetClosedTrade(index)
	if trade == nil {
		return ""
	}
	return trade.ExitID
}

func (ta *TradeAccessor) ClosedTradeExitPrice(index int) float64 {
	trade := ta.GetClosedTrade(index)
	if trade == nil {
		return math.NaN()
	}
	return trade.ExitPrice
}

func (ta *TradeAccessor) ClosedTradeExitTime(index int) float64 {
	trade := ta.GetClosedTrade(index)
	if trade == nil {
		return math.NaN()
	}
	return float64(trade.ExitTime)
}

func (ta *TradeAccessor) ClosedTradeMaxDrawdown(index int) float64 {
	trade := ta.GetClosedTrade(index)
	if trade == nil {
		return math.NaN()
	}
	return trade.MaxDrawdown
}

func (ta *TradeAccessor) ClosedTradeMaxDrawdownPercent(index int) float64 {
	trade := ta.GetClosedTrade(index)
	if trade == nil {
		return math.NaN()
	}
	basis := trade.EntryPrice * trade.Size
	if basis == 0 {
		return math.NaN()
	}
	return (trade.MaxDrawdown / basis) * 100.0
}

func (ta *TradeAccessor) ClosedTradeMaxRunup(index int) float64 {
	trade := ta.GetClosedTrade(index)
	if trade == nil {
		return math.NaN()
	}
	return trade.MaxRunup
}

func (ta *TradeAccessor) ClosedTradeMaxRunupPercent(index int) float64 {
	trade := ta.GetClosedTrade(index)
	if trade == nil {
		return math.NaN()
	}
	basis := trade.EntryPrice * trade.Size
	if basis == 0 {
		return math.NaN()
	}
	return (trade.MaxRunup / basis) * 100.0
}

func (ta *TradeAccessor) ClosedTradeProfit(index int) float64 {
	trade := ta.GetClosedTrade(index)
	if trade == nil {
		return math.NaN()
	}
	return trade.Profit
}

func (ta *TradeAccessor) ClosedTradeProfitPercent(index int) float64 {
	trade := ta.GetClosedTrade(index)
	if trade == nil {
		return math.NaN()
	}
	basis := trade.EntryPrice * trade.Size
	if basis == 0 {
		return math.NaN()
	}
	return (trade.Profit / basis) * 100.0
}

func (ta *TradeAccessor) ClosedTradeSize(index int) float64 {
	trade := ta.GetClosedTrade(index)
	if trade == nil {
		return math.NaN()
	}
	if trade.Direction == Short {
		return -trade.Size
	}
	return trade.Size
}

func (ta *TradeAccessor) OpenTradeCommission(index int) float64 {
	trade := ta.GetOpenTrade(index)
	if trade == nil {
		return math.NaN()
	}
	return trade.Commission
}

func (ta *TradeAccessor) OpenTradeEntryBarIndex(index int) float64 {
	trade := ta.GetOpenTrade(index)
	if trade == nil {
		return math.NaN()
	}
	return float64(trade.EntryBar)
}

func (ta *TradeAccessor) OpenTradeEntryComment(index int) string {
	trade := ta.GetOpenTrade(index)
	if trade == nil {
		return ""
	}
	return trade.EntryComment
}

func (ta *TradeAccessor) OpenTradeEntryID(index int) string {
	trade := ta.GetOpenTrade(index)
	if trade == nil {
		return ""
	}
	return trade.EntryID
}

func (ta *TradeAccessor) OpenTradeEntryPrice(index int) float64 {
	trade := ta.GetOpenTrade(index)
	if trade == nil {
		return math.NaN()
	}
	return trade.EntryPrice
}

func (ta *TradeAccessor) OpenTradeEntryTime(index int) float64 {
	trade := ta.GetOpenTrade(index)
	if trade == nil {
		return math.NaN()
	}
	return float64(trade.EntryTime)
}

func (ta *TradeAccessor) OpenTradeMaxDrawdown(index int) float64 {
	trade := ta.GetOpenTrade(index)
	if trade == nil {
		return math.NaN()
	}
	return trade.MaxDrawdown
}

func (ta *TradeAccessor) OpenTradeMaxDrawdownPercent(index int) float64 {
	trade := ta.GetOpenTrade(index)
	if trade == nil {
		return math.NaN()
	}
	basis := trade.EntryPrice * trade.Size
	if basis == 0 {
		return math.NaN()
	}
	return (trade.MaxDrawdown / basis) * 100.0
}

func (ta *TradeAccessor) OpenTradeMaxRunup(index int) float64 {
	trade := ta.GetOpenTrade(index)
	if trade == nil {
		return math.NaN()
	}
	return trade.MaxRunup
}

func (ta *TradeAccessor) OpenTradeMaxRunupPercent(index int) float64 {
	trade := ta.GetOpenTrade(index)
	if trade == nil {
		return math.NaN()
	}
	basis := trade.EntryPrice * trade.Size
	if basis == 0 {
		return math.NaN()
	}
	return (trade.MaxRunup / basis) * 100.0
}

func (ta *TradeAccessor) OpenTradeProfit(index int) float64 {
	trade := ta.GetOpenTrade(index)
	if trade == nil {
		return math.NaN()
	}
	return trade.Profit
}

func (ta *TradeAccessor) OpenTradeProfitPercent(index int) float64 {
	trade := ta.GetOpenTrade(index)
	if trade == nil {
		return math.NaN()
	}
	basis := trade.EntryPrice * trade.Size
	if basis == 0 {
		return math.NaN()
	}
	return (trade.Profit / basis) * 100.0
}

func (ta *TradeAccessor) OpenTradeSize(index int) float64 {
	trade := ta.GetOpenTrade(index)
	if trade == nil {
		return math.NaN()
	}
	if trade.Direction == Short {
		return -trade.Size
	}
	return trade.Size
}
