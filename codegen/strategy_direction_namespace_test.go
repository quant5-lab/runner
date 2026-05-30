package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// TestResolveStrategyDirectionMember locks the AST-shape recognition for
// `strategy.direction.{long,short,all}` — a nested MemberExpression where the
// outer Object is itself `strategy.direction`.
//
// Removing the resolver re-surfaces the Hull-strategy codegen failure
// ("unsupported string member expression: ...") that leaves the strat_dir_value
// variable empty and silently disables strategy.risk.allow_entry_in().
func TestResolveStrategyDirectionMember(t *testing.T) {
	makeNested := func(prop string) *ast.MemberExpression {
		return &ast.MemberExpression{
			Object: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "direction"},
			},
			Property: &ast.Identifier{Name: prop},
		}
	}

	tests := []struct {
		name     string
		expr     ast.Expression
		wantOK   bool
		wantCode string
	}{
		{
			name:     "strategy.direction.long resolves to runtime constant",
			expr:     makeNested("long"),
			wantOK:   true,
			wantCode: "strategy.DirectionLong",
		},
		{
			name:     "strategy.direction.short resolves to runtime constant",
			expr:     makeNested("short"),
			wantOK:   true,
			wantCode: "strategy.DirectionShort",
		},
		{
			name:     "strategy.direction.all resolves to runtime constant",
			expr:     makeNested("all"),
			wantOK:   true,
			wantCode: "strategy.DirectionAll",
		},
		{
			name:   "strategy.direction.unknown is rejected",
			expr:   makeNested("unknown"),
			wantOK: false,
		},
		{
			name: "non-strategy outer object is rejected",
			expr: &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "position"},
					Property: &ast.Identifier{Name: "direction"},
				},
				Property: &ast.Identifier{Name: "long"},
			},
			wantOK: false,
		},
		{
			name: "non-direction middle property is rejected",
			expr: &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "risk"},
				},
				Property: &ast.Identifier{Name: "long"},
			},
			wantOK: false,
		},
		{
			name: "single-level strategy.long is not matched here",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "long"},
			},
			wantOK: false,
		},
		{
			name:   "non-member expression returns false",
			expr:   &ast.Identifier{Name: "foo"},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := resolveStrategyDirectionMember(tt.expr)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v (got=%q)", ok, tt.wantOK, got)
			}
			if got != tt.wantCode {
				t.Errorf("code = %q, want %q", got, tt.wantCode)
			}
		})
	}
}

// TestGenerateStringVariableInit_StrategyDirectionTernary locks the end-to-end
// codegen path for the Hull-strategy pattern:
//
//	strat_dir_input == "long" ? strategy.direction.long : ... : strategy.direction.all
//
// Prior to the fix this produced "unsupported string member expression" and the
// emitted assignment was discarded, silently breaking strategy.risk.allow_entry_in.
func TestGenerateStringVariableInit_StrategyDirectionTernary(t *testing.T) {
	g := newTestGenerator()

	expr := &ast.ConditionalExpression{
		Test: &ast.BinaryExpression{
			Operator: "==",
			Left:     &ast.Identifier{Name: "strat_dir_input"},
			Right:    &ast.Literal{Value: "long"},
		},
		Consequent: &ast.MemberExpression{
			Object: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "direction"},
			},
			Property: &ast.Identifier{Name: "long"},
		},
		Alternate: &ast.ConditionalExpression{
			Test: &ast.BinaryExpression{
				Operator: "==",
				Left:     &ast.Identifier{Name: "strat_dir_input"},
				Right:    &ast.Literal{Value: "short"},
			},
			Consequent: &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "direction"},
				},
				Property: &ast.Identifier{Name: "short"},
			},
			Alternate: &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "strategy"},
					Property: &ast.Identifier{Name: "direction"},
				},
				Property: &ast.Identifier{Name: "all"},
			},
		},
	}
	g.variables["strat_dir_input"] = "string"

	code, err := g.generateStringVariableInit("strat_dir_value", expr)
	if err != nil {
		t.Fatalf("generateStringVariableInit returned error: %v", err)
	}

	for _, want := range []string{
		"strategy.DirectionLong",
		"strategy.DirectionShort",
		"strategy.DirectionAll",
	} {
		if !contains(code, want) {
			t.Errorf("expected %q in generated code, got:\n%s", want, code)
		}
	}
}
