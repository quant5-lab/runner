package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/parser"
)

/* Integration tests for inline comma-separated statement lists in function bodies */

type inlineStatementListTestCase struct {
	name                   string
	fixture                string
	expectedFunctionName   string
	expectedStatementCount int
}

func TestInlineStatementList_RealWorldPatterns(t *testing.T) {
	t.Parallel()
	testCases := []inlineStatementListTestCase{
		{
			name:                   "variable initialization chain",
			fixture:                "01_variable_initialization.pine",
			expectedFunctionName:   "init",
			expectedStatementCount: 3,
		},
		{
			name:                   "intermediate calculations",
			fixture:                "02_intermediate_calculations.pine",
			expectedFunctionName:   "momentum",
			expectedStatementCount: 3,
		},
		{
			name:                   "state transformation",
			fixture:                "03_state_transformation.pine",
			expectedFunctionName:   "normalize",
			expectedStatementCount: 3,
		},
		{
			name:                   "multi-step TA calculation",
			fixture:                "04_multi_step_ta.pine",
			expectedFunctionName:   "customRsi",
			expectedStatementCount: 6,
		},
		{
			name:                   "complex expression chain",
			fixture:                "05_complex_expression.pine",
			expectedFunctionName:   "signal",
			expectedStatementCount: 6,
		},
		{
			name:                   "reassignment pattern",
			fixture:                "06_reassignment.pine",
			expectedFunctionName:   "accumulate",
			expectedStatementCount: 4,
		},
		{
			name:                   "ternary final expression",
			fixture:                "07_ternary_final.pine",
			expectedFunctionName:   "adaptive",
			expectedStatementCount: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fixturePath := filepath.Join("../fixtures/inline_statement_list", tc.fixture)
			content, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatalf("Failed to read fixture %s: %v", tc.fixture, err)
			}

			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseString(tc.fixture, string(content))
			if err != nil {
				t.Fatalf("Parse failed for %s: %v", tc.fixture, err)
			}

			if script == nil {
				t.Fatal("Script AST should not be nil")
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("ESTree conversion failed: %v", err)
			}

			arrowFunc := extractArrowFunction(t, program, tc.expectedFunctionName)
			if arrowFunc == nil {
				t.Fatalf("Arrow function %s not found in AST", tc.expectedFunctionName)
			}

			actualCount := len(arrowFunc.Body)
			if actualCount != tc.expectedStatementCount {
				t.Errorf("Expected %d statements in function body, got %d",
					tc.expectedStatementCount, actualCount)
			}

			verifyFinalExpressionStatement(t, arrowFunc)
		})
	}
}

func TestInlineStatementList_AllFixturesParseCleanly(t *testing.T) {
	t.Parallel()
	fixturesDir := "../fixtures/inline_statement_list"

	entries, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Fatalf("Failed to read fixtures directory: %v", err)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	successCount := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".pine" {
			continue
		}

		filePath := filepath.Join(fixturesDir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Failed to read fixture %s: %v", entry.Name(), err)
			continue
		}

		ast, err := p.ParseString(entry.Name(), string(content))
		if err != nil {
			t.Errorf("Parse failed for %s: %v", entry.Name(), err)
			continue
		}

		if ast == nil {
			t.Errorf("AST is nil for %s", entry.Name())
			continue
		}

		successCount++
	}

	if successCount == 0 {
		t.Fatal("No fixtures parsed successfully")
	}

	t.Logf("Successfully parsed %d inline statement list fixtures", successCount)
}

func TestInlineStatementList_JSONSerializationStability(t *testing.T) {
	t.Parallel()
	fixturePath := "../fixtures/inline_statement_list/02_intermediate_calculations.pine"
	content, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseString("test.pine", string(content))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := parser.NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("ESTree conversion failed: %v", err)
	}

	jsonBytes, err := converter.ToJSON(program)
	if err != nil {
		t.Fatalf("JSON serialization failed: %v", err)
	}

	var parsed map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsed)
	if err != nil {
		t.Fatalf("JSON deserialization failed: %v", err)
	}

	if len(jsonBytes) == 0 {
		t.Fatal("Generated JSON should not be empty")
	}
}

func extractArrowFunction(t *testing.T, program *ast.Program, funcName string) *ast.ArrowFunctionExpression {
	for _, stmt := range program.Body {
		varDecl, ok := stmt.(*ast.VariableDeclaration)
		if !ok {
			continue
		}

		for _, decl := range varDecl.Declarations {
			idNode, ok := decl.ID.(*ast.Identifier)
			if !ok {
				continue
			}
			if idNode.Name == funcName {
				arrowFunc, ok := decl.Init.(*ast.ArrowFunctionExpression)
				if !ok {
					t.Fatalf("Expected ArrowFunctionExpression for %s, got %T", funcName, decl.Init)
				}
				return arrowFunc
			}
		}
	}
	return nil
}

func verifyFinalExpressionStatement(t *testing.T, arrowFunc *ast.ArrowFunctionExpression) {
	if len(arrowFunc.Body) == 0 {
		t.Fatal("Arrow function body should not be empty")
	}

	lastStmt := arrowFunc.Body[len(arrowFunc.Body)-1]
	_, ok := lastStmt.(*ast.ExpressionStatement)
	if !ok {
		t.Errorf("Final statement should be ExpressionStatement, got %T", lastStmt)
	}
}
