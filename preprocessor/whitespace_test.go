package preprocessor

import (
	"strings"
	"testing"
)

func TestExpandTabs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no tabs preserves input",
			input:    "if condition\n    x = 1\n    y = 2",
			expected: "if condition\n    x = 1\n    y = 2",
		},
		{
			name:     "leading single tab",
			input:    "\tx = 1",
			expected: "    x = 1",
		},
		{
			name:     "leading multiple tabs create nested indentation",
			input:    "\tx = 1\n\t\ty = 2\n\t\t\tz = 3",
			expected: "    x = 1\n        y = 2\n            z = 3",
		},
		{
			name:     "tabs and spaces on different lines maintain structure",
			input:    "\tx = 1\n    y = 2",
			expected: "    x = 1\n    y = 2",
		},
		{
			name:     "mid-line tabs expand to four spaces",
			input:    "x\t=\t1",
			expected: "x    =    1",
		},
		{
			name:     "tabs at line end expand correctly",
			input:    "line1\t",
			expected: "line1    ",
		},
		{
			name:     "consecutive tabs multiply correctly",
			input:    "x\t\t=\t\t\t1",
			expected: "x        =            1",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only tabs",
			input:    "\t\t\t",
			expected: "            ",
		},
		{
			name:     "only spaces preserved",
			input:    "    ",
			expected: "    ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandTabs(tt.input)
			if result != tt.expected {
				t.Errorf("Input:    %q\nGot:      %q\nExpected: %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExpandTabs_LineEndings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "preserves LF newlines",
			input:    "line1\t\nline2\t\nline3",
			expected: "line1    \nline2    \nline3",
		},
		{
			name:     "preserves CRLF newlines",
			input:    "line1\t\r\nline2\t\r\nline3",
			expected: "line1    \r\nline2    \r\nline3",
		},
		{
			name:     "mixed line endings preserved",
			input:    "line1\t\nline2\t\r\nline3",
			expected: "line1    \nline2    \r\nline3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandTabs(tt.input)
			if result != tt.expected {
				t.Errorf("Input:    %q\nGot:      %q\nExpected: %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExpandTabs_PineScriptPatterns(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, result string)
	}{
		{
			name:  "nested if blocks with mixed indentation",
			input: "if close > open\n\trsi_val = ta.rsi(close, 14)\n    if rsi_val > 70\n        x = 1",
			check: func(t *testing.T, result string) {
				if strings.Contains(result, "\t") {
					t.Error("Result contains tabs after expansion")
				}
				lines := strings.Split(result, "\n")
				if len(lines) != 4 {
					t.Errorf("Expected 4 lines, got %d", len(lines))
				}
				if !strings.HasPrefix(lines[1], "    rsi_val") {
					t.Errorf("Line 2 should start with 4 spaces, got: %q", lines[1])
				}
			},
		},
		{
			name:  "function definitions with tab indentation",
			input: "myFunc() =>\n\tresult = ta.sma(close, 10)\n\tresult",
			check: func(t *testing.T, result string) {
				lines := strings.Split(result, "\n")
				for i, line := range lines[1:] {
					if strings.HasPrefix(line, "\t") {
						t.Errorf("Line %d still has tab prefix: %q", i+2, line)
					}
				}
			},
		},
		{
			name:  "for loop with tab indented body",
			input: "for i = 0 to 10\n\tsum := sum + i",
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "    sum") {
					t.Errorf("Tab not expanded in for loop body: %q", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandTabs(tt.input)
			tt.check(t, result)
		})
	}
}

func TestExpandTabs_Properties(t *testing.T) {
	t.Run("idempotent - applying twice produces same result", func(t *testing.T) {
		input := "if a\n\tx = 1\n\t\ty = 2"
		firstPass := ExpandTabs(input)
		secondPass := ExpandTabs(firstPass)

		if firstPass != secondPass {
			t.Errorf("Not idempotent\nFirst:  %q\nSecond: %q", firstPass, secondPass)
		}
	})

	t.Run("character count increases by 3 per tab", func(t *testing.T) {
		input := "\t\t\t"
		result := ExpandTabs(input)

		expectedLen := 12 // 3 tabs × 4 spaces = 12
		if len(result) != expectedLen {
			t.Errorf("Expected %d characters, got %d", expectedLen, len(result))
		}
	})

	t.Run("no tabs means no changes", func(t *testing.T) {
		inputs := []string{
			"no tabs here",
			"    spaces only",
			"x = 1\ny = 2",
			"",
		}

		for _, input := range inputs {
			result := ExpandTabs(input)
			if result != input {
				t.Errorf("Input changed unexpectedly: %q -> %q", input, result)
			}
		}
	})
}
