package codegen

import "testing"

func TestArrayConstructorTypeResolver_ResolveVariableType(t *testing.T) {
	resolver := NewArrayConstructorTypeResolver()

	tests := []struct {
		funcName     string
		expectedType ArrayElementType
		expectedOk   bool
	}{
		{"array.new_float", ArrayElementFloat64, true},
		{"array.new_int", ArrayElementFloat64, true},
		{"array.new_bool", ArrayElementFloat64, true},
		{"array.new_color", ArrayElementFloat64, true},
		{"array.new_string", ArrayElementString, true},
		{"array.from", ArrayElementFloat64, true},
		{"array.new_label", ArrayElementFloat64, true},
		{"array.new_line", ArrayElementFloat64, true},
		{"array.new_box", ArrayElementFloat64, true},
		{"array.new_table", ArrayElementFloat64, true},
		{"array.new_linefill", ArrayElementFloat64, true},
		{"ta.pivot_point_levels", ArrayElementFloat64, true},
		{"pivot_point_levels", ArrayElementFloat64, true},
		{"array.push", 0, false},
		{"array.get", 0, false},
		{"ta.sma", 0, false},
		{"unknown", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			gotType, gotOk := resolver.ResolveVariableType(tt.funcName)
			if gotOk != tt.expectedOk {
				t.Errorf("ResolveVariableType(%q) ok = %v, want %v", tt.funcName, gotOk, tt.expectedOk)
			}
			if gotType != tt.expectedType {
				t.Errorf("ResolveVariableType(%q) type = %v, want %v", tt.funcName, gotType, tt.expectedType)
			}
		})
	}
}
