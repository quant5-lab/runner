package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

type conditionExpressionCase struct {
	name         string
	expr         ast.Expression
	wantExact    string
	wantContains []string
	wantAbsent   []string
}

func newConditionExpressionTestGenerator() *generator {
	typeSystem := NewTypeInferenceEngine()
	return &generator{
		variables:        make(map[string]string),
		varInits:         make(map[string]ast.Expression),
		constants:        make(map[string]interface{}),
		builtinHandler:   NewBuiltinIdentifierHandler(),
		typeSystem:       typeSystem,
		boolConverter:    NewBooleanConverter(typeSystem),
		literalFormatter: NewLiteralFormatter(),
	}
}

func assertConditionExpression(t *testing.T, gen *generator, tc conditionExpressionCase) {
	t.Helper()
	code, err := gen.generateConditionExpression(tc.expr)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if tc.wantExact != "" && code != tc.wantExact {
		t.Errorf("got %q, want %q", code, tc.wantExact)
	}
	for _, pattern := range tc.wantContains {
		if !strings.Contains(code, pattern) {
			t.Errorf("missing %q in: %s", pattern, code)
		}
	}
	for _, absent := range tc.wantAbsent {
		if strings.Contains(code, absent) {
			t.Errorf("unexpected %q in: %s", absent, code)
		}
	}
}

func identifierCase(name, wantExact string, wantContains, wantAbsent []string) conditionExpressionCase {
	return conditionExpressionCase{
		name:         name,
		expr:         &ast.Identifier{Name: name},
		wantExact:    wantExact,
		wantContains: wantContains,
		wantAbsent:   wantAbsent,
	}
}

func TestConditionExpression_BuiltinIdentifiers(t *testing.T) {
	gen := newConditionExpressionTestGenerator()

	tests := []conditionExpressionCase{
		identifierCase("close", "bar.Close", nil, nil),
		identifierCase("open", "bar.Open", nil, nil),
		identifierCase("high", "bar.High", nil, nil),
		identifierCase("low", "bar.Low", nil, nil),
		identifierCase("volume", "bar.Volume", nil, nil),
		identifierCase("bar_index", "float64(i)", nil, nil),
		identifierCase("na", "math.NaN()", nil, nil),
		identifierCase("hl2", "", []string{"bar.High", "bar.Low", "/ 2"}, []string{"Series"}),
		identifierCase("hlc3", "", []string{"bar.High", "bar.Low", "bar.Close", "/ 3"}, []string{"Series"}),
		identifierCase("ohlc4", "", []string{"bar.Open", "bar.High", "bar.Low", "bar.Close", "/ 4"}, []string{"Series"}),
		identifierCase("hlcc4", "", []string{"bar.High", "bar.Low", "bar.Close", "/ 4"}, []string{"Series"}),
		identifierCase("tr", "", []string{"math.Max", "bar.High", "bar.Low"}, nil),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertConditionExpression(t, gen, tt) })
	}
}

func TestConditionExpression_SecurityContext(t *testing.T) {
	gen := newConditionExpressionTestGenerator()
	gen.inSecurityContext = true

	tests := []conditionExpressionCase{
		identifierCase("close", "", []string{"closeSeries.GetCurrent()"}, []string{"bar."}),
		identifierCase("bar_index", "", []string{"float64(ctx.BarIndex)"}, []string{"float64(i)"}),
		identifierCase("hl2", "", []string{"highSeries.GetCurrent()", "lowSeries.GetCurrent()", "/ 2"}, []string{"bar."}),
		identifierCase("hlc3", "", []string{"highSeries.GetCurrent()", "lowSeries.GetCurrent()", "closeSeries.GetCurrent()", "/ 3"}, []string{"bar."}),
		identifierCase("ohlc4", "", []string{"openSeries.GetCurrent()", "highSeries.GetCurrent()", "lowSeries.GetCurrent()", "closeSeries.GetCurrent()", "/ 4"}, []string{"bar."}),
		identifierCase("hlcc4", "", []string{"highSeries.GetCurrent()", "lowSeries.GetCurrent()", "closeSeries.GetCurrent()", "/ 4"}, []string{"bar."}),
		identifierCase("tr", "", []string{"highSeries.GetCurrent()", "lowSeries.GetCurrent()", "closeSeries.Get(1)"}, []string{"bar."}),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertConditionExpression(t, gen, tt) })
	}
}

func TestConditionExpression_NaContextIndependence(t *testing.T) {
	normal := newConditionExpressionTestGenerator()
	security := newConditionExpressionTestGenerator()
	security.inSecurityContext = true

	naExpr := &ast.Identifier{Name: "na"}

	normalCode, err := normal.generateConditionExpression(naExpr)
	if err != nil {
		t.Fatalf("normal context error: %v", err)
	}
	securityCode, err := security.generateConditionExpression(naExpr)
	if err != nil {
		t.Fatalf("security context error: %v", err)
	}

	if normalCode != "math.NaN()" {
		t.Errorf("normal context: got %q, want %q", normalCode, "math.NaN()")
	}
	if normalCode != securityCode {
		t.Errorf("context mismatch: normal=%q, security=%q", normalCode, securityCode)
	}
}

func TestConditionExpression_BuiltinPriorityOverConstants(t *testing.T) {
	gen := newConditionExpressionTestGenerator()
	gen.constants["close"] = 42
	gen.constants["hl2"] = 99

	tests := []conditionExpressionCase{
		identifierCase("close", "bar.Close", nil, []string{"42"}),
		identifierCase("hl2", "", []string{"bar.High", "bar.Low", "/ 2"}, []string{"99", "Series"}),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertConditionExpression(t, gen, tt) })
	}
}

func TestConditionExpression_IdentifierFallthrough(t *testing.T) {
	gen := newConditionExpressionTestGenerator()
	gen.constants["length"] = 14
	gen.constants["src"] = "input.source"

	tests := []conditionExpressionCase{
		{name: "user variable", expr: &ast.Identifier{Name: "myVar"}, wantExact: "myVarSeries.GetCurrent()"},
		{name: "numeric constant", expr: &ast.Identifier{Name: "length"}, wantExact: "length"},
		{name: "input.source constant", expr: &ast.Identifier{Name: "src"}, wantExact: "srcSeries.GetCurrent()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertConditionExpression(t, gen, tt) })
	}
}

func TestConditionExpression_BuiltinsInCompoundExpressions(t *testing.T) {
	gen := newConditionExpressionTestGenerator()

	tests := []conditionExpressionCase{
		{
			name: "close > hl2",
			expr: &ast.BinaryExpression{
				Left: &ast.Identifier{Name: "close"}, Operator: ">", Right: &ast.Identifier{Name: "hl2"},
			},
			wantContains: []string{"bar.Close", "bar.High", "bar.Low", ">"},
			wantAbsent:   []string{"hl2Series"},
		},
		{
			name: "ohlc4 + hlc3",
			expr: &ast.BinaryExpression{
				Left: &ast.Identifier{Name: "ohlc4"}, Operator: "+", Right: &ast.Identifier{Name: "hlc3"},
			},
			wantContains: []string{"bar.Open", "/ 4", "/ 3"},
			wantAbsent:   []string{"ohlc4Series", "hlc3Series"},
		},
		{
			name: "high == low",
			expr: &ast.BinaryExpression{
				Left: &ast.Identifier{Name: "high"}, Operator: "==", Right: &ast.Identifier{Name: "low"},
			},
			wantContains: []string{"bar.High", "bar.Low", "=="},
		},
		{
			name: "close > 100",
			expr: &ast.BinaryExpression{
				Left: &ast.Identifier{Name: "close"}, Operator: ">", Right: &ast.Literal{Value: 100.0},
			},
			wantContains: []string{"bar.Close", ">", "100"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertConditionExpression(t, gen, tt) })
	}
}
