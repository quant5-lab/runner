package codegen

import "testing"

func TestIsTickerConstructorFunction(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"ticker.heikinashi", true},
		{"ticker.renko", true},
		{"ticker.kagi", true},
		{"ticker.linebreak", true},
		{"ticker.pointfigure", true},
		{"ticker.range", true},
		{"ticker.new", true},
		{"ticker.modify", true},
		{"ticker.standard", true},
		{"ticker.inherit", true},

		{"heikinashi", true},
		{"heikenashi", true},
		{"renko", true},
		{"kagi", true},
		{"linebreak", true},
		{"pointfigure", true},
		{"range", true},

		{"ta.sma", false},
		{"ta.ema", false},
		{"strategy.entry", false},
		{"request.security", false},
		{"plot", false},
		{"math.abs", false},
		{"str.tostring", false},
		{"", false},

		{"TICKER.HEIKINASHI", false},
		{"Heikinashi", false},
		{"RANGE", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTickerConstructorFunction(tt.name)
			if got != tt.want {
				t.Errorf("IsTickerConstructorFunction(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
