package preprocessor

import (
	"fmt"
	"strings"
	"testing"
)

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

	expected := input

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

	if !strings.Contains(result, "if longCondition\n    strategy.entry") {
		t.Errorf("Expected first if block with entry statement\nGot:\n%s", result)
	}

	if !strings.Contains(result, "    plot(close, \"Entry Price\", color=color.green)") {
		t.Errorf("Expected plot statement in first if block\nGot:\n%s", result)
	}

	if !strings.Contains(result, "if shortCondition\n    strategy.entry") {
		t.Errorf("Expected second if block with entry statement\nGot:\n%s", result)
	}

	if !strings.Contains(result, "    plot(close, \"Entry Price\", color=color.red)") {
		t.Errorf("Expected plot statement in second if block\nGot:\n%s", result)
	}

	ifLongCount := strings.Count(result, "if longCondition")
	if ifLongCount != 1 {
		t.Errorf("Expected 1 'if longCondition', got %d\nResult:\n%s", ifLongCount, result)
	}

	ifShortCount := strings.Count(result, "if shortCondition")
	if ifShortCount != 1 {
		t.Errorf("Expected 1 'if shortCondition', got %d\nResult:\n%s", ifShortCount, result)
	}
}

func TestNormalizeIfBlocks_ConditionEvaluationAtomicity(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(result string) error
	}{
		{
			name: "multiple variable reassignments in single condition",
			input: `state = false
state := state[1]
flag = false
flag := flag[1]
if condition
    state := true
    flag := true
x = 1`,
			validate: func(result string) error {
				ifCount := strings.Count(result, "if condition")
				if ifCount != 1 {
					return fmt.Errorf("expected 1 'if condition', got %d", ifCount)
				}
				lines := strings.Split(result, "\n")
				ifLineIdx := -1
				for i, line := range lines {
					if strings.Contains(line, "if condition") {
						ifLineIdx = i
						break
					}
				}
				if ifLineIdx == -1 {
					return fmt.Errorf("if statement not found")
				}
				bodyLines := 0
				for i := ifLineIdx + 1; i < len(lines); i++ {
					trimmed := strings.TrimSpace(lines[i])
					if trimmed == "" || strings.HasPrefix(trimmed, "//") {
						continue
					}
					if strings.HasPrefix(lines[i], "    ") && !strings.HasPrefix(trimmed, "if") {
						bodyLines++
					} else {
						break
					}
				}
				if bodyLines != 2 {
					return fmt.Errorf("expected 2 body statements, got %d", bodyLines)
				}
				return nil
			},
		},
		{
			name: "three assignments in single if block",
			input: `if trigger
    a := 1
    b := 2
    c := 3`,
			validate: func(result string) error {
				if strings.Count(result, "if trigger") != 1 {
					return fmt.Errorf("condition duplicated")
				}
				if !strings.Contains(result, "a := 1") || !strings.Contains(result, "b := 2") || !strings.Contains(result, "c := 3") {
					return fmt.Errorf("missing assignments")
				}
				return nil
			},
		},
		{
			name: "mixed assignment and function calls maintain order",
			input: `if ready
    count := 0
    strategy.entry("LONG", strategy.long)
    active := true
    plot(close)`,
			validate: func(result string) error {
				countIdx := strings.Index(result, "count := 0")
				entryIdx := strings.Index(result, "strategy.entry")
				activeIdx := strings.Index(result, "active := true")
				plotIdx := strings.Index(result, "plot(close)")
				if countIdx == -1 || entryIdx == -1 || activeIdx == -1 || plotIdx == -1 {
					return fmt.Errorf("missing statements")
				}
				if !(countIdx < entryIdx && entryIdx < activeIdx && activeIdx < plotIdx) {
					return fmt.Errorf("statement order not preserved: count=%d entry=%d active=%d plot=%d", countIdx, entryIdx, activeIdx, plotIdx)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeIfBlocks(tt.input)
			if err := tt.validate(result); err != nil {
				t.Errorf("%s\nInput:\n%s\nOutput:\n%s", err, tt.input, result)
			}
		})
	}
}

func TestNormalizeIfBlocks_MultipleConsecutiveIfBlocks(t *testing.T) {
	input := `x = 0
if condition1
    a := 1
    b := 2
if condition2
    c := 3
    d := 4
if condition3
    e := 5
y = 0`

	result := NormalizeIfBlocks(input)

	if1Count := strings.Count(result, "if condition1")
	if2Count := strings.Count(result, "if condition2")
	if3Count := strings.Count(result, "if condition3")

	if if1Count != 1 || if2Count != 1 || if3Count != 1 {
		t.Errorf("Expected each condition exactly once, got: if1=%d, if2=%d, if3=%d\nResult:\n%s", if1Count, if2Count, if3Count, result)
	}

	lines := strings.Split(result, "\n")
	var if1Idx, if2Idx, if3Idx int
	for i, line := range lines {
		if strings.Contains(line, "if condition1") {
			if1Idx = i
		}
		if strings.Contains(line, "if condition2") {
			if2Idx = i
		}
		if strings.Contains(line, "if condition3") {
			if3Idx = i
		}
	}

	if !(if1Idx < if2Idx && if2Idx < if3Idx) {
		t.Errorf("If blocks not in correct order: if1=%d, if2=%d, if3=%d", if1Idx, if2Idx, if3Idx)
	}
}

func TestNormalizeIfBlocks_DifferentIndentationLevels(t *testing.T) {
	input := `if outer
    if inner
        x := 1
        y := 2
    z := 3`

	result := NormalizeIfBlocks(input)

	outerCount := strings.Count(result, "if outer")
	innerCount := strings.Count(result, "if inner")

	if outerCount != 1 {
		t.Errorf("Expected 1 'if outer', got %d", outerCount)
	}
	if innerCount != 1 {
		t.Errorf("Expected 1 'if inner', got %d", innerCount)
	}
}

func TestNormalizeIfBlocks_TabIndentation(t *testing.T) {
	input := "if close > open\n\ta := 1\n\tb := 2\nx = 3"

	result := NormalizeIfBlocks(input)

	ifCount := strings.Count(result, "if close > open")
	if ifCount != 1 {
		t.Errorf("Expected 1 if statement with tab indentation, got %d\nResult:\n%s", ifCount, result)
	}

	if !strings.Contains(result, "a := 1") || !strings.Contains(result, "b := 2") {
		t.Errorf("Assignments not preserved with tab indentation\nResult:\n%s", result)
	}
}

func TestNormalizeIfBlocks_MixedTabSpaceIndentation(t *testing.T) {
	input := "if trigger\n  \tx := 1\n  \ty := 2"

	result := NormalizeIfBlocks(input)

	ifCount := strings.Count(result, "if trigger")
	if ifCount != 1 {
		t.Errorf("Expected 1 if statement with mixed indentation, got %d", ifCount)
	}
}

func TestNormalizeIfBlocks_EmptyIfBlock(t *testing.T) {
	input := `x = 1
if condition
y = 2`

	result := NormalizeIfBlocks(input)

	if strings.Contains(result, "if condition") {
		t.Errorf("Empty if block should not generate if statement\nResult:\n%s", result)
	}
}

func TestNormalizeIfBlocks_OnlyCommentsInBody(t *testing.T) {
	input := `if condition
    // This is a comment
    // Another comment
x = 1`

	result := NormalizeIfBlocks(input)

	if strings.Contains(result, "if condition") {
		t.Errorf("If block with only comments should not generate if statement\nResult:\n%s", result)
	}
}

func TestNormalizeIfBlocks_TrailingEmptyLines(t *testing.T) {
	input := `if condition
    x := 1
    y := 2


z = 3`

	result := NormalizeIfBlocks(input)

	ifCount := strings.Count(result, "if condition")
	if ifCount != 1 {
		t.Errorf("Expected 1 if statement, got %d\nResult:\n%s", ifCount, result)
	}

	lines := strings.Split(result, "\n")
	bodyStatements := 0
	inIfBlock := false
	for _, line := range lines {
		if strings.Contains(line, "if condition") {
			inIfBlock = true
			continue
		}
		if inIfBlock {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			if strings.HasPrefix(line, "    ") && !strings.HasPrefix(trimmed, "//") {
				bodyStatements++
			} else {
				break
			}
		}
	}

	if bodyStatements != 2 {
		t.Errorf("Expected 2 body statements, got %d", bodyStatements)
	}
}

func TestNormalizeIfBlocks_ComplexMultiLineCondition(t *testing.T) {
	input := `if close > open and
   high > high[1] and
   low > low[1] and
   volume > volume[1] and
   rsi < 30
    entry := true
    active := true
    count := count + 1`

	result := NormalizeIfBlocks(input)

	ifCount := strings.Count(result, "if close > open and high > high[1] and low > low[1] and volume > volume[1] and rsi < 30")
	if ifCount != 1 {
		t.Errorf("Expected 1 collapsed condition, got %d\nResult:\n%s", ifCount, result)
	}

	if !strings.Contains(result, "entry := true") || !strings.Contains(result, "active := true") || !strings.Contains(result, "count := count + 1") {
		t.Errorf("Not all assignments present\nResult:\n%s", result)
	}
}

func TestNormalizeIfBlocks_ConsecutiveIfBlocksNoBlankLine(t *testing.T) {
	input := `if condition1
    x := 1
if condition2
    y := 2`

	result := NormalizeIfBlocks(input)

	if1Count := strings.Count(result, "if condition1")
	if2Count := strings.Count(result, "if condition2")

	if if1Count != 1 || if2Count != 1 {
		t.Errorf("Expected each condition once, got: if1=%d, if2=%d\nResult:\n%s", if1Count, if2Count, result)
	}
}

func TestNormalizeIfBlocks_ReassignmentOfSameVariable(t *testing.T) {
	input := `if trigger
    state := false
    state := true
    state := maybe`

	result := NormalizeIfBlocks(input)

	ifCount := strings.Count(result, "if trigger")
	if ifCount != 1 {
		t.Errorf("Expected 1 if statement, got %d\nResult:\n%s", ifCount, result)
	}

	stateCount := strings.Count(result, "state :=")
	if stateCount != 3 {
		t.Errorf("Expected 3 state assignments, got %d\nResult:\n%s", stateCount, result)
	}
}

func TestNormalizeIfBlocks_AtEndOfFile(t *testing.T) {
	input := `x = 1
if condition
    y := 2
    z := 3`

	result := NormalizeIfBlocks(input)

	ifCount := strings.Count(result, "if condition")
	if ifCount != 1 {
		t.Errorf("Expected 1 if statement at EOF, got %d\nResult:\n%s", ifCount, result)
	}
}

func TestNormalizeIfBlocks_BodyStatementClassification(t *testing.T) {
	tests := []struct {
		name                string
		input               string
		expectedIfCount     int
		expectedBodyStmtMin int
	}{
		{
			name: "strategy namespace calls",
			input: `if ready
    strategy.entry("L", strategy.long)
    strategy.exit("X", "L")
    strategy.close("L")`,
			expectedIfCount:     1,
			expectedBodyStmtMin: 3,
		},
		{
			name: "plot namespace calls",
			input: `if show
    plot(close)
    plotshape(true, style=shape.circle)
    plotchar(close, char="A")`,
			expectedIfCount:     1,
			expectedBodyStmtMin: 3,
		},
		{
			name: "mixed assignments",
			input: `if active
    x := 1
    y = 2
    z := x + y`,
			expectedIfCount:     1,
			expectedBodyStmtMin: 3,
		},
		{
			name: "user function calls",
			input: `if trigger
    myFunc(a, b)
    otherFunc()`,
			expectedIfCount:     1,
			expectedBodyStmtMin: 2,
		},
		{
			name: "control flow keywords",
			input: `if done
    break`,
			expectedIfCount:     1,
			expectedBodyStmtMin: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeIfBlocks(tt.input)
			ifCount := strings.Count(result, "if ")
			if ifCount != tt.expectedIfCount {
				t.Errorf("Expected %d if statements, got %d\nResult:\n%s", tt.expectedIfCount, ifCount, result)
			}

			lines := strings.Split(result, "\n")
			bodyStmts := 0
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(line, "    ") && trimmed != "" && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "if ") {
					bodyStmts++
				}
			}

			if bodyStmts < tt.expectedBodyStmtMin {
				t.Errorf("Expected at least %d body statements, got %d\nResult:\n%s", tt.expectedBodyStmtMin, bodyStmts, result)
			}
		})
	}
}

func TestNormalizeIfBlocks_WhitespaceNormalization(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "leading empty lines before body",
			input: `if condition

    x := 1
    y := 2`,
		},
		{
			name: "trailing empty lines after body",
			input: `if condition
    x := 1
    y := 2

z = 3`,
		},
		{
			name: "multiple empty lines between statements",
			input: `if condition
    x := 1


    y := 2`,
		},
		{
			name: "mixed empty and comment lines",
			input: `if condition

    // comment
    x := 1

    // another
    y := 2`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeIfBlocks(tt.input)
			ifCount := strings.Count(result, "if condition")
			if ifCount != 1 {
				t.Errorf("Whitespace handling failed: expected 1 if statement, got %d\nResult:\n%s", ifCount, result)
			}

			if !strings.Contains(result, "x := 1") || !strings.Contains(result, "y := 2") {
				t.Errorf("Statements not preserved\nResult:\n%s", result)
			}
		})
	}
}

/* break/continue inside if-body are preserved, not absorbed into condition */
func TestNormalizeIfBlocks_ControlFlowKeywordsPreserved(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "break preserved in if body",
			input: `if i > 5
    break`,
			expected: `if i > 5
    break`,
		},
		{
			name: "continue preserved in if body",
			input: `if i == 3
    continue`,
			expected: `if i == 3
    continue`,
		},
		{
			name: "break with assignment in same body",
			input: `if done
    x := 1
    break`,
			expected: `if done
    x := 1
    break`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeIfBlocks(tt.input)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}
