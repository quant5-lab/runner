package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

func TestSumConditionalPreRegistration(t *testing.T) {
	pineCode := `//@version=5
strategy("Test", overlay=true)

sr_a1 = close[1]
sr_a = close

sr_gains = sum(sr_a1 > sr_a ? 1 : 0, 20)
sr_losses = sum(sr_a1 < sr_a ? 1 : 0, 20)

plot(sr_gains)
plot(sr_losses)
`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	parseResult, err := p.ParseBytes("test.pine", []byte(pineCode))
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
		t.Fatalf("Generate error: %v", err)
	}

	code := result.FunctionBody

	// Check if ternary vars are declared
	if !strings.Contains(code, "var ternary_") {
		t.Logf("Generated code (first 200 lines):")
		lines := strings.Split(code, "\n")
		for i, line := range lines {
			if i >= 200 {
				break
			}
			t.Logf("%d: %s", i+1, line)
		}
		t.Errorf("Missing ternary Series declaration")
	} else {
		t.Logf("Found ternary declarations!")
	}
}
