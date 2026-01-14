package parser

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// TestBooleanLiterals_InTernary verifies true/false parse as Literals, not Identifiers
func TestBooleanLiterals_InTernary(t *testing.T) {
	tests := []struct {
		name           string
		script         string
		expectConValue interface{}
		expectAltValue interface{}
	}{
		{
			name: "false consequent, true alternate",
			script: `//@version=5
indicator("Test")
x = na(close) ? false : true`,
			expectConValue: false,
			expectAltValue: true,
		},
		{
			name: "true consequent, false alternate",
			script: `//@version=5
indicator("Test")
x = close > 100 ? true : false`,
			expectConValue: true,
			expectAltValue: false,
		},
		{
			name: "both false",
			script: `//@version=5
indicator("Test")
x = close > 100 ? false : false`,
			expectConValue: false,
			expectAltValue: false,
		},
		{
			name: "both true",
			script: `//@version=5
indicator("Test")
x = close > 100 ? true : true`,
			expectConValue: true,
			expectAltValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			parseResult, err := p.ParseBytes("test.pine", []byte(tt.script))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(parseResult)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			// Find variable declaration
			var condExpr *ast.ConditionalExpression
			for _, stmt := range program.Body {
				if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
					if len(varDecl.Declarations) > 0 {
						if cond, ok := varDecl.Declarations[0].Init.(*ast.ConditionalExpression); ok {
							condExpr = cond
							break
						}
					}
				}
			}

			if condExpr == nil {
				t.Fatal("ConditionalExpression not found")
			}

			// Verify consequent is Literal with correct value
			conLit, ok := condExpr.Consequent.(*ast.Literal)
			if !ok {
				t.Errorf("Consequent is %T, expected *ast.Literal", condExpr.Consequent)
			} else {
				if conLit.NodeType != ast.TypeLiteral {
					t.Errorf("Consequent NodeType = %s, expected %s", conLit.NodeType, ast.TypeLiteral)
				}
				if conLit.Value != tt.expectConValue {
					t.Errorf("Consequent Value = %v, expected %v", conLit.Value, tt.expectConValue)
				}
			}

			// Verify alternate is Literal with correct value
			altLit, ok := condExpr.Alternate.(*ast.Literal)
			if !ok {
				t.Errorf("Alternate is %T, expected *ast.Literal", condExpr.Alternate)
			} else {
				if altLit.NodeType != ast.TypeLiteral {
					t.Errorf("Alternate NodeType = %s, expected %s", altLit.NodeType, ast.TypeLiteral)
				}
				if altLit.Value != tt.expectAltValue {
					t.Errorf("Alternate Value = %v, expected %v", altLit.Value, tt.expectAltValue)
				}
			}
		})
	}
}

// TestBooleanLiterals_InComparison verifies true/false work in comparisons
func TestBooleanLiterals_InComparison(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{
			name: "compare with true",
			script: `//@version=5
indicator("Test")
x = close > 100 == true`,
		},
		{
			name: "compare with false",
			script: `//@version=5
indicator("Test")
x = close > 100 == false`,
		},
		{
			name: "true and false in logical expression",
			script: `//@version=5
indicator("Test")
x = true and false`,
		},
		{
			name: "true or false in logical expression",
			script: `//@version=5
indicator("Test")
x = true or false`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			parseResult, err := p.ParseBytes("test.pine", []byte(tt.script))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			_, err = converter.ToESTree(parseResult)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}
		})
	}
}

// TestBooleanLiterals_RegressionSafety ensures booleans don't become Identifiers
func TestBooleanLiterals_RegressionSafety(t *testing.T) {
	script := `//@version=5
indicator("Test")
session_open = na(time(timeframe.period, "0950-1345")) ? false : true
is_entry = time(timeframe.period, "1000-1200") ? true : false`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	parseResult, err := p.ParseBytes("test.pine", []byte(script))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(parseResult)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	// Count boolean Literals
	boolCount := 0
	identifierCount := 0

	var countBooleans func(ast.Node)
	countBooleans = func(node ast.Node) {
		if lit, ok := node.(*ast.Literal); ok {
			if _, isBool := lit.Value.(bool); isBool {
				boolCount++
			}
		}
		if ident, ok := node.(*ast.Identifier); ok {
			if ident.Name == "true" || ident.Name == "false" {
				identifierCount++
			}
		}

		// Recursively check children
		switch n := node.(type) {
		case *ast.Program:
			for _, stmt := range n.Body {
				countBooleans(stmt)
			}
		case *ast.VariableDeclaration:
			for _, decl := range n.Declarations {
				if decl.Init != nil {
					countBooleans(decl.Init)
				}
			}
		case *ast.ConditionalExpression:
			countBooleans(n.Test)
			countBooleans(n.Consequent)
			countBooleans(n.Alternate)
		case *ast.CallExpression:
			for _, arg := range n.Arguments {
				countBooleans(arg)
			}
		case *ast.BinaryExpression:
			countBooleans(n.Left)
			countBooleans(n.Right)
		}
	}

	countBooleans(program)

	// Expect 4 boolean Literals (2 false, 2 true), 0 Identifiers named "true"/"false"
	if boolCount != 4 {
		t.Errorf("Expected 4 boolean Literals, found %d", boolCount)
	}
	if identifierCount > 0 {
		t.Errorf("REGRESSION: Found %d Identifiers with name 'true' or 'false' (should be 0)", identifierCount)
	}
}
