//go:build integration

package integration

import (
	"os"
	"testing"

	"github.com/quant5-lab/runner/codegen"
	"github.com/quant5-lab/runner/parser"
)

func TestCumBasicIntegration(t *testing.T) {
	t.Parallel()
	content, err := os.ReadFile("../../e2e/fixtures/strategies/test-cum-basic.pine")
	if err != nil {
		t.Fatalf("test-cum-basic.pine not found: %v", err)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("parser creation failed: %v", err)
	}

	ast, err := p.ParseString("test-cum-basic.pine", string(content))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	converter := parser.NewConverter()
	estree, err := converter.ToESTree(ast)
	if err != nil {
		t.Fatalf("ESTree conversion failed: %v", err)
	}

	strategyCode, err := codegen.GenerateStrategyCodeFromAST(estree)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}

	if strategyCode.FunctionBody == "" {
		t.Fatal("generated code empty")
	}

	if !containsSubstr(strategyCode.FunctionBody, "ta.cum(") {
		t.Error("generated code missing ta.cum comment")
	}
	if !containsSubstr(strategyCode.FunctionBody, "cumGainSeries.Set") {
		t.Error("generated code missing cumGain Series.Set")
	}
	if !containsSubstr(strategyCode.FunctionBody, "cumLossSeries.Set") {
		t.Error("generated code missing cumLoss Series.Set")
	}
	if !containsSubstr(strategyCode.FunctionBody, "netCumSeries.Set") {
		t.Error("generated code missing netCum Series.Set")
	}
}

func TestCumEdgeCasesIntegration(t *testing.T) {
	t.Parallel()
	content, err := os.ReadFile("../../e2e/fixtures/strategies/test-cum-edge-cases.pine")
	if err != nil {
		t.Fatalf("test-cum-edge-cases.pine not found: %v", err)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("parser creation failed: %v", err)
	}

	ast, err := p.ParseString("test-cum-edge-cases.pine", string(content))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	converter := parser.NewConverter()
	estree, err := converter.ToESTree(ast)
	if err != nil {
		t.Fatalf("ESTree conversion failed: %v", err)
	}

	strategyCode, err := codegen.GenerateStrategyCodeFromAST(estree)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}

	if !containsSubstr(strategyCode.FunctionBody, "math.IsNaN") {
		t.Error("generated code missing NaN handling")
	}
	if !containsSubstr(strategyCode.FunctionBody, "cumNaNSeries.Set") {
		t.Error("generated code missing cumNaN Series.Set")
	}
	if !containsSubstr(strategyCode.FunctionBody, "cumNegSeries.Set") {
		t.Error("generated code missing cumNeg Series.Set")
	}
	if !containsSubstr(strategyCode.FunctionBody, "cumZeroSeries.Set") {
		t.Error("generated code missing cumZero Series.Set")
	}
	if !containsSubstr(strategyCode.FunctionBody, "cumLeadingSeries.Set") {
		t.Error("generated code missing cumLeading Series.Set")
	}
}

func TestCumInlineIntegration(t *testing.T) {
	t.Parallel()
	content, err := os.ReadFile("../../e2e/fixtures/strategies/test-cum-arrow.pine")
	if err != nil {
		t.Fatalf("test-cum-arrow.pine not found: %v", err)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("parser creation failed: %v", err)
	}

	ast, err := p.ParseString("test-cum-arrow.pine", string(content))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	converter := parser.NewConverter()
	estree, err := converter.ToESTree(ast)
	if err != nil {
		t.Fatalf("ESTree conversion failed: %v", err)
	}

	strategyCode, err := codegen.GenerateStrategyCodeFromAST(estree)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}

	if !containsSubstr(strategyCode.FunctionBody, "ta.cum(") {
		t.Error("generated code missing ta.cum comment")
	}
	if !containsSubstr(strategyCode.FunctionBody, "cumVolumeSeries.Set") {
		t.Error("generated code missing cumVolume Series.Set")
	}
	if !containsSubstr(strategyCode.FunctionBody, "cumCloseSeries.Set") {
		t.Error("generated code missing cumClose Series.Set")
	}
}
