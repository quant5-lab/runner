package strategy

import (
	"fmt"
	"math"
)

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

type Order struct {
	ID            string
	ExitID        string // Semantic exit ID propagated to Trade.ExitID on fill
	Action        string // OrderActionEntry, OrderActionClose, OrderActionCloseAll
	Direction     string
	Qty           float64
	UseDefaultQty bool // qty deferred to fill time: computed after reversal close at fill price
	Type          string
	CreatedBar    int
	EntryComment  string
	FromEntry     string // Target entry ID for close orders
	ExitComment   string // Comment for close orders
}

type OrderManager struct {
	orders      []Order
	nextOrderID int
}

func NewOrderManager() *OrderManager {
	return &OrderManager{
		orders:      []Order{},
		nextOrderID: 1,
	}
}

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

func (om *OrderManager) CreateEntryOrderWithDefaultQty(id, direction string, createdBar int, comment string) Order {
	om.removeOrderByID(id)
	order := Order{
		ID:            id,
		Action:        OrderActionEntry,
		Direction:     direction,
		UseDefaultQty: true,
		Type:          "market",
		CreatedBar:    createdBar,
		EntryComment:  comment,
	}
	om.orders = append(om.orders, order)
	return order
}

/* legacy compatibility */
func (om *OrderManager) CreateOrder(id, direction string, qty float64, createdBar int, comment string) Order {
	return om.CreateEntryOrder(id, direction, qty, createdBar, comment)
}

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

func (om *OrderManager) GetPendingOrders(currentBar int) []Order {
	pending := []Order{}
	for _, order := range om.orders {
		if order.CreatedBar < currentBar {
			pending = append(pending, order)
		}
	}
	return pending
}

func (om *OrderManager) GetCurrentBarOrders(currentBar int) []Order {
	var current []Order
	for _, order := range om.orders {
		if order.CreatedBar == currentBar {
			current = append(current, order)
		}
	}
	return current
}

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

type PositionTracker struct {
	positionSize     float64
	positionAvgPrice float64
	totalCost        float64
}

func NewPositionTracker() *PositionTracker {
	return &PositionTracker{}
}

func (pt *PositionTracker) UpdatePosition(qty, price float64, direction string) {
	sizeChange := qty
	if direction == Short {
		sizeChange = -qty
	}

	if (pt.positionSize > 0 && sizeChange < 0) || (pt.positionSize < 0 && sizeChange > 0) {
		pt.positionSize += sizeChange
		if pt.positionSize == 0 {
			pt.positionAvgPrice = 0
			pt.totalCost = 0
		} else {
			pt.totalCost = pt.positionAvgPrice * abs(pt.positionSize)
		}
	} else {
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

func (pt *PositionTracker) GetPositionSize() float64 {
	return pt.positionSize
}

func (pt *PositionTracker) GetAvgPrice() float64 {
	return pt.positionAvgPrice
}

type TradeHistory struct {
	openTrades   []Trade
	closedTrades []Trade
}

func NewTradeHistory() *TradeHistory {
	return &TradeHistory{
		openTrades:   []Trade{},
		closedTrades: []Trade{},
	}
}

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
		Profit:       profit - commission,
		MaxDrawdown:  open.MaxDrawdown * metricsScale,
		MaxRunup:     open.MaxRunup * metricsScale,
		Commission:   commission,
	}
}

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

/* totalExitCommission is allocated proportionally across closed lots. */
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

func (th *TradeHistory) GetOpenTrades() []Trade {
	return th.openTrades
}

func (th *TradeHistory) MatchingOpenTrades(entryID string) []Trade {
	var matching []Trade
	for _, trade := range th.openTrades {
		if trade.EntryID == entryID {
			matching = append(matching, trade)
		}
	}
	return matching
}

func (th *TradeHistory) AllOpenTradesSnapshot() []Trade {
	return append([]Trade{}, th.openTrades...)
}

func (th *TradeHistory) GetClosedTrades() []Trade {
	return th.closedTrades
}

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

type EquityCalculator struct {
	initialCapital float64
	realizedProfit float64
}

func NewEquityCalculator(initialCapital float64) *EquityCalculator {
	return &EquityCalculator{
		initialCapital: initialCapital,
		realizedProfit: 0,
	}
}

func (ec *EquityCalculator) UpdateFromClosedTrade(trade Trade) {
	ec.realizedProfit += trade.Profit
}

func (ec *EquityCalculator) GetEquity(unrealizedProfit float64) float64 {
	return ec.initialCapital + ec.realizedProfit + unrealizedProfit
}

func (ec *EquityCalculator) GetNetProfit() float64 {
	return ec.realizedProfit
}

type Strategy struct {
	context              interface{} // Context with OHLCV data
	orderManager         *OrderManager
	positionTracker      *PositionTracker
	tradeHistory         *TradeHistory
	equityCalculator     *EquityCalculator
	reversalHandler      *PositionReversalHandler
	defaultQtyCalc       *DefaultQtyCalculator
	currencyConverter    *CurrencyConverter
	pendingExitManager   *PendingExitManager
	initialized          bool
	currentBar           int
	currentPrice         float64
	currentBarTime       int64
	pyramiding           int
	commissionValue      float64
	commissionType       string
	defaultQtyValue      float64
	defaultQtyType       string
	qtyStep              float64
	allowedDirection     string
	processOrdersOnClose bool
}

func NewStrategy() *Strategy {
	om := NewOrderManager()
	pt := NewPositionTracker()
	th := NewTradeHistory()
	ec := NewEquityCalculator(10000)
	rh := NewPositionReversalHandler(th, pt, ec)

	return &Strategy{
		orderManager:       om,
		positionTracker:    pt,
		tradeHistory:       th,
		equityCalculator:   ec,
		reversalHandler:    rh,
		defaultQtyCalc:     NewDefaultQtyCalculator(),
		currencyConverter:  NewCurrencyConverter(),
		pendingExitManager: NewPendingExitManager(),
		initialized:        false,
		pyramiding:         -1,
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
	s.reversalHandler.commissionCalc = s.calcCommission
}

func (s *Strategy) SetDefaultQty(value float64, qtyType string) {
	s.defaultQtyValue = value
	s.defaultQtyType = qtyType
}

func (s *Strategy) SetQtyStep(step float64) {
	s.qtyStep = step
}

func (s *Strategy) applyQtyStep(qty float64) float64 {
	return floorToStep(qty, s.qtyStep)
}

func (s *Strategy) DefaultEntryQty(fillPrice float64) float64 {
	return s.defaultQtyCalc.CalculateQty(s.defaultQtyType, s.defaultQtyValue, fillPrice, s.Equity())
}

func (s *Strategy) ConvertToAccount(value float64) float64 {
	return s.currencyConverter.ToAccount(value)
}

func (s *Strategy) ConvertToSymbol(value float64) float64 {
	return s.currencyConverter.ToSymbol(value)
}

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
	return s.scheduleEntry(id, direction, qty, false, comment)
}

func (s *Strategy) EntryWithDefaultQty(id, direction, comment string) error {
	return s.scheduleEntry(id, direction, 0, true, comment)
}

func (s *Strategy) scheduleEntry(id, direction string, qty float64, useDefaultQty bool, comment string) error {
	if !s.initialized {
		return fmt.Errorf("strategy not initialized")
	}

	if s.allowedDirection != "" && s.allowedDirection != DirectionAll {
		if direction != s.allowedDirection {
			// TV semantics: a blocked entry still closes any open opposite-direction
			// positions — the reversal "close" half of `strategy.entry` runs even
			// when the "open" half is disabled by `strategy.risk.allow_entry_in`.
			s.scheduleDirectionCloseForBlockedEntry(id, direction, comment)
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

		// TV semantics: pyramiding=N caps simultaneous same-direction entries at N
		// (the count must be strictly less than the cap for a new entry to land).
		// Pine treats pyramiding=0 as the default that still permits one entry, so
		// the effective limit is max(1, pyramiding).
		limit := s.pyramiding
		if limit < 1 {
			limit = 1
		}
		if sameDirectionCount >= limit {
			return nil
		}
	}

	if useDefaultQty {
		s.orderManager.CreateEntryOrderWithDefaultQty(id, direction, s.currentBar, comment)
	} else {
		s.orderManager.CreateOrder(id, direction, qty, s.currentBar, comment)
	}
	return nil
}

/* ignores pyramiding; may cross zero from long to short or vice-versa */
func (s *Strategy) Order(id, direction string, qty float64, comment string) error {
	if !s.initialized {
		return fmt.Errorf("strategy not initialized")
	}

	if s.allowedDirection != "" && s.allowedDirection != DirectionAll {
		if direction != s.allowedDirection {
			s.scheduleDirectionCloseForBlockedEntry(id, direction, comment)
			return nil
		}
	}

	s.orderManager.CreateNetOrder(id, direction, qty, s.currentBar, comment)
	return nil
}

// scheduleDirectionCloseForBlockedEntry schedules a close order against every
// currently-open trade in the direction OPPOSITE to the blocked entry. Mirrors
// the "close existing opposite + open new same" reversal flow of strategy.entry
// but skips the open half when the configured allowedDirection blocks it.
// Idempotent per-bar — CreateCloseOrder dedupes by id.
func (s *Strategy) scheduleDirectionCloseForBlockedEntry(blockedEntryID, blockedDirection, comment string) {
	opposite := Long
	if blockedDirection == Long {
		opposite = Short
	}
	for _, trade := range s.tradeHistory.GetOpenTrades() {
		if trade.Direction == opposite {
			s.orderManager.CreateCloseOrder(blockedEntryID, trade.EntryID, s.currentBar, comment)
		}
	}
}

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

/* fills at next bar open — TV pending-order model */
func (s *Strategy) Close(id string, currentPrice float64, currentTime int64, comment string) {
	if !s.initialized {
		return
	}

	openTrades := s.tradeHistory.GetOpenTrades()
	for _, trade := range openTrades {
		if trade.EntryID == id {
			// Per Pine semantics: strategy.close("buy") → exit_id = "buy"
			s.orderManager.CreateCloseOrder(id, id, s.currentBar, comment)
			return
		}
	}
}

func (s *Strategy) closeTrades(trades []Trade, exitID string, fillPrice float64, fillBar int, fillTime int64, comment string) {
	for _, trade := range trades {
		exitCommission := s.calcCommission(trade.Size, fillPrice)
		closedTrade := s.tradeHistory.CloseTrade(trade.EntryID, exitID, fillPrice, fillBar, fillTime, comment, exitCommission)
		if closedTrade != nil {
			oppositeDir := Long
			if trade.Direction == Long {
				oppositeDir = Short
			}
			s.positionTracker.UpdatePosition(trade.Size, fillPrice, oppositeDir)
			s.equityCalculator.UpdateFromClosedTrade(*closedTrade)
		}
	}
}

func (s *Strategy) executeCloseOrder(exitID, entryID string, fillPrice float64, fillBar int, fillTime int64, comment string) {
	s.closeTrades(s.tradeHistory.MatchingOpenTrades(entryID), exitID, fillPrice, fillBar, fillTime, comment)
	s.pendingExitManager.RemoveAllExitsForEntry(entryID)
}

/*
	TV broker emulator: pending exits are cancelled immediately so OnBarMetrics cannot fill them

on the same bar as a close_all call. Fills at next bar open.
*/
func (s *Strategy) CloseAll(currentPrice float64, currentTime int64, comment string) {
	if !s.initialized {
		return
	}

	openTrades := s.tradeHistory.GetOpenTrades()
	if len(openTrades) > 0 {
		for _, trade := range openTrades {
			s.pendingExitManager.RemoveAllExitsForEntry(trade.EntryID)
		}
		s.orderManager.CreateCloseAllOrder(s.currentBar, comment)
	}
}

func (s *Strategy) executeCloseAllOrder(fillPrice float64, fillBar int, fillTime int64, comment string) {
	for _, trade := range s.tradeHistory.AllOpenTradesSnapshot() {
		s.pendingExitManager.RemoveAllExitsForEntry(trade.EntryID)
	}
	s.closeTrades(s.tradeHistory.AllOpenTradesSnapshot(), "", fillPrice, fillBar, fillTime, comment)
}

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

/* declarative per Pine: re-calls update levels but do not reset the eligibility gate */
func (s *Strategy) ExitWithLevels(exitID, fromEntry string, stopLevel, limitLevel, barHigh, barLow, barClose float64, barTime int64, comment string) {
	if !s.initialized {
		return
	}
	s.pendingExitManager.RegisterExit(exitID, fromEntry, stopLevel, limitLevel, s.currentBar, comment)
}

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

			qty := order.Qty
			if order.UseDefaultQty {
				qty = s.DefaultEntryQty(openPrice)
			}
			qty = s.applyQtyStep(qty)

			s.positionTracker.UpdatePosition(qty, openPrice, order.Direction)

			entryCommission := s.calcCommission(qty, openPrice)
			s.tradeHistory.AddOpenTrade(Trade{
				EntryID:      order.ID,
				Direction:    order.Direction,
				Size:         qty,
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
			s.executeNetOrder(order.ID, order.Direction, s.applyQtyStep(order.Qty), openPrice, currentBar, openTime, order.EntryComment)
		}

		s.orderManager.RemoveOrder(order.ID)
	}
}

func (s *Strategy) SetProcessOrdersOnClose(v bool) {
	s.processOrdersOnClose = v
}

// OnBarClose must be called after all bar-N signals and before advancing to bar N+1.
func (s *Strategy) OnBarClose(closePrice float64, closeTime int64) {
	if !s.initialized || !s.processOrdersOnClose {
		return
	}
	pendingOrders := s.orderManager.GetCurrentBarOrders(s.currentBar)
	for _, order := range pendingOrders {
		switch order.Action {
		case OrderActionEntry:
			s.reversalHandler.HandleReversal(order.Direction, closePrice, s.currentBar, closeTime)
			qty := order.Qty
			if order.UseDefaultQty {
				qty = s.DefaultEntryQty(closePrice)
			}
			qty = s.applyQtyStep(qty)
			s.positionTracker.UpdatePosition(qty, closePrice, order.Direction)
			entryCommission := s.calcCommission(qty, closePrice)
			s.tradeHistory.AddOpenTrade(Trade{
				EntryID:      order.ID,
				Direction:    order.Direction,
				Size:         qty,
				EntryPrice:   closePrice,
				EntryBar:     s.currentBar,
				EntryTime:    s.currentBarTime,
				EntryComment: order.EntryComment,
				Commission:   entryCommission,
			})
		case OrderActionClose:
			s.executeCloseOrder(order.ExitID, order.FromEntry, closePrice, s.currentBar, closeTime, order.ExitComment)
		case OrderActionCloseAll:
			s.executeCloseAllOrder(closePrice, s.currentBar, closeTime, order.ExitComment)
		case OrderActionOrder:
			s.executeNetOrder(order.ID, order.Direction, s.applyQtyStep(order.Qty), closePrice, s.currentBar, closeTime, order.EntryComment)
		}
		s.orderManager.RemoveOrder(order.ID)
	}
}

func (s *Strategy) OnBarMetrics(barOpen, barHigh, barLow float64, barTime int64) {
	if !s.initialized {
		return
	}
	s.currentBarTime = barTime
	s.tradeHistory.UpdateOpenTradeMetrics(barHigh, barLow)
	s.checkAndFillPendingExits(barOpen, barHigh, barLow, barTime)
}

func (s *Strategy) checkAndFillPendingExits(barOpen, barHigh, barLow float64, barTime int64) {
	for _, trade := range s.tradeHistory.GetOpenTrades() {
		for _, exit := range s.pendingExitManager.GetExitsForEntry(trade.EntryID) {
			triggered, fillPrice, _ := s.pendingExitManager.CheckExitTriggered(exit, trade, s.currentBar, barOpen, barHigh, barLow)
			if triggered {
				s.executeCloseOrder(exit.ExitID, trade.EntryID, fillPrice, s.currentBar, barTime, exit.Comment)
				s.pendingExitManager.RemoveAllExitsForEntry(trade.EntryID)
				break
			}
		}
	}
}

func (s *Strategy) GetPositionSize() float64 {
	return s.positionTracker.GetPositionSize()
}

func (s *Strategy) GetPositionAvgPrice() float64 {
	avgPrice := s.positionTracker.GetAvgPrice()
	if avgPrice == 0 {
		return math.NaN()
	}
	return avgPrice
}

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

func (s *Strategy) Equity() float64 {
	return s.GetEquity(s.currentPrice)
}

func (s *Strategy) GetNetProfit() float64 {
	return s.equityCalculator.GetNetProfit()
}

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

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
