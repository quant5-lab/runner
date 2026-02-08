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
