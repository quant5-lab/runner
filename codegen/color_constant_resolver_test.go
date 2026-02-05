package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestColorConstantResolver_BareIdentifierResolution(t *testing.T) {
	resolver := NewColorConstantResolver()

	t.Run("valid color identifiers", func(t *testing.T) {
		colors := []string{"blue", "red", "green"}
		for _, color := range colors {
			hex, found := resolver.ResolveIdentifierToHex(color)
			if !found {
				t.Errorf("ResolveIdentifierToHex(%q) should find color", color)
			}
			if len(hex) != 7 || hex[0] != '#' {
				t.Errorf("ResolveIdentifierToHex(%q) returned invalid hex: %q", color, hex)
			}
		}
	})

	t.Run("case sensitivity", func(t *testing.T) {
		cases := []string{"RED", "Blue", "GREEN"}
		for _, name := range cases {
			if _, found := resolver.ResolveIdentifierToHex(name); found {
				t.Errorf("ResolveIdentifierToHex(%q) should be case-sensitive", name)
			}
		}
	})

	t.Run("non-color identifiers", func(t *testing.T) {
		nonColors := []string{"close", "open", "sma", "myvar", "blu", "blue_var", ""}
		for _, name := range nonColors {
			if hex, found := resolver.ResolveIdentifierToHex(name); found {
				t.Errorf("ResolveIdentifierToHex(%q) should not find color, got %q", name, hex)
			}
		}
	})
}

func TestColorConstantResolver_MemberExpressionResolution(t *testing.T) {
	resolver := NewColorConstantResolver()

	t.Run("valid color namespace", func(t *testing.T) {
		expr := &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "color"},
			Property: &ast.Identifier{Name: "blue"},
		}
		hex, found := resolver.ResolveMemberExpressionToHex(expr)
		if !found {
			t.Error("ResolveMemberExpressionToHex(color.blue) should find color")
		}
		if len(hex) != 7 || hex[0] != '#' {
			t.Errorf("ResolveMemberExpressionToHex returned invalid hex: %q", hex)
		}
	})

	t.Run("wrong namespace", func(t *testing.T) {
		wrongNamespaces := []string{"colour", "colors", "strategy", "ta", ""}
		for _, ns := range wrongNamespaces {
			expr := &ast.MemberExpression{
				Object:   &ast.Identifier{Name: ns},
				Property: &ast.Identifier{Name: "blue"},
			}
			if hex, found := resolver.ResolveMemberExpressionToHex(expr); found {
				t.Errorf("ResolveMemberExpressionToHex(%s.blue) should not find color, got %q", ns, hex)
			}
		}
	})

	t.Run("invalid color property", func(t *testing.T) {
		expr := &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "color"},
			Property: &ast.Identifier{Name: "invalid"},
		}
		if hex, found := resolver.ResolveMemberExpressionToHex(expr); found {
			t.Errorf("ResolveMemberExpressionToHex(color.invalid) should not find color, got %q", hex)
		}
	})

	t.Run("nil safety", func(t *testing.T) {
		nilCases := []*ast.MemberExpression{
			nil,
			{Object: nil, Property: &ast.Identifier{Name: "blue"}},
			{Object: &ast.Identifier{Name: "color"}, Property: nil},
		}
		for i, expr := range nilCases {
			if hex, found := resolver.ResolveMemberExpressionToHex(expr); found {
				t.Errorf("Case %d: ResolveMemberExpressionToHex(nil) should not panic or find color, got %q", i, hex)
			}
		}
	})

	t.Run("malformed AST nodes", func(t *testing.T) {
		malformedCases := []*ast.MemberExpression{
			{
				Object:   &ast.Literal{Value: "color"},
				Property: &ast.Identifier{Name: "blue"},
			},
			{
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Literal{Value: "blue"},
			},
			{
				Object: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "red"},
				},
				Property: &ast.Identifier{Name: "blue"},
			},
		}
		for i, expr := range malformedCases {
			if hex, found := resolver.ResolveMemberExpressionToHex(expr); found {
				t.Errorf("Case %d: malformed AST should not resolve, got %q", i, hex)
			}
		}
	})
}

func TestColorConstantResolver_IsColorIdentifier(t *testing.T) {
	resolver := NewColorConstantResolver()

	t.Run("positive classification", func(t *testing.T) {
		colors := []string{"red", "green", "blue", "silver", "aqua"}
		for _, name := range colors {
			if !resolver.IsColorIdentifier(name) {
				t.Errorf("IsColorIdentifier(%q) should return true", name)
			}
		}
	})

	t.Run("negative classification", func(t *testing.T) {
		nonColors := []string{"close", "open", "RED", "myvar", ""}
		for _, name := range nonColors {
			if resolver.IsColorIdentifier(name) {
				t.Errorf("IsColorIdentifier(%q) should return false", name)
			}
		}
	})
}

func TestColorConstantResolver_ConsistencyWithRegistry(t *testing.T) {
	resolver := NewColorConstantResolver()
	registry := NewPineConstantRegistry()

	allColors := []string{
		"aqua", "black", "blue", "fuchsia", "gray", "green", "lime",
		"maroon", "navy", "olive", "orange", "purple", "red",
		"silver", "teal", "white", "yellow",
	}

	for _, colorName := range allColors {
		t.Run(colorName, func(t *testing.T) {
			hexFromResolver, foundResolver := resolver.ResolveIdentifierToHex(colorName)
			hexFromRegistry, foundRegistry := registry.GetColorHex(colorName)

			if foundResolver != foundRegistry {
				t.Errorf("Inconsistency: resolver found=%v, registry found=%v", foundResolver, foundRegistry)
			}

			if foundResolver && foundRegistry && hexFromResolver != hexFromRegistry {
				t.Errorf("Inconsistency: resolver hex=%q, registry hex=%q", hexFromResolver, hexFromRegistry)
			}

			memberExpr := &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Identifier{Name: colorName},
			}
			hexFromMember, foundMember := resolver.ResolveMemberExpressionToHex(memberExpr)

			if foundResolver && foundMember && hexFromResolver != hexFromMember {
				t.Errorf("Inconsistency: bare identifier hex=%q, member expression hex=%q", hexFromResolver, hexFromMember)
			}
		})
	}
}
