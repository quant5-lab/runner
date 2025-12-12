package parser

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// TestPostfixExpr_SimpleSubscript verifies basic subscript parsing
func TestPostfixExpr_SimpleSubscript(t *testing.T) {
	pineScript := `//@version=5
indicator("Test")
x = close[1]
`
	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(pineScript))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	// Verify AST structure
	if len(program.Body) < 2 {
		t.Fatalf("Expected at least 2 statements, got %d", len(program.Body))
	}

	varDecl, ok := program.Body[1].(*ast.VariableDeclaration)
	if !ok {
		t.Fatalf("Expected VariableDeclaration, got %T", program.Body[1])
	}

	memberExpr, ok := varDecl.Declarations[0].Init.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("Expected MemberExpression for close[1], got %T", varDecl.Declarations[0].Init)
	}

	if !memberExpr.Computed {
		t.Error("Expected computed property (subscript)")
	}

	ident, ok := memberExpr.Object.(*ast.Identifier)
	if !ok || ident.Name != "close" {
		t.Errorf("Expected Object to be Identifier 'close', got %T", memberExpr.Object)
	}
}

// TestPostfixExpr_FunctionCallWithSubscript verifies func()[offset] parsing
func TestPostfixExpr_FunctionCallWithSubscript(t *testing.T) {
	pineScript := `//@version=5
indicator("Test")
pivot = pivothigh(5, 5)[1]
`
	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(pineScript))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	varDecl := program.Body[1].(*ast.VariableDeclaration)
	memberExpr, ok := varDecl.Declarations[0].Init.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("Expected MemberExpression for pivothigh()[1], got %T", varDecl.Declarations[0].Init)
	}

	// Verify Object is CallExpression
	callExpr, ok := memberExpr.Object.(*ast.CallExpression)
	if !ok {
		t.Fatalf("Expected Object to be CallExpression, got %T", memberExpr.Object)
	}

	// Verify CallExpression callee
	callee, ok := callExpr.Callee.(*ast.Identifier)
	if !ok || callee.Name != "pivothigh" {
		t.Errorf("Expected callee 'pivothigh', got %v", callExpr.Callee)
	}

	// Verify subscript
	if !memberExpr.Computed {
		t.Error("Expected computed property (subscript)")
	}

	literal, ok := memberExpr.Property.(*ast.Literal)
	if !ok {
		t.Fatalf("Expected Property to be Literal, got %T", memberExpr.Property)
	}

	if literal.Value != float64(1) {
		t.Errorf("Expected subscript [1], got %v", literal.Value)
	}
}

// TestPostfixExpr_NestedSubscript verifies fixnan(func()[1]) parsing
func TestPostfixExpr_NestedSubscript(t *testing.T) {
	pineScript := `//@version=5
indicator("Test")
filled = fixnan(pivothigh(5, 5)[1])
`
	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(pineScript))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	varDecl := program.Body[1].(*ast.VariableDeclaration)

	// Outer call: fixnan(...)
	outerCall, ok := varDecl.Declarations[0].Init.(*ast.CallExpression)
	if !ok {
		t.Fatalf("Expected CallExpression for fixnan(...), got %T", varDecl.Declarations[0].Init)
	}

	outerCallee, ok := outerCall.Callee.(*ast.Identifier)
	if !ok || outerCallee.Name != "fixnan" {
		t.Errorf("Expected outer callee 'fixnan', got %v", outerCall.Callee)
	}

	// Argument: pivothigh()[1]
	memberExpr, ok := outerCall.Arguments[0].(*ast.MemberExpression)
	if !ok {
		t.Fatalf("Expected MemberExpression for pivothigh()[1], got %T", outerCall.Arguments[0])
	}

	// Inner call: pivothigh(5, 5)
	innerCall, ok := memberExpr.Object.(*ast.CallExpression)
	if !ok {
		t.Fatalf("Expected Object to be CallExpression, got %T", memberExpr.Object)
	}

	innerCallee, ok := innerCall.Callee.(*ast.Identifier)
	if !ok || innerCallee.Name != "pivothigh" {
		t.Errorf("Expected inner callee 'pivothigh', got %v", innerCall.Callee)
	}

	// Verify subscript [1]
	if !memberExpr.Computed {
		t.Error("Expected computed property (subscript)")
	}

	literal, ok := memberExpr.Property.(*ast.Literal)
	if !ok || literal.Value != float64(1) {
		t.Errorf("Expected subscript [1], got %v", memberExpr.Property)
	}
}

// TestPostfixExpr_NamespacedFunctionWithSubscript verifies ta.sma()[1] parsing
func TestPostfixExpr_NamespacedFunctionWithSubscript(t *testing.T) {
	pineScript := `//@version=5
indicator("Test")
x = ta.sma(close, 20)[1]
`
	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(pineScript))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	varDecl := program.Body[1].(*ast.VariableDeclaration)
	memberExpr, ok := varDecl.Declarations[0].Init.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("Expected MemberExpression for ta.sma()[1], got %T", varDecl.Declarations[0].Init)
	}

	// Verify Object is CallExpression
	callExpr, ok := memberExpr.Object.(*ast.CallExpression)
	if !ok {
		t.Fatalf("Expected Object to be CallExpression, got %T", memberExpr.Object)
	}

	// Verify callee is ta.sma (MemberExpression)
	calleeMember, ok := callExpr.Callee.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("Expected callee to be MemberExpression (ta.sma), got %T", callExpr.Callee)
	}

	obj, ok := calleeMember.Object.(*ast.Identifier)
	if !ok || obj.Name != "ta" {
		t.Errorf("Expected namespace 'ta', got %v", calleeMember.Object)
	}

	prop, ok := calleeMember.Property.(*ast.Identifier)
	if !ok || prop.Name != "sma" {
		t.Errorf("Expected function 'sma', got %v", calleeMember.Property)
	}
}

// TestPostfixExpr_IdentifierWithoutSubscript verifies plain identifiers still work
func TestPostfixExpr_IdentifierWithoutSubscript(t *testing.T) {
	pineScript := `//@version=5
indicator("Test")
x = close
`
	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(pineScript))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	varDecl := program.Body[1].(*ast.VariableDeclaration)

	// Should be plain Identifier, not MemberExpression
	ident, ok := varDecl.Declarations[0].Init.(*ast.Identifier)
	if !ok {
		t.Fatalf("Expected Identifier for plain 'close', got %T", varDecl.Declarations[0].Init)
	}

	if ident.Name != "close" {
		t.Errorf("Expected identifier 'close', got %s", ident.Name)
	}
}

// TestPostfixExpr_CallWithoutSubscript verifies plain function calls still work
func TestPostfixExpr_CallWithoutSubscript(t *testing.T) {
	pineScript := `//@version=5
indicator("Test")
x = ta.sma(close, 20)
`
	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(pineScript))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	varDecl := program.Body[1].(*ast.VariableDeclaration)

	// Should be CallExpression, not MemberExpression
	callExpr, ok := varDecl.Declarations[0].Init.(*ast.CallExpression)
	if !ok {
		t.Fatalf("Expected CallExpression for ta.sma(...), got %T", varDecl.Declarations[0].Init)
	}

	// Verify it's ta.sma
	calleeMember, ok := callExpr.Callee.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("Expected callee to be MemberExpression, got %T", callExpr.Callee)
	}

	obj, ok := calleeMember.Object.(*ast.Identifier)
	if !ok || obj.Name != "ta" {
		t.Errorf("Expected namespace 'ta', got %v", calleeMember.Object)
	}
}

// TestPostfixExpr_VariableOffsetSubscript verifies dynamic offset like close[length]
func TestPostfixExpr_VariableOffsetSubscript(t *testing.T) {
	pineScript := `//@version=5
indicator("Test")
offset = 5
x = close[offset]
`
	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(pineScript))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	varDecl := program.Body[2].(*ast.VariableDeclaration)
	memberExpr, ok := varDecl.Declarations[0].Init.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("Expected MemberExpression for close[offset], got %T", varDecl.Declarations[0].Init)
	}

	// Verify Property is Identifier (variable offset)
	offsetIdent, ok := memberExpr.Property.(*ast.Identifier)
	if !ok {
		t.Fatalf("Expected Property to be Identifier, got %T", memberExpr.Property)
	}

	if offsetIdent.Name != "offset" {
		t.Errorf("Expected offset variable 'offset', got %s", offsetIdent.Name)
	}
}

// TestPostfixExpr_InCondition verifies subscripts work in conditions
func TestPostfixExpr_InCondition(t *testing.T) {
	pineScript := `//@version=5
indicator("Test")
signal = close[0] > close[1] ? 1 : 0
`
	p, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(pineScript))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	varDecl := program.Body[1].(*ast.VariableDeclaration)
	condExpr, ok := varDecl.Declarations[0].Init.(*ast.ConditionalExpression)
	if !ok {
		t.Fatalf("Expected ConditionalExpression, got %T", varDecl.Declarations[0].Init)
	}

	// Verify test condition has subscripts
	binExpr, ok := condExpr.Test.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("Expected BinaryExpression in test, got %T", condExpr.Test)
	}

	// Left: close[0]
	leftMember, ok := binExpr.Left.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("Expected MemberExpression for close[0], got %T", binExpr.Left)
	}
	if !leftMember.Computed {
		t.Error("Expected computed subscript for close[0]")
	}

	// Right: close[1]
	rightMember, ok := binExpr.Right.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("Expected MemberExpression for close[1], got %T", binExpr.Right)
	}
	if !rightMember.Computed {
		t.Error("Expected computed subscript for close[1]")
	}
}
