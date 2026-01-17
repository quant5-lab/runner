package strategyintegration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type StrategyTestResult struct {
	Trades      []Trade `json:"trades"`
	OpenTrades  []Trade `json:"openTrades"`
	Equity      float64 `json:"equity"`
	NetProfit   float64 `json:"netProfit"`
	TotalTrades int     `json:"totalTrades"`
}

type Trade struct {
	EntryID      string  `json:"entryId"`
	EntryPrice   float64 `json:"entryPrice"`
	EntryBar     int     `json:"entryBar"`
	EntryTime    int64   `json:"entryTime"`
	EntryComment string  `json:"entryComment"`
	ExitPrice    float64 `json:"exitPrice"`
	ExitBar      int     `json:"exitBar"`
	ExitTime     int64   `json:"exitTime"`
	ExitComment  string  `json:"exitComment"`
	Size         float64 `json:"size"`
	Profit       float64 `json:"profit"`
	Direction    string  `json:"direction"`
}

type ChartData struct {
	Strategy *StrategyTestResult `json:"strategy"`
}

/* StrategyTestCase defines a single isolated strategy test */
type StrategyTestCase struct {
	Name           string
	PineFile       string
	DataFile       string
	ValidateTrades func(t *testing.T, result *StrategyTestResult)
}

/* runStrategyTest executes Pine→Go→Binary→JSON pipeline */
func runStrategyTest(t *testing.T, tc StrategyTestCase) *StrategyTestResult {
	t.Helper()

	baseDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Get working directory: %v", err)
	}

	golangPortDir := filepath.Join(baseDir, "../..")
	pineGenPath := filepath.Join(golangPortDir, "pine-gen")
	pineFile := filepath.Join(golangPortDir, "tests", "fixtures", "strategy", tc.PineFile)
	dataFile := filepath.Join(baseDir, "testdata", tc.DataFile)

	binaryPath := filepath.Join(os.TempDir(), "strategy-test-"+tc.Name)
	outputPath := filepath.Join(os.TempDir(), "strategy-output-"+tc.Name+".json")

	genCmd := exec.Command(pineGenPath, "-input", pineFile, "-output", binaryPath)
	genCmd.Dir = golangPortDir
	genOut, err := genCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("pine-gen failed: %v\nOutput: %s", err, genOut)
	}

	var generatedGoFile string
	outputLines := strings.Split(string(genOut), "\n")
	for _, line := range outputLines {
		if strings.HasPrefix(line, "Generated:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				generatedGoFile = parts[1]
				break
			}
		}
	}
	if generatedGoFile == "" {
		t.Fatalf("Could not find generated Go file in output:\n%s", genOut)
	}

	compileCmd := exec.Command("go", "build", "-o", binaryPath, generatedGoFile)
	compileCmd.Dir = golangPortDir
	compileOut, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Compile failed: %v\nOutput: %s", err, compileOut)
	}

	execCmd := exec.Command(binaryPath, "-symbol", "TEST", "-data", dataFile, "-output", outputPath)
	execOut, err := execCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Strategy execution failed: %v\nOutput: %s", err, execOut)
	}

	jsonBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Read output JSON: %v", err)
	}

	var chartData ChartData
	if err := json.Unmarshal(jsonBytes, &chartData); err != nil {
		t.Fatalf("Parse JSON: %v", err)
	}

	if chartData.Strategy == nil {
		t.Fatal("No strategy data in output")
	}

	return chartData.Strategy
}

/* TestEntryBasic verifies entry orders execute on next bar */
func TestEntryBasic(t *testing.T) {
	tc := StrategyTestCase{
		Name:     "entry-basic",
		PineFile: "test-entry-basic.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			if len(result.Trades)+len(result.OpenTrades) < 1 {
				t.Errorf("Expected at least 1 trade, got %d trades + %d open",
					len(result.Trades), len(result.OpenTrades))
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

/* TestExitStop verifies stop loss triggers when barLow reaches stop level */
func TestExitStop(t *testing.T) {
	tc := StrategyTestCase{
		Name:     "exit-stop",
		PineFile: "test-exit-stop.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			if len(result.Trades) < 1 {
				t.Errorf("Expected at least 1 closed trade from stop trigger, got %d", len(result.Trades))
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

/* TestExitLimit verifies take profit triggers when barHigh reaches limit level */
func TestExitLimit(t *testing.T) {
	tc := StrategyTestCase{
		Name:     "exit-limit",
		PineFile: "test-exit-limit.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			if len(result.Trades) < 1 {
				t.Errorf("Expected at least 1 closed trade from limit trigger, got %d", len(result.Trades))
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

/* TestEntryShort validates short position entry and negative position tracking */
func TestEntryShort(t *testing.T) {
	tc := StrategyTestCase{
		Name:     "entry-short",
		PineFile: "test-entry-short.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			if len(result.OpenTrades) < 1 {
				t.Fatal("Expected at least 1 open short trade")
			}
			trade := result.OpenTrades[0]
			if trade.Direction != "short" {
				t.Errorf("Expected direction 'short', got '%s'", trade.Direction)
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

/* TestPositionReversal validates long→short transition closes long and opens short */
func TestPositionReversal(t *testing.T) {
	tc := StrategyTestCase{
		Name:     "position-reversal",
		PineFile: "test-position-reversal.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			if len(result.Trades) < 2 {
				t.Errorf("Expected at least 2 closed trades from reversals, got %d", len(result.Trades))
			}
			for i := 1; i < len(result.Trades); i++ {
				if result.Trades[i].Direction == result.Trades[i-1].Direction {
					t.Error("Expected alternating directions in position reversals")
					break
				}
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

/* TestExitStopAndLimit validates both stop and limit set simultaneously */
func TestExitStopAndLimit(t *testing.T) {
	tc := StrategyTestCase{
		Name:     "exit-stop-and-limit",
		PineFile: "test-exit-stop-and-limit.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			if len(result.Trades) < 1 {
				t.Error("Expected at least 1 closed trade")
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

/* TestCloseAll validates strategy.close_all() closes all open positions */
func TestCloseAll(t *testing.T) {
	tc := StrategyTestCase{
		Name:     "close-all",
		PineFile: "test-close-all.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			if len(result.Trades) < 1 {
				t.Error("Expected at least 1 closed trade from close_all")
			}
			if len(result.OpenTrades) > 0 {
				t.Errorf("Expected 0 open trades after close_all, got %d", len(result.OpenTrades))
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

/* TestEntryMultiple validates pyramiding/scaling into positions */
func TestEntryMultiple(t *testing.T) {
	tc := StrategyTestCase{
		Name:     "entry-multiple",
		PineFile: "test-entry-multiple.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			totalSize := 0.0
			for _, trade := range result.OpenTrades {
				totalSize += trade.Size
			}
			if totalSize < 2 {
				t.Errorf("Expected multiple entries (total size >= 2), got total size %.0f", totalSize)
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

/* TestLogicalOR validates strategy.position_size in OR logical expressions */
func TestLogicalOR(t *testing.T) {
	tc := StrategyTestCase{
		Name:     "logical-or",
		PineFile: "test-logical-or.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			if len(result.OpenTrades) < 1 {
				t.Error("Expected at least 1 open trade from OR condition")
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

/* TestAvgPriceCondition validates strategy.position_avg_price in logical expressions */
func TestAvgPriceCondition(t *testing.T) {
	tc := StrategyTestCase{
		Name:     "avg-price-condition",
		PineFile: "test-avg-price-condition.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			if len(result.Trades) < 1 {
				t.Error("Expected at least 1 closed trade from avg_price condition")
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

/* TestExactPriceTrigger validates stop/limit triggered on exact price boundaries */
func TestExactPriceTrigger(t *testing.T) {
	tc := StrategyTestCase{
		Name:     "exact-price-trigger",
		PineFile: "test-exact-price-trigger.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			if len(result.Trades) < 1 {
				t.Error("Expected at least 1 closed trade from exact price trigger")
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

/* TestEquityWithUnrealized validates equity calculation includes open position P&L */
func TestEquityWithUnrealized(t *testing.T) {
	tc := StrategyTestCase{
		Name:     "entry-basic",
		PineFile: "test-entry-basic.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			if len(result.OpenTrades) == 0 {
				t.Skip("No open trades to test unrealized P&L")
			}
			if result.Equity == 10000 {
				t.Error("Expected equity != 10000 with open position")
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

/* TestNetProfitAccumulation validates net profit sums all closed trade profits */
func TestNetProfitAccumulation(t *testing.T) {
	tc := StrategyTestCase{
		Name:     "exit-limit",
		PineFile: "test-exit-limit.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			if len(result.Trades) == 0 {
				t.Skip("No closed trades to validate net profit")
			}
			expectedProfit := 0.0
			for _, trade := range result.Trades {
				expectedProfit += trade.Profit
			}
			if result.NetProfit != expectedProfit {
				t.Errorf("Expected net profit %.2f, got %.2f", expectedProfit, result.NetProfit)
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

/* TestCommentIntegration verifies end-to-end comment propagation from PineScript to JSON */
func TestCommentIntegration(t *testing.T) {
	tc := StrategyTestCase{
		Name:     "comment-integration",
		PineFile: "test-comment-integration.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			if len(result.Trades) < 1 {
				t.Fatalf("Expected at least 1 closed trade, got %d", len(result.Trades))
			}

			/* Verify entry and exit comments present */
			trade := result.Trades[0]
			if trade.EntryComment == "" {
				t.Errorf("Expected non-empty entry comment, got empty string")
			}
			if trade.ExitComment == "" {
				t.Errorf("Expected non-empty exit comment, got empty string")
			}

			/* Verify specific comment strings */
			if trade.EntryComment != "Bullish candle entry" {
				t.Errorf("Expected entry comment 'Bullish candle entry', got %q", trade.EntryComment)
			}
			if trade.ExitComment != "Position close" {
				t.Errorf("Expected exit comment 'Position close', got %q", trade.ExitComment)
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

/* TestEntryWhenTrue validates conditional entry execution with when parameter */
func TestEntryWhenTrue(t *testing.T) {
	tc := StrategyTestCase{
		Name:     "entry-when-true",
		PineFile: "test-entry-when-true.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			totalTrades := len(result.Trades) + len(result.OpenTrades)
			if totalTrades < 1 {
				t.Errorf("Expected at least 1 trade (when condition met), got %d", totalTrades)
				return
			}

			/* Entry should happen when buySignal becomes true (close > 105) */
			var trade Trade
			if len(result.Trades) > 0 {
				trade = result.Trades[0]
			} else {
				trade = result.OpenTrades[0]
			}

			if trade.EntryBar < 1 {
				t.Errorf("Entry bar %d too early (expected >= 1 when condition becomes true)", trade.EntryBar)
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

/* TestEntryWhenFalse validates entry suppression with always-false when condition */
func TestEntryWhenFalse(t *testing.T) {
	tc := StrategyTestCase{
		Name:     "entry-when-false",
		PineFile: "test-entry-when-false.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			/* Entry with when=false should never execute */
			totalTrades := len(result.Trades) + len(result.OpenTrades)
			if totalTrades != 0 {
				t.Errorf("Expected 0 trades with when=false, got %d", totalTrades)
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

/* TestEntryWhenMultiple validates different entries with different when conditions */
func TestEntryWhenMultiple(t *testing.T) {
	tc := StrategyTestCase{
		Name:     "entry-when-multiple",
		PineFile: "test-entry-when-multiple.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			/* Should have both long and short entries based on different conditions */
			if len(result.Trades) < 2 {
				t.Errorf("Expected at least 2 trades (long and short with different when), got %d", len(result.Trades))
			}

			/* Verify both directions present */
			hasLong := false
			hasShort := false
			for _, trade := range result.Trades {
				if trade.Direction == "long" {
					hasLong = true
				}
				if trade.Direction == "short" {
					hasShort = true
				}
			}

			if !hasLong {
				t.Error("Expected at least 1 long trade from longCondition")
			}
			if !hasShort {
				t.Error("Expected at least 1 short trade from shortCondition")
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

/* TestExitDynamicLevels validates exit level updates each bar (trailing stop pattern) */
func TestExitDynamicLevels(t *testing.T) {
	tc := StrategyTestCase{
		Name:     "exit-dynamic-levels",
		PineFile: "test-exit-dynamic-levels.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			/* Exit levels update each bar - should use latest level when triggered */
			if len(result.Trades) < 1 {
				t.Error("Expected at least 1 closed trade from dynamic stop level")
			}

			/* Verify exit triggered (not still open) */
			for _, trade := range result.Trades {
				if trade.ExitBar == 0 {
					t.Error("Trade has exitBar=0, exit did not trigger properly")
				}
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

/* TestExitPersistenceMultibar validates exit orders persist across multiple bars until triggered */
func TestExitPersistenceMultibar(t *testing.T) {
	tc := StrategyTestCase{
		Name:     "exit-persistence-multibar",
		PineFile: "test-exit-persistence-multibar.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			if len(result.Trades) < 1 {
				t.Error("Expected at least 1 closed trade from persistent exit order")
				return
			}

			/* Verify exit happened (not same bar as entry) */
			for _, trade := range result.Trades {
				if trade.ExitBar == trade.EntryBar {
					t.Errorf("Exit bar %d same as entry bar %d (expected persistence across bars)",
						trade.ExitBar, trade.EntryBar)
				}
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}

/* TestEntryWhenComplex validates when parameter with complex boolean expressions */
func TestEntryWhenComplex(t *testing.T) {
	tc := StrategyTestCase{
		Name:     "entry-when-complex",
		PineFile: "test-entry-when-complex.pine",
		DataFile: "simple-bars.json",
		ValidateTrades: func(t *testing.T, result *StrategyTestResult) {
			/* Complex when expression should gate entry correctly */
			if len(result.Trades)+len(result.OpenTrades) < 1 {
				t.Errorf("Expected at least 1 trade (complex when condition met), got %d trades + %d open",
					len(result.Trades), len(result.OpenTrades))
			}
		},
	}

	result := runStrategyTest(t, tc)
	tc.ValidateTrades(t, result)
}
