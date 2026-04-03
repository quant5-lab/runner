package strategy

type PositionReversalHandler struct {
	tradeHistory     *TradeHistory
	positionTracker  *PositionTracker
	equityCalculator *EquityCalculator
}

func NewPositionReversalHandler(
	tradeHistory *TradeHistory,
	positionTracker *PositionTracker,
	equityCalculator *EquityCalculator,
) *PositionReversalHandler {
	return &PositionReversalHandler{
		tradeHistory:     tradeHistory,
		positionTracker:  positionTracker,
		equityCalculator: equityCalculator,
	}
}

func (h *PositionReversalHandler) HandleReversal(
	entryDirection string,
	exitPrice float64,
	exitBar int,
	exitTime int64,
) {
	oppositeDirection := h.getOppositeDirection(entryDirection)
	h.closeTradesInDirection(oppositeDirection, exitPrice, exitBar, exitTime)
}

func (h *PositionReversalHandler) getOppositeDirection(direction string) string {
	if direction == Long {
		return Short
	}
	return Long
}

func (h *PositionReversalHandler) closeTradesInDirection(
	direction string,
	exitPrice float64,
	exitBar int,
	exitTime int64,
) {
	var tradesToClose []Trade
	openTrades := h.tradeHistory.GetOpenTrades()

	for _, trade := range openTrades {
		if trade.Direction == direction {
			tradesToClose = append(tradesToClose, trade)
		}
	}

	for _, trade := range tradesToClose {
		h.closeTrade(trade, exitPrice, exitBar, exitTime)
	}
}

func (h *PositionReversalHandler) closeTrade(
	trade Trade,
	exitPrice float64,
	exitBar int,
	exitTime int64,
) {
	closedTrade := h.tradeHistory.CloseTrade(
		trade.EntryID,
		"",
		exitPrice,
		exitBar,
		exitTime,
		"Position reversal",
		0.0,
	)

	if closedTrade != nil {
		oppositeDirection := h.getOppositeDirection(trade.Direction)
		h.positionTracker.UpdatePosition(trade.Size, exitPrice, oppositeDirection)
		h.equityCalculator.UpdateFromClosedTrade(*closedTrade)
	}
}
