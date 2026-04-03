package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestTACallResolver_ATR_SingleArgument(t *testing.T) {
	registry := NewTASignatureRegistry()
	resolver := NewTACallResolver(registry)

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Literal{Value: "14"},
		},
	}

	resolved, err := resolver.Resolve("ta.atr", call)
	if err != nil {
		t.Fatalf("Failed to resolve ta.atr(14): %v", err)
	}

	if !resolved.RequiresOHLC {
		t.Error("ta.atr should require implicit OHLC")
	}

	if len(resolved.SeriesArguments) != 0 {
		t.Errorf("ta.atr(14) should have 0 series arguments, got %d", len(resolved.SeriesArguments))
	}

	if len(resolved.ScalarArguments) != 1 {
		t.Errorf("ta.atr(14) should have 1 scalar argument, got %d", len(resolved.ScalarArguments))
	}
}

func TestTACallResolver_Pivothigh_TwoArguments(t *testing.T) {
	registry := NewTASignatureRegistry()
	resolver := NewTACallResolver(registry)

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Literal{Value: "5"},
			&ast.Literal{Value: "5"},
		},
	}

	resolved, err := resolver.Resolve("ta.pivothigh", call)
	if err != nil {
		t.Fatalf("Failed to resolve ta.pivothigh(5, 5): %v", err)
	}

	if !resolved.DefaultSourceApplied {
		t.Error("ta.pivothigh(5, 5) should apply default source 'high'")
	}

	if resolved.DefaultSourceName != "high" {
		t.Errorf("ta.pivothigh default source = %s, want 'high'", resolved.DefaultSourceName)
	}

	if len(resolved.ScalarArguments) != 2 {
		t.Errorf("ta.pivothigh(5, 5) should have 2 scalar arguments, got %d", len(resolved.ScalarArguments))
	}
}

func TestTACallResolver_Pivothigh_ThreeArguments(t *testing.T) {
	registry := NewTASignatureRegistry()
	resolver := NewTACallResolver(registry)

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: "5"},
			&ast.Literal{Value: "5"},
		},
	}

	resolved, err := resolver.Resolve("ta.pivothigh", call)
	if err != nil {
		t.Fatalf("Failed to resolve ta.pivothigh(close, 5, 5): %v", err)
	}

	if resolved.DefaultSourceApplied {
		t.Error("ta.pivothigh(close, 5, 5) should not apply default source")
	}

	if len(resolved.SeriesArguments) != 1 {
		t.Errorf("ta.pivothigh(close, 5, 5) should have 1 series argument, got %d", len(resolved.SeriesArguments))
	}

	if len(resolved.ScalarArguments) != 2 {
		t.Errorf("ta.pivothigh(close, 5, 5) should have 2 scalar arguments, got %d", len(resolved.ScalarArguments))
	}
}

func TestTACallResolver_SMA_SingleArgument(t *testing.T) {
	registry := NewTASignatureRegistry()
	resolver := NewTACallResolver(registry)

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Literal{Value: "14"},
		},
	}

	resolved, err := resolver.Resolve("ta.sma", call)
	if err != nil {
		t.Fatalf("Failed to resolve ta.sma(14): %v", err)
	}

	if !resolved.DefaultSourceApplied {
		t.Error("ta.sma(14) should apply default source 'close'")
	}

	if resolved.DefaultSourceName != "close" {
		t.Errorf("ta.sma default source = %s, want 'close'", resolved.DefaultSourceName)
	}
}

func TestTACallResolver_SMA_TwoArguments(t *testing.T) {
	registry := NewTASignatureRegistry()
	resolver := NewTACallResolver(registry)

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "open"},
			&ast.Literal{Value: "14"},
		},
	}

	resolved, err := resolver.Resolve("ta.sma", call)
	if err != nil {
		t.Fatalf("Failed to resolve ta.sma(open, 14): %v", err)
	}

	if resolved.DefaultSourceApplied {
		t.Error("ta.sma(open, 14) should not apply default source")
	}

	if len(resolved.SeriesArguments) != 1 {
		t.Errorf("ta.sma(open, 14) should have 1 series argument, got %d", len(resolved.SeriesArguments))
	}

	if len(resolved.ScalarArguments) != 1 {
		t.Errorf("ta.sma(open, 14) should have 1 scalar argument, got %d", len(resolved.ScalarArguments))
	}
}

func TestTACallResolver_UnknownFunction_FallbackPattern(t *testing.T) {
	registry := NewTASignatureRegistry()
	resolver := NewTACallResolver(registry)

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: "20"},
		},
	}

	resolved, err := resolver.Resolve("ta.unknown", call)
	if err != nil {
		t.Fatalf("Failed to resolve unknown function with 2 args: %v", err)
	}

	if len(resolved.SeriesArguments) != 1 {
		t.Errorf("Fallback pattern should have 1 series argument, got %d", len(resolved.SeriesArguments))
	}

	if len(resolved.ScalarArguments) != 1 {
		t.Errorf("Fallback pattern should have 1 scalar argument, got %d", len(resolved.ScalarArguments))
	}
}

func TestTACallResolver_UnknownFunction_InvalidArgCount(t *testing.T) {
	registry := NewTASignatureRegistry()
	resolver := NewTACallResolver(registry)

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Literal{Value: "14"},
		},
	}

	_, err := resolver.Resolve("ta.unknown", call)
	if err == nil {
		t.Error("Unknown function with 1 argument should fail")
	}
}

func TestTACallResolver_Change_SingleArgument(t *testing.T) {
	registry := NewTASignatureRegistry()
	resolver := NewTACallResolver(registry)

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
		},
	}

	resolved, err := resolver.Resolve("ta.change", call)
	if err != nil {
		t.Fatalf("Failed to resolve ta.change(close): %v", err)
	}

	if len(resolved.SeriesArguments) != 1 {
		t.Errorf("ta.change(close) should have 1 series argument, got %d", len(resolved.SeriesArguments))
	}
}

func TestTACallResolver_Change_TwoArguments(t *testing.T) {
	registry := NewTASignatureRegistry()
	resolver := NewTACallResolver(registry)

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: "2"},
		},
	}

	resolved, err := resolver.Resolve("ta.change", call)
	if err != nil {
		t.Fatalf("Failed to resolve ta.change(close, 2): %v", err)
	}

	if len(resolved.SeriesArguments) != 1 {
		t.Errorf("ta.change(close, 2) should have 1 series argument, got %d", len(resolved.SeriesArguments))
	}

	if len(resolved.ScalarArguments) != 1 {
		t.Errorf("ta.change(close, 2) should have 1 scalar argument, got %d", len(resolved.ScalarArguments))
	}
}

/* TestTACallResolver_SourceDetectionWithExtendedOverloads exercises the
 * HasOverloadWithSourceAt fix: when a function has overloads up to N args
 * but a K-arg call (K < N) matches an overload whose first arg is a series,
 * the resolver must NOT inject a default source.
 *
 * This set covers every combination of (has-source | no-source) × (at max | below max)
 * for functions whose max arg count exceeds the provided arg count.
 */
func TestTACallResolver_SourceDetectionWithExtendedOverloads(t *testing.T) {
	registry := NewTASignatureRegistry()
	resolver := NewTACallResolver(registry)

	tests := []struct {
		name               string
		funcName           string
		args               []ast.Expression
		wantDefaultApplied bool
		wantSeriesCount    int
		wantScalarCount    int
	}{
		// ta.kcw: max=4, provided=2, no source → default applied
		{
			name:     "kcw_2args_no_source_gets_default",
			funcName: "ta.kcw",
			args: []ast.Expression{
				&ast.Literal{Value: 20.0},
				&ast.Literal{Value: 1.5},
			},
			wantDefaultApplied: true,
			wantSeriesCount:    1,
			wantScalarCount:    2,
		},
		// ta.kcw: max=4, provided=3, first arg is series → no default
		{
			name:     "kcw_3args_with_source_no_default",
			funcName: "ta.kcw",
			args: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 20.0},
				&ast.Literal{Value: 1.5},
			},
			wantDefaultApplied: false,
			wantSeriesCount:    1,
			wantScalarCount:    2,
		},
		// ta.kcw: max=4, provided=4, first arg is series → no default
		{
			name:     "kcw_4args_with_source_no_default",
			funcName: "ta.kcw",
			args: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 20.0},
				&ast.Literal{Value: 1.5},
				&ast.Literal{Value: true},
			},
			wantDefaultApplied: false,
			wantSeriesCount:    1,
			wantScalarCount:    3,
		},
		// ta.kc: same overload structure, same expectations
		{
			name:     "kc_2args_no_source_gets_default",
			funcName: "ta.kc",
			args: []ast.Expression{
				&ast.Literal{Value: 20.0},
				&ast.Literal{Value: 1.5},
			},
			wantDefaultApplied: true,
			wantSeriesCount:    1,
			wantScalarCount:    2,
		},
		{
			name:     "kc_3args_with_source_no_default",
			funcName: "ta.kc",
			args: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 20.0},
				&ast.Literal{Value: 1.5},
			},
			wantDefaultApplied: false,
			wantSeriesCount:    1,
			wantScalarCount:    2,
		},
		{
			name:     "kc_4args_with_source_no_default",
			funcName: "ta.kc",
			args: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 20.0},
				&ast.Literal{Value: 1.5},
				&ast.Literal{Value: false},
			},
			wantDefaultApplied: false,
			wantSeriesCount:    1,
			wantScalarCount:    3,
		},
		// ta.sma: max=2, provided=1, no source → default applied (baseline)
		{
			name:     "sma_1arg_baseline_default_applied",
			funcName: "ta.sma",
			args: []ast.Expression{
				&ast.Literal{Value: 14.0},
			},
			wantDefaultApplied: true,
			wantSeriesCount:    1,
			wantScalarCount:    1,
		},
		// ta.sma: max=2, provided=2, source explicit → no default
		{
			name:     "sma_2args_baseline_no_default",
			funcName: "ta.sma",
			args: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: 14.0},
			},
			wantDefaultApplied: false,
			wantSeriesCount:    1,
			wantScalarCount:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{Arguments: tt.args}
			resolved, err := resolver.Resolve(tt.funcName, call)
			if err != nil {
				t.Fatalf("Resolve(%s, %d args) error = %v", tt.funcName, len(tt.args), err)
			}
			if resolved.DefaultSourceApplied != tt.wantDefaultApplied {
				t.Errorf("DefaultSourceApplied = %v, want %v", resolved.DefaultSourceApplied, tt.wantDefaultApplied)
			}
			if len(resolved.SeriesArguments) != tt.wantSeriesCount {
				t.Errorf("SeriesArguments = %d, want %d", len(resolved.SeriesArguments), tt.wantSeriesCount)
			}
			if len(resolved.ScalarArguments) != tt.wantScalarCount {
				t.Errorf("ScalarArguments = %d, want %d", len(resolved.ScalarArguments), tt.wantScalarCount)
			}
		})
	}
}
