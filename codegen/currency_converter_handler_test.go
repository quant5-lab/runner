package codegen

import (
	"regexp"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

func TestCurrencyConverterHandler_CanHandle(t *testing.T) {
	handler := NewCurrencyConverterHandler()

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{"strategy.convert_to_account", "strategy.convert_to_account", true},
		{"strategy.convert_to_symbol", "strategy.convert_to_symbol", true},
		{"strategy.entry", "strategy.entry", false},
		{"ta.sma", "ta.sma", false},
		{"convert_to_account", "convert_to_account", false},
		{"convert_to_symbol", "convert_to_symbol", false},
		{"math.abs", "math.abs", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.CanHandle(tt.funcName)
			if got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestCurrencyConverterHandler_GenerateCode(t *testing.T) {
	tests := []struct {
		name          string
		script        string
		expectedRegex string
	}{
		{
			name: "convert_to_account with literal",
			script: `//@version=5
strategy('test')
x = strategy.convert_to_account(100)`,
			expectedRegex: `strat\.ConvertToAccount\(\s*100\s*\)`,
		},
		{
			name: "convert_to_symbol with literal",
			script: `//@version=5
strategy('test')
x = strategy.convert_to_symbol(200.5)`,
			expectedRegex: `strat\.ConvertToSymbol\(\s*200\.5\s*\)`,
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
				t.Fatalf("Generation failed: %v", err)
			}

			code := result.FunctionBody
			matched, err := regexp.MatchString(tt.expectedRegex, code)
			if err != nil {
				t.Fatalf("Regex compilation error: %v", err)
			}
			if !matched {
				t.Errorf("Generated code missing expected pattern:\n%s\n\nGenerated:\n%s", tt.expectedRegex, code)
			}
		})
	}
}
