package codegen

import "testing"

func TestStrategyAnnotationIndex(t *testing.T) {
	tests := []struct {
		funcName  string
		parameter string
		want      int
	}{
		{"strategy.entry", "comment", 7},
		{"strategy.order", "comment", 7},
		{"strategy.exit", "comment", 12},
		{"strategy.exit", "comment_profit", 13},
		{"strategy.exit", "comment_loss", 14},
		{"strategy.exit", "comment_trailing", 15},
		{"strategy.close", "comment", 1},
		{"strategy.close_all", "comment", 0},
		{"strategy.entry", "comment_profit", noAnnotationPosition},
		{"strategy.close_all", "comment_trailing", noAnnotationPosition},
		{"unknown.function", "comment", noAnnotationPosition},
	}

	for _, tt := range tests {
		t.Run(tt.funcName+"/"+tt.parameter, func(t *testing.T) {
			if got := strategyAnnotationIndex(tt.funcName, tt.parameter); got != tt.want {
				t.Fatalf("strategyAnnotationIndex(%q, %q) = %d, want %d", tt.funcName, tt.parameter, got, tt.want)
			}
		})
	}
}

func TestIsPresentationOnlyStrategyPositionalArgMatchesAnnotationSlots(t *testing.T) {
	tests := []struct {
		funcName string
		index    int
		want     bool
	}{
		{"strategy.entry", 7, true},
		{"strategy.entry", 8, true},
		{"strategy.entry", 9, true},
		{"strategy.entry", 3, false},
		{"strategy.order", 3, false},
		{"strategy.exit", 12, true},
		{"strategy.exit", 20, true},
		{"strategy.exit", 21, false},
		{"strategy.close", 1, true},
		{"strategy.close", 2, false},
		{"strategy.close_all", 0, true},
		{"strategy.close_all", 1, true},
		{"strategy.close_all", 2, false},
		{"unknown.function", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			if got := isPresentationOnlyStrategyPositionalArg(tt.funcName, tt.index); got != tt.want {
				t.Fatalf("isPresentationOnlyStrategyPositionalArg(%q, %d) = %v, want %v", tt.funcName, tt.index, got, tt.want)
			}
		})
	}
}
