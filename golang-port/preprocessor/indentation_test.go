package preprocessor

import (
	"strings"
	"testing"
)

/* Test IfBlockNormalizer functionality */

func TestNormalizeIfBlocks_SingleLineConditionSingleBody(t *testing.T) {
	input := `x = 1
if close > open
    strategy.entry("LONG", strategy.long)
y = 2`

	expected := `x = 1
if close > open
    strategy.entry("LONG", strategy.long)
y = 2`

	result := NormalizeIfBlocks(input)
	if result != expected {
		t.Errorf("Single-line condition + single body failed\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestNormalizeIfBlocks_SingleLineConditionMultipleBodies(t *testing.T) {
	input := `x = 1
if close > open
    strategy.entry("LONG", strategy.long)
    plot(close, color=color.blue)
y = 2`

	expected := `x = 1
if close > open
    strategy.entry("LONG", strategy.long)
if close > open
    plot(close, color=color.blue)
y = 2`

	result := NormalizeIfBlocks(input)
	if result != expected {
		t.Errorf("Single-line condition + multiple bodies failed\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestNormalizeIfBlocks_MultiLineCondition(t *testing.T) {
	input := `x = 1
if close > open and
   volume > volume[1] and
   rsi < 30
    strategy.entry("LONG", strategy.long)
    plot(close, color=color.blue)
y = 2`

	expected := `x = 1
if close > open and volume > volume[1] and rsi < 30
    strategy.entry("LONG", strategy.long)
if close > open and volume > volume[1] and rsi < 30
    plot(close, color=color.blue)
y = 2`

	result := NormalizeIfBlocks(input)
	if result != expected {
		t.Errorf("Multi-line condition failed\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestNormalizeIfBlocks_NestedIf(t *testing.T) {
	input := `x = 1
if close > open
    if volume > 1000
        strategy.entry("LONG", strategy.long)
y = 2`

	expected := `x = 1
if close > open if volume > 1000
    strategy.entry("LONG", strategy.long)
y = 2`

	result := NormalizeIfBlocks(input)
	if result != expected {
		t.Errorf("Nested if failed\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestNormalizeIfBlocks_EmptyLinesInBody(t *testing.T) {
	input := `if close > open
    strategy.entry("LONG", strategy.long)

    plot(close, color=color.blue)
y = 2`

	expected := `if close > open
    strategy.entry("LONG", strategy.long)
if close > open
    plot(close, color=color.blue)
y = 2`

	result := NormalizeIfBlocks(input)
	if result != expected {
		t.Errorf("Empty lines in body failed\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestNormalizeIfBlocks_CommentsInBody(t *testing.T) {
	input := `if close > open
    // Enter long position
    strategy.entry("LONG", strategy.long)
    // Show price
    plot(close, color=color.blue)
y = 2`

	expected := `if close > open
    strategy.entry("LONG", strategy.long)
if close > open
    plot(close, color=color.blue)
y = 2`

	result := NormalizeIfBlocks(input)
	if result != expected {
		t.Errorf("Comments in body failed\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestNormalizeIfBlocks_AssignmentInBody(t *testing.T) {
	input := `if close > open
    x := close
    y = open
y = 2`

	expected := `if close > open
    x := close
if close > open
    y = open
y = 2`

	result := NormalizeIfBlocks(input)
	if result != expected {
		t.Errorf("Assignment in body failed\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestNormalizeIfBlocks_FunctionCallInBody(t *testing.T) {
	input := `if close > open
    ta.sma(close, 20)
    plotshape(true, style=shape.circle)
y = 2`

	expected := `if close > open ta.sma(close, 20)
    plotshape(true, style=shape.circle)
y = 2`

	result := NormalizeIfBlocks(input)
	if result != expected {
		t.Errorf("Function call in body failed\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestNormalizeIfBlocks_NoIfStatements(t *testing.T) {
	input := `x = 1
y = 2
plot(close)`

	expected := input // Should remain unchanged

	result := NormalizeIfBlocks(input)
	if result != expected {
		t.Errorf("No if statements failed\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestNormalizeIfBlocks_IndentationPreserved(t *testing.T) {
	input := `    if close > open
        strategy.entry("LONG", strategy.long)
        plot(close, color=color.blue)`

	expected := `    if close > open
        strategy.entry("LONG", strategy.long)
    if close > open
        plot(close, color=color.blue)`

	result := NormalizeIfBlocks(input)
	if result != expected {
		t.Errorf("Indentation preservation failed\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestLooksLikeBodyStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"strategy.entry(\"LONG\", strategy.long)", true},
		{"plot(close)", true},
		{"x := 10", true},
		{"y = 20", true},
		{"plotshape(true)", true},
		{"ta.sma(close, 20)", false}, // TA calls treated as condition continuations unless prefixed with assignment
		{"close > open", false},      // Condition continuation
		{"and volume > 1000", false}, // Condition continuation
		{"or rsi < 30", false},       // Condition continuation
		{"// comment", false},        // Comment (handled separately)
		{"", false},                  // Empty line
	}

	for _, tt := range tests {
		result := looksLikeBodyStatement(tt.input)
		if result != tt.expected {
			t.Errorf("looksLikeBodyStatement(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestGetIndentation(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"no indent", 0},
		{"  two spaces", 2},
		{"    four spaces", 4},
		{"        eight spaces", 8},
		{"\tone tab", 4},    // Tab = 4 spaces
		{"\t\ttwo tabs", 8}, // 2 tabs = 8 spaces
		{"  \tmixed", 6},    // 2 spaces + 1 tab = 6 spaces
	}

	for _, tt := range tests {
		result := getIndentation(tt.input)
		if result != tt.expected {
			t.Errorf("getIndentation(%q) = %d, expected %d", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeIfBlocks_RealWorldExample(t *testing.T) {
	input := `// Strategy logic
longCondition = close > ta.sma(close, 50) and volume > ta.sma(volume, 20)
if longCondition
    strategy.entry("LONG", strategy.long)
    plot(close, "Entry Price", color=color.green)

shortCondition = close < ta.sma(close, 50)
if shortCondition
    strategy.entry("SHORT", strategy.short)
    plot(close, "Entry Price", color=color.red)
`

	result := NormalizeIfBlocks(input)

	// Verify each if block expanded
	if !strings.Contains(result, "if longCondition\n    strategy.entry") {
		t.Errorf("Expected first if block to be expanded\nGot:\n%s", result)
	}

	if !strings.Contains(result, "if longCondition\n    plot(close") {
		t.Errorf("Expected second statement under first if\nGot:\n%s", result)
	}

	if !strings.Contains(result, "if shortCondition\n    strategy.entry") {
		t.Errorf("Expected third if block to be expanded\nGot:\n%s", result)
	}

	if !strings.Contains(result, "if shortCondition\n    plot(close") {
		t.Errorf("Expected fourth statement under second if\nGot:\n%s", result)
	}
}
