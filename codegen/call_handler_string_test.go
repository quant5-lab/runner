package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestStringNamespaceHandler_CanHandle validates string function recognition
 *
 * Tests that StringNamespaceHandler correctly identifies str.* namespace functions
 * and defers non-string functions to subsequent handlers in the chain.
 */
func TestStringNamespaceHandler_CanHandle(t *testing.T) {
	handler := NewStringNamespaceHandler()

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		// Registered str.* functions
		{"str.tostring", "str.tostring", true},
		{"str.tonumber", "str.tonumber", true},
		{"str.length", "str.length", true},
		{"str.format", "str.format", true},
		{"str.contains", "str.contains", true},
		{"str.pos", "str.pos", true},
		{"str.substring", "str.substring", true},
		{"str.lower", "str.lower", true},
		{"str.upper", "str.upper", true},
		{"str.trim", "str.trim", true},
		{"str.replace", "str.replace", true},
		{"str.replace_all", "str.replace_all", true},
		{"str.split", "str.split", true},
		{"str.startswith", "str.startswith", true},
		{"str.endswith", "str.endswith", true},
		{"str.repeat", "str.repeat", true},
		{"str.match", "str.match", true},
		{"str.format_time", "str.format_time", true},

		// Non-string functions
		{"ta.sma", "ta.sma", false},
		{"math.abs", "math.abs", false},
		{"plot", "plot", false},
		{"strategy.entry", "strategy.entry", false},
		{"array.new_int", "array.new_int", false},
		{"map.new", "map.new", false},
		{"request.security", "request.security", false},
		{"ticker.heikinashi", "ticker.heikinashi", false},
		{"unknown_func", "unknown", false},
		{"empty", "", false},

		// Case sensitivity validation
		{"STR.UPPER", "STR.UPPER", false},
		{"Str.Lower", "Str.Lower", false},
		{"str.CONTAINS", "str.CONTAINS", false},

		// Partial namespace matches should fail
		{"string", "string", false},
		{"str", "str", false},
		{"str.", "str.", false},
		{"tostring", "tostring", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.CanHandle(tt.funcName)
			if got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

/* TestStringNamespaceHandler_GenerateCode_Conversions tests str.tostring and str.tonumber
 *
 * Validates type conversion functions with various argument types and format specifiers.
 */
func TestStringNamespaceHandler_GenerateCode_Conversions(t *testing.T) {
	tests := []struct {
		name           string
		funcName       string
		args           []ast.Expression
		expectError    bool
		validateOutput func(t *testing.T, code string)
	}{
		{
			name:     "str.tostring with identifier default format",
			funcName: "str.tostring",
			args: []ast.Expression{
				&ast.Identifier{Name: "close"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "fmt.Sprintf") {
					t.Errorf("Expected fmt.Sprintf for str.tostring")
				}
				if !strings.Contains(code, "%.10g") {
					t.Errorf("Expected default format")
				}
				if !strings.Contains(code, "close") {
					t.Errorf("Expected 'close' argument")
				}
			},
		},
		{
			name:     "str.tostring with custom format",
			funcName: "str.tostring",
			args: []ast.Expression{
				&ast.Identifier{Name: "price"},
				&ast.Literal{Value: "%.2f"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "fmt.Sprintf") {
					t.Errorf("Expected fmt.Sprintf")
				}
				if !strings.Contains(code, "%.2f") {
					t.Errorf("Expected custom format")
				}
				if !strings.Contains(code, "price") {
					t.Errorf("Expected 'price' argument")
				}
			},
		},
		{
			name:     "str.tostring with literal number",
			funcName: "str.tostring",
			args: []ast.Expression{
				&ast.Literal{Value: 42.5},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "fmt.Sprintf") {
					t.Error("Expected fmt.Sprintf")
				}
				if !strings.Contains(code, "42.5") {
					t.Error("Expected literal value 42.5")
				}
			},
		},
		{
			name:     "str.tostring with call expression result",
			funcName: "str.tostring",
			args: []ast.Expression{
				&ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "math"},
						Property: &ast.Identifier{Name: "abs"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "value"},
					},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "fmt.Sprintf") {
					t.Error("Expected fmt.Sprintf")
				}
			},
		},
		{
			name:     "str.tonumber with string literal",
			funcName: "str.tonumber",
			args: []ast.Expression{
				&ast.Literal{Value: "123.45"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "strconv.ParseFloat") {
					t.Error("Expected strconv.ParseFloat for str.tonumber")
				}
				if !strings.Contains(code, "123.45") {
					t.Error("Expected string literal 123.45")
				}
				if !strings.Contains(code, ", 64)") {
					t.Error("Expected 64-bit float parsing")
				}
			},
		},
		{
			name:     "str.tonumber with identifier",
			funcName: "str.tonumber",
			args: []ast.Expression{
				&ast.Identifier{Name: "strValue"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "strconv.ParseFloat") {
					t.Error("Expected strconv.ParseFloat")
				}
				if !strings.Contains(code, "strValue") {
					t.Error("Expected 'strValue' identifier")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			handler := NewStringNamespaceHandler()

			call := &ast.CallExpression{
				Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "str"}, Property: &ast.Identifier{Name: strings.TrimPrefix(tt.funcName, "str.")}},
				Arguments: tt.args,
			}

			code, err := handler.GenerateCode(g, call)
			if (err != nil) != tt.expectError {
				t.Fatalf("GenerateCode() error = %v, expectError = %v", err, tt.expectError)
			}
			if !tt.expectError && tt.validateOutput != nil {
				tt.validateOutput(t, code)
			}
		})
	}
}

/* TestStringNamespaceHandler_GenerateCode_SimpleOperations tests single-arg functions
 *
 * Validates str.lower, str.upper, str.trim, str.length behavior with various inputs.
 */
func TestStringNamespaceHandler_GenerateCode_SimpleOperations(t *testing.T) {
	tests := []struct {
		name           string
		funcName       string
		args           []ast.Expression
		expectError    bool
		validateOutput func(t *testing.T, code string)
	}{
		{
			name:     "str.lower with literal",
			funcName: "str.lower",
			args: []ast.Expression{
				&ast.Literal{Value: "HELLO"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "strings.ToLower") {
					t.Error("Expected strings.ToLower")
				}
				if !strings.Contains(code, "HELLO") {
					t.Error("Expected 'HELLO' argument")
				}
			},
		},
		{
			name:     "str.upper with identifier",
			funcName: "str.upper",
			args: []ast.Expression{
				&ast.Identifier{Name: "text"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "strings.ToUpper") {
					t.Error("Expected strings.ToUpper")
				}
				if !strings.Contains(code, "text") {
					t.Error("Expected 'text' identifier")
				}
			},
		},
		{
			name:     "str.trim with identifier",
			funcName: "str.trim",
			args: []ast.Expression{
				&ast.Identifier{Name: "input"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "strings.TrimSpace") {
					t.Error("Expected strings.TrimSpace")
				}
				if !strings.Contains(code, "input") {
					t.Error("Expected 'input' identifier")
				}
			},
		},
		{
			name:     "str.length with literal",
			funcName: "str.length",
			args: []ast.Expression{
				&ast.Literal{Value: "test"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "len(") {
					t.Error("Expected len() function")
				}
				if !strings.Contains(code, "test") {
					t.Error("Expected 'test' literal")
				}
			},
		},
		{
			name:     "str.length with call expression",
			funcName: "str.length",
			args: []ast.Expression{
				&ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "str"},
						Property: &ast.Identifier{Name: "upper"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "name"},
					},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "len(") {
					t.Error("Expected len() function")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			handler := NewStringNamespaceHandler()

			call := &ast.CallExpression{
				Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "str"}, Property: &ast.Identifier{Name: strings.TrimPrefix(tt.funcName, "str.")}},
				Arguments: tt.args,
			}

			code, err := handler.GenerateCode(g, call)
			if (err != nil) != tt.expectError {
				t.Fatalf("GenerateCode() error = %v, expectError = %v", err, tt.expectError)
			}
			if !tt.expectError && tt.validateOutput != nil {
				tt.validateOutput(t, code)
			}
		})
	}
}

/* TestStringNamespaceHandler_GenerateCode_TwoArgFunctions tests binary string operations
 *
 * Validates str.contains, str.startswith, str.endswith, str.pos, str.split, str.match.
 */
func TestStringNamespaceHandler_GenerateCode_TwoArgFunctions(t *testing.T) {
	tests := []struct {
		name           string
		funcName       string
		args           []ast.Expression
		expectError    bool
		validateOutput func(t *testing.T, code string)
	}{
		{
			name:     "str.contains with literals",
			funcName: "str.contains",
			args: []ast.Expression{
				&ast.Literal{Value: "hello world"},
				&ast.Literal{Value: "world"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "strings.Contains") {
					t.Error("Expected strings.Contains")
				}
			},
		},
		{
			name:     "str.startswith with identifiers",
			funcName: "str.startswith",
			args: []ast.Expression{
				&ast.Identifier{Name: "text"},
				&ast.Literal{Value: "prefix"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "strings.HasPrefix") {
					t.Error("Expected strings.HasPrefix")
				}
				if !strings.Contains(code, "text") {
					t.Error("Expected 'text' identifier")
				}
				if !strings.Contains(code, "prefix") {
					t.Error("Expected 'prefix' literal")
				}
			},
		},
		{
			name:     "str.endswith with identifiers",
			funcName: "str.endswith",
			args: []ast.Expression{
				&ast.Identifier{Name: "filename"},
				&ast.Literal{Value: ".txt"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "strings.HasSuffix") {
					t.Error("Expected strings.HasSuffix")
				}
			},
		},
		{
			name:     "str.pos with identifiers",
			funcName: "str.pos",
			args: []ast.Expression{
				&ast.Identifier{Name: "haystack"},
				&ast.Identifier{Name: "needle"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "strings.Index") {
					t.Error("Expected strings.Index")
				}
			},
		},
		{
			name:     "str.split with delimiter",
			funcName: "str.split",
			args: []ast.Expression{
				&ast.Literal{Value: "a,b,c"},
				&ast.Literal{Value: ","},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "strings.Split") {
					t.Error("Expected strings.Split")
				}
			},
		},
		{
			name:     "str.match with regex pattern",
			funcName: "str.match",
			args: []ast.Expression{
				&ast.Identifier{Name: "input"},
				&ast.Literal{Value: "[0-9]+"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "regexp.MustCompile") {
					t.Error("Expected regexp.MustCompile")
				}
				if !strings.Contains(code, "FindString") {
					t.Error("Expected FindString method")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			handler := NewStringNamespaceHandler()

			call := &ast.CallExpression{
				Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "str"}, Property: &ast.Identifier{Name: strings.TrimPrefix(tt.funcName, "str.")}},
				Arguments: tt.args,
			}

			code, err := handler.GenerateCode(g, call)
			if (err != nil) != tt.expectError {
				t.Fatalf("GenerateCode() error = %v, expectError = %v", err, tt.expectError)
			}
			if !tt.expectError && tt.validateOutput != nil {
				tt.validateOutput(t, code)
			}
		})
	}
}

/* TestStringNamespaceHandler_GenerateCode_SubstringOperations tests substring extraction
 *
 * Validates str.substring with 2-arg (from index to end) and 3-arg (from..to) forms.
 */
func TestStringNamespaceHandler_GenerateCode_SubstringOperations(t *testing.T) {
	tests := []struct {
		name           string
		args           []ast.Expression
		expectError    bool
		validateOutput func(t *testing.T, code string)
	}{
		{
			name: "str.substring with start index only",
			args: []ast.Expression{
				&ast.Identifier{Name: "text"},
				&ast.Literal{Value: 5.0},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "text[5:]") {
					t.Error("Expected slice notation text[5:]")
				}
			},
		},
		{
			name: "str.substring with start and end index",
			args: []ast.Expression{
				&ast.Identifier{Name: "text"},
				&ast.Literal{Value: 2.0},
				&ast.Literal{Value: 8.0},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "text[2:8]") {
					t.Error("Expected slice notation text[2:8]")
				}
			},
		},
		{
			name: "str.substring with zero start index",
			args: []ast.Expression{
				&ast.Literal{Value: "hello"},
				&ast.Literal{Value: 0.0},
				&ast.Literal{Value: 3.0},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "[0:3]") {
					t.Error("Expected slice notation [0:3]")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			handler := NewStringNamespaceHandler()

			call := &ast.CallExpression{
				Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "str"}, Property: &ast.Identifier{Name: "substring"}},
				Arguments: tt.args,
			}

			code, err := handler.GenerateCode(g, call)
			if (err != nil) != tt.expectError {
				t.Fatalf("GenerateCode() error = %v, expectError = %v", err, tt.expectError)
			}
			if !tt.expectError && tt.validateOutput != nil {
				tt.validateOutput(t, code)
			}
		})
	}
}

/* TestStringNamespaceHandler_GenerateCode_ReplaceOperations tests str.replace and str.replace_all
 *
 * Validates replacement with optional occurrence count for str.replace.
 */
func TestStringNamespaceHandler_GenerateCode_ReplaceOperations(t *testing.T) {
	tests := []struct {
		name           string
		funcName       string
		args           []ast.Expression
		expectError    bool
		validateOutput func(t *testing.T, code string)
	}{
		{
			name:     "str.replace with default count -1",
			funcName: "str.replace",
			args: []ast.Expression{
				&ast.Identifier{Name: "text"},
				&ast.Literal{Value: "old"},
				&ast.Literal{Value: "new"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "strings.Replace") {
					t.Error("Expected strings.Replace")
				}
				if !strings.Contains(code, ", -1)") {
					t.Error("Expected default count -1 for replace all occurrences")
				}
			},
		},
		{
			name:     "str.replace with explicit count",
			funcName: "str.replace",
			args: []ast.Expression{
				&ast.Identifier{Name: "text"},
				&ast.Literal{Value: "old"},
				&ast.Literal{Value: "new"},
				&ast.Literal{Value: 1.0},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "strings.Replace") {
					t.Error("Expected strings.Replace")
				}
				if !strings.Contains(code, ", 1)") {
					t.Error("Expected explicit count 1")
				}
			},
		},
		{
			name:     "str.replace_all always replaces all",
			funcName: "str.replace_all",
			args: []ast.Expression{
				&ast.Identifier{Name: "text"},
				&ast.Literal{Value: "old"},
				&ast.Literal{Value: "new"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "strings.ReplaceAll") {
					t.Error("Expected strings.ReplaceAll")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			handler := NewStringNamespaceHandler()

			call := &ast.CallExpression{
				Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "str"}, Property: &ast.Identifier{Name: strings.TrimPrefix(tt.funcName, "str.")}},
				Arguments: tt.args,
			}

			code, err := handler.GenerateCode(g, call)
			if (err != nil) != tt.expectError {
				t.Fatalf("GenerateCode() error = %v, expectError = %v", err, tt.expectError)
			}
			if !tt.expectError && tt.validateOutput != nil {
				tt.validateOutput(t, code)
			}
		})
	}
}

/* TestStringNamespaceHandler_GenerateCode_RepeatOperation tests str.repeat
 *
 * Validates string repetition with optional separator.
 */
func TestStringNamespaceHandler_GenerateCode_RepeatOperation(t *testing.T) {
	tests := []struct {
		name           string
		args           []ast.Expression
		expectError    bool
		validateOutput func(t *testing.T, code string)
	}{
		{
			name: "str.repeat without separator",
			args: []ast.Expression{
				&ast.Literal{Value: "abc"},
				&ast.Literal{Value: 3.0},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "strings.Repeat") {
					t.Error("Expected strings.Repeat")
				}
				if !strings.Contains(code, `+ ""`) {
					t.Error("Expected empty separator concatenation")
				}
			},
		},
		{
			name: "str.repeat with separator",
			args: []ast.Expression{
				&ast.Literal{Value: "abc"},
				&ast.Literal{Value: 3.0},
				&ast.Literal{Value: "-"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "strings.Repeat") {
					t.Error("Expected strings.Repeat")
				}
				if !strings.Contains(code, `+ "-"`) {
					t.Error("Expected separator '-' concatenation")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			handler := NewStringNamespaceHandler()

			call := &ast.CallExpression{
				Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "str"}, Property: &ast.Identifier{Name: "repeat"}},
				Arguments: tt.args,
			}

			code, err := handler.GenerateCode(g, call)
			if (err != nil) != tt.expectError {
				t.Fatalf("GenerateCode() error = %v, expectError = %v", err, tt.expectError)
			}
			if !tt.expectError && tt.validateOutput != nil {
				tt.validateOutput(t, code)
			}
		})
	}
}

/* TestStringNamespaceHandler_GenerateCode_FormatOperations tests str.format and str.format_time
 *
 * Validates printf-style formatting and timestamp formatting.
 */
func TestStringNamespaceHandler_GenerateCode_FormatOperations(t *testing.T) {
	tests := []struct {
		name           string
		funcName       string
		args           []ast.Expression
		expectError    bool
		validateOutput func(t *testing.T, code string)
	}{
		{
			name:     "str.format with template only",
			funcName: "str.format",
			args: []ast.Expression{
				&ast.Literal{Value: "Hello"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if code != `"Hello"` {
					t.Errorf("Expected literal string, got %q", code)
				}
			},
		},
		{
			name:     "str.format with single argument",
			funcName: "str.format",
			args: []ast.Expression{
				&ast.Literal{Value: "Value: {0}"},
				&ast.Identifier{Name: "price"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "fmt.Sprintf") {
					t.Error("Expected fmt.Sprintf")
				}
				if !strings.Contains(code, "price") {
					t.Error("Expected 'price' argument")
				}
			},
		},
		{
			name:     "str.format with multiple arguments",
			funcName: "str.format",
			args: []ast.Expression{
				&ast.Literal{Value: "{0} + {1} = {2}"},
				&ast.Literal{Value: 1.0},
				&ast.Literal{Value: 2.0},
				&ast.Literal{Value: 3.0},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "fmt.Sprintf") {
					t.Error("Expected fmt.Sprintf")
				}
			},
		},
		{
			name:     "str.format_time with default UTC timezone",
			funcName: "str.format_time",
			args: []ast.Expression{
				&ast.Identifier{Name: "timestamp"},
				&ast.Literal{Value: "2006-01-02"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "time.Unix") {
					t.Error("Expected time.Unix conversion")
				}
				if !strings.Contains(code, "/1000") {
					t.Error("Expected millisecond to second conversion")
				}
				if !strings.Contains(code, `"UTC"`) {
					t.Error("Expected default UTC timezone")
				}
				if !strings.Contains(code, ".Format(") {
					t.Error("Expected .Format() call")
				}
			},
		},
		{
			name:     "str.format_time with custom timezone",
			funcName: "str.format_time",
			args: []ast.Expression{
				&ast.Identifier{Name: "timestamp"},
				&ast.Literal{Value: "15:04:05"},
				&ast.Literal{Value: "America/New_York"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "time.Unix") {
					t.Error("Expected time.Unix conversion")
				}
				if !strings.Contains(code, "America/New_York") {
					t.Error("Expected custom timezone")
				}
				if !strings.Contains(code, "time.LoadLocation") {
					t.Error("Expected time.LoadLocation for timezone")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			handler := NewStringNamespaceHandler()

			call := &ast.CallExpression{
				Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "str"}, Property: &ast.Identifier{Name: strings.TrimPrefix(tt.funcName, "str.")}},
				Arguments: tt.args,
			}

			code, err := handler.GenerateCode(g, call)
			if (err != nil) != tt.expectError {
				t.Fatalf("GenerateCode() error = %v, expectError = %v", err, tt.expectError)
			}
			if !tt.expectError && tt.validateOutput != nil {
				tt.validateOutput(t, code)
			}
		})
	}
}

/* TestStringNamespaceHandler_ArgumentCountValidation tests signature validation
 *
 * Validates that functions enforce min/max argument counts correctly.
 */
func TestStringNamespaceHandler_ArgumentCountValidation(t *testing.T) {
	tests := []struct {
		name        string
		funcName    string
		argCount    int
		expectError bool
		errorPhrase string
	}{
		// Single argument functions
		{"str.lower with 0 args", "str.lower", 0, true, "expected 1 arguments"},
		{"str.lower with 1 arg", "str.lower", 1, false, ""},
		{"str.lower with 2 args", "str.lower", 2, true, "expected 1 arguments"},

		{"str.upper with 0 args", "str.upper", 0, true, "expected 1 arguments"},
		{"str.upper with 1 arg", "str.upper", 1, false, ""},

		{"str.trim with 0 args", "str.trim", 0, true, "expected 1 arguments"},
		{"str.trim with 1 arg", "str.trim", 1, false, ""},

		{"str.length with 0 args", "str.length", 0, true, "expected 1 arguments"},
		{"str.length with 1 arg", "str.length", 1, false, ""},

		{"str.tonumber with 0 args", "str.tonumber", 0, true, "expected 1 arguments"},
		{"str.tonumber with 1 arg", "str.tonumber", 1, false, ""},

		// Two argument functions
		{"str.contains with 1 arg", "str.contains", 1, true, "expected 2 arguments"},
		{"str.contains with 2 args", "str.contains", 2, false, ""},
		{"str.contains with 3 args", "str.contains", 3, true, "expected 2 arguments"},

		{"str.startswith with 1 arg", "str.startswith", 1, true, "expected 2 arguments"},
		{"str.startswith with 2 args", "str.startswith", 2, false, ""},

		{"str.endswith with 1 arg", "str.endswith", 1, true, "expected 2 arguments"},
		{"str.endswith with 2 args", "str.endswith", 2, false, ""},

		{"str.pos with 1 arg", "str.pos", 1, true, "expected 2 arguments"},
		{"str.pos with 2 args", "str.pos", 2, false, ""},

		{"str.split with 1 arg", "str.split", 1, true, "expected 2 arguments"},
		{"str.split with 2 args", "str.split", 2, false, ""},

		{"str.match with 1 arg", "str.match", 1, true, "expected 2 arguments"},
		{"str.match with 2 args", "str.match", 2, false, ""},

		// Variable argument functions
		{"str.tostring with 0 args", "str.tostring", 0, true, "expected 1-2 arguments"},
		{"str.tostring with 1 arg", "str.tostring", 1, false, ""},
		{"str.tostring with 2 args", "str.tostring", 2, false, ""},
		{"str.tostring with 3 args", "str.tostring", 3, true, "expected 1-2 arguments"},

		{"str.substring with 1 arg", "str.substring", 1, true, "expected 2-3 arguments"},
		{"str.substring with 2 args", "str.substring", 2, false, ""},
		{"str.substring with 3 args", "str.substring", 3, false, ""},
		{"str.substring with 4 args", "str.substring", 4, true, "expected 2-3 arguments"},

		{"str.replace with 2 args", "str.replace", 2, true, "expected 3-4 arguments"},
		{"str.replace with 3 args", "str.replace", 3, false, ""},
		{"str.replace with 4 args", "str.replace", 4, false, ""},
		{"str.replace with 5 args", "str.replace", 5, true, "expected 3-4 arguments"},

		{"str.replace_all with 2 args", "str.replace_all", 2, true, "expected 3 arguments"},
		{"str.replace_all with 3 args", "str.replace_all", 3, false, ""},
		{"str.replace_all with 4 args", "str.replace_all", 4, true, "expected 3 arguments"},

		{"str.repeat with 1 arg", "str.repeat", 1, true, "expected 2-3 arguments"},
		{"str.repeat with 2 args", "str.repeat", 2, false, ""},
		{"str.repeat with 3 args", "str.repeat", 3, false, ""},
		{"str.repeat with 4 args", "str.repeat", 4, true, "expected 2-3 arguments"},

		{"str.format_time with 1 arg", "str.format_time", 1, true, "expected 2-3 arguments"},
		{"str.format_time with 2 args", "str.format_time", 2, false, ""},
		{"str.format_time with 3 args", "str.format_time", 3, false, ""},
		{"str.format_time with 4 args", "str.format_time", 4, true, "expected 2-3 arguments"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			handler := NewStringNamespaceHandler()

			args := make([]ast.Expression, tt.argCount)
			for i := 0; i < tt.argCount; i++ {
				args[i] = &ast.Literal{Value: "arg"}
			}

			call := &ast.CallExpression{
				Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "str"}, Property: &ast.Identifier{Name: strings.TrimPrefix(tt.funcName, "str.")}},
				Arguments: args,
			}

			_, err := handler.GenerateCode(g, call)
			if (err != nil) != tt.expectError {
				t.Fatalf("GenerateCode() error = %v, expectError = %v", err, tt.expectError)
			}
			if tt.expectError && err != nil && !strings.Contains(err.Error(), tt.errorPhrase) {
				t.Errorf("Expected error containing %q, got %q", tt.errorPhrase, err.Error())
			}
		})
	}
}

/* TestStringNamespaceHandler_NestedExpressionArguments tests complex nested arguments
 *
 * Validates that nested call expressions, conditionals, and binary expressions work as arguments.
 */
func TestStringNamespaceHandler_NestedExpressionArguments(t *testing.T) {
	tests := []struct {
		name           string
		funcName       string
		args           []ast.Expression
		expectError    bool
		validateOutput func(t *testing.T, code string)
	}{
		{
			name:     "str.lower with nested str.upper call",
			funcName: "str.lower",
			args: []ast.Expression{
				&ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "str"},
						Property: &ast.Identifier{Name: "upper"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "text"},
					},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "strings.ToLower") {
					t.Error("Expected outer strings.ToLower")
				}
				if !strings.Contains(code, "strings.ToUpper") {
					t.Error("Expected nested strings.ToUpper")
				}
			},
		},
		{
			name:     "str.contains with conditional argument",
			funcName: "str.contains",
			args: []ast.Expression{
				&ast.ConditionalExpression{
					Test: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "x"},
						Operator: ">",
						Right:    &ast.Literal{Value: 0.0},
					},
					Consequent: &ast.Literal{Value: "positive"},
					Alternate:  &ast.Literal{Value: "negative"},
				},
				&ast.Literal{Value: "pos"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "strings.Contains") {
					t.Error("Expected strings.Contains")
				}
			},
		},
		{
			name:     "str.substring with identifier indices",
			funcName: "str.substring",
			args: []ast.Expression{
				&ast.Identifier{Name: "text"},
				&ast.Identifier{Name: "startIdx"},
				&ast.Identifier{Name: "endIdx"},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "[") || !strings.Contains(code, "]") {
					t.Error("Expected slice notation")
				}
				if !strings.Contains(code, "startIdx") || !strings.Contains(code, "endIdx") {
					t.Error("Expected identifier indices")
				}
			},
		},
		{
			name:     "str.format with nested str.tostring",
			funcName: "str.format",
			args: []ast.Expression{
				&ast.Literal{Value: "Price: {0}"},
				&ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "str"},
						Property: &ast.Identifier{Name: "tostring"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: "%.2f"},
					},
				},
			},
			expectError: false,
			validateOutput: func(t *testing.T, code string) {
				if !strings.Contains(code, "fmt.Sprintf") {
					t.Error("Expected fmt.Sprintf")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			handler := NewStringNamespaceHandler()

			call := &ast.CallExpression{
				Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "str"}, Property: &ast.Identifier{Name: strings.TrimPrefix(tt.funcName, "str.")}},
				Arguments: tt.args,
			}

			code, err := handler.GenerateCode(g, call)
			if (err != nil) != tt.expectError {
				t.Fatalf("GenerateCode() error = %v, expectError = %v", err, tt.expectError)
			}
			if !tt.expectError && tt.validateOutput != nil {
				tt.validateOutput(t, code)
			}
		})
	}
}

/* TestStringNamespaceHandler_RouterIntegration tests handler integration with CallExpressionRouter
 *
 * Validates that StringNamespaceHandler is correctly positioned in the handler chain
 * and doesn't interfere with other handlers.
 */
func TestStringNamespaceHandler_RouterIntegration(t *testing.T) {
	tests := []struct {
		name        string
		funcName    string
		shouldRoute bool
		description string
	}{
		{"str.lower routed to StringNamespaceHandler", "str.lower", true, "String function should be handled"},
		{"str.format routed to StringNamespaceHandler", "str.format", true, "String function should be handled"},
		{"ta.sma not routed to StringNamespaceHandler", "ta.sma", false, "TA function should skip to TAIndicatorCallHandler"},
		{"math.abs not routed to StringNamespaceHandler", "math.abs", false, "Math function should skip to MathCallHandler"},
		{"ticker.heikinashi not routed to StringNamespaceHandler", "ticker.heikinashi", false, "Ticker function should skip to TickerFunctionHandler"},
		{"unknown_func falls through", "unknown_func", false, "Unknown function should reach UnknownFunctionHandler"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewCallExpressionRouter()
			handler := NewStringNamespaceHandler()

			canHandle := handler.CanHandle(tt.funcName)
			if canHandle != tt.shouldRoute {
				t.Errorf("Handler routing mismatch for %q: canHandle=%v, want=%v", tt.funcName, canHandle, tt.shouldRoute)
			}

			foundIdx := -1
			for i, h := range router.handlers {
				if h.CanHandle(tt.funcName) {
					foundIdx = i
					break
				}
			}

			if tt.shouldRoute && foundIdx == -1 {
				t.Errorf("Function %q should be routed but no handler found", tt.funcName)
			}
			if !tt.shouldRoute && foundIdx >= 0 {
				if _, ok := router.handlers[foundIdx].(*StringNamespaceHandler); ok {
					t.Errorf("Function %q should NOT be routed to StringNamespaceHandler but was", tt.funcName)
				}
			}
		})
	}
}

/* TestStringNamespaceHandler_NonStringFunctionsIgnored validates handler scope
 *
 * Ensures StringNamespaceHandler returns empty string for non-str.* functions,
 * allowing them to pass through to subsequent handlers.
 */
func TestStringNamespaceHandler_NonStringFunctionsIgnored(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
	}{
		{"ta.sma ignored", "ta.sma"},
		{"math.abs ignored", "math.abs"},
		{"plot ignored", "plot"},
		{"strategy.entry ignored", "strategy.entry"},
		{"array.new_int ignored", "array.new_int"},
		{"map.new ignored", "map.new"},
		{"request.security ignored", "request.security"},
		{"ticker.heikinashi ignored", "ticker.heikinashi"},
		{"color.new ignored", "color.new"},
		{"year ignored", "year"},
		{"timeframe.in_seconds ignored", "timeframe.in_seconds"},
		{"unknown_function ignored", "unknown_function"},
		{"empty string ignored", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			handler := NewStringNamespaceHandler()

			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: tt.funcName},
				Arguments: []ast.Expression{&ast.Literal{Value: "test"}},
			}

			code, err := handler.GenerateCode(g, call)
			if err != nil {
				t.Fatalf("GenerateCode() should not error for non-handled functions, got: %v", err)
			}
			if code != "" {
				t.Errorf("GenerateCode() should return empty string for %q, got: %q", tt.funcName, code)
			}
		})
	}
}
