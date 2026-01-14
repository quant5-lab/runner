package codegen

import "testing"

func TestPeriodClassifier_CompileTimeConstant(t *testing.T) {
	classifier := NewPeriodClassifier()

	tests := []struct {
		name        string
		periodValue int
		periodExpr  string
		expected    PeriodType
	}{
		{
			name:        "constant period 14",
			periodValue: 14,
			periodExpr:  "",
			expected:    PeriodCompileTimeConstant,
		},
		{
			name:        "constant period 20",
			periodValue: 20,
			periodExpr:  "",
			expected:    PeriodCompileTimeConstant,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.Classify(tt.periodValue, tt.periodExpr)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPeriodClassifier_RuntimeDynamic(t *testing.T) {
	classifier := NewPeriodClassifier()

	tests := []struct {
		name        string
		periodValue int
		periodExpr  string
		expected    PeriodType
	}{
		{
			name:        "runtime expression",
			periodValue: 0,
			periodExpr:  "input.int(14, 'Period')",
			expected:    PeriodRuntimeDynamic,
		},
		{
			name:        "variable reference",
			periodValue: 0,
			periodExpr:  "periodVar",
			expected:    PeriodRuntimeDynamic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.Classify(tt.periodValue, tt.periodExpr)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
