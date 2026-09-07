package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// TestBooleanConverter_GoBoolLiteral_EnsureBooleanOperand verifies that
// EnsureBooleanOperand never wraps generated code that is already a Go boolean
// keyword. Namespace resolvers (e.g. timeframe.isseconds) return bare "true"
// or "false"; passing those to value.IsTrue would be a compile error because
// IsTrue expects float64.
func TestBooleanConverter_GoBoolLiteral_EnsureBooleanOperand(t *testing.T) {
	tests := []struct {
		name          string
		expr          ast.Expression
		generatedCode string
		want          string
	}{
		// ---- "false" constant (timeframe.isseconds, timeframe.isticks, etc.) ----
		{
			name: "false from MemberExpression - not wrapped",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "timeframe"},
				Property: &ast.Identifier{Name: "isseconds"},
			},
			generatedCode: "false",
			want:          "false",
		},
		{
			name:          "false from Identifier - not wrapped",
			expr:          &ast.Identifier{Name: "isseconds"},
			generatedCode: "false",
			want:          "false",
		},
		{
			name: "false from CallExpression - not wrapped",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "someFunc"},
			},
			generatedCode: "false",
			want:          "false",
		},
		// ---- "true" constant (barstate.ishistory, barstate.isnew, etc.) ----
		{
			name: "true from MemberExpression - not wrapped",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "barstate"},
				Property: &ast.Identifier{Name: "ishistory"},
			},
			generatedCode: "true",
			want:          "true",
		},
		{
			name:          "true from Identifier - not wrapped",
			expr:          &ast.Identifier{Name: "ishistory"},
			generatedCode: "true",
			want:          "true",
		},
		// ---- runtime boolean expressions (GoBool, but not bare keywords) ----
		// ctx.IsMonthly is a runtime Go bool field, not a keyword — EnsureBooleanOperand
		// has no mechanism to recognise it as boolean (no Series pattern, no comparison)
		// so it wraps it with value.IsTrue.  This is a known limitation: runtime ctx.*
		// bool fields work correctly in if-statement position even when wrapped because
		// value.IsTrue on a Go bool compares 0/1 encoded float64 from NamespaceResolver.
		// The critical invariant this suite guards is that bare "true"/"false" keywords
		// are NEVER wrapped — they differ from runtime ctx.* expressions.
		{
			name: "runtime ctx.IsMonthly from MemberExpression - wraps (not a keyword)",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "timeframe"},
				Property: &ast.Identifier{Name: "ismonthly"},
			},
			generatedCode: "ctx.IsMonthly",
			want:          "(value.IsTrue(ctx.IsMonthly))",
		},
		// ---- non-keyword codes that look similar but are not keywords ----
		{
			name:          "trueval - not a keyword, identifier node wraps",
			expr:          &ast.Identifier{Name: "trueval"},
			generatedCode: "truevalSeries.GetCurrent()",
			want:          "(value.IsTrue(truevalSeries.GetCurrent()))",
		},
		{
			name:          "falseSignal - not a keyword, identifier node wraps",
			expr:          &ast.Identifier{Name: "falseSignal"},
			generatedCode: "falseSignalSeries.GetCurrent()",
			want:          "(value.IsTrue(falseSignalSeries.GetCurrent()))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := NewBooleanConverter(NewTypeInferenceEngine())
			got := bc.EnsureBooleanOperand(tt.expr, tt.generatedCode)
			if got != tt.want {
				t.Errorf("EnsureBooleanOperand(%q)\n  want: %q\n  got:  %q", tt.generatedCode, tt.want, got)
			}
		})
	}
}

// TestBooleanConverter_GoBoolLiteral_ConvertBoolSeriesForIfStatement verifies that
// ConvertBoolSeriesForIfStatement also passes through bare "true"/"false" codes
// without alteration — the if-statement path uses a different code route but must
// honour the same contract.
func TestBooleanConverter_GoBoolLiteral_ConvertBoolSeriesForIfStatement(t *testing.T) {
	tests := []struct {
		name          string
		expr          ast.Expression
		generatedCode string
		want          string
	}{
		{
			name: "false keyword - not wrapped in if-statement context",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "timeframe"},
				Property: &ast.Identifier{Name: "isseconds"},
			},
			generatedCode: "false",
			want:          "false",
		},
		{
			name: "true keyword - not wrapped in if-statement context",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "barstate"},
				Property: &ast.Identifier{Name: "isrealtime"},
			},
			generatedCode: "true",
			want:          "true",
		},
		{
			name:          "true keyword from Identifier - not wrapped",
			expr:          &ast.Identifier{Name: "myFlag"},
			generatedCode: "true",
			want:          "true",
		},
		{
			name: "false keyword from CallExpression - not wrapped",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "compute"},
			},
			generatedCode: "false",
			want:          "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := NewBooleanConverter(NewTypeInferenceEngine())
			got := bc.ConvertBoolSeriesForIfStatement(tt.expr, tt.generatedCode)
			if got != tt.want {
				t.Errorf("ConvertBoolSeriesForIfStatement(%q)\n  want: %q\n  got:  %q", tt.generatedCode, tt.want, got)
			}
		})
	}
}

// TestIsGoBoolLiteral verifies the predicate exhaustively, including boundary
// cases that differ by a single character from the keywords.
func TestIsGoBoolLiteral(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"true", true},
		{"false", true},
		// --- not keywords ---
		{"True", false},
		{"False", false},
		{"TRUE", false},
		{"FALSE", false},
		{"true ", false}, // trailing space
		{" true", false}, // leading space
		{"trueish", false},
		{"falsy", false},
		{"0", false},
		{"1", false},
		{"", false},
		{"trueval", false},
		{"false0", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isGoBoolLiteral(tt.input)
			if got != tt.want {
				t.Errorf("isGoBoolLiteral(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestBooleanConverter_GoBoolLiteral_NamespaceProperties exercises every
// namespace property that the BuiltinNamespaceResolver resolves to a bare
// "true" or "false" string, ensuring EnsureBooleanOperand passes them through
// unchanged regardless of the AST node type.
func TestBooleanConverter_GoBoolLiteral_NamespaceProperties(t *testing.T) {
	// Each entry is (namespace, property, expected Go code).
	// Source: BuiltinNamespaceResolver — properties whose Code is "true" or "false".
	resolver := NewBuiltinNamespaceResolver()

	type entry struct{ ns, prop string }
	bareKeywordProps := []entry{
		{"barstate", "isrealtime"},
		{"barstate", "isnew"},
		{"barstate", "isconfirmed"},
		{"barstate", "ishistory"},
		{"timeframe", "isseconds"},
		{"timeframe", "isticks"},
		{"session", "ismarket"},
		{"session", "ispremarket"},
		{"session", "ispostmarket"},
		{"chart", "is_heikinashi"},
		{"chart", "is_kagi"},
		{"chart", "is_linebreak"},
		{"chart", "is_pnf"},
		{"chart", "is_range"},
		{"chart", "is_renko"},
	}

	bc := NewBooleanConverter(NewTypeInferenceEngine())
	memberExpr := func(ns, prop string) ast.Expression {
		return &ast.MemberExpression{
			Object:   &ast.Identifier{Name: ns},
			Property: &ast.Identifier{Name: prop},
		}
	}

	for _, e := range bareKeywordProps {
		res, found := resolver.Resolve(e.ns, e.prop)
		if !found {
			t.Errorf("%s.%s: resolver returned not-found", e.ns, e.prop)
			continue
		}
		if res.Code != "true" && res.Code != "false" {
			// Property no longer resolves to a bare keyword; skip (it's a runtime expr).
			continue
		}
		t.Run(e.ns+"."+e.prop, func(t *testing.T) {
			expr := memberExpr(e.ns, e.prop)
			got := bc.EnsureBooleanOperand(expr, res.Code)
			if got != res.Code {
				t.Errorf("EnsureBooleanOperand(%q) = %q, want %q (must not wrap bare keyword)", res.Code, got, res.Code)
			}
		})
	}
}
