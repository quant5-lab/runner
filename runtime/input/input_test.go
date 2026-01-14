package input

import "testing"

func TestNewManager(t *testing.T) {
	m := NewManager(nil)
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}
	if m.overrides == nil {
		t.Error("Manager.overrides not initialized")
	}
}

func TestIntWithOverride(t *testing.T) {
	overrides := map[string]interface{}{
		"Length": 20,
	}
	m := NewManager(overrides)

	got := m.Int(10, "Length")
	if got != 20 {
		t.Errorf("Int() = %d, want 20", got)
	}
}

func TestIntWithoutOverride(t *testing.T) {
	m := NewManager(nil)

	got := m.Int(10, "Length")
	if got != 10 {
		t.Errorf("Int() = %d, want 10 (default)", got)
	}
}

func TestIntWithFloat64Override(t *testing.T) {
	overrides := map[string]interface{}{
		"Length": 25.7,
	}
	m := NewManager(overrides)

	got := m.Int(10, "Length")
	if got != 25 {
		t.Errorf("Int() = %d, want 25 (converted from float64)", got)
	}
}

func TestFloatWithOverride(t *testing.T) {
	overrides := map[string]interface{}{
		"Factor": 2.5,
	}
	m := NewManager(overrides)

	got := m.Float(1.0, "Factor")
	if got != 2.5 {
		t.Errorf("Float() = %f, want 2.5", got)
	}
}

func TestFloatWithIntOverride(t *testing.T) {
	overrides := map[string]interface{}{
		"Factor": 3,
	}
	m := NewManager(overrides)

	got := m.Float(1.0, "Factor")
	if got != 3.0 {
		t.Errorf("Float() = %f, want 3.0 (converted from int)", got)
	}
}

func TestFloatWithoutOverride(t *testing.T) {
	m := NewManager(nil)

	got := m.Float(1.5, "Factor")
	if got != 1.5 {
		t.Errorf("Float() = %f, want 1.5 (default)", got)
	}
}

func TestStringWithOverride(t *testing.T) {
	overrides := map[string]interface{}{
		"Title": "Custom Title",
	}
	m := NewManager(overrides)

	got := m.String("Default", "Title")
	if got != "Custom Title" {
		t.Errorf("String() = %s, want Custom Title", got)
	}
}

func TestStringWithoutOverride(t *testing.T) {
	m := NewManager(nil)

	got := m.String("Default", "Title")
	if got != "Default" {
		t.Errorf("String() = %s, want Default", got)
	}
}

func TestBoolWithOverride(t *testing.T) {
	overrides := map[string]interface{}{
		"Enabled": true,
	}
	m := NewManager(overrides)

	got := m.Bool(false, "Enabled")
	if got != true {
		t.Errorf("Bool() = %v, want true", got)
	}
}

func TestBoolWithoutOverride(t *testing.T) {
	m := NewManager(nil)

	got := m.Bool(false, "Enabled")
	if got != false {
		t.Errorf("Bool() = %v, want false (default)", got)
	}
}

func TestEmptyTitleReturnsDefault(t *testing.T) {
	overrides := map[string]interface{}{
		"Something": 100,
	}
	m := NewManager(overrides)

	if got := m.Int(10, ""); got != 10 {
		t.Errorf("Int with empty title = %d, want 10", got)
	}
	if got := m.Float(1.5, ""); got != 1.5 {
		t.Errorf("Float with empty title = %f, want 1.5", got)
	}
	if got := m.String("test", ""); got != "test" {
		t.Errorf("String with empty title = %s, want test", got)
	}
	if got := m.Bool(true, ""); got != true {
		t.Errorf("Bool with empty title = %v, want true", got)
	}
}

func TestWrongTypeReturnsDefault(t *testing.T) {
	overrides := map[string]interface{}{
		"Length": "not a number",
		"Factor": "not a float",
		"Title":  123,
		"Flag":   "not a bool",
	}
	m := NewManager(overrides)

	if got := m.Int(10, "Length"); got != 10 {
		t.Errorf("Int with wrong type = %d, want 10 (default)", got)
	}
	if got := m.Float(1.5, "Factor"); got != 1.5 {
		t.Errorf("Float with wrong type = %f, want 1.5 (default)", got)
	}
	if got := m.String("default", "Title"); got != "default" {
		t.Errorf("String with wrong type = %s, want default", got)
	}
	if got := m.Bool(false, "Flag"); got != false {
		t.Errorf("Bool with wrong type = %v, want false (default)", got)
	}
}
