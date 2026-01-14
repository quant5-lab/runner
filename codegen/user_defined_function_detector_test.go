package codegen

import "testing"

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
