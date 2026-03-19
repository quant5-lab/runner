package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

type Converter struct {
	factory *StatementConverterFactory
}

func buildNestedMemberExpression(object string, properties []string) ast.Expression {
	var current ast.Expression = &ast.Identifier{
		NodeType: ast.TypeIdentifier,
		Name:     object,
	}

	for _, prop := range properties {
		current = &ast.MemberExpression{
			NodeType: ast.TypeMemberExpression,
			Object:   current,
			Property: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     prop,
			},
			Computed: false,
		}
	}

	return current
}

func NewConverter() *Converter {
	c := &Converter{}
	c.factory = NewStatementConverterFactory(
		c.convertExpression,
		c.convertOrExpr,
		c.convertArithExpr,
		c.convertStatement,
		c,
	)
	return c
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
	return c.factory.Convert(stmt)
}

func (c *Converter) convertExpression(expr *Expression) (ast.Expression, error) {
	if expr.ForInExpr != nil {
		return c.convertForInExprToStatement(expr.ForInExpr)
	}
	if expr.ForExpr != nil {
		return c.convertForExprToStatement(expr.ForExpr)
	}
	if expr.WhileExpr != nil {
		return c.convertWhileExprToStatement(expr.WhileExpr)
	}
	if expr.IfExpr != nil {
		return c.convertIfExprToStatement(expr.IfExpr)
	}
	if expr.SwitchExpr != nil {
		return c.convertSwitchExprToStatement(expr.SwitchExpr)
	}
	if expr.Array != nil {
		elements := []ast.Expression{}
		for _, elem := range expr.Array.Elements {
			astExpr, err := c.convertTernaryExpr(elem)
			if err != nil {
				return nil, err
			}
			elements = append(elements, astExpr)
		}
		return &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    elements,
			Raw:      "[...]",
		}, nil
	}
	if expr.Ternary != nil {
		return c.convertTernaryExpr(expr.Ternary)
	}
	if expr.MemberAccess != nil {
		return buildNestedMemberExpression(expr.MemberAccess.Object, expr.MemberAccess.Properties), nil
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
	if expr.HexColor != nil {
		return &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    *expr.HexColor,
			Raw:      fmt.Sprintf("'%s'", *expr.HexColor),
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
		return buildNestedMemberExpression(term.MemberAccess.Object, term.MemberAccess.Properties), nil
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
		callee = buildNestedMemberExpression(call.Callee.MemberAccess.Object, call.Callee.MemberAccess.Properties)
	} else if call.Callee.Ident != nil {
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
		converted, err := c.convertExpression(arg.Value)
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

	if postfix.Primary.Paren != nil {
		baseExpr, err = c.convertExpression(postfix.Primary.Paren)
		if err != nil {
			return nil, err
		}
	} else if postfix.Primary.Call != nil {
		baseExpr, err = c.convertCallExpr(postfix.Primary.Call)
		if err != nil {
			return nil, err
		}
	} else if postfix.Primary.MemberAccess != nil {
		baseExpr = buildNestedMemberExpression(postfix.Primary.MemberAccess.Object, postfix.Primary.MemberAccess.Properties)
	} else if postfix.Primary.Ident != nil {
		baseExpr = &ast.Identifier{
			NodeType: ast.TypeIdentifier,
			Name:     *postfix.Primary.Ident,
		}
	} else {
		return nil, fmt.Errorf("postfix primary must have paren, call, member access, or ident")
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
		return buildNestedMemberExpression(val.MemberAccess.Object, val.MemberAccess.Properties), nil
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

func (c *Converter) buildLeftAssociativeArithExpr(leftOperand ast.Expression, op string, rightGrammar *ArithExpr) (ast.Expression, error) {
	operands := []ast.Expression{leftOperand}
	operators := []string{op}

	current := rightGrammar
	for current != nil {
		operand, err := c.convertTerm(current.Left)
		if err != nil {
			return nil, err
		}
		operands = append(operands, operand)

		if current.Op != nil && current.Right != nil {
			operators = append(operators, *current.Op)
			current = current.Right
		} else {
			current = nil
		}
	}

	result := operands[0]
	for i := 0; i < len(operators); i++ {
		result = &ast.BinaryExpression{
			NodeType: ast.TypeBinaryExpression,
			Operator: operators[i],
			Left:     result,
			Right:    operands[i+1],
		}
	}

	return result, nil
}

func (c *Converter) buildLeftAssociativeTerm(leftOperand ast.Expression, op string, rightGrammar *Term) (ast.Expression, error) {
	operands := []ast.Expression{leftOperand}
	operators := []string{op}

	current := rightGrammar
	for current != nil {
		operand, err := c.convertFactor(current.Left)
		if err != nil {
			return nil, err
		}
		operands = append(operands, operand)

		if current.Op != nil && current.Right != nil {
			operators = append(operators, *current.Op)
			current = current.Right
		} else {
			current = nil
		}
	}

	result := operands[0]
	for i := 0; i < len(operators); i++ {
		result = &ast.BinaryExpression{
			NodeType: ast.TypeBinaryExpression,
			Operator: operators[i],
			Left:     result,
			Right:    operands[i+1],
		}
	}

	return result, nil
}

func (c *Converter) convertArithExpr(arith *ArithExpr) (ast.Expression, error) {
	left, err := c.convertTerm(arith.Left)
	if err != nil {
		return nil, err
	}

	if arith.Op == nil || arith.Right == nil {
		return left, nil
	}

	return c.buildLeftAssociativeArithExpr(left, *arith.Op, arith.Right)
}

func (c *Converter) convertTerm(term *Term) (ast.Expression, error) {
	left, err := c.convertFactor(term.Left)
	if err != nil {
		return nil, err
	}

	if term.Op == nil || term.Right == nil {
		return left, nil
	}

	return c.buildLeftAssociativeTerm(left, *term.Op, term.Right)
}

func (c *Converter) convertFactor(factor *Factor) (ast.Expression, error) {
	if factor.Array != nil {
		elements := []ast.Expression{}
		for _, elem := range factor.Array.Elements {
			astExpr, err := c.convertTernaryExpr(elem)
			if err != nil {
				return nil, err
			}
			elements = append(elements, astExpr)
		}
		return &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    elements,
			Raw:      "[...]",
		}, nil
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
		return buildNestedMemberExpression(factor.MemberAccess.Object, factor.MemberAccess.Properties), nil
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

	if factor.HexColor != nil {
		return &ast.Literal{
			NodeType: ast.TypeLiteral,
			Value:    *factor.HexColor,
			Raw:      fmt.Sprintf("'%s'", *factor.HexColor),
		}, nil
	}

	return nil, fmt.Errorf("empty factor")
}

func (c *Converter) convertForExprToStatement(forExpr *ForExpr) (ast.Expression, error) {
	fromExpr, err := c.convertArithExpr(forExpr.From)
	if err != nil {
		return nil, fmt.Errorf("converting for-loop from expression: %w", err)
	}

	toExpr, err := c.convertArithExpr(forExpr.To)
	if err != nil {
		return nil, fmt.Errorf("converting for-loop to expression: %w", err)
	}

	var stepExpr ast.Expression
	if forExpr.Step != nil {
		stepExpr, err = c.convertArithExpr(forExpr.Step)
		if err != nil {
			return nil, fmt.Errorf("converting for-loop step expression: %w", err)
		}
	}

	body := []ast.Node{}
	for _, stmt := range forExpr.Body {
		node, err := c.convertStatement(stmt)
		if err != nil {
			return nil, fmt.Errorf("converting for-loop body statement: %w", err)
		}
		if node != nil {
			body = append(body, node)
		}
	}

	return &ast.ForStatement{
		NodeType: ast.TypeForStatement,
		Counter:  forExpr.Counter,
		From:     fromExpr,
		To:       toExpr,
		Step:     stepExpr,
		Body:     body,
	}, nil
}

func (c *Converter) convertForInExprToStatement(forInExpr *ForInExpr) (ast.Expression, error) {
	return convertForInToAST(
		forInExpr.Vars, forInExpr.Collection, forInExpr.Body,
		c.convertArithExpr, c.convertStatement,
	)
}

func (c *Converter) convertWhileExprToStatement(whileExpr *WhileExpr) (ast.Expression, error) {
	condition, err := c.convertOrExpr(whileExpr.Condition)
	if err != nil {
		return nil, fmt.Errorf("converting while-expression condition: %w", err)
	}

	body := []ast.Node{}
	for _, stmt := range whileExpr.Body {
		node, err := c.convertStatement(stmt)
		if err != nil {
			return nil, fmt.Errorf("converting while-expression body statement: %w", err)
		}
		if node != nil {
			body = append(body, node)
		}
	}

	return &ast.WhileStatement{
		NodeType:  ast.TypeWhileStatement,
		Condition: condition,
		Body:      body,
	}, nil
}

func (c *Converter) convertIfExprToStatement(ifExpr *IfExpr) (ast.Expression, error) {
	test, err := c.convertOrExpr(ifExpr.Condition)
	if err != nil {
		return nil, fmt.Errorf("converting if-expression condition: %w", err)
	}

	body := []ast.Node{}
	for _, stmt := range ifExpr.Body {
		node, err := c.convertStatement(stmt)
		if err != nil {
			return nil, fmt.Errorf("converting if-expression body statement: %w", err)
		}
		if node != nil {
			body = append(body, node)
		}
	}

	alternate, err := c.convertIfExprElseClause(ifExpr.ElseClause)
	if err != nil {
		return nil, fmt.Errorf("converting if-expression else clause: %w", err)
	}

	return &ast.IfStatement{
		NodeType:   ast.TypeIfStatement,
		Test:       test,
		Consequent: body,
		Alternate:  alternate,
	}, nil
}

func (c *Converter) convertIfExprElseClause(ec *ElseClause) ([]ast.Node, error) {
	if ec == nil {
		return []ast.Node{}, nil
	}
	if ec.ElseIf != nil {
		node, err := c.convertIfGrammarNode(ec.ElseIf)
		if err != nil {
			return nil, err
		}
		return []ast.Node{node}, nil
	}
	nodes := []ast.Node{}
	for _, stmt := range ec.ElseBody {
		node, err := c.convertStatement(stmt)
		if err != nil {
			return nil, err
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

func (c *Converter) convertIfGrammarNode(ifGram *IfStatement) (ast.Node, error) {
	test, err := c.convertOrExpr(ifGram.Condition)
	if err != nil {
		return nil, err
	}
	body := []ast.Node{}
	for _, stmt := range ifGram.Body {
		node, err := c.convertStatement(stmt)
		if err != nil {
			return nil, err
		}
		if node != nil {
			body = append(body, node)
		}
	}
	alternate, err := c.convertIfExprElseClause(ifGram.ElseClause)
	if err != nil {
		return nil, err
	}
	return &ast.IfStatement{
		NodeType:   ast.TypeIfStatement,
		Test:       test,
		Consequent: body,
		Alternate:  alternate,
	}, nil
}

func (c *Converter) convertSwitchExprToStatement(switchExpr *SwitchExpr) (ast.Expression, error) {
	bodyResolver := NewSwitchCaseBodyResolver(c.convertStatement, c.convertExpression)
	lowering := NewSwitchLowering(c.convertOrExpr, bodyResolver)
	return lowering.Lower(switchExpr)
}

func (c *Converter) ToJSON(program *ast.Program) ([]byte, error) {
	return json.MarshalIndent(program, "", "  ")
}
