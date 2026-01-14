package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/parser"
)

func parseArrowFunctionForInitializer(t *testing.T, source string) *ast.ArrowFunctionExpression {
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(source))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := parser.NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	var arrowFunc *ast.ArrowFunctionExpression
	for _, stmt := range program.Body {
		if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
			for _, decl := range varDecl.Declarations {
				if arrow, ok := decl.Init.(*ast.ArrowFunctionExpression); ok {
					arrowFunc = arrow
					break
				}
			}
		}
	}

	if arrowFunc == nil {
		t.Fatal("Arrow function not found in parsed AST")
	}

	return arrowFunc
}

func TestArrowLocalSeriesInitializer_GenerateInitializations(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		wantInCode []string
		wantNone   []string
	}{
		{
			name: "variables used in TA functions",
			source: `
dirmov(len) =>
    up = change(high)
    down = change(low)
    truerange = rma(tr, len)
    plus = rma(up, len)
    [plus, 0]
`,
			wantInCode: []string{
				`upSeries := arrowCtx.GetOrCreateSeries("up")`,
			},
			wantNone: []string{
				`downSeries`,
				`plusSeries`,
				`truerangeSeries`,
			},
		},
		{
			name: "simple arithmetic without TA",
			source: `
add(a, b) =>
    result = a + b
    result
`,
			wantInCode: []string{},
			wantNone: []string{
				`resultSeries`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrowFunc := parseArrowFunctionForInitializer(t, tt.source)

			initializer := NewArrowLocalSeriesInitializer("\t")
			code := initializer.GenerateInitializations(arrowFunc)

			for _, expected := range tt.wantInCode {
				if !strings.Contains(code, expected) {
					t.Errorf("Generated code missing expected pattern %q\nGot:\n%s", expected, code)
				}
			}

			for _, notExpected := range tt.wantNone {
				if strings.Contains(code, notExpected) {
					t.Errorf("Generated code should not contain %q\nGot:\n%s", notExpected, code)
				}
			}
		})
	}
}
