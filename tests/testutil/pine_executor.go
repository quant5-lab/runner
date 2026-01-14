package testutil

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// PineExecutor runs PineScript through full Pine→Go→Binary→JSON pipeline
type PineExecutor struct {
	ProjectRoot  string
	DataFilePath string
	Symbol       string
}

// NewPineExecutor creates executor with project root auto-detection
func NewPineExecutor(t *testing.T) *PineExecutor {
	t.Helper()

	projectRoot := findProjectRoot(t)
	dataPath := FetchTestData(t, "SPY", "1D", 500)

	return &PineExecutor{
		ProjectRoot:  projectRoot,
		DataFilePath: dataPath,
		Symbol:       "SPY",
	}
}

// ExecuteScript runs inline PineScript with default SPY data
func (e *PineExecutor) ExecuteScript(t *testing.T, name, script string) *PineScriptOutput {
	t.Helper()
	return e.executePipeline(t, name, script, e.DataFilePath, e.Symbol)
}

// ExecuteScriptWithCustomData runs inline PineScript with custom bar data
func (e *PineExecutor) ExecuteScriptWithCustomData(t *testing.T, name, script string, customBars []map[string]interface{}) *PineScriptOutput {
	t.Helper()

	tmpDir := t.TempDir()
	dataPath := e.prepareDataFile(t, tmpDir, customBars)
	return e.executePipeline(t, name, script, dataPath, "TEST")
}

// ExecuteScriptWithCustomDataRaw runs inline PineScript with custom bar data and returns raw JSON
func (e *PineExecutor) ExecuteScriptWithCustomDataRaw(t *testing.T, name, script string, customBars []map[string]interface{}) []byte {
	t.Helper()

	tmpDir := t.TempDir()
	dataPath := e.prepareDataFile(t, tmpDir, customBars)
	return e.executePipelineRaw(t, name, script, dataPath, "TEST")
}

// GenerateCode runs Parse→Generate step and returns generated Go code
func (e *PineExecutor) GenerateCode(t *testing.T, name, script string) (string, string) {
	t.Helper()

	tmpDir := t.TempDir()
	pineFile := filepath.Join(tmpDir, name+".pine")
	if err := os.WriteFile(pineFile, []byte(script), 0644); err != nil {
		t.Fatalf("Write pine file: %v", err)
	}

	builderPath := filepath.Join(e.ProjectRoot, "cmd", "pine-gen", "main.go")
	templatePath := filepath.Join(e.ProjectRoot, "template", "main.go.tmpl")
	binaryPath := filepath.Join(tmpDir, "test_binary")

	buildCmd := exec.Command("go", "run", builderPath, "-input", pineFile, "-output", binaryPath, "-template", templatePath)
	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Pine compilation failed: %v\n%s", err, buildOut)
	}

	var generatedFile string
	for _, line := range strings.Split(string(buildOut), "\n") {
		if strings.HasPrefix(line, "Generated:") {
			generatedFile = strings.TrimSpace(strings.TrimPrefix(line, "Generated:"))
			break
		}
	}
	if generatedFile == "" {
		t.Fatalf("Could not find generated file in output:\n%s", buildOut)
	}

	generatedCode, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("Read generated file: %v", err)
	}

	return string(generatedCode), generatedFile
}

// CompileCode compiles generated Go code to verify it builds successfully
func (e *PineExecutor) CompileCode(t *testing.T, generatedCode string) error {
	t.Helper()

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "generated.go")
	binaryPath := filepath.Join(tmpDir, "test_binary")

	if err := os.WriteFile(goFile, []byte(generatedCode), 0644); err != nil {
		t.Fatalf("Write generated code: %v", err)
	}

	compileCmd := exec.Command("go", "build", "-o", binaryPath, goFile)
	if _, err := compileCmd.CombinedOutput(); err != nil {
		return err
	}

	return nil
}

// executePipeline handles Parse→Generate→Compile→Execute→Parse flow
func (e *PineExecutor) executePipeline(t *testing.T, name, script, dataFilePath, symbol string) *PineScriptOutput {
	t.Helper()

	tmpDir := t.TempDir()
	pineFile := filepath.Join(tmpDir, name+".pine")
	if err := os.WriteFile(pineFile, []byte(script), 0644); err != nil {
		t.Fatalf("Write pine file: %v", err)
	}

	builderPath := filepath.Join(e.ProjectRoot, "cmd", "pine-gen", "main.go")
	templatePath := filepath.Join(e.ProjectRoot, "template", "main.go.tmpl")
	binaryPath := filepath.Join(tmpDir, "test_binary")

	buildCmd := exec.Command("go", "run", builderPath, "-input", pineFile, "-output", binaryPath, "-template", templatePath)
	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Pine compilation failed: %v\n%s", err, buildOut)
	}

	var generatedFile string
	for _, line := range strings.Split(string(buildOut), "\n") {
		if strings.HasPrefix(line, "Generated:") {
			generatedFile = strings.TrimSpace(strings.TrimPrefix(line, "Generated:"))
			break
		}
	}
	if generatedFile == "" {
		t.Fatalf("Could not find generated file in output:\n%s", buildOut)
	}

	compileCmd := exec.Command("go", "build", "-o", binaryPath, generatedFile)
	if compileOut, err := compileCmd.CombinedOutput(); err != nil {
		t.Fatalf("Go build failed: %v\n%s", err, compileOut)
	}

	resultPath := filepath.Join(tmpDir, "result.json")
	execCmd := exec.Command(binaryPath, "-symbol", symbol, "-data", dataFilePath, "-output", resultPath)
	if execOut, err := execCmd.CombinedOutput(); err != nil {
		t.Fatalf("Execution failed: %v\n%s", err, execOut)
	}

	jsonBytes, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatalf("Read result: %v", err)
	}

	var rawOutput struct {
		Indicators map[string]struct {
			Data []struct {
				Time  int64   `json:"time"`
				Value float64 `json:"value"`
			} `json:"data"`
		} `json:"indicators"`
		Strategy struct {
			ClosedTrades []struct {
				EntryBar  int   `json:"entry_bar"`
				ExitBar   int   `json:"exit_bar"`
				EntryTime int64 `json:"entry_time"`
				ExitTime  int64 `json:"exit_time"`
			} `json:"closed_trades"`
		} `json:"strategy"`
	}

	if err := json.Unmarshal(jsonBytes, &rawOutput); err != nil {
		t.Fatalf("Parse JSON: %v", err)
	}

	output := &PineScriptOutput{Plots: make([]StrategyPlot, 0)}
	for title, indicator := range rawOutput.Indicators {
		plot := StrategyPlot{Title: title}
		for _, d := range indicator.Data {
			plot.Data = append(plot.Data, PlotPoint{Time: d.Time, Value: d.Value})
		}
		output.Plots = append(output.Plots, plot)
	}

	for _, tr := range rawOutput.Strategy.ClosedTrades {
		output.Strategy.ClosedTrades = append(output.Strategy.ClosedTrades, StrategyTrade{
			EntryBar:  tr.EntryBar,
			ExitBar:   tr.ExitBar,
			EntryTime: tr.EntryTime,
			ExitTime:  tr.ExitTime,
		})
	}

	return output
}

// executePipelineRaw runs pipeline and returns raw JSON bytes
func (e *PineExecutor) executePipelineRaw(t *testing.T, name, script, dataFilePath, symbol string) []byte {
	t.Helper()

	tmpDir := t.TempDir()
	pineFile := filepath.Join(tmpDir, name+".pine")
	if err := os.WriteFile(pineFile, []byte(script), 0644); err != nil {
		t.Fatalf("Write pine file: %v", err)
	}

	builderPath := filepath.Join(e.ProjectRoot, "cmd", "pine-gen", "main.go")
	templatePath := filepath.Join(e.ProjectRoot, "template", "main.go.tmpl")
	binaryPath := filepath.Join(tmpDir, "test_binary")

	buildCmd := exec.Command("go", "run", builderPath, "-input", pineFile, "-output", binaryPath, "-template", templatePath)
	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Pine compilation failed: %v\n%s", err, buildOut)
	}

	var generatedFile string
	for _, line := range strings.Split(string(buildOut), "\n") {
		if strings.HasPrefix(line, "Generated:") {
			generatedFile = strings.TrimSpace(strings.TrimPrefix(line, "Generated:"))
			break
		}
	}
	if generatedFile == "" {
		t.Fatalf("Could not find generated file in output:\n%s", buildOut)
	}

	compileCmd := exec.Command("go", "build", "-o", binaryPath, generatedFile)
	if compileOut, err := compileCmd.CombinedOutput(); err != nil {
		t.Fatalf("Go build failed: %v\n%s", err, compileOut)
	}

	resultPath := filepath.Join(tmpDir, "result.json")
	execCmd := exec.Command(binaryPath, "-symbol", symbol, "-data", dataFilePath, "-output", resultPath)
	if execOut, err := execCmd.CombinedOutput(); err != nil {
		t.Fatalf("Execution failed: %v\n%s", err, execOut)
	}

	jsonBytes, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatalf("Read result: %v", err)
	}

	return jsonBytes
}

// prepareDataFile writes custom bar data to JSON file
func (e *PineExecutor) prepareDataFile(t *testing.T, tmpDir string, customBars []map[string]interface{}) string {
	t.Helper()

	dataPath := filepath.Join(tmpDir, "custom_data.json")
	dataJSON, err := json.Marshal(customBars)
	if err != nil {
		t.Fatalf("Marshal custom data: %v", err)
	}

	if err := os.WriteFile(dataPath, dataJSON, 0644); err != nil {
		t.Fatalf("Write custom data file: %v", err)
	}

	return dataPath
}

// ExtractPlotValues extracts numeric values from plot by title
func (e *PineExecutor) ExtractPlotValues(t *testing.T, output *PineScriptOutput, plotTitle string) []float64 {
	t.Helper()

	for _, plot := range output.Plots {
		if strings.Contains(plot.Title, plotTitle) {
			values := make([]float64, len(plot.Data))
			for i, point := range plot.Data {
				values[i] = point.Value
			}
			return values
		}
	}

	t.Fatalf("Plot %q not found in output", plotTitle)
	return nil
}

// PineScriptOutput represents strategy execution output
type PineScriptOutput struct {
	Plots    []StrategyPlot
	Strategy StrategyData
}

type StrategyPlot struct {
	Title string
	Data  []PlotPoint
}

type PlotPoint struct {
	Time  int64
	Value float64
}

type StrategyData struct {
	ClosedTrades []StrategyTrade
}

type StrategyTrade struct {
	EntryBar  int
	ExitBar   int
	EntryTime int64
	ExitTime  int64
}
