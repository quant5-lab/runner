package validation

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestUserFunctionEvaluator_ParameterlessFunctions(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		funcBody ast.Expression
		expected float64
	}{
		{
			name:     "constant_literal",
			funcName: "getPeriod",
			funcBody: &ast.Literal{Value: 14},
			expected: 14.0,
		},
		{
			name:     "constant_float",
			funcName: "getMultiplier",
			funcBody: &ast.Literal{Value: 2.5},
			expected: 2.5,
		},
		{
			name:     "negative_value",
			funcName: "getOffset",
			funcBody: &ast.Literal{Value: -5},
			expected: -5.0,
		},
		{
			name:     "zero",
			funcName: "getZero",
			funcBody: &ast.Literal{Value: 0},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constReg := NewConstantRegistry()
			funcReg := NewFunctionRegistry()

			funcReg.Set(tt.funcName, &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{},
				Body: []ast.Node{
					&ast.ExpressionStatement{Expression: tt.funcBody},
				},
			})

			eval := NewExpressionEvaluatorWithFunctions(constReg, funcReg)

			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: tt.funcName},
				Arguments: []ast.Expression{},
			}

			result := eval.Evaluate(call)

			if math.IsNaN(result) {
				t.Errorf("expected %.2f, got NaN", tt.expected)
				return
			}
			if math.Abs(result-tt.expected) > 0.0001 {
				t.Errorf("expected %.4f, got %.4f", tt.expected, result)
			}
		})
	}
}

func TestUserFunctionEvaluator_ExpressionBody(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		funcBody ast.Expression
		expected float64
	}{
		{
			name:     "binary_multiplication",
			funcName: "getPeriod",
			funcBody: &ast.BinaryExpression{
				Operator: "*",
				Left:     &ast.Literal{Value: 7},
				Right:    &ast.Literal{Value: 2},
			},
			expected: 14.0,
		},
		{
			name:     "binary_addition",
			funcName: "getSum",
			funcBody: &ast.BinaryExpression{
				Operator: "+",
				Left:     &ast.Literal{Value: 10},
				Right:    &ast.Literal{Value: 4},
			},
			expected: 14.0,
		},
		{
			name:     "complex_expression",
			funcName: "getComplex",
			funcBody: &ast.BinaryExpression{
				Operator: "+",
				Left: &ast.BinaryExpression{
					Operator: "*",
					Left:     &ast.Literal{Value: 3},
					Right:    &ast.Literal{Value: 4},
				},
				Right: &ast.Literal{Value: 2},
			},
			expected: 14.0,
		},
		{
			name:     "unary_negation",
			funcName: "getNegative",
			funcBody: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.Literal{Value: 14},
			},
			expected: -14.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constReg := NewConstantRegistry()
			funcReg := NewFunctionRegistry()

			funcReg.Set(tt.funcName, &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{},
				Body: []ast.Node{
					&ast.ExpressionStatement{Expression: tt.funcBody},
				},
			})

			eval := NewExpressionEvaluatorWithFunctions(constReg, funcReg)

			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: tt.funcName},
				Arguments: []ast.Expression{},
			}

			result := eval.Evaluate(call)

			if math.IsNaN(result) {
				t.Errorf("expected %.2f, got NaN", tt.expected)
				return
			}
			if math.Abs(result-tt.expected) > 0.0001 {
				t.Errorf("expected %.4f, got %.4f", tt.expected, result)
			}
		})
	}
}

func TestUserFunctionEvaluator_ChainedFunctions(t *testing.T) {
	constReg := NewConstantRegistry()
	funcReg := NewFunctionRegistry()

	// getBase() => 7
	funcReg.Set("getBase", &ast.ArrowFunctionExpression{
		Params: []ast.Identifier{},
		Body: []ast.Node{
			&ast.ExpressionStatement{Expression: &ast.Literal{Value: 7}},
		},
	})

	// getPeriod() => getBase() * 2
	funcReg.Set("getPeriod", &ast.ArrowFunctionExpression{
		Params: []ast.Identifier{},
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.BinaryExpression{
					Operator: "*",
					Left: &ast.CallExpression{
						Callee:    &ast.Identifier{Name: "getBase"},
						Arguments: []ast.Expression{},
					},
					Right: &ast.Literal{Value: 2},
				},
			},
		},
	})

	eval := NewExpressionEvaluatorWithFunctions(constReg, funcReg)

	call := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "getPeriod"},
		Arguments: []ast.Expression{},
	}

	result := eval.Evaluate(call)

	if math.IsNaN(result) {
		t.Error("expected 14.0, got NaN")
		return
	}
	if math.Abs(result-14.0) > 0.0001 {
		t.Errorf("expected 14.0, got %.4f", result)
	}
}

func TestUserFunctionEvaluator_FunctionsWithParameters(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		params   []string
		body     ast.Expression
		args     []ast.Expression
		expected float64
	}{
		{
			name:     "identity_function",
			funcName: "identity",
			params:   []string{"x"},
			body:     &ast.Identifier{Name: "x"},
			args:     []ast.Expression{&ast.Literal{Value: 14}},
			expected: 14.0,
		},
		{
			name:     "add_one",
			funcName: "addOne",
			params:   []string{"x"},
			body: &ast.BinaryExpression{
				Operator: "+",
				Left:     &ast.Identifier{Name: "x"},
				Right:    &ast.Literal{Value: 1},
			},
			args:     []ast.Expression{&ast.Literal{Value: 13}},
			expected: 14.0,
		},
		{
			name:     "multiply",
			funcName: "multiply",
			params:   []string{"a", "b"},
			body: &ast.BinaryExpression{
				Operator: "*",
				Left:     &ast.Identifier{Name: "a"},
				Right:    &ast.Identifier{Name: "b"},
			},
			args:     []ast.Expression{&ast.Literal{Value: 7}, &ast.Literal{Value: 2}},
			expected: 14.0,
		},
		{
			name:     "scale",
			funcName: "scale",
			params:   []string{"x"},
			body: &ast.BinaryExpression{
				Operator: "*",
				Left:     &ast.Identifier{Name: "x"},
				Right:    &ast.Literal{Value: 2},
			},
			args:     []ast.Expression{&ast.Literal{Value: 7}},
			expected: 14.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constReg := NewConstantRegistry()
			funcReg := NewFunctionRegistry()

			params := make([]ast.Identifier, len(tt.params))
			for i, p := range tt.params {
				params[i] = ast.Identifier{Name: p}
			}

			funcReg.Set(tt.funcName, &ast.ArrowFunctionExpression{
				Params: params,
				Body: []ast.Node{
					&ast.ExpressionStatement{Expression: tt.body},
				},
			})

			eval := NewExpressionEvaluatorWithFunctions(constReg, funcReg)

			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: tt.funcName},
				Arguments: tt.args,
			}

			result := eval.Evaluate(call)

			if math.IsNaN(result) {
				t.Errorf("expected %.2f, got NaN", tt.expected)
				return
			}
			if math.Abs(result-tt.expected) > 0.0001 {
				t.Errorf("expected %.4f, got %.4f", tt.expected, result)
			}
		})
	}
}

func TestUserFunctionEvaluator_ParameterScopeCleanup(t *testing.T) {
	constReg := NewConstantRegistry()
	funcReg := NewFunctionRegistry()

	constReg.Set("x", 100.0)

	funcReg.Set("addOne", &ast.ArrowFunctionExpression{
		Params: []ast.Identifier{{Name: "x"}},
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.BinaryExpression{
					Operator: "+",
					Left:     &ast.Identifier{Name: "x"},
					Right:    &ast.Literal{Value: 1},
				},
			},
		},
	})

	eval := NewExpressionEvaluatorWithFunctions(constReg, funcReg)

	call := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "addOne"},
		Arguments: []ast.Expression{&ast.Literal{Value: 13}},
	}
	result := eval.Evaluate(call)
	if math.Abs(result-14.0) > 0.0001 {
		t.Errorf("expected 14.0 from addOne(13), got %.4f", result)
	}

	resolvedX := eval.Evaluate(&ast.Identifier{Name: "x"})
	if math.Abs(resolvedX-100.0) > 0.0001 {
		t.Errorf("x should still be 100.0 after function call, got %.4f", resolvedX)
	}
}

func TestUserFunctionEvaluator_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*FunctionRegistry)
		call      *ast.CallExpression
	}{
		{
			name:      "unknown_function",
			setupFunc: func(fr *FunctionRegistry) {},
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "unknown"},
				Arguments: []ast.Expression{},
			},
		},
		{
			name: "wrong_argument_count",
			setupFunc: func(fr *FunctionRegistry) {
				fr.Set("addOne", &ast.ArrowFunctionExpression{
					Params: []ast.Identifier{{Name: "x"}},
					Body: []ast.Node{
						&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "x"}},
					},
				})
			},
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "addOne"},
				Arguments: []ast.Expression{},
			},
		},
		{
			name: "non_constant_argument",
			setupFunc: func(fr *FunctionRegistry) {
				fr.Set("identity", &ast.ArrowFunctionExpression{
					Params: []ast.Identifier{{Name: "x"}},
					Body: []ast.Node{
						&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "x"}},
					},
				})
			},
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "identity"},
				Arguments: []ast.Expression{&ast.Identifier{Name: "unknownVar"}},
			},
		},
		{
			name: "empty_function_body",
			setupFunc: func(fr *FunctionRegistry) {
				fr.Set("empty", &ast.ArrowFunctionExpression{
					Params: []ast.Identifier{},
					Body:   []ast.Node{},
				})
			},
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "empty"},
				Arguments: []ast.Expression{},
			},
		},
		{
			name:      "nil_call",
			setupFunc: func(fr *FunctionRegistry) {},
			call:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constReg := NewConstantRegistry()
			funcReg := NewFunctionRegistry()
			tt.setupFunc(funcReg)

			eval := NewExpressionEvaluatorWithFunctions(constReg, funcReg)

			result := eval.Evaluate(tt.call)

			if !math.IsNaN(result) {
				t.Errorf("expected NaN for error case, got %.4f", result)
			}
		})
	}
}

func TestUserFunctionEvaluator_RecursionLimit(t *testing.T) {
	constReg := NewConstantRegistry()
	funcReg := NewFunctionRegistry()

	funcReg.Set("infinite", &ast.ArrowFunctionExpression{
		Params: []ast.Identifier{},
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee:    &ast.Identifier{Name: "infinite"},
					Arguments: []ast.Expression{},
				},
			},
		},
	})

	eval := NewExpressionEvaluatorWithFunctions(constReg, funcReg)

	call := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "infinite"},
		Arguments: []ast.Expression{},
	}

	result := eval.Evaluate(call)

	if !math.IsNaN(result) {
		t.Errorf("expected NaN for infinite recursion, got %.4f", result)
	}
}

func TestUserFunctionEvaluator_MathFunctionFallback(t *testing.T) {
	constReg := NewConstantRegistry()
	funcReg := NewFunctionRegistry()

	eval := NewExpressionEvaluatorWithFunctions(constReg, funcReg)

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "math"},
			Property: &ast.Identifier{Name: "ceil"},
		},
		Arguments: []ast.Expression{&ast.Literal{Value: 13.5}},
	}

	result := eval.Evaluate(call)

	if math.IsNaN(result) {
		t.Error("expected 14.0, got NaN")
		return
	}
	if math.Abs(result-14.0) > 0.0001 {
		t.Errorf("expected 14.0, got %.4f", result)
	}
}

func TestUserFunctionEvaluator_FunctionUsingConstants(t *testing.T) {
	constReg := NewConstantRegistry()
	funcReg := NewFunctionRegistry()

	constReg.Set("baseValue", 7.0)

	funcReg.Set("getPeriod", &ast.ArrowFunctionExpression{
		Params: []ast.Identifier{},
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.BinaryExpression{
					Operator: "*",
					Left:     &ast.Identifier{Name: "baseValue"},
					Right:    &ast.Literal{Value: 2},
				},
			},
		},
	})

	eval := NewExpressionEvaluatorWithFunctions(constReg, funcReg)

	call := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "getPeriod"},
		Arguments: []ast.Expression{},
	}

	result := eval.Evaluate(call)

	if math.IsNaN(result) {
		t.Error("expected 14.0, got NaN")
		return
	}
	if math.Abs(result-14.0) > 0.0001 {
		t.Errorf("expected 14.0, got %.4f", result)
	}
}

func TestFunctionRegistry_Operations(t *testing.T) {
	registry := NewFunctionRegistry()

	arrowFunc := &ast.ArrowFunctionExpression{
		Params: []ast.Identifier{},
		Body: []ast.Node{
			&ast.ExpressionStatement{Expression: &ast.Literal{Value: 14}},
		},
	}

	t.Run("set_and_get", func(t *testing.T) {
		registry.Set("getPeriod", arrowFunc)

		fn, exists := registry.GetFunction("getPeriod")
		if !exists {
			t.Error("expected function to exist")
			return
		}
		if fn != arrowFunc {
			t.Error("retrieved function doesn't match")
		}
	})

	t.Run("get_nonexistent", func(t *testing.T) {
		_, exists := registry.GetFunction("nonexistent")
		if exists {
			t.Error("expected function to not exist")
		}
	})

	t.Run("clear", func(t *testing.T) {
		registry.Set("test", arrowFunc)
		registry.Clear()

		_, exists := registry.GetFunction("test")
		if exists {
			t.Error("expected registry to be empty after Clear()")
		}
	})
}
