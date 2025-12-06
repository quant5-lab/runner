package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestExpressionAnalyzer_SimpleCallExpression tests detection of single TA function call */
func TestExpressionAnalyzer_SimpleCallExpression(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
	analyzer := NewExpressionAnalyzer(g)

	// Create: ta.sma(close, 20)
	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 20},
		},
	}

	calls := analyzer.FindNestedCalls(call)

	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}

	if calls[0].FuncName != "ta.sma" {
		t.Errorf("Expected funcName 'ta.sma', got %q", calls[0].FuncName)
	}

	if calls[0].ArgHash == "" {
		t.Error("Expected non-empty ArgHash")
	}
}

/* TestExpressionAnalyzer_NestedCalls tests detection of nested TA calls */
func TestExpressionAnalyzer_NestedCalls(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
	analyzer := NewExpressionAnalyzer(g)

	// Create: rma(max(change(close), 0), 9)
	innerCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "change"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
		},
	}

	midCall := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "max"},
		Arguments: []ast.Expression{innerCall, &ast.Literal{Value: 0}},
	}

	outerCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "rma"},
		},
		Arguments: []ast.Expression{midCall, &ast.Literal{Value: 9}},
	}

	calls := analyzer.FindNestedCalls(outerCall)

	// Should find: rma, max, change (outer to inner order)
	if len(calls) != 3 {
		t.Fatalf("Expected 3 calls, got %d", len(calls))
	}

	expectedFuncs := []string{"ta.rma", "max", "ta.change"}
	for i, call := range calls {
		if call.FuncName != expectedFuncs[i] {
			t.Errorf("Call %d: expected %q, got %q", i, expectedFuncs[i], call.FuncName)
		}
	}
}

/* TestExpressionAnalyzer_BinaryExpression tests detection in binary operations */
func TestExpressionAnalyzer_BinaryExpression(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
	analyzer := NewExpressionAnalyzer(g)

	// Create: ta.sma(close, 50) > ta.sma(close, 200)
	leftCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 50},
		},
	}

	rightCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 200},
		},
	}

	binExpr := &ast.BinaryExpression{
		Operator: ">",
		Left:     leftCall,
		Right:    rightCall,
	}

	calls := analyzer.FindNestedCalls(binExpr)

	if len(calls) != 2 {
		t.Fatalf("Expected 2 calls, got %d", len(calls))
	}

	// Both should be ta.sma but with different hashes (different periods)
	if calls[0].FuncName != "ta.sma" || calls[1].FuncName != "ta.sma" {
		t.Error("Expected both calls to be ta.sma")
	}

	if calls[0].ArgHash == calls[1].ArgHash {
		t.Error("Expected different ArgHash for different periods (50 vs 200)")
	}
}

/* TestExpressionAnalyzer_HashUniqueness tests that different arguments produce different hashes */
func TestExpressionAnalyzer_HashUniqueness(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
	analyzer := NewExpressionAnalyzer(g)

	// sma(close, 50)
	call1 := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 50},
		},
	}

	// sma(close, 200)
	call2 := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 200},
		},
	}

	hash1 := analyzer.computeArgHash(call1)
	hash2 := analyzer.computeArgHash(call2)

	if hash1 == hash2 {
		t.Error("Expected different hashes for sma(close,50) vs sma(close,200)")
	}
}

/* TestExpressionAnalyzer_HashConsistency tests that same arguments produce same hash */
func TestExpressionAnalyzer_HashConsistency(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
	analyzer := NewExpressionAnalyzer(g)

	// Create same call twice
	createCall := func() *ast.CallExpression {
		return &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "sma"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 50},
			},
		}
	}

	call1 := createCall()
	call2 := createCall()

	hash1 := analyzer.computeArgHash(call1)
	hash2 := analyzer.computeArgHash(call2)

	if hash1 != hash2 {
		t.Errorf("Expected consistent hash for identical calls, got %q vs %q", hash1, hash2)
	}
}

/* TestExpressionAnalyzer_NoCallsInLiterals tests that literals don't produce calls */
func TestExpressionAnalyzer_NoCallsInLiterals(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
	analyzer := NewExpressionAnalyzer(g)

	literal := &ast.Literal{Value: 42.0}
	calls := analyzer.FindNestedCalls(literal)

	if len(calls) != 0 {
		t.Errorf("Expected 0 calls from literal, got %d", len(calls))
	}
}

/* TestExpressionAnalyzer_ConditionalExpression tests detection in ternary operators */
func TestExpressionAnalyzer_ConditionalExpression(t *testing.T) {
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
	analyzer := NewExpressionAnalyzer(g)

	// Create: condition ? ta.sma(close, 20) : ta.ema(close, 10)
	smaCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 20},
		},
	}

	emaCall := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "ema"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 10},
		},
	}

	conditional := &ast.ConditionalExpression{
		Test:       &ast.Literal{Value: true},
		Consequent: smaCall,
		Alternate:  emaCall,
	}

	calls := analyzer.FindNestedCalls(conditional)

	if len(calls) != 2 {
		t.Fatalf("Expected 2 calls, got %d", len(calls))
	}

	funcNames := []string{calls[0].FuncName, calls[1].FuncName}
	hasSma := false
	hasEma := false
	for _, fn := range funcNames {
		if fn == "ta.sma" {
			hasSma = true
		}
		if fn == "ta.ema" {
			hasEma = true
		}
	}
	if !hasSma || !hasEma {
		t.Errorf("Expected ta.sma and ta.ema, got %v", funcNames)
	}
}
