package strategy

import "math"

/*
	FirstRegisteredBar is preserved across level updates so the eligibility gate

is never reset by a declarative re-call of strategy.exit().
*/
type ExitOrder struct {
	ExitID             string
	FromEntry          string
	StopLevel          float64
	LimitLevel         float64
	Comment            string
	FirstRegisteredBar int
}

// Pine default: a stop/limit placed during bar N fills no earlier than bar N+1
// (calc_on_every_tick=false, process_orders_on_close=false)
func (e ExitOrder) IsEligible(currentBar int) bool {
	return currentBar > e.FirstRegisteredBar
}

type PendingExitManager struct {
	exitOrders []ExitOrder
}

func NewPendingExitManager() *PendingExitManager {
	return &PendingExitManager{exitOrders: []ExitOrder{}}
}

func (pem *PendingExitManager) RegisterExit(exitID, fromEntry string, stopLevel, limitLevel float64, currentBar int, comment string) {
	firstBar := pem.firstRegisteredBarFor(exitID, fromEntry, currentBar)
	pem.RemoveExit(exitID, fromEntry)
	pem.exitOrders = append(pem.exitOrders, ExitOrder{
		ExitID:             exitID,
		FromEntry:          fromEntry,
		StopLevel:          stopLevel,
		LimitLevel:         limitLevel,
		Comment:            comment,
		FirstRegisteredBar: firstBar,
	})
}

func (pem *PendingExitManager) firstRegisteredBarFor(exitID, fromEntry string, fallback int) int {
	for _, e := range pem.exitOrders {
		if e.ExitID == exitID && e.FromEntry == fromEntry {
			return e.FirstRegisteredBar
		}
	}
	return fallback
}

func (pem *PendingExitManager) GetExitsForEntry(entryID string) []ExitOrder {
	exits := []ExitOrder{}
	for _, e := range pem.exitOrders {
		if e.FromEntry == "" || e.FromEntry == entryID {
			exits = append(exits, e)
		}
	}
	return exits
}

func (pem *PendingExitManager) RemoveExit(exitID, fromEntry string) {
	filtered := pem.exitOrders[:0]
	for _, e := range pem.exitOrders {
		if !(e.ExitID == exitID && e.FromEntry == fromEntry) {
			filtered = append(filtered, e)
		}
	}
	pem.exitOrders = filtered
}

func (pem *PendingExitManager) RemoveAllExitsForEntry(entryID string) {
	filtered := pem.exitOrders[:0]
	for _, e := range pem.exitOrders {
		if e.FromEntry != entryID && e.FromEntry != "" {
			filtered = append(filtered, e)
		}
	}
	pem.exitOrders = filtered
}

func (pem *PendingExitManager) CheckExitTriggered(exit ExitOrder, trade Trade, currentBar int, barOpen, barHigh, barLow float64) (bool, float64, string) {
	if !exit.IsEligible(currentBar) {
		return false, 0, ""
	}
	return checkPriceTrigger(exit, trade, barOpen, barHigh, barLow)
}

func checkPriceTrigger(exit ExitOrder, trade Trade, barOpen, barHigh, barLow float64) (bool, float64, string) {
	isLong := trade.Direction == Long
	stopHit := isStopBreached(exit.StopLevel, isLong, barHigh, barLow)
	limitHit := isLimitBreached(exit.LimitLevel, isLong, barHigh, barLow)

	switch {
	case !stopHit && !limitHit:
		return false, 0, ""
	case stopHit && !limitHit:
		return true, stopFillPrice(exit.StopLevel, isLong, barOpen), "stop"
	case !stopHit && limitHit:
		return true, exit.LimitLevel, "limit"
	default:
		return fillOnIntrabarPath(exit, isLong, barOpen, barHigh, barLow)
	}
}

// stopFillPrice matches the TV broker-emulator gap-fill convention: a stop that
// is gapped through at open fills at open, not at the stop level.
func stopFillPrice(stopLevel float64, isLong bool, barOpen float64) float64 {
	if isLong && barOpen < stopLevel {
		return barOpen
	}
	if !isLong && barOpen > stopLevel {
		return barOpen
	}
	return stopLevel
}

func isStopBreached(stopLevel float64, isLong bool, barHigh, barLow float64) bool {
	if math.IsNaN(stopLevel) {
		return false
	}
	if isLong {
		return barLow <= stopLevel
	}
	return barHigh >= stopLevel
}

func isLimitBreached(limitLevel float64, isLong bool, barHigh, barLow float64) bool {
	if math.IsNaN(limitLevel) {
		return false
	}
	if isLong {
		return barHigh >= limitLevel
	}
	return barLow <= limitLevel
}

// fillOnIntrabarPath handles the case where both stop and limit breach within
// the same bar. TV broker emulator applies a conservative-fill rule: without
// sub-bar tick data, intrabar ordering is indeterminate, so stop always fills
// first (worse-for-trader outcome). Precondition: both levels are within
// [barLow, barHigh] as confirmed by checkPriceTrigger.
func fillOnIntrabarPath(exit ExitOrder, isLong bool, barOpen, barHigh, barLow float64) (bool, float64, string) {
	return true, stopFillPrice(exit.StopLevel, isLong, barOpen), "stop"
}
