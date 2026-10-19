package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestStrategyDirectionExtractor_GeneratorIntegration(t *testing.T) {
	type directionCase struct {
		name    string
		expr    ast.Expression
		wantDir string
	}

	directions := []directionCase{
		{
			"strategy.long (v5)",
			&ast.MemberExpression{Object: &ast.Identifier{Name: "strategy"}, Property: &ast.Identifier{Name: "long"}},
			"strategy.Long",
		},
		{
			"strategy.short (v5)",
			&ast.MemberExpression{Object: &ast.Identifier{Name: "strategy"}, Property: &ast.Identifier{Name: "short"}},
			"strategy.Short",
		},
		{"bool true (v4)", &ast.Literal{Value: true}, "strategy.Long"},
		{"bool false (v4)", &ast.Literal{Value: false}, "strategy.Short"},
		{"identifier true (v4)", &ast.Identifier{Name: "true"}, "strategy.Long"},
		{"identifier false (v4)", &ast.Identifier{Name: "false"}, "strategy.Short"},
	}

	type argStyle struct {
		name      string
		buildArgs func(id string, dirExpr ast.Expression) []ast.Expression
	}

	argStyles := []argStyle{
		{
			name: "positional",
			buildArgs: func(id string, dirExpr ast.Expression) []ast.Expression {
				return []ast.Expression{&ast.Literal{Value: id}, dirExpr}
			},
		},
		{
			name: "named-arg",
			buildArgs: func(id string, dirExpr ast.Expression) []ast.Expression {
				return []ast.Expression{
					&ast.Literal{Value: id},
					&ast.ObjectExpression{
						NodeType: ast.TypeObjectExpression,
						Properties: []ast.Property{
							{Key: &ast.Identifier{Name: "long"}, Value: dirExpr},
						},
					},
				}
			},
		},
	}

	type callSite struct {
		name        string
		callee      string
		id          string
		wantRuntime string
	}

	callSites := []callSite{
		{"strategy.entry", "entry", "E", "strat.Entry"},
		{"strategy.order", "order", "O", "strat.Order"},
	}

	handler := NewStrategyActionHandler()

	for _, site := range callSites {
		for _, style := range argStyles {
			for _, dir := range directions {
				t.Run(site.name+"/"+style.name+"/"+dir.name, func(t *testing.T) {
					g := newTestGenerator()
					call := &ast.CallExpression{
						Callee: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "strategy"},
							Property: &ast.Identifier{Name: site.callee},
						},
						Arguments: style.buildArgs(site.id, dir.expr),
					}
					code, err := handler.GenerateCode(g, call)
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					if !strings.Contains(code, dir.wantDir) {
						t.Errorf("want %q in generated code:\n%s", dir.wantDir, code)
					}
					if !strings.Contains(code, site.wantRuntime) {
						t.Errorf("want %q in generated code:\n%s", site.wantRuntime, code)
					}
				})
			}
		}
	}
}

func TestStrategyDirectionExtractor_AllowEntryIn_DirectionConstants(t *testing.T) {
	cases := []struct {
		prop    string
		wantDir string
	}{
		{"long", "strategy.DirectionLong"},
		{"short", "strategy.DirectionShort"},
		{"all", "strategy.DirectionAll"},
	}

	for _, tc := range cases {
		t.Run("strategy.direction."+tc.prop, func(t *testing.T) {
			g := newTestGenerator()
			handler := NewStrategyActionHandler()

			directionExpr := &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "direction"},
				},
				Property: &ast.Identifier{Name: tc.prop},
			}
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "risk"},
					},
					Property: &ast.Identifier{Name: "allow_entry_in"},
				},
				Arguments: []ast.Expression{directionExpr},
			}

			code, err := handler.GenerateCode(g, call)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(code, tc.wantDir) {
				t.Errorf("want %q in generated code:\n%s", tc.wantDir, code)
			}
			if !strings.Contains(code, "SetAllowedDirection") {
				t.Errorf("want SetAllowedDirection in generated code:\n%s", code)
			}
		})
	}
}

// TestStrategyDirectionExtractor_MixedVersionScenarios guards against direction
// state inadvertently accumulating between calls on a shared generator instance.
func TestStrategyDirectionExtractor_MixedVersionScenarios(t *testing.T) {
	g := newTestGenerator()
	handler := NewStrategyActionHandler()

	tests := []struct {
		name         string
		entryID      string
		directionArg ast.Expression
		wantDir      string
	}{
		{
			name:    "v5 positional long",
			entryID: "entry1",
			directionArg: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "long"},
			},
			wantDir: "strategy.Long",
		},
		{
			name:         "v4 positional false",
			entryID:      "entry2",
			directionArg: &ast.Literal{Value: false},
			wantDir:      "strategy.Short",
		},
		{
			name:         "v4 positional true identifier",
			entryID:      "entry3",
			directionArg: &ast.Identifier{Name: "true"},
			wantDir:      "strategy.Long",
		},
		{
			name:    "v5 named-arg short",
			entryID: "entry4",
			directionArg: &ast.ObjectExpression{
				NodeType: ast.TypeObjectExpression,
				Properties: []ast.Property{
					{Key: &ast.Identifier{Name: "long"}, Value: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "short"},
					}},
				},
			},
			wantDir: "strategy.Short",
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

func TestStrategyDirectionExtractor_LazyInitialization(t *testing.T) {
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

// TestStrategyDirectionExtractor_UnresolvableDirectionFailsGeneration is the
// Anchor-5 integration guard: every strategy call site that takes a direction
// argument must fail code generation when the direction is unresolvable, rather
// than silently emitting a long entry.
func TestStrategyDirectionExtractor_UnresolvableDirectionFailsGeneration(t *testing.T) {
	type callFactory struct {
		callee string
		build  func(dirExpr ast.Expression) *ast.CallExpression
	}

	callFactories := []callFactory{
		{
			callee: "strategy.entry",
			build: func(dirExpr ast.Expression) *ast.CallExpression {
				return &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "entry"},
					},
					Arguments: []ast.Expression{&ast.Literal{Value: "E"}, dirExpr},
				}
			},
		},
		{
			callee: "strategy.order",
			build: func(dirExpr ast.Expression) *ast.CallExpression {
				return &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "order"},
					},
					Arguments: []ast.Expression{&ast.Literal{Value: "O"}, dirExpr},
				}
			},
		},
		{
			callee: "strategy.risk.allow_entry_in",
			build: func(dirExpr ast.Expression) *ast.CallExpression {
				return &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object: &ast.MemberExpression{
							Object:   &ast.Identifier{Name: "strategy"},
							Property: &ast.Identifier{Name: "risk"},
						},
						Property: &ast.Identifier{Name: "allow_entry_in"},
					},
					Arguments: []ast.Expression{dirExpr},
				}
			},
		},
	}

	unresolvable := []struct {
		name string
		expr ast.Expression
	}{
		{"string literal", &ast.Literal{Value: "invalid"}},
		{"numeric literal", &ast.Literal{Value: 42.0}},
		{"arithmetic expression", &ast.BinaryExpression{
			Left: &ast.Identifier{Name: "x"}, Operator: "+", Right: &ast.Literal{Value: 1.0},
		}},
		{"call expression", &ast.CallExpression{
			Callee: &ast.Identifier{Name: "getDirection"},
		}},
		{"nil expression", nil},
		{"named-arg with unresolvable long value", &ast.ObjectExpression{
			NodeType: ast.TypeObjectExpression,
			Properties: []ast.Property{
				{Key: &ast.Identifier{Name: "long"}, Value: &ast.Literal{Value: "invalid"}},
			},
		}},
		{"named-arg without long property", &ast.ObjectExpression{
			NodeType: ast.TypeObjectExpression,
			Properties: []ast.Property{
				{Key: &ast.Identifier{Name: "qty"}, Value: &ast.Literal{Value: 1.0}},
			},
		}},
		{"empty named-arg ObjectExpression", &ast.ObjectExpression{NodeType: ast.TypeObjectExpression}},
	}

	handler := NewStrategyActionHandler()

	for _, factory := range callFactories {
		for _, dir := range unresolvable {
			t.Run(factory.callee+"/"+dir.name, func(t *testing.T) {
				g := newTestGenerator()
				_, err := handler.GenerateCode(g, factory.build(dir.expr))
				if err == nil {
					t.Errorf("%s: want generation error for unresolvable direction %T, got success",
						factory.callee, dir.expr)
				}
			})
		}
	}
}

func TestStrategyDirectionExtractor_BackwardCompatibility(t *testing.T) {
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

	baseStructure := codes[0]
	for i, code := range codes[1:] {
		if code != baseStructure {
			t.Errorf("Code structure inconsistency at index %d:\nExpected:\n%s\nGot:\n%s",
				i+1, baseStructure, code)
		}
	}
}
