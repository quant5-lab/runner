package visual

import (
	"math"
	"testing"
)

func TestPineColorNew(t *testing.T) {
	tests := []struct {
		name     string
		baseHex  string
		transp   float64
		expected string
	}{
		{"fully opaque", "#FF5252", 0, "#FF5252FF"},
		{"fully transparent", "#FF5252", 100, "#FF525200"},
		{"50% transparent", "#FF5252", 50, "#FF525280"},
		{"with 8-digit input replaces alpha", "#FF525280", 0, "#FF5252FF"},
		{"blue opaque", "#2962FF", 0, "#2962FFFF"},
		{"lime 70% transp", "#00E676", 70, "#00E6764D"},
		{"negative transp clamps to 0", "#FFFFFF", -10, "#FFFFFFFF"},
		{"over 100 transp clamps to 100", "#FFFFFF", 150, "#FFFFFF00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PineColorNew(tt.baseHex, tt.transp)
			if got != tt.expected {
				t.Errorf("PineColorNew(%q, %v) = %q, want %q", tt.baseHex, tt.transp, got, tt.expected)
			}
		})
	}
}

func TestPineColorRGB(t *testing.T) {
	tests := []struct {
		name     string
		r, g, b  float64
		transp   float64
		expected string
	}{
		{"pure red opaque", 255, 0, 0, 0, "#FF0000FF"},
		{"pure green opaque", 0, 255, 0, 0, "#00FF00FF"},
		{"pure blue opaque", 0, 0, 255, 0, "#0000FFFF"},
		{"white transparent", 255, 255, 255, 100, "#FFFFFF00"},
		{"mid gray 50%", 128, 128, 128, 50, "#80808080"},
		{"clamp overflow", 300, -10, 255, 0, "#FF00FFFF"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PineColorRGB(tt.r, tt.g, tt.b, tt.transp)
			if got != tt.expected {
				t.Errorf("PineColorRGB(%v, %v, %v, %v) = %q, want %q",
					tt.r, tt.g, tt.b, tt.transp, got, tt.expected)
			}
		})
	}
}

func TestPineColorComponents(t *testing.T) {
	t.Run("R from 6-digit", func(t *testing.T) {
		if got := PineColorR("#FF5252"); got != 255 {
			t.Errorf("PineColorR(#FF5252) = %v, want 255", got)
		}
	})
	t.Run("G from 6-digit", func(t *testing.T) {
		if got := PineColorG("#FF5252"); got != 82 {
			t.Errorf("PineColorG(#FF5252) = %v, want 82", got)
		}
	})
	t.Run("B from 6-digit", func(t *testing.T) {
		if got := PineColorB("#FF5252"); got != 82 {
			t.Errorf("PineColorB(#FF5252) = %v, want 82", got)
		}
	})
	t.Run("T from 6-digit (opaque)", func(t *testing.T) {
		if got := PineColorT("#FF5252"); got != 0 {
			t.Errorf("PineColorT(#FF5252) = %v, want 0", got)
		}
	})
	t.Run("T from 8-digit fully transparent", func(t *testing.T) {
		if got := PineColorT("#FF525200"); got != 100 {
			t.Errorf("PineColorT(#FF525200) = %v, want 100", got)
		}
	})
	t.Run("T from 8-digit half transparent", func(t *testing.T) {
		got := PineColorT("#FF525280")
		if math.Abs(got-50) > 1 {
			t.Errorf("PineColorT(#FF525280) = %v, want ~50", got)
		}
	})
	t.Run("R from 8-digit", func(t *testing.T) {
		if got := PineColorR("#2962FF80"); got != 41 {
			t.Errorf("PineColorR(#2962FF80) = %v, want 41", got)
		}
	})
}

func TestPineColorFromGradient(t *testing.T) {
	tests := []struct {
		name        string
		value       float64
		bottomValue float64
		topValue    float64
		bottomColor string
		topColor    string
		expected    string
	}{
		{"at bottom", 0, 0, 100, "#000000", "#FFFFFF", "#000000FF"},
		{"at top", 100, 0, 100, "#000000", "#FFFFFF", "#FFFFFFFF"},
		{"at midpoint", 50, 0, 100, "#000000", "#FFFFFF", "#808080FF"},
		{"below range clamps", -50, 0, 100, "#000000", "#FFFFFF", "#000000FF"},
		{"above range clamps", 200, 0, 100, "#000000", "#FFFFFF", "#FFFFFFFF"},
		{"equal range returns bottom", 50, 50, 50, "#FF0000", "#00FF00", "#FF0000FF"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PineColorFromGradient(tt.value, tt.bottomValue, tt.topValue, tt.bottomColor, tt.topColor)
			if got != tt.expected {
				t.Errorf("PineColorFromGradient(%v, %v, %v, %q, %q) = %q, want %q",
					tt.value, tt.bottomValue, tt.topValue, tt.bottomColor, tt.topColor, got, tt.expected)
			}
		})
	}
}

func TestPineColorNewRoundTrip(t *testing.T) {
	base := "#FF5252"
	transp := 30.0
	result := PineColorNew(base, transp)
	gotT := PineColorT(result)
	if math.Abs(gotT-transp) > 1 {
		t.Errorf("Round-trip: PineColorNew(%q, %v) → %q → PineColorT = %v, want ~%v",
			base, transp, result, gotT, transp)
	}
	gotR := PineColorR(result)
	if gotR != 255 {
		t.Errorf("Round-trip R: got %v, want 255", gotR)
	}
}

func TestParseHexColorEdgeCases(t *testing.T) {
	t.Run("short hex", func(t *testing.T) {
		r, g, b, a := parseHexColor("#FFF")
		if r != 0 || g != 0 || b != 0 || a != 255 {
			t.Errorf("Short hex should default to 0,0,0,255 — got %d,%d,%d,%d", r, g, b, a)
		}
	})
	t.Run("no hash prefix", func(t *testing.T) {
		r, _, _, _ := parseHexColor("FF5252")
		if r != 255 {
			t.Errorf("Without # prefix: R = %d, want 255", r)
		}
	})
	t.Run("lowercase hex", func(t *testing.T) {
		r, g, b, _ := parseHexColor("#ff5252")
		if r != 255 || g != 82 || b != 82 {
			t.Errorf("Lowercase hex: got %d,%d,%d — want 255,82,82", r, g, b)
		}
	})
}
