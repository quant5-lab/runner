//go:build integration

package integration

import (
	"os"
	"testing"

	"github.com/quant5-lab/runner/codegen"
	"github.com/quant5-lab/runner/parser"
)

func TestVWMABasicIntegration(t *testing.T) {
	t.Parallel()
	content, err := os.ReadFile("../../e2e/fixtures/strategies/test-vwma-basic.pine")
	if err != nil {
		t.Fatalf("test-vwma-basic.pine not found: %v", err)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("parser creation failed: %v", err)
	}

	ast, err := p.ParseString("test-vwma-basic.pine", string(content))
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

	if !containsSubstr(strategyCode.FunctionBody, "ta.vwma(") {
		t.Error("generated code missing ta.vwma comment")
	}
	if !containsSubstr(strategyCode.FunctionBody, "vwma14Series.Set") {
		t.Error("generated code missing vwma14 Series.Set")
	}
	if !containsSubstr(strategyCode.FunctionBody, "vwma20Series.Set") {
		t.Error("generated code missing vwma20 Series.Set")
	}
	if !containsSubstr(strategyCode.FunctionBody, "vwma50Series.Set") {
		t.Error("generated code missing vwma50 Series.Set")
	}
	if !containsSubstr(strategyCode.FunctionBody, "weightedSum") {
		t.Error("generated code missing weightedSum accumulation")
	}
	if !containsSubstr(strategyCode.FunctionBody, "volumeSum") {
		t.Error("generated code missing volumeSum accumulation")
	}
	if !containsSubstr(strategyCode.FunctionBody, "bar.Volume") || !containsSubstr(strategyCode.FunctionBody, "Volume") {
		t.Error("generated code missing Volume access")
	}
}

func TestVWMASourceTypesIntegration(t *testing.T) {
	t.Parallel()
	content, err := os.ReadFile("../../e2e/fixtures/strategies/test-vwma-sources.pine")
	if err != nil {
		t.Fatalf("test-vwma-sources.pine not found: %v", err)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("parser creation failed: %v", err)
	}

	ast, err := p.ParseString("test-vwma-sources.pine", string(content))
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

	if !containsSubstr(strategyCode.FunctionBody, "vwmaCloseSeries.Set") {
		t.Error("generated code missing vwmaClose Series.Set")
	}
	if !containsSubstr(strategyCode.FunctionBody, "vwmaHighSeries.Set") {
		t.Error("generated code missing vwmaHigh Series.Set")
	}
	if !containsSubstr(strategyCode.FunctionBody, "vwmaLowSeries.Set") {
		t.Error("generated code missing vwmaLow Series.Set")
	}
	if !containsSubstr(strategyCode.FunctionBody, "vwmaHL2Series.Set") {
		t.Error("generated code missing vwmaHL2 Series.Set")
	}
	if !containsSubstr(strategyCode.FunctionBody, "vwmaHLC3Series.Set") {
		t.Error("generated code missing vwmaHLC3 Series.Set")
	}
	if !containsSubstr(strategyCode.FunctionBody, "vwmaOHLC4Series.Set") {
		t.Error("generated code missing vwmaOHLC4 Series.Set")
	}
	if !containsSubstr(strategyCode.FunctionBody, "vwmaExprSeries.Set") {
		t.Error("generated code missing vwmaExpr Series.Set")
	}
	if !containsSubstr(strategyCode.FunctionBody, "vwmaNestedSeries.Set") {
		t.Error("generated code missing vwmaNested Series.Set")
	}
}

func TestVWMAEdgeCasesIntegration(t *testing.T) {
	t.Parallel()
	content, err := os.ReadFile("../../e2e/fixtures/strategies/test-vwma-edge-cases.pine")
	if err != nil {
		t.Fatalf("test-vwma-edge-cases.pine not found: %v", err)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("parser creation failed: %v", err)
	}

	ast, err := p.ParseString("test-vwma-edge-cases.pine", string(content))
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

	if !containsSubstr(strategyCode.FunctionBody, "vwmaShortSeries.Set") {
		t.Error("generated code missing vwmaShort Series.Set")
	}
	if !containsSubstr(strategyCode.FunctionBody, "vwmaLongSeries.Set") {
		t.Error("generated code missing vwmaLong Series.Set")
	}
	if !containsSubstr(strategyCode.FunctionBody, "vwmaBareSeries.Set") {
		t.Error("generated code missing vwmaBare Series.Set")
	}
	if !containsSubstr(strategyCode.FunctionBody, "vwmaConditionalSeries.Set") {
		t.Error("generated code missing vwmaConditional Series.Set")
	}
	if !containsSubstr(strategyCode.FunctionBody, "math.IsNaN") {
		t.Error("generated code missing NaN handling")
	}
}

func TestVWMAVsWMAComparison(t *testing.T) {
	t.Parallel()

	vwmaContent := `//@version=5
strategy("VWMA Test", overlay=true)
result = ta.vwma(close, 20)
plot(result)
`

	wmaContent := `//@version=5
strategy("WMA Test", overlay=true)
result = ta.wma(close, 20)
plot(result)
`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("parser creation failed: %v", err)
	}

	vwmaAST, err := p.ParseString("vwma-test.pine", vwmaContent)
	if err != nil {
		t.Fatalf("VWMA parse failed: %v", err)
	}

	wmaAST, err := p.ParseString("wma-test.pine", wmaContent)
	if err != nil {
		t.Fatalf("WMA parse failed: %v", err)
	}

	converter := parser.NewConverter()

	vwmaEstree, err := converter.ToESTree(vwmaAST)
	if err != nil {
		t.Fatalf("VWMA ESTree conversion failed: %v", err)
	}

	wmaEstree, err := converter.ToESTree(wmaAST)
	if err != nil {
		t.Fatalf("WMA ESTree conversion failed: %v", err)
	}

	vwmaCode, err := codegen.GenerateStrategyCodeFromAST(vwmaEstree)
	if err != nil {
		t.Fatalf("VWMA codegen failed: %v", err)
	}

	wmaCode, err := codegen.GenerateStrategyCodeFromAST(wmaEstree)
	if err != nil {
		t.Fatalf("WMA codegen failed: %v", err)
	}

	if !containsSubstr(vwmaCode.FunctionBody, "Volume") {
		t.Error("VWMA must use Volume weighting")
	}
	if containsSubstr(wmaCode.FunctionBody, "volumeSum") {
		t.Error("WMA should not use volumeSum (only VWMA does)")
	}
	if !containsSubstr(vwmaCode.FunctionBody, "volumeSum") {
		t.Error("VWMA must accumulate volumeSum")
	}
	if !containsSubstr(vwmaCode.FunctionBody, "weightedSum / volumeSum") {
		t.Error("VWMA must divide weightedSum by volumeSum")
	}
}

func TestVWMANaNPropagation(t *testing.T) {
	t.Parallel()

	content := `//@version=5
strategy("VWMA NaN Test", overlay=true)
nanValue = close / 0.0
vwmaNaN = ta.vwma(nanValue, 10)
vwmaValid = ta.vwma(close, 10)
plot(vwmaNaN)
plot(vwmaValid)
`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("parser creation failed: %v", err)
	}

	ast, err := p.ParseString("vwma-nan-test.pine", content)
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
		t.Error("VWMA must check for NaN values")
	}
	if !containsSubstr(strategyCode.FunctionBody, "hasNaN") {
		t.Error("VWMA must track NaN presence")
	}
	if !containsSubstr(strategyCode.FunctionBody, "if hasNaN") {
		t.Error("VWMA must handle hasNaN flag")
	}
}

func TestVWMABareAliasCodeEquivalence(t *testing.T) {
	t.Parallel()

	namespacedContent := `//@version=5
strategy("Namespaced VWMA", overlay=true)
result = ta.vwma(close, 14)
plot(result)
`

	bareContent := `//@version=5
strategy("Bare VWMA", overlay=true)
result = vwma(close, 14)
plot(result)
`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("parser creation failed: %v", err)
	}

	namespacedAST, err := p.ParseString("namespaced.pine", namespacedContent)
	if err != nil {
		t.Fatalf("namespaced parse failed: %v", err)
	}

	bareAST, err := p.ParseString("bare.pine", bareContent)
	if err != nil {
		t.Fatalf("bare parse failed: %v", err)
	}

	converter := parser.NewConverter()

	namespacedEstree, err := converter.ToESTree(namespacedAST)
	if err != nil {
		t.Fatalf("namespaced ESTree conversion failed: %v", err)
	}

	bareEstree, err := converter.ToESTree(bareAST)
	if err != nil {
		t.Fatalf("bare ESTree conversion failed: %v", err)
	}

	namespacedCode, err := codegen.GenerateStrategyCodeFromAST(namespacedEstree)
	if err != nil {
		t.Fatalf("namespaced codegen failed: %v", err)
	}

	bareCode, err := codegen.GenerateStrategyCodeFromAST(bareEstree)
	if err != nil {
		t.Fatalf("bare codegen failed: %v", err)
	}

	if !containsSubstr(namespacedCode.FunctionBody, "weightedSum") {
		t.Error("namespaced VWMA missing weightedSum")
	}
	if !containsSubstr(bareCode.FunctionBody, "weightedSum") {
		t.Error("bare VWMA missing weightedSum")
	}
	if !containsSubstr(namespacedCode.FunctionBody, "volumeSum") {
		t.Error("namespaced VWMA missing volumeSum")
	}
	if !containsSubstr(bareCode.FunctionBody, "volumeSum") {
		t.Error("bare VWMA missing volumeSum")
	}
}
