package parser

import (
	"testing"
)

// TestSwitchStatement_NestedSwitch verifies switch statements nested inside switch statements
func TestSwitchStatement_NestedSwitch(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedDepth  int
		outerCaseCount int
	}{
		{
			name: "switch as first statement in switch case",
			source: `switch outer
    1 =>
        switch inner
            10 =>
                x = 1`,
			expectedDepth:  2,
			outerCaseCount: 1,
		},
		{
			name: "switch after statement in switch case",
			source: `switch outer
    1 =>
        y = 1
        switch inner
            10 =>
                x = 1`,
			expectedDepth:  2,
			outerCaseCount: 1,
		},
		{
			name: "triple nested switch",
			source: `switch a
    1 =>
        switch b
            10 =>
                switch c
                    100 =>
                        x = 1`,
			expectedDepth:  3,
			outerCaseCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("parser creation: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			if len(script.Statements) == 0 {
				t.Fatal("no statements parsed")
			}

			outerSwitch := script.Statements[0].Core.Switch
			if outerSwitch == nil {
				t.Fatal("expected outer switch statement")
			}

			if len(outerSwitch.Cases) != tt.outerCaseCount {
				t.Errorf("expected %d cases in outer switch, got %d", tt.outerCaseCount, len(outerSwitch.Cases))
			}

			// Count nesting depth
			depth := 1
			current := outerSwitch
			for {
				found := false
				for _, caseStmt := range current.Cases {
					for _, stmt := range caseStmt.Body {
						if stmt.Core.Switch != nil {
							depth++
							current = stmt.Core.Switch
							found = true
							break
						}
					}
					if found {
						break
					}
				}
				if !found {
					break
				}
			}

			if depth != tt.expectedDepth {
				t.Errorf("expected nesting depth %d, got %d", tt.expectedDepth, depth)
			}
		})
	}
}

// TestSwitchStatement_NestedIf verifies if statements nested inside switch statements
func TestSwitchStatement_NestedIf(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		outerCaseCount int
		hasNestedIf    bool
	}{
		{
			name: "if as first statement in switch case",
			source: `switch mode
    1 =>
        if condition
            x = 1`,
			outerCaseCount: 1,
			hasNestedIf:    true,
		},
		{
			name: "if after statement in switch case",
			source: `switch mode
    1 =>
        y = 1
        if condition
            x = 1`,
			outerCaseCount: 1,
			hasNestedIf:    true,
		},
		{
			name: "multiple nested ifs",
			source: `switch mode
    1 =>
        if cond1
            x = 1
    2 =>
        if cond2
            x = 2`,
			outerCaseCount: 2,
			hasNestedIf:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("parser creation: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			switchStmt := script.Statements[0].Core.Switch
			if switchStmt == nil {
				t.Fatal("expected switch statement")
			}

			if len(switchStmt.Cases) != tt.outerCaseCount {
				t.Errorf("expected %d cases, got %d", tt.outerCaseCount, len(switchStmt.Cases))
			}

			hasIf := false
			for _, caseStmt := range switchStmt.Cases {
				for _, stmt := range caseStmt.Body {
					if stmt.Core.If != nil {
						hasIf = true
						break
					}
				}
				if hasIf {
					break
				}
			}

			if hasIf != tt.hasNestedIf {
				t.Errorf("expected hasNestedIf=%v, got %v", tt.hasNestedIf, hasIf)
			}
		})
	}
}

// TestSwitchStatement_NestedFor verifies for loops nested inside switch statements
func TestSwitchStatement_NestedFor(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		outerCaseCount int
		hasNestedFor   bool
	}{
		{
			name: "for as first statement in switch case",
			source: `switch mode
    1 =>
        for i = 1 to 10
            x = i`,
			outerCaseCount: 1,
			hasNestedFor:   true,
		},
		{
			name: "for after statement in switch case",
			source: `switch mode
    1 =>
        y = 1
        for i = 1 to 10
            x = i`,
			outerCaseCount: 1,
			hasNestedFor:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("parser creation: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			switchStmt := script.Statements[0].Core.Switch
			if switchStmt == nil {
				t.Fatal("expected switch statement")
			}

			if len(switchStmt.Cases) != tt.outerCaseCount {
				t.Errorf("expected %d cases, got %d", tt.outerCaseCount, len(switchStmt.Cases))
			}

			hasFor := false
			for _, caseStmt := range switchStmt.Cases {
				for _, stmt := range caseStmt.Body {
					if stmt.Core.For != nil {
						hasFor = true
						break
					}
				}
				if hasFor {
					break
				}
			}

			if hasFor != tt.hasNestedFor {
				t.Errorf("expected hasNestedFor=%v, got %v", tt.hasNestedFor, hasFor)
			}
		})
	}
}

// TestIfStatement_NestedSwitch verifies switch statements nested inside if statements
func TestIfStatement_NestedSwitch(t *testing.T) {
	source := `if condition
    switch mode
        1 =>
            x = 1
        2 =>
            x = 2`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("parser creation: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(source))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	ifStmt := script.Statements[0].Core.If
	if ifStmt == nil {
		t.Fatal("expected if statement")
	}

	hasSwitch := false
	for _, stmt := range ifStmt.Body {
		if stmt.Core.Switch != nil {
			hasSwitch = true
			break
		}
	}

	if !hasSwitch {
		t.Error("expected nested switch in if body")
	}
}

// TestForStatement_NestedSwitch verifies switch statements nested inside for loops
func TestForStatement_NestedSwitch(t *testing.T) {
	source := `for i = 1 to 10
    switch i
        5 =>
            x = 1
        =>
            x = 0`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("parser creation: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(source))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	forStmt := script.Statements[0].Core.For
	if forStmt == nil {
		t.Fatal("expected for statement")
	}

	hasSwitch := false
	for _, stmt := range forStmt.Body {
		if stmt.Core.Switch != nil {
			hasSwitch = true
			break
		}
	}

	if !hasSwitch {
		t.Error("expected nested switch in for body")
	}
}

// TestSwitchStatement_ComplexNesting verifies complex nesting patterns with mixed control flow
func TestSwitchStatement_ComplexNesting(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "switch with if and for in different cases",
			source: `switch mode
    1 =>
        if condition
            x = 1
    2 =>
        for i = 1 to 10
            y = i`,
		},
		{
			name: "if containing switch containing for",
			source: `if condition
    switch mode
        1 =>
            for i = 1 to 5
                x = i`,
		},
		{
			name: "for containing switch containing if",
			source: `for i = 1 to 10
    switch i
        5 =>
            if special
                x = 1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("parser creation: %v", err)
			}

			_, err = p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}
		})
	}
}
