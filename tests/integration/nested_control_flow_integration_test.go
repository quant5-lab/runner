package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

func TestNestedControlFlowIntegration(t *testing.T) {
	t.Parallel()
	fixturePath := filepath.Join("..", "fixtures", "blockers", "test-nested-control-flow.pine")

	content, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	script, err := p.ParseBytes(fixturePath, content)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(script.Statements) == 0 {
		t.Fatal("No statements parsed from comprehensive nested control flow test")
	}

	// Verify script parses successfully
	t.Logf("Successfully parsed %d statements from nested control flow fixture", len(script.Statements))

	// Convert to ESTree to ensure AST is valid
	converter := parser.NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("AST conversion failed: %v", err)
	}

	if len(program.Body) == 0 {
		t.Fatal("ESTree program body is empty")
	}

	t.Logf("Successfully converted to ESTree with %d top-level nodes", len(program.Body))
}
