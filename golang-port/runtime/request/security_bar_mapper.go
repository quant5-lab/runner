package request

import (
	"time"

	"github.com/quant5-lab/runner/runtime/context"
)

type SecurityBarMapper struct {
	ranges []BarRange
}

func NewSecurityBarMapper() *SecurityBarMapper {
	return &SecurityBarMapper{
		ranges: []BarRange{},
	}
}

func (m *SecurityBarMapper) BuildMapping(
	higherTimeframeBars []context.OHLCV,
	lowerTimeframeBars []context.OHLCV,
) {
	if len(higherTimeframeBars) == 0 || len(lowerTimeframeBars) == 0 {
		return
	}

	// Detect same-timeframe case: if timestamps align exactly, do 1:1 mapping
	if m.isSameTimeframeMapping(higherTimeframeBars, lowerTimeframeBars) {
		m.buildExactMapping(higherTimeframeBars, lowerTimeframeBars)
		return
	}

	m.ranges = make([]BarRange, 0, len(higherTimeframeBars))
	lowerIdx := 0

	for dailyIdx, dailyBar := range higherTimeframeBars {
		startIdx := lowerIdx
		dailyDate := extractDate(dailyBar.Time)

		for lowerIdx < len(lowerTimeframeBars) {
			lowerBarDate := extractDate(lowerTimeframeBars[lowerIdx].Time)

			if lowerBarDate != dailyDate {
				break
			}

			lowerIdx++
		}

		endIdx := lowerIdx - 1

		if endIdx >= startIdx {
			m.ranges = append(m.ranges, NewBarRange(dailyIdx, startIdx, endIdx))
		}
	}
}

// isSameTimeframeMapping checks if both datasets have matching timestamps (same-timeframe case)
func (m *SecurityBarMapper) isSameTimeframeMapping(higher, lower []context.OHLCV) bool {
	minLen := len(higher)
	if len(lower) < minLen {
		minLen = len(lower)
	}

	samplesToCheck := 3
	if minLen < samplesToCheck {
		samplesToCheck = minLen
	}

	matches := 0
	for i := 0; i < samplesToCheck; i++ {
		higherIdx := len(higher) - 1 - i
		lowerIdx := len(lower) - 1 - i
		if higher[higherIdx].Time == lower[lowerIdx].Time {
			matches++
		}
	}

	return matches == samplesToCheck
}

/* buildExactMapping creates 1:1 index mapping for same-timeframe */
func (m *SecurityBarMapper) buildExactMapping(higher, lower []context.OHLCV) {
	offset := len(higher) - len(lower)
	if offset < 0 {
		offset = 0
	}

	m.ranges = make([]BarRange, len(lower))
	for i := 0; i < len(lower); i++ {
		m.ranges[i] = NewBarRange(i+offset, i, i)
	}
}

func (m *SecurityBarMapper) FindDailyBarIndex(hourlyIndex int, lookahead bool) int {
	if len(m.ranges) == 0 {
		return -1
	}

	containingRangeIdx := m.findContainingRange(hourlyIndex)

	if containingRangeIdx >= 0 {
		return m.selectBarFromContainingRange(containingRangeIdx, lookahead)
	}

	if hourlyIndex < m.ranges[0].StartHourlyIndex {
		return m.handleBeforeFirstRange(lookahead)
	}

	return m.handleAfterLastRange(lookahead)
}

func (m *SecurityBarMapper) findContainingRange(hourlyIndex int) int {
	for i, r := range m.ranges {
		if r.Contains(hourlyIndex) {
			return i
		}
	}
	return -1
}

func (m *SecurityBarMapper) selectBarFromContainingRange(rangeIdx int, lookahead bool) int {
	if lookahead {
		return m.ranges[rangeIdx].DailyBarIndex
	}

	// For lookahead=false, try to use the previous bar
	if rangeIdx > 0 {
		return m.ranges[rangeIdx-1].DailyBarIndex
	}

	// First range with lookahead=false: check if we have warmup bars before it
	// If DailyBarIndex > 0, it means there are warmup bars available
	if m.ranges[0].DailyBarIndex > 0 {
		return m.ranges[0].DailyBarIndex - 1
	}

	/* Same-timeframe 1:1 mapping: use current bar when no previous exists */
	if len(m.ranges) > 0 && m.ranges[0].DailyBarIndex == 0 && rangeIdx == 0 {
		r := m.ranges[0]
		if r.StartHourlyIndex == r.EndHourlyIndex {
			return 0
		}
	}

	return -1
}

func (m *SecurityBarMapper) binarySearchRange(hourlyIndex int) int {
	left, right := 0, len(m.ranges)-1

	for left <= right {
		mid := (left + right) / 2
		r := m.ranges[mid]

		if r.Contains(hourlyIndex) {
			return mid
		}

		if r.IsBeforeRange(hourlyIndex) {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return left
}

func (m *SecurityBarMapper) handleBeforeFirstRange(lookahead bool) int {
	if lookahead {
		return 0
	}
	return -1
}

func (m *SecurityBarMapper) handleAfterLastRange(lookahead bool) int {
	lastIdx := len(m.ranges) - 1
	return m.ranges[lastIdx].DailyBarIndex
}

func extractDate(timestamp int64) string {
	t := time.Unix(timestamp, 0).UTC()
	return t.Format("2006-01-02")
}
