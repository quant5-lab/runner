package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestMemberExpressionDirectionExtractor_PineV5Syntax verifies v5 strategy direction constants */
func TestMemberExpressionDirectionExtractor_PineV5Syntax(t *testing.T) {
	extractor := &MemberExpressionDirectionExtractor{}

	tests := []struct {
		name         string
		objectName   string
		propertyName string
		expected     string
		found        bool
	}{
		{
			name:         "strategy.long constant",
			objectName:   "strategy",
			propertyName: "long",
			expected:     "strategy.Long",
			found:        true,
		},
		{
			name:         "strategy.short constant",
			objectName:   "strategy",
			propertyName: "short",
			expected:     "strategy.Short",
			found:        true,
		},
		{
			name:         "unknown property rejected",
			objectName:   "strategy",
			propertyName: "unknown",
			expected:     "",
			found:        false,
		},
		{
			name:         "any object with long property accepted",
			objectName:   "position",
			propertyName: "long",
			expected:     "strategy.Long",
			found:        true,
		},
		{
			name:         "case sensitivity enforced",
			objectName:   "strategy",
			propertyName: "Long",
			expected:     "",
			found:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.MemberExpression{
				Object:   &ast.Identifier{Name: tt.objectName},
				Property: &ast.Identifier{Name: tt.propertyName},
			}

			result, found := extractor.Extract(expr)
			if found != tt.found {
				t.Errorf("expected found=%v, got %v", tt.found, found)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

/* TestMemberExpressionDirectionExtractor_TypeSafety verifies type checking robustness */
func TestMemberExpressionDirectionExtractor_TypeSafety(t *testing.T) {
	extractor := &MemberExpressionDirectionExtractor{}

	tests := []struct {
		name     string
		expr     ast.Expression
		expected string
		found    bool
	}{
		{
			name:     "non-member expression rejected",
			expr:     &ast.Literal{Value: true},
			expected: "",
			found:    false,
		},
		{
			name:     "identifier rejected",
			expr:     &ast.Identifier{Name: "long"},
			expected: "",
			found:    false,
		},
		{
			name: "nil property rejected",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: nil,
			},
			expected: "",
			found:    false,
		},
		{
			name: "non-identifier property rejected",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Literal{Value: "long"},
			},
			expected: "",
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, found := extractor.Extract(tt.expr)
			if found != tt.found {
				t.Errorf("expected found=%v, got %v", tt.found, found)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

/* TestBooleanLiteralDirectionExtractor_PineV4BooleanSemantics verifies v4 boolean mapping */
func TestBooleanLiteralDirectionExtractor_PineV4BooleanSemantics(t *testing.T) {
	extractor := &BooleanLiteralDirectionExtractor{}

	tests := []struct {
		name     string
		value    interface{}
		expected string
		found    bool
	}{
		{
			name:     "true maps to Long",
			value:    true,
			expected: "strategy.Long",
			found:    true,
		},
		{
			name:     "false maps to Short",
			value:    false,
			expected: "strategy.Short",
			found:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.Literal{Value: tt.value}

			result, found := extractor.Extract(expr)
			if found != tt.found {
				t.Errorf("expected found=%v, got %v", tt.found, found)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

/* TestBooleanLiteralDirectionExtractor_TypeDiscrimination verifies strict type checking */
func TestBooleanLiteralDirectionExtractor_TypeDiscrimination(t *testing.T) {
	extractor := &BooleanLiteralDirectionExtractor{}

	tests := []struct {
		name  string
		expr  ast.Expression
		found bool
	}{
		{
			name:  "string literal rejected",
			expr:  &ast.Literal{Value: "true"},
			found: false,
		},
		{
			name:  "integer literal rejected",
			expr:  &ast.Literal{Value: 1},
			found: false,
		},
		{
			name:  "float literal rejected",
			expr:  &ast.Literal{Value: 1.0},
			found: false,
		},
		{
			name:  "nil value rejected",
			expr:  &ast.Literal{Value: nil},
			found: false,
		},
		{
			name:  "identifier rejected",
			expr:  &ast.Identifier{Name: "true"},
			found: false,
		},
		{
			name:  "member expression rejected",
			expr:  &ast.MemberExpression{Object: &ast.Identifier{Name: "x"}, Property: &ast.Identifier{Name: "y"}},
			found: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, found := extractor.Extract(tt.expr)
			if found != tt.found {
				t.Errorf("expected found=%v, got %v", tt.found, found)
			}
		})
	}
}

/* TestIdentifierDirectionExtractor_PineV4StringLiterals verifies identifier-as-boolean */
func TestIdentifierDirectionExtractor_PineV4StringLiterals(t *testing.T) {
	extractor := &IdentifierDirectionExtractor{}

	tests := []struct {
		name      string
		identName string
		expected  string
		found     bool
	}{
		{
			name:      "true identifier to Long",
			identName: "true",
			expected:  "strategy.Long",
			found:     true,
		},
		{
			name:      "false identifier to Short",
			identName: "false",
			expected:  "strategy.Short",
			found:     true,
		},
		{
			name:      "case sensitivity enforced",
			identName: "True",
			expected:  "",
			found:     false,
		},
		{
			name:      "case sensitivity enforced",
			identName: "FALSE",
			expected:  "",
			found:     false,
		},
		{
			name:      "arbitrary identifier rejected",
			identName: "longDirection",
			expected:  "",
			found:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.Identifier{Name: tt.identName}

			result, found := extractor.Extract(expr)
			if found != tt.found {
				t.Errorf("expected found=%v, got %v", tt.found, found)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

/* TestIdentifierDirectionExtractor_BoundaryConditions verifies edge cases */
func TestIdentifierDirectionExtractor_BoundaryConditions(t *testing.T) {
	extractor := &IdentifierDirectionExtractor{}

	tests := []struct {
		name  string
		expr  ast.Expression
		found bool
	}{
		{
			name:  "non-identifier expression rejected",
			expr:  &ast.Literal{Value: "true"},
			found: false,
		},
		{
			name:  "empty identifier rejected",
			expr:  &ast.Identifier{Name: ""},
			found: false,
		},
		{
			name:  "whitespace identifier rejected",
			expr:  &ast.Identifier{Name: " true "},
			found: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, found := extractor.Extract(tt.expr)
			if found != tt.found {
				t.Errorf("expected found=%v, got %v", tt.found, found)
			}
		})
	}
}

/* TestChainDirectionExtractor_MultiVersionCompatibility verifies version detection */
func TestChainDirectionExtractor_MultiVersionCompatibility(t *testing.T) {
	extractor := NewDefaultDirectionExtractor()

	tests := []struct {
		name        string
		expr        ast.Expression
		expected    string
		pineVersion string
	}{
		{
			name: "Pine v5 strategy.long",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "long"},
			},
			expected:    "strategy.Long",
			pineVersion: "v5",
		},
		{
			name: "Pine v5 strategy.short",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "short"},
			},
			expected:    "strategy.Short",
			pineVersion: "v5",
		},
		{
			name:        "Pine v4 true boolean literal",
			expr:        &ast.Literal{Value: true},
			expected:    "strategy.Long",
			pineVersion: "v4",
		},
		{
			name:        "Pine v4 false boolean literal",
			expr:        &ast.Literal{Value: false},
			expected:    "strategy.Short",
			pineVersion: "v4",
		},
		{
			name:        "Pine v4 true identifier",
			expr:        &ast.Identifier{Name: "true"},
			expected:    "strategy.Long",
			pineVersion: "v4",
		},
		{
			name:        "Pine v4 false identifier",
			expr:        &ast.Identifier{Name: "false"},
			expected:    "strategy.Short",
			pineVersion: "v4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.Extract(tt.expr)
			if result != tt.expected {
				t.Errorf("%s syntax: expected %q, got %q", tt.pineVersion, tt.expected, result)
			}
		})
	}
}

/* TestChainDirectionExtractor_FallbackBehavior verifies default handling */
func TestChainDirectionExtractor_FallbackBehavior(t *testing.T) {
	extractor := NewDefaultDirectionExtractor()

	tests := []struct {
		name     string
		expr     ast.Expression
		expected string
	}{
		{
			name:     "unknown expression defaults to Long",
			expr:     &ast.Literal{Value: "unknown"},
			expected: "strategy.Long",
		},
		{
			name:     "numeric literal defaults to Long",
			expr:     &ast.Literal{Value: 42},
			expected: "strategy.Long",
		},
		{
			name: "call expression defaults to Long",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "getDirection"},
			},
			expected: "strategy.Long",
		},
		{
			name: "unary expression defaults to Long",
			expr: &ast.UnaryExpression{
				Operator: "!",
				Argument: &ast.Identifier{Name: "x"},
			},
			expected: "strategy.Long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.Extract(tt.expr)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

/* TestChainDirectionExtractor_ChainOrdering verifies precedence rules */
func TestChainDirectionExtractor_ChainOrdering(t *testing.T) {
	/* Verify that v5 member expressions take precedence over v4 identifiers */
	extractor := NewDefaultDirectionExtractor()

	/* This tests that if AST contains strategy.long, it's matched before
	   any identifier named "long" could be processed */
	expr := &ast.MemberExpression{
		Object:   &ast.Identifier{Name: "strategy"},
		Property: &ast.Identifier{Name: "long"},
	}

	result := extractor.Extract(expr)
	if result != "strategy.Long" {
		t.Errorf("Member expression should match before identifier fallback, got %q", result)
	}
}

/* TestChainDirectionExtractor_ExtensibilityContract verifies new extractor addition */
func TestChainDirectionExtractor_ExtensibilityContract(t *testing.T) {
	/* Verify chain accepts new extractors without modifying existing code */
	chain := NewChainDirectionExtractor(
		&MemberExpressionDirectionExtractor{},
		&BooleanLiteralDirectionExtractor{},
		&IdentifierDirectionExtractor{},
		/* Future extractors can be added here without modifying existing extractors */
	)

	/* Verify existing behavior unchanged when adding extractors */
	result := chain.Extract(&ast.Literal{Value: true})
	if result != "strategy.Long" {
		t.Errorf("Adding extractors should not break existing chain, got %q", result)
	}

	/* Verify chain length is correct */
	if len(chain.extractors) != 3 {
		t.Errorf("Expected 3 extractors in default chain, got %d", len(chain.extractors))
	}
}

/* TestChainDirectionExtractor_NilSafety verifies nil handling */
func TestChainDirectionExtractor_NilSafety(t *testing.T) {
	extractor := NewDefaultDirectionExtractor()

	/* Nil expression should not panic, should default to Long */
	result := extractor.Extract(nil)
	if result != "strategy.Long" {
		t.Errorf("Nil expression should default to Long, got %q", result)
	}
}
