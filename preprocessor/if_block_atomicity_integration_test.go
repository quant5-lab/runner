package preprocessor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

/* Integration tests for if block atomicity using actual .pine files */

func parseAndNormalize(t *testing.T, filename string) *parser.Script {
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

func countIfStatementsInScript(script *parser.Script) int {
	count := 0
	var visitStatements func([]*parser.Statement)
	visitStatements = func(statements []*parser.Statement) {
		for _, stmt := range statements {
			if stmt.Core.If != nil {
				count++
				visitStatements(stmt.Core.If.Body)
			}
			if stmt.Core.FunctionDecl != nil && stmt.Core.FunctionDecl.MultiLineBody != nil {
				visitStatements(stmt.Core.FunctionDecl.MultiLineBody)
			}
		}
	}
	visitStatements(script.Statements)
	return count
}

func findIfStatementsInScript(script *parser.Script) []*parser.IfStatement {
	var ifNodes []*parser.IfStatement
	var visitStatements func([]*parser.Statement)
	visitStatements = func(statements []*parser.Statement) {
		for _, stmt := range statements {
			if stmt.Core.If != nil {
				ifNodes = append(ifNodes, stmt.Core.If)
				visitStatements(stmt.Core.If.Body)
			}
			if stmt.Core.FunctionDecl != nil && stmt.Core.FunctionDecl.MultiLineBody != nil {
				visitStatements(stmt.Core.FunctionDecl.MultiLineBody)
			}
		}
	}
	visitStatements(script.Statements)
	return ifNodes
}

func TestIfBlockAtomicity_BasicMultipleAssignments(t *testing.T) {
	result := parseAndNormalize(t, "test-if-atomicity-basic.pine")

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

func TestIfBlockAtomicity_StateMachine(t *testing.T) {
	result := parseAndNormalize(t, "test-if-atomicity-state-machine.pine")

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

func TestIfBlockAtomicity_ComplexConditions(t *testing.T) {
	result := parseAndNormalize(t, "test-if-atomicity-complex.pine")

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

func TestIfBlockAtomicity_NestedBlocks(t *testing.T) {
	result := parseAndNormalize(t, "test-if-atomicity-nested.pine")

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

func TestIfBlockAtomicity_ConsecutiveBlocks(t *testing.T) {
	result := parseAndNormalize(t, "test-if-atomicity-consecutive.pine")

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

func TestIfBlockAtomicity_MixedStatements(t *testing.T) {
	result := parseAndNormalize(t, "test-if-atomicity-mixed.pine")

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

func TestIfBlockAtomicity_NoRegressionOnExistingStrategies(t *testing.T) {
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

func TestIfBlockAtomicity_ConditionEvaluationCount(t *testing.T) {
	pineCode := `
if trigger
    var1 := 1
    var2 := 2
    var3 := 3
`

	normalized := NormalizeIfBlocks(pineCode)

	lines := strings.Split(normalized, "\n")
	ifLines := []string{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "if ") {
			ifLines = append(ifLines, line)
		}
	}

	if len(ifLines) != 1 {
		t.Errorf("Expected 1 'if' line in normalized output, got %d", len(ifLines))
	}
}

func TestIfBlockAtomicity_RealWorldPattern(t *testing.T) {
	pineCode := `
if close_all_avg
    pos_size_long := 0
    pos_size_short := 0
`

	normalized := NormalizeIfBlocks(pineCode)

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	result, err := p.ParseString("test", normalized)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ifNodes := findIfStatementsInScript(result)
	if len(ifNodes) != 1 {
		t.Fatalf("Expected 1 if statement, got %d", len(ifNodes))
	}

	body := ifNodes[0].Body
	if len(body) != 2 {
		t.Errorf("Expected 2 reassignments in single if block, got %d", len(body))
	}

	for i, stmt := range body {
		if stmt.Core.Reassignment == nil {
			t.Errorf("Statement[%d]: expected Reassignment, got Assignment=%v", i, stmt.Core.Assignment != nil)
		}
	}
}
