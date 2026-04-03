package codegen

import "testing"

func TestIsColorFunction(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		expected bool
	}{
		{"color.new", "color.new", true},
		{"color.rgb", "color.rgb", true},
		{"color.from_gradient", "color.from_gradient", true},
		{"color.r", "color.r", true},
		{"color.g", "color.g", true},
		{"color.b", "color.b", true},
		{"color.t", "color.t", true},
		{"ta.sma is not color", "ta.sma", false},
		{"math.abs is not color", "math.abs", false},
		{"input.bool is not color", "input.bool", false},
		{"unknown color property", "color.unknown", false},
		{"bare color prefix", "color", false},
		{"empty string", "", false},
		{"partial match", "color.ne", false},
		{"extra suffix", "color.new.extra", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsColorFunction(tt.funcName); got != tt.expected {
				t.Errorf("IsColorFunction(%q) = %v, want %v", tt.funcName, got, tt.expected)
			}
		})
	}
}

func TestColorFunctionReturnType(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		expected string
	}{
		{"color.new returns string", "color.new", "string"},
		{"color.rgb returns string", "color.rgb", "string"},
		{"color.from_gradient returns string", "color.from_gradient", "string"},
		{"color.r returns float64", "color.r", "float64"},
		{"color.g returns float64", "color.g", "float64"},
		{"color.b returns float64", "color.b", "float64"},
		{"color.t returns float64", "color.t", "float64"},
		{"unknown function returns empty", "ta.sma", ""},
		{"empty string returns empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ColorFunctionReturnType(tt.funcName); got != tt.expected {
				t.Errorf("ColorFunctionReturnType(%q) = %q, want %q", tt.funcName, got, tt.expected)
			}
		})
	}
}

func TestColorFunctionGoName(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		expected string
	}{
		{"color.new", "color.new", "visual.PineColorNew"},
		{"color.rgb", "color.rgb", "visual.PineColorRGB"},
		{"color.from_gradient", "color.from_gradient", "visual.PineColorFromGradient"},
		{"color.r", "color.r", "visual.PineColorR"},
		{"color.g", "color.g", "visual.PineColorG"},
		{"color.b", "color.b", "visual.PineColorB"},
		{"color.t", "color.t", "visual.PineColorT"},
		{"unknown function returns empty", "ta.sma", ""},
		{"empty string returns empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ColorFunctionGoName(tt.funcName); got != tt.expected {
				t.Errorf("ColorFunctionGoName(%q) = %q, want %q", tt.funcName, got, tt.expected)
			}
		})
	}
}

func TestColorFunctionRegistry_Consistency(t *testing.T) {
	allFunctions := []string{
		"color.new", "color.rgb", "color.from_gradient",
		"color.r", "color.g", "color.b", "color.t",
	}

	for _, name := range allFunctions {
		t.Run(name, func(t *testing.T) {
			if !IsColorFunction(name) {
				t.Fatal("IsColorFunction returned false")
			}
			if ColorFunctionReturnType(name) == "" {
				t.Error("ColorFunctionReturnType returned empty")
			}
			if ColorFunctionGoName(name) == "" {
				t.Error("ColorFunctionGoName returned empty")
			}
		})
	}
}
