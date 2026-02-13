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

func TestColorConstantResolver_ResolveExpression(t *testing.T) {
	resolver := NewColorConstantResolver()

	t.Run("named color dispatch", func(t *testing.T) {
		sample := map[string]string{
			"red":   "#FF5252",
			"blue":  "#2962FF",
			"green": "#4CAF50",
			"white": "#FFFFFF",
			"black": "#363A45",
		}
		for name, expectedHex := range sample {
			t.Run(name, func(t *testing.T) {
				expr := &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: name},
				}
				hex, ok := resolver.ResolveExpression(expr)
				if !ok || hex != expectedHex {
					t.Errorf("expected %q, got %q (ok=%v)", expectedHex, hex, ok)
				}
			})
		}
	})

	t.Run("color.rgb", func(t *testing.T) {
		tests := []struct {
			name     string
			args     []ast.Expression
			expected string
		}{
			{
				name: "pure red",
				args: []ast.Expression{
					&ast.Literal{Value: float64(255)},
					&ast.Literal{Value: float64(0)},
					&ast.Literal{Value: float64(0)},
				},
				expected: "#FF0000",
			},
			{
				name: "white",
				args: []ast.Expression{
					&ast.Literal{Value: float64(255)},
					&ast.Literal{Value: float64(255)},
					&ast.Literal{Value: float64(255)},
				},
				expected: "#FFFFFF",
			},
			{
				name: "black",
				args: []ast.Expression{
					&ast.Literal{Value: float64(0)},
					&ast.Literal{Value: float64(0)},
					&ast.Literal{Value: float64(0)},
				},
				expected: "#000000",
			},
			{
				name: "integer literal args",
				args: []ast.Expression{
					&ast.Literal{Value: 128},
					&ast.Literal{Value: 64},
					&ast.Literal{Value: 32},
				},
				expected: "#804020",
			},
			{
				name: "transparency arg ignored",
				args: []ast.Expression{
					&ast.Literal{Value: float64(128)},
					&ast.Literal{Value: float64(64)},
					&ast.Literal{Value: float64(32)},
					&ast.Literal{Value: float64(50)},
				},
				expected: "#804020",
			},
			{
				name: "clamps above 255",
				args: []ast.Expression{
					&ast.Literal{Value: float64(300)},
					&ast.Literal{Value: float64(256)},
					&ast.Literal{Value: float64(128)},
				},
				expected: "#FFFF80",
			},
			{
				name: "clamps below 0",
				args: []ast.Expression{
					&ast.Literal{Value: float64(-10)},
					&ast.Literal{Value: float64(-1)},
					&ast.Literal{Value: float64(0)},
				},
				expected: "#000000",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				call := &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "color"},
						Property: &ast.Identifier{Name: "rgb"},
					},
					Arguments: tt.args,
				}
				hex, ok := resolver.ResolveExpression(call)
				if !ok || hex != tt.expected {
					t.Errorf("expected %q, got %q (ok=%v)", tt.expected, hex, ok)
				}
			})
		}
	})

	t.Run("color.new", func(t *testing.T) {
		t.Run("named base with transparency", func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "new"},
				},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "color"},
						Property: &ast.Identifier{Name: "red"},
					},
					&ast.Literal{Value: float64(50)},
				},
			}
			hex, ok := resolver.ResolveExpression(call)
			if !ok || hex != "#FF5252" {
				t.Errorf("expected #FF5252, got %q (ok=%v)", hex, ok)
			}
		})

		t.Run("rgb base nested", func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "new"},
				},
				Arguments: []ast.Expression{
					&ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "color"},
							Property: &ast.Identifier{Name: "rgb"},
						},
						Arguments: []ast.Expression{
							&ast.Literal{Value: float64(100)},
							&ast.Literal{Value: float64(200)},
							&ast.Literal{Value: float64(50)},
						},
					},
					&ast.Literal{Value: float64(80)},
				},
			}
			hex, ok := resolver.ResolveExpression(call)
			if !ok || hex != "#64C832" {
				t.Errorf("expected #64C832, got %q (ok=%v)", hex, ok)
			}
		})

		t.Run("unresolvable base rejected", func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "new"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "myColor"},
					&ast.Literal{Value: float64(50)},
				},
			}
			if _, ok := resolver.ResolveExpression(call); ok {
				t.Error("color.new with unresolvable base should not resolve")
			}
		})
	})

	t.Run("hex literal", func(t *testing.T) {
		tests := []struct {
			name     string
			value    interface{}
			expected string
			ok       bool
		}{
			{"6-digit hex", "#FF0000", "#FF0000", true},
			{"8-digit hex with alpha", "#FF000080", "#FF000080", true},
			{"lowercase hex", "#aabbcc", "#aabbcc", true},
			{"mixed case hex", "#AaBbCc", "#AaBbCc", true},
			{"8-digit lowercase", "#00ff0080", "#00ff0080", true},
			{"too short 1 digit", "#F", "", false},
			{"too short 3 digits", "#FFF", "", false},
			{"too short 5 digits", "#12345", "", false},
			{"too long 7 digits", "#1234567", "", false},
			{"too long 10 digits", "#1234567890", "", false},
			{"non-hex characters 6 chars", "#GGHHII", "", false},
			{"non-hex characters 8 chars", "#NotAHex!", "", false},
			{"hash only", "#", "", false},
			{"non-hex string rejected", "red", "", false},
			{"empty string rejected", "", "", false},
			{"numeric rejected", float64(42), "", false},
			{"bool rejected", true, "", false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				hex, ok := resolver.ResolveExpression(&ast.Literal{Value: tt.value})
				if ok != tt.ok {
					t.Errorf("expected ok=%v, got ok=%v (hex=%q)", tt.ok, ok, hex)
				}
				if ok && hex != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, hex)
				}
			})
		}
	})

	t.Run("rejection", func(t *testing.T) {
		cases := []struct {
			name string
			expr ast.Expression
		}{
			{"nil expression", nil},
			{"bare identifier", &ast.Identifier{Name: "red"}},
			{"non-color namespace call", &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "rgb"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: float64(255)},
					&ast.Literal{Value: float64(0)},
					&ast.Literal{Value: float64(0)},
				},
			}},
			{"unknown color function", &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "from_gradient"},
				},
				Arguments: []ast.Expression{&ast.Literal{Value: float64(50)}},
			}},
			{"non-MemberExpression callee", &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "rgb"},
				Arguments: []ast.Expression{&ast.Literal{Value: float64(255)}},
			}},
			{"color.rgb insufficient args", &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "rgb"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: float64(255)},
					&ast.Literal{Value: float64(0)},
				},
			}},
			{"color.rgb zero args", &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "rgb"},
				},
				Arguments: []ast.Expression{},
			}},
			{"color.rgb non-numeric arg", &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "rgb"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: float64(0)},
					&ast.Literal{Value: float64(0)},
				},
			}},
			{"color.new empty args", &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "new"},
				},
				Arguments: []ast.Expression{},
			}},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				if hex, ok := resolver.ResolveExpression(tt.expr); ok {
					t.Errorf("should not resolve, got %q", hex)
				}
			})
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
