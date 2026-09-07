package testutil

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type StrategyRunner struct {
	pineGenPath   string
	workspaceRoot string
	tempDir       string
}

func NewStrategyRunner(t *testing.T) *StrategyRunner {
	t.Helper()

	workspaceRoot, err := findWorkspaceRoot()
	if err != nil {
		t.Fatalf("Find workspace root: %v", err)
	}

	return &StrategyRunner{
		pineGenPath:   filepath.Join(workspaceRoot, "build", "pine-gen"),
		workspaceRoot: workspaceRoot,
		tempDir:       t.TempDir(),
	}
}

func (r *StrategyRunner) Execute(t *testing.T, strategyPath, dataPath, symbol, timeframe string) *StrategyResult {
	t.Helper()

	binaryPath := filepath.Join(r.tempDir, "strategy-bin")
	outputPath := filepath.Join(r.tempDir, "output.json")

	goSourcePath := r.generateGoCode(t, strategyPath, binaryPath)
	r.compileStrategy(t, goSourcePath, binaryPath)
	r.runStrategy(t, binaryPath, dataPath, outputPath, symbol, timeframe)

	return r.parseOutput(t, outputPath)
}

func (r *StrategyRunner) generateGoCode(t *testing.T, strategyPath, binaryPath string) string {
	t.Helper()

	cmd := exec.Command(r.pineGenPath, "-input", strategyPath, "-output", binaryPath)
	cmd.Dir = r.workspaceRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Generate Go code: %v\nOutput: %s", err, output)
	}

	goFilePath := extractGeneratedFilePath(string(output))
	if goFilePath == "" {
		t.Fatalf("Generated Go file path not found in output:\n%s", output)
	}

	return goFilePath
}

func (r *StrategyRunner) compileStrategy(t *testing.T, goSourcePath, binaryPath string) {
	t.Helper()

	cmd := exec.Command("go", "build", "-o", binaryPath, goSourcePath)
	cmd.Dir = r.workspaceRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Compile strategy: %v\nOutput: %s", err, output)
	}
}

func (r *StrategyRunner) runStrategy(t *testing.T, binaryPath, dataPath, outputPath, symbol, timeframe string) {
	t.Helper()

	dataDir := filepath.Dir(dataPath)

	cmd := exec.Command(binaryPath,
		"-symbol", symbol,
		"-timeframe", timeframe,
		"-data", dataPath,
		"-datadir", dataDir,
		"-output", outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Execute strategy: %v\nOutput: %s", err, output)
	}
}

func (r *StrategyRunner) parseOutput(t *testing.T, outputPath string) *StrategyResult {
	t.Helper()

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Read output file: %v", err)
	}

	var chartOutput ChartOutput
	if err := json.Unmarshal(data, &chartOutput); err != nil {
		t.Fatalf("Parse output JSON: %v", err)
	}

	if chartOutput.Strategy == nil {
		t.Fatal("No strategy data in output")
	}

	if chartOutput.Strategy.Plots == nil {
		chartOutput.Strategy.Plots = make(map[string][]float64)
	}

	for _, plot := range chartOutput.Plots {
		chartOutput.Strategy.Plots[plot.Title] = plot.Values
	}

	chartOutput.Strategy.Indicators = make(map[string][]float64, len(chartOutput.Indicators))
	for name, ind := range chartOutput.Indicators {
		chartOutput.Strategy.Indicators[name] = indicatorToFloats(ind)
	}

	if n := len(chartOutput.Candlestick); n > 0 {
		chartOutput.Strategy.MarkClose = chartOutput.Candlestick[n-1].Close
		chartOutput.Strategy.FromRunner = true
	}

	return chartOutput.Strategy
}

func indicatorToFloats(ind IndicatorSeries) []float64 {
	out := make([]float64, len(ind.Data))
	for i, bar := range ind.Data {
		if bar.Value != nil {
			out[i] = *bar.Value
		} else {
			out[i] = math.NaN()
		}
	}
	return out
}

func extractGeneratedFilePath(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Generated:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1]
			}
		}
	}
	return ""
}

func FindWorkspaceRoot() (string, error) { return findWorkspaceRoot() }

func findWorkspaceRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("workspace root not found (no go.mod)")
		}
		dir = parent
	}
}
