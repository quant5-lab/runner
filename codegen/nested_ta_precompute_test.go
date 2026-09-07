package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/parser"
)

// arrowParamFactoryFixture builds an ArrowAwareAccessorFactory wired to an arrow scope
// where the given names are UDF parameters (not arrow-local series), so the factory's
// symbolTable is non-nil and createBinaryAccessor takes the precomputer path.
func arrowParamFactoryFixture(gen *generator, paramNames ...string) *ArrowAwareAccessorFactory {
	symTable := NewSymbolTable()
	gen.symbolTable = symTable

	accessResolver := NewArrowSeriesAccessResolver()
	for _, n := range paramNames {
		accessResolver.RegisterParameter(n)
	}
	identifierResolver := NewArrowIdentifierResolver(accessResolver)
	exprGen := &legacyArrowExpressionGenerator{gen: gen}
	return NewArrowAwareAccessorFactory(identifierResolver, exprGen, gen, symTable)
}

func buildTACall(funcName string, src, period ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee:    &ast.Identifier{Name: funcName},
		Arguments: []ast.Expression{src, period},
	}
}

// TestNestedTACallPrecomputer_Scan verifies that Process identifies TA calls embedded in
// every expression tree shape and ignores non-TA calls, literals, and plain identifiers.
func TestNestedTACallPrecomputer_Scan(t *testing.T) {
	src := &ast.Identifier{Name: "src"}
	period14 := &ast.Literal{Value: float64(14)}
	period28 := &ast.Literal{Value: float64(28)}

	wmaCall := buildTACall("wma", src, period14)
	smaCall := buildTACall("sma", src, period28)
	mathAbsCall := &ast.CallExpression{Callee: &ast.Identifier{Name: "abs"}, Arguments: []ast.Expression{src}}

	tests := []struct {
		name            string
		expr            ast.Expression
		wantNested      bool
		preambleMust    []string
		preambleMustNot []string
	}{
		{
			name:         "binary with single TA call",
			expr:         &ast.BinaryExpression{Left: wmaCall, Operator: "*", Right: &ast.Literal{Value: 2.0}},
			wantNested:   true,
			preambleMust: []string{"arrowCtx.GetOrCreateSeries", ".Set("},
		},
		{
			name: "binary with two distinct TA calls",
			expr: &ast.BinaryExpression{
				Left:     &ast.BinaryExpression{Left: &ast.Literal{Value: 2.0}, Operator: "*", Right: wmaCall},
				Operator: "-",
				Right:    smaCall,
			},
			wantNested:   true,
			preambleMust: []string{"arrowCtx.GetOrCreateSeries", ".Set("},
		},
		{
			name:            "binary with only literals and identifiers",
			expr:            &ast.BinaryExpression{Left: src, Operator: "+", Right: period14},
			wantNested:      false,
			preambleMustNot: []string{"arrowCtx.GetOrCreateSeries"},
		},
		{
			name: "conditional with TA in test branch",
			expr: &ast.ConditionalExpression{
				Test:       &ast.BinaryExpression{Left: wmaCall, Operator: ">", Right: src},
				Consequent: &ast.Literal{Value: 1.0},
				Alternate:  &ast.Literal{Value: 0.0},
			},
			wantNested:   true,
			preambleMust: []string{"arrowCtx.GetOrCreateSeries"},
		},
		{
			name: "conditional with TA in alternate branch",
			expr: &ast.ConditionalExpression{
				Test:       &ast.BinaryExpression{Left: src, Operator: ">", Right: &ast.Literal{Value: 0.0}},
				Consequent: &ast.Literal{Value: 0.0},
				Alternate:  smaCall,
			},
			wantNested:   true,
			preambleMust: []string{"arrowCtx.GetOrCreateSeries"},
		},
		{
			name:            "non-TA call (math.abs) produces no entries",
			expr:            &ast.BinaryExpression{Left: mathAbsCall, Operator: "*", Right: period14},
			wantNested:      false,
			preambleMustNot: []string{"arrowCtx.GetOrCreateSeries"},
		},
		{
			name:         "unary minus wrapping a TA call",
			expr:         &ast.UnaryExpression{Operator: "-", Argument: wmaCall},
			wantNested:   true,
			preambleMust: []string{"arrowCtx.GetOrCreateSeries"},
		},
		{
			name:            "bare identifier",
			expr:            src,
			wantNested:      false,
			preambleMustNot: []string{"arrowCtx.GetOrCreateSeries"},
		},
		{
			name:            "literal",
			expr:            period14,
			wantNested:      false,
			preambleMustNot: []string{"arrowCtx.GetOrCreateSeries"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			exprGen := &legacyArrowExpressionGenerator{gen: gen}
			pre := newNestedTACallPrecomputer(exprGen)

			if err := pre.Process(tt.expr); err != nil {
				t.Fatalf("Process() error: %v", err)
			}

			if pre.HasNestedCalls() != tt.wantNested {
				t.Errorf("HasNestedCalls() = %v, want %v", pre.HasNestedCalls(), tt.wantNested)
			}

			preamble := pre.Preamble()
			for _, must := range tt.preambleMust {
				if !strings.Contains(preamble, must) {
					t.Errorf("Preamble() missing %q; got: %q", must, preamble)
				}
			}
			for _, notWant := range tt.preambleMustNot {
				if strings.Contains(preamble, notWant) {
					t.Errorf("Preamble() must not contain %q; got: %q", notWant, preamble)
				}
			}
		})
	}
}

// TestNestedTACallPrecomputer_Deduplication verifies that the same call expression
// appearing multiple times in the tree produces exactly one preamble entry.
func TestNestedTACallPrecomputer_Deduplication(t *testing.T) {
	src := &ast.Identifier{Name: "src"}
	wmaCall := buildTACall("wma", src, &ast.Literal{Value: float64(14)})

	expr := &ast.BinaryExpression{
		Left: &ast.BinaryExpression{
			Left:     &ast.BinaryExpression{Left: &ast.Literal{Value: 3.0}, Operator: "*", Right: wmaCall},
			Operator: "-",
			Right:    wmaCall,
		},
		Operator: "+",
		Right:    wmaCall,
	}

	gen := newTestGenerator()
	exprGen := &legacyArrowExpressionGenerator{gen: gen}
	pre := newNestedTACallPrecomputer(exprGen)

	if err := pre.Process(expr); err != nil {
		t.Fatalf("Process() error: %v", err)
	}

	count := strings.Count(pre.Preamble(), "arrowCtx.GetOrCreateSeries")
	if count != 1 {
		t.Errorf("expected 1 GetOrCreateSeries for deduplicated call, got %d; preamble: %q", count, pre.Preamble())
	}
}

// TestNestedTACallPrecomputer_BuildLookup verifies that the lookup resolves precomputed
// calls by hash and returns empty for calls that were never processed.
func TestNestedTACallPrecomputer_BuildLookup(t *testing.T) {
	src := &ast.Identifier{Name: "src"}
	wmaCall := buildTACall("wma", src, &ast.Literal{Value: float64(14)})
	smaCall := buildTACall("sma", src, &ast.Literal{Value: float64(20)})

	gen := newTestGenerator()
	exprGen := &legacyArrowExpressionGenerator{gen: gen}
	pre := newNestedTACallPrecomputer(exprGen)

	if err := pre.Process(wmaCall); err != nil {
		t.Fatalf("Process() error: %v", err)
	}

	lookup := pre.BuildLookup()

	t.Run("resolves precomputed call", func(t *testing.T) {
		name := lookup(wmaCall)
		if name == "" {
			t.Fatal("BuildLookup() returned empty for precomputed call")
		}
		if !strings.HasPrefix(name, "_ta_src_") {
			t.Errorf("expected _ta_src_ prefix, got %q", name)
		}
	})

	t.Run("same call resolves to same name", func(t *testing.T) {
		if lookup(wmaCall) != lookup(wmaCall) {
			t.Error("identical call must always resolve to the same series name")
		}
	})

	t.Run("unregistered call returns empty", func(t *testing.T) {
		if name := lookup(smaCall); name != "" {
			t.Errorf("unregistered call should return empty, got %q", name)
		}
	})
}

// TestNestedTACallPrecomputer_ZeroState verifies the zero-value state before any processing.
func TestNestedTACallPrecomputer_ZeroState(t *testing.T) {
	gen := newTestGenerator()
	exprGen := &legacyArrowExpressionGenerator{gen: gen}
	pre := newNestedTACallPrecomputer(exprGen)

	if pre.HasNestedCalls() {
		t.Error("HasNestedCalls() must be false before Process()")
	}
	if p := pre.Preamble(); p != "" {
		t.Errorf("Preamble() must be empty before Process(), got %q", p)
	}
	lookup := pre.BuildLookup()
	anyCall := buildTACall("wma", &ast.Identifier{Name: "x"}, &ast.Literal{Value: 14.0})
	if name := lookup(anyCall); name != "" {
		t.Errorf("BuildLookup() must return empty before any precomputation, got %q", name)
	}
}

// TestSeriesExpressionAccessor_Preamble verifies the preamble field: default empty,
// set via WithPreamble, readable via GetPreamble, and chainable.
func TestSeriesExpressionAccessor_Preamble(t *testing.T) {
	expr := &ast.Identifier{Name: "src"}
	symTable := NewSymbolTable()

	tests := []struct {
		name     string
		preamble string
		want     string
	}{
		{"default is empty", "", ""},
		{"single statement", `x := arrowCtx.GetOrCreateSeries("s")`, `x := arrowCtx.GetOrCreateSeries("s")`},
		{"multiple statements", "a := 1; b := 2", "a := 1; b := 2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := NewSeriesExpressionAccessor(expr, symTable, nil)

			if got := acc.GetPreamble(); got != "" {
				t.Errorf("GetPreamble() before WithPreamble must be empty, got %q", got)
			}

			if tt.preamble != "" {
				if returned := acc.WithPreamble(tt.preamble); returned != acc {
					t.Error("WithPreamble() must return the receiver")
				}
			}

			if got := acc.GetPreamble(); got != tt.want {
				t.Errorf("GetPreamble() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestArrowAwareAccessorFactory_BinaryNestedTA verifies that binary expressions containing
// embedded TA calls produce an accessor with a non-empty preamble and loop access that
// references the precomputed arrowCtx series rather than math.NaN().
func TestArrowAwareAccessorFactory_BinaryNestedTA(t *testing.T) {
	src := &ast.Identifier{Name: "_src"}
	period := &ast.Literal{Value: float64(14)}
	period2 := &ast.Literal{Value: float64(28)}

	tests := []struct {
		name         string
		expr         ast.Expression
		wantPreamble bool
		loopMust     []string
		loopMustNot  []string
	}{
		{
			name:         "single nested TA call: 2*wma(src,14)",
			expr:         &ast.BinaryExpression{Left: &ast.Literal{Value: 2.0}, Operator: "*", Right: buildTACall("wma", src, period)},
			wantPreamble: true,
			loopMust:     []string{"_ta_src_", ".Get(j)"},
			loopMustNot:  []string{"math.NaN()"},
		},
		{
			name: "two distinct nested TA calls: 2*wma(src,14) - sma(src,28)",
			expr: &ast.BinaryExpression{
				Left:     &ast.BinaryExpression{Left: &ast.Literal{Value: 2.0}, Operator: "*", Right: buildTACall("wma", src, period)},
				Operator: "-",
				Right:    buildTACall("sma", src, period2),
			},
			wantPreamble: true,
			loopMust:     []string{"_ta_src_", ".Get(j)"},
			loopMustNot:  []string{"math.NaN()"},
		},
		{
			name:         "plain arithmetic without TA calls",
			expr:         &ast.BinaryExpression{Left: src, Operator: "+", Right: period},
			wantPreamble: false,
			loopMustNot:  []string{"math.NaN()"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			factory := arrowParamFactoryFixture(gen, "_src")

			accessor, err := factory.CreateAccessorForExpression(tt.expr)
			if err != nil {
				t.Fatalf("CreateAccessorForExpression() error: %v", err)
			}

			pa, hasPreamble := accessor.(interface{ GetPreamble() string })

			if tt.wantPreamble {
				if !hasPreamble || pa.GetPreamble() == "" {
					t.Errorf("expected non-empty preamble (hasPreamble=%v)", hasPreamble)
				} else if !strings.Contains(pa.GetPreamble(), "arrowCtx.GetOrCreateSeries") {
					t.Errorf("preamble missing GetOrCreateSeries; got: %q", pa.GetPreamble())
				}
			} else if hasPreamble && pa.GetPreamble() != "" {
				t.Errorf("expected empty preamble, got: %q", pa.GetPreamble())
			}

			loopCode := accessor.GenerateLoopValueAccess("j")
			for _, must := range tt.loopMust {
				if !strings.Contains(loopCode, must) {
					t.Errorf("loop access missing %q; got: %q", must, loopCode)
				}
			}
			for _, notWant := range tt.loopMustNot {
				if strings.Contains(loopCode, notWant) {
					t.Errorf("loop access must not contain %q; got: %q", notWant, loopCode)
				}
			}
		})
	}
}

// TestArrowAwareAccessorFactory_ConditionalNestedTA verifies the same precomputer treatment
// for conditional expressions whose branches contain embedded TA calls.
func TestArrowAwareAccessorFactory_ConditionalNestedTA(t *testing.T) {
	src := &ast.Identifier{Name: "_src"}
	period := &ast.Literal{Value: float64(14)}
	positiveTest := &ast.BinaryExpression{Left: src, Operator: ">", Right: &ast.Literal{Value: 0.0}}

	tests := []struct {
		name         string
		expr         *ast.ConditionalExpression
		wantPreamble bool
		loopMustNot  []string
	}{
		{
			name: "TA call in consequent branch",
			expr: &ast.ConditionalExpression{
				Test: positiveTest, Consequent: buildTACall("wma", src, period), Alternate: &ast.Literal{Value: 0.0},
			},
			wantPreamble: true,
			loopMustNot:  []string{"math.NaN()"},
		},
		{
			name: "TA call in alternate branch",
			expr: &ast.ConditionalExpression{
				Test: positiveTest, Consequent: &ast.Literal{Value: 0.0}, Alternate: buildTACall("sma", src, period),
			},
			wantPreamble: true,
			loopMustNot:  []string{"math.NaN()"},
		},
		{
			name: "no TA calls",
			expr: &ast.ConditionalExpression{
				Test: positiveTest, Consequent: &ast.Literal{Value: 1.0}, Alternate: &ast.Literal{Value: 0.0},
			},
			wantPreamble: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			factory := arrowParamFactoryFixture(gen, "_src")

			accessor, err := factory.CreateAccessorForExpression(tt.expr)
			if err != nil {
				t.Fatalf("CreateAccessorForExpression() error: %v", err)
			}

			pa, hasPreamble := accessor.(interface{ GetPreamble() string })

			if tt.wantPreamble {
				if !hasPreamble || pa.GetPreamble() == "" {
					t.Errorf("expected preamble (hasPreamble=%v)", hasPreamble)
				}
			} else if hasPreamble && pa.GetPreamble() != "" {
				t.Errorf("expected empty preamble; got: %q", pa.GetPreamble())
			}

			loopCode := accessor.GenerateLoopValueAccess("j")
			for _, notWant := range tt.loopMustNot {
				if strings.Contains(loopCode, notWant) {
					t.Errorf("loop access must not contain %q; got: %q", notWant, loopCode)
				}
			}
		})
	}
}

// TestArrowAwareAccessorFactory_BinaryNestedTA_PreambleDeduplication verifies that two
// occurrences of the same nested TA call produce exactly one series in the preamble.
func TestArrowAwareAccessorFactory_BinaryNestedTA_PreambleDeduplication(t *testing.T) {
	src := &ast.Identifier{Name: "_src"}
	wmaCall := buildTACall("wma", src, &ast.Literal{Value: float64(14)})

	expr := &ast.BinaryExpression{
		Left:     &ast.BinaryExpression{Left: &ast.Literal{Value: 2.0}, Operator: "*", Right: wmaCall},
		Operator: "-",
		Right:    wmaCall,
	}

	gen := newTestGenerator()
	factory := arrowParamFactoryFixture(gen, "_src")

	accessor, err := factory.CreateAccessorForExpression(expr)
	if err != nil {
		t.Fatalf("CreateAccessorForExpression() error: %v", err)
	}

	pa, ok := accessor.(interface{ GetPreamble() string })
	if !ok || pa.GetPreamble() == "" {
		t.Fatal("expected non-empty preamble")
	}

	count := strings.Count(pa.GetPreamble(), "arrowCtx.GetOrCreateSeries")
	if count != 1 {
		t.Errorf("expected 1 GetOrCreateSeries for deduplicated call, got %d; preamble: %q", count, pa.GetPreamble())
	}
}

// TestArrowNestedTA_HMAPatternCodegen is an end-to-end codegen test verifying that UDFs
// containing TA-inside-TA call patterns (HMA, EHMA) produce precomputed arrowCtx series
// in the generated Go code and do not propagate math.NaN() as the outer TA source operand.
func TestArrowNestedTA_HMAPatternCodegen(t *testing.T) {
	tests := []struct {
		name           string
		script         string
		mustContain    []string
		mustNotContain []string
	}{
		{
			name: "HMA: wma of (2*wma - wma)",
			script: `//@version=5
indicator("HMA Test")
HMA(src, length) =>
    wma(2 * wma(src, length / 2) - wma(src, length), math.round(math.sqrt(length)))
plot(HMA(close, 55))
`,
			mustContain:    []string{"_ta_src_", "arrowCtx.GetOrCreateSeries", ".Get("},
			mustNotContain: []string{"2 * math.NaN()", "math.NaN()) - math.NaN()"},
		},
		{
			name: "EHMA: ema of (2*ema - ema)",
			script: `//@version=5
indicator("EHMA Test")
EHMA(src, length) =>
    ta.ema(2 * ta.ema(src, length / 2) - ta.ema(src, length), math.round(math.sqrt(length)))
plot(EHMA(close, 55))
`,
			mustContain:    []string{"_ta_src_", "arrowCtx.GetOrCreateSeries"},
			mustNotContain: []string{"2 * math.NaN()"},
		},
		{
			name: "plain WMA with identifier source incurs no _ta_src_ overhead",
			script: `//@version=5
indicator("Plain WMA Test")
myWMA(src, len) => wma(src, len)
plot(myWMA(close, 20))
`,
			mustNotContain: []string{"2 * math.NaN()", "math.NaN()) -"},
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
					t.Errorf("generated code missing %q", want)
				}
			}
			for _, notWant := range tt.mustNotContain {
				if strings.Contains(full, notWant) {
					t.Errorf("generated code must not contain %q", notWant)
				}
			}
		})
	}
}
