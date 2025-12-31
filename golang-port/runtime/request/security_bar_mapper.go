package request

import (
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
	m.BuildMappingWithDateFilter(higherTimeframeBars, lowerTimeframeBars, DateRange{}, "UTC")
}

func (m *SecurityBarMapper) BuildMappingWithDateFilter(
	higherTimeframeBars []context.OHLCV,
	lowerTimeframeBars []context.OHLCV,
	baseDateRange DateRange,
	timezone string,
) {
	if len(higherTimeframeBars) == 0 || len(lowerTimeframeBars) == 0 {
		return
	}

	if timezone == "" {
		timezone = "UTC"
	}

	m.ranges = make([]BarRange, 0, len(higherTimeframeBars))

	for dailyIdx, dailyBar := range higherTimeframeBars {
		dailyDate := ExtractDateInTimezone(dailyBar.Time, timezone)

		startIdx := -1
		endIdx := -1

		for hourlyIdx, hourlyBar := range lowerTimeframeBars {
			hourlyDate := ExtractDateInTimezone(hourlyBar.Time, timezone)

			if hourlyDate == dailyDate {
				if startIdx == -1 {
					startIdx = hourlyIdx
				}
				endIdx = hourlyIdx
			}
		}

		/* Ensures ATR calculations use same daily bar indices regardless of base TF length */
		if startIdx < 0 {
			startIdx = -1
			endIdx = -1
		}

		m.ranges = append(m.ranges, NewBarRange(dailyIdx, startIdx, endIdx))
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

	if rangeIdx > 0 {
		return m.ranges[rangeIdx-1].DailyBarIndex
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

func (m *SecurityBarMapper) GetRanges() []BarRange {
	return m.ranges
}
