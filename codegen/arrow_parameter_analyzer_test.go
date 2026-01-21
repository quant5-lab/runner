package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestParameterUsageAnalyzer_AnalyzeArrowFunction validates parameter classification */
func TestParameterUsageAnalyzer_AnalyzeArrowFunction(t *testing.T) {
	tests := []struct {
		name           string
		arrowFunc      *ast.ArrowFunctionExpression
		expectedUsages map[string]ParameterUsageType
	}{
		{
			name: "two-arg TA call - first param is series",
			arrowFunc: &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{
					{Name: "src"},
					{Name: "len"},
				},
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "sma"},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "src"},
								&ast.Identifier{Name: "len"},
							},
						},
					},
				},
			},
			expectedUsages: map[string]ParameterUsageType{
				"src": ParameterUsageSeries,
				"len": ParameterUsageScalar,
			},
		},
		{
			name: "one-arg TA call - param defaults to scalar",
			arrowFunc: &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{
					{Name: "period"},
				},
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "highest"},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "period"},
							},
						},
					},
				},
			},
			expectedUsages: map[string]ParameterUsageType{
				"period": ParameterUsageScalar,
			},
		},
		{
			name: "ta.prefix function recognition",
			arrowFunc: &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{
					{Name: "source"},
					{Name: "length"},
				},
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "ta"},
								Property: &ast.Identifier{Name: "ema"},
							},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "source"},
								&ast.Identifier{Name: "length"},
							},
						},
					},
				},
			},
			expectedUsages: map[string]ParameterUsageType{
				"source": ParameterUsageSeries,
				"length": ParameterUsageScalar,
			},
		},
		{
			name: "multiple TA calls - parameter usage accumulates",
			arrowFunc: &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{
					{Name: "src"},
					{Name: "fast"},
					{Name: "slow"},
				},
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "fastMA"},
								Init: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "ema"},
									Arguments: []ast.Expression{
										&ast.Identifier{Name: "src"},
										&ast.Identifier{Name: "fast"},
									},
								},
							},
						},
					},
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "slowMA"},
								Init: &ast.CallExpression{
									Callee: &ast.Identifier{Name: "sma"},
									Arguments: []ast.Expression{
										&ast.Identifier{Name: "src"},
										&ast.Identifier{Name: "slow"},
									},
								},
							},
						},
					},
				},
			},
			expectedUsages: map[string]ParameterUsageType{
				"src":  ParameterUsageSeries,
				"fast": ParameterUsageScalar,
				"slow": ParameterUsageScalar,
			},
		},
		{
			name: "non-TA function - all params remain scalar",
			arrowFunc: &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{
					{Name: "a"},
					{Name: "b"},
				},
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "userFunc"},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "a"},
								&ast.Identifier{Name: "b"},
							},
						},
					},
				},
			},
			expectedUsages: map[string]ParameterUsageType{
				"a": ParameterUsageScalar,
				"b": ParameterUsageScalar,
			},
		},
		{
			name: "binary expression - parameters remain scalar",
			arrowFunc: &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{
					{Name: "x"},
					{Name: "y"},
				},
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Left:     &ast.Identifier{Name: "x"},
							Operator: "+",
							Right:    &ast.Identifier{Name: "y"},
						},
					},
				},
			},
			expectedUsages: map[string]ParameterUsageType{
				"x": ParameterUsageScalar,
				"y": ParameterUsageScalar,
			},
		},
		{
			name: "conditional expression with TA call",
			arrowFunc: &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{
					{Name: "src"},
					{Name: "len"},
					{Name: "threshold"},
				},
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.ConditionalExpression{
							Test: &ast.BinaryExpression{
								Left:     &ast.Identifier{Name: "threshold"},
								Operator: ">",
								Right:    &ast.Literal{Value: 0.0},
							},
							Consequent: &ast.CallExpression{
								Callee: &ast.Identifier{Name: "sma"},
								Arguments: []ast.Expression{
									&ast.Identifier{Name: "src"},
									&ast.Identifier{Name: "len"},
								},
							},
							Alternate: &ast.Literal{Value: 0.0},
						},
					},
				},
			},
			expectedUsages: map[string]ParameterUsageType{
				"src":       ParameterUsageSeries,
				"len":       ParameterUsageScalar,
				"threshold": ParameterUsageScalar,
			},
		},
		{
			name: "zero parameters",
			arrowFunc: &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{},
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.Literal{Value: 42.0},
					},
				},
			},
			expectedUsages: map[string]ParameterUsageType{},
		},
		{
			name: "for-loop with subscript access marks parameter as series",
			arrowFunc: &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{
					{Name: "src"},
					{Name: "len"},
				},
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID:   &ast.Identifier{Name: "sum"},
								Init: &ast.Literal{Value: 0.0},
							},
						},
					},
					&ast.ForStatement{
						Counter: "i",
						From:    &ast.Literal{Value: 0.0},
						To:      &ast.Identifier{Name: "len"},
						Body: []ast.Node{
							&ast.VariableDeclaration{
								Kind: "var",
								Declarations: []ast.VariableDeclarator{
									{
										ID: &ast.Identifier{Name: "sum"},
										Init: &ast.BinaryExpression{
											Left:     &ast.Identifier{Name: "sum"},
											Operator: "+",
											Right: &ast.MemberExpression{
												Object:   &ast.Identifier{Name: "src"},
												Property: &ast.Identifier{Name: "i"},
												Computed: true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedUsages: map[string]ParameterUsageType{
				"src": ParameterUsageSeries,
				"len": ParameterUsageScalar,
			},
		},
		{
			name: "parameter unused in body",
			arrowFunc: &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{
					{Name: "unused"},
				},
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.Literal{Value: 100.0},
					},
				},
			},
			expectedUsages: map[string]ParameterUsageType{
				"unused": ParameterUsageScalar,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewParameterUsageAnalyzer()
			result := analyzer.AnalyzeArrowFunction(tt.arrowFunc)

			if len(result) != len(tt.expectedUsages) {
				t.Fatalf("Usage count mismatch: got %d, want %d", len(result), len(tt.expectedUsages))
			}

			for paramName, expectedType := range tt.expectedUsages {
				actualType, exists := result[paramName]
				if !exists {
					t.Errorf("Parameter %q not found in result", paramName)
					continue
				}
				if actualType != expectedType {
					t.Errorf("Parameter %q: got %v, want %v", paramName, actualType, expectedType)
				}
			}
		})
	}
}

/* TestParameterUsageAnalyzer_TAFunctionRecognition validates TA function detection */
func TestParameterUsageAnalyzer_TAFunctionRecognition(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		isTAFunc bool
	}{
		{"sma without prefix", "sma", true},
		{"ema without prefix", "ema", true},
		{"rma without prefix", "rma", true},
		{"wma without prefix", "wma", true},
		{"stdev without prefix", "stdev", true},
		{"highest without prefix", "highest", true},
		{"lowest without prefix", "lowest", true},
		{"ta.sma with prefix", "ta.sma", true},
		{"ta.ema with prefix", "ta.ema", true},
		{"ta.rma with prefix", "ta.rma", true},
		{"ta.wma with prefix", "ta.wma", true},
		{"ta.stdev with prefix", "ta.stdev", true},
		{"ta.highest with prefix", "ta.highest", true},
		{"ta.lowest with prefix", "ta.lowest", true},
		{"user function", "myFunc", false},
		{"plot function", "plot", false},
		{"strategy.entry", "strategy.entry", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTAIndicatorFunction(tt.funcName)
			if result != tt.isTAFunc {
				t.Errorf("isTAIndicatorFunction(%q) = %v, want %v", tt.funcName, result, tt.isTAFunc)
			}
		})
	}
}

/* TestParameterUsageAnalyzer_NestedExpressions validates recursive analysis */
func TestParameterUsageAnalyzer_NestedExpressions(t *testing.T) {
	arrowFunc := &ast.ArrowFunctionExpression{
		Params: []ast.Identifier{
			{Name: "src"},
			{Name: "len"},
		},
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.BinaryExpression{
					Left: &ast.CallExpression{
						Callee: &ast.Identifier{Name: "sma"},
						Arguments: []ast.Expression{
							&ast.Identifier{Name: "src"},
							&ast.Identifier{Name: "len"},
						},
					},
					Operator: "+",
					Right: &ast.CallExpression{
						Callee: &ast.Identifier{Name: "ema"},
						Arguments: []ast.Expression{
							&ast.Identifier{Name: "src"},
							&ast.Identifier{Name: "len"},
						},
					},
				},
			},
		},
	}

	analyzer := NewParameterUsageAnalyzer()
	result := analyzer.AnalyzeArrowFunction(arrowFunc)

	if result["src"] != ParameterUsageSeries {
		t.Errorf("src should be series, got %v", result["src"])
	}
	if result["len"] != ParameterUsageScalar {
		t.Errorf("len should be scalar, got %v", result["len"])
	}
}

/* TestParameterUsageAnalyzer_UnaryExpression validates unary operator handling */
func TestParameterUsageAnalyzer_UnaryExpression(t *testing.T) {
	arrowFunc := &ast.ArrowFunctionExpression{
		Params: []ast.Identifier{
			{Name: "src"},
			{Name: "len"},
		},
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.UnaryExpression{
					Operator: "-",
					Argument: &ast.CallExpression{
						Callee: &ast.Identifier{Name: "sma"},
						Arguments: []ast.Expression{
							&ast.Identifier{Name: "src"},
							&ast.Identifier{Name: "len"},
						},
					},
				},
			},
		},
	}

	analyzer := NewParameterUsageAnalyzer()
	result := analyzer.AnalyzeArrowFunction(arrowFunc)

	if result["src"] != ParameterUsageSeries {
		t.Errorf("src should be series in unary expression, got %v", result["src"])
	}
	if result["len"] != ParameterUsageScalar {
		t.Errorf("len should be scalar in unary expression, got %v", result["len"])
	}
}

/* TestParameterUsageAnalyzer_ArrayLiteral validates array element analysis */
func TestParameterUsageAnalyzer_ArrayLiteral(t *testing.T) {
	arrowFunc := &ast.ArrowFunctionExpression{
		Params: []ast.Identifier{
			{Name: "src"},
			{Name: "len"},
		},
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.Literal{
					Value: []ast.Expression{
						&ast.CallExpression{
							Callee: &ast.Identifier{Name: "sma"},
							Arguments: []ast.Expression{
								&ast.Identifier{Name: "src"},
								&ast.Identifier{Name: "len"},
							},
						},
						&ast.Literal{Value: 0.0},
					},
				},
			},
		},
	}

	analyzer := NewParameterUsageAnalyzer()
	result := analyzer.AnalyzeArrowFunction(arrowFunc)

	if result["src"] != ParameterUsageSeries {
		t.Errorf("src should be series in array literal, got %v", result["src"])
	}
}

/* TestParameterUsageAnalyzer_EdgeCases validates boundary conditions */
func TestParameterUsageAnalyzer_EdgeCases(t *testing.T) {
	t.Run("nil arrow function", func(t *testing.T) {
		analyzer := NewParameterUsageAnalyzer()

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for nil arrow function")
			}
		}()

		analyzer.AnalyzeArrowFunction(nil)
	})

	t.Run("empty body", func(t *testing.T) {
		arrowFunc := &ast.ArrowFunctionExpression{
			Params: []ast.Identifier{
				{Name: "param"},
			},
			Body: []ast.Node{},
		}

		analyzer := NewParameterUsageAnalyzer()
		result := analyzer.AnalyzeArrowFunction(arrowFunc)

		if result["param"] != ParameterUsageScalar {
			t.Errorf("Parameter with empty body should default to scalar, got %v", result["param"])
		}
	})

	t.Run("TA call with non-identifier first argument", func(t *testing.T) {
		arrowFunc := &ast.ArrowFunctionExpression{
			Params: []ast.Identifier{
				{Name: "len"},
			},
			Body: []ast.Node{
				&ast.ExpressionStatement{
					Expression: &ast.CallExpression{
						Callee: &ast.Identifier{Name: "sma"},
						Arguments: []ast.Expression{
							&ast.Literal{Value: 42.0},
							&ast.Identifier{Name: "len"},
						},
					},
				},
			},
		}

		analyzer := NewParameterUsageAnalyzer()
		result := analyzer.AnalyzeArrowFunction(arrowFunc)

		if result["len"] != ParameterUsageScalar {
			t.Errorf("len should remain scalar, got %v", result["len"])
		}
	})

	t.Run("parameter name with special characters", func(t *testing.T) {
		arrowFunc := &ast.ArrowFunctionExpression{
			Params: []ast.Identifier{
				{Name: "_src_123"},
				{Name: "len"},
			},
			Body: []ast.Node{
				&ast.ExpressionStatement{
					Expression: &ast.CallExpression{
						Callee: &ast.Identifier{Name: "sma"},
						Arguments: []ast.Expression{
							&ast.Identifier{Name: "_src_123"},
							&ast.Identifier{Name: "len"},
						},
					},
				},
			},
		}

		analyzer := NewParameterUsageAnalyzer()
		result := analyzer.AnalyzeArrowFunction(arrowFunc)

		if result["_src_123"] != ParameterUsageSeries {
			t.Errorf("_src_123 should be series, got %v", result["_src_123"])
		}
	})
}

/* TestParameterUsageAnalyzer_Idempotency validates consistent analysis */
func TestParameterUsageAnalyzer_Idempotency(t *testing.T) {
	arrowFunc := &ast.ArrowFunctionExpression{
		Params: []ast.Identifier{
			{Name: "src"},
			{Name: "len"},
		},
		Body: []ast.Node{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Callee: &ast.Identifier{Name: "sma"},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "src"},
						&ast.Identifier{Name: "len"},
					},
				},
			},
		},
	}

	analyzer1 := NewParameterUsageAnalyzer()
	result1 := analyzer1.AnalyzeArrowFunction(arrowFunc)

	analyzer2 := NewParameterUsageAnalyzer()
	result2 := analyzer2.AnalyzeArrowFunction(arrowFunc)

	if len(result1) != len(result2) {
		t.Fatalf("Result count differs between runs: %d vs %d", len(result1), len(result2))
	}

	for param, usage1 := range result1 {
		usage2, exists := result2[param]
		if !exists {
			t.Errorf("Parameter %q missing in second run", param)
			continue
		}
		if usage1 != usage2 {
			t.Errorf("Parameter %q usage differs: %v vs %v", param, usage1, usage2)
		}
	}
}
