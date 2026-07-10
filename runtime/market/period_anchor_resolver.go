package market

import (
	"time"

	"github.com/quant5-lab/runner/runtime/context"
)

// SessionAnchorFor derives the PeriodAnchor for a resolved Profile using the
// exchange's known session table.  Returns a midnight-origin UTC anchor for
// exchanges not in the table.
func SessionAnchorFor(profile Profile) context.PeriodAnchor {
	tz := profile.Timezone
	if tz == "" {
		tz = "UTC"
	}
	return context.PeriodAnchor{
		Timezone:          tz,
		SessionOpenMinute: profile.SessionOpenMinute,
	}
}

// For known exchanges (MOEX, NYSE, Binance, …) the curated table value is the
// authoritative TV session-open anchor; bars are not consulted.  Fixtures may
// include pre-trading bars (e.g. MOEX 09:00 MSK auction, 09:00 MSK pre-open)
// that fall inside the defined session window but are not the canonical TV
// boundary origin — observation would pick those up and shift the 3h (or any
// multi-hour) grid to a wrong slot.
//
// For unknown exchanges the mode of per-day first-bar clock-minutes is used.
// Unrecognised-exchange tickers observe rather than default, so
// crypto-unknown datasets naturally return 0 (midnight) while unrecognised
// stock exchanges return their actual session open, with no exchange silently
// falling back to UTC midnight.  When bars are absent the curated table entry
// serves as fallback.
func DeriveSessionAnchor(profile Profile, bars []context.OHLCV) context.PeriodAnchor {
	base := SessionAnchorFor(profile)
	if len(bars) == 0 || profile.Exchange != ExchangeUnknown {
		return base
	}
	return context.PeriodAnchor{
		Timezone:          base.Timezone,
		SessionOpenMinute: observedSessionOpenMinute(bars, base.Timezone),
	}
}

// observedSessionOpenMinute uses the mode of per-day first-bar clock-minutes
// rather than the global minimum so that extended-hours or pre-market outliers
// on a minority of days do not drag the anchor away from the dominant regular
// session open.
func observedSessionOpenMinute(bars []context.OHLCV, timezone string) int {
	loc := loadTimezone(timezone)
	return modeMinute(dailyFirstBarMinutes(bars, loc))
}

func dailyFirstBarMinutes(bars []context.OHLCV, loc *time.Location) []int {
	type dateKey struct {
		year  int
		month time.Month
		day   int
	}
	dayMin := make(map[dateKey]int, 32)
	for _, bar := range bars {
		t := time.Unix(bar.Time, 0).In(loc)
		key := dateKey{t.Year(), t.Month(), t.Day()}
		m := t.Hour()*60 + t.Minute()
		if prev, seen := dayMin[key]; !seen || m < prev {
			dayMin[key] = m
		}
	}
	result := make([]int, 0, len(dayMin))
	for _, m := range dayMin {
		result = append(result, m)
	}
	return result
}

// modeMinute breaks ties toward the smaller value so that a genuine session-open
// minute always wins over a later anomaly when counts are equal.
func modeMinute(minutes []int) int {
	if len(minutes) == 0 {
		return 0
	}
	freq := make(map[int]int, len(minutes))
	for _, m := range minutes {
		freq[m]++
	}
	best, bestFreq := 0, 0
	for m, f := range freq {
		if f > bestFreq || (f == bestFreq && m < best) {
			best, bestFreq = m, f
		}
	}
	return best
}

func loadTimezone(tz string) *time.Location {
	if tz == "" || tz == "UTC" {
		return time.UTC
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.UTC
	}
	return loc
}
