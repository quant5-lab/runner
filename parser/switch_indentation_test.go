package parser

import (
	"testing"
)

/* Switch statement indentation tests */

func TestSwitchStatement_IndentationLevels(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedCases int
	}{
		{
			name: "two cases",
			source: `switch val
    1 =>
        x = 1
    2 =>
        x = 2`,
			expectedCases: 2,
		},
		{
			name: "three cases",
			source: `switch mode
    1 =>
        doA()
    2 =>
        doB()
    3 =>
        doC()`,
			expectedCases: 3,
		},
		{
			name: "five cases",
			source: `switch state
    1 =>
        a = 1
    2 =>
        a = 2
    3 =>
        a = 3
    4 =>
        a = 4
    5 =>
        a = 5`,
			expectedCases: 5,
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

			switchStmt := script.Statements[0].Core.Switch
			if switchStmt == nil {
				t.Fatal("expected switch statement")
			}

			if len(switchStmt.Cases) != tt.expectedCases {
				t.Errorf("expected %d cases, got %d", tt.expectedCases, len(switchStmt.Cases))
			}
		})
	}
}

func TestSwitchStatement_MultipleSequential(t *testing.T) {
	source := `switch a
    1 =>
        x = 1

switch b
    2 =>
        y = 2

switch c
    3 =>
        z = 3`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("parser creation: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(source))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(script.Statements) != 3 {
		t.Fatalf("expected 3 switch statements, got %d", len(script.Statements))
	}

	for i, stmt := range script.Statements {
		if stmt.Core.Switch == nil {
			t.Errorf("statement %d: expected switch statement", i)
		}
	}
}

func TestSwitchStatement_WithEmptyLines(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "empty line before switch",
			source: `
switch val
    1 =>
        x = 1`,
		},
		{
			name: "empty line after switch",
			source: `switch val
    1 =>
        x = 1

y = 2`,
		},
		{
			name: "empty lines between cases",
			source: `switch val
    1 =>
        x = 1

    2 =>
        x = 2`,
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

func TestSwitchStatement_WithComments(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "comment before switch",
			source: `// Choose action
switch val
    1 =>
        x = 1`,
		},
		{
			name: "comment before case",
			source: `switch val
    // First option
    1 =>
        x = 1`,
		},
		{
			name: "comment in case body",
			source: `switch val
    1 =>
        // Do something
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

func TestSwitchStatement_MixedWithOtherStatements(t *testing.T) {
	source := `a = 1
switch val
    1 =>
        x = 1
    2 =>
        x = 2
b = 2
if condition
    c = 3
for i = 1 to 10
    d = i`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("parser creation: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(source))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(script.Statements) != 5 {
		t.Fatalf("expected 5 statements, got %d", len(script.Statements))
	}

	hasSwitch := script.Statements[1].Core.Switch != nil
	hasIf := script.Statements[3].Core.If != nil
	hasFor := script.Statements[4].Core.For != nil

	if !hasSwitch || !hasIf || !hasFor {
		t.Error("expected switch, if, and for statements in correct positions")
	}
}

func TestSwitchStatement_MultiStatementCaseBodies(t *testing.T) {
	source := `switch mode
    1 =>
        a = 1
        b = 2
        c = 3
    2 =>
        x = 10
        y = 20`

	p, err := NewParser()
	if err != nil {
		t.Fatalf("parser creation: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(source))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	switchStmt := script.Statements[0].Core.Switch
	if switchStmt == nil {
		t.Fatal("expected switch statement")
	}

	if len(switchStmt.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(switchStmt.Cases))
	}

	if len(switchStmt.Cases[0].Body) != 3 {
		t.Errorf("case 0: expected 3 body statements, got %d", len(switchStmt.Cases[0].Body))
	}

	if len(switchStmt.Cases[1].Body) != 2 {
		t.Errorf("case 1: expected 2 body statements, got %d", len(switchStmt.Cases[1].Body))
	}
}

func TestSwitchStatement_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		shouldErr bool
	}{
		{
			name: "switch without indent",
			source: `switch val
1 =>
    x = 1`,
			shouldErr: true,
		},
		{
			name:      "empty switch body",
			source:    `switch val`,
			shouldErr: true,
		},
		{
			name: "case without arrow",
			source: `switch val
    1
        x = 1`,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("parser creation: %v", err)
			}

			_, err = p.ParseBytes("test.pine", []byte(tt.source))
			if tt.shouldErr && err == nil {
				t.Error("expected parse error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}
