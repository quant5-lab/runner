package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
	"github.com/quant5-lab/runner/preprocessor"
)

/* TestStrategyRuntimeSamplingOrder validates execution order for strategy runtime state sampling */
func TestStrategyRuntimeSamplingOrder(t *testing.T) {
	script := `//@version=5
strategy("Test", overlay=true)

posAvg = strategy.position_avg_price

if close > 100
    strategy.entry("Long", strategy.long, 1.0)

plot(posAvg)
`

	script = preprocessor.NormalizeIfBlocks(script)

	pineParser, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Parser creation failed: %v", err)
	}

	parsedAST, err := pineParser.ParseString("test.pine", script)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	astConverter := parser.NewConverter()
	program, err := astConverter.ToESTree(parsedAST)
	if err != nil {
		t.Fatalf("AST conversion failed: %v", err)
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Codegen failed: %v", err)
	}

	lines := strings.Split(code.FunctionBody, "\n")

	var (
		onBarUpdateIdx      = -1
		sampleCurrentBarIdx = -1
		posAvgSetIdx        = -1
		advanceCursorsIdx   = -1
	)

	for i, line := range lines {
		if strings.Contains(line, "strat.OnBarUpdate") {
			onBarUpdateIdx = i
		}
		if strings.Contains(line, "sm.SampleCurrentBar") {
			sampleCurrentBarIdx = i
		}
		if strings.Contains(line, "posAvgSeries.Set") {
			posAvgSetIdx = i
		}
		if strings.Contains(line, "sm.AdvanceCursors") {
			advanceCursorsIdx = i
		}
	}

	if onBarUpdateIdx == -1 {
		t.Fatal("strat.OnBarUpdate not found")
	}
	if sampleCurrentBarIdx == -1 {
		t.Fatal("sm.SampleCurrentBar not found")
	}
	if posAvgSetIdx == -1 {
		t.Fatal("posAvgSeries.Set not found")
	}
	if advanceCursorsIdx == -1 {
		t.Fatal("sm.AdvanceCursors not found")
	}

	if sampleCurrentBarIdx <= onBarUpdateIdx {
		t.Errorf("sm.SampleCurrentBar (line %d) must come AFTER strat.OnBarUpdate (line %d)",
			sampleCurrentBarIdx, onBarUpdateIdx)
	}

	if posAvgSetIdx <= sampleCurrentBarIdx {
		t.Errorf("posAvgSeries.Set (line %d) must come AFTER sm.SampleCurrentBar (line %d)",
			posAvgSetIdx, sampleCurrentBarIdx)
	}

	if advanceCursorsIdx <= posAvgSetIdx {
		t.Errorf("sm.AdvanceCursors (line %d) must come AFTER posAvgSeries.Set (line %d)",
			advanceCursorsIdx, posAvgSetIdx)
	}
}

/* TestStrategyRuntimeWithoutAccess ensures StateManager only created when strategy runtime values accessed */
func TestStrategyRuntimeWithoutAccess(t *testing.T) {
	script := `//@version=5
strategy("No Runtime Access", overlay=true)

sma20 = ta.sma(close, 20)

if close > sma20
    strategy.entry("Long", strategy.long, 1.0)

plot(sma20)
`

	script = preprocessor.NormalizeIfBlocks(script)

	pineParser, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Parser creation failed: %v", err)
	}

	parsedAST, err := pineParser.ParseString("test.pine", script)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	astConverter := parser.NewConverter()
	program, err := astConverter.ToESTree(parsedAST)
	if err != nil {
		t.Fatalf("AST conversion failed: %v", err)
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Codegen failed: %v", err)
	}

	if strings.Contains(code.FunctionBody, "sm.SampleCurrentBar") {
		t.Error("Unexpected sm.SampleCurrentBar when strategy runtime values not accessed")
	}

	if strings.Contains(code.FunctionBody, "sm := strategy.NewStateManager") {
		t.Error("Unexpected StateManager when strategy runtime values not accessed")
	}
}

/* TestStrategyRuntimeMultipleAccess validates single sampling for multiple runtime values */
func TestStrategyRuntimeMultipleAccess(t *testing.T) {
	script := `//@version=5
strategy("Multiple Access", overlay=true)

posAvg = strategy.position_avg_price
posSize = strategy.position_size
eq = strategy.equity

if close > 100
    strategy.entry("Long", strategy.long, 1.0)

plot(posAvg)
plot(posSize)
plot(eq)
`

	script = preprocessor.NormalizeIfBlocks(script)

	pineParser, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Parser creation failed: %v", err)
	}

	parsedAST, err := pineParser.ParseString("test.pine", script)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	astConverter := parser.NewConverter()
	program, err := astConverter.ToESTree(parsedAST)
	if err != nil {
		t.Fatalf("AST conversion failed: %v", err)
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Codegen failed: %v", err)
	}

	sampleCalls := strings.Count(code.FunctionBody, "sm.SampleCurrentBar")
	if sampleCalls != 1 {
		t.Errorf("Expected exactly 1 sm.SampleCurrentBar call, found %d", sampleCalls)
	}

	requiredSeries := []string{
		"strategy_position_avg_priceSeries",
		"strategy_position_sizeSeries",
		"strategy_equitySeries",
	}

	for _, seriesName := range requiredSeries {
		if !strings.Contains(code.FunctionBody, seriesName) {
			t.Errorf("Missing required Series: %s", seriesName)
		}
	}
}
