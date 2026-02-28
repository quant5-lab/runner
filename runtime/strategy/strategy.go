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

/* AllowedDirection constants — matches strategy.direction.* Pine namespace */
const (
	DirectionAll   = "all"
	DirectionLong  = "long"
	DirectionShort = "short"
)

/* OrderAction constants - TradingView order execution model */
const (
	OrderActionEntry    = "entry"
	OrderActionClose    = "close"
	OrderActionCloseAll = "close_all"
	OrderActionOrder    = "order" // strategy.order: net-adds to position, ignores pyramiding
)

/* Commission type constants - matches strategy.commission.* Pine namespace */
const (
	CommissionPercent         = "percent"
	CommissionCashPerOrder    = "cash_per_order"
	CommissionCashPerContract = "cash_per_contract"
)

/* Trade represents a single trade (open or closed) */
type Trade struct {
	EntryID      string  `json:"entryId"`
	ExitID       string  `json:"exitId"`
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
	MaxDrawdown  float64 `json:"maxDrawdown"`
	MaxRunup     float64 `json:"maxRunup"`
	Commission   float64 `json:"commission"`
}

/* Order represents a pending order - unified for entry and close operations */
type Order struct {
	ID           string
	ExitID       string // Semantic exit ID propagated to Trade.ExitID on fill
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
func (om *OrderManager) CreateCloseOrder(exitID, fromEntry string, createdBar int, comment string) Order {
	orderID := fmt.Sprintf("_close_%s_%d", fromEntry, createdBar)
	om.removeOrderByID(orderID)

	order := Order{
		ID:          orderID,
		ExitID:      exitID,
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

/* CreateNetOrder creates a pending strategy.order (net-position) order */
func (om *OrderManager) CreateNetOrder(id, direction string, qty float64, createdBar int, comment string) Order {
	om.removeOrderByID(id)

	order := Order{
		ID:           id,
		Action:       OrderActionOrder,
		Direction:    direction,
		Qty:          qty,
		Type:         "market",
		CreatedBar:   createdBar,
		EntryComment: comment,
	}
	om.orders = append(om.orders, order)
	return order
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

func (om *OrderManager) ClearAll() {
	om.orders = om.orders[:0]
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

func directionMultiplier(direction string) float64 {
	if direction == Short {
		return -1.0
	}
	return 1.0
}

func buildClosedTrade(open Trade, closeSize, profit, commission, metricsScale float64, exitID string, exitPrice float64, exitBar int, exitTime int64, exitComment string) Trade {
	return Trade{
		EntryID:      open.EntryID,
		ExitID:       exitID,
		Direction:    open.Direction,
		Size:         closeSize,
		EntryPrice:   open.EntryPrice,
		EntryBar:     open.EntryBar,
		EntryTime:    open.EntryTime,
		EntryComment: open.EntryComment,
		ExitPrice:    exitPrice,
		ExitBar:      exitBar,
		ExitTime:     exitTime,
		ExitComment:  exitComment,
		Profit:       profit,
		MaxDrawdown:  open.MaxDrawdown * metricsScale,
		MaxRunup:     open.MaxRunup * metricsScale,
		Commission:   commission,
	}
}

/* CloseTrade closes a trade by entry ID */
func (th *TradeHistory) CloseTrade(entryID, exitID string, exitPrice float64, exitBar int, exitTime int64, exitComment string, exitCommission float64) *Trade {
	for i, trade := range th.openTrades {
		if trade.EntryID == entryID {
			profit := (exitPrice - trade.EntryPrice) * trade.Size * directionMultiplier(trade.Direction)
			closed := buildClosedTrade(trade, trade.Size, profit, trade.Commission+exitCommission, 1.0, exitID, exitPrice, exitBar, exitTime, exitComment)
			th.closedTrades = append(th.closedTrades, closed)
			th.openTrades = append(th.openTrades[:i], th.openTrades[i+1:]...)
			return &th.closedTrades[len(th.closedTrades)-1]
		}
	}
	return nil
}

/*
	PartialCloseTrades closes up to qty units from open trades in direction (FIFO).

Returns (actual qty closed, newly closed trades). totalExitCommission is allocated proportionally.
*/
func (th *TradeHistory) PartialCloseTrades(direction, exitID string, qty, exitPrice float64, exitBar int, exitTime int64, exitComment string, totalExitCommission float64) (float64, []Trade) {
	remaining := qty
	var newlyClosed []Trade
	i := 0
	for i < len(th.openTrades) && remaining > 0 {
		trade := th.openTrades[i]
		if trade.Direction != direction {
			i++
			continue
		}

		multiplier := directionMultiplier(trade.Direction)

		if trade.Size <= remaining {
			exitCommission := totalExitCommission * (trade.Size / qty)
			profit := (exitPrice - trade.EntryPrice) * trade.Size * multiplier
			closed := buildClosedTrade(trade, trade.Size, profit, trade.Commission+exitCommission, 1.0, exitID, exitPrice, exitBar, exitTime, exitComment)
			remaining -= trade.Size
			th.closedTrades = append(th.closedTrades, closed)
			newlyClosed = append(newlyClosed, closed)
			th.openTrades = append(th.openTrades[:i], th.openTrades[i+1:]...)
		} else {
			entryProportion := remaining / trade.Size
			exitCommission := totalExitCommission * (remaining / qty)
			commission := trade.Commission*entryProportion + exitCommission
			profit := (exitPrice - trade.EntryPrice) * remaining * multiplier
			closed := buildClosedTrade(trade, remaining, profit, commission, entryProportion, exitID, exitPrice, exitBar, exitTime, exitComment)
			th.closedTrades = append(th.closedTrades, closed)
			newlyClosed = append(newlyClosed, closed)
			th.openTrades[i].Size -= remaining
			th.openTrades[i].Commission -= trade.Commission * entryProportion
			remaining = 0
		}
	}
	return qty - remaining, newlyClosed
}

/* GetOpenTrades returns open trades */
func (th *TradeHistory) GetOpenTrades() []Trade {
	return th.openTrades
}

/* GetClosedTrades returns closed trades */
func (th *TradeHistory) GetClosedTrades() []Trade {
	return th.closedTrades
}

/* UpdateOpenTradeMetrics updates MaxDrawdown and MaxRunup for all open trades based on bar data */
func (th *TradeHistory) UpdateOpenTradeMetrics(barHigh, barLow float64) {
	for i := range th.openTrades {
		trade := &th.openTrades[i]
		var favorable, adverse float64
		if trade.Direction == Long {
			favorable = (barHigh - trade.EntryPrice) * trade.Size
			adverse = (trade.EntryPrice - barLow) * trade.Size
		} else {
			favorable = (trade.EntryPrice - barLow) * trade.Size
			adverse = (barHigh - trade.EntryPrice) * trade.Size
		}
		if favorable < 0 {
			favorable = 0
		}
		if adverse < 0 {
			adverse = 0
		}
		if favorable > trade.MaxRunup {
			trade.MaxRunup = favorable
		}
		if adverse > trade.MaxDrawdown {
			trade.MaxDrawdown = adverse
		}
	}
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
	commissionValue  float64
	commissionType   string
	allowedDirection string
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

func (s *Strategy) SetCommission(value float64, commType string) {
	s.commissionValue = value
	s.commissionType = commType
}

/* SetAllowedDirection restricts entry direction; DirectionAll permits both */
func (s *Strategy) SetAllowedDirection(direction string) {
	s.allowedDirection = direction
}

func (s *Strategy) Cancel(id string) {
	s.orderManager.RemoveOrder(id)
}

func (s *Strategy) CancelAll() {
	s.orderManager.ClearAll()
}

func (s *Strategy) GetPositionEntryName() string {
	openTrades := s.tradeHistory.GetOpenTrades()
	if len(openTrades) == 0 {
		return ""
	}
	return openTrades[len(openTrades)-1].EntryID
}

/* calcCommission computes commission for one side of a trade (entry or exit) */
func (s *Strategy) calcCommission(qty, price float64) float64 {
	switch s.commissionType {
	case CommissionPercent:
		return qty * price * s.commissionValue / 100.0
	case CommissionCashPerOrder:
		return s.commissionValue
	case CommissionCashPerContract:
		return qty * s.commissionValue
	}
	return 0.0
}

func (s *Strategy) Entry(id, direction string, qty float64, comment string) error {
	if !s.initialized {
		return fmt.Errorf("strategy not initialized")
	}

	if s.allowedDirection != "" && s.allowedDirection != DirectionAll {
		if direction != s.allowedDirection {
			return nil
		}
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

/*
	Order places a net-position order: ignores pyramiding, nets arithmetically against current position.

Long adds to position; Short reduces it (and may cross zero into a short position).
*/
func (s *Strategy) Order(id, direction string, qty float64, comment string) error {
	if !s.initialized {
		return fmt.Errorf("strategy not initialized")
	}

	if s.allowedDirection != "" && s.allowedDirection != DirectionAll {
		if direction != s.allowedDirection {
			return nil
		}
	}

	s.orderManager.CreateNetOrder(id, direction, qty, s.currentBar, comment)
	return nil
}

/* executeNetOrder fills a strategy.order at fillPrice by netting the position arithmetically (FIFO). */
func (s *Strategy) executeNetOrder(id, direction string, qty, fillPrice float64, fillBar int, fillTime int64, comment string) {
	currentSize := s.positionTracker.GetPositionSize()

	dirSign := 1.0
	if direction == Short {
		dirSign = -1.0
	}
	delta := qty * dirSign
	newSize := currentSize + delta

	isReducing := (currentSize > 0 && delta < 0) || (currentSize < 0 && delta > 0)

	if !isReducing {
		s.openNetTrade(id, direction, qty, fillPrice, fillBar, fillTime, comment)
		return
	}

	closeQty := math.Min(math.Abs(currentSize), math.Abs(delta))

	// closeDirection is the direction of existing open trades, opposite of the incoming order
	closeDirection := Long
	if delta > 0 {
		closeDirection = Short
	}

	exitCommission := s.calcCommission(closeQty, fillPrice)
	_, newlyClosed := s.tradeHistory.PartialCloseTrades(closeDirection, id, closeQty, fillPrice, fillBar, fillTime, comment, exitCommission)
	if len(newlyClosed) > 0 {
		s.positionTracker.UpdatePosition(closeQty, fillPrice, direction)
		for _, t := range newlyClosed {
			s.equityCalculator.UpdateFromClosedTrade(t)
		}
	}

	if math.Abs(delta) > math.Abs(currentSize)+1e-9 {
		openQty := math.Abs(newSize)
		s.openNetTrade(id, direction, openQty, fillPrice, fillBar, fillTime, comment)
	}
}

func (s *Strategy) openNetTrade(id, direction string, qty, fillPrice float64, fillBar int, fillTime int64, comment string) {
	entryCommission := s.calcCommission(qty, fillPrice)
	s.positionTracker.UpdatePosition(qty, fillPrice, direction)
	s.tradeHistory.AddOpenTrade(Trade{
		EntryID:      id,
		Direction:    direction,
		Size:         qty,
		EntryPrice:   fillPrice,
		EntryBar:     fillBar,
		EntryTime:    fillTime,
		EntryComment: comment,
		Commission:   entryCommission,
	})
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
			// Per Pine semantics: strategy.close("buy") → exit_id = "buy"
			s.orderManager.CreateCloseOrder(id, id, s.currentBar, comment)
			return
		}
	}
}

/* executeCloseOrder executes a close order at given price - internal use only */
func (s *Strategy) executeCloseOrder(exitID, entryID string, fillPrice float64, fillBar int, fillTime int64, comment string) {
	openTrades := s.tradeHistory.GetOpenTrades()
	for _, trade := range openTrades {
		if trade.EntryID == entryID {
			exitCommission := s.calcCommission(trade.Size, fillPrice)
			closedTrade := s.tradeHistory.CloseTrade(trade.EntryID, exitID, fillPrice, fillBar, fillTime, comment, exitCommission)
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
		exitCommission := s.calcCommission(trade.Size, fillPrice)
		closedTrade := s.tradeHistory.CloseTrade(trade.EntryID, "", fillPrice, fillBar, fillTime, comment, exitCommission)
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
	if !s.initialized {
		return
	}
	openTrades := s.tradeHistory.GetOpenTrades()
	for _, trade := range openTrades {
		if trade.EntryID == fromEntry {
			// Per Pine semantics: strategy.exit("my_exit_id", "from_entry") → exit_id = "my_exit_id"
			s.orderManager.CreateCloseOrder(id, fromEntry, s.currentBar, comment)
			return
		}
	}
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
			s.executeCloseOrder(exitID, fromEntry, stopLevel, s.currentBar, barTime, comment)
			return
		}
		if trade.Direction == Short && barHigh >= stopLevel {
			s.executeCloseOrder(exitID, fromEntry, stopLevel, s.currentBar, barTime, comment)
			return
		}
	}

	// Check take profit (long: high >= limit, short: low <= limit)
	if !math.IsNaN(limitLevel) {
		if trade.Direction == Long && barHigh >= limitLevel {
			s.executeCloseOrder(exitID, fromEntry, limitLevel, s.currentBar, barTime, comment)
			return
		}
		if trade.Direction == Short && barLow <= limitLevel {
			s.executeCloseOrder(exitID, fromEntry, limitLevel, s.currentBar, barTime, comment)
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

			entryCommission := s.calcCommission(order.Qty, openPrice)
			s.tradeHistory.AddOpenTrade(Trade{
				EntryID:      order.ID,
				Direction:    order.Direction,
				Size:         order.Qty,
				EntryPrice:   openPrice,
				EntryBar:     currentBar,
				EntryTime:    openTime,
				EntryComment: order.EntryComment,
				Commission:   entryCommission,
			})

		case OrderActionClose:
			s.executeCloseOrder(order.ExitID, order.FromEntry, openPrice, currentBar, openTime, order.ExitComment)

		case OrderActionCloseAll:
			s.executeCloseAllOrder(openPrice, currentBar, openTime, order.ExitComment)

		case OrderActionOrder:
			s.executeNetOrder(order.ID, order.Direction, order.Qty, openPrice, currentBar, openTime, order.EntryComment)
		}

		s.orderManager.RemoveOrder(order.ID)
	}
}

/* OnBarMetrics updates per-trade MaxDrawdown and MaxRunup using bar high/low data */
func (s *Strategy) OnBarMetrics(barHigh, barLow float64) {
	if !s.initialized {
		return
	}
	s.tradeHistory.UpdateOpenTradeMetrics(barHigh, barLow)
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

func (s *Strategy) GetOpenProfit(currentPrice float64) float64 {
	return CalcOpenProfit(s.tradeHistory.GetOpenTrades(), currentPrice)
}

func (s *Strategy) GetOpenTradesCount() int {
	return len(s.tradeHistory.GetOpenTrades())
}

func (s *Strategy) GetAvgTrade() float64 {
	return AvgTrade(s.tradeHistory.GetClosedTrades())
}

func (s *Strategy) GetAvgWinningTrade() float64 {
	return AvgWinningTrade(s.tradeHistory.GetClosedTrades())
}

func (s *Strategy) GetAvgLosingTrade() float64 {
	return AvgLosingTrade(s.tradeHistory.GetClosedTrades())
}

/* Helper function */
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
