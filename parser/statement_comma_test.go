package parser

import (
	"testing"
)

func TestStatementComma_SyntaxVariations(t *testing.T) {
	tests := []struct {
		name             string
		source           string
		expectedStmts    int
		hasTrailingComma bool
	}{
		{
			name:             "no comma",
			source:           "//@version=5\nindicator(\"t\")\na = 1",
			expectedStmts:    2,
			hasTrailingComma: false,
		},
		{
			name:             "trailing comma",
			source:           "//@version=5\nindicator(\"t\")\na = 1,",
			expectedStmts:    2,
			hasTrailingComma: true,
		},
		{
			name:          "comma separator",
			source:        "//@version=5\nindicator(\"t\")\na = 1, b = 2",
			expectedStmts: 3,
		},
		{
			name:             "comma separator with trailing",
			source:           "//@version=5\nindicator(\"t\")\na = 1, b = 2,",
			expectedStmts:    3,
			hasTrailingComma: true,
		},
		{
			name:          "three comma-separated",
			source:        "//@version=5\nindicator(\"t\")\na = 1, b = 2, c = 3",
			expectedStmts: 4,
		},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() failed: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, err := p.ParseString("test.pine", tt.source)
			if err != nil {
				t.Fatalf("ParseString() failed: %v", err)
			}

			if len(script.Statements) != tt.expectedStmts {
				t.Errorf("Expected %d statements, got %d", tt.expectedStmts, len(script.Statements))
			}

			lastStmt := script.Statements[len(script.Statements)-1]
			hasComma := lastStmt.TrailingComma != nil

			if hasComma != tt.hasTrailingComma {
				t.Errorf("Expected hasTrailingComma=%v, got %v", tt.hasTrailingComma, hasComma)
			}
		})
	}
}

func TestStatementComma_StatementTypes(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		stmtType    string
		expectComma bool
	}{
		{
			name:        "assignment",
			source:      "//@version=5\nindicator(\"t\")\na = 1,",
			stmtType:    "assignment",
			expectComma: true,
		},
		{
			name:        "reassignment",
			source:      "//@version=5\nindicator(\"t\")\nvar x = 0\nx := 1,",
			stmtType:    "reassignment",
			expectComma: true,
		},
		{
			name:        "expression",
			source:      "//@version=5\nindicator(\"t\")\nplot(close),",
			stmtType:    "expression",
			expectComma: true,
		},
		{
			name:        "tuple assignment",
			source:      "//@version=5\nindicator(\"t\")\n[a, b] = ta.bb(close, 20, 2),",
			stmtType:    "tuple",
			expectComma: true,
		},
		{
			name:        "typed assignment",
			source:      "//@version=5\nindicator(\"t\")\nfloat x = 1,",
			stmtType:    "typed",
			expectComma: true,
		},
		{
			name:        "if statement",
			source:      "//@version=5\nindicator(\"t\")\nif true\n    a = 1",
			stmtType:    "if",
			expectComma: false,
		},
		{
			name:        "for statement",
			source:      "//@version=5\nindicator(\"t\")\nfor i = 0 to 10\n    a = i",
			stmtType:    "for",
			expectComma: false,
		},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() failed: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, err := p.ParseString("test.pine", tt.source)
			if err != nil {
				t.Fatalf("ParseString() failed: %v", err)
			}

			if len(script.Statements) == 0 {
				t.Fatal("Expected at least one statement")
			}

			lastStmt := script.Statements[len(script.Statements)-1]
			hasComma := lastStmt.TrailingComma != nil

			if hasComma != tt.expectComma {
				t.Errorf("Expected comma=%v, got %v", tt.expectComma, hasComma)
			}

			verifyStatementType(t, lastStmt, tt.stmtType)
		})
	}
}

func TestStatementComma_NestedContexts(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "comma in if block",
			source: `//@version=5
indicator("t")
if close > open
    a = 1, b = 2`,
		},
		{
			name: "trailing comma in if block",
			source: `//@version=5
indicator("t")
if close > open
    a = 1,`,
		},
		{
			name: "comma in for loop",
			source: `//@version=5
indicator("t")
for i = 0 to 10
    x = i, y = i * 2`,
		},
		{
			name: "nested if with comma",
			source: `//@version=5
indicator("t")
if true
    if false
        a = 1,`,
		},
		{
			name: "for in if with comma",
			source: `//@version=5
indicator("t")
if true
    for i = 0 to 5
        x = i,`,
		},
		{
			name: "function with commas",
			source: `//@version=5
indicator("t")
f(x) =>
    a = x, b = x * 2
    b`,
		},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() failed: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, err := p.ParseString("test.pine", tt.source)
			if err != nil {
				t.Errorf("ParseString() failed for nested context: %v", err)
			}

			if script == nil {
				t.Error("Expected non-nil script")
			}
		})
	}
}

func TestStatementComma_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		shouldParse bool
	}{
		{
			name:        "comma at script end",
			source:      "//@version=5\nindicator(\"t\")\na = 1,",
			shouldParse: true,
		},
		{
			name:        "multiple trailing commas across statements",
			source:      "//@version=5\nindicator(\"t\")\na = 1,\nb = 2,\nc = 3,",
			shouldParse: true,
		},
		{
			name:        "mixed comma usage",
			source:      "//@version=5\nindicator(\"t\")\na = 1,\nb = 2, c = 3\nd = 4,",
			shouldParse: true,
		},
		{
			name:        "comma after ternary",
			source:      "//@version=5\nindicator(\"t\")\nx = close > open ? high : low,",
			shouldParse: true,
		},
		{
			name:        "comma after array",
			source:      "//@version=5\nindicator(\"t\")\narr = [1, 2, 3],",
			shouldParse: true,
		},
		{
			name:        "comma after complex call",
			source:      "//@version=5\nindicator(\"t\")\nta.sma(ta.ema(close, 10), 20),",
			shouldParse: true,
		},
		{
			name:        "empty statement not created by comma",
			source:      "//@version=5\nindicator(\"t\")\na = 1, b = 2",
			shouldParse: true,
		},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() failed: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, err := p.ParseString("test.pine", tt.source)

			if tt.shouldParse {
				if err != nil {
					t.Errorf("Expected parse success, got: %v", err)
				}
				if script == nil {
					t.Error("Expected non-nil script")
				}
			} else {
				if err == nil {
					t.Error("Expected parse failure, got success")
				}
			}
		})
	}
}

func TestStatementComma_ConverterBehavior(t *testing.T) {
	tests := []struct {
		name             string
		source           string
		expectedVarNames []string
	}{
		{
			name:             "single assignment with comma",
			source:           "//@version=5\nindicator(\"t\")\na = 1,",
			expectedVarNames: []string{"a"},
		},
		{
			name:             "comma-separated assignments",
			source:           "//@version=5\nindicator(\"t\")\na = 1, b = 2, c = 3",
			expectedVarNames: []string{"a", "b", "c"},
		},
		{
			name:             "mixed statements with commas",
			source:           "//@version=5\nindicator(\"t\")\nx = 1, y = 2,\nz = 3",
			expectedVarNames: []string{"x", "y", "z"},
		},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() failed: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, err := p.ParseString("test.pine", tt.source)
			if err != nil {
				t.Fatalf("ParseString() failed: %v", err)
			}

			varCount := 0
			for _, stmt := range script.Statements {
				if stmt.Core != nil && stmt.Core.Assignment != nil {
					if varCount < len(tt.expectedVarNames) {
						expectedName := tt.expectedVarNames[varCount]
						actualName := stmt.Core.Assignment.Name
						if actualName != expectedName {
							t.Errorf("Variable %d: expected '%s', got '%s'",
								varCount, expectedName, actualName)
						}
					}
					varCount++
				}
			}

			if varCount != len(tt.expectedVarNames) {
				t.Errorf("Expected %d variables, got %d", len(tt.expectedVarNames), varCount)
			}
		})
	}
}

func TestStatementComma_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "traditional newline separation",
			source: "//@version=5\nindicator(\"t\")\na = 1\nb = 2\nc = 3",
		},
		{
			name: "multiline function without commas",
			source: `//@version=5
indicator("t")
f(x) =>
    a = x + 1
    b = a * 2
    b`,
		},
		{
			name: "control flow without commas",
			source: `//@version=5
indicator("t")
if close > open
    a = 1
    b = 2`,
		},
		{
			name: "loop without commas",
			source: `//@version=5
indicator("t")
for i = 0 to 10
    x = i
    y = i * 2`,
		},
		{
			name:   "sequential calls without commas",
			source: "//@version=5\nindicator(\"t\")\nplot(close)\nplot(open)\nplot(high)",
		},
	}

	p, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser() failed: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
					t.Errorf("Statement %d: traditional syntax should not have trailing comma", i)
				}
			}
		})
	}
}

func verifyStatementType(t *testing.T, stmt *Statement, expectedType string) {
	t.Helper()

	if stmt.Core == nil {
		t.Fatal("Statement.Core is nil")
	}

	switch expectedType {
	case "assignment":
		if stmt.Core.Assignment == nil {
			t.Error("Expected assignment statement")
		}
	case "reassignment":
		if stmt.Core.Reassignment == nil {
			t.Error("Expected reassignment statement")
		}
	case "expression":
		if stmt.Core.Expression == nil {
			t.Error("Expected expression statement")
		}
	case "tuple":
		if stmt.Core.TupleAssignment == nil {
			t.Error("Expected tuple assignment statement")
		}
	case "typed":
		if stmt.Core.TypedAssignment == nil {
			t.Error("Expected typed assignment statement")
		}
	case "if":
		if stmt.Core.If == nil {
			t.Error("Expected if statement")
		}
	case "for":
		if stmt.Core.For == nil {
			t.Error("Expected for statement")
		}
	default:
		t.Errorf("Unknown statement type: %s", expectedType)
	}
}
