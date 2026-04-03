package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestInputValueKindOf(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		want     InputValueKind
	}{
		{"float explicit", "input.float", InputValueFloat},
		{"price maps to float", "input.price", InputValueFloat},
		{"int explicit", "input.int", InputValueInt},
		{"time maps to int", "input.time", InputValueInt},
		{"bool", "input.bool", InputValueBool},
		{"string explicit", "input.string", InputValueString},
		{"symbol maps to string", "input.symbol", InputValueString},
		{"timeframe maps to string", "input.timeframe", InputValueString},
		{"text_area maps to string", "input.text_area", InputValueString},
		{"session distinct kind", "input.session", InputValueSession},
		{"color", "input.color", InputValueColor},
		{"source series kind", "input.source", InputValueSource},
		{"unimplemented input.enum", "input.enum", InputValueUnknown},
		{"non-input function", "ta.sma", InputValueUnknown},
		{"empty string", "", InputValueUnknown},
		{"partial match input prefix", "input", InputValueUnknown},
		{"case sensitivity", "Input.Float", InputValueUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InputValueKindOf(tt.funcName)
			if got != tt.want {
				t.Errorf("InputValueKindOf(%q) = %d, want %d", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestInputValueKindOf_AllRegisteredKinds(t *testing.T) {
	allFuncNames := []string{
		"input.float", "input.int", "input.bool", "input.string",
		"input.session", "input.source", "input.symbol", "input.timeframe",
		"input.text_area", "input.price", "input.time", "input.color",
	}

	seenKinds := make(map[InputValueKind]bool)
	for _, name := range allFuncNames {
		kind := InputValueKindOf(name)
		if kind == InputValueUnknown {
			t.Errorf("registered input function %q returned InputValueUnknown", name)
		}
		seenKinds[kind] = true
	}

	expectedKinds := []InputValueKind{
		InputValueFloat, InputValueInt, InputValueBool,
		InputValueString, InputValueSession, InputValueColor, InputValueSource,
	}
	for _, expected := range expectedKinds {
		if !seenKinds[expected] {
			t.Errorf("expected kind %d never returned by any registered function", expected)
		}
	}
}

func TestIsInputFuncName(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{"float", "input.float", true},
		{"int", "input.int", true},
		{"bool", "input.bool", true},
		{"string", "input.string", true},
		{"session", "input.session", true},
		{"source series", "input.source", true},
		{"symbol", "input.symbol", true},
		{"timeframe", "input.timeframe", true},
		{"text_area", "input.text_area", true},
		{"price", "input.price", true},
		{"time", "input.time", true},
		{"color", "input.color", true},
		{"unimplemented enum", "input.enum", false},
		{"ta function", "ta.sma", false},
		{"empty", "", false},
		{"prefix only", "input", false},
		{"case mismatch", "Input.Float", false},
		{"whitespace", " input.float", false},
		{"suffix whitespace", "input.float ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsInputFuncName(tt.funcName)
			if got != tt.want {
				t.Errorf("IsInputFuncName(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestIsInputFuncName_ConsistencyWithInputValueKindOf(t *testing.T) {
	testCases := []string{
		"input.float", "input.int", "input.bool", "input.string",
		"input.session", "input.source", "input.color",
		"input.enum", "ta.sma", "", "unknown",
	}

	for _, name := range testCases {
		isInput := IsInputFuncName(name)
		kind := InputValueKindOf(name)
		kindIndicatesInput := kind != InputValueUnknown

		if isInput != kindIndicatesInput {
			t.Errorf("inconsistency for %q: IsInputFuncName=%v but kind=%d", name, isInput, kind)
		}
	}
}

func TestIsInputConstantFuncName(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{"float constant", "input.float", true},
		{"int constant", "input.int", true},
		{"bool constant", "input.bool", true},
		{"string constant", "input.string", true},
		{"session constant", "input.session", true},
		{"symbol constant", "input.symbol", true},
		{"timeframe constant", "input.timeframe", true},
		{"text_area constant", "input.text_area", true},
		{"price constant", "input.price", true},
		{"time constant", "input.time", true},
		{"color constant", "input.color", true},
		{"source is series not constant", "input.source", false},
		{"unimplemented enum", "input.enum", false},
		{"ta function", "ta.sma", false},
		{"empty", "", false},
		{"prefix only", "input", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsInputConstantFuncName(tt.funcName)
			if got != tt.want {
				t.Errorf("IsInputConstantFuncName(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestIsInputConstantFuncName_ConsistencyWithInputValueKindOf(t *testing.T) {
	allFuncNames := []string{
		"input.float", "input.int", "input.bool", "input.string",
		"input.session", "input.source", "input.symbol", "input.timeframe",
		"input.text_area", "input.price", "input.time", "input.color",
	}

	for _, name := range allFuncNames {
		isConstant := IsInputConstantFuncName(name)
		kind := InputValueKindOf(name)

		if kind == InputValueSource && isConstant {
			t.Errorf("%q has kind InputValueSource but IsInputConstantFuncName returned true", name)
		}

		if kind != InputValueSource && kind != InputValueUnknown && !isConstant {
			t.Errorf("%q has constant kind %d but IsInputConstantFuncName returned false", name, kind)
		}
	}
}

func TestInputFuncNameFromLiteral(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  string
	}{
		{"int_positive", 10, "input.int"},
		{"int_negative", -5, "input.int"},
		{"int_zero", 0, "input.int"},
		{"float64_whole_number", 20.0, "input.int"},
		{"float64_fractional", 3.14, "input.float"},
		{"float64_negative", -2.5, "input.float"},
		{"float64_zero", 0.0, "input.int"},
		{"bool_true", true, "input.bool"},
		{"bool_false", false, "input.bool"},
		{"string_value", "EMA", "input.string"},
		{"string_empty", "", "input.string"},
		{"nil_value", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inputFuncNameFromLiteral(&ast.Literal{Value: tt.value})
			if got != tt.want {
				t.Errorf("inputFuncNameFromLiteral(%v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

/* extractDefvalExpression normalizes defval from positional or named arguments */

func TestExtractDefvalExpression(t *testing.T) {
	tests := []struct {
		name string
		call *ast.CallExpression
		want interface{}
	}{
		{"positional_literal", &ast.CallExpression{
			Arguments: []ast.Expression{&ast.Literal{Value: 14}},
		}, 14},
		{"positional_identifier", &ast.CallExpression{
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}, "close"},
		{"named_defval", inputCallWithNamedDefval(&ast.Literal{Value: true}), true},
		{"named_defval_identifier", inputCallWithNamedDefval(&ast.Identifier{Name: "hl2"}), "hl2"},
		{"no_arguments", &ast.CallExpression{Arguments: []ast.Expression{}}, nil},
		{"nil_arguments", &ast.CallExpression{}, nil},
		{"object_without_defval", &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.ObjectExpression{Properties: []ast.Property{
					{Key: &ast.Identifier{Name: "title"}, Value: &ast.Literal{Value: "X"}},
				}},
			},
		}, nil},
		{"empty_object", &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.ObjectExpression{Properties: []ast.Property{}},
			},
		}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDefvalExpression(tt.call)
			if tt.want == nil {
				if got != nil {
					t.Errorf("extractDefvalExpression() = %v, want nil", got)
				}
				return
			}
			switch expected := tt.want.(type) {
			case int:
				lit, ok := got.(*ast.Literal)
				if !ok || lit.Value != expected {
					t.Errorf("extractDefvalExpression() = %v, want Literal{%v}", got, expected)
				}
			case bool:
				lit, ok := got.(*ast.Literal)
				if !ok || lit.Value != expected {
					t.Errorf("extractDefvalExpression() = %v, want Literal{%v}", got, expected)
				}
			case string:
				ident, ok := got.(*ast.Identifier)
				if !ok || ident.Name != expected {
					t.Errorf("extractDefvalExpression() = %v, want Identifier{%s}", got, expected)
				}
			}
		})
	}
}

func TestFindPropertyValue(t *testing.T) {
	obj := &ast.ObjectExpression{
		Properties: []ast.Property{
			{Key: &ast.Identifier{Name: "defval"}, Value: &ast.Literal{Value: 42}},
			{Key: &ast.Identifier{Name: "title"}, Value: &ast.Literal{Value: "Test"}},
		},
	}

	if got := findPropertyValue(obj, "defval"); got == nil {
		t.Error("findPropertyValue(defval) returned nil, want non-nil")
	}
	if got := findPropertyValue(obj, "missing"); got != nil {
		t.Error("findPropertyValue(missing) returned non-nil, want nil")
	}

	/* Non-Identifier key (e.g. Literal) is safely skipped */
	objWithLiteralKey := &ast.ObjectExpression{
		Properties: []ast.Property{
			{Key: &ast.Literal{Value: "defval"}, Value: &ast.Literal{Value: 42}},
		},
	}
	if got := findPropertyValue(objWithLiteralKey, "defval"); got != nil {
		t.Error("findPropertyValue() matched non-Identifier key, want nil")
	}

	objEmpty := &ast.ObjectExpression{}
	if got := findPropertyValue(objEmpty, "defval"); got != nil {
		t.Error("findPropertyValue() returned non-nil for empty properties")
	}
}

/* Searches all ObjectExpression arguments for named property */

func TestFindNamedArgValue(t *testing.T) {
	/* Property found in second ObjectExpression */
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.ObjectExpression{Properties: []ast.Property{
				{Key: &ast.Identifier{Name: "title"}, Value: &ast.Literal{Value: "Length"}},
			}},
			&ast.ObjectExpression{Properties: []ast.Property{
				{Key: &ast.Identifier{Name: "minval"}, Value: &ast.Literal{Value: 1}},
			}},
		},
	}

	if got := findNamedArgValue(call, "minval"); got == nil {
		t.Error("findNamedArgValue(minval) returned nil, want non-nil")
	}
	if got := findNamedArgValue(call, "missing"); got != nil {
		t.Error("findNamedArgValue(missing) returned non-nil, want nil")
	}

	/* Non-ObjectExpression arguments are skipped */
	callMixed := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Literal{Value: 14},
			&ast.ObjectExpression{Properties: []ast.Property{
				{Key: &ast.Identifier{Name: "title"}, Value: &ast.Literal{Value: "X"}},
			}},
		},
	}
	if got := findNamedArgValue(callMixed, "title"); got == nil {
		t.Error("findNamedArgValue(title) skipped non-object arg but missed property")
	}

	/* No ObjectExpression arguments at all */
	callPositional := &ast.CallExpression{
		Arguments: []ast.Expression{&ast.Literal{Value: 14}},
	}
	if got := findNamedArgValue(callPositional, "title"); got != nil {
		t.Error("findNamedArgValue() returned non-nil for call with no ObjectExpressions")
	}

	callEmpty := &ast.CallExpression{}
	if got := findNamedArgValue(callEmpty, "title"); got != nil {
		t.Error("findNamedArgValue() returned non-nil for empty call")
	}
}

/* v4 type=input.* explicit type parameter detection */

func TestResolveFromExplicitTypeParam(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		want     string
	}{
		{"integer", "integer", "input.int"},
		{"float", "float", "input.float"},
		{"bool", "bool", "input.bool"},
		{"string", "string", "input.string"},
		{"session", "session", "input.session"},
		{"source", "source", "input.source"},
		{"symbol", "symbol", "input.symbol"},
		{"time", "time", "input.time"},
		{"timeframe", "timeframe", "input.timeframe"},
		{"price", "price", "input.price"},
		{"color", "color", "input.color"},
		{"text_area", "text_area", "input.text_area"},
		{"unknown_type", "custom", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := inputCallWithTypeParam(tt.typeName)
			got := resolveFromExplicitTypeParam(call)
			if got != tt.want {
				t.Errorf("resolveFromExplicitTypeParam() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveFromExplicitTypeParam_NonMemberExpression(t *testing.T) {
	/* type=bool (bare Identifier, not MemberExpression) returns "" */
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.ObjectExpression{Properties: []ast.Property{
				{Key: &ast.Identifier{Name: "type"}, Value: &ast.Identifier{Name: "bool"}},
			}},
		},
	}
	if got := resolveFromExplicitTypeParam(call); got != "" {
		t.Errorf("resolveFromExplicitTypeParam(bare Identifier) = %q, want empty", got)
	}
}

func TestResolveFromExplicitTypeParam_WrongObjectPrefix(t *testing.T) {
	/* type=wrong.integer returns "" */
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.ObjectExpression{Properties: []ast.Property{
				{Key: &ast.Identifier{Name: "type"}, Value: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "wrong"},
					Property: &ast.Identifier{Name: "integer"},
				}},
			}},
		},
	}
	if got := resolveFromExplicitTypeParam(call); got != "" {
		t.Errorf("resolveFromExplicitTypeParam(wrong prefix) = %q, want empty", got)
	}
}

func TestResolveFromExplicitTypeParam_NoTypeParam(t *testing.T) {
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.ObjectExpression{Properties: []ast.Property{
				{Key: &ast.Identifier{Name: "defval"}, Value: &ast.Literal{Value: 10}},
			}},
		},
	}
	if got := resolveFromExplicitTypeParam(call); got != "" {
		t.Errorf("resolveFromExplicitTypeParam(no type) = %q, want empty", got)
	}
}

func TestResolveFromDefvalType_Literals(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  string
	}{
		{"int", 10, "input.int"},
		{"float", 3.14, "input.float"},
		{"bool", true, "input.bool"},
		{"string", "EMA", "input.string"},
	}

	for _, tt := range tests {
		t.Run("positional_"+tt.name, func(t *testing.T) {
			call := &ast.CallExpression{Arguments: []ast.Expression{&ast.Literal{Value: tt.value}}}
			if got := resolveFromDefvalType(call); got != tt.want {
				t.Errorf("resolveFromDefvalType() = %q, want %q", got, tt.want)
			}
		})

		t.Run("named_"+tt.name, func(t *testing.T) {
			call := inputCallWithNamedDefval(&ast.Literal{Value: tt.value})
			if got := resolveFromDefvalType(call); got != tt.want {
				t.Errorf("resolveFromDefvalType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveFromDefvalType_SourceIdentifiers(t *testing.T) {
	sources := []string{"hl2", "hlc3", "ohlc4", "hlcc4", "open", "high", "low", "close", "volume"}

	for _, src := range sources {
		t.Run(src, func(t *testing.T) {
			call := inputCallWithNamedDefval(&ast.Identifier{Name: src})
			if got := resolveFromDefvalType(call); got != "input.source" {
				t.Errorf("resolveFromDefvalType(%s) = %q, want %q", src, got, "input.source")
			}
		})
	}
}

func TestResolveFromDefvalType_UnknownIdentifier(t *testing.T) {
	call := &ast.CallExpression{
		Arguments: []ast.Expression{&ast.Identifier{Name: "customVar"}},
	}
	if got := resolveFromDefvalType(call); got != "" {
		t.Errorf("resolveFromDefvalType(unknown identifier) = %q, want empty", got)
	}
}

func TestResolveFromDefvalType_ComplexExpression(t *testing.T) {
	/* CallExpression as defval is unresolvable */
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.CallExpression{Callee: &ast.Identifier{Name: "syminfo.tickerid"}},
		},
	}
	if got := resolveFromDefvalType(call); got != "" {
		t.Errorf("resolveFromDefvalType(CallExpression) = %q, want empty", got)
	}
}

/* Strategy chain: explicit type takes precedence over defval inference */

func TestResolveInputFuncName_ExplicitTypePrecedence(t *testing.T) {
	/* Explicit type=input.float overrides defval integer inference */
	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.ObjectExpression{Properties: []ast.Property{
				{Key: &ast.Identifier{Name: "defval"}, Value: &ast.Literal{Value: 10}},
				{Key: &ast.Identifier{Name: "type"}, Value: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "float"},
				}},
			}},
		},
	}

	if got := resolveInputFuncName(call); got != "input.float" {
		t.Errorf("resolveInputFuncName() = %q, want %q", got, "input.float")
	}
}

func TestResolveInputFuncName_FallsBackToDefval(t *testing.T) {
	call := inputCallWithNamedDefval(&ast.Literal{Value: true})
	if got := resolveInputFuncName(call); got != "input.bool" {
		t.Errorf("resolveInputFuncName() = %q, want %q", got, "input.bool")
	}
}

func TestResolveInputFuncName_Unresolvable(t *testing.T) {
	tests := []struct {
		name string
		call *ast.CallExpression
	}{
		{"no_arguments", &ast.CallExpression{Arguments: []ast.Expression{}}},
		{"nil_arguments", &ast.CallExpression{}},
		{"unknown_identifier", &ast.CallExpression{
			Arguments: []ast.Expression{&ast.Identifier{Name: "customValue"}},
		}},
		{"object_without_defval_or_type", &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.ObjectExpression{Properties: []ast.Property{
					{Key: &ast.Identifier{Name: "title"}, Value: &ast.Literal{Value: "Length"}},
				}},
			},
		}},
		{"unknown_type_no_defval", &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.ObjectExpression{Properties: []ast.Property{
					{Key: &ast.Identifier{Name: "type"}, Value: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "input"},
						Property: &ast.Identifier{Name: "custom"},
					}},
				}},
			},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveInputFuncName(tt.call); got != "" {
				t.Errorf("resolveInputFuncName() = %q, want empty", got)
			}
		})
	}
}

/* Helpers */

func inputCallWithTypeParam(typeName string) *ast.CallExpression {
	return &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.ObjectExpression{Properties: []ast.Property{
				{Key: &ast.Identifier{Name: "title"}, Value: &ast.Literal{Value: "Test"}},
				{Key: &ast.Identifier{Name: "type"}, Value: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: typeName},
				}},
				{Key: &ast.Identifier{Name: "defval"}, Value: &ast.Literal{Value: 10}},
			}},
		},
	}
}

func inputCallWithNamedDefval(value ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.ObjectExpression{Properties: []ast.Property{
				{Key: &ast.Identifier{Name: "defval"}, Value: value},
				{Key: &ast.Identifier{Name: "title"}, Value: &ast.Literal{Value: "Test"}},
			}},
		},
	}
}
