package strategy

import (
	"fmt"
)

/* Direction constants */
const (
	Long  = "long"
	Short = "short"
)

/* Trade represents a single trade (open or closed) */
type Trade struct {
	EntryID    string
	Direction  string
	Size       float64
	EntryPrice float64
	EntryBar   int
	EntryTime  int64
	ExitPrice  float64
	ExitBar    int
	ExitTime   int64
	Profit     float64
}

/* Order represents a pending order */
type Order struct {
	ID         string
	Direction  string
	Qty        float64
	Type       string
	CreatedBar int
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

/* CreateOrder creates or replaces an order */
func (om *OrderManager) CreateOrder(id, direction string, qty float64, createdBar int) Order {
	// Remove existing order with same ID
	for i, order := range om.orders {
		if order.ID == id {
			om.orders = append(om.orders[:i], om.orders[i+1:]...)
			break
		}
	}

	order := Order{
		ID:         id,
		Direction:  direction,
		Qty:        qty,
		Type:       "market",
		CreatedBar: createdBar,
	}
	om.orders = append(om.orders, order)
	return order
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
func (th *TradeHistory) CloseTrade(entryID string, exitPrice float64, exitBar int, exitTime int64) *Trade {
	for i, trade := range th.openTrades {
		if trade.EntryID == entryID {
			trade.ExitPrice = exitPrice
			trade.ExitBar = exitBar
			trade.ExitTime = exitTime
			
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
	initialized      bool
	currentBar       int
}

/* NewStrategy creates a new strategy */
func NewStrategy() *Strategy {
	return &Strategy{
		orderManager:     NewOrderManager(),
		positionTracker:  NewPositionTracker(),
		tradeHistory:     NewTradeHistory(),
		equityCalculator: NewEquityCalculator(10000),
		initialized:      false,
	}
}

/* Call initializes strategy with name and options */
func (s *Strategy) Call(strategyName string, initialCapital float64) {
	s.initialized = true
	s.equityCalculator = NewEquityCalculator(initialCapital)
}

/* Entry places an entry order */
func (s *Strategy) Entry(id, direction string, qty float64) error {
	if !s.initialized {
		return fmt.Errorf("strategy not initialized")
	}
	s.orderManager.CreateOrder(id, direction, qty, s.currentBar)
	return nil
}

/* Close closes position by entry ID */
func (s *Strategy) Close(id string, currentPrice float64, currentTime int64) {
	if !s.initialized {
		return
	}

	openTrades := s.tradeHistory.GetOpenTrades()
	for _, trade := range openTrades {
		if trade.EntryID == id {
			closedTrade := s.tradeHistory.CloseTrade(trade.EntryID, currentPrice, s.currentBar, currentTime)
			if closedTrade != nil {
				// Update position tracker
				oppositeDir := Long
				if trade.Direction == Long {
					oppositeDir = Short
				}
				s.positionTracker.UpdatePosition(trade.Size, currentPrice, oppositeDir)
				
				// Update equity
				s.equityCalculator.UpdateFromClosedTrade(*closedTrade)
			}
		}
	}
}

/* CloseAll closes all open positions */
func (s *Strategy) CloseAll(currentPrice float64, currentTime int64) {
	if !s.initialized {
		return
	}

	openTrades := s.tradeHistory.GetOpenTrades()
	for _, trade := range openTrades {
		closedTrade := s.tradeHistory.CloseTrade(trade.EntryID, currentPrice, s.currentBar, currentTime)
		if closedTrade != nil {
			// Update position tracker
			oppositeDir := Long
			if trade.Direction == Long {
				oppositeDir = Short
			}
			s.positionTracker.UpdatePosition(trade.Size, currentPrice, oppositeDir)
			
			// Update equity
			s.equityCalculator.UpdateFromClosedTrade(*closedTrade)
		}
	}
}

/* Exit exits with stop/limit orders (simplified - just closes) */
func (s *Strategy) Exit(id, fromEntry string, currentPrice float64, currentTime int64) {
	s.Close(fromEntry, currentPrice, currentTime)
}

/* OnBarUpdate processes pending orders at bar open */
func (s *Strategy) OnBarUpdate(currentBar int, openPrice float64, openTime int64) {
	if !s.initialized {
		return
	}
	
	s.currentBar = currentBar
	pendingOrders := s.orderManager.GetPendingOrders(currentBar)
	
	for _, order := range pendingOrders {
		// Update position
		s.positionTracker.UpdatePosition(order.Qty, openPrice, order.Direction)
		
		// Add to open trades
		s.tradeHistory.AddOpenTrade(Trade{
			EntryID:    order.ID,
			Direction:  order.Direction,
			Size:       order.Qty,
			EntryPrice: openPrice,
			EntryBar:   currentBar,
			EntryTime:  openTime,
		})
		
		// Remove order
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
		return 0 // Return 0 instead of NaN for simplicity
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

/* GetNetProfit returns realized profit */
func (s *Strategy) GetNetProfit() float64 {
	return s.equityCalculator.GetNetProfit()
}

/* Helper function */
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
