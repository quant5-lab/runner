package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* Validates all dispatch outcomes: actionable code, empty output, comment-only output, nil router */
func TestExpressionPositionDispatcher_DispatchOutcomes(t *testing.T) {
	gen := newTestGenerator()

	tests := []struct {
		name         string
		router       *CallExpressionRouter
		call         *ast.CallExpression
		wantErr      bool
		wantNonEmpty bool
	}{
		{
			name:   "actionable handler output",
			router: gen.callRouter,
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "heikinashi"},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "syminfo"},
						Property: &ast.Identifier{Name: "tickerid"},
					},
				},
			},
			wantErr:      false,
			wantNonEmpty: true,
		},
		{
			name:   "empty handler output",
			router: gen.callRouter,
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "indicator"},
			},
			wantErr: true,
		},
		{
			name:   "comment-only handler output",
			router: gen.callRouter,
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "nonexistent"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
				},
			},
			wantErr: true,
		},
		{
			name:   "nil router",
			router: nil,
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "anyFunction"},
				Arguments: []ast.Expression{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dispatcher := NewExpressionPositionDispatcher(tt.router)
			code, err := dispatcher.Dispatch(gen, tt.call)

			if (err != nil) != tt.wantErr {
				t.Errorf("Dispatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantNonEmpty {
				if code == "" {
					t.Error("Expected non-empty code")
				}
				if code != strings.TrimSpace(code) {
					t.Errorf("Output should be trimmed, got: %q", code)
				}
			}
		})
	}
}
