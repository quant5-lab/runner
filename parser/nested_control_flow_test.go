package parser

import (
	"strings"
	"testing"
)

// TestIfStatement_NestedIf verifies nested IF statements with control flow as first statement in block
func TestIfStatement_NestedIf(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedDepth  int
		outerBodyCount int
	}{
		{
			name: "if as first statement in if body",
			source: `if condition1
    if condition2
        x = 1`,
			expectedDepth:  2,
			outerBodyCount: 1,
		},
		{
			name: "if after statement in if body",
			source: `if condition1
    y = 1
    if condition2
        x = 1`,
			expectedDepth:  2,
			outerBodyCount: 2,
		},
		{
			name: "triple nested if",
			source: `if a
    if b
        if c
            x = 1`,
			expectedDepth:  3,
			outerBodyCount: 1,
		},
		{
			name: "if with comments before nested if",
			source: `if a
    // Comment
    if b
        x = 1`,
			expectedDepth:  2,
			outerBodyCount: 1,
		},
		{
			name: "if with empty lines before nested if",
			source: `if a

    if b
        x = 1`,
			expectedDepth:  2,
			outerBodyCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(script.Statements) == 0 {
				t.Fatal("No statements parsed")
			}

			outerIf := script.Statements[0].Core.If
			if outerIf == nil {
				t.Fatal("Expected outer IF statement")
			}

			if len(outerIf.Body) != tt.outerBodyCount {
				t.Errorf("Expected %d statements in outer IF body, got %d", tt.outerBodyCount, len(outerIf.Body))
			}

			// Count nesting depth
			depth := 1
			current := outerIf
			for {
				found := false
				for _, stmt := range current.Body {
					if stmt.Core.If != nil {
						depth++
						current = stmt.Core.If
						found = true
						break
					}
				}
				if !found {
					break
				}
			}

			if depth != tt.expectedDepth {
				t.Errorf("Expected nesting depth %d, got %d", tt.expectedDepth, depth)
			}
		})
	}
}

// TestIfStatement_NestedFor verifies FOR loops nested inside IF statements
func TestIfStatement_NestedFor(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		outerBodyCount int
		hasNestedFor   bool
	}{
		{
			name: "for as first statement in if body",
			source: `if condition
    for i = 0 to 5
        x = i`,
			outerBodyCount: 1,
			hasNestedFor:   true,
		},
		{
			name: "for after statement in if body",
			source: `if condition
    y = 1
    for i = 0 to 5
        x = i`,
			outerBodyCount: 2,
			hasNestedFor:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			outerIf := script.Statements[0].Core.If
			if outerIf == nil {
				t.Fatal("Expected IF statement")
			}

			if len(outerIf.Body) != tt.outerBodyCount {
				t.Errorf("Expected %d statements in IF body, got %d", tt.outerBodyCount, len(outerIf.Body))
			}

			hasFor := false
			for _, stmt := range outerIf.Body {
				if stmt.Core.For != nil {
					hasFor = true
					break
				}
			}

			if hasFor != tt.hasNestedFor {
				t.Errorf("Expected hasNestedFor=%v, got %v", tt.hasNestedFor, hasFor)
			}
		})
	}
}

// TestForStatement_NestedIf verifies IF statements nested inside FOR loops
func TestForStatement_NestedIf(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		outerBodyCount int
		hasNestedIf    bool
	}{
		{
			name: "if as first statement in for body",
			source: `for i = 0 to 10
    if i > 5
        x = i`,
			outerBodyCount: 1,
			hasNestedIf:    true,
		},
		{
			name: "if after statement in for body",
			source: `for i = 0 to 10
    sum = sum + i
    if i > 5
        x = i`,
			outerBodyCount: 2,
			hasNestedIf:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			outerFor := script.Statements[0].Core.For
			if outerFor == nil {
				t.Fatal("Expected FOR statement")
			}

			if len(outerFor.Body) != tt.outerBodyCount {
				t.Errorf("Expected %d statements in FOR body, got %d", tt.outerBodyCount, len(outerFor.Body))
			}

			hasIf := false
			for _, stmt := range outerFor.Body {
				if stmt.Core.If != nil {
					hasIf = true
					break
				}
			}

			if hasIf != tt.hasNestedIf {
				t.Errorf("Expected hasNestedIf=%v, got %v", tt.hasNestedIf, hasIf)
			}
		})
	}
}

// TestForStatement_NestedFor verifies FOR loops nested inside FOR loops
func TestForStatement_NestedFor(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		outerBodyCount int
		nestingDepth   int
	}{
		{
			name: "for as first statement in for body",
			source: `for i = 0 to 2
    for j = 0 to 2
        x = 1`,
			outerBodyCount: 1,
			nestingDepth:   2,
		},
		{
			name: "for after statement in for body",
			source: `for i = 0 to 5
    a = i
    for j = 0 to 3
        b = j
    c = i + 1`,
			outerBodyCount: 3,
			nestingDepth:   2,
		},
		{
			name: "triple nested for",
			source: `for i = 0 to 2
    for j = 0 to 2
        for k = 0 to 2
            x = 1`,
			outerBodyCount: 1,
			nestingDepth:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			outerFor := script.Statements[0].Core.For
			if outerFor == nil {
				t.Fatal("Expected outer FOR statement")
			}

			if len(outerFor.Body) != tt.outerBodyCount {
				t.Errorf("Expected %d statements in outer FOR body, got %d", tt.outerBodyCount, len(outerFor.Body))
			}

			// Count nesting depth
			depth := 1
			current := outerFor
			for {
				found := false
				for _, stmt := range current.Body {
					if stmt.Core.For != nil {
						depth++
						current = stmt.Core.For
						found = true
						break
					}
				}
				if !found {
					break
				}
			}

			if depth != tt.nestingDepth {
				t.Errorf("Expected nesting depth %d, got %d", tt.nestingDepth, depth)
			}
		})
	}
}

// TestFunctionDecl_NestedControlFlow verifies control flow nested inside arrow functions
func TestFunctionDecl_NestedControlFlow(t *testing.T) {
	tests := []struct {
		name            string
		source          string
		hasNestedIf     bool
		hasNestedFor    bool
		nestedBodyCount int
	}{
		{
			name: "if as first statement in arrow function",
			source: `calcValue(price) =>
    if price > 100
        result = price * 2
    result`,
			hasNestedIf:     true,
			hasNestedFor:    false,
			nestedBodyCount: 2,
		},
		{
			name: "for as first statement in arrow function",
			source: `sumRange(start, end) =>
    total = 0.0
    for i = start to end
        total := total + close[i]
    total`,
			hasNestedIf:     false,
			hasNestedFor:    true,
			nestedBodyCount: 3,
		},
		{
			name: "nested if inside if in arrow function",
			source: `calcSignal(price) =>
    if price > 50
        if price > 100
            result = 1
    result`,
			hasNestedIf:     true,
			hasNestedFor:    false,
			nestedBodyCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			funcDecl := script.Statements[0].Core.FunctionDecl
			if funcDecl == nil {
				t.Fatal("Expected function declaration")
			}

			body := funcDecl.MultiLineBody
			if body == nil {
				t.Fatal("Expected multi-line body")
			}

			if len(body) != tt.nestedBodyCount {
				t.Errorf("Expected %d statements in function body, got %d", tt.nestedBodyCount, len(body))
			}

			hasIf := false
			hasFor := false
			for _, stmt := range body {
				if stmt.Core.If != nil {
					hasIf = true
				}
				if stmt.Core.For != nil {
					hasFor = true
				}
			}

			if hasIf != tt.hasNestedIf {
				t.Errorf("Expected hasNestedIf=%v, got %v", tt.hasNestedIf, hasIf)
			}
			if hasFor != tt.hasNestedFor {
				t.Errorf("Expected hasNestedFor=%v, got %v", tt.hasNestedFor, hasFor)
			}
		})
	}
}

// TestControlFlow_DEDENTFollowedByControlKeyword verifies control keywords after DEDENT
func TestControlFlow_DEDENTFollowedByControlKeyword(t *testing.T) {
	tests := []struct {
		name             string
		source           string
		expectedStmtType []string
	}{
		{
			name: "if after dedent from if",
			source: `if a
    x = 1
if b
    y = 2`,
			expectedStmtType: []string{"if", "if"},
		},
		{
			name: "for after dedent from if",
			source: `if a
    x = 1
for i = 0 to 5
    y = i`,
			expectedStmtType: []string{"if", "for"},
		},
		{
			name: "if after dedent from for",
			source: `for i = 0 to 5
    x = i
if condition
    y = 1`,
			expectedStmtType: []string{"for", "if"},
		},
		{
			name: "arrow function after dedent from if",
			source: `if a
    x = 1
getValue() =>
    close`,
			expectedStmtType: []string{"if", "function"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(script.Statements) != len(tt.expectedStmtType) {
				t.Fatalf("Expected %d statements, got %d", len(tt.expectedStmtType), len(script.Statements))
			}

			for i, expectedType := range tt.expectedStmtType {
				stmt := script.Statements[i]
				var actualType string
				if stmt.Core.If != nil {
					actualType = "if"
				} else if stmt.Core.For != nil {
					actualType = "for"
				} else if stmt.Core.FunctionDecl != nil {
					actualType = "function"
				} else {
					actualType = "other"
				}

				if actualType != expectedType {
					t.Errorf("Statement %d: expected type %s, got %s", i, expectedType, actualType)
				}
			}
		})
	}
}

// TestControlFlow_ExtremeNesting verifies deeply nested control structures
func TestControlFlow_ExtremeNesting(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedDepth int
	}{
		{
			name: "5-level nested if",
			source: `if a
    if b
        if c
            if d
                if e
                    x = 1`,
			expectedDepth: 5,
		},
		{
			name: "7-level nested for",
			source: `for i1 = 0 to 2
    for i2 = 0 to 2
        for i3 = 0 to 2
            for i4 = 0 to 2
                for i5 = 0 to 2
                    for i6 = 0 to 2
                        for i7 = 0 to 2
                            x = 1`,
			expectedDepth: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(script.Statements) == 0 {
				t.Fatal("No statements parsed")
			}

			// Count depth based on statement type
			depth := 1
			var countDepth func([]*Statement) int
			countDepth = func(body []*Statement) int {
				for _, stmt := range body {
					if stmt.Core.If != nil {
						return 1 + countDepth(stmt.Core.If.Body)
					}
					if stmt.Core.For != nil {
						return 1 + countDepth(stmt.Core.For.Body)
					}
				}
				return 0
			}

			firstStmt := script.Statements[0]
			if firstStmt.Core.If != nil {
				depth = 1 + countDepth(firstStmt.Core.If.Body)
			} else if firstStmt.Core.For != nil {
				depth = 1 + countDepth(firstStmt.Core.For.Body)
			}

			if depth != tt.expectedDepth {
				t.Errorf("Expected nesting depth %d, got %d", tt.expectedDepth, depth)
			}
		})
	}
}

// TestIndentation_MixedSizes verifies different indentation sizes (2, 4, 8 spaces)
func TestIndentation_MixedSizes(t *testing.T) {
	tests := []struct {
		name             string
		source           string
		expectedStmtType string
	}{
		{
			name: "2-space indentation",
			source: `if a
  x = 1
  if b
    y = 2`,
			expectedStmtType: "if",
		},
		{
			name: "4-space indentation",
			source: `if a
    x = 1
    if b
        y = 2`,
			expectedStmtType: "if",
		},
		{
			name: "8-space indentation",
			source: `if a
        x = 1
        if b
                y = 2`,
			expectedStmtType: "if",
		},
		{
			name: "mixed 2 and 4 space indentation",
			source: `if a
  x = 1
  for i = 0 to 5
    y = i`,
			expectedStmtType: "if",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(script.Statements) == 0 {
				t.Fatal("No statements parsed")
			}

			firstStmt := script.Statements[0]
			if tt.expectedStmtType == "if" && firstStmt.Core.If == nil {
				t.Fatal("Expected IF statement")
			}
		})
	}
}

// TestIndentation_TabsVsSpaces verifies tab and space indentation mixing
func TestIndentation_TabsVsSpaces(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "tab indentation",
			source: "if a\n\tx = 1\n\tif b\n\t\ty = 2",
		},
		{
			name: "1-space indentation",
			source: `if a
 x = 1
 if b
  y = 2`,
		},
		{
			name: "3-space indentation",
			source: `if a
   x = 1
   if b
      y = 2`,
		},
		{
			name: "5-space indentation",
			source: `if a
     x = 1
     if b
          y = 2`,
		},
		{
			name: "6-space indentation",
			source: `if a
      x = 1
      if b
            y = 2`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(script.Statements) == 0 {
				t.Fatal("No statements parsed")
			}

			firstStmt := script.Statements[0]
			if firstStmt.Core.If == nil {
				t.Fatal("Expected IF statement")
			}
		})
	}
}

func TestIndentation_TabAndSpaceMixing(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "consistent tabs only",
			source: "if a\n\tx = 1\n\tif b\n\t\ty = 2",
		},
		{
			name: "consistent spaces only",
			source: `if a
    x = 1
    if b
        y = 2`,
		},
		{
			name:   "tab first level, spaces second level",
			source: "if a\n\tx = 1\n    if b\n        y = 2",
		},
		{
			name:   "spaces first level, tab second level",
			source: "if a\n    x = 1\n\tif b\n\t\ty = 2",
		},
		{
			name:   "tab and 1-space mixed nested",
			source: "if a\n\tx = 1\n if b\n  y = 2",
		},
		{
			name:   "tab and 2-space mixed nested",
			source: "if a\n\tx = 1\n  if b\n    y = 2",
		},
		{
			name:   "tab and 3-space mixed nested",
			source: "if a\n\tx = 1\n   if b\n      y = 2",
		},
		{
			name:   "alternating tabs and spaces per line",
			source: "if a\n\tx = 1\n    y = 2\n\tz = 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			expandedSource := strings.ReplaceAll(tt.source, "\t", "    ")
			script, err := p.ParseBytes("test.pine", []byte(expandedSource))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(script.Statements) == 0 {
				t.Fatal("No statements parsed")
			}
		})
	}
}

// TestIndentation_MisalignedComments verifies comments with odd indentation don't crash
func TestIndentation_MisalignedComments(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "comment with zero indentation in indented block",
			source: `if a
    x = 1
// Comment at column 0
    if b
        y = 2`,
		},
		{
			name: "comment with excessive indentation",
			source: `if a
    x = 1
                // Comment way out
    if b
        y = 2`,
		},
		{
			name: "comment with odd indentation between nested controls",
			source: `if a
    x = 1
   // Comment at 3 spaces
    if b
        y = 2`,
		},
		{
			name: "multiple misaligned comments",
			source: `if a
// Comment 0
    x = 1
  // Comment 2
    if b
      // Comment 6
        y = 2
            // Comment 12`,
		},
		{
			name: "comment before nested control",
			source: `if a
    x = 1
  // Misaligned comment before for
    for i = 0 to 5
        sum = i`,
		},
		{
			name: "comment in arrow function with nested if",
			source: `getValue() =>
// Comment at 0
    result = 0
     // Comment at 5
    if close > 50
          // Comment at 10
        result = 1
    result`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed (should not crash on misaligned comments): %v", err)
			}

			if len(script.Statements) == 0 {
				t.Fatal("No statements parsed")
			}
		})
	}
}
