package parser

import "github.com/quant5-lab/runner/ast"

func buildIdentifier(name string) *ast.Identifier {
	return &ast.Identifier{
		NodeType: ast.TypeIdentifier,
		Name:     name,
	}
}

func buildIdentifiers(names []string) []ast.Identifier {
	identifiers := make([]ast.Identifier, len(names))
	for i, name := range names {
		identifiers[i] = ast.Identifier{
			NodeType: ast.TypeIdentifier,
			Name:     name,
		}
	}
	return identifiers
}

func buildArrayPattern(names []string) *ast.ArrayPattern {
	return &ast.ArrayPattern{
		NodeType: ast.TypeArrayPattern,
		Elements: buildIdentifiers(names),
	}
}

func buildVariableDeclaration(id ast.Pattern, init ast.Expression, kind string) *ast.VariableDeclaration {
	return &ast.VariableDeclaration{
		NodeType: ast.TypeVariableDeclaration,
		Declarations: []ast.VariableDeclarator{
			{
				NodeType: ast.TypeVariableDeclarator,
				ID:       id,
				Init:     init,
			},
		},
		Kind: kind,
	}
}
