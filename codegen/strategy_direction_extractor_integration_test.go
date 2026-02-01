package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestStrategyDirectionExtractor_GeneratorIntegration verifies end-to-end extraction through generator */
func TestStrategyDirectionExtractor_GeneratorIntegration(t *testing.T) {
	tests := []struct {
		name         string
		directionArg ast.Expression
		wantContains string
		pineVersion  string
	}{
		{
			name: "Pine v5 strategy.long integration",
			directionArg: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "long"},
			},
			wantContains: "strategy.Long",
			pineVersion:  "v5",
		},
		{
			name: "Pine v5 strategy.short integration",
			directionArg: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "short"},
			},
			wantContains: "strategy.Short",
			pineVersion:  "v5",
		},
		{
			name:         "Pine v4 true literal integration",
			directionArg: &ast.Literal{Value: true},
			wantContains: "strategy.Long",
			pineVersion:  "v4",
		},
		{
			name:         "Pine v4 false literal integration",
			directionArg: &ast.Literal{Value: false},
			wantContains: "strategy.Short",
			pineVersion:  "v4",
		},
		{
			name:         "Pine v4 true identifier integration",
			directionArg: &ast.Identifier{Name: "true"},
			wantContains: "strategy.Long",
			pineVersion:  "v4",
		},
		{
			name:         "Pine v4 false identifier integration",
			directionArg: &ast.Identifier{Name: "false"},
			wantContains: "strategy.Short",
			pineVersion:  "v4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			handler := NewStrategyActionHandler()

			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "TestEntry"},
					tt.directionArg,
				},
			}

			code, err := handler.GenerateCode(g, call)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tt.pineVersion, err)
			}

			if !strings.Contains(code, tt.wantContains) {
				t.Errorf("%s: expected code to contain %q\nGenerated:\n%s",
					tt.pineVersion, tt.wantContains, code)
			}

			if !strings.Contains(code, "strat.Entry") {
				t.Errorf("%s: expected Entry call in generated code:\n%s", tt.pineVersion, code)
			}
		})
	}
}

/* TestStrategyDirectionExtractor_MixedVersionScenarios verifies handling of mixed syntax */
func TestStrategyDirectionExtractor_MixedVersionScenarios(t *testing.T) {
	g := newTestGenerator()
	handler := NewStrategyActionHandler()

	/* Scenario: Strategy uses both v4 and v5 syntax in different entries */
	tests := []struct {
		name         string
		entryID      string
		directionArg ast.Expression
		wantDir      string
	}{
		{
			name:    "First entry uses v5 long",
			entryID: "entry1",
			directionArg: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "long"},
			},
			wantDir: "strategy.Long",
		},
		{
			name:         "Second entry uses v4 false",
			entryID:      "entry2",
			directionArg: &ast.Literal{Value: false},
			wantDir:      "strategy.Short",
		},
		{
			name:         "Third entry uses v4 true identifier",
			entryID:      "entry3",
			directionArg: &ast.Identifier{Name: "true"},
			wantDir:      "strategy.Long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: tt.entryID},
					tt.directionArg,
				},
			}

			code, err := handler.GenerateCode(g, call)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(code, tt.wantDir) {
				t.Errorf("expected direction %q in:\n%s", tt.wantDir, code)
			}
		})
	}
}

/* TestStrategyDirectionExtractor_LazyInitialization verifies generator field setup */
func TestStrategyDirectionExtractor_LazyInitialization(t *testing.T) {
	/* Generator without explicit direction extractor should auto-initialize */
	g := &generator{
		strategyConfig: &StrategyConfig{
			DefaultQtyType:  "fixed",
			DefaultQtyValue: 1,
		},
	}

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Identifier{Name: "entry"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "Test"},
			&ast.Literal{Value: true},
		},
	}

	handler := NewStrategyActionHandler()
	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Fatalf("lazy initialization failed: %v", err)
	}

	if !strings.Contains(code, "strategy.Long") {
		t.Errorf("lazy-initialized extractor should handle v4 boolean: %s", code)
	}
}

/* TestStrategyDirectionExtractor_ErrorResilience verifies graceful degradation */
func TestStrategyDirectionExtractor_ErrorResilience(t *testing.T) {
	g := newTestGenerator()
	handler := NewStrategyActionHandler()

	tests := []struct {
		name         string
		directionArg ast.Expression
		wantDefault  string
	}{
		{
			name:         "invalid direction falls back to Long",
			directionArg: &ast.Literal{Value: "invalid"},
			wantDefault:  "strategy.Long",
		},
		{
			name:         "numeric direction falls back to Long",
			directionArg: &ast.Literal{Value: 42},
			wantDefault:  "strategy.Long",
		},
		{
			name: "complex expression falls back to Long",
			directionArg: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "x"},
				Operator: "+",
				Right:    &ast.Literal{Value: 1},
			},
			wantDefault: "strategy.Long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "entry"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "TestEntry"},
					tt.directionArg,
				},
			}

			code, err := handler.GenerateCode(g, call)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(code, tt.wantDefault) {
				t.Errorf("expected fallback to %q\nGenerated:\n%s",
					tt.wantDefault, code)
			}
		})
	}
}

/* TestStrategyDirectionExtractor_BackwardCompatibility verifies existing code unchanged */
func TestStrategyDirectionExtractor_BackwardCompatibility(t *testing.T) {
	/* Verify that strategies using old v5 syntax continue to work identically */
	g := newTestGenerator()
	handler := NewStrategyActionHandler()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Identifier{Name: "entry"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "LegacyEntry"},
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "long"},
			},
			&ast.Literal{Value: 10.0},
		},
	}

	code, err := handler.GenerateCode(g, call)
	if err != nil {
		t.Fatalf("backward compatibility broken: %v", err)
	}

	requiredPatterns := []string{
		"strat.Entry",
		`"LegacyEntry"`,
		"strategy.Long",
		"10",
	}

	for _, pattern := range requiredPatterns {
		if !strings.Contains(code, pattern) {
			t.Errorf("backward compatibility: missing pattern %q in:\n%s", pattern, code)
		}
	}
}

/* TestStrategyDirectionExtractor_CodeStructureConsistency verifies output format */
func TestStrategyDirectionExtractor_CodeStructureConsistency(t *testing.T) {
	g := newTestGenerator()
	handler := NewStrategyActionHandler()

	directions := []ast.Expression{
		&ast.MemberExpression{Object: &ast.Identifier{Name: "strategy"}, Property: &ast.Identifier{Name: "long"}},
		&ast.Literal{Value: true},
		&ast.Identifier{Name: "true"},
	}

	codes := make([]string, 0, len(directions))

	for _, dir := range directions {
		call := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "entry"},
			},
			Arguments: []ast.Expression{
				&ast.Literal{Value: "Entry"},
				dir,
			},
		}

		code, err := handler.GenerateCode(g, call)
		if err != nil {
			t.Fatalf("code generation failed: %v", err)
		}
		codes = append(codes, code)
	}

	/* All three versions should produce identical output structure */
	baseStructure := codes[0]
	for i, code := range codes[1:] {
		if code != baseStructure {
			t.Errorf("Code structure inconsistency at index %d:\nExpected:\n%s\nGot:\n%s",
				i+1, baseStructure, code)
		}
	}
}
