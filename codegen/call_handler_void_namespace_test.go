package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// TestVoidNamespaceCallHandler_CanHandle covers the full decision boundary of the
// handler: all five chart-only namespaces must match for any dotted suffix and for
// the bare constructor form; everything outside those namespaces must not match.
func TestVoidNamespaceCallHandler_CanHandle(t *testing.T) {
	h := &VoidNamespaceCallHandler{}

	tests := []struct {
		funcName string
		want     bool
	}{
		// Bare type constructors
		{"label", true},
		{"line", true},
		{"box", true},
		{"table", true},
		{"linefill", true},
		// All three method families for each namespace
		{"label.new", true},
		{"label.delete", true},
		{"label.set_xy", true},
		{"label.set_text", true},
		{"label.set_size", true},
		{"label.set_color", true},
		{"label.set_textcolor", true},
		{"label.set_style", true},
		{"label.get_x", true},
		{"label.get_y", true},
		{"label.get_text", true},
		{"label.copy", true},
		{"line.new", true},
		{"line.delete", true},
		{"line.set_color", true},
		{"line.set_width", true},
		{"line.set_style", true},
		{"line.get_x1", true},
		{"line.get_y1", true},
		{"line.get_x2", true},
		{"line.get_y2", true},
		{"box.new", true},
		{"box.delete", true},
		{"box.set_bgcolor", true},
		{"box.set_border_color", true},
		{"box.get_top", true},
		{"table.new", true},
		{"table.delete", true},
		{"table.cell", true},
		{"table.set_bgcolor", true},
		{"table.merge_cells", true},
		{"linefill.new", true},
		{"linefill.delete", true},
		{"linefill.get_line1", true},
		// Unknown methods on chart namespaces still match (forward compatibility)
		{"label.set_future_property", true},
		{"line.do_something_new", true},
		// Non-chart namespaces
		{"alert", false},
		{"alertcondition", false},
		{"plot", false},
		{"plotshape", false},
		{"ta.sma", false},
		{"ta.ema", false},
		{"strategy.entry", false},
		{"strategy.close", false},
		{"math.abs", false},
		{"str.lower", false},
		{"request.security", false},
		{"array.new_float", false},
		// Substring collisions: a namespace that contains a chart-namespace name
		// as a prefix of a longer word must NOT match
		{"labelcustom", false},
		{"linestyle", false},
		{"tablename", false},
		{"boxing", false},
		{"linefilling", false},
		// Other namespace prefixes that share substrings
		{"ta.label", false},
		{"mylib.label", false},
		{"ta.line", false},
		// Case sensitivity
		{"Label", false},
		{"Label.new", false},
		{"LINE", false},
		{"BOX.new", false},
		// Empty
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			if got := h.CanHandle(tt.funcName); got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

// TestVoidNamespaceCallHandler_GenerateCode_MatchingCalls verifies the split
// between silent presentation calls and calculation-bearing drawing getters.
func TestVoidNamespaceCallHandler_GenerateCode_MatchingCalls(t *testing.T) {
	h := &VoidNamespaceCallHandler{}

	tests := []struct {
		name         string
		call         *ast.CallExpression
		wantContains string
		wantGap      string
	}{
		{
			name:         "bare label constructor with na",
			wantContains: "math.NaN()",
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "label"},
				Arguments: []ast.Expression{&ast.Identifier{Name: "na"}},
			},
		},
		{
			name:         "label.new with all positional arguments",
			wantContains: "math.NaN()",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "label"},
					Property: &ast.Identifier{Name: "new"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 100.0},
					&ast.Literal{Value: 200.0},
					&ast.Literal{Value: "text"},
					&ast.MemberExpression{Object: &ast.Identifier{Name: "xloc"}, Property: &ast.Identifier{Name: "bar_time"}},
					&ast.MemberExpression{Object: &ast.Identifier{Name: "yloc"}, Property: &ast.Identifier{Name: "price"}},
				},
			},
		},
		{
			name:         "label.delete with label reference",
			wantContains: "math.NaN()",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "label"},
					Property: &ast.Identifier{Name: "delete"},
				},
				Arguments: []ast.Expression{&ast.Identifier{Name: "myLabel"}},
			},
		},
		{
			name:         "label.set_xy mutator",
			wantContains: "math.NaN()",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "label"},
					Property: &ast.Identifier{Name: "set_xy"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "lbl"},
					&ast.Literal{Value: 100.0},
					&ast.Literal{Value: 200.0},
				},
			},
		},
		{
			name:         "label.get_x getter",
			wantContains: `featuregap.Record("label.get_x"`,
			wantGap:      "label.get_x",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "label"},
					Property: &ast.Identifier{Name: "get_x"},
				},
				Arguments: []ast.Expression{&ast.Identifier{Name: "lbl"}},
			},
		},
		{
			name:         "label.copy returns new reference",
			wantContains: "math.NaN()",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "label"},
					Property: &ast.Identifier{Name: "copy"},
				},
				Arguments: []ast.Expression{&ast.Identifier{Name: "lbl"}},
			},
		},
		{
			name:         "line.new with coordinate arguments",
			wantContains: "math.NaN()",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "line"},
					Property: &ast.Identifier{Name: "new"},
				},
				Arguments: []ast.Expression{},
			},
		},
		{
			name:         "line.delete",
			wantContains: "math.NaN()",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "line"},
					Property: &ast.Identifier{Name: "delete"},
				},
				Arguments: []ast.Expression{&ast.Identifier{Name: "ln"}},
			},
		},
		{
			name:         "box.new",
			wantContains: "math.NaN()",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "box"},
					Property: &ast.Identifier{Name: "new"},
				},
				Arguments: []ast.Expression{},
			},
		},
		{
			name:         "table.new with rows and cols",
			wantContains: "math.NaN()",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "table"},
					Property: &ast.Identifier{Name: "new"},
				},
				Arguments: []ast.Expression{
					&ast.MemberExpression{Object: &ast.Identifier{Name: "position"}, Property: &ast.Identifier{Name: "top_right"}},
					&ast.Literal{Value: 3.0},
					&ast.Literal{Value: 4.0},
				},
			},
		},
		{
			name:         "table.cell populates a cell",
			wantContains: "math.NaN()",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "table"},
					Property: &ast.Identifier{Name: "cell"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "tbl"},
					&ast.Literal{Value: 0.0},
					&ast.Literal{Value: 0.0},
					&ast.Literal{Value: "header"},
				},
			},
		},
		{
			name:         "linefill.new",
			wantContains: "math.NaN()",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "linefill"},
					Property: &ast.Identifier{Name: "new"},
				},
				Arguments: []ast.Expression{},
			},
		},
		{
			name:         "call with no arguments",
			wantContains: "math.NaN()",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "label"},
					Property: &ast.Identifier{Name: "new"},
				},
				Arguments: []ast.Expression{},
			},
		},
		{
			name:         "call with complex expression argument",
			wantContains: "math.NaN()",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "label"},
					Property: &ast.Identifier{Name: "new"},
				},
				Arguments: []ast.Expression{
					&ast.BinaryExpression{
						Operator: "+",
						Left:     &ast.Identifier{Name: "time"},
						Right:    &ast.Literal{Value: 3600.0},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			code, err := h.GenerateCode(g, tt.call)
			if err != nil {
				t.Fatalf("GenerateCode() unexpected error: %v", err)
			}
			if !contains(code, tt.wantContains) {
				t.Errorf("GenerateCode() = %q, want containing %q", code, tt.wantContains)
			}
			if tt.wantGap != "" && !containsGap(g.featureGaps, tt.wantGap) {
				t.Errorf("featureGaps = %v, want %q", g.featureGaps, tt.wantGap)
			}
			if tt.wantGap == "" && len(g.featureGaps) != 0 {
				t.Errorf("featureGaps = %v, want none", g.featureGaps)
			}
		})
	}
}

// TestVoidNamespaceCallHandler_GenerateCode_NonMatchingCallsPassThrough verifies that
// the handler returns ("", nil) for calls it does not own, so the router can
// continue to the next handler in the chain.
func TestVoidNamespaceCallHandler_GenerateCode_NonMatchingCallsPassThrough(t *testing.T) {
	h := &VoidNamespaceCallHandler{}
	g := newTestGenerator()

	tests := []struct {
		name string
		call *ast.CallExpression
	}{
		{
			name: "alert flat builtin",
			call: &ast.CallExpression{Callee: &ast.Identifier{Name: "alert"}},
		},
		{
			name: "ta.sma TA function",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{Object: &ast.Identifier{Name: "ta"}, Property: &ast.Identifier{Name: "sma"}},
			},
		},
		{
			name: "strategy.entry",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{Object: &ast.Identifier{Name: "strategy"}, Property: &ast.Identifier{Name: "entry"}},
			},
		},
		{
			name: "plot",
			call: &ast.CallExpression{Callee: &ast.Identifier{Name: "plot"}},
		},
		{
			name: "unknown function",
			call: &ast.CallExpression{Callee: &ast.Identifier{Name: "never_seen_before_func"}},
		},
		{
			name: "ta.label is not a chart namespace (ta prefix dominates)",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{Object: &ast.Identifier{Name: "ta"}, Property: &ast.Identifier{Name: "label"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := h.GenerateCode(g, tt.call)
			if err != nil {
				t.Fatalf("GenerateCode() unexpected error: %v", err)
			}
			if code != "" {
				t.Errorf("GenerateCode() for non-matching call = %q, want empty string (pass-through)", code)
			}
		})
	}
}

// TestVoidNamespaceCallHandler_GeneratorStateUnchanged verifies that handling a
// chart-only call leaves the generator's featureGaps and other mutable state
// unmodified. Chart drawing calls must not pollute the featuregap log.
func TestVoidNamespaceCallHandler_GeneratorStateUnchanged(t *testing.T) {
	h := &VoidNamespaceCallHandler{}

	calls := []*ast.CallExpression{
		{
			Callee: &ast.MemberExpression{Object: &ast.Identifier{Name: "label"}, Property: &ast.Identifier{Name: "new"}},
		},
		{
			Callee: &ast.MemberExpression{Object: &ast.Identifier{Name: "label"}, Property: &ast.Identifier{Name: "delete"}},
		},
		{
			Callee:    &ast.Identifier{Name: "label"},
			Arguments: []ast.Expression{&ast.Identifier{Name: "na"}},
		},
		{
			Callee: &ast.MemberExpression{Object: &ast.Identifier{Name: "line"}, Property: &ast.Identifier{Name: "new"}},
		},
		{
			Callee: &ast.MemberExpression{Object: &ast.Identifier{Name: "table"}, Property: &ast.Identifier{Name: "cell"}},
		},
	}

	for _, call := range calls {
		t.Run(extractCallFunctionName(call), func(t *testing.T) {
			g := newTestGenerator()
			gapsBefore := len(g.featureGaps)

			_, _ = h.GenerateCode(g, call)

			if len(g.featureGaps) != gapsBefore {
				t.Errorf("GenerateCode() added %d featureGap entries; chart-only calls must not pollute the featuregap log",
					len(g.featureGaps)-gapsBefore)
			}
		})
	}
}
