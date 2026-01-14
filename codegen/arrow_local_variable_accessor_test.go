package codegen

import "testing"

func TestArrowLocalVariableAccessor_RegisterAndCheck(t *testing.T) {
	accessor := NewArrowLocalVariableAccessor()

	accessor.RegisterLocalVariable("up")
	accessor.RegisterLocalVariable("down")

	if !accessor.IsLocalVariable("up") {
		t.Error("Expected 'up' to be registered as local variable")
	}

	if !accessor.IsLocalVariable("down") {
		t.Error("Expected 'down' to be registered as local variable")
	}

	if accessor.IsLocalVariable("notRegistered") {
		t.Error("Expected 'notRegistered' to NOT be registered")
	}
}

func TestArrowLocalVariableAccessor_GenerateAccess(t *testing.T) {
	tests := []struct {
		name    string
		varName string
		offset  int
		want    string
	}{
		{
			name:    "current bar scalar access",
			varName: "up",
			offset:  0,
			want:    "up",
		},
		{
			name:    "historical series access offset 1",
			varName: "up",
			offset:  1,
			want:    "upSeries.Get(1)",
		},
		{
			name:    "historical series access offset 2",
			varName: "down",
			offset:  2,
			want:    "downSeries.Get(2)",
		},
		{
			name:    "historical series access offset 10",
			varName: "truerange",
			offset:  10,
			want:    "truerangeSeries.Get(10)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := NewArrowLocalVariableAccessor()
			accessor.RegisterLocalVariable(tt.varName)

			got := accessor.GenerateAccess(tt.varName, tt.offset)
			if got != tt.want {
				t.Errorf("GenerateAccess(%q, %d) = %q, want %q", tt.varName, tt.offset, got, tt.want)
			}
		})
	}
}

func TestArrowLocalVariableAccessor_DualAccessPattern(t *testing.T) {
	accessor := NewArrowLocalVariableAccessor()
	accessor.RegisterLocalVariable("up")
	accessor.RegisterLocalVariable("down")

	currentUp := accessor.GenerateAccess("up", 0)
	if currentUp != "up" {
		t.Errorf("Current bar access should be scalar 'up', got: %s", currentUp)
	}

	historicalUp := accessor.GenerateAccess("up", 1)
	if historicalUp != "upSeries.Get(1)" {
		t.Errorf("Historical access should be 'upSeries.Get(1)', got: %s", historicalUp)
	}
}
