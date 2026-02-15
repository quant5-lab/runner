package codegen

import "testing"

func TestSecurityBarFieldExpression_DirectFields(t *testing.T) {
	bar := "secCtx.Data[secBarIdx]"

	tests := []struct {
		field string
		want  string
	}{
		{"close", bar + ".Close"},
		{"open", bar + ".Open"},
		{"high", bar + ".High"},
		{"low", bar + ".Low"},
		{"volume", bar + ".Volume"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			got, ok := SecurityBarFieldExpression(tt.field, bar)
			if !ok {
				t.Fatalf("SecurityBarFieldExpression(%q) returned ok=false", tt.field)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSecurityBarFieldExpression_DerivedPrices(t *testing.T) {
	bar := "b"

	tests := []struct {
		field string
		want  string
	}{
		{"ohlc4", "(b.Open + b.High + b.Low + b.Close) / 4"},
		{"hlc3", "(b.High + b.Low + b.Close) / 3"},
		{"hl2", "(b.High + b.Low) / 2"},
		{"hlcc4", "(b.High + b.Low + b.Close + b.Close) / 4"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			got, ok := SecurityBarFieldExpression(tt.field, bar)
			if !ok {
				t.Fatalf("SecurityBarFieldExpression(%q) returned ok=false", tt.field)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSecurityBarFieldExpression_UnknownField(t *testing.T) {
	unknowns := []string{"", "invalid", "Close", "bar_index", "time"}
	for _, field := range unknowns {
		t.Run(field, func(t *testing.T) {
			_, ok := SecurityBarFieldExpression(field, "secCtx.Data[0]")
			if ok {
				t.Errorf("SecurityBarFieldExpression(%q) should return ok=false for unknown field", field)
			}
		})
	}
}
