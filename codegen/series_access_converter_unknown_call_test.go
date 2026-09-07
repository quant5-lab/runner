package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestSeriesAccessConverter_UnknownCallFallback validates graceful degradation of
 * unresolvable call expressions across all ConvertExpression dispatch paths.
 */
func TestSeriesAccessConverter_UnknownCallFallback(t *testing.T) {
	closeIdent := func() ast.Expression { return &ast.Identifier{Name: "close"} }

	t.Run("direct unknown call returns math.NaN() and reports gap", func(t *testing.T) {
		st := NewSymbolTable()
		var reported []string
		conv := NewSeriesAccessConverter(st, "j", nil)
		conv.ReportUnknownCall = func(name string) { reported = append(reported, name) }

		call := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "unknown_func"},
			Arguments: []ast.Expression{closeIdent()},
		}

		code, err := conv.ConvertExpression(call)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != "math.NaN()" {
			t.Errorf("got %q, want math.NaN()", code)
		}
		if len(reported) != 1 || reported[0] != "unknown_func" {
			t.Errorf("reported gaps = %v, want [unknown_func]", reported)
		}
	})

	t.Run("unknown call nested in binary expression reports gap once", func(t *testing.T) {
		st := NewSymbolTable()
		var reported []string
		conv := NewSeriesAccessConverter(st, "j", nil)
		conv.ReportUnknownCall = func(name string) { reported = append(reported, name) }

		expr := &ast.BinaryExpression{
			Left: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "unknown_func"},
				Arguments: []ast.Expression{closeIdent()},
			},
			Operator: "+",
			Right:    closeIdent(),
		}

		code, err := conv.ConvertExpression(expr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != "(math.NaN() + closeSeries.Get(j))" {
			t.Errorf("got %q, want (math.NaN() + closeSeries.Get(j))", code)
		}
		if len(reported) != 1 || reported[0] != "unknown_func" {
			t.Errorf("reported gaps = %v, want [unknown_func]", reported)
		}
	})

	t.Run("outer unknown call with nested inner unknown reports outer once", func(t *testing.T) {
		st := NewSymbolTable()
		var reported []string
		conv := NewSeriesAccessConverter(st, "j", nil)
		conv.ReportUnknownCall = func(name string) { reported = append(reported, name) }

		inner := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "inner_func"},
			Arguments: []ast.Expression{closeIdent()},
		}
		outer := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "outer_func"},
			Arguments: []ast.Expression{inner},
		}

		code, err := conv.ConvertExpression(outer)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != "math.NaN()" {
			t.Errorf("got %q, want math.NaN()", code)
		}
		// short-circuits at outer_func — inner args are not visited
		if len(reported) != 1 || reported[0] != "outer_func" {
			t.Errorf("reported gaps = %v, want [outer_func]", reported)
		}
	})

	t.Run("member-expression unknown call returns math.NaN() and reports namespaced name", func(t *testing.T) {
		st := NewSymbolTable()
		var reported []string
		conv := NewSeriesAccessConverter(st, "j", nil)
		conv.ReportUnknownCall = func(name string) { reported = append(reported, name) }

		call := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ns"},
				Property: &ast.Identifier{Name: "unknown"},
			},
			Arguments: []ast.Expression{closeIdent()},
		}

		code, err := conv.ConvertExpression(call)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != "math.NaN()" {
			t.Errorf("got %q, want math.NaN()", code)
		}
		if len(reported) != 1 || reported[0] != "ns.unknown" {
			t.Errorf("reported gaps = %v, want [ns.unknown]", reported)
		}
	})

	t.Run("nil ReportUnknownCall does not panic", func(t *testing.T) {
		st := NewSymbolTable()
		conv := NewSeriesAccessConverter(st, "j", nil)
		call := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "unknown_func"},
			Arguments: []ast.Expression{closeIdent()},
		}

		code, err := conv.ConvertExpression(call)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != "math.NaN()" {
			t.Errorf("got %q, want math.NaN()", code)
		}
	})

	t.Run("known math function is not reported as unknown", func(t *testing.T) {
		st := NewSymbolTable()
		var reported []string
		conv := NewSeriesAccessConverter(st, "j", nil)
		conv.ReportUnknownCall = func(name string) { reported = append(reported, name) }

		call := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "math.abs"},
			Arguments: []ast.Expression{closeIdent()},
		}

		_, err := conv.ConvertExpression(call)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(reported) != 0 {
			t.Errorf("math function should not be reported, got: %v", reported)
		}
	})

	t.Run("CallVarLookup hit prevents unknown report", func(t *testing.T) {
		st := NewSymbolTable()
		var reported []string

		targetCall := &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "unknown_func"},
			Arguments: []ast.Expression{closeIdent()},
		}

		conv := NewSeriesAccessConverter(st, "j", func(c *ast.CallExpression) string {
			if c == targetCall {
				return "ta_unknown_func_abc"
			}
			return ""
		})
		conv.ReportUnknownCall = func(name string) { reported = append(reported, name) }

		code, err := conv.ConvertExpression(targetCall)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != "ta_unknown_func_abcSeries.Get(j)" {
			t.Errorf("got %q, want series access", code)
		}
		if len(reported) != 0 {
			t.Errorf("should not report when temp var resolves, got: %v", reported)
		}
	})

	t.Run("unknown call in unary expression returns math.NaN() and reports gap", func(t *testing.T) {
		st := NewSymbolTable()
		var reported []string
		conv := NewSeriesAccessConverter(st, "j", nil)
		conv.ReportUnknownCall = func(name string) { reported = append(reported, name) }

		expr := &ast.UnaryExpression{
			Operator: "-",
			Argument: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "unknown_func"},
				Arguments: []ast.Expression{closeIdent()},
			},
		}

		code, err := conv.ConvertExpression(expr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != "-math.NaN()" {
			t.Errorf("got %q, want -math.NaN()", code)
		}
		if len(reported) != 1 || reported[0] != "unknown_func" {
			t.Errorf("reported gaps = %v, want [unknown_func]", reported)
		}
	})

	t.Run("unknown call in conditional expression branches degrade independently", func(t *testing.T) {
		st := NewSymbolTable()
		var reported []string
		conv := NewSeriesAccessConverter(st, "j", nil)
		conv.ReportUnknownCall = func(name string) { reported = append(reported, name) }

		expr := &ast.ConditionalExpression{
			Test: &ast.BinaryExpression{
				Left:     closeIdent(),
				Operator: ">",
				Right:    &ast.Literal{Value: 0.0},
			},
			Consequent: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "cond_true_gap"},
				Arguments: []ast.Expression{closeIdent()},
			},
			Alternate: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "cond_false_gap"},
				Arguments: []ast.Expression{closeIdent()},
			},
		}

		code, err := conv.ConvertExpression(expr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Count(code, "math.NaN()") < 2 {
			t.Errorf("both branches must degrade to math.NaN(), got: %s", code)
		}
		reportedNames := make(map[string]bool, len(reported))
		for _, name := range reported {
			reportedNames[name] = true
		}
		if !reportedNames["cond_true_gap"] || !reportedNames["cond_false_gap"] {
			t.Errorf("both branch gaps must be reported, got: %v", reported)
		}
	})
}
