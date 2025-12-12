package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestSeriesSourceClassifier_ClassifyAST_Identifiers tests classification of identifier nodes */
func TestSeriesSourceClassifier_ClassifyAST_Identifiers(t *testing.T) {
	classifier := NewSeriesSourceClassifier()

	tests := []struct {
		name          string
		expr          ast.Expression
		wantType      SourceType
		wantFieldName string
		wantVarName   string
	}{
		{
			name:          "close identifier",
			expr:          &ast.Identifier{Name: "close"},
			wantType:      SourceTypeOHLCVField,
			wantFieldName: "Close",
		},
		{
			name:          "open identifier",
			expr:          &ast.Identifier{Name: "open"},
			wantType:      SourceTypeOHLCVField,
			wantFieldName: "Open",
		},
		{
			name:          "high identifier",
			expr:          &ast.Identifier{Name: "high"},
			wantType:      SourceTypeOHLCVField,
			wantFieldName: "High",
		},
		{
			name:          "low identifier",
			expr:          &ast.Identifier{Name: "low"},
			wantType:      SourceTypeOHLCVField,
			wantFieldName: "Low",
		},
		{
			name:          "volume identifier",
			expr:          &ast.Identifier{Name: "volume"},
			wantType:      SourceTypeOHLCVField,
			wantFieldName: "Volume",
		},
		{
			name:        "user variable identifier",
			expr:        &ast.Identifier{Name: "myValue"},
			wantType:    SourceTypeSeriesVariable,
			wantVarName: "myValue",
		},
		{
			name:        "temp variable identifier",
			expr:        &ast.Identifier{Name: "ta_sma_20_abc123"},
			wantType:    SourceTypeSeriesVariable,
			wantVarName: "ta_sma_20_abc123",
		},
		{
			name:        "underscore variable",
			expr:        &ast.Identifier{Name: "my_var"},
			wantType:    SourceTypeSeriesVariable,
			wantVarName: "my_var",
		},
		{
			name:        "empty identifier",
			expr:        &ast.Identifier{Name: ""},
			wantType:    SourceTypeSeriesVariable,
			wantVarName: "",
		},
		{
			name:        "case sensitivity - Close uppercase",
			expr:        &ast.Identifier{Name: "Close"},
			wantType:    SourceTypeSeriesVariable,
			wantVarName: "Close",
		},
		{
			name:        "mixed case ohlcv (CLOSE not recognized)",
			expr:        &ast.Identifier{Name: "CLOSE"},
			wantType:    SourceTypeSeriesVariable,
			wantVarName: "CLOSE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.ClassifyAST(tt.expr)

			if result.Type != tt.wantType {
				t.Errorf("ClassifyAST() type = %v, want %v", result.Type, tt.wantType)
			}

			if tt.wantType == SourceTypeOHLCVField {
				if result.FieldName != tt.wantFieldName {
					t.Errorf("ClassifyAST() fieldName = %q, want %q", result.FieldName, tt.wantFieldName)
				}
				if !result.IsOHLCVField() {
					t.Error("IsOHLCVField() = false, want true")
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
		name          string
		expr          ast.Expression
		wantType      SourceType
		wantFieldName string
		wantVarName   string
	}{
		{
			name: "close[1] - historical OHLCV access",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 1},
				Computed: true,
			},
			wantType:      SourceTypeOHLCVField,
			wantFieldName: "Close",
		},
		{
			name: "close[4] - multi-bar lookback",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 4},
				Computed: true,
			},
			wantType:      SourceTypeOHLCVField,
			wantFieldName: "Close",
		},
		{
			name: "high[10] - high field lookback",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "high"},
				Property: &ast.Literal{Value: 10},
				Computed: true,
			},
			wantType:      SourceTypeOHLCVField,
			wantFieldName: "High",
		},
		{
			name: "volume[0] - current bar",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "volume"},
				Property: &ast.Literal{Value: 0},
				Computed: true,
			},
			wantType:      SourceTypeOHLCVField,
			wantFieldName: "Volume",
		},
		{
			name: "myVar[1] - user series subscript",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "myVar"},
				Property: &ast.Literal{Value: 1},
				Computed: true,
			},
			wantType:    SourceTypeSeriesVariable,
			wantVarName: "myVar",
		},
		{
			name: "tempVar[5] - temp variable subscript",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta_sma_50_xyz"},
				Property: &ast.Literal{Value: 5},
				Computed: true,
			},
			wantType:    SourceTypeSeriesVariable,
			wantVarName: "ta_sma_50_xyz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.ClassifyAST(tt.expr)

			if result.Type != tt.wantType {
				t.Errorf("ClassifyAST() type = %v, want %v", result.Type, tt.wantType)
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

/* TestSeriesSourceClassifier_ClassifyAST_EdgeCases tests boundary conditions */
func TestSeriesSourceClassifier_ClassifyAST_EdgeCases(t *testing.T) {
	classifier := NewSeriesSourceClassifier()

	tests := []struct {
		name          string
		expr          ast.Expression
		wantType      SourceType
		wantFieldName string
		wantVarName   string
	}{
		{
			name: "non-computed member expression",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "bar"},
				Property: &ast.Identifier{Name: "Close"},
				Computed: false,
			},
			wantType:      SourceTypeOHLCVField,
			wantFieldName: "Close",
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
			wantType:      SourceTypeOHLCVField,
			wantFieldName: "Close",
		},
		{
			name:          "nil expression defaults to Close",
			expr:          nil,
			wantType:      SourceTypeOHLCVField,
			wantFieldName: "Close",
		},
		{
			name: "member expression with variable object",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "myArray"},
				Property: &ast.Literal{Value: 0},
				Computed: true,
			},
			wantType:    SourceTypeSeriesVariable,
			wantVarName: "myArray",
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
			wantType:      SourceTypeOHLCVField,
			wantFieldName: "Close",
		},
		{
			name: "call expression - fallback to Close",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.sma"},
			},
			wantType:      SourceTypeOHLCVField,
			wantFieldName: "Close",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.ClassifyAST(tt.expr)

			if result.Type != tt.wantType {
				t.Errorf("ClassifyAST() type = %v, want %v", result.Type, tt.wantType)
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

/* TestSeriesSourceClassifier_ClassifyAST_AllOHLCVFields validates all OHLCV field mappings */
func TestSeriesSourceClassifier_ClassifyAST_AllOHLCVFields(t *testing.T) {
	classifier := NewSeriesSourceClassifier()

	fields := []struct {
		input    string
		expected string
	}{
		{"close", "Close"},
		{"open", "Open"},
		{"high", "High"},
		{"low", "Low"},
		{"volume", "Volume"},
	}

	for _, field := range fields {
		t.Run(field.input, func(t *testing.T) {
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
