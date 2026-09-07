package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestBuildArrowSymbolTable_LocalsRegisteredAsSeries(t *testing.T) {
	table := buildArrowSymbolTable(nil, []string{"alpha", "beta"}, nil, nil)

	for _, name := range []string{"alpha", "beta"} {
		if !table.IsSeries(name) {
			t.Errorf("local variable %q must be registered as VariableTypeSeries", name)
		}
	}
}

func TestBuildArrowSymbolTable_InheritsMainScope(t *testing.T) {
	main := NewSymbolTable()
	main.Register("outerSeries", VariableTypeSeries)
	main.Register("outerScalar", VariableTypeScalar)

	table := buildArrowSymbolTable(main, []string{"localVar"}, nil, nil)

	if !table.IsSeries("outerSeries") {
		t.Error("outer series variable must be inherited")
	}
	if !table.IsScalar("outerScalar") {
		t.Error("outer scalar variable must be inherited")
	}
	if !table.IsSeries("localVar") {
		t.Error("local variable declared in arrow body must be VariableTypeSeries")
	}
}

func TestBuildArrowSymbolTable_MainScopeIsNotMutated(t *testing.T) {
	main := NewSymbolTable()
	main.Register("existing", VariableTypeSeries)

	buildArrowSymbolTable(main, []string{"local"}, nil, nil)

	if main.Lookup("local") != VariableTypeUnknown {
		t.Error("buildArrowSymbolTable must not mutate the main-scope table")
	}
}

func TestBuildArrowSymbolTable_NilMainScopeCreatesValidTable(t *testing.T) {
	table := buildArrowSymbolTable(nil, []string{"x"}, nil, nil)

	if table == nil {
		t.Fatal("returned table must not be nil when mainScope is nil")
	}
	if !table.IsSeries("x") {
		t.Error("local variable must be VariableTypeSeries even when mainScope is nil")
	}
}

func TestBuildArrowSymbolTable_ParamUsageClassification(t *testing.T) {
	params := []ast.Identifier{{Name: "src"}, {Name: "len"}, {Name: "sp"}}
	usage := map[string]ParameterUsageType{
		"src": ParameterUsageTASource,
		"len": ParameterUsageScalar,
		"sp":  ParameterUsageSeries,
	}

	table := buildArrowSymbolTable(nil, nil, params, usage)

	if !table.IsSeries("src") {
		t.Error("ParameterUsageTASource must register as VariableTypeSeries")
	}
	if !table.IsSeries("sp") {
		t.Error("ParameterUsageSeries must register as VariableTypeSeries")
	}
	if !table.IsScalar("len") {
		t.Error("ParameterUsageScalar must register as VariableTypeScalar")
	}
}

func TestBuildArrowSymbolTable_DefaultParamUsageIsScalar(t *testing.T) {
	params := []ast.Identifier{{Name: "n"}}
	usage := map[string]ParameterUsageType{}

	table := buildArrowSymbolTable(nil, nil, params, usage)

	if !table.IsScalar("n") {
		t.Errorf("parameter with no explicit usage must default to VariableTypeScalar, got IsSeries=%v", table.IsSeries("n"))
	}
}

func TestBuildArrowSymbolTable_LocalsOverrideMainScopeType(t *testing.T) {
	main := NewSymbolTable()
	main.Register("v", VariableTypeScalar)

	table := buildArrowSymbolTable(main, []string{"v"}, nil, nil)

	if !table.IsSeries("v") {
		t.Error("arrow-local declaration must override inherited scalar type with VariableTypeSeries")
	}
}

func TestActiveSymbolTable_ReturnsArrowTableWhenSet(t *testing.T) {
	gen := newTestGenerator()
	arrowTable := NewSymbolTable()
	arrowTable.Register("arrowLocal", VariableTypeSeries)
	gen.arrowSymbolTable = arrowTable

	got := activeSymbolTable(gen)

	if !got.IsSeries("arrowLocal") {
		t.Error("activeSymbolTable must return arrowSymbolTable when it is set")
	}
}

func TestActiveSymbolTable_FallsBackToMainWhenArrowTableNil(t *testing.T) {
	gen := newTestGenerator()
	gen.symbolTable = NewSymbolTable()
	gen.symbolTable.Register("mainEntry", VariableTypeSeries)
	gen.arrowSymbolTable = nil

	got := activeSymbolTable(gen)

	if !got.IsSeries("mainEntry") {
		t.Error("activeSymbolTable must return symbolTable when arrowSymbolTable is nil")
	}
}

func TestActiveSymbolTable_ArrowTableIsolatedFromMain(t *testing.T) {
	gen := newTestGenerator()
	gen.symbolTable = NewSymbolTable()
	gen.symbolTable.Register("mainOnly", VariableTypeSeries)

	arrowTable := NewSymbolTable()
	arrowTable.Register("arrowOnly", VariableTypeSeries)
	gen.arrowSymbolTable = arrowTable

	got := activeSymbolTable(gen)

	if got.IsSeries("mainOnly") {
		t.Error("activeSymbolTable must not expose main-scope entries when arrowSymbolTable is set independently")
	}
	if !got.IsSeries("arrowOnly") {
		t.Error("activeSymbolTable must expose arrow-scope entries")
	}
}
