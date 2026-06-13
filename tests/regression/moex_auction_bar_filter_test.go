package regression

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/quant5-lab/runner/runtime/market"
)

func TestRegularSessionNormalization_IdempotentOnMOEXFixture(t *testing.T) {
	root := projectRootFromCwd()
	raw := loadOHLCVBars(t, filepath.Join(root, "tests", "golden", "fixtures", "data", "SBERP-1h.json"))
	bars := fixtureBarsToRuntimeBars(raw)

	once, _, _ := market.NormalizeBarsWithReferenceSession("SBERP", "1h", "", "regular", bars)
	twice, _, _ := market.NormalizeBarsWithReferenceSession("SBERP", "1h", "", "regular", once)

	if len(twice) != len(once) {
		t.Fatalf("normalization is not idempotent: first pass=%d bars, second pass=%d bars", len(once), len(twice))
	}
}

func TestRegularSessionNormalization_RemovesBarsBeyondAlwaysOpen_MOEX(t *testing.T) {
	root := projectRootFromCwd()
	raw := loadOHLCVBars(t, filepath.Join(root, "tests", "golden", "fixtures", "data", "SBERP-1h.json"))
	bars := fixtureBarsToRuntimeBars(raw)

	alwaysOpen, _, _ := market.NormalizeBarsWithReferenceSession("SBERP", "1h", "Europe/Moscow", "always-open", bars)
	regular, _, _ := market.NormalizeBarsWithReferenceSession("SBERP", "1h", "", "regular", bars)

	if len(regular) >= len(alwaysOpen) {
		t.Fatalf("regular session must remove bars beyond always-open: always-open=%d regular=%d", len(alwaysOpen), len(regular))
	}
}

func TestRegularSessionNormalization_EveryOutputBarAcceptedByCalendar_MOEX(t *testing.T) {
	root := projectRootFromCwd()
	raw := loadOHLCVBars(t, filepath.Join(root, "tests", "golden", "fixtures", "data", "SBERP-1h.json"))
	bars := fixtureBarsToRuntimeBars(raw)

	normalized, _, _ := market.NormalizeBarsWithReferenceSession("SBERP", "1h", "", "regular", bars)

	cal := market.RegularCalendarForExchange(market.ExchangeMOEX, market.SourceMetadata{})
	for _, bar := range normalized {
		if !cal.Accepts(bar, "1h", "Europe/Moscow") {
			t.Fatalf("normalized bar at time=%d rejected by MOEX regular calendar", bar.Time)
		}
	}
}

func TestAlwaysOpenNormalization_PreservesAllBars_MOEX(t *testing.T) {
	root := projectRootFromCwd()
	raw := loadOHLCVBars(t, filepath.Join(root, "tests", "golden", "fixtures", "data", "SBERP-1h.json"))
	bars := fixtureBarsToRuntimeBars(raw)

	normalized, _, _ := market.NormalizeBarsWithReferenceSession("SBERP", "1h", "Europe/Moscow", "always-open", bars)

	if len(normalized) != len(bars) {
		t.Fatalf("always-open normalization must preserve all bars: before=%d after=%d", len(bars), len(normalized))
	}
}

func TestDefaultSessionNormalization_FiltersMOEXAuctionBars(t *testing.T) {
	root := projectRootFromCwd()
	raw := loadOHLCVBars(t, filepath.Join(root, "tests", "golden", "fixtures", "data", "SBERP-1h.json"))
	bars := fixtureBarsToRuntimeBars(raw)

	normalized, _ := market.NormalizeBars("SBERP", "1h", "Europe/Moscow", bars)

	if len(normalized) >= len(bars) {
		t.Fatalf("MOEX default session must filter pre-session auction bars: before=%d after=%d", len(bars), len(normalized))
	}

	loc := moscowLocation(t)
	for _, bar := range normalized {
		instant := time.Unix(normalizedUnixSecond(bar.Time), 0).In(loc)
		minOfDay := instant.Hour()*60 + instant.Minute()
		wd := instant.Weekday()
		if wd == time.Saturday || wd == time.Sunday {
			if minOfDay < 600 || minOfDay >= 1140 {
				t.Fatalf("MOEX weekend bar outside [10:00,19:00) session passed through: %s", instant)
			}
		} else {
			if minOfDay < 420 {
				t.Fatalf("MOEX weekday pre-session bar (before 07:00 MSK) passed through: %s", instant)
			}
		}
	}
}

func moscowLocation(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		t.Fatalf("load Europe/Moscow: %v", err)
	}
	return loc
}

func TestRegularSessionNormalization_PreservesAllBars_Crypto(t *testing.T) {
	root := projectRootFromCwd()
	raw := loadOHLCVBars(t, filepath.Join(root, "tests", "golden", "fixtures", "data", "BTCUSDT-1h.json"))
	bars := fixtureBarsToRuntimeBars(raw)

	normalized, _, _ := market.NormalizeBarsWithReferenceSession("BTCUSDT", "1h", "", "regular", bars)

	if len(normalized) != len(bars) {
		t.Fatalf("BTCUSDT regular session must preserve all bars: before=%d after=%d", len(bars), len(normalized))
	}
}

func TestDefaultSessionNormalization_PreservesMOEXDailyBars(t *testing.T) {
	root := projectRootFromCwd()
	raw := loadOHLCVBars(t, filepath.Join(root, "tests", "golden", "fixtures", "data", "SBERP_1D.json"))
	bars := fixtureBarsToRuntimeBars(raw)

	normalized, _ := market.NormalizeBars("SBERP", "1D", "Europe/Moscow", bars)

	if len(normalized) != len(bars) {
		t.Fatalf("MOEX 1D normalization must preserve all bars: before=%d after=%d", len(bars), len(normalized))
	}
}

func TestMOEXWeekendSession_AcceptsMidSessionBars(t *testing.T) {
	root := projectRootFromCwd()
	raw := loadOHLCVBars(t, filepath.Join(root, "tests", "golden", "fixtures", "data", "SBERP-1h.json"))
	bars := fixtureBarsToRuntimeBars(raw)

	normalized, _ := market.NormalizeBars("SBERP", "1h", "Europe/Moscow", bars)

	loc := moscowLocation(t)
	midSessionCount := 0
	for _, bar := range normalized {
		instant := time.Unix(normalizedUnixSecond(bar.Time), 0).In(loc)
		wd := instant.Weekday()
		if (wd == time.Saturday || wd == time.Sunday) && instant.Hour() >= 10 && instant.Hour() < 19 {
			midSessionCount++
		}
	}

	if midSessionCount == 0 {
		t.Fatal("MOEX normalization must preserve weekend mid-session bars (10:00–18:59 MSK): none found")
	}
}

func TestMOEXRegularSessionNormalization_RemovedBarsAreClassified(t *testing.T) {
	cases := []struct {
		symbol string
		file   string
	}{
		{symbol: "SBERP", file: "SBERP-1h.json"},
		{symbol: "CNRU", file: "CNRU-1h.json"},
	}

	for _, tc := range cases {
		t.Run(tc.symbol, func(t *testing.T) {
			root := projectRootFromCwd()
			path := filepath.Join(root, "tests", "golden", "fixtures", "data", tc.file)
			raw := loadOHLCVBars(t, path)
			bars := fixtureBarsToRuntimeBars(raw)
			metadata := loadMOEXFixtureMetadata(t, path)

			normalized, _, err := market.NormalizeBarsWithMetadataE(tc.symbol, "1h", metadata, bars)
			if err != nil {
				t.Fatalf("normalize %s: %v", tc.file, err)
			}
			kept := map[int64]struct{}{}
			for _, bar := range normalized {
				kept[normalizedUnixSecond(bar.Time)] = struct{}{}
			}

			removed := 0
			for _, bar := range bars {
				if _, ok := kept[normalizedUnixSecond(bar.Time)]; ok {
					continue
				}
				removed++
				if got := classifyMOEXRegularSessionRemoval(t, bar.Time); got == "other" {
					t.Fatalf("removed bar has no accepted category: file=%s time=%s", tc.file, moscowInstant(t, bar.Time).Format("2006-01-02 Mon 15:04"))
				}
			}
			if removed == 0 {
				t.Fatalf("%s regular-session normalization removed no bars", tc.file)
			}
		})
	}
}

func loadMOEXFixtureMetadata(t *testing.T, path string) market.SourceMetadata {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture metadata %s: %v", filepath.Base(path), err)
	}
	var fixture struct {
		Timezone  string   `json:"timezone"`
		OpenDates []string `json:"openDates"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("parse fixture metadata %s: %v", filepath.Base(path), err)
	}
	return market.SourceMetadata{
		Exchange:         "MOEX",
		Timezone:         fixture.Timezone,
		ReferenceSession: "regular",
		OpenDates:        fixture.OpenDates,
	}
}

func classifyMOEXRegularSessionRemoval(t *testing.T, unixTime int64) string {
	instant := moscowInstant(t, unixTime)
	minute := instant.Hour()*60 + instant.Minute()
	weekend := instant.Weekday() == time.Saturday || instant.Weekday() == time.Sunday
	date := instant.Format("2006-01-02")

	if date == "2025-11-05" && minute == 6*60 {
		return "special holiday/auction-print"
	}
	if weekend && minute < 10*60 {
		return "weekend pre-session"
	}
	if weekend && minute >= 19*60 {
		return "weekend post-session"
	}
	if !weekend && minute < 7*60 {
		return "weekday pre-session"
	}
	return "other"
}

func moscowInstant(t *testing.T, unixTime int64) time.Time {
	t.Helper()
	return time.Unix(normalizedUnixSecond(unixTime), 0).In(moscowLocation(t))
}

func normalizedUnixSecond(unixTime int64) int64 {
	if unixTime > 10_000_000_000 {
		return unixTime / 1000
	}
	return unixTime
}
