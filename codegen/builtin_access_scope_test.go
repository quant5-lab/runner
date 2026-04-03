package codegen

import "testing"

func TestAccessScope_Constants(t *testing.T) {
	tests := []struct {
		name  string
		scope AccessScope
		value int
	}{
		{"BarLoopScope is 0", BarLoopScope, 0},
		{"SecurityScope is 1", SecurityScope, 1},
		{"ArrowScope is 2", ArrowScope, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.scope) != tt.value {
				t.Errorf("AccessScope %s = %d, want %d", tt.name, int(tt.scope), tt.value)
			}
		})
	}
}

func TestScopeFromSecurityFlag(t *testing.T) {
	tests := []struct {
		name     string
		flag     bool
		expected AccessScope
	}{
		{"true returns SecurityScope", true, SecurityScope},
		{"false returns BarLoopScope", false, BarLoopScope},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScopeFromSecurityFlag(tt.flag)
			if result != tt.expected {
				t.Errorf("ScopeFromSecurityFlag(%v) = %v, want %v", tt.flag, result, tt.expected)
			}
		})
	}
}
