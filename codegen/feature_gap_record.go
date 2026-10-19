package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/*
runtimeFeatureGap emits the runtime call that records one observable feature gap
and yields the numeric sentinel in expression position. The source position is
carried so the resulting diagnostic is source-located instead of anonymous.
*/
func runtimeFeatureGap(funcName, source string, location ast.SourceLocation) string {
	return fmt.Sprintf("featuregap.Record(%q, %q, %q, %d, %d, ctx.BarIndex)",
		funcName, source, location.File, location.Line, location.Column)
}

/*
runtimeFeatureGapString is the string-typed counterpart, used where the
substituted annotation must yield a string sentinel.
*/
func runtimeFeatureGapString(funcName, source string, location ast.SourceLocation) string {
	return fmt.Sprintf("featuregap.RecordString(%q, %q, %q, %d, %d, ctx.BarIndex)",
		funcName, source, location.File, location.Line, location.Column)
}
