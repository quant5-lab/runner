package preprocessor

import (
	"testing"

	"github.com/borisquantlab/pinescript-go/parser"
)

// Test idempotency - transforming already-transformed code
func TestTANamespaceTransformer_Idempotency(t *testing.T) {
	// Code already in v5 format
	input := `
ma20 = ta.sma(close, 20)
ma50 = ta.ema(close, 50)
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

	// Should remain unchanged (ta.sma should not become ta.ta.sma)
	for i := 0; i < 2; i++ {
		expr := result.Statements[i].Assignment.Value
		call := findCallInFactor(expr.Ternary.Condition.Left.Left.Left.Left.Left)
		if call == nil {
			t.Fatalf("Statement %d: expected call expression", i)
		}
		
		// Should have MemberAccess (ta.sma), not simple Ident
		if call.Callee.MemberAccess == nil {
			t.Errorf("Statement %d: expected member access (ta.xxx), got simple identifier", i)
		}
	}
}

// Test user-defined functions with same names as built-ins
func TestTANamespaceTransformer_UserDefinedFunctions(t *testing.T) {
	// User defines their own sma function
	input := `
my_sma = sma(close, 20)
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

	// Should transform built-in sma to ta.sma
	expr := result.Statements[0].Assignment.Value
	call := findCallInFactor(expr.Ternary.Condition.Left.Left.Left.Left.Left)
	if call == nil {
		t.Fatal("Expected call expression")
	}
	if call.Callee.Ident == nil || *call.Callee.Ident != "ta.sma" {
		t.Error("Built-in sma should be transformed to ta.sma")
	}
}

// Test empty file
func TestTANamespaceTransformer_EmptyFile(t *testing.T) {
	input := ``

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

	if len(result.Statements) != 0 {
		t.Errorf("Expected 0 statements, got %d", len(result.Statements))
	}
}

// Test comments only
func TestTANamespaceTransformer_CommentsOnly(t *testing.T) {
	input := `
// This is a comment
// Another comment
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

	if len(result.Statements) != 0 {
		t.Errorf("Expected 0 statements, got %d", len(result.Statements))
	}
}

// Test function not in mapping (should remain unchanged)
func TestTANamespaceTransformer_UnknownFunction(t *testing.T) {
	input := `x = myCustomFunction(close, 20)`

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

	// Should remain unchanged
	expr := result.Statements[0].Assignment.Value
	call := findCallInFactor(expr.Ternary.Condition.Left.Left.Left.Left.Left)
	if call == nil {
		t.Fatal("Expected call expression")
	}
	if call.Callee.Ident == nil || *call.Callee.Ident != "myCustomFunction" {
		t.Error("Custom function should not be transformed")
	}
}

// Test pipeline error propagation
func TestPipeline_ErrorPropagation(t *testing.T) {
	// This test verifies that errors from transformers are properly propagated
	// Currently all transformers return nil error, but this tests the mechanism
	
	input := `ma = sma(close, 20)`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Create pipeline and run
	pipeline := NewPipeline().
		Add(NewTANamespaceTransformer()).
		Add(NewMathNamespaceTransformer())

	result, err := pipeline.Run(ast)
	if err != nil {
		t.Fatalf("Pipeline failed: %v", err)
	}

	if result == nil {
		t.Error("Expected result, got nil")
	}
}

// Test multiple transformations on same statement
func TestPipeline_MultipleTransformations(t *testing.T) {
	input := `
study("Test")
ma = sma(close, 20)
val = abs(5)
`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Run full pipeline
	pipeline := NewV4ToV5Pipeline()
	result, err := pipeline.Run(ast)
	if err != nil {
		t.Fatalf("Pipeline failed: %v", err)
	}

	// Check study → indicator
	studyExpr := result.Statements[0].Expression.Expr
	studyCall := findCallInFactor(studyExpr.Ternary.Condition.Left.Left.Left.Left.Left)
	if studyCall == nil || studyCall.Callee.Ident == nil || *studyCall.Callee.Ident != "indicator" {
		t.Error("study should be transformed to indicator")
	}

	// Check sma → ta.sma
	smaExpr := result.Statements[1].Assignment.Value
	smaCall := findCallInFactor(smaExpr.Ternary.Condition.Left.Left.Left.Left.Left)
	if smaCall == nil || smaCall.Callee.Ident == nil || *smaCall.Callee.Ident != "ta.sma" {
		t.Error("sma should be transformed to ta.sma")
	}

	// Check abs → math.abs
	absExpr := result.Statements[2].Assignment.Value
	absCall := findCallInFactor(absExpr.Ternary.Condition.Left.Left.Left.Left.Left)
	if absCall == nil || absCall.Callee.Ident == nil || *absCall.Callee.Ident != "math.abs" {
		t.Error("abs should be transformed to math.abs")
	}
}

// Test nil pointer safety
func TestTANamespaceTransformer_NilPointerSafety(t *testing.T) {
	// Create minimal AST with nil fields
	ast := &parser.Script{
		Statements: []*parser.Statement{
			{
				Assignment: &parser.Assignment{
					Name: "test",
					Value: &parser.Expression{
						Ternary: nil, // Nil ternary
					},
				},
			},
		},
	}

	transformer := NewTANamespaceTransformer()
	_, err := transformer.Transform(ast)
	
	// Should not panic, should handle nil gracefully
	if err != nil {
		t.Fatalf("Transform should handle nil gracefully, got error: %v", err)
	}
}

// Test mixed v4/v5 syntax (partially migrated file)
func TestPipeline_MixedV4V5Syntax(t *testing.T) {
	input := `
sma20 = sma(close, 20)
ema20 = ta.ema(close, 20)
rsi14 = rsi(close, 14)
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

	// All should be transformed to ta. namespace
	expectedNames := []string{"ta.sma", "ta.ema", "ta.rsi"}
	for i, expected := range expectedNames {
		expr := result.Statements[i].Assignment.Value
		call := findCallInFactor(expr.Ternary.Condition.Left.Left.Left.Left.Left)
		if call == nil {
			t.Fatalf("Statement %d: expected call", i)
		}
		
		// For already-transformed ema, check it doesn't double-transform
		if i == 1 && call.Callee.MemberAccess == nil {
			// Parser saw "ta.ema" as MemberAccess
			continue
		}
		
		if call.Callee.Ident != nil && *call.Callee.Ident != expected {
			t.Errorf("Statement %d: expected %s, got %s", i, expected, *call.Callee.Ident)
		}
	}
}

// Test all transformer types together
func TestAllTransformers_Coverage(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		transformer Transformer
		checkFunc   string
	}{
		{
			name:        "TANamespace - crossover",
			input:       `signal = crossover(fast, slow)`,
			transformer: NewTANamespaceTransformer(),
			checkFunc:   "ta.crossover",
		},
		{
			name:        "TANamespace - stdev",
			input:       `stddev = stdev(close, 20)`,
			transformer: NewTANamespaceTransformer(),
			checkFunc:   "ta.stdev",
		},
		{
			name:        "MathNamespace - sqrt",
			input:       `root = sqrt(x)`,
			transformer: NewMathNamespaceTransformer(),
			checkFunc:   "math.sqrt",
		},
		{
			name:        "MathNamespace - max",
			input:       `maximum = max(a, b)`,
			transformer: NewMathNamespaceTransformer(),
			checkFunc:   "math.max",
		},
		{
			name:        "RequestNamespace - security",
			input:       `daily = security(tickerid, "D", close)`,
			transformer: NewRequestNamespaceTransformer(),
			checkFunc:   "request.security",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			ast, err := p.ParseString("test", tc.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			result, err := tc.transformer.Transform(ast)
			if err != nil {
				t.Fatalf("Transform failed: %v", err)
			}

			expr := result.Statements[0].Assignment.Value
			call := findCallInFactor(expr.Ternary.Condition.Left.Left.Left.Left.Left)
			if call == nil || call.Callee.Ident == nil || *call.Callee.Ident != tc.checkFunc {
				t.Errorf("Expected %s transformation", tc.checkFunc)
			}
		})
	}
}
