package codegen

import "github.com/quant5-lab/runner/ast"

type VariableDeclarationRegistrar struct {
	gen *generator
}

func NewVariableDeclarationRegistrar(gen *generator) *VariableDeclarationRegistrar {
	return &VariableDeclarationRegistrar{gen: gen}
}

func (r *VariableDeclarationRegistrar) RegisterDeclaration(decl *ast.VariableDeclaration) {
	for _, declarator := range decl.Declarations {
		r.RegisterDeclarator(declarator)
	}
}

func (r *VariableDeclarationRegistrar) RegisterDeclarator(d ast.VariableDeclarator) {
	if r.registerArrayPattern(d) {
		return
	}

	id, ok := d.ID.(*ast.Identifier)
	if !ok {
		return
	}

	if _, ok := d.Init.(*ast.ArrowFunctionExpression); ok {
		return
	}

	varName := SanitizeGoIdentifier(id.Name)
	r.RegisterVariable(varName, d.Init)
}

func (r *VariableDeclarationRegistrar) RegisterVariable(varName string, init ast.Expression) {
	r.gen.scanForSubscriptedCalls(init)

	if callExpr, ok := init.(*ast.CallExpression); ok {
		r.gen.collectNestedVariables(varName, callExpr)
	}

	if r.gen.constantRegistry.IsConstant(varName) {
		return
	}

	varType := r.gen.inferVariableType(init)
	r.gen.variables[varName] = varType
	r.gen.typeSystem.RegisterVariable(varName, varType)

	r.registerStringConstant(varName, varType, init)
}

func (r *VariableDeclarationRegistrar) registerArrayPattern(d ast.VariableDeclarator) bool {
	arrayPattern, ok := d.ID.(*ast.ArrayPattern)
	if !ok {
		return false
	}
	for _, elem := range arrayPattern.Elements {
		varName := SanitizeGoIdentifier(elem.Name)
		varType := r.gen.inferVariableType(d.Init)
		r.gen.variables[varName] = varType
		r.gen.typeSystem.RegisterVariable(varName, varType)
	}
	return true
}

func (r *VariableDeclarationRegistrar) registerStringConstant(varName, varType string, init ast.Expression) {
	if varType != "string" {
		return
	}
	lit, ok := init.(*ast.Literal)
	if !ok {
		return
	}
	strVal, ok := lit.Value.(string)
	if !ok {
		return
	}
	r.gen.constants[varName] = strVal
	r.gen.constantRegistry.Register(varName, strVal)
}
