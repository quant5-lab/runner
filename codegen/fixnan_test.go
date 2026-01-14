package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

func TestFixnanHandler_CanHandle(t *testing.T) {
	handler := &FixnanHandler{}

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{"fixnan function", "fixnan", true},
		{"ta.sma not handled", "ta.sma", false},
		{"random function", "foo", false},
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

func TestFixnanIntegration(t *testing.T) {
	pineScript := `//@version=5
indicator("Fixnan Integration", overlay=true)
pivot = pivothigh(5, 5)
filled = fixnan(pivot)
plot(filled, title="Filled Pivot")
`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(pineScript))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := parser.NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	result, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	code := result.FunctionBody

	requiredPatterns := []string{
		"var fixnanState_filled = math.NaN()",
		"if !math.IsNaN(pivotSeries.GetCurrent())",
		"fixnanState_filled = pivotSeries.GetCurrent()",
		"filledSeries.Set(fixnanState_filled)",
	}

	for _, pattern := range requiredPatterns {
		if !strings.Contains(code, pattern) {
			t.Errorf("Generated code missing pattern %q", pattern)
		}
	}
}
