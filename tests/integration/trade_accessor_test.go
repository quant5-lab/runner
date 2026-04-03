//go:build integration

package integration

import (
	"os"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/codegen"
	"github.com/quant5-lab/runner/parser"
)

func TestTradeAccessorIntegration(t *testing.T) {
	t.Parallel()
	content, err := os.ReadFile("../../e2e/fixtures/strategies/test-trade-accessor.pine")
	if err != nil {
		t.Fatalf("test-trade-accessor.pine not found: %v", err)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("parser creation failed: %v", err)
	}

	ast, err := p.ParseString("test-trade-accessor.pine", string(content))
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

	closedTradeAccessors := []string{
		"tradeAccessor.ClosedTradeProfit(int(0))",
		"tradeAccessor.ClosedTradeSize(int(0))",
		"tradeAccessor.ClosedTradeEntryPrice(int(0))",
		"tradeAccessor.ClosedTradeExitPrice(int(0))",
		"tradeAccessor.ClosedTradeEntryID(int(0))",
		"tradeAccessor.ClosedTradeEntryComment(int(0))",
		"tradeAccessor.ClosedTradeExitComment(int(0))",
		"tradeAccessor.ClosedTradeEntryBarIndex(int(0))",
		"tradeAccessor.ClosedTradeExitBarIndex(int(0))",
		"tradeAccessor.ClosedTradeProfit(int(1))",
		"tradeAccessor.ClosedTradeSize(int(1))",
	}

	openTradeAccessors := []string{
		"tradeAccessor.OpenTradeProfit(int(0))",
		"tradeAccessor.OpenTradeSize(int(0))",
		"tradeAccessor.OpenTradeEntryPrice(int(0))",
		"tradeAccessor.OpenTradeEntryID(int(0))",
		"tradeAccessor.OpenTradeEntryComment(int(0))",
	}

	for _, accessor := range closedTradeAccessors {
		if !strings.Contains(strategyCode.FunctionBody, accessor) {
			t.Errorf("generated code missing: %s", accessor)
		}
	}

	for _, accessor := range openTradeAccessors {
		if !strings.Contains(strategyCode.FunctionBody, accessor) {
			t.Errorf("generated code missing: %s", accessor)
		}
	}

	if !strings.Contains(strategyCode.FunctionBody, "lastClosedProfitSeries.Set") {
		t.Error("generated code missing lastClosedProfit Series.Set")
	}
	if !strings.Contains(strategyCode.FunctionBody, "openTradeProfitSeries.Set") {
		t.Error("generated code missing openTradeProfit Series.Set")
	}
}
