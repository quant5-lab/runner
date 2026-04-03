package codegen

import (
	"testing"
)

/* TestStrategyConfig_NewStrategyConfig verifies default values initialization */
func TestStrategyConfig_NewStrategyConfig(t *testing.T) {
	config := NewStrategyConfig()

	if config.Name != "Generated Strategy" {
		t.Errorf("Expected default name 'Generated Strategy', got '%s'", config.Name)
	}
	if config.InitialCapital != defaultInitialCapital {
		t.Errorf("Expected initial_capital %.2f, got %.2f", defaultInitialCapital, config.InitialCapital)
	}
	if config.DefaultQtyValue != defaultQtyValue {
		t.Errorf("Expected default_qty_value %.2f, got %.2f", defaultQtyValue, config.DefaultQtyValue)
	}
	if config.DefaultQtyType != "" {
		t.Errorf("Expected empty default_qty_type, got '%s'", config.DefaultQtyType)
	}
	if config.CommissionType != "" {
		t.Errorf("Expected empty commission_type, got '%s'", config.CommissionType)
	}
	if config.CommissionValue != defaultCommissionValue {
		t.Errorf("Expected commission_value %.2f, got %.2f", defaultCommissionValue, config.CommissionValue)
	}
	if config.Pyramiding != defaultPyramiding {
		t.Errorf("Expected pyramiding %d, got %d", defaultPyramiding, config.Pyramiding)
	}
}

/* TestStrategyConfig_MergeFrom_NilHandling verifies nil safety */
func TestStrategyConfig_MergeFrom_NilHandling(t *testing.T) {
	config := NewStrategyConfig()
	originalName := config.Name
	originalCapital := config.InitialCapital

	config.MergeFrom(nil)

	if config.Name != originalName {
		t.Error("MergeFrom(nil) should not modify config")
	}
	if config.InitialCapital != originalCapital {
		t.Error("MergeFrom(nil) should not modify config")
	}
}

/* TestStrategyConfig_MergeFrom_NameMerge verifies name merge behavior */
func TestStrategyConfig_MergeFrom_NameMerge(t *testing.T) {
	tests := []struct {
		name         string
		otherName    string
		expectMerge  bool
		expectedName string
	}{
		{
			name:         "custom name merges",
			otherName:    "My Custom Strategy",
			expectMerge:  true,
			expectedName: "My Custom Strategy",
		},
		{
			name:         "default name does not merge",
			otherName:    "Generated Strategy",
			expectMerge:  false,
			expectedName: "Generated Strategy",
		},
		{
			name:         "empty name does not merge",
			otherName:    "",
			expectMerge:  false,
			expectedName: "Generated Strategy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewStrategyConfig()
			other := &StrategyConfig{Name: tt.otherName}

			config.MergeFrom(other)

			if config.Name != tt.expectedName {
				t.Errorf("Expected name '%s', got '%s'", tt.expectedName, config.Name)
			}
		})
	}
}

/* TestStrategyConfig_MergeFrom_InitialCapitalMerge verifies capital merge behavior */
func TestStrategyConfig_MergeFrom_InitialCapitalMerge(t *testing.T) {
	tests := []struct {
		name            string
		otherCapital    float64
		expectMerge     bool
		expectedCapital float64
	}{
		{
			name:            "positive capital merges",
			otherCapital:    50000.0,
			expectMerge:     true,
			expectedCapital: 50000.0,
		},
		{
			name:            "zero capital does not merge",
			otherCapital:    0.0,
			expectMerge:     false,
			expectedCapital: defaultInitialCapital,
		},
		{
			name:            "negative capital does not merge",
			otherCapital:    -1000.0,
			expectMerge:     false,
			expectedCapital: defaultInitialCapital,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewStrategyConfig()
			other := &StrategyConfig{InitialCapital: tt.otherCapital}

			config.MergeFrom(other)

			if config.InitialCapital != tt.expectedCapital {
				t.Errorf("Expected initial_capital %.2f, got %.2f", tt.expectedCapital, config.InitialCapital)
			}
		})
	}
}

/* TestStrategyConfig_MergeFrom_DefaultQtyValueMerge verifies qty value merge behavior */
func TestStrategyConfig_MergeFrom_DefaultQtyValueMerge(t *testing.T) {
	tests := []struct {
		name        string
		otherQty    float64
		expectMerge bool
		expectedQty float64
	}{
		{
			name:        "positive qty merges",
			otherQty:    5.0,
			expectMerge: true,
			expectedQty: 5.0,
		},
		{
			name:        "fractional qty merges",
			otherQty:    0.5,
			expectMerge: true,
			expectedQty: 0.5,
		},
		{
			name:        "zero qty does not merge",
			otherQty:    0.0,
			expectMerge: false,
			expectedQty: defaultQtyValue,
		},
		{
			name:        "negative qty does not merge",
			otherQty:    -2.0,
			expectMerge: false,
			expectedQty: defaultQtyValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewStrategyConfig()
			other := &StrategyConfig{DefaultQtyValue: tt.otherQty}

			config.MergeFrom(other)

			if config.DefaultQtyValue != tt.expectedQty {
				t.Errorf("Expected default_qty_value %.2f, got %.2f", tt.expectedQty, config.DefaultQtyValue)
			}
		})
	}
}

/* TestStrategyConfig_MergeFrom_DefaultQtyTypeMerge verifies qty type merge behavior */
func TestStrategyConfig_MergeFrom_DefaultQtyTypeMerge(t *testing.T) {
	tests := []struct {
		name         string
		otherType    string
		expectMerge  bool
		expectedType string
	}{
		{
			name:         "non-empty type merges",
			otherType:    "strategy.percent_of_equity",
			expectMerge:  true,
			expectedType: "strategy.percent_of_equity",
		},
		{
			name:         "empty type does not merge",
			otherType:    "",
			expectMerge:  false,
			expectedType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewStrategyConfig()
			other := &StrategyConfig{DefaultQtyType: tt.otherType}

			config.MergeFrom(other)

			if config.DefaultQtyType != tt.expectedType {
				t.Errorf("Expected default_qty_type '%s', got '%s'", tt.expectedType, config.DefaultQtyType)
			}
		})
	}
}

/* TestStrategyConfig_MergeFrom_MultipleFields verifies all fields merge together */
func TestStrategyConfig_MergeFrom_MultipleFields(t *testing.T) {
	config := NewStrategyConfig()
	other := &StrategyConfig{
		Name:            "Multi-Field Strategy",
		InitialCapital:  25000.0,
		DefaultQtyValue: 3.0,
		DefaultQtyType:  "strategy.fixed",
		CommissionType:  "percent",
		CommissionValue: 0.1,
	}

	config.MergeFrom(other)

	if config.Name != "Multi-Field Strategy" {
		t.Errorf("Expected name 'Multi-Field Strategy', got '%s'", config.Name)
	}
	if config.InitialCapital != 25000.0 {
		t.Errorf("Expected initial_capital 25000.0, got %.2f", config.InitialCapital)
	}
	if config.DefaultQtyValue != 3.0 {
		t.Errorf("Expected default_qty_value 3.0, got %.2f", config.DefaultQtyValue)
	}
	if config.DefaultQtyType != "strategy.fixed" {
		t.Errorf("Expected default_qty_type 'strategy.fixed', got '%s'", config.DefaultQtyType)
	}
	if config.CommissionType != "percent" {
		t.Errorf("Expected commission_type 'percent', got '%s'", config.CommissionType)
	}
	if config.CommissionValue != 0.1 {
		t.Errorf("Expected commission_value 0.1, got %.4f", config.CommissionValue)
	}
}

/* TestStrategyConfig_MergeFrom_PartialMerge verifies selective field merging */
func TestStrategyConfig_MergeFrom_PartialMerge(t *testing.T) {
	config := NewStrategyConfig()
	config.Name = "Original Name"

	other := &StrategyConfig{
		Name:            "Generated Strategy",
		InitialCapital:  30000.0,
		DefaultQtyValue: 0.0,
	}

	config.MergeFrom(other)

	if config.Name != "Original Name" {
		t.Errorf("Default sentinel name must not overwrite existing, got '%s'", config.Name)
	}
	if config.InitialCapital != 30000.0 {
		t.Errorf("Expected initial_capital 30000.0, got %.2f", config.InitialCapital)
	}
	if config.DefaultQtyValue != defaultQtyValue {
		t.Errorf("Zero qty must not overwrite, expected %.2f, got %.2f", defaultQtyValue, config.DefaultQtyValue)
	}
}

/* TestStrategyConfig_MergeFrom_Idempotency verifies repeated merges behave correctly */
func TestStrategyConfig_MergeFrom_Idempotency(t *testing.T) {
	config := NewStrategyConfig()
	other := &StrategyConfig{
		Name:            "Test Strategy",
		InitialCapital:  15000.0,
		DefaultQtyValue: 2.0,
		CommissionType:  "percent",
		CommissionValue: 0.5,
	}

	config.MergeFrom(other)
	snapshot := *config

	config.MergeFrom(other)

	if config.Name != snapshot.Name {
		t.Error("Second merge must not change already merged name")
	}
	if config.InitialCapital != snapshot.InitialCapital {
		t.Error("Second merge must not change already merged capital")
	}
	if config.CommissionType != snapshot.CommissionType {
		t.Error("Second merge must not change already merged commission_type")
	}
	if config.CommissionValue != snapshot.CommissionValue {
		t.Error("Second merge must not change already merged commission_value")
	}
}

/*
	TestStrategyConfig_MergeFrom_CommissionCascade verifies commission field independence: type and

value update separately, so merging only one field preserves the other on a pre-set base config.
*/
func TestStrategyConfig_MergeFrom_CommissionCascade(t *testing.T) {
	tests := []struct {
		name          string
		baseType      string
		baseValue     float64
		incomingType  string
		incomingValue float64
		expectedType  string
		expectedValue float64
	}{
		{
			name:     "type update preserves existing value",
			baseType: "percent", baseValue: 0.1,
			incomingType: "cash_per_order", incomingValue: 0.0,
			expectedType: "cash_per_order", expectedValue: 0.1,
		},
		{
			name:     "value update preserves existing type",
			baseType: "percent", baseValue: 0.1,
			incomingType: "", incomingValue: 5.0,
			expectedType: "percent", expectedValue: 5.0,
		},
		{
			name:     "both fields update together",
			baseType: "percent", baseValue: 0.1,
			incomingType: "cash_per_contract", incomingValue: 2.0,
			expectedType: "cash_per_contract", expectedValue: 2.0,
		},
		{
			name:     "neither updates when empty type and zero value",
			baseType: "percent", baseValue: 0.1,
			incomingType: "", incomingValue: 0.0,
			expectedType: "percent", expectedValue: 0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewStrategyConfig()
			config.CommissionType = tt.baseType
			config.CommissionValue = tt.baseValue

			other := &StrategyConfig{CommissionType: tt.incomingType, CommissionValue: tt.incomingValue}
			config.MergeFrom(other)

			if config.CommissionType != tt.expectedType {
				t.Errorf("Expected commission_type %q, got %q", tt.expectedType, config.CommissionType)
			}
			if config.CommissionValue != tt.expectedValue {
				t.Errorf("Expected commission_value %.4f, got %.4f", tt.expectedValue, config.CommissionValue)
			}
		})
	}
}

/* TestStrategyConfig_MergeFrom_Commission verifies commission fields merge correctly */
func TestStrategyConfig_MergeFrom_Commission(t *testing.T) {
	tests := []struct {
		name                    string
		otherType               string
		otherValue              float64
		expectedCommissionType  string
		expectedCommissionValue float64
	}{
		{
			name:                    "percent type with value merges",
			otherType:               "percent",
			otherValue:              0.1,
			expectedCommissionType:  "percent",
			expectedCommissionValue: 0.1,
		},
		{
			name:                    "cash_per_order type merges",
			otherType:               "cash_per_order",
			otherValue:              5.0,
			expectedCommissionType:  "cash_per_order",
			expectedCommissionValue: 5.0,
		},
		{
			name:                    "cash_per_contract type merges",
			otherType:               "cash_per_contract",
			otherValue:              2.0,
			expectedCommissionType:  "cash_per_contract",
			expectedCommissionValue: 2.0,
		},
		{
			name:                    "empty type does not overwrite",
			otherType:               "",
			otherValue:              1.0,
			expectedCommissionType:  "",
			expectedCommissionValue: 1.0,
		},
		{
			name:                    "zero value does not overwrite",
			otherType:               "percent",
			otherValue:              0.0,
			expectedCommissionType:  "percent",
			expectedCommissionValue: defaultCommissionValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewStrategyConfig()
			other := &StrategyConfig{CommissionType: tt.otherType, CommissionValue: tt.otherValue}

			config.MergeFrom(other)

			if config.CommissionType != tt.expectedCommissionType {
				t.Errorf("Expected commission_type '%s', got '%s'", tt.expectedCommissionType, config.CommissionType)
			}
			if config.CommissionValue != tt.expectedCommissionValue {
				t.Errorf("Expected commission_value %.4f, got %.4f", tt.expectedCommissionValue, config.CommissionValue)
			}
		})
	}
}
