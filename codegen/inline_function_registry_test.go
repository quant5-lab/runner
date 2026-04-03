package codegen

import "testing"

/* TestInlineFunctionRegistry_IsInlineOnly tests inline-only function detection */
func TestInlineFunctionRegistry_IsInlineOnly(t *testing.T) {
	registry := NewInlineFunctionRegistry()

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{
			name:     "valuewhen not inline-only (uses Series temp vars)",
			funcName: "valuewhen",
			want:     false,
		},
		{
			name:     "ta.valuewhen not inline-only (uses Series temp vars)",
			funcName: "ta.valuewhen",
			want:     false,
		},
		{
			name:     "sma is not inline-only",
			funcName: "ta.sma",
			want:     false,
		},
		{
			name:     "ema is not inline-only",
			funcName: "ta.ema",
			want:     false,
		},
		{
			name:     "unknown function",
			funcName: "unknown.func",
			want:     false,
		},
		{
			name:     "empty string is not inline-only",
			funcName: "",
			want:     false,
		},
		{
			name:     "barstate.isfirst not registered by default",
			funcName: "barstate.isfirst",
			want:     false,
		},
		{
			name:     "barstate.islast not registered by default",
			funcName: "barstate.islast",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.IsInlineOnly(tt.funcName)
			if got != tt.want {
				t.Errorf("IsInlineOnly(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

/* TestInlineFunctionRegistry_Register tests custom function registration */
func TestInlineFunctionRegistry_Register(t *testing.T) {
	registry := NewInlineFunctionRegistry()

	customFunc := "custom.inline"
	if registry.IsInlineOnly(customFunc) {
		t.Error("Custom function should not be registered initially")
	}

	registry.Register(customFunc)

	if !registry.IsInlineOnly(customFunc) {
		t.Error("Custom function should be registered after Register()")
	}
}

/* TestInlineFunctionRegistry_Isolation tests registry instance isolation */
func TestInlineFunctionRegistry_Isolation(t *testing.T) {
	registry1 := NewInlineFunctionRegistry()
	registry2 := NewInlineFunctionRegistry()

	registry1.Register("custom1")
	registry2.Register("custom2")

	if registry1.IsInlineOnly("custom2") {
		t.Error("Registry1 should not contain Registry2's custom function")
	}

	if registry2.IsInlineOnly("custom1") {
		t.Error("Registry2 should not contain Registry1's custom function")
	}
}

/* TestInlineFunctionRegistry_CaseSensitivity tests function name case handling */
func TestInlineFunctionRegistry_CaseSensitivity(t *testing.T) {
	registry := NewInlineFunctionRegistry()

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{"lowercase valuewhen", "valuewhen", false},
		{"uppercase VALUEWHEN", "VALUEWHEN", false},
		{"mixed case ValueWhen", "ValueWhen", false},
		{"mixed case Valuewhen", "Valuewhen", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.IsInlineOnly(tt.funcName)
			if got != tt.want {
				t.Errorf("IsInlineOnly(%q) = %v, want %v (case-sensitive check)",
					tt.funcName, got, tt.want)
			}
		})
	}
}

/* TestInlineFunctionRegistry_RepeatedRegistration tests duplicate registration handling */
func TestInlineFunctionRegistry_RepeatedRegistration(t *testing.T) {
	registry := NewInlineFunctionRegistry()

	funcName := "custom.test"

	registry.Register(funcName)
	if !registry.IsInlineOnly(funcName) {
		t.Fatalf("Function %q should be inline-only after first registration", funcName)
	}

	registry.Register(funcName)
	if !registry.IsInlineOnly(funcName) {
		t.Errorf("Function %q should remain inline-only after repeated registration", funcName)
	}

	registry.Register(funcName)
	if !registry.IsInlineOnly(funcName) {
		t.Errorf("Function %q should remain inline-only after third registration", funcName)
	}
}

/* TestInlineFunctionRegistry_EmptyStringRegistration tests empty function name handling */
func TestInlineFunctionRegistry_EmptyStringRegistration(t *testing.T) {
	registry := NewInlineFunctionRegistry()

	registry.Register("")

	if !registry.IsInlineOnly("") {
		t.Error("Empty string should be registered after explicit Register call")
	}
}

/* TestInlineFunctionRegistry_BulkOperations tests performance with many functions */
func TestInlineFunctionRegistry_BulkOperations(t *testing.T) {
	registry := NewInlineFunctionRegistry()

	const bulkCount = 1000
	for i := 0; i < bulkCount; i++ {
		registry.Register("bulk.func" + string(rune(i)))
	}

	notFoundCount := 0
	for i := 0; i < bulkCount*2; i++ {
		funcName := "bulk.func" + string(rune(i))
		if !registry.IsInlineOnly(funcName) {
			notFoundCount++
		}
	}

	if notFoundCount < bulkCount {
		t.Errorf("Expected at least %d not found functions, got %d", bulkCount, notFoundCount)
	}
}

/* TestInlineFunctionRegistry_Immutability tests registry state consistency */
func TestInlineFunctionRegistry_Immutability(t *testing.T) {
	registry := NewInlineFunctionRegistry()

	check1 := registry.IsInlineOnly("valuewhen")
	check2 := registry.IsInlineOnly("ta.sma")

	registry.Register("custom.test")

	check3 := registry.IsInlineOnly("valuewhen")
	check4 := registry.IsInlineOnly("ta.sma")
	check5 := registry.IsInlineOnly("custom.test")

	if check1 != check3 {
		t.Error("Built-in valuewhen detection changed after custom registration")
	}

	if check2 != check4 {
		t.Error("Non-inline function detection changed after custom registration")
	}

	if !check5 {
		t.Error("Custom registered function not detected as inline-only")
	}
}

/* TestInlineFunctionRegistry_NamespaceVariations tests namespace prefix handling */
func TestInlineFunctionRegistry_NamespaceVariations(t *testing.T) {
	registry := NewInlineFunctionRegistry()

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{"ta.valuewhen not inline-only", "ta.valuewhen", false},
		{"valuewhen not inline-only", "valuewhen", false},
		{"custom.valuewhen wrong namespace", "custom.valuewhen", false},
		{"barstate.isfirst not registered", "barstate.isfirst", false},
		{"isfirst without namespace", "isfirst", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.IsInlineOnly(tt.funcName)
			if got != tt.want {
				t.Errorf("IsInlineOnly(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}
