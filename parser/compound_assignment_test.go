package parser

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// TestCompoundAssignment_Desugar verifies that each compound assignment operator
// parses and is desugared into the canonical `x := x OP rhs` form.
//
// The converter must produce:
//   - A VariableDeclaration with kind "var" (same as Reassignment)
//   - A BinaryExpression as initialiser with the left operand being the same
//     identifier as the assignment target
//   - The correct binary operator stripped of its trailing `=`
func TestCompoundAssignment_Desugar(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		wantVarName string
		wantOp      string
	}{
		{
			name:        "+= desugars to x := x + rhs",
			source:      "count += 1",
			wantVarName: "count",
			wantOp:      "+",
		},
		{
			name:        "-= desugars to x := x - rhs",
			source:      "total -= fee",
			wantVarName: "total",
			wantOp:      "-",
		},
		{
			name:        "*= desugars to x := x * rhs",
			source:      "factor *= 2",
			wantVarName: "factor",
			wantOp:      "*",
		},
		{
			name:        "/= desugars to x := x / rhs",
			source:      "ratio /= divisor",
			wantVarName: "ratio",
			wantOp:      "/",
		},
		{
			name:        "%= desugars to x := x % rhs",
			source:      "remainder %= 10",
			wantVarName: "remainder",
			wantOp:      "%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog := parseOne(t, tt.source)

			if len(prog.Body) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(prog.Body))
			}

			decl, ok := prog.Body[0].(*ast.VariableDeclaration)
			if !ok {
				t.Fatalf("expected VariableDeclaration, got %T", prog.Body[0])
			}
			if decl.Kind != "var" {
				t.Errorf("expected kind=var, got %q", decl.Kind)
			}
			if len(decl.Declarations) != 1 {
				t.Fatalf("expected 1 declarator, got %d", len(decl.Declarations))
			}

			d := decl.Declarations[0]
			id, ok := d.ID.(*ast.Identifier)
			if !ok {
				t.Fatalf("expected Identifier ID, got %T", d.ID)
			}
			if id.Name != tt.wantVarName {
				t.Errorf("target name: want %q, got %q", tt.wantVarName, id.Name)
			}

			bin, ok := d.Init.(*ast.BinaryExpression)
			if !ok {
				t.Fatalf("expected BinaryExpression init, got %T", d.Init)
			}
			if bin.Operator != tt.wantOp {
				t.Errorf("operator: want %q, got %q", tt.wantOp, bin.Operator)
			}

			leftId, ok := bin.Left.(*ast.Identifier)
			if !ok {
				t.Fatalf("expected Identifier left operand, got %T", bin.Left)
			}
			if leftId.Name != tt.wantVarName {
				t.Errorf("left operand: want %q, got %q", tt.wantVarName, leftId.Name)
			}
		})
	}
}

// TestCompoundAssignment_InsideIfBody verifies that compound assignments are parsed
// correctly when they appear inside an indented if-body — the same context that
// caused the original parse failure in support_resistance_pivot_levels.pine.
func TestCompoundAssignment_InsideIfBody(t *testing.T) {
	source := "if close > open\n    idx += 1"

	prog := parseOne(t, source)
	if len(prog.Body) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Body))
	}

	ifStmt, ok := prog.Body[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement, got %T", prog.Body[0])
	}
	if len(ifStmt.Consequent) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(ifStmt.Consequent))
	}

	decl, ok := ifStmt.Consequent[0].(*ast.VariableDeclaration)
	if !ok {
		t.Fatalf("expected VariableDeclaration in if body, got %T", ifStmt.Consequent[0])
	}

	bin, ok := decl.Declarations[0].Init.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("expected BinaryExpression, got %T", decl.Declarations[0].Init)
	}
	if bin.Operator != "+" {
		t.Errorf("operator: want +, got %q", bin.Operator)
	}
}

// TestCompoundAssignment_NoRegressionOnExistingOperators verifies that the new
// compound-operator tokens do not interfere with existing two-character tokens
// that share a leading character (e.g. `:=`, `==`, `!=`).
func TestCompoundAssignment_NoRegressionOnExistingOperators(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   ":= reassignment still parses",
			source: "x := close + 1",
		},
		{
			name:   "== comparison still parses",
			source: "y = close == open",
		},
		{
			name:   "!= comparison still parses",
			source: "z = close != open",
		},
		{
			name:   "+= adjacent to - does not confuse lexer",
			source: "a += b - 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(parseOne(t, tt.source).Body) == 0 {
				t.Error("expected at least one statement")
			}
		})
	}
}
