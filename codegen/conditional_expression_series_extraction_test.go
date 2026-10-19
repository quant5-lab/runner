package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/*
TestConditionalExpressionWithUnsupportedOperandFailsLoudly guards the failure half of the argument
position lowering: an operand the condition generators cannot lower must surface as a marker that
stops the generated Go from compiling, never as a constant that silently disables the condition.
*/
func TestConditionalExpressionWithUnsupportedOperandFailsLoudly(t *testing.T) {
	gen := &generator{}

	result := gen.extractSeriesExpression(&ast.ConditionalExpression{
		Test:       &ast.ObjectExpression{},
		Consequent: literal(1.0),
		Alternate:  literal(0.0),
	})

	if !strings.Contains(result, "ERROR") {
		t.Errorf("expected a loud failure marker, got: %s", result)
	}
}
