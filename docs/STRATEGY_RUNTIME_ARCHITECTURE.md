# Strategy Runtime Architecture

## Overview
The Go strategy runtime follows **SOLID**, **DRY**, and **KISS** principles, aligning with the PineTS reference implementation while optimizing for Go's performance and type safety.

## Design Pattern: Separation of Concerns

The runtime is decomposed into 4 specialized components plus 1 orchestrator:

```
Strategy (Orchestrator)
├── OrderManager      (Order lifecycle)
├── PositionTracker   (Position state)
├── TradeHistory      (Trade records)
└── EquityCalculator  (P&L tracking)
```

This matches the PineTS pattern exactly:
- **PineTS**: `OrderManager.class.ts`, `PositionTracker.class.ts`, `TradeHistory.class.ts`, `EquityCalculator.class.ts`
- **Go Port**: Same 4 classes embedded in `strategy.go`

## Component Responsibilities

### 1. OrderManager
**Purpose**: Manages pending orders before execution.

**Operations**:
- `CreateOrder(id, direction, qty, createdBar)` - Place or replace order
- `GetPendingOrders(currentBar)` - Get orders ready to execute
- `RemoveOrder(id)` - Cancel order

**State**:
- `orders []Order` - Pending orders awaiting execution
- `nextOrderID int` - Auto-increment ID generator

**Logic**:
```go
// Market orders execute on next bar open (bar after creation)
func (om *OrderManager) GetPendingOrders(currentBar int) []Order {
    pending := []Order{}
    for _, order := range om.orders {
        if order.CreatedBar < currentBar {
            pending = append(pending, order)
        }
    }
    return pending
}
```

### 2. PositionTracker
**Purpose**: Tracks current net position and average entry price.

**Operations**:
- `UpdatePosition(qty, price, direction)` - Adjust position from trade
- `GetPositionSize()` - Current position (positive=long, negative=short)
- `GetAvgPrice()` - Average entry price of open position

**State**:
- `positionSize float64` - Net position (long/short)
- `positionAvgPrice float64` - Weighted average entry price
- `totalCost float64` - Total capital invested in position

**Logic**:
```go
// Average price remains constant when reducing position
// Average price recalculates when adding to position
func (pt *PositionTracker) UpdatePosition(qty, price float64, direction string) {
    sizeChange := qty
    if direction == Short {
        sizeChange = -qty
    }
    
    if (pt.positionSize > 0 && sizeChange < 0) || (pt.positionSize < 0 && sizeChange > 0) {
        // Closing - avg price stays constant
        pt.positionSize += sizeChange
        if pt.positionSize == 0 {
            pt.positionAvgPrice = 0
            pt.totalCost = 0
        } else {
            pt.totalCost = pt.positionAvgPrice * abs(pt.positionSize)
        }
    } else {
        // Opening/adding - recalculate avg price
        addedCost := qty * price
        pt.totalCost += addedCost
        pt.positionSize += sizeChange
        
        if pt.positionSize != 0 {
            pt.positionAvgPrice = pt.totalCost / abs(pt.positionSize)
        }
    }
}
```

### 3. TradeHistory
**Purpose**: Records open and closed trades for reporting.

**Operations**:
- `AddOpenTrade(trade)` - Record new trade
- `CloseTrade(entryID, exitPrice, exitBar, exitTime)` - Close trade and calculate P&L
- `GetOpenTrades()` - Current open positions
- `GetClosedTrades()` - Historical closed trades

**State**:
- `openTrades []Trade` - Active trades
- `closedTrades []Trade` - Completed trades with P&L

**Logic**:
```go
func (th *TradeHistory) CloseTrade(entryID string, exitPrice float64, exitBar int, exitTime int64) *Trade {
    idx := th.openTrades.findIndex(t => t.entryID === entryID)
    if idx == -1 return nil
    
    trade := th.openTrades[idx]
    trade.exitPrice = exitPrice
    trade.exitBar = exitBar
    trade.exitTime = exitTime
    
    // Calculate realized P&L
    priceDiff := exitPrice - trade.entryPrice
    multiplier := 1.0
    if trade.direction == Short {
        multiplier = -1.0
    }
    trade.profit = priceDiff * trade.size * multiplier
    
    th.closedTrades = append(th.closedTrades, trade)
    th.openTrades = remove(th.openTrades, idx)
    return &trade
}
```

### 4. EquityCalculator
**Purpose**: Calculates account equity and net profit.

**Operations**:
- `UpdateFromClosedTrade(trade)` - Add realized P&L
- `GetEquity(unrealizedProfit)` - Total equity (realized + unrealized)
- `GetNetProfit()` - Realized P&L only

**State**:
- `initialCapital float64` - Starting account balance
- `realizedProfit float64` - Sum of closed trade P&L

**Logic**:
```go
func (ec *EquityCalculator) GetEquity(unrealizedProfit float64) float64 {
    return ec.initialCapital + ec.realizedProfit + unrealizedProfit
}

func (ec *EquityCalculator) UpdateFromClosedTrade(trade Trade) {
    ec.realizedProfit += trade.profit
}
```

### 5. Strategy (Orchestrator)
**Purpose**: Coordinates all components and provides Pine Script API.

**Operations**:
- `Call(name, initialCapital)` - Initialize strategy
- `Entry(id, direction, qty)` - Place entry order
- `Close(id, currentPrice, currentTime)` - Close position by ID
- `CloseAll(currentPrice, currentTime)` - Close all positions
- `OnBarUpdate(bar, openPrice, openTime)` - Execute pending orders at bar open

**State** (delegates to components):
- `orderManager *OrderManager`
- `positionTracker *PositionTracker`
- `tradeHistory *TradeHistory`
- `equityCalculator *EquityCalculator`

**Coordination Logic**:
```go
func (s *Strategy) OnBarUpdate(currentBar int, openPrice float64, openTime int64) {
    s.currentBar = currentBar
    pendingOrders := s.orderManager.GetPendingOrders(currentBar)
    
    for _, order := range pendingOrders {
        // Update position tracker
        s.positionTracker.UpdatePosition(order.Qty, openPrice, order.Direction)
        
        // Record trade
        s.tradeHistory.AddOpenTrade(Trade{
            EntryID:    order.ID,
            Direction:  order.Direction,
            Size:       order.Qty,
            EntryPrice: openPrice,
            EntryBar:   currentBar,
            EntryTime:  openTime,
        })
        
        // Remove executed order
        s.orderManager.RemoveOrder(order.ID)
    }
}

func (s *Strategy) Close(id string, currentPrice float64, currentTime int64) {
    openTrades := s.tradeHistory.GetOpenTrades()
    for _, trade := range openTrades {
        if trade.EntryID == id {
            // Close trade in history
            closedTrade := s.tradeHistory.CloseTrade(trade.EntryID, currentPrice, s.currentBar, currentTime)
            
            if closedTrade != nil {
                // Update position (opposite direction)
                oppositeDir := Long
                if trade.Direction == Long {
                    oppositeDir = Short
                }
                s.positionTracker.UpdatePosition(trade.Size, currentPrice, oppositeDir)
                
                // Update equity with realized P&L
                s.equityCalculator.UpdateFromClosedTrade(*closedTrade)
            }
        }
    }
}
```

## SOLID Principles Applied

### Single Responsibility Principle (SRP)
Each component has one clear purpose:
- OrderManager: ONLY manages order lifecycle
- PositionTracker: ONLY tracks position state
- TradeHistory: ONLY records trades
- EquityCalculator: ONLY computes P&L
- Strategy: ONLY coordinates components

### Open/Closed Principle (OCP)
Components can be extended without modification:
- Add new order types (limit, stop) in OrderManager
- Add position analytics in PositionTracker
- Add trade metrics in TradeHistory
- Add performance metrics in EquityCalculator

### Liskov Substitution Principle (LSP)
Components can be mocked/replaced for testing:
```go
type IOrderManager interface {
    CreateOrder(id, direction string, qty float64, createdBar int) Order
    GetPendingOrders(currentBar int) []Order
}

type MockOrderManager struct {}
func (m *MockOrderManager) CreateOrder(...) Order { return mockOrder }
```

### Interface Segregation Principle (ISP)
Each component exposes minimal API:
- PositionTracker: Only 3 methods (Update, GetSize, GetAvgPrice)
- EquityCalculator: Only 3 methods (Update, GetEquity, GetNetProfit)

### Dependency Inversion Principle (DIP)
Strategy depends on abstractions (component interfaces), not concrete implementations.

## DRY (Don't Repeat Yourself)

**Position calculation logic** exists ONLY in PositionTracker:
- Not duplicated in Strategy.Entry()
- Not duplicated in Strategy.Close()
- Single source of truth for avg price calculation

**P&L calculation logic** exists ONLY in TradeHistory.CloseTrade():
- Not duplicated in EquityCalculator
- Not duplicated in chart output generation

## KISS (Keep It Simple, Stupid)

### Simple Data Flow
```
Pine Script Call → Strategy Method → Component Update → State Change
```

### No Hidden State
- All state visible in component structs
- No global variables
- No singletons

### Explicit Dependencies
```go
func NewStrategy() *Strategy {
    return &Strategy{
        orderManager:     NewOrderManager(),
        positionTracker:  NewPositionTracker(),
        tradeHistory:     NewTradeHistory(),
        equityCalculator: NewEquityCalculator(10000),
    }
}
```

## Comparison: PineTS vs Go Port

| Aspect | PineTS | Go Port |
|--------|--------|---------|
| **Language** | TypeScript | Go |
| **Pattern** | 4 separate classes | 4 embedded components |
| **Instantiation** | `new PositionTracker()` | `NewPositionTracker()` |
| **Encapsulation** | `private` fields | Unexported fields |
| **State** | Class properties | Struct fields |
| **Methods** | `this.updatePosition()` | `pt.UpdatePosition()` |
| **Error Handling** | Exceptions | Error returns |

### Code Alignment Example

**PineTS**:
```typescript
export class PositionTracker {
    private positionSize: number = 0;
    private positionAvgPrice: number = 0;
    
    updatePosition(qty: number, price: number, direction: 'long' | 'short'): void {
        const sizeChange = qty * (direction === 'long' ? 1 : -1);
        // ... position logic
    }
}
```

**Go Port**:
```go
type PositionTracker struct {
    positionSize     float64
    positionAvgPrice float64
}

func (pt *PositionTracker) UpdatePosition(qty, price float64, direction string) {
    sizeChange := qty
    if direction == Short {
        sizeChange = -qty
    }
    // ... position logic
}
```

## Testing Strategy

### Unit Tests
Each component tested independently:
```go
func TestPositionTracker_UpdatePosition(t *testing.T) {
    pt := NewPositionTracker()
    pt.UpdatePosition(10, 100, Long)
    assert.Equal(t, 10.0, pt.GetPositionSize())
    assert.Equal(t, 100.0, pt.GetAvgPrice())
}
```

### Integration Tests
Full strategy execution:
```go
func TestStrategyExecution(t *testing.T) {
    strat := NewStrategy()
    strat.Entry("long1", Long, 10)
    strat.OnBarUpdate(1, 100, 1000)
    strat.Close("long1", 110, 2000)
    
    assert.Equal(t, 100.0, strat.GetNetProfit())
}
```

## Performance Optimizations

### Memory Efficiency
- **Slices not arrays**: Dynamic growth without reallocation
- **Pointer receivers**: Avoid struct copying
- **Pre-allocated capacity**: `make([]Trade, 0, 100)`

### CPU Efficiency
- **O(1) position updates**: No loops in PositionTracker
- **O(n) trade closure**: Linear search through open trades
- **Zero allocation**: Reuse trade structs when possible

## Future Extensibility

### Additional Components (Candidates)
- **RiskManager**: Position sizing, max drawdown limits
- **PerformanceAnalyzer**: Sharpe ratio, win rate, profit factor
- **OrderValidator**: Margin checks, position limits

### Strategy Features (Pending)
- **Limit Orders**: OrderManager extension
- **Stop Orders**: OrderManager extension
- **Bracket Orders**: Multi-leg order coordination
- **Position Pyramiding**: Multiple entries to same position

## Code Generation Integration

The Go codegen generates strategy calls:

```go
// Generated from Pine Script:
// strategy.entry("Long", strategy.long, 10)
strat.Entry("Long", strategy.Long, 10)

// Generated from Pine Script:
// strategy.close("Long")
strat.Close("Long", bar.Close, bar.Time)

// Generated from Pine Script:
// strategy.close_all()
strat.CloseAll(bar.Close, bar.Time)
```

## References

- **Implementation**: `golang-port/runtime/strategy/strategy.go`
- **Tests**: `golang-port/runtime/strategy/strategy_test.go`
- **PineTS Reference**: Attached files (OrderManager, PositionTracker, TradeHistory, EquityCalculator)
- **Codegen**: `golang-port/codegen/generator.go` (strategy.close support)
