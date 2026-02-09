package parser

import (
	"strings"
	"testing"
)

/* Switch indentation variant tests: verifies arbitrary indent sizes parse correctly */

func TestSwitchStatement_IndentSizeVariants(t *testing.T) {
	tests := []struct {
		name   string
		indent int
	}{
		{name: "1-space", indent: 1},
		{name: "2-space", indent: 2},
		{name: "3-space", indent: 3},
		{name: "4-space", indent: 4},
		{name: "5-space", indent: 5},
		{name: "6-space", indent: 6},
		{name: "7-space", indent: 7},
		{name: "8-space", indent: 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caseIndent := strings.Repeat(" ", tt.indent)
			bodyIndent := strings.Repeat(" ", tt.indent*2)
			source := "switch val\n" +
				caseIndent + "1 =>\n" +
				bodyIndent + "x = 1\n" +
				caseIndent + "2 =>\n" +
				bodyIndent + "x = 2"

			p, err := NewParser()
			if err != nil {
				t.Fatalf("parser creation: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(source))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			sw := script.Statements[0].Core.Switch
			if sw == nil {
				t.Fatal("expected switch statement")
			}
			if len(sw.Cases) != 2 {
				t.Errorf("expected 2 cases, got %d", len(sw.Cases))
			}
		})
	}
}

func TestSwitchStatement_UnequalIndentLevels(t *testing.T) {
	tests := []struct {
		name       string
		caseIndent int
		bodyIndent int
	}{
		{name: "case=2 body=3", caseIndent: 2, bodyIndent: 3},
		{name: "case=2 body=5", caseIndent: 2, bodyIndent: 5},
		{name: "case=3 body=7", caseIndent: 3, bodyIndent: 7},
		{name: "case=4 body=6", caseIndent: 4, bodyIndent: 6},
		{name: "case=1 body=4", caseIndent: 1, bodyIndent: 4},
		{name: "case=3 body=4", caseIndent: 3, bodyIndent: 4},
		{name: "case=5 body=9", caseIndent: 5, bodyIndent: 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := strings.Repeat(" ", tt.caseIndent)
			bi := strings.Repeat(" ", tt.bodyIndent)
			source := "switch val\n" +
				ci + "1 =>\n" +
				bi + "x = 1\n" +
				ci + "2 =>\n" +
				bi + "x = 2"

			p, err := NewParser()
			if err != nil {
				t.Fatalf("parser creation: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(source))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			sw := script.Statements[0].Core.Switch
			if sw == nil {
				t.Fatal("expected switch statement")
			}
			if len(sw.Cases) != 2 {
				t.Errorf("expected 2 cases, got %d", len(sw.Cases))
			}
		})
	}
}

func TestSwitchStatement_TabIndentation(t *testing.T) {
	source := "switch val\n" +
		"\t1 =>\n" +
		"\t\tx = 1\n" +
		"\t2 =>\n" +
		"\t\tx = 2"

	p, err := NewParser()
	if err != nil {
		t.Fatalf("parser creation: %v", err)
	}

	script, err := p.ParseBytes("test.pine", []byte(source))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	sw := script.Statements[0].Core.Switch
	if sw == nil {
		t.Fatal("expected switch statement")
	}
	if len(sw.Cases) != 2 {
		t.Errorf("expected 2 cases, got %d", len(sw.Cases))
	}
}

func TestSwitchStatement_Form2IndentSizeVariants(t *testing.T) {
	tests := []struct {
		name   string
		indent int
	}{
		{name: "2-space", indent: 2},
		{name: "3-space", indent: 3},
		{name: "4-space", indent: 4},
		{name: "5-space", indent: 5},
		{name: "7-space", indent: 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := strings.Repeat(" ", tt.indent)
			bi := strings.Repeat(" ", tt.indent*2)
			source := "switch\n" +
				ci + "x > 10 =>\n" +
				bi + "a = 1\n" +
				ci + "x > 0 =>\n" +
				bi + "a = 2\n" +
				ci + "=>\n" +
				bi + "a = 0"

			p, err := NewParser()
			if err != nil {
				t.Fatalf("parser creation: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(source))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			sw := script.Statements[0].Core.Switch
			if sw == nil {
				t.Fatal("expected switch statement")
			}
			if len(sw.Cases) != 3 {
				t.Errorf("expected 3 cases, got %d", len(sw.Cases))
			}
		})
	}
}

func TestSwitchStatement_NestedUnequalIndent(t *testing.T) {
	tests := []struct {
		name  string
		outer int
		inner int
		body  int
	}{
		{name: "outer=2 inner=5 body=8", outer: 2, inner: 5, body: 8},
		{name: "outer=3 inner=6 body=9", outer: 3, inner: 6, body: 9},
		{name: "outer=4 inner=8 body=12", outer: 4, inner: 8, body: 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oi := strings.Repeat(" ", tt.outer)
			ii := strings.Repeat(" ", tt.inner)
			bi := strings.Repeat(" ", tt.body)
			source := "if condition\n" +
				oi + "switch val\n" +
				ii + "1 =>\n" +
				bi + "x = 1\n" +
				ii + "2 =>\n" +
				bi + "x = 2"

			p, err := NewParser()
			if err != nil {
				t.Fatalf("parser creation: %v", err)
			}

			_, err = p.ParseBytes("test.pine", []byte(source))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
		})
	}
}

func TestSwitchStatement_MultiStatementBodyUnequalIndent(t *testing.T) {
	tests := []struct {
		name       string
		caseIndent int
		bodyIndent int
		bodyLines  int
	}{
		{name: "case=2 body=5 lines=3", caseIndent: 2, bodyIndent: 5, bodyLines: 3},
		{name: "case=3 body=7 lines=2", caseIndent: 3, bodyIndent: 7, bodyLines: 2},
		{name: "case=1 body=3 lines=4", caseIndent: 1, bodyIndent: 3, bodyLines: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := strings.Repeat(" ", tt.caseIndent)
			bi := strings.Repeat(" ", tt.bodyIndent)

			var bodyLines []string
			for i := 0; i < tt.bodyLines; i++ {
				bodyLines = append(bodyLines, bi+"x = "+string(rune('0'+i)))
			}
			body := strings.Join(bodyLines, "\n")

			source := "switch val\n" +
				ci + "1 =>\n" +
				body + "\n" +
				ci + "2 =>\n" +
				bi + "y = 9"

			p, err := NewParser()
			if err != nil {
				t.Fatalf("parser creation: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(source))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			sw := script.Statements[0].Core.Switch
			if sw == nil {
				t.Fatal("expected switch statement")
			}
			if len(sw.Cases) != 2 {
				t.Errorf("expected 2 cases, got %d", len(sw.Cases))
			}
			if len(sw.Cases[0].Body) != tt.bodyLines {
				t.Errorf("case 0: expected %d body statements, got %d", tt.bodyLines, len(sw.Cases[0].Body))
			}
		})
	}
}
