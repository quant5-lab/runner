package codegen

import "fmt"

// HistoricalOffset represents a lookback offset in series data access.
type HistoricalOffset struct {
	value int
}

// NewHistoricalOffset creates an offset with the given value.
func NewHistoricalOffset(value int) HistoricalOffset {
	return HistoricalOffset{value: value}
}

// NoOffset returns a zero offset for current bar access.
func NoOffset() HistoricalOffset {
	return HistoricalOffset{value: 0}
}

func (o HistoricalOffset) Value() int {
	return o.value
}

func (o HistoricalOffset) IsZero() bool {
	return o.value == 0
}

// Add combines this offset with an additional offset.
func (o HistoricalOffset) Add(other int) int {
	return o.value + other
}

// FormatLoopAccess generates loop access expression accounting for this offset.
func (o HistoricalOffset) FormatLoopAccess(loopVar string) string {
	if o.IsZero() {
		return loopVar
	}
	return fmt.Sprintf("%s+%d", loopVar, o.value)
}

