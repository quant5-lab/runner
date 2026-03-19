package preprocessor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

func parsePineFile(t *testing.T, filename string) *parser.Script {
	t.Helper()

	filePath := filepath.Join("..", "e2e", "fixtures", "strategies", filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", filename, err)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	result, err := p.ParseString(filename, string(content))
	if err != nil {
		t.Fatalf("Failed to parse %s: %v", filename, err)
	}

	return result
}

func collectIfStatements(statements []*parser.Statement) []*parser.IfStatement {
	var result []*parser.IfStatement
	for _, stmt := range statements {
		if stmt.Core.If != nil {
			result = append(result, stmt.Core.If)
			result = append(result, collectIfStatements(stmt.Core.If.Body)...)
		}
		if stmt.Core.FunctionDecl != nil && stmt.Core.FunctionDecl.MultiLineBody != nil {
			result = append(result, collectIfStatements(stmt.Core.FunctionDecl.MultiLineBody)...)
		}
	}
	return result
}

func findIfStatementsInScript(script *parser.Script) []*parser.IfStatement {
	return collectIfStatements(script.Statements)
}

func countIfStatementsInScript(script *parser.Script) int {
	return len(findIfStatementsInScript(script))
}

func TestIfBlockParsing_BasicMultipleAssignments(t *testing.T) {
	result := parsePineFile(t, "test-if-atomicity-basic.pine")

	ifCount := countIfStatementsInScript(result)
	if ifCount != 1 {
		t.Errorf("Expected 1 if statement, got %d", ifCount)
	}

	ifNodes := findIfStatementsInScript(result)
	if len(ifNodes) != 1 {
		t.Fatalf("Expected 1 IfStatement node, got %d", len(ifNodes))
	}

	body := ifNodes[0].Body
	if len(body) != 3 {
		t.Errorf("Expected 3 statements in if body, got %d", len(body))
	}

	for i, stmt := range body {
		if stmt.Core.Reassignment == nil {
			t.Errorf("Statement[%d]: expected Reassignment, got Assignment=%v", i, stmt.Core.Assignment != nil)
		}
	}
}

func TestIfBlockParsing_StateMachine(t *testing.T) {
	result := parsePineFile(t, "test-if-atomicity-state-machine.pine")

	ifNodes := findIfStatementsInScript(result)

	if len(ifNodes) != 2 {
		t.Fatalf("Expected 2 if statements (entry + exit), got %d", len(ifNodes))
	}

	entryBlock := ifNodes[0]
	if len(entryBlock.Body) != 3 {
		t.Errorf("Entry block: expected 3 assignments, got %d", len(entryBlock.Body))
	}

	exitBlock := ifNodes[1]
	if len(exitBlock.Body) != 3 {
		t.Errorf("Exit block: expected 3 assignments, got %d", len(exitBlock.Body))
	}

	for blockIdx, block := range ifNodes {
		for stmtIdx, stmt := range block.Body {
			if stmt.Core.Reassignment == nil {
				t.Errorf("Block[%d] Statement[%d]: expected Reassignment", blockIdx, stmtIdx)
			}
		}
	}
}

func TestIfBlockParsing_ComplexConditions(t *testing.T) {
	result := parsePineFile(t, "test-if-atomicity-complex.pine")

	ifNodes := findIfStatementsInScript(result)

	if len(ifNodes) != 2 {
		t.Fatalf("Expected 2 if statements, got %d", len(ifNodes))
	}

	complexBlock := ifNodes[0]
	if len(complexBlock.Body) < 4 {
		t.Errorf("Complex condition block: expected at least 4 assignments, got %d",
			len(complexBlock.Body))
	}

	if complexBlock.Condition == nil {
		t.Error("If statement should have condition")
	}

	resetBlock := ifNodes[1]
	if len(resetBlock.Body) != 3 {
		t.Errorf("Reset block: expected 3 assignments, got %d", len(resetBlock.Body))
	}
}

func TestIfBlockParsing_NestedBlocks(t *testing.T) {
	result := parsePineFile(t, "test-if-atomicity-nested.pine")

	ifNodes := findIfStatementsInScript(result)

	if len(ifNodes) != 4 {
		t.Fatalf("Expected 4 if statements, got %d", len(ifNodes))
	}

	outerBlock := ifNodes[0]
	if len(outerBlock.Body) < 2 {
		t.Errorf("Outer block: expected at least 2 statements, got %d", len(outerBlock.Body))
	}

	hasReassignment := false
	hasNestedIf := false
	for _, stmt := range outerBlock.Body {
		if stmt.Core.Reassignment != nil {
			hasReassignment = true
		}
		if stmt.Core.If != nil {
			hasNestedIf = true
		}
	}

	if !hasReassignment {
		t.Error("Outer block: should contain at least one reassignment")
	}
	if !hasNestedIf {
		t.Error("Outer block: should contain nested if statement")
	}

	resetBlock := ifNodes[3]
	if len(resetBlock.Body) != 3 {
		t.Errorf("Reset block: expected 3 assignments, got %d", len(resetBlock.Body))
	}
}

func TestIfBlockParsing_ConsecutiveBlocks(t *testing.T) {
	result := parsePineFile(t, "test-if-atomicity-consecutive.pine")

	ifNodes := findIfStatementsInScript(result)

	if len(ifNodes) != 4 {
		t.Fatalf("Expected 4 if statements, got %d", len(ifNodes))
	}

	for i := 0; i < 3; i++ {
		if len(ifNodes[i].Body) != 2 {
			t.Errorf("If block[%d]: expected 2 assignments, got %d", i, len(ifNodes[i].Body))
		}

		for stmtIdx, stmt := range ifNodes[i].Body {
			if stmt.Core.Reassignment == nil {
				t.Errorf("Block[%d] Statement[%d]: expected Reassignment", i, stmtIdx)
			}
		}
	}

	resetBlock := ifNodes[3]
	if len(resetBlock.Body) != 6 {
		t.Errorf("Reset block: expected 6 assignments, got %d", len(resetBlock.Body))
	}
}

func TestIfBlockParsing_MixedStatements(t *testing.T) {
	result := parsePineFile(t, "test-if-atomicity-mixed.pine")

	ifNodes := findIfStatementsInScript(result)

	if len(ifNodes) != 2 {
		t.Fatalf("Expected 2 if statements, got %d", len(ifNodes))
	}

	entryBody := ifNodes[0].Body
	if len(entryBody) != 4 {
		t.Errorf("Entry block: expected 4 assignments, got %d", len(entryBody))
	}

	exitBody := ifNodes[1].Body
	if len(exitBody) != 3 {
		t.Errorf("Exit block: expected 3 assignments, got %d", len(exitBody))
	}

	for blockIdx, block := range ifNodes {
		for stmtIdx, stmt := range block.Body {
			if stmt.Core.Reassignment == nil {
				t.Errorf("Block[%d] Statement[%d]: expected Reassignment", blockIdx, stmtIdx)
			}
		}
	}
}

func TestIfBlockParsing_NoRegressionOnExistingStrategies(t *testing.T) {
	strategies := []struct {
		filename string
		minIfs   int
	}{
		{"bb-strategy-9-rus.pine", 1},
		{"daily-lines-simple.pine", 0},
	}

	for _, tc := range strategies {
		t.Run(tc.filename, func(t *testing.T) {
			filePath := filepath.Join("..", "strategies", tc.filename)
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Skipf("Cannot read %s: %v", tc.filename, err)
			}

			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			result, err := p.ParseString(tc.filename, string(content))
			if err != nil {
				t.Fatalf("Parse failed for %s: %v", tc.filename, err)
			}

			ifCount := countIfStatementsInScript(result)
			if ifCount < tc.minIfs {
				t.Errorf("%s: expected at least %d if statements, got %d", tc.filename, tc.minIfs, ifCount)
			}
		})
	}
}
