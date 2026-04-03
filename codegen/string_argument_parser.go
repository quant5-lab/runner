package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type StringArgumentParser struct {
	gen *generator
}

func NewStringArgumentParser(g *generator) *StringArgumentParser {
	return &StringArgumentParser{gen: g}
}

func (p *StringArgumentParser) ParseArguments(call *ast.CallExpression) ([]string, error) {
	var parsedArgs []string

	for _, arg := range call.Arguments {
		argCode, err := p.parseArgument(arg)
		if err != nil {
			return nil, err
		}
		parsedArgs = append(parsedArgs, argCode)
	}

	return parsedArgs, nil
}

func (p *StringArgumentParser) parseArgument(expr ast.Expression) (string, error) {
	switch e := expr.(type) {
	case *ast.Literal:
		return p.parseLiteral(e)
	case *ast.Identifier:
		return p.parseIdentifier(e)
	case *ast.CallExpression:
		return p.gen.generateCallExpression(e)
	case *ast.BinaryExpression:
		return p.gen.generateBinaryExpression(e)
	case *ast.MemberExpression:
		return p.gen.generateMemberExpression(e)
	case *ast.ConditionalExpression:
		return p.gen.generateConditionExpression(e)
	default:
		return "", fmt.Errorf("unsupported argument type: %T", expr)
	}
}

func (p *StringArgumentParser) parseLiteral(lit *ast.Literal) (string, error) {
	switch v := lit.Value.(type) {
	case string:
		return fmt.Sprintf("%q", v), nil
	case float64:
		return fmt.Sprintf("%v", v), nil
	case int:
		return fmt.Sprintf("%d", v), nil
	case int64:
		return fmt.Sprintf("%d", v), nil
	case bool:
		return fmt.Sprintf("%t", v), nil
	default:
		return "", fmt.Errorf("unsupported literal type: %T", v)
	}
}

func (p *StringArgumentParser) parseIdentifier(id *ast.Identifier) (string, error) {
	if p.gen.variables[id.Name] != "" {
		return id.Name, nil
	}
	if _, exists := p.gen.constants[id.Name]; exists {
		return id.Name, nil
	}
	return id.Name, nil
}
