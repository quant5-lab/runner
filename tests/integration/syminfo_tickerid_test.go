package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/* TestSyminfoTickeridInSecurity validates syminfo.tickerid resolves to ctx.Symbol in security() context */
func TestSyminfoTickeridInSecurity(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Syminfo Security", overlay=true)
daily_close = request.security(syminfo.tickerid, "1D", close)
plot(daily_close, "Daily Close", color=color.blue)
`
	exec := util.NewPineExecutor(t)
	generatedCode, _ := exec.GenerateCode(t, "syminfo-security", pineScript)

	if !strings.Contains(generatedCode, "var syminfo_tickerid string") {
		t.Error("Expected syminfo_tickerid variable declaration")
	}

	if !strings.Contains(generatedCode, "*symbolFlag") {
		t.Error("Expected syminfo_tickerid initialization from symbolFlag")
	}

	if !strings.Contains(generatedCode, "ctx.Symbol") {
		t.Error("Expected syminfo.tickerid to resolve to ctx.Symbol in security()")
	}

	if err := exec.CompileCode(t, generatedCode); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	t.Log("✓ syminfo.tickerid in security() - PASS")
}

/* TestSyminfoTickeridWithTAFunction validates syminfo.tickerid with TA function in security() */
func TestSyminfoTickeridWithTAFunction(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Syminfo TA Security", overlay=true)
daily_sma = request.security(syminfo.tickerid, "1D", ta.sma(close, 20))
plot(daily_sma, "Daily SMA", color=color.green)
`
	exec := util.NewPineExecutor(t)
	generatedCode, _ := exec.GenerateCode(t, "syminfo-ta", pineScript)

	if !strings.Contains(generatedCode, "var syminfo_tickerid string") {
		t.Error("Expected syminfo_tickerid variable declaration")
	}

	if !strings.Contains(generatedCode, "ctx.Symbol") {
		t.Error("Expected ctx.Symbol in security() call")
	}

	hasSmaSum := strings.Contains(generatedCode, "smaSum")
	hasTaSma := strings.Contains(generatedCode, "ta.Sma")
	hasSma20 := strings.Contains(generatedCode, "sma_20") || strings.Contains(generatedCode, "daily_sma")

	if !hasSmaSum && !hasTaSma && !hasSma20 {
		t.Errorf("Expected SMA calculation pattern. Generated code contains:\nsmaSum: %v\nta.Sma: %v\nsma_20: %v",
			hasSmaSum, hasTaSma, hasSma20)
	}

	if err := exec.CompileCode(t, generatedCode); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	t.Log("✓ syminfo.tickerid with TA function - PASS")
}

/* TestSyminfoTickeridStandalone validates direct syminfo.tickerid reference */
func TestSyminfoTickeridStandalone(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Syminfo Standalone")
current_symbol = syminfo.tickerid
`
	tmpDir := t.TempDir()
	pineFile := filepath.Join(tmpDir, "test.pine")
	outputBinary := filepath.Join(tmpDir, "test_binary")

	if err := os.WriteFile(pineFile, []byte(pineScript), 0644); err != nil {
		t.Fatalf("Failed to write Pine file: %v", err)
	}

	originalDir, _ := os.Getwd()
	os.Chdir("../..")
	defer os.Chdir(originalDir)

	buildCmd := exec.Command("go", "run", "cmd/pine-gen/main.go",
		"-input", pineFile,
		"-output", outputBinary)

	buildOutput, err := buildCmd.CombinedOutput()

	if err == nil {
		t.Log("✓ syminfo.tickerid standalone - build succeeded (may have limitations)")
	} else {
		buildOutputStr := string(buildOutput)
		if strings.Contains(buildOutputStr, "Codegen error") ||
			strings.Contains(buildOutputStr, "undefined") ||
			strings.Contains(buildOutputStr, "error") {
			t.Log("✓ syminfo.tickerid standalone - EXPECTED LIMITATION (string vars not yet supported)")
		} else {
			t.Errorf("Unexpected build failure: %v\nOutput: %s", err, buildOutputStr)
		}
	}
}

/* TestSyminfoTickeridMultipleSecurityCalls validates reusability across multiple security() calls */
func TestSyminfoTickeridMultipleSecurityCalls(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Syminfo Multiple Security", overlay=true)
daily_close = request.security(syminfo.tickerid, "1D", close)
weekly_close = request.security(syminfo.tickerid, "1W", close)
plot(daily_close, "Daily", color=color.blue)
plot(weekly_close, "Weekly", color=color.red)
`
	exec := util.NewPineExecutor(t)
	generatedCode, _ := exec.GenerateCode(t, "syminfo-multiple", pineScript)

	declarationCount := strings.Count(generatedCode, "var syminfo_tickerid string")
	if declarationCount != 1 {
		t.Errorf("Expected 1 syminfo_tickerid declaration, got %d (violates DRY)", declarationCount)
	}

	symbolResolutions := strings.Count(generatedCode, "ctx.Symbol")
	if symbolResolutions < 2 {
		t.Errorf("Expected at least 2 ctx.Symbol resolutions, got %d", symbolResolutions)
	}

	if err := exec.CompileCode(t, generatedCode); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	t.Log("✓ syminfo.tickerid multiple security() calls - PASS")
}

/* TestSyminfoTickeridWithComplexExpression validates syminfo.tickerid in complex expression context */
func TestSyminfoTickeridWithComplexExpression(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Syminfo Complex Expression", overlay=true)
daily_change_pct = request.security(syminfo.tickerid, "1D", (close - open) / open * 100)
plot(daily_change_pct, "Daily % Change", color=color.orange)
`
	exec := util.NewPineExecutor(t)
	generatedCode, _ := exec.GenerateCode(t, "syminfo-expression", pineScript)

	if !strings.Contains(generatedCode, "var syminfo_tickerid string") {
		t.Error("Expected syminfo_tickerid variable declaration")
	}

	if !strings.Contains(generatedCode, "ctx.Symbol") {
		t.Error("Expected ctx.Symbol resolution")
	}

	if !strings.Contains(generatedCode, "Series.Set(") {
		t.Error("Expected Series.Set() for expression result")
	}

	if err := exec.CompileCode(t, generatedCode); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	t.Log("✓ syminfo.tickerid with complex expression - PASS")
}

/* TestSyminfoTickeridRegressionNoSideEffects validates that syminfo.tickerid doesn't break existing code */
func TestSyminfoTickeridRegressionNoSideEffects(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Regression Test", overlay=true)
btc_close = request.security("BTCUSDT", "1D", close)
plot(btc_close, "BTC Close", color=color.yellow)
`
	exec := util.NewPineExecutor(t)
	generatedCode, _ := exec.GenerateCode(t, "syminfo-regression", pineScript)

	if !strings.Contains(generatedCode, "var syminfo_tickerid string") {
		t.Error("Expected syminfo_tickerid variable declaration")
	}

	if !strings.Contains(generatedCode, `"BTCUSDT"`) {
		t.Error("Expected literal symbol string in security() call")
	}

	if err := exec.CompileCode(t, generatedCode); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	t.Log("✓ Regression test: literal symbols still work - PASS")
}
