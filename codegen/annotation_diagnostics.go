package codegen

import (
	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/featuregap"
)

// commentParameterSource tags a record for a secondary comment parameter the
// strategy runtime cannot carry into the generated result.
const commentParameterSource = "generator.comment_unrepresented"

/*
recordAnnotationDiagnostics records unsupported values that reach only a strategy
call's comment parameter, for a call the caller has already established is a
strategy execution sink. The record is observable, never critical, so a
substituted annotation cannot invalidate the backtest, and it is produced at
analysis time so it exists whether or not the call is reached at run time.
*/
func (a *compatibilityAnalyzer) recordAnnotationDiagnostics(call *ast.CallExpression, critical compatibilityFeatureSet, out map[string]compatibilityDiagnostic) {
	sink := extractCallFunctionName(call)
	for _, annotation := range strategyCommentExpressions(sink, call.Arguments) {
		for _, feature := range sortedFeatures(a.featuresInExpression(annotation, a.tainted, "generator.comment_unknown", nil)) {
			if _, alsoCritical := critical[feature.ID]; alsoCritical {
				continue
			}
			addDiagnostic(out, feature, "observable-non-backtest", featuregap.SubstitutionEmptyString, "")
		}
	}
	a.recordUncarriedCommentParameters(call, sink, out)
}

/*
recordUncarriedCommentParameters reports every secondary comment parameter
(comment_profit, comment_loss, comment_trailing) the runtime cannot carry into
the generated result, so an annotation is substituted with provenance instead of
being dropped silently. Only the primary "comment" parameter reaches the result.
An annotation whose expression already carries a feature gap is skipped, because
that gap is recorded by the feature pass above.
*/
func (a *compatibilityAnalyzer) recordUncarriedCommentParameters(call *ast.CallExpression, sink string, out map[string]compatibilityDiagnostic) {
	arguments := newStrategyCallArguments(call.Arguments)
	for _, parameter := range strategyCommentParameters {
		if parameter == "comment" {
			continue
		}
		expr, present := arguments.value(parameter, strategyAnnotationIndex(sink, parameter))
		if !present || isNaIdentifier(expr) {
			continue
		}
		if len(a.featuresInExpression(expr, a.tainted, "generator.comment_unknown", nil)) != 0 {
			continue
		}
		addDiagnostic(out, compatibilityFeature{
			ID:       sink + "." + parameter,
			Source:   commentParameterSource,
			Location: call.Location,
		}, "observable-non-backtest", featuregap.SubstitutionEmptyString, "")
	}
}

// isNaIdentifier reports whether an argument is Pine's absent-value marker, which
// a positional call writes into every comment slot the caller left unset.
func isNaIdentifier(expr ast.Expression) bool {
	identifier, ok := expr.(*ast.Identifier)
	return ok && identifier.Name == "na"
}
