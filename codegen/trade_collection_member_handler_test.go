package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestTradeCollectionMemberHandler_CanHandle(t *testing.T) {
	handler := NewTradeCollectionMemberHandler()

	tests := []struct {
		name   string
		object string
		member string
		want   bool
	}{
		// Valid closedtrades properties
		{"closedtrades_profit", "closedtrades", "profit", true},
		{"closedtrades_entry_price", "closedtrades", "entry_price", true},
		{"closedtrades_exit_price", "closedtrades", "exit_price", true},
		{"closedtrades_size", "closedtrades", "size", true},
		{"closedtrades_entry_id", "closedtrades", "entry_id", true},
		{"closedtrades_exit_id", "closedtrades", "exit_id", true},
		{"closedtrades_entry_comment", "closedtrades", "entry_comment", true},
		{"closedtrades_exit_comment", "closedtrades", "exit_comment", true},
		{"closedtrades_commission", "closedtrades", "commission", true},
		{"closedtrades_profit_percent", "closedtrades", "profit_percent", true},
		{"closedtrades_entry_bar_index", "closedtrades", "entry_bar_index", true},
		{"closedtrades_exit_bar_index", "closedtrades", "exit_bar_index", true},
		{"closedtrades_entry_time", "closedtrades", "entry_time", true},
		{"closedtrades_exit_time", "closedtrades", "exit_time", true},
		{"closedtrades_max_runup", "closedtrades", "max_runup", true},
		{"closedtrades_max_drawdown", "closedtrades", "max_drawdown", true},
		{"closedtrades_max_runup_percent", "closedtrades", "max_runup_percent", true},
		{"closedtrades_max_drawdown_percent", "closedtrades", "max_drawdown_percent", true},

		// Valid opentrades properties
		{"opentrades_profit", "opentrades", "profit", true},
		{"opentrades_entry_price", "opentrades", "entry_price", true},
		{"opentrades_size", "opentrades", "size", true},
		{"opentrades_entry_id", "opentrades", "entry_id", true},
		{"opentrades_entry_comment", "opentrades", "entry_comment", true},
		{"opentrades_commission", "opentrades", "commission", true},
		{"opentrades_profit_percent", "opentrades", "profit_percent", true},
		{"opentrades_entry_bar_index", "opentrades", "entry_bar_index", true},
		{"opentrades_entry_time", "opentrades", "entry_time", true},
		{"opentrades_max_runup", "opentrades", "max_runup", true},
		{"opentrades_max_drawdown", "opentrades", "max_drawdown", true},
		{"opentrades_max_runup_percent", "opentrades", "max_runup_percent", true},
		{"opentrades_max_drawdown_percent", "opentrades", "max_drawdown_percent", true},

		// Invalid: wrong object
		{"wrong_object", "trades", "profit", false},
		{"strategy_prefix", "strategy.closedtrades", "profit", false},
		{"unknown_collection", "allTrades", "profit", false},

		// Invalid: wrong property
		{"unknown_property", "closedtrades", "unknown", false},
		{"typo_property", "closedtrades", "proffit", false},
		{"case_mismatch", "closedtrades", "Profit", false},

		// Edge cases
		{"empty_object", "", "profit", false},
		{"empty_member", "closedtrades", "", false},
		{"both_empty", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.CanHandle(tt.object, tt.member)
			if got != tt.want {
				t.Errorf("CanHandle(%q, %q) = %v, want %v", tt.object, tt.member, got, tt.want)
			}
		})
	}
}

func TestTradeCollectionMemberHandler_GenerateAccess_ArgumentValidation(t *testing.T) {
	handler := NewTradeCollectionMemberHandler()
	mockGen := &mockExpressionGenerator{}

	tests := []struct {
		name    string
		object  string
		member  string
		args    []ast.Expression
		wantErr string
	}{
		{
			name:    "no_arguments",
			object:  "closedtrades",
			member:  "profit",
			args:    []ast.Expression{},
			wantErr: "requires exactly 1 argument",
		},
		{
			name:    "multiple_arguments",
			object:  "closedtrades",
			member:  "profit",
			args:    []ast.Expression{&ast.Literal{Value: "0"}, &ast.Literal{Value: "1"}},
			wantErr: "requires exactly 1 argument",
		},
		{
			name:    "nil_arguments_slice",
			object:  "closedtrades",
			member:  "profit",
			args:    nil,
			wantErr: "requires exactly 1 argument",
		},
		{
			name:    "unknown_property",
			object:  "closedtrades",
			member:  "unknown_prop",
			args:    []ast.Expression{&ast.Literal{Value: "0"}},
			wantErr: "unknown trade property",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.GenerateAccess(tt.object, tt.member, tt.args, mockGen)
			if err == nil {
				t.Errorf("GenerateAccess() expected error containing %q, got nil", tt.wantErr)
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("GenerateAccess() error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestTradeCollectionMemberHandler_GenerateAccess_CodeGeneration(t *testing.T) {
	handler := NewTradeCollectionMemberHandler()

	tests := []struct {
		name     string
		object   string
		member   string
		args     []ast.Expression
		wantCode string
	}{
		// Closedtrades numeric properties with literal index
		{"closedtrades_profit_literal", "closedtrades", "profit", []ast.Expression{&ast.Literal{Value: "5"}}, "tradeAccessor.ClosedTradeProfit(int(5))"},
		{"closedtrades_entry_price_zero", "closedtrades", "entry_price", []ast.Expression{&ast.Literal{Value: "0"}}, "tradeAccessor.ClosedTradeEntryPrice(int(0))"},
		{"closedtrades_exit_price", "closedtrades", "exit_price", []ast.Expression{&ast.Literal{Value: "10"}}, "tradeAccessor.ClosedTradeExitPrice(int(10))"},
		{"closedtrades_size_literal", "closedtrades", "size", []ast.Expression{&ast.Literal{Value: "3"}}, "tradeAccessor.ClosedTradeSize(int(3))"},
		{"closedtrades_commission", "closedtrades", "commission", []ast.Expression{&ast.Literal{Value: "0"}}, "tradeAccessor.ClosedTradeCommission(int(0))"},
		{"closedtrades_profit_percent", "closedtrades", "profit_percent", []ast.Expression{&ast.Literal{Value: "1"}}, "tradeAccessor.ClosedTradeProfitPercent(int(1))"},
		{"closedtrades_entry_bar_index", "closedtrades", "entry_bar_index", []ast.Expression{&ast.Literal{Value: "2"}}, "tradeAccessor.ClosedTradeEntryBarIndex(int(2))"},
		{"closedtrades_exit_bar_index", "closedtrades", "exit_bar_index", []ast.Expression{&ast.Literal{Value: "2"}}, "tradeAccessor.ClosedTradeExitBarIndex(int(2))"},
		{"closedtrades_entry_time", "closedtrades", "entry_time", []ast.Expression{&ast.Literal{Value: "0"}}, "tradeAccessor.ClosedTradeEntryTime(int(0))"},
		{"closedtrades_exit_time", "closedtrades", "exit_time", []ast.Expression{&ast.Literal{Value: "0"}}, "tradeAccessor.ClosedTradeExitTime(int(0))"},
		{"closedtrades_max_runup", "closedtrades", "max_runup", []ast.Expression{&ast.Literal{Value: "0"}}, "tradeAccessor.ClosedTradeMaxRunup(int(0))"},
		{"closedtrades_max_drawdown", "closedtrades", "max_drawdown", []ast.Expression{&ast.Literal{Value: "0"}}, "tradeAccessor.ClosedTradeMaxDrawdown(int(0))"},

		// Closedtrades with identifier index
		{"closedtrades_profit_identifier", "closedtrades", "profit", []ast.Expression{&ast.Identifier{Name: "tradeIdx"}}, "tradeAccessor.ClosedTradeProfit(int(tradeIdx))"},
		{"closedtrades_entry_price_var", "closedtrades", "entry_price", []ast.Expression{&ast.Identifier{Name: "i"}}, "tradeAccessor.ClosedTradeEntryPrice(int(i))"},

		// Opentrades properties
		{"opentrades_profit_literal", "opentrades", "profit", []ast.Expression{&ast.Literal{Value: "2"}}, "tradeAccessor.OpenTradeProfit(int(2))"},
		{"opentrades_entry_price", "opentrades", "entry_price", []ast.Expression{&ast.Literal{Value: "0"}}, "tradeAccessor.OpenTradeEntryPrice(int(0))"},
		{"opentrades_size", "opentrades", "size", []ast.Expression{&ast.Literal{Value: "1"}}, "tradeAccessor.OpenTradeSize(int(1))"},
		{"opentrades_commission", "opentrades", "commission", []ast.Expression{&ast.Literal{Value: "0"}}, "tradeAccessor.OpenTradeCommission(int(0))"},
		{"opentrades_profit_percent", "opentrades", "profit_percent", []ast.Expression{&ast.Literal{Value: "0"}}, "tradeAccessor.OpenTradeProfitPercent(int(0))"},
		{"opentrades_entry_bar_index", "opentrades", "entry_bar_index", []ast.Expression{&ast.Literal{Value: "0"}}, "tradeAccessor.OpenTradeEntryBarIndex(int(0))"},
		{"opentrades_max_runup", "opentrades", "max_runup", []ast.Expression{&ast.Literal{Value: "0"}}, "tradeAccessor.OpenTradeMaxRunup(int(0))"},

		// String properties (entry_id, exit_id, comments)
		{"closedtrades_entry_id", "closedtrades", "entry_id", []ast.Expression{&ast.Literal{Value: "0"}}, "tradeAccessor.ClosedTradeEntryID(int(0))"},
		{"closedtrades_exit_id", "closedtrades", "exit_id", []ast.Expression{&ast.Literal{Value: "0"}}, "tradeAccessor.ClosedTradeExitID(int(0))"},
		{"closedtrades_entry_comment", "closedtrades", "entry_comment", []ast.Expression{&ast.Literal{Value: "0"}}, "tradeAccessor.ClosedTradeEntryComment(int(0))"},
		{"closedtrades_exit_comment", "closedtrades", "exit_comment", []ast.Expression{&ast.Literal{Value: "0"}}, "tradeAccessor.ClosedTradeExitComment(int(0))"},
		{"opentrades_entry_id", "opentrades", "entry_id", []ast.Expression{&ast.Literal{Value: "0"}}, "tradeAccessor.OpenTradeEntryID(int(0))"},
		{"opentrades_entry_comment", "opentrades", "entry_comment", []ast.Expression{&ast.Literal{Value: "0"}}, "tradeAccessor.OpenTradeEntryComment(int(0))"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGen := &mockExpressionGenerator{}
			code, err := handler.GenerateAccess(tt.object, tt.member, tt.args, mockGen)

			if err != nil {
				t.Fatalf("GenerateAccess() unexpected error: %v", err)
			}

			if code != tt.wantCode {
				t.Errorf("GenerateAccess() code = %q, want %q", code, tt.wantCode)
			}
		})
	}
}

func TestTradeCollectionMemberHandler_ComplexArgumentExpressions(t *testing.T) {
	handler := NewTradeCollectionMemberHandler()

	tests := []struct {
		name         string
		arg          ast.Expression
		mockResponse string
		wantCode     string
	}{
		{
			name:         "binary_expression_index",
			arg:          &ast.BinaryExpression{Left: &ast.Identifier{Name: "count"}, Operator: "-", Right: &ast.Literal{Value: "1"}},
			mockResponse: "(count - 1)",
			wantCode:     "tradeAccessor.ClosedTradeProfit(int((count - 1)))",
		},
		{
			name:         "call_expression_index",
			arg:          &ast.CallExpression{Callee: &ast.Identifier{Name: "getIndex"}},
			mockResponse: "getIndex()",
			wantCode:     "tradeAccessor.ClosedTradeProfit(int(getIndex()))",
		},
		{
			name:         "conditional_expression_index",
			arg:          &ast.ConditionalExpression{Test: &ast.Identifier{Name: "useLast"}, Consequent: &ast.Literal{Value: "0"}, Alternate: &ast.Literal{Value: "1"}},
			mockResponse: "(useLast ? 0 : 1)",
			wantCode:     "tradeAccessor.ClosedTradeProfit(int((useLast ? 0 : 1)))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGen := &mockExpressionGenerator{returnValue: tt.mockResponse}
			code, err := handler.GenerateAccess("closedtrades", "profit", []ast.Expression{tt.arg}, mockGen)

			if err != nil {
				t.Fatalf("GenerateAccess() error = %v", err)
			}

			if code != tt.wantCode {
				t.Errorf("GenerateAccess() code = %q, want %q", code, tt.wantCode)
			}

			if mockGen.callCount != 1 {
				t.Errorf("Generator called %d times, want 1", mockGen.callCount)
			}
		})
	}
}

func TestTradeCollectionMemberHandler_PropertyMapping(t *testing.T) {
	handler := NewTradeCollectionMemberHandler()

	/* Verify all 18 properties in the property map generate valid code */
	propertyTests := []struct {
		pineProperty string
		goMethodPart string
	}{
		{"commission", "Commission"},
		{"entry_bar_index", "EntryBarIndex"},
		{"entry_comment", "EntryComment"},
		{"entry_id", "EntryID"},
		{"entry_price", "EntryPrice"},
		{"entry_time", "EntryTime"},
		{"exit_bar_index", "ExitBarIndex"},
		{"exit_comment", "ExitComment"},
		{"exit_id", "ExitID"},
		{"exit_price", "ExitPrice"},
		{"exit_time", "ExitTime"},
		{"max_drawdown", "MaxDrawdown"},
		{"max_drawdown_percent", "MaxDrawdown"},
		{"max_runup", "MaxRunup"},
		{"max_runup_percent", "MaxRunup"},
		{"profit", "Profit"},
		{"profit_percent", "ProfitPercent"},
		{"size", "Size"},
	}

	for _, tt := range propertyTests {
		t.Run(tt.pineProperty, func(t *testing.T) {
			mockGen := &mockExpressionGenerator{}
			code, err := handler.GenerateAccess("closedtrades", tt.pineProperty, []ast.Expression{&ast.Literal{Value: "0"}}, mockGen)

			if err != nil {
				t.Fatalf("GenerateAccess() error = %v", err)
			}

			expectedCode := "tradeAccessor.ClosedTrade" + tt.goMethodPart + "(int(0))"
			if code != expectedCode {
				t.Errorf("Property %q mapped to %q, want %q", tt.pineProperty, code, expectedCode)
			}
		})
	}
}

func TestTradeCollectionMemberHandler_EdgeCases(t *testing.T) {
	handler := NewTradeCollectionMemberHandler()

	t.Run("empty_property_name", func(t *testing.T) {
		got := handler.CanHandle("closedtrades", "")
		if got {
			t.Error("CanHandle() should return false for empty property name")
		}
	})

	t.Run("whitespace_in_property", func(t *testing.T) {
		tests := []string{"profit ", " profit", " profit "}
		for _, prop := range tests {
			if handler.CanHandle("closedtrades", prop) {
				t.Errorf("CanHandle() should reject property with whitespace: %q", prop)
			}
		}
	})

	t.Run("case_sensitive_object", func(t *testing.T) {
		tests := []string{"ClosedTrades", "CLOSEDTRADES", "closedTrades", "Closedtrades"}
		for _, obj := range tests {
			if handler.CanHandle(obj, "profit") {
				t.Errorf("CanHandle() should be case-sensitive for object, but accepted %q", obj)
			}
		}
	})

	t.Run("case_sensitive_property", func(t *testing.T) {
		tests := []string{"Profit", "PROFIT", "pRoFiT"}
		for _, prop := range tests {
			if handler.CanHandle("closedtrades", prop) {
				t.Errorf("CanHandle() should be case-sensitive for property, but accepted %q", prop)
			}
		}
	})

	t.Run("unicode_in_names", func(t *testing.T) {
		if handler.CanHandle("closedtrades™", "profit") {
			t.Error("CanHandle() should reject unicode in object names")
		}
		if handler.CanHandle("closedtrades", "profit™") {
			t.Error("CanHandle() should reject unicode in property names")
		}
	})

	t.Run("special_characters", func(t *testing.T) {
		specialChars := []string{"profit-value", "profit.value", "profit/value", "profit#value"}
		for _, prop := range specialChars {
			if handler.CanHandle("closedtrades", prop) {
				t.Errorf("CanHandle() should reject property with special chars: %q", prop)
			}
		}
	})

	t.Run("expression_generator_error_propagation", func(t *testing.T) {
		failingGen := &mockExpressionGenerator{returnError: &mockError{}}
		expr := &ast.BinaryExpression{Left: &ast.Identifier{Name: "x"}, Operator: "+", Right: &ast.Literal{Value: "1"}}

		_, err := handler.GenerateAccess("closedtrades", "profit", []ast.Expression{expr}, failingGen)
		if err == nil {
			t.Error("GenerateAccess() should propagate generator errors")
		}
	})
}

type mockError struct{}

func (m *mockError) Error() string { return "mock error" }

type mockExpressionGenerator struct {
	callCount   int
	returnValue string
	returnError error
}

func (m *mockExpressionGenerator) Generate(expr ast.Expression) (string, error) {
	m.callCount++

	if m.returnError != nil {
		return "", m.returnError
	}

	if m.returnValue != "" {
		return m.returnValue, nil
	}

	// Default behavior: simple string representation
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Name, nil
	case *ast.Literal:
		return e.Value.(string), nil
	default:
		return "expr", nil
	}
}
