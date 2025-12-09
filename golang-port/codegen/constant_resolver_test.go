package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestConstantValue_TypeSafety(t *testing.T) {
	tests := []struct {
		name         string
		constant     ConstantValue
		expectBool   bool
		expectInt    bool
		expectFloat  bool
		expectString bool
		boolValue    bool
		intValue     int
		floatValue   float64
		stringValue  string
	}{
		{
			name:         "bool constant",
			constant:     NewBoolConstant(true),
			expectBool:   true,
			expectInt:    false,
			expectFloat:  false,
			expectString: false,
			boolValue:    true,
		},
		{
			name:         "int constant",
			constant:     NewIntConstant(42),
			expectBool:   false,
			expectInt:    true,
			expectFloat:  false,
			expectString: false,
			intValue:     42,
		},
		{
			name:         "float constant",
			constant:     NewFloatConstant(3.14),
			expectBool:   false,
			expectInt:    false,
			expectFloat:  true,
			expectString: false,
			floatValue:   3.14,
		},
		{
			name:         "string constant",
			constant:     NewStringConstant("test"),
			expectBool:   false,
			expectInt:    false,
			expectFloat:  false,
			expectString: true,
			stringValue:  "test",
		},
		{
			name:       "bool false constant",
			constant:   NewBoolConstant(false),
			expectBool: true,
			boolValue:  false,
		},
		{
			name:       "negative int constant",
			constant:   NewIntConstant(-1),
			expectBool: false,
			expectInt:  true,
			intValue:   -1,
		},
		{
			name:       "zero int constant",
			constant:   NewIntConstant(0),
			expectBool: false,
			expectInt:  true,
			intValue:   0,
		},
		{
			name:        "negative float constant",
			constant:    NewFloatConstant(-2.5),
			expectBool:  false,
			expectFloat: true,
			floatValue:  -2.5,
		},
		{
			name:         "empty string constant",
			constant:     NewStringConstant(""),
			expectBool:   false,
			expectString: true,
			stringValue:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.constant.IsBool(); got != tt.expectBool {
				t.Errorf("IsBool() = %v, want %v", got, tt.expectBool)
			}
			if got := tt.constant.IsInt(); got != tt.expectInt {
				t.Errorf("IsInt() = %v, want %v", got, tt.expectInt)
			}
			if got := tt.constant.IsFloat(); got != tt.expectFloat {
				t.Errorf("IsFloat() = %v, want %v", got, tt.expectFloat)
			}
			if got := tt.constant.IsString(); got != tt.expectString {
				t.Errorf("IsString() = %v, want %v", got, tt.expectString)
			}

			if tt.expectBool {
				val, ok := tt.constant.AsBool()
				if !ok {
					t.Errorf("AsBool() failed for bool constant")
				}
				if val != tt.boolValue {
					t.Errorf("AsBool() = %v, want %v", val, tt.boolValue)
				}
			} else {
				if _, ok := tt.constant.AsBool(); ok {
					t.Errorf("AsBool() succeeded for non-bool constant")
				}
			}

			if tt.expectInt {
				val, ok := tt.constant.AsInt()
				if !ok {
					t.Errorf("AsInt() failed for int constant")
				}
				if val != tt.intValue {
					t.Errorf("AsInt() = %v, want %v", val, tt.intValue)
				}
			} else {
				if _, ok := tt.constant.AsInt(); ok {
					t.Errorf("AsInt() succeeded for non-int constant")
				}
			}

			if tt.expectFloat {
				val, ok := tt.constant.AsFloat()
				if !ok {
					t.Errorf("AsFloat() failed for float constant")
				}
				if val != tt.floatValue {
					t.Errorf("AsFloat() = %v, want %v", val, tt.floatValue)
				}
			} else {
				if _, ok := tt.constant.AsFloat(); ok {
					t.Errorf("AsFloat() succeeded for non-float constant")
				}
			}

			if tt.expectString {
				val, ok := tt.constant.AsString()
				if !ok {
					t.Errorf("AsString() failed for string constant")
				}
				if val != tt.stringValue {
					t.Errorf("AsString() = %v, want %v", val, tt.stringValue)
				}
			} else {
				if _, ok := tt.constant.AsString(); ok {
					t.Errorf("AsString() succeeded for non-string constant")
				}
			}
		})
	}
}

func TestConstantKeyExtractor_ExpressionTypes(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		expected string
		shouldOk bool
	}{
		{
			name: "valid member expression",
			expr: &ast.MemberExpression{
				NodeType: ast.TypeMemberExpression,
				Object:   &ast.Identifier{Name: "barmerge"},
				Property: &ast.Identifier{Name: "lookahead_on"},
				Computed: false,
			},
			expected: "barmerge.lookahead_on",
			shouldOk: true,
		},
		{
			name: "strategy constant",
			expr: &ast.MemberExpression{
				NodeType: ast.TypeMemberExpression,
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "long"},
				Computed: false,
			},
			expected: "strategy.long",
			shouldOk: true,
		},
		{
			name: "color constant",
			expr: &ast.MemberExpression{
				NodeType: ast.TypeMemberExpression,
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Identifier{Name: "red"},
				Computed: false,
			},
			expected: "color.red",
			shouldOk: true,
		},
		{
			name: "literal expression",
			expr: &ast.Literal{
				NodeType: ast.TypeLiteral,
				Value:    true,
			},
			shouldOk: false,
		},
		{
			name: "identifier expression",
			expr: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     "variable",
			},
			shouldOk: false,
		},
		{
			name: "binary expression",
			expr: &ast.BinaryExpression{
				NodeType: ast.TypeBinaryExpression,
				Operator: "+",
				Left:     &ast.Literal{Value: 1},
				Right:    &ast.Literal{Value: 2},
			},
			shouldOk: false,
		},
		{
			name: "computed member expression",
			expr: &ast.MemberExpression{
				NodeType: ast.TypeMemberExpression,
				Object:   &ast.Identifier{Name: "array"},
				Property: &ast.Literal{Value: 0},
				Computed: true,
			},
			shouldOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewConstantKeyExtractor()
			key, ok := extractor.ExtractFromExpression(tt.expr)

			if ok != tt.shouldOk {
				t.Errorf("ExtractFromExpression() ok = %v, want %v", ok, tt.shouldOk)
			}

			if tt.shouldOk && key != tt.expected {
				t.Errorf("ExtractFromExpression() key = %q, want %q", key, tt.expected)
			}
		})
	}
}

func TestPineConstantRegistry_AllNamespaces(t *testing.T) {
	tests := []struct {
		namespace string
		constants map[string]interface{}
	}{
		{
			namespace: "barmerge",
			constants: map[string]interface{}{
				"lookahead_on":  true,
				"lookahead_off": false,
				"gaps_on":       true,
				"gaps_off":      false,
			},
		},
		{
			namespace: "strategy",
			constants: map[string]interface{}{
				"long":  1,
				"short": -1,
				"cash":  "cash",
			},
		},
		{
			namespace: "color",
			constants: map[string]interface{}{
				"red":   "#FF0000",
				"green": "#00FF00",
				"blue":  "#0000FF",
				"black": "#000000",
				"white": "#FFFFFF",
			},
		},
		{
			namespace: "plot",
			constants: map[string]interface{}{
				"style_line":      "line",
				"style_stepline":  "stepline",
				"style_histogram": "histogram",
				"style_cross":     "cross",
				"style_area":      "area",
				"style_columns":   "columns",
				"style_circles":   "circles",
			},
		},
	}

	registry := NewPineConstantRegistry()

	for _, namespace := range tests {
		t.Run(namespace.namespace, func(t *testing.T) {
			for name, expected := range namespace.constants {
				key := namespace.namespace + "." + name

				val, ok := registry.Get(key)
				if !ok {
					t.Errorf("Expected %s to be registered", key)
					continue
				}

				switch expectedVal := expected.(type) {
				case bool:
					if actual, ok := val.AsBool(); !ok || actual != expectedVal {
						t.Errorf("%s: expected %v, got %v", key, expectedVal, actual)
					}
				case int:
					if actual, ok := val.AsInt(); !ok || actual != expectedVal {
						t.Errorf("%s: expected %v, got %v", key, expectedVal, actual)
					}
				case string:
					if actual, ok := val.AsString(); !ok || actual != expectedVal {
						t.Errorf("%s: expected %q, got %q", key, expectedVal, actual)
					}
				}
			}
		})
	}
}

func TestPineConstantRegistry_EdgeCases(t *testing.T) {
	registry := NewPineConstantRegistry()

	t.Run("unknown constant", func(t *testing.T) {
		if _, ok := registry.Get("unknown.constant"); ok {
			t.Error("should fail for unknown constant")
		}
	})

	t.Run("unknown namespace", func(t *testing.T) {
		if _, ok := registry.Get("unknown_namespace.constant"); ok {
			t.Error("should fail for unknown namespace")
		}
	})

	t.Run("empty key", func(t *testing.T) {
		if _, ok := registry.Get(""); ok {
			t.Error("should fail for empty key")
		}
	})

	t.Run("invalid key format", func(t *testing.T) {
		if _, ok := registry.Get("nodot"); ok {
			t.Error("should fail for key without dot separator")
		}
	})

	t.Run("multiple dots in key", func(t *testing.T) {
		if _, ok := registry.Get("strategy.long.extra"); ok {
			t.Error("should fail for key with multiple dots")
		}
	})

	t.Run("case sensitivity", func(t *testing.T) {
		if _, ok := registry.Get("STRATEGY.LONG"); ok {
			t.Error("should fail for uppercase key (case sensitive)")
		}
		if _, ok := registry.Get("Strategy.Long"); ok {
			t.Error("should fail for mixed case key (case sensitive)")
		}
	})

	t.Run("trailing/leading spaces", func(t *testing.T) {
		if _, ok := registry.Get(" strategy.long"); ok {
			t.Error("should fail for key with leading space")
		}
		if _, ok := registry.Get("strategy.long "); ok {
			t.Error("should fail for key with trailing space")
		}
	})

	t.Run("type mismatch access", func(t *testing.T) {
		val, ok := registry.Get("strategy.long")
		if !ok {
			t.Fatal("strategy.long should be registered")
		}

		if _, ok := val.AsBool(); ok {
			t.Error("should fail accessing int constant as bool")
		}
		if _, ok := val.AsFloat(); ok {
			t.Error("should fail accessing int constant as float")
		}
		if _, ok := val.AsString(); ok {
			t.Error("should fail accessing int constant as string")
		}
	})
}

func TestConstantResolver_BoolResolution(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		expected bool
		shouldOk bool
	}{
		{
			name:     "literal true",
			expr:     &ast.Literal{Value: true},
			expected: true,
			shouldOk: true,
		},
		{
			name:     "literal false",
			expr:     &ast.Literal{Value: false},
			expected: false,
			shouldOk: true,
		},
		{
			name: "barmerge.lookahead_on",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "barmerge"},
				Property: &ast.Identifier{Name: "lookahead_on"},
			},
			expected: true,
			shouldOk: true,
		},
		{
			name: "barmerge.lookahead_off",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "barmerge"},
				Property: &ast.Identifier{Name: "lookahead_off"},
			},
			expected: false,
			shouldOk: true,
		},
		{
			name: "barmerge.gaps_on",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "barmerge"},
				Property: &ast.Identifier{Name: "gaps_on"},
			},
			expected: true,
			shouldOk: true,
		},
		{
			name: "barmerge.gaps_off",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "barmerge"},
				Property: &ast.Identifier{Name: "gaps_off"},
			},
			expected: false,
			shouldOk: true,
		},
		{
			name: "unknown constant",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "unknown"},
				Property: &ast.Identifier{Name: "constant"},
			},
			shouldOk: false,
		},
		{
			name: "int constant requested as bool",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "long"},
			},
			shouldOk: false,
		},
		{
			name: "string constant requested as bool",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Identifier{Name: "red"},
			},
			shouldOk: false,
		},
		{
			name:     "non-bool literal",
			expr:     &ast.Literal{Value: 42},
			shouldOk: false,
		},
		{
			name:     "string literal",
			expr:     &ast.Literal{Value: "true"},
			shouldOk: false,
		},
		{
			name:     "identifier",
			expr:     &ast.Identifier{Name: "variable"},
			shouldOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewConstantResolver()
			val, ok := resolver.ResolveToBool(tt.expr)

			if ok != tt.shouldOk {
				t.Errorf("ResolveToBool() ok = %v, want %v", ok, tt.shouldOk)
			}

			if tt.shouldOk && val != tt.expected {
				t.Errorf("ResolveToBool() val = %v, want %v", val, tt.expected)
			}
		})
	}
}

func TestConstantResolver_IntResolution(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		expected int
		shouldOk bool
	}{
		{
			name:     "literal int",
			expr:     &ast.Literal{Value: 42},
			expected: 42,
			shouldOk: true,
		},
		{
			name: "strategy.long",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "long"},
			},
			expected: 1,
			shouldOk: true,
		},
		{
			name: "strategy.short",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "short"},
			},
			expected: -1,
			shouldOk: true,
		},
		{
			name: "bool constant requested as int",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "barmerge"},
				Property: &ast.Identifier{Name: "lookahead_on"},
			},
			shouldOk: false,
		},
		{
			name:     "float literal",
			expr:     &ast.Literal{Value: 3.14},
			shouldOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewConstantResolver()
			val, ok := resolver.ResolveToInt(tt.expr)

			if ok != tt.shouldOk {
				t.Errorf("ResolveToInt() ok = %v, want %v", ok, tt.shouldOk)
			}

			if tt.shouldOk && val != tt.expected {
				t.Errorf("ResolveToInt() val = %v, want %v", val, tt.expected)
			}
		})
	}
}

func TestConstantResolver_FloatResolution(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		expected float64
		shouldOk bool
	}{
		{
			name:     "literal float",
			expr:     &ast.Literal{Value: 3.14},
			expected: 3.14,
			shouldOk: true,
		},
		{
			name:     "literal int converted to float",
			expr:     &ast.Literal{Value: 42},
			expected: 42.0,
			shouldOk: true,
		},
		{
			name: "bool constant requested as float",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "barmerge"},
				Property: &ast.Identifier{Name: "lookahead_on"},
			},
			shouldOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewConstantResolver()
			val, ok := resolver.ResolveToFloat(tt.expr)

			if ok != tt.shouldOk {
				t.Errorf("ResolveToFloat() ok = %v, want %v", ok, tt.shouldOk)
			}

			if tt.shouldOk && val != tt.expected {
				t.Errorf("ResolveToFloat() val = %v, want %v", val, tt.expected)
			}
		})
	}
}

func TestConstantResolver_StringResolution(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		expected string
		shouldOk bool
	}{
		{
			name:     "literal string",
			expr:     &ast.Literal{Value: "test"},
			expected: "test",
			shouldOk: true,
		},
		{
			name: "color.red",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Identifier{Name: "red"},
			},
			expected: "#FF0000",
			shouldOk: true,
		},
		{
			name: "strategy.cash",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "cash"},
			},
			expected: "cash",
			shouldOk: true,
		},
		{
			name: "plot.style_line",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "plot"},
				Property: &ast.Identifier{Name: "style_line"},
			},
			expected: "line",
			shouldOk: true,
		},
		{
			name:     "empty string",
			expr:     &ast.Literal{Value: ""},
			expected: "",
			shouldOk: true,
		},
		{
			name: "int constant requested as string",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "long"},
			},
			shouldOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewConstantResolver()
			val, ok := resolver.ResolveToString(tt.expr)

			if ok != tt.shouldOk {
				t.Errorf("ResolveToString() ok = %v, want %v", ok, tt.shouldOk)
			}

			if tt.shouldOk && val != tt.expected {
				t.Errorf("ResolveToString() val = %q, want %q", val, tt.expected)
			}
		})
	}
}

func TestConstantResolver_EdgeCases(t *testing.T) {
	t.Run("nil expression", func(t *testing.T) {
		resolver := NewConstantResolver()

		if _, ok := resolver.ResolveToBool(nil); ok {
			t.Error("ResolveToBool should fail for nil expression")
		}
		if _, ok := resolver.ResolveToInt(nil); ok {
			t.Error("ResolveToInt should fail for nil expression")
		}
		if _, ok := resolver.ResolveToFloat(nil); ok {
			t.Error("ResolveToFloat should fail for nil expression")
		}
		if _, ok := resolver.ResolveToString(nil); ok {
			t.Error("ResolveToString should fail for nil expression")
		}
	})

	t.Run("member expression with nil object", func(t *testing.T) {
		resolver := NewConstantResolver()
		expr := &ast.MemberExpression{
			Object:   nil,
			Property: &ast.Identifier{Name: "constant"},
		}

		if _, ok := resolver.ResolveToBool(expr); ok {
			t.Error("should fail for member expression with nil object")
		}
	})

	t.Run("member expression with nil property", func(t *testing.T) {
		resolver := NewConstantResolver()
		expr := &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "namespace"},
			Property: nil,
		}

		if _, ok := resolver.ResolveToBool(expr); ok {
			t.Error("should fail for member expression with nil property")
		}
	})

	t.Run("member expression with non-identifier object", func(t *testing.T) {
		resolver := NewConstantResolver()
		expr := &ast.MemberExpression{
			Object:   &ast.Literal{Value: "not_identifier"},
			Property: &ast.Identifier{Name: "constant"},
		}

		if _, ok := resolver.ResolveToBool(expr); ok {
			t.Error("should fail for member expression with non-identifier object")
		}
	})

	t.Run("member expression with non-identifier property", func(t *testing.T) {
		resolver := NewConstantResolver()
		expr := &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "namespace"},
			Property: &ast.Literal{Value: 0},
		}

		if _, ok := resolver.ResolveToBool(expr); ok {
			t.Error("should fail for member expression with non-identifier property")
		}
	})
}
