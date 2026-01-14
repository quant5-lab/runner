package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/validation"
)

/* TestTAArgumentExtractor_Extract_IdentifierSources tests classification of simple identifiers */
func TestTAArgumentExtractor_Extract_IdentifierSources(t *testing.T) {
	analyzer := validation.NewWarmupAnalyzer()
	g := &generator{
		variables:      make(map[string]string),
		constants:      make(map[string]interface{}),
		constEvaluator: analyzer,
	}

	extractor := NewTAArgumentExtractor(g)

	tests := []struct {
		name           string
		sourceExpr     ast.Expression
		period         int
		wantSourceType SourceType
		wantFieldName  string
		wantVarName    string
		wantNeedsNaN   bool
	}{
		{
			name:           "close field",
			sourceExpr:     &ast.Identifier{Name: "close"},
			period:         20,
			wantSourceType: SourceTypeOHLCVField,
			wantFieldName:  "Close",
			wantNeedsNaN:   false,
		},
		{
			name:           "open field",
			sourceExpr:     &ast.Identifier{Name: "open"},
			period:         50,
			wantSourceType: SourceTypeOHLCVField,
			wantFieldName:  "Open",
			wantNeedsNaN:   false,
		},
		{
			name:           "high field",
			sourceExpr:     &ast.Identifier{Name: "high"},
			period:         100,
			wantSourceType: SourceTypeOHLCVField,
			wantFieldName:  "High",
			wantNeedsNaN:   false,
		},
		{
			name:           "low field",
			sourceExpr:     &ast.Identifier{Name: "low"},
			period:         10,
			wantSourceType: SourceTypeOHLCVField,
			wantFieldName:  "Low",
			wantNeedsNaN:   false,
		},
		{
			name:           "volume field",
			sourceExpr:     &ast.Identifier{Name: "volume"},
			period:         14,
			wantSourceType: SourceTypeOHLCVField,
			wantFieldName:  "Volume",
			wantNeedsNaN:   false,
		},
		{
			name:           "user series variable",
			sourceExpr:     &ast.Identifier{Name: "myValue"},
			period:         30,
			wantSourceType: SourceTypeSeriesVariable,
			wantVarName:    "myValue",
			wantNeedsNaN:   true,
		},
		{
			name:           "temp variable with hash",
			sourceExpr:     &ast.Identifier{Name: "ta_sma_50_abc123"},
			period:         20,
			wantSourceType: SourceTypeSeriesVariable,
			wantVarName:    "ta_sma_50_abc123",
			wantNeedsNaN:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					tt.sourceExpr,
					&ast.Literal{Value: tt.period},
				},
			}

			comp, err := extractor.Extract(call, "ta.sma")
			if err != nil {
				t.Fatalf("Extract() error = %v", err)
			}

			if comp.Period != tt.period {
				t.Errorf("Period = %d, want %d", comp.Period, tt.period)
			}

			if comp.SourceInfo.Type != tt.wantSourceType {
				t.Errorf("SourceInfo.Type = %v, want %v", comp.SourceInfo.Type, tt.wantSourceType)
			}

			if tt.wantSourceType == SourceTypeOHLCVField {
				if comp.SourceInfo.FieldName != tt.wantFieldName {
					t.Errorf("SourceInfo.FieldName = %s, want %s", comp.SourceInfo.FieldName, tt.wantFieldName)
				}
				if !comp.SourceInfo.IsOHLCVField() {
					t.Error("IsOHLCVField() = false, want true")
				}
			}

			if tt.wantSourceType == SourceTypeSeriesVariable {
				if comp.SourceInfo.VariableName != tt.wantVarName {
					t.Errorf("SourceInfo.VariableName = %s, want %s", comp.SourceInfo.VariableName, tt.wantVarName)
				}
				if !comp.SourceInfo.IsSeriesVariable() {
					t.Error("IsSeriesVariable() = false, want true")
				}
			}

			if comp.NeedsNaNCheck != tt.wantNeedsNaN {
				t.Errorf("NeedsNaNCheck = %v, want %v", comp.NeedsNaNCheck, tt.wantNeedsNaN)
			}

			if comp.AccessGen == nil {
				t.Fatal("AccessGen is nil")
			}
		})
	}
}

/* TestTAArgumentExtractor_Extract_MemberExpressions tests subscripted historical access */
func TestTAArgumentExtractor_Extract_MemberExpressions(t *testing.T) {
	analyzer := validation.NewWarmupAnalyzer()
	g := &generator{
		variables:      make(map[string]string),
		constants:      make(map[string]interface{}),
		constEvaluator: analyzer,
	}

	extractor := NewTAArgumentExtractor(g)

	tests := []struct {
		name           string
		sourceExpr     ast.Expression
		period         int
		wantSourceType SourceType
		wantFieldName  string
		wantVarName    string
	}{
		{
			name: "close[1] - single bar lookback",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 1},
				Computed: true,
			},
			period:         20,
			wantSourceType: SourceTypeOHLCVField,
			wantFieldName:  "Close",
		},
		{
			name: "close[4] - multi bar lookback",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 4},
				Computed: true,
			},
			period:         200,
			wantSourceType: SourceTypeOHLCVField,
			wantFieldName:  "Close",
		},
		{
			name: "high[10]",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "high"},
				Property: &ast.Literal{Value: 10},
				Computed: true,
			},
			period:         50,
			wantSourceType: SourceTypeOHLCVField,
			wantFieldName:  "High",
		},
		{
			name: "low[5]",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "low"},
				Property: &ast.Literal{Value: 5},
				Computed: true,
			},
			period:         14,
			wantSourceType: SourceTypeOHLCVField,
			wantFieldName:  "Low",
		},
		{
			name: "volume[0] - current bar",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "volume"},
				Property: &ast.Literal{Value: 0},
				Computed: true,
			},
			period:         21,
			wantSourceType: SourceTypeOHLCVField,
			wantFieldName:  "Volume",
		},
		{
			name: "userVar[1] - series variable subscript",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "sma20"},
				Property: &ast.Literal{Value: 1},
				Computed: true,
			},
			period:         10,
			wantSourceType: SourceTypeSeriesVariable,
			wantVarName:    "sma20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					tt.sourceExpr,
					&ast.Literal{Value: tt.period},
				},
			}

			comp, err := extractor.Extract(call, "ta.sma")
			if err != nil {
				t.Fatalf("Extract() error = %v", err)
			}

			if comp.Period != tt.period {
				t.Errorf("Period = %d, want %d", comp.Period, tt.period)
			}

			if comp.SourceInfo.Type != tt.wantSourceType {
				t.Errorf("SourceInfo.Type = %v, want %v", comp.SourceInfo.Type, tt.wantSourceType)
			}

			if tt.wantSourceType == SourceTypeOHLCVField {
				if comp.SourceInfo.FieldName != tt.wantFieldName {
					t.Errorf("SourceInfo.FieldName = %s, want %s", comp.SourceInfo.FieldName, tt.wantFieldName)
				}
			}

			if tt.wantSourceType == SourceTypeSeriesVariable {
				if comp.SourceInfo.VariableName != tt.wantVarName {
					t.Errorf("SourceInfo.VariableName = %s, want %s", comp.SourceInfo.VariableName, tt.wantVarName)
				}
			}
		})
	}
}

/* TestTAArgumentExtractor_Extract_PeriodVariations tests period extraction from various sources */
func TestTAArgumentExtractor_Extract_PeriodVariations(t *testing.T) {
	tests := []struct {
		name       string
		periodExpr ast.Expression
		constants  map[string]interface{}
		wantPeriod int
		wantError  bool
	}{
		{
			name:       "integer literal",
			periodExpr: &ast.Literal{Value: 20},
			wantPeriod: 20,
		},
		{
			name:       "float literal",
			periodExpr: &ast.Literal{Value: 50.0},
			wantPeriod: 50,
		},
		{
			name:       "small period",
			periodExpr: &ast.Literal{Value: 2},
			wantPeriod: 2,
		},
		{
			name:       "large period",
			periodExpr: &ast.Literal{Value: 500},
			wantPeriod: 500,
		},
		{
			name:       "constant variable",
			periodExpr: &ast.Identifier{Name: "length"},
			constants:  map[string]interface{}{"length": 50},
			wantPeriod: 50,
		},
		{
			name:       "string literal - invalid",
			periodExpr: &ast.Literal{Value: "invalid"},
			wantError:  true,
		},
		{
			name:       "undefined constant",
			periodExpr: &ast.Identifier{Name: "undefined"},
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := validation.NewWarmupAnalyzer()
			for k, v := range tt.constants {
				if fv, ok := v.(int); ok {
					analyzer.AddConstant(k, float64(fv))
				} else if fv, ok := v.(float64); ok {
					analyzer.AddConstant(k, fv)
				}
			}

			g := &generator{
				variables:      make(map[string]string),
				constants:      tt.constants,
				constEvaluator: analyzer,
			}

			extractor := NewTAArgumentExtractor(g)

			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					tt.periodExpr,
				},
			}

			comp, err := extractor.Extract(call, "ta.sma")

			if tt.wantError {
				if err == nil {
					t.Error("Extract() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Fatalf("Extract() unexpected error = %v", err)
			}

			if comp.Period != tt.wantPeriod {
				t.Errorf("Period = %d, want %d", comp.Period, tt.wantPeriod)
			}
		})
	}
}

/* TestTAArgumentExtractor_Extract_ValidationErrors tests error conditions */
func TestTAArgumentExtractor_Extract_ValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		call      *ast.CallExpression
		funcName  string
		wantError string
	}{
		{
			name: "no arguments",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{},
			},
			funcName:  "ta.sma",
			wantError: "requires at least 2 arguments",
		},
		{
			name: "single argument",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
				},
			},
			funcName:  "ta.ema",
			wantError: "requires at least 2 arguments",
		},
		{
			name: "invalid period type",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: "not-a-number"},
				},
			},
			funcName:  "ta.stdev",
			wantError: "period must be numeric",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := validation.NewWarmupAnalyzer()
			g := &generator{
				variables:      make(map[string]string),
				constants:      make(map[string]interface{}),
				constEvaluator: analyzer,
			}

			extractor := NewTAArgumentExtractor(g)

			_, err := extractor.Extract(tt.call, tt.funcName)
			if err == nil {
				t.Fatal("Extract() error = nil, want error")
			}

			if tt.wantError != "" {
				errMsg := err.Error()
				found := false
				for i := 0; i <= len(errMsg)-len(tt.wantError); i++ {
					if errMsg[i:i+len(tt.wantError)] == tt.wantError {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Extract() error = %q, want substring %q", errMsg, tt.wantError)
				}
			}
		})
	}
}

/* TestTAArgumentExtractor_Extract_AccessGeneratorTypes validates correct generator creation */
func TestTAArgumentExtractor_Extract_AccessGeneratorTypes(t *testing.T) {
	analyzer := validation.NewWarmupAnalyzer()
	g := &generator{
		variables:      make(map[string]string),
		constants:      make(map[string]interface{}),
		constEvaluator: analyzer,
	}

	extractor := NewTAArgumentExtractor(g)

	tests := []struct {
		name       string
		sourceExpr ast.Expression
	}{
		{
			name:       "OHLCV field",
			sourceExpr: &ast.Identifier{Name: "close"},
		},
		{
			name: "subscripted OHLCV",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "high"},
				Property: &ast.Literal{Value: 5},
				Computed: true,
			},
		},
		{
			name:       "series variable",
			sourceExpr: &ast.Identifier{Name: "myVar"},
		},
		{
			name: "subscripted series",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "sma20"},
				Property: &ast.Literal{Value: 1},
				Computed: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					tt.sourceExpr,
					&ast.Literal{Value: 20},
				},
			}

			comp, err := extractor.Extract(call, "ta.sma")
			if err != nil {
				t.Fatalf("Extract() error = %v", err)
			}

			if comp.AccessGen == nil {
				t.Fatal("AccessGen is nil")
			}

			loopAccess := comp.AccessGen.GenerateLoopValueAccess("j")
			if loopAccess == "" {
				t.Error("GenerateLoopValueAccess returned empty string")
			}

			initialAccess := comp.AccessGen.GenerateInitialValueAccess(20)
			if initialAccess == "" {
				t.Error("GenerateInitialValueAccess returned empty string")
			}
		})
	}
}

/* TestTAArgumentExtractor_Integration tests full workflow with multiple indicators */
func TestTAArgumentExtractor_Integration(t *testing.T) {
	analyzer := validation.NewWarmupAnalyzer()
	g := &generator{
		variables:      make(map[string]string),
		constants:      make(map[string]interface{}),
		constEvaluator: analyzer,
	}

	extractor := NewTAArgumentExtractor(g)

	indicators := []struct {
		funcName   string
		sourceExpr ast.Expression
		period     int
	}{
		{"ta.sma", &ast.Identifier{Name: "close"}, 20},
		{"ta.ema", &ast.Identifier{Name: "close"}, 50},
		{"ta.rma", &ast.Identifier{Name: "close"}, 14},
		{"ta.wma", &ast.Identifier{Name: "high"}, 10},
		{"ta.stdev", &ast.Identifier{Name: "close"}, 20},
	}

	for _, ind := range indicators {
		t.Run(ind.funcName, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					ind.sourceExpr,
					&ast.Literal{Value: ind.period},
				},
			}

			comp, err := extractor.Extract(call, ind.funcName)
			if err != nil {
				t.Fatalf("Extract(%s) error = %v", ind.funcName, err)
			}

			if comp.Period != ind.period {
				t.Errorf("%s: Period = %d, want %d", ind.funcName, comp.Period, ind.period)
			}

			if comp.AccessGen == nil {
				t.Errorf("%s: AccessGen is nil", ind.funcName)
			}

			if comp.SourceInfo.Type == SourceTypeUnknown {
				t.Errorf("%s: SourceInfo.Type is Unknown", ind.funcName)
			}
		})
	}
}
