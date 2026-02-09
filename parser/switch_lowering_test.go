package parser

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestSwitchLowering_Form1WithSubject(t *testing.T) {
	source := `x = switch close
    1 =>
        val1
    2 =>
        val2
    =>
        defaultVal`

	program := parseSwitchSource(t, source)

	varDecl, ok := program.Body[0].(*ast.VariableDeclaration)
	if !ok {
		t.Fatalf("expected VariableDeclaration, got %T", program.Body[0])
	}

	ifStmt, ok := varDecl.Declarations[0].Init.(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement (lowered switch), got %T", varDecl.Declarations[0].Init)
	}

	assertBinaryTest(t, ifStmt, "==", "first case")

	if len(ifStmt.Alternate) != 1 {
		t.Fatalf("expected 1 alternate node (else-if chain), got %d", len(ifStmt.Alternate))
	}

	nestedIf, ok := ifStmt.Alternate[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected nested IfStatement in alternate, got %T", ifStmt.Alternate[0])
	}

	assertBinaryTest(t, nestedIf, "==", "second case")

	if len(nestedIf.Alternate) != 1 {
		t.Fatalf("expected 1 default alternate node, got %d", len(nestedIf.Alternate))
	}
}

func TestSwitchLowering_Form2WithoutSubject(t *testing.T) {
	source := `x = switch
    close > open =>
        1
    close < open =>
        2`

	program := parseSwitchSource(t, source)

	varDecl := program.Body[0].(*ast.VariableDeclaration)
	ifStmt, ok := varDecl.Declarations[0].Init.(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement (lowered switch), got %T", varDecl.Declarations[0].Init)
	}

	if _, isBinary := ifStmt.Test.(*ast.BinaryExpression); !isBinary {
		t.Fatalf("form 2 should use condition directly as BinaryExpression, got %T", ifStmt.Test)
	}

	if len(ifStmt.Alternate) != 1 {
		t.Fatalf("expected 1 alternate (else-if), got %d", len(ifStmt.Alternate))
	}

	nestedIf := ifStmt.Alternate[0].(*ast.IfStatement)
	if len(nestedIf.Alternate) != 0 {
		t.Fatalf("expected 0 alternate (no default), got %d", len(nestedIf.Alternate))
	}
}

func TestSwitchLowering_StatementLevel(t *testing.T) {
	source := `switch action
    1 =>
        doSomething()
    2 =>
        doOther()`

	program := parseSwitchSource(t, source)

	ifStmt, ok := program.Body[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement at statement level, got %T", program.Body[0])
	}

	assertBinaryTest(t, ifStmt, "==", "statement-level switch")
}

func TestSwitchLowering_DefaultOnly(t *testing.T) {
	source := `x = switch
    =>
        42`

	program := parseSwitchSource(t, source)

	varDecl := program.Body[0].(*ast.VariableDeclaration)
	ifStmt, ok := varDecl.Declarations[0].Init.(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement, got %T", varDecl.Declarations[0].Init)
	}

	lit, ok := ifStmt.Test.(*ast.Literal)
	if !ok || lit.Value != true {
		t.Fatalf("default-only switch should have true test, got %v", ifStmt.Test)
	}
}

func TestSwitchLowering_SingleCase(t *testing.T) {
	source := `x = switch val
    1 =>
        result`

	program := parseSwitchSource(t, source)

	varDecl := program.Body[0].(*ast.VariableDeclaration)
	ifStmt, ok := varDecl.Declarations[0].Init.(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement, got %T", varDecl.Declarations[0].Init)
	}

	assertBinaryTest(t, ifStmt, "==", "single case")

	if len(ifStmt.Alternate) != 0 {
		t.Fatalf("single case without default should have empty alternate, got %d", len(ifStmt.Alternate))
	}
}

func TestSwitchLowering_MultipleCasesWithDefault(t *testing.T) {
	source := `x = switch mode
    1 =>
        a
    2 =>
        b
    3 =>
        c
    =>
        d`

	program := parseSwitchSource(t, source)

	varDecl := program.Body[0].(*ast.VariableDeclaration)
	ifStmt := varDecl.Declarations[0].Init.(*ast.IfStatement)

	depth := countIfChainDepth(ifStmt)
	if depth != 3 {
		t.Fatalf("expected if-chain depth 3 (3 cases), got %d", depth)
	}

	lastIf := getLastIfInChain(ifStmt)
	if len(lastIf.Alternate) != 1 {
		t.Fatalf("expected default body in last alternate, got %d nodes", len(lastIf.Alternate))
	}
}

func parseSwitchSource(t *testing.T, source string) *ast.Program {
	t.Helper()

	p, err := NewParser()
	if err != nil {
		t.Fatalf("parser creation: %v", err)
	}

	script, err := p.ParseString("", source)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("conversion: %v", err)
	}

	return program
}

func assertBinaryTest(t *testing.T, ifStmt *ast.IfStatement, operator string, context string) {
	t.Helper()

	binExpr, ok := ifStmt.Test.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("[%s] expected BinaryExpression test, got %T", context, ifStmt.Test)
	}
	if binExpr.Operator != operator {
		t.Fatalf("[%s] expected operator %q, got %q", context, operator, binExpr.Operator)
	}
}

func countIfChainDepth(ifStmt *ast.IfStatement) int {
	depth := 1
	current := ifStmt
	for len(current.Alternate) == 1 {
		nested, ok := current.Alternate[0].(*ast.IfStatement)
		if !ok {
			break
		}
		depth++
		current = nested
	}
	return depth
}

func getLastIfInChain(ifStmt *ast.IfStatement) *ast.IfStatement {
	current := ifStmt
	for len(current.Alternate) == 1 {
		nested, ok := current.Alternate[0].(*ast.IfStatement)
		if !ok {
			break
		}
		current = nested
	}
	return current
}
