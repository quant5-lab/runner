package codegen

import (
	"testing"
)

// TestSeriesDataAccessor_Construction validates Series accessor creation
func TestSeriesDataAccessor_Construction(t *testing.T) {
	tests := []struct {
		name         string
		variableName string
		offset       int
		period       int
		wantInitial  string
		wantLoop     string
	}{
		{
			name:         "no offset - myVar, period 20",
			variableName: "myVar",
			offset:       0,
			period:       20,
			wantInitial:  "myVarSeries.Get(19)",
			wantLoop:     "myVarSeries.Get(j)",
		},
		{
			name:         "offset 1 - myVar[1], period 20",
			variableName: "myVar",
			offset:       1,
			period:       20,
			wantInitial:  "myVarSeries.Get(20)",
			wantLoop:     "myVarSeries.Get(j+1)",
		},
		{
			name:         "offset 2 - ema[2], period 10",
			variableName: "ema",
			offset:       2,
			period:       10,
			wantInitial:  "emaSeries.Get(11)",
			wantLoop:     "emaSeries.Get(j+2)",
		},
		{
			name:         "large offset - data[50], period 5",
			variableName: "data",
			offset:       50,
			period:       5,
			wantInitial:  "dataSeries.Get(54)",
			wantLoop:     "dataSeries.Get(j+50)",
		},
		{
			name:         "minimal period - value[0], period 1",
			variableName: "value",
			offset:       0,
			period:       1,
			wantInitial:  "valueSeries.Get(0)",
			wantLoop:     "valueSeries.Get(j)",
		},
		{
			name:         "long variable name - longTermIndicator[5], period 100",
			variableName: "longTermIndicator",
			offset:       5,
			period:       100,
			wantInitial:  "longTermIndicatorSeries.Get(104)",
			wantLoop:     "longTermIndicatorSeries.Get(j+5)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := NewHistoricalOffset(tt.offset)
			accessor := NewSeriesDataAccessor(tt.variableName, offset)

			gotInitial := accessor.GenerateInitialValueAccess(tt.period)
			if gotInitial != tt.wantInitial {
				t.Errorf("GenerateInitialValueAccess(%d) = %q, want %q",
					tt.period, gotInitial, tt.wantInitial)
			}

			gotLoop := accessor.GenerateLoopValueAccess("j")
			if gotLoop != tt.wantLoop {
				t.Errorf("GenerateLoopValueAccess(\"j\") = %q, want %q",
					gotLoop, tt.wantLoop)
			}
		})
	}
}

// TestOHLCVDataAccessor_Construction validates OHLCV accessor creation
func TestOHLCVDataAccessor_Construction(t *testing.T) {
	tests := []struct {
		name        string
		fieldName   string
		offset      int
		period      int
		wantInitial string
		wantLoop    string
	}{
		{
			name:        "no offset - Close, period 20",
			fieldName:   "Close",
			offset:      0,
			period:      20,
			wantInitial: "ctx.Data[ctx.BarIndex-19].Close",
			wantLoop:    "ctx.Data[ctx.BarIndex-j].Close",
		},
		{
			name:        "offset 1 - Close[1], period 20",
			fieldName:   "Close",
			offset:      1,
			period:      20,
			wantInitial: "ctx.Data[ctx.BarIndex-20].Close",
			wantLoop:    "ctx.Data[ctx.BarIndex-(j+1)].Close",
		},
		{
			name:        "offset 4 - Close[4], period 20 (BB7 bug case)",
			fieldName:   "Close",
			offset:      4,
			period:      20,
			wantInitial: "ctx.Data[ctx.BarIndex-23].Close",
			wantLoop:    "ctx.Data[ctx.BarIndex-(j+4)].Close",
		},
		{
			name:        "High field with offset - High[10], period 50",
			fieldName:   "High",
			offset:      10,
			period:      50,
			wantInitial: "ctx.Data[ctx.BarIndex-59].High",
			wantLoop:    "ctx.Data[ctx.BarIndex-(j+10)].High",
		},
		{
			name:        "Low field no offset - Low, period 14",
			fieldName:   "Low",
			offset:      0,
			period:      14,
			wantInitial: "ctx.Data[ctx.BarIndex-13].Low",
			wantLoop:    "ctx.Data[ctx.BarIndex-j].Low",
		},
		{
			name:        "Open field with offset - Open[2], period 5",
			fieldName:   "Open",
			offset:      2,
			period:      5,
			wantInitial: "ctx.Data[ctx.BarIndex-6].Open",
			wantLoop:    "ctx.Data[ctx.BarIndex-(j+2)].Open",
		},
		{
			name:        "Volume field with large offset - Volume[100], period 1",
			fieldName:   "Volume",
			offset:      100,
			period:      1,
			wantInitial: "ctx.Data[ctx.BarIndex-100].Volume",
			wantLoop:    "ctx.Data[ctx.BarIndex-(j+100)].Volume",
		},
		{
			name:        "minimal period - Close[0], period 1",
			fieldName:   "Close",
			offset:      0,
			period:      1,
			wantInitial: "ctx.Data[ctx.BarIndex-0].Close",
			wantLoop:    "ctx.Data[ctx.BarIndex-j].Close",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := NewHistoricalOffset(tt.offset)
			accessor := NewOHLCVDataAccessor(tt.fieldName, offset)

			gotInitial := accessor.GenerateInitialValueAccess(tt.period)
			if gotInitial != tt.wantInitial {
				t.Errorf("GenerateInitialValueAccess(%d) = %q, want %q",
					tt.period, gotInitial, tt.wantInitial)
			}

			gotLoop := accessor.GenerateLoopValueAccess("j")
			if gotLoop != tt.wantLoop {
				t.Errorf("GenerateLoopValueAccess(\"j\") = %q, want %q",
					gotLoop, tt.wantLoop)
			}
		})
	}
}

// TestDataAccessFactory_CreateAccessor validates factory pattern
func TestDataAccessFactory_CreateAccessor(t *testing.T) {
	factory := &DataAccessFactory{}

	tests := []struct {
		name              string
		sourceInfo        SourceInfo
		period            int
		wantInitialAccess string
		wantLoopAccess    string
	}{
		{
			name: "Series variable no offset",
			sourceInfo: SourceInfo{
				Type:         SourceTypeSeriesVariable,
				VariableName: "myVar",
				BaseOffset:   0,
			},
			period:            20,
			wantInitialAccess: "myVarSeries.Get(19)",
			wantLoopAccess:    "myVarSeries.Get(j)",
		},
		{
			name: "Series variable with offset 2",
			sourceInfo: SourceInfo{
				Type:         SourceTypeSeriesVariable,
				VariableName: "ema",
				BaseOffset:   2,
			},
			period:            10,
			wantInitialAccess: "emaSeries.Get(11)",
			wantLoopAccess:    "emaSeries.Get(j+2)",
		},
		{
			name: "OHLCV field no offset",
			sourceInfo: SourceInfo{
				Type:       SourceTypeOHLCVField,
				FieldName:  "Close",
				BaseOffset: 0,
			},
			period:            20,
			wantInitialAccess: "ctx.Data[ctx.BarIndex-19].Close",
			wantLoopAccess:    "ctx.Data[ctx.BarIndex-j].Close",
		},
		{
			name: "OHLCV field with offset 4 (BB7 case)",
			sourceInfo: SourceInfo{
				Type:       SourceTypeOHLCVField,
				FieldName:  "Close",
				BaseOffset: 4,
			},
			period:            20,
			wantInitialAccess: "ctx.Data[ctx.BarIndex-23].Close",
			wantLoopAccess:    "ctx.Data[ctx.BarIndex-(j+4)].Close",
		},
		{
			name: "High field with large offset",
			sourceInfo: SourceInfo{
				Type:       SourceTypeOHLCVField,
				FieldName:  "High",
				BaseOffset: 50,
			},
			period:            10,
			wantInitialAccess: "ctx.Data[ctx.BarIndex-59].High",
			wantLoopAccess:    "ctx.Data[ctx.BarIndex-(j+50)].High",
		},
		{
			name: "Series variable with large offset",
			sourceInfo: SourceInfo{
				Type:         SourceTypeSeriesVariable,
				VariableName: "longTerm",
				BaseOffset:   100,
			},
			period:            5,
			wantInitialAccess: "longTermSeries.Get(104)",
			wantLoopAccess:    "longTermSeries.Get(j+100)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := factory.CreateAccessor(tt.sourceInfo)

			gotInitial := accessor.GenerateInitialValueAccess(tt.period)
			if gotInitial != tt.wantInitialAccess {
				t.Errorf("GenerateInitialValueAccess(%d) = %q, want %q",
					tt.period, gotInitial, tt.wantInitialAccess)
			}

			gotLoop := accessor.GenerateLoopValueAccess("j")
			if gotLoop != tt.wantLoopAccess {
				t.Errorf("GenerateLoopValueAccess(\"j\") = %q, want %q",
					gotLoop, tt.wantLoopAccess)
			}
		})
	}
}

// TestDataAccessFactory_TypeDiscrimination validates factory creates correct type
func TestDataAccessFactory_TypeDiscrimination(t *testing.T) {
	factory := &DataAccessFactory{}

	t.Run("creates SeriesDataAccessor for SeriesVariable", func(t *testing.T) {
		source := SourceInfo{
			Type:         SourceTypeSeriesVariable,
			VariableName: "test",
			BaseOffset:   0,
		}

		accessor := factory.CreateAccessor(source)
		_, ok := accessor.(*SeriesDataAccessor)
		if !ok {
			t.Errorf("Expected *SeriesDataAccessor, got %T", accessor)
		}
	})

	t.Run("creates OHLCVDataAccessor for OHLCVField", func(t *testing.T) {
		source := SourceInfo{
			Type:       SourceTypeOHLCVField,
			FieldName:  "Close",
			BaseOffset: 0,
		}

		accessor := factory.CreateAccessor(source)
		_, ok := accessor.(*OHLCVDataAccessor)
		if !ok {
			t.Errorf("Expected *OHLCVDataAccessor, got %T", accessor)
		}
	})
}

// TestDataAccessStrategy_LoopVariableNames validates different loop variable names
func TestDataAccessStrategy_LoopVariableNames(t *testing.T) {
	tests := []struct {
		name       string
		accessor   DataAccessStrategy
		loopVar    string
		wantFormat string
	}{
		{
			name:       "Series accessor with j",
			accessor:   NewSeriesDataAccessor("myVar", NewHistoricalOffset(2)),
			loopVar:    "j",
			wantFormat: "myVarSeries.Get(j+2)",
		},
		{
			name:       "Series accessor with i",
			accessor:   NewSeriesDataAccessor("myVar", NewHistoricalOffset(2)),
			loopVar:    "i",
			wantFormat: "myVarSeries.Get(i+2)",
		},
		{
			name:       "Series accessor with idx",
			accessor:   NewSeriesDataAccessor("myVar", NewHistoricalOffset(2)),
			loopVar:    "idx",
			wantFormat: "myVarSeries.Get(idx+2)",
		},
		{
			name:       "OHLCV accessor with j",
			accessor:   NewOHLCVDataAccessor("Close", NewHistoricalOffset(4)),
			loopVar:    "j",
			wantFormat: "ctx.Data[ctx.BarIndex-(j+4)].Close",
		},
		{
			name:       "OHLCV accessor with loopIndex",
			accessor:   NewOHLCVDataAccessor("Close", NewHistoricalOffset(4)),
			loopVar:    "loopIndex",
			wantFormat: "ctx.Data[ctx.BarIndex-(loopIndex+4)].Close",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.accessor.GenerateLoopValueAccess(tt.loopVar)
			if got != tt.wantFormat {
				t.Errorf("GenerateLoopValueAccess(%q) = %q, want %q",
					tt.loopVar, got, tt.wantFormat)
			}
		})
	}
}

// TestDataAccessStrategy_AllOHLCVFields validates all OHLCV fields work correctly
func TestDataAccessStrategy_AllOHLCVFields(t *testing.T) {
	fields := []string{"Close", "Open", "High", "Low", "Volume"}
	offset := NewHistoricalOffset(3)
	period := 10

	for _, field := range fields {
		t.Run(field, func(t *testing.T) {
			accessor := NewOHLCVDataAccessor(field, offset)

			wantInitial := "ctx.Data[ctx.BarIndex-12]." + field
			wantLoop := "ctx.Data[ctx.BarIndex-(j+3)]." + field

			gotInitial := accessor.GenerateInitialValueAccess(period)
			if gotInitial != wantInitial {
				t.Errorf("GenerateInitialValueAccess(%d) = %q, want %q",
					period, gotInitial, wantInitial)
			}

			gotLoop := accessor.GenerateLoopValueAccess("j")
			if gotLoop != wantLoop {
				t.Errorf("GenerateLoopValueAccess(\"j\") = %q, want %q",
					gotLoop, wantLoop)
			}
		})
	}
}

// TestDataAccessStrategy_EdgeCasePeriods validates edge case period values
func TestDataAccessStrategy_EdgeCasePeriods(t *testing.T) {
	tests := []struct {
		name        string
		period      int
		offset      int
		wantSeriesInit string
		wantOHLCVInit  string
	}{
		{
			name:           "period 1, offset 0",
			period:         1,
			offset:         0,
			wantSeriesInit: "testSeries.Get(0)",
			wantOHLCVInit:  "ctx.Data[ctx.BarIndex-0].Close",
		},
		{
			name:           "period 1, offset 5",
			period:         1,
			offset:         5,
			wantSeriesInit: "testSeries.Get(5)",
			wantOHLCVInit:  "ctx.Data[ctx.BarIndex-5].Close",
		},
		{
			name:           "period 200, offset 4",
			period:         200,
			offset:         4,
			wantSeriesInit: "testSeries.Get(203)",
			wantOHLCVInit:  "ctx.Data[ctx.BarIndex-203].Close",
		},
		{
			name:           "period 100, offset 100",
			period:         100,
			offset:         100,
			wantSeriesInit: "testSeries.Get(199)",
			wantOHLCVInit:  "ctx.Data[ctx.BarIndex-199].Close",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offsetObj := NewHistoricalOffset(tt.offset)

			seriesAccessor := NewSeriesDataAccessor("test", offsetObj)
			gotSeries := seriesAccessor.GenerateInitialValueAccess(tt.period)
			if gotSeries != tt.wantSeriesInit {
				t.Errorf("Series: GenerateInitialValueAccess(%d) = %q, want %q",
					tt.period, gotSeries, tt.wantSeriesInit)
			}

			ohlcvAccessor := NewOHLCVDataAccessor("Close", offsetObj)
			gotOHLCV := ohlcvAccessor.GenerateInitialValueAccess(tt.period)
			if gotOHLCV != tt.wantOHLCVInit {
				t.Errorf("OHLCV: GenerateInitialValueAccess(%d) = %q, want %q",
					tt.period, gotOHLCV, tt.wantOHLCVInit)
			}
		})
	}
}

// BenchmarkDataAccessFactory measures factory performance
func BenchmarkDataAccessFactory(b *testing.B) {
	factory := &DataAccessFactory{}

	b.Run("CreateSeriesAccessor", func(b *testing.B) {
		source := SourceInfo{
			Type:         SourceTypeSeriesVariable,
			VariableName: "myVar",
			BaseOffset:   4,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = factory.CreateAccessor(source)
		}
	})

	b.Run("CreateOHLCVAccessor", func(b *testing.B) {
		source := SourceInfo{
			Type:       SourceTypeOHLCVField,
			FieldName:  "Close",
			BaseOffset: 4,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = factory.CreateAccessor(source)
		}
	})

	b.Run("GenerateSeriesAccess", func(b *testing.B) {
		accessor := NewSeriesDataAccessor("myVar", NewHistoricalOffset(4))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = accessor.GenerateInitialValueAccess(20)
			_ = accessor.GenerateLoopValueAccess("j")
		}
	})

	b.Run("GenerateOHLCVAccess", func(b *testing.B) {
		accessor := NewOHLCVDataAccessor("Close", NewHistoricalOffset(4))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = accessor.GenerateInitialValueAccess(20)
			_ = accessor.GenerateLoopValueAccess("j")
		}
	})
}
