package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestAnalyzeCompatibility_BacktestCriticalSinks(t *testing.T) {
	tests := []struct {
		name        string
		call        *ast.CallExpression
		wantImpact  string
		wantFeature string
		wantSink    string
		wantSource  string
	}{
		{
			name:        "unknown variable controls entry condition",
			call:        strategyEntryCall(identifier("x")),
			wantImpact:  "backtest-critical",
			wantFeature: "unsupported_entry_source",
			wantSink:    "strategy.entry",
			wantSource:  "generator.variable_init_unknown",
		},
		{
			name:        "drawing getter controls entry condition",
			call:        strategyEntryCall(chartGetterCall("line", "get_price")),
			wantImpact:  "backtest-critical",
			wantFeature: "line.get_price",
			wantSink:    "strategy.entry",
			wantSource:  "chart_namespace_getter",
		},
		{
			name: "unknown variable in named quantity",
			call: call(member("strategy", "entry"),
				literal("L"),
				member("strategy", "long"),
				objectArg("qty", identifier("x")),
			),
			wantImpact:  "backtest-critical",
			wantFeature: "unsupported_entry_source",
			wantSink:    "strategy.entry",
		},
		{
			name: "unknown variable in named comment",
			call: call(member("strategy", "entry"),
				literal("L"),
				member("strategy", "long"),
				objectArg("comment", identifier("x")),
			),
			wantImpact:  "observable-non-backtest",
			wantFeature: "unsupported_entry_source",
			wantSource:  "generator.variable_init_unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := compatibilityProgram(tt.call)
			got := analyzeCompatibility(program, []string{"unsupported_entry_source"})
			if tt.wantImpact == "" {
				if len(got) != 0 {
					t.Fatalf("diagnostics = %+v, want none", got)
				}
				return
			}
			assertCompatibilityImpact(t, got, tt.wantFeature, tt.wantImpact, tt.wantSink)
			if tt.wantSource != "" && got[0].Source != tt.wantSource {
				t.Fatalf("diagnostic source = %q, want %q", got[0].Source, tt.wantSource)
			}
		})
	}
}

func TestAnalyzeCompatibility_StrategyArgumentImpactClassification(t *testing.T) {
	tests := []struct {
		name       string
		call       *ast.CallExpression
		wantImpact string
		wantSink   string
	}{
		{
			name:       "entry id selects the order",
			call:       call(member("strategy", "entry"), objectArg("id", identifier("x"))),
			wantImpact: "backtest-critical",
			wantSink:   "strategy.entry",
		},
		{
			name:       "exit from_entry selects the position",
			call:       call(member("strategy", "exit"), objectArg("from_entry", identifier("x"))),
			wantImpact: "backtest-critical",
			wantSink:   "strategy.exit",
		},
		{
			name:       "order oca name selects the cancellation group",
			call:       call(member("strategy", "order"), objectArg("oca_name", identifier("x"))),
			wantImpact: "backtest-critical",
			wantSink:   "strategy.order",
		},
		{
			name:       "order oca type changes cancellation semantics",
			call:       call(member("strategy", "order"), objectArg("oca_type", identifier("x"))),
			wantImpact: "backtest-critical",
			wantSink:   "strategy.order",
		},
		{
			name:       "cancel positional id selects the order",
			call:       call(member("strategy", "cancel"), identifier("x")),
			wantImpact: "backtest-critical",
			wantSink:   "strategy.cancel",
		},
		{
			name:       "close positional id selects the position",
			call:       call(member("strategy", "close"), identifier("x")),
			wantImpact: "backtest-critical",
			wantSink:   "strategy.close",
		},
		{
			name:       "entry comment is observable only",
			call:       call(member("strategy", "entry"), objectArg("comment", identifier("x"))),
			wantImpact: "observable-non-backtest",
		},
		{
			name:       "exit profit comment is observable only",
			call:       call(member("strategy", "exit"), objectArg("comment_profit", identifier("x"))),
			wantImpact: "observable-non-backtest",
		},
		{
			name: "close alert message is not a comment parameter",
			call: call(member("strategy", "close"), objectArg("alert_message", identifier("x"))),
		},
		{
			name: "close all disable alert is not a comment parameter",
			call: call(member("strategy", "close_all"), objectArg("disable_alert", identifier("x"))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzeCompatibility(compatibilityProgram(tt.call), []string{"unsupported_entry_source"})
			if tt.wantImpact == "" {
				if len(got) != 0 {
					t.Fatalf("diagnostics = %+v, want no record for a non-comment argument", got)
				}
				return
			}
			assertCompatibilityImpact(t, got, "unsupported_entry_source", tt.wantImpact, tt.wantSink)
			if len(got) != 1 {
				t.Fatalf("diagnostics = %+v, want exactly one record", got)
			}
		})
	}
}

func TestAnalyzeCompatibility_NonBacktestConsumersRemainObservableOnly(t *testing.T) {
	program := &ast.Program{Body: []ast.Node{
		compatVarDecl("x", call(identifier("unsupported_entry_source"), identifier("close"))),
		&ast.ExpressionStatement{Expression: call(identifier("plot"), identifier("x"))},
	}}

	if got := analyzeCompatibility(program, []string{"unsupported_entry_source"}); len(got) != 0 {
		t.Fatalf("diagnostics = %+v, want no static backtest-critical diagnostics", got)
	}
}

func TestAnalyzeCompatibility_ChartHandleProducersOnlyReportWhenGetterConsumesHandle(t *testing.T) {
	tests := []struct {
		name         string
		body         []ast.Node
		wantFeatures []string
	}{
		{
			name: "display-only constructor remains silent",
			body: []ast.Node{
				compatVarDecl("line_handle", chartCall("line", "new", identifier("bar_index"), identifier("close"))),
				&ast.ExpressionStatement{Expression: call(identifier("plot"), identifier("close"))},
			},
		},
		{
			name: "getter consuming constructor handle reaches entry",
			body: []ast.Node{
				compatVarDecl("line_handle", chartCall("line", "new", identifier("bar_index"), identifier("close"))),
				compatVarDecl("price", chartCall("line", "get_price", identifier("line_handle"), identifier("bar_index"))),
				&ast.IfStatement{
					Test:       identifier("price"),
					Consequent: []ast.Node{&ast.ExpressionStatement{Expression: strategyEntryCall(identifier("close"))}},
				},
			},
			wantFeatures: []string{"line.get_price", "line.new"},
		},
		{
			name: "alias consumed by getter preserves originating constructor",
			body: []ast.Node{
				compatVarDecl("source_handle", chartCall("box", "new", identifier("bar_index"), identifier("close"))),
				compatVarDecl("alias_handle", identifier("source_handle")),
				compatVarDecl("top", chartCall("box", "get_top", identifier("alias_handle"))),
				&ast.IfStatement{
					Test:       identifier("top"),
					Consequent: []ast.Node{&ast.ExpressionStatement{Expression: strategyEntryCall(identifier("close"))}},
				},
			},
			wantFeatures: []string{"box.get_top", "box.new"},
		},
		{
			name: "handle controlling strategy branch is calculation bearing",
			body: []ast.Node{
				compatVarDecl("line_handle", chartCall("line", "new", identifier("bar_index"), identifier("close"))),
				&ast.IfStatement{
					Test:       identifier("line_handle"),
					Consequent: []ast.Node{&ast.ExpressionStatement{Expression: strategyEntryCall(identifier("close"))}},
				},
			},
			wantFeatures: []string{"line.new"},
		},
		{
			name: "inline handle producer controlling strategy branch is calculation bearing",
			body: []ast.Node{
				&ast.IfStatement{
					Test:       chartCall("line", "new", identifier("bar_index"), identifier("close")),
					Consequent: []ast.Node{&ast.ExpressionStatement{Expression: strategyEntryCall(identifier("close"))}},
				},
			},
			wantFeatures: []string{"line.new"},
		},
		{
			name: "handle in critical strategy argument is calculation bearing",
			body: []ast.Node{
				compatVarDecl("line_handle", chartCall("line", "new", identifier("bar_index"), identifier("close"))),
				&ast.ExpressionStatement{Expression: call(member("strategy", "entry"),
					literal("L"),
					member("strategy", "long"),
					objectArg("qty", identifier("line_handle")),
				)},
			},
			wantFeatures: []string{"line.new"},
		},
		{
			name: "handle in strategy comment remains presentation only",
			body: []ast.Node{
				compatVarDecl("line_handle", chartCall("line", "new", identifier("bar_index"), identifier("close"))),
				&ast.ExpressionStatement{Expression: call(member("strategy", "entry"),
					literal("L"),
					member("strategy", "long"),
					objectArg("comment", identifier("line_handle")),
				)},
			},
		},
		{
			name: "handle in for boundary controlling strategy branch is calculation bearing",
			body: []ast.Node{
				compatVarDecl("line_handle", chartCall("line", "new", identifier("bar_index"), identifier("close"))),
				&ast.ForStatement{
					Counter: "i",
					From:    literal(0.0),
					To:      identifier("line_handle"),
					Body:    []ast.Node{&ast.ExpressionStatement{Expression: strategyEntryCall(identifier("close"))}},
				},
			},
			wantFeatures: []string{"line.new"},
		},
		{
			name: "handle in while condition controlling strategy branch is calculation bearing",
			body: []ast.Node{
				compatVarDecl("line_handle", chartCall("line", "new", identifier("bar_index"), identifier("close"))),
				&ast.WhileStatement{
					Condition: identifier("line_handle"),
					Body:      []ast.Node{&ast.ExpressionStatement{Expression: strategyEntryCall(identifier("close"))}},
				},
			},
			wantFeatures: []string{"line.new"},
		},
		{
			name: "handle in for-in collection controlling strategy branch is calculation bearing",
			body: []ast.Node{
				compatVarDecl("line_handle", chartCall("line", "new", identifier("bar_index"), identifier("close"))),
				&ast.ForInStatement{
					ElementVar: "item",
					Collection: identifier("line_handle"),
					Body:       []ast.Node{&ast.ExpressionStatement{Expression: strategyEntryCall(identifier("close"))}},
				},
			},
			wantFeatures: []string{"line.new"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzeCompatibility(&ast.Program{Body: tt.body}, nil)
			if len(tt.wantFeatures) == 0 {
				if len(got) != 0 {
					t.Fatalf("diagnostics = %+v, want display-only chart handle to stay silent", got)
				}
				return
			}
			for _, feature := range tt.wantFeatures {
				assertCompatibilityDiagnostic(t, got, feature, "strategy.entry")
			}
			if len(got) != len(tt.wantFeatures) {
				t.Fatalf("diagnostics = %+v, want exactly %v", got, tt.wantFeatures)
			}
		})
	}
}

func TestAnalyzeCompatibility_InputGapIsNotDependencyTracked(t *testing.T) {
	program := &ast.Program{Body: []ast.Node{
		compatVarDecl("x", call(identifier("input"), literal(1.0))),
		&ast.ExpressionStatement{Expression: strategyEntryCall(identifier("x"))},
	}}

	if got := analyzeCompatibility(program, []string{"input"}); len(got) != 0 {
		t.Fatalf("diagnostics = %+v, want input gap excluded", got)
	}
}

func TestCompatibilitySetupCode_StableAndEscaped(t *testing.T) {
	diagnostics := []compatibilityDiagnostic{
		{
			FeatureID:    `custom"gap`,
			Source:       "generator.variable_init_unknown",
			Location:     ast.SourceLocation{File: "custom.pine", Line: 7, Column: 3},
			Phase:        "codegen",
			Impact:       "backtest-critical",
			Substitution: "NaN",
			Sinks:        []string{"strategy.close", "strategy.entry"},
		},
	}

	got := compatibilitySetupCode(diagnostics)
	for _, want := range []string{
		`featuregap.RecordStaticAt("custom\"gap", "generator.variable_init_unknown", "custom.pine", 7, 3, "codegen", "backtest-critical", "NaN", []string{"strategy.close", "strategy.entry"})`,
		"\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("compatibilitySetupCode() missing %q in %q", want, got)
		}
	}
}

func TestAnalyzeCompatibility_UnsupportedStatementRecordsObservableDiagnostic(t *testing.T) {
	program := &ast.Program{Body: []ast.Node{
		&ast.ExpressionStatement{Expression: callWithLoc(
			ast.SourceLocation{File: "statement_gap.pine", Line: 3, Column: 1},
			identifier("unsupported_statement"),
			identifier("close"),
		)},
	}}

	got := analyzeCompatibility(program, []string{"unsupported_statement"})
	if len(got) != 1 {
		t.Fatalf("diagnostics = %+v, want one statement diagnostic", got)
	}
	diag := got[0]
	if diag.FeatureID != "unsupported_statement" || diag.Impact != "observable-non-backtest" {
		t.Fatalf("diagnostic = %+v, want observable unsupported_statement", diag)
	}
	if diag.Location.File != "statement_gap.pine" || diag.Location.Line != 3 || diag.Location.Column != 1 {
		t.Fatalf("location = %+v, want statement_gap.pine:3:1", diag.Location)
	}
}

func TestAnalyzeCompatibility_UDFReturnTaintReachesBacktestSink(t *testing.T) {
	program := &ast.Program{Body: []ast.Node{
		compatUDF("f", []string{"v"}, &ast.ExpressionStatement{Expression: callWithLoc(
			ast.SourceLocation{File: "udf_gap.pine", Line: 3, Column: 9},
			identifier("unsupported_entry_source"),
			identifier("v"),
		)}),
		compatVarDecl("x", call(identifier("f"), identifier("close"))),
		&ast.IfStatement{
			Test:       identifier("x"),
			Consequent: []ast.Node{&ast.ExpressionStatement{Expression: strategyEntryCall(identifier("x"))}},
		},
	}}

	got := analyzeCompatibility(program, []string{"unsupported_entry_source"})
	assertCompatibilityDiagnostic(t, got, "unsupported_entry_source", "strategy.entry")
	if len(got) != 1 {
		t.Fatalf("diagnostics = %+v, want one UDF-mediated critical diagnostic", got)
	}
	if got[0].Source != "arrow_expression" {
		t.Fatalf("source = %q, want arrow_expression", got[0].Source)
	}
	if got[0].Location.Line != 3 || got[0].Location.Column != 9 {
		t.Fatalf("location = %+v, want udf body call location", got[0].Location)
	}
}

func TestAnalyzeCompatibility_AllControlFlowGapsReachNestedSink(t *testing.T) {
	program := &ast.Program{Body: []ast.Node{
		compatVarDecl("a", call(identifier("gap_a"), identifier("close"))),
		compatVarDecl("b", call(identifier("gap_b"), identifier("close"))),
		&ast.IfStatement{
			Test: logical(
				call(identifier("nz"), identifier("a"), literal(1.0)),
				call(identifier("nz"), identifier("b"), literal(1.0)),
			),
			Consequent: []ast.Node{&ast.ExpressionStatement{Expression: strategyEntryCall(identifier("a"))}},
		},
	}}

	got := analyzeCompatibility(program, []string{"gap_a", "gap_b"})
	assertCompatibilityDiagnostic(t, got, "gap_a", "strategy.entry")
	assertCompatibilityDiagnostic(t, got, "gap_b", "strategy.entry")
	if len(got) != 2 {
		t.Fatalf("diagnostics = %+v, want both control-flow gaps", got)
	}
}

func TestAnalyzeCompatibility_InputFalseConditionalPrunesUnreachedUDFGap(t *testing.T) {
	program := &ast.Program{Body: []ast.Node{
		compatVarDecl("enabled", call(identifier("input"), literal(false))),
		compatUDF("f", []string{"v"}, &ast.ExpressionStatement{Expression: callWithLoc(
			ast.SourceLocation{File: "session.pine", Line: 8, Column: 18},
			identifier("unsupported_entry_source"),
			identifier("v"),
		)}),
		compatVarDecl("x", call(identifier("f"), identifier("close"))),
		compatVarDecl("gate", conditional(identifier("enabled"), identifier("x"), literal(true))),
		&ast.IfStatement{
			Test:       identifier("gate"),
			Consequent: []ast.Node{&ast.ExpressionStatement{Expression: strategyEntryCall(identifier("close"))}},
		},
	}}

	if got := analyzeCompatibility(program, []string{"unsupported_entry_source"}); len(got) != 0 {
		t.Fatalf("diagnostics = %+v, want no dependency through statically false input branch", got)
	}
}

func TestAnalyzeCompatibility_ConstantConditionalBranches(t *testing.T) {
	tests := []struct {
		name         string
		enabled      ast.Expression
		wantCritical bool
	}{
		{
			name:         "positional input true reaches consequent",
			enabled:      call(identifier("input"), literal(true)),
			wantCritical: true,
		},
		{
			name:         "named defval false reaches alternate",
			enabled:      call(identifier("input"), objectArg("defval", literal(false))),
			wantCritical: false,
		},
		{
			name:         "literal true reaches consequent",
			enabled:      literal(true),
			wantCritical: true,
		},
		{
			name:         "literal false reaches alternate",
			enabled:      literal(false),
			wantCritical: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := &ast.Program{Body: []ast.Node{
				compatVarDecl("enabled", tt.enabled),
				compatVarDecl("x", call(identifier("unsupported_entry_source"), identifier("close"))),
				compatVarDecl("gate", conditional(identifier("enabled"), identifier("x"), literal(true))),
				&ast.IfStatement{
					Test:       identifier("gate"),
					Consequent: []ast.Node{&ast.ExpressionStatement{Expression: strategyEntryCall(identifier("close"))}},
				},
			}}

			got := analyzeCompatibility(program, []string{"unsupported_entry_source"})
			if tt.wantCritical {
				assertCompatibilityDiagnostic(t, got, "unsupported_entry_source", "strategy.entry")
				return
			}
			if len(got) != 0 {
				t.Fatalf("diagnostics = %+v, want no dependency through unreachable branch", got)
			}
		})
	}
}

func TestAnalyzeCompatibility_ConstantLogicalBranches(t *testing.T) {
	tests := []struct {
		name         string
		test         ast.Expression
		wantCritical bool
	}{
		{
			name:         "false and tainted is independent of tainted right side",
			test:         logical(literal(false), identifier("x")),
			wantCritical: false,
		},
		{
			name:         "true or tainted is independent of tainted right side",
			test:         logicalOr(literal(true), identifier("x")),
			wantCritical: false,
		},
		{
			name:         "true and tainted depends on tainted right side",
			test:         logical(literal(true), identifier("x")),
			wantCritical: true,
		},
		{
			name:         "false or tainted depends on tainted right side",
			test:         logicalOr(literal(false), identifier("x")),
			wantCritical: true,
		},
		{
			name:         "tainted and false is independent of tainted left side",
			test:         logical(identifier("x"), literal(false)),
			wantCritical: false,
		},
		{
			name:         "tainted or true is independent of tainted left side",
			test:         logicalOr(identifier("x"), literal(true)),
			wantCritical: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := &ast.Program{Body: []ast.Node{
				compatVarDecl("x", call(identifier("unsupported_entry_source"), identifier("close"))),
				&ast.IfStatement{
					Test:       tt.test,
					Consequent: []ast.Node{&ast.ExpressionStatement{Expression: strategyEntryCall(identifier("close"))}},
				},
			}}

			got := analyzeCompatibility(program, []string{"unsupported_entry_source"})
			if tt.wantCritical {
				assertCompatibilityDiagnostic(t, got, "unsupported_entry_source", "strategy.entry")
				return
			}
			if len(got) != 0 {
				t.Fatalf("diagnostics = %+v, want no dependency through constant logical result", got)
			}
		})
	}
}

func TestAnalyzeCompatibility_UnknownConditionalKeepsUDFGapConservative(t *testing.T) {
	program := &ast.Program{Body: []ast.Node{
		compatUDF("f", []string{"v"}, &ast.ExpressionStatement{Expression: callWithLoc(
			ast.SourceLocation{File: "session.pine", Line: 8, Column: 18},
			identifier("unsupported_entry_source"),
			identifier("v"),
		)}),
		compatVarDecl("x", call(identifier("f"), identifier("close"))),
		compatVarDecl("gate", conditional(identifier("enabled"), identifier("x"), literal(true))),
		&ast.IfStatement{
			Test:       identifier("gate"),
			Consequent: []ast.Node{&ast.ExpressionStatement{Expression: strategyEntryCall(identifier("close"))}},
		},
	}}

	got := analyzeCompatibility(program, []string{"unsupported_entry_source"})
	assertCompatibilityDiagnostic(t, got, "unsupported_entry_source", "strategy.entry")
}

func TestAnalyzeCompatibility_ConflictingBoolAssignmentsAreNotConstants(t *testing.T) {
	program := &ast.Program{Body: []ast.Node{
		compatVarDecl("enabled", literal(false)),
		compatVarDecl("enabled", literal(true)),
		compatVarDecl("x", call(identifier("unsupported_entry_source"), identifier("close"))),
		compatVarDecl("gate", conditional(identifier("enabled"), identifier("x"), literal(true))),
		&ast.IfStatement{
			Test:       identifier("gate"),
			Consequent: []ast.Node{&ast.ExpressionStatement{Expression: strategyEntryCall(identifier("close"))}},
		},
	}}

	got := analyzeCompatibility(program, []string{"unsupported_entry_source"})
	assertCompatibilityDiagnostic(t, got, "unsupported_entry_source", "strategy.entry")
}

func assertCompatibilityDiagnostic(t *testing.T, got []compatibilityDiagnostic, featureID, sink string) {
	t.Helper()
	assertCompatibilityImpact(t, got, featureID, "backtest-critical", sink)
}

func assertCompatibilityImpact(t *testing.T, got []compatibilityDiagnostic, featureID, impact, sink string) {
	t.Helper()
	for _, diag := range got {
		if diag.FeatureID != featureID || diag.Impact != impact {
			continue
		}
		if sink == "" || containsString(diag.Sinks, sink) {
			return
		}
	}
	t.Fatalf("missing feature=%s impact=%s sink=%s in %+v", featureID, impact, sink, got)
}

func compatibilityProgram(sinkCall *ast.CallExpression) *ast.Program {
	return &ast.Program{Body: []ast.Node{
		compatVarDecl("x", call(identifier("unsupported_entry_source"), identifier("close"))),
		&ast.ExpressionStatement{Expression: sinkCall},
	}}
}

func strategyEntryCall(condition ast.Expression) *ast.CallExpression {
	return call(member("strategy", "entry"),
		literal("L"),
		member("strategy", "long"),
		objectArg("when", condition),
	)
}

func chartGetterCall(namespace, getter string) *ast.CallExpression {
	return call(member(namespace, getter), identifier("handle"), identifier("bar_index"))
}

func chartCall(namespace, method string, args ...ast.Expression) *ast.CallExpression {
	return call(member(namespace, method), args...)
}

func compatVarDecl(name string, init ast.Expression) *ast.VariableDeclaration {
	return &ast.VariableDeclaration{
		Declarations: []ast.VariableDeclarator{{
			ID:   identifier(name),
			Init: init,
		}},
	}
}

func call(callee ast.Expression, args ...ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee:    callee,
		Arguments: args,
	}
}

func callWithLoc(location ast.SourceLocation, callee ast.Expression, args ...ast.Expression) *ast.CallExpression {
	call := call(callee, args...)
	call.Location = location
	return call
}

func logical(left, right ast.Expression) *ast.LogicalExpression {
	return &ast.LogicalExpression{
		Operator: "&&",
		Left:     left,
		Right:    right,
	}
}

func logicalOr(left, right ast.Expression) *ast.LogicalExpression {
	return &ast.LogicalExpression{
		Operator: "||",
		Left:     left,
		Right:    right,
	}
}

func conditional(test, consequent, alternate ast.Expression) *ast.ConditionalExpression {
	return &ast.ConditionalExpression{
		Test:       test,
		Consequent: consequent,
		Alternate:  alternate,
	}
}

func compatUDF(name string, params []string, body ...ast.Node) *ast.VariableDeclaration {
	return &ast.VariableDeclaration{
		Declarations: []ast.VariableDeclarator{{
			ID: identifier(name),
			Init: &ast.ArrowFunctionExpression{
				Params: buildTestIdentifiers(params),
				Body:   body,
			},
		}},
	}
}

func buildTestIdentifiers(names []string) []ast.Identifier {
	out := make([]ast.Identifier, 0, len(names))
	for _, name := range names {
		out = append(out, ast.Identifier{Name: name})
	}
	return out
}

func member(object, property string) *ast.MemberExpression {
	return &ast.MemberExpression{
		Object:   identifier(object),
		Property: identifier(property),
	}
}

func identifier(name string) *ast.Identifier {
	return &ast.Identifier{Name: name}
}

func literal(value interface{}) *ast.Literal {
	return &ast.Literal{Value: value}
}

func objectArg(key string, value ast.Expression) *ast.ObjectExpression {
	return &ast.ObjectExpression{
		Properties: []ast.Property{{
			Key:   identifier(key),
			Value: value,
		}},
	}
}
