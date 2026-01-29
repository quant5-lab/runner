package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/security"
)

func TestAnalyzeAndGeneratePrefetch_NoSecurityCalls(t *testing.T) {

	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body:     []ast.Node{},
	}

	injection, err := AnalyzeAndGeneratePrefetch(program)
	if err != nil {
		t.Fatalf("AnalyzeAndGeneratePrefetch failed: %v", err)
	}

	if injection.PrefetchCode != "" {
		t.Error("Expected empty prefetch code when no security() calls")
	}

	if len(injection.ImportPaths) != 0 {
		t.Errorf("Expected 0 imports, got %d", len(injection.ImportPaths))
	}
}

func TestAnalyzeAndGeneratePrefetch_WithSecurityCall(t *testing.T) {

	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.VariableDeclaration{
				NodeType: ast.TypeVariableDeclaration,
				Kind:     "var",
				Declarations: []ast.VariableDeclarator{
					{
						NodeType: ast.TypeVariableDeclarator,
						ID: &ast.Identifier{
							NodeType: ast.TypeIdentifier,
							Name:     "dailyClose",
						},
						Init: &ast.CallExpression{
							NodeType: ast.TypeCallExpression,
							Callee: &ast.MemberExpression{
								NodeType: ast.TypeMemberExpression,
								Object: &ast.Identifier{
									NodeType: ast.TypeIdentifier,
									Name:     "request",
								},
								Property: &ast.Identifier{
									NodeType: ast.TypeIdentifier,
									Name:     "security",
								},
							},
							Arguments: []ast.Expression{
								&ast.Literal{NodeType: ast.TypeLiteral, Value: "BTCUSDT"},
								&ast.Literal{NodeType: ast.TypeLiteral, Value: "1D"},
								&ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
							},
						},
					},
				},
			},
		},
	}

	injection, err := AnalyzeAndGeneratePrefetch(program)
	if err != nil {
		t.Fatalf("AnalyzeAndGeneratePrefetch failed: %v", err)
	}

	if injection.PrefetchCode == "" {
		t.Error("Expected non-empty prefetch code")
	}

	requiredStrings := []string{
		"fetcher.Fetch",
		"context.New",
		"securityContexts",
		"BTCUSDT",
		"1D",
	}

	for _, required := range requiredStrings {
		if !contains(injection.PrefetchCode, required) {
			t.Errorf("Prefetch code missing required string: %q", required)
		}
	}

	if len(injection.ImportPaths) != 3 {
		t.Errorf("Expected 3 imports, got %d", len(injection.ImportPaths))
	}

	expectedImports := []string{
		"github.com/quant5-lab/runner/datafetcher",
		"github.com/quant5-lab/runner/security",
		"github.com/quant5-lab/runner/ast",
	}
	for _, expected := range expectedImports {
		found := false
		for _, imp := range injection.ImportPaths {
			if imp == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing import: %q", expected)
		}
	}
}

func TestGenerateSecurityLookup(t *testing.T) {

	secCall := &security.SecurityCall{
		Symbol:     "TEST",
		Timeframe:  "1h",
		ExprName:   "unnamed",
		Expression: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
	}

	code := GenerateSecurityLookup(secCall, "testVar")

	requiredStrings := []string{
		"testVar_values",
		"securityCache.GetExpression",
		"TEST",
		"1h",
		"ctx.BarIndex",
		"math.NaN()",
	}

	for _, required := range requiredStrings {
		if !contains(code, required) {
			t.Errorf("Lookup code missing required string: %q", required)
		}
	}
}

func TestInjectSecurityCode_NoSecurityCalls(t *testing.T) {
	originalCode := &StrategyCode{
		FunctionBody: "\t// Original strategy code\n",
		StrategyName: "Test Strategy",
	}

	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body:     []ast.Node{},
	}

	injectedCode, err := InjectSecurityCode(originalCode, program)
	if err != nil {
		t.Fatalf("InjectSecurityCode failed: %v", err)
	}

	if injectedCode.FunctionBody != originalCode.FunctionBody {
		t.Error("Function body should remain unchanged when no security() calls")
	}
}

func TestInjectSecurityCode_WithSecurityCall(t *testing.T) {
	originalCode := &StrategyCode{
		FunctionBody: "\t// Original strategy code\n",
		StrategyName: "Test Strategy",
	}

	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.VariableDeclaration{
				NodeType: ast.TypeVariableDeclaration,
				Kind:     "var",
				Declarations: []ast.VariableDeclarator{
					{
						NodeType: ast.TypeVariableDeclarator,
						ID:       &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "dailyClose"},
						Init: &ast.CallExpression{
							NodeType: ast.TypeCallExpression,
							Callee: &ast.MemberExpression{
								NodeType: ast.TypeMemberExpression,
								Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "request"},
								Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "security"},
							},
							Arguments: []ast.Expression{
								&ast.Literal{NodeType: ast.TypeLiteral, Value: "BTCUSDT"},
								&ast.Literal{NodeType: ast.TypeLiteral, Value: "1D"},
								&ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
							},
						},
					},
				},
			},
		},
	}

	injectedCode, err := InjectSecurityCode(originalCode, program)
	if err != nil {
		t.Fatalf("InjectSecurityCode failed: %v", err)
	}

	if !contains(injectedCode.FunctionBody, "fetcher.Fetch") {
		t.Error("Expected security prefetch code to be injected")
	}

	if !contains(injectedCode.FunctionBody, "// Original strategy code") {
		t.Error("Original strategy code should be preserved")
	}
}

func TestAnalyzeAndGeneratePrefetch_RuntimeSymbolResolution(t *testing.T) {

	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.VariableDeclaration{
				NodeType: ast.TypeVariableDeclaration,
				Kind:     "var",
				Declarations: []ast.VariableDeclarator{
					{
						NodeType: ast.TypeVariableDeclarator,
						ID:       &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "dailyClose"},
						Init: &ast.CallExpression{
							NodeType: ast.TypeCallExpression,
							Callee: &ast.MemberExpression{
								NodeType: ast.TypeMemberExpression,
								Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "request"},
								Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "security"},
							},
							Arguments: []ast.Expression{
								&ast.MemberExpression{
									NodeType: ast.TypeMemberExpression,
									Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "syminfo"},
									Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "tickerid"},
								},
								&ast.Literal{NodeType: ast.TypeLiteral, Value: "1D"},
								&ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
							},
						},
					},
				},
			},
		},
	}

	injection, err := AnalyzeAndGeneratePrefetch(program)
	if err != nil {
		t.Fatalf("AnalyzeAndGeneratePrefetch failed: %v", err)
	}

	if !contains(injection.PrefetchCode, "ctx.Symbol") {
		t.Error("Expected ctx.Symbol runtime reference for syminfo.tickerid")
	}

	if !contains(injection.PrefetchCode, "fetcher.Fetch(ctx.Symbol") {
		t.Error("Expected fetcher.Fetch to use ctx.Symbol for runtime symbol")
	}
}

func TestAnalyzeAndGeneratePrefetch_RuntimeTimeframeResolution(t *testing.T) {

	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.VariableDeclaration{
				NodeType: ast.TypeVariableDeclaration,
				Kind:     "var",
				Declarations: []ast.VariableDeclarator{
					{
						NodeType: ast.TypeVariableDeclarator,
						ID:       &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "currentClose"},
						Init: &ast.CallExpression{
							NodeType: ast.TypeCallExpression,
							Callee: &ast.MemberExpression{
								NodeType: ast.TypeMemberExpression,
								Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "request"},
								Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "security"},
							},
							Arguments: []ast.Expression{
								&ast.Literal{NodeType: ast.TypeLiteral, Value: "BTCUSDT"},
								&ast.MemberExpression{
									NodeType: ast.TypeMemberExpression,
									Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "timeframe"},
									Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "period"},
								},
								&ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
							},
						},
					},
				},
			},
		},
	}

	injection, err := AnalyzeAndGeneratePrefetch(program)
	if err != nil {
		t.Fatalf("AnalyzeAndGeneratePrefetch failed: %v", err)
	}

	if !contains(injection.PrefetchCode, "ctx.Timeframe") {
		t.Error("Expected ctx.Timeframe runtime reference for timeframe.period")
	}

	if !contains(injection.PrefetchCode, "fetcher.Fetch") && !contains(injection.PrefetchCode, "ctx.Timeframe") {
		t.Error("Expected fetcher.Fetch to use ctx.Timeframe for runtime timeframe")
	}
}

func TestAnalyzeAndGeneratePrefetch_CombinedRuntimeResolution(t *testing.T) {

	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.VariableDeclaration{
				NodeType: ast.TypeVariableDeclaration,
				Kind:     "var",
				Declarations: []ast.VariableDeclarator{
					{
						NodeType: ast.TypeVariableDeclarator,
						ID:       &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "dynamicClose"},
						Init: &ast.CallExpression{
							NodeType: ast.TypeCallExpression,
							Callee: &ast.MemberExpression{
								NodeType: ast.TypeMemberExpression,
								Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "request"},
								Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "security"},
							},
							Arguments: []ast.Expression{
								&ast.MemberExpression{
									NodeType: ast.TypeMemberExpression,
									Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "syminfo"},
									Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "tickerid"},
								},
								&ast.MemberExpression{
									NodeType: ast.TypeMemberExpression,
									Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "timeframe"},
									Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "period"},
								},
								&ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
							},
						},
					},
				},
			},
		},
	}

	injection, err := AnalyzeAndGeneratePrefetch(program)
	if err != nil {
		t.Fatalf("AnalyzeAndGeneratePrefetch failed: %v", err)
	}

	if !contains(injection.PrefetchCode, "ctx.Symbol") {
		t.Error("Expected ctx.Symbol runtime reference")
	}
	if !contains(injection.PrefetchCode, "ctx.Timeframe") {
		t.Error("Expected ctx.Timeframe runtime reference")
	}

	if !contains(injection.PrefetchCode, "fetcher.Fetch(ctx.Symbol, ctx.Timeframe") {
		t.Error("Expected fetcher.Fetch to use both ctx.Symbol and ctx.Timeframe")
	}

	if !contains(injection.PrefetchCode, "context.New(ctx.Symbol, ctx.Timeframe") {
		t.Error("Expected context.New to use both runtime references")
	}
}

func TestAnalyzeAndGeneratePrefetch_MixedLiteralAndRuntime(t *testing.T) {

	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.VariableDeclaration{
				NodeType: ast.TypeVariableDeclaration,
				Kind:     "var",
				Declarations: []ast.VariableDeclarator{
					{
						NodeType: ast.TypeVariableDeclarator,
						ID:       &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "dailyClose"},
						Init: &ast.CallExpression{
							NodeType: ast.TypeCallExpression,
							Callee: &ast.MemberExpression{
								NodeType: ast.TypeMemberExpression,
								Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "request"},
								Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "security"},
							},
							Arguments: []ast.Expression{
								&ast.Literal{NodeType: ast.TypeLiteral, Value: "BTCUSDT"},
								&ast.Literal{NodeType: ast.TypeLiteral, Value: "1D"},
								&ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
							},
						},
					},
					{
						NodeType: ast.TypeVariableDeclarator,
						ID:       &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "currentClose"},
						Init: &ast.CallExpression{
							NodeType: ast.TypeCallExpression,
							Callee: &ast.MemberExpression{
								NodeType: ast.TypeMemberExpression,
								Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "request"},
								Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "security"},
							},
							Arguments: []ast.Expression{
								&ast.MemberExpression{
									NodeType: ast.TypeMemberExpression,
									Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "syminfo"},
									Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "tickerid"},
								},
								&ast.MemberExpression{
									NodeType: ast.TypeMemberExpression,
									Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "timeframe"},
									Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "period"},
								},
								&ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
							},
						},
					},
				},
			},
		},
	}

	injection, err := AnalyzeAndGeneratePrefetch(program)
	if err != nil {
		t.Fatalf("AnalyzeAndGeneratePrefetch failed: %v", err)
	}

	if !contains(injection.PrefetchCode, "BTCUSDT") {
		t.Error("Expected literal symbol BTCUSDT")
	}
	if !contains(injection.PrefetchCode, "ctx.Symbol") {
		t.Error("Expected runtime ctx.Symbol")
	}
	if !contains(injection.PrefetchCode, "1D") {
		t.Error("Expected literal timeframe 1D")
	}
	if !contains(injection.PrefetchCode, "ctx.Timeframe") {
		t.Error("Expected runtime ctx.Timeframe")
	}
}

func TestAnalyzeAndGeneratePrefetch_RuntimeDeduplication(t *testing.T) {

	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.VariableDeclaration{
				NodeType: ast.TypeVariableDeclaration,
				Kind:     "var",
				Declarations: []ast.VariableDeclarator{
					{
						NodeType: ast.TypeVariableDeclarator,
						ID:       &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "currentClose"},
						Init: &ast.CallExpression{
							NodeType: ast.TypeCallExpression,
							Callee: &ast.MemberExpression{
								NodeType: ast.TypeMemberExpression,
								Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "request"},
								Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "security"},
							},
							Arguments: []ast.Expression{
								&ast.MemberExpression{
									NodeType: ast.TypeMemberExpression,
									Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "syminfo"},
									Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "tickerid"},
								},
								&ast.MemberExpression{
									NodeType: ast.TypeMemberExpression,
									Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "timeframe"},
									Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "period"},
								},
								&ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
							},
						},
					},
					{
						NodeType: ast.TypeVariableDeclarator,
						ID:       &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "currentHigh"},
						Init: &ast.CallExpression{
							NodeType: ast.TypeCallExpression,
							Callee: &ast.MemberExpression{
								NodeType: ast.TypeMemberExpression,
								Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "request"},
								Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "security"},
							},
							Arguments: []ast.Expression{
								&ast.MemberExpression{
									NodeType: ast.TypeMemberExpression,
									Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "syminfo"},
									Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "tickerid"},
								},
								&ast.MemberExpression{
									NodeType: ast.TypeMemberExpression,
									Object:   &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "timeframe"},
									Property: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "period"},
								},
								&ast.Identifier{NodeType: ast.TypeIdentifier, Name: "high"},
							},
						},
					},
				},
			},
		},
	}

	injection, err := AnalyzeAndGeneratePrefetch(program)
	if err != nil {
		t.Fatalf("AnalyzeAndGeneratePrefetch failed: %v", err)
	}

	fetchCount := countOccurrences(injection.PrefetchCode, "fetcher.Fetch")
	if fetchCount != 1 {
		t.Errorf("Expected 1 fetcher.Fetch call (deduplicated), got %d", fetchCount)
	}

	contextCount := countOccurrences(injection.PrefetchCode, "context.New")
	if contextCount != 1 {
		t.Errorf("Expected 1 context.New call (deduplicated), got %d", contextCount)
	}
}

func countOccurrences(haystack, needle string) int {
	count := 0
	offset := 0
	for {
		idx := strings.Index(haystack[offset:], needle)
		if idx == -1 {
			break
		}
		count++
		offset += idx + len(needle)
	}
	return count
}
