package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestCallExpressionRouter_VariableAssignmentRouting validates CallExpressionRouter
 * integration with generateVariableFromCall() across all registered handlers.
 * Tests that ALL handlers (str.*, ticker.*, math.*, value.*, etc.) route correctly
 * in variable assignment expressions, not just conditionals.
 */
func TestCallExpressionRouter_VariableAssignmentRouting(t *testing.T) {
	gen := newTestGenerator()

	tests := []struct {
		name         string
		call         *ast.CallExpression
		wantContains string /* Expected code pattern in output */
		wantErr      bool
	}{
		{
			name: "str.tostring in variable assignment",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "str"},
					Property: &ast.Identifier{Name: "tostring"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
				},
			},
			wantContains: "fmt.Sprintf",
			wantErr:      false,
		},
		{
			name: "str.length in variable assignment",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "str"},
					Property: &ast.Identifier{Name: "length"},
				},
				Arguments: []ast.Expression{
					&ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "str"},
							Property: &ast.Identifier{Name: "tostring"},
						},
						Arguments: []ast.Expression{
							&ast.Identifier{Name: "close"},
						},
					},
				},
			},
			wantContains: "len(",
			wantErr:      false,
		},
		{
			name: "ticker.heikinashi in variable assignment",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "heikinashi"},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "syminfo"},
						Property: &ast.Identifier{Name: "tickerid"},
					},
				},
			},
			wantContains: "ticker.Heikinashi", /* Actual output is ticker.Heikinashi */
			wantErr:      false,
		},
		{
			name: "math.abs in variable assignment",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: "abs"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
				},
			},
			wantContains: "math.Abs",
			wantErr:      false,
		},
		{
			name: "math.pow in variable assignment",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: "pow"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 2.0},
				},
			},
			wantContains: "math.Pow",
			wantErr:      false,
		},
		{
			name: "value functions in variable assignment",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "na"},
			},
			wantContains: "true", /* na() generates "true", not math.NaN() */
			wantErr:      false,
		},
		{
			name: "unregistered function falls back to TODO",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "unregistered_func"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
				},
			},
			wantContains: "TODO: implement unregistered_func",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := gen.generateVariableFromCall("testVar", tt.call)

			if (err != nil) != tt.wantErr {
				t.Errorf("generateVariableFromCall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !strings.Contains(code, tt.wantContains) {
				t.Errorf("generateVariableFromCall() code = %q, want to contain %q", code, tt.wantContains)
			}

			/* Verify output structure: must have Series.Set() call */
			if !strings.Contains(code, "testVarSeries.Set(") {
				t.Errorf("Expected testVarSeries.Set() in output, got: %q", code)
			}
		})
	}
}

/* TestCallExpressionRouter_ConditionalExpressionRouting validates handler routing
 * in conditional expressions (if statements, ternary operators, etc.).
 * This is the complementary test to VariableAssignmentRouting.
 */
func TestCallExpressionRouter_ConditionalExpressionRouting(t *testing.T) {
	gen := newTestGenerator()

	tests := []struct {
		name         string
		call         *ast.CallExpression
		wantContains string
		wantErr      bool
	}{
		{
			name: "str.tostring in conditional",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "str"},
					Property: &ast.Identifier{Name: "tostring"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
				},
			},
			wantContains: "fmt.Sprintf",
			wantErr:      false,
		},
		{
			name: "math function in conditional",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: "abs"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
				},
			},
			wantContains: "math.Abs",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			/* Use router directly - conditionals don't go through generateVariableFromCall */
			code, err := gen.callRouter.RouteCall(gen, tt.call)

			if (err != nil) != tt.wantErr {
				t.Errorf("RouteCall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !strings.Contains(code, tt.wantContains) {
				t.Errorf("RouteCall() code = %q, want to contain %q", code, tt.wantContains)
			}
		})
	}
}

/* TestCallExpressionRouter_PlotTitleRouting validates handler routing in plot title expressions.
 * Plot titles are special - they're evaluated once at strategy initialization, not per-bar.
 */
func TestCallExpressionRouter_PlotTitleRouting(t *testing.T) {
	gen := newTestGenerator()

	tests := []struct {
		name         string
		call         *ast.CallExpression
		wantContains string
		wantErr      bool
	}{
		{
			name: "str.tostring in plot title",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "str"},
					Property: &ast.Identifier{Name: "tostring"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
				},
			},
			wantContains: "fmt.Sprintf",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := gen.callRouter.RouteCall(gen, tt.call)

			if (err != nil) != tt.wantErr {
				t.Errorf("RouteCall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !strings.Contains(code, tt.wantContains) {
				t.Errorf("RouteCall() code = %q, want to contain %q", code, tt.wantContains)
			}
		})
	}
}

/* TestCallExpressionRouter_NestedCallRouting validates handler routing for nested call expressions.
 * Example: str.length(str.tostring(...)) should route both calls correctly.
 */
func TestCallExpressionRouter_NestedCallRouting(t *testing.T) {
	gen := newTestGenerator()

	tests := []struct {
		name         string
		call         *ast.CallExpression
		wantContains []string /* All expected patterns in nested output */
		wantErr      bool
	}{
		{
			name: "str.length(str.tostring(...))",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "str"},
					Property: &ast.Identifier{Name: "length"},
				},
				Arguments: []ast.Expression{
					&ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "str"},
							Property: &ast.Identifier{Name: "tostring"},
						},
						Arguments: []ast.Expression{
							&ast.Identifier{Name: "close"},
						},
					},
				},
			},
			wantContains: []string{"len(", "fmt.Sprintf"},
			wantErr:      false,
		},
		{
			name: "math.abs(math.pow(...))",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: "abs"},
				},
				Arguments: []ast.Expression{
					&ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "math"},
							Property: &ast.Identifier{Name: "pow"},
						},
						Arguments: []ast.Expression{
							&ast.Identifier{Name: "close"},
							&ast.Literal{Value: 2.0},
						},
					},
				},
			},
			wantContains: []string{"math.Abs", "math.Pow"},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := gen.generateVariableFromCall("testVar", tt.call)

			if (err != nil) != tt.wantErr {
				t.Errorf("generateVariableFromCall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, wantPattern := range tt.wantContains {
				if !strings.Contains(code, wantPattern) {
					t.Errorf("generateVariableFromCall() code = %q, want to contain %q", code, wantPattern)
				}
			}
		})
	}
}

/* TestCallExpressionRouter_EdgeCases validates routing behavior for edge cases:
 * - Empty arguments
 * - Nil expressions
 * - Invalid function names
 * - Mixed registered/unregistered calls
 */
func TestCallExpressionRouter_EdgeCases(t *testing.T) {
	gen := newTestGenerator()

	tests := []struct {
		name         string
		call         *ast.CallExpression
		wantContains string
		wantErr      bool
	}{
		{
			name: "str function with no arguments",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "str"},
					Property: &ast.Identifier{Name: "toupper"},
				},
				Arguments: []ast.Expression{},
			},
			wantContains: "TODO", /* Handler validation error → falls back to TODO */
			wantErr:      false,
		},
		{
			name: "unknown namespace function",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "unknown"},
					Property: &ast.Identifier{Name: "func"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "test"},
				},
			},
			wantContains: "TODO: implement unknown.func",
			wantErr:      false,
		},
		{
			name: "bare unknown function",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "unknown_bare"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "test"},
				},
			},
			wantContains: "TODO: implement unknown_bare",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := gen.generateVariableFromCall("testVar", tt.call)

			if (err != nil) != tt.wantErr {
				t.Errorf("generateVariableFromCall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !strings.Contains(code, tt.wantContains) {
				t.Errorf("generateVariableFromCall() code = %q, want to contain %q", code, tt.wantContains)
			}
		})
	}
}

/* TestCallExpressionRouter_AllHandlersCovered validates that ALL registered handlers
 * route correctly through generateVariableFromCall(). This prevents regressions where
 * new handlers are added to CallExpressionRouter but variable assignment path bypasses them.
 */
func TestCallExpressionRouter_AllHandlersCovered(t *testing.T) {
	gen := newTestGenerator()

	/* Sample call for each handler - one per registered handler */
	handlerSamples := []struct {
		handlerName string
		call        *ast.CallExpression
		wantNotTODO bool /* True if handler should generate real code, not TODO */
	}{
		{
			handlerName: "MetaFunctionHandler",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "indicator"},
			},
			wantNotTODO: false, /* Meta functions are filtered out earlier */
		},
		{
			handlerName: "StrategyActionHandler",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Buy"},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "long"},
					},
				},
			},
			wantNotTODO: true,
		},
		{
			handlerName: "InputHandler",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "int"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 10},
				},
			},
			wantNotTODO: false, /* input.* functions are filtered at strategy level */
		},
		{
			handlerName: "ValueHandler",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "na"},
			},
			wantNotTODO: true,
		},
		{
			handlerName: "MathFunctionHandler",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: "abs"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
				},
			},
			wantNotTODO: true,
		},
		{
			handlerName: "TickerHandler",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "heikinashi"},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "syminfo"},
						Property: &ast.Identifier{Name: "tickerid"},
					},
				},
			},
			wantNotTODO: true,
		},
		{
			handlerName: "StringNamespaceHandler",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "str"},
					Property: &ast.Identifier{Name: "tostring"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
				},
			},
			wantNotTODO: true,
		},
	}

	for _, tt := range handlerSamples {
		t.Run(tt.handlerName, func(t *testing.T) {
			code, err := gen.generateVariableFromCall("testVar", tt.call)
			if err != nil {
				t.Fatalf("generateVariableFromCall() error = %v", err)
			}

			hasTODO := strings.Contains(code, "TODO")

			if tt.wantNotTODO && hasTODO {
				t.Errorf("%s handler generated TODO comment, expected real code: %q", tt.handlerName, code)
			}
		})
	}
}
