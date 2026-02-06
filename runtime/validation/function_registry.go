package validation

import "github.com/quant5-lab/runner/ast"

type FunctionStore interface {
	GetFunction(name string) (*ast.ArrowFunctionExpression, bool)
}

type FunctionRegistry struct {
	store map[string]*ast.ArrowFunctionExpression
}

func NewFunctionRegistry() *FunctionRegistry {
	return &FunctionRegistry{
		store: make(map[string]*ast.ArrowFunctionExpression),
	}
}

func (r *FunctionRegistry) Set(name string, fn *ast.ArrowFunctionExpression) {
	r.store[name] = fn
}

func (r *FunctionRegistry) GetFunction(name string) (*ast.ArrowFunctionExpression, bool) {
	fn, exists := r.store[name]
	return fn, exists
}

func (r *FunctionRegistry) Clear() {
	r.store = make(map[string]*ast.ArrowFunctionExpression)
}
