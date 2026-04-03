package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestStrategyConfigExtractor_EmptyCall verifies behavior with no arguments */
func TestStrategyConfigExtractor_EmptyCall(t *testing.T) {
	extractor := NewStrategyConfigExtractor()
	call := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "strategy"},
		Arguments: []ast.Expression{},
	}

	config := extractor.ExtractFromCall(call)

	if config == nil {
		t.Fatal("ExtractFromCall should return non-nil config for empty call")
	}
	if config.Name != "Generated Strategy" {
		t.Errorf("Expected default name, got '%s'", config.Name)
	}
	if config.InitialCapital != defaultInitialCapital {
		t.Errorf("Expected default capital %.2f, got %.2f", defaultInitialCapital, config.InitialCapital)
	}
}

/* TestStrategyConfigExtractor_NameOnly verifies extraction with title argument */
func TestStrategyConfigExtractor_NameOnly(t *testing.T) {
	tests := []struct {
		name         string
		arg          ast.Expression
		expectedName string
	}{
		{
			name:         "string literal name",
			arg:          &ast.Literal{Value: "My Strategy"},
			expectedName: "My Strategy",
		},
		{
			name:         "empty string name",
			arg:          &ast.Literal{Value: ""},
			expectedName: "",
		},
		{
			name:         "non-string literal ignored",
			arg:          &ast.Literal{Value: 42},
			expectedName: "Generated Strategy",
		},
		{
			name:         "identifier ignored",
			arg:          &ast.Identifier{Name: "strategyName"},
			expectedName: "Generated Strategy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewStrategyConfigExtractor()
			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "strategy"},
				Arguments: []ast.Expression{tt.arg},
			}

			config := extractor.ExtractFromCall(call)

			if config.Name != tt.expectedName {
				t.Errorf("Expected name '%s', got '%s'", tt.expectedName, config.Name)
			}
		})
	}
}

/* TestStrategyConfigExtractor_InitialCapital verifies initial_capital extraction */
func TestStrategyConfigExtractor_InitialCapital(t *testing.T) {
	tests := []struct {
		name            string
		value           interface{}
		expectedCapital float64
	}{
		{
			name:            "float capital",
			value:           50000.0,
			expectedCapital: 50000.0,
		},
		{
			name:            "integer capital",
			value:           25000,
			expectedCapital: 25000.0,
		},
		{
			name:            "zero capital",
			value:           0.0,
			expectedCapital: 0.0,
		},
		{
			name:            "fractional capital",
			value:           12345.67,
			expectedCapital: 12345.67,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewStrategyConfigExtractor()
			call := &ast.CallExpression{
				Callee: &ast.Identifier{Name: "strategy"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Test"},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "initial_capital"},
								Value: &ast.Literal{Value: tt.value},
							},
						},
					},
				},
			}

			config := extractor.ExtractFromCall(call)

			if config.InitialCapital != tt.expectedCapital {
				t.Errorf("Expected initial_capital %.2f, got %.2f", tt.expectedCapital, config.InitialCapital)
			}
		})
	}
}

/* TestStrategyConfigExtractor_DefaultQtyValue verifies default_qty_value extraction */
func TestStrategyConfigExtractor_DefaultQtyValue(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		expectedQty float64
	}{
		{
			name:        "float qty",
			value:       3.5,
			expectedQty: 3.5,
		},
		{
			name:        "integer qty",
			value:       10,
			expectedQty: 10.0,
		},
		{
			name:        "zero qty",
			value:       0.0,
			expectedQty: 0.0,
		},
		{
			name:        "fractional qty",
			value:       0.25,
			expectedQty: 0.25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewStrategyConfigExtractor()
			call := &ast.CallExpression{
				Callee: &ast.Identifier{Name: "strategy"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Test"},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "default_qty_value"},
								Value: &ast.Literal{Value: tt.value},
							},
						},
					},
				},
			}

			config := extractor.ExtractFromCall(call)

			if config.DefaultQtyValue != tt.expectedQty {
				t.Errorf("Expected default_qty_value %.2f, got %.2f", tt.expectedQty, config.DefaultQtyValue)
			}
		})
	}
}

/* TestStrategyConfigExtractor_DefaultQtyType verifies default_qty_type extraction from identifiers and member expressions */
func TestStrategyConfigExtractor_DefaultQtyType(t *testing.T) {
	tests := []struct {
		name         string
		value        ast.Expression
		expectedType string
	}{
		// Simple identifiers (unprefixed)
		{
			name:         "simple identifier fixed",
			value:        &ast.Identifier{Name: "fixed"},
			expectedType: "fixed",
		},
		{
			name:         "simple identifier cash",
			value:        &ast.Identifier{Name: "cash"},
			expectedType: "cash",
		},
		{
			name:         "simple identifier percent_of_equity",
			value:        &ast.Identifier{Name: "percent_of_equity"},
			expectedType: "percent_of_equity",
		},
		// Member expressions (strategy.* prefix)
		{
			name: "member expression strategy.fixed",
			value: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "fixed"},
			},
			expectedType: "strategy.fixed",
		},
		{
			name: "member expression strategy.cash",
			value: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "cash"},
			},
			expectedType: "strategy.cash",
		},
		{
			name: "member expression strategy.percent_of_equity",
			value: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "percent_of_equity"},
			},
			expectedType: "strategy.percent_of_equity",
		},
		// Edge cases - testing parser behavior with invalid inputs
		{
			name:         "string literal (invalid - should not parse as identifier)",
			value:        &ast.Literal{Value: "fixed"},
			expectedType: "",
		},
		{
			name:         "string literal strategy.cash (invalid - should not parse as identifier)",
			value:        &ast.Literal{Value: "strategy.cash"},
			expectedType: "",
		},
		// Empty/invalid cases
		{
			name:         "empty identifier",
			value:        &ast.Identifier{Name: ""},
			expectedType: "",
		},
		{
			name: "invalid member expression - non-identifier object",
			value: &ast.MemberExpression{
				Object:   &ast.Literal{Value: 42},
				Property: &ast.Identifier{Name: "cash"},
			},
			expectedType: "",
		},
		{
			name:         "numeric literal (invalid)",
			value:        &ast.Literal{Value: 100},
			expectedType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewStrategyConfigExtractor()
			call := &ast.CallExpression{
				Callee: &ast.Identifier{Name: "strategy"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Test"},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "default_qty_type"},
								Value: tt.value,
							},
						},
					},
				},
			}

			config := extractor.ExtractFromCall(call)

			if config.DefaultQtyType != tt.expectedType {
				t.Errorf("Expected default_qty_type '%s', got '%s'", tt.expectedType, config.DefaultQtyType)
			}
		})
	}
}

/* TestStrategyConfigExtractor_AllProperties verifies extraction of all properties together */
func TestStrategyConfigExtractor_AllProperties(t *testing.T) {
	extractor := NewStrategyConfigExtractor()
	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "strategy"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "Comprehensive Strategy"},
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key:   &ast.Identifier{Name: "initial_capital"},
						Value: &ast.Literal{Value: 75000.0},
					},
					{
						Key:   &ast.Identifier{Name: "default_qty_value"},
						Value: &ast.Literal{Value: 5.0},
					},
					{
						Key:   &ast.Identifier{Name: "default_qty_type"},
						Value: &ast.Identifier{Name: "fixed"},
					},
				},
			},
		},
	}

	config := extractor.ExtractFromCall(call)

	if config.Name != "Comprehensive Strategy" {
		t.Errorf("Expected name 'Comprehensive Strategy', got '%s'", config.Name)
	}
	if config.InitialCapital != 75000.0 {
		t.Errorf("Expected initial_capital 75000.0, got %.2f", config.InitialCapital)
	}
	if config.DefaultQtyValue != 5.0 {
		t.Errorf("Expected default_qty_value 5.0, got %.2f", config.DefaultQtyValue)
	}
	if config.DefaultQtyType != "fixed" {
		t.Errorf("Expected default_qty_type 'fixed', got '%s'", config.DefaultQtyType)
	}
}

/* TestStrategyConfigExtractor_MultipleObjectExpressions verifies handling of multiple config objects */
func TestStrategyConfigExtractor_MultipleObjectExpressions(t *testing.T) {
	extractor := NewStrategyConfigExtractor()
	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "strategy"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "Test"},
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key:   &ast.Identifier{Name: "initial_capital"},
						Value: &ast.Literal{Value: 20000.0},
					},
				},
			},
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key:   &ast.Identifier{Name: "default_qty_value"},
						Value: &ast.Literal{Value: 2.0},
					},
				},
			},
		},
	}

	config := extractor.ExtractFromCall(call)

	if config.InitialCapital != 20000.0 {
		t.Errorf("Expected initial_capital 20000.0 from first object, got %.2f", config.InitialCapital)
	}
	if config.DefaultQtyValue != 2.0 {
		t.Errorf("Expected default_qty_value 2.0 from second object, got %.2f", config.DefaultQtyValue)
	}
}

/* TestStrategyConfigExtractor_EmptyObjectExpression verifies handling of empty config object */
func TestStrategyConfigExtractor_EmptyObjectExpression(t *testing.T) {
	extractor := NewStrategyConfigExtractor()
	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "strategy"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "Empty Config"},
			&ast.ObjectExpression{
				Properties: []ast.Property{},
			},
		},
	}

	config := extractor.ExtractFromCall(call)

	if config.Name != "Empty Config" {
		t.Errorf("Expected name 'Empty Config', got '%s'", config.Name)
	}
	if config.InitialCapital != defaultInitialCapital {
		t.Errorf("Expected default capital, got %.2f", config.InitialCapital)
	}
	if config.DefaultQtyValue != defaultQtyValue {
		t.Errorf("Expected default qty, got %.2f", config.DefaultQtyValue)
	}
}

/* TestStrategyConfigExtractor_IrrelevantProperties verifies ignoring of unknown properties */
func TestStrategyConfigExtractor_IrrelevantProperties(t *testing.T) {
	extractor := NewStrategyConfigExtractor()
	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "strategy"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "Test"},
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key:   &ast.Identifier{Name: "overlay"},
						Value: &ast.Literal{Value: true},
					},
					{
						Key:   &ast.Identifier{Name: "precision"},
						Value: &ast.Literal{Value: 2},
					},
					{
						Key:   &ast.Identifier{Name: "initial_capital"},
						Value: &ast.Literal{Value: 15000.0},
					},
				},
			},
		},
	}

	config := extractor.ExtractFromCall(call)

	if config.InitialCapital != 15000.0 {
		t.Errorf("Expected initial_capital 15000.0, got %.2f", config.InitialCapital)
	}
	if config.DefaultQtyValue != defaultQtyValue {
		t.Errorf("Irrelevant properties should not affect defaults, got %.2f", config.DefaultQtyValue)
	}
}

/* TestStrategyConfigExtractor_MixedArgumentTypes verifies handling of non-object arguments */
func TestStrategyConfigExtractor_MixedArgumentTypes(t *testing.T) {
	extractor := NewStrategyConfigExtractor()
	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "strategy"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "Mixed Args"},
			&ast.Identifier{Name: "someVar"},
			&ast.ObjectExpression{
				Properties: []ast.Property{
					{
						Key:   &ast.Identifier{Name: "default_qty_value"},
						Value: &ast.Literal{Value: 4.0},
					},
				},
			},
			&ast.Literal{Value: 123},
		},
	}

	config := extractor.ExtractFromCall(call)

	if config.Name != "Mixed Args" {
		t.Errorf("Expected name 'Mixed Args', got '%s'", config.Name)
	}
	if config.DefaultQtyValue != 4.0 {
		t.Errorf("Expected default_qty_value 4.0 from object expression, got %.2f", config.DefaultQtyValue)
	}
}

/* TestStrategyConfigExtractor_CommissionType verifies commission_type extraction */
func TestStrategyConfigExtractor_CommissionType(t *testing.T) {
	tests := []struct {
		name                   string
		commissionTypeExpr     ast.Expression
		expectedCommissionType string
	}{
		{
			name: "strategy.commission.percent (3-level)",
			commissionTypeExpr: &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "commission"},
				},
				Property: &ast.Identifier{Name: "percent"},
			},
			expectedCommissionType: "percent",
		},
		{
			name: "strategy.commission.cash_per_order (3-level)",
			commissionTypeExpr: &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "commission"},
				},
				Property: &ast.Identifier{Name: "cash_per_order"},
			},
			expectedCommissionType: "cash_per_order",
		},
		{
			name: "strategy.commission.cash_per_contract (3-level)",
			commissionTypeExpr: &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "commission"},
				},
				Property: &ast.Identifier{Name: "cash_per_contract"},
			},
			expectedCommissionType: "cash_per_contract",
		},
		{
			name: "commission.percent (2-level)",
			commissionTypeExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "commission"},
				Property: &ast.Identifier{Name: "percent"},
			},
			expectedCommissionType: "percent",
		},
		{
			name:                   "bare identifier percent",
			commissionTypeExpr:     &ast.Identifier{Name: "percent"},
			expectedCommissionType: "percent",
		},
		{
			name:                   "bare identifier cash_per_order",
			commissionTypeExpr:     &ast.Identifier{Name: "cash_per_order"},
			expectedCommissionType: "cash_per_order",
		},
		{
			name:                   "unknown identifier passes through",
			commissionTypeExpr:     &ast.Identifier{Name: "my_custom_commission"},
			expectedCommissionType: "my_custom_commission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewStrategyConfigExtractor()
			call := &ast.CallExpression{
				Callee: &ast.Identifier{Name: "strategy"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Commission Test"},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "commission_type"},
								Value: tt.commissionTypeExpr,
							},
						},
					},
				},
			}

			config := extractor.ExtractFromCall(call)

			if config.CommissionType != tt.expectedCommissionType {
				t.Errorf("Expected commission_type '%s', got '%s'", tt.expectedCommissionType, config.CommissionType)
			}
		})
	}
}

/* TestStrategyConfigExtractor_CommissionValue verifies commission_value extraction */
func TestStrategyConfigExtractor_CommissionValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected float64
	}{
		{"positive fractional", 0.1, 0.1},
		{"positive whole", 5.0, 5.0},
		{"large value", 100.0, 100.0},
		{"zero stays default", 0.0, defaultCommissionValue},
		{"negative extracted as-is", -1.0, -1.0},
		{"string not extracted", "0.1", defaultCommissionValue},
		{"nil not extracted", nil, defaultCommissionValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewStrategyConfigExtractor()
			call := &ast.CallExpression{
				Callee: &ast.Identifier{Name: "strategy"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Commission Value Test"},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{
								Key:   &ast.Identifier{Name: "commission_value"},
								Value: &ast.Literal{Value: tt.value},
							},
						},
					},
				},
			}

			config := extractor.ExtractFromCall(call)

			if config.CommissionValue != tt.expected {
				t.Errorf("commission_value: got %.4f, want %.4f", config.CommissionValue, tt.expected)
			}
		})
	}
}

/* TestStrategyConfigExtractor_CommissionTypeAndValue verifies both fields extracted together */
func TestStrategyConfigExtractor_CommissionTypeAndValue(t *testing.T) {
	tests := []struct {
		name         string
		typeExpr     ast.Expression
		valueExpr    ast.Expression
		expectedType string
		expectedVal  float64
	}{
		{
			name: "percent with value",
			typeExpr: &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "commission"},
				},
				Property: &ast.Identifier{Name: "percent"},
			},
			valueExpr:    &ast.Literal{Value: 0.1},
			expectedType: "percent",
			expectedVal:  0.1,
		},
		{
			name: "cash_per_order with value",
			typeExpr: &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "commission"},
				},
				Property: &ast.Identifier{Name: "cash_per_order"},
			},
			valueExpr:    &ast.Literal{Value: 1.5},
			expectedType: "cash_per_order",
			expectedVal:  1.5,
		},
		{
			name: "cash_per_contract with value",
			typeExpr: &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "commission"},
				},
				Property: &ast.Identifier{Name: "cash_per_contract"},
			},
			valueExpr:    &ast.Literal{Value: 0.5},
			expectedType: "cash_per_contract",
			expectedVal:  0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewStrategyConfigExtractor()
			call := &ast.CallExpression{
				Callee: &ast.Identifier{Name: "strategy"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Commission Combined Test"},
					&ast.ObjectExpression{
						Properties: []ast.Property{
							{Key: &ast.Identifier{Name: "commission_type"}, Value: tt.typeExpr},
							{Key: &ast.Identifier{Name: "commission_value"}, Value: tt.valueExpr},
						},
					},
				},
			}

			config := extractor.ExtractFromCall(call)

			if config.CommissionType != tt.expectedType {
				t.Errorf("commission_type: got %q, want %q", config.CommissionType, tt.expectedType)
			}
			if config.CommissionValue != tt.expectedVal {
				t.Errorf("commission_value: got %.4f, want %.4f", config.CommissionValue, tt.expectedVal)
			}
		})
	}
}

/* TestNormalizeCommissionType verifies all Pine commission identifier forms */
func TestNormalizeCommissionType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"strategy.commission.percent", "percent"},
		{"commission.percent", "percent"},
		{"percent", "percent"},
		{"strategy.commission.cash_per_order", "cash_per_order"},
		{"commission.cash_per_order", "cash_per_order"},
		{"cash_per_order", "cash_per_order"},
		{"strategy.commission.cash_per_contract", "cash_per_contract"},
		{"commission.cash_per_contract", "cash_per_contract"},
		{"cash_per_contract", "cash_per_contract"},
		{"unknown_type", "unknown_type"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeCommissionType(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeCommissionType(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
