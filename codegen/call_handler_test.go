package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// TestCallExpressionRouter_Registration verifies handler registration and chain ordering
func TestCallExpressionRouter_Registration(t *testing.T) {
	router := NewCallExpressionRouter()

	if router == nil {
		t.Fatal("NewCallExpressionRouter() returned nil")
	}

	if len(router.handlers) == 0 {
		t.Error("Router has no registered handlers")
	}

	lastHandler := router.handlers[len(router.handlers)-1]
	if _, ok := lastHandler.(*UnknownFunctionHandler); !ok {
		t.Error("Last handler should be UnknownFunctionHandler (catch-all)")
	}
}

// TestCallExpressionRouter_HandlersCanHandleCorrectFunctions tests that each handler
// only claims functions it should handle (no overlap, no gaps)
func TestCallExpressionRouter_HandlersCanHandleCorrectFunctions(t *testing.T) {
	router := NewCallExpressionRouter()

	tests := []struct {
		funcName       string
		wantHandlerIdx int // Index of handler that should handle this
		handlerType    string
	}{
		{"indicator", 0, "MetaFunctionHandler"},
		{"strategy", 0, "MetaFunctionHandler"},
		{"plot", 1, "PlotFunctionHandler"},
		{"strategy.entry", 2, "StrategyActionHandler"},
		{"strategy.close", 2, "StrategyActionHandler"},
		{"strategy.close_all", 2, "StrategyActionHandler"},
		{"strategy.closedtrades.profit", 3, "TradeCollectionCallHandler"},
		{"strategy.opentrades.size", 3, "TradeCollectionCallHandler"},
		{"abs", 4, "MathCallHandler"},
		{"math.abs", 4, "MathCallHandler"},
		{"nz", 5, "ValueCallHandler"},
		{"na", 5, "ValueCallHandler"},
		{"ta.sma", 6, "TAIndicatorCallHandler"},
		{"ta.ema", 6, "TAIndicatorCallHandler"},
		{"ta.crossover", 6, "TAIndicatorCallHandler"},
		{"valuewhen", 6, "TAIndicatorCallHandler"},
		{"color.new", 8, "ColorCallHandler"},
		{"color.rgb", 8, "ColorCallHandler"},
		{"color.from_gradient", 8, "ColorCallHandler"},
		{"color.r", 8, "ColorCallHandler"},
		{"color.g", 8, "ColorCallHandler"},
		{"color.b", 8, "ColorCallHandler"},
		{"color.t", 8, "ColorCallHandler"},
		{"year", 9, "CalendarCallHandler"},
		{"timestamp", 9, "CalendarCallHandler"},
		{"dayofweek", 9, "CalendarCallHandler"},
		{"timeframe.in_seconds", 10, "TimeframeFuncCallHandler"},
		{"timeframe.from_seconds", 10, "TimeframeFuncCallHandler"},
		{"timeframe.change", 10, "TimeframeFuncCallHandler"},
		{"unknown_function", 13, "UnknownFunctionHandler"},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			foundHandlerIdx := -1
			for i, handler := range router.handlers {
				if handler.CanHandle(tt.funcName) {
					foundHandlerIdx = i
					break
				}
			}

			if foundHandlerIdx == -1 {
				t.Errorf("No handler claims %q", tt.funcName)
			} else if foundHandlerIdx != tt.wantHandlerIdx {
				t.Errorf("Function %q handled by index %d, want index %d",
					tt.funcName, foundHandlerIdx, tt.wantHandlerIdx)
			}
		})
	}
}

// TestCallExpressionRouter_RouteCall verifies routing delegates to correct handler
func TestCallExpressionRouter_RouteCall(t *testing.T) {
	g := newTestGenerator()
	router := NewCallExpressionRouter()

	tests := []struct {
		name     string
		call     *ast.CallExpression
		wantCode string // Expected code pattern
		wantErr  bool
	}{
		{
			name: "meta function indicator",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "indicator"},
			},
			wantCode: "", // Meta functions produce no code
			wantErr:  false,
		},
		{
			name: "meta function strategy",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "strategy"},
			},
			wantCode: "",
			wantErr:  false,
		},
		{
			name: "strategy.entry valid",
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
			wantCode: "strat.Entry(",
			wantErr:  false,
		},
		{
			name: "strategy.entry invalid args",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Buy"},
				},
			},
			wantCode: "// strategy.entry() - invalid arguments",
			wantErr:  false,
		},
		{
			name: "strategy.close valid",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "close"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Buy"},
				},
			},
			wantCode: "strat.Close(",
			wantErr:  false,
		},
		{
			name: "strategy.close_all",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "close_all"},
				},
			},
			wantCode: "strat.CloseAll(",
			wantErr:  false,
		},
		{
			name: "unknown function",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "unknown_func"},
			},
			wantCode: "// unknown_func() - TODO: implement",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := router.RouteCall(g, tt.call)

			if (err != nil) != tt.wantErr {
				t.Errorf("RouteCall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantCode != "" && !strings.Contains(code, tt.wantCode) {
				t.Errorf("RouteCall() code = %q, want to contain %q", code, tt.wantCode)
			}
		})
	}
}

// TestExtractCallFunctionName verifies function name extraction from various AST structures
func TestExtractCallFunctionName(t *testing.T) {
	tests := []struct {
		name string
		call *ast.CallExpression
		want string
	}{
		{
			name: "simple identifier",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "plot"},
			},
			want: "plot",
		},
		{
			name: "member expression ta.sma",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
			},
			want: "ta.sma",
		},
		{
			name: "member expression strategy.entry",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
			},
			want: "strategy.entry",
		},
		{
			name: "nested member expression (edge case)",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "outer"},
						Property: &ast.Identifier{Name: "inner"},
					},
					Property: &ast.Identifier{Name: "prop"},
				},
			},
			want: "outer.inner.prop", // Nested member expressions now supported (for strategy.closedtrades.profit etc)
		},
		{
			name: "non-identifier property",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "obj"},
					Property: &ast.Literal{Value: "prop"},
				},
			},
			want: "", // Non-identifier property returns empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCallFunctionName(tt.call)
			if got != tt.want {
				t.Errorf("extractCallFunctionName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCallExpressionRouter_NilSafety tests router behavior with nil inputs
func TestCallExpressionRouter_NilSafety(t *testing.T) {
	router := NewCallExpressionRouter()
	g := newTestGenerator()

	tests := []struct {
		name string
		call *ast.CallExpression
	}{
		{
			name: "nil callee",
			call: &ast.CallExpression{
				Callee: nil,
			},
		},
		{
			name: "empty member expression",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := router.RouteCall(g, tt.call)
			if err != nil {
				// Error is acceptable for invalid input
				return
			}
			// Fallback to unknown handler TODO comment
			if !strings.Contains(code, "TODO") && code != "" {
				t.Errorf("Expected TODO comment or empty code for invalid input, got: %q", code)
			}
		})
	}
}

// TestCallExpressionRouter_HandlerPriority verifies handler priority in chain
// More specific handlers should be checked before generic ones
func TestCallExpressionRouter_HandlerPriority(t *testing.T) {
	router := NewCallExpressionRouter()

	// ta.sma should be handled by TAIndicatorCallHandler, not UnknownFunctionHandler
	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
	}

	g := newTestGenerator()
	code, err := router.RouteCall(g, call)
	if err != nil {
		t.Fatalf("RouteCall() error = %v", err)
	}

	// Should produce empty code (handled in declarations), not TODO
	if strings.Contains(code, "TODO") {
		t.Errorf("ta.sma should be handled by specific handler, not unknown handler. Got: %q", code)
	}
}

// TestCallExpressionRouter_CustomHandlerRegistration tests dynamic handler registration
func TestCallExpressionRouter_CustomHandlerRegistration(t *testing.T) {
	router := &CallExpressionRouter{
		handlers: make([]CallExpressionHandler, 0),
	}

	// Register custom handler
	customHandler := &MetaFunctionHandler{}
	router.RegisterHandler(customHandler)

	if len(router.handlers) != 1 {
		t.Errorf("Expected 1 handler, got %d", len(router.handlers))
	}

	// Verify registered handler works
	if !router.handlers[0].CanHandle("indicator") {
		t.Error("Registered handler should handle 'indicator'")
	}
}
