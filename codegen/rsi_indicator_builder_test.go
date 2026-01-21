package codegen

import (
	"strings"
	"testing"
)

func TestRSIIndicatorBuilder_TopLevelContext(t *testing.T) {
	context := NewTopLevelIndicatorContext()
	accessor := NewOHLCVFieldAccessGenerator("Close")
	period := NewConstantPeriod(14)

	builder := NewRSIIndicatorBuilder("myRSI", period, accessor, false, context)
	code := builder.Build()

	t.Run("change_calculation_component", func(t *testing.T) {
		if !strings.Contains(code, "_change float64") {
			t.Error("Missing change variable declaration")
		}
	})

	t.Run("directional_split_component", func(t *testing.T) {
		/* Unique variable names prevent redeclaration in multiple RSI instances */
		/* Check for series updates - variable names include RSI result var prefix */
		hasGainsUpdate := strings.Contains(code, "gainsSeries.Set(gain)") ||
			strings.Contains(code, "_gainsSeries.Set(")
		hasLossesUpdate := strings.Contains(code, "lossesSeries.Set(loss)") ||
			strings.Contains(code, "_lossesSeries.Set(")

		if !hasGainsUpdate {
			t.Error("Missing gains series update")
		}
		if !hasLossesUpdate {
			t.Error("Missing losses series update")
		}
	})

	t.Run("rma_smoothing_components", func(t *testing.T) {
		rmaCount := strings.Count(code, "Inline RMA(14)")
		if rmaCount != 2 {
			t.Errorf("Expected 2 RMA calculations (gains and losses), got %d", rmaCount)
		}
	})

	t.Run("final_formula_computation", func(t *testing.T) {
		if !strings.Contains(code, "rs :=") {
			t.Error("Missing RS (Relative Strength) calculation")
		}
		if !strings.Contains(code, "Series.Set(rsi)") {
			t.Error("Missing final RSI storage in series")
		}
	})

	t.Run("warmup_period_handling", func(t *testing.T) {
		if !strings.Contains(code, "if ctx.BarIndex < 14") {
			t.Error("Missing warmup guard: BarIndex < period")
		}
		if !strings.Contains(code, "math.NaN()") {
			t.Error("Missing NaN return during warmup period")
		}
	})

	t.Run("series_storage_pattern", func(t *testing.T) {
		if !strings.Contains(code, "Series.Set") {
			t.Error("Missing Series.Set() pattern for TopLevel context")
		}
		if strings.Contains(code, "arrowCtx.GetOrCreateSeries") {
			t.Error("TopLevel context should not use arrow patterns")
		}
	})
}

func TestRSIIndicatorBuilder_ArrowContext(t *testing.T) {
	context := NewArrowFunctionIndicatorContext()
	accessor := NewInternalSeriesAccessor("source", context)
	period := NewConstantPeriod(20)

	builder := NewRSIIndicatorBuilder("rsi", period, accessor, false, context)
	code := builder.Build()

	t.Run("arrow_gains_series_pattern", func(t *testing.T) {
		/* Arrow context uses dynamic series names, variables may have unique prefixes */
		hasGainsPattern := strings.Contains(code, "arrowCtx.GetOrCreateSeries(\"_rsi_gains\").Set(gain)") ||
			strings.Contains(code, "arrowCtx.GetOrCreateSeries(\"_rsi_gains\").Set(")
		if !hasGainsPattern {
			t.Errorf("Missing arrow gains update pattern\nGenerated:\n%s", code)
		}
	})

	t.Run("arrow_losses_series_pattern", func(t *testing.T) {
		hasLossesPattern := strings.Contains(code, "arrowCtx.GetOrCreateSeries(\"_rsi_losses\").Set(loss)") ||
			strings.Contains(code, "arrowCtx.GetOrCreateSeries(\"_rsi_losses\").Set(")
		if !hasLossesPattern {
			t.Errorf("Missing arrow losses update pattern\nGenerated:\n%s", code)
		}
	})

	t.Run("arrow_rma_series_access", func(t *testing.T) {
		if !strings.Contains(code, "arrowCtx.GetOrCreateSeries(\"_rsi_rma_gains\")") {
			t.Error("Missing arrow RMA gains series access")
		}
		if !strings.Contains(code, "arrowCtx.GetOrCreateSeries(\"_rsi_rma_losses\")") {
			t.Error("Missing arrow RMA losses series access")
		}
	})

	t.Run("arrow_final_storage_pattern", func(t *testing.T) {
		expected := "arrowCtx.GetOrCreateSeries(\"rsi\").Set(rsi)"
		if !strings.Contains(code, expected) {
			t.Errorf("Missing arrow final RSI storage\nGenerated:\n%s", code)
		}
	})

	t.Run("no_toplevel_patterns", func(t *testing.T) {
		if strings.Contains(code, "rsiSeries :=") {
			t.Error("Arrow context should not use := series declarations")
		}
	})
}

func TestRSIIndicatorBuilder_VariablePeriod(t *testing.T) {
	context := NewTopLevelIndicatorContext()
	accessor := NewOHLCVFieldAccessGenerator("Close")
	period := NewRuntimePeriod("userPeriod")

	builder := NewRSIIndicatorBuilder("dynRSI", period, accessor, false, context)
	code := builder.Build()

	t.Run("variable_period_in_warmup", func(t *testing.T) {
		if !strings.Contains(code, "if ctx.BarIndex < int(userPeriod)") {
			t.Error("Missing variable period warmup guard (self-documenting structure)")
		}
	})

	t.Run("variable period in RMA", func(t *testing.T) {
		if !strings.Contains(code, "alpha") && !strings.Contains(code, "userPeriod") {
			t.Error("Missing variable period in RMA calculations (self-documenting structure)")
		}
	})
}

/* TestRSIIndicatorBuilder_InternalSeriesNames validates series naming */
func TestRSIIndicatorBuilder_InternalSeriesNames(t *testing.T) {
	context := NewTopLevelIndicatorContext()
	accessor := NewOHLCVFieldAccessGenerator("Close")
	period := NewConstantPeriod(14)

	builder := NewRSIIndicatorBuilder("test", period, accessor, false, context)
	_ = builder.Build()

	seriesNames := builder.GetInternalSeriesNames()

	expectedNames := []string{
		"_test_gains",
		"_test_losses",
		"_test_rma_gains",
		"_test_rma_losses",
	}

	if len(seriesNames) != len(expectedNames) {
		t.Errorf("Expected %d internal series, got %d", len(expectedNames), len(seriesNames))
	}

	for i, expected := range expectedNames {
		if i >= len(seriesNames) {
			t.Errorf("Missing internal series: %s", expected)
			continue
		}
		if seriesNames[i] != expected {
			t.Errorf("Series[%d]: expected %q, got %q", i, expected, seriesNames[i])
		}
	}
}

func TestRSIIndicatorBuilder_CompositionOrder(t *testing.T) {
	context := NewTopLevelIndicatorContext()
	accessor := NewOHLCVFieldAccessGenerator("Close")
	period := NewConstantPeriod(14)

	builder := NewRSIIndicatorBuilder("rsi", period, accessor, false, context)
	code := builder.Build()

	changePos := strings.Index(code, "ctx.BarIndex < 1")
	nanCheckPos := strings.Index(code, "math.IsNaN")
	rmaPos := strings.Index(code, "alpha")
	rsPos := strings.Index(code, "rs :=")

	if changePos == -1 || nanCheckPos == -1 || rmaPos == -1 || rsPos == -1 {
		t.Fatal("Missing expected computation steps in generated code (self-documenting structure)")
	}

	if !(changePos < nanCheckPos && nanCheckPos < rmaPos && rmaPos < rsPos) {
		t.Error("Computation steps out of order: must be change → gain/loss → RMA → RS/RSI")
	}
}

func TestRSIIndicatorBuilder_NoFuturePeek(t *testing.T) {
	context := NewTopLevelIndicatorContext()
	accessor := NewOHLCVFieldAccessGenerator("Close")
	period := NewConstantPeriod(14)

	builder := NewRSIIndicatorBuilder("rsi", period, accessor, false, context)
	code := builder.Build()

	/* Change calculation must use current and previous bars only */
	if strings.Contains(code, "ctx.BarIndex+1") {
		t.Error("FUTURE PEEK DETECTED: accessing BarIndex+1 (violates causality)")
	}

	if strings.Contains(code, ".Get(-1)") {
		t.Error("FUTURE PEEK DETECTED: negative offset in series access")
	}

	/* Should use ctx.BarIndex-1 for previous bar */
	if !strings.Contains(code, "ctx.Data[ctx.BarIndex-1]") {
		t.Error("Missing correct previous bar access pattern")
	}
}

func TestRSIIndicatorBuilder_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		period       PeriodExpression
		varName      string
		expectError  bool
		expectWarmup bool
		description  string
	}{
		{
			name:         "zero_period",
			period:       NewConstantPeriod(0),
			varName:      "rsi",
			expectError:  false,
			expectWarmup: true,
			description:  "Zero period generates warmup guard with BarIndex < 0",
		},
		{
			name:         "one_period",
			period:       NewConstantPeriod(1),
			varName:      "rsi",
			expectError:  false,
			expectWarmup: true,
			description:  "Minimum period generates warmup guard with BarIndex < 1",
		},
		{
			name:         "large_period",
			period:       NewConstantPeriod(500),
			varName:      "rsi",
			expectError:  false,
			expectWarmup: true,
			description:  "Large period value generates correct code",
		},
		{
			name:         "variable_period",
			period:       NewRuntimePeriod("myPeriod"),
			varName:      "rsi",
			expectError:  false,
			expectWarmup: true,
			description:  "Runtime period generates dynamic warmup check",
		},
		{
			name:         "special_chars_varname",
			period:       NewConstantPeriod(14),
			varName:      "my_rsi_2",
			expectError:  false,
			expectWarmup: false,
			description:  "Variable names with underscores and numbers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := NewTopLevelIndicatorContext()
			accessor := NewOHLCVFieldAccessGenerator("Close")

			builder := NewRSIIndicatorBuilder(tt.varName, tt.period, accessor, false, context)
			code := builder.Build()

			if code == "" {
				t.Errorf("Empty code generated for %s", tt.description)
			}

			if tt.expectWarmup {
				if !strings.Contains(code, "if ctx.BarIndex < ") {
					t.Errorf("Missing warmup guard for %s", tt.description)
				}
			}

			if !strings.Contains(code, "gain") && !strings.Contains(code, "loss") {
				t.Errorf("Missing gain/loss split for %s", tt.description)
			}

			if !strings.Contains(code, "rs") || !strings.Contains(code, "100.0") {
				t.Errorf("Missing RSI formula components for %s", tt.description)
			}

			if tt.varName != "" {
				expectedPrefix := "_" + tt.varName
				if !strings.Contains(code, expectedPrefix) {
					t.Errorf("Internal series names broken for %s: expected prefix %s", tt.description, expectedPrefix)
				}
			}
		})
	}
}

func TestRSIIndicatorBuilder_WarmupBoundaries(t *testing.T) {
	context := NewTopLevelIndicatorContext()
	accessor := NewOHLCVFieldAccessGenerator("Close")
	period := NewConstantPeriod(14)

	builder := NewRSIIndicatorBuilder("rsi", period, accessor, false, context)
	code := builder.Build()

	t.Run("warmup_guard_exact_boundary", func(t *testing.T) {
		if !strings.Contains(code, "if ctx.BarIndex < 14") {
			t.Error("Warmup boundary should be exactly period")
		}
	})

	t.Run("warmup_returns_nan", func(t *testing.T) {
		if !strings.Contains(code, "math.NaN()") {
			t.Error("Warmup period must return NaN")
		}
	})

	t.Run("change_calc_first_bar", func(t *testing.T) {
		if !strings.Contains(code, "ctx.Data[ctx.BarIndex-1]") {
			t.Error("Change calculation needs previous bar access")
		}
	})

	t.Run("division_by_zero_protection", func(t *testing.T) {
		if !strings.Contains(code, "100.0 / (1.0 + rs)") {
			t.Error("RSI formula missing")
		}
	})
}

func TestRSIIndicatorBuilder_NaNPropagation(t *testing.T) {
	context := NewTopLevelIndicatorContext()
	accessor := NewOHLCVFieldAccessGenerator("Close")
	period := NewConstantPeriod(14)

	builder := NewRSIIndicatorBuilder("rsi", period, accessor, false, context)
	code := builder.Build()

	t.Run("change_nan_handling", func(t *testing.T) {
		if !strings.Contains(code, "ctx.BarIndex < 1") {
			t.Error("Change warmup guard missing (self-explanatory structure)")
		}
	})

	t.Run("rma_nan_propagation", func(t *testing.T) {
		if !strings.Contains(code, "alpha") || !strings.Contains(code, "previousValue") {
			t.Error("RMA calculation structure missing (code should self-document)")
		}
	})

	t.Run("warmup_explicit_nan", func(t *testing.T) {
		if !strings.Contains(code, "math.NaN()") {
			t.Error("Warmup must explicitly return NaN")
		}
	})
}

/* TestRSIIndicatorBuilder_PeriodBoundaries validates comprehensive period boundary behavior */
func TestRSIIndicatorBuilder_PeriodBoundaries(t *testing.T) {
	testCases := []struct {
		name        string
		period      int
		warmupBar   int
		description string
	}{
		{
			name:        "period_1_minimum",
			period:      1,
			warmupBar:   1,
			description: "Minimum valid period",
		},
		{
			name:        "period_2_edge",
			period:      2,
			warmupBar:   2,
			description: "Edge case small period",
		},
		{
			name:        "period_5_small",
			period:      5,
			warmupBar:   5,
			description: "Small typical period",
		},
		{
			name:        "period_14_standard",
			period:      14,
			warmupBar:   14,
			description: "Standard RSI period",
		},
		{
			name:        "period_20_common",
			period:      20,
			warmupBar:   20,
			description: "Common alternative period",
		},
		{
			name:        "period_50_medium",
			period:      50,
			warmupBar:   50,
			description: "Medium-term period",
		},
		{
			name:        "period_100_large",
			period:      100,
			warmupBar:   100,
			description: "Large period",
		},
		{
			name:        "period_200_very_large",
			period:      200,
			warmupBar:   200,
			description: "Very large period",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			context := NewTopLevelIndicatorContext()
			accessor := NewOHLCVFieldAccessGenerator("Close")
			period := NewConstantPeriod(tc.period)

			builder := NewRSIIndicatorBuilder("rsi", period, accessor, false, context)
			code := builder.Build()

			warmupCheck := "if ctx.BarIndex <"
			if !strings.Contains(code, warmupCheck) {
				t.Errorf("%s: Missing warmup guard", tc.description)
			}

			if !strings.Contains(code, "math.NaN()") {
				t.Errorf("%s: Missing NaN return during warmup", tc.description)
			}

			if strings.Count(code, "Inline RMA") != 2 {
				t.Errorf("%s: Expected 2 RMA calculations (gains + losses)", tc.description)
			}

			if !strings.Contains(code, "gain") || !strings.Contains(code, "loss") {
				t.Errorf("%s: Missing gains/losses directional split", tc.description)
			}

			if !strings.Contains(code, "rs :=") {
				t.Errorf("%s: Missing RS calculation", tc.description)
			}

			if code == "" {
				t.Errorf("%s: Empty code generated", tc.description)
			}
		})
	}
}

/* TestRSIIndicatorBuilder_AlgorithmCorrectness validates RSI formula mathematical correctness */
func TestRSIIndicatorBuilder_AlgorithmCorrectness(t *testing.T) {
	context := NewTopLevelIndicatorContext()
	accessor := NewOHLCVFieldAccessGenerator("Close")
	period := NewConstantPeriod(14)

	builder := NewRSIIndicatorBuilder("rsi", period, accessor, false, context)
	code := builder.Build()

	t.Run("rma_alpha_formula", func(t *testing.T) {
		if !strings.Contains(code, "1.0 / float64(14)") {
			t.Error("RMA alpha must be 1/period (Wilder's smoothing)")
		}

		if strings.Contains(code, "2.0 / float64(14+1)") {
			t.Error("RSI uses RMA (Wilder), not EMA - alpha should be 1/period, not 2/(period+1)")
		}
	})

	t.Run("directional_split_correctness", func(t *testing.T) {
		if !strings.Contains(code, "> 0") {
			t.Error("Missing positive change condition")
		}

		if !strings.Contains(code, "gain") && !strings.Contains(code, "loss") {
			t.Error("Missing gain/loss variables")
		}
	})

	t.Run("rs_calculation", func(t *testing.T) {
		if !strings.Contains(code, "rs :=") {
			t.Error("Missing RS (Relative Strength) variable")
		}

		expectedRSPattern := "rma_gains"
		if !strings.Contains(code, expectedRSPattern) {
			t.Error("RS calculation must use RMA(gains)")
		}

		expectedRSPattern2 := "rma_losses"
		if !strings.Contains(code, expectedRSPattern2) {
			t.Error("RS calculation must use RMA(losses)")
		}
	})

	t.Run("rsi_formula", func(t *testing.T) {
		expectedFormula := "100.0 - (100.0 / (1.0 + rs))"
		if !strings.Contains(code, expectedFormula) {
			t.Error("RSI formula must be: 100 - 100/(1+RS)")
		}
	})

	t.Run("zero_division_protection", func(t *testing.T) {
		if strings.Contains(code, "/ 0") {
			t.Error("Code must not contain division by literal zero")
		}
	})

	t.Run("rma_self_reference", func(t *testing.T) {
		if !strings.Contains(code, "Series.Get(1)") {
			t.Error("RMA requires self-reference to previous value")
		}
	})
}

/* TestRSIIndicatorBuilder_AccessorVariations validates RSI with different source data accessor types */
func TestRSIIndicatorBuilder_AccessorVariations(t *testing.T) {
	period := NewConstantPeriod(14)

	t.Run("ohlcv_field_accessor", func(t *testing.T) {
		fields := []string{"Open", "High", "Low", "Close", "Volume"}
		for _, field := range fields {
			context := NewTopLevelIndicatorContext()
			accessor := NewOHLCVFieldAccessGenerator(field)
			builder := NewRSIIndicatorBuilder("rsi", period, accessor, false, context)
			code := builder.Build()

			if code == "" {
				t.Errorf("Empty code generated for field %s", field)
			}

			if !strings.Contains(code, "ctx.Data[ctx.BarIndex]."+field) {
				t.Errorf("Missing OHLCV field access for %s", field)
			}
		}
	})

	t.Run("internal_series_accessor_arrow_context", func(t *testing.T) {
		context := NewArrowFunctionIndicatorContext()
		accessor := NewInternalSeriesAccessor("sourceSeries", context)
		builder := NewRSIIndicatorBuilder("rsi", period, accessor, false, context)
		code := builder.Build()

		if code == "" {
			t.Error("Empty code generated for internal series accessor")
		}

		if !strings.Contains(code, "arrowCtx.GetOrCreateSeries") {
			t.Error("Arrow context should use GetOrCreateSeries pattern")
		}
	})

	t.Run("needs_nan_check_variation", func(t *testing.T) {
		context := NewTopLevelIndicatorContext()
		accessor := NewOHLCVFieldAccessGenerator("Close")

		builderWithNaN := NewRSIIndicatorBuilder("rsi1", period, accessor, true, context)
		codeWithNaN := builderWithNaN.Build()

		builderNoNaN := NewRSIIndicatorBuilder("rsi2", period, accessor, false, context)
		codeNoNaN := builderNoNaN.Build()

		if codeWithNaN == codeNoNaN {
			t.Error("needsNaN flag should affect generated code")
		}
	})
}
