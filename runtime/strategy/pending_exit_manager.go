package strategy

import "math"

/*
ExitOrder represents a pending exit order for stop-loss or take-profit.

Design: Value object - immutable state per exit order
Lifecycle: Created by ExitWithLevels(), checked each bar, removed when filled
*/
type ExitOrder struct {
	ExitID     string
	FromEntry  string
	StopLevel  float64
	LimitLevel float64
	Comment    string
	CreatedBar int
}

/*
PendingExitManager manages pending exit orders per PineScript semantics.

Single Responsibility: Store and match exit orders to open trades
Design: Stateless query interface - holds exit orders until filled or cancelled
PineScript Semantics: strategy.exit() registers orders, not immediate execution
*/
type PendingExitManager struct {
	exitOrders []ExitOrder
}

/*
NewPendingExitManager creates exit order manager.
*/
func NewPendingExitManager() *PendingExitManager {
	return &PendingExitManager{
		exitOrders: []ExitOrder{},
	}
}

/*
RegisterExit registers or updates an exit order.

PineScript: strategy.exit() called each bar updates exit levels
Behavior: Replaces existing exit with same exitID + fromEntry
*/
func (pem *PendingExitManager) RegisterExit(exitID, fromEntry string, stopLevel, limitLevel float64, createdBar int, comment string) {
	// Remove existing exit with same exitID + fromEntry
	pem.RemoveExit(exitID, fromEntry)

	pem.exitOrders = append(pem.exitOrders, ExitOrder{
		ExitID:     exitID,
		FromEntry:  fromEntry,
		StopLevel:  stopLevel,
		LimitLevel: limitLevel,
		Comment:    comment,
		CreatedBar: createdBar,
	})
}

/*
GetExitsForEntry returns all pending exits for specific entry ID.
*/
func (pem *PendingExitManager) GetExitsForEntry(entryID string) []ExitOrder {
	exits := []ExitOrder{}
	for _, exit := range pem.exitOrders {
		if exit.FromEntry == "" || exit.FromEntry == entryID {
			exits = append(exits, exit)
		}
	}
	return exits
}

/*
RemoveExit removes exit by exitID and fromEntry.
*/
func (pem *PendingExitManager) RemoveExit(exitID, fromEntry string) {
	filtered := []ExitOrder{}
	for _, exit := range pem.exitOrders {
		if !(exit.ExitID == exitID && exit.FromEntry == fromEntry) {
			filtered = append(filtered, exit)
		}
	}
	pem.exitOrders = filtered
}

/*
RemoveAllExitsForEntry removes all exits targeting specific entry.
*/
func (pem *PendingExitManager) RemoveAllExitsForEntry(entryID string) {
	filtered := []ExitOrder{}
	for _, exit := range pem.exitOrders {
		if exit.FromEntry != entryID && exit.FromEntry != "" {
			filtered = append(filtered, exit)
		}
	}
	pem.exitOrders = filtered
}

/*
CheckExitTriggered checks if exit order should trigger on current bar.

Returns: (triggered, exitPrice, exitType)
- triggered: true if stop or limit hit
- exitPrice: price at which exit executes
- exitType: "stop" or "limit"

PineScript Semantics:
- Stop: Triggers when price reaches stopLevel or worse (unfavorable)
- Limit: Triggers when price reaches limitLevel or better (favorable)
- Long: stop=low check, limit=high check
- Short: stop=high check, limit=low check
*/
func (pem *PendingExitManager) CheckExitTriggered(exit ExitOrder, trade Trade, barHigh, barLow float64) (bool, float64, string) {
	isLong := trade.Direction == Long

	// Check stop loss (unfavorable direction)
	if !math.IsNaN(exit.StopLevel) {
		if isLong && barLow <= exit.StopLevel {
			return true, exit.StopLevel, "stop"
		}
		if !isLong && barHigh >= exit.StopLevel {
			return true, exit.StopLevel, "stop"
		}
	}

	// Check take profit (favorable direction)
	if !math.IsNaN(exit.LimitLevel) {
		if isLong && barHigh >= exit.LimitLevel {
			return true, exit.LimitLevel, "limit"
		}
		if !isLong && barLow <= exit.LimitLevel {
			return true, exit.LimitLevel, "limit"
		}
	}

	return false, 0, ""
}
