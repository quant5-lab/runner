package preprocessor

import (
	"strings"
	"testing"
)

func TestRewriteGenericTypeSyntax_ArrayConstructors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "array_new_float",
			input:    "arr = array.new<float>(10)",
			expected: "arr = array.new_float(10)",
		},
		{
			name:     "array_new_int",
			input:    "arr = array.new<int>(5, 0)",
			expected: "arr = array.new_int(5, 0)",
		},
		{
			name:     "array_new_bool",
			input:    "arr = array.new<bool>(3, true)",
			expected: "arr = array.new_bool(3, true)",
		},
		{
			name:     "array_new_color",
			input:    "arr = array.new<color>(2)",
			expected: "arr = array.new_color(2)",
		},
		{
			name:     "array_new_string",
			input:    "arr = array.new<string>(4, \"\")",
			expected: "arr = array.new_string(4, \"\")",
		},
		{
			name:     "multiple_array_constructors",
			input:    "a = array.new<float>(5)\nb = array.new<int>(3)",
			expected: "a = array.new_float(5)\nb = array.new_int(3)",
		},
		{
			name:     "array_constructor_in_expression",
			input:    "result = array.push(array.new<float>(0), close)",
			expected: "result = array.push(array.new_float(0), close)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RewriteGenericTypeSyntax(tt.input)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestRewriteGenericTypeSyntax_MatrixConstructors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "matrix_new_float",
			input:    "m = matrix.new<float>(3, 3)",
			expected: "m = matrix.new_float(3, 3)",
		},
		{
			name:     "matrix_new_int",
			input:    "m = matrix.new<int>(2, 4, 0)",
			expected: "m = matrix.new_int(2, 4, 0)",
		},
		{
			name:     "matrix_new_bool",
			input:    "m = matrix.new<bool>(5, 5)",
			expected: "m = matrix.new_bool(5, 5)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RewriteGenericTypeSyntax(tt.input)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestRewriteGenericTypeSyntax_MapConstructors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "map_string_float_no_spaces",
			input:    "m = map.new<string,float>()",
			expected: "m = map.new_string_float()",
		},
		{
			name:     "map_string_float_with_spaces",
			input:    "m = map.new<string, float>()",
			expected: "m = map.new_string_float()",
		},
		{
			name:     "map_int_bool",
			input:    "m = map.new<int, bool>()",
			expected: "m = map.new_int_bool()",
		},
		{
			name:     "map_float_string",
			input:    "m = map.new<float,string>()",
			expected: "m = map.new_float_string()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RewriteGenericTypeSyntax(tt.input)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestRewriteGenericTypeSyntax_NoRewritePreserved(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "legacy_array_new_float",
			input: "arr = array.new_float(10)",
		},
		{
			name:  "comparison_operators",
			input: "result = a < b and c > d",
		},
		{
			name:  "angle_brackets_in_comments",
			input: "// array.new<float> example",
		},
		{
			name:  "user_defined_type_placeholder",
			input: "arr = array.new<MyType>(5)",
		},
		{
			name:  "array_get_access",
			input: "val = array.get(arr, 0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RewriteGenericTypeSyntax(tt.input)
			if result != tt.input {
				t.Errorf("Input should not be modified.\nExpected:\n%s\nGot:\n%s", tt.input, result)
			}
		})
	}
}

func TestRewriteGenericTypeSyntax_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty_string",
			input:    "",
			expected: "",
		},
		{
			name:     "only_whitespace",
			input:    "   \n\t  ",
			expected: "   \n\t  ",
		},
		{
			name:     "mixed_constructors_single_line",
			input:    "a = array.new<float>(5); b = matrix.new<int>(2, 2)",
			expected: "a = array.new_float(5); b = matrix.new_int(2, 2)",
		},
		{
			name:     "constructor_with_complex_arguments",
			input:    "arr = array.new<float>(bar_index, close[1] + open[0])",
			expected: "arr = array.new_float(bar_index, close[1] + open[0])",
		},
		{
			name:     "nested_function_calls",
			input:    "result = array.push(array.new<float>(0), ta.sma(close, 10))",
			expected: "result = array.push(array.new_float(0), ta.sma(close, 10))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RewriteGenericTypeSyntax(tt.input)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestRewriteGenericTypeSyntax_MultilineScript(t *testing.T) {
	input := `//@version=5
indicator("Generic Type Test", overlay=true)

var prices = array.new<float>(0)
var volumes = array.new<int>(0)
var flags = array.new<bool>(10, false)

m = matrix.new<float>(3, 3, 0.0)
lookup = map.new<string, float>()

array.push(prices, close)
array.push(volumes, int(volume))

plot(array.size(prices), "Count")`

	expected := `//@version=5
indicator("Generic Type Test", overlay=true)

var prices = array.new_float(0)
var volumes = array.new_int(0)
var flags = array.new_bool(10, false)

m = matrix.new_float(3, 3, 0.0)
lookup = map.new_string_float()

array.push(prices, close)
array.push(volumes, int(volume))

plot(array.size(prices), "Count")`

	result := RewriteGenericTypeSyntax(input)
	if result != expected {
		t.Errorf("Multiline script rewrite failed.\nExpected:\n%s\n\nGot:\n%s", expected, result)
	}
}

func TestRewriteGenericTypeSyntax_IdempotenceInvariant(t *testing.T) {
	inputs := []string{
		"arr = array.new<float>(5)",
		"m = matrix.new<int>(2, 2)",
		"lookup = map.new<string, float>()",
		"mixed = array.new<bool>(3) and matrix.new<color>(1, 1)",
	}

	for _, input := range inputs {
		t.Run("idempotence_"+input[:20], func(t *testing.T) {
			firstPass := RewriteGenericTypeSyntax(input)
			secondPass := RewriteGenericTypeSyntax(firstPass)
			if firstPass != secondPass {
				t.Errorf("Rewriter is not idempotent.\nFirst pass:\n%s\nSecond pass:\n%s", firstPass, secondPass)
			}
		})
	}
}

func TestRewriteGenericTypeSyntax_BoundaryWordMatching(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "prefix_match_rejected",
			input:    "myarray.new<float>(5)",
			expected: "myarray.new<float>(5)",
		},
		{
			name:     "suffix_match_rejected",
			input:    "arrayutil.new<float>(5)",
			expected: "arrayutil.new<float>(5)",
		},
		{
			name:     "exact_namespace_match_accepted",
			input:    "array.new<float>(5)",
			expected: "array.new_float(5)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RewriteGenericTypeSyntax(tt.input)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestRewriteGenericTypeSyntax_WhitespaceVariations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "map_no_space_after_comma",
			input:    "m = map.new<string,float>()",
			expected: "m = map.new_string_float()",
		},
		{
			name:     "map_single_space_after_comma",
			input:    "m = map.new<string, float>()",
			expected: "m = map.new_string_float()",
		},
		{
			name:     "map_multiple_spaces_after_comma",
			input:    "m = map.new<string,  float>()",
			expected: "m = map.new_string_float()",
		},
		{
			name:     "map_tab_after_comma",
			input:    "m = map.new<string,\tfloat>()",
			expected: "m = map.new_string_float()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RewriteGenericTypeSyntax(tt.input)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestRewriteGenericTypeSyntax_CaseSensitivity(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase_array",
			input:    "arr = array.new<float>(5)",
			expected: "arr = array.new_float(5)",
		},
		{
			name:     "uppercase_array_no_match",
			input:    "arr = ARRAY.new<float>(5)",
			expected: "arr = ARRAY.new<float>(5)",
		},
		{
			name:     "mixed_case_array_no_match",
			input:    "arr = Array.new<float>(5)",
			expected: "arr = Array.new<float>(5)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RewriteGenericTypeSyntax(tt.input)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestRewriteGenericTypeSyntax_PreservesComments(t *testing.T) {
	input := `// array.new<float> is new syntax
arr = array.new<float>(10)  // Inline comment
/* Block comment with array.new<int> */
x = array.new<int>(5)`

	result := RewriteGenericTypeSyntax(input)

	if !strings.Contains(result, "// array.new<float> is new syntax") {
		t.Error("Line comment was incorrectly modified")
	}
	if !strings.Contains(result, "// Inline comment") {
		t.Error("Inline comment was incorrectly modified")
	}
	if !strings.Contains(result, "/* Block comment with array.new<int> */") {
		t.Error("Block comment was incorrectly modified")
	}
	if !strings.Contains(result, "arr = array.new_float(10)") {
		t.Error("Code was not rewritten")
	}
	if !strings.Contains(result, "x = array.new_int(5)") {
		t.Error("Code was not rewritten")
	}
}
