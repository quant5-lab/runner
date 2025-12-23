package parser

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func assertNestedMemberStructure(t *testing.T, expr ast.Expression, expectedChain []string) {
	t.Helper()

	if len(expectedChain) == 0 {
		t.Fatal("expectedChain cannot be empty")
	}

	current := expr
	for i := len(expectedChain) - 1; i >= 0; i-- {
		if i == 0 {
			ident, ok := current.(*ast.Identifier)
			if !ok {
				t.Fatalf("Expected base Identifier, got %T", current)
			}
			if ident.Name != expectedChain[0] {
				t.Errorf("Base object: expected=%q got=%q", expectedChain[0], ident.Name)
			}
		} else {
			member, ok := current.(*ast.MemberExpression)
			if !ok {
				t.Fatalf("Level %d: expected MemberExpression, got %T", i, current)
			}
			if member.Computed {
				t.Errorf("Level %d: unexpected computed member access", i)
			}
			propIdent, ok := member.Property.(*ast.Identifier)
			if !ok {
				t.Fatalf("Level %d property: expected Identifier, got %T", i, member.Property)
			}
			if propIdent.Name != expectedChain[i] {
				t.Errorf("Level %d property: expected=%q got=%q", i, expectedChain[i], propIdent.Name)
			}
			current = member.Object
		}
	}
}

func TestMemberExpression_NestingDepths(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedChain []string
	}{
		{
			name:          "two-level",
			input:         "x = strategy.cash",
			expectedChain: []string{"strategy", "cash"},
		},
		{
			name:          "three-level",
			input:         "x = strategy.commission.percent",
			expectedChain: []string{"strategy", "commission", "percent"},
		},
		{
			name:          "four-level",
			input:         "x = a.b.c.d",
			expectedChain: []string{"a", "b", "c", "d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if len(program.Body) != 1 {
				t.Fatalf("Expected 1 statement, got %d", len(program.Body))
			}

			varDecl := program.Body[0].(*ast.VariableDeclaration)
			init := varDecl.Declarations[0].Init

			assertNestedMemberStructure(t, init, tt.expectedChain)
		})
	}
}

func TestMemberExpression_SyntacticContexts(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedChain []string
	}{
		{
			name:          "as assignment value",
			input:         "x = strategy.commission.percent",
			expectedChain: []string{"strategy", "commission", "percent"},
		},
		{
			name:          "as function argument",
			input:         "f(strategy.commission.percent)",
			expectedChain: []string{"strategy", "commission", "percent"},
		},
		{
			name:          "as named argument value",
			input:         "f(val=strategy.commission.percent)",
			expectedChain: []string{"strategy", "commission", "percent"},
		},
		{
			name:          "in binary expression",
			input:         "x = strategy.commission.percent + 0.1",
			expectedChain: []string{"strategy", "commission", "percent"},
		},
		{
			name:          "in comparison",
			input:         "x = strategy.commission.percent > 0",
			expectedChain: []string{"strategy", "commission", "percent"},
		},
		{
			name:          "in ternary condition",
			input:         "x = strategy.commission.percent ? 1 : 0",
			expectedChain: []string{"strategy", "commission", "percent"},
		},
		{
			name:          "in ternary consequent",
			input:         "x = cond ? strategy.commission.percent : 0",
			expectedChain: []string{"strategy", "commission", "percent"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			// Find member expression in AST (may be nested in various nodes)
			foundMember := findMemberExpression(program, tt.expectedChain)

			if foundMember == nil {
				t.Fatal("Expected member expression not found in AST")
			}

			assertNestedMemberStructure(t, foundMember, tt.expectedChain)
		})
	}
}

func TestMemberExpression_SpecialIdentifiers(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedChain []string
	}{
		{
			name:          "underscores",
			input:         "x = my_namespace.my_sub.my_prop",
			expectedChain: []string{"my_namespace", "my_sub", "my_prop"},
		},
		{
			name:          "numbers",
			input:         "x = ta2.sma20.value100",
			expectedChain: []string{"ta2", "sma20", "value100"},
		},
		{
			name:          "mixed case",
			input:         "x = MyNamespace.SubModule.PropertyName",
			expectedChain: []string{"MyNamespace", "SubModule", "PropertyName"},
		},
		{
			name:          "long identifiers",
			input:         "x = " + strings.Repeat("a", 50) + "." + strings.Repeat("b", 50) + "." + strings.Repeat("c", 50),
			expectedChain: []string{strings.Repeat("a", 50), strings.Repeat("b", 50), strings.Repeat("c", 50)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if len(program.Body) != 1 {
				t.Fatalf("Expected 1 statement, got %d", len(program.Body))
			}

			varDecl := program.Body[0].(*ast.VariableDeclaration)
			init := varDecl.Declarations[0].Init

			assertNestedMemberStructure(t, init, tt.expectedChain)
		})
	}
}

func TestMemberExpression_RealWorldPatterns(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		chains [][]string // Multiple member expressions expected
	}{
		{
			name: "strategy configuration",
			input: `x = strategy.cash
y = strategy.commission.percent`,
			chains: [][]string{
				{"strategy", "cash"},
				{"strategy", "commission", "percent"},
			},
		},
		{
			name: "multiple nested namespaces",
			input: `x = ta.sma(close, 20)
y = strategy.commission.percent
z = request.security.data.close`,
			chains: [][]string{
				{"ta", "sma"},
				{"strategy", "commission", "percent"},
				{"request", "security", "data", "close"},
			},
		},
		{
			name: "conditional with nested member",
			input: `if strategy.commission.percent > 0
    x = 1`,
			chains: [][]string{
				{"strategy", "commission", "percent"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			// Collect all non-computed member expressions
			foundMembers := collectMemberExpressions(program)

			if len(foundMembers) < len(tt.chains) {
				t.Errorf("Expected at least %d member expressions, found %d", len(tt.chains), len(foundMembers))
			}

			// Verify each expected chain is present
			for _, expectedChain := range tt.chains {
				found := false
				for _, foundChain := range foundMembers {
					if len(foundChain) == len(expectedChain) {
						matches := true
						for i := range foundChain {
							if foundChain[i] != expectedChain[i] {
								matches = false
								break
							}
						}
						if matches {
							found = true
							break
						}
					}
				}
				if !found {
					t.Errorf("Expected chain %v not found in AST", expectedChain)
				}
			}
		})
	}
}

func TestMemberExpression_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedChain []string
	}{
		{
			name:          "ta namespace",
			input:         "x = ta.sma(close, 20)",
			expectedChain: []string{"ta", "sma"},
		},
		{
			name:          "math namespace",
			input:         "x = math.max(a, b)",
			expectedChain: []string{"math", "max"},
		},
		{
			name:          "request namespace",
			input:         "x = request.security(symbol, tf, close)",
			expectedChain: []string{"request", "security"},
		},
		{
			name:          "strategy namespace",
			input:         "x = strategy.entry(id, direction)",
			expectedChain: []string{"strategy", "entry"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			// Find call expression
			callExpr := findCallExpression(program)

			if callExpr == nil {
				t.Fatal("CallExpression not found")
			}

			assertNestedMemberStructure(t, callExpr.Callee, tt.expectedChain)
		})
	}
}

func TestMemberExpression_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{
			name:      "single identifier - not member expression",
			input:     "x = value",
			shouldErr: false,
		},
		{
			name:      "two-level member",
			input:     "x = a.b",
			shouldErr: false,
		},
		{
			name:      "three-level member",
			input:     "x = a.b.c",
			shouldErr: false,
		},
		{
			name:      "four-level member",
			input:     "x = a.b.c.d",
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.input)
			if tt.shouldErr {
				if err == nil {
					t.Error("Expected parse error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected parse error: %v", err)
			}

			converter := NewConverter()
			_, err = converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}
		})
	}
}

func TestMemberExpression_ConverterRobustness(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		description string
	}{
		{
			name:        "single property",
			input:       "x = a.b",
			expectError: false,
			description: "buildNestedMemberExpression with single property",
		},
		{
			name:        "two properties",
			input:       "x = a.b.c",
			expectError: false,
			description: "buildNestedMemberExpression with two properties",
		},
		{
			name:        "five properties",
			input:       "x = a.b.c.d",
			expectError: false,
			description: "buildNestedMemberExpression with multiple properties",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected conversion error for: %s", tt.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Conversion failed for %s: %v", tt.description, err)
			}

			if len(program.Body) == 0 {
				t.Fatalf("Empty program body for: %s", tt.description)
			}
		})
	}
}

func TestMemberExpression_ComputedVsNonComputed(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectComputed bool
		expectedAccess string
		description    string
	}{
		{
			name:           "non-computed dot notation",
			input:          "x = strategy.cash",
			expectComputed: false,
			expectedAccess: "property access via dot",
			description:    "Dot notation creates non-computed member expression",
		},
		{
			name:           "non-computed multi-level",
			input:          "x = strategy.commission.percent",
			expectComputed: false,
			expectedAccess: "nested property access",
			description:    "Multi-level dot notation remains non-computed at each level",
		},
		{
			name:           "computed bracket notation",
			input:          "x = close[1]",
			expectComputed: true,
			expectedAccess: "array subscript",
			description:    "Bracket notation creates computed member expression",
		},
		{
			name:           "computed with zero offset",
			input:          "x = close[0]",
			expectComputed: true,
			expectedAccess: "array subscript",
			description:    "Even [0] creates computed member expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Parser creation failed: %v", err)
			}

			script, err := p.ParseString("", tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if len(program.Body) != 1 {
				t.Fatalf("Expected 1 statement, got %d", len(program.Body))
			}

			varDecl := program.Body[0].(*ast.VariableDeclaration)
			init := varDecl.Declarations[0].Init

			member, ok := init.(*ast.MemberExpression)
			if !ok {
				t.Fatalf("Expected MemberExpression, got %T", init)
			}

			if member.Computed != tt.expectComputed {
				t.Errorf("Computed flag: expected=%v got=%v (%s)",
					tt.expectComputed, member.Computed, tt.description)
			}
		})
	}
}

func extractMemberChain(member *ast.MemberExpression) []string {
	var chain []string
	current := ast.Expression(member)

	for {
		if m, ok := current.(*ast.MemberExpression); ok {
			if propIdent, ok := m.Property.(*ast.Identifier); ok {
				chain = append([]string{propIdent.Name}, chain...)
			}
			current = m.Object
		} else if ident, ok := current.(*ast.Identifier); ok {
			chain = append([]string{ident.Name}, chain...)
			break
		} else {
			break
		}
	}

	return chain
}

type astVisitor struct {
	visitNode       func(ast.Node) bool
	visitExpression func(ast.Expression) bool
}

func (v *astVisitor) traverseProgram(program *ast.Program) {
	for _, node := range program.Body {
		v.traverseNode(node)
	}
}

func (v *astVisitor) traverseNode(node ast.Node) bool {
	if v.visitNode != nil && !v.visitNode(node) {
		return false
	}

	switch n := node.(type) {
	case *ast.VariableDeclaration:
		for _, decl := range n.Declarations {
			if !v.traverseExpression(decl.Init) {
				return false
			}
		}
	case *ast.ExpressionStatement:
		return v.traverseExpression(n.Expression)
	case *ast.IfStatement:
		if !v.traverseExpression(n.Test) {
			return false
		}
		for _, stmt := range n.Consequent {
			if !v.traverseNode(stmt) {
				return false
			}
		}
		for _, stmt := range n.Alternate {
			if !v.traverseNode(stmt) {
				return false
			}
		}
	}
	return true
}

func (v *astVisitor) traverseExpression(expr ast.Expression) bool {
	if expr == nil {
		return true
	}

	if v.visitExpression != nil && !v.visitExpression(expr) {
		return false
	}

	switch e := expr.(type) {
	case *ast.MemberExpression:
		return v.traverseExpression(e.Object)
	case *ast.CallExpression:
		if !v.traverseExpression(e.Callee) {
			return false
		}
		for _, arg := range e.Arguments {
			if obj, ok := arg.(*ast.ObjectExpression); ok {
				for _, prop := range obj.Properties {
					if !v.traverseExpression(prop.Value) {
						return false
					}
				}
			} else if !v.traverseExpression(arg) {
				return false
			}
		}
	case *ast.BinaryExpression:
		if !v.traverseExpression(e.Left) {
			return false
		}
		return v.traverseExpression(e.Right)
	case *ast.ConditionalExpression:
		if !v.traverseExpression(e.Test) {
			return false
		}
		if !v.traverseExpression(e.Consequent) {
			return false
		}
		return v.traverseExpression(e.Alternate)
	}
	return true
}

func findMemberExpression(program *ast.Program, expectedChain []string) ast.Expression {
	var result ast.Expression
	visitor := &astVisitor{
		visitExpression: func(expr ast.Expression) bool {
			if member, ok := expr.(*ast.MemberExpression); ok && !member.Computed {
				chain := extractMemberChain(member)
				if len(chain) == len(expectedChain) {
					matches := true
					for i := range chain {
						if chain[i] != expectedChain[i] {
							matches = false
							break
						}
					}
					if matches {
						result = member
						return false
					}
				}
			}
			return true
		},
	}
	visitor.traverseProgram(program)
	return result
}

func collectMemberExpressions(program *ast.Program) [][]string {
	var members [][]string
	visitor := &astVisitor{
		visitExpression: func(expr ast.Expression) bool {
			if member, ok := expr.(*ast.MemberExpression); ok && !member.Computed {
				chain := extractMemberChain(member)
				if len(chain) >= 2 {
					members = append(members, chain)
				}
			}
			return true
		},
	}
	visitor.traverseProgram(program)
	return members
}

func findCallExpression(program *ast.Program) *ast.CallExpression {
	var result *ast.CallExpression
	visitor := &astVisitor{
		visitExpression: func(expr ast.Expression) bool {
			if call, ok := expr.(*ast.CallExpression); ok {
				result = call
				return false
			}
			return true
		},
	}
	visitor.traverseProgram(program)
	return result
}
