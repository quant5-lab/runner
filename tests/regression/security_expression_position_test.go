package regression

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

/* Security in expression position must route through full evaluator, not bare IIFE */
func TestSecurity_ExpressionPosition_TernaryWithUserVariable(t *testing.T) {
	testDir := t.TempDir()

	strategy := `//@version=4
strategy("Security Expression Position", overlay=true)

threshold = sma(close, 5)
signal = security(syminfo.tickerid, "1D", close) > threshold ? 1 : 0

plot(signal, "Signal")
`
	strategyPath := filepath.Join(testDir, "test_strategy.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}

	hourlyData := generateTestOHLCV(240, 3600)
	hourlyPath := filepath.Join(testDir, "EXPRTEST_1h.json")
	if err := os.WriteFile(hourlyPath, []byte(hourlyData), 0644); err != nil {
		t.Fatal(err)
	}

	dailyData := generateTestOHLCV(10, 86400)
	dailyPath := filepath.Join(testDir, "EXPRTEST_1D.json")
	if err := os.WriteFile(dailyPath, []byte(dailyData), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, hourlyPath, testDir, projectRoot, "EXPRTEST", testDir)

	signal, ok := result.Indicators["Signal"]
	if !ok {
		t.Fatalf("Expected 'Signal' indicator, got: %v", getIndicatorNames(result.Indicators))
	}

	nonNull := countNonNull(signal.Data)
	if nonNull < 200 {
		t.Errorf("Expression-position security produced only %d non-null values, expected >200 from 240 bars", nonNull)
	}

	for i, bar := range signal.Data {
		if val, ok := getFloatValue(bar); ok {
			if val != 0.0 && val != 1.0 {
				t.Errorf("Bar %d: ternary signal %.2f outside valid set {0, 1}", i, val)
				break
			}
		}
	}
}

/* Security with comparison expression argument hoisted through full evaluator */
func TestSecurity_ExpressionPosition_ComparisonArgument(t *testing.T) {
	testDir := t.TempDir()

	strategy := `//@version=4
strategy("Security Comparison Arg", overlay=true)

bb_upper = sma(close, 20) + 2 * stdev(close, 20)
is_above = security(syminfo.tickerid, "1D", close > bb_upper) ? 1 : 0

plot(is_above, "Above BB")
`
	strategyPath := filepath.Join(testDir, "test_strategy.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}

	hourlyData := generateTestOHLCV(240, 3600)
	hourlyPath := filepath.Join(testDir, "CMPTEST_1h.json")
	if err := os.WriteFile(hourlyPath, []byte(hourlyData), 0644); err != nil {
		t.Fatal(err)
	}

	dailyData := generateTestOHLCV(10, 86400)
	dailyPath := filepath.Join(testDir, "CMPTEST_1D.json")
	if err := os.WriteFile(dailyPath, []byte(dailyData), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, hourlyPath, testDir, projectRoot, "CMPTEST", testDir)

	above, ok := result.Indicators["Above BB"]
	if !ok {
		t.Fatalf("Expected 'Above BB' indicator, got: %v", getIndicatorNames(result.Indicators))
	}

	nonNull := countNonNull(above.Data)
	if nonNull < 200 {
		t.Errorf("Comparison-arg security produced only %d non-null values, expected >200 from 240 bars", nonNull)
	}
}

/* Multiple security calls in expression position with shared user variables */
func TestSecurity_ExpressionPosition_MultipleCallsSharedVariable(t *testing.T) {
	testDir := t.TempDir()

	strategy := `//@version=4
strategy("Security Multi Expression", overlay=true)

bb_len = 20
bb_basis = sma(close, bb_len)
bb_dev = stdev(close, bb_len)
bb_upper = bb_basis + 2 * bb_dev
bb_lower = bb_basis - 2 * bb_dev

above_upper = security(syminfo.tickerid, "1D", close > bb_upper) ? 1 : 0
below_lower = security(syminfo.tickerid, "1D", close < bb_lower) ? 1 : 0
combined = above_upper + below_lower

plot(combined, "BB Signal")
`
	strategyPath := filepath.Join(testDir, "test_strategy.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}

	hourlyData := generateTestOHLCV(240, 3600)
	hourlyPath := filepath.Join(testDir, "MULTITEST_1h.json")
	if err := os.WriteFile(hourlyPath, []byte(hourlyData), 0644); err != nil {
		t.Fatal(err)
	}

	dailyData := generateTestOHLCV(10, 86400)
	dailyPath := filepath.Join(testDir, "MULTITEST_1D.json")
	if err := os.WriteFile(dailyPath, []byte(dailyData), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, hourlyPath, testDir, projectRoot, "MULTITEST", testDir)

	signal, ok := result.Indicators["BB Signal"]
	if !ok {
		t.Fatalf("Expected 'BB Signal' indicator, got: %v", getIndicatorNames(result.Indicators))
	}

	nonNull := countNonNull(signal.Data)
	if nonNull < 200 {
		t.Errorf("Multi-security expression produced only %d non-null values, expected >200 from 240 bars", nonNull)
	}

	for i, bar := range signal.Data {
		if val, ok := getFloatValue(bar); ok {
			if val < 0 || val > 2 {
				t.Errorf("Bar %d: combined signal %.1f outside valid range [0, 2]", i, val)
				break
			}
		}
	}
}

/* Security in binary expression (addition) must hoist through full evaluator */
func TestSecurity_ExpressionPosition_BinaryArithmetic(t *testing.T) {
	testDir := t.TempDir()

	strategy := `//@version=4
strategy("Security Binary Arithmetic", overlay=true)

daily_close = security(syminfo.tickerid, "1D", close)
daily_open = security(syminfo.tickerid, "1D", open)
spread = daily_close - daily_open + security(syminfo.tickerid, "1D", high) * 0.5

plot(spread, "Spread")
`
	strategyPath := filepath.Join(testDir, "test_strategy.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}

	hourlyData := generateTestOHLCV(240, 3600)
	hourlyPath := filepath.Join(testDir, "BINTEST_1h.json")
	if err := os.WriteFile(hourlyPath, []byte(hourlyData), 0644); err != nil {
		t.Fatal(err)
	}

	dailyData := generateTestOHLCV(10, 86400)
	dailyPath := filepath.Join(testDir, "BINTEST_1D.json")
	if err := os.WriteFile(dailyPath, []byte(dailyData), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, hourlyPath, testDir, projectRoot, "BINTEST", testDir)

	spread, ok := result.Indicators["Spread"]
	if !ok {
		t.Fatalf("Expected 'Spread' indicator, got: %v", getIndicatorNames(result.Indicators))
	}

	nonNull := countNonNull(spread.Data)
	if nonNull < 200 {
		t.Errorf("Binary arithmetic security produced only %d non-null values, expected >200 from 240 bars", nonNull)
	}

	for i, bar := range spread.Data {
		if val, ok := getFloatValue(bar); ok {
			if val == 0.0 {
				continue
			}
			if val < -100000 || val > 200000 {
				t.Errorf("Bar %d: spread value %.2f outside reasonable range", i, val)
				break
			}
		}
	}
}

/* Verifies generated code structure for expression-position security */
func TestSecurity_ExpressionPosition_CodegenHoisting(t *testing.T) {
	testDir := t.TempDir()

	strategy := `//@version=4
strategy("Codegen Verify", overlay=true)

myVar = sma(close, 10)
result = security(syminfo.tickerid, "1D", close > myVar) ? 1 : 0

plot(result, "Result")
`
	strategyPath := filepath.Join(testDir, "test_strategy.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))
	builderPath := filepath.Join(projectRoot, "cmd", "pine-gen", "main.go")
	templatePath := filepath.Join(projectRoot, "template", "main.go.tmpl")
	outputGoPath := filepath.Join(testDir, "output.go")

	compileCmd := exec.Command(
		"go", "run", builderPath,
		"-input", strategyPath,
		"-output", outputGoPath,
		"-template", templatePath,
	)
	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Pine compilation failed: %v\n%s", err, compileOutput)
	}

	generatedFile := extractGeneratedPath(t, compileOutput)

	generatedCode, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	code := string(generatedCode)

	if !strings.Contains(code, "request_security_") && !strings.Contains(code, "Series.Set(") {
		t.Error("Expression-position security should produce hoisted temp var with Series.Set pattern")
	}

	if !strings.Contains(code, "SetVariableRegistry") {
		t.Error("Hoisted security must use full evaluator with SetVariableRegistry")
	}

	if !strings.Contains(code, "SetBarIndexMapper") || !strings.Contains(code, "NewBarIndexMapper") {
		t.Error("Hoisted security must use BarIndexMapper for variable resolution")
	}
}

func extractGeneratedPath(t *testing.T, output []byte) string {
	t.Helper()
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "Generated: ") {
			return filepath.Clean(strings.TrimPrefix(line, "Generated: "))
		}
	}
	t.Fatalf("Failed to parse generated file path from output: %s", output)
	return ""
}
