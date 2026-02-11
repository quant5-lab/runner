package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

func TestCommaSupport_RealWorldPatterns(t *testing.T) {
	t.Parallel()
	fixturesDir := filepath.Join("..", "fixtures", "trailing_comma")

	tests := []struct {
		name               string
		fixture            string
		expectedStatements int
	}{
		{
			name:               "assignment with trailing comma",
			fixture:            "01_assignment_trailing_comma.pine",
			expectedStatements: 3,
		},
		{
			name:               "same-line comma-separated statements",
			fixture:            "02_same_line_comma_separation.pine",
			expectedStatements: 5,
		},
		{
			name:               "expression statement with trailing comma",
			fixture:            "03_call_trailing_comma.pine",
			expectedStatements: 3,
		},
		{
			name:               "mixed comma usage",
			fixture:            "04_mixed_comma_usage.pine",
			expectedStatements: 6,
		},
		{
			name:               "reassignment with trailing comma",
			fixture:            "05_reassignment_trailing_comma.pine",
			expectedStatements: 4,
		},
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(fixturesDir, tt.fixture)
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read fixture %s: %v", tt.fixture, err)
			}

			script, err := p.ParseString(path, string(content))
			if err != nil {
				t.Fatalf("Parse failed for %s: %v", tt.fixture, err)
			}

			if script == nil {
				t.Fatal("Expected non-nil script")
			}

			if len(script.Statements) != tt.expectedStatements {
				t.Errorf("Expected %d statements, got %d", tt.expectedStatements, len(script.Statements))
			}
		})
	}
}

func TestCommaSupport_StatementTypes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		source           string
		hasTrailingComma bool
		statementType    string
	}{
		{
			name:             "assignment with trailing comma",
			source:           "//@version=5\nindicator(\"test\")\na = close,",
			hasTrailingComma: true,
			statementType:    "assignment",
		},
		{
			name:             "reassignment with trailing comma",
			source:           "//@version=5\nindicator(\"test\")\nvar x = 0\nx := close,",
			hasTrailingComma: true,
			statementType:    "reassignment",
		},
		{
			name:             "expression statement with trailing comma",
			source:           "//@version=5\nindicator(\"test\")\nplot(close),",
			hasTrailingComma: true,
			statementType:    "expression",
		},
		{
			name:             "tuple assignment with trailing comma",
			source:           "//@version=5\nindicator(\"test\")\n[a, b] = ta.bb(close, 20, 2),",
			hasTrailingComma: true,
			statementType:    "tuple",
		},
		{
			name:             "typed assignment with trailing comma",
			source:           "//@version=5\nindicator(\"test\")\nfloat x = close,",
			hasTrailingComma: true,
			statementType:    "typed",
		},
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			script, err := p.ParseString("test.pine", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(script.Statements) < 2 {
				t.Fatalf("Expected at least 2 statements, got %d", len(script.Statements))
			}

			lastStmt := script.Statements[len(script.Statements)-1]
			hasComma := lastStmt.TrailingComma != nil

			if hasComma != tt.hasTrailingComma {
				t.Errorf("Expected trailing comma = %v, got %v", tt.hasTrailingComma, hasComma)
			}
		})
	}
}

func TestCommaSupport_CommaSeparation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		source             string
		expectedStatements int
		expectedNames      []string
	}{
		{
			name:               "two same-line assignments",
			source:             "//@version=5\nindicator(\"test\")\na = 1, b = 2",
			expectedStatements: 3,
			expectedNames:      []string{"a", "b"},
		},
		{
			name:               "three same-line assignments",
			source:             "//@version=5\nindicator(\"test\")\nx = 1, y = 2, z = 3",
			expectedStatements: 4,
			expectedNames:      []string{"x", "y", "z"},
		},
		{
			name:               "mixed statement types same-line",
			source:             "//@version=5\nindicator(\"test\")\na = 1, plot(a), b = 2",
			expectedStatements: 4,
			expectedNames:      []string{"a", "b"},
		},
		{
			name:               "same-line with trailing comma",
			source:             "//@version=5\nindicator(\"test\")\na = 1, b = 2,",
			expectedStatements: 3,
			expectedNames:      []string{"a", "b"},
		},
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			script, err := p.ParseString("test.pine", tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(script.Statements) != tt.expectedStatements {
				t.Errorf("Expected %d statements, got %d", tt.expectedStatements, len(script.Statements))
			}

			assignmentCount := 0
			for _, stmt := range script.Statements {
				if stmt.Core != nil && stmt.Core.Assignment != nil {
					if assignmentCount < len(tt.expectedNames) {
						expectedName := tt.expectedNames[assignmentCount]
						actualName := stmt.Core.Assignment.Name
						if actualName != expectedName {
							t.Errorf("Assignment %d: expected name '%s', got '%s'",
								assignmentCount, expectedName, actualName)
						}
					}
					assignmentCount++
				}
			}

			if assignmentCount != len(tt.expectedNames) {
				t.Errorf("Expected %d assignments, found %d", len(tt.expectedNames), assignmentCount)
			}
		})
	}
}

func TestCommaSupport_EdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		source      string
		shouldParse bool
	}{
		{
			name:        "trailing comma on last script statement",
			source:      "//@version=5\nindicator(\"test\")\na = 1\nb = 2,",
			shouldParse: true,
		},
		{
			name:        "multiple trailing commas mixed",
			source:      "//@version=5\nindicator(\"test\")\na = 1,\nb = 2,\nc = 3",
			shouldParse: true,
		},
		{
			name:        "comma separation within control flow",
			source:      "//@version=5\nindicator(\"test\")\nif close > open\n    a = 1, b = 2",
			shouldParse: true,
		},
		{
			name:        "trailing comma in nested control flow",
			source:      "//@version=5\nindicator(\"test\")\nif true\n    for i = 0 to 10\n        a = i,",
			shouldParse: true,
		},
		{
			name:        "comma after complex expression",
			source:      "//@version=5\nindicator(\"test\")\nx = close > open ? high : low,",
			shouldParse: true,
		},
		{
			name:        "comma after function call with multiple args",
			source:      "//@version=5\nindicator(\"test\")\nta.sma(close, 20),",
			shouldParse: true,
		},
		{
			name:        "comma after array literal assignment",
			source:      "//@version=5\nindicator(\"test\")\narr = [1, 2, 3],",
			shouldParse: true,
		},
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			script, err := p.ParseString("test.pine", tt.source)

			if tt.shouldParse {
				if err != nil {
					t.Errorf("Expected parse success, got error: %v", err)
				}
				if script == nil {
					t.Error("Expected non-nil script")
				}
			} else {
				if err == nil {
					t.Error("Expected parse error, got success")
				}
			}
		})
	}
}

func TestCommaSupport_BackwardCompatibility(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "no commas - traditional style",
			source: "//@version=5\nindicator(\"test\")\na = 1\nb = 2\nc = 3",
		},
		{
			name:   "statements on separate lines",
			source: "//@version=5\nindicator(\"test\")\nx = close\ny = open\nz = high",
		},
		{
			name: "multiline function",
			source: `//@version=5
indicator("test")
f(x) =>
    a = x + 1
    b = a * 2
    b`,
		},
		{
			name: "control flow without commas",
			source: `//@version=5
indicator("test")
if close > open
    a = 1
    b = 2`,
		},
		{
			name:   "function calls without trailing commas",
			source: "//@version=5\nindicator(\"test\")\nplot(close)\nplot(open)",
		},
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			script, err := p.ParseString("test.pine", tt.source)
			if err != nil {
				t.Errorf("Traditional syntax should parse: %v", err)
			}

			if script == nil {
				t.Error("Expected non-nil script")
				return
			}

			for i, stmt := range script.Statements {
				if stmt.TrailingComma != nil {
					t.Errorf("Statement %d should not have trailing comma in traditional syntax", i)
				}
			}
		})
	}
}
