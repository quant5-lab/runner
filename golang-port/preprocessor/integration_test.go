package preprocessor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/borisquantlab/pinescript-go/parser"
)

// TestIntegration_DailyLinesSimple tests the full v4→v5 pipeline with the actual file
func TestIntegration_DailyLinesSimple(t *testing.T) {
	// Find the strategies directory
	strategyPath := filepath.Join("..", "..", "strategies", "daily-lines-simple.pine")

	// Read the actual file
	content, err := os.ReadFile(strategyPath)
	if err != nil {
		t.Skipf("Skipping integration test: cannot read %s: %v", strategyPath, err)
	}

	// Parse the v4 code
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := p.ParseString("daily-lines-simple.pine", string(content))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Run full v4→v5 pipeline
	pipeline := NewV4ToV5Pipeline()
	result, err := pipeline.Run(ast)
	if err != nil {
		t.Fatalf("Pipeline failed: %v", err)
	}

	// Verify transformations
	if len(result.Statements) < 4 {
		t.Fatalf("Expected at least 4 statements (study + 3 SMAs), got %d", len(result.Statements))
	}

	// Statement 0: study() → indicator()
	studyExpr := result.Statements[0].Expression
	if studyExpr == nil || studyExpr.Expr == nil {
		t.Fatal("Expected study/indicator call in first statement")
	}
	studyCall := findCallInFactor(studyExpr.Expr.Ternary.Condition.Left.Left.Left.Left.Left)
	if studyCall == nil {
		t.Fatal("Expected call expression for study/indicator")
	}
	if studyCall.Callee.Ident == nil || *studyCall.Callee.Ident != "indicator" {
		t.Errorf("Expected 'indicator', got '%v'", studyCall.Callee.Ident)
	}

	// Statements 1-3: sma() → ta.sma()
	expectedVars := []string{"ma20", "ma50", "ma200"}
	for i, varName := range expectedVars {
		stmt := result.Statements[i+1]
		if stmt.Assignment == nil {
			t.Fatalf("Statement %d: expected assignment", i+1)
		}
		if stmt.Assignment.Name != varName {
			t.Errorf("Statement %d: expected variable '%s', got '%s'", i+1, varName, stmt.Assignment.Name)
		}

		expr := stmt.Assignment.Value
		call := findCallInFactor(expr.Ternary.Condition.Left.Left.Left.Left.Left)
		if call == nil {
			t.Fatalf("Statement %d: expected call expression", i+1)
		}
		if call.Callee.Ident == nil || *call.Callee.Ident != "ta.sma" {
			t.Errorf("Statement %d: expected 'ta.sma', got '%v'", i+1, call.Callee.Ident)
		}
	}
}

// TestIntegration_PipelineIdempotency tests that running pipeline twice gives same result
func TestIntegration_PipelineIdempotency(t *testing.T) {
	input := `
study("Test")
ma = sma(close, 20)
`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast1, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// First transformation
	pipeline := NewV4ToV5Pipeline()
	result1, err := pipeline.Run(ast1)
	if err != nil {
		t.Fatalf("First pipeline run failed: %v", err)
	}

	// Parse again for second transformation
	ast2, err := p.ParseString("test", input)
	if err != nil {
		t.Fatalf("Second parse failed: %v", err)
	}

	// Second transformation
	result2, err := pipeline.Run(ast2)
	if err != nil {
		t.Fatalf("Second pipeline run failed: %v", err)
	}

	// Results should be identical
	if len(result1.Statements) != len(result2.Statements) {
		t.Errorf("Different number of statements: %d vs %d", len(result1.Statements), len(result2.Statements))
	}

	// Check first statement (study → indicator)
	call1 := findCallInFactor(result1.Statements[0].Expression.Expr.Ternary.Condition.Left.Left.Left.Left.Left)
	call2 := findCallInFactor(result2.Statements[0].Expression.Expr.Ternary.Condition.Left.Left.Left.Left.Left)

	if call1 == nil || call2 == nil {
		t.Fatal("Expected call expressions")
	}

	if call1.Callee.Ident == nil || call2.Callee.Ident == nil {
		t.Fatal("Expected callee identifiers")
	}

	if *call1.Callee.Ident != *call2.Callee.Ident {
		t.Errorf("Different transformations: %s vs %s", *call1.Callee.Ident, *call2.Callee.Ident)
	}
}

// TestIntegration_AllNamespaces tests file with mixed ta/math/request functions
func TestIntegration_AllNamespaces(t *testing.T) {
	input := `
study("Mixed Namespaces")
ma = sma(close, 20)
stddev = stdev(close, 20)
absVal = abs(ma)
dailyHigh = security(syminfo.tickerid, "D", high)
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

	// Verify all transformations
	expectedTransformations := []struct {
		stmtIndex int
		expected  string
	}{
		{0, "indicator"},        // study → indicator
		{1, "ta.sma"},           // sma → ta.sma
		{2, "ta.stdev"},         // stdev → ta.stdev
		{3, "math.abs"},         // abs → math.abs
		{4, "request.security"}, // security → request.security
	}

	for _, exp := range expectedTransformations {
		var call *parser.CallExpr

		stmt := result.Statements[exp.stmtIndex]
		if stmt.Expression != nil {
			call = findCallInFactor(stmt.Expression.Expr.Ternary.Condition.Left.Left.Left.Left.Left)
		} else if stmt.Assignment != nil {
			call = findCallInFactor(stmt.Assignment.Value.Ternary.Condition.Left.Left.Left.Left.Left)
		}

		if call == nil {
			t.Fatalf("Statement %d: expected call expression", exp.stmtIndex)
		}

		if call.Callee.Ident == nil || *call.Callee.Ident != exp.expected {
			t.Errorf("Statement %d: expected '%s', got '%v'", exp.stmtIndex, exp.expected, call.Callee.Ident)
		}
	}
}

// TestIntegration_ErrorRecovery tests that parser errors are handled gracefully
func TestIntegration_InvalidSyntax(t *testing.T) {
	input := `
study("Test"
ma = sma(close 20)
` // Missing closing paren and comma

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	_, err = p.ParseString("test", input)
	if err == nil {
		t.Error("Expected parse error for invalid syntax")
	}

	// Error should be descriptive
	if err != nil && len(err.Error()) == 0 {
		t.Error("Parse error should have descriptive message")
	}
}

// TestIntegration_LargeFile tests performance with realistic file size
func TestIntegration_LargeFile(t *testing.T) {
	// Build a large file with many function calls
	input := "study(\"Large File\")\n"
	for i := 0; i < 100; i++ {
		input += "ma" + string(rune('a'+i%26)) + " = sma(close, 20)\n"
	}

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

	// Should have 101 statements (1 study + 100 assignments)
	if len(result.Statements) != 101 {
		t.Errorf("Expected 101 statements, got %d", len(result.Statements))
	}

	// Spot check a few transformations
	for _, idx := range []int{1, 50, 100} {
		expr := result.Statements[idx].Assignment.Value
		call := findCallInFactor(expr.Ternary.Condition.Left.Left.Left.Left.Left)
		if call == nil || call.Callee.Ident == nil || *call.Callee.Ident != "ta.sma" {
			t.Errorf("Statement %d: expected ta.sma transformation", idx)
		}
	}
}
