package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

func TestDrawingArrayConstructors_ParseAndCodegen(t *testing.T) {
	tests := []struct {
		name         string
		script       string
		wantContains string
		wantAbsent   []string
	}{
		{
			name: "empty line array",
			script: `//@version=5
indicator("test")
lines = array.new_line()
plot(close)
`,
			wantContains: "[]float64{}",
		},
		{
			name: "label array with size",
			script: `//@version=5
indicator("test")
labels = array.new_label(10)
plot(close)
`,
			wantContains: "make([]float64, int(10))",
		},
		{
			name: "empty box array",
			script: `//@version=5
indicator("test")
boxes = array.new_box()
plot(close)
`,
			wantContains: "[]float64{}",
		},
		{
			name: "table array with size",
			script: `//@version=5
indicator("test")
tables = array.new_table(5)
plot(close)
`,
			wantContains: "make([]float64, int(5))",
		},
		{
			name: "empty linefill array",
			script: `//@version=5
indicator("test")
fills = array.new_linefill()
plot(close)
`,
			wantContains: "[]float64{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := mustGenerateFromScript(t, tt.script)

			if !strings.Contains(code, tt.wantContains) {
				t.Errorf("want %q in generated code, got:\n%s", tt.wantContains, code)
			}

			for _, absent := range []string{"TODO", "unknown"} {
				if strings.Contains(code, absent) {
					t.Errorf("generated code must not contain %q", absent)
				}
			}
		})
	}
}

func mustGenerateFromScript(t *testing.T, script string) string {
	t.Helper()

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("parser creation: %v", err)
	}

	parsed, err := p.ParseBytes("test.pine", []byte(script))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	converter := parser.NewConverter()
	program, err := converter.ToESTree(parsed)
	if err != nil {
		t.Fatalf("conversion: %v", err)
	}

	result, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("codegen: %v", err)
	}

	return result.FunctionBody
}
