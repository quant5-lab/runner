package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func annotationGapCall(name string) *ast.CallExpression {
	return callWithLoc(annotationGapLocation(), identifier(name), identifier("close"))
}

func annotationGapLocation() ast.SourceLocation {
	return ast.SourceLocation{File: "annotation.pine", Line: 4, Column: 12}
}

func strategyEntryWithComment(value ast.Expression) *ast.CallExpression {
	return call(member("strategy", "entry"), literal("L"), member("strategy", "long"), objectArg("comment", value))
}

func assertSingleLocatedObservableAnnotation(t *testing.T, got []compatibilityDiagnostic, featureID string) {
	t.Helper()
	if len(got) != 1 {
		t.Fatalf("diagnostics = %+v, want exactly one annotation diagnostic", got)
	}
	diag := got[0]
	if diag.FeatureID != featureID || diag.Impact != "observable-non-backtest" || diag.Substitution != "empty-string" {
		t.Fatalf("diagnostic = %+v, want observable empty-string %s", diag, featureID)
	}
	if diag.Location != annotationGapLocation() {
		t.Fatalf("location = %+v, want %+v", diag.Location, annotationGapLocation())
	}
}

func TestAnalyzeCompatibility_AnnotationGapInUnreachedBranchIsReportedAndLocated(t *testing.T) {
	program := &ast.Program{Body: []ast.Node{
		compatVarDecl("msg", annotationGapCall("unsupported_annotation")),
		&ast.IfStatement{
			Test:       identifier("close"),
			Consequent: []ast.Node{&ast.ExpressionStatement{Expression: strategyEntryWithComment(identifier("msg"))}},
		},
	}}

	assertSingleLocatedObservableAnnotation(t, analyzeCompatibility(program, []string{"unsupported_annotation"}), "unsupported_annotation")
}

func TestAnalyzeCompatibility_DirectAnnotationGapIsReportedAndLocated(t *testing.T) {
	program := &ast.Program{Body: []ast.Node{
		&ast.ExpressionStatement{Expression: strategyEntryWithComment(annotationGapCall("unsupported_annotation"))},
	}}

	assertSingleLocatedObservableAnnotation(t, analyzeCompatibility(program, []string{"unsupported_annotation"}), "unsupported_annotation")
}

func TestAnalyzeCompatibility_AnnotationGapThatAlsoFeedsCriticalArgumentStaysCritical(t *testing.T) {
	program := &ast.Program{Body: []ast.Node{
		compatVarDecl("msg", annotationGapCall("unsupported_annotation")),
		&ast.ExpressionStatement{Expression: call(member("strategy", "entry"),
			literal("L"),
			member("strategy", "long"),
			strategyNamedArgs(
				strategyNamedProp("qty", identifier("msg")),
				strategyNamedProp("comment", identifier("msg")),
			),
		)},
	}}

	got := analyzeCompatibility(program, []string{"unsupported_annotation"})
	if len(got) != 1 {
		t.Fatalf("diagnostics = %+v, want a single critical record", got)
	}
	if got[0].Impact != "backtest-critical" || !containsString(got[0].Sinks, "strategy.entry") {
		t.Fatalf("diagnostic = %+v, want backtest-critical on strategy.entry", got[0])
	}
}

func exitCallWithAnnotation(parameter, shape string) *ast.CallExpression {
	annotation := annotationGapCall("unsupported_annotation")
	if shape == "named" {
		return exitCallWithNamedComment(ast.SourceLocation{}, parameter, annotation)
	}
	return exitCallWithPositionalComment(ast.SourceLocation{}, parameter, annotation)
}

func TestAnalyzeCompatibility_EveryExitCommentParameterIsReportedInEveryCallShape(t *testing.T) {
	for _, parameter := range []string{"comment", "comment_profit", "comment_loss", "comment_trailing"} {
		for _, shape := range []string{"named", "positional"} {
			t.Run(parameter+"/"+shape, func(t *testing.T) {
				program := &ast.Program{Body: []ast.Node{
					&ast.ExpressionStatement{Expression: exitCallWithAnnotation(parameter, shape)},
				}}
				assertSingleLocatedObservableAnnotation(t, analyzeCompatibility(program, []string{"unsupported_annotation"}), "unsupported_annotation")
			})
		}
	}
}

func TestAnalyzeCompatibility_CommentOnNonStrategyCallIsNotRecorded(t *testing.T) {
	program := &ast.Program{Body: []ast.Node{
		compatVarDecl("msg", annotationGapCall("unsupported_annotation")),
		&ast.ExpressionStatement{Expression: call(identifier("plot"), identifier("close"), objectArg("comment", identifier("msg")))},
	}}

	if got := analyzeCompatibility(program, []string{"unsupported_annotation"}); len(got) != 0 {
		t.Fatalf("diagnostics = %+v, want the annotation sink closed outside strategy calls", got)
	}
}

func TestAnalyzeCompatibility_SupportedCommentLeavesNoDiagnostic(t *testing.T) {
	program := &ast.Program{Body: []ast.Node{
		compatVarDecl("msg", literal("carried")),
		&ast.ExpressionStatement{Expression: strategyEntryWithComment(identifier("msg"))},
	}}

	if got := analyzeCompatibility(program, []string{"unsupported_annotation"}); len(got) != 0 {
		t.Fatalf("diagnostics = %+v, want a carried comment to stay silent", got)
	}
}

// The runtime carries only the primary "comment" parameter into the generated
// result, so each secondary comment variant is reported at the strategy call's
// location in both call forms instead of being dropped silently.
func TestAnalyzeCompatibility_SecondaryCommentVariantIsReported(t *testing.T) {
	location := ast.SourceLocation{File: "annotation.pine", Line: 9, Column: 3}
	forms := map[string]func(parameter string, value ast.Expression) *ast.CallExpression{
		"named": func(parameter string, value ast.Expression) *ast.CallExpression {
			return exitCallWithNamedComment(location, parameter, value)
		},
		"positional": func(parameter string, value ast.Expression) *ast.CallExpression {
			return exitCallWithPositionalComment(location, parameter, value)
		},
	}

	for _, variant := range []string{"comment_profit", "comment_loss", "comment_trailing"} {
		for _, form := range []string{"named", "positional"} {
			t.Run(variant+"/"+form, func(t *testing.T) {
				assertSecondaryCommentRecord(t, forms[form](variant, literal("tag")), variant, location)
			})
		}
	}

	t.Run("string na is a real annotation", func(t *testing.T) {
		assertSecondaryCommentRecord(t, exitCallWithNamedComment(location, "comment_profit", literal("na")), "comment_profit", location)
	})

	t.Run("variable reference is a real annotation", func(t *testing.T) {
		assertSecondaryCommentRecord(t, exitCallWithNamedComment(location, "comment_profit", identifier("msg")), "comment_profit", location)
	})
}

// An argument that is not written at all, or written as Pine's na marker, is an
// absent annotation rather than an unrepresentable one.
func TestAnalyzeCompatibility_AbsentCommentArgumentIsNotReported(t *testing.T) {
	location := ast.SourceLocation{File: "annotation.pine", Line: 9, Column: 3}
	tests := []struct {
		name string
		call *ast.CallExpression
	}{
		{
			name: "only the primary comment is written",
			call: exitCallWithNamedComment(location, "comment", literal("base")),
		},
		{
			name: "na positional placeholders",
			call: exitCallWithNaCommentPlaceholders(location),
		},
		{
			name: "na named value",
			call: exitCallWithNamedComment(location, "comment_profit", naIdentifier()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := analyzeCompatibility(strategyExitProgram(tt.call), nil); len(got) != 0 {
				t.Fatalf("diagnostics = %+v, want no record for an absent comment argument", got)
			}
		})
	}
}

func strategyExitProgram(call *ast.CallExpression) *ast.Program {
	return &ast.Program{Body: []ast.Node{&ast.ExpressionStatement{Expression: call}}}
}

func exitCallWithNamedComment(location ast.SourceLocation, parameter string, value ast.Expression) *ast.CallExpression {
	return callWithLoc(location, member("strategy", "exit"),
		literal("X"), literal("L"),
		strategyNamedArgs(strategyNamedProp(parameter, value)))
}

// exitCallWithPositionalComment writes the annotation at its real Pine slot and
// fills every earlier argument with na, the shape a positional Pine call has.
func exitCallWithPositionalComment(location ast.SourceLocation, parameter string, value ast.Expression) *ast.CallExpression {
	slot := strategyAnnotationIndex("strategy.exit", parameter)
	args := make([]ast.Expression, slot+1)
	for i := range args {
		args[i] = naIdentifier()
	}
	args[0] = literal("X")
	args[1] = literal("L")
	args[slot] = value
	return callWithLoc(location, member("strategy", "exit"), args...)
}

// exitCallWithNaCommentPlaceholders keeps the primary comment and writes na into
// every secondary comment slot, the shape a caller leaves behind when it omits
// those optional annotations positionally.
func exitCallWithNaCommentPlaceholders(location ast.SourceLocation) *ast.CallExpression {
	call := exitCallWithPositionalComment(location, "comment_trailing", naIdentifier())
	call.Arguments[strategyAnnotationIndex("strategy.exit", "comment")] = literal("base")
	return call
}

func assertSecondaryCommentRecord(t *testing.T, call *ast.CallExpression, parameter string, location ast.SourceLocation) {
	t.Helper()
	got := analyzeCompatibility(strategyExitProgram(call), nil)
	if len(got) != 1 {
		t.Fatalf("diagnostics = %+v, want exactly one record for %s", got, parameter)
	}
	diag := got[0]
	wantID := "strategy.exit." + parameter
	if diag.FeatureID != wantID || diag.Source != commentParameterSource ||
		diag.Impact != "observable-non-backtest" || diag.Substitution != "empty-string" {
		t.Fatalf("diagnostic = %+v, want observable empty-string %s", diag, wantID)
	}
	if diag.Location != location {
		t.Fatalf("%s location = %+v, want %+v", parameter, diag.Location, location)
	}
}
