package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

type Converter struct{}

func NewConverter() *Converter {
	return &Converter{}
}

func (c *Converter) ToESTree(script *Script) (*ast.Program, error) {
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body:     []ast.Node{},
	}

	for _, stmt := range script.Statements {
		node, err := c.convertStatement(stmt)
		if err != nil {
			return nil, err
		}
		if node != nil {
			program.Body = append(program.Body, node)
		}
	}

	return program, nil
}

func (c *Converter) convertStatement(stmt *Statement) (ast.Node, error) {
	if stmt.Assignment != nil {
		init, err := c.convertExpression(stmt.Assignment.Value)
		if err != nil {
			return nil, err
		}
		return &ast.VariableDeclaration{
			NodeType: ast.TypeVariableDeclaration,
			Declarations: []ast.VariableDeclarator{
				{
					NodeType: ast.TypeVariableDeclarator,
					ID: ast.Identifier{
						NodeType: ast.TypeIdentifier,
						Name:     stmt.Assignment.Name,
					},
					Init: init,
				},
			},
			Kind: "let",
		}, nil
	}

	if stmt.Reassignment != nil {
		init, err := c.convertExpression(stmt.Reassignment.Value)
		if err != nil {
			return nil, err
		}
		return &ast.VariableDeclaration{
			NodeType: ast.TypeVariableDeclaration,
			Declarations: []ast.VariableDeclarator{
				{
					NodeType: ast.TypeVariableDeclarator,
					ID: ast.Identifier{
						NodeType: ast.TypeIdentifier,
						Name:     stmt.Reassignment.Name,
					},
					Init: init,
				},
			},
			Kind: "var",
		}, nil
	}

	if stmt.If != nil {
		test, err := c.convertComparison(stmt.If.Condition)
		if err != nil {
			return nil, err
		}

		consequent := []ast.Node{}
		if stmt.If.Body != nil {
			node, err := c.convertStatement(stmt.If.Body)
			if err != nil {
				return nil, err
			}
			if node != nil {
				consequent = append(consequent, node)
			}
		}

		return &ast.IfStatement{
			NodeType:   ast.TypeIfStatement,
			Test:       test,
			Consequent: consequent,
			Alternate:  []ast.Node{},
		}, nil
	}

	if stmt.Expression != nil {
		expr, err := c.convertExpression(stmt.Expression.Expr)
		if err != nil {
			return nil, err
		}
		return &ast.ExpressionStatement{
			NodeType:   ast.TypeExpressionStatement,
			Expression: expr,
		}, nil
	}

	return nil, fmt.Errorf("empty statement")
}

func (c *Converter) convertExpression(expr *Expression) (ast.Expression, error) {
	if expr.Ternary != nil {
		return c.convertTernaryExpr(expr.Ternary)
	}
	if expr.MemberAccess != nil {
		return &ast.MemberExpression{
			NodeType: ast.TypeMemberExpression,
			Object: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     expr.MemberAccess.Object,
			},
			Property: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     expr.MemberAccess.Property,
			},
			Computed: false,
		}, nil
	}
	if expr.Call != nil {
		return c.convertCallExpr(expr.Call)
	}
	if expr.Ident != nil {
		return &ast.MemberExpression{
			NodeType: ast.TypeMemberExpression,
			Object: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     *expr.Ident,
			},
			Property: &ast.Literal{
				NodeType: ast.TypeLiteral,
				Value:    0,
				Raw:      "0",
			},
			Computed: true,
		}, nil
	}
	if expr.Number != nil {
		return &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    *expr.Number,
			Raw:      fmt.Sprintf("%v", *expr.Number),
		}, nil
	}
	if expr.String != nil {
		cleaned := strings.Trim(strings.Trim(*expr.String, `"`), `'`)
		return &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    cleaned,
			Raw:      fmt.Sprintf("'%s'", cleaned),
		}, nil
	}
	return nil, fmt.Errorf("empty expression")
}

func (c *Converter) convertComparison(comp *Comparison) (ast.Expression, error) {
	left, err := c.convertComparisonTerm(comp.Left)
	if err != nil {
		return nil, err
	}

	// No operator means just a simple expression
	if comp.Op == nil {
		return left, nil
	}

	right, err := c.convertComparisonTerm(comp.Right)
	if err != nil {
		return nil, err
	}

	return &ast.BinaryExpression{
		NodeType: ast.TypeBinaryExpression,
		Operator: *comp.Op,
		Left:     left,
		Right:    right,
	}, nil
}

func (c *Converter) convertComparisonTerm(term *ComparisonTerm) (ast.Expression, error) {
	if term.Postfix != nil {
		return c.convertPostfixExpr(term.Postfix)
	}

	if term.MemberAccess != nil {
		return &ast.MemberExpression{
			NodeType: ast.TypeMemberExpression,
			Object: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     term.MemberAccess.Object,
			},
			Property: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     term.MemberAccess.Property,
			},
			Computed: false,
		}, nil
	}
	if term.True != nil {
		return &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    true,
			Raw:      "true",
		}, nil
	}
	if term.False != nil {
		return &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    false,
			Raw:      "false",
		}, nil
	}
	if term.Ident != nil {
		return &ast.MemberExpression{
			NodeType: ast.TypeMemberExpression,
			Object: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     *term.Ident,
			},
			Property: &ast.Literal{
				NodeType: ast.TypeLiteral,
				Value:    0,
				Raw:      "0",
			},
			Computed: true,
		}, nil
	}
	if term.Number != nil {
		return &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    *term.Number,
			Raw:      fmt.Sprintf("%v", *term.Number),
		}, nil
	}
	if term.String != nil {
		cleaned := strings.Trim(*term.String, `"`)
		return &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    cleaned,
			Raw:      fmt.Sprintf("'%s'", cleaned),
		}, nil
	}
	return nil, fmt.Errorf("empty comparison term")
}

func (c *Converter) convertCallExpr(call *CallExpr) (ast.Expression, error) {
	var callee ast.Expression

	if call.Callee.MemberAccess != nil {
		// ta.sma(...) -> MemberExpression as callee
		callee = &ast.MemberExpression{
			NodeType: ast.TypeMemberExpression,
			Object: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     call.Callee.MemberAccess.Object,
			},
			Property: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     call.Callee.MemberAccess.Property,
			},
			Computed: false,
		}
	} else if call.Callee.Ident != nil {
		// plot(...) -> Identifier as callee
		callee = &ast.Identifier{
			NodeType: ast.TypeIdentifier,
			Name:     *call.Callee.Ident,
		}
	} else {
		return nil, fmt.Errorf("empty callee")
	}

	args := []ast.Expression{}
	namedArgs := make(map[string]ast.Expression)

	for _, arg := range call.Args {
		converted, err := c.convertTernaryExpr(arg.Value)
		if err != nil {
			return nil, err
		}

		if arg.Name != nil {
			namedArgs[*arg.Name] = converted
		} else {
			args = append(args, converted)
		}
	}

	if len(namedArgs) > 0 {
		props := []ast.Property{}
		for key, val := range namedArgs {
			props = append(props, ast.Property{
				NodeType: ast.TypeProperty,
				Key: &ast.Identifier{
					NodeType: ast.TypeIdentifier,
					Name:     key,
				},
				Value:     val,
				Kind:      "init",
				Method:    false,
				Shorthand: false,
				Computed:  false,
			})
		}
		args = append(args, &ast.ObjectExpression{
			NodeType:   ast.TypeObjectExpression,
			Properties: props,
		})
	}

	return &ast.CallExpression{
		NodeType:  ast.TypeCallExpression,
		Callee:    callee,
		Arguments: args,
	}, nil
}

func (c *Converter) convertPostfixExpr(postfix *PostfixExpr) (ast.Expression, error) {
	var baseExpr ast.Expression
	var err error

	if postfix.Primary.Call != nil {
		baseExpr, err = c.convertCallExpr(postfix.Primary.Call)
		if err != nil {
			return nil, err
		}
	} else if postfix.Primary.MemberAccess != nil {
		baseExpr = &ast.MemberExpression{
			NodeType: ast.TypeMemberExpression,
			Object: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     postfix.Primary.MemberAccess.Object,
			},
			Property: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     postfix.Primary.MemberAccess.Property,
			},
			Computed: false,
		}
	} else if postfix.Primary.Ident != nil {
		baseExpr = &ast.Identifier{
			NodeType: ast.TypeIdentifier,
			Name:     *postfix.Primary.Ident,
		}
	} else {
		return nil, fmt.Errorf("postfix primary must have call, member access, or ident")
	}

	if postfix.Subscript != nil {
		indexExpr, err := c.convertArithExpr(postfix.Subscript)
		if err != nil {
			return nil, err
		}

		return &ast.MemberExpression{
			NodeType: ast.TypeMemberExpression,
			Object:   baseExpr,
			Property: indexExpr,
			Computed: true,
		}, nil
	}

	return baseExpr, nil
}

func (c *Converter) convertValue(val *Value) (ast.Expression, error) {
	if val.Postfix != nil {
		return c.convertPostfixExpr(val.Postfix)
	}

	if val.MemberAccess != nil {
		return &ast.MemberExpression{
			NodeType: ast.TypeMemberExpression,
			Object: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     val.MemberAccess.Object,
			},
			Property: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     val.MemberAccess.Property,
			},
			Computed: false,
		}, nil
	}
	if val.True != nil {
		return &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    true,
			Raw:      "true",
		}, nil
	}
	if val.False != nil {
		return &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    false,
			Raw:      "false",
		}, nil
	}
	if val.Ident != nil {
		return &ast.MemberExpression{
			NodeType: ast.TypeMemberExpression,
			Object: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     *val.Ident,
			},
			Property: &ast.Literal{
				NodeType: ast.TypeLiteral,
				Value:    0,
				Raw:      "0",
			},
			Computed: true,
		}, nil
	}
	if val.Number != nil {
		return &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    *val.Number,
			Raw:      fmt.Sprintf("%v", *val.Number),
		}, nil
	}
	if val.String != nil {
		cleaned := strings.Trim(strings.Trim(*val.String, `"`), `'`)
		return &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    cleaned,
			Raw:      fmt.Sprintf("'%s'", cleaned),
		}, nil
	}
	return nil, fmt.Errorf("empty value")
}

func (c *Converter) parseCallee(name string) (ast.Expression, error) {
	if strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		return &ast.MemberExpression{
			NodeType: ast.TypeMemberExpression,
			Object: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     parts[0],
			},
			Property: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     parts[1],
			},
			Computed: false,
		}, nil
	}
	return &ast.Identifier{
		NodeType: ast.TypeIdentifier,
		Name:     name,
	}, nil
}

func (c *Converter) convertTernaryExpr(ternary *TernaryExpr) (ast.Expression, error) {
	// Check if it's actually a ternary (has ? :) or just a simple expression
	if ternary.TrueVal == nil && ternary.FalseVal == nil {
		// No ternary, just convert the condition as expression
		return c.convertOrExpr(ternary.Condition)
	}

	test, err := c.convertOrExpr(ternary.Condition)
	if err != nil {
		return nil, err
	}

	consequent, err := c.convertExpression(ternary.TrueVal)
	if err != nil {
		return nil, err
	}

	alternate, err := c.convertExpression(ternary.FalseVal)
	if err != nil {
		return nil, err
	}

	return &ast.ConditionalExpression{
		NodeType:   ast.TypeConditionalExpression,
		Test:       test,
		Consequent: consequent,
		Alternate:  alternate,
	}, nil
}

func (c *Converter) convertOrExpr(or *OrExpr) (ast.Expression, error) {
	left, err := c.convertAndExpr(or.Left)
	if err != nil {
		return nil, err
	}

	if or.Right == nil {
		return left, nil
	}

	right, err := c.convertOrExpr(or.Right)
	if err != nil {
		return nil, err
	}

	return &ast.LogicalExpression{
		NodeType: ast.TypeLogicalExpression,
		Operator: "||",
		Left:     left,
		Right:    right,
	}, nil
}

func (c *Converter) convertAndExpr(and *AndExpr) (ast.Expression, error) {
	left, err := c.convertCompExpr(and.Left)
	if err != nil {
		return nil, err
	}

	if and.Right == nil {
		return left, nil
	}

	right, err := c.convertAndExpr(and.Right)
	if err != nil {
		return nil, err
	}

	return &ast.LogicalExpression{
		NodeType: ast.TypeLogicalExpression,
		Operator: "&&",
		Left:     left,
		Right:    right,
	}, nil
}

func (c *Converter) convertCompExpr(comp *CompExpr) (ast.Expression, error) {
	left, err := c.convertArithExpr(comp.Left)
	if err != nil {
		return nil, err
	}

	if comp.Op == nil || comp.Right == nil {
		return left, nil
	}

	right, err := c.convertCompExpr(comp.Right)
	if err != nil {
		return nil, err
	}

	return &ast.BinaryExpression{
		NodeType: ast.TypeBinaryExpression,
		Operator: *comp.Op,
		Left:     left,
		Right:    right,
	}, nil
}

func (c *Converter) convertArithExpr(arith *ArithExpr) (ast.Expression, error) {
	left, err := c.convertTerm(arith.Left)
	if err != nil {
		return nil, err
	}

	if arith.Op == nil || arith.Right == nil {
		return left, nil
	}

	right, err := c.convertArithExpr(arith.Right)
	if err != nil {
		return nil, err
	}

	return &ast.BinaryExpression{
		NodeType: ast.TypeBinaryExpression,
		Operator: *arith.Op,
		Left:     left,
		Right:    right,
	}, nil
}

func (c *Converter) convertTerm(term *Term) (ast.Expression, error) {
	left, err := c.convertFactor(term.Left)
	if err != nil {
		return nil, err
	}

	if term.Op == nil || term.Right == nil {
		return left, nil
	}

	right, err := c.convertTerm(term.Right)
	if err != nil {
		return nil, err
	}

	return &ast.BinaryExpression{
		NodeType: ast.TypeBinaryExpression,
		Operator: *term.Op,
		Left:     left,
		Right:    right,
	}, nil
}

func (c *Converter) convertFactor(factor *Factor) (ast.Expression, error) {
	if factor.Paren != nil {
		return c.convertTernaryExpr(factor.Paren)
	}

	if factor.Unary != nil {
		// Convert unary expression like -1 or +x
		operand, err := c.convertFactor(factor.Unary.Operand)
		if err != nil {
			return nil, err
		}

		return &ast.UnaryExpression{
			NodeType: ast.TypeUnaryExpression,
			Operator: factor.Unary.Op,
			Argument: operand,
			Prefix:   true,
		}, nil
	}

	if factor.Postfix != nil {
		return c.convertPostfixExpr(factor.Postfix)
	}

	if factor.MemberAccess != nil {
		return &ast.MemberExpression{
			NodeType: ast.TypeMemberExpression,
			Object: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     factor.MemberAccess.Object,
			},
			Property: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     factor.MemberAccess.Property,
			},
			Computed: false,
		}, nil
	}

	if factor.True != nil {
		return &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    true,
			Raw:      "true",
		}, nil
	}

	if factor.False != nil {
		return &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    false,
			Raw:      "false",
		}, nil
	}

	if factor.Ident != nil {
		// Special handling for built-in identifiers that are NOT series
		if *factor.Ident == "na" {
			return &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     *factor.Ident,
			}, nil
		}

		return &ast.MemberExpression{
			NodeType: ast.TypeMemberExpression,
			Object: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     *factor.Ident,
			},
			Property: &ast.Literal{
				NodeType: ast.TypeLiteral,
				Value:    0,
				Raw:      "0",
			},
			Computed: true,
		}, nil
	}

	if factor.Number != nil {
		return &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    *factor.Number,
			Raw:      fmt.Sprintf("%v", *factor.Number),
		}, nil
	}

	if factor.String != nil {
		cleaned := strings.Trim(strings.Trim(*factor.String, `"`), `'`)
		return &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    cleaned,
			Raw:      fmt.Sprintf("'%s'", cleaned),
		}, nil
	}

	return nil, fmt.Errorf("empty factor")
}

func (c *Converter) ToJSON(program *ast.Program) ([]byte, error) {
	return json.MarshalIndent(program, "", "  ")
}
