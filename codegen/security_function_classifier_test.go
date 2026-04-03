package codegen

import "testing"

/* Validates security function name recognition */
func TestIsSecurityFunction(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		expected bool
	}{
		{"bare security", "security", true},
		{"namespaced request.security", "request.security", true},
		{"empty string", "", false},
		{"wrong property request.seed", "request.seed", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := IsSecurityFunction(tt.funcName); result != tt.expected {
				t.Errorf("IsSecurityFunction(%q) = %v, want %v", tt.funcName, result, tt.expected)
			}
		})
	}
}
