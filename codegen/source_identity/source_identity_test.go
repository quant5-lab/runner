package source_identity

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestSourceIdentifier_ValueObject tests immutability and value semantics */
func TestSourceIdentifier_ValueObject(t *testing.T) {
	id1 := SourceIdentifier{hash: "abc123"}
	id2 := SourceIdentifier{hash: "abc123"}
	id3 := SourceIdentifier{hash: "def456"}

	/* Test value equality */
	if id1.Hash() != id2.Hash() {
		t.Errorf("identical values should have same string representation: %q != %q", id1.Hash(), id2.Hash())
	}

	/* Test value inequality */
	if id1.Hash() == id3.Hash() {
		t.Errorf("different values should have different string representation: %q == %q", id1.Hash(), id3.Hash())
	}

	/* Test immutability - String() should always return same value */
	firstCall := id1.Hash()
	secondCall := id1.Hash()
	if firstCall != secondCall {
		t.Errorf("String() should be stable: %q != %q", firstCall, secondCall)
	}
}

/* TestIdentifierFactory_DeterministicHashing tests hash stability across calls */
func TestIdentifierFactory_DeterministicHashing(t *testing.T) {
	factory := NewIdentifierFactory()

	tests := []struct {
		name string
		expr ast.Expression
	}{
		{
			name: "simple identifier",
			expr: &ast.Identifier{Name: "close"},
		},
		{
			name: "literal number",
			expr: &ast.Literal{Value: 14.0},
		},
		{
			name: "binary expression",
			expr: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "close"},
				Operator: "+",
				Right:    &ast.Literal{Value: 1.0},
			},
		},
		{
			name: "member expression",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Literal{Value: 1.0},
				Computed: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			/* Generate hash multiple times */
			id1 := factory.CreateFromExpression(tt.expr)
			id2 := factory.CreateFromExpression(tt.expr)
			id3 := factory.CreateFromExpression(tt.expr)

			/* All should be identical (deterministic) */
			if id1.Hash() != id2.Hash() {
				t.Errorf("hash unstable between calls: %q != %q", id1.Hash(), id2.Hash())
			}
			if id2.Hash() != id3.Hash() {
				t.Errorf("hash unstable between calls: %q != %q", id2.Hash(), id3.Hash())
			}
		})
	}
}

/* TestIdentifierFactory_UniquenessAcrossDifferentExpressions tests collision resistance */
func TestIdentifierFactory_UniquenessAcrossDifferentExpressions(t *testing.T) {
	factory := NewIdentifierFactory()

	tests := []struct {
		name  string
		expr1 ast.Expression
		expr2 ast.Expression
	}{
		{
			name:  "different identifiers",
			expr1: &ast.Identifier{Name: "close"},
			expr2: &ast.Identifier{Name: "open"},
		},
		{
			name:  "different literals",
			expr1: &ast.Literal{Value: 14.0},
			expr2: &ast.Literal{Value: 18.0},
		},
		{
			name: "different operators",
			expr1: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "close"},
				Operator: "+",
				Right:    &ast.Literal{Value: 1.0},
			},
			expr2: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "close"},
				Operator: "-",
				Right:    &ast.Literal{Value: 1.0},
			},
		},
		{
			name: "different operands",
			expr1: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "close"},
				Operator: "+",
				Right:    &ast.Literal{Value: 1.0},
			},
			expr2: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "open"},
				Operator: "+",
				Right:    &ast.Literal{Value: 1.0},
			},
		},
		{
			name: "same structure different order - commutative operations",
			expr1: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "a"},
				Operator: "+",
				Right:    &ast.Identifier{Name: "b"},
			},
			expr2: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "b"},
				Operator: "+",
				Right:    &ast.Identifier{Name: "a"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id1 := factory.CreateFromExpression(tt.expr1)
			id2 := factory.CreateFromExpression(tt.expr2)

			/* Different expressions should produce different hashes */
			if id1.Hash() == id2.Hash() {
				t.Errorf("hash collision detected: both expressions produce %q", id1.Hash())
			}
		})
	}
}

/* TestIdentifierFactory_StructuralEquivalence tests expressions with same AST structure produce same hash */
func TestIdentifierFactory_StructuralEquivalence(t *testing.T) {
	factory := NewIdentifierFactory()

	/* Create two structurally identical but separate AST nodes */
	expr1 := &ast.BinaryExpression{
		Left:     &ast.Identifier{Name: "close"},
		Operator: "-",
		Right: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "close"},
			Property: &ast.Literal{Value: 1.0},
			Computed: true,
		},
	}

	expr2 := &ast.BinaryExpression{
		Left:     &ast.Identifier{Name: "close"},
		Operator: "-",
		Right: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "close"},
			Property: &ast.Literal{Value: 1.0},
			Computed: true,
		},
	}

	id1 := factory.CreateFromExpression(expr1)
	id2 := factory.CreateFromExpression(expr2)

	/* Structurally identical expressions should produce same hash */
	if id1.Hash() != id2.Hash() {
		t.Errorf("structurally identical expressions should hash the same: %q != %q", id1.Hash(), id2.Hash())
	}
}

/* TestIdentifierFactory_NestedExpressionHandling tests complex nested structures */
func TestIdentifierFactory_NestedExpressionHandling(t *testing.T) {
	factory := NewIdentifierFactory()

	tests := []struct {
		name string
		expr ast.Expression
	}{
		{
			name: "deeply nested binary expressions",
			expr: &ast.BinaryExpression{
				Left: &ast.BinaryExpression{
					Left: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "a"},
						Operator: "+",
						Right:    &ast.Identifier{Name: "b"},
					},
					Operator: "*",
					Right:    &ast.Identifier{Name: "c"},
				},
				Operator: "-",
				Right:    &ast.Identifier{Name: "d"},
			},
		},
		{
			name: "nested member expressions",
			expr: &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "obj"},
					Property: &ast.Identifier{Name: "prop1"},
					Computed: false,
				},
				Property: &ast.Identifier{Name: "prop2"},
				Computed: false,
			},
		},
		{
			name: "call expression with multiple arguments",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "rma"},
					Computed: false,
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "source"},
					&ast.Literal{Value: 14.0},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			/* Should not panic and should produce valid hash */
			id := factory.CreateFromExpression(tt.expr)
			hash := id.Hash()

			if hash == "" {
				t.Error("hash should not be empty for complex expression")
			}

			if len(hash) != 8 {
				t.Errorf("hash length should be 8 characters, got %d: %q", len(hash), hash)
			}

			/* Should be deterministic */
			id2 := factory.CreateFromExpression(tt.expr)
			if id.Hash() != id2.Hash() {
				t.Errorf("complex expression hash unstable: %q != %q", id.Hash(), id2.Hash())
			}
		})
	}
}

/* TestIdentifierFactory_NilHandling tests nil expression handling */
func TestIdentifierFactory_NilHandling(t *testing.T) {
	factory := NewIdentifierFactory()

	/* Nil expression produces empty identifier */
	id := factory.CreateFromExpression(nil)
	hash := id.Hash()

	if hash != "" {
		t.Errorf("nil expression should produce empty hash, got %q", hash)
	}

	/* Empty identifier should be identifiable */
	if !id.IsEmpty() {
		t.Error("nil expression should produce empty identifier")
	}

	/* Should be deterministic even for nil */
	id2 := factory.CreateFromExpression(nil)
	if id.Hash() != id2.Hash() {
		t.Errorf("nil handling should be deterministic: %q != %q", id.Hash(), id2.Hash())
	}
}

/* TestIdentifierFactory_HashLength tests hash output format */
func TestIdentifierFactory_HashLength(t *testing.T) {
	factory := NewIdentifierFactory()

	exprs := []ast.Expression{
		&ast.Identifier{Name: "x"},
		&ast.Literal{Value: 1.0},
		&ast.BinaryExpression{
			Left:     &ast.Identifier{Name: "a"},
			Operator: "+",
			Right:    &ast.Identifier{Name: "b"},
		},
	}

	for _, expr := range exprs {
		id := factory.CreateFromExpression(expr)
		hash := id.Hash()

		/* Hash should be exactly 8 characters (first 8 of SHA256 hex) */
		if len(hash) != 8 {
			t.Errorf("hash length should be 8, got %d: %q", len(hash), hash)
		}

		/* Should only contain hex characters */
		for _, ch := range hash {
			if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
				t.Errorf("hash contains non-hex character %q in %q", ch, hash)
			}
		}
	}
}
