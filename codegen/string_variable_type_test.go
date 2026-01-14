package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* Validates string type detection for strategy constants */
func TestTypeInference_StringConstants(t *testing.T) {
	tests := []struct {
		name         string
		expr         ast.Expression
		expectedType string
		description  string
	}{
		{
			name: "strategy.long is string",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "long"},
			},
			expectedType: "string",
			description:  "strategy.long constant returns string type",
		},
		{
			name: "strategy.short is string",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "short"},
			},
			expectedType: "string",
			description:  "strategy.short constant returns string type",
		},
		{
			name: "syminfo.tickerid is string",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "syminfo"},
				Property: &ast.Identifier{Name: "tickerid"},
			},
			expectedType: "string",
			description:  "syminfo.tickerid returns string type",
		},
		{
			name: "ternary with strategy constants",
			expr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Operator: ">",
					Left:     &ast.Identifier{Name: "close"},
					Right:    &ast.Identifier{Name: "open"},
				},
				Consequent: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "long"},
				},
				Alternate: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "short"},
				},
			},
			expectedType: "string",
			description:  "Ternary inherits string type from consequent",
		},
		{
			name: "strategy.position_size is float64",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "position_size"},
			},
			expectedType: "float64",
			description:  "Non-direction strategy members are float64",
		},
		{
			name: "random member expression defaults to float64",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "custom"},
				Property: &ast.Identifier{Name: "value"},
			},
			expectedType: "float64",
			description:  "Unknown member expressions default to float64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewTypeInferenceEngine()
			result := engine.InferType(tt.expr)
			if result != tt.expectedType {
				t.Errorf("%s\nexpected type: %s\ngot type:      %s",
					tt.description, tt.expectedType, result)
			}
		})
	}
}

/* Validates string variables generate scalar declarations not Series */
func TestStringVariableCodeGeneration(t *testing.T) {
	tests := []struct {
		name            string
		program         *ast.Program
		mustHaveDecl    string
		mustNotHaveDecl string
		mustHaveInit    string
		mustNotHaveInit string
		mustHaveUnused  string
		mustNotHaveNext string
		description     string
	}{
		{
			name: "string variable with strategy.long",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "direction"},
								Init: &ast.MemberExpression{
									Object:   &ast.Identifier{Name: "strategy"},
									Property: &ast.Identifier{Name: "long"},
								},
							},
						},
					},
				},
			},
			mustHaveDecl:    "var direction string",
			mustNotHaveDecl: "var directionSeries",
			mustHaveInit:    "direction = strategy.Long",
			mustNotHaveInit: "directionSeries = series.NewSeries",
			mustHaveUnused:  "_ = direction",
			mustNotHaveNext: "directionSeries.Next()",
			description:     "String variable uses scalar, not Series",
		},
		{
			name: "string variable with ternary",
			program: &ast.Program{
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.Identifier{Name: "side"},
								Init: &ast.ConditionalExpression{
									Test: &ast.BinaryExpression{
										Operator: ">",
										Left:     &ast.Identifier{Name: "close"},
										Right:    &ast.Identifier{Name: "open"},
									},
									Consequent: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "strategy"},
										Property: &ast.Identifier{Name: "long"},
									},
									Alternate: &ast.MemberExpression{
										Object:   &ast.Identifier{Name: "strategy"},
										Property: &ast.Identifier{Name: "short"},
									},
								},
							},
						},
					},
				},
			},
			mustHaveDecl:    "var side string",
			mustNotHaveDecl: "var sideSeries",
			mustHaveInit:    "side = func() string {",
			mustNotHaveInit: "sideSeries = series.NewSeries",
			mustHaveUnused:  "_ = side",
			mustNotHaveNext: "sideSeries.Next()",
			description:     "Ternary string variable generates conditional scalar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := GenerateStrategyCodeFromAST(tt.program)
			if err != nil {
				t.Fatalf("generation failed: %v", err)
			}

			body := code.FunctionBody

			/* Declaration checks */
			if !strings.Contains(body, tt.mustHaveDecl) {
				t.Errorf("%s: missing declaration\nexpected: %s\n", tt.description, tt.mustHaveDecl)
			}
			if strings.Contains(body, tt.mustNotHaveDecl) {
				t.Errorf("%s: should NOT have Series declaration: %s\n", tt.description, tt.mustNotHaveDecl)
			}

			/* Initialization checks */
			if !strings.Contains(body, tt.mustHaveInit) {
				t.Errorf("%s: missing initialization\nexpected: %s\n", tt.description, tt.mustHaveInit)
			}
			if strings.Contains(body, tt.mustNotHaveInit) {
				t.Errorf("%s: should NOT have Series init: %s\n", tt.description, tt.mustNotHaveInit)
			}

			/* Unused suppression check */
			if !strings.Contains(body, tt.mustHaveUnused) {
				t.Errorf("%s: missing unused suppression: %s\n", tt.description, tt.mustHaveUnused)
			}

			/* Series.Next() exclusion */
			if strings.Contains(body, tt.mustNotHaveNext) {
				t.Errorf("%s: should NOT call Series.Next(): %s\n", tt.description, tt.mustNotHaveNext)
			}
		})
	}
}

/* Validates string scalars vs float64 Series in codegen */
func TestStringVariableVsFloatVariable(t *testing.T) {
	program := &ast.Program{
		Body: []ast.Node{
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "entry_type"},
						Init: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "strategy"},
							Property: &ast.Identifier{Name: "long"},
						},
					},
				},
			},
			&ast.VariableDeclaration{
				Declarations: []ast.VariableDeclarator{
					{
						ID: &ast.Identifier{Name: "signal"},
						Init: &ast.BinaryExpression{
							Operator: ">",
							Left:     &ast.Identifier{Name: "close"},
							Right:    &ast.Identifier{Name: "open"},
						},
					},
				},
			},
		},
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	body := code.FunctionBody

	/* String variable checks */
	if !strings.Contains(body, "var entry_type string") {
		t.Error("String variable should declare as string")
	}
	if strings.Contains(body, "var entry_typeSeries") {
		t.Error("String variable should NOT have Series declaration")
	}
	if strings.Contains(body, "entry_typeSeries = series.NewSeries") {
		t.Error("String variable should NOT initialize Series")
	}
	if !strings.Contains(body, "_ = entry_type") {
		t.Error("String variable should suppress unused warning")
	}

	/* Float variable checks */
	if !strings.Contains(body, "var signalSeries *series.Series") {
		t.Error("Bool variable should declare as Series")
	}
	if strings.Contains(body, "var signal string") {
		t.Error("Bool variable should NOT be string type")
	}
	if !strings.Contains(body, "signalSeries = series.NewSeries") {
		t.Error("Bool variable should initialize Series")
	}
	if !strings.Contains(body, "_ = signalSeries") {
		t.Error("Bool variable should suppress Series unused warning")
	}
	if !strings.Contains(body, "signalSeries.Next()") {
		t.Error("Bool variable should call Series.Next()")
	}
}
