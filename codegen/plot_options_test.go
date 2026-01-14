package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestParsePlotOptions_SimpleVariable(t *testing.T) {
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "sma20"},
		},
	}

	opts := ParsePlotOptions(call)

	if opts.Variable != "sma20" {
		t.Errorf("Expected variable 'sma20', got '%s'", opts.Variable)
	}
	if opts.Title != "sma20" {
		t.Errorf("Expected title 'sma20', got '%s'", opts.Title)
	}
}

func TestParsePlotOptions_MemberExpression(t *testing.T) {
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "sma50"},
				Property: &ast.Literal{Value: 0},
				Computed: true,
			},
		},
	}

	opts := ParsePlotOptions(call)

	if opts.Variable != "sma50" {
		t.Errorf("Expected variable 'sma50', got '%s'", opts.Variable)
	}
}

func TestParsePlotOptions_WithTitle(t *testing.T) {
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "ema20"},
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key:   &ast.Identifier{Name: "title"},
						Value: &ast.Literal{Value: "EMA 20"},
					},
				},
			},
		},
	}

	opts := ParsePlotOptions(call)

	if opts.Variable != "ema20" {
		t.Errorf("Expected variable 'ema20', got '%s'", opts.Variable)
	}
	if opts.Title != "EMA 20" {
		t.Errorf("Expected title 'EMA 20', got '%s'", opts.Title)
	}
}

func TestParsePlotOptions_EmptyCall(t *testing.T) {
	call := &ast.CallExpression{
		Arguments: []ast.Expression{},
	}

	opts := ParsePlotOptions(call)

	if opts.Variable != "" {
		t.Errorf("Expected empty variable, got '%s'", opts.Variable)
	}
	if opts.Title != "" {
		t.Errorf("Expected empty title, got '%s'", opts.Title)
	}
}

func TestParsePlotOptions_MultipleProperties(t *testing.T) {
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "rsi"},
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key:   &ast.Identifier{Name: "color"},
						Value: &ast.Identifier{Name: "blue"},
					},
					{
						Key:   &ast.Identifier{Name: "title"},
						Value: &ast.Literal{Value: "RSI Indicator"},
					},
					{
						Key:   &ast.Identifier{Name: "linewidth"},
						Value: &ast.Literal{Value: 2},
					},
				},
			},
		},
	}

	opts := ParsePlotOptions(call)

	if opts.Variable != "rsi" {
		t.Errorf("Expected variable 'rsi', got '%s'", opts.Variable)
	}
	if opts.Title != "RSI Indicator" {
		t.Errorf("Expected title 'RSI Indicator', got '%s'", opts.Title)
	}
}

// TestParsePlotOptions_StyleParameter verifies style expression parsing
func TestParsePlotOptions_StyleParameter(t *testing.T) {
	tests := []struct {
		name      string
		styleExpr ast.Expression
		wantNil   bool
	}{
		{
			name:      "style as constant",
			styleExpr: &ast.MemberExpression{Object: &ast.Identifier{Name: "plot"}, Property: &ast.Identifier{Name: "style_circles"}},
			wantNil:   false,
		},
		{
			name:      "style as string literal",
			styleExpr: &ast.Literal{Value: "circles"},
			wantNil:   false,
		},
		{
			name:      "style as linebr constant",
			styleExpr: &ast.MemberExpression{Object: &ast.Identifier{Name: "plot"}, Property: &ast.Identifier{Name: "style_linebr"}},
			wantNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{Key: &ast.Identifier{Name: "style"}, Value: tt.styleExpr},
						},
					},
				},
			}

			opts := ParsePlotOptions(call)

			if tt.wantNil && opts.StyleExpr != nil {
				t.Error("Expected StyleExpr to be nil")
			}
			if !tt.wantNil && opts.StyleExpr == nil {
				t.Error("Expected StyleExpr to be set")
			}
		})
	}
}

// TestParsePlotOptions_LineWidthParameter verifies linewidth expression parsing
func TestParsePlotOptions_LineWidthParameter(t *testing.T) {
	tests := []struct {
		name          string
		linewidthExpr ast.Expression
		wantNil       bool
	}{
		{name: "linewidth 1", linewidthExpr: &ast.Literal{Value: float64(1)}, wantNil: false},
		{name: "linewidth 2", linewidthExpr: &ast.Literal{Value: float64(2)}, wantNil: false},
		{name: "linewidth 8", linewidthExpr: &ast.Literal{Value: float64(8)}, wantNil: false},
		{name: "linewidth 10", linewidthExpr: &ast.Literal{Value: float64(10)}, wantNil: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "sma"},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{Key: &ast.Identifier{Name: "linewidth"}, Value: tt.linewidthExpr},
						},
					},
				},
			}

			opts := ParsePlotOptions(call)

			if tt.wantNil && opts.LineWidthExpr != nil {
				t.Error("Expected LineWidthExpr to be nil")
			}
			if !tt.wantNil && opts.LineWidthExpr == nil {
				t.Error("Expected LineWidthExpr to be set")
			}
		})
	}
}

// TestParsePlotOptions_TranspParameter verifies transparency expression parsing
func TestParsePlotOptions_TranspParameter(t *testing.T) {
	tests := []struct {
		name       string
		transpExpr ast.Expression
		wantNil    bool
	}{
		{name: "transp 0", transpExpr: &ast.Literal{Value: float64(0)}, wantNil: false},
		{name: "transp 30", transpExpr: &ast.Literal{Value: float64(30)}, wantNil: false},
		{name: "transp 50", transpExpr: &ast.Literal{Value: float64(50)}, wantNil: false},
		{name: "transp 100", transpExpr: &ast.Literal{Value: float64(100)}, wantNil: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "ema"},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{Key: &ast.Identifier{Name: "transp"}, Value: tt.transpExpr},
						},
					},
				},
			}

			opts := ParsePlotOptions(call)

			if tt.wantNil && opts.TranspExpr != nil {
				t.Error("Expected TranspExpr to be nil")
			}
			if !tt.wantNil && opts.TranspExpr == nil {
				t.Error("Expected TranspExpr to be set")
			}
		})
	}
}

// TestParsePlotOptions_PaneParameter verifies pane expression parsing
func TestParsePlotOptions_PaneParameter(t *testing.T) {
	tests := []struct {
		name     string
		paneExpr ast.Expression
		wantNil  bool
	}{
		{name: "pane indicator", paneExpr: &ast.Literal{Value: "indicator"}, wantNil: false},
		{name: "pane main", paneExpr: &ast.Literal{Value: "main"}, wantNil: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "rsi"},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{Key: &ast.Identifier{Name: "pane"}, Value: tt.paneExpr},
						},
					},
				},
			}

			opts := ParsePlotOptions(call)

			if tt.wantNil && opts.PaneExpr != nil {
				t.Error("Expected PaneExpr to be nil")
			}
			if !tt.wantNil && opts.PaneExpr == nil {
				t.Error("Expected PaneExpr to be set")
			}
		})
	}
}

// TestParsePlotOptions_AllParameters verifies all parameters parsed together
func TestParsePlotOptions_AllParameters(t *testing.T) {
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "macd"},
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{Key: &ast.Identifier{Name: "title"}, Value: &ast.Literal{Value: "MACD Line"}},
					{Key: &ast.Identifier{Name: "color"}, Value: &ast.MemberExpression{Object: &ast.Identifier{Name: "color"}, Property: &ast.Identifier{Name: "blue"}}},
					{Key: &ast.Identifier{Name: "style"}, Value: &ast.MemberExpression{Object: &ast.Identifier{Name: "plot"}, Property: &ast.Identifier{Name: "style_line"}}},
					{Key: &ast.Identifier{Name: "linewidth"}, Value: &ast.Literal{Value: float64(2)}},
					{Key: &ast.Identifier{Name: "transp"}, Value: &ast.Literal{Value: float64(20)}},
					{Key: &ast.Identifier{Name: "offset"}, Value: &ast.Literal{Value: float64(-1)}},
					{Key: &ast.Identifier{Name: "pane"}, Value: &ast.Literal{Value: "indicator"}},
				},
			},
		},
	}

	opts := ParsePlotOptions(call)

	if opts.Variable != "macd" {
		t.Errorf("Expected variable 'macd', got '%s'", opts.Variable)
	}
	if opts.Title != "MACD Line" {
		t.Errorf("Expected title 'MACD Line', got '%s'", opts.Title)
	}
	if opts.ColorExpr == nil {
		t.Error("Expected ColorExpr to be set")
	}
	if opts.StyleExpr == nil {
		t.Error("Expected StyleExpr to be set")
	}
	if opts.LineWidthExpr == nil {
		t.Error("Expected LineWidthExpr to be set")
	}
	if opts.TranspExpr == nil {
		t.Error("Expected TranspExpr to be set")
	}
	if opts.OffsetExpr == nil {
		t.Error("Expected OffsetExpr to be set")
	}
	if opts.PaneExpr == nil {
		t.Error("Expected PaneExpr to be set")
	}
}
