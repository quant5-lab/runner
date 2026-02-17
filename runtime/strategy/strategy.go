package strategy

import (
	"fmt"
	"math"
)

/* Direction constants */
const (
	Long  = "long"
	Short = "short"
)

/* OrderAction constants - TradingView order execution model */
const (
	OrderActionEntry    = "entry"
	OrderActionClose    = "close"
	OrderActionCloseAll = "close_all"
)

/* Trade represents a single trade (open or closed) */
type Trade struct {
	EntryID      string  `json:"entryId"`
	Direction    string  `json:"direction"`
	Size         float64 `json:"size"`
	EntryPrice   float64 `json:"entryPrice"`
	EntryBar     int     `json:"entryBar"`
	EntryTime    int64   `json:"entryTime"`
	EntryComment string  `json:"entryComment"`
	ExitPrice    float64 `json:"exitPrice"`
	ExitBar      int     `json:"exitBar"`
	ExitTime     int64   `json:"exitTime"`
	ExitComment  string  `json:"exitComment"`
	Profit       float64 `json:"profit"`
}

/* Order represents a pending order - unified for entry and close operations */
type Order struct {
	ID           string
	Action       string // OrderActionEntry, OrderActionClose, OrderActionCloseAll
	Direction    string
	Qty          float64
	Type         string
	CreatedBar   int
	EntryComment string
	FromEntry    string // Target entry ID for close orders
	ExitComment  string // Comment for close orders
}

/* OrderManager manages pending orders */
type OrderManager struct {
	orders      []Order
	nextOrderID int
}

/* NewOrderManager creates a new order manager */
func NewOrderManager() *OrderManager {
	return &OrderManager{
		orders:      []Order{},
		nextOrderID: 1,
	}
}

/* CreateEntryOrder creates a pending entry order */
func (om *OrderManager) CreateEntryOrder(id, direction string, qty float64, createdBar int, comment string) Order {
	om.removeOrderByID(id)

	order := Order{
		ID:           id,
		Action:       OrderActionEntry,
		Direction:    direction,
		Qty:          qty,
		Type:         "market",
		CreatedBar:   createdBar,
		EntryComment: comment,
	}
	om.orders = append(om.orders, order)
	return order
}

/* CreateCloseOrder creates a pending close order for specific entry */
func (om *OrderManager) CreateCloseOrder(fromEntry string, createdBar int, comment string) Order {
	orderID := fmt.Sprintf("_close_%s_%d", fromEntry, createdBar)
	om.removeOrderByID(orderID)

	order := Order{
		ID:          orderID,
		Action:      OrderActionClose,
		Type:        "market",
		CreatedBar:  createdBar,
		FromEntry:   fromEntry,
		ExitComment: comment,
	}
	om.orders = append(om.orders, order)
	return order
}

/* CreateCloseAllOrder creates a pending close-all order */
func (om *OrderManager) CreateCloseAllOrder(createdBar int, comment string) Order {
	orderID := fmt.Sprintf("_close_all_%d", createdBar)
	om.removeOrderByID(orderID)

	order := Order{
		ID:          orderID,
		Action:      OrderActionCloseAll,
		Type:        "market",
		CreatedBar:  createdBar,
		ExitComment: comment,
	}
	om.orders = append(om.orders, order)
	return order
}

/* CreateOrder creates or replaces an order - legacy compatibility */
func (om *OrderManager) CreateOrder(id, direction string, qty float64, createdBar int, comment string) Order {
	return om.CreateEntryOrder(id, direction, qty, createdBar, comment)
}

func (om *OrderManager) removeOrderByID(id string) {
	for i, order := range om.orders {
		if order.ID == id {
			om.orders = append(om.orders[:i], om.orders[i+1:]...)
			return
		}
	}
}

/* GetPendingOrders returns orders ready to execute */
func (om *OrderManager) GetPendingOrders(currentBar int) []Order {
	pending := []Order{}
	for _, order := range om.orders {
		if order.CreatedBar < currentBar {
			pending = append(pending, order)
		}
	}
	return pending
}

/* RemoveOrder removes an order by ID */
func (om *OrderManager) RemoveOrder(id string) {
	for i, order := range om.orders {
		if order.ID == id {
			om.orders = append(om.orders[:i], om.orders[i+1:]...)
			return
		}
	}
}

/* PositionTracker tracks current position */
type PositionTracker struct {
	positionSize     float64
	positionAvgPrice float64
	totalCost        float64
}

/* NewPositionTracker creates a new position tracker */
func NewPositionTracker() *PositionTracker {
	return &PositionTracker{}
}

/* UpdatePosition updates position from trade */
func (pt *PositionTracker) UpdatePosition(qty, price float64, direction string) {
	sizeChange := qty
	if direction == Short {
		sizeChange = -qty
	}

	// Check if closing or opening position
	if (pt.positionSize > 0 && sizeChange < 0) || (pt.positionSize < 0 && sizeChange > 0) {
		// Closing or reducing position
		pt.positionSize += sizeChange
		if pt.positionSize == 0 {
			pt.positionAvgPrice = 0
			pt.totalCost = 0
		} else {
			pt.totalCost = pt.positionAvgPrice * abs(pt.positionSize)
		}
	} else {
		// Opening or adding to position
		addedCost := qty * price
		pt.totalCost += addedCost
		pt.positionSize += sizeChange
		if pt.positionSize != 0 {
			pt.positionAvgPrice = pt.totalCost / abs(pt.positionSize)
		} else {
			pt.positionAvgPrice = 0
		}
	}
}

/* GetPositionSize returns current position size */
func (pt *PositionTracker) GetPositionSize() float64 {
	return pt.positionSize
}

/* GetAvgPrice returns average entry price */
func (pt *PositionTracker) GetAvgPrice() float64 {
	return pt.positionAvgPrice
}

/* TradeHistory tracks open and closed trades */
type TradeHistory struct {
	openTrades   []Trade
	closedTrades []Trade
}

/* NewTradeHistory creates a new trade history */
func NewTradeHistory() *TradeHistory {
	return &TradeHistory{
		openTrades:   []Trade{},
		closedTrades: []Trade{},
	}
}

/* AddOpenTrade adds a new open trade */
func (th *TradeHistory) AddOpenTrade(trade Trade) {
	th.openTrades = append(th.openTrades, trade)
}

/* CloseTrade closes a trade by entry ID */
func (th *TradeHistory) CloseTrade(entryID string, exitPrice float64, exitBar int, exitTime int64, exitComment string) *Trade {
	for i, trade := range th.openTrades {
		if trade.EntryID == entryID {
			trade.ExitPrice = exitPrice
			trade.ExitBar = exitBar
			trade.ExitTime = exitTime
			trade.ExitComment = exitComment

			// Calculate profit
			priceDiff := exitPrice - trade.EntryPrice
			multiplier := 1.0
			if trade.Direction == Short {
				multiplier = -1.0
			}
			trade.Profit = priceDiff * trade.Size * multiplier

			th.closedTrades = append(th.closedTrades, trade)
			th.openTrades = append(th.openTrades[:i], th.openTrades[i+1:]...)
			return &trade
		}
	}
	return nil
}

/* GetOpenTrades returns open trades */
func (th *TradeHistory) GetOpenTrades() []Trade {
	return th.openTrades
}

/* GetClosedTrades returns closed trades */
func (th *TradeHistory) GetClosedTrades() []Trade {
	return th.closedTrades
}

/* EquityCalculator calculates equity */
type EquityCalculator struct {
	initialCapital float64
	realizedProfit float64
}

/* NewEquityCalculator creates a new equity calculator */
func NewEquityCalculator(initialCapital float64) *EquityCalculator {
	return &EquityCalculator{
		initialCapital: initialCapital,
		realizedProfit: 0,
	}
}

/* UpdateFromClosedTrade updates realized profit from closed trade */
func (ec *EquityCalculator) UpdateFromClosedTrade(trade Trade) {
	ec.realizedProfit += trade.Profit
}

/* GetEquity returns current equity including unrealized profit */
func (ec *EquityCalculator) GetEquity(unrealizedProfit float64) float64 {
	return ec.initialCapital + ec.realizedProfit + unrealizedProfit
}

/* GetNetProfit returns realized profit */
func (ec *EquityCalculator) GetNetProfit() float64 {
	return ec.realizedProfit
}

/* Strategy implements strategy operations */
type Strategy struct {
	context          interface{} // Context with OHLCV data
	orderManager     *OrderManager
	positionTracker  *PositionTracker
	tradeHistory     *TradeHistory
	equityCalculator *EquityCalculator
	reversalHandler  *PositionReversalHandler
	initialized      bool
	currentBar       int
	currentPrice     float64
	pyramiding       int
}

func NewStrategy() *Strategy {
	om := NewOrderManager()
	pt := NewPositionTracker()
	th := NewTradeHistory()
	ec := NewEquityCalculator(10000)
	rh := NewPositionReversalHandler(th, pt, ec)

	return &Strategy{
		orderManager:     om,
		positionTracker:  pt,
		tradeHistory:     th,
		equityCalculator: ec,
		reversalHandler:  rh,
		initialized:      false,
		pyramiding:       -1,
	}
}

func (s *Strategy) Call(strategyName string, initialCapital float64) {
	s.initialized = true
	s.equityCalculator = NewEquityCalculator(initialCapital)
	s.reversalHandler.equityCalculator = s.equityCalculator
	s.pyramiding = -1
}

func (s *Strategy) CallWithPyramiding(strategyName string, initialCapital float64, pyramiding int) {
	s.initialized = true
	s.equityCalculator = NewEquityCalculator(initialCapital)
	s.reversalHandler.equityCalculator = s.equityCalculator
	s.pyramiding = pyramiding
}

func (s *Strategy) Entry(id, direction string, qty float64, comment string) error {
	if !s.initialized {
		return fmt.Errorf("strategy not initialized")
	}

	if s.pyramiding >= 0 {
		openTrades := s.tradeHistory.GetOpenTrades()
		sameDirectionCount := 0
		for _, trade := range openTrades {
			if trade.Direction == direction {
				sameDirectionCount++
			}
		}

		pendingOrders := s.orderManager.GetPendingOrders(s.currentBar + 1)
		for _, order := range pendingOrders {
			if order.Direction == direction {
				sameDirectionCount++
			}
		}

		if sameDirectionCount > s.pyramiding {
			return nil
		}
	}

	s.orderManager.CreateOrder(id, direction, qty, s.currentBar, comment)
	return nil
}

/* Close creates a pending close order - fills at next bar open per TradingView model */
func (s *Strategy) Close(id string, currentPrice float64, currentTime int64, comment string) {
	if !s.initialized {
		return
	}

	// Verify trade exists before creating close order
	openTrades := s.tradeHistory.GetOpenTrades()
	for _, trade := range openTrades {
		if trade.EntryID == id {
			s.orderManager.CreateCloseOrder(id, s.currentBar, comment)
			return
		}
	}
}

/* executeCloseOrder executes a close order at given price - internal use only */
func (s *Strategy) executeCloseOrder(entryID string, fillPrice float64, fillBar int, fillTime int64, comment string) {
	openTrades := s.tradeHistory.GetOpenTrades()
	for _, trade := range openTrades {
		if trade.EntryID == entryID {
			closedTrade := s.tradeHistory.CloseTrade(trade.EntryID, fillPrice, fillBar, fillTime, comment)
			if closedTrade != nil {
				// Update position tracker
				oppositeDir := Long
				if trade.Direction == Long {
					oppositeDir = Short
				}
				s.positionTracker.UpdatePosition(trade.Size, fillPrice, oppositeDir)

				// Update equity
				s.equityCalculator.UpdateFromClosedTrade(*closedTrade)
			}
			return
		}
	}
}

/* CloseAll creates a pending close-all order - fills at next bar open per TradingView model */
func (s *Strategy) CloseAll(currentPrice float64, currentTime int64, comment string) {
	if !s.initialized {
		return
	}

	// Only create close-all order if there are open trades
	openTrades := s.tradeHistory.GetOpenTrades()
	if len(openTrades) > 0 {
		s.orderManager.CreateCloseAllOrder(s.currentBar, comment)
	}
}

/* executeCloseAllOrder executes a close-all order at given price - internal use only */
func (s *Strategy) executeCloseAllOrder(fillPrice float64, fillBar int, fillTime int64, comment string) {
	openTrades := s.tradeHistory.GetOpenTrades()
	for _, trade := range openTrades {
		closedTrade := s.tradeHistory.CloseTrade(trade.EntryID, fillPrice, fillBar, fillTime, comment)
		if closedTrade != nil {
			// Update position tracker
			oppositeDir := Long
			if trade.Direction == Long {
				oppositeDir = Short
			}
			s.positionTracker.UpdatePosition(trade.Size, fillPrice, oppositeDir)

			// Update equity
			s.equityCalculator.UpdateFromClosedTrade(*closedTrade)
		}
	}
}

/* Exit exits with stop/limit orders (simplified - just closes) */
func (s *Strategy) Exit(id, fromEntry string, currentPrice float64, currentTime int64, comment string) {
	s.Close(fromEntry, currentPrice, currentTime, comment)
}

/* ExitWithLevels checks stop/limit levels and closes if triggered */
func (s *Strategy) ExitWithLevels(exitID, fromEntry string, stopLevel, limitLevel, barHigh, barLow, barClose float64, barTime int64, comment string) {
	if !s.initialized {
		return
	}

	// Find open trade by entry ID
	openTrades := s.tradeHistory.GetOpenTrades()
	var trade *Trade
	for i := range openTrades {
		if openTrades[i].EntryID == fromEntry {
			trade = &openTrades[i]
			break
		}
	}

	if trade == nil {
		return
	}

	// Check stop loss (long: low <= stop, short: high >= stop)
	if !math.IsNaN(stopLevel) {
		if trade.Direction == Long && barLow <= stopLevel {
			s.executeCloseOrder(fromEntry, stopLevel, s.currentBar, barTime, comment)
			return
		}
		if trade.Direction == Short && barHigh >= stopLevel {
			s.executeCloseOrder(fromEntry, stopLevel, s.currentBar, barTime, comment)
			return
		}
	}

	// Check take profit (long: high >= limit, short: low <= limit)
	if !math.IsNaN(limitLevel) {
		if trade.Direction == Long && barHigh >= limitLevel {
			s.executeCloseOrder(fromEntry, limitLevel, s.currentBar, barTime, comment)
			return
		}
		if trade.Direction == Short && barLow <= limitLevel {
			s.executeCloseOrder(fromEntry, limitLevel, s.currentBar, barTime, comment)
			return
		}
	}
}

/* OnBarUpdate processes pending orders at bar open */
func (s *Strategy) OnBarUpdate(currentBar int, openPrice float64, openTime int64) {
	if !s.initialized {
		return
	}

	s.currentBar = currentBar
	s.currentPrice = openPrice
	pendingOrders := s.orderManager.GetPendingOrders(currentBar)

	for _, order := range pendingOrders {
		switch order.Action {
		case OrderActionEntry:
			s.reversalHandler.HandleReversal(order.Direction, openPrice, currentBar, openTime)

			s.positionTracker.UpdatePosition(order.Qty, openPrice, order.Direction)

			s.tradeHistory.AddOpenTrade(Trade{
				EntryID:      order.ID,
				Direction:    order.Direction,
				Size:         order.Qty,
				EntryPrice:   openPrice,
				EntryBar:     currentBar,
				EntryTime:    openTime,
				EntryComment: order.EntryComment,
			})

		case OrderActionClose:
			s.executeCloseOrder(order.FromEntry, openPrice, currentBar, openTime, order.ExitComment)

		case OrderActionCloseAll:
			s.executeCloseAllOrder(openPrice, currentBar, openTime, order.ExitComment)
		}

		s.orderManager.RemoveOrder(order.ID)
	}
}

/* GetPositionSize returns current position size */
func (s *Strategy) GetPositionSize() float64 {
	return s.positionTracker.GetPositionSize()
}

/* GetPositionAvgPrice returns average entry price */
func (s *Strategy) GetPositionAvgPrice() float64 {
	avgPrice := s.positionTracker.GetAvgPrice()
	if avgPrice == 0 {
		return math.NaN()
	}
	return avgPrice
}

/* GetEquity returns current equity including unrealized P&L */
func (s *Strategy) GetEquity(currentPrice float64) float64 {
	unrealizedPL := 0.0
	openTrades := s.tradeHistory.GetOpenTrades()

	for _, trade := range openTrades {
		priceDiff := currentPrice - trade.EntryPrice
		multiplier := 1.0
		if trade.Direction == Short {
			multiplier = -1.0
		}
		unrealizedPL += priceDiff * trade.Size * multiplier
	}

	return s.equityCalculator.GetEquity(unrealizedPL)
}

/* Equity returns current equity */
func (s *Strategy) Equity() float64 {
	return s.GetEquity(s.currentPrice)
}

/* GetNetProfit returns realized profit */
func (s *Strategy) GetNetProfit() float64 {
	return s.equityCalculator.GetNetProfit()
}

/* GetTradeHistory returns trade history (for chart data export) */
func (s *Strategy) GetTradeHistory() *TradeHistory {
	return s.tradeHistory
}

func (s *Strategy) GetInitialCapital() float64 {
	return s.equityCalculator.initialCapital
}

func (s *Strategy) GetGrossProfit() float64 {
	return AggregateGrossProfit(s.tradeHistory.GetClosedTrades())
}

func (s *Strategy) GetGrossLoss() float64 {
	return AggregateGrossLoss(s.tradeHistory.GetClosedTrades())
}

func (s *Strategy) GetWinningTradesCount() int {
	return CountWinningTrades(s.tradeHistory.GetClosedTrades())
}

func (s *Strategy) GetLosingTradesCount() int {
	return CountLosingTrades(s.tradeHistory.GetClosedTrades())
}

func (s *Strategy) GetEvenTradesCount() int {
	return CountEvenTrades(s.tradeHistory.GetClosedTrades())
}

/* Helper function */
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
