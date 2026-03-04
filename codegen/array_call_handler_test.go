package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// newGeneratorWithArrayVar returns a test generator with the named variable
// registered as "array_series".
func newGeneratorWithArrayVar(varName string) *generator {
	g := newTestGenerator()
	g.variables[varName] = "array_series"
	return g
}

// arrayGetCall constructs an array.get(arrayArg, indexArg) call.
func arrayGetCall(arrayArg ast.Expression, indexArg ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "array"},
			Property: &ast.Identifier{Name: "get"},
		},
		Arguments: []ast.Expression{arrayArg, indexArg},
	}
}

// arraySizeCall constructs an array.size(arrayArg) call.
func arraySizeCall(arrayArg ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "array"},
			Property: &ast.Identifier{Name: "size"},
		},
		Arguments: []ast.Expression{arrayArg},
	}
}

// historySubscript builds a computed MemberExpression: varName[offset].
func historySubscript(varName string, offset int) *ast.MemberExpression {
	return &ast.MemberExpression{
		Object:   &ast.Identifier{Name: varName},
		Property: &ast.Literal{Value: float64(offset)},
		Computed: true,
	}
}

// --- CanHandle ---

func TestArrayMethodCallHandler_CanHandle(t *testing.T) {
	h := &ArrayMethodCallHandler{}

	accept := []string{"array.get", "array.size"}
	for _, name := range accept {
		if !h.CanHandle(name) {
			t.Errorf("ArrayMethodCallHandler must accept %q", name)
		}
	}

	reject := []string{"array.new_float", "array.push", "array", "array.set", "ta.array", ""}
	for _, name := range reject {
		if h.CanHandle(name) {
			t.Errorf("ArrayMethodCallHandler must not accept %q", name)
		}
	}
}

// --- array.get ---

func TestArrayMethodCallHandler_Get_IdentifierCurrentBar(t *testing.T) {
	h := &ArrayMethodCallHandler{}
	g := newGeneratorWithArrayVar("levels")

	code, err := h.GenerateCode(g, arrayGetCall(
		&ast.Identifier{Name: "levels"},
		&ast.Literal{Value: float64(0)},
	))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != "levelsArraySeries.Elem(0, 0)" {
		t.Errorf("got: %s, want: levelsArraySeries.Elem(0, 0)", code)
	}
}

func TestArrayMethodCallHandler_Get_IdentifierNonZeroIndex(t *testing.T) {
	h := &ArrayMethodCallHandler{}
	g := newGeneratorWithArrayVar("levels")

	code, err := h.GenerateCode(g, arrayGetCall(
		&ast.Identifier{Name: "levels"},
		&ast.Literal{Value: float64(5)},
	))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != "levelsArraySeries.Elem(0, 5)" {
		t.Errorf("got: %s, want: levelsArraySeries.Elem(0, 5)", code)
	}
}

func TestArrayMethodCallHandler_Get_HistorySubscriptAccess(t *testing.T) {
	h := &ArrayMethodCallHandler{}
	g := newGeneratorWithArrayVar("levels")

	// array.get(levels[1], 0) — access previous bar's array element
	code, err := h.GenerateCode(g, arrayGetCall(
		historySubscript("levels", 1),
		&ast.Literal{Value: float64(0)},
	))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != "levelsArraySeries.Elem(1, 0)" {
		t.Errorf("got: %s, want: levelsArraySeries.Elem(1, 0)", code)
	}
}

func TestArrayMethodCallHandler_Get_HistorySubscriptLargerOffset(t *testing.T) {
	h := &ArrayMethodCallHandler{}
	g := newGeneratorWithArrayVar("pivots")

	// array.get(pivots[3], 2)
	code, err := h.GenerateCode(g, arrayGetCall(
		historySubscript("pivots", 3),
		&ast.Literal{Value: float64(2)},
	))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != "pivotsArraySeries.Elem(3, 2)" {
		t.Errorf("got: %s, want: pivotsArraySeries.Elem(3, 2)", code)
	}
}

func TestArrayMethodCallHandler_Get_DynamicIndexExpression(t *testing.T) {
	h := &ArrayMethodCallHandler{}
	g := newGeneratorWithArrayVar("levels")

	// array.get(levels, someVar)
	code, err := h.GenerateCode(g, arrayGetCall(
		&ast.Identifier{Name: "levels"},
		&ast.Identifier{Name: "someVar"},
	))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(code, "levelsArraySeries.Elem(0,") {
		t.Errorf("expected levelsArraySeries.Elem(0, ...) for dynamic index, got: %s", code)
	}
	if !strings.Contains(code, "int(") {
		t.Errorf("dynamic index must be cast to int, got: %s", code)
	}
}

func TestArrayMethodCallHandler_Get_WrongArgCountReturnsError(t *testing.T) {
	h := &ArrayMethodCallHandler{}
	g := newGeneratorWithArrayVar("levels")

	for _, argCount := range []int{0, 1, 3} {
		args := make([]ast.Expression, argCount)
		for i := range args {
			args[i] = &ast.Literal{Value: float64(i)}
		}
		call := &ast.CallExpression{
			Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "array"}, Property: &ast.Identifier{Name: "get"}},
			Arguments: args,
		}
		_, err := h.GenerateCode(g, call)
		if err == nil {
			t.Errorf("expected error for array.get with %d arguments, got nil", argCount)
		}
	}
}

func TestArrayMethodCallHandler_Get_NonArraySeriesVariableReturnsError(t *testing.T) {
	h := &ArrayMethodCallHandler{}
	g := newTestGenerator()
	g.variables["myVar"] = "function" // not array_series

	_, err := h.GenerateCode(g, arrayGetCall(
		&ast.Identifier{Name: "myVar"},
		&ast.Literal{Value: float64(0)},
	))
	if err == nil {
		t.Error("expected error when calling array.get on non-array-series variable")
	}
}

func TestArrayMethodCallHandler_Get_UnregisteredVariableReturnsError(t *testing.T) {
	h := &ArrayMethodCallHandler{}
	g := newTestGenerator() // levels not registered

	_, err := h.GenerateCode(g, arrayGetCall(
		&ast.Identifier{Name: "levels"},
		&ast.Literal{Value: float64(0)},
	))
	if err == nil {
		t.Error("expected error when array.get target variable is not registered")
	}
}

// --- array.size ---

func TestArrayMethodCallHandler_Size_Identifier(t *testing.T) {
	h := &ArrayMethodCallHandler{}
	g := newGeneratorWithArrayVar("levels")

	code, err := h.GenerateCode(g, arraySizeCall(&ast.Identifier{Name: "levels"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != "float64(len(levelsArraySeries.Get(0)))" {
		t.Errorf("got: %s, want: float64(len(levelsArraySeries.Get(0)))", code)
	}
}

func TestArrayMethodCallHandler_Size_HistorySubscript(t *testing.T) {
	h := &ArrayMethodCallHandler{}
	g := newGeneratorWithArrayVar("levels")

	// array.size(levels[2]) — size of previous bar's array
	code, err := h.GenerateCode(g, arraySizeCall(historySubscript("levels", 2)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != "float64(len(levelsArraySeries.Get(2)))" {
		t.Errorf("got: %s, want: float64(len(levelsArraySeries.Get(2)))", code)
	}
}

func TestArrayMethodCallHandler_Size_WrongArgCountReturnsError(t *testing.T) {
	h := &ArrayMethodCallHandler{}
	g := newGeneratorWithArrayVar("levels")

	for _, argCount := range []int{0, 2, 3} {
		args := make([]ast.Expression, argCount)
		for i := range args {
			args[i] = &ast.Identifier{Name: "levels"}
		}
		call := &ast.CallExpression{
			Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "array"}, Property: &ast.Identifier{Name: "size"}},
			Arguments: args,
		}
		_, err := h.GenerateCode(g, call)
		if err == nil {
			t.Errorf("expected error for array.size with %d arguments, got nil", argCount)
		}
	}
}

// --- isArraySeriesVariable ---

func TestIsArraySeriesVariable(t *testing.T) {
	g := newTestGenerator()
	g.variables["levels"] = "array_series"
	g.variables["ema"] = "function"
	g.variables["label"] = "string"

	if !g.isArraySeriesVariable("levels") {
		t.Error("levels registered as array_series must return true")
	}
	if g.isArraySeriesVariable("ema") {
		t.Error("ema registered as function must return false")
	}
	if g.isArraySeriesVariable("label") {
		t.Error("label registered as string must return false")
	}
	if g.isArraySeriesVariable("notRegistered") {
		t.Error("unregistered variable must return false")
	}
}

// --- resolveSubscriptOffset ---

func TestResolveSubscriptOffset_IntLiteral(t *testing.T) {
	offset, isDynamic, _, err := resolveSubscriptOffset(&ast.Literal{Value: 3})
	if err != nil || isDynamic || offset != 3 {
		t.Errorf("int literal: got offset=%d isDynamic=%v err=%v, want 3, false, nil", offset, isDynamic, err)
	}
}

func TestResolveSubscriptOffset_FloatLiteral(t *testing.T) {
	offset, isDynamic, _, err := resolveSubscriptOffset(&ast.Literal{Value: float64(7)})
	if err != nil || isDynamic || offset != 7 {
		t.Errorf("float64 literal: got offset=%d isDynamic=%v err=%v, want 7, false, nil", offset, isDynamic, err)
	}
}

func TestResolveSubscriptOffset_DynamicExpression(t *testing.T) {
	_, isDynamic, _, err := resolveSubscriptOffset(&ast.Identifier{Name: "n"})
	if err != nil || !isDynamic {
		t.Errorf("identifier: got isDynamic=%v err=%v, want true, nil", isDynamic, err)
	}
}

// --- Registration ---

func TestArrayMethodCallHandler_RegisteredInRouter(t *testing.T) {
	router := NewCallExpressionRouter()
	h := &ArrayMethodCallHandler{}

	for _, name := range []string{"array.get", "array.size"} {
		found := false
		for _, handler := range router.handlers {
			if handler.CanHandle(name) {
				_, ok := handler.(*ArrayMethodCallHandler)
				if ok {
					found = true
					_ = h // ensure type is used
					break
				}
			}
		}
		if !found {
			t.Errorf("ArrayMethodCallHandler not found in router for %q", name)
		}
	}
}
