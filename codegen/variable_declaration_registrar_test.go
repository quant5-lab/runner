package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/validation"
)

/* newDeclarationTestGenerator creates a minimal generator for declaration-related unit tests */
func newDeclarationTestGenerator() *generator {
	constantRegistry := NewConstantRegistry()
	typeSystem := NewTypeInferenceEngine()

	return &generator{
		variables:         make(map[string]string),
		varInits:          make(map[string]ast.Expression),
		constants:         make(map[string]interface{}),
		reassignedVars:    make(map[string]bool),
		constantRegistry:  constantRegistry,
		typeSystem:        typeSystem,
		runtimeOnlyFilter: NewRuntimeOnlyFunctionFilter(),
		constEvaluator:    validation.NewWarmupAnalyzer(),
	}
}

/* TestVariableDeclarationRegistrar validates registration across all AST init expression types */
func TestVariableDeclarationRegistrar(t *testing.T) {
	tests := []struct {
		name         string
		declarator   ast.VariableDeclarator
		wantVarName  string
		wantType     string
		wantSkipped  bool
		wantConstant interface{}
		description  string
	}{
		{
			name: "float literal",
			declarator: ast.VariableDeclarator{
				ID:   &ast.Identifier{Name: "price"},
				Init: &ast.Literal{Value: 42.0},
			},
			wantVarName: "price",
			wantType:    "float64",
			description: "numeric float literal infers float64",
		},
		{
			name: "integer literal",
			declarator: ast.VariableDeclarator{
				ID:   &ast.Identifier{Name: "count"},
				Init: &ast.Literal{Value: 10},
			},
			wantVarName: "count",
			wantType:    "float64",
			description: "integer literal infers float64 (PineScript numeric unification)",
		},
		{
			name: "boolean literal",
			declarator: ast.VariableDeclarator{
				ID:   &ast.Identifier{Name: "flag"},
				Init: &ast.Literal{Value: true},
			},
			wantVarName: "flag",
			wantType:    "bool",
			description: "boolean literal infers bool type",
		},
		{
			name: "string literal registers constant",
			declarator: ast.VariableDeclarator{
				ID:   &ast.Identifier{Name: "ticker"},
				Init: &ast.Literal{Value: "AAPL"},
			},
			wantVarName:  "ticker",
			wantType:     "string",
			wantConstant: "AAPL",
			description:  "string literal registered in both variables and constants maps",
		},
		{
			name: "identifier init",
			declarator: ast.VariableDeclarator{
				ID:   &ast.Identifier{Name: "src"},
				Init: &ast.Identifier{Name: "close"},
			},
			wantVarName: "src",
			wantType:    "float64",
			description: "identifier init infers float64 for known series names",
		},
		{
			name: "binary expression",
			declarator: ast.VariableDeclarator{
				ID: &ast.Identifier{Name: "spread"},
				Init: &ast.BinaryExpression{
					Operator: "-",
					Left:     &ast.Identifier{Name: "high"},
					Right:    &ast.Identifier{Name: "low"},
				},
			},
			wantVarName: "spread",
			wantType:    "float64",
			description: "binary arithmetic expression infers float64",
		},
		{
			name: "conditional expression",
			declarator: ast.VariableDeclarator{
				ID: &ast.Identifier{Name: "val"},
				Init: &ast.ConditionalExpression{
					Test:       &ast.Literal{Value: true},
					Consequent: &ast.Literal{Value: 1.0},
					Alternate:  &ast.Literal{Value: 0.0},
				},
			},
			wantVarName: "val",
			wantType:    "float64",
			description: "ternary expression infers type from consequent branch",
		},
		{
			name: "call expression",
			declarator: ast.VariableDeclarator{
				ID: &ast.Identifier{Name: "avg"},
				Init: &ast.CallExpression{
					Callee:    &ast.Identifier{Name: "sma"},
					Arguments: []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: 20}},
				},
			},
			wantVarName: "avg",
			wantType:    "float64",
			description: "function call expression infers float64 return type",
		},
		{
			name: "arrow function skipped",
			declarator: ast.VariableDeclarator{
				ID: &ast.Identifier{Name: "myFunc"},
				Init: &ast.ArrowFunctionExpression{
					Params: []ast.Identifier{{Name: "x"}},
					Body:   []ast.Node{},
				},
			},
			wantSkipped: true,
			description: "arrow functions are UDFs, not variables",
		},
		{
			name: "pre-registered constant skipped",
			declarator: ast.VariableDeclarator{
				ID:   &ast.Identifier{Name: "length"},
				Init: &ast.Literal{Value: 14.0},
			},
			wantSkipped: true,
			description: "constants already in registry not re-registered as variables",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newDeclarationTestGenerator()
			registrar := NewVariableDeclarationRegistrar(gen)

			if tt.name == "pre-registered constant skipped" {
				gen.constantRegistry.Register("length", 14.0)
			}

			registrar.RegisterDeclarator(tt.declarator)

			if tt.wantSkipped {
				varName := ""
				if id, ok := tt.declarator.ID.(*ast.Identifier); ok {
					varName = id.Name
				}
				if _, ok := gen.variables[varName]; ok {
					t.Errorf("expected %q skipped, found in variables (%s)", varName, tt.description)
				}
				return
			}

			varType, ok := gen.variables[tt.wantVarName]
			if !ok {
				t.Fatalf("expected %q in variables (%s)", tt.wantVarName, tt.description)
			}
			if varType != tt.wantType {
				t.Errorf("type mismatch for %q: want %q, got %q (%s)", tt.wantVarName, tt.wantType, varType, tt.description)
			}

			if tt.wantConstant != nil {
				if gen.constants[tt.wantVarName] != tt.wantConstant {
					t.Errorf("constant mismatch for %q: want %v, got %v (%s)",
						tt.wantVarName, tt.wantConstant, gen.constants[tt.wantVarName], tt.description)
				}
			}
		})
	}
}

/* TestVariableDeclarationRegistrar_ArrayPattern validates tuple destructuring registration */
func TestVariableDeclarationRegistrar_ArrayPattern(t *testing.T) {
	gen := newDeclarationTestGenerator()
	registrar := NewVariableDeclarationRegistrar(gen)

	registrar.RegisterDeclarator(ast.VariableDeclarator{
		ID: &ast.ArrayPattern{
			Elements: []ast.Identifier{
				{Name: "upper"},
				{Name: "middle"},
				{Name: "lower"},
			},
		},
		Init: &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "bb"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: 20}},
		},
	})

	for _, name := range []string{"upper", "middle", "lower"} {
		if _, ok := gen.variables[name]; !ok {
			t.Errorf("expected %q in variables from array pattern destructuring", name)
		}
	}
}

/* TestVariableDeclarationRegistrar_BatchDeclaration validates RegisterDeclaration processes all declarators */
func TestVariableDeclarationRegistrar_BatchDeclaration(t *testing.T) {
	gen := newDeclarationTestGenerator()
	registrar := NewVariableDeclarationRegistrar(gen)

	registrar.RegisterDeclaration(&ast.VariableDeclaration{
		Declarations: []ast.VariableDeclarator{
			{ID: &ast.Identifier{Name: "a"}, Init: &ast.Literal{Value: 1.0}},
			{ID: &ast.Identifier{Name: "b"}, Init: &ast.Literal{Value: true}},
			{ID: &ast.Identifier{Name: "c"}, Init: &ast.Literal{Value: "text"}},
		},
	})

	expected := map[string]string{"a": "float64", "b": "bool", "c": "string"}
	for name, wantType := range expected {
		gotType, ok := gen.variables[name]
		if !ok {
			t.Errorf("expected %q in variables map", name)
			continue
		}
		if gotType != wantType {
			t.Errorf("variable %q: expected type %q, got %q", name, wantType, gotType)
		}
	}
}
