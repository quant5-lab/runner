package strategy

// IntrabarPricePath is the assumed intrabar price movement sequence used by
// TradingView's broker emulator to determine order fill priority when multiple
// exit levels are breached within the same bar.
type IntrabarPricePath int

const (
	// PathHighBeforeLow — open is closer to (or equidistant from) the bar high:
	// assumed sequence open → high → low → close.
	PathHighBeforeLow IntrabarPricePath = iota
	// PathLowBeforeHigh — open is closer to the bar low:
	// assumed sequence open → low → high → close.
	PathLowBeforeHigh
)

// IntrabarPath returns the assumed intrabar price path based on the relative
// distance of the bar open to the high vs. the low — the same heuristic used
// by TradingView's broker emulator (process_orders_on_close=false).
//
// Tie (open equidistant from high and low, or open == high == low) resolves to
// PathHighBeforeLow, matching TradingView's default behaviour.
func IntrabarPath(barOpen, barHigh, barLow float64) IntrabarPricePath {
	if barHigh-barOpen <= barOpen-barLow {
		return PathHighBeforeLow
	}
	return PathLowBeforeHigh
}
