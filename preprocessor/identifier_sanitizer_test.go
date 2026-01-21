package preprocessor

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestIdentifierSanitizer_VariableDeclaration(t *testing.T) {
	tests := []struct {
		name            string
		originalName    string
		expectedName    string
		shouldBeRenamed bool
	}{
		{"len keyword", "len", "len_", true},
		{"type keyword", "type", "type_", true},
		{"regular name", "length", "length", false},
		{"src name", "src", "src", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID:   &ast.Identifier{Name: tt.originalName},
								Init: &ast.Literal{Value: 14},
							},
						},
					},
				},
			}

			sanitizer := NewIdentifierSanitizer()
			result, err := sanitizer.Transform(program)
			if err != nil {
				t.Fatalf("Transform failed: %v", err)
			}

			varDecl := result.Body[0].(*ast.VariableDeclaration)
			id := varDecl.Declarations[0].ID.(*ast.Identifier)

			if id.Name != tt.expectedName {
				t.Errorf("Expected %q, got %q", tt.expectedName, id.Name)
			}

			renamed := sanitizer.GetRenamedIdentifiers()
			_, wasRenamed := renamed[tt.originalName]
			if wasRenamed != tt.shouldBeRenamed {
				t.Errorf("Expected renamed=%v, got %v", tt.shouldBeRenamed, wasRenamed)
			}
		})
	}
}

func TestIdentifierSanitizer_IdentifierReference(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "len"},
						Init: &ast.Literal{Value: 14},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "result"},
						Init: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "ta"},
								Property: &ast.Identifier{Name: "rsi"},
							},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "close"},
								&ast.Identifier{Name: "len"},
							},
						},
					},
				},
			},
		},
	}

	sanitizer := NewIdentifierSanitizer()
	result, err := sanitizer.Transform(program)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	varDecl1 := result.Body[0].(*ast.VariableDeclaration)
	id1 := varDecl1.Declarations[0].ID.(*ast.Identifier)
	if id1.Name != "len_" {
		t.Errorf("Declaration: expected %q, got %q", "len_", id1.Name)
	}

	varDecl2 := result.Body[1].(*ast.VariableDeclaration)
	call := varDecl2.Declarations[0].Init.(*ast.CallExpression)
	lenArg := call.Arguments[1].(*ast.Identifier)
	if lenArg.Name != "len_" {
		t.Errorf("Reference: expected %q, got %q", "len_", lenArg.Name)
	}
}

func TestIdentifierSanitizer_TupleDestructuring(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.ArrayPattern{
							Elements: []ast.Identifier{
								{Name: "map"},
								{Name: "type"},
								{Name: "value"},
							},
						},
						Init: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "ta"},
								Property: &ast.Identifier{Name: "rsi"},
							},
						},
					},
				},
			},
		},
	}

	sanitizer := NewIdentifierSanitizer()
	result, err := sanitizer.Transform(program)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	varDecl := result.Body[0].(*ast.VariableDeclaration)
	arrayPattern := varDecl.Declarations[0].ID.(*ast.ArrayPattern)

	expected := []string{"map_", "type_", "value"}
	for i, elem := range arrayPattern.Elements {
		if elem.Name != expected[i] {
			t.Errorf("Element %d: expected %q, got %q", i, expected[i], elem.Name)
		}
	}
}

func TestIdentifierSanitizer_ArrowFunction(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "myFunc"},
						Init: &ast.ArrowFunctionExpression{
							Params: []ast.Identifier{
								{Name: "len"},
								{Name: "type"},
							},
							Body: []ast.Node{
								&ast.ExpressionStatement{
									Expression: &ast.BinaryExpression{
										Left:     &ast.Identifier{Name: "len"},
										Operator: "*",
										Right:    &ast.Identifier{Name: "type"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	sanitizer := NewIdentifierSanitizer()
	result, err := sanitizer.Transform(program)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	varDecl := result.Body[0].(*ast.VariableDeclaration)
	arrow := varDecl.Declarations[0].Init.(*ast.ArrowFunctionExpression)

	if arrow.Params[0].Name != "len_" {
		t.Errorf("Param 1: expected %q, got %q", "len_", arrow.Params[0].Name)
	}

	if arrow.Params[1].Name != "type_" {
		t.Errorf("Param 2: expected %q, got %q", "type_", arrow.Params[1].Name)
	}

	exprStmt := arrow.Body[0].(*ast.ExpressionStatement)
	body := exprStmt.Expression.(*ast.BinaryExpression)
	left := body.Left.(*ast.Identifier)
	right := body.Right.(*ast.Identifier)

	if left.Name != "len_" {
		t.Errorf("Body left: expected %q, got %q", "len_", left.Name)
	}
	if right.Name != "type_" {
		t.Errorf("Body right: expected %q, got %q", "type_", right.Name)
	}
}

func TestIdentifierSanitizer_ObjectPropertyNotRenamed(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "config"},
						Init: &ast.ObjectExpression{
							Properties: []ast.Property{
								{
									Key:   &ast.Identifier{Name: "type"},
									Value: &ast.Literal{Value: "test"},
								},
							},
						},
					},
				},
			},
		},
	}

	sanitizer := NewIdentifierSanitizer()
	result, err := sanitizer.Transform(program)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	varDecl := result.Body[0].(*ast.VariableDeclaration)
	obj := varDecl.Declarations[0].Init.(*ast.ObjectExpression)
	key := obj.Properties[0].Key.(*ast.Identifier)

	if key.Name != "type" {
		t.Errorf("Object key should not be renamed: expected %q, got %q", "type", key.Name)
	}
}

func TestIdentifierSanitizer_ConsistentRenaming(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "len"},
						Init: &ast.Literal{Value: 14},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID:   &ast.Identifier{Name: "len2"},
						Init: &ast.Identifier{Name: "len"},
					},
				},
			},
			&ast.ExpressionStatement{
				Expression: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "len"},
					Operator: "+",
					Right:    &ast.Literal{Value: 1},
				},
			},
		},
	}

	sanitizer := NewIdentifierSanitizer()
	result, err := sanitizer.Transform(program)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	decl1 := result.Body[0].(*ast.VariableDeclaration)
	id1 := decl1.Declarations[0].ID.(*ast.Identifier)

	decl2 := result.Body[1].(*ast.VariableDeclaration)
	ref2 := decl2.Declarations[0].Init.(*ast.Identifier)

	expr3 := result.Body[2].(*ast.ExpressionStatement)
	ref3 := expr3.Expression.(*ast.BinaryExpression).Left.(*ast.Identifier)

	if id1.Name != "len_" || ref2.Name != "len_" || ref3.Name != "len_" {
		t.Errorf("Inconsistent renaming: got %q, %q, %q", id1.Name, ref2.Name, ref3.Name)
	}

	renamed := sanitizer.GetRenamedIdentifiers()
	if len(renamed) != 1 || renamed["len"] != "len_" {
		t.Errorf("Expected single entry len→len_, got %v", renamed)
	}
}
