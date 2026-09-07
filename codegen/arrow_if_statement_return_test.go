package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestInferIfStatementReturnType(t *testing.T) {
	engine := NewTypeInferenceEngine()

	tests := []struct {
		name   string
		ifStmt *ast.IfStatement
		want   string
	}{
		{
			name: "all branches return string literals → string",
			ifStmt: &ast.IfStatement{
				Test: &ast.BinaryExpression{Operator: "==", Left: &ast.Identifier{Name: "tf"}, Right: &ast.Literal{Value: "1m"}},
				Consequent: []ast.Node{
					&ast.ExpressionStatement{Expression: &ast.Literal{Value: "1"}},
				},
				Alternate: []ast.Node{
					&ast.ExpressionStatement{Expression: &ast.Literal{Value: "3"}},
				},
			},
			want: "string",
		},
		{
			name: "branches return float literals → float64",
			ifStmt: &ast.IfStatement{
				Test: &ast.BinaryExpression{Operator: ">", Left: &ast.Identifier{Name: "close"}, Right: &ast.Literal{Value: 100.0}},
				Consequent: []ast.Node{
					&ast.ExpressionStatement{Expression: &ast.Literal{Value: 1.0}},
				},
				Alternate: []ast.Node{
					&ast.ExpressionStatement{Expression: &ast.Literal{Value: 0.0}},
				},
			},
			want: "float64",
		},
		{
			name: "consequent is string → string even if alternate is empty",
			ifStmt: &ast.IfStatement{
				Test: &ast.Identifier{Name: "flag"},
				Consequent: []ast.Node{
					&ast.ExpressionStatement{Expression: &ast.Literal{Value: "D"}},
				},
			},
			want: "string",
		},
		{
			name: "nested else-if chain all returning strings → string",
			ifStmt: &ast.IfStatement{
				Test: &ast.BinaryExpression{Operator: "==", Left: &ast.Identifier{Name: "tf"}, Right: &ast.Literal{Value: "1m"}},
				Consequent: []ast.Node{
					&ast.ExpressionStatement{Expression: &ast.Literal{Value: "1"}},
				},
				Alternate: []ast.Node{
					&ast.IfStatement{
						Test: &ast.BinaryExpression{Operator: "==", Left: &ast.Identifier{Name: "tf"}, Right: &ast.Literal{Value: "5m"}},
						Consequent: []ast.Node{
							&ast.ExpressionStatement{Expression: &ast.Literal{Value: "5"}},
						},
						Alternate: []ast.Node{
							&ast.ExpressionStatement{Expression: &ast.Literal{Value: "60"}},
						},
					},
				},
			},
			want: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferIfStatementReturnType(tt.ifStmt, engine)
			if got != tt.want {
				t.Errorf("inferIfStatementReturnType: got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateIfStatementReturn_StringSwitch(t *testing.T) {
	gen := newTestGenerator()

	switchStmt := &ast.IfStatement{
		Test: &ast.BinaryExpression{
			Operator: "==",
			Left:     &ast.Identifier{Name: "tf"},
			Right:    &ast.Literal{Value: "1m"},
		},
		Consequent: []ast.Node{
			&ast.ExpressionStatement{Expression: &ast.Literal{Value: "1"}},
		},
		Alternate: []ast.Node{
			&ast.IfStatement{
				Test: &ast.BinaryExpression{
					Operator: "==",
					Left:     &ast.Identifier{Name: "tf"},
					Right:    &ast.Literal{Value: "5m"},
				},
				Consequent: []ast.Node{
					&ast.ExpressionStatement{Expression: &ast.Literal{Value: "5"}},
				},
				Alternate: []ast.Node{
					&ast.ExpressionStatement{Expression: &ast.Literal{Value: "60"}},
				},
			},
		},
	}

	arrowCodegen := NewArrowFunctionCodegen(gen)
	code, err := arrowCodegen.generateIfStatementReturn(switchStmt)
	if err != nil {
		t.Fatalf("generateIfStatementReturn: %v", err)
	}
	if !strings.Contains(code, "func() string") {
		t.Errorf("expected func() string IIFE, got: %s", code)
	}
	if !strings.Contains(code, `"1"`) || !strings.Contains(code, `"5"`) || !strings.Contains(code, `"60"`) {
		t.Errorf("expected branch values in output, got: %s", code)
	}
	if !strings.Contains(code, "return ") {
		t.Errorf("expected return statement, got: %s", code)
	}
}

/* --------------------------------------------------------------------------
 * TestInferBranchBodyType
 *
 * Verifies type resolution for a single branch body across all node types
 * that the function handles (ExpressionStatement, VariableDeclaration, empty).
 * -------------------------------------------------------------------------*/

func TestInferBranchBodyType(t *testing.T) {
	engine := NewTypeInferenceEngine()

	tests := []struct {
		name string
		body []ast.Node
		want string
	}{
		{
			name: "empty body → float64",
			body: []ast.Node{},
			want: "float64",
		},
		{
			name: "nil body → float64",
			body: nil,
			want: "float64",
		},
		{
			name: "ExpressionStatement with string literal → string",
			body: []ast.Node{
				&ast.ExpressionStatement{Expression: &ast.Literal{Value: "label"}},
			},
			want: "string",
		},
		{
			name: "ExpressionStatement with float literal → float64",
			body: []ast.Node{
				&ast.ExpressionStatement{Expression: &ast.Literal{Value: 42.0}},
			},
			want: "float64",
		},
		{
			name: "VariableDeclaration with string init → string",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Kind: "var",
					Declarations: []ast.VariableDeclarator{
						{ID: &ast.Identifier{Name: "v"}, Init: &ast.Literal{Value: "hello"}},
					},
				},
			},
			want: "string",
		},
		{
			name: "VariableDeclaration with float init → float64",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Kind: "var",
					Declarations: []ast.VariableDeclarator{
						{ID: &ast.Identifier{Name: "v"}, Init: &ast.Literal{Value: 1.0}},
					},
				},
			},
			want: "float64",
		},
		{
			name: "VariableDeclaration with empty declarations → float64",
			body: []ast.Node{
				&ast.VariableDeclaration{Kind: "var"},
			},
			want: "float64",
		},
		{
			name: "last node wins — multiple statements, last is string",
			body: []ast.Node{
				&ast.ExpressionStatement{Expression: &ast.Literal{Value: 1.0}},
				&ast.ExpressionStatement{Expression: &ast.Literal{Value: "last"}},
			},
			want: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferBranchBodyType(tt.body, engine)
			if got != tt.want {
				t.Errorf("inferBranchBodyType: got %q, want %q", got, tt.want)
			}
		})
	}
}

/* --------------------------------------------------------------------------
 * TestInferAlternateBranchType
 *
 * Verifies that the alternate path correctly delegates to inferIfStatementReturnType
 * for nested if-statements and to inferBranchBodyType otherwise.
 * -------------------------------------------------------------------------*/

func TestInferAlternateBranchType(t *testing.T) {
	engine := NewTypeInferenceEngine()

	tests := []struct {
		name      string
		alternate []ast.Node
		want      string
	}{
		{
			name:      "empty alternate → float64",
			alternate: nil,
			want:      "float64",
		},
		{
			name: "alternate is flat string expression → string",
			alternate: []ast.Node{
				&ast.ExpressionStatement{Expression: &ast.Literal{Value: "fallback"}},
			},
			want: "string",
		},
		{
			name: "alternate is flat float expression → float64",
			alternate: []ast.Node{
				&ast.ExpressionStatement{Expression: &ast.Literal{Value: 0.0}},
			},
			want: "float64",
		},
		{
			name: "alternate is nested IfStatement returning string → string",
			alternate: []ast.Node{
				&ast.IfStatement{
					Test: &ast.Identifier{Name: "x"},
					Consequent: []ast.Node{
						&ast.ExpressionStatement{Expression: &ast.Literal{Value: "yes"}},
					},
					Alternate: []ast.Node{
						&ast.ExpressionStatement{Expression: &ast.Literal{Value: "no"}},
					},
				},
			},
			want: "string",
		},
		{
			name: "alternate is nested IfStatement returning float → float64",
			alternate: []ast.Node{
				&ast.IfStatement{
					Test: &ast.Identifier{Name: "x"},
					Consequent: []ast.Node{
						&ast.ExpressionStatement{Expression: &ast.Literal{Value: 1.0}},
					},
					Alternate: []ast.Node{
						&ast.ExpressionStatement{Expression: &ast.Literal{Value: 0.0}},
					},
				},
			},
			want: "float64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferAlternateBranchType(tt.alternate, engine)
			if got != tt.want {
				t.Errorf("inferAlternateBranchType: got %q, want %q", got, tt.want)
			}
		})
	}
}

/* --------------------------------------------------------------------------
 * TestGenerateIfStatementReturn_Float64
 *
 * generateIfStatementReturn must emit a func() float64 IIFE for if-expressions
 * whose branches all produce numeric values.
 * -------------------------------------------------------------------------*/

func TestGenerateIfStatementReturn_Float64(t *testing.T) {
	gen := newTestGenerator()

	ifStmt := &ast.IfStatement{
		Test: &ast.BinaryExpression{
			Operator: ">",
			Left:     &ast.Identifier{Name: "close"},
			Right:    &ast.Identifier{Name: "open"},
		},
		Consequent: []ast.Node{
			&ast.ExpressionStatement{Expression: &ast.Literal{Value: 1.0}},
		},
		Alternate: []ast.Node{
			&ast.ExpressionStatement{Expression: &ast.Literal{Value: 0.0}},
		},
	}

	arrowCodegen := NewArrowFunctionCodegen(gen)
	code, err := arrowCodegen.generateIfStatementReturn(ifStmt)
	if err != nil {
		t.Fatalf("generateIfStatementReturn: %v", err)
	}
	if !strings.Contains(code, "func() float64") {
		t.Errorf("expected func() float64 IIFE, got: %s", code)
	}
	if !strings.Contains(code, "return ") {
		t.Errorf("expected return statement, got: %s", code)
	}
}

/* --------------------------------------------------------------------------
 * TestGenerateIfStatementReturn_StringFallback
 *
 * When a string-returning if-expression has no else branch, the fallback
 * return must be "" (empty string), not math.NaN(). A float64-returning
 * if-expression must still fall back to math.NaN().
 * -------------------------------------------------------------------------*/

func TestGenerateIfStatementReturn_StringFallback(t *testing.T) {
	tests := []struct {
		name         string
		consequent   ast.Expression
		wantFuncType string
		wantFallback string
		wantAbsent   string
	}{
		{
			name:         "string branch → empty string fallback",
			consequent:   &ast.Literal{Value: "ok"},
			wantFuncType: "func() string",
			wantFallback: `return ""`,
			wantAbsent:   "math.NaN()",
		},
		{
			name:         "float64 branch → NaN fallback",
			consequent:   &ast.Literal{Value: 1.0},
			wantFuncType: "func() float64",
			wantFallback: "return math.NaN()",
			wantAbsent:   `return ""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			// if-statement with no else branch
			ifStmt := &ast.IfStatement{
				Test: &ast.BinaryExpression{
					Operator: "==",
					Left:     &ast.Identifier{Name: "x"},
					Right:    &ast.Literal{Value: 1.0},
				},
				Consequent: []ast.Node{
					&ast.ExpressionStatement{Expression: tt.consequent},
				},
				Alternate: nil,
			}

			arrowCodegen := NewArrowFunctionCodegen(gen)
			code, err := arrowCodegen.generateIfStatementReturn(ifStmt)
			if err != nil {
				t.Fatalf("generateIfStatementReturn: %v", err)
			}
			if !strings.Contains(code, tt.wantFuncType) {
				t.Errorf("expected IIFE type %q in output, got: %s", tt.wantFuncType, code)
			}
			if !strings.Contains(code, tt.wantFallback) {
				t.Errorf("expected fallback %q in output, got: %s", tt.wantFallback, code)
			}
			if strings.Contains(code, tt.wantAbsent) {
				t.Errorf("must not contain %q in output, got: %s", tt.wantAbsent, code)
			}
		})
	}
}
