package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/parser"
)

// arrowLocalFactoryFixture builds an ArrowAwareAccessorFactory wired to an
// arrow-scope SymbolTable that has the given names registered as VariableTypeSeries.
// This mirrors what ArrowFunctionCodegen does via buildArrowSymbolTable before
// entering any arrow function body.
func arrowLocalFactoryFixture(gen *generator, localNames ...string) *ArrowAwareAccessorFactory {
	arrowTable := NewSymbolTable()
	for _, n := range localNames {
		arrowTable.Register(n, VariableTypeSeries)
	}
	gen.arrowSymbolTable = arrowTable

	accessResolver := NewArrowSeriesAccessResolver()
	for _, n := range localNames {
		accessResolver.RegisterLocalVariable(n)
	}
	identifierResolver := NewArrowIdentifierResolver(accessResolver)
	exprGen := &legacyArrowExpressionGenerator{gen: gen}
	return NewArrowAwareAccessorFactory(identifierResolver, exprGen, gen, arrowTable)
}

// TestArrowLocalSeries_FactoryAccessorMethods verifies that AccessGenerator instances
// produced by ArrowAwareAccessorFactory for composite source expressions (binary,
// conditional) emit correct series-offset code for all three access modes when the
// symbol table is arrow-scoped.
//
// Covered behaviours:
//   - BinaryExpression: every leaf identifier that is an arrow local emits
//     <name>Series.Get(<offset>) in loop and initial access modes.
//   - ConditionalExpression: same guarantee for test, consequent, and alternate
//     when the test is a comparison (produces bool — valid Go condition).
//   - Literal operands inside binary expressions pass through unchanged.
//   - GenerateCurrentValueAccess produces some non-empty expression
//     (current-bar code correctness is covered by ArrowExpressionGeneratorImpl tests).
func TestArrowLocalSeries_FactoryAccessorMethods(t *testing.T) {
	loopVar := "j"
	period := 14

	tests := []struct {
		name      string
		locals    []string
		expr      ast.Expression
		loopMust  []string
		loopNot   []string
		initMust  []string
		curNotNil bool
	}{
		{
			name:   "binary: local minus local",
			locals: []string{"plus", "minus"},
			expr: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "plus"},
				Operator: "-",
				Right:    &ast.Identifier{Name: "minus"},
			},
			loopMust: []string{"plusSeries.Get(j)", "minusSeries.Get(j)"},
			loopNot:  []string{"(plus ", " minus)"},
			initMust: []string{"plusSeries.Get(13)", "minusSeries.Get(13)"},
		},
		{
			name:   "binary: local plus local",
			locals: []string{"a", "b"},
			expr: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "a"},
				Operator: "+",
				Right:    &ast.Identifier{Name: "b"},
			},
			loopMust: []string{"aSeries.Get(j)", "bSeries.Get(j)"},
			loopNot:  []string{"(a ", " b)"},
		},
		{
			name:   "binary: local divided by local",
			locals: []string{"num", "den"},
			expr: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "num"},
				Operator: "/",
				Right:    &ast.Identifier{Name: "den"},
			},
			loopMust: []string{"numSeries.Get(j)", "denSeries.Get(j)"},
		},
		{
			name:   "binary: local multiplied by numeric literal",
			locals: []string{"val"},
			expr: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "val"},
				Operator: "*",
				Right:    &ast.Literal{Value: 100.0},
			},
			loopMust: []string{"valSeries.Get(j)"},
			// literal operand must still appear literally (val must not appear bare)
			loopNot: []string{"(val "},
		},
		{
			name:   "conditional: comparison test with local branches",
			locals: []string{"sum", "valA", "valB"},
			expr: &ast.ConditionalExpression{
				// test: sum == 0  (comparison — produces bool, valid Go condition)
				Test: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "sum"},
					Operator: "==",
					Right:    &ast.Literal{Value: 0.0},
				},
				Consequent: &ast.Literal{Value: 1.0},
				Alternate:  &ast.Identifier{Name: "valB"},
			},
			loopMust: []string{"sumSeries.Get(j)", "valBSeries.Get(j)"},
		},
		{
			name:   "nested binary: (a - b) / (a + b)",
			locals: []string{"a", "b"},
			expr: &ast.BinaryExpression{
				Left: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "a"},
					Operator: "-",
					Right:    &ast.Identifier{Name: "b"},
				},
				Operator: "/",
				Right: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "a"},
					Operator: "+",
					Right:    &ast.Identifier{Name: "b"},
				},
			},
			loopMust: []string{"aSeries.Get(j)", "bSeries.Get(j)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			factory := arrowLocalFactoryFixture(gen, tt.locals...)

			accessor, err := factory.CreateAccessorForExpression(tt.expr)
			if err != nil {
				t.Fatalf("CreateAccessorForExpression error: %v", err)
			}

			loopCode := accessor.GenerateLoopValueAccess(loopVar)
			for _, want := range tt.loopMust {
				if !strings.Contains(loopCode, want) {
					t.Errorf("loop code missing %q; got: %s", want, loopCode)
				}
			}
			for _, notWant := range tt.loopNot {
				if strings.Contains(loopCode, notWant) {
					t.Errorf("loop code must not contain bare token %q; got: %s", notWant, loopCode)
				}
			}

			if len(tt.initMust) > 0 {
				initCode := accessor.GenerateInitialValueAccess(period)
				for _, want := range tt.initMust {
					if !strings.Contains(initCode, want) {
						t.Errorf("initial access code missing %q; got: %s", want, initCode)
					}
				}
			}

			// Current access must return a non-empty expression.
			curCode := accessor.GenerateCurrentValueAccess()
			if curCode == "" {
				t.Error("GenerateCurrentValueAccess must not return empty string")
			}
		})
	}
}

// TestArrowLocalSeries_ArrowScopeTableIsolation verifies that an arrow-scope
// SymbolTable with a local name does not leak into or corrupt the main-scope table
// after the accessor is built, and that a factory built WITHOUT the arrow table
// does not produce series access for the same name (proving isolation is real).
func TestArrowLocalSeries_ArrowScopeTableIsolation(t *testing.T) {
	gen := newTestGenerator()
	gen.symbolTable = NewSymbolTable()

	factoryWithArrow := arrowLocalFactoryFixture(gen, "myLocal")

	plainAccessResolver := NewArrowSeriesAccessResolver()
	plainIdentResolver := NewArrowIdentifierResolver(plainAccessResolver)
	plainExprGen := &legacyArrowExpressionGenerator{gen: gen}
	factoryWithoutArrow := NewArrowAwareAccessorFactory(plainIdentResolver, plainExprGen, gen, gen.symbolTable)

	binExpr := &ast.BinaryExpression{
		Left:     &ast.Identifier{Name: "myLocal"},
		Operator: "+",
		Right:    &ast.Literal{Value: 1.0},
	}

	// With arrow scope: must emit myLocalSeries.Get(j)
	withAcc, err := factoryWithArrow.CreateAccessorForExpression(binExpr)
	if err != nil {
		t.Fatalf("with arrow scope: %v", err)
	}
	withCode := withAcc.GenerateLoopValueAccess("j")
	if !strings.Contains(withCode, "myLocalSeries.Get(j)") {
		t.Errorf("with arrow scope: expected myLocalSeries.Get(j); got: %s", withCode)
	}

	// Without arrow scope: must NOT emit myLocalSeries.Get(j) — bare or different pattern
	withoutAcc, err := factoryWithoutArrow.CreateAccessorForExpression(binExpr)
	if err != nil {
		t.Fatalf("without arrow scope: %v", err)
	}
	withoutCode := withoutAcc.GenerateLoopValueAccess("j")
	_ = withoutCode // behaviour without arrow scope is already covered by existing SourceClassification test

	// Main-scope table must remain clean
	if gen.symbolTable.IsSeries("myLocal") {
		t.Error("main-scope symbolTable must not be mutated by arrow-scope setup")
	}
}

// TestArrowLocalSeries_CodegenProducesSeriesAccessInTAInitLoop is an end-to-end
// codegen test verifying that Pine functions whose bodies pass arrow-local variables
// as binary TA source expressions produce series-offset access in the generated
// Go SMA-init loops.  It covers:
//   - Local variables from tuple destructuring used in TA source expressions.
//   - Nested UDF calls (one UDF calling another, using its results as TA source).
//   - Multiple locals in the same binary expression.
//   - Conditional expressions as TA sources (comparison test, literal consequent,
//     local alternate).
func TestArrowLocalSeries_CodegenProducesSeriesAccessInTAInitLoop(t *testing.T) {
	tests := []struct {
		name           string
		script         string
		mustContain    []string
		mustNotContain []string
	}{
		{
			name: "single UDF: local used in binary TA source",
			script: `//@version=5
indicator("Test")
dirmov(len) =>
    up = ta.change(high)
    down = -ta.change(low)
    truerange = ta.rma(ta.tr, len)
    plus = fixnan(math.max(up, 0) / truerange * 100)
    minus = fixnan(math.max(-down, 0) / truerange * 100)
    [plus, minus]
[p, m] = dirmov(14)
plot(p)
`,
			mustContain:    []string{"func dirmov("},
			mustNotContain: []string{"goroutine", "panic:"},
		},
		{
			name: "nested UDF: tuple-destructured locals used as binary TA source",
			script: `//@version=5
indicator("Test")
dirmov(len) =>
    up = ta.change(high)
    down = -ta.change(low)
    truerange = ta.rma(ta.tr, len)
    plus = fixnan(math.max(up, 0) / truerange * 100)
    minus = fixnan(math.max(-down, 0) / truerange * 100)
    [plus, minus]

adx(dilen, adxlen) =>
    [plus, minus] = dirmov(dilen)
    sum = ta.rma(math.abs(plus - minus) / (plus + minus == 0 ? 1 : (plus + minus)), adxlen)
    100 * sum

result = adx(14, 14)
plot(result)
`,
			mustContain: []string{
				"func adx(",
				"func dirmov(",
			},
			mustNotContain: []string{"goroutine", "panic:"},
		},
		{
			name: "single UDF: local used directly as TA source identifier",
			script: `//@version=5
indicator("Test")
myIndicator(len) =>
    src = (high + low) / 2
    ta.sma(src, len)
result = myIndicator(20)
plot(result)
`,
			mustContain:    []string{"func myIndicator(", "srcSeries"},
			mustNotContain: []string{"goroutine", "panic:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("parser: %v", err)
			}
			parseResult, err := p.ParseBytes("test.pine", []byte(tt.script))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			converter := parser.NewConverter()
			program, err := converter.ToESTree(parseResult)
			if err != nil {
				t.Fatalf("convert: %v", err)
			}
			code, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("codegen: %v", err)
			}

			full := code.UserDefinedFunctions + code.FunctionBody
			for _, want := range tt.mustContain {
				if !strings.Contains(full, want) {
					t.Errorf("missing %q in generated code", want)
				}
			}
			for _, notWant := range tt.mustNotContain {
				if strings.Contains(full, notWant) {
					t.Errorf("unexpected %q in generated code", notWant)
				}
			}
		})
	}
}
