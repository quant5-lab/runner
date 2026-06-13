package datafetcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileFetcher_FetchSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "BTC_1h.json")

	testData := `[
		{"time": 1700000000, "open": 100, "high": 105, "low": 95, "close": 102, "volume": 1000},
		{"time": 1700003600, "open": 102, "high": 107, "low": 97, "close": 104, "volume": 1100}
	]`

	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	fetcher := NewFileFetcher(tmpDir, 0)

	bars, err := fetcher.Fetch("BTC", "1h", 0)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if len(bars) != 2 {
		t.Errorf("Expected 2 bars, got %d", len(bars))
	}

	if bars[0].Close != 102 {
		t.Errorf("Expected first close 102, got %.2f", bars[0].Close)
	}
}

func TestFileFetcher_FetchWithLimit(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "ETH_1D.json")

	testData := `[
		{"time": 1700000000, "open": 100, "high": 105, "low": 95, "close": 102, "volume": 1000},
		{"time": 1700086400, "open": 102, "high": 107, "low": 97, "close": 104, "volume": 1100},
		{"time": 1700172800, "open": 104, "high": 109, "low": 99, "close": 106, "volume": 1200}
	]`

	os.WriteFile(testFile, []byte(testData), 0644)

	fetcher := NewFileFetcher(tmpDir, 0)

	bars, err := fetcher.Fetch("ETH", "1D", 2)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if len(bars) != 2 {
		t.Errorf("Expected 2 bars, got %d", len(bars))
	}

	if bars[0].Close != 104 {
		t.Errorf("Expected first close 104, got %.2f", bars[0].Close)
	}
}

func TestFileFetcher_FetchWithMetadata_PreservesSourceMetadataAndLimit(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "SBERP_1h.json")

	testData := `{
		"timezone": "Europe/Moscow",
			"exchange": "MOEX",
			"referenceSession": "regular",
			"qtyStep": 0.00001,
			"openDates": ["2025-08-16"],
		"bars": [
			{"time": 1700000000, "open": 100, "high": 105, "low": 95, "close": 102, "volume": 1000},
			{"time": 1700003600, "open": 102, "high": 107, "low": 97, "close": 104, "volume": 1100},
			{"time": 1700007200, "open": 104, "high": 109, "low": 99, "close": 106, "volume": 1200}
		]
	}`
	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	fetcher := NewFileFetcher(tmpDir, 0)
	marketData, err := fetcher.FetchWithMetadata("SBERP", "1h", 2)
	if err != nil {
		t.Fatalf("FetchWithMetadata failed: %v", err)
	}

	if marketData.Timezone != "Europe/Moscow" {
		t.Fatalf("timezone = %q, want Europe/Moscow", marketData.Timezone)
	}
	if marketData.ReferenceSession != "regular" {
		t.Fatalf("reference session = %q, want regular", marketData.ReferenceSession)
	}
	if marketData.Exchange != "MOEX" {
		t.Fatalf("exchange = %q, want MOEX", marketData.Exchange)
	}
	if marketData.QtyStep != 0.00001 {
		t.Fatalf("qty step = %.8f, want 0.00001000", marketData.QtyStep)
	}
	if len(marketData.OpenDates) != 1 || marketData.OpenDates[0] != "2025-08-16" {
		t.Fatalf("open dates = %#v, want [2025-08-16]", marketData.OpenDates)
	}
	if len(marketData.Bars) != 2 {
		t.Fatalf("bars = %d, want 2", len(marketData.Bars))
	}
	if marketData.Bars[0].Close != 104 {
		t.Fatalf("first limited close = %.2f, want 104.00", marketData.Bars[0].Close)
	}
}

func TestFileFetcher_FetchWithMetadata_ArrayInputHasEmptyMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "BTCUSDT_1h.json")

	testData := `[{"time": 1700000000, "open": 100, "high": 105, "low": 95, "close": 102, "volume": 1000}]`
	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	fetcher := NewFileFetcher(tmpDir, 0)
	marketData, err := fetcher.FetchWithMetadata("BTCUSDT", "1h", 0)
	if err != nil {
		t.Fatalf("FetchWithMetadata failed: %v", err)
	}

	if marketData.Timezone != "" {
		t.Fatalf("timezone = %q, want empty", marketData.Timezone)
	}
	if marketData.ReferenceSession != "" {
		t.Fatalf("reference session = %q, want empty", marketData.ReferenceSession)
	}
	if len(marketData.Bars) != 1 {
		t.Fatalf("bars = %d, want 1", len(marketData.Bars))
	}
}

func TestFileFetcher_FetchWithMetadata_ObjectInputAllowsEmptyBars(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "SBERP_1h.json")

	testData := `{"timezone":"Europe/Moscow","referenceSession":"regular","bars":[]}`
	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	fetcher := NewFileFetcher(tmpDir, 0)
	marketData, err := fetcher.FetchWithMetadata("SBERP", "1h", 0)
	if err != nil {
		t.Fatalf("FetchWithMetadata failed: %v", err)
	}

	if marketData.Timezone != "Europe/Moscow" {
		t.Fatalf("timezone = %q, want Europe/Moscow", marketData.Timezone)
	}
	if marketData.ReferenceSession != "regular" {
		t.Fatalf("reference session = %q, want regular", marketData.ReferenceSession)
	}
	if len(marketData.Bars) != 0 {
		t.Fatalf("bars = %d, want 0", len(marketData.Bars))
	}
}

func TestFileFetcher_FetchWithMetadata_InvalidObjectBarsFails(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "BAD_1h.json")

	testData := `{"timezone":"UTC","bars":{}}`
	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	fetcher := NewFileFetcher(tmpDir, 0)
	if _, err := fetcher.FetchWithMetadata("BAD", "1h", 0); err == nil {
		t.Fatal("expected invalid object bars error")
	}
}

func TestFileFetcher_SimulatedLatency(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "TEST_1m.json")

	testData := `[{"time": 1700000000, "open": 100, "high": 105, "low": 95, "close": 102, "volume": 1000}]`
	os.WriteFile(testFile, []byte(testData), 0644)

	fetcher := NewFileFetcher(tmpDir, 50*time.Millisecond)

	start := time.Now()
	_, err := fetcher.Fetch("TEST", "1m", 0)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if elapsed < 50*time.Millisecond {
		t.Errorf("Expected latency >=50ms, got %v", elapsed)
	}
}

func TestFileFetcher_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	fetcher := NewFileFetcher(tmpDir, 0)

	_, err := fetcher.Fetch("NONEXISTENT", "1h", 0)
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestFileFetcher_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "BAD_1h.json")

	os.WriteFile(testFile, []byte("not valid json"), 0644)

	fetcher := NewFileFetcher(tmpDir, 0)

	_, err := fetcher.Fetch("BAD", "1h", 0)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}
