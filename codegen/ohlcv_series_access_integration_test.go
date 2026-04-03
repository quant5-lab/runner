package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestOHLCVSeriesAccessPatterns validates ForwardSeriesBuffer pattern consistency across all code paths */
func TestOHLCVSeriesAccessPatterns(t *testing.T) {
	t.Run("all OHLCV fields generate consistent Series access", func(t *testing.T) {
		fields := []struct {
			fieldName  string
			seriesName string
		}{
			{"Close", "closeSeries"},
			{"High", "highSeries"},
			{"Low", "lowSeries"},
			{"Open", "openSeries"},
			{"Volume", "volumeSeries"},
		}

		for _, field := range fields {
			t.Run(field.fieldName, func(t *testing.T) {
				gen := NewOHLCVFieldAccessGenerator(field.fieldName)

				current := gen.GenerateCurrentValueAccess()
				if !strings.Contains(current, field.seriesName+".GetCurrent()") {
					t.Errorf("Current access = %q, want %s.GetCurrent()", current, field.seriesName)
				}

				loop := gen.GenerateLoopValueAccess("j")
				if !strings.Contains(loop, field.seriesName+".Get(j)") {
					t.Errorf("Loop access = %q, want %s.Get(j)", loop, field.seriesName)
				}

				initial := gen.GenerateInitialValueAccess(20)
				if !strings.Contains(initial, field.seriesName+".Get(19)") {
					t.Errorf("Initial access = %q, want %s.Get(19)", initial, field.seriesName)
				}
			})
		}
	})

	t.Run("derived prices generate Series access", func(t *testing.T) {
		derivedPrices := []struct {
			name     string
			mustHave []string
		}{
			{
				name:     "hl2",
				mustHave: []string{"highSeries.Get", "lowSeries.Get"},
			},
			{
				name:     "hlc3",
				mustHave: []string{"highSeries.Get", "lowSeries.Get", "closeSeries.Get"},
			},
			{
				name:     "ohlc4",
				mustHave: []string{"openSeries.Get", "highSeries.Get", "lowSeries.Get", "closeSeries.Get"},
			},
			{
				name:     "hlcc4",
				mustHave: []string{"highSeries.Get", "lowSeries.Get", "closeSeries.Get"},
			},
		}

		for _, dp := range derivedPrices {
			t.Run(dp.name, func(t *testing.T) {
				st := NewSymbolTable()
				conv := NewSeriesAccessConverter(st, "j", nil)
				expr := &ast.Identifier{Name: dp.name}

				code, err := conv.ConvertExpression(expr)
				if err != nil {
					t.Fatalf("ConvertExpression(%s) failed: %v", dp.name, err)
				}

				for _, pattern := range dp.mustHave {
					if !strings.Contains(code, pattern) {
						t.Errorf("%s: missing pattern %q in generated code: %s", dp.name, pattern, code)
					}
				}
			})
		}
	})

	t.Run("offset variations maintain Series pattern", func(t *testing.T) {
		offsets := []struct {
			offset     int
			loopVar    string
			wantSuffix string
		}{
			{0, "j", ".Get(j)"},
			{1, "j", ".Get(j+1)"},
			{5, "i", ".Get(i+5)"},
			{10, "idx", ".Get(idx+10)"},
		}

		for _, tc := range offsets {
			t.Run(tc.loopVar+" offset "+string(rune(tc.offset+'0')), func(t *testing.T) {
				gen := NewOHLCVFieldAccessGeneratorWithOffset("Close", tc.offset)
				loop := gen.GenerateLoopValueAccess(tc.loopVar)

				if !strings.Contains(loop, "closeSeries"+tc.wantSuffix) {
					t.Errorf("Loop access = %q, want pattern closeSeries%s", loop, tc.wantSuffix)
				}
			})
		}
	})

	t.Run("complex expressions preserve Series pattern", func(t *testing.T) {
		st := NewSymbolTable()
		st.Register("myVar", VariableTypeSeries)
		conv := NewSeriesAccessConverter(st, "j", nil)

		expr := &ast.BinaryExpression{
			Left:     &ast.Identifier{Name: "close"},
			Operator: "-",
			Right:    &ast.Identifier{Name: "myVar"},
		}

		code, err := conv.ConvertExpression(expr)
		if err != nil {
			t.Fatalf("ConvertExpression failed: %v", err)
		}

		if !strings.Contains(code, "closeSeries.Get(j)") {
			t.Errorf("Missing closeSeries.Get(j) in: %s", code)
		}
		if !strings.Contains(code, "myVarSeries.Get(j)") {
			t.Errorf("Missing myVarSeries.Get(j) in: %s", code)
		}
	})

	t.Run("no legacy ctx.Data patterns in any accessor", func(t *testing.T) {
		fields := []string{"Close", "High", "Low", "Open", "Volume"}

		for _, field := range fields {
			gen := NewOHLCVFieldAccessGenerator(field)

			outputs := []string{
				gen.GenerateCurrentValueAccess(),
				gen.GenerateLoopValueAccess("j"),
				gen.GenerateInitialValueAccess(20),
			}

			for _, output := range outputs {
				if strings.Contains(output, "ctx.Data[i-") || strings.Contains(output, "ctx.Data[ctx.BarIndex-") {
					t.Errorf("Legacy pattern detected in %s output: %s", field, output)
				}
			}
		}
	})
}

/* TestSeriesAccessConsistencyAcrossGenerators validates cross-generator pattern consistency */
func TestSeriesAccessConsistencyAcrossGenerators(t *testing.T) {
	t.Run("OHLCVFieldAccessGenerator vs SeriesAccessConverter consistency", func(t *testing.T) {
		fields := []string{"close", "high", "low", "open", "volume"}

		for _, field := range fields {
			t.Run(field, func(t *testing.T) {
				fieldCapitalized := strings.ToUpper(field[:1]) + field[1:]
				gen1 := NewOHLCVFieldAccessGenerator(fieldCapitalized)

				st := NewSymbolTable()
				conv := NewSeriesAccessConverter(st, "j", nil)
				expr := &ast.Identifier{Name: field}
				code2, err := conv.ConvertExpression(expr)
				if err != nil {
					t.Fatalf("SeriesAccessConverter failed: %v", err)
				}

				gen1Loop := gen1.GenerateLoopValueAccess("j")

				if gen1Loop != code2 {
					t.Errorf("Inconsistent patterns:\n  OHLCVFieldAccessGenerator: %s\n  SeriesAccessConverter: %s",
						gen1Loop, code2)
				}
			})
		}
	})

	t.Run("all generators avoid legacy array access", func(t *testing.T) {
		legacyPatterns := []string{
			"ctx.Data[i-",
			"ctx.Data[ctx.BarIndex-i]",
		}

		generators := []struct {
			name   string
			output string
		}{
			{
				"OHLCVFieldAccessGenerator",
				NewOHLCVFieldAccessGenerator("Close").GenerateLoopValueAccess("j"),
			},
			{
				"BuiltinIdentifierAccessor",
				NewBuiltinIdentifierAccessor("ctx.Data[ctx.BarIndex].Close").GenerateLoopValueAccess("j"),
			},
			{
				"OHLCVFieldAccessor",
				NewOHLCVFieldAccessor("Close").GetAccessExpression("j"),
			},
		}

		for _, gen := range generators {
			for _, pattern := range legacyPatterns {
				if strings.Contains(gen.output, pattern) {
					t.Errorf("%s uses legacy pattern %q: %s", gen.name, pattern, gen.output)
				}
			}
		}
	})
}

/* TestOHLCVSeriesAccessEdgeCases validates boundary conditions and special cases */
func TestOHLCVSeriesAccessEdgeCases(t *testing.T) {
	t.Run("zero offset explicit", func(t *testing.T) {
		gen := NewOHLCVFieldAccessGeneratorWithOffset("Close", 0)
		loop := gen.GenerateLoopValueAccess("j")

		if loop != "closeSeries.Get(j)" {
			t.Errorf("Zero offset should generate .Get(j), got: %s", loop)
		}
	})

	t.Run("large offset values", func(t *testing.T) {
		gen := NewOHLCVFieldAccessGeneratorWithOffset("High", 1000)
		loop := gen.GenerateLoopValueAccess("j")

		if !strings.Contains(loop, "highSeries.Get(j+1000)") {
			t.Errorf("Large offset not handled correctly: %s", loop)
		}
	})

	t.Run("different loop variable names", func(t *testing.T) {
		gen := NewOHLCVFieldAccessGenerator("Low")
		loopVars := []string{"j", "i", "idx", "k", "offset"}

		for _, loopVar := range loopVars {
			loop := gen.GenerateLoopValueAccess(loopVar)
			expected := "lowSeries.Get(" + loopVar + ")"

			if loop != expected {
				t.Errorf("Loop var %s: got %s, want %s", loopVar, loop, expected)
			}
		}
	})

	t.Run("all OHLCV fields with GetCurrent", func(t *testing.T) {
		fields := []struct {
			name   string
			series string
		}{
			{"Close", "closeSeries"},
			{"High", "highSeries"},
			{"Low", "lowSeries"},
			{"Open", "openSeries"},
			{"Volume", "volumeSeries"},
		}

		for _, field := range fields {
			gen := NewOHLCVFieldAccessGenerator(field.name)
			current := gen.GenerateCurrentValueAccess()
			expected := field.series + ".GetCurrent()"

			if current != expected {
				t.Errorf("%s current access: got %s, want %s", field.name, current, expected)
			}
		}
	})

	t.Run("InitialValueAccess uses correct lookback", func(t *testing.T) {
		periods := []int{1, 5, 10, 20, 50, 100, 200}

		for _, period := range periods {
			gen := NewOHLCVFieldAccessGenerator("Close")
			initial := gen.GenerateInitialValueAccess(period)

			if !strings.Contains(initial, "closeSeries.Get") {
				t.Errorf("Period %d: missing Series.Get pattern: %s", period, initial)
			}
		}
	})
}
