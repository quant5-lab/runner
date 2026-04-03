package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestTimeframeFuncCallHandler_CanHandle(t *testing.T) {
	h := NewTimeframeFuncCallHandler()

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{"in_seconds", "timeframe.in_seconds", true},
		{"from_seconds", "timeframe.from_seconds", true},
		{"change", "timeframe.change", true},

		{"period_is_variable", "timeframe.period", false},
		{"multiplier_is_variable", "timeframe.multiplier", false},
		{"ta_sma", "ta.sma", false},
		{"bare_change", "change", false},
		{"empty", "", false},
		{"case_sensitive", "Timeframe.Change", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := h.CanHandle(tt.funcName); got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestTimeframeFuncCallHandler_GenerateCode(t *testing.T) {
	h := NewTimeframeFuncCallHandler()
	g := newTestGenerator()

	tests := []struct {
		name        string
		funcName    string
		args        []ast.Expression
		wantContain []string
		wantAbsent  []string
	}{
		{
			"in_seconds_no_args",
			"timeframe.in_seconds",
			nil,
			[]string{"TimeframeToSeconds", "ctx.Timeframe"},
			nil,
		},
		{
			"in_seconds_literal_arg",
			"timeframe.in_seconds",
			[]ast.Expression{&ast.Literal{Value: "1D"}},
			[]string{"TimeframeToSeconds"},
			[]string{"ctx.Timeframe"},
		},
		{
			"in_seconds_identifier_arg",
			"timeframe.in_seconds",
			[]ast.Expression{&ast.Identifier{Name: "tf"}},
			[]string{"TimeframeToSeconds"},
			nil,
		},
		{
			"from_seconds_literal",
			"timeframe.from_seconds",
			[]ast.Expression{&ast.Literal{Value: float64(86400)}},
			[]string{"TimeframeFromSeconds", "int64"},
			nil,
		},
		{
			"from_seconds_identifier",
			"timeframe.from_seconds",
			[]ast.Expression{&ast.Identifier{Name: "secs"}},
			[]string{"TimeframeFromSeconds"},
			nil,
		},
		{
			"change_returns_float64",
			"timeframe.change",
			[]ast.Expression{&ast.Literal{Value: "1D"}},
			[]string{"float64", "AlignTimestampToPeriod"},
			[]string{"bool"},
		},
		{
			"change_with_variable_arg",
			"timeframe.change",
			[]ast.Expression{&ast.Identifier{Name: "higherTf"}},
			[]string{"float64", "AlignTimestampToPeriod"},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "timeframe"},
					Property: &ast.Identifier{Name: strings.TrimPrefix(tt.funcName, "timeframe.")},
				},
				Arguments: tt.args,
			}

			code, err := h.GenerateCode(g, call)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, want := range tt.wantContain {
				if !contains(code, want) {
					t.Errorf("missing %q in: %s", want, code)
				}
			}
			for _, absent := range tt.wantAbsent {
				if contains(code, absent) {
					t.Errorf("unexpected %q in: %s", absent, code)
				}
			}
		})
	}
}

func TestTimeframeFuncCallHandler_ErrorCases(t *testing.T) {
	h := NewTimeframeFuncCallHandler()
	g := newTestGenerator()

	tests := []struct {
		name     string
		funcName string
		args     []ast.Expression
	}{
		{"from_seconds_no_args", "from_seconds", nil},
		{"change_no_args", "change", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "timeframe"},
					Property: &ast.Identifier{Name: tt.funcName},
				},
				Arguments: tt.args,
			}

			_, err := h.GenerateCode(g, call)
			if err == nil {
				t.Errorf("timeframe.%s with no args should error", tt.funcName)
			}
		})
	}
}

func TestTimeframeChangeInlineHandler_CanHandle(t *testing.T) {
	h := NewTimeframeChangeInlineHandler()

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{"timeframe.change", "timeframe.change", true},
		{"ta.change", "ta.change", false},
		{"bare_change", "change", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := h.CanHandle(tt.funcName); got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestTimeframeChangeInlineHandler_GenerateInline(t *testing.T) {
	h := NewTimeframeChangeInlineHandler()
	g := newTestGenerator()

	tests := []struct {
		name        string
		args        []ast.Expression
		wantContain []string
		wantAbsent  []string
	}{
		{
			"returns_bool",
			[]ast.Expression{&ast.Literal{Value: "1D"}},
			[]string{"bool", "AlignTimestampToPeriod"},
			[]string{"float64"},
		},
		{
			"variable_timeframe",
			[]ast.Expression{&ast.Identifier{Name: "higherTf"}},
			[]string{"bool", "higherTf"},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "timeframe"},
					Property: &ast.Identifier{Name: "change"},
				},
				Arguments: tt.args,
			}

			code, err := h.GenerateInline(expr, g)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, want := range tt.wantContain {
				if !contains(code, want) {
					t.Errorf("missing %q in: %s", want, code)
				}
			}
			for _, absent := range tt.wantAbsent {
				if contains(code, absent) {
					t.Errorf("unexpected %q in: %s", absent, code)
				}
			}
		})
	}
}

func TestTimeframeChangeInlineHandler_ErrorCases(t *testing.T) {
	h := NewTimeframeChangeInlineHandler()
	g := newTestGenerator()

	expr := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "timeframe"},
			Property: &ast.Identifier{Name: "change"},
		},
	}

	if _, err := h.GenerateInline(expr, g); err == nil {
		t.Error("GenerateInline with no args should error")
	}
}
