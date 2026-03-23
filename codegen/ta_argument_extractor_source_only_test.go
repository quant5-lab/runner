package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/validation"
)

func newExtractSourceOnlyExtractor() *TAArgumentExtractor {
	analyzer := validation.NewWarmupAnalyzer()
	g := &generator{
		variables:      make(map[string]string),
		constants:      make(map[string]interface{}),
		constEvaluator: analyzer,
	}
	return NewTAArgumentExtractor(g)
}

func TestTAArgumentExtractor_ExtractSourceOnly_OHLCVIdentifiers(t *testing.T) {
	extractor := newExtractSourceOnlyExtractor()

	fields := []struct {
		name      string
		fieldName string
	}{
		{"close", "Close"},
		{"open", "Open"},
		{"high", "High"},
		{"low", "Low"},
		{"volume", "Volume"},
	}

	for _, tt := range fields {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: tt.name},
				},
			}

			comp, err := extractor.ExtractSourceOnly(call, "ta.swma")
			if err != nil {
				t.Fatalf("ExtractSourceOnly() error = %v", err)
			}

			if comp.Period != 0 {
				t.Errorf("Period = %d, want 0 (no period for source-only)", comp.Period)
			}

			if comp.SourceInfo.FieldName != tt.fieldName {
				t.Errorf("FieldName = %s, want %s", comp.SourceInfo.FieldName, tt.fieldName)
			}

			if !comp.SourceInfo.IsOHLCVField() {
				t.Error("IsOHLCVField() = false, want true")
			}

			if comp.NeedsNaNCheck {
				t.Error("NeedsNaNCheck = true for OHLCV field, want false")
			}

			if comp.AccessGen == nil {
				t.Fatal("AccessGen is nil")
			}
		})
	}
}

func TestTAArgumentExtractor_ExtractSourceOnly_SeriesVariable(t *testing.T) {
	extractor := newExtractSourceOnlyExtractor()

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "myCustomSeries"},
		},
	}

	comp, err := extractor.ExtractSourceOnly(call, "ta.swma")
	if err != nil {
		t.Fatalf("ExtractSourceOnly() error = %v", err)
	}

	if comp.Period != 0 {
		t.Errorf("Period = %d, want 0", comp.Period)
	}

	if !comp.SourceInfo.IsSeriesVariable() {
		t.Error("IsSeriesVariable() = false, want true")
	}

	if comp.SourceInfo.VariableName != "myCustomSeries" {
		t.Errorf("VariableName = %s, want myCustomSeries", comp.SourceInfo.VariableName)
	}

	if !comp.NeedsNaNCheck {
		t.Error("NeedsNaNCheck = false for series variable, want true")
	}
}

func TestTAArgumentExtractor_ExtractSourceOnly_NoArguments(t *testing.T) {
	extractor := newExtractSourceOnlyExtractor()

	call := &ast.CallExpression{Arguments: []ast.Expression{}}

	_, err := extractor.ExtractSourceOnly(call, "ta.swma")
	if err == nil {
		t.Fatal("Expected error for zero arguments, got nil")
	}
}

func TestTAArgumentExtractor_ExtractSourceOnly_HLCComposites(t *testing.T) {
	extractor := newExtractSourceOnlyExtractor()

	composites := []string{"hlc3", "hl2", "ohlc4", "hlcc4"}

	for _, name := range composites {
		t.Run(name, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: name},
				},
			}

			comp, err := extractor.ExtractSourceOnly(call, "ta.swma")
			if err != nil {
				t.Fatalf("ExtractSourceOnly(%s) error = %v", name, err)
			}

			if comp.Period != 0 {
				t.Errorf("Period = %d, want 0", comp.Period)
			}

			if comp.AccessGen == nil {
				t.Fatalf("AccessGen is nil for %s", name)
			}
		})
	}
}

func TestTAArgumentExtractor_ExtractSourceOnly_TrBuiltin(t *testing.T) {
	tests := []struct {
		name       string
		sourceExpr ast.Expression
	}{
		{
			name:       "bare tr identifier",
			sourceExpr: &ast.Identifier{Name: "tr"},
		},
		{
			name: "ta.tr member expression",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "tr"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := newExtractSourceOnlyExtractor()
			call := &ast.CallExpression{
				Arguments: []ast.Expression{tt.sourceExpr},
			}

			comp, err := extractor.ExtractSourceOnly(call, "ta.swma")
			if err != nil {
				t.Fatalf("ExtractSourceOnly() error = %v", err)
			}

			if comp.Period != 0 {
				t.Errorf("Period = %d, want 0", comp.Period)
			}
			if _, ok := comp.AccessGen.(*TrueRangeAccessGenerator); !ok {
				t.Errorf("AccessGen = %T, want *TrueRangeAccessGenerator", comp.AccessGen)
			}
			if !comp.NeedsNaNCheck {
				t.Error("NeedsNaNCheck = false, want true: ta.tr needs NaN check")
			}
		})
	}
}

func TestTAArgumentExtractor_ExtractSourceOnly_ExtraArgsIgnored(t *testing.T) {
	/* Extra arguments beyond the source must be silently ignored (like SWMA's fixed period) */
	extractor := newExtractSourceOnlyExtractor()

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 4.0},
		},
	}

	comp, err := extractor.ExtractSourceOnly(call, "ta.swma")
	if err != nil {
		t.Fatalf("ExtractSourceOnly() error = %v with extra arg: %v", err, err)
	}

	if comp.SourceInfo.FieldName != "Close" {
		t.Errorf("FieldName = %s, want Close", comp.SourceInfo.FieldName)
	}
}
