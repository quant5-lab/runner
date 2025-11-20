package preprocessor

import (
	"testing"

	"github.com/borisquantlab/pinescript-go/parser"
)

func TestTANamespaceTransformer_SimpleAssignment(t *testing.T) {
	input := `ma20 = sma(close, 20)`

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

	// Check that sma was renamed to ta.sma
	if result.Statements[0].Assignment == nil {
		t.Fatal("Expected assignment statement")
	}

	// The Call is nested inside Ternary.Condition.Left...Left.Left.Call
	expr := result.Statements[0].Assignment.Value
	if expr.Ternary == nil || expr.Ternary.Condition == nil {
		t.Fatal("Expected ternary with condition")
	}

	// Navigate through the nested structure to find the Call
	call := findCallInFactor(expr.Ternary.Condition.Left.Left.Left.Left.Left)
	if call == nil {
		t.Fatal("Expected call expression in nested structure")
	}
	if call.Callee == nil || call.Callee.Ident == nil {
		t.Fatal("Expected callee identifier")
	}
	if *call.Callee.Ident != "ta.sma" {
		t.Errorf("Expected callee 'ta.sma', got '%s'", *call.Callee.Ident)
	}
}

// Helper to extract Call from Factor
func findCallInFactor(factor *parser.Factor) *parser.CallExpr {
	if factor == nil {
		return nil
	}
	return factor.Call
}

func TestTANamespaceTransformer_MultipleIndicators(t *testing.T) {
	input := `
ma20 = sma(close, 20)
ma50 = ema(close, 50)
rsiVal = rsi(close, 14)
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

	// Check all three were transformed
	expectedCallees := []string{"ta.sma", "ta.ema", "ta.rsi"}
	for i, expected := range expectedCallees {
		expr := result.Statements[i].Assignment.Value
		call := findCallInFactor(expr.Ternary.Condition.Left.Left.Left.Left.Left)
		if call == nil {
			t.Fatalf("Statement %d: expected call expression", i)
		}
		if call.Callee == nil || call.Callee.Ident == nil {
			t.Fatalf("Statement %d: expected callee identifier", i)
		}
		if *call.Callee.Ident != expected {
			t.Errorf("Statement %d: expected callee '%s', got '%s'", i, expected, *call.Callee.Ident)
		}
	}
}

func TestTANamespaceTransformer_Crossover(t *testing.T) {
	input := `bullish = crossover(fast, slow)`

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

	expr := result.Statements[0].Assignment.Value
	call := findCallInFactor(expr.Ternary.Condition.Left.Left.Left.Left.Left)
	if call == nil {
		t.Fatal("Expected call expression")
	}
	if call.Callee == nil || call.Callee.Ident == nil {
		t.Fatal("Expected callee identifier")
	}
	if *call.Callee.Ident != "ta.crossover" {
		t.Errorf("Expected callee 'ta.crossover', got '%s'", *call.Callee.Ident)
	}
}

func TestTANamespaceTransformer_DailyLinesSimple(t *testing.T) {
	// This is the actual daily-lines-simple.pine content
	input := `
ma20 = sma(close, 20)
ma50 = sma(close, 50)
ma200 = sma(close, 200)
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

	// All three sma calls should be transformed to ta.sma
	for i := 0; i < 3; i++ {
		expr := result.Statements[i].Assignment.Value
		call := findCallInFactor(expr.Ternary.Condition.Left.Left.Left.Left.Left)
		if call == nil {
			t.Fatalf("Statement %d: expected call expression", i)
		}
		if call.Callee == nil || call.Callee.Ident == nil {
			t.Fatalf("Statement %d: expected callee identifier", i)
		}
		if *call.Callee.Ident != "ta.sma" {
			t.Errorf("Statement %d: expected callee 'ta.sma', got '%s'", i, *call.Callee.Ident)
		}
	}
}

func TestStudyToIndicatorTransformer(t *testing.T) {
	input := `study(title="Test", shorttitle="T", overlay=true)`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	transformer := NewStudyToIndicatorTransformer()
	result, err := transformer.Transform(ast)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	expr := result.Statements[0].Expression.Expr
	call := findCallInFactor(expr.Ternary.Condition.Left.Left.Left.Left.Left)
	if call == nil {
		t.Fatal("Expected call expression")
	}
	if call.Callee == nil || call.Callee.Ident == nil {
		t.Fatal("Expected callee identifier")
	}
	if *call.Callee.Ident != "indicator" {
		t.Errorf("Expected callee 'indicator', got '%s'", *call.Callee.Ident)
	}
}

func TestV4ToV5Pipeline(t *testing.T) {
	// Full daily-lines-simple.pine (v4 syntax)
	input := `
study(title="20-50-200 SMA", shorttitle="SMA Lines", overlay=true)
ma20 = sma(close, 20)
ma50 = sma(close, 50)
ma200 = sma(close, 200)
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
	if studyCall == nil {
		t.Fatal("Expected study call expression")
	}
	if studyCall.Callee == nil || studyCall.Callee.Ident == nil {
		t.Fatal("Expected study callee identifier")
	}
	if *studyCall.Callee.Ident != "indicator" {
		t.Errorf("Expected callee 'indicator', got '%s'", *studyCall.Callee.Ident)
	}

	// Check sma → ta.sma (3 occurrences)
	for i := 1; i <= 3; i++ {
		expr := result.Statements[i].Assignment.Value
		call := findCallInFactor(expr.Ternary.Condition.Left.Left.Left.Left.Left)
		if call == nil {
			t.Fatalf("Statement %d: expected call expression", i)
		}
		if call.Callee == nil || call.Callee.Ident == nil {
			t.Fatalf("Statement %d: expected callee identifier", i)
		}
		if *call.Callee.Ident != "ta.sma" {
			t.Errorf("Statement %d: expected callee 'ta.sma', got '%s'", i, *call.Callee.Ident)
		}
	}
}
