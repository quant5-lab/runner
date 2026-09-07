package regression

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type closeAllTrade struct {
	EntryBar  int     `json:"entryBar"`
	ExitBar   int     `json:"exitBar"`
	ExitPrice float64 `json:"exitPrice"`
}

type closeAllOutput struct {
	Strategy struct {
		Trades     []closeAllTrade `json:"trades"`
		OpenTrades []closeAllTrade `json:"openTrades"`
	} `json:"strategy"`
}

func runCloseAllScript(t *testing.T, pineScript, ohlcvJSON string) closeAllOutput {
	t.Helper()
	root := projectRootFromCwd()
	tmpDir := t.TempDir()

	built, ok := codegenAndBuild(t, tmpDir, "closeall_test", pineScript, root)
	if !ok {
		t.Fatal("close_all script codegen/build failed")
	}

	dataPath := filepath.Join(tmpDir, "bars.json")
	if err := os.WriteFile(dataPath, []byte(ohlcvJSON), 0644); err != nil {
		t.Fatalf("write data: %v", err)
	}
	outputPath := filepath.Join(tmpDir, "out.json")
	cmd := exec.Command(built.BinaryPath,
		"-symbol", "TEST",
		"-data", dataPath,
		"-output", outputPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("strategy run failed: %v\n%s", err, out)
	}
	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	var result closeAllOutput
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("parse output: %v", err)
	}
	return result
}

func makeCloseAllBars(specs [][4]float64, intervalSec int64) string {
	type Bar struct {
		Time   int64   `json:"time"`
		Open   float64 `json:"open"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Close  float64 `json:"close"`
		Volume float64 `json:"volume"`
	}
	type Envelope struct {
		Timezone string `json:"timezone"`
		Bars     []Bar  `json:"bars"`
	}
	const startTime int64 = 1640000000
	bars := make([]Bar, len(specs))
	for i, s := range specs {
		bars[i] = Bar{
			Time:   startTime + int64(i)*intervalSec,
			Open:   s[0],
			High:   s[1],
			Low:    s[2],
			Close:  s[3],
			Volume: 100,
		}
	}
	b, _ := json.Marshal(Envelope{Timezone: "UTC", Bars: bars})
	return string(b)
}

func TestCloseAll_PriceStopCondition_ClosesAtNextOpen(t *testing.T) {
	tests := []struct {
		name        string
		pine        string
		bars        [][4]float64
		wantExitBar int
		wantExitPx  float64
	}{
		{
			// Long: stop at 95; bar 2 low=93 breaches → close fills at bar 3 open=102.
			name: "long_stop_breach_exits_at_next_open",
			pine: `
//@version=4
strategy("CloseAll Stop Long", overlay=true, default_qty_type=strategy.fixed, default_qty_value=1, initial_capital=10000)
stopLevel = 95.0
if bar_index == 0 and strategy.position_size == 0
    strategy.entry("e", strategy.long)
if strategy.position_size > 0 and low < stopLevel
    strategy.close_all()
`,
			bars: [][4]float64{
				{99, 101, 98, 100},   // bar 0: entry signal (bar_index==0)
				{100, 105, 99, 104},  // bar 1: entry fills at open=100; no breach (low=99>95)
				{104, 106, 93, 100},  // bar 2: low=93 < stop=95 → close_all queued
				{102, 108, 101, 106}, // bar 3: close_all fills at open=102
				{106, 110, 104, 108}, // bar 4: post-fill; no re-entry (bar_index≠0)
			},
			wantExitBar: 3,
			wantExitPx:  102,
		},
		{
			// Short: stop at 110; bar 2 high=112 breaches → close fills at bar 3 open=98.
			name: "short_stop_breach_exits_at_next_open",
			pine: `
//@version=4
strategy("CloseAll Stop Short", overlay=true, default_qty_type=strategy.fixed, default_qty_value=1, initial_capital=10000)
stopLevel = 110.0
if bar_index == 0 and strategy.position_size == 0
    strategy.entry("e", strategy.short)
if strategy.position_size < 0 and high > stopLevel
    strategy.close_all()
`,
			bars: [][4]float64{
				{105, 106, 103, 104}, // bar 0: entry signal (bar_index==0)
				{104, 106, 102, 103}, // bar 1: entry fills at open=104; no breach (high=106<110)
				{103, 112, 101, 105}, // bar 2: high=112 > stop=110 → close_all queued
				{98, 100, 96, 99},    // bar 3: close_all fills at open=98
				{99, 103, 97, 100},   // bar 4: post-fill
			},
			wantExitBar: 3,
			wantExitPx:  98,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			result := runCloseAllScript(t, tt.pine, makeCloseAllBars(tt.bars, 3600))

			if len(result.Strategy.Trades) != 1 {
				t.Fatalf("closed trades: got %d, want 1", len(result.Strategy.Trades))
			}
			trade := result.Strategy.Trades[0]
			if trade.ExitBar != tt.wantExitBar {
				t.Errorf("exit bar: got %d, want %d", trade.ExitBar, tt.wantExitBar)
			}
			if math.Abs(trade.ExitPrice-tt.wantExitPx) > 0.01 {
				t.Errorf("exit price: got %.4f, want %.4f (next bar open)", trade.ExitPrice, tt.wantExitPx)
			}
			if len(result.Strategy.OpenTrades) != 0 {
				t.Errorf("open trades after close_all: got %d, want 0", len(result.Strategy.OpenTrades))
			}
		})
	}
}

func TestCloseAll_ConditionFalse_PositionPersists(t *testing.T) {
	// stop at 80; all bar lows stay above 90 → condition never fires.
	pine := `
//@version=4
strategy("CloseAll No-Trigger", overlay=true, default_qty_type=strategy.fixed, default_qty_value=1, initial_capital=10000)
if bar_index == 0 and strategy.position_size == 0
    strategy.entry("e", strategy.long)
if strategy.position_size > 0 and low < 80
    strategy.close_all()
`
	// 8 bars, all low >= 91 — condition never breaches.
	bars := [][4]float64{
		{99, 101, 97, 100},
		{100, 105, 99, 104},
		{104, 108, 96, 107},
		{107, 110, 95, 109},
		{109, 112, 92, 111},
		{111, 114, 91, 113},
		{113, 116, 93, 115},
		{115, 118, 94, 117},
	}

	result := runCloseAllScript(t, pine, makeCloseAllBars(bars, 3600))

	if len(result.Strategy.Trades) != 0 {
		t.Errorf("closed trades: got %d, want 0 (condition never fires)", len(result.Strategy.Trades))
	}
	if len(result.Strategy.OpenTrades) != 1 {
		t.Errorf("open trades: got %d, want 1 (position stays open)", len(result.Strategy.OpenTrades))
	}
}

func TestCloseAll_MultipleConditionalBlocks_IdempotentClose(t *testing.T) {
	// Two conditions that both fire on bar 2: low < 90 AND high > 110.
	// Bar 3 open = 102 should be the single exit price.
	pine := `
//@version=4
strategy("CloseAll Idempotent", overlay=true, default_qty_type=strategy.fixed, default_qty_value=1, initial_capital=10000)
if bar_index == 0 and strategy.position_size == 0
    strategy.entry("e", strategy.long)
if strategy.position_size > 0 and low < 90
    strategy.close_all()
if strategy.position_size > 0 and high > 110
    strategy.close_all()
`
	bars := [][4]float64{
		{99, 101, 97, 100},   // bar 0: entry signal (bar_index==0); fills bar 1 open
		{100, 105, 99, 104},  // bar 1: entry fills at open=100; no breach
		{104, 115, 85, 108},  // bar 2: both low<90 AND high>110 → both close_all blocks fire
		{102, 106, 100, 105}, // bar 3: close-all fills here at open=102
		{105, 109, 103, 107}, // bar 4: post-fill; no re-entry (bar_index≠0)
	}

	result := runCloseAllScript(t, pine, makeCloseAllBars(bars, 3600))

	if len(result.Strategy.Trades) != 1 {
		t.Fatalf("closed trades: got %d, want exactly 1 (idempotent close_all)", len(result.Strategy.Trades))
	}
	if len(result.Strategy.OpenTrades) != 0 {
		t.Errorf("open trades after close_all: got %d, want 0", len(result.Strategy.OpenTrades))
	}
	trade := result.Strategy.Trades[0]
	if trade.ExitBar != 3 {
		t.Errorf("exit bar: got %d, want 3 (one bar after both conditions fire)", trade.ExitBar)
	}
	const wantExitPx = 102.0
	if math.Abs(trade.ExitPrice-wantExitPx) > 0.01 {
		t.Errorf("exit price: got %.4f, want %.4f (bar 3 open)", trade.ExitPrice, wantExitPx)
	}
}

func TestCloseAll_PersistsAcrossBarsUntilConditionMet(t *testing.T) {
	const quietBars = 4
	const stopLevel = 90.0

	pine := fmt.Sprintf(`//@version=4
strategy("CloseAll Delayed", overlay=true, default_qty_type=strategy.fixed, default_qty_value=1, initial_capital=10000)
if strategy.position_size == 0 and bar_index == 1
    strategy.entry("e", strategy.long)
if strategy.position_size > 0 and low < %.1f
    strategy.close_all()
`, stopLevel)

	// bar 0: no signal
	// bar 1: entry (bar_index=1); fills at bar 2 open
	// bars 2..quietBars+1: quiet (low = 95, above stop)
	// bar quietBars+2: breach (low = 88 < 90)
	// bar quietBars+3: close fills here
	specCount := quietBars + 4
	bars := make([][4]float64, specCount)
	bars[0] = [4]float64{99, 101, 97, 100}
	bars[1] = [4]float64{100, 105, 99, 104}
	// Quiet bars: low stays above stopLevel
	for i := 2; i <= quietBars+1; i++ {
		bars[i] = [4]float64{104, 108, 95, 107}
	}
	// Breach bar
	bars[quietBars+2] = [4]float64{107, 109, 88, 108} // low=88 < 90
	// Fill bar
	bars[quietBars+3] = [4]float64{103, 106, 101, 105}

	result := runCloseAllScript(t, pine, makeCloseAllBars(bars, 3600))

	if len(result.Strategy.Trades) != 1 {
		t.Fatalf("closed trades: got %d, want 1", len(result.Strategy.Trades))
	}
	trade := result.Strategy.Trades[0]
	wantExitBar := quietBars + 3
	if trade.ExitBar != wantExitBar {
		t.Errorf("exit bar: got %d, want %d (one bar after first breach at bar %d)",
			trade.ExitBar, wantExitBar, quietBars+2)
	}
	if len(result.Strategy.OpenTrades) != 0 {
		t.Errorf("open trades after close: got %d, want 0", len(result.Strategy.OpenTrades))
	}
}

// close_all registered bar N fills at bar N+1 OnBarUpdate (before OnBarMetrics), so a pending stop breach on bar N+1 never fires.
func TestCloseAll_PendingExitClearedByNextBarFill(t *testing.T) {
	pine := `
//@version=4
strategy("CloseAll Clears Pending Stop", overlay=true, default_qty_type=strategy.fixed, default_qty_value=1, initial_capital=10000)
if bar_index == 0 and strategy.position_size == 0
    strategy.entry("e", strategy.long)
    strategy.exit("x", "e", stop=90)
if bar_index == 1 and strategy.position_size > 0
    strategy.close_all()
`
	bars := [][4]float64{
		{99, 101, 97, 100},   // bar 0: entry + exit registered
		{100, 105, 99, 104},  // bar 1: fills at open=100; low=99>90 (no breach); close_all queued
		{102, 107, 85, 106},  // bar 2: close_all fills at open=102; low=85 would breach stop but exit is cleared
		{106, 110, 104, 108}, // bar 3: post
	}

	result := runCloseAllScript(t, pine, makeCloseAllBars(bars, 3600))

	if len(result.Strategy.Trades) != 1 {
		t.Fatalf("closed trades: got %d, want 1", len(result.Strategy.Trades))
	}
	trade := result.Strategy.Trades[0]
	if trade.ExitBar != 2 {
		t.Errorf("exit bar: got %d, want 2 (close_all fill bar)", trade.ExitBar)
	}
	const wantExitPx = 102.0
	if math.Abs(trade.ExitPrice-wantExitPx) > 0.01 {
		t.Errorf("exit price: got %.4f, want %.4f (bar 2 open via close_all, not stop 90)",
			trade.ExitPrice, wantExitPx)
	}
	if len(result.Strategy.OpenTrades) != 0 {
		t.Errorf("open trades: got %d, want 0", len(result.Strategy.OpenTrades))
	}
}
