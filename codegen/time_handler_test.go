package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestSessionArgumentParser_ParseLiteral(t *testing.T) {
	parser := NewSessionArgumentParser()

	tests := []struct {
		name     string
		input    ast.Expression
		expected SessionArgument
	}{
		{
			name: "string literal with double quotes",
			input: &ast.Literal{
				Value: `"0950-1645"`,
			},
			expected: SessionArgument{
				Type:  ArgumentTypeLiteral,
				Value: "0950-1645",
			},
		},
		{
			name: "string literal with single quotes",
			input: &ast.Literal{
				Value: `'0950-1645'`,
			},
			expected: SessionArgument{
				Type:  ArgumentTypeLiteral,
				Value: "0950-1645",
			},
		},
		{
			name: "non-string literal",
			input: &ast.Literal{
				Value: 123,
			},
			expected: SessionArgument{
				Type: ArgumentTypeUnknown,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)
			if result.Type != tt.expected.Type {
				t.Errorf("expected type %v, got %v", tt.expected.Type, result.Type)
			}
			if result.Value != tt.expected.Value {
				t.Errorf("expected value %q, got %q", tt.expected.Value, result.Value)
			}
		})
	}
}

func TestSessionArgumentParser_ParseIdentifier(t *testing.T) {
	parser := NewSessionArgumentParser()

	tests := []struct {
		name     string
		input    ast.Expression
		expected SessionArgument
	}{
		{
			name: "simple identifier",
			input: &ast.Identifier{
				Name: "entry_time_input",
			},
			expected: SessionArgument{
				Type:  ArgumentTypeIdentifier,
				Value: "entry_time_input",
			},
		},
		{
			name: "identifier with underscores",
			input: &ast.Identifier{
				Name: "my_session_var",
			},
			expected: SessionArgument{
				Type:  ArgumentTypeIdentifier,
				Value: "my_session_var",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)
			if result.Type != tt.expected.Type {
				t.Errorf("expected type %v, got %v", tt.expected.Type, result.Type)
			}
			if result.Value != tt.expected.Value {
				t.Errorf("expected value %q, got %q", tt.expected.Value, result.Value)
			}
		})
	}
}

func TestSessionArgumentParser_ParseWrappedIdentifier(t *testing.T) {
	parser := NewSessionArgumentParser()

	tests := []struct {
		name     string
		input    ast.Expression
		expected SessionArgument
	}{
		{
			name: "wrapped identifier with [0]",
			input: &ast.MemberExpression{
				Computed: true,
				Object: &ast.Identifier{
					Name: "my_session",
				},
				Property: &ast.Literal{
					Value: 0,
				},
			},
			expected: SessionArgument{
				Type:  ArgumentTypeWrappedIdentifier,
				Value: "my_session",
			},
		},
		{
			name: "non-computed member expression",
			input: &ast.MemberExpression{
				Computed: false,
				Object: &ast.Identifier{
					Name: "obj",
				},
				Property: &ast.Identifier{
					Name: "prop",
				},
			},
			expected: SessionArgument{
				Type:  ArgumentTypeWrappedIdentifier,
				Value: "obj.prop",
			},
		},
		{
			name: "wrapped with non-zero index",
			input: &ast.MemberExpression{
				Computed: true,
				Object: &ast.Identifier{
					Name: "my_session",
				},
				Property: &ast.Literal{
					Value: 1,
				},
			},
			expected: SessionArgument{
				Type: ArgumentTypeUnknown,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)
			if result.Type != tt.expected.Type {
				t.Errorf("expected type %v, got %v", tt.expected.Type, result.Type)
			}
			if result.Value != tt.expected.Value {
				t.Errorf("expected value %q, got %q", tt.expected.Value, result.Value)
			}
		})
	}
}

func TestSessionArgument_IsVariable(t *testing.T) {
	tests := []struct {
		name     string
		arg      SessionArgument
		expected bool
	}{
		{
			name:     "identifier is variable",
			arg:      SessionArgument{Type: ArgumentTypeIdentifier, Value: "var1"},
			expected: true,
		},
		{
			name:     "wrapped identifier is variable",
			arg:      SessionArgument{Type: ArgumentTypeWrappedIdentifier, Value: "var2"},
			expected: true,
		},
		{
			name:     "literal is not variable",
			arg:      SessionArgument{Type: ArgumentTypeLiteral, Value: "0950-1645"},
			expected: false,
		},
		{
			name:     "unknown is not variable",
			arg:      SessionArgument{Type: ArgumentTypeUnknown},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.arg.IsVariable()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestSessionArgument_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		arg      SessionArgument
		expected bool
	}{
		{
			name:     "literal with value is valid",
			arg:      SessionArgument{Type: ArgumentTypeLiteral, Value: "0950-1645"},
			expected: true,
		},
		{
			name:     "identifier with value is valid",
			arg:      SessionArgument{Type: ArgumentTypeIdentifier, Value: "var1"},
			expected: true,
		},
		{
			name:     "unknown type is invalid",
			arg:      SessionArgument{Type: ArgumentTypeUnknown, Value: "value"},
			expected: false,
		},
		{
			name:     "empty value is invalid",
			arg:      SessionArgument{Type: ArgumentTypeLiteral, Value: ""},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.arg.IsValid()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestTimeCodeGenerator_BareTimestampPaths(t *testing.T) {
	type genFunc func(varName string) string

	tests := []struct {
		name    string
		indent  string
		varName string
		method  func(g *TimeCodeGenerator) genFunc
	}{
		{"no_args_tab_indent", "\t", "myVar", func(g *TimeCodeGenerator) genFunc { return g.GenerateNoArguments }},
		{"no_args_double_indent", "\t\t", "tSeries", func(g *TimeCodeGenerator) genFunc { return g.GenerateNoArguments }},
		{"single_arg_tab_indent", "\t", "myVar", func(g *TimeCodeGenerator) genFunc { return g.GenerateSingleArgument }},
		{"single_arg_double_indent", "\t\t", "tSeries", func(g *TimeCodeGenerator) genFunc { return g.GenerateSingleArgument }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewTimeCodeGenerator(tt.indent)
			result := tt.method(gen)(tt.varName)
			if !strings.Contains(result, barTimestampMsExpr) {
				t.Errorf("want ms timestamp %q in output, got:\n%s", barTimestampMsExpr, result)
			}
			if !strings.Contains(result, tt.varName+"Series.Set(") {
				t.Errorf("want %sSeries.Set(...) in output, got:\n%s", tt.varName, result)
			}
			if !strings.HasPrefix(result, tt.indent) {
				t.Errorf("want indentation %q prefix, got:\n%s", tt.indent, result)
			}
		})
	}
}

func TestTimeCodeGenerator_SessionPaths(t *testing.T) {
	tests := []struct {
		name           string
		session        SessionArgument
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:         "literal_session_quoted",
			session:      SessionArgument{Type: ArgumentTypeLiteral, Value: "0950-1645"},
			wantContains: []string{`"0950-1645"`, "session.TimeFunc", "ctx.Timezone"},
		},
		{
			name:           "identifier_session_unquoted",
			session:        SessionArgument{Type: ArgumentTypeIdentifier, Value: "entry_time_input"},
			wantContains:   []string{"entry_time_input", "session.TimeFunc"},
			wantNotContain: []string{`"entry_time_input"`},
		},
		{
			name:         "invalid_session_emits_nan",
			session:      SessionArgument{Type: ArgumentTypeUnknown},
			wantContains: []string{"math.NaN()"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewTimeCodeGenerator("\t")
			result := gen.GenerateWithSession("myVar", tt.session)
			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("want %q in output, got:\n%s", want, result)
				}
			}
			for _, notWant := range tt.wantNotContain {
				if strings.Contains(result, notWant) {
					t.Errorf("must not contain %q in output, got:\n%s", notWant, result)
				}
			}
		})
	}
}

func TestTimeHandler_HandleVariableInit(t *testing.T) {
	tests := []struct {
		name           string
		args           []ast.Expression
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:         "zero_args_emits_ms_timestamp",
			args:         []ast.Expression{},
			wantContains: []string{barTimestampMsExpr},
		},
		{
			name:         "single_arg_timeframe_emits_ms_timestamp",
			args:         []ast.Expression{&ast.Identifier{Name: "timeframe.period"}},
			wantContains: []string{barTimestampMsExpr},
		},
		{
			name: "two_args_literal_session_emits_session_filter",
			args: []ast.Expression{
				&ast.Identifier{Name: "timeframe.period"},
				&ast.Literal{Value: `"0950-1645"`},
			},
			wantContains:   []string{"session.TimeFunc", `"0950-1645"`},
			wantNotContain: []string{"math.NaN()"},
		},
		{
			name: "two_args_variable_session_unquoted",
			args: []ast.Expression{
				&ast.Identifier{Name: "timeframe.period"},
				&ast.Identifier{Name: "my_session_var"},
			},
			wantContains:   []string{"session.TimeFunc", "my_session_var"},
			wantNotContain: []string{`"my_session_var"`},
		},
		{
			name: "two_args_invalid_session_emits_nan",
			args: []ast.Expression{
				&ast.Identifier{Name: "timeframe.period"},
				&ast.Literal{Value: 123},
			},
			wantContains: []string{"math.NaN()"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewTimeHandler("\t")
			result := handler.HandleVariableInit("testVar", &ast.CallExpression{Arguments: tt.args})
			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("want %q in output, got:\n%s", want, result)
				}
			}
			for _, notWant := range tt.wantNotContain {
				if strings.Contains(result, notWant) {
					t.Errorf("must not contain %q in output, got:\n%s", notWant, result)
				}
			}
		})
	}
}

func TestTimeHandler_HandleInlineExpression(t *testing.T) {
	tests := []struct {
		name           string
		args           []ast.Expression
		wantContains   []string
		wantExact      string
		wantNotContain []string
	}{
		{
			name:      "zero_args_returns_ms_timestamp",
			args:      []ast.Expression{},
			wantExact: barTimestampMsExpr,
		},
		{
			name:      "single_arg_timeframe_only_returns_ms_timestamp",
			args:      []ast.Expression{&ast.Identifier{Name: "timeframe.period"}},
			wantExact: barTimestampMsExpr,
		},
		{
			name: "two_args_literal_session_emits_session_filter",
			args: []ast.Expression{
				&ast.Identifier{Name: "timeframe.period"},
				&ast.Literal{Value: `"0950-1645"`},
			},
			wantContains:   []string{"session.TimeFunc", `"0950-1645"`},
			wantNotContain: []string{"math.NaN()"},
		},
		{
			name: "two_args_variable_session_unquoted",
			args: []ast.Expression{
				&ast.Identifier{Name: "timeframe.period"},
				&ast.Identifier{Name: "entry_time"},
			},
			wantContains:   []string{"session.TimeFunc", "entry_time"},
			wantNotContain: []string{`"entry_time"`},
		},
		{
			name: "two_args_invalid_session_returns_nan",
			args: []ast.Expression{
				&ast.Identifier{Name: "timeframe.period"},
				&ast.Literal{Value: 123},
			},
			wantExact: "math.NaN()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewTimeHandler("\t")
			result := handler.HandleInlineExpression(tt.args)
			if tt.wantExact != "" && result != tt.wantExact {
				t.Errorf("want exact %q, got %q", tt.wantExact, result)
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("want %q in output, got: %s", want, result)
				}
			}
			for _, notWant := range tt.wantNotContain {
				if strings.Contains(result, notWant) {
					t.Errorf("must not contain %q in output, got: %s", notWant, result)
				}
			}
		})
	}
}

// The registry constructs TimeHandlers at setup time with a v5 default; strategies may be v4.
func TestTimeHandler_GenerateInline_PineVersionBinding(t *testing.T) {
	sessionArgs := []ast.Expression{
		&ast.Identifier{Name: "timeframe.period"},
		&ast.Literal{Value: `"0930-1600"`},
	}
	expr := &ast.CallExpression{Arguments: sessionArgs}

	tests := []struct {
		name         string
		handlerVer   int
		generatorVer int
		wantVersion  string
	}{
		{"handler_v4_no_generator", 4, 0, ", 4)"},
		{"handler_v5_no_generator", 5, 0, ", 5)"},
		{"generator_v4_overrides_handler_v5", 5, 4, ", 4)"},
		{"generator_v5_overrides_handler_v4", 4, 5, ", 5)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewTimeHandlerWithVersion("", tt.handlerVer)
			var g *generator
			if tt.generatorVer != 0 {
				g = newTestGenerator()
				g.pineVersion = tt.generatorVer
			}
			result, err := handler.GenerateInline(expr, g)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(result, tt.wantVersion) {
				t.Errorf("want version suffix %q in output, got: %s", tt.wantVersion, result)
			}
		})
	}
}
