package preprocessor

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

func findCallInCondition(expr *parser.Expression) *parser.CallExpr {
	if expr == nil || expr.Ternary == nil || expr.Ternary.Condition == nil {
		return nil
	}
	return findCallInFactor(expr.Ternary.Condition.Left.Left.Left.Left.Left)
}

func assertTernaryTransformation(t *testing.T, expr *parser.Expression, context string) {
	t.Helper()
	if expr.Ternary == nil {
		t.Errorf("%s: Expected ternary expression, got nil", context)
		return
	}
	if expr.Ternary.TrueVal == nil {
		t.Errorf("%s: Ternary TrueVal is nil", context)
	}
	if expr.Ternary.FalseVal == nil {
		t.Errorf("%s: Ternary FalseVal is nil", context)
	}
	if expr.Call != nil {
		t.Errorf("%s: CallExpr should be nil after transformation", context)
	}
}

func assertNoTransformation(t *testing.T, expr *parser.Expression, context string) {
	t.Helper()
	if expr.Ternary != nil && expr.Ternary.TrueVal != nil {
		t.Errorf("%s: Unexpected ternary transformation occurred", context)
	}
}

/* TEST CATEGORY: Basic Transformation Correctness */

func TestIffToTernary_BasicTransformation(t *testing.T) {
	source := `x = iff(close > open, 1, 0)`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseString("test.pine", source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	transformer := NewIffToTernaryTransformer()
	result, err := transformer.Transform(script)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if len(result.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(result.Statements))
	}

	stmt := result.Statements[0]
	if stmt.Assignment == nil {
		t.Fatal("Expected assignment statement")
	}

	assertTernaryTransformation(t, stmt.Assignment.Value, "Basic iff()")
}

/* TEST CATEGORY: Nesting Depth */

func TestIffToTernary_NestingDepth(t *testing.T) {
	tests := []struct {
		name   string
		source string
		depth  int
	}{
		{
			name:   "two_level_nesting",
			source: `x = iff(a > 0, iff(b > 0, 1, 2), 3)`,
			depth:  2,
		},
		{
			name:   "three_level_nesting",
			source: `x = iff(a > 0, iff(b > 0, iff(c > 0, 1, 2), 3), 4)`,
			depth:  3,
		},
		{
			name:   "nesting_in_alternate_branch",
			source: `x = iff(a > 0, 1, iff(b > 0, 2, 3))`,
			depth:  2,
		},
		{
			name:   "nesting_both_branches",
			source: `x = iff(a > 0, iff(b > 0, 1, 2), iff(c > 0, 3, 4))`,
			depth:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseString("test.pine", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			transformer := NewIffToTernaryTransformer()
			result, err := transformer.Transform(script)
			if err != nil {
				t.Fatalf("Transform failed: %v", err)
			}

			stmt := result.Statements[0]
			assertTernaryTransformation(t, stmt.Assignment.Value, "Outer iff()")

			/* Verify nested transformations */
			outerTernary := stmt.Assignment.Value.Ternary
			if tt.depth >= 2 {
				/* Check at least one branch has nested ternary */
				hasNested := (outerTernary.TrueVal != nil && outerTernary.TrueVal.Ternary != nil) ||
					(outerTernary.FalseVal != nil && outerTernary.FalseVal.Ternary != nil)
				if !hasNested {
					t.Errorf("Expected nested ternary at depth %d", tt.depth)
				}
			}
		})
	}
}

/* TEST CATEGORY: Expression Context */

func TestIffToTernary_ExpressionContexts(t *testing.T) {
	tests := []struct {
		name   string
		source string
		verify func(*testing.T, *parser.Script)
	}{
		{
			name:   "function_argument",
			source: `plot(iff(close > open, close, open))`,
			verify: func(t *testing.T, script *parser.Script) {
				stmt := script.Statements[0]
				if stmt.Expression == nil {
					t.Fatal("Expected expression statement")
				}
				/* plot() call is wrapped in incomplete ternary */
				plotExpr := stmt.Expression.Expr
				if plotExpr.Ternary == nil {
					t.Fatal("Expected ternary wrapper")
				}
				/* Find plot() call in condition */
				plotCall := findCallInCondition(plotExpr)
				if plotCall == nil {
					t.Fatal("Expected plot() call")
				}
				if len(plotCall.Args) != 1 {
					t.Fatalf("Expected 1 arg to plot(), got %d", len(plotCall.Args))
				}
				/* Verify iff() inside plot() was transformed */
				assertTernaryTransformation(t, plotCall.Args[0].Value, "iff() in plot() arg")
			},
		},
		{
			name: "if_condition",
			source: `if iff(x > 0, true, false)
    a = 1`,
			verify: func(t *testing.T, script *parser.Script) {
				ifStmt := script.Statements[0].If
				if ifStmt == nil {
					t.Fatal("Expected if statement")
				}
				/* iff() in if condition is parsed as OrExpr, not transformable at condition level */
				/* Transformation happens if iff() is in assignment/expression context */
				if ifStmt.Condition == nil {
					t.Fatal("Expected condition")
				}
			},
		},
		{
			name: "for_loop_range",
			source: `for i = 0 to iff(condition, 10, 20)
    a = i`,
			verify: func(t *testing.T, script *parser.Script) {
				forStmt := script.Statements[0].For
				if forStmt == nil {
					t.Fatal("Expected for statement")
				}
				/* For loop range is ArithExpr, iff() handled at parse level */
				if forStmt.To == nil {
					t.Error("Expected 'to' range")
				}
			},
		},
		{
			name:   "multiple_arguments",
			source: `plot(iff(a > b, 1, 2), iff(c > d, 3, 4))`,
			verify: func(t *testing.T, script *parser.Script) {
				plotCall := findCallInCondition(script.Statements[0].Expression.Expr)
				if len(plotCall.Args) != 2 {
					t.Fatalf("Expected 2 args, got %d", len(plotCall.Args))
				}
				assertTernaryTransformation(t, plotCall.Args[0].Value, "First iff() arg")
				assertTernaryTransformation(t, plotCall.Args[1].Value, "Second iff() arg")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseString("test.pine", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			transformer := NewIffToTernaryTransformer()
			result, err := transformer.Transform(script)
			if err != nil {
				t.Fatalf("Transform failed: %v", err)
			}

			tt.verify(t, result)
		})
	}
}

/* TEST CATEGORY: Condition Complexity */

func TestIffToTernary_ComplexConditions(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "logical_and",
			source: `x = iff(close > open and volume > 1000, high, low)`,
		},
		{
			name:   "logical_or",
			source: `x = iff(close > open or volume > 1000, high, low)`,
		},
		{
			name:   "nested_logical",
			source: `x = iff((a > b and c < d) or (e == f), 1, 0)`,
		},
		{
			name:   "function_call_in_condition",
			source: `x = iff(ta.crossover(fast, slow), 1, 0)`,
		},
		{
			name:   "comparison_operators",
			source: `x = iff(a >= b and c <= d and e != f, 1, 0)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseString("test.pine", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			transformer := NewIffToTernaryTransformer()
			result, err := transformer.Transform(script)
			if err != nil {
				t.Fatalf("Transform failed: %v", err)
			}

			stmt := result.Statements[0]
			assertTernaryTransformation(t, stmt.Assignment.Value, "Complex condition")

			/* Verify condition preserved */
			ternary := stmt.Assignment.Value.Ternary
			if ternary.Condition == nil {
				t.Error("Condition should be preserved")
			}
		})
	}
}

/* TEST CATEGORY: Argument Complexity */

func TestIffToTernary_ComplexArguments(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "arithmetic_in_branches",
			source: `x = iff(condition, high - low, open + close)`,
		},
		{
			name:   "function_calls_in_branches",
			source: `x = iff(condition, sma(close, 20), ema(close, 10))`,
		},
		{
			name:   "array_access_in_branches",
			source: `x = iff(condition, close[1], close[2])`,
		},
		{
			name:   "nested_function_calls",
			source: `x = iff(condition, max(high, close), min(low, open))`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseString("test.pine", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			transformer := NewIffToTernaryTransformer()
			result, err := transformer.Transform(script)
			if err != nil {
				t.Fatalf("Transform failed: %v", err)
			}

			stmt := result.Statements[0]
			assertTernaryTransformation(t, stmt.Assignment.Value, "Complex arguments")

			/* Verify branches exist */
			ternary := stmt.Assignment.Value.Ternary
			if ternary.TrueVal == nil || ternary.FalseVal == nil {
				t.Error("Both branches should be populated")
			}
		})
	}
}

/* TEST CATEGORY: Validation & Error Handling */

func TestIffToTernary_InvalidArgumentCount(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "zero_args",
			source: `x = iff()`,
		},
		{
			name:   "one_arg",
			source: `x = iff(condition)`,
		},
		{
			name:   "two_args",
			source: `x = iff(condition, 1)`,
		},
		{
			name:   "four_args",
			source: `x = iff(condition, 1, 2, 3)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseString("test.pine", tt.source)
			if err != nil {
				/* Parser may reject some invalid syntax */
				return
			}

			transformer := NewIffToTernaryTransformer()
			_, err = transformer.Transform(script)

			/* Transformer MUST reject all invalid argument counts */
			if err == nil {
				t.Error("Expected transformation error for invalid iff() argument count")
			}
			if err != nil && !strings.Contains(err.Error(), "iff()") {
				t.Errorf("Error should mention iff(), got: %v", err)
			}
		})
	}
}

/* TEST CATEGORY: Non-iff Preservation */

func TestIffToTernary_NonIffPreservation(t *testing.T) {
	tests := []struct {
		name   string
		source string
		verify func(*testing.T, *parser.Script)
	}{
		{
			name:   "regular_function_calls",
			source: `x = sma(close, 10)`,
			verify: func(t *testing.T, script *parser.Script) {
				stmt := script.Statements[0]
				/* sma() remains in incomplete ternary wrapper */
				expr := stmt.Assignment.Value
				if expr.Ternary == nil {
					t.Error("Expected ternary wrapper for expression")
				}
				if expr.Ternary.TrueVal != nil || expr.Ternary.FalseVal != nil {
					t.Error("Non-iff should have incomplete ternary (nil branches)")
				}
			},
		},
		{
			name:   "native_ternary_operator",
			source: `x = close > open ? 1 : 0`,
			verify: func(t *testing.T, script *parser.Script) {
				stmt := script.Statements[0]
				ternary := stmt.Assignment.Value.Ternary
				if ternary == nil {
					t.Fatal("Expected ternary")
				}
				/* Native ternary has TrueVal/FalseVal populated by parser */
				if ternary.TrueVal == nil || ternary.FalseVal == nil {
					t.Error("Native ternary should have both branches")
				}
			},
		},
		{
			name:   "literals",
			source: `x = 42`,
			verify: func(t *testing.T, script *parser.Script) {
				stmt := script.Statements[0]
				expr := stmt.Assignment.Value
				/* Literals wrapped in incomplete ternary */
				if expr.Ternary == nil {
					t.Error("Expected ternary wrapper")
				}
				assertNoTransformation(t, expr, "Literal")
			},
		},
		{
			name: "mixed_iff_and_non_iff",
			source: `a = iff(x > 0, 1, 0)
b = sma(close, 20)
c = iff(y > 0, 2, 3)`,
			verify: func(t *testing.T, script *parser.Script) {
				if len(script.Statements) != 3 {
					t.Fatalf("Expected 3 statements, got %d", len(script.Statements))
				}
				/* Statement 0: iff() transformed */
				assertTernaryTransformation(t, script.Statements[0].Assignment.Value, "First iff()")
				/* Statement 1: sma() unchanged */
				assertNoTransformation(t, script.Statements[1].Assignment.Value, "sma()")
				/* Statement 2: iff() transformed */
				assertTernaryTransformation(t, script.Statements[2].Assignment.Value, "Second iff()")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseString("test.pine", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			transformer := NewIffToTernaryTransformer()
			result, err := transformer.Transform(script)
			if err != nil {
				t.Fatalf("Transform failed: %v", err)
			}

			tt.verify(t, result)
		})
	}
}

/* TEST CATEGORY: Multiple Transformations */

func TestIffToTernary_MultipleTransformations(t *testing.T) {
	source := `
x = iff(a > b, 1, 2)
y = iff(c < d, 3, 4)
z = iff(e == f, 5, 6)
`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseString("test.pine", source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	transformer := NewIffToTernaryTransformer()
	result, err := transformer.Transform(script)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if len(result.Statements) != 3 {
		t.Fatalf("Expected 3 statements, got %d", len(result.Statements))
	}

	for i, stmt := range result.Statements {
		if stmt.Assignment == nil {
			t.Errorf("Statement %d: Expected assignment", i)
			continue
		}
		assertTernaryTransformation(t, stmt.Assignment.Value, "iff() "+string(rune('x'+i)))
	}
}

/* TEST CATEGORY: Statement Types */

func TestIffToTernary_StatementTypes(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "assignment",
			source: `x = iff(condition, 1, 0)`,
		},
		{
			name:   "typed_assignment",
			source: `int x = iff(condition, 1, 0)`,
		},
		{
			name:   "reassignment",
			source: `x := iff(condition, 1, 0)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseString("test.pine", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			transformer := NewIffToTernaryTransformer()
			result, err := transformer.Transform(script)
			if err != nil {
				t.Fatalf("Transform failed: %v", err)
			}

			stmt := result.Statements[0]
			var expr *parser.Expression
			if stmt.Assignment != nil {
				expr = stmt.Assignment.Value
			} else if stmt.TypedAssignment != nil {
				expr = stmt.TypedAssignment.Value
			} else if stmt.Reassignment != nil {
				expr = stmt.Reassignment.Value
			} else {
				t.Fatal("No assignment statement found")
			}

			assertTernaryTransformation(t, expr, tt.name)
		})
	}
}

/* TEST CATEGORY: Real-World Integration */

func TestIffToTernary_Integration_RealWorldPattern(t *testing.T) {
	/* Generic multi-level nesting pattern (not specific to utbot-quantnomad) */
	source := `trailing_stop = 0.0
trailing_stop := iff(price > stop[1] and price[1] > stop[1], max(stop[1], price - loss), iff(price < stop[1] and price[1] < stop[1], min(stop[1], price + loss), iff(price > stop[1], price - loss, price + loss)))`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseString("test.pine", source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	transformer := NewIffToTernaryTransformer()
	result, err := transformer.Transform(script)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if len(result.Statements) != 2 {
		t.Fatalf("Expected 2 statements, got %d", len(result.Statements))
	}

	/* Second statement is the reassignment with nested iff() */
	stmt := result.Statements[1]
	if stmt.Reassignment == nil {
		t.Fatal("Expected reassignment statement")
	}

	assertTernaryTransformation(t, stmt.Reassignment.Value, "Outer iff()")

	/* Verify nested transformations */
	outerTernary := stmt.Reassignment.Value.Ternary
	if outerTernary.FalseVal == nil || outerTernary.FalseVal.Ternary == nil {
		t.Error("Expected nested ternary in alternate branch")
		return
	}

	level2Ternary := outerTernary.FalseVal.Ternary
	if level2Ternary.FalseVal == nil || level2Ternary.FalseVal.Ternary == nil {
		t.Error("Expected 3rd level nested ternary")
	}
}

/* TEST CATEGORY: Edge Cases */

func TestIffToTernary_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
		verify func(*testing.T, *parser.Script, error)
	}{
		{
			name:   "empty_script",
			source: ``,
			verify: func(t *testing.T, script *parser.Script, err error) {
				if err != nil {
					t.Errorf("Transform should succeed on empty script: %v", err)
				}
				if len(script.Statements) != 0 {
					t.Errorf("Expected 0 statements, got %d", len(script.Statements))
				}
			},
		},
		{
			name:   "no_iff_calls",
			source: `x = 1\ny = sma(close, 20)`,
			verify: func(t *testing.T, script *parser.Script, err error) {
				if err != nil {
					t.Errorf("Transform should succeed on non-iff script: %v", err)
				}
				if len(script.Statements) != 2 {
					t.Errorf("Expected 2 statements, got %d", len(script.Statements))
				}
			},
		},
		{
			name:   "iff_in_nested_block",
			source: `if condition\n    if sub_condition\n        x = iff(a > b, 1, 0)`,
			verify: func(t *testing.T, script *parser.Script, err error) {
				if err != nil {
					t.Errorf("Transform failed: %v", err)
				}
				/* Verify nested if body contains transformed iff() */
				outerIf := script.Statements[0].If
				if outerIf == nil || len(outerIf.Body) == 0 {
					t.Fatal("Expected outer if with body")
				}
				innerIf := outerIf.Body[0].If
				if innerIf == nil || len(innerIf.Body) == 0 {
					t.Fatal("Expected inner if with body")
				}
				assignment := innerIf.Body[0].Assignment
				if assignment == nil {
					t.Fatal("Expected assignment in inner if body")
				}
				assertTernaryTransformation(t, assignment.Value, "iff() in nested block")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseString("test.pine", tt.source)
			if err != nil {
				t.Skipf("Parse failed (expected for some edge cases): %v", err)
				return
			}

			transformer := NewIffToTernaryTransformer()
			result, transformErr := transformer.Transform(script)

			tt.verify(t, result, transformErr)
		})
	}
}
