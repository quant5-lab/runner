package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestUserDefinedFunctionDetector_IsUserDefinedFunction(t *testing.T) {
	tests := []struct {
		name       string
		registry   map[string]string
		funcName   string
		wantResult bool
	}{
		{
			name: "detect arrow function",
			registry: map[string]string{
				"dirmov": "function",
				"adx":    "function",
				"sma20":  "float",
			},
			funcName:   "dirmov",
			wantResult: true,
		},
		{
			name: "reject non-function variable",
			registry: map[string]string{
				"sma20": "float",
				"ema50": "float",
			},
			funcName:   "sma20",
			wantResult: false,
		},
		{
			name: "reject unknown identifier",
			registry: map[string]string{
				"dirmov": "function",
			},
			funcName:   "unknown_func",
			wantResult: false,
		},
		{
			name:       "reject on empty registry",
			registry:   map[string]string{},
			funcName:   "any_func",
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewUserDefinedFunctionDetector(tt.registry)
			result := detector.IsUserDefinedFunction(tt.funcName)

			if result != tt.wantResult {
				t.Errorf("IsUserDefinedFunction(%q) = %v, want %v", tt.funcName, result, tt.wantResult)
			}
		})
	}
}

/* IsUserDefinedCall answers the same question for an expression that may not be a call at all,
 * which is the shape a caller holds before it decides how to lower the expression. */
func TestUserDefinedFunctionDetector_IsUserDefinedCall(t *testing.T) {
	registry := map[string]string{
		"myUdf": "function",
		"sma20": "float",
	}

	var typedNilCall *ast.CallExpression

	tests := []struct {
		name string
		expr ast.Expression
		want bool
	}{
		{
			name: "identifier callee of a registered function",
			expr: &ast.CallExpression{Callee: &ast.Identifier{Name: "myUdf"}},
			want: true,
		},
		{
			name: "identifier callee of a non-function variable",
			expr: &ast.CallExpression{Callee: &ast.Identifier{Name: "sma20"}},
			want: false,
		},
		{
			name: "identifier callee of an unknown name",
			expr: &ast.CallExpression{Callee: &ast.Identifier{Name: "unknown"}},
			want: false,
		},
		{
			name: "namespace callee is never a user-defined function",
			expr: &ast.CallExpression{Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "myUdf"},
				Property: &ast.Identifier{Name: "tail"},
			}},
			want: false,
		},
		{
			name: "identifier expression that is not a call",
			expr: &ast.Identifier{Name: "myUdf"},
			want: false,
		},
		{
			name: "literal expression",
			expr: &ast.Literal{Value: 1.0},
			want: false,
		},
		{
			name: "nil expression",
			expr: nil,
			want: false,
		},
		{
			name: "typed nil call expression",
			expr: typedNilCall,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewUserDefinedFunctionDetector(registry)
			if result := detector.IsUserDefinedCall(tt.expr); result != tt.want {
				t.Errorf("IsUserDefinedCall() = %v, want %v", result, tt.want)
			}
		})
	}
}
