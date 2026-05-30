package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

// inferIfStatementReturnType finds the first non-float64 branch type so that
// Pine switch expressions compiled to if-chains produce the correct IIFE return type.
func inferIfStatementReturnType(ifStmt *ast.IfStatement, typeSystem *TypeInferenceEngine) string {
	if t := inferBranchBodyType(ifStmt.Consequent, typeSystem); t != "" && t != "float64" {
		return t
	}
	return inferAlternateBranchType(ifStmt.Alternate, typeSystem)
}

func inferAlternateBranchType(alternate []ast.Node, typeSystem *TypeInferenceEngine) string {
	if len(alternate) == 0 {
		return "float64"
	}
	if nested, ok := alternate[0].(*ast.IfStatement); ok {
		return inferIfStatementReturnType(nested, typeSystem)
	}
	if t := inferBranchBodyType(alternate, typeSystem); t != "" && t != "float64" {
		return t
	}
	return "float64"
}

func inferBranchBodyType(body []ast.Node, typeSystem *TypeInferenceEngine) string {
	if len(body) == 0 {
		return "float64"
	}
	last := body[len(body)-1]
	switch n := last.(type) {
	case *ast.ExpressionStatement:
		return typeSystem.InferType(n.Expression)
	case *ast.VariableDeclaration:
		if len(n.Declarations) > 0 && n.Declarations[0].Init != nil {
			return typeSystem.InferType(n.Declarations[0].Init)
		}
	}
	return "float64"
}

// generateIfStatementReturn emits a type-aware IIFE for a Pine switch expression
// compiled to an if-statement, so the enclosing arrow function return type is satisfied.
func (a *ArrowFunctionCodegen) generateIfStatementReturn(ifStmt *ast.IfStatement) (string, error) {
	returnType := inferIfStatementReturnType(ifStmt, a.gen.typeSystem)
	cfGen := NewControlFlowExpressionGenerator(a.gen)
	iife, err := cfGen.GenerateIfExpressionAsTypedIIFE(ifStmt, returnType)
	if err != nil {
		return "", err
	}
	return a.gen.ind() + "return " + iife + "\n", nil
}
