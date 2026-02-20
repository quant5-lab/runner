package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestTradeCollectionCallHandler_CanHandle(t *testing.T) {
	handler := NewTradeCollectionCallHandler()

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		// Closedtrades members
		{"closedtrades.profit", "strategy.closedtrades.profit", true},
		{"closedtrades.entry_price", "strategy.closedtrades.entry_price", true},
		{"closedtrades.exit_price", "strategy.closedtrades.exit_price", true},
		{"closedtrades.size", "strategy.closedtrades.size", true},

		// Opentrades members
		{"opentrades.profit", "strategy.opentrades.profit", true},
		{"opentrades.entry_price", "strategy.opentrades.entry_price", true},
		{"opentrades.size", "strategy.opentrades.size", true},

		// Invalid patterns
		{"wrong_namespace", "strat.closedtrades.profit", false},
		{"two_level_only", "strategy.closedtrades", false},
		{"unknown_collection", "strategy.allTrades.profit", false},
		{"unknown_property", "strategy.closedtrades.unknown", false},
		{"not_strategy", "ta.closedtrades.profit", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.CanHandle(tt.funcName)
			if got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestTradeCollectionCallHandler_GenerateCode(t *testing.T) {
	tests := []struct {
		name        string
		call        *ast.CallExpression
		wantCode    string
		wantErr     bool
		errContains string
	}{
		{
			name: "closedtrades.profit with index",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "closedtrades"},
					},
					Property: &ast.Identifier{Name: "profit"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "5"},
				},
			},
			wantCode: "tradeAccessor.ClosedTradeProfit(int(5))",
			wantErr:  false,
		},
		{
			name: "opentrades.entry_price with variable",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "opentrades"},
					},
					Property: &ast.Identifier{Name: "entry_price"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "tradeIndex"},
				},
			},
			wantCode: "tradeAccessor.OpenTradeEntryPrice(int(tradeIndex))",
			wantErr:  false,
		},
		{
			name: "closedtrades.size with zero",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "closedtrades"},
					},
					Property: &ast.Identifier{Name: "size"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "0"},
				},
			},
			wantCode: "tradeAccessor.ClosedTradeSize(int(0))",
			wantErr:  false,
		},
		{
			name: "invalid callee type",
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "profit"},
				Arguments: []ast.Expression{&ast.Literal{Value: "0"}},
			},
			wantCode: "",
			wantErr:  false,
		},
		{
			name: "non-identifier property",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "closedtrades"},
					},
					Property: &ast.Literal{Value: "profit"},
				},
				Arguments: []ast.Expression{&ast.Literal{Value: "0"}},
			},
			wantCode: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewTradeCollectionCallHandler()
			g := &generator{
				variables: make(map[string]string),
				constants: make(map[string]interface{}),
			}

			code, err := handler.GenerateCode(g, tt.call)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GenerateCode() expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GenerateCode() error = %q, want error containing %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("GenerateCode() unexpected error: %v", err)
				return
			}

			if code != tt.wantCode {
				t.Errorf("GenerateCode() code = %q, want %q", code, tt.wantCode)
			}
		})
	}
}

func TestTradeCollectionCallHandler_IntegrationWithRouter(t *testing.T) {
	// Verify handler integrates properly with an isolated router (not shared global state)
	router := &CallExpressionRouter{handlers: make([]CallExpressionHandler, 0)}
	router.RegisterHandler(NewTradeCollectionCallHandler())

	g := &generator{
		variables:  make(map[string]string),
		constants:  make(map[string]interface{}),
		callRouter: router,
	}

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "closedtrades"},
			},
			Property: &ast.Identifier{Name: "profit"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "0"},
		},
	}

	code, err := router.RouteCall(g, call)
	if err != nil {
		t.Fatalf("RouteCall() error = %v", err)
	}

	want := "tradeAccessor.ClosedTradeProfit(int(0))"
	if code != want {
		t.Errorf("RouteCall() code = %q, want %q", code, want)
	}
}
