package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestSeriesSourceClassifier_ClassifyAST_Identifiers tests classification of identifier nodes */
func TestSeriesSourceClassifier_ClassifyAST_Identifiers(t *testing.T) {
	classifier := NewSeriesSourceClassifier()

	tests := []struct {
		name           string
		expr           ast.Expression
		wantType       SourceType
		wantFieldName  string
		wantVarName    string
		wantBaseOffset int
	}{
		{
			name:           "close identifier",
			expr:           &ast.Identifier{Name: "close"},
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "Close",
			wantBaseOffset: 0,
		},
		{
			name:           "open identifier",
			expr:           &ast.Identifier{Name: "open"},
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "Open",
			wantBaseOffset: 0,
		},
		{
			name:           "high identifier",
			expr:           &ast.Identifier{Name: "high"},
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "High",
			wantBaseOffset: 0,
		},
		{
			name:           "low identifier",
			expr:           &ast.Identifier{Name: "low"},
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "Low",
			wantBaseOffset: 0,
		},
		{
			name:           "volume identifier",
			expr:           &ast.Identifier{Name: "volume"},
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "Volume",
			wantBaseOffset: 0,
		},
		{
			name:           "user variable identifier",
			expr:           &ast.Identifier{Name: "myValue"},
			wantType:       SourceTypeSeriesVariable,
			wantVarName:    "myValue",
			wantBaseOffset: 0,
		},
		{
			name:           "temp variable identifier",
			expr:           &ast.Identifier{Name: "ta_sma_20_abc123"},
			wantType:       SourceTypeSeriesVariable,
			wantVarName:    "ta_sma_20_abc123",
			wantBaseOffset: 0,
		},
		{
			name:           "underscore variable",
			expr:           &ast.Identifier{Name: "my_var"},
			wantType:       SourceTypeSeriesVariable,
			wantVarName:    "my_var",
			wantBaseOffset: 0,
		},
		{
			name:           "empty identifier",
			expr:           &ast.Identifier{Name: ""},
			wantType:       SourceTypeSeriesVariable,
			wantVarName:    "",
			wantBaseOffset: 0,
		},
		{
			name:           "case sensitivity - Close uppercase",
			expr:           &ast.Identifier{Name: "Close"},
			wantType:       SourceTypeSeriesVariable,
			wantVarName:    "Close",
			wantBaseOffset: 0,
		},
		{
			name:           "mixed case ohlcv (CLOSE not recognized)",
			expr:           &ast.Identifier{Name: "CLOSE"},
			wantType:       SourceTypeSeriesVariable,
			wantVarName:    "CLOSE",
			wantBaseOffset: 0,
		},
		{
			name:           "hl2 derived price",
			expr:           &ast.Identifier{Name: "hl2"},
			wantType:       SourceTypeDerivedPrice,
			wantFieldName:  "hl2",
			wantBaseOffset: 0,
		},
		{
			name:           "hlc3 derived price",
			expr:           &ast.Identifier{Name: "hlc3"},
			wantType:       SourceTypeDerivedPrice,
			wantFieldName:  "hlc3",
			wantBaseOffset: 0,
		},
		{
			name:           "ohlc4 derived price",
			expr:           &ast.Identifier{Name: "ohlc4"},
			wantType:       SourceTypeDerivedPrice,
			wantFieldName:  "ohlc4",
			wantBaseOffset: 0,
		},
		{
			name:           "hlcc4 derived price",
			expr:           &ast.Identifier{Name: "hlcc4"},
			wantType:       SourceTypeDerivedPrice,
			wantFieldName:  "hlcc4",
			wantBaseOffset: 0,
		},
		{
			name:           "HL2 uppercase not derived price",
			expr:           &ast.Identifier{Name: "HL2"},
			wantType:       SourceTypeSeriesVariable,
			wantVarName:    "HL2",
			wantBaseOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.ClassifyAST(tt.expr)

			if result.Type != tt.wantType {
				t.Errorf("ClassifyAST() type = %v, want %v", result.Type, tt.wantType)
			}

			if result.BaseOffset != tt.wantBaseOffset {
				t.Errorf("ClassifyAST() BaseOffset = %d, want %d", result.BaseOffset, tt.wantBaseOffset)
			}

			if tt.wantType == SourceTypeOHLCVField {
				if result.FieldName != tt.wantFieldName {
					t.Errorf("ClassifyAST() fieldName = %q, want %q", result.FieldName, tt.wantFieldName)
				}
				if !result.IsOHLCVField() {
					t.Error("IsOHLCVField() = false, want true")
				}
			}

			if tt.wantType == SourceTypeDerivedPrice {
				if result.PriceName != tt.wantFieldName {
					t.Errorf("ClassifyAST() PriceName = %q, want %q", result.PriceName, tt.wantFieldName)
				}
				if !result.IsDerivedPrice() {
					t.Error("IsDerivedPrice() = false, want true")
				}
				if result.IsOHLCVField() {
					t.Error("IsDerivedPrice should not be OHLCV field")
				}
				if result.IsSeriesVariable() {
					t.Error("IsDerivedPrice should not be series variable")
				}
			}

			if tt.wantType == SourceTypeSeriesVariable {
				if result.VariableName != tt.wantVarName {
					t.Errorf("ClassifyAST() variableName = %q, want %q", result.VariableName, tt.wantVarName)
				}
				if !result.IsSeriesVariable() {
					t.Error("IsSeriesVariable() = false, want true")
				}
			}
		})
	}
}

/* TestSeriesSourceClassifier_ClassifyAST_MemberExpressions tests subscript access classification */
func TestSeriesSourceClassifier_ClassifyAST_MemberExpressions(t *testing.T) {
	classifier := NewSeriesSourceClassifier()

	tests := []struct {
		name           string
		expr           ast.Expression
		wantType       SourceType
		wantFieldName  string
		wantVarName    string
		wantBaseOffset int
	}{
		{
			name: "close[1] - historical OHLCV access",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 1},
				Computed: true,
			},
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "Close",
			wantBaseOffset: 1,
		},
		{
			name: "close[4] - multi-bar lookback",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 4},
				Computed: true,
			},
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "Close",
			wantBaseOffset: 4,
		},
		{
			name: "high[10] - high field lookback",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "high"},
				Property: &ast.Literal{Value: 10},
				Computed: true,
			},
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "High",
			wantBaseOffset: 10,
		},
		{
			name: "volume[0] - current bar",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "volume"},
				Property: &ast.Literal{Value: 0},
				Computed: true,
			},
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "Volume",
			wantBaseOffset: 0,
		},
		{
			name: "myVar[1] - user series subscript",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "myVar"},
				Property: &ast.Literal{Value: 1},
				Computed: true,
			},
			wantType:       SourceTypeSeriesVariable,
			wantVarName:    "myVar",
			wantBaseOffset: 1,
		},
		{
			name: "tempVar[5] - temp variable subscript",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta_sma_50_xyz"},
				Property: &ast.Literal{Value: 5},
				Computed: true,
			},
			wantType:       SourceTypeSeriesVariable,
			wantVarName:    "ta_sma_50_xyz",
			wantBaseOffset: 5,
		},
		{
			name: "hl2[1] - derived price with offset",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "hl2"},
				Property: &ast.Literal{Value: 1},
				Computed: true,
			},
			wantType:       SourceTypeDerivedPrice,
			wantFieldName:  "hl2",
			wantBaseOffset: 1,
		},
		{
			name: "hlc3[2] - derived price multi-bar",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "hlc3"},
				Property: &ast.Literal{Value: 2},
				Computed: true,
			},
			wantType:       SourceTypeDerivedPrice,
			wantFieldName:  "hlc3",
			wantBaseOffset: 2,
		},
		{
			name: "ohlc4[0] - derived price current bar",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ohlc4"},
				Property: &ast.Literal{Value: 0},
				Computed: true,
			},
			wantType:       SourceTypeDerivedPrice,
			wantFieldName:  "ohlc4",
			wantBaseOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.ClassifyAST(tt.expr)

			if result.Type != tt.wantType {
				t.Errorf("ClassifyAST() type = %v, want %v", result.Type, tt.wantType)
			}

			if result.BaseOffset != tt.wantBaseOffset {
				t.Errorf("ClassifyAST() BaseOffset = %d, want %d", result.BaseOffset, tt.wantBaseOffset)
			}

			if tt.wantType == SourceTypeOHLCVField && result.FieldName != tt.wantFieldName {
				t.Errorf("ClassifyAST() fieldName = %q, want %q", result.FieldName, tt.wantFieldName)
			}

			if tt.wantType == SourceTypeDerivedPrice && result.PriceName != tt.wantFieldName {
				t.Errorf("ClassifyAST() PriceName = %q, want %q", result.PriceName, tt.wantFieldName)
			}

			if tt.wantType == SourceTypeSeriesVariable && result.VariableName != tt.wantVarName {
				t.Errorf("ClassifyAST() variableName = %q, want %q", result.VariableName, tt.wantVarName)
			}
		})
	}
}

/* TestSeriesSourceClassifier_ClassifyAST_EdgeCases tests boundary conditions */
func TestSeriesSourceClassifier_ClassifyAST_EdgeCases(t *testing.T) {
	classifier := NewSeriesSourceClassifier()

	tests := []struct {
		name           string
		expr           ast.Expression
		wantType       SourceType
		wantFieldName  string
		wantVarName    string
		wantBaseOffset int
	}{
		{
			name: "non-computed member expression",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "bar"},
				Property: &ast.Identifier{Name: "Close"},
				Computed: false,
			},
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "Close",
			wantBaseOffset: 0,
		},
		{
			name: "nested member expression",
			expr: &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ctx"},
					Property: &ast.Identifier{Name: "Data"},
					Computed: false,
				},
				Property: &ast.Identifier{Name: "Close"},
				Computed: false,
			},
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "Close",
			wantBaseOffset: 0,
		},
		{
			name:           "nil expression defaults to Close",
			expr:           nil,
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "Close",
			wantBaseOffset: 0,
		},
		{
			name: "member expression with variable object",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "myArray"},
				Property: &ast.Literal{Value: 0},
				Computed: true,
			},
			wantType:       SourceTypeSeriesVariable,
			wantVarName:    "myArray",
			wantBaseOffset: 0,
		},
		{
			name: "deeply nested member - only innermost identifier matters",
			expr: &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ctx"},
						Property: &ast.Identifier{Name: "Data"},
						Computed: false,
					},
					Property: &ast.Identifier{Name: "BarIndex"},
					Computed: false,
				},
				Property: &ast.Identifier{Name: "Close"},
				Computed: false,
			},
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "Close",
			wantBaseOffset: 0,
		},
		{
			name: "call expression - fallback to Close",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.sma"},
			},
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "Close",
			wantBaseOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.ClassifyAST(tt.expr)

			if result.Type != tt.wantType {
				t.Errorf("ClassifyAST() type = %v, want %v", result.Type, tt.wantType)
			}

			if result.BaseOffset != tt.wantBaseOffset {
				t.Errorf("ClassifyAST() BaseOffset = %d, want %d", result.BaseOffset, tt.wantBaseOffset)
			}

			if tt.wantType == SourceTypeOHLCVField && result.FieldName != tt.wantFieldName {
				t.Errorf("ClassifyAST() fieldName = %q, want %q", result.FieldName, tt.wantFieldName)
			}

			if tt.wantType == SourceTypeSeriesVariable && result.VariableName != tt.wantVarName {
				t.Errorf("ClassifyAST() variableName = %q, want %q", result.VariableName, tt.wantVarName)
			}
		})
	}
}

/* TestSeriesSourceClassifier_ClassifyAST_BaseOffsetEdgeCases tests comprehensive BaseOffset extraction scenarios */
func TestSeriesSourceClassifier_ClassifyAST_BaseOffsetEdgeCases(t *testing.T) {
	classifier := NewSeriesSourceClassifier()

	tests := []struct {
		name           string
		expr           ast.Expression
		wantType       SourceType
		wantFieldName  string
		wantVarName    string
		wantBaseOffset int
	}{
		{
			name: "large offset - close[100]",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 100},
				Computed: true,
			},
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "Close",
			wantBaseOffset: 100,
		},
		{
			name: "float offset rounded - close[3.7]",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 3.7},
				Computed: true,
			},
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "Close",
			wantBaseOffset: 3,
		},
		{
			name: "int literal offset - close[2]",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: int(2)},
				Computed: true,
			},
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "Close",
			wantBaseOffset: 2,
		},
		{
			name: "non-literal property - close[barOffset] defaults to 0",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Identifier{Name: "barOffset"},
				Computed: true,
			},
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "Close",
			wantBaseOffset: 0,
		},
		{
			name: "series variable with large offset - myVar[50]",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "myVar"},
				Property: &ast.Literal{Value: 50},
				Computed: true,
			},
			wantType:       SourceTypeSeriesVariable,
			wantVarName:    "myVar",
			wantBaseOffset: 50,
		},
		{
			name: "volume with zero offset - volume[0]",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "volume"},
				Property: &ast.Literal{Value: 0},
				Computed: true,
			},
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "Volume",
			wantBaseOffset: 0,
		},
		{
			name: "negative offset in literal - high[-1] (should extract as -1)",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "high"},
				Property: &ast.Literal{Value: -1},
				Computed: true,
			},
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "High",
			wantBaseOffset: -1,
		},
		{
			name: "computed=false with literal property - no offset extraction",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 5},
				Computed: false,
			},
			wantType:       SourceTypeOHLCVField,
			wantFieldName:  "Close",
			wantBaseOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.ClassifyAST(tt.expr)

			if result.Type != tt.wantType {
				t.Errorf("ClassifyAST() type = %v, want %v", result.Type, tt.wantType)
			}

			if result.BaseOffset != tt.wantBaseOffset {
				t.Errorf("ClassifyAST() BaseOffset = %d, want %d", result.BaseOffset, tt.wantBaseOffset)
			}

			if tt.wantType == SourceTypeOHLCVField && result.FieldName != tt.wantFieldName {
				t.Errorf("ClassifyAST() fieldName = %q, want %q", result.FieldName, tt.wantFieldName)
			}

			if tt.wantType == SourceTypeSeriesVariable && result.VariableName != tt.wantVarName {
				t.Errorf("ClassifyAST() variableName = %q, want %q", result.VariableName, tt.wantVarName)
			}
		})
	}
}

/* TestSeriesSourceClassifier_ClassifyAST_AllBuiltinFields validates all builtin field mappings */
func TestSeriesSourceClassifier_ClassifyAST_AllBuiltinFields(t *testing.T) {
	classifier := NewSeriesSourceClassifier()

	ohlcvFields := []struct {
		input    string
		expected string
	}{
		{"close", "Close"},
		{"open", "Open"},
		{"high", "High"},
		{"low", "Low"},
		{"volume", "Volume"},
	}

	for _, field := range ohlcvFields {
		t.Run("OHLCV:"+field.input, func(t *testing.T) {
			expr := &ast.Identifier{Name: field.input}
			result := classifier.ClassifyAST(expr)

			if result.Type != SourceTypeOHLCVField {
				t.Errorf("Expected SourceTypeOHLCVField, got %v", result.Type)
			}

			if result.FieldName != field.expected {
				t.Errorf("FieldName = %q, want %q", result.FieldName, field.expected)
			}
		})
	}

	derivedPrices := []string{"hl2", "hlc3", "ohlc4", "hlcc4"}

	for _, price := range derivedPrices {
		t.Run("DerivedPrice:"+price, func(t *testing.T) {
			expr := &ast.Identifier{Name: price}
			result := classifier.ClassifyAST(expr)

			if result.Type != SourceTypeDerivedPrice {
				t.Errorf("Expected SourceTypeDerivedPrice, got %v", result.Type)
			}

			if result.PriceName != price {
				t.Errorf("PriceName = %q, want %q", result.PriceName, price)
			}
		})
	}
}

/* TestSeriesSourceClassifier_ClassifyAST_Consistency validates AST and string methods produce consistent results */
func TestSeriesSourceClassifier_ClassifyAST_Consistency(t *testing.T) {
	classifier := NewSeriesSourceClassifier()

	tests := []struct {
		name       string
		expr       ast.Expression
		stringExpr string
	}{
		{
			name:       "close identifier",
			expr:       &ast.Identifier{Name: "close"},
			stringExpr: "close",
		},
		{
			name:       "user variable",
			expr:       &ast.Identifier{Name: "myVar"},
			stringExpr: "myVarSeries.GetCurrent()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			astResult := classifier.ClassifyAST(tt.expr)
			strResult := classifier.Classify(tt.stringExpr)

			if astResult.Type != strResult.Type {
				t.Errorf("Inconsistent classification: AST=%v, String=%v", astResult.Type, strResult.Type)
			}
		})
	}
}

/* BenchmarkClassifyAST measures performance of AST-based classification */
func BenchmarkClassifyAST(b *testing.B) {
	classifier := NewSeriesSourceClassifier()

	b.Run("Identifier_OHLCV", func(b *testing.B) {
		expr := &ast.Identifier{Name: "close"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			classifier.ClassifyAST(expr)
		}
	})

	b.Run("Identifier_UserVariable", func(b *testing.B) {
		expr := &ast.Identifier{Name: "myValue"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			classifier.ClassifyAST(expr)
		}
	})

	b.Run("MemberExpression_HistoricalAccess", func(b *testing.B) {
		expr := &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "close"},
			Property: &ast.Literal{Value: 4},
			Computed: true,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			classifier.ClassifyAST(expr)
		}
	})
}
