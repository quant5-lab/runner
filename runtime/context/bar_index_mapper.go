package context

/* BarIndexMapper maps base-timeframe bar indices to security-timeframe bar indices */
type BarIndexMapper interface {
	FindDailyBarIndex(barIndex int, lookahead bool) int
}
