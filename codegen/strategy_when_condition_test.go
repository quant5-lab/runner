package codegen

import (
	"fmt"
	"strings"
	"testing"
)

// These tests drive the real parser and generator, so a condition's emitted shape is asserted
// through the same path a script takes rather than through a hand-built AST.

/*
TestStrategyWhenComparisonConditionNeedsNoCoercion guards a comparison condition, whose lowering
is already a Go bool.
*/
func TestStrategyWhenComparisonConditionNeedsNoCoercion(t *testing.T) {
	code, err := compilePineScript(`
//@version=5
strategy("Comparison condition", overlay=true)
strategy.entry("L", strategy.long, when=close > 100)
`)
	if err != nil {
		t.Fatalf("code generation failed: %v", err)
	}

	if !strings.Contains(code, "if (bar.Close > 100) {") {
		t.Errorf("expected a bare Go bool condition, got:\n%s", code)
	}

	if strings.Contains(code, "value.IsTrue((bar.Close") {
		t.Errorf("a comparison operand must not be coerced with value.IsTrue, got:\n%s", code)
	}
}

/*
TestStrategyWhenLogicalConditionCoercesPerOperand guards that the enclosing float64 IIFE keeps
its coercion while each already-boolean operand is emitted bare.
*/
func TestStrategyWhenLogicalConditionCoercesPerOperand(t *testing.T) {
	code, err := compilePineScript(`
//@version=5
strategy("Logical condition", overlay=true)
strategy.entry("L", strategy.long, when=close > 100 and close < 200)
`)
	if err != nil {
		t.Fatalf("code generation failed: %v", err)
	}

	if !strings.Contains(code, "((bar.Close > 100) && (bar.Close < 200))") {
		t.Errorf("expected both comparison operands as bare Go bools, got:\n%s", code)
	}

	if strings.Contains(code, "value.IsTrue((bar.Close") {
		t.Errorf("a comparison operand must not be coerced with value.IsTrue, got:\n%s", code)
	}
}

/*
TestStrategyWhenNumericConditionStaysCoerced guards every operand kind whose lowering is float64:
the coercion must remain, and the coercion must wrap the whole operand.
*/
func TestStrategyWhenNumericConditionStaysCoerced(t *testing.T) {
	tests := []struct {
		name          string
		condition     string
		wantFragment  string
		justification string
	}{
		{
			name:          "bare numeric series",
			condition:     "close",
			wantFragment:  "value.IsTrue(bar.Close)",
			justification: "numeric truthiness needs the coercion",
		},
		{
			name:          "unary-minus negation",
			condition:     "-close",
			wantFragment:  "value.IsTrue(-bar.Close)",
			justification: "the operand is arithmetic, not a boolean",
		},
		{
			name:          "numeric call",
			condition:     "ta.sma(close, 20)",
			wantFragment:  "value.IsTrue(ta_sma_",
			justification: "a numeric call lowers to a float64 series",
		},
		{
			name:          "bool-returning call",
			condition:     "ta.crossover(close, ta.sma(close, 20))",
			wantFragment:  "value.IsTrue(ta_crossover",
			justification: "IsAlreadyBoolean calls this boolean, but the emitted binding is float64",
		},
		{
			name:          "bool variable",
			condition:     "enabled",
			wantFragment:  "value.IsTrue(enabledSeries.GetCurrent())",
			justification: "bool variables are stored as 0.0/1.0 series",
		},
		{
			name:          "logical over numeric operands",
			condition:     "enabled and close",
			wantFragment:  "value.IsTrue(enabledSeries.GetCurrent()) && value.IsTrue(bar.Close)",
			justification: "a logical over numeric operands is a float64 IIFE, coerced per operand",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(fmt.Sprintf(`
//@version=5
strategy("Numeric condition", overlay=true)
enabled = close > 100
strategy.entry("L", strategy.long, when=%s)
`, tt.condition))
			if err != nil {
				t.Fatalf("code generation failed: %v", err)
			}

			if !strings.Contains(code, tt.wantFragment) {
				t.Errorf("expected %q: %s, got:\n%s", tt.wantFragment, tt.justification, code)
			}
		})
	}
}

/*
TestIfStatementComparisonNeedsNoCoercion guards that a comparison condition keeps emitting a bare
Go bool in an if statement, which is a separate lowering from the strategy argument path.
*/
func TestIfStatementComparisonNeedsNoCoercion(t *testing.T) {
	code, err := compilePineScript(`
//@version=5
strategy("If condition", overlay=true)
if close > 100
    strategy.entry("L", strategy.long)
`)
	if err != nil {
		t.Fatalf("code generation failed: %v", err)
	}

	if !strings.Contains(code, "if (bar.Close > 100) {") {
		t.Errorf("expected a bare Go bool condition, got:\n%s", code)
	}
}

/*
TestStrategyWhenGoBoolConditionNeedsNoCoercion guards every when= operand whose lowering is
already a Go bool: the value handler's na, the same negated, a namespace member the resolver types
GoBool, that member inside a logical combination, and a negation the extractor coerces its operand
for. Each must reach the block bare, because value.IsTrue accepts a float64.
*/
func TestStrategyWhenGoBoolConditionNeedsNoCoercion(t *testing.T) {
	tests := []struct {
		name         string
		condition    string
		wantFragment string
		mustNot      string
	}{
		{
			name:         "na call",
			condition:    "na(close)",
			wantFragment: "if math.IsNaN(bar.Close) {",
			mustNot:      "value.IsTrue(math.IsNaN(bar.Close))",
		},
		{
			name:         "negated na call",
			condition:    "not na(close)",
			wantFragment: "if !math.IsNaN(bar.Close) {",
			mustNot:      "value.IsTrue(!math.IsNaN(bar.Close))",
		},
		{
			name:         "namespace member",
			condition:    "barstate.islast",
			wantFragment: "if (ctx.BarIndex == len(ctx.Data)-1) {",
			mustNot:      "value.IsTrue((ctx.BarIndex",
		},
		{
			name:         "namespace member inside a logical combination",
			condition:    "barstate.islast and close > 100",
			wantFragment: "((ctx.BarIndex == len(ctx.Data)-1) && (bar.Close > 100))",
			mustNot:      "value.IsTrue((ctx.BarIndex",
		},
		{
			name:         "negated bool variable",
			condition:    "not enabled",
			wantFragment: "if !value.IsTrue(enabledSeries.GetCurrent()) {",
			mustNot:      "value.IsTrue(!",
		},
		{
			name:         "negated numeric series",
			condition:    "not close",
			wantFragment: "if !value.IsTrue(bar.Close) {",
			mustNot:      "value.IsTrue(!bar.Close",
		},
		{
			name:         "negated bool literal",
			condition:    "not true",
			wantFragment: "if !value.IsTrue(1.0) {",
			mustNot:      "value.IsTrue(!1.0)",
		},
		{
			name:         "numeric namespace member keeps its coercion",
			condition:    "dayofweek.monday",
			wantFragment: "if value.IsTrue(2.0) {",
			mustNot:      "if 2.0 {",
		},
		{
			name:         "conditional expression carries its test",
			condition:    "close > 100 ? true : false",
			wantFragment: "value.IsTrue(func() float64 { if (bar.Close > 100) { return 1.0 } else { return 0.0 } }())",
			mustNot:      "value.IsTrue(0.0)",
		},
		{
			name:         "conditional expression with series branches",
			condition:    "close > 100 ? close : open",
			wantFragment: "value.IsTrue(func() float64 { if (bar.Close > 100) { return bar.Close } else { return bar.Open } }())",
			mustNot:      "value.IsTrue(0.0)",
		},
		{
			name:         "session member",
			condition:    "session.isfirstbar",
			wantFragment: "if session_isfirstbarSeries.GetCurrent() == 1.0 {",
			mustNot:      "value.IsTrue(session_isfirstbarSeries",
		},
		{
			name:         "strategy runtime member",
			condition:    "strategy.position_size > 0",
			wantFragment: "strategy_position_sizeSeries.Get(0) > 0",
			mustNot:      "value.IsTrue((strategy_position_sizeSeries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(fmt.Sprintf(`
//@version=5
strategy("Go bool condition", overlay=true)
enabled = close > 100
strategy.entry("L", strategy.long, when=%s)
`, tt.condition))
			if err != nil {
				t.Fatalf("code generation failed: %v", err)
			}

			if !strings.Contains(code, tt.wantFragment) {
				t.Errorf("expected %q, got:\n%s", tt.wantFragment, code)
			}

			if strings.Contains(code, tt.mustNot) {
				t.Errorf("a Go bool operand must reach the block bare, got:\n%s", code)
			}
		})
	}
}

/*
TestVariableDeclarationLogicalMemberNeedsNoCoercion guards the same Go-bool decision in a variable
initialiser, which reaches the condition lowering through the logical operands rather than through
the argument extractor.
*/
func TestVariableDeclarationLogicalMemberNeedsNoCoercion(t *testing.T) {
	code, err := compilePineScript(`
//@version=5
strategy("Variable declaration condition", overlay=true)
signal = barstate.islast and close > 100
strategy.entry("L", strategy.long, when=signal)
`)
	if err != nil {
		t.Fatalf("code generation failed: %v", err)
	}

	if !strings.Contains(code, "((ctx.BarIndex == len(ctx.Data)-1) && (bar.Close > 100))") {
		t.Errorf("expected the member to reach the initialiser bare, got:\n%s", code)
	}

	if strings.Contains(code, "value.IsTrue((ctx.BarIndex") {
		t.Errorf("a Go bool member must not be coerced with value.IsTrue, got:\n%s", code)
	}
}
