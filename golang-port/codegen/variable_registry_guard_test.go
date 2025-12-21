package codegen

import "testing"

/* TestVariableRegistryGuard_TypePreservation tests type preservation rules */
func TestVariableRegistryGuard_TypePreservation(t *testing.T) {
	tests := []struct {
		name         string
		existingType string
		newType      string
		wantPreserve bool
	}{
		{
			name:         "preserve function from float64 overwrite",
			existingType: "function",
			newType:      "float64",
			wantPreserve: true,
		},
		{
			name:         "preserve function from bool overwrite",
			existingType: "function",
			newType:      "bool",
			wantPreserve: true,
		},
		{
			name:         "preserve function from string overwrite",
			existingType: "function",
			newType:      "string",
			wantPreserve: true,
		},
		{
			name:         "preserve function from int overwrite",
			existingType: "function",
			newType:      "int",
			wantPreserve: true,
		},
		{
			name:         "allow function to function update",
			existingType: "function",
			newType:      "function",
			wantPreserve: false,
		},
		{
			name:         "allow float64 to bool transition",
			existingType: "float64",
			newType:      "bool",
			wantPreserve: false,
		},
		{
			name:         "allow bool to float64 transition",
			existingType: "bool",
			newType:      "float64",
			wantPreserve: false,
		},
		{
			name:         "allow string to int transition",
			existingType: "string",
			newType:      "int",
			wantPreserve: false,
		},
		{
			name:         "allow empty type transition",
			existingType: "",
			newType:      "float64",
			wantPreserve: false,
		},
		{
			name:         "function to empty type blocked",
			existingType: "function",
			newType:      "",
			wantPreserve: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := map[string]string{"testVar": tt.existingType}
			guard := NewVariableRegistryGuard(registry)
			got := guard.ShouldPreserveExistingType("testVar", tt.newType)
			if got != tt.wantPreserve {
				t.Errorf("ShouldPreserveExistingType(existingType=%q, newType=%q) = %v, want %v",
					tt.existingType, tt.newType, got, tt.wantPreserve)
			}
		})
	}
}

/* TestVariableRegistryGuard_SafeRegister tests safe registration behavior */
func TestVariableRegistryGuard_SafeRegister(t *testing.T) {
	tests := []struct {
		name           string
		initialReg     map[string]string
		varName        string
		varType        string
		wantRegistered bool
		wantFinalType  string
	}{
		{
			name:           "register new variable",
			initialReg:     map[string]string{},
			varName:        "x",
			varType:        "float64",
			wantRegistered: true,
			wantFinalType:  "float64",
		},
		{
			name:           "block function overwrite with float64",
			initialReg:     map[string]string{"adx": "function"},
			varName:        "adx",
			varType:        "float64",
			wantRegistered: false,
			wantFinalType:  "function",
		},
		{
			name:           "block function overwrite with bool",
			initialReg:     map[string]string{"fn": "function"},
			varName:        "fn",
			varType:        "bool",
			wantRegistered: false,
			wantFinalType:  "function",
		},
		{
			name:           "allow function to function update",
			initialReg:     map[string]string{"adx": "function"},
			varName:        "adx",
			varType:        "function",
			wantRegistered: true,
			wantFinalType:  "function",
		},
		{
			name:           "allow non-function type change",
			initialReg:     map[string]string{"x": "float64"},
			varName:        "x",
			varType:        "bool",
			wantRegistered: true,
			wantFinalType:  "bool",
		},
		{
			name:           "allow overwrite with same type",
			initialReg:     map[string]string{"x": "float64"},
			varName:        "x",
			varType:        "float64",
			wantRegistered: true,
			wantFinalType:  "float64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := make(map[string]string)
			for k, v := range tt.initialReg {
				registry[k] = v
			}

			guard := NewVariableRegistryGuard(registry)
			registered := guard.SafeRegister(tt.varName, tt.varType)

			if registered != tt.wantRegistered {
				t.Errorf("SafeRegister() returned %v, want %v", registered, tt.wantRegistered)
			}

			if finalType := registry[tt.varName]; finalType != tt.wantFinalType {
				t.Errorf("Final type = %q, want %q", finalType, tt.wantFinalType)
			}
		})
	}
}

/* TestVariableRegistryGuard_EdgeCases tests boundary conditions */
func TestVariableRegistryGuard_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		initialReg     map[string]string
		varName        string
		varType        string
		wantRegistered bool
	}{
		{
			name:           "empty variable name",
			initialReg:     map[string]string{},
			varName:        "",
			varType:        "float64",
			wantRegistered: true,
		},
		{
			name:           "empty type name",
			initialReg:     map[string]string{},
			varName:        "x",
			varType:        "",
			wantRegistered: true,
		},
		{
			name:           "special characters in variable name",
			initialReg:     map[string]string{},
			varName:        "var_with-special.chars",
			varType:        "float64",
			wantRegistered: true,
		},
		{
			name:           "unicode variable name",
			initialReg:     map[string]string{},
			varName:        "变量",
			varType:        "float64",
			wantRegistered: true,
		},
		{
			name:           "function with empty name blocked from overwrite",
			initialReg:     map[string]string{"": "function"},
			varName:        "",
			varType:        "float64",
			wantRegistered: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := make(map[string]string)
			for k, v := range tt.initialReg {
				registry[k] = v
			}

			guard := NewVariableRegistryGuard(registry)
			registered := guard.SafeRegister(tt.varName, tt.varType)

			if registered != tt.wantRegistered {
				t.Errorf("SafeRegister(%q, %q) = %v, want %v",
					tt.varName, tt.varType, registered, tt.wantRegistered)
			}
		})
	}
}

/* TestVariableRegistryGuard_SequentialOperations tests multiple operations in sequence */
func TestVariableRegistryGuard_SequentialOperations(t *testing.T) {
	registry := make(map[string]string)
	guard := NewVariableRegistryGuard(registry)

	if !guard.SafeRegister("myFunc", "function") {
		t.Fatal("Failed to register function initially")
	}

	if guard.SafeRegister("myFunc", "float64") {
		t.Error("Function overwrite with float64 should be blocked")
	}

	if registry["myFunc"] != "function" {
		t.Errorf("Function type changed to %q, should remain 'function'", registry["myFunc"])
	}

	if guard.SafeRegister("myFunc", "bool") {
		t.Error("Function overwrite with bool should be blocked")
	}

	if !guard.SafeRegister("myFunc", "function") {
		t.Error("Function update to function should succeed")
	}

	if !guard.SafeRegister("x", "float64") {
		t.Error("New variable registration should succeed")
	}

	if !guard.SafeRegister("x", "bool") {
		t.Error("Non-function type change should succeed")
	}

	if registry["myFunc"] != "function" {
		t.Errorf("Final myFunc type = %q, want 'function'", registry["myFunc"])
	}
	if registry["x"] != "bool" {
		t.Errorf("Final x type = %q, want 'bool'", registry["x"])
	}
}

/* TestVariableRegistryGuard_Isolation tests guard instance independence */
func TestVariableRegistryGuard_Isolation(t *testing.T) {
	registry1 := map[string]string{"func1": "function"}
	registry2 := map[string]string{"func2": "function"}

	guard1 := NewVariableRegistryGuard(registry1)
	guard2 := NewVariableRegistryGuard(registry2)

	guard1.SafeRegister("var1", "float64")

	if _, exists := registry2["var1"]; exists {
		t.Error("Guard1 operations affected Guard2's registry")
	}

	guard2.SafeRegister("var2", "bool")

	if _, exists := registry1["var2"]; exists {
		t.Error("Guard2 operations affected Guard1's registry")
	}

	guard1.SafeRegister("func1", "float64")
	guard2.SafeRegister("func2", "int")

	if registry1["func1"] != "function" {
		t.Error("Guard1 function protection failed")
	}
	if registry2["func2"] != "function" {
		t.Error("Guard2 function protection failed")
	}
}

/* TestVariableRegistryGuard_StateConsistency tests registry state remains consistent */
func TestVariableRegistryGuard_StateConsistency(t *testing.T) {
	registry := map[string]string{
		"fn1":  "function",
		"fn2":  "function",
		"var1": "float64",
	}

	guard := NewVariableRegistryGuard(registry)

	initialCount := len(registry)
	fn1Type := registry["fn1"]
	fn2Type := registry["fn2"]
	var1Type := registry["var1"]

	guard.SafeRegister("fn1", "float64")
	guard.SafeRegister("fn1", "bool")
	guard.SafeRegister("fn2", "int")

	guard.SafeRegister("var1", "bool")
	guard.SafeRegister("var2", "string")
	guard.SafeRegister("var3", "float64")

	if registry["fn1"] != fn1Type {
		t.Errorf("fn1 type changed from %q to %q", fn1Type, registry["fn1"])
	}
	if registry["fn2"] != fn2Type {
		t.Errorf("fn2 type changed from %q to %q", fn2Type, registry["fn2"])
	}
	if registry["var1"] == var1Type {
		t.Error("var1 type should have changed but didn't")
	}
	if len(registry) != initialCount+2 {
		t.Errorf("Registry size = %d, want %d", len(registry), initialCount+2)
	}
}

/* TestVariableRegistryGuard_BulkOperations tests performance with many variables */
func TestVariableRegistryGuard_BulkOperations(t *testing.T) {
	registry := make(map[string]string)
	guard := NewVariableRegistryGuard(registry)

	const bulkCount = 1000

	for i := 0; i < bulkCount; i++ {
		varName := "func" + string(rune('A'+i%26)) + string(rune(i))
		if !guard.SafeRegister(varName, "function") {
			t.Fatalf("Failed to register function at iteration %d", i)
		}
	}

	protectedCount := 0
	for i := 0; i < bulkCount; i++ {
		varName := "func" + string(rune('A'+i%26)) + string(rune(i))
		if !guard.SafeRegister(varName, "float64") {
			protectedCount++
		}
	}

	if protectedCount != bulkCount {
		t.Errorf("Protected %d/%d functions", protectedCount, bulkCount)
	}

	for i := 0; i < bulkCount; i++ {
		varName := "func" + string(rune('A'+i%26)) + string(rune(i))
		if registry[varName] != "function" {
			t.Errorf("Variable %q type = %q, want 'function'", varName, registry[varName])
		}
	}
}

/* TestVariableRegistryGuard_TypeStringVariations tests type string format edge cases */
func TestVariableRegistryGuard_TypeStringVariations(t *testing.T) {
	tests := []struct {
		name          string
		initialType   string
		newType       string
		expectedFinal string
	}{
		{
			name:          "case sensitive function detection",
			initialType:   "Function",
			newType:       "float64",
			expectedFinal: "float64", // "Function" != "function", not protected
		},
		{
			name:          "exact match required",
			initialType:   "function",
			newType:       "float64",
			expectedFinal: "function", // Protected
		},
		{
			name:          "whitespace in type not trimmed",
			initialType:   "function ",
			newType:       "float64",
			expectedFinal: "float64", // "function " != "function"
		},
		{
			name:          "complex type strings allowed",
			initialType:   "*context.ArrowContext",
			newType:       "map[string]string",
			expectedFinal: "map[string]string",
		},
		{
			name:          "empty string to empty string",
			initialType:   "",
			newType:       "",
			expectedFinal: "", // Not protected, allows update
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := map[string]string{"var": tt.initialType}
			guard := NewVariableRegistryGuard(registry)

			registered := guard.SafeRegister("var", tt.newType)

			if registry["var"] != tt.expectedFinal {
				t.Errorf("Final type = %q, want %q", registry["var"], tt.expectedFinal)
			}

			// Check if initial type was actually preserved (blocked from update)
			wasBlocked := (tt.initialType == "function" && tt.newType != "function")

			if wasBlocked && registered {
				t.Errorf("Registration should be blocked but succeeded")
			}
			if !wasBlocked && !registered {
				t.Errorf("Registration should succeed but was blocked")
			}
		})
	}
}

/* TestVariableRegistryGuard_VariableNameVariations tests variable name edge cases */
func TestVariableRegistryGuard_VariableNameVariations(t *testing.T) {
	tests := []struct {
		name    string
		varName string
		varType string
	}{
		{"very long variable name", "thisIsAVeryLongVariableNameThatExceedsTypicalLengthConstraintsInMostProgrammingContexts", "float64"},
		{"variable with numbers", "var123", "float64"},
		{"variable starting with underscore", "_privateVar", "function"},
		{"variable with multiple underscores", "var__name", "function"},
		{"camelCase variable", "myVariableName", "float64"},
		{"PascalCase variable", "MyVariableName", "function"},
		{"snake_case variable", "my_variable_name", "float64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := make(map[string]string)
			guard := NewVariableRegistryGuard(registry)

			if !guard.SafeRegister(tt.varName, tt.varType) {
				t.Errorf("Failed to register variable %q with type %q", tt.varName, tt.varType)
			}

			if registry[tt.varName] != tt.varType {
				t.Errorf("Variable %q type = %q, want %q", tt.varName, registry[tt.varName], tt.varType)
			}

			// Test function protection works with varied names
			if tt.varType == "function" {
				if guard.SafeRegister(tt.varName, "float64") {
					t.Errorf("Function type for %q should be protected from overwrite", tt.varName)
				}
				if registry[tt.varName] != "function" {
					t.Errorf("Function type changed for %q", tt.varName)
				}
			}
		})
	}
}

/* TestVariableRegistryGuard_MultiPhaseRegistration simulates multi-phase codegen workflow */
func TestVariableRegistryGuard_MultiPhaseRegistration(t *testing.T) {
	tests := []struct {
		name          string
		phases        [][]struct{ varName, varType string }
		expectedFinal map[string]string
	}{
		{
			name: "three-phase codegen simulation",
			phases: [][]struct{ varName, varType string }{
				// Phase 1: First pass variable collection
				{
					{"x", "float64"},
					{"y", "float64"},
					{"result", "float64"},
				},
				// Phase 2: Arrow function registration
				{
					{"myFunc", "function"},
					{"calculate", "function"},
				},
				// Phase 3: Statement generation (attempts overwrites)
				{
					{"myFunc", "float64"},
					{"calculate", "bool"},
					{"x", "bool"},
					{"z", "string"},
				},
			},
			expectedFinal: map[string]string{
				"x":         "bool",
				"y":         "float64",
				"result":    "float64",
				"myFunc":    "function",
				"calculate": "function",
				"z":         "string",
			},
		},
		{
			name: "arrow function tuple call pattern",
			phases: [][]struct{ varName, varType string }{
				// Phase 1: Tuple destructuring variables
				{
					{"ADX", "float64"},
					{"up", "float64"},
					{"down", "float64"},
				},
				// Phase 2: Arrow function definition
				{
					{"adx", "function"},
				},
				// Phase 3: Re-registration attempt during tuple call
				{
					{"adx", "float64"},
					{"ADX", "float64"},
				},
			},
			expectedFinal: map[string]string{
				"ADX":  "float64",
				"up":   "float64",
				"down": "float64",
				"adx":  "function",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := make(map[string]string)
			guard := NewVariableRegistryGuard(registry)

			// Execute all phases
			for phaseNum, phase := range tt.phases {
				for _, reg := range phase {
					guard.SafeRegister(reg.varName, reg.varType)
				}
				t.Logf("After phase %d: %v", phaseNum+1, registry)
			}

			// Verify final state
			if len(registry) != len(tt.expectedFinal) {
				t.Errorf("Registry size = %d, want %d", len(registry), len(tt.expectedFinal))
			}

			for varName, expectedType := range tt.expectedFinal {
				if actualType, exists := registry[varName]; !exists {
					t.Errorf("Variable %q missing from registry", varName)
				} else if actualType != expectedType {
					t.Errorf("Variable %q type = %q, want %q", varName, actualType, expectedType)
				}
			}
		})
	}
}

/* TestVariableRegistryGuard_NilAndEmptyHandling tests nil and empty state handling */
func TestVariableRegistryGuard_NilAndEmptyHandling(t *testing.T) {
	t.Run("nil registry pointer", func(t *testing.T) {
		var nilMap map[string]string
		guard := &VariableRegistryGuard{registry: nilMap}

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when accessing nil map, but no panic occurred")
			}
		}()

		guard.SafeRegister("x", "float64")
	})

	t.Run("empty registry operations", func(t *testing.T) {
		registry := make(map[string]string)
		guard := NewVariableRegistryGuard(registry)

		// Multiple operations on empty registry
		for i := 0; i < 10; i++ {
			varName := "var" + string(rune('0'+i))
			if !guard.SafeRegister(varName, "float64") {
				t.Errorf("Failed to register %q in empty registry", varName)
			}
		}

		if len(registry) != 10 {
			t.Errorf("Registry size = %d, want 10", len(registry))
		}
	})

	t.Run("zero-value guard struct", func(t *testing.T) {
		var guard VariableRegistryGuard

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic with zero-value guard, but no panic occurred")
			}
		}()

		guard.SafeRegister("x", "float64")
	})
}

/* TestVariableRegistryGuard_StressTest tests performance with large variable sets */
func TestVariableRegistryGuard_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	const (
		totalVars         = 10000
		functionVars      = 1000
		overwriteAttempts = 5
	)

	registry := make(map[string]string)
	guard := NewVariableRegistryGuard(registry)

	// Phase 1: Register mixed variables
	for i := 0; i < totalVars; i++ {
		varName := "var" + string(rune(i))
		varType := "float64"
		if i < functionVars {
			varType = "function"
		}
		if !guard.SafeRegister(varName, varType) {
			t.Fatalf("Failed initial registration at iteration %d", i)
		}
	}

	if len(registry) != totalVars {
		t.Fatalf("Registry size after initial registration = %d, want %d", len(registry), totalVars)
	}

	// Phase 2: Attempt to overwrite all function types multiple times
	protectedCount := 0
	for attempt := 0; attempt < overwriteAttempts; attempt++ {
		for i := 0; i < functionVars; i++ {
			varName := "var" + string(rune(i))
			if !guard.SafeRegister(varName, "float64") {
				protectedCount++
			}
		}
	}

	expectedProtected := functionVars * overwriteAttempts
	if protectedCount != expectedProtected {
		t.Errorf("Protected %d overwrites, want %d", protectedCount, expectedProtected)
	}

	// Phase 3: Verify all function types remain intact
	functionIntact := 0
	for i := 0; i < functionVars; i++ {
		varName := "var" + string(rune(i))
		if registry[varName] == "function" {
			functionIntact++
		}
	}

	if functionIntact != functionVars {
		t.Errorf("Function types intact = %d, want %d", functionIntact, functionVars)
	}
}

/* TestVariableRegistryGuard_TypeTransitionMatrix tests all type transition combinations */
func TestVariableRegistryGuard_TypeTransitionMatrix(t *testing.T) {
	types := []string{"function", "float64", "bool", "int", "string", ""}

	for _, fromType := range types {
		for _, toType := range types {
			t.Run(fromType+"_to_"+toType, func(t *testing.T) {
				registry := make(map[string]string)
				if fromType != "" {
					registry["var"] = fromType
				}

				guard := NewVariableRegistryGuard(registry)
				registered := guard.SafeRegister("var", toType)

				// Function type should be preserved unless updating to function
				shouldBeBlocked := (fromType == "function" && toType != "function")

				if shouldBeBlocked {
					if registered {
						t.Errorf("Registration should be blocked but succeeded")
					}
					if registry["var"] != fromType {
						t.Errorf("Type changed from %q to %q, should be preserved", fromType, registry["var"])
					}
				} else {
					if !registered {
						t.Errorf("Registration should succeed but was blocked")
					}
					if registry["var"] != toType {
						t.Errorf("Type = %q, want %q", registry["var"], toType)
					}
				}
			})
		}
	}
}
