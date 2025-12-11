package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestValuewhenHandler_CanHandle(t *testing.T) {
	handler := &ValuewhenHandler{}

	tests := []struct {
		funcName string
		want     bool
	}{
		{"ta.valuewhen", true},
		{"valuewhen", true},
		{"ta.sma", false},
		{"ta.change", false},
		{"other", false},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			if got := handler.CanHandle(tt.funcName); got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestValuewhenHandler_GenerateCode_ArgumentValidation(t *testing.T) {
	handler := &ValuewhenHandler{}
	g := newTestGenerator()

	tests := []struct {
		name    string
		args    []ast.Expression
		wantErr string
	}{
		{
			name:    "no arguments",
			args:    []ast.Expression{},
			wantErr: "requires 3 arguments",
		},
		{
			name: "one argument",
			args: []ast.Expression{
				&ast.Identifier{Name: "cond"},
			},
			wantErr: "requires 3 arguments",
		},
		{
			name: "two arguments",
			args: []ast.Expression{
				&ast.Identifier{Name: "cond"},
				&ast.Identifier{Name: "src"},
			},
			wantErr: "requires 3 arguments",
		},
		{
			name: "non-literal occurrence",
			args: []ast.Expression{
				&ast.Identifier{Name: "cond"},
				&ast.Identifier{Name: "src"},
				&ast.Identifier{Name: "occ"},
			},
			wantErr: "occurrence must be literal",
		},
		{
			name: "string occurrence",
			args: []ast.Expression{
				&ast.Identifier{Name: "cond"},
				&ast.Identifier{Name: "src"},
				&ast.Literal{Value: "invalid"},
			},
			wantErr: "period must be numeric",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "valuewhen"},
				Arguments: tt.args,
			}

			_, err := handler.GenerateCode(g, "test", call)
			if err == nil {
				t.Error("expected error, got nil")
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestValuewhenHandler_GenerateCode_ValidCases(t *testing.T) {
	handler := &ValuewhenHandler{}

	tests := []struct {
		name            string
		conditionExpr   ast.Expression
		sourceExpr      ast.Expression
		occurrence      int
		expectCondition string
		expectSource    string
		expectOccur     string
	}{
		{
			name:            "series condition, builtin source, occurrence 0",
			conditionExpr:   &ast.Identifier{Name: "bullish"},
			sourceExpr:      &ast.MemberExpression{Object: &ast.Identifier{Name: "bar"}, Property: &ast.Identifier{Name: "Close"}},
			occurrence:      0,
			expectCondition: "bullishSeries.Get(lookbackOffset)",
			expectSource:    "ctx.Data[i-lookbackOffset].Close",
			expectOccur:     "0",
		},
		{
			name:            "series condition, series source, occurrence 1",
			conditionExpr:   &ast.Identifier{Name: "crossover"},
			sourceExpr:      &ast.Identifier{Name: "high"},
			occurrence:      1,
			expectCondition: "crossoverSeries.Get(lookbackOffset)",
			expectSource:    "highSeries.Get(lookbackOffset)",
			expectOccur:     "1",
		},
		{
			name:            "series condition, series source, high occurrence",
			conditionExpr:   &ast.Identifier{Name: "signal"},
			sourceExpr:      &ast.Identifier{Name: "price"},
			occurrence:      5,
			expectCondition: "signalSeries.Get(lookbackOffset)",
			expectSource:    "priceSeries.Get(lookbackOffset)",
			expectOccur:     "5",
		},
		{
			name:            "builtin bar field sources",
			conditionExpr:   &ast.Identifier{Name: "cond"},
			sourceExpr:      &ast.MemberExpression{Object: &ast.Identifier{Name: "bar"}, Property: &ast.Identifier{Name: "High"}},
			occurrence:      0,
			expectCondition: "condSeries.Get(lookbackOffset)",
			expectSource:    "ctx.Data[i-lookbackOffset].High",
			expectOccur:     "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()

			call := &ast.CallExpression{
				Callee: &ast.Identifier{Name: "valuewhen"},
				Arguments: []ast.Expression{
					tt.conditionExpr,
					tt.sourceExpr,
					&ast.Literal{Value: float64(tt.occurrence)},
				},
			}

			code, err := handler.GenerateCode(g, "result", call)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(code, "Inline valuewhen") {
				t.Error("expected inline valuewhen comment")
			}

			if !strings.Contains(code, "resultSeries.Set(func() float64 {") {
				t.Error("expected Series.Set() with IIFE")
			}

			if !strings.Contains(code, "occurrenceCount := 0") {
				t.Error("expected occurrenceCount initialization")
			}

			if !strings.Contains(code, "for lookbackOffset := 0; lookbackOffset <= i; lookbackOffset++") {
				t.Error("expected lookback loop")
			}

			if !strings.Contains(code, tt.expectCondition+" != 0") {
				t.Errorf("expected condition check %q in generated code", tt.expectCondition)
			}

			if !strings.Contains(code, "occurrenceCount == "+tt.expectOccur) {
				t.Errorf("expected occurrence check %q in generated code", tt.expectOccur)
			}

			if !strings.Contains(code, "return") || !strings.Contains(code, "lookbackOffset") {
				t.Error("expected return statement with lookbackOffset-based access")
			}

			if !strings.Contains(code, "occurrenceCount++") {
				t.Error("expected occurrenceCount increment")
			}

			if !strings.Contains(code, "return math.NaN()") {
				t.Error("expected NaN fallback return")
			}
		})
	}
}

func TestValuewhenHandler_IntegrationWithGenerator(t *testing.T) {
	handler := &ValuewhenHandler{}

	tests := []struct {
		name       string
		varName    string
		condition  ast.Expression
		source     ast.Expression
		occurrence int
	}{
		{
			name:       "simple identifier condition and source",
			varName:    "lastValue",
			condition:  &ast.Identifier{Name: "trigger"},
			source:     &ast.Identifier{Name: "value"},
			occurrence: 0,
		},
		{
			name:    "bar field source",
			varName: "lastClose",
			condition: &ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.MemberExpression{Object: &ast.Identifier{Name: "bar"}, Property: &ast.Identifier{Name: "Close"}},
				Right:    &ast.MemberExpression{Object: &ast.Identifier{Name: "bar"}, Property: &ast.Identifier{Name: "Open"}},
			},
			source:     &ast.MemberExpression{Object: &ast.Identifier{Name: "bar"}, Property: &ast.Identifier{Name: "Close"}},
			occurrence: 0,
		},
		{
			name:       "historical occurrence",
			varName:    "nthValue",
			condition:  &ast.Identifier{Name: "signal"},
			source:     &ast.Identifier{Name: "price"},
			occurrence: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()

			call := &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.valuewhen"},
				Arguments: []ast.Expression{
					tt.condition,
					tt.source,
					&ast.Literal{Value: float64(tt.occurrence)},
				},
			}

			code, err := handler.GenerateCode(g, tt.varName, call)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(code, tt.varName+"Series.Set") {
				t.Errorf("expected %sSeries.Set in generated code", tt.varName)
			}

			if !strings.Contains(code, "func() float64") {
				t.Error("expected IIFE pattern")
			}

			if strings.Count(code, "for lookbackOffset") != 1 {
				t.Error("expected exactly one lookback loop")
			}

			if strings.Count(code, "return") != 2 {
				t.Error("expected two return statements (match and NaN fallback)")
			}
		})
	}
}

func TestGenerator_ConvertSeriesAccessToOffset(t *testing.T) {
	g := newTestGenerator()

	tests := []struct {
		name       string
		seriesCode string
		offsetVar  string
		want       string
	}{
		{
			name:       "bar.Close with offset",
			seriesCode: "bar.Close",
			offsetVar:  "lookbackOffset",
			want:       "ctx.Data[i-lookbackOffset].Close",
		},
		{
			name:       "bar.High with offset",
			seriesCode: "bar.High",
			offsetVar:  "lookbackOffset",
			want:       "ctx.Data[i-lookbackOffset].High",
		},
		{
			name:       "bar.Low with offset",
			seriesCode: "bar.Low",
			offsetVar:  "offset",
			want:       "ctx.Data[i-offset].Low",
		},
		{
			name:       "bar.Open with offset",
			seriesCode: "bar.Open",
			offsetVar:  "o",
			want:       "ctx.Data[i-o].Open",
		},
		{
			name:       "bar.Volume with offset",
			seriesCode: "bar.Volume",
			offsetVar:  "lookbackOffset",
			want:       "ctx.Data[i-lookbackOffset].Volume",
		},
		{
			name:       "Series.GetCurrent() to Get(offset)",
			seriesCode: "priceSeries.GetCurrent()",
			offsetVar:  "lookbackOffset",
			want:       "priceSeries.Get(lookbackOffset)",
		},
		{
			name:       "different series name",
			seriesCode: "sma20Series.GetCurrent()",
			offsetVar:  "lookbackOffset",
			want:       "sma20Series.Get(lookbackOffset)",
		},
		{
			name:       "Series.Get(0) to Get(offset)",
			seriesCode: "valueSeries.Get(0)",
			offsetVar:  "lookbackOffset",
			want:       "valueSeries.Get(lookbackOffset)",
		},
		{
			name:       "Series.Get(N) to Get(offset) - replaces existing offset",
			seriesCode: "dataSeries.Get(5)",
			offsetVar:  "newOffset",
			want:       "dataSeries.Get(newOffset)",
		},
		{
			name:       "non-series expression returns unchanged",
			seriesCode: "42.0",
			offsetVar:  "lookbackOffset",
			want:       "42.0",
		},
		{
			name:       "literal identifier returns unchanged",
			seriesCode: "someConstant",
			offsetVar:  "lookbackOffset",
			want:       "someConstant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := g.convertSeriesAccessToOffset(tt.seriesCode, tt.offsetVar)
			if got != tt.want {
				t.Errorf("convertSeriesAccessToOffset(%q, %q) = %q, want %q",
					tt.seriesCode, tt.offsetVar, got, tt.want)
			}
		})
	}
}
