package datafetcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileFetcher_FetchSuccess(t *testing.T) {
	/* Create temp directory with test data */
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "BTC_1h.json")

	testData := `[
		{"time": 1700000000, "open": 100, "high": 105, "low": 95, "close": 102, "volume": 1000},
		{"time": 1700003600, "open": 102, "high": 107, "low": 97, "close": 104, "volume": 1100}
	]`

	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	/* Create fetcher with no latency */
	fetcher := NewFileFetcher(tmpDir, 0)

	/* Fetch data */
	bars, err := fetcher.Fetch("BTC", "1h", 0)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	/* Verify data */
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

	/* Fetch with limit */
	bars, err := fetcher.Fetch("ETH", "1D", 2)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	/* Should return last 2 bars */
	if len(bars) != 2 {
		t.Errorf("Expected 2 bars, got %d", len(bars))
	}

	if bars[0].Close != 104 {
		t.Errorf("Expected first close 104, got %.2f", bars[0].Close)
	}
}

func TestFileFetcher_SimulatedLatency(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "TEST_1m.json")

	testData := `[{"time": 1700000000, "open": 100, "high": 105, "low": 95, "close": 102, "volume": 1000}]`
	os.WriteFile(testFile, []byte(testData), 0644)

	/* Create fetcher with 50ms latency */
	fetcher := NewFileFetcher(tmpDir, 50*time.Millisecond)

	start := time.Now()
	_, err := fetcher.Fetch("TEST", "1m", 0)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	/* Should take at least 50ms */
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

	/* Write invalid JSON */
	os.WriteFile(testFile, []byte("not valid json"), 0644)

	fetcher := NewFileFetcher(tmpDir, 0)

	_, err := fetcher.Fetch("BAD", "1h", 0)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}
