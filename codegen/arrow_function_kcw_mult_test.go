package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

/* TestArrowFunctionKCW_MultParameterHandling verifies KCW mult argument handling in arrow function context */
func TestArrowFunctionKCW_MultParameterHandling(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		mustContain    []string
		mustNotContain []string
	}{
		{
			name: "mult_as_runtime_parameter_uses_identifier",
			/* Period must be a constant literal; KCW returns math.NaN() for dynamic period. */
			source: `
calcKCW(src, mult) => ta.kcw(src, 20, mult)
result = calcKCW(close, 3.0)
`,
			mustContain: []string{
				"2.0 * mult *",
			},
			mustNotContain: []string{
				"2.0 * 1.5",
				"2.0 * 3",
			},
		},
		{
			name: "mult_as_literal_constant_inlines_value",
			source: `
calcKCW(src) => ta.kcw(src, 20, 2.0)
result = calcKCW(close)
`,
			mustContain: []string{
				"2.0 * 2",
			},
			mustNotContain: []string{
				"2.0 * mult",
			},
		},
		{
			name: "implicit_source_form_extracts_mult_from_arg1",
			/* Verifies mult index selection when source is implicit: period must be a constant literal. */
			source: `
calcKCW() => ta.kcw(20, 1.5)
result = calcKCW()
`,
			mustContain: []string{
				"1.5",
				"ctx.Data",
			},
			mustNotContain: []string{
				"srcSeries",
			},
		},
		{
			name: "fractional_mult_literal_preserved",
			source: `
calcKCW(src) => ta.kcw(src, 14, 2.5)
result = calcKCW(close)
`,
			mustContain: []string{
				"2.0 * 2.5",
			},
			mustNotContain: []string{
				"2.0 * 2 ",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("parser creation failed: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("conversion failed: %v", err)
			}

			result, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("codegen failed: %v", err)
			}

			code := result.UserDefinedFunctions

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("missing pattern %q\nArrow functions:\n%s", pattern, code)
				}
			}

			for _, pattern := range tt.mustNotContain {
				if strings.Contains(code, pattern) {
					t.Errorf("must not contain pattern %q\nArrow functions:\n%s", pattern, code)
				}
			}
		})
	}
}

/* TestArrowFunctionKCW_FormulaStructure verifies KCW formula components in generated arrow code */
func TestArrowFunctionKCW_FormulaStructure(t *testing.T) {
	/* Period must be a constant literal for KCW to generate stateful code in arrow context. */
	source := `
width(src) => ta.kcw(src, 20, 1.5)
result = width(close)
`
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("parser creation failed: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(source))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	converter := parser.NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("conversion failed: %v", err)
	}

	result, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}

	code := result.UserDefinedFunctions

	t.Run("ema_component", func(t *testing.T) {
		if !strings.Contains(code, "ema") {
			t.Error("KCW arrow code must compute EMA component")
		}
	})

	t.Run("atr_component", func(t *testing.T) {
		if !strings.Contains(code, "atr") {
			t.Error("KCW arrow code must compute ATR component")
		}
	})

	t.Run("bandwidth_factor", func(t *testing.T) {
		if !strings.Contains(code, "2.0 *") {
			t.Error("KCW formula must include 2.0 * mult factor")
		}
	})

	t.Run("zero_ema_guard", func(t *testing.T) {
		if !strings.Contains(code, "== 0") {
			t.Error("KCW must guard against zero EMA in generated code")
		}
	})

	t.Run("warmup_nan", func(t *testing.T) {
		if !strings.Contains(code, "math.NaN()") {
			t.Error("KCW must return NaN during warmup")
		}
	})
}

/* TestArrowFunctionKCW_DynamicPeriodFallsBackToNaN verifies KCW returns math.NaN() when period is runtime dynamic */
func TestArrowFunctionKCW_DynamicPeriodFallsBackToNaN(t *testing.T) {
	/* KCW uses stateful EMA/RMA series whose size must be known at codegen time.
	   A dynamic period (arrow function parameter) prevents this, so KCWIIFEGenerator
	   returns math.NaN() to signal the unsupported configuration. */
	source := `
calcKCW(src, len) => ta.kcw(src, len, 1.5)
result = calcKCW(close, 20)
`
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("parser creation failed: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(source))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	converter := parser.NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("conversion failed: %v", err)
	}

	result, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}

	code := result.UserDefinedFunctions
	if !strings.Contains(code, "math.NaN()") {
		t.Errorf("KCW with dynamic period must return math.NaN() in arrow context\nArrow functions:\n%s", code)
	}
}

/* TestArrowFunctionKCW_EdgeCases verifies boundary conditions */
func TestArrowFunctionKCW_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		mustContain []string
	}{
		{
			name: "multiple_kcw_calls_with_different_mults",
			/* Same source+period with different mults must produce distinct series and formulas. */
			source: `
compare(src) =>
    narrow = ta.kcw(src, 20, 1.0)
    wide = ta.kcw(src, 20, 3.0)
    [narrow, wide]
[n, w] = compare(close)
`,
			mustContain: []string{
				"2.0 * 1",
				"2.0 * 3",
			},
		},
		{
			name: "bare_alias_routes_through_generate_kcw_call",
			/* Bare alias must honour the same mult extraction path as the ta.kcw-prefixed form. */
			source: `
calcWidth(src) => kcw(src, 14, 1.5)
result = calcWidth(close)
`,
			mustContain: []string{
				"1.5",
				"srcSeries *series.Series",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("parser creation failed: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("conversion failed: %v", err)
			}

			result, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("codegen failed: %v", err)
			}

			code := result.UserDefinedFunctions

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("missing pattern %q\nFull code:\n%s", pattern, code)
				}
			}
		})
	}
}
