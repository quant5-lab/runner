package codegen

import (
	"testing"
)

func TestConstantRegistry_Register(t *testing.T) {
	tests := []struct {
		name      string
		constName string
		value     interface{}
	}{
		{
			name:      "register bool constant",
			constName: "enabled",
			value:     true,
		},
		{
			name:      "register float constant",
			constName: "multiplier",
			value:     1.5,
		},
		{
			name:      "register int constant",
			constName: "length",
			value:     20,
		},
		{
			name:      "register string constant",
			constName: "symbol",
			value:     "BTCUSDT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewConstantRegistry()
			registry.Register(tt.constName, tt.value)

			if !registry.IsConstant(tt.constName) {
				t.Errorf("constant %q not registered", tt.constName)
			}

			retrieved, exists := registry.Get(tt.constName)
			if !exists {
				t.Fatalf("failed to retrieve constant %q", tt.constName)
			}

			if retrieved != tt.value {
				t.Errorf("expected value %v, got %v", tt.value, retrieved)
			}
		})
	}
}

func TestConstantRegistry_ExtractFromGeneratedCode_Bool(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected interface{}
	}{
		{
			name:     "extract true",
			code:     "const enabled = true\n",
			expected: true,
		},
		{
			name:     "extract false",
			code:     "const showTrades = false\n",
			expected: false,
		},
		{
			name:     "malformed bool constant matches bool pattern",
			code:     "const invalid = truefalse\n",
			expected: true, // fmt.Sscanf parses "true" prefix successfully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewConstantRegistry()
			result := registry.ExtractFromGeneratedCode(tt.code)

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestConstantRegistry_ExtractFromGeneratedCode_Float(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected interface{}
	}{
		{
			name:     "extract float with decimals",
			code:     "const multiplier = 1.50\n",
			expected: 1.5,
		},
		{
			name:     "extract float zero",
			code:     "const factor = 0.00\n",
			expected: 0.0,
		},
		{
			name:     "malformed float matches float pattern (partial parse)",
			code:     "const invalid = 1.5.0\n",
			expected: 1.5, // fmt.Sscanf parses "1.5" successfully, stops at second dot
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewConstantRegistry()
			result := registry.ExtractFromGeneratedCode(tt.code)

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestConstantRegistry_ExtractFromGeneratedCode_Int(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected float64 // fmt.Sscanf tries %f before %d, so ints parse as floats
	}{
		{
			name:     "extract positive int parsed as float",
			code:     "const length = 20\n",
			expected: 20.0,
		},
		{
			name:     "extract zero int parsed as float",
			code:     "const period = 0\n",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewConstantRegistry()
			result := registry.ExtractFromGeneratedCode(tt.code)

			if resultFloat, ok := result.(float64); ok {
				if resultFloat != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, resultFloat)
				}
			} else {
				t.Errorf("expected float64, got %T", result)
			}
		})
	}
}

func TestConstantRegistry_ExtractFromGeneratedCode_String(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected interface{}
	}{
		{
			name:     "string literal attempts bool parse (returns false for strings)",
			code:     `const symbol = "BTCUSDT"` + "\n",
			expected: false, // Sscanf tries %t first, parses "BTCUSDT" as false
		},
		{
			name:     "empty string attempts bool parse",
			code:     `const empty = ""` + "\n",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewConstantRegistry()
			result := registry.ExtractFromGeneratedCode(tt.code)

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestConstantRegistry_IsBoolConstant(t *testing.T) {
	registry := NewConstantRegistry()
	registry.Register("enabled", true)
	registry.Register("multiplier", 1.5)
	registry.Register("length", 20)

	tests := []struct {
		name     string
		expected bool
	}{
		{"enabled", true},
		{"multiplier", false},
		{"length", false},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.IsBoolConstant(tt.name)
			if result != tt.expected {
				t.Errorf("IsBoolConstant(%q) expected %v, got %v", tt.name, tt.expected, result)
			}
		})
	}
}

func TestConstantRegistry_GetAll(t *testing.T) {
	registry := NewConstantRegistry()
	registry.Register("enabled", true)
	registry.Register("multiplier", 1.5)
	registry.Register("length", 20)

	all := registry.GetAll()
	if len(all) != 3 {
		t.Errorf("expected 3 constants, got %d", len(all))
	}

	if _, ok := all["enabled"].(bool); !ok {
		t.Errorf("enabled type mismatch")
	}
	if _, ok := all["multiplier"].(float64); !ok {
		t.Errorf("multiplier type mismatch")
	}
	if _, ok := all["length"].(int); !ok {
		t.Errorf("length type mismatch")
	}
}

func TestConstantRegistry_Count(t *testing.T) {
	registry := NewConstantRegistry()

	if registry.Count() != 0 {
		t.Errorf("expected empty registry, got count %d", registry.Count())
	}

	registry.Register("enabled", true)
	if registry.Count() != 1 {
		t.Errorf("expected count 1, got %d", registry.Count())
	}

	registry.Register("multiplier", 1.5)
	registry.Register("length", 20)
	if registry.Count() != 3 {
		t.Errorf("expected count 3, got %d", registry.Count())
	}
}

func TestConstantRegistry_EdgeCases(t *testing.T) {
	t.Run("Get non-existent constant returns nil", func(t *testing.T) {
		registry := NewConstantRegistry()
		result, exists := registry.Get("nonexistent")
		if exists || result != nil {
			t.Errorf("expected (nil, false), got (%v, %v)", result, exists)
		}
	})

	t.Run("IsConstant with non-existent constant returns false", func(t *testing.T) {
		registry := NewConstantRegistry()
		result := registry.IsConstant("nonexistent")
		if result {
			t.Error("expected false for non-existent constant")
		}
	})

	t.Run("ExtractFromGeneratedCode with empty string returns nil", func(t *testing.T) {
		registry := NewConstantRegistry()
		result := registry.ExtractFromGeneratedCode("")
		if result != nil {
			t.Errorf("expected nil for empty code, got %v", result)
		}
	})

	t.Run("ExtractFromGeneratedCode with malformed const returns nil", func(t *testing.T) {
		registry := NewConstantRegistry()
		result := registry.ExtractFromGeneratedCode("const malformed\n")
		if result != nil {
			t.Errorf("expected nil for malformed const, got %v", result)
		}
	})

	t.Run("Register duplicate constant overwrites", func(t *testing.T) {
		registry := NewConstantRegistry()
		registry.Register("value", 1.0)
		registry.Register("value", 2.0)

		constant, _ := registry.Get("value")
		if constant.(float64) != 2.0 {
			t.Errorf("expected overwritten value 2.0, got %v", constant)
		}
	})
}

func TestConstantRegistry_Integration_MultipleTypes(t *testing.T) {
	registry := NewConstantRegistry()

	registry.Register("enabled", true)
	registry.Register("multiplier", 1.5)
	registry.Register("length", 20)

	if registry.Count() != 3 {
		t.Errorf("expected 3 constants, got %d", registry.Count())
	}

	tests := []struct {
		name  string
		value interface{}
	}{
		{"enabled", true},
		{"multiplier", 1.5},
		{"length", 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constant, exists := registry.Get(tt.name)
			if !exists {
				t.Fatalf("constant %q not found", tt.name)
			}
			if constant != tt.value {
				t.Errorf("expected value %v, got %v", tt.value, constant)
			}
		})
	}
}
