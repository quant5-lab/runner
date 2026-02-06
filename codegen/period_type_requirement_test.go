package codegen

import (
	"testing"
)

func TestPeriodTypeQualifier_AllowsRuntimeDynamic(t *testing.T) {
	tests := []struct {
		name      string
		qualifier PeriodTypeQualifier
		expected  bool
	}{
		{
			name:      "simple int does not allow runtime dynamic",
			qualifier: PeriodSimpleInt,
			expected:  false,
		},
		{
			name:      "series int allows runtime dynamic",
			qualifier: PeriodSeriesInt,
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.qualifier.AllowsRuntimeDynamic(); got != tt.expected {
				t.Errorf("AllowsRuntimeDynamic() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPeriodRequirementRepository_GetSpec(t *testing.T) {
	repo := NewPeriodRequirementRepository()

	tests := []struct {
		name              string
		functionName      string
		expectExists      bool
		expectedQualifier PeriodTypeQualifier
		expectedPosition  int
	}{
		{
			name:              "ta.sma has series int",
			functionName:      "ta.sma",
			expectExists:      true,
			expectedQualifier: PeriodSeriesInt,
			expectedPosition:  1,
		},
		{
			name:              "ta.ema has simple int",
			functionName:      "ta.ema",
			expectExists:      true,
			expectedQualifier: PeriodSimpleInt,
			expectedPosition:  1,
		},
		{
			name:              "ta.rsi has simple int",
			functionName:      "ta.rsi",
			expectExists:      true,
			expectedQualifier: PeriodSimpleInt,
			expectedPosition:  1,
		},
		{
			name:              "ta.stdev has series int",
			functionName:      "ta.stdev",
			expectExists:      true,
			expectedQualifier: PeriodSeriesInt,
			expectedPosition:  1,
		},
		{
			name:              "ta.highest has series int at position 0",
			functionName:      "ta.highest",
			expectExists:      true,
			expectedQualifier: PeriodSeriesInt,
			expectedPosition:  0,
		},
		{
			name:              "ta.atr has simple int at position 0",
			functionName:      "ta.atr",
			expectExists:      true,
			expectedQualifier: PeriodSimpleInt,
			expectedPosition:  0,
		},
		{
			name:         "unknown function",
			functionName: "ta.unknown",
			expectExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, exists := repo.GetSpec(tt.functionName)

			if exists != tt.expectExists {
				t.Fatalf("GetSpec(%s) exists = %v, want %v", tt.functionName, exists, tt.expectExists)
			}

			if tt.expectExists {
				if spec.PeriodQualifier != tt.expectedQualifier {
					t.Errorf("PeriodQualifier = %d, want %d", spec.PeriodQualifier, tt.expectedQualifier)
				}
				if spec.ParameterPosition != tt.expectedPosition {
					t.Errorf("ParameterPosition = %d, want %d", spec.ParameterPosition, tt.expectedPosition)
				}
			}
		})
	}
}

func TestPeriodRequirementRepository_AllowsRuntimeDynamic(t *testing.T) {
	repo := NewPeriodRequirementRepository()

	tests := []struct {
		name         string
		functionName string
		expected     bool
	}{
		{
			name:         "ta.sma allows runtime dynamic",
			functionName: "ta.sma",
			expected:     true,
		},
		{
			name:         "ta.ema does not allow runtime dynamic",
			functionName: "ta.ema",
			expected:     false,
		},
		{
			name:         "ta.stdev allows runtime dynamic",
			functionName: "ta.stdev",
			expected:     true,
		},
		{
			name:         "unknown function returns false",
			functionName: "ta.unknown",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := repo.AllowsRuntimeDynamic(tt.functionName); got != tt.expected {
				t.Errorf("AllowsRuntimeDynamic(%s) = %v, want %v", tt.functionName, got, tt.expected)
			}
		})
	}
}
