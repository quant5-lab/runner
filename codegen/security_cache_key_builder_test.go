package codegen

import (
	"testing"
)

func TestSecurityCacheKeyBuilder_AllRuntimeCombinations(t *testing.T) {
	builder := NewSecurityCacheKeyBuilder()

	tests := []struct {
		name            string
		symbolResult    *ExtractionResult
		timeframeResult *ExtractionResult
		wantKeyPattern  string
		wantFormatArgs  string
	}{
		{
			name:            "both runtime",
			symbolResult:    &ExtractionResult{Code: "ctx.Symbol", IsRuntime: true},
			timeframeResult: &ExtractionResult{Code: "ctx.Timeframe", IsRuntime: true},
			wantKeyPattern:  "%s:%s",
			wantFormatArgs:  "ctx.Symbol, ctx.Timeframe",
		},
		{
			name:            "runtime symbol, constant timeframe",
			symbolResult:    &ExtractionResult{Code: "ctx.Symbol", IsRuntime: true},
			timeframeResult: &ExtractionResult{Code: `"1D"`, IsRuntime: false},
			wantKeyPattern:  "%s:1D",
			wantFormatArgs:  "ctx.Symbol",
		},
		{
			name:            "constant symbol, runtime timeframe",
			symbolResult:    &ExtractionResult{Code: `"BTCUSDT"`, IsRuntime: false},
			timeframeResult: &ExtractionResult{Code: "ctx.Timeframe", IsRuntime: true},
			wantKeyPattern:  "BTCUSDT:%s",
			wantFormatArgs:  "ctx.Timeframe",
		},
		{
			name:            "both constant",
			symbolResult:    &ExtractionResult{Code: `"BTCUSDT"`, IsRuntime: false},
			timeframeResult: &ExtractionResult{Code: `"1D"`, IsRuntime: false},
			wantKeyPattern:  "BTCUSDT:1D",
			wantFormatArgs:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.Build(tt.symbolResult, tt.timeframeResult)

			if result.KeyPattern != tt.wantKeyPattern {
				t.Errorf("KeyPattern = %q, want %q", result.KeyPattern, tt.wantKeyPattern)
			}

			if result.FormatArgs != tt.wantFormatArgs {
				t.Errorf("FormatArgs = %q, want %q", result.FormatArgs, tt.wantFormatArgs)
			}
		})
	}
}

func TestSecurityCacheKeyBuilder_SpecialCharactersInSymbol(t *testing.T) {
	builder := NewSecurityCacheKeyBuilder()

	tests := []struct {
		symbolCode     string
		timeframeCode  string
		wantKeyPattern string
	}{
		{
			symbolCode:     `"BTC-USDT"`,
			timeframeCode:  `"1D"`,
			wantKeyPattern: "BTC-USDT:1D",
		},
		{
			symbolCode:     `"BINANCE:BTCUSDT"`,
			timeframeCode:  `"1H"`,
			wantKeyPattern: "BINANCE:BTCUSDT:1H",
		},
		{
			symbolCode:     `"BTC/USDT"`,
			timeframeCode:  `"5m"`,
			wantKeyPattern: "BTC/USDT:5m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.symbolCode, func(t *testing.T) {
			symbolResult := &ExtractionResult{Code: tt.symbolCode, IsRuntime: false}
			timeframeResult := &ExtractionResult{Code: tt.timeframeCode, IsRuntime: false}

			result := builder.Build(symbolResult, timeframeResult)

			if result.KeyPattern != tt.wantKeyPattern {
				t.Errorf("KeyPattern = %q, want %q", result.KeyPattern, tt.wantKeyPattern)
			}

			if result.FormatArgs != "" {
				t.Errorf("FormatArgs = %q, want empty string for constant key", result.FormatArgs)
			}
		})
	}
}

func TestSecurityCacheKeyBuilder_UserVariableScenarios(t *testing.T) {
	builder := NewSecurityCacheKeyBuilder()

	tests := []struct {
		name            string
		symbolResult    *ExtractionResult
		timeframeResult *ExtractionResult
		wantKeyPattern  string
		wantFormatArgs  string
	}{
		{
			name:            "user symbol variable, constant timeframe",
			symbolResult:    &ExtractionResult{Code: "mySymbol", IsRuntime: true},
			timeframeResult: &ExtractionResult{Code: `"1D"`, IsRuntime: false},
			wantKeyPattern:  "%s:1D",
			wantFormatArgs:  "mySymbol",
		},
		{
			name:            "constant symbol, user timeframe variable",
			symbolResult:    &ExtractionResult{Code: `"BTCUSDT"`, IsRuntime: false},
			timeframeResult: &ExtractionResult{Code: "myTimeframe", IsRuntime: true},
			wantKeyPattern:  "BTCUSDT:%s",
			wantFormatArgs:  "myTimeframe",
		},
		{
			name:            "both user variables",
			symbolResult:    &ExtractionResult{Code: "mySymbol", IsRuntime: true},
			timeframeResult: &ExtractionResult{Code: "myTimeframe", IsRuntime: true},
			wantKeyPattern:  "%s:%s",
			wantFormatArgs:  "mySymbol, myTimeframe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.Build(tt.symbolResult, tt.timeframeResult)

			if result.KeyPattern != tt.wantKeyPattern {
				t.Errorf("KeyPattern = %q, want %q", result.KeyPattern, tt.wantKeyPattern)
			}

			if result.FormatArgs != tt.wantFormatArgs {
				t.Errorf("FormatArgs = %q, want %q", result.FormatArgs, tt.wantFormatArgs)
			}
		})
	}
}

func TestSecurityCacheKeyBuilder_EmptyFormatArgsForConstantKeys(t *testing.T) {
	builder := NewSecurityCacheKeyBuilder()

	symbolResult := &ExtractionResult{Code: `"BTCUSDT"`, IsRuntime: false}
	timeframeResult := &ExtractionResult{Code: `"1D"`, IsRuntime: false}

	result := builder.Build(symbolResult, timeframeResult)

	if result.FormatArgs != "" {
		t.Errorf("FormatArgs should be empty for fully constant key, got %q", result.FormatArgs)
	}

	if result.KeyPattern != "BTCUSDT:1D" {
		t.Errorf("KeyPattern = %q, want %q", result.KeyPattern, "BTCUSDT:1D")
	}
}
