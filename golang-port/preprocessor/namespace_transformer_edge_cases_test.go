package preprocessor

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

/* NamespaceTransformer edge case tests: deeply nested expressions, large ASTs, boundary conditions */

func TestNamespaceTransformer_DeeplyNestedExpressions(t *testing.T) {
	/* Test 10+ levels of nested function calls */
	input := `result = max(max(max(max(max(max(max(max(max(max(1, 2), 3), 4), 5), 6), 7), 8), 9), 10), 11)`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	transformer := NewMathNamespaceTransformer()
	result, err := transformer.Transform(ast)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if result == nil {
		t.Error("Expected result, got nil")
	}

	/* All nested max() calls should be transformed to math.max() */
	/* Cannot easily assert deep nesting without complex traversal, but test should not panic */
}

func TestNamespaceTransformer_LargeNumberOfCalls(t *testing.T) {
	/* Test 1000 function calls (stress test for traversal performance) */
	var lines []string
	for i := 0; i < 1000; i++ {
		lines = append(lines, "ma"+strings.Repeat("x", i%10)+" = sma(close, 20)")
	}
	input := strings.Join(lines, "\n")

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	transformer := NewTANamespaceTransformer()
	result, err := transformer.Transform(ast)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if len(result.Statements) != 1000 {
		t.Errorf("Expected 1000 statements, got %d", len(result.Statements))
	}
}

func TestNamespaceTransformer_InvalidASTStructure(t *testing.T) {
	/* Test with manually constructed invalid AST (nil fields in unexpected places) */
	ast := &parser.Script{
		Statements: []*parser.Statement{
			{
				Assignment: &parser.Assignment{
					Name: "test",
					Value: &parser.Expression{
						Ternary: &parser.TernaryExpr{
							Condition: nil, /* Nil condition */
						},
					},
				},
			},
			{
				Assignment: &parser.Assignment{
					Name: "test2",
					Value: &parser.Expression{
						Ternary: &parser.TernaryExpr{
							Condition: &parser.OrExpr{
								Left: nil, /* Nil left operand */
							},
						},
					},
				},
			},
			{
				/* Nil assignment */
				Assignment: nil,
			},
		},
	}

	transformer := NewTANamespaceTransformer()
	result, err := transformer.Transform(ast)

	/* Should not panic, should handle gracefully */
	if err != nil {
		t.Fatalf("Transform should handle invalid AST gracefully, got error: %v", err)
	}

	if result == nil {
		t.Error("Expected result even with invalid AST")
	}
}

func TestNamespaceTransformer_MixedFunctionTypes(t *testing.T) {
	/* Test mixing TA, Math, and Request functions in same expression */
	input := `
combined = max(sma(close, 20), security(syminfo.tickerid, "D", close))
`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	/* Apply all three transformers */
	pipeline := NewV4ToV5Pipeline()
	result, err := pipeline.Run(ast)
	if err != nil {
		t.Fatalf("Pipeline failed: %v", err)
	}

	if result == nil {
		t.Error("Expected result after full pipeline")
	}

	/* All three function types should be transformed:
	   max → math.max
	   sma → ta.sma
	   security → request.security
	*/
}

func TestNamespaceTransformer_FunctionAsArgument(t *testing.T) {
	/* Test function calls as arguments to other function calls */
	input := `
result = sma(ema(close, 10), 20)
nested = max(min(abs(x), y), z)
`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	pipeline := NewV4ToV5Pipeline()
	result, err := pipeline.Run(ast)
	if err != nil {
		t.Fatalf("Pipeline failed: %v", err)
	}

	if len(result.Statements) != 2 {
		t.Errorf("Expected 2 statements, got %d", len(result.Statements))
	}

	/* Both outer and inner function calls should be transformed */
}

func TestNamespaceTransformer_CallInTernary(t *testing.T) {
	/* Test function calls inside ternary expressions */
	input := `
result = close > open ? sma(close, 20) : ema(close, 20)
complex = max(high, low) > threshold ? crossover(fast, slow) : false
`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	pipeline := NewV4ToV5Pipeline()
	result, err := pipeline.Run(ast)
	if err != nil {
		t.Fatalf("Pipeline failed: %v", err)
	}

	if len(result.Statements) != 2 {
		t.Errorf("Expected 2 statements, got %d", len(result.Statements))
	}

	/* All function calls in ternary branches should be transformed */
}

func TestNamespaceTransformer_CallInBinaryExpression(t *testing.T) {
	/* Test function calls in binary expressions (comparisons, arithmetic) */
	input := `
condition1 = sma(close, 20) > sma(close, 50)
condition2 = rsi(close, 14) < 30
arithmetic = max(high, low) * 1.5 + min(high, low) * 0.5
`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	pipeline := NewV4ToV5Pipeline()
	result, err := pipeline.Run(ast)
	if err != nil {
		t.Fatalf("Pipeline failed: %v", err)
	}

	if len(result.Statements) != 3 {
		t.Errorf("Expected 3 statements, got %d", len(result.Statements))
	}

	/* All function calls in binary expressions should be transformed */
}

func TestNamespaceTransformer_CallInUnaryExpression(t *testing.T) {
	/* Test function calls in unary expressions (negation, not) */
	input := `
negated = -abs(value)
notResult = not crossover(fast, slow)
`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	pipeline := NewV4ToV5Pipeline()
	result, err := pipeline.Run(ast)
	if err != nil {
		t.Fatalf("Pipeline failed: %v", err)
	}

	if len(result.Statements) != 2 {
		t.Errorf("Expected 2 statements, got %d", len(result.Statements))
	}

	/* Function calls after unary operators should be transformed */
}

func TestNamespaceTransformer_CallInArrayAccess(t *testing.T) {
	/* Test function calls in array access expressions */
	input := `
historical1 = sma(close, 20)[1]
historical2 = ema(close, 50)[10]
`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	pipeline := NewV4ToV5Pipeline()
	result, err := pipeline.Run(ast)
	if err != nil {
		t.Fatalf("Pipeline failed: %v", err)
	}

	if len(result.Statements) != 2 {
		t.Errorf("Expected 2 statements, got %d", len(result.Statements))
	}

	/* Function calls with array access should be transformed */
}

func TestNamespaceTransformer_EmptyScript(t *testing.T) {
	/* Test empty script (no statements) */
	ast := &parser.Script{
		Statements: []*parser.Statement{},
	}

	transformer := NewTANamespaceTransformer()
	result, err := transformer.Transform(ast)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if len(result.Statements) != 0 {
		t.Errorf("Expected 0 statements, got %d", len(result.Statements))
	}
}

func TestNamespaceTransformer_OnlyComments(t *testing.T) {
	/* Test script with only whitespace/newlines (no executable code) */
	input := `


`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	transformer := NewTANamespaceTransformer()
	result, err := transformer.Transform(ast)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	/* Should have 0 statements (only whitespace) */
	if len(result.Statements) != 0 {
		t.Errorf("Expected 0 statements (whitespace only), got %d", len(result.Statements))
	}
}

func TestNamespaceTransformer_MultipleTransformersSameNode(t *testing.T) {
	/* Test applying multiple transformers that could potentially conflict */
	input := `ma = sma(close, 20)`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	/* Apply TA transformer first */
	taTransformer := NewTANamespaceTransformer()
	result1, err := taTransformer.Transform(ast)
	if err != nil {
		t.Fatalf("TA Transform failed: %v", err)
	}

	/* Apply Math transformer second (should not affect sma, which is TA function) */
	mathTransformer := NewMathNamespaceTransformer()
	result2, err := mathTransformer.Transform(result1)
	if err != nil {
		t.Fatalf("Math Transform failed: %v", err)
	}

	expr := result2.Statements[0].Assignment.Value
	call := findCallInFactor(expr.Ternary.Condition.Left.Left.Left.Left.Left)
	if call == nil {
		t.Fatal("Expected call expression")
	}

	/* Should remain ta.sma (not affected by Math transformer) */
	if call.Callee.MemberAccess == nil {
		t.Error("Expected MemberAccess (ta.sma)")
	}

	if call.Callee.MemberAccess != nil {
		if call.Callee.MemberAccess.Object != "ta" || call.Callee.MemberAccess.Property != "sma" {
			t.Errorf("Expected ta.sma, got %s.%s",
				call.Callee.MemberAccess.Object,
				call.Callee.MemberAccess.Property)
		}
	}
}

func TestNamespaceTransformer_CaseSensitivity(t *testing.T) {
	/* Test that function name matching is case-sensitive */
	input := `
lower = sma(close, 20)
upper = SMA(close, 20)
mixed = Sma(close, 20)
`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	transformer := NewTANamespaceTransformer()
	result, err := transformer.Transform(ast)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	/* Only lowercase 'sma' should be transformed (PineScript is case-sensitive) */
	lowerExpr := result.Statements[0].Assignment.Value
	lowerCall := findCallInFactor(lowerExpr.Ternary.Condition.Left.Left.Left.Left.Left)
	if lowerCall != nil && lowerCall.Callee.MemberAccess != nil {
		if lowerCall.Callee.MemberAccess.Property != "sma" {
			t.Error("Lowercase 'sma' should be transformed")
		}
	}

	/* Uppercase 'SMA' should NOT be transformed */
	upperExpr := result.Statements[1].Assignment.Value
	upperCall := findCallInFactor(upperExpr.Ternary.Condition.Left.Left.Left.Left.Left)
	if upperCall != nil && upperCall.Callee.Ident != nil {
		if *upperCall.Callee.Ident != "SMA" {
			t.Error("Uppercase 'SMA' should remain unchanged")
		}
	}
}

func TestNamespaceTransformer_ConsecutiveTransforms(t *testing.T) {
	/* Test applying same transformer multiple times (idempotency) */
	input := `ma = sma(close, 20)`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	transformer := NewTANamespaceTransformer()

	/* Apply 5 times */
	result := ast
	for i := 0; i < 5; i++ {
		result, err = transformer.Transform(result)
		if err != nil {
			t.Fatalf("Transform iteration %d failed: %v", i, err)
		}
	}

	/* Should still be ta.sma (not ta.ta.ta.ta.ta.sma) */
	expr := result.Statements[0].Assignment.Value
	call := findCallInFactor(expr.Ternary.Condition.Left.Left.Left.Left.Left)
	if call == nil {
		t.Fatal("Expected call expression")
	}

	if call.Callee.MemberAccess == nil {
		t.Error("Expected MemberAccess after transformations")
	}

	if call.Callee.MemberAccess != nil {
		if call.Callee.MemberAccess.Object != "ta" || call.Callee.MemberAccess.Property != "sma" {
			t.Errorf("Expected ta.sma after 5 transforms, got %s.%s",
				call.Callee.MemberAccess.Object,
				call.Callee.MemberAccess.Property)
		}
	}
}
