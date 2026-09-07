package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestTupleSecurityEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		varNames      []string
		call          *ast.CallExpression
		shouldError   bool
		errorContains string
		description   string
	}{
		{
			name:     "zero-element tuple",
			varNames: []string{},
			call: buildSecurityCallWithExprs("security", []ast.Expression{
				&ast.Identifier{Name: "close"},
			}, nil),
			shouldError:   true,
			errorContains: "empty tuple pattern",
			description:   "empty tuple should fail validation",
		},
		{
			name:     "cardinality mismatch - more vars than exprs",
			varNames: []string{"a", "b", "c"},
			call: buildSecurityCallWithExprs("security", []ast.Expression{
				&ast.Identifier{Name: "close"},
			}, nil),
			shouldError:   true,
			errorContains: "cardinality mismatch",
			description:   "3 variables but 1 expression",
		},
		{
			name:     "cardinality mismatch - fewer vars than exprs",
			varNames: []string{"a"},
			call: buildSecurityCallWithExprs("security", []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Identifier{Name: "open"},
			}, nil),
			shouldError:   true,
			errorContains: "cardinality mismatch",
			description:   "1 variable but 2 expressions",
		},
		{
			name:     "non-array third argument",
			varNames: []string{"a"},
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "AAPL"},
					&ast.Literal{Value: "1D"},
					&ast.Identifier{Name: "close"},
				},
			},
			shouldError:   true,
			errorContains: "expected array literal",
			description:   "third arg must be array, not identifier",
		},
		{
			name:     "missing third argument",
			varNames: []string{"a"},
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "AAPL"},
					&ast.Literal{Value: "1D"},
				},
			},
			shouldError:   true,
			errorContains: "at least 3 arguments",
			description:   "requires symbol, timeframe, expressions",
		},
		{
			name:     "only two arguments",
			varNames: []string{"a"},
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "AAPL"},
				},
			},
			shouldError:   true,
			errorContains: "at least 3 arguments",
			description:   "missing timeframe and expressions",
		},
		{
			name:          "empty expression array",
			varNames:      []string{},
			call:          buildSecurityCallWithExprs("security", []ast.Expression{}, nil),
			shouldError:   true,
			errorContains: "empty tuple pattern",
			description:   "array literal cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGeneratorForSecurityTupleTests()
			declarator := buildTupleDeclarator(tt.varNames, tt.call)

			_, err := gen.generateTupleDestructuringDeclaration(declarator)

			if tt.shouldError {
				if err == nil {
					t.Fatalf("expected error containing %q, got none", tt.errorContains)
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestTupleSecuritySimpleOHLCV(t *testing.T) {
	tests := []struct {
		name     string
		varNames []string
		fields   []string
		funcName string
	}{
		{
			name:     "close and open",
			varNames: []string{"c", "o"},
			fields:   []string{"close", "open"},
			funcName: "request.security",
		},
		{
			name:     "all OHLCV",
			varNames: []string{"o", "h", "l", "c", "v"},
			fields:   []string{"open", "high", "low", "close", "volume"},
			funcName: "request.security",
		},
		{
			name:     "v4 syntax",
			varNames: []string{"c", "v"},
			fields:   []string{"close", "volume"},
			funcName: "security",
		},
		{
			name:     "single element tuple",
			varNames: []string{"c"},
			fields:   []string{"close"},
			funcName: "request.security",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGeneratorForSecurityTupleTests()
			call := buildSecurityCall(tt.funcName, tt.fields, nil)
			declarator := buildTupleDeclarator(tt.varNames, call)

			code, err := gen.generateTupleDestructuringDeclaration(declarator)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assertContains(t, code, "secKey :=")
			assertContains(t, code, "secCtx, secFound := securityContexts[secKey]")
			assertContains(t, code, "securityBarMapper, mapperFound := securityBarMappers[secKey]")
			assertContains(t, code, "secBarIdx := securityBarMapper.FindDailyBarIndex(ctx.BarIndex, secLookahead)")

			for i, varName := range tt.varNames {
				goField := ohlcvGoField(tt.fields[i])
				expected := fmt.Sprintf("%sSeries.Set(secCtx.Data[secBarIdx].%s)", varName, goField)
				assertContains(t, code, expected)
			}

			nanCount := strings.Count(code, "Series.Set(math.NaN())")
			expectedNanSets := len(tt.varNames) * 3
			if nanCount != expectedNanSets {
				t.Errorf("expected %d NaN fallbacks (%d vars × 3 guards), got %d\n%s",
					expectedNanSets, len(tt.varNames), nanCount, code)
			}
		})
	}
}

func TestTupleSecurityComplexExpressions(t *testing.T) {
	tests := []struct {
		name     string
		varNames []string
		exprs    []ast.Expression
	}{
		{
			name:     "TA call expressions",
			varNames: []string{"emaVal", "smaVal"},
			exprs: []ast.Expression{
				buildTACallExpression("ema", 10.0),
				buildTACallExpression("sma", 20.0),
			},
		},
		{
			name:     "binary expression",
			varNames: []string{"spread"},
			exprs: []ast.Expression{
				&ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "high"},
					Operator: "-",
					Right:    &ast.Identifier{Name: "low"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGeneratorForSecurityTupleTests()
			call := buildSecurityCallWithExprs("request.security", tt.exprs, nil)
			declarator := buildTupleDeclarator(tt.varNames, call)

			code, err := gen.generateTupleDestructuringDeclaration(declarator)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assertContains(t, code, "secBarEvaluator == nil")
			assertContains(t, code, "security.NewStreamingBarEvaluator()")

			for _, varName := range tt.varNames {
				assertContains(t, code, fmt.Sprintf("%sSeries.Set(secValue)", varName))
			}
		})
	}
}

/* Scope isolation: each complex expression evaluated in separate block to prevent Go redeclaration errors */
func TestTupleSecurityMultipleComplexExpressions(t *testing.T) {
	tests := []struct {
		name      string
		numExprs  int
		buildFunc func(int) []ast.Expression
	}{
		{
			name:     "single complex expression",
			numExprs: 1,
			buildFunc: func(n int) []ast.Expression {
				return []ast.Expression{buildTACallExpression("ema", 10.0)}
			},
		},
		{
			name:     "two complex expressions",
			numExprs: 2,
			buildFunc: func(n int) []ast.Expression {
				return []ast.Expression{
					buildTACallExpression("ema", 10.0),
					buildTACallExpression("sma", 20.0),
				}
			},
		},
		{
			name:     "three complex expressions",
			numExprs: 3,
			buildFunc: func(n int) []ast.Expression {
				return []ast.Expression{
					buildTACallExpression("ema", 10.0),
					buildTACallExpression("sma", 20.0),
					buildTACallExpression("rsi", 14.0),
				}
			},
		},
		{
			name:     "five complex expressions",
			numExprs: 5,
			buildFunc: func(n int) []ast.Expression {
				indicators := []string{"ema", "sma", "rsi", "cci", "atr"}
				exprs := make([]ast.Expression, n)
				for i := 0; i < n; i++ {
					exprs[i] = buildTACallExpression(indicators[i], float64(10+i*5))
				}
				return exprs
			},
		},
		{
			name:     "eight complex expressions",
			numExprs: 8,
			buildFunc: func(n int) []ast.Expression {
				indicators := []string{"ema", "sma", "rsi", "cci", "atr", "adx", "obv", "roc"}
				exprs := make([]ast.Expression, n)
				for i := 0; i < n; i++ {
					exprs[i] = buildTACallExpression(indicators[i], float64(10+i*2))
				}
				return exprs
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGeneratorForSecurityTupleTests()
			exprs := tt.buildFunc(tt.numExprs)
			varNames := make([]string, tt.numExprs)
			for i := 0; i < tt.numExprs; i++ {
				varNames[i] = fmt.Sprintf("var%d", i)
			}

			call := buildSecurityCallWithExprs("request.security", exprs, nil)
			declarator := buildTupleDeclarator(varNames, call)

			code, err := gen.generateTupleDestructuringDeclaration(declarator)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			evalCount := strings.Count(code, "secValue, err := secBarEvaluator.EvaluateAtBar(")
			if evalCount != tt.numExprs {
				t.Errorf("expected %d EvaluateAtBar calls, got %d", tt.numExprs, evalCount)
			}

			for _, varName := range varNames {
				expected := fmt.Sprintf("%sSeries.Set(secValue)", varName)
				if !strings.Contains(code, expected) {
					t.Errorf("missing Series.Set for %s", varName)
				}
			}

			lines := strings.Split(code, "\n")
			standaloneOpenBraces := 0
			for _, line := range lines {
				if strings.TrimSpace(line) == "{" {
					standaloneOpenBraces++
				}
			}

			expectedMinBraces := tt.numExprs + 1
			if standaloneOpenBraces < expectedMinBraces {
				t.Errorf("expected at least %d scope blocks (1 outer + %d elements), got %d",
					expectedMinBraces, tt.numExprs, standaloneOpenBraces)
			}
		})
	}
}

func TestTupleSecurityMixedExpressions(t *testing.T) {
	gen := newTestGeneratorForSecurityTupleTests()

	exprs := []ast.Expression{
		&ast.Identifier{Name: "close"},
		buildTACallExpression("ema", 10.0),
		&ast.Identifier{Name: "volume"},
	}

	varNames := []string{"c", "emaVal", "v"}
	call := buildSecurityCallWithExprs("request.security", exprs, nil)
	declarator := buildTupleDeclarator(varNames, call)

	code, err := gen.generateTupleDestructuringDeclaration(declarator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertContains(t, code, "cSeries.Set(secCtx.Data[secBarIdx].Close)")
	assertContains(t, code, "vSeries.Set(secCtx.Data[secBarIdx].Volume)")
	assertContains(t, code, "emaValSeries.Set(secValue)")
}

func TestTupleSecurityLookahead(t *testing.T) {
	tests := []struct {
		name     string
		opts     []ast.Expression
		expected string
	}{
		{
			name:     "no lookahead argument",
			opts:     nil,
			expected: "secLookahead := false",
		},
		{
			name: "positional true literal",
			opts: []ast.Expression{
				&ast.Literal{Value: true},
			},
			expected: "secLookahead := true",
		},
		{
			name: "positional false literal",
			opts: []ast.Expression{
				&ast.Literal{Value: false},
			},
			expected: "secLookahead := false",
		},
		{
			name: "named lookahead via ObjectExpression true",
			opts: []ast.Expression{
				&ast.ObjectExpression{
					Properties: []ast.Property{
						{
							Key:   &ast.Identifier{Name: "lookahead"},
							Value: &ast.Literal{Value: true},
						},
					},
				},
			},
			expected: "secLookahead := true",
		},
		{
			name: "named lookahead via ObjectExpression false",
			opts: []ast.Expression{
				&ast.ObjectExpression{
					Properties: []ast.Property{
						{
							Key:   &ast.Identifier{Name: "lookahead"},
							Value: &ast.Literal{Value: false},
						},
					},
				},
			},
			expected: "secLookahead := false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGeneratorForSecurityTupleTests()
			call := buildSecurityCall("request.security", []string{"close", "open"}, tt.opts)
			declarator := buildTupleDeclarator([]string{"c", "o"}, call)

			code, err := gen.generateTupleDestructuringDeclaration(declarator)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assertContains(t, code, tt.expected)
		})
	}
}

func TestTupleSecurityOHLCVArbitraryCardinality(t *testing.T) {
	cardinalities := []int{1, 2, 3, 5, 8}

	for _, n := range cardinalities {
		t.Run(fmt.Sprintf("cardinality_%d", n), func(t *testing.T) {
			gen := newTestGeneratorForSecurityTupleTests()

			varNames := make([]string, n)
			fields := make([]string, n)
			ohlcvCycle := []string{"open", "high", "low", "close", "volume"}
			for i := 0; i < n; i++ {
				varNames[i] = fmt.Sprintf("v%d", i)
				fields[i] = ohlcvCycle[i%len(ohlcvCycle)]
			}

			call := buildSecurityCall("request.security", fields, nil)
			declarator := buildTupleDeclarator(varNames, call)

			code, err := gen.generateTupleDestructuringDeclaration(declarator)
			if err != nil {
				t.Fatalf("unexpected error for cardinality %d: %v", n, err)
			}

			for _, varName := range varNames {
				assertContains(t, code, varName+"Series.Set(")
			}

			nanCount := strings.Count(code, "Series.Set(math.NaN())")
			if nanCount != n*3 {
				t.Errorf("expected %d NaN fallbacks (%d vars × 3 guards), got %d", n*3, n, nanCount)
			}
		})
	}
}

func TestTupleSecurityComplexArbitraryCardinality(t *testing.T) {
	cardinalities := []int{1, 2, 4, 6}

	for _, n := range cardinalities {
		t.Run(fmt.Sprintf("cardinality_%d", n), func(t *testing.T) {
			gen := newTestGeneratorForSecurityTupleTests()

			indicators := []string{"ema", "sma", "rsi", "cci", "atr", "adx"}
			varNames := make([]string, n)
			exprs := make([]ast.Expression, n)
			for i := 0; i < n; i++ {
				varNames[i] = fmt.Sprintf("ind%d", i)
				exprs[i] = buildTACallExpression(indicators[i%len(indicators)], float64(10+i*5))
			}

			call := buildSecurityCallWithExprs("security", exprs, nil)
			declarator := buildTupleDeclarator(varNames, call)

			code, err := gen.generateTupleDestructuringDeclaration(declarator)
			if err != nil {
				t.Fatalf("unexpected error for cardinality %d: %v", n, err)
			}

			evalCount := strings.Count(code, "secValue, err := secBarEvaluator.EvaluateAtBar(")
			if evalCount != n {
				t.Errorf("expected %d EvaluateAtBar calls for %d complex exprs, got %d", n, n, evalCount)
			}

			for _, varName := range varNames {
				expected := fmt.Sprintf("%sSeries.Set(secValue)", varName)
				assertContains(t, code, expected)
			}
		})
	}
}

func TestTupleSecurityMixedArbitraryCardinality(t *testing.T) {
	tests := []struct {
		name         string
		ohlcvCount   int
		complexCount int
	}{
		{"1 OHLCV + 1 complex", 1, 1},
		{"2 OHLCV + 1 complex", 2, 1},
		{"1 OHLCV + 2 complex", 1, 2},
		{"3 OHLCV + 2 complex", 3, 2},
		{"2 OHLCV + 3 complex", 2, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGeneratorForSecurityTupleTests()

			totalCount := tt.ohlcvCount + tt.complexCount
			varNames := make([]string, totalCount)
			exprs := make([]ast.Expression, totalCount)

			ohlcvFields := []string{"close", "open", "high", "low", "volume"}
			indicators := []string{"ema", "sma", "rsi"}

			for i := 0; i < tt.ohlcvCount; i++ {
				varNames[i] = fmt.Sprintf("ohlcv%d", i)
				exprs[i] = &ast.Identifier{Name: ohlcvFields[i%len(ohlcvFields)]}
			}

			for i := 0; i < tt.complexCount; i++ {
				idx := tt.ohlcvCount + i
				varNames[idx] = fmt.Sprintf("ind%d", i)
				exprs[idx] = buildTACallExpression(indicators[i%len(indicators)], float64(10+i*5))
			}

			call := buildSecurityCallWithExprs("request.security", exprs, nil)
			declarator := buildTupleDeclarator(varNames, call)

			code, err := gen.generateTupleDestructuringDeclaration(declarator)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			evalCount := strings.Count(code, "secValue, err := secBarEvaluator.EvaluateAtBar(")
			if evalCount != tt.complexCount {
				t.Errorf("expected %d EvaluateAtBar calls, got %d", tt.complexCount, evalCount)
			}

			directAccessCount := 0
			for i := 0; i < tt.ohlcvCount; i++ {
				field := ohlcvFields[i%len(ohlcvFields)]
				goField := ohlcvGoField(field)
				if strings.Contains(code, fmt.Sprintf("secCtx.Data[secBarIdx].%s)", goField)) {
					directAccessCount++
				}
			}
			if directAccessCount != tt.ohlcvCount {
				t.Errorf("expected %d OHLCV direct accesses, got %d", tt.ohlcvCount, directAccessCount)
			}
		})
	}
}

func TestTupleSecurityRuntimeArguments(t *testing.T) {
	tests := []struct {
		name          string
		symbolArg     ast.Expression
		timeframeArg  ast.Expression
		expectFmt     bool
		expectCtxSym  bool
		shouldError   bool
		errorContains string
		description   string
	}{
		{
			name:         "runtime symbol via syminfo.tickerid",
			symbolArg:    &ast.MemberExpression{Object: &ast.Identifier{Name: "syminfo"}, Property: &ast.Identifier{Name: "tickerid"}},
			timeframeArg: &ast.Literal{Value: "1D"},
			expectFmt:    true,
			expectCtxSym: true,
			shouldError:  false,
			description:  "symbol from context, timeframe literal",
		},
		{
			name:          "runtime timeframe via variable (unsupported)",
			symbolArg:     &ast.Literal{Value: "AAPL"},
			timeframeArg:  &ast.Identifier{Name: "tf"},
			shouldError:   true,
			errorContains: "unsupported timeframe identifier",
			description:   "variable timeframe not yet supported",
		},
		{
			name:          "runtime timeframe via member expression (unsupported)",
			symbolArg:     &ast.MemberExpression{Object: &ast.Identifier{Name: "syminfo"}, Property: &ast.Identifier{Name: "tickerid"}},
			timeframeArg:  &ast.MemberExpression{Object: &ast.Identifier{Name: "input"}, Property: &ast.Identifier{Name: "timeframe"}},
			shouldError:   true,
			errorContains: "unsupported member expression for timeframe",
			description:   "member expression timeframe not yet supported",
		},
		{
			name:         "literal symbol and timeframe",
			symbolArg:    &ast.Literal{Value: "BTCUSDT"},
			timeframeArg: &ast.Literal{Value: "1H"},
			expectFmt:    false,
			expectCtxSym: false,
			shouldError:  false,
			description:  "both compile-time literals",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGeneratorForSecurityTupleTests()
			call := &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					tt.symbolArg,
					tt.timeframeArg,
					&ast.Literal{
						Value: []ast.Expression{
							&ast.Identifier{Name: "close"},
							&ast.Identifier{Name: "open"},
						},
						Raw: "[close, open]",
					},
				},
			}
			declarator := buildTupleDeclarator([]string{"c", "o"}, call)

			code, err := gen.generateTupleDestructuringDeclaration(declarator)

			if tt.shouldError {
				if err == nil {
					t.Fatalf("expected error containing %q, got none", tt.errorContains)
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got: %v", tt.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectFmt {
				assertContains(t, code, "fmt.Sprintf")
			} else {
				assertContains(t, code, "secKey := ")
			}

			if tt.expectCtxSym {
				assertContains(t, code, "ctx.Symbol")
			}
		})
	}
}

func TestExtractTupleExpressionElements(t *testing.T) {
	t.Run("valid array literal", func(t *testing.T) {
		expr := &ast.Literal{
			Value: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Identifier{Name: "open"},
			},
		}
		elems, err := extractTupleExpressionElements(expr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(elems) != 2 {
			t.Fatalf("expected 2 elements, got %d", len(elems))
		}
	})

	t.Run("non-literal expression", func(t *testing.T) {
		_, err := extractTupleExpressionElements(&ast.Identifier{Name: "close"})
		if err == nil {
			t.Fatal("expected error for non-literal")
		}
	})

	t.Run("literal with non-array value", func(t *testing.T) {
		_, err := extractTupleExpressionElements(&ast.Literal{Value: "string"})
		if err == nil {
			t.Fatal("expected error for non-array literal value")
		}
	})

	t.Run("empty array", func(t *testing.T) {
		_, err := extractTupleExpressionElements(&ast.Literal{Value: []ast.Expression{}})
		if err == nil {
			t.Fatal("expected error for empty array")
		}
	})
}

func buildTACallExpression(indicator string, period float64) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: indicator},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: period},
		},
	}
}

func buildSecurityCall(funcName string, fields []string, extraArgs []ast.Expression) *ast.CallExpression {
	exprs := make([]ast.Expression, len(fields))
	for i, f := range fields {
		exprs[i] = &ast.Identifier{Name: f}
	}
	return buildSecurityCallWithExprs(funcName, exprs, extraArgs)
}

func buildSecurityCallWithExprs(funcName string, exprs []ast.Expression, extraArgs []ast.Expression) *ast.CallExpression {
	var callee ast.Expression
	if funcName == "request.security" {
		callee = &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "request"},
			Property: &ast.Identifier{Name: "security"},
		}
	} else {
		callee = &ast.Identifier{Name: funcName}
	}

	args := []ast.Expression{
		&ast.Literal{Value: "AAPL"},
		&ast.Literal{Value: "1D"},
		&ast.Literal{
			Value: exprs,
			Raw:   "[...]",
		},
	}
	args = append(args, extraArgs...)

	return &ast.CallExpression{
		Callee:    callee,
		Arguments: args,
	}
}

func buildTupleDeclarator(varNames []string, call *ast.CallExpression) ast.VariableDeclarator {
	elements := make([]ast.Identifier, len(varNames))
	for i, name := range varNames {
		elements[i] = ast.Identifier{Name: name}
	}
	return ast.VariableDeclarator{
		ID:   &ast.ArrayPattern{Elements: elements},
		Init: call,
	}
}

func newTestGeneratorForSecurityTupleTests() *generator {
	gen := newTestGeneratorForTupleTests()
	gen.symbolTable = NewSymbolTable()
	return gen
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("expected output to contain %q\ngot:\n%s", needle, haystack)
	}
}

func ohlcvGoField(field string) string {
	switch field {
	case "close":
		return "Close"
	case "open":
		return "Open"
	case "high":
		return "High"
	case "low":
		return "Low"
	case "volume":
		return "Volume"
	default:
		return "Close"
	}
}

// TestTupleSecurityUDFCallDispatch verifies that [vars...] = security(sym, tf, UDF(args...))
// delegates to SecurityUDFCallGenerator.EmitTupleCall for any registered UDF regardless of
// argument count, var count, or UDF name. It covers the full matrix of screener patterns
// that embed Pine user-defined functions as the third security() argument.
func TestTupleSecurityUDFCallDispatch(t *testing.T) {
	tests := []struct {
		name     string
		udfName  string
		udfArgs  []ast.Expression
		varNames []string
		// assertions on emitted code
		wantContains []string
	}{
		{
			name:    "two-var two-arg UDF (Pmax-style screener)",
			udfName: "Pmax",
			udfArgs: []ast.Expression{
				&ast.Identifier{Name: "Multiplier"},
				&ast.Identifier{Name: "Periods"},
			},
			varNames: []string{"trend", "tsl"},
			wantContains: []string{
				"context.NewArrowContext(secCtx)",
				"Pmax(",
				"trendSeries.Set(trend)",
				"tslSeries.Set(tsl)",
				"secKey :=",
				"secCtx, secFound := securityContexts[secKey]",
				"securityBarMapper, mapperFound := securityBarMappers[secKey]",
				"secBarIdx := securityBarMapper.FindDailyBarIndex(ctx.BarIndex, secLookahead)",
			},
		},
		{
			name:     "single-var no-arg UDF",
			udfName:  "GetSignal",
			udfArgs:  []ast.Expression{},
			varNames: []string{"sig"},
			wantContains: []string{
				"context.NewArrowContext(secCtx)",
				// arrow context is always injected as first arg, so no-arg UDF
				// appears as GetSignal(arrowCtxVar) — match the prefix only.
				"GetSignal(",
				"sigSeries.Set(sig)",
			},
		},
		{
			name:    "single-var multi-arg UDF",
			udfName: "Score",
			udfArgs: []ast.Expression{
				&ast.Literal{Value: 14.0},
				&ast.Identifier{Name: "close"},
				&ast.Identifier{Name: "volume"},
			},
			varNames: []string{"score"},
			wantContains: []string{
				"context.NewArrowContext(secCtx)",
				"Score(",
				"scoreSeries.Set(score)",
			},
		},
		{
			name:     "three-var UDF returning triple",
			udfName:  "TripleOutput",
			udfArgs:  []ast.Expression{&ast.Identifier{Name: "src"}},
			varNames: []string{"a", "b", "c"},
			wantContains: []string{
				"context.NewArrowContext(secCtx)",
				"TripleOutput(",
				"aSeries.Set(a)",
				"bSeries.Set(b)",
				"cSeries.Set(c)",
			},
		},
		{
			name:     "UDF with underscore-prefixed name",
			udfName:  "_internal_helper",
			udfArgs:  []ast.Expression{&ast.Identifier{Name: "close"}},
			varNames: []string{"out1", "out2"},
			wantContains: []string{
				"_internal_helper(",
				"out1Series.Set(out1)",
				"out2Series.Set(out2)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGeneratorForSecurityTupleTests()
			gen.variables[tt.udfName] = "function"

			udfCall := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: tt.udfName},
				Arguments: tt.udfArgs,
			}
			call := &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "AAPL"},
					&ast.Literal{Value: "60"},
					udfCall,
				},
			}

			declarator := buildTupleDeclarator(tt.varNames, call)
			code, err := gen.generateTupleDestructuringDeclaration(declarator)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, want := range tt.wantContains {
				assertContains(t, code, want)
			}
		})
	}
}

// TestTupleSecurityThirdArgRejectedKinds verifies that third-argument forms that are
// neither an array literal nor a registered UDF call produce a consistent error.
// Each case tests a structurally different argument type to ensure no form silently
// slips through to undefined behavior.
func TestTupleSecurityThirdArgRejectedKinds(t *testing.T) {
	tests := []struct {
		name    string
		arg     ast.Expression
		varsCnt int
		wantErr string
	}{
		{
			name:    "non-UDF builtin call (ta.sma)",
			varsCnt: 1,
			arg: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 14.0},
				},
			},
			wantErr: "expected array literal",
		},
		{
			name:    "non-UDF flat call (math.round)",
			varsCnt: 1,
			arg: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: "round"},
				},
				Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
			},
			wantErr: "expected array literal",
		},
		{
			name:    "unregistered identifier-call",
			varsCnt: 1,
			arg: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "notAUDF"},
			},
			wantErr: "expected array literal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGeneratorForSecurityTupleTests()
			// deliberately do NOT register the callee as a UDF

			varNames := make([]string, tt.varsCnt)
			for i := range varNames {
				varNames[i] = fmt.Sprintf("v%d", i)
			}

			call := &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "AAPL"},
					&ast.Literal{Value: "60"},
					tt.arg,
				},
			}

			declarator := buildTupleDeclarator(varNames, call)
			_, err := gen.generateTupleDestructuringDeclaration(declarator)
			if err == nil {
				t.Fatal("expected error but got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected %q in error, got: %v", tt.wantErr, err)
			}
		})
	}
}

// TestResolveTupleExpressionArg exercises every branch of the shared helper, including
// invariants about when cardinality is and is not checked.
func TestResolveTupleExpressionArg(t *testing.T) {
	gen := newTestGeneratorForSecurityTupleTests()
	gen.variables["MyUDF"] = "function"
	detector := NewUserDefinedFunctionDetector(gen.variables)

	t.Run("UDF call yields udfCall with nil elements", func(t *testing.T) {
		call := &ast.CallExpression{Callee: &ast.Identifier{Name: "MyUDF"}}
		elems, udfCall, err := resolveTupleExpressionArg(detector, call, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if udfCall == nil {
			t.Fatal("expected udfCall to be non-nil")
		}
		if elems != nil {
			t.Fatalf("expected elements to be nil for UDF path, got %v", elems)
		}
	})

	t.Run("UDF path bypasses cardinality check", func(t *testing.T) {
		// The UDF determines its output count at runtime; the caller's var count
		// is irrelevant at the parse stage and must not cause a cardinality error.
		call := &ast.CallExpression{Callee: &ast.Identifier{Name: "MyUDF"}}
		_, udfCall, err := resolveTupleExpressionArg(detector, call, 999)
		if err != nil {
			t.Fatalf("UDF path should not check cardinality, got error: %v", err)
		}
		if udfCall == nil {
			t.Fatal("expected udfCall to be non-nil")
		}
	})

	t.Run("array literal with matching cardinality returns elements", func(t *testing.T) {
		lit := &ast.Literal{
			Value: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Identifier{Name: "open"},
			},
		}
		elems, udfCall, err := resolveTupleExpressionArg(detector, lit, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if udfCall != nil {
			t.Fatalf("expected udfCall to be nil for literal path")
		}
		if len(elems) != 2 {
			t.Fatalf("expected 2 elements, got %d", len(elems))
		}
	})

	t.Run("array literal cardinality mismatch returns error", func(t *testing.T) {
		lit := &ast.Literal{
			Value: []ast.Expression{&ast.Identifier{Name: "close"}},
		}
		_, _, err := resolveTupleExpressionArg(detector, lit, 3)
		if err == nil {
			t.Fatal("expected cardinality error")
		}
		if !strings.Contains(err.Error(), "cardinality mismatch") {
			t.Errorf("expected 'cardinality mismatch' in error, got: %v", err)
		}
	})

	t.Run("empty array literal rejected", func(t *testing.T) {
		// An empty literal [] has no expressions to bind to variables.
		// extractTupleExpressionElements catches this before cardinality is reached.
		lit := &ast.Literal{Value: []ast.Expression{}}
		_, _, err := resolveTupleExpressionArg(detector, lit, 2)
		if err == nil {
			t.Fatal("expected error for empty literal")
		}
		if !strings.Contains(err.Error(), "empty expression array") {
			t.Errorf("expected 'empty expression array' in error, got: %v", err)
		}
	})

	t.Run("non-UDF CallExpression rejected", func(t *testing.T) {
		call := &ast.CallExpression{Callee: &ast.Identifier{Name: "unknownFunc"}}
		_, _, err := resolveTupleExpressionArg(detector, call, 1)
		if err == nil {
			t.Fatal("expected error for non-UDF call")
		}
		if !strings.Contains(err.Error(), "expected array literal") {
			t.Errorf("expected 'expected array literal' in error, got: %v", err)
		}
	})

	t.Run("bare identifier rejected", func(t *testing.T) {
		_, _, err := resolveTupleExpressionArg(detector, &ast.Identifier{Name: "close"}, 1)
		if err == nil {
			t.Fatal("expected error for identifier")
		}
		if !strings.Contains(err.Error(), "expected array literal") {
			t.Errorf("expected 'expected array literal' in error, got: %v", err)
		}
	})

	t.Run("binary expression rejected", func(t *testing.T) {
		expr := &ast.BinaryExpression{
			Operator: "+",
			Left:     &ast.Identifier{Name: "close"},
			Right:    &ast.Literal{Value: 1.0},
		}
		_, _, err := resolveTupleExpressionArg(detector, expr, 1)
		if err == nil {
			t.Fatal("expected error for binary expression")
		}
		if !strings.Contains(err.Error(), "expected array literal") {
			t.Errorf("expected 'expected array literal' in error, got: %v", err)
		}
	})
}
