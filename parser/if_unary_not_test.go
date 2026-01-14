package parser

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestUnaryOperator_IfCondition validates unary operator parsing in if statement conditions */
func TestUnaryOperator_IfCondition(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantCondition string
		wantOperator  string
	}{
		{
			name:          "not with identifier",
			input:         "if not has_trade\n    x = 1",
			wantCondition: "UnaryExpression",
			wantOperator:  "not",
		},
		{
			name:          "not with function call",
			input:         "if not na(close)\n    y = 2",
			wantCondition: "UnaryExpression",
			wantOperator:  "not",
		},
		{
			name:          "not with logical expression",
			input:         "if not has_trade and buy_signal\n    z = 3",
			wantCondition: "LogicalExpression",
			wantOperator:  "not",
		},
		{
			name:          "not with parenthesized comparison",
			input:         "if not (close > open)\n    w = 4",
			wantCondition: "UnaryExpression",
			wantOperator:  "not",
		},
		{
			name:          "exclamation mark negation",
			input:         "if !enabled\n    a = 5",
			wantCondition: "UnaryExpression",
			wantOperator:  "!",
		},
		{
			name:          "arithmetic negation",
			input:         "if -delta > threshold\n    b = 6",
			wantCondition: "BinaryExpression",
			wantOperator:  "-",
		},
		{
			name:          "positive unary",
			input:         "if +value == 0\n    c = 7",
			wantCondition: "BinaryExpression",
			wantOperator:  "+",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := parser.ParseString("", tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			conv := NewConverter()
			program, err := conv.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if len(program.Body) == 0 {
				t.Fatal("Expected at least one statement")
			}

			ifStmt, ok := program.Body[0].(*ast.IfStatement)
			if !ok {
				t.Fatalf("Expected IfStatement, got %T", program.Body[0])
			}

			jsonBytes, err := json.MarshalIndent(ifStmt.Test, "", "  ")
			if err != nil {
				t.Fatalf("Failed to marshal AST: %v", err)
			}
			astJSON := string(jsonBytes)

			if !strings.Contains(astJSON, tt.wantCondition) {
				t.Errorf("Expected condition type %q in AST, got:\n%s", tt.wantCondition, astJSON)
			}

			if strings.Contains(tt.input, "not ") && strings.Contains(astJSON, `"name": "not"`) {
				if !strings.Contains(astJSON, `"operator": "not"`) {
					t.Error("VIOLATION: 'not' parsed as Identifier instead of UnaryExpression operator")
				}
			}
		})
	}
}

/* TestUnaryOperator_ComplexNesting validates nested unary operator structures */
func TestUnaryOperator_ComplexNesting(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		validateStructure func(*testing.T, ast.Expression)
	}{
		{
			name:  "not with logical and",
			input: "if not has_trade and buy_signal\n    x = 1",
			validateStructure: func(t *testing.T, expr ast.Expression) {
				logicalExpr, ok := expr.(*ast.LogicalExpression)
				if !ok {
					t.Fatalf("Expected LogicalExpression, got %T", expr)
				}

				unaryExpr, ok := logicalExpr.Left.(*ast.UnaryExpression)
				if !ok {
					t.Fatalf("Expected left side to be UnaryExpression, got %T", logicalExpr.Left)
				}

				if unaryExpr.Operator != "not" {
					t.Errorf("Expected unary operator 'not', got %q", unaryExpr.Operator)
				}

				if _, ok := logicalExpr.Right.(*ast.Identifier); !ok {
					t.Errorf("Expected right side to be Identifier, got %T", logicalExpr.Right)
				}
			},
		},
		{
			name:  "not with logical or",
			input: "if not enabled or force_entry\n    y = 2",
			validateStructure: func(t *testing.T, expr ast.Expression) {
				logicalExpr, ok := expr.(*ast.LogicalExpression)
				if !ok {
					t.Fatalf("Expected LogicalExpression, got %T", expr)
				}

				if logicalExpr.Operator != "||" {
					t.Errorf("Expected operator '||', got %q", logicalExpr.Operator)
				}
			},
		},
		{
			name:  "double negation",
			input: "if not (not condition)\n    z = 3",
			validateStructure: func(t *testing.T, expr ast.Expression) {
				outerUnary, ok := expr.(*ast.UnaryExpression)
				if !ok {
					t.Fatalf("Expected outer UnaryExpression, got %T", expr)
				}

				innerUnary, ok := outerUnary.Argument.(*ast.UnaryExpression)
				if !ok {
					t.Fatalf("Expected inner UnaryExpression, got %T", outerUnary.Argument)
				}

				if innerUnary.Operator != "not" {
					t.Errorf("Expected inner operator 'not', got %q", innerUnary.Operator)
				}
			},
		},
		{
			name:  "not with comparison",
			input: "if not (close > open)\n    w = 4",
			validateStructure: func(t *testing.T, expr ast.Expression) {
				unaryExpr, ok := expr.(*ast.UnaryExpression)
				if !ok {
					t.Fatalf("Expected UnaryExpression, got %T", expr)
				}

				binaryExpr, ok := unaryExpr.Argument.(*ast.BinaryExpression)
				if !ok {
					t.Fatalf("Expected argument to be BinaryExpression, got %T", unaryExpr.Argument)
				}

				if binaryExpr.Operator != ">" {
					t.Errorf("Expected comparison operator '>', got %q", binaryExpr.Operator)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := parser.ParseString("", tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			conv := NewConverter()
			program, err := conv.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if len(program.Body) == 0 {
				t.Fatal("Expected at least one statement")
			}

			ifStmt, ok := program.Body[0].(*ast.IfStatement)
			if !ok {
				t.Fatalf("Expected IfStatement, got %T", program.Body[0])
			}

			tt.validateStructure(t, ifStmt.Test)
		})
	}
}

/* TestUnaryOperator_EdgeCases validates unary operator edge case handling */
func TestUnaryOperator_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{
			name:      "not with member expression",
			input:     "if not strategy.position_size\n    x = 1",
			shouldErr: false,
		},
		{
			name:      "not with subscript",
			input:     "if not close[1]\n    y = 2",
			shouldErr: false,
		},
		{
			name:      "not with ternary",
			input:     "if not (enabled ? true : false)\n    z = 3",
			shouldErr: false,
		},
		{
			name:      "arithmetic negation with series",
			input:     "if -high > -low\n    w = 4",
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := parser.ParseString("", tt.input)
			if (err != nil) != tt.shouldErr {
				t.Fatalf("Parse error = %v, shouldErr = %v", err, tt.shouldErr)
			}

			if tt.shouldErr {
				return
			}

			conv := NewConverter()
			program, err := conv.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if len(program.Body) == 0 {
				t.Fatal("Expected at least one statement")
			}

			if _, ok := program.Body[0].(*ast.IfStatement); !ok {
				t.Fatalf("Expected IfStatement, got %T", program.Body[0])
			}
		})
	}
}

/* TestIfStatement_ComparisonBackwardCompatibility ensures existing if conditions still work */
func TestIfStatement_ComparisonBackwardCompatibility(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantTestType string
	}{
		{
			name:         "simple comparison",
			input:        "if close > open\n    x = 1",
			wantTestType: "BinaryExpression",
		},
		{
			name:         "logical and",
			input:        "if has_trade and buy_signal\n    y = 2",
			wantTestType: "LogicalExpression",
		},
		{
			name:         "logical or",
			input:        "if sell_signal or stop_loss\n    z = 3",
			wantTestType: "LogicalExpression",
		},
		{
			name:         "complex logical and",
			input:        "if has_trade and volume > 1000\n    w = 4",
			wantTestType: "LogicalExpression",
		},
		{
			name:         "equality comparison",
			input:        "if state == 1\n    a = 5",
			wantTestType: "BinaryExpression",
		},
		{
			name:         "inequality comparison",
			input:        "if state != 0\n    b = 6",
			wantTestType: "BinaryExpression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := parser.ParseString("", tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			conv := NewConverter()
			program, err := conv.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if len(program.Body) == 0 {
				t.Fatal("Expected at least one statement")
			}

			ifStmt, ok := program.Body[0].(*ast.IfStatement)
			if !ok {
				t.Fatalf("Expected IfStatement, got %T", program.Body[0])
			}

			if ifStmt.Test == nil {
				t.Error("Expected non-nil test condition")
			}

			jsonBytes, _ := json.MarshalIndent(ifStmt.Test, "", "  ")
			astJSON := string(jsonBytes)

			if !strings.Contains(astJSON, tt.wantTestType) {
				t.Errorf("Expected test type %q, got:\n%s", tt.wantTestType, astJSON)
			}
		})
	}
}
