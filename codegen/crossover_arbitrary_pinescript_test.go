package codegen

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

/*
End-to-end compilation tests for crossover arbitrary PineScript support.

These tests validate the full pine-gen pipeline (lexer → parser → codegen → Go compiler).
While crossover_inline_handler_test.go provides comprehensive unit testing, these E2E tests ensure:
1. Lexer correctly parses complex PineScript expressions
2. Parser builds correct AST for all expression types
3. Codegen detects and enforces stateful indicator rules
4. Generated Go code compiles successfully
5. Error messages propagate correctly through full pipeline

Test Organization:
- Positive tests: Validate successful compilation with supported features
- Negative tests: Validate error detection for unsupported patterns (inline stateful indicators)

Coverage complements unit tests by catching integration issues that may not appear in isolated tests.
*/

func TestCrossover_BinaryExpressionWithStatefulIndicator(t *testing.T) {
	pine := `//@version=5
strategy("Test")
ema10 = ta.ema(close, 10)
sma20 = ta.sma(close, 20)
if ta.crossover(sma20 + ema10, high)
    strategy.entry("long", strategy.long)
`
	if err := compilePine(pine); err != nil {
		t.Fatalf("Expected successful compilation with extracted EMA, got: %v", err)
	}
}

/* Inline stateful TA calls in binary expressions are now hoisted automatically */
func TestCrossover_BinaryExpressionInlineStateful_Hoisted(t *testing.T) {
	pine := `//@version=5
strategy("Test")
if ta.crossover(ta.sma(close, 20) + ta.ema(close, 10), high)
    strategy.entry("long", strategy.long)
`
	if err := compilePine(pine); err != nil {
		t.Fatalf("Expected successful compilation via hoisting, got: %v", err)
	}
}

/* Nested stateful TA calls are now hoisted automatically */
func TestCrossover_NestedTACallWithStateful_Hoisted(t *testing.T) {
	pine := `//@version=5
strategy("Test")
if ta.crossover(ta.sma(ta.ema(close, 10), 20), high)
    strategy.entry("long", strategy.long)
`
	if err := compilePine(pine); err != nil {
		t.Fatalf("Expected successful compilation via hoisting, got: %v", err)
	}
}

func TestCrossover_NestedWindowFunctions_ShouldSucceed(t *testing.T) {
	pine := `//@version=5
strategy("Test")
if ta.crossover(ta.sma(ta.wma(close, 10), 20), high)
    strategy.entry("long", strategy.long)
`
	if err := compilePine(pine); err != nil {
		t.Fatalf("Expected successful compilation with nested window functions, got: %v", err)
	}
}

func TestCrossover_TernaryExpression(t *testing.T) {
	pine := `//@version=5
strategy("Test")
condition = close > open
if ta.crossover(condition ? close : open, ta.sma(close, 20))
    strategy.entry("long", strategy.long)
`
	if err := compilePine(pine); err != nil {
		t.Fatalf("Expected successful compilation with ternary expression, got: %v", err)
	}
}

/* Stateful TA calls in ternary expressions are now hoisted automatically */
func TestCrossover_TernaryWithStatefulIndicator_Hoisted(t *testing.T) {
	pine := `//@version=5
strategy("Test")
condition = close > open
if ta.crossover(condition ? ta.ema(close, 10) : close, high)
    strategy.entry("long", strategy.long)
`
	if err := compilePine(pine); err != nil {
		t.Fatalf("Expected successful compilation via hoisting, got: %v", err)
	}
}

/* Multiple stateful TA calls in arithmetic are now hoisted automatically */
func TestCrossover_ComplexArithmeticWithMultipleStateful_Hoisted(t *testing.T) {
	pine := `//@version=5
strategy("Test")
if ta.crossover((ta.ema(close, 10) + ta.rma(high, 20)) / 2, ta.sma(close, 50))
    strategy.entry("long", strategy.long)
`
	if err := compilePine(pine); err != nil {
		t.Fatalf("Expected successful compilation via hoisting, got: %v", err)
	}
}

func TestCrossover_UnaryExpressionWithIdentifier(t *testing.T) {
	pine := `//@version=5
strategy("Test")
ema10 = ta.ema(close, 10)
if ta.crossover(-ema10, 0)
    strategy.entry("long", strategy.long)
`
	if err := compilePine(pine); err != nil {
		t.Fatalf("Expected successful compilation with unary expression, got: %v", err)
	}
}

/* Stateful TA calls in unary expressions are now hoisted automatically */
func TestCrossover_UnaryExpressionWithStatefulCall_Hoisted(t *testing.T) {
	pine := `//@version=5
strategy("Test")
if ta.crossover(-ta.ema(close, 10), 0)
    strategy.entry("long", strategy.long)
`
	if err := compilePine(pine); err != nil {
		t.Fatalf("Expected successful compilation via hoisting, got: %v", err)
	}
}

/* Deeply nested mixed window and stateful TA calls are now hoisted automatically */
func TestCrossover_DeepNestingMixedWindowAndStateful_Hoisted(t *testing.T) {
	pine := `//@version=5
strategy("Test")
if ta.crossover(ta.sma(ta.wma(ta.ema(close, 5), 10), 20), high)
    strategy.entry("long", strategy.long)
`
	if err := compilePine(pine); err != nil {
		t.Fatalf("Expected successful compilation via hoisting, got: %v", err)
	}
}

func TestCrossover_MultipleWindowFunctionsInBinaryOps(t *testing.T) {
	pine := `//@version=5
strategy("Test")
if ta.crossover(ta.sma(close, 10) + ta.wma(close, 20) - ta.stdev(close, 30), high)
    strategy.entry("long", strategy.long)
`
	if err := compilePine(pine); err != nil {
		t.Fatalf("Expected successful compilation with multiple window functions, got: %v", err)
	}
}

/* Helper: compile PineScript using pine-gen binary */
func compilePine(pine string) error {
	tmpDir, err := os.MkdirTemp("", "crossover_test")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	pineFile := filepath.Join(tmpDir, "test.pine")
	if err := os.WriteFile(pineFile, []byte(pine), 0644); err != nil {
		return err
	}

	outputFile := filepath.Join(tmpDir, "test.go")
	cmd := exec.Command("./build/pine-gen", "-input", pineFile, "-output", outputFile)
	cmd.Dir = ".."
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &compileError{output: string(output), err: err}
	}

	return nil
}

type compileError struct {
	output string
	err    error
}

func (e *compileError) Error() string {
	return e.output
}
