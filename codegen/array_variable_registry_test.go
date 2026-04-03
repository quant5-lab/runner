package codegen

import "testing"

func TestArrayVariableRegistry_RegisterAndLookup(t *testing.T) {
	registry := NewArrayVariableRegistry()

	registry.Register("prices", ArrayElementFloat64)
	registry.Register("labels", ArrayElementString)

	tests := []struct {
		name         string
		varName      string
		expectedType ArrayElementType
		expectedOk   bool
	}{
		{"float array found", "prices", ArrayElementFloat64, true},
		{"string array found", "labels", ArrayElementString, true},
		{"unregistered variable not found", "unknown", 0, false},
		{"empty name not found", "", 0, false},
		{"case sensitive mismatch", "Prices", 0, false},
		{"partial match not found", "price", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotOk := registry.Lookup(tt.varName)
			if gotOk != tt.expectedOk {
				t.Errorf("Lookup(%q) ok = %v, want %v", tt.varName, gotOk, tt.expectedOk)
			}
			if gotType != tt.expectedType {
				t.Errorf("Lookup(%q) type = %v, want %v", tt.varName, gotType, tt.expectedType)
			}
		})
	}
}

func TestArrayVariableRegistry_IsArrayVariable(t *testing.T) {
	registry := NewArrayVariableRegistry()

	registry.Register("prices", ArrayElementFloat64)
	registry.Register("labels", ArrayElementString)

	tests := []struct {
		name     string
		varName  string
		expected bool
	}{
		{"float array is array variable", "prices", true},
		{"string array is array variable", "labels", true},
		{"unregistered variable is not array", "unknown", false},
		{"empty name is not array", "", false},
		{"case sensitive check", "Prices", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := registry.IsArrayVariable(tt.varName); got != tt.expected {
				t.Errorf("IsArrayVariable(%q) = %v, want %v", tt.varName, got, tt.expected)
			}
		})
	}
}

func TestArrayVariableRegistry_ReregisterOverwrites(t *testing.T) {
	registry := NewArrayVariableRegistry()

	registry.Register("data", ArrayElementFloat64)
	gotType, ok := registry.Lookup("data")
	if !ok || gotType != ArrayElementFloat64 {
		t.Fatalf("Initial register failed: got (%v, %v), want (ArrayElementFloat64, true)", gotType, ok)
	}

	registry.Register("data", ArrayElementString)
	gotType, ok = registry.Lookup("data")
	if !ok || gotType != ArrayElementString {
		t.Errorf("Reregister did not overwrite: got (%v, %v), want (ArrayElementString, true)", gotType, ok)
	}
}

func TestArrayVariableRegistry_EmptyRegistry(t *testing.T) {
	registry := NewArrayVariableRegistry()

	tests := []string{"prices", "labels", "", "unknown"}
	for _, varName := range tests {
		t.Run(varName, func(t *testing.T) {
			gotType, ok := registry.Lookup(varName)
			if ok {
				t.Errorf("Lookup(%q) in empty registry returned ok=true", varName)
			}
			if gotType != 0 {
				t.Errorf("Lookup(%q) in empty registry returned type=%v, want 0", varName, gotType)
			}
			if registry.IsArrayVariable(varName) {
				t.Errorf("IsArrayVariable(%q) in empty registry returned true", varName)
			}
		})
	}
}

func TestArrayVariableRegistry_MixedTypes(t *testing.T) {
	registry := NewArrayVariableRegistry()

	floatVars := []string{"prices", "volumes", "indicators"}
	for _, v := range floatVars {
		registry.Register(v, ArrayElementFloat64)
	}

	stringVars := []string{"labels", "tooltips", "names"}
	for _, v := range stringVars {
		registry.Register(v, ArrayElementString)
	}

	for _, v := range floatVars {
		t.Run("float_"+v, func(t *testing.T) {
			gotType, ok := registry.Lookup(v)
			if !ok {
				t.Errorf("Lookup(%q) not found", v)
			}
			if gotType != ArrayElementFloat64 {
				t.Errorf("Lookup(%q) = %v, want ArrayElementFloat64", v, gotType)
			}
		})
	}

	for _, v := range stringVars {
		t.Run("string_"+v, func(t *testing.T) {
			gotType, ok := registry.Lookup(v)
			if !ok {
				t.Errorf("Lookup(%q) not found", v)
			}
			if gotType != ArrayElementString {
				t.Errorf("Lookup(%q) = %v, want ArrayElementString", v, gotType)
			}
		})
	}
}
