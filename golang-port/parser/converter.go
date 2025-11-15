package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/borisquantlab/pinescript-go/ast"
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
	if expr.Call != nil {
		return c.convertCallExpr(expr.Call)
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
		cleaned := strings.Trim(*expr.String, `"`)
		return &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    cleaned,
			Raw:      fmt.Sprintf("'%s'", cleaned),
		}, nil
	}
	return nil, fmt.Errorf("empty expression")
}

func (c *Converter) convertCallExpr(call *CallExpr) (ast.Expression, error) {
	fullName := call.Function
	if call.Namespace != nil {
		fullName = *call.Namespace + "." + call.Function
	}
	
	callee, err := c.parseCallee(fullName)
	if err != nil {
		return nil, err
	}

	args := []ast.Expression{}
	namedArgs := make(map[string]ast.Expression)

	for _, arg := range call.Args {
		converted, err := c.convertValue(arg.Value)
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

func (c *Converter) convertValue(val *Value) (ast.Expression, error) {
	if val.Member != nil {
		return &ast.MemberExpression{
			NodeType: ast.TypeMemberExpression,
			Object: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     val.Member.Object,
			},
			Property: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     val.Member.Property,
			},
			Computed: false,
		}, nil
	}
	if val.Boolean != nil {
		boolVal := *val.Boolean == "true"
		return &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    boolVal,
			Raw:      *val.Boolean,
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
		cleaned := strings.Trim(*val.String, `"`)
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



func (c *Converter) ToJSON(program *ast.Program) ([]byte, error) {
	return json.MarshalIndent(program, "", "  ")
}
