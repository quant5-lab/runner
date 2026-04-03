package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestBuiltinUsageDetector(t *testing.T) {
	tests := []struct {
		name       string
		targets    []string
		program    *ast.Program
		wantFound  []string
		wantAbsent []string
	}{
		{
			name:    "identifier in variable declaration init",
			targets: []string{"dayofweek", "hour"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{Init: &ast.Identifier{Name: "dayofweek"}},
						},
					},
				},
			},
			wantFound:  []string{"dayofweek"},
			wantAbsent: []string{"hour"},
		},
		{
			name:    "binary expression operands",
			targets: []string{"hour", "minute"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Left:     &ast.Identifier{Name: "hour"},
							Operator: "+",
							Right:    &ast.Identifier{Name: "minute"},
						},
					},
				},
			},
			wantFound: []string{"hour", "minute"},
		},
		{
			name:    "if condition",
			targets: []string{"month"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.IfStatement{
						Test: &ast.BinaryExpression{
							Left:     &ast.Identifier{Name: "month"},
							Operator: "==",
							Right:    &ast.Literal{Value: float64(12)},
						},
						Consequent: []ast.Node{},
					},
				},
			},
			wantFound: []string{"month"},
		},
		{
			name:    "if consequent and alternate bodies",
			targets: []string{"hour", "minute"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.IfStatement{
						Test:       &ast.Literal{Value: true},
						Consequent: []ast.Node{&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "hour"}}},
						Alternate:  []ast.Node{&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "minute"}}},
					},
				},
			},
			wantFound: []string{"hour", "minute"},
		},
		{
			name:    "subscript member expression",
			targets: []string{"dayofweek"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "dayofweek"},
							Property: &ast.Literal{Value: float64(1)},
							Computed: true,
						},
					},
				},
			},
			wantFound: []string{"dayofweek"},
		},
		{
			name:    "call expression arguments",
			targets: []string{"year"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee:    &ast.Identifier{Name: "plot"},
							Arguments: []ast.Expression{&ast.Identifier{Name: "year"}},
						},
					},
				},
			},
			wantFound: []string{"year"},
		},
		{
			name:    "for loop body and bounds",
			targets: []string{"second", "hour"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.ForStatement{
						Counter: "i",
						From:    &ast.Identifier{Name: "hour"},
						To:      &ast.Literal{Value: float64(10)},
						Body: []ast.Node{
							&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "second"}},
						},
					},
				},
			},
			wantFound: []string{"second", "hour"},
		},
		{
			name:    "for-in collection and body",
			targets: []string{"minute", "year"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.ForInStatement{
						ElementVar: "val",
						Collection: &ast.Identifier{Name: "minute"},
						Body: []ast.Node{
							&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "year"}},
						},
					},
				},
			},
			wantFound: []string{"minute", "year"},
		},
		{
			name:    "while condition and body",
			targets: []string{"hour", "second"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.WhileStatement{
						Condition: &ast.Identifier{Name: "hour"},
						Body: []ast.Node{
							&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "second"}},
						},
					},
				},
			},
			wantFound: []string{"hour", "second"},
		},
		{
			name:    "conditional (ternary) all branches",
			targets: []string{"dayofweek", "hour", "minute", "year"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{Init: &ast.ConditionalExpression{
								Test:       &ast.Identifier{Name: "dayofweek"},
								Consequent: &ast.Identifier{Name: "hour"},
								Alternate:  &ast.Identifier{Name: "minute"},
							}},
						},
					},
				},
			},
			wantFound:  []string{"dayofweek", "hour", "minute"},
			wantAbsent: []string{"year"},
		},
		{
			name:    "unary expression operand",
			targets: []string{"dayofweek"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.UnaryExpression{
							Operator: "-",
							Argument: &ast.Identifier{Name: "dayofweek"},
						},
					},
				},
			},
			wantFound: []string{"dayofweek"},
		},
		{
			name:    "logical expression both sides",
			targets: []string{"dayofweek", "hour"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.LogicalExpression{
							Operator: "and",
							Left: &ast.BinaryExpression{
								Left: &ast.Identifier{Name: "dayofweek"}, Operator: "==", Right: &ast.Literal{Value: float64(1)},
							},
							Right: &ast.BinaryExpression{
								Left: &ast.Identifier{Name: "hour"}, Operator: ">", Right: &ast.Literal{Value: float64(9)},
							},
						},
					},
				},
			},
			wantFound: []string{"dayofweek", "hour"},
		},
		{
			name:    "arrow function body",
			targets: []string{"hour"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{Init: &ast.ArrowFunctionExpression{
								Params: []ast.Identifier{{Name: "x"}},
								Body: []ast.Node{
									&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "hour"}},
								},
							}},
						},
					},
				},
			},
			wantFound: []string{"hour"},
		},
		{
			name:    "deeply nested: call in binary in conditional",
			targets: []string{"month", "year"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{Init: &ast.ConditionalExpression{
								Test: &ast.Literal{Value: true},
								Consequent: &ast.BinaryExpression{
									Left: &ast.CallExpression{
										Callee:    &ast.Identifier{Name: "fn"},
										Arguments: []ast.Expression{&ast.Identifier{Name: "month"}},
									},
									Operator: "+",
									Right:    &ast.Identifier{Name: "year"},
								},
								Alternate: &ast.Literal{Value: float64(0)},
							}},
						},
					},
				},
			},
			wantFound: []string{"month", "year"},
		},
		{
			name:    "multiple targets across multiple statements",
			targets: []string{"hour", "minute", "second"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "hour"}},
					&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "minute"}},
				},
			},
			wantFound:  []string{"hour", "minute"},
			wantAbsent: []string{"second"},
		},
		{
			name:       "empty program body",
			targets:    []string{"hour"},
			program:    &ast.Program{Body: []ast.Node{}},
			wantAbsent: []string{"hour"},
		},
		{
			name:    "empty targets list detects nothing",
			targets: []string{},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "hour"}},
				},
			},
			wantAbsent: []string{"hour"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewBuiltinUsageDetector(tt.targets)
			found := detector.Detect(tt.program)

			for _, name := range tt.wantFound {
				if !found[name] {
					t.Errorf("expected %q detected", name)
				}
			}
			for _, name := range tt.wantAbsent {
				if found[name] {
					t.Errorf("expected %q not detected", name)
				}
			}
		})
	}
}

func TestBuiltinUsageDetector_NilProgram(t *testing.T) {
	detector := NewBuiltinUsageDetector([]string{"hour"})
	found := detector.Detect(nil)

	if found != nil {
		t.Error("nil program should return nil")
	}
}

func TestBuiltinUsageDetector_MemberExpressions(t *testing.T) {
	tests := []struct {
		name       string
		members    []string
		program    *ast.Program
		wantFound  []string
		wantAbsent []string
	}{
		{
			name:    "session.isfirstbar in expression",
			members: []string{"session.isfirstbar", "session.islastbar"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "session"},
							Property: &ast.Identifier{Name: "isfirstbar"},
						},
					},
				},
			},
			wantFound:  []string{"session.isfirstbar"},
			wantAbsent: []string{"session.islastbar"},
		},
		{
			name:    "session.isfirstbar subscript access",
			members: []string{"session.isfirstbar"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.MemberExpression{
							Object: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "session"},
								Property: &ast.Identifier{Name: "isfirstbar"},
							},
							Property: &ast.Literal{Value: float64(1)},
							Computed: true,
						},
					},
				},
			},
			wantFound: []string{"session.isfirstbar"},
		},
		{
			name:    "multiple session members in binary",
			members: []string{"session.isfirstbar", "session.islastbar"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Left: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "session"},
								Property: &ast.Identifier{Name: "isfirstbar"},
							},
							Operator: "or",
							Right: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "session"},
								Property: &ast.Identifier{Name: "islastbar"},
							},
						},
					},
				},
			},
			wantFound: []string{"session.isfirstbar", "session.islastbar"},
		},
		{
			name:    "member in if condition",
			members: []string{"session.isfirstbar"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.IfStatement{
						Test: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "session"},
							Property: &ast.Identifier{Name: "isfirstbar"},
						},
						Consequent: []ast.Node{},
					},
				},
			},
			wantFound: []string{"session.isfirstbar"},
		},
		{
			name:    "member in call argument",
			members: []string{"session.islastbar"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "plot"},
							Arguments: []ast.Expression{
								&ast.MemberExpression{
									Object:   &ast.Identifier{Name: "session"},
									Property: &ast.Identifier{Name: "islastbar"},
								},
							},
						},
					},
				},
			},
			wantFound: []string{"session.islastbar"},
		},
		{
			name:    "non-target member expression ignored",
			members: []string{"session.isfirstbar"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "syminfo"},
							Property: &ast.Identifier{Name: "ticker"},
						},
					},
				},
			},
			wantAbsent: []string{"session.isfirstbar", "syminfo.ticker"},
		},
		{
			name:    "member in variable declaration",
			members: []string{"session.isfirstbar_regular"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{Init: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "session"},
								Property: &ast.Identifier{Name: "isfirstbar_regular"},
							}},
						},
					},
				},
			},
			wantFound: []string{"session.isfirstbar_regular"},
		},
		{
			name:    "member in ternary",
			members: []string{"session.isfirstbar", "session.islastbar_regular"},
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{Init: &ast.ConditionalExpression{
								Test: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "session"},
									Property: &ast.Identifier{Name: "isfirstbar"},
								},
								Consequent: &ast.Literal{Value: float64(1)},
								Alternate: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "session"},
									Property: &ast.Identifier{Name: "islastbar_regular"},
								},
							}},
						},
					},
				},
			},
			wantFound: []string{"session.isfirstbar", "session.islastbar_regular"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewBuiltinUsageDetectorWithMembers(nil, tt.members)
			found := detector.Detect(tt.program)

			for _, name := range tt.wantFound {
				if !found[name] {
					t.Errorf("expected %q detected", name)
				}
			}
			for _, name := range tt.wantAbsent {
				if found[name] {
					t.Errorf("expected %q not detected", name)
				}
			}
		})
	}
}

/* Mixed detection: both identifiers and member expressions */
func TestBuiltinUsageDetector_MixedDetection(t *testing.T) {
	detector := NewBuiltinUsageDetectorWithMembers(
		[]string{"dayofweek", "hour"},
		[]string{"session.isfirstbar"},
	)

	program := &ast.Program{
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "dayofweek"},
					Operator: "and",
					Right: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "session"},
						Property: &ast.Identifier{Name: "isfirstbar"},
					},
				},
			},
		},
	}

	found := detector.Detect(program)

	if !found["dayofweek"] {
		t.Error("expected dayofweek detected")
	}
	if !found["session.isfirstbar"] {
		t.Error("expected session.isfirstbar detected")
	}
	if found["hour"] {
		t.Error("hour not in program, should not be detected")
	}
}
