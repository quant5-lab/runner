package codegen

import (
	"strings"
	"testing"
)

func TestArrowVarInitResult_Construction(t *testing.T) {
	tests := []struct {
		name       string
		preamble   string
		assignment string
	}{
		{
			name:       "both preamble and assignment",
			preamble:   "temp := expr\n",
			assignment: "\tvar := func() { return temp }()\n",
		},
		{
			name:       "assignment only",
			preamble:   "",
			assignment: "\tvar := value\n",
		},
		{
			name:       "empty result",
			preamble:   "",
			assignment: "",
		},
		{
			name:       "multiline preamble",
			preamble:   "temp1 := a\ntemp2 := b\n",
			assignment: "\tvar := temp1 + temp2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewArrowVarInitResult(tt.preamble, tt.assignment)

			if result.Preamble != tt.preamble {
				t.Errorf("Expected preamble %q, got %q", tt.preamble, result.Preamble)
			}

			if result.Assignment != tt.assignment {
				t.Errorf("Expected assignment %q, got %q", tt.assignment, result.Assignment)
			}
		})
	}
}

func TestArrowVarInitResult_HasPreamble(t *testing.T) {
	tests := []struct {
		name     string
		preamble string
		expected bool
	}{
		{
			name:     "with preamble",
			preamble: "temp := value\n",
			expected: true,
		},
		{
			name:     "empty preamble",
			preamble: "",
			expected: false,
		},
		{
			name:     "whitespace only",
			preamble: "   ",
			expected: true,
		},
		{
			name:     "newline only",
			preamble: "\n",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewArrowVarInitResult(tt.preamble, "assignment")

			if result.HasPreamble() != tt.expected {
				t.Errorf("Expected HasPreamble() = %v, got %v", tt.expected, result.HasPreamble())
			}
		})
	}
}

func TestArrowVarInitResult_CombinedCode(t *testing.T) {
	tests := []struct {
		name       string
		preamble   string
		assignment string
		expected   string
	}{
		{
			name:       "concatenates preamble and assignment",
			preamble:   "temp := a + b\n",
			assignment: "\tvar := temp * 2\n",
			expected:   "temp := a + b\n\tvar := temp * 2\n",
		},
		{
			name:       "assignment only",
			preamble:   "",
			assignment: "\tvar := value\n",
			expected:   "\tvar := value\n",
		},
		{
			name:       "preamble only",
			preamble:   "temp := value\n",
			assignment: "",
			expected:   "temp := value\n",
		},
		{
			name:       "both empty",
			preamble:   "",
			assignment: "",
			expected:   "",
		},
		{
			name:       "multiple preamble statements",
			preamble:   "temp1 := a\ntemp2 := b\ntemp3 := c\n",
			assignment: "\tvar := combine(temp1, temp2, temp3)\n",
			expected:   "temp1 := a\ntemp2 := b\ntemp3 := c\n\tvar := combine(temp1, temp2, temp3)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewArrowVarInitResult(tt.preamble, tt.assignment)
			combined := result.CombinedCode()

			if combined != tt.expected {
				t.Errorf("Expected combined code:\n%s\nGot:\n%s", tt.expected, combined)
			}
		})
	}
}

func TestArrowVarInitResult_EdgeCases(t *testing.T) {
	t.Run("nil safety", func(t *testing.T) {
		result := NewArrowVarInitResult("", "")
		if result == nil {
			t.Fatal("NewArrowVarInitResult returned nil")
		}
	})

	t.Run("large preamble", func(t *testing.T) {
		largePreamble := strings.Repeat("temp := value\n", 1000)
		result := NewArrowVarInitResult(largePreamble, "var := final\n")

		if !result.HasPreamble() {
			t.Error("Expected HasPreamble() to be true for large preamble")
		}

		combined := result.CombinedCode()
		if !strings.Contains(combined, largePreamble) {
			t.Error("Combined code does not contain full preamble")
		}
	})

	t.Run("unicode in preamble", func(t *testing.T) {
		preamble := "temp := 日本語\n"
		assignment := "\tvar := temp\n"
		result := NewArrowVarInitResult(preamble, assignment)

		if result.Preamble != preamble {
			t.Errorf("Unicode not preserved in preamble")
		}
	})

	t.Run("special characters", func(t *testing.T) {
		preamble := "temp := \"\\n\\t\\r\"\n"
		assignment := "\tvar := temp\n"
		result := NewArrowVarInitResult(preamble, assignment)

		combined := result.CombinedCode()
		if !strings.Contains(combined, "\\n\\t\\r") {
			t.Error("Special characters not preserved")
		}
	})
}

func TestArrowVarInitResult_RealWorldPatterns(t *testing.T) {
	tests := []struct {
		name        string
		preamble    string
		assignment  string
		description string
	}{
		{
			name:        "fixnan with complex expression",
			preamble:    "fixnan_source_temp := (100 * rma(upSeries, 20) / truerange)\n",
			assignment:  "\tplus := func() float64 { val := fixnan_source_temp; if math.IsNaN(val) { return 0.0 }; return val }()\n",
			description: "Complex arithmetic expression with TA function",
		},
		{
			name:        "nested TA calls",
			preamble:    "ternary_source_temp := func() float64 { if condition { return a } else { return b } }()\n",
			assignment:  "\tresult := func() float64 { return rma(ternary_source_temp, period) }()\n",
			description: "Conditional expression feeding into TA function",
		},
		{
			name:        "binary expression with series",
			preamble:    "binary_source_temp := (plusSeries.GetCurrent() - minusSeries.GetCurrent())\n",
			assignment:  "\tdelta := func() float64 { return binary_source_temp * multiplier }()\n",
			description: "Series arithmetic with post-processing",
		},
		{
			name:        "simple identifier assignment",
			preamble:    "",
			assignment:  "\tvar := identifier\n",
			description: "Direct identifier reference without preamble",
		},
		{
			name:        "literal value",
			preamble:    "",
			assignment:  "\tconst := 42.0\n",
			description: "Literal value assignment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewArrowVarInitResult(tt.preamble, tt.assignment)

			if tt.preamble != "" && !result.HasPreamble() {
				t.Errorf("[%s] Expected preamble to be detected", tt.description)
			}

			combined := result.CombinedCode()
			if tt.preamble != "" && !strings.Contains(combined, tt.preamble) {
				t.Errorf("[%s] Preamble not found in combined code", tt.description)
			}

			if !strings.Contains(combined, tt.assignment) {
				t.Errorf("[%s] Assignment not found in combined code", tt.description)
			}

			if tt.preamble != "" {
				preambleIdx := strings.Index(combined, tt.preamble)
				assignmentIdx := strings.Index(combined, tt.assignment)
				if preambleIdx > assignmentIdx {
					t.Errorf("[%s] Preamble must appear before assignment in combined code", tt.description)
				}
			}
		})
	}
}
