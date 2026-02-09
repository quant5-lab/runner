package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestTA_ImplicitOHLCPattern(t *testing.T) {
	resolver := NewArrowTACallSignatureResolver()

	tests := []struct {
		name     string
		function string
		args     []ast.Expression
		wantOHLC bool
	}{
		{
			name:     "ta.atr single scalar argument",
			function: "ta.atr",
			args: []ast.Expression{
				&ast.Literal{Value: "14"},
			},
			wantOHLC: true,
		},
		{
			name:     "ta.tr no arguments",
			function: "ta.tr",
			args:     []ast.Expression{},
			wantOHLC: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{Arguments: tt.args}
			resolved, err := resolver.ResolveCall(tt.function, call)

			if err != nil {
				t.Fatalf("%s failed: %v", tt.name, err)
			}
			if resolved == nil {
				t.Fatalf("%s returned nil", tt.name)
			}

			if tt.wantOHLC {
				if resolved.SourceExpr != nil {
					t.Errorf("%s should have nil SourceExpr for implicit OHLC, got %T",
						tt.name, resolved.SourceExpr)
				}
			}
		})
	}
}

func TestTA_MultiArgumentPatterns(t *testing.T) {
	resolver := NewArrowTACallSignatureResolver()

	tests := []struct {
		name              string
		function          string
		args              []ast.Expression
		wantSourceNil     bool
		wantLengthNil     bool
		wantDefaultSource bool
		defaultSourceName string
	}{
		{
			name:     "ta.pivothigh three args explicit source",
			function: "ta.pivothigh",
			args: []ast.Expression{
				&ast.Identifier{Name: "source"},
				&ast.Literal{Value: "5"},
				&ast.Literal{Value: "5"},
			},
			wantSourceNil:     false,
			wantLengthNil:     false,
			wantDefaultSource: false,
		},
		{
			name:     "ta.pivothigh two args default source",
			function: "ta.pivothigh",
			args: []ast.Expression{
				&ast.Literal{Value: "5"},
				&ast.Literal{Value: "5"},
			},
			wantSourceNil:     true,
			wantLengthNil:     false,
			wantDefaultSource: true,
			defaultSourceName: "high",
		},
		{
			name:     "ta.pivotlow two args default source",
			function: "ta.pivotlow",
			args: []ast.Expression{
				&ast.Literal{Value: "3"},
				&ast.Literal{Value: "3"},
			},
			wantSourceNil:     true,
			wantLengthNil:     false,
			wantDefaultSource: true,
			defaultSourceName: "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{Arguments: tt.args}
			resolved, err := resolver.ResolveCall(tt.function, call)

			if err != nil {
				t.Fatalf("%s failed: %v", tt.name, err)
			}
			if resolved == nil {
				t.Fatalf("%s returned nil", tt.name)
			}

			if tt.wantSourceNil && resolved.SourceExpr != nil {
				t.Errorf("%s should have nil SourceExpr, got %T", tt.name, resolved.SourceExpr)
			}
			if !tt.wantSourceNil && resolved.SourceExpr == nil {
				t.Errorf("%s should have non-nil SourceExpr", tt.name)
			}

			if tt.wantLengthNil && resolved.LengthExpr != nil {
				t.Errorf("%s should have nil LengthExpr, got %T", tt.name, resolved.LengthExpr)
			}
			if !tt.wantLengthNil && resolved.LengthExpr == nil {
				t.Errorf("%s should have non-nil LengthExpr", tt.name)
			}

			if tt.wantDefaultSource != resolved.NeedsDefaultSource {
				t.Errorf("%s NeedsDefaultSource = %v, want %v",
					tt.name, resolved.NeedsDefaultSource, tt.wantDefaultSource)
			}

			if tt.defaultSourceName != "" && resolved.DefaultSourceName != tt.defaultSourceName {
				t.Errorf("%s DefaultSourceName = %q, want %q",
					tt.name, resolved.DefaultSourceName, tt.defaultSourceName)
			}
		})
	}
}
