package regression

import (
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

	for _, bar := range normalized {
		ts := bar.Time
		if ts > 10_000_000_000 {
			ts /= 1000
		}
		instant := time.Unix(ts, 0).In(moscowLocation(t))
		if instant.Hour() < 7 {
			t.Fatalf("normalized MOEX bar at %s passed through — pre-session bar not filtered", instant)
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
