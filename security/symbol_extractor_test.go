package security

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/ticker"
)

func TestSymbolExtractor_Literal(t *testing.T) {
	extractor := NewSymbolExtractor()

	expr := &ast.Literal{Value: "BTCUSDT"}
	symbol, modType := extractor.Extract(expr)

	if symbol != "BTCUSDT" {
		t.Errorf("symbol = %q, want BTCUSDT", symbol)
	}

	if modType != "" {
		t.Errorf("modType = %q, want empty", modType)
	}
}

func TestSymbolExtractor_Identifier(t *testing.T) {
	extractor := NewSymbolExtractor()

	expr := &ast.Identifier{Name: "mySymbol"}
	symbol, modType := extractor.Extract(expr)

	if symbol != "mySymbol" {
		t.Errorf("symbol = %q, want mySymbol", symbol)
	}

	if modType != "" {
		t.Errorf("modType = %q, want empty", modType)
	}
}

func TestSymbolExtractor_MemberExpression(t *testing.T) {
	extractor := NewSymbolExtractor()

	expr := &ast.MemberExpression{
		Object:   &ast.Identifier{Name: "syminfo"},
		Property: &ast.Identifier{Name: "tickerid"},
	}

	symbol, modType := extractor.Extract(expr)

	if symbol != "syminfo.tickerid" {
		t.Errorf("symbol = %q, want syminfo.tickerid", symbol)
	}

	if modType != "" {
		t.Errorf("modType = %q, want empty", modType)
	}
}

func TestSymbolExtractor_HeikinashiCall(t *testing.T) {
	extractor := NewSymbolExtractor()

	expr := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "heikinashi"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "BTCUSDT"},
		},
	}

	symbol, modType := extractor.Extract(expr)

	if symbol != "BTCUSDT" {
		t.Errorf("symbol = %q, want BTCUSDT", symbol)
	}

	if modType != ticker.ModifierHeikinAshi {
		t.Errorf("modType = %q, want %q", modType, ticker.ModifierHeikinAshi)
	}
}

func TestSymbolExtractor_TickerHeikinashiCall(t *testing.T) {
	extractor := NewSymbolExtractor()

	expr := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ticker"},
			Property: &ast.Identifier{Name: "heikinashi"},
		},
		Arguments: []ast.Expression{
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "syminfo"},
				Property: &ast.Identifier{Name: "tickerid"},
			},
		},
	}

	symbol, modType := extractor.Extract(expr)

	if symbol != "syminfo.tickerid" {
		t.Errorf("symbol = %q, want syminfo.tickerid", symbol)
	}

	if modType != ticker.ModifierHeikinAshi {
		t.Errorf("modType = %q, want %q", modType, ticker.ModifierHeikinAshi)
	}
}

func TestSymbolExtractor_RenkoCall(t *testing.T) {
	extractor := NewSymbolExtractor()

	expr := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "renko"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "ETHUSDT"},
			&ast.Literal{Value: "ATR"},
			&ast.Literal{Value: 14.0},
		},
	}

	symbol, modType := extractor.Extract(expr)

	if symbol != "ETHUSDT" {
		t.Errorf("symbol = %q, want ETHUSDT", symbol)
	}

	if modType != ticker.ModifierRenko {
		t.Errorf("modType = %q, want %q", modType, ticker.ModifierRenko)
	}
}

func TestSymbolExtractor_KagiCall(t *testing.T) {
	extractor := NewSymbolExtractor()

	expr := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "kagi"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "AAPL"},
			&ast.Literal{Value: 3.0},
		},
	}

	symbol, modType := extractor.Extract(expr)

	if symbol != "AAPL" {
		t.Errorf("symbol = %q, want AAPL", symbol)
	}

	if modType != ticker.ModifierKagi {
		t.Errorf("modType = %q, want %q", modType, ticker.ModifierKagi)
	}
}

func TestSymbolExtractor_LinebreakCall(t *testing.T) {
	extractor := NewSymbolExtractor()

	expr := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "linebreak"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "TSLA"},
			&ast.Literal{Value: 3.0},
		},
	}

	symbol, modType := extractor.Extract(expr)

	if symbol != "TSLA" {
		t.Errorf("symbol = %q, want TSLA", symbol)
	}

	if modType != ticker.ModifierLineBreak {
		t.Errorf("modType = %q, want %q", modType, ticker.ModifierLineBreak)
	}
}

func TestSymbolExtractor_RangeCall(t *testing.T) {
	extractor := NewSymbolExtractor()

	expr := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "range"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "BTCUSDT"},
		},
	}

	symbol, modType := extractor.Extract(expr)

	if symbol != "BTCUSDT" {
		t.Errorf("symbol = %q, want BTCUSDT", symbol)
	}
	if modType != ticker.ModifierRange {
		t.Errorf("modType = %q, want %q", modType, ticker.ModifierRange)
	}
}

func TestSymbolExtractor_TickerRangeCall(t *testing.T) {
	extractor := NewSymbolExtractor()

	expr := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ticker"},
			Property: &ast.Identifier{Name: "range"},
		},
		Arguments: []ast.Expression{
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "syminfo"},
				Property: &ast.Identifier{Name: "tickerid"},
			},
		},
	}

	symbol, modType := extractor.Extract(expr)

	if symbol != "syminfo.tickerid" {
		t.Errorf("symbol = %q, want syminfo.tickerid", symbol)
	}
	if modType != ticker.ModifierRange {
		t.Errorf("modType = %q, want %q", modType, ticker.ModifierRange)
	}
}

func TestSymbolExtractor_TickerStandardCall(t *testing.T) {
	extractor := NewSymbolExtractor()

	expr := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ticker"},
			Property: &ast.Identifier{Name: "standard"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "HEIKINASHI:BTCUSDT"},
		},
	}

	symbol, modType := extractor.Extract(expr)

	if symbol != "BTCUSDT" {
		t.Errorf("symbol = %q, want BTCUSDT", symbol)
	}

	if modType != "" {
		t.Errorf("modType = %q, want empty (standard extracts base symbol)", modType)
	}
}

func TestSymbolExtractor_ModifiedSymbolString(t *testing.T) {
	extractor := NewSymbolExtractor()

	expr := &ast.Literal{Value: "HEIKINASHI:ETHUSDT"}
	symbol, modType := extractor.Extract(expr)

	if symbol != "ETHUSDT" {
		t.Errorf("symbol = %q, want ETHUSDT", symbol)
	}

	if modType != ticker.ModifierHeikinAshi {
		t.Errorf("modType = %q, want %q", modType, ticker.ModifierHeikinAshi)
	}
}

func TestSymbolExtractor_InsufficientArguments(t *testing.T) {
	extractor := NewSymbolExtractor()

	tests := []struct {
		name string
		expr *ast.CallExpression
	}{
		{
			name: "heikinashi with no args",
			expr: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "heikinashi"},
				Arguments: []ast.Expression{},
			},
		},
		{
			name: "renko with one arg",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "renko"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
				},
			},
		},
		{
			name: "kagi with no args",
			expr: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "kagi"},
				Arguments: []ast.Expression{},
			},
		},
		{
			name: "range with no args",
			expr: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "range"},
				Arguments: []ast.Expression{},
			},
		},
		{
			name: "ticker.range with no args",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ticker"},
					Property: &ast.Identifier{Name: "range"},
				},
				Arguments: []ast.Expression{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symbol, _ := extractor.Extract(tt.expr)
			if symbol != "" {
				t.Errorf("Expected empty symbol for insufficient args, got %q", symbol)
			}
		})
	}
}

func TestSymbolExtractor_NonTickerCall(t *testing.T) {
	extractor := NewSymbolExtractor()

	expr := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "ta.sma"},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 20.0},
		},
	}

	symbol, modType := extractor.Extract(expr)

	if symbol != "" {
		t.Errorf("symbol = %q, want empty for non-ticker call", symbol)
	}

	if modType != "" {
		t.Errorf("modType = %q, want empty for non-ticker call", modType)
	}
}
