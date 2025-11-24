package codegen

import (
	"testing"
)

func TestSeriesSourceClassifier_ClassifySeriesVariable(t *testing.T) {
	classifier := NewSeriesSourceClassifier()

	tests := []struct {
		name        string
		sourceExpr  string
		wantType    SourceType
		wantVarName string
	}{
		{
			name:        "simple series variable",
			sourceExpr:  "cagr5Series.Get(0)",
			wantType:    SourceTypeSeriesVariable,
			wantVarName: "cagr5",
		},
		{
			name:        "series variable with GetCurrent",
			sourceExpr:  "myValueSeries.GetCurrent()",
			wantType:    SourceTypeSeriesVariable,
			wantVarName: "myValue",
		},
		{
			name:        "underscore in variable name",
			sourceExpr:  "my_var_Series.Get(10)",
			wantType:    SourceTypeSeriesVariable,
			wantVarName: "my_var_",
		},
		{
			name:        "number in variable name",
			sourceExpr:  "value123Series.Get(5)",
			wantType:    SourceTypeSeriesVariable,
			wantVarName: "value123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.Classify(tt.sourceExpr)

			if result.Type != tt.wantType {
				t.Errorf("Classify(%q) type = %v, want %v", tt.sourceExpr, result.Type, tt.wantType)
			}

			if result.VariableName != tt.wantVarName {
				t.Errorf("Classify(%q) variableName = %q, want %q", tt.sourceExpr, result.VariableName, tt.wantVarName)
			}

			if !result.IsSeriesVariable() {
				t.Errorf("IsSeriesVariable() = false, want true")
			}
		})
	}
}

func TestSeriesSourceClassifier_ClassifyOHLCVField(t *testing.T) {
	classifier := NewSeriesSourceClassifier()

	tests := []struct {
		name          string
		sourceExpr    string
		wantType      SourceType
		wantFieldName string
	}{
		{
			name:          "close field with prefix",
			sourceExpr:    "bar.Close",
			wantType:      SourceTypeOHLCVField,
			wantFieldName: "Close",
		},
		{
			name:          "close field standalone",
			sourceExpr:    "close",
			wantType:      SourceTypeOHLCVField,
			wantFieldName: "close",
		},
		{
			name:          "high field",
			sourceExpr:    "ctx.Data[i].High",
			wantType:      SourceTypeOHLCVField,
			wantFieldName: "High",
		},
		{
			name:          "volume field",
			sourceExpr:    "Volume",
			wantType:      SourceTypeOHLCVField,
			wantFieldName: "Volume",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.Classify(tt.sourceExpr)

			if result.Type != tt.wantType {
				t.Errorf("Classify(%q) type = %v, want %v", tt.sourceExpr, result.Type, tt.wantType)
			}

			if result.FieldName != tt.wantFieldName {
				t.Errorf("Classify(%q) fieldName = %q, want %q", tt.sourceExpr, result.FieldName, tt.wantFieldName)
			}

			if !result.IsOHLCVField() {
				t.Errorf("IsOHLCVField() = false, want true")
			}
		})
	}
}

func TestSeriesSourceClassifier_EdgeCases(t *testing.T) {
	classifier := NewSeriesSourceClassifier()

	tests := []struct {
		name       string
		sourceExpr string
		wantType   SourceType
	}{
		{
			name:       "empty string",
			sourceExpr: "",
			wantType:   SourceTypeOHLCVField,
		},
		{
			name:       "just dots",
			sourceExpr: "...",
			wantType:   SourceTypeOHLCVField,
		},
		{
			name:       "series without Get",
			sourceExpr: "valueSeries",
			wantType:   SourceTypeOHLCVField,
		},
		{
			name:       "Get without Series prefix",
			sourceExpr: "something.Get(0)",
			wantType:   SourceTypeOHLCVField,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.Classify(tt.sourceExpr)

			if result.Type != tt.wantType {
				t.Errorf("Classify(%q) type = %v, want %v", tt.sourceExpr, result.Type, tt.wantType)
			}
		})
	}
}

func TestSeriesVariableAccessGenerator(t *testing.T) {
	gen := NewSeriesVariableAccessGenerator("myVar")

	t.Run("GenerateInitialValueAccess", func(t *testing.T) {
		tests := []struct {
			period int
			want   string
		}{
			{period: 5, want: "myVarSeries.Get(5-1)"},
			{period: 20, want: "myVarSeries.Get(20-1)"},
			{period: 60, want: "myVarSeries.Get(60-1)"},
		}

		for _, tt := range tests {
			got := gen.GenerateInitialValueAccess(tt.period)
			if got != tt.want {
				t.Errorf("GenerateInitialValueAccess(%d) = %q, want %q", tt.period, got, tt.want)
			}
		}
	})

	t.Run("GenerateLoopValueAccess", func(t *testing.T) {
		tests := []struct {
			loopVar string
			want    string
		}{
			{loopVar: "j", want: "myVarSeries.Get(j)"},
			{loopVar: "i", want: "myVarSeries.Get(i)"},
			{loopVar: "idx", want: "myVarSeries.Get(idx)"},
		}

		for _, tt := range tests {
			got := gen.GenerateLoopValueAccess(tt.loopVar)
			if got != tt.want {
				t.Errorf("GenerateLoopValueAccess(%q) = %q, want %q", tt.loopVar, got, tt.want)
			}
		}
	})
}

func TestOHLCVFieldAccessGenerator(t *testing.T) {
	gen := NewOHLCVFieldAccessGenerator("Close")

	t.Run("GenerateInitialValueAccess", func(t *testing.T) {
		tests := []struct {
			period int
			want   string
		}{
			{period: 5, want: "ctx.Data[ctx.BarIndex-(5-1)].Close"},
			{period: 20, want: "ctx.Data[ctx.BarIndex-(20-1)].Close"},
			{period: 60, want: "ctx.Data[ctx.BarIndex-(60-1)].Close"},
		}

		for _, tt := range tests {
			got := gen.GenerateInitialValueAccess(tt.period)
			if got != tt.want {
				t.Errorf("GenerateInitialValueAccess(%d) = %q, want %q", tt.period, got, tt.want)
			}
		}
	})

	t.Run("GenerateLoopValueAccess", func(t *testing.T) {
		tests := []struct {
			loopVar string
			want    string
		}{
			{loopVar: "j", want: "ctx.Data[ctx.BarIndex-j].Close"},
			{loopVar: "i", want: "ctx.Data[ctx.BarIndex-i].Close"},
			{loopVar: "idx", want: "ctx.Data[ctx.BarIndex-idx].Close"},
		}

		for _, tt := range tests {
			got := gen.GenerateLoopValueAccess(tt.loopVar)
			if got != tt.want {
				t.Errorf("GenerateLoopValueAccess(%q) = %q, want %q", tt.loopVar, got, tt.want)
			}
		}
	})
}

func TestCreateAccessGenerator(t *testing.T) {
	t.Run("creates SeriesVariableAccessGenerator", func(t *testing.T) {
		source := SourceInfo{
			Type:         SourceTypeSeriesVariable,
			VariableName: "cagr5",
		}

		gen := CreateAccessGenerator(source)

		got := gen.GenerateInitialValueAccess(60)
		want := "cagr5Series.Get(60-1)"

		if got != want {
			t.Errorf("CreateAccessGenerator for series variable: got %q, want %q", got, want)
		}
	})

	t.Run("creates OHLCVFieldAccessGenerator", func(t *testing.T) {
		source := SourceInfo{
			Type:      SourceTypeOHLCVField,
			FieldName: "Close",
		}

		gen := CreateAccessGenerator(source)

		got := gen.GenerateInitialValueAccess(20)
		want := "ctx.Data[ctx.BarIndex-(20-1)].Close"

		if got != want {
			t.Errorf("CreateAccessGenerator for OHLCV field: got %q, want %q", got, want)
		}
	})
}

func BenchmarkClassifySeriesVariable(b *testing.B) {
	classifier := NewSeriesSourceClassifier()
	expr := "cagr5Series.Get(0)"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		classifier.Classify(expr)
	}
}

func BenchmarkClassifyOHLCVField(b *testing.B) {
	classifier := NewSeriesSourceClassifier()
	expr := "bar.Close"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		classifier.Classify(expr)
	}
}
