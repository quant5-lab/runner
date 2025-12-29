package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

func TestSecurityInlineInConditionals(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		mustContain []string
	}{
		{
			name: "security() in ternary condition",
			script: `//@version=4
strategy("Test", overlay=true)
sma_1d = security(syminfo.tickerid, "1D", sma(close, 20))
signal = sma_1d > close ? 1 : 0
plot(signal)`,
			mustContain: []string{
				"sma_1dSeries.Set",
				"signalSeries.Set",
				"securityContexts[secKey]",
			},
		},
		{
			name: "security() directly in ternary test",
			script: `//@version=4
strategy("Test", overlay=true)
signal = security(syminfo.tickerid, "1D", close) > close ? 1 : 0
plot(signal)`,
			mustContain: []string{
				"signalSeries.Set",
				"(func() float64 {",
				"securityContexts[secKey]",
				"secCtx.Data[secBarIdx].Close",
			},
		},
		{
			name: "security() with comparison in ternary",
			script: `//@version=4
strategy("Test", overlay=true)
bb_upper = sma(close, 20) + 2 * stdev(close, 20)
signal = security(syminfo.tickerid, "1D", low > bb_upper) ? 1 : 0
plot(signal)`,
			mustContain: []string{
				"signalSeries.Set",
				"(func() float64 {",
				"secBarEvaluator.EvaluateAtBar",
				"&ast.BinaryExpression",
			},
		},
		{
			name: "nested security() calls inline",
			script: `//@version=4
strategy("Test", overlay=true)
close_1d = security(syminfo.tickerid, "1D", close)
open_1d = security(syminfo.tickerid, "1D", open)
candle_type = close_1d > open_1d ? security(syminfo.tickerid, "1D", high) : security(syminfo.tickerid, "1D", low)
plot(candle_type)`,
			mustContain: []string{
				"close_1dSeries.Set",
				"open_1dSeries.Set",
				"candle_typeSeries.Set",
				"securityContexts[secKey]",
			},
		},
		{
			name: "security() with TA function inline",
			script: `//@version=4
strategy("Test", overlay=true)
is_bullish = security(syminfo.tickerid, "1D", sma(close, 20) > sma(close, 50)) ? 1 : 0
plot(is_bullish)`,
			mustContain: []string{
				"is_bullishSeries.Set",
				"secBarEvaluator.EvaluateAtBar",
				"&ast.BinaryExpression",
			},
		},
		{
			name: "security() with lookahead in ternary",
			script: `//@version=4
strategy("Test", overlay=true)
open_1d = security(syminfo.tickerid, "D", open, lookahead=barmerge.lookahead_on)
signal = open_1d > close ? 1 : 0
plot(signal)`,
			mustContain: []string{
				"open_1dSeries.Set",
				"secLookahead := true",
				"securityContexts[secKey]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			parseResult, err := p.ParseBytes("test.pine", []byte(tt.script))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(parseResult)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			result, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("GenerateStrategyCodeFromAST failed: %v", err)
			}

			code := result.FunctionBody

			for _, substr := range tt.mustContain {
				if !strings.Contains(code, substr) {
					t.Errorf("missing substring %q in generated code", substr)
				}
			}

			if strings.Contains(code, "undefined:") {
				t.Errorf("generated code contains undefined references")
			}
		})
	}
}

func TestSecurityInlineTimeframeNormalization(t *testing.T) {
	tests := []struct {
		name      string
		timeframe string
		expectKey string
	}{
		{
			name:      "daily short form D",
			timeframe: "D",
			expectKey: "%s:1D",
		},
		{
			name:      "weekly short form W",
			timeframe: "W",
			expectKey: "%s:1W",
		},
		{
			name:      "monthly short form M",
			timeframe: "M",
			expectKey: "%s:1M",
		},
		{
			name:      "explicit 1D",
			timeframe: "1D",
			expectKey: "%s:1D",
		},
		{
			name:      "intraday 5m",
			timeframe: "5m",
			expectKey: "%s:5m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := `//@version=4
strategy("Test", overlay=true)
val = security(syminfo.tickerid, "` + tt.timeframe + `", close)
plot(val)`

			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			parseResult, err := p.ParseBytes("test.pine", []byte(script))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(parseResult)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			result, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("GenerateStrategyCodeFromAST failed: %v", err)
			}

			code := result.FunctionBody

			if !strings.Contains(code, tt.expectKey) {
				t.Errorf("expected cache key pattern %q not found", tt.expectKey)
			}
		})
	}
}

func TestSecurityInlineErrorRecovery(t *testing.T) {
	tests := []struct {
		name      string
		script    string
		expectNaN []string
	}{
		{
			name: "cache not found recovery",
			script: `//@version=4
strategy("Test", overlay=true)
signal = security(syminfo.tickerid, "1D", close) > 0 ? 1 : 0
plot(signal)`,
			expectNaN: []string{
				"if !secFound { return math.NaN() }",
				"if !mapperFound { return math.NaN() }",
				"if secBarIdx < 0 { return math.NaN() }",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			parseResult, err := p.ParseBytes("test.pine", []byte(tt.script))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(parseResult)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			result, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("GenerateStrategyCodeFromAST failed: %v", err)
			}

			code := result.FunctionBody

			for _, nanCheck := range tt.expectNaN {
				if !strings.Contains(code, nanCheck) {
					t.Errorf("missing NaN recovery path: %q", nanCheck)
				}
			}
		})
	}
}
