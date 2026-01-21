package preprocessor

import (
	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/util"
)

/*
IdentifierSanitizer renames Pine identifiers conflicting with Go reserved words.

Architecture: AST transformation pass before code generation.
Responsibility: Ensure all Pine identifiers are valid Go identifiers.
Exclusions: Pine built-in bar fields (close, open, high, low, volume) are NOT renamed.
*/
type IdentifierSanitizer struct {
	renamedIdentifiers map[string]string
	pineBuiltins       map[string]bool
}

func NewIdentifierSanitizer() *IdentifierSanitizer {
	return &IdentifierSanitizer{
		renamedIdentifiers: make(map[string]string),
		pineBuiltins: map[string]bool{
			"close": true, "open": true, "high": true, "low": true, "volume": true,
			"time": true, "bar_index": true, "na": true,
		},
	}
}

/*
Transform walks AST and renames conflicting identifiers.

Strategy: Replace all Identifier.Name fields with sanitized versions.
Consistency: Same Pine name always maps to same Go name.
*/
func (s *IdentifierSanitizer) Transform(program *ast.Program) (*ast.Program, error) {
	s.walkNodes(program.Body)
	return program, nil
}

func (s *IdentifierSanitizer) walkNodes(nodes []ast.Node) {
	for _, node := range nodes {
		s.walkNode(node)
	}
}

func (s *IdentifierSanitizer) walkNode(node ast.Node) {
	switch n := node.(type) {
	case *ast.VariableDeclaration:
		s.sanitizeVariableDeclaration(n)
	case *ast.ExpressionStatement:
		s.walkExpression(n.Expression)
	case *ast.IfStatement:
		s.walkExpression(n.Test)
		s.walkNodes(n.Consequent)
		if n.Alternate != nil {
			s.walkNodes(n.Alternate)
		}
	case *ast.ForStatement:
		s.walkExpression(n.From)
		s.walkExpression(n.To)
		s.walkExpression(n.Step)
		s.walkNodes(n.Body)
	}
}

func (s *IdentifierSanitizer) sanitizeVariableDeclaration(decl *ast.VariableDeclaration) {
	for i := range decl.Declarations {
		s.sanitizeDeclarator(&decl.Declarations[i])
		if decl.Declarations[i].Init != nil {
			s.walkExpression(decl.Declarations[i].Init)
		}
	}
}

func (s *IdentifierSanitizer) sanitizeDeclarator(declarator *ast.VariableDeclarator) {
	switch id := declarator.ID.(type) {
	case *ast.Identifier:
		id.Name = s.sanitizeIdentifierName(id.Name)
	case *ast.ArrayPattern:
		for i := range id.Elements {
			id.Elements[i].Name = s.sanitizeIdentifierName(id.Elements[i].Name)
		}
	}
}

func (s *IdentifierSanitizer) walkExpression(expr ast.Expression) {
	if expr == nil {
		return
	}

	switch node := expr.(type) {
	case *ast.Identifier:
		node.Name = s.sanitizeIdentifierName(node.Name)
	case *ast.CallExpression:
		s.walkExpression(node.Callee)
		for _, arg := range node.Arguments {
			s.walkExpression(arg)
		}
	case *ast.MemberExpression:
		s.walkExpression(node.Object)
	case *ast.BinaryExpression:
		s.walkExpression(node.Left)
		s.walkExpression(node.Right)
	case *ast.UnaryExpression:
		s.walkExpression(node.Argument)
	case *ast.ConditionalExpression:
		s.walkExpression(node.Test)
		s.walkExpression(node.Consequent)
		s.walkExpression(node.Alternate)
	case *ast.LogicalExpression:
		s.walkExpression(node.Left)
		s.walkExpression(node.Right)
	case *ast.ArrowFunctionExpression:
		s.sanitizeArrowFunction(node)
	case *ast.ObjectExpression:
		for _, prop := range node.Properties {
			s.walkExpression(prop.Value)
		}
	}
}

func (s *IdentifierSanitizer) sanitizeArrowFunction(fn *ast.ArrowFunctionExpression) {
	for i := range fn.Params {
		fn.Params[i].Name = s.sanitizeIdentifierName(fn.Params[i].Name)
	}
	s.walkNodes(fn.Body)
}

func (s *IdentifierSanitizer) sanitizeIdentifierName(name string) string {
	/* Pine built-ins are never renamed */
	if s.pineBuiltins[name] {
		return name
	}

	if sanitized, exists := s.renamedIdentifiers[name]; exists {
		return sanitized
	}

	sanitized := util.SanitizeGoIdentifier(name)
	if sanitized != name {
		s.renamedIdentifiers[name] = sanitized
	}
	return sanitized
}

func (s *IdentifierSanitizer) GetRenamedIdentifiers() map[string]string {
	return s.renamedIdentifiers
}
