package request

import (
	"github.com/quant5-lab/runner/runtime/context"
)

type MappingMode int

const (
	ModeDownscaling MappingMode = iota // Security TF < Base TF (e.g., H→D)
	ModeUpscaling                      // Security TF > Base TF (e.g., M→D, W→D)
	ModeIdentity                       // Same TF or HA on same TF
)

/*
SecurityBarMapper maps bar indices between different timeframes in security() calls.

Mode determines lookup algorithm:
  - ModeDownscaling: Containment search for which target bar contains source index
  - ModeUpscaling: Direct lookup from source index to target bar range

Thread-safe for reads after initialization (immutable ranges and mode).
*/
type SecurityBarMapper struct {
	ranges []BarRange
	mode   MappingMode
}

func NewSecurityBarMapper() *SecurityBarMapper {
	return &SecurityBarMapper{
		ranges: []BarRange{},
		mode:   ModeDownscaling,
	}
}

func (m *SecurityBarMapper) BuildMapping(
	higherTimeframeBars []context.OHLCV,
	lowerTimeframeBars []context.OHLCV,
) {
	m.BuildMappingWithDateFilter(higherTimeframeBars, lowerTimeframeBars, DateRange{}, "UTC")
}

/*
BuildMappingWithDateFilter creates downscaling mappings (Higher TF → Lower TF bar ranges).

Each higher TF bar owns lower TF bars from its date up to the next higher TF bar's date.
Gaps in higher TF data (weekends, holidays) extend the previous bar's range, matching
TradingView behavior.

Example (with weekend gap):
  - Daily Fri 2025-01-03 → Hourly [Fri 09:00..Sun 18:00] (extends through weekend)
  - Daily Mon 2025-01-06 → Hourly [Mon 09:00..Mon 18:00]
*/
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

	m.mode = ModeDownscaling
	m.ranges = make([]BarRange, 0, len(higherTimeframeBars))
	lowerIdx := 0

	if len(higherTimeframeBars) > 0 && len(lowerTimeframeBars) > 0 {
		firstHigherDate := ExtractDateInTimezone(higherTimeframeBars[0].Time, timezone)
		for lowerIdx < len(lowerTimeframeBars) {
			lowerDate := ExtractDateInTimezone(lowerTimeframeBars[lowerIdx].Time, timezone)
			if lowerDate >= firstHigherDate {
				break
			}
			lowerIdx++
		}
	}

	for dailyIdx := range higherTimeframeBars {
		startIdx := lowerIdx

		nextDailyDate := ""
		if dailyIdx+1 < len(higherTimeframeBars) {
			nextDailyDate = ExtractDateInTimezone(higherTimeframeBars[dailyIdx+1].Time, timezone)
		}

		for lowerIdx < len(lowerTimeframeBars) {
			lowerBarDate := ExtractDateInTimezone(lowerTimeframeBars[lowerIdx].Time, timezone)

			if nextDailyDate != "" && lowerBarDate >= nextDailyDate {
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

/*
BuildIdentityMapping creates 1:1 mappings for same-timeframe security calls.

Used when security and base timeframes are identical (e.g., HA on same TF).
Each bar index maps directly to itself: secBarIdx = mainBarIdx.

Parameters:
  - barCount: Number of bars to map
*/
func (m *SecurityBarMapper) BuildIdentityMapping(barCount int) {
	m.mode = ModeIdentity
	m.ranges = nil // Identity mode doesn't use ranges
}

/*
BuildMappingForUpscaling creates upscaling mappings (Lower TF → Higher TF bar ranges).

Used when security timeframe > base timeframe (e.g., Weekly base with Daily security).
Maps each lower TF bar to all higher TF bars within its time period.

Example: Weekly → Daily upscaling
  - Weekly bar #0 (Jan 2-6) → Daily bars [0..4] (Mon-Fri)
  - Weekly bar #1 (Jan 9-13) → Daily bars [5..9] (Mon-Fri)

Allows direct lookup: ranges[weeklyIdx] returns Daily bar range for that week.
No future peeking: Returns StartIdx (first Daily bar) by default.

Parameters:
  - higherFreqBars: Target security timeframe bars (higher frequency, e.g., Daily)
  - lowerFreqBars: Base execution timeframe bars (lower frequency, e.g., Weekly)
  - timezone: Timezone for period calculation (default "UTC")
*/
func (m *SecurityBarMapper) BuildMappingForUpscaling(
	higherFreqBars []context.OHLCV,
	lowerFreqBars []context.OHLCV,
	timezone string,
) {
	if len(higherFreqBars) == 0 || len(lowerFreqBars) == 0 {
		return
	}

	if timezone == "" {
		timezone = "UTC"
	}

	m.mode = ModeUpscaling
	m.ranges = make([]BarRange, 0, len(lowerFreqBars))

	for loIdx, loBar := range lowerFreqBars {
		startIdx := -1
		endIdx := -1

		nextLoBarTime := int64(1<<63 - 1) // max int64
		if loIdx+1 < len(lowerFreqBars) {
			nextLoBarTime = lowerFreqBars[loIdx+1].Time
		}

		for hiIdx, hiBar := range higherFreqBars {
			if hiBar.Time >= loBar.Time && hiBar.Time < nextLoBarTime {
				if startIdx == -1 {
					startIdx = hiIdx
				}
				endIdx = hiIdx
			}
		}

		if startIdx < 0 {
			startIdx = -1
			endIdx = -1
		}

		m.ranges = append(m.ranges, NewBarRange(loIdx, startIdx, endIdx))
	}
}

/*
FindDailyBarIndex dispatches to the appropriate lookup algorithm based on mapping mode.

IDENTITY MODE (same TF, e.g., HA on same TF):
  - Direct 1:1 mapping: returns barIndex unchanged
  - Ignores lookahead parameter

UPSCALING MODE (security TF > base TF, e.g., M→D, W→D):
  - Direct index lookup: ranges[baseBarIndex] contains the security bar range
  - Returns StartIdx (first bar in period) by default
  - Returns EndIdx (last bar in period) with lookahead=true

DOWNSCALING MODE (security TF < base TF, e.g., H→D):
  - Containment search: finds which security bar contains baseBarIndex
  - Returns the security bar index for that containing range
  - With lookahead=false: returns previous security bar
  - With lookahead=true: returns current security bar

Returns -1 if no valid mapping found.
Thread-safe after mapper initialization.
*/
func (m *SecurityBarMapper) FindDailyBarIndex(barIndex int, lookahead bool) int {
	if m.mode == ModeIdentity {
		return barIndex
	}
	if m.mode == ModeUpscaling {
		return m.findUpscalingIndex(barIndex, lookahead)
	}
	return m.findDownscalingIndex(barIndex, lookahead)
}

func (m *SecurityBarMapper) findUpscalingIndex(baseBarIndex int, lookahead bool) int {
	if baseBarIndex < 0 || baseBarIndex >= len(m.ranges) {
		return -1
	}

	r := m.ranges[baseBarIndex]
	if r.StartHourlyIndex < 0 {
		return -1
	}

	if lookahead {
		return r.EndHourlyIndex
	}
	return r.StartHourlyIndex
}

func (m *SecurityBarMapper) findDownscalingIndex(sourceBarIndex int, lookahead bool) int {
	if len(m.ranges) == 0 {
		return -1
	}

	for i, r := range m.ranges {
		if r.Contains(sourceBarIndex) {
			if lookahead {
				return r.DailyBarIndex
			}
			if i > 0 {
				return m.ranges[i-1].DailyBarIndex
			}
			/* First range with lookahead=false returns current daily bar (no previous available) */
			return r.DailyBarIndex
		}
	}

	if len(m.ranges) > 0 {
		lastRange := m.ranges[len(m.ranges)-1]
		if sourceBarIndex > lastRange.EndHourlyIndex {
			return lastRange.DailyBarIndex
		}
	}

	return -1
}

/*
FindTargetBarIndexByContainment finds which target TF bar contains the given source bar index.

Legacy method maintained for backward compatibility.
Prefer using FindDailyBarIndex which dispatches based on mapping mode.

Returns -1 if no containing range found and sourceBarIndex is before first range.
Returns last target bar index if sourceBarIndex is after all ranges.
*/
func (m *SecurityBarMapper) FindTargetBarIndexByContainment(sourceBarIndex int, lookahead bool) int {
	return m.findDownscalingIndex(sourceBarIndex, lookahead)
}

func (m *SecurityBarMapper) GetRanges() []BarRange {
	return m.ranges
}
