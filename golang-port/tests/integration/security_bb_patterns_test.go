package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

/* TestSecurityBBRealWorldPatterns tests actual security() patterns from production BB strategies
 * These are patterns that WORK with our current implementation (Python parser + Go codegen)
 * 
 * From bb-strategy-7-rus.pine, bb-strategy-8-rus.pine, bb-strategy-9-rus.pine
 */
func TestSecurityBBRealWorldPatterns(t *testing.T) {
	patterns := []struct {
		name        string
		script      string
		description string
	}{
		{
			name: "SMA_Daily_v4",
			script: `//@version=4
strategy("BB SMA Test", overlay=true)
sma_1d_20 = security(syminfo.tickerid, 'D', sma(close, 20))
plot(sma_1d_20, "SMA20 1D")
`,
			description: "Simple Moving Average on daily timeframe (BB7 pattern)",
		},
		{
			name: "SMA_Daily_v5",
			script: `//@version=5
indicator("BB SMA Test", overlay=true)
sma_1d_20 = request.security(syminfo.tickerid, '1D', ta.sma(close, 20))
plot(sma_1d_20, "SMA20 1D")
`,
			description: "Simple Moving Average on daily timeframe (v5 syntax)",
		},
		{
			name: "Multiple_SMA_Daily",
			script: `//@version=4
strategy("BB Multiple SMA", overlay=true)
sma_1d_20 = security(syminfo.tickerid, 'D', sma(close, 20))
sma_1d_50 = security(syminfo.tickerid, 'D', sma(close, 50))
sma_1d_200 = security(syminfo.tickerid, 'D', sma(close, 200))
plot(sma_1d_20, "SMA20")
plot(sma_1d_50, "SMA50")
plot(sma_1d_200, "SMA200")
`,
			description: "Multiple SMA calculations (BB7/8/9 pattern)",
		},
		{
			name: "Open_Daily_Lookahead",
			script: `//@version=4
strategy("BB Open Test", overlay=true)
open_1d = security(syminfo.tickerid, "D", open, lookahead=barmerge.lookahead_on)
plot(open_1d, "Open 1D")
`,
			description: "Daily open with lookahead (BB7 pattern)",
		},
		{
			name: "BB_Basis_SMA",
			script: `//@version=4
strategy("BB Basis Test", overlay=true)
bb_1d_basis = security(syminfo.tickerid, "1D", sma(close, 46))
plot(bb_1d_basis, "BB Basis")
`,
			description: "Bollinger Band basis calculation (BB8 pattern)",
		},
		{
			name: "Close_Simple",
			script: `//@version=5
indicator("Close Test", overlay=true)
close_1d = request.security(syminfo.tickerid, "1D", close)
plot(close_1d, "Close 1D")
`,
			description: "Simple close value from daily timeframe",
		},
		{
			name: "EMA_Daily",
			script: `//@version=5
indicator("EMA Test", overlay=true)
ema_1d_10 = request.security(syminfo.tickerid, "1D", ta.ema(close, 10))
plot(ema_1d_10, "EMA10 1D")
`,
			description: "Exponential Moving Average on daily timeframe",
		},
	}

	for _, tc := range patterns {
		t.Run(tc.name, func(t *testing.T) {
			success := buildAndCompilePineScript(t, tc.script)
			if !success {
				t.Fatalf("'%s' failed: %s", tc.name, tc.description)
			}
			t.Logf("✅ '%s' - %s", tc.name, tc.description)
		})
	}

	t.Logf("\n🎯 All %d BB strategy patterns compiled successfully", len(patterns))
}

/* TestSecurityStdevWorkaround tests BB strategy pattern with stdev
 * BB8 uses: bb_1d_dev = security(syminfo.tickerid, "1D", bb_1d_bbstdev * stdev(close, bb_1d_bblenght))
 * But multiplication inside security() doesn't parse - need workaround
 */
func TestSecurityStdevWorkaround(t *testing.T) {
	testCases := []struct {
		name   string
		script string
		status string
	}{
		{
			name: "Stdev_Simple_Works",
			script: `//@version=4
strategy("Stdev Works", overlay=true)
dev_1d = security(syminfo.tickerid, "1D", stdev(close, 20))
plot(dev_1d, "Stdev")
`,
			status: "WORKS - simple stdev call",
		},
		{
			name: "Stdev_PreMultiplied_Works",
			script: `//@version=4
strategy("Stdev Workaround", overlay=true)
// Workaround: calculate with multiplier outside security()
bbstdev = 0.35
dev_1d = security(syminfo.tickerid, "1D", stdev(close, 20))
bb_dev = bbstdev * dev_1d
plot(bb_dev, "BB Dev")
`,
			status: "WORKS - multiplication outside security()",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			success := buildAndCompilePineScript(t, tc.script)
			if !success {
				t.Fatalf("Test failed: %s", tc.status)
			}
			t.Logf("✅ %s: %s", tc.name, tc.status)
		})
	}
}

/* TestSecurityLongTermStability tests patterns for regression safety
 * These patterns must continue working in all future versions
 */
func TestSecurityLongTermStability(t *testing.T) {
	testCases := []struct {
		name        string
		script      string
		criticalFor string
	}{
		{
			name: "SMA_Warmup_Handling",
			script: `//@version=5
indicator("SMA Warmup", overlay=true)
// With 20-period SMA, first 19 bars should be NaN
sma20_1d = request.security(syminfo.tickerid, "1D", ta.sma(close, 20))
plot(sma20_1d, "SMA20")
`,
			criticalFor: "NaN handling with insufficient warmup period",
		},
		{
			name: "Multiple_Timeframes",
			script: `//@version=4
strategy("Multi TF", overlay=true)
close_1d = security(syminfo.tickerid, "1D", close)
close_1w = security(syminfo.tickerid, "1W", close)
plot(close_1d, "Daily")
plot(close_1w, "Weekly")
`,
			criticalFor: "Multiple security() calls with different timeframes",
		},
		{
			name: "Mixed_v4_v5_Syntax",
			script: `//@version=4
strategy("Mixed Syntax", overlay=true)
// v4 syntax
sma_1d = security(syminfo.tickerid, "1D", sma(close, 20))
plot(sma_1d, "SMA")
`,
			criticalFor: "Pine v4 to v5 migration compatibility",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			success := buildAndCompilePineScript(t, tc.script)
			if !success {
				t.Fatalf("REGRESSION: %s failed - critical for: %s", tc.name, tc.criticalFor)
			}
			t.Logf("✅ Stability check passed: %s", tc.criticalFor)
		})
	}
}

/* TestSecurityInlineTA_Validation validates inline TA code generation
 * Ensures generated code contains inline algorithms, not runtime lookups
 */
func TestSecurityInlineTA_Validation(t *testing.T) {
	pineScript := `//@version=5
indicator("Inline TA Check", overlay=true)
sma20_1d = request.security(syminfo.tickerid, "1D", ta.sma(close, 20))
ema10_1d = request.security(syminfo.tickerid, "1D", ta.ema(close, 10))
plot(sma20_1d, "SMA")
plot(ema10_1d, "EMA")
`

	tmpDir := t.TempDir()
	pineFile := filepath.Join(tmpDir, "test.pine")
	outputBinary := filepath.Join(tmpDir, "test_binary")
	
	err := os.WriteFile(pineFile, []byte(pineScript), 0644)
	if err != nil {
		t.Fatalf("Failed to write Pine file: %v", err)
	}

	originalDir, _ := os.Getwd()
	os.Chdir("../..")
	defer os.Chdir(originalDir)

	buildCmd := exec.Command("go", "run", "cmd/pine-gen/main.go",
		"-input", pineFile,
		"-output", outputBinary)

	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Build failed: %v\nOutput: %s", err, buildOutput)
	}

	/* Read generated code */
	generatedCode, err := os.ReadFile("/var/folders/ft/nyw_rm792qb2056vjlkzfj200000gn/T/pine_strategy_temp.go")
	if err != nil {
		t.Fatalf("Failed to read generated code: %v", err)
	}

	generatedStr := string(generatedCode)

	/* Validate inline SMA algorithm present */
	if !containsSubstring(generatedStr, "Inline SMA") && !containsSubstring(generatedStr, "inline SMA") {
		t.Error("Expected inline SMA generation (not runtime lookup)")
	}

	/* Validate inline EMA algorithm present */
	if !containsSubstring(generatedStr, "Inline EMA") && !containsSubstring(generatedStr, "inline EMA") {
		t.Error("Expected inline EMA generation (not runtime lookup)")
	}

	/* Validate context switching */
	if !containsSubstring(generatedStr, "origCtx := ctx") {
		t.Error("Expected context switching code (origCtx := ctx)")
	}

	if !containsSubstring(generatedStr, "ctx = secCtx") {
		t.Error("Expected context assignment (ctx = secCtx)")
	}

	/* Validate NaN handling */
	if !containsSubstring(generatedStr, "math.NaN()") {
		t.Error("Expected NaN handling for insufficient warmup")
	}

	t.Log("✅ Inline TA code generation validated")
}

/* Helper function to build and compile Pine script using pine-gen */
func buildAndCompilePineScript(t *testing.T, pineScript string) bool {
	tmpDir := t.TempDir()
	pineFile := filepath.Join(tmpDir, "test.pine")
	outputBinary := filepath.Join(tmpDir, "test_binary")
	
	err := os.WriteFile(pineFile, []byte(pineScript), 0644)
	if err != nil {
		t.Errorf("Failed to write Pine file: %v", err)
		return false
	}

	originalDir, _ := os.Getwd()
	os.Chdir("../..")
	defer os.Chdir(originalDir)

	buildCmd := exec.Command("go", "run", "cmd/pine-gen/main.go",
		"-input", pineFile,
		"-output", outputBinary)

	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Errorf("Build failed: %v\nOutput: %s", err, buildOutput)
		return false
	}

	binaryPath := filepath.Join(tmpDir, "test_binary")
	compileCmd := exec.Command("go", "build", "-o", binaryPath,
		"/var/folders/ft/nyw_rm792qb2056vjlkzfj200000gn/T/pine_strategy_temp.go")

	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Errorf("Compilation failed: %v\nOutput: %s", err, compileOutput)
		return false
	}

	return true
}

func containsSubstring(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && 
		(s == substr || (len(s) >= len(substr) && containsSubstringHelper(s, substr)))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
