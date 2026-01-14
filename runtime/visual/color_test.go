package visual

import "testing"

func TestColorConstants(t *testing.T) {
	tests := []struct {
		name  string
		color string
		hex   string
	}{
		{"Aqua matches TradingView", Aqua, "#00BCD4"},
		{"Black matches TradingView", Black, "#363A45"},
		{"Blue matches TradingView", Blue, "#2962FF"},
		{"Fuchsia matches TradingView", Fuchsia, "#E040FB"},
		{"Gray matches TradingView", Gray, "#787B86"},
		{"Green matches TradingView", Green, "#4CAF50"},
		{"Lime matches TradingView", Lime, "#00E676"},
		{"Maroon matches TradingView", Maroon, "#880E4F"},
		{"Navy matches TradingView", Navy, "#311B92"},
		{"Olive matches TradingView", Olive, "#808000"},
		{"Orange matches TradingView", Orange, "#FF9800"},
		{"Purple matches TradingView", Purple, "#9C27B0"},
		{"Red matches TradingView", Red, "#FF5252"},
		{"Silver matches TradingView", Silver, "#B2B5BE"},
		{"Teal matches TradingView", Teal, "#00897B"},
		{"White matches TradingView", White, "#FFFFFF"},
		{"Yellow matches TradingView", Yellow, "#FFEB3B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.color != tt.hex {
				t.Errorf("Color constant = %s, want %s", tt.color, tt.hex)
			}
		})
	}
}
