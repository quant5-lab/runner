package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* --------------------------------------------------------------------------
 * Helpers
 * -------------------------------------------------------------------------*/

func newScanner(constants map[string]interface{}, variables map[string]string) *UDFCallSiteArgumentTypeScanner {
	if constants == nil {
		constants = map[string]interface{}{}
	}
	if variables == nil {
		variables = map[string]string{}
	}
	return NewUDFCallSiteArgumentTypeScanner(constants, variables)
}

// buildProgram wraps statements in an *ast.Program.
func buildProgram(stmts ...ast.Node) *ast.Program {
	return &ast.Program{Body: stmts}
}

// callStmt builds an ExpressionStatement that calls funcName with the given args.
func callStmt(funcName string, args ...ast.Expression) *ast.ExpressionStatement {
	return &ast.ExpressionStatement{
		Expression: &ast.CallExpression{
			Callee:    &ast.Identifier{Name: funcName},
			Arguments: args,
		},
	}
}

// varDeclInit wraps a call in a VariableDeclaration initialiser.
func varDeclInit(varName, funcName string, args ...ast.Expression) *ast.VariableDeclaration {
	return &ast.VariableDeclaration{
		Kind: "var",
		Declarations: []ast.VariableDeclarator{
			{ID: &ast.Identifier{Name: varName}, Init: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: funcName},
				Arguments: args,
			}},
		},
	}
}

// strLitExpr returns an *ast.Literal with a string value.
func strLitExpr(s string) *ast.Literal { return &ast.Literal{Value: s} }

// numLitExpr returns an *ast.Literal with a float64 value.
func numLitExpr(f float64) *ast.Literal { return &ast.Literal{Value: f} }

// boolLitExpr returns an *ast.Literal with a bool value.
func boolLitExpr(b bool) *ast.Literal { return &ast.Literal{Value: b} }

/* --------------------------------------------------------------------------
 * TestUDFCallSiteArgumentTypeScanner_InferExpressionType
 *
 * Verifies type inference for individual argument expressions, covering all
 * resolution paths: literal, constant identifier, variable type identifier,
 * and the scalar fallback.
 * -------------------------------------------------------------------------*/

func TestUDFCallSiteArgumentTypeScanner_InferExpressionType(t *testing.T) {
	tests := []struct {
		name      string
		expr      ast.Expression
		constants map[string]interface{}
		variables map[string]string
		want      ParameterUsageType
	}{
		// --- string literal ---
		{
			name: "string literal → String",
			expr: strLitExpr("0900-1700"),
			want: ParameterUsageString,
		},
		{
			name: "empty string literal → String",
			expr: strLitExpr(""),
			want: ParameterUsageString,
		},
		// --- numeric / bool literals ---
		{
			name: "float literal → Scalar",
			expr: numLitExpr(3.14),
			want: ParameterUsageScalar,
		},
		{
			name: "bool literal → Scalar",
			expr: boolLitExpr(true),
			want: ParameterUsageScalar,
		},
		// --- identifier: constant lookup ---
		{
			name:      "identifier whose constant is string → String",
			expr:      &ast.Identifier{Name: "SESS"},
			constants: map[string]interface{}{"SESS": "0900-1700"},
			want:      ParameterUsageString,
		},
		{
			name:      "identifier whose constant is float → Scalar",
			expr:      &ast.Identifier{Name: "LENGTH"},
			constants: map[string]interface{}{"LENGTH": 14.0},
			want:      ParameterUsageScalar,
		},
		{
			name:      "identifier whose constant is bool → Scalar",
			expr:      &ast.Identifier{Name: "FILTER"},
			constants: map[string]interface{}{"FILTER": false},
			want:      ParameterUsageScalar,
		},
		// --- identifier: variable type lookup ---
		{
			name:      "identifier whose variable type is string → String",
			expr:      &ast.Identifier{Name: "label"},
			variables: map[string]string{"label": "string"},
			want:      ParameterUsageString,
		},
		{
			name:      "identifier whose variable type is color → String",
			expr:      &ast.Identifier{Name: "col"},
			variables: map[string]string{"col": "color"},
			want:      ParameterUsageString,
		},
		{
			name:      "identifier whose variable type is float → Scalar",
			expr:      &ast.Identifier{Name: "src"},
			variables: map[string]string{"src": "float"},
			want:      ParameterUsageScalar,
		},
		{
			name:      "identifier whose variable type is bool → Scalar",
			expr:      &ast.Identifier{Name: "flag"},
			variables: map[string]string{"flag": "bool"},
			want:      ParameterUsageScalar,
		},
		// --- identifier: unknown (no constant, no variable entry) ---
		{
			name: "unknown identifier → Scalar fallback",
			expr: &ast.Identifier{Name: "mystery"},
			want: ParameterUsageScalar,
		},
		// --- non-identifier expressions ---
		{
			name: "binary expression → Scalar",
			expr: &ast.BinaryExpression{
				Operator: "+",
				Left:     &ast.Identifier{Name: "a"},
				Right:    numLitExpr(1),
			},
			want: ParameterUsageScalar,
		},
		{
			name: "call expression → Scalar",
			expr: &ast.CallExpression{Callee: &ast.Identifier{Name: "ta.sma"}},
			want: ParameterUsageScalar,
		},
		// --- input.source tag must not be classified as string ---
		{
			name:      "input.source tag in constants with float variable → Scalar",
			expr:      &ast.Identifier{Name: "src"},
			constants: map[string]interface{}{"src": "input.source"},
			variables: map[string]string{"src": "float"},
			want:      ParameterUsageScalar,
		},
		// --- non-identifier, non-literal expressions ---
		{
			name: "member expression (color.red) → Scalar fallback",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Identifier{Name: "red"},
			},
			want: ParameterUsageScalar,
		},
		// --- real string constant with no variable entry: still String ---
		{
			name:      "real string constant and no variable entry → String",
			expr:      &ast.Identifier{Name: "sess"},
			constants: map[string]interface{}{"sess": "0900-1700"},
			want:      ParameterUsageString,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newScanner(tt.constants, tt.variables)
			got := s.inferExpressionType(tt.expr)
			if got != tt.want {
				t.Errorf("inferExpressionType: got %v, want %v", got, tt.want)
			}
		})
	}
}

/* --------------------------------------------------------------------------
 * TestUDFCallSiteArgumentTypeScanner_ScanProgram_ArgumentTypes
 *
 * Verifies that ScanProgram correctly identifies UDF calls and captures the
 * per-position types from the argument expressions.
 * -------------------------------------------------------------------------*/

func TestUDFCallSiteArgumentTypeScanner_ScanProgram_ArgumentTypes(t *testing.T) {
	tests := []struct {
		name      string
		variables map[string]string
		constants map[string]interface{}
		program   *ast.Program
		wantTypes map[string][]ParameterUsageType // funcName → per-position types
	}{
		{
			name:      "single string arg at position 0",
			variables: map[string]string{"fn": "function"},
			program:   buildProgram(callStmt("fn", strLitExpr("0900-1700"))),
			wantTypes: map[string][]ParameterUsageType{
				"fn": {ParameterUsageString},
			},
		},
		{
			name:      "scalar arg at position 0",
			variables: map[string]string{"fn": "function"},
			program:   buildProgram(callStmt("fn", numLitExpr(14))),
			wantTypes: map[string][]ParameterUsageType{
				"fn": {ParameterUsageScalar},
			},
		},
		{
			name:      "mixed string and scalar args",
			variables: map[string]string{"fn": "function"},
			program:   buildProgram(callStmt("fn", strLitExpr("sess"), numLitExpr(1))),
			wantTypes: map[string][]ParameterUsageType{
				"fn": {ParameterUsageString, ParameterUsageScalar},
			},
		},
		{
			name:      "string constant identifier at position 1",
			variables: map[string]string{"fn": "function"},
			constants: map[string]interface{}{"LABEL": "hello"},
			program: buildProgram(callStmt("fn",
				numLitExpr(10),
				&ast.Identifier{Name: "LABEL"},
			)),
			wantTypes: map[string][]ParameterUsageType{
				"fn": {ParameterUsageScalar, ParameterUsageString},
			},
		},
		{
			name:      "string variable type identifier",
			variables: map[string]string{"fn": "function", "mode": "string"},
			program:   buildProgram(callStmt("fn", &ast.Identifier{Name: "mode"})),
			wantTypes: map[string][]ParameterUsageType{
				"fn": {ParameterUsageString},
			},
		},
		{
			name:      "multiple distinct UDFs each independently recorded",
			variables: map[string]string{"f1": "function", "f2": "function"},
			program: buildProgram(
				callStmt("f1", strLitExpr("x"), numLitExpr(5)),
				callStmt("f2", numLitExpr(1)),
			),
			wantTypes: map[string][]ParameterUsageType{
				"f1": {ParameterUsageString, ParameterUsageScalar},
				"f2": {ParameterUsageScalar},
			},
		},
		{
			name:      "non-UDF call is not recorded",
			variables: map[string]string{"fn": "function"},
			program:   buildProgram(callStmt("sma", numLitExpr(14))),
			wantTypes: map[string][]ParameterUsageType{},
		},
		{
			name:      "zero-arg UDF produces empty type slice",
			variables: map[string]string{"myFn": "function"},
			program:   buildProgram(callStmt("myFn")),
			wantTypes: map[string][]ParameterUsageType{"myFn": {}},
		},
		{
			name:      "UDF inside VariableDeclaration initialiser",
			variables: map[string]string{"calc": "function"},
			program:   buildProgram(varDeclInit("result", "calc", strLitExpr("A"), numLitExpr(2))),
			wantTypes: map[string][]ParameterUsageType{
				"calc": {ParameterUsageString, ParameterUsageScalar},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newScanner(tt.constants, tt.variables)
			got := s.ScanProgram(tt.program)
			if len(got) != len(tt.wantTypes) {
				t.Fatalf("result map length: got %d, want %d (got=%v)", len(got), len(tt.wantTypes), got)
			}
			for funcName, wantSlice := range tt.wantTypes {
				gotSlice, ok := got[funcName]
				if !ok {
					t.Errorf("missing entry for UDF %q", funcName)
					continue
				}
				if len(gotSlice) != len(wantSlice) {
					t.Errorf("UDF %q: arg count got %d, want %d", funcName, len(gotSlice), len(wantSlice))
					continue
				}
				for i, wantType := range wantSlice {
					if gotSlice[i] != wantType {
						t.Errorf("UDF %q pos %d: got %v, want %v", funcName, i, gotSlice[i], wantType)
					}
				}
			}
		})
	}
}

/* --------------------------------------------------------------------------
 * TestUDFCallSiteArgumentTypeScanner_FirstCallSiteWins
 *
 * The scanner records only the first call site per UDF. Subsequent calls with
 * different argument types must not overwrite the first result.
 * -------------------------------------------------------------------------*/

func TestUDFCallSiteArgumentTypeScanner_FirstCallSiteWins(t *testing.T) {
	variables := map[string]string{"fn": "function"}

	// First call: string arg; second call: scalar arg.
	program := buildProgram(
		callStmt("fn", strLitExpr("first")),
		callStmt("fn", numLitExpr(99)),
	)

	got := newScanner(nil, variables).ScanProgram(program)

	if len(got["fn"]) != 1 {
		t.Fatalf("expected 1 type entry, got %d", len(got["fn"]))
	}
	if got["fn"][0] != ParameterUsageString {
		t.Errorf("first-call-site: got %v, want ParameterUsageString", got["fn"][0])
	}
}

/* --------------------------------------------------------------------------
 * TestUDFCallSiteArgumentTypeScanner_NestedCallsScanned
 *
 * UDF calls that appear inside the arguments of another call or inside
 * if-statement bodies are still discovered during the traversal.
 * -------------------------------------------------------------------------*/

func TestUDFCallSiteArgumentTypeScanner_NestedCallsScanned(t *testing.T) {
	t.Run("UDF call nested inside argument of another call", func(t *testing.T) {
		variables := map[string]string{"inner": "function"}

		// outer(inner("sess")) — outer is a builtin, inner is UDF
		outerCall := &ast.ExpressionStatement{
			Expression: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "plot"},
				Arguments: []ast.Expression{
					&ast.CallExpression{
						Callee:    &ast.Identifier{Name: "inner"},
						Arguments: []ast.Expression{strLitExpr("sess")},
					},
				},
			},
		}

		got := newScanner(nil, variables).ScanProgram(buildProgram(outerCall))
		if got["inner"][0] != ParameterUsageString {
			t.Errorf("inner call inside argument: got %v, want ParameterUsageString", got["inner"][0])
		}
	})

	t.Run("UDF call inside if-statement consequent body", func(t *testing.T) {
		variables := map[string]string{"fn": "function"}

		ifStmt := &ast.IfStatement{
			Test: &ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Identifier{Name: "open"},
			},
			Consequent: []ast.Node{callStmt("fn", strLitExpr("body"))},
		}

		got := newScanner(nil, variables).ScanProgram(buildProgram(ifStmt))
		if got["fn"][0] != ParameterUsageString {
			t.Errorf("UDF in if body: got %v, want ParameterUsageString", got["fn"][0])
		}
	})
}

/* --------------------------------------------------------------------------
 * TestUDFCallSiteArgumentTypeScanner_EmptyAndNilSafety
 * -------------------------------------------------------------------------*/

func TestUDFCallSiteArgumentTypeScanner_EmptyAndNilSafety(t *testing.T) {
	t.Run("empty program returns empty map", func(t *testing.T) {
		got := newScanner(nil, nil).ScanProgram(buildProgram())
		if len(got) != 0 {
			t.Errorf("expected empty result, got %v", got)
		}
	})

	t.Run("program with no calls returns empty map", func(t *testing.T) {
		program := buildProgram(&ast.VariableDeclaration{
			Kind: "var",
			Declarations: []ast.VariableDeclarator{
				{ID: &ast.Identifier{Name: "x"}, Init: numLitExpr(1)},
			},
		})
		got := newScanner(nil, nil).ScanProgram(program)
		if len(got) != 0 {
			t.Errorf("expected empty result for program with no calls, got %v", got)
		}
	})
}
