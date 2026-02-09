package codegen

import (
	"strings"
	"testing"
)

/* Switch nested control flow code generation */

func TestSwitchCodegen_NestedSwitch(t *testing.T) {
	source := `switch outer
    1 =>
        switch inner
            10 =>
                x = 1
            20 =>
                x = 2
    2 =>
        y = 2`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilation: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustNotContain("switch")

	// Should have nested if-else chains
	ifCount := strings.Count(code, "if ")
	if ifCount < 2 {
		t.Errorf("expected at least 2 if statements (outer + inner), got %d", ifCount)
	}
}

func TestSwitchCodegen_NestedIf(t *testing.T) {
	source := `switch mode
    1 =>
        if condition
            x = 1
        else
            x = 2
    2 =>
        if otherCondition
            y = 1`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilation: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustNotContain("switch")
	verifier.MustContain("if")
	verifier.MustContain("else")
}

func TestSwitchCodegen_NestedFor(t *testing.T) {
	source := `switch mode
    1 =>
        for i = 1 to 10
            sum = sum + i
    2 =>
        result = 0`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilation: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustNotContain("switch")
	verifier.MustContain("for")
}

func TestSwitchCodegen_IfContainingSwitch(t *testing.T) {
	source := `if condition
    switch mode
        1 =>
            x = 1
        2 =>
            x = 2`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilation: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustNotContain("switch")
	verifier.MustContain("if")
}

func TestSwitchCodegen_ForContainingSwitch(t *testing.T) {
	source := `for i = 1 to 10
    switch i
        5 =>
            x = 1
        =>
            x = 0`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilation: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustNotContain("switch")
	verifier.MustContain("for")
}

func TestSwitchCodegen_DeepNesting(t *testing.T) {
	source := `switch a
    1 =>
        switch b
            10 =>
                switch c
                    100 =>
                        x = 1`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilation: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustNotContain("switch")

	// Should have deeply nested if statements
	ifCount := strings.Count(code, "if ")
	if ifCount < 3 {
		t.Errorf("expected at least 3 nested if statements, got %d", ifCount)
	}
}

func TestSwitchCodegen_MultipleNestingPatterns(t *testing.T) {
	source := `switch mode
    1 =>
        if condition
            x = 1
    2 =>
        for i = 1 to 5
            y = i
    3 =>
        switch submode
            10 =>
                z = 10`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilation: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustNotContain("switch")
	verifier.MustContain("if", "for")
}

/* Switch expression IIFE generation */

func TestSwitchCodegen_ExpressionIIFEBasic(t *testing.T) {
	source := `x = switch mode
    1 =>
        10
    2 =>
        20`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilation: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustContain("func() float64", "return")
	verifier.MustNotContain("switch")

	// Should have IIFE pattern
	if !strings.Contains(code, "(func() float64 {") {
		t.Error("expected IIFE pattern for switch expression")
	}
}

func TestSwitchCodegen_ExpressionIIFEWithDefault(t *testing.T) {
	source := `x = switch mode
    1 =>
        10
    =>
        0`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilation: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustContain("func() float64", "} else {")

	// Should have at least 2 return statements (case + default)
	returnCount := strings.Count(code, "return")
	if returnCount < 2 {
		t.Errorf("expected at least 2 return statements, got %d", returnCount)
	}
}

func TestSwitchCodegen_ExpressionIIFEManyBranches(t *testing.T) {
	source := `x = switch mode
    1 =>
        10
    2 =>
        20
    3 =>
        30
    4 =>
        40
    5 =>
        50`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilation: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustContain("func() float64")

	// Should have 5 return statements (one per case)
	returnCount := strings.Count(code, "return")
	if returnCount < 5 {
		t.Errorf("expected at least 5 return statements, got %d", returnCount)
	}

	// Should have else-if chains
	elseIfCount := strings.Count(code, "} else if")
	if elseIfCount < 3 {
		t.Errorf("expected at least 3 else-if chains, got %d", elseIfCount)
	}
}

/* Switch-expression na semantics: missing default returns math.NaN() per PineScript spec */
func TestSwitchCodegen_ExpressionAlternateBehavior(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		mustHaveNaN  bool
		mustHaveIIFE bool
		mustNotHave0 bool
	}{
		{
			name: "no default returns na",
			source: `x = switch mode
    1 =>
        10
    2 =>
        20`,
			mustHaveNaN:  true,
			mustHaveIIFE: true,
			mustNotHave0: true,
		},
		{
			name: "with default returns value",
			source: `x = switch mode
    1 =>
        10
    =>
        99`,
			mustHaveNaN:  false,
			mustHaveIIFE: true,
			mustNotHave0: false,
		},
		{
			name: "form 2 no default returns na",
			source: `x = switch
    close > open =>
        1
    close < open =>
        -1`,
			mustHaveNaN:  true,
			mustHaveIIFE: true,
			mustNotHave0: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.source)
			if err != nil {
				t.Fatalf("compilation: %v", err)
			}

			verifier := NewCodeVerifier(code, t)

			if tt.mustHaveIIFE {
				verifier.MustContain("func() float64")
			}

			if tt.mustHaveNaN {
				verifier.MustContain("return math.NaN()")
			}

			if tt.mustNotHave0 {
				verifier.MustNotContain("return 0.0")
			}
		})
	}
}

/* Switch edge cases */

func TestSwitchCodegen_EmptyDefault(t *testing.T) {
	source := `switch val
    1 =>
        x = 1
    =>
        x = 0`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilation: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustContain("} else {")
	verifier.MustNotContain("switch")
}

func TestSwitchCodegen_MultiStatementCases(t *testing.T) {
	source := `switch mode
    1 =>
        a = 1
        b = 2
        c = 3
    2 =>
        x = 10
        y = 20`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilation: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustNotContain("switch")
	verifier.MustContain("if")

	// Should have all assignments
	for _, varName := range []string{"a", "b", "c", "x", "y"} {
		if !strings.Contains(code, varName) {
			t.Errorf("expected variable %s in generated code", varName)
		}
	}
}

func TestSwitchCodegen_Form1ScalarComparison(t *testing.T) {
	source := `result = switch mode
    1 =>
        1
    2 =>
        -1
    =>
        0`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilation: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustNotContain("switch")
}

func TestSwitchCodegen_Form2ComplexConditions(t *testing.T) {
	source := `result = switch
    close > open and volume > 1000 =>
        1
    close < open or volume < 500 =>
        -1
    =>
        0`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilation: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustContain("&&", "||")
	verifier.MustNotContain("switch")
}

func TestSwitchCodegen_MixedWithOtherControlFlow(t *testing.T) {
	source := `if condition1
    x = 1

switch mode
    1 =>
        y = 1
    2 =>
        y = 2

for i = 1 to 10
    z = i`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilation: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustContain("if", "for")
	verifier.MustNotContain("switch")
}

/* Real-world switch patterns */

func TestSwitchCodegen_TradingSignal(t *testing.T) {
	source := `signal = switch
    ta.crossover(close, ta.sma(close, 20)) =>
        1
    ta.crossunder(close, ta.sma(close, 20)) =>
        -1
    =>
        0`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilation: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustContain("func() float64")
	verifier.MustNotContain("switch")
}

func TestSwitchCodegen_StateTransition(t *testing.T) {
	source := `switch currentState
    1 =>
        if trigger
            currentState := 2
    2 =>
        if exitCondition
            currentState := 0`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("compilation: %v", err)
	}

	verifier := NewCodeVerifier(code, t)
	verifier.MustNotContain("switch")
}
