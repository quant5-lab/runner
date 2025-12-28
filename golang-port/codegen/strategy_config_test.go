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
}

/* TestStrategyConfig_MergeFrom_PartialMerge verifies selective field merging */
func TestStrategyConfig_MergeFrom_PartialMerge(t *testing.T) {
	config := NewStrategyConfig()
	config.Name = "Original Name"

	other := &StrategyConfig{
		Name:            "Generated Strategy", // Should not merge
		InitialCapital:  30000.0,              // Should merge
		DefaultQtyValue: 0.0,                  // Should not merge
	}

	config.MergeFrom(other)

	if config.Name != "Original Name" {
		t.Errorf("Name should not be overwritten by default name, got '%s'", config.Name)
	}
	if config.InitialCapital != 30000.0 {
		t.Errorf("Expected initial_capital 30000.0, got %.2f", config.InitialCapital)
	}
	if config.DefaultQtyValue != defaultQtyValue {
		t.Errorf("Zero qty should not merge, expected %.2f, got %.2f", defaultQtyValue, config.DefaultQtyValue)
	}
}

/* TestStrategyConfig_MergeFrom_Idempotency verifies repeated merges behave correctly */
func TestStrategyConfig_MergeFrom_Idempotency(t *testing.T) {
	config := NewStrategyConfig()
	other := &StrategyConfig{
		Name:            "Test Strategy",
		InitialCapital:  15000.0,
		DefaultQtyValue: 2.0,
	}

	config.MergeFrom(other)
	firstMergeName := config.Name
	firstMergeCapital := config.InitialCapital

	config.MergeFrom(other)

	if config.Name != firstMergeName {
		t.Error("Second merge should not change already merged name")
	}
	if config.InitialCapital != firstMergeCapital {
		t.Error("Second merge should not change already merged capital")
	}
}
