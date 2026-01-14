package parser

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* Integration tests for tuple destructuring with real-world PineScript patterns */

// TestTupleAssignment_RealWorld_IndicatorPatterns verifies common indicator patterns
func TestTupleAssignment_RealWorld_IndicatorPatterns(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		tupleCount  int
		elementName string
	}{
		{
			name: "ADX with DMI components",
			source: `//@version=4
study("ADX Test")
[ADX, up, down] = adx(14, 16)
plot(ADX)`,
			tupleCount:  3,
			elementName: "ADX",
		},
		{
			name: "Bollinger Bands",
			source: `//@version=5
indicator("BB Test")
[basis, upper, lower] = ta.bb(close, 20, 2.0)
plot(basis)`,
			tupleCount:  3,
			elementName: "basis",
		},
		{
			name: "MACD with signal and histogram",
			source: `//@version=5
indicator("MACD")
[macdLine, signalLine, histLine] = ta.macd(close, 12, 26, 9)
plot(macdLine)`,
			tupleCount:  3,
			elementName: "macdLine",
		},
		{
			name: "Stochastic oscillator",
			source: `//@version=5
indicator("Stoch")
[k, d] = ta.stoch(close, high, low, 14)
plot(k)`,
			tupleCount:  2,
			elementName: "k",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if len(program.Body) < 2 {
				t.Fatal("Expected at least 2 statements")
			}

			/* Find the tuple assignment statement */
			var tupleDecl *ast.VariableDeclaration
			for _, stmt := range program.Body {
				if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
					if _, ok := varDecl.Declarations[0].ID.(*ast.ArrayPattern); ok {
						tupleDecl = varDecl
						break
					}
				}
			}

			if tupleDecl == nil {
				t.Fatal("Expected to find tuple assignment in program body")
			}

			arrayPattern := tupleDecl.Declarations[0].ID.(*ast.ArrayPattern)
			if len(arrayPattern.Elements) != tt.tupleCount {
				t.Errorf("Expected %d elements, got %d", tt.tupleCount, len(arrayPattern.Elements))
			}

			if arrayPattern.Elements[0].Name != tt.elementName {
				t.Errorf("Expected first element '%s', got '%s'", tt.elementName, arrayPattern.Elements[0].Name)
			}
		})
	}
}

// TestTupleAssignment_RealWorld_WithCalculations verifies tuple assignments used in calculations
func TestTupleAssignment_RealWorld_WithCalculations(t *testing.T) {
	source := `//@version=5
indicator("BB Strategy")
length = input.int(20, "Length")
mult = input.float(2.0, "Multiplier")

[basis, upper, lower] = ta.bb(close, length, mult)
bandwidth = (upper - lower) / basis
pctB = (close - lower) / (upper - lower)

plot(basis)
plot(bandwidth)
plot(pctB)`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(source))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	/* Count tuple assignments vs regular assignments */
	tupleCount := 0
	regularCount := 0

	for _, stmt := range program.Body {
		if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
			if _, ok := varDecl.Declarations[0].ID.(*ast.ArrayPattern); ok {
				tupleCount++
			} else if _, ok := varDecl.Declarations[0].ID.(*ast.Identifier); ok {
				regularCount++
			}
		}
	}

	if tupleCount != 1 {
		t.Errorf("Expected 1 tuple assignment, got %d", tupleCount)
	}

	if regularCount < 4 {
		t.Errorf("Expected at least 4 regular assignments, got %d", regularCount)
	}
}

// TestTupleAssignment_RealWorld_MultipleInSameScript verifies multiple tuple assignments
func TestTupleAssignment_RealWorld_MultipleInSameScript(t *testing.T) {
	source := `//@version=5
indicator("Multi-Indicator")

[macd, signal, hist] = ta.macd(close, 12, 26, 9)
[k, d] = ta.stoch(close, high, low, 14)
[basis, upper, lower] = ta.bb(close, 20, 2.0)

plot(macd)
plot(k)
plot(basis)`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(source))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	/* Count tuple assignments */
	tupleAssignments := []int{}

	for _, stmt := range program.Body {
		if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
			if arrayPattern, ok := varDecl.Declarations[0].ID.(*ast.ArrayPattern); ok {
				tupleAssignments = append(tupleAssignments, len(arrayPattern.Elements))
			}
		}
	}

	if len(tupleAssignments) != 3 {
		t.Fatalf("Expected 3 tuple assignments, got %d", len(tupleAssignments))
	}

	expectedSizes := []int{3, 2, 3}
	for i, expected := range expectedSizes {
		if tupleAssignments[i] != expected {
			t.Errorf("Tuple %d: expected %d elements, got %d", i, expected, tupleAssignments[i])
		}
	}
}

// TestTupleAssignment_RealWorld_StrategyContext verifies tuple assignments in strategy scripts
func TestTupleAssignment_RealWorld_StrategyContext(t *testing.T) {
	source := `//@version=5
strategy("ADX Strategy", overlay=true)

adxLength = input.int(14, "ADX Length")
adxSmooth = input.int(14, "ADX Smoothing")

[ADX, plusDI, minusDI] = ta.dmi(adxLength, adxSmooth)

longCondition = ADX > 25 and plusDI > minusDI
shortCondition = ADX > 25 and minusDI > plusDI

if longCondition
    strategy.entry("Long", strategy.long)

if shortCondition
    strategy.entry("Short", strategy.short)`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(source))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	/* Verify tuple assignment exists and has correct structure */
	var foundTuple bool
	for _, stmt := range program.Body {
		if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
			if arrayPattern, ok := varDecl.Declarations[0].ID.(*ast.ArrayPattern); ok {
				if len(arrayPattern.Elements) == 3 &&
					arrayPattern.Elements[0].Name == "ADX" &&
					arrayPattern.Elements[1].Name == "plusDI" &&
					arrayPattern.Elements[2].Name == "minusDI" {
					foundTuple = true
					break
				}
			}
		}
	}

	if !foundTuple {
		t.Fatal("Expected to find ADX tuple assignment with correct elements")
	}
}
