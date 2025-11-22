package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/security"
)

func TestAnalyzeAndGeneratePrefetch_NoSecurityCalls(t *testing.T) {
	/* Program without security() calls */
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
	/* Program with request.security() call */
	program := &ast.Program{
		NodeType: ast.TypeProgram,
		Body: []ast.Node{
			&ast.VariableDeclaration{
				NodeType: ast.TypeVariableDeclaration,
				Kind:     "var",
				Declarations: []ast.VariableDeclarator{
					{
						NodeType: ast.TypeVariableDeclarator,
						ID: ast.Identifier{
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

	/* Verify prefetch code contains key elements */
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

	/* Verify imports - only datafetcher needed now */
	if len(injection.ImportPaths) != 1 {
		t.Errorf("Expected 1 import, got %d", len(injection.ImportPaths))
	}

	expectedImport := "github.com/quant5-lab/runner/datafetcher"
	found := false
	for _, imp := range injection.ImportPaths {
		if imp == expectedImport {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Missing import: %q", expectedImport)
	}
}

func TestGenerateSecurityLookup(t *testing.T) {
	/* Create SecurityCall matching analyzer output */
	secCall := &security.SecurityCall{
		Symbol:     "TEST",
		Timeframe:  "1h",
		ExprName:   "unnamed",
		Expression: &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "close"},
	}

	code := GenerateSecurityLookup(secCall, "testVar")

	/* Verify generated lookup code */
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
		FunctionBody:       "\t// Original strategy code\n",
		StrategyName:       "Test Strategy",
		NeedsSeriesPreCalc: false,
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
		FunctionBody:       "\t// Original strategy code\n",
		StrategyName:       "Test Strategy",
		NeedsSeriesPreCalc: false,
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
						ID:       ast.Identifier{NodeType: ast.TypeIdentifier, Name: "dailyClose"},
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

	/* Verify prefetch code was injected */
	if !contains(injectedCode.FunctionBody, "fetcher.Fetch") {
		t.Error("Expected security prefetch code to be injected")
	}

	/* Verify original code is still present */
	if !contains(injectedCode.FunctionBody, "// Original strategy code") {
		t.Error("Original strategy code should be preserved")
	}

	/* Verify NeedsSeriesPreCalc flag set */
	if !injectedCode.NeedsSeriesPreCalc {
		t.Error("Expected NeedsSeriesPreCalc to be true after security injection")
	}
}
