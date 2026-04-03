package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// pivotCallee builds a ta.pivot_point_levels callee expression.
func pivotCallee() ast.Expression {
	return &ast.MemberExpression{
		Object:   &ast.Identifier{Name: "ta"},
		Property: &ast.Identifier{Name: "pivot_point_levels"},
	}
}

// pivotCall2 builds a 2-argument call (type, anchor).
func pivotCall2(pivotType string, anchor ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee:    pivotCallee(),
		Arguments: []ast.Expression{&ast.Literal{Value: pivotType}, anchor},
	}
}

// pivotCall3 builds a 3-argument call (type, anchor, developing).
func pivotCall3(pivotType string, anchor, developing ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee:    pivotCallee(),
		Arguments: []ast.Expression{&ast.Literal{Value: pivotType}, anchor, developing},
	}
}

// boolLit is a convenience for boolean AST literals.
func boolLit(v bool) *ast.Literal { return &ast.Literal{Value: v} }

// --- CanHandle ---

func TestPivotPointLevelsHandler_CanHandle(t *testing.T) {
	h := &PivotPointLevelsHandler{}

	accept := []string{"ta.pivot_point_levels", "pivot_point_levels"}
	for _, name := range accept {
		if !h.CanHandle(name) {
			t.Errorf("PivotPointLevelsHandler must accept %q", name)
		}
	}

	reject := []string{"ta.pivotpoints", "pivot_points", "ta.pivot", "ta.pivot_point_levels_2", ""}
	for _, name := range reject {
		if h.CanHandle(name) {
			t.Errorf("PivotPointLevelsHandler must not accept %q", name)
		}
	}
}

// --- GetInternalSeriesNames ---

func TestPivotPointLevelsHandler_GetInternalSeriesNamesLayout(t *testing.T) {
	h := &PivotPointLevelsHandler{}
	names, err := h.GetInternalSeriesNames("levels", nil)
	if err != nil {
		t.Fatalf("GetInternalSeriesNames() error = %v", err)
	}
	if len(names) != 4 {
		t.Fatalf("expected 4 internal series, got %d", len(names))
	}

	want := []string{"_levels_periodH", "_levels_periodL", "_levels_periodC", "_levels_periodO"}
	for i, w := range want {
		if names[i] != w {
			t.Errorf("names[%d] = %q, want %q", i, names[i], w)
		}
	}
}

func TestPivotPointLevelsHandler_GetInternalSeriesNamesReflectVarName(t *testing.T) {
	h := &PivotPointLevelsHandler{}
	names, err := h.GetInternalSeriesNames("weeklyPivot", nil)
	if err != nil {
		t.Fatalf("GetInternalSeriesNames() error = %v", err)
	}

	for _, name := range names {
		if !strings.HasPrefix(name, "_weeklyPivot_") {
			t.Errorf("series name %q does not reflect varName 'weeklyPivot'", name)
		}
	}
}

// --- extractPivotArguments ---

func TestExtractPivotArguments_TwoArgs(t *testing.T) {
	call := pivotCall2("Traditional", boolLit(true))
	typeExpr, anchorExpr, developingExpr, err := extractPivotArguments(call)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if typeExpr == nil || anchorExpr == nil {
		t.Error("typeExpr and anchorExpr must not be nil for 2-arg form")
	}
	if developingExpr != nil {
		t.Error("developingExpr must be nil for 2-arg form")
	}
}

func TestExtractPivotArguments_ThreeArgs(t *testing.T) {
	call := pivotCall3("Traditional", boolLit(true), boolLit(false))
	typeExpr, anchorExpr, developingExpr, err := extractPivotArguments(call)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if typeExpr == nil || anchorExpr == nil || developingExpr == nil {
		t.Error("all three expressions must be non-nil for 3-arg form")
	}
}

func TestExtractPivotArguments_WrongArgCountReturnsError(t *testing.T) {
	for _, argCount := range []int{0, 1, 4, 5} {
		args := make([]ast.Expression, argCount)
		for i := range args {
			args[i] = &ast.Literal{Value: float64(i)}
		}
		call := &ast.CallExpression{Callee: pivotCallee(), Arguments: args}
		_, _, _, err := extractPivotArguments(call)
		if err == nil {
			t.Errorf("expected error for %d arguments, got nil", argCount)
		}
	}
}

// --- resolveDevelopingCode ---

func TestResolveDevelopingCode_NilDefaultsFalse(t *testing.T) {
	g := newTestGenerator()
	code, err := resolveDevelopingCode(g, nil)
	if err != nil || code != "false" {
		t.Errorf("nil developing: got %q err=%v, want \"false\", nil", code, err)
	}
}

func TestResolveDevelopingCode_LiteralTrue(t *testing.T) {
	g := newTestGenerator()
	code, err := resolveDevelopingCode(g, boolLit(true))
	if err != nil || code != "true" {
		t.Errorf("literal true: got %q err=%v, want \"true\", nil", code, err)
	}
}

func TestResolveDevelopingCode_LiteralFalse(t *testing.T) {
	g := newTestGenerator()
	code, err := resolveDevelopingCode(g, boolLit(false))
	if err != nil || code != "false" {
		t.Errorf("literal false: got %q err=%v, want \"false\", nil", code, err)
	}
}

// --- resolvePivotTypeCode ---

func TestResolvePivotTypeCode_StringLiteralProducesTypedCast(t *testing.T) {
	g := newTestGenerator()
	code, err := resolvePivotTypeCode(g, &ast.Literal{Value: "Traditional"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != `ta.PivotType("Traditional")` {
		t.Errorf("got %q, want ta.PivotType(\"Traditional\")", code)
	}
}

func TestResolvePivotTypeCode_StringLiteralsForAllSixTypes(t *testing.T) {
	g := newTestGenerator()
	types := []string{"Traditional", "Fibonacci", "Woodie", "Classic", "DM", "Camarilla"}
	for _, pt := range types {
		code, err := resolvePivotTypeCode(g, &ast.Literal{Value: pt})
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", pt, err)
		}
		want := `ta.PivotType("` + pt + `")`
		if code != want {
			t.Errorf("%s: got %q, want %q", pt, code, want)
		}
	}
}

// --- buildPivotPerBarBlock ---

func TestBuildPivotPerBarBlock_ContainsAnchorBranch(t *testing.T) {
	code := buildPivotPerBarBlock("levels", `ta.PivotType("Traditional")`, "_anchor", "false")

	if !strings.Contains(code, "if _anchor {") {
		t.Error("generated block must contain anchor conditional branch")
	}
}

func TestBuildPivotPerBarBlock_OnAnchorComputesLevelsAndResets(t *testing.T) {
	code := buildPivotPerBarBlock("levels", `ta.PivotType("Traditional")`, "_anchor", "false")

	if !strings.Contains(code, "levelsArraySeries.Set(ta.ComputePivotLevels(") {
		t.Error("anchor branch must call ta.ComputePivotLevels and store result")
	}
	// Accumulator reset: all four internal series updated from current bar
	for _, field := range []string{"_levels_periodHSeries.Set(", "_levels_periodLSeries.Set(", "_levels_periodCSeries.Set(", "_levels_periodOSeries.Set("} {
		if !strings.Contains(code, field) {
			t.Errorf("anchor branch must reset accumulator series %q", field)
		}
	}
}

func TestBuildPivotPerBarBlock_PrevStateReadViaGet1(t *testing.T) {
	code := buildPivotPerBarBlock("levels", `ta.PivotType("Traditional")`, "_anchor", "false")

	// Get(1) is the NaN sentinel for "no previous state" on bar 0
	if !strings.Contains(code, "_levels_periodHSeries.Get(1)") {
		t.Error("block must read previous H state via Get(1)")
	}
	if !strings.Contains(code, "_levels_periodLSeries.Get(1)") {
		t.Error("block must read previous L state via Get(1)")
	}
}

func TestBuildPivotPerBarBlock_OffAnchorDevelopingFalseCarriesForward(t *testing.T) {
	code := buildPivotPerBarBlock("levels", `ta.PivotType("Traditional")`, "_anchor", "false")

	// Off-anchor, not developing: carry previous bar's levels forward
	if !strings.Contains(code, "levelsArraySeries.Get(1)") {
		t.Error("carry-forward path must read previous bar's array via Get(1)")
	}
	if !strings.Contains(code, "ta.NaNLevels()") {
		t.Error("fallback when no prior levels must emit ta.NaNLevels()")
	}
}

func TestBuildPivotPerBarBlock_OffAnchorDevelopingTrueRecomputes(t *testing.T) {
	code := buildPivotPerBarBlock("levels", `ta.PivotType("Traditional")`, "_anchor", "true")

	// With developing=true the else-branch must always recompute from accumulator
	if !strings.Contains(code, "_developing := true") {
		t.Error("block must emit '_developing := true' when developing is true")
	}
}

func TestBuildPivotPerBarBlock_VarNamesReflectVarName(t *testing.T) {
	code := buildPivotPerBarBlock("myPivot", `ta.PivotType("Camarilla")`, "_anchor", "false")

	for _, want := range []string{
		"myPivotArraySeries",
		"_myPivot_periodH",
		"_myPivot_prevH",
		"_myPivot_newH",
	} {
		if !strings.Contains(code, want) {
			t.Errorf("expected %q in generated block for varName 'myPivot'", want)
		}
	}
}

func TestBuildPivotPerBarBlock_IsNaNGuardForFirstBar(t *testing.T) {
	code := buildPivotPerBarBlock("levels", `ta.PivotType("Traditional")`, "_anchor", "false")

	// math.IsNaN check initialises accumulator on bar 0 without extra pre-loop code
	if !strings.Contains(code, "math.IsNaN(") {
		t.Error("block must guard first-bar accumulator init with math.IsNaN check")
	}
}

// --- GenerateCode (integration) ---

func TestPivotPointLevelsHandler_GenerateCode_TwoArgForm(t *testing.T) {
	h := &PivotPointLevelsHandler{}
	g := newTestGenerator()

	code, err := h.GenerateCode(g, "levels", pivotCall2("Traditional", boolLit(true)))
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	NewCodeVerifier(code, t).
		MustContain("levelsArraySeries", "ta.ComputePivotLevels", "ta.NaNLevels()")
}

func TestPivotPointLevelsHandler_GenerateCode_ThreeArgDevelopingTrue(t *testing.T) {
	h := &PivotPointLevelsHandler{}
	g := newTestGenerator()

	code, err := h.GenerateCode(g, "weekly", pivotCall3("Fibonacci", boolLit(true), boolLit(true)))
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	NewCodeVerifier(code, t).
		MustContain("weeklyArraySeries", "_developing := true")
}

func TestPivotPointLevelsHandler_GenerateCode_WrongArgCountReturnsError(t *testing.T) {
	h := &PivotPointLevelsHandler{}
	g := newTestGenerator()

	for _, argCount := range []int{0, 1, 4} {
		args := make([]ast.Expression, argCount)
		for i := range args {
			args[i] = &ast.Literal{Value: float64(i)}
		}
		call := &ast.CallExpression{Callee: pivotCallee(), Arguments: args}
		_, err := h.GenerateCode(g, "levels", call)
		if err == nil {
			t.Errorf("expected error for %d arguments, got nil", argCount)
		}
	}
}

// --- Registration ---

func TestPivotPointLevelsHandler_RegisteredInTAFunctionRegistry(t *testing.T) {
	reg := NewTAFunctionRegistry()

	for _, name := range []string{"ta.pivot_point_levels", "pivot_point_levels"} {
		if reg.FindHandler(name) == nil {
			t.Errorf("TAFunctionRegistry must route %q via PivotPointLevelsHandler", name)
		}
	}
}

func TestPivotPointLevelsHandler_RegisteredInCompositeIndicatorRegistry(t *testing.T) {
	reg := NewCompositeIndicatorRegistry()
	reg.Register("ta.pivot_point_levels", &PivotPointLevelsHandler{})
	reg.Register("pivot_point_levels", &PivotPointLevelsHandler{})

	call := pivotCall2("Traditional", boolLit(true))
	for _, funcName := range []string{"ta.pivot_point_levels", "pivot_point_levels"} {
		names := reg.GetInternalSeriesNames(funcName, "levels", call)
		if len(names) != 4 {
			t.Errorf("%s: compositeIndicatorRegistry returned %d internal series, want 4", funcName, len(names))
		}
	}
}
