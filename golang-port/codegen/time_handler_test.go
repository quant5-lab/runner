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
				Type: ArgumentTypeUnknown,
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

func TestTimeCodeGenerator_GenerateNoArguments(t *testing.T) {
	gen := NewTimeCodeGenerator("\t")
	result := gen.GenerateNoArguments("myVar")

	expected := "\tmyVarSeries.Set(float64(ctx.Data[ctx.BarIndex].Time))\n"
	if result != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestTimeCodeGenerator_GenerateWithSession_Literal(t *testing.T) {
	gen := NewTimeCodeGenerator("\t")
	session := SessionArgument{
		Type:  ArgumentTypeLiteral,
		Value: "0950-1645",
	}

	result := gen.GenerateWithSession("myVar", session)

	if !strings.Contains(result, `"0950-1645"`) {
		t.Errorf("expected literal session string in quotes, got:\n%s", result)
	}
	if !strings.Contains(result, "session.TimeFunc") {
		t.Errorf("expected session.TimeFunc call, got:\n%s", result)
	}
	if !strings.Contains(result, "ctx.Timezone") {
		t.Errorf("expected ctx.Timezone parameter, got:\n%s", result)
	}
}

func TestTimeCodeGenerator_GenerateWithSession_Variable(t *testing.T) {
	gen := NewTimeCodeGenerator("\t")
	session := SessionArgument{
		Type:  ArgumentTypeIdentifier,
		Value: "entry_time_input",
	}

	result := gen.GenerateWithSession("myVar", session)

	if !strings.Contains(result, "entry_time_input") {
		t.Errorf("expected variable name without quotes, got:\n%s", result)
	}
	if strings.Contains(result, `"entry_time_input"`) {
		t.Errorf("variable should not be quoted, got:\n%s", result)
	}
	if !strings.Contains(result, "session.TimeFunc") {
		t.Errorf("expected session.TimeFunc call, got:\n%s", result)
	}
}

func TestTimeCodeGenerator_GenerateWithSession_Invalid(t *testing.T) {
	gen := NewTimeCodeGenerator("\t")
	session := SessionArgument{
		Type: ArgumentTypeUnknown,
	}

	result := gen.GenerateWithSession("myVar", session)

	if !strings.Contains(result, "math.NaN()") {
		t.Errorf("expected NaN for invalid session, got:\n%s", result)
	}
}

func TestTimeHandler_HandleVariableInit_NoArguments(t *testing.T) {
	handler := NewTimeHandler("\t")
	call := &ast.CallExpression{
		Arguments: []ast.Expression{},
	}

	result := handler.HandleVariableInit("testVar", call)

	if !strings.Contains(result, "float64(ctx.Data[ctx.BarIndex].Time)") {
		t.Errorf("expected timestamp without session filtering, got:\n%s", result)
	}
}

func TestTimeHandler_HandleVariableInit_SingleArgument(t *testing.T) {
	handler := NewTimeHandler("\t")
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "timeframe.period"},
		},
	}

	result := handler.HandleVariableInit("testVar", call)

	if !strings.Contains(result, "float64(ctx.Data[ctx.BarIndex].Time)") {
		t.Errorf("expected timestamp without session filtering, got:\n%s", result)
	}
}

func TestTimeHandler_HandleVariableInit_TwoArguments_Literal(t *testing.T) {
	handler := NewTimeHandler("\t")
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "timeframe.period"},
			&ast.Literal{Value: `"0950-1645"`},
		},
	}

	result := handler.HandleVariableInit("testVar", call)

	if !strings.Contains(result, "session.TimeFunc") {
		t.Errorf("expected session.TimeFunc call, got:\n%s", result)
	}
	if !strings.Contains(result, `"0950-1645"`) {
		t.Errorf("expected quoted session string, got:\n%s", result)
	}
}

func TestTimeHandler_HandleVariableInit_TwoArguments_Variable(t *testing.T) {
	handler := NewTimeHandler("\t")
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "timeframe.period"},
			&ast.Identifier{Name: "my_session_var"},
		},
	}

	result := handler.HandleVariableInit("testVar", call)

	if !strings.Contains(result, "session.TimeFunc") {
		t.Errorf("expected session.TimeFunc call, got:\n%s", result)
	}
	if !strings.Contains(result, "my_session_var") {
		t.Errorf("expected variable name, got:\n%s", result)
	}
	if strings.Contains(result, `"my_session_var"`) {
		t.Errorf("variable should not be quoted, got:\n%s", result)
	}
}

func TestTimeHandler_HandleInlineExpression_NoSession(t *testing.T) {
	handler := NewTimeHandler("\t")
	args := []ast.Expression{
		&ast.Identifier{Name: "timeframe.period"},
	}

	result := handler.HandleInlineExpression(args)

	expected := "float64(ctx.Data[ctx.BarIndex].Time)"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestTimeHandler_HandleInlineExpression_WithLiteralSession(t *testing.T) {
	handler := NewTimeHandler("\t")
	args := []ast.Expression{
		&ast.Identifier{Name: "timeframe.period"},
		&ast.Literal{Value: `"0950-1645"`},
	}

	result := handler.HandleInlineExpression(args)

	if !strings.Contains(result, "session.TimeFunc") {
		t.Errorf("expected session.TimeFunc call, got: %s", result)
	}
	if !strings.Contains(result, `"0950-1645"`) {
		t.Errorf("expected quoted session string, got: %s", result)
	}
}

func TestTimeHandler_HandleInlineExpression_WithVariableSession(t *testing.T) {
	handler := NewTimeHandler("\t")
	args := []ast.Expression{
		&ast.Identifier{Name: "timeframe.period"},
		&ast.Identifier{Name: "entry_time"},
	}

	result := handler.HandleInlineExpression(args)

	if !strings.Contains(result, "session.TimeFunc") {
		t.Errorf("expected session.TimeFunc call, got: %s", result)
	}
	if !strings.Contains(result, "entry_time") {
		t.Errorf("expected variable name, got: %s", result)
	}
}

func TestTimeHandler_HandleInlineExpression_InvalidSession(t *testing.T) {
	handler := NewTimeHandler("\t")
	args := []ast.Expression{
		&ast.Identifier{Name: "timeframe.period"},
		&ast.Literal{Value: 123}, // Invalid: not a string
	}

	result := handler.HandleInlineExpression(args)

	expected := "math.NaN()"
	if result != expected {
		t.Errorf("expected %q for invalid session, got %q", expected, result)
	}
}
