package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

/* TestSecurityTACombination tests inline TA combination inside security()
 * Pattern: security(symbol, "1D", ta.sma(close, 20) + ta.ema(close, 10))
 * Critical for regression safety - ensures inline TA + binary operations work
 */
func TestSecurityTACombination(t *testing.T) {
	pineScript := `//@version=5
indicator("TA Combo Security", overlay=true)
combined = request.security(syminfo.tickerid, "1D", ta.sma(close, 20) + ta.ema(close, 10))
plot(combined, "Combined", color=color.blue)
`

	/* Write Pine script to temp file */
	tmpDir := t.TempDir()
	pineFile := filepath.Join(tmpDir, "test.pine")
	outputBinary := filepath.Join(tmpDir, "test_binary")

	err := os.WriteFile(pineFile, []byte(pineScript), 0644)
	if err != nil {
		t.Fatalf("Failed to write Pine file: %v", err)
	}

	/* Build using pine-gen */
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

	/* Read generated code to validate inline TA */
	generatedCode, err := os.ReadFile("/var/folders/ft/nyw_rm792qb2056vjlkzfj200000gn/T/pine_strategy_temp.go")
	if err != nil {
		t.Fatalf("Failed to read generated code: %v", err)
	}

	generatedStr := string(generatedCode)

	/* Validate inline SMA and EMA present */
	if !contains(generatedStr, "Inline SMA") && !contains(generatedStr, "inline SMA") {
		t.Error("Expected inline SMA generation in security context")
	}

	if !contains(generatedStr, "Inline EMA") && !contains(generatedStr, "inline EMA") {
		t.Error("Expected inline EMA generation in security context")
	}

	/* Compile the generated code */
	binaryPath := filepath.Join(tmpDir, "test_binary")
	compileCmd := exec.Command("go", "build", "-o", binaryPath,
		"/var/folders/ft/nyw_rm792qb2056vjlkzfj200000gn/T/pine_strategy_temp.go")

	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Compilation failed: %v\nOutput: %s", err, compileOutput)
	}

	t.Log("✅ TA combination security() compiled successfully")
}

/* TestSecurityArithmeticExpression tests arithmetic expressions inside security()
 * Pattern: security(symbol, "1D", (high - low) / close * 100)
 * Critical for regression safety - ensures binary operations work in security context
 */
func TestSecurityArithmeticExpression(t *testing.T) {
	pineScript := `//@version=5
indicator("Arithmetic Security", overlay=true)
volatility = request.security(syminfo.tickerid, "1D", (high - low) / close * 100)
plot(volatility, "Volatility %", color=color.red)
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

	generatedCode, err := os.ReadFile("/var/folders/ft/nyw_rm792qb2056vjlkzfj200000gn/T/pine_strategy_temp.go")
	if err != nil {
		t.Fatalf("Failed to read generated code: %v", err)
	}

	generatedStr := string(generatedCode)

	/* Validate field access (high, low, close) */
	requiredFields := []string{"High", "Low", "Close"}
	for _, field := range requiredFields {
		if !contains(generatedStr, field) {
			t.Errorf("Expected field '%s' access in generated code", field)
		}
	}

	/* Compile the generated code */
	binaryPath := filepath.Join(tmpDir, "test_binary")
	compileCmd := exec.Command("go", "build", "-o", binaryPath,
		"/var/folders/ft/nyw_rm792qb2056vjlkzfj200000gn/T/pine_strategy_temp.go")

	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Compilation failed: %v\nOutput: %s", err, compileOutput)
	}

	t.Log("✅ Arithmetic expression security() compiled successfully")
}

/* TestSecurityBBStrategy7Patterns tests real-world patterns from bb-strategy-7-rus.pine
 * Validates all security() patterns used in production strategy
 */
func TestSecurityBBStrategy7Patterns(t *testing.T) {
	patterns := []struct {
		name   string
		script string
	}{
		{
			name: "SMA on daily timeframe",
			script: `//@version=4
strategy("BB7-SMA", overlay=true)
sma_1d_20 = security(syminfo.tickerid, 'D', sma(close, 20))
plot(sma_1d_20)`,
		},
		{
			name: "ATR on daily timeframe",
			script: `//@version=4
strategy("BB7-ATR", overlay=true)
atr_1d = security(syminfo.tickerid, "1D", atr(14))
plot(atr_1d)`,
		},
		{
			name: "Open with lookahead",
			script: `//@version=4
strategy("BB7-Open", overlay=true)
open_1d = security(syminfo.tickerid, "D", open, lookahead=barmerge.lookahead_on)
plot(open_1d)`,
		},
	}

	for _, tc := range patterns {
		t.Run(tc.name, func(t *testing.T) {
			success := buildAndCompilePine(t, tc.script)
			if !success {
				t.Fatalf("Pattern '%s' failed", tc.name)
			}
			t.Logf("✅ BB7 pattern '%s' compiled successfully", tc.name)
		})
	}
}

/* TestSecurityBBStrategy8Patterns tests real-world patterns from bb-strategy-8-rus.pine
 * Includes complex expressions with stdev, comparisons, valuewhen
 */
func TestSecurityBBStrategy8Patterns(t *testing.T) {
	patterns := []struct {
		name   string
		script string
	}{
		{
			name: "BB basis with SMA",
			script: `//@version=4
strategy("BB8-Basis", overlay=true)
bb_1d_basis = security(syminfo.tickerid, "1D", sma(close, 46))
plot(bb_1d_basis)`,
		},
		{
			name: "BB deviation with stdev multiplication",
			script: `//@version=4
strategy("BB8-Dev", overlay=true)
bb_1d_dev = security(syminfo.tickerid, "1D", 0.35 * stdev(close, 46))
plot(bb_1d_dev)`,
		},
	}

	for _, tc := range patterns {
		t.Run(tc.name, func(t *testing.T) {
			success := buildAndCompilePine(t, tc.script)
			if !success {
				t.Fatalf("Pattern '%s' failed", tc.name)
			}
			t.Logf("✅ BB8 pattern '%s' compiled successfully", tc.name)
		})
	}
}

/* TestSecurityStability_RegressionSuite comprehensive regression test suite
 * Ensures all complex expression types continue to work
 */
func TestSecurityStability_RegressionSuite(t *testing.T) {
	testCases := []struct {
		name        string
		script      string
		description string
	}{
		{
			name: "TA_Combo_Add",
			script: `//@version=5
indicator("Test")
result = request.security(syminfo.tickerid, "1D", ta.sma(close, 20) + ta.ema(close, 10))
plot(result)`,
			description: "SMA + EMA combination",
		},
		{
			name: "TA_Combo_Subtract",
			script: `//@version=5
indicator("Test")
result = request.security(syminfo.tickerid, "1D", ta.sma(close, 20) - ta.ema(close, 10))
plot(result)`,
			description: "SMA - EMA subtraction",
		},
		{
			name: "TA_Combo_Multiply",
			script: `//@version=5
indicator("Test")
result = request.security(syminfo.tickerid, "1D", ta.sma(close, 20) * 1.5)
plot(result)`,
			description: "SMA multiplication by constant",
		},
		{
			name: "Arithmetic_HighLow",
			script: `//@version=5
indicator("Test")
result = request.security(syminfo.tickerid, "1D", (high - low) / close * 100)
plot(result)`,
			description: "High-Low volatility percentage",
		},
		{
			name: "Arithmetic_OHLC",
			script: `//@version=5
indicator("Test")
result = request.security(syminfo.tickerid, "1D", (open + high + low + close) / 4)
plot(result)`,
			description: "OHLC average",
		},
		{
			name: "Stdev_Multiplication",
			script: `//@version=4
strategy("Test", overlay=true)
dev = security(syminfo.tickerid, "1D", 2.0 * stdev(close, 20))
plot(dev)`,
			description: "Stdev with constant multiplication (BB pattern)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			success := buildAndCompilePine(t, tc.script)
			if !success {
				t.Fatalf("'%s' failed: %s", tc.name, tc.description)
			}
			t.Logf("✅ '%s' - %s", tc.name, tc.description)
		})
	}

	t.Logf("\n🎯 All %d regression test cases passed", len(testCases))
}

/* TestSecurityNaN_Handling ensures NaN values are handled correctly
 * Critical for long-term stability - avoid crashes with insufficient data
 */
func TestSecurityNaN_Handling(t *testing.T) {
	pineScript := `//@version=5
indicator("NaN Test", overlay=true)
sma20 = request.security(syminfo.tickerid, "1D", ta.sma(close, 20))
plot(sma20, "SMA20")`

	success := buildAndCompilePine(t, pineScript)
	if !success {
		t.Fatal("NaN handling test failed")
	}

	/* Read generated code to validate NaN handling */
	generatedCode, err := os.ReadFile("/var/folders/ft/nyw_rm792qb2056vjlkzfj200000gn/T/pine_strategy_temp.go")
	if err != nil {
		t.Fatalf("Failed to read generated code: %v", err)
	}

	if !contains(string(generatedCode), "math.NaN()") {
		t.Error("Expected NaN handling in generated code for insufficient warmup")
	}

	t.Log("✅ NaN handling compiled successfully")
}

/* Helper function to build and compile Pine script using pine-gen */
func buildAndCompilePine(t *testing.T, pineScript string) bool {
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

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || (len(s) >= len(substr) && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
