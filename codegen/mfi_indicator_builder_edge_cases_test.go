package codegen

import (
	"strings"
	"testing"
)

func TestMFIIndicatorBuilder_AlgorithmCorrectness(t *testing.T) {
	testCases := []struct {
		name   string
		period int
	}{
		{"small_period", 2},
		{"typical_period", 14},
		{"large_period", 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			context := NewTopLevelIndicatorContext()
			accessor := NewOHLCVFieldAccessGenerator("Close")
			builder := NewMFIIndicatorBuilder("test", NewConstantPeriod(tc.period), accessor, context)
			code := builder.Build()

			t.Run("five_step_structure", func(t *testing.T) {
				steps := []string{
					"change", "rawMF", "_positive_mf", "_negative_mf", "mfi",
				}
				for _, step := range steps {
					if !strings.Contains(code, step) && !strings.Contains(code, strings.Replace(step, "_", "", -1)) {
						t.Errorf("Missing step in algorithm: %s", step)
					}
				}
			})

			t.Run("zero_denominator_protection", func(t *testing.T) {
				if !strings.Contains(code, "negSum == 0") {
					t.Error("Missing zero-denominator protection")
				}
				if !strings.Contains(code, "mfi := 100.0") {
					t.Error("Missing MFI=100 for zero denominator case")
				}
			})

			t.Run("warmup_handling", func(t *testing.T) {
				if !strings.Contains(code, "if ctx.BarIndex <") {
					t.Error("Missing warmup check")
				}
				if !strings.Contains(code, "NaN()") {
					t.Error("Missing NaN return during warmup")
				}
			})

			t.Run("forward_window_summation", func(t *testing.T) {
				if !strings.Contains(code, "for j := 0; j <") {
					t.Error("Missing forward window summation loop")
				}
				if !strings.Contains(code, "Get(j)") {
					t.Error("Window summation should use Get(offset)")
				}
			})

			t.Run("period_used_correctly", func(t *testing.T) {
				if !strings.Contains(code, "for j := 0; j <") {
					t.Error("Period should control loop boundary")
				}
			})
		})
	}
}

func TestMFIIndicatorBuilder_PeriodBoundaries(t *testing.T) {
	testCases := []struct {
		period int
	}{
		{1}, {2}, {14}, {100}, {200}, {500},
	}

	for _, tc := range testCases {
		context := NewTopLevelIndicatorContext()
		accessor := NewOHLCVFieldAccessGenerator("Close")
		builder := NewMFIIndicatorBuilder("mfi", NewConstantPeriod(tc.period), accessor, context)
		code := builder.Build()

		if code == "" {
			t.Errorf("Period %d generated empty code", tc.period)
		}

		if !strings.Contains(code, "ctx.BarIndex") {
			t.Errorf("Period %d missing warmup check", tc.period)
		}

		if !strings.Contains(code, "for j := 0; j <") {
			t.Errorf("Period %d missing window summation", tc.period)
		}

		if !strings.Contains(code, "negSum == 0") {
			t.Errorf("Period %d missing zero-denominator check", tc.period)
		}
	}
}

func TestMFIIndicatorBuilder_InternalSeriesNaming(t *testing.T) {
	context := NewTopLevelIndicatorContext()
	accessor := NewOHLCVFieldAccessGenerator("Close")
	builder := NewMFIIndicatorBuilder("myMFI", NewConstantPeriod(14), accessor, context)
	builder.Build()
	seriesNames := builder.GetInternalSeriesNames()

	if len(seriesNames) != 2 {
		t.Errorf("Expected 2 internal series, got %d: %v", len(seriesNames), seriesNames)
	}

	expectedPrefix := "_myMFI_"
	for _, name := range seriesNames {
		if !strings.HasPrefix(name, expectedPrefix) {
			t.Errorf("Series name %q missing prefix %q", name, expectedPrefix)
		}
	}

	hasPositive := false
	hasNegative := false
	for _, name := range seriesNames {
		if strings.Contains(name, "positive_mf") {
			hasPositive = true
		}
		if strings.Contains(name, "negative_mf") {
			hasNegative = true
		}
	}

	if !hasPositive {
		t.Error("Missing positive_mf series")
	}
	if !hasNegative {
		t.Error("Missing negative_mf series")
	}
}

func TestMFIIndicatorBuilder_CodeStructureInvariants(t *testing.T) {
	context := NewTopLevelIndicatorContext()
	accessor := NewOHLCVFieldAccessGenerator("Close")
	builder := NewMFIIndicatorBuilder("test", NewConstantPeriod(14), accessor, context)
	code := builder.Build()

	t.Run("balanced_braces", func(t *testing.T) {
		openCount := strings.Count(code, "{")
		closeCount := strings.Count(code, "}")
		if openCount != closeCount {
			t.Errorf("Unbalanced braces: %d open, %d close", openCount, closeCount)
		}
	})

	t.Run("proper_indentation", func(t *testing.T) {
		lines := strings.Split(code, "\n")
		hasIndentation := false
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			if strings.HasPrefix(line, "\t") || strings.HasPrefix(line, "    ") {
				hasIndentation = true
				break
			}
		}
		if !hasIndentation {
			t.Error("Code should have proper indentation")
		}
	})

	t.Run("series_creation_present", func(t *testing.T) {
		if len(code) < 200 {
			t.Error("Builder should generate substantial code")
		}
	})

	t.Run("series_updates", func(t *testing.T) {
		if !strings.Contains(code, "_positive_mf") && !strings.Contains(code, "posMF") {
			t.Error("Missing positive money flow series updates")
		}
		if !strings.Contains(code, "_negative_mf") && !strings.Contains(code, "negMF") {
			t.Error("Missing negative money flow series updates")
		}
	})

	t.Run("code_completeness", func(t *testing.T) {
		if len(code) < 200 {
			t.Error("Generated code should be substantial")
		}
	})
}

func TestMFIIndicatorBuilder_ChangeCalculatorIntegration(t *testing.T) {
	context := NewTopLevelIndicatorContext()
	accessor := NewOHLCVFieldAccessGenerator("Close")
	builder := NewMFIIndicatorBuilder("test", NewConstantPeriod(14), accessor, context)
	code := builder.Build()

	t.Run("change_calculation_present", func(t *testing.T) {
		if !strings.Contains(code, "change") {
			t.Error("Missing change variable")
		}
	})

	t.Run("nan_check_for_change", func(t *testing.T) {
		if !strings.Contains(code, "math.IsNaN") {
			t.Error("Missing NaN check for change calculation")
		}
	})

	t.Run("change_retrieval", func(t *testing.T) {
		if !strings.Contains(code, "Get(1)") || !strings.Contains(code, "GetCurrent()") {
			t.Error("Change calculation should use previous and current values")
		}
	})
}

func TestMFIIndicatorBuilder_DirectionalSplitLogic(t *testing.T) {
	context := NewTopLevelIndicatorContext()
	accessor := NewOHLCVFieldAccessGenerator("Close")
	builder := NewMFIIndicatorBuilder("test", NewConstantPeriod(14), accessor, context)
	code := builder.Build()

	t.Run("positive_change_branch", func(t *testing.T) {
		if !strings.Contains(code, "> 0") {
			t.Error("Missing positive change check")
		}
	})

	t.Run("negative_change_branch", func(t *testing.T) {
		if !strings.Contains(code, "< 0") {
			t.Error("Missing negative change branch")
		}
	})

	t.Run("zero_change_excluded_from_both_flows", func(t *testing.T) {
		if got := strings.Count(code, "posMF = 0.0"); got < 3 {
			t.Errorf("posMF must be zeroed in at least 3 branches (NaN+neg+zero), got %d", got)
		}
		if got := strings.Count(code, "negMF = 0.0"); got < 3 {
			t.Errorf("negMF must be zeroed in at least 3 branches (NaN+pos+zero), got %d", got)
		}
	})

	t.Run("nan_zero_assignment", func(t *testing.T) {
		if !strings.Contains(code, "math.IsNaN") {
			t.Error("Missing NaN check")
		}
	})

	t.Run("uses_raw_money_flow", func(t *testing.T) {
		if !strings.Contains(code, "rawMF") && !strings.Contains(code, "raw") {
			t.Error("Missing raw money flow in directional split")
		}
	})

	t.Run("volume_multiplication", func(t *testing.T) {
		if !strings.Contains(code, "Volume") {
			t.Error("Raw money flow should multiply by volume")
		}
	})
}

func TestMFIIndicatorBuilder_WindowSummationLogic(t *testing.T) {
	testCases := []struct {
		period int
	}{
		{1}, {2}, {14}, {100},
	}

	for _, tc := range testCases {
		context := NewTopLevelIndicatorContext()
		accessor := NewOHLCVFieldAccessGenerator("Close")
		builder := NewMFIIndicatorBuilder("mfi", NewConstantPeriod(tc.period), accessor, context)
		code := builder.Build()

		t.Run("forward_summation_loop", func(t *testing.T) {
			if !strings.Contains(code, "for j := 0; j <") {
				t.Errorf("Period %d missing forward summation loop", tc.period)
			}
		})

		t.Run("loop_boundary", func(t *testing.T) {
			if !strings.Contains(code, "j++") {
				t.Errorf("Period %d missing loop increment", tc.period)
			}
		})

		t.Run("accumulation_variables", func(t *testing.T) {
			if !strings.Contains(code, "Sum") && !strings.Contains(code, "sum") {
				t.Errorf("Period %d missing sum variables", tc.period)
			}
		})

		t.Run("series_get_offset", func(t *testing.T) {
			if !strings.Contains(code, "Get(j)") {
				t.Errorf("Period %d should use Get(offset) for window access", tc.period)
			}
		})

		t.Run("both_directions_summed", func(t *testing.T) {
			sumCount := strings.Count(code, "Sum")
			if sumCount < 2 {
				t.Errorf("Period %d should sum both positive and negative flows (found %d Sum variables)", tc.period, sumCount)
			}
		})
	}
}

func TestMFIIndicatorBuilder_EdgeCaseHandling(t *testing.T) {
	t.Run("period_one", func(t *testing.T) {
		context := NewTopLevelIndicatorContext()
		accessor := NewOHLCVFieldAccessGenerator("Close")
		builder := NewMFIIndicatorBuilder("test", NewConstantPeriod(1), accessor, context)
		code := builder.Build()

		if code == "" {
			t.Error("Period 1 should generate valid code")
		}

		if !strings.Contains(code, "for j := 0; j <") {
			t.Error("Period 1 should still have summation loop structure")
		}

		if !strings.Contains(code, "negSum == 0") {
			t.Error("Period 1 should still check zero denominator")
		}
	})

	t.Run("empty_varname", func(t *testing.T) {
		context := NewTopLevelIndicatorContext()
		accessor := NewOHLCVFieldAccessGenerator("Close")
		builder := NewMFIIndicatorBuilder("", NewConstantPeriod(14), accessor, context)
		code := builder.Build()

		if code == "" {
			t.Error("Empty varName should generate code")
		}
	})

	t.Run("complex_source_expression", func(t *testing.T) {
		context := NewTopLevelIndicatorContext()
		accessor := NewTrueRangeAccessGenerator()
		builder := NewMFIIndicatorBuilder("test", NewConstantPeriod(14), accessor, context)
		code := builder.Build()

		if code == "" {
			t.Error("Complex source should generate code")
		}
	})

	t.Run("zero_period", func(t *testing.T) {
		context := NewTopLevelIndicatorContext()
		accessor := NewOHLCVFieldAccessGenerator("Close")
		builder := NewMFIIndicatorBuilder("test", NewConstantPeriod(0), accessor, context)
		code := builder.Build()

		_ = code
	})

	t.Run("negative_period", func(t *testing.T) {
		context := NewTopLevelIndicatorContext()
		accessor := NewOHLCVFieldAccessGenerator("Close")
		builder := NewMFIIndicatorBuilder("test", NewConstantPeriod(-14), accessor, context)
		code := builder.Build()

		_ = code
	})

	t.Run("very_large_period", func(t *testing.T) {
		context := NewTopLevelIndicatorContext()
		accessor := NewOHLCVFieldAccessGenerator("Close")
		builder := NewMFIIndicatorBuilder("test", NewConstantPeriod(5000), accessor, context)
		code := builder.Build()

		if code == "" {
			t.Error("Large period should generate code")
		}

		if !strings.Contains(code, "if ctx.BarIndex <") {
			t.Error("Large period should have warmup check")
		}
	})
}

func TestMFIIndicatorBuilder_MFIFormulaAccuracy(t *testing.T) {
	context := NewTopLevelIndicatorContext()
	accessor := NewOHLCVFieldAccessGenerator("Close")
	builder := NewMFIIndicatorBuilder("test", NewConstantPeriod(14), accessor, context)
	code := builder.Build()

	t.Run("money_flow_ratio_calculation", func(t *testing.T) {
		if !strings.Contains(code, "mfr") && !strings.Contains(code, "ratio") {
			t.Error("Missing money flow ratio variable")
		}

		if !strings.Contains(code, "posSum") && !strings.Contains(code, "Sum") {
			t.Error("MFR should use sum of positive flows")
		}

		if !strings.Contains(code, "negSum") && !strings.Contains(code, "Sum") {
			t.Error("MFR should use sum of negative flows")
		}
	})

	t.Run("mfi_scaling", func(t *testing.T) {
		if !strings.Contains(code, "100") {
			t.Error("MFI formula should use constant 100")
		}
	})

	t.Run("mfi_calculation", func(t *testing.T) {
		if !strings.Contains(code, "mfi") {
			t.Error("Missing mfi variable")
		}

		if !strings.Contains(code, "1.0 + mfr") {
			t.Error("MFI formula should use 1.0 + mfr pattern")
		}
	})

	t.Run("zero_denominator_result", func(t *testing.T) {
		if !strings.Contains(code, "negSum == 0") {
			t.Error("Missing zero-denominator check")
		}

		lines := strings.Split(code, "\n")
		foundCheck := false
		foundUpdate := false
		for i, line := range lines {
			if strings.Contains(line, "negSum == 0") {
				foundCheck = true
				if i+1 < len(lines) && strings.Contains(lines[i+1], "100") {
					foundUpdate = true
					break
				}
			}
		}

		if !foundCheck {
			t.Error("Missing negSum == 0 check")
		}
		if !foundUpdate {
			t.Error("Missing assignment of 100 after zero-denom check")
		}
	})
}

func TestMFIIndicatorBuilder_ZeroChangeExplicitlyZeroInBothFlows(t *testing.T) {
	cases := []struct {
		name   string
		period int
		src    AccessGenerator
	}{
		{"period_1_close", 1, NewOHLCVFieldAccessGenerator("Close")},
		{"period_14_close", 14, NewOHLCVFieldAccessGenerator("Close")},
		{"period_14_high", 14, NewOHLCVFieldAccessGenerator("High")},
		{"period_14_truerange", 14, NewTrueRangeAccessGenerator()},
		{"period_100_close", 100, NewOHLCVFieldAccessGenerator("Close")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := NewTopLevelIndicatorContext()
			builder := NewMFIIndicatorBuilder("mfi", NewConstantPeriod(tc.period), tc.src, ctx)
			code := builder.Build()

			if got := strings.Count(code, "= rawMF"); got != 2 {
				t.Errorf("rawMF must appear in exactly 2 branches (pos + neg), got %d", got)
			}

			if got := strings.Count(code, "posMF = 0.0"); got < 3 {
				t.Errorf("posMF must be zeroed in at least 3 branches (NaN+neg+zero), got %d", got)
			}
			if got := strings.Count(code, "negMF = 0.0"); got < 3 {
				t.Errorf("negMF must be zeroed in at least 3 branches (NaN+pos+zero), got %d", got)
			}
		})
	}
}
