package codegen

import "github.com/quant5-lab/runner/ast"

type NestedVariableScanner struct {
	registrar *VariableDeclarationRegistrar
	gen       *generator
}

func NewNestedVariableScanner(gen *generator) *NestedVariableScanner {
	return &NestedVariableScanner{
		registrar: NewVariableDeclarationRegistrar(gen),
		gen:       gen,
	}
}

func (s *NestedVariableScanner) ScanIfBlock(ifStmt *ast.IfStatement) {
	for _, node := range ifStmt.Consequent {
		s.scanStatement(node)
	}
	for _, node := range ifStmt.Alternate {
		s.scanStatement(node)
	}
}

func (s *NestedVariableScanner) scanStatement(stmt ast.Node) {
	switch n := stmt.(type) {
	case *ast.IfStatement:
		s.ScanIfBlock(n)
	case *ast.VariableDeclaration:
		s.registrar.RegisterDeclaration(n)
	}
}

func (s *NestedVariableScanner) ScanReassignments(stmt ast.Node) {
	switch n := stmt.(type) {
	case *ast.IfStatement:
		for _, node := range n.Consequent {
			s.ScanReassignments(node)
		}
		for _, node := range n.Alternate {
			s.ScanReassignments(node)
		}

	case *ast.ForStatement:
		for _, node := range n.Body {
			s.ScanReassignments(node)
		}

	case *ast.ForInStatement:
		for _, node := range n.Body {
			s.ScanReassignments(node)
		}

	case *ast.WhileStatement:
		for _, node := range n.Body {
			s.ScanReassignments(node)
		}

	case *ast.VariableDeclaration:
		if n.Kind == "var" {
			for _, declarator := range n.Declarations {
				if id, ok := declarator.ID.(*ast.Identifier); ok {
					s.gen.reassignedVars[id.Name] = true
				}
			}
		}
	}
}
