package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func strategyNamedCall(funcName string, args ...ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Identifier{Name: funcName},
		},
		Arguments: args,
	}
}

func strategyNamedArgs(props ...ast.Property) ast.Expression {
	return &ast.ObjectExpression{Properties: props}
}

func strategyNamedProp(name string, value ast.Expression) ast.Property {
	return ast.Property{Key: &ast.Identifier{Name: name}, Value: value}
}

func strategyLongMember() ast.Expression {
	return &ast.MemberExpression{
		Object:   &ast.Identifier{Name: "strategy"},
		Property: &ast.Identifier{Name: "long"},
	}
}

func unsupportedCall(name string) *ast.CallExpression {
	return &ast.CallExpression{
		Callee:    &ast.Identifier{Name: name},
		Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
	}
}

// TestStrategyCallArgumentsResolveNamedAndPositional verifies a strategy
// command's parameters resolve identically whether they are written
// positionally, by name, or mixed, and that a name wins over the same slot.
func TestStrategyCallArgumentsResolveNamedAndPositional(t *testing.T) {
	idLiteral := &ast.Literal{Value: "L"}
	direction := strategyLongMember()
	positionalComment := &ast.Literal{Value: "positional-comment"}
	namedComment := &ast.Literal{Value: "named-comment"}

	positionalWithComment := func() []ast.Expression {
		return []ast.Expression{
			idLiteral, direction,
			naIdentifier(), naIdentifier(), naIdentifier(), naIdentifier(), naIdentifier(),
			positionalComment,
		}
	}

	tests := []struct {
		name      string
		args      []ast.Expression
		parameter string
		slot      int
		want      ast.Expression
		wantOK    bool
	}{
		{name: "positional argument by slot", args: []ast.Expression{idLiteral, direction}, parameter: "id", slot: 0, want: idLiteral, wantOK: true},
		{name: "named argument by name", args: []ast.Expression{strategyNamedArgs(strategyNamedProp("id", idLiteral))}, parameter: "id", slot: 0, want: idLiteral, wantOK: true},
		{name: "named argument wins over the positional slot", args: append(positionalWithComment(), strategyNamedArgs(strategyNamedProp("comment", namedComment))), parameter: "comment", slot: 7, want: namedComment, wantOK: true},
		{name: "positional argument used when no name matches", args: positionalWithComment(), parameter: "comment", slot: 7, want: positionalComment, wantOK: true},
		{name: "named-only parameter resolves by name", args: []ast.Expression{strategyNamedArgs(strategyNamedProp("when", idLiteral))}, parameter: "when", slot: -1, want: idLiteral, wantOK: true},
		{name: "absent parameter", args: []ast.Expression{idLiteral, direction}, parameter: "comment", slot: 7, wantOK: false},
		{name: "absent named-only parameter", args: []ast.Expression{idLiteral}, parameter: "when", slot: -1, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := newStrategyCallArguments(tt.args).value(tt.parameter, tt.slot)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && got != tt.want {
				t.Fatalf("value = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestStrategyCallArgumentsBoundaries verifies the resolver's empty, negative
// slot, out-of-range, and container-fallback behaviour.
func TestStrategyCallArgumentsBoundaries(t *testing.T) {
	empty := newStrategyCallArguments(nil)
	if _, ok := empty.value("id", 0); ok {
		t.Fatal("empty argument list must not resolve a parameter")
	}
	if _, ok := empty.positionalValue(-1); ok {
		t.Fatal("negative slot must not resolve")
	}
	if _, ok := empty.positionalValue(0); ok {
		t.Fatal("out-of-range slot must not resolve")
	}
	if _, ok := empty.valueOrContainer(1); ok {
		t.Fatal("no positional argument and no named container must not resolve")
	}

	container := strategyNamedArgs(strategyNamedProp("direction", strategyLongMember()))
	got, ok := newStrategyCallArguments([]ast.Expression{&ast.Literal{Value: "L"}, container}).valueOrContainer(1)
	if !ok || got != container {
		t.Fatalf("valueOrContainer must return the named container when the slot is absent: got %v, ok %v", got, ok)
	}
}

// TestStrategyCallArgumentsMergeMultipleNamedContainers verifies a call whose
// named arguments are split across several containers resolves every parameter,
// whichever container holds it.
func TestStrategyCallArgumentsMergeMultipleNamedContainers(t *testing.T) {
	direction := strategyLongMember()
	profit := &ast.Literal{Value: "TP"}
	loss := &ast.Literal{Value: "SL"}
	first := strategyNamedArgs(
		strategyNamedProp("direction", direction),
		strategyNamedProp("comment_profit", profit),
	)
	second := strategyNamedArgs(strategyNamedProp("comment_loss", loss))

	call := newStrategyCallArguments([]ast.Expression{&ast.Literal{Value: "X"}, first, second})

	for _, want := range []struct {
		parameter string
		value     ast.Expression
	}{
		{"direction", direction},
		{"comment_profit", profit},
		{"comment_loss", loss},
	} {
		got, ok := call.namedValue(want.parameter)
		if !ok || got != want.value {
			t.Fatalf("%s = %v, ok %v, want the value from its own container", want.parameter, got, ok)
		}
	}
}

// TestCommentExpressionSubstitution verifies a comment value is flagged as
// substituted exactly when it resolves to a feature-gap sentinel.
func TestCommentExpressionSubstitution(t *testing.T) {
	gen := newTestGenerator()
	gen.variables["comment_msg"] = "string"
	extractor := &ArgumentExtractor{generator: gen}

	tests := []struct {
		name            string
		expr            ast.Expression
		wantSubstituted bool
	}{
		{name: "string literal", expr: &ast.Literal{Value: "x"}},
		{name: "known string variable", expr: &ast.Identifier{Name: "comment_msg"}},
		{name: "unknown variable keeps series access", expr: &ast.Identifier{Name: "legacy_msg"}},
		{name: "unsupported call substitutes", expr: unsupportedCall("unsupported_comment"), wantSubstituted: true},
		{
			name: "conditional substitutes when a branch substitutes",
			expr: &ast.ConditionalExpression{
				Test:       &ast.Identifier{Name: "flag"},
				Consequent: unsupportedCall("unsupported_comment"),
				Alternate:  &ast.Literal{Value: "fallback"},
			},
			wantSubstituted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, substituted := extractor.commentExpression(tt.expr)
			if substituted != tt.wantSubstituted {
				t.Fatalf("substituted = %v, want %v", substituted, tt.wantSubstituted)
			}
		})
	}
}

// TestCommentAnnotationsSelectPresentParameters verifies only present comment
// parameters are returned, in Pine parameter order, with the named form winning.
func TestCommentAnnotationsSelectPresentParameters(t *testing.T) {
	extractor := &ArgumentExtractor{generator: newTestGenerator()}

	named := extractor.commentAnnotations("strategy.exit", strategyNamedCall("exit",
		&ast.Literal{Value: "X"}, &ast.Literal{Value: "L"},
		strategyNamedArgs(
			strategyNamedProp("comment_profit", &ast.Literal{Value: "TP"}),
			strategyNamedProp("comment", &ast.Literal{Value: "C"}),
		)).Arguments)

	if len(named) != 2 {
		t.Fatalf("annotations = %+v, want comment and comment_profit", named)
	}
	if named[0].parameter != "comment" || named[0].code != `"C"` {
		t.Fatalf("annotation[0] = %+v, want comment=%q", named[0], `"C"`)
	}
	if named[1].parameter != "comment_profit" || named[1].code != `"TP"` {
		t.Fatalf("annotation[1] = %+v, want comment_profit=%q", named[1], `"TP"`)
	}

	positionalArgs := make([]ast.Expression, 13)
	for i := range positionalArgs {
		positionalArgs[i] = naIdentifier()
	}
	positionalArgs[0] = &ast.Literal{Value: "X"}
	positionalArgs[1] = &ast.Literal{Value: "L"}
	positionalArgs[12] = &ast.Literal{Value: "positional-C"}

	mixed := extractor.commentAnnotations("strategy.exit", append(
		append([]ast.Expression{}, positionalArgs...),
		strategyNamedArgs(strategyNamedProp("comment", &ast.Literal{Value: "named-C"})),
	))

	if len(mixed) != 1 || mixed[0].code != `"named-C"` {
		t.Fatalf("mixed annotations = %+v, want the named comment to win", mixed)
	}
}

// TestStrategyEntryResolvesAcrossCallShapes verifies strategy.entry resolves its
// id and direction in every Pine call shape instead of emitting an invalid stub.
func TestStrategyEntryResolvesAcrossCallShapes(t *testing.T) {
	tests := []struct {
		name string
		call *ast.CallExpression
	}{
		{name: "positional id and direction", call: strategyNamedCall("entry", &ast.Literal{Value: "L"}, strategyLongMember())},
		{name: "positional id and named direction", call: strategyNamedCall("entry", &ast.Literal{Value: "L"}, strategyNamedArgs(strategyNamedProp("direction", strategyLongMember())))},
		{name: "positional id and legacy named long", call: strategyNamedCall("entry", &ast.Literal{Value: "L"}, strategyNamedArgs(strategyNamedProp("long", strategyLongMember())))},
		{name: "fully named", call: strategyNamedCall("entry", strategyNamedArgs(
			strategyNamedProp("id", &ast.Literal{Value: "L"}),
			strategyNamedProp("direction", strategyLongMember()),
		))},
	}

	handler := NewStrategyActionHandler()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := handler.GenerateCode(newTestGenerator(), tt.call)
			if err != nil {
				t.Fatalf("GenerateCode() error: %v", err)
			}
			if !strings.Contains(code, `strat.Entry("L", strategy.Long`) {
				t.Fatalf("entry not resolved from shape: %q", code)
			}
			if strings.Contains(code, "invalid arguments") {
				t.Fatalf("valid shape emitted an invalid-args stub: %q", code)
			}
		})
	}
}

// TestStrategyEntryNamedCommentCarriesSentinel verifies a comment written
// alongside a named direction still reaches the generated result as a typed
// sentinel, for both the v5 direction= and the legacy long= key.
func TestStrategyEntryNamedCommentCarriesSentinel(t *testing.T) {
	handler := NewStrategyActionHandler()
	for _, key := range []string{"direction", "long"} {
		t.Run(key, func(t *testing.T) {
			call := strategyNamedCall("entry", &ast.Literal{Value: "L"},
				strategyNamedArgs(
					strategyNamedProp(key, strategyLongMember()),
					strategyNamedProp("comment", unsupportedCall("unsupported_comment")),
				))

			code, err := handler.GenerateCode(newTestGenerator(), call)
			if err != nil {
				t.Fatalf("GenerateCode() error: %v", err)
			}
			if !strings.Contains(code, `featuregap.RecordString("unsupported_comment"`) {
				t.Fatalf("named-direction comment dropped: %q", code)
			}
		})
	}
}

// TestStrategyCloseResolvesIDAcrossCallShapes verifies strategy.close keeps its
// entry id whether the id is positional or named.
func TestStrategyCloseResolvesIDAcrossCallShapes(t *testing.T) {
	tests := []struct {
		name string
		call *ast.CallExpression
	}{
		{name: "positional id", call: strategyNamedCall("close", &ast.Literal{Value: "L"})},
		{name: "named id", call: strategyNamedCall("close", strategyNamedArgs(strategyNamedProp("id", &ast.Literal{Value: "L"})))},
		{name: "named id with comment", call: strategyNamedCall("close", strategyNamedArgs(
			strategyNamedProp("id", &ast.Literal{Value: "L"}),
			strategyNamedProp("comment", &ast.Literal{Value: "c"}),
		))},
	}

	handler := NewStrategyActionHandler()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := handler.GenerateCode(newTestGenerator(), tt.call)
			if err != nil {
				t.Fatalf("GenerateCode() error: %v", err)
			}
			if !strings.Contains(code, `strat.Close("L"`) {
				t.Fatalf("close id lost: %q", code)
			}
		})
	}
}

// TestStrategyExitCommentVariantDiagnostics verifies each exit comment variant
// surfaces an unsupported value as a provenance record and stays silent for a
// supported one.
func TestStrategyExitCommentVariantDiagnostics(t *testing.T) {
	handler := NewStrategyActionHandler()
	values := []struct {
		name           string
		value          ast.Expression
		wantDiagnostic bool
	}{
		{name: "unsupported value emits a runtime reference", value: unsupportedCall("unsupported_variant"), wantDiagnostic: true},
		{name: "supported literal emits no runtime reference", value: &ast.Literal{Value: "TP"}},
	}

	for _, variant := range []string{"comment_profit", "comment_loss", "comment_trailing"} {
		for _, value := range values {
			t.Run(variant+"/"+value.name, func(t *testing.T) {
				call := strategyNamedCall("exit", &ast.Literal{Value: "X"}, &ast.Literal{Value: "L"},
					strategyNamedArgs(strategyNamedProp(variant, value.value)))

				code, err := handler.GenerateCode(newTestGenerator(), call)
				if err != nil {
					t.Fatalf("GenerateCode() error: %v", err)
				}
				if got := strings.Contains(code, "_ = featuregap.RecordString"); got != value.wantDiagnostic {
					t.Fatalf("diagnostic presence = %v, want %v; code = %q", got, value.wantDiagnostic, code)
				}
			})
		}
	}
}

// TestExpressionPositionDispatcher_DispatchStringAnnotationOutcomes verifies the
// dispatcher reports substitution only for a sentinel fallback and errors on an
// unsupported namespaced call or a missing router.
func TestExpressionPositionDispatcher_DispatchStringAnnotationOutcomes(t *testing.T) {
	gen := newTestGenerator()
	knownCall := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "heikinashi"},
		Arguments: []ast.Expression{
			&ast.MemberExpression{Object: &ast.Identifier{Name: "syminfo"}, Property: &ast.Identifier{Name: "tickerid"}},
		},
	}
	namespacedCall := &ast.CallExpression{
		Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "ta"}, Property: &ast.Identifier{Name: "nonexistent"}},
		Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
	}

	tests := []struct {
		name            string
		router          *CallExpressionRouter
		call            *ast.CallExpression
		wantSubstituted bool
		wantSentinel    bool
		wantErr         bool
	}{
		{name: "unknown bare call substitutes a sentinel", router: gen.callRouter, call: unsupportedCall("unsupported_comment"), wantSubstituted: true, wantSentinel: true},
		{name: "known function carries its value", router: gen.callRouter, call: knownCall},
		{name: "namespaced unknown errors", router: gen.callRouter, call: namespacedCall, wantErr: true},
		{name: "nil router errors", router: nil, call: unsupportedCall("unsupported_comment"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dispatcher := NewExpressionPositionDispatcher(tt.router)
			code, substituted, err := dispatcher.DispatchStringAnnotation(gen, tt.call, "generator.comment_unknown")

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if substituted != tt.wantSubstituted {
				t.Fatalf("substituted = %v, want %v", substituted, tt.wantSubstituted)
			}
			if tt.wantSentinel && !strings.HasPrefix(code, "featuregap.RecordString(") {
				t.Fatalf("code = %q, want a string sentinel", code)
			}
			if !tt.wantSentinel && strings.HasPrefix(code, "featuregap.RecordString(") {
				t.Fatalf("code = %q, want a carried value", code)
			}
		})
	}
}
