package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestConditionalCodeGenerator_GenerateSetCallsForExpression(t *testing.T) {
	g := newTestGenerator()
	analyzer := NewConditionalArgumentAnalyzer(&ExpressionHasher{})
	gen := NewConditionalCodeGenerator(g, analyzer, g.tempVarMgr)

	conditional := &ast.ConditionalExpression{
		Test: &ast.BinaryExpression{
			Left:     &ast.Identifier{Name: "close"},
			Operator: ">",
			Right:    &ast.Identifier{Name: "open"},
		},
		Consequent: &ast.Identifier{Name: "high"},
		Alternate:  &ast.Identifier{Name: "low"},
	}

	callExpr := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "ta.sma"},
		Arguments: []ast.Expression{
			conditional,
			&ast.Literal{Value: 14.0},
		},
	}

	hasher := &ExpressionHasher{}
	hash := hasher.Hash(conditional)
	if len(hash) > 8 {
		hash = hash[:8]
	}
	g.tempVarMgr.RegisterConditional(hash, conditional)

	code := gen.GenerateSetCallsForExpression(callExpr, func() string { return "" })

	if !strings.Contains(code, "Series.Set(") {
		t.Errorf("Expected generated code to contain .Set() call, got: %s", code)
	}
}

func TestConditionalCodeGenerator_AccessModes(t *testing.T) {
	tests := []struct {
		name       string
		mode       AccessMode
		consequent ast.Expression
		alternate  ast.Expression
		wantSuffix string
	}{
		{
			name:       "ValueMode with literals",
			mode:       AccessModeValue,
			consequent: &ast.Literal{Value: 1.0},
			alternate:  &ast.Literal{Value: 0.0},
			wantSuffix: ".Get(0)",
		},
		{
			name:       "SeriesMode with identifiers",
			mode:       AccessModeSeries,
			consequent: &ast.Identifier{Name: "high"},
			alternate:  &ast.Identifier{Name: "low"},
			wantSuffix: "Series",
		},
		{
			name:       "CurrentMode with identifiers",
			mode:       AccessModeCurrent,
			consequent: &ast.Identifier{Name: "high"},
			alternate:  &ast.Identifier{Name: "low"},
			wantSuffix: ".GetCurrent()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			analyzer := NewConditionalArgumentAnalyzer(&ExpressionHasher{})
			gen := NewConditionalCodeGenerator(g, analyzer, g.tempVarMgr)

			conditional := &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "close"},
					Operator: ">",
					Right:    &ast.Identifier{Name: "open"},
				},
				Consequent: tt.consequent,
				Alternate:  tt.alternate,
			}

			hasher := &ExpressionHasher{}
			hash := hasher.Hash(conditional)
			if len(hash) > 8 {
				hash = hash[:8]
			}
			g.tempVarMgr.RegisterConditional(hash, conditional)

			ref, found := gen.GetTempVarReference(conditional, tt.mode)
			if !found {
				t.Error("Expected to find temp var reference")
			}
			if !strings.Contains(ref, tt.wantSuffix) {
				t.Errorf("Expected reference to contain %q, got: %s", tt.wantSuffix, ref)
			}
		})
	}
}

func TestConditionalCodeGenerator_GetTempVarReference(t *testing.T) {
	tests := []struct {
		name       string
		mode       AccessMode
		wantSuffix string
	}{
		{
			name:       "ValueMode",
			mode:       AccessModeValue,
			wantSuffix: ".Get(0)",
		},
		{
			name:       "SeriesMode",
			mode:       AccessModeSeries,
			wantSuffix: "Series",
		},
		{
			name:       "CurrentMode",
			mode:       AccessModeCurrent,
			wantSuffix: ".GetCurrent()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			analyzer := NewConditionalArgumentAnalyzer(&ExpressionHasher{})
			gen := NewConditionalCodeGenerator(g, analyzer, g.tempVarMgr)

			conditional := &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "close"},
					Operator: ">",
					Right:    &ast.Identifier{Name: "open"},
				},
				Consequent: &ast.Identifier{Name: "high"},
				Alternate:  &ast.Identifier{Name: "low"},
			}

			hasher := &ExpressionHasher{}
			hash := hasher.Hash(conditional)
			if len(hash) > 8 {
				hash = hash[:8]
			}
			g.tempVarMgr.RegisterConditional(hash, conditional)

			got, found := gen.GetTempVarReference(conditional, tt.mode)
			if !found {
				t.Error("Expected to find temp var reference")
			}
			if !strings.Contains(got, tt.wantSuffix) {
				t.Errorf("Expected reference to contain %q, got: %s", tt.wantSuffix, got)
			}
		})
	}
}

func TestConditionalCodeGenerator_GenerateSetCallsForStatement(t *testing.T) {
	g := newTestGenerator()
	analyzer := NewConditionalArgumentAnalyzer(&ExpressionHasher{})
	gen := NewConditionalCodeGenerator(g, analyzer, g.tempVarMgr)

	conditional := &ast.ConditionalExpression{
		Test: &ast.BinaryExpression{
			Left:     &ast.Identifier{Name: "close"},
			Operator: ">",
			Right:    &ast.Identifier{Name: "open"},
		},
		Consequent: &ast.Identifier{Name: "high"},
		Alternate:  &ast.Identifier{Name: "low"},
	}

	stmt := &ast.ExpressionStatement{
		Expression: &ast.CallExpression{
			Callee: &ast.Identifier{Name: "plot"},
			Arguments: []ast.Expression{
				conditional,
			},
		},
	}

	hasher := &ExpressionHasher{}
	hash := hasher.Hash(conditional)
	if len(hash) > 8 {
		hash = hash[:8]
	}
	g.tempVarMgr.RegisterConditional(hash, conditional)

	code := gen.GenerateSetCallsForStatement(stmt, func() string { return "" })

	if !strings.Contains(code, "Series.Set(") {
		t.Errorf("Expected statement code to contain .Set() call, got: %s", code)
	}
}

func TestConditionalCodeGenerator_NilSafety(t *testing.T) {
	g := newTestGenerator()
	analyzer := NewConditionalArgumentAnalyzer(&ExpressionHasher{})
	gen := NewConditionalCodeGenerator(g, analyzer, g.tempVarMgr)

	code := gen.GenerateSetCallsForExpression(nil, func() string { return "" })
	if code != "" {
		t.Errorf("Expected empty code for nil expression, got: %s", code)
	}

	code = gen.GenerateSetCallsForStatement(nil, func() string { return "" })
	if code != "" {
		t.Errorf("Expected empty code for nil statement, got: %s", code)
	}
}
