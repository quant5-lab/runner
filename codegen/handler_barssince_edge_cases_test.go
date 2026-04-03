package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestBarsSinceHandler_EdgeCases(t *testing.T) {
	handler := &BarsSinceHandler{}
	gen := newTestGenerator()

	t.Run("empty_arguments", func(t *testing.T) {
		call := &ast.CallExpression{Arguments: []ast.Expression{}}
		_, err := handler.GenerateCode(gen, "test", call)
		if err == nil {
			t.Error("Expected error for empty arguments")
		}
	})

	t.Run("extra_arguments_accepted", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "condition"},
				&ast.Literal{Value: 42},
			},
		}
		_, err := handler.GenerateCode(gen, "test", call)
		if err != nil {
			t.Errorf("Handler accepts extra arguments, got error: %v", err)
		}
	})

	t.Run("empty_variable_name", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{&ast.Identifier{Name: "condition"}},
		}
		code, err := handler.GenerateCode(gen, "", call)
		if err != nil {
			t.Fatalf("Should handle empty varName, got error: %v", err)
		}
		if code == "" {
			t.Error("Should generate code even with empty varName")
		}
	})
}

/* TestBarsSinceHandler_AlgorithmCorrectness validates algorithmic properties
 *
 * Tests that barssince follows correct counter algorithm:
 * - Three branches: condition true, not first bar, first bar
 * - Series-based state (no external variables)
 * - Self-reference for previous counter (Get(1))
 * - Counter reset when condition true
 * - Counter increment when condition false
 * - NaN on first bar when condition false
 */
func TestBarsSinceHandler_AlgorithmCorrectness(t *testing.T) {
	handler := &BarsSinceHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{
		Arguments: []ast.Expression{&ast.Identifier{Name: "signal"}},
	}

	code, err := handler.GenerateCode(gen, "counter", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("three_branch_structure", func(t *testing.T) {
		/* Branch 1: condition true → reset to 0 */
		if !strings.Contains(code, "if value.IsTrue(") {
			t.Error("Missing condition true branch check")
		}
		if !strings.Contains(code, "Series.Set(0") {
			t.Error("Missing counter reset to 0")
		}

		/* Branch 2: not first bar AND previous not NaN → increment */
		if !strings.Contains(code, "else if i > 0") {
			t.Error("Missing bar index boundary check (uses i > 0, not ctx.BarIndex > 0)")
		}
		if !strings.Contains(code, "!math.IsNaN") && !strings.Contains(code, "Get(1)") {
			t.Error("Missing NaN check for previous value")
		}
		if !strings.Contains(code, "Get(1) + 1") {
			t.Error("Missing counter increment logic")
		}

		/* Branch 3: first bar OR previous was NaN → NaN */
		if !strings.Contains(code, "} else {") {
			t.Error("Missing else branch for NaN assignment")
		}
		if !strings.Contains(code, "Series.Set(math.NaN())") {
			t.Error("Missing NaN assignment in else branch")
		}
	})

	t.Run("series_based_state", func(t *testing.T) {
		/* Must use Get(1) for previous value, not external variables */
		if !strings.Contains(code, "counterSeries.Get(1)") {
			t.Error("Missing series-based previous value retrieval")
		}

		/* No external state variables */
		if strings.Contains(code, "var previousCount") {
			t.Error("Should not declare external state variables")
		}
	})

	t.Run("self_reference_pattern", func(t *testing.T) {
		if !strings.Contains(code, "Series.Get(1)") {
			t.Error("Missing self-reference to previous counter value")
		}
	})

	t.Run("forward_only_logic", func(t *testing.T) {
		/* No loops - single bar calculation using previous bar state */
		if strings.Contains(code, "for ") {
			t.Error("barssince should not use loops - violates stateful principle")
		}
	})
}

/* TestBarsSinceHandler_ConditionTypes validates different condition expressions */
func TestBarsSinceHandler_ConditionTypes(t *testing.T) {
	handler := &BarsSinceHandler{}
	gen := newTestGenerator()

	testCases := []struct {
		name      string
		condition ast.Expression
	}{
		{
			name:      "identifier_condition",
			condition: &ast.Identifier{Name: "bullish"},
		},
		{
			name:      "member_expression",
			condition: &ast.MemberExpression{Object: &ast.Identifier{Name: "signal"}, Property: &ast.Identifier{Name: "cross"}},
		},
		{
			name:      "call_expression",
			condition: &ast.CallExpression{Callee: &ast.Identifier{Name: "ta.crossover"}},
		},
		{
			name: "binary_comparison",
			condition: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "close"},
				Operator: ">",
				Right:    &ast.Identifier{Name: "open"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			call := &ast.CallExpression{Arguments: []ast.Expression{tc.condition}}
			code, err := handler.GenerateCode(gen, "test", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			if !strings.Contains(code, "value.IsTrue(") {
				t.Error("Missing value.IsTrue check")
			}

			if !strings.Contains(code, "testSeries.Set(") {
				t.Error("Missing series assignment")
			}
		})
	}
}

func TestBarsSinceHandler_CodeStructureInvariants(t *testing.T) {
	handler := &BarsSinceHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{
		Arguments: []ast.Expression{&ast.Identifier{Name: "cond"}},
	}

	code, err := handler.GenerateCode(gen, "bs", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

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

	t.Run("balanced_braces", func(t *testing.T) {
		openCount := strings.Count(code, "{")
		closeCount := strings.Count(code, "}")
		if openCount != closeCount {
			t.Errorf("Unbalanced braces: %d open, %d close", openCount, closeCount)
		}
	})

	t.Run("consistent_naming", func(t *testing.T) {
		if !strings.Contains(code, "bsSeries") {
			t.Error("Missing consistent series variable naming")
		}
	})

	t.Run("code_completeness", func(t *testing.T) {
		if len(code) < 50 {
			t.Error("Generated code should be substantial")
		}
	})

	t.Run("imports_required_packages", func(t *testing.T) {
		if !strings.Contains(code, "math.IsNaN") {
			t.Error("Code should use math.IsNaN for NaN checks")
		}
	})
}

func TestBarsSinceHandler_IntegrationWithRegistry(t *testing.T) {
	registry := NewTAFunctionRegistry()

	t.Run("handler_registered", func(t *testing.T) {
		handler := registry.FindHandler("ta.barssince")
		if handler == nil {
			t.Fatal("barssince handler not registered in TAFunctionRegistry")
		}

		if _, ok := handler.(*BarsSinceHandler); !ok {
			t.Errorf("Wrong handler type: %T", handler)
		}
	})

	t.Run("supports_function_names", func(t *testing.T) {
		if !registry.IsSupported("ta.barssince") {
			t.Error("Registry should support 'ta.barssince'")
		}
		if !registry.IsSupported("barssince") {
			t.Error("Registry should support 'barssince' (Pine v4 syntax)")
		}
	})

	t.Run("generates_code_via_registry", func(t *testing.T) {
		gen := newTestGenerator()
		call := &ast.CallExpression{
			Arguments: []ast.Expression{&ast.Identifier{Name: "signal"}},
		}

		code, err := registry.GenerateInlineTA(gen, "test", "ta.barssince", call)
		if err != nil {
			t.Fatalf("GenerateInlineTA() error = %v", err)
		}

		if code == "" {
			t.Error("Registry should generate non-empty code")
		}

		if !strings.Contains(code, "testSeries.Set") {
			t.Error("Registry-generated code missing series assignment")
		}
	})
}
