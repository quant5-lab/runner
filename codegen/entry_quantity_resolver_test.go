package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// qtyLiteralEval resolves only AST float/int literals to (value, true).
// Zero and negative values yield (0, false) because they are not valid position sizes.
func qtyLiteralEval(expr ast.Expression) (float64, bool) {
	lit, ok := expr.(*ast.Literal)
	if !ok {
		return 0, false
	}
	switch v := lit.Value.(type) {
	case float64:
		if v > 0 {
			return v, true
		}
	case int:
		if v > 0 {
			return float64(v), true
		}
	}
	return 0, false
}

// qtyConstEval resolves literals and also looks up identifiers in a symbol table.
// Models the behaviour of makeQtyEvaluator, where identifiers bound to
// input constants resolve to their defval.
func qtyConstEval(consts map[string]float64) QtyEvaluator {
	return func(expr ast.Expression) (float64, bool) {
		if v, ok := qtyLiteralEval(expr); ok {
			return v, true
		}
		if id, ok := expr.(*ast.Identifier); ok {
			if v, exists := consts[id.Name]; exists && v > 0 {
				return v, true
			}
		}
		return 0, false
	}
}

// -- builders --

func qtyResolver() *EntryQuantityResolver { return NewEntryQuantityResolver() }

func positionalQtyArgs(qty ast.Expression) []ast.Expression {
	return []ast.Expression{
		&ast.Literal{Value: "Long"},
		&ast.Identifier{Name: "strategy.long"},
		qty,
	}
}

// allNamedArgs models strategy.entry(id="Long", long=strategy.long, qty=…, when=…)
// where every parameter is named, producing a single ObjectExpression argument.
func allNamedArgs(props ...ast.Property) []ast.Expression {
	return []ast.Expression{&ast.ObjectExpression{Properties: props}}
}

// mixedNamedArgs models strategy.entry("Long", long=strategy.long, qty=…, when=…)
// where the id is positional but direction/qty/when are named.
func mixedNamedArgs(props ...ast.Property) []ast.Expression {
	return []ast.Expression{
		&ast.Literal{Value: "Long"},
		&ast.ObjectExpression{Properties: props},
	}
}

func eqProp(key string, value ast.Expression) ast.Property {
	return ast.Property{Key: &ast.Identifier{Name: key}, Value: value}
}

func qtyProp(value ast.Expression) ast.Property { return eqProp("qty", value) }

func intLit(v int) *ast.Literal { return &ast.Literal{Value: v} }

// TestEntryQuantityResolver_ArgCount verifies the arg-count boundary:
// fewer than 3 positional args without a named-arg object default to defaultQty.
func TestEntryQuantityResolver_ArgCount(t *testing.T) {
	cases := []struct {
		name    string
		args    []ast.Expression
		wantQty float64
	}{
		{
			name:    "nil args",
			args:    nil,
			wantQty: 5.0,
		},
		{
			name:    "zero args",
			args:    []ast.Expression{},
			wantQty: 5.0,
		},
		{
			name:    "one arg (id only)",
			args:    []ast.Expression{fltLit(1)},
			wantQty: 5.0,
		},
		{
			name:    "two args (id + direction)",
			args:    []ast.Expression{&ast.Literal{Value: "Long"}, ident("strategy.long")},
			wantQty: 5.0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := qtyResolver().ResolveQuantity(tc.args, tc.wantQty, qtyLiteralEval)
			if got != tc.wantQty {
				t.Errorf("want default %v, got %v", tc.wantQty, got)
			}
		})
	}
}

// TestEntryQuantityResolver_PositionalLiteral verifies that a float or int literal
// at the third positional slot (args[2]) is extracted as the qty.
func TestEntryQuantityResolver_PositionalLiteral(t *testing.T) {
	cases := []struct {
		name    string
		arg     ast.Expression
		wantQty float64
	}{
		{name: "float64", arg: fltLit(10000.0), wantQty: 10000.0},
		{name: "fractional", arg: fltLit(0.5), wantQty: 0.5},
		{name: "int", arg: intLit(50), wantQty: 50.0},
		{name: "large", arg: fltLit(1e6), wantQty: 1e6},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := qtyResolver().ResolveQuantity(positionalQtyArgs(tc.arg), 1.0, qtyLiteralEval)
			if got != tc.wantQty {
				t.Errorf("want %v, got %v", tc.wantQty, got)
			}
		})
	}
}

// TestEntryQuantityResolver_PositionalLiteralFallback verifies that non-positive
// positional literals fall back to defaultQty.
func TestEntryQuantityResolver_PositionalLiteralFallback(t *testing.T) {
	cases := []struct {
		name string
		arg  ast.Expression
	}{
		{name: "zero", arg: fltLit(0.0)},
		{name: "negative", arg: fltLit(-1.0)},
		{name: "negative int", arg: intLit(-10)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			const want = 7.0
			got := qtyResolver().ResolveQuantity(positionalQtyArgs(tc.arg), want, qtyLiteralEval)
			if got != want {
				t.Errorf("non-positive positional literal must use default: want %v, got %v", want, got)
			}
		})
	}
}

// TestEntryQuantityResolver_PositionalConstantIdentifier verifies that a variable
// identifier at args[2] resolves through the evaluator when that identifier
// is bound to a known constant (e.g., an input defval).
func TestEntryQuantityResolver_PositionalConstantIdentifier(t *testing.T) {
	cases := []struct {
		name   string
		consts map[string]float64
		id     string
		want   float64
	}{
		{
			name:   "known constant resolves to its value",
			consts: map[string]float64{"tradeQty": 250.0},
			id:     "tradeQty",
			want:   250.0,
		},
		{
			name:   "unknown identifier falls back to default",
			consts: map[string]float64{},
			id:     "dynamicQty",
			want:   1.0,
		},
		{
			name:   "constant registered as zero falls back to default",
			consts: map[string]float64{"zeroQty": 0.0},
			id:     "zeroQty",
			want:   1.0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := qtyResolver().ResolveQuantity(
				positionalQtyArgs(ident(tc.id)),
				1.0,
				qtyConstEval(tc.consts),
			)
			if got != tc.want {
				t.Errorf("want %v, got %v", tc.want, got)
			}
		})
	}
}

// TestEntryQuantityResolver_PositionalObjectExpressionSkipped verifies that an
// ObjectExpression at args[2] (the named-arg blob in all-positional-then-named calls)
// is not interpreted as the qty value.
func TestEntryQuantityResolver_PositionalObjectExpressionSkipped(t *testing.T) {
	cases := []struct {
		name string
		obj  *ast.ObjectExpression
	}{
		{
			name: "empty object",
			obj:  &ast.ObjectExpression{},
		},
		{
			name: "object with when= only",
			obj: &ast.ObjectExpression{Properties: []ast.Property{
				eqProp("when", ident("signal")),
			}},
		},
		{
			name: "object with comment= only",
			obj: &ast.ObjectExpression{Properties: []ast.Property{
				eqProp("comment", &ast.Literal{Value: "buy"}),
			}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := []ast.Expression{&ast.Literal{Value: "Long"}, ident("strategy.long"), tc.obj}
			const want = 9.0
			got := qtyResolver().ResolveQuantity(args, want, qtyLiteralEval)
			if got != want {
				t.Errorf("ObjectExpression at positional slot must not be treated as qty: want %v, got %v", want, got)
			}
		})
	}
}

// TestEntryQuantityResolver_NamedQtyLiteral verifies qty= extraction from named
// arguments packed as an ObjectExpression.  Both call shapes are covered:
// mixed (positional id + named rest) and all-named (every parameter named).
func TestEntryQuantityResolver_NamedQtyLiteral(t *testing.T) {
	cases := []struct {
		name    string
		args    []ast.Expression
		wantQty float64
	}{
		{
			name:    "mixed: id positional, rest named",
			args:    mixedNamedArgs(eqProp("long", ident("strategy.long")), qtyProp(fltLit(500.0))),
			wantQty: 500.0,
		},
		{
			name: "mixed: qty among multiple named params",
			args: mixedNamedArgs(
				eqProp("long", ident("strategy.long")),
				qtyProp(fltLit(100.0)),
				eqProp("comment", &ast.Literal{Value: "entry"}),
				eqProp("when", ident("signal")),
			),
			wantQty: 100.0,
		},
		{
			name: "all-named: id also in ObjectExpression",
			args: allNamedArgs(
				eqProp("id", &ast.Literal{Value: "Long"}),
				eqProp("long", ident("strategy.long")),
				qtyProp(fltLit(200.0)),
			),
			wantQty: 200.0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := qtyResolver().ResolveQuantity(tc.args, 1.0, qtyLiteralEval)
			if got != tc.wantQty {
				t.Errorf("want %v, got %v", tc.wantQty, got)
			}
		})
	}
}

// TestEntryQuantityResolver_NamedQtyLiteralFallback verifies that a qty= property
// whose value is non-positive falls back to defaultQty, consistent with the
// positional path behaviour.
func TestEntryQuantityResolver_NamedQtyLiteralFallback(t *testing.T) {
	cases := []struct {
		name string
		val  ast.Expression
	}{
		{name: "zero", val: fltLit(0.0)},
		{name: "negative", val: fltLit(-5.0)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := mixedNamedArgs(qtyProp(tc.val))
			const want = 3.0
			got := qtyResolver().ResolveQuantity(args, want, qtyLiteralEval)
			if got != want {
				t.Errorf("non-positive named qty must use default: want %v, got %v", want, got)
			}
		})
	}
}

// TestEntryQuantityResolver_NamedQtyConstantIdentifier verifies that when qty=
// is a named identifier bound to a known constant, the constant value is used.
func TestEntryQuantityResolver_NamedQtyConstantIdentifier(t *testing.T) {
	cases := []struct {
		name    string
		consts  map[string]float64
		id      string
		wantQty float64
	}{
		{
			name:    "known constant resolves",
			consts:  map[string]float64{"tradeSize": 10000.0},
			id:      "tradeSize",
			wantQty: 10000.0,
		},
		{
			name:    "unknown identifier falls back to default",
			consts:  map[string]float64{},
			id:      "dynamicQty",
			wantQty: 2.0,
		},
		{
			name:    "constant registered as zero falls back to default",
			consts:  map[string]float64{"empty": 0.0},
			id:      "empty",
			wantQty: 2.0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := mixedNamedArgs(eqProp("long", ident("strategy.long")), qtyProp(ident(tc.id)))
			got := qtyResolver().ResolveQuantity(args, 2.0, qtyConstEval(tc.consts))
			if got != tc.wantQty {
				t.Errorf("want %v, got %v", tc.wantQty, got)
			}
		})
	}
}

// TestEntryQuantityResolver_NamedBlobWithoutQtyProperty verifies that an
// ObjectExpression containing only non-qty properties does not produce a qty —
// defaultQty is returned.
func TestEntryQuantityResolver_NamedBlobWithoutQtyProperty(t *testing.T) {
	cases := []struct {
		name string
		args []ast.Expression
	}{
		{
			name: "only direction in named blob",
			args: mixedNamedArgs(eqProp("long", ident("strategy.long"))),
		},
		{
			name: "direction + when in named blob",
			args: mixedNamedArgs(eqProp("long", ident("strategy.long")), eqProp("when", ident("signal"))),
		},
		{
			name: "empty object expression",
			args: []ast.Expression{&ast.ObjectExpression{}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			const want = 4.0
			got := qtyResolver().ResolveQuantity(tc.args, want, qtyLiteralEval)
			if got != want {
				t.Errorf("no qty property: want default %v, got %v", want, got)
			}
		})
	}
}

// TestEntryQuantityResolver_ExtraPositionalArgs verifies that when more than three
// positional args are present, only args[2] is consulted for qty.
func TestEntryQuantityResolver_ExtraPositionalArgs(t *testing.T) {
	args := []ast.Expression{
		&ast.Literal{Value: "Long"},
		ident("strategy.long"),
		fltLit(8.0),
		&ast.ObjectExpression{},
		fltLit(99.0),
	}
	got := qtyResolver().ResolveQuantity(args, 1.0, qtyLiteralEval)
	if got != 8.0 {
		t.Errorf("third arg must be used, want 8.0, got %v", got)
	}
}
