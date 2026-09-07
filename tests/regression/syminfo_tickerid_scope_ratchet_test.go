package regression

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSyminfoTickerid_TemplateDeclarationPlacement guards against the
// syminfo_tickerid declaration drifting back into a function body.
//
// The non-arrow resolver emits "syminfo_tickerid" from executeStrategy, UDFs,
// and IIFEs — Go functions where a main()-local variable is invisible. Package
// scope is the only scope that dominates every emission site.
func TestSyminfoTickerid_TemplateDeclarationPlacement(t *testing.T) {
	root := projectRootFromCwd()

	tmplBytes, err := os.ReadFile(filepath.Join(root, "template", "main.go.tmpl"))
	if err != nil {
		t.Fatalf("read template: %v", err)
	}
	tmpl := string(tmplBytes)

	declIdx := strings.Index(tmpl, "var syminfo_tickerid string")
	if declIdx == -1 {
		t.Fatal("var syminfo_tickerid string not found in template/main.go.tmpl")
	}

	strategyFuncIdx := strings.Index(tmpl, "{{STRATEGY_FUNC}}")
	if strategyFuncIdx == -1 {
		t.Fatal("{{STRATEGY_FUNC}} placeholder not found in template")
	}

	if declIdx > strategyFuncIdx {
		t.Errorf(
			"var syminfo_tickerid string appears after {{STRATEGY_FUNC}} (offset %d > %d) — "+
				"must be at package scope before the strategy function block",
			declIdx, strategyFuncIdx,
		)
	}

	if strings.Contains(tmpl[:declIdx], "\nfunc ") {
		t.Error("var syminfo_tickerid string appears after a func declaration — must be at package scope")
	}
}

// TestSyminfoTickerid_PackageScopeInGeneratedBinary verifies that assigning
// syminfo.tickerid to a string variable and passing it to security() compiles
// without error and that the declaration is at package scope in the generated Go.
//
// This pattern (ticker-via-string-variable → security symbol) requires the Go
// name to be visible from executeStrategy, where the non-arrow resolver emits it.
func TestSyminfoTickerid_PackageScopeInGeneratedBinary(t *testing.T) {
	root := projectRootFromCwd()
	tmpDir := t.TempDir()

	pine := `//@version=4
strategy("ticker-string-security", overlay=true)
sym = syminfo.tickerid
sz = security(sym, "60", close)
plot(sz)
`
	built, ok := codegenAndBuild(t, tmpDir, "ticker_string_security", pine, root)
	if !ok {
		t.Fatal("codegen/build failed — syminfo_tickerid scope regression suspected")
	}

	src, err := os.ReadFile(built.GeneratedPath)
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	assertSyminfoAtPackageScope(t, string(src))
}

// assertSyminfoAtPackageScope fails if "var syminfo_tickerid string" is missing,
// duplicated, or appears after the first top-level func in the generated source.
func assertSyminfoAtPackageScope(t *testing.T, generated string) {
	t.Helper()

	declIdx := strings.Index(generated, "var syminfo_tickerid string")
	if declIdx == -1 {
		t.Fatal("var syminfo_tickerid string not found in generated code")
	}

	firstFuncIdx := strings.Index(generated, "\nfunc ")
	if firstFuncIdx == -1 {
		t.Fatal("no top-level func declaration found in generated code")
	}

	if declIdx > firstFuncIdx {
		end := min(firstFuncIdx+80, len(generated))
		t.Errorf(
			"syminfo_tickerid declared inside a function (offset %d > first func at %d); "+
				"first func context: %q — keep var syminfo_tickerid at package scope in template/main.go.tmpl",
			declIdx, firstFuncIdx, generated[firstFuncIdx:end],
		)
	}

	if count := strings.Count(generated, "var syminfo_tickerid string"); count != 1 {
		t.Errorf("expected exactly 1 'var syminfo_tickerid string', got %d", count)
	}

	const assignment = "syminfo_tickerid = *symbolFlag"
	if !strings.Contains(generated, assignment) {
		t.Errorf("assignment %q not found in generated code — syminfo_tickerid will be empty string at runtime", assignment)
	}
}

// TestSyminfoTickerid_TemplateAssignmentOrdering verifies that the template
// assigns syminfo_tickerid its runtime value in the correct position within main:
// after flag.Parse() (so *symbolFlag is populated) and before executeStrategy()
// (so the value is visible to the strategy execution phase).
func TestSyminfoTickerid_TemplateAssignmentOrdering(t *testing.T) {
	root := projectRootFromCwd()

	tmplBytes, err := os.ReadFile(filepath.Join(root, "template", "main.go.tmpl"))
	if err != nil {
		t.Fatalf("read template: %v", err)
	}
	tmpl := string(tmplBytes)

	const assignment = "syminfo_tickerid = *symbolFlag"
	assignIdx := strings.Index(tmpl, assignment)
	if assignIdx == -1 {
		t.Fatalf("%q not found in template/main.go.tmpl", assignment)
	}

	const parseMark = "flag.Parse()"
	parseIdx := strings.Index(tmpl, parseMark)
	if parseIdx == -1 {
		t.Fatalf("%q not found in template/main.go.tmpl", parseMark)
	}

	const execMark = "executeStrategy("
	execIdx := strings.Index(tmpl, execMark)
	if execIdx == -1 {
		t.Fatalf("%q not found in template/main.go.tmpl", execMark)
	}

	if assignIdx < parseIdx {
		t.Errorf("syminfo_tickerid assigned before flag.Parse() (offsets: assign=%d parse=%d) — *symbolFlag would be empty", assignIdx, parseIdx)
	}
	if assignIdx > execIdx {
		t.Errorf("syminfo_tickerid assigned after executeStrategy( (offsets: assign=%d exec=%d) — strategy runs before the symbol is set", assignIdx, execIdx)
	}
}

// TestSyminfoTickerid_AllIdentityAliasesCompileAndReach verifies that every
// syminfo property that resolves to the syminfo_tickerid package variable
// (tickerid, ticker, description) compiles successfully, and that the generated
// source declares the variable at package scope in every case.
//
// tickerid and ticker are valid security() symbol arguments; description is a
// string value and is tested via a standalone variable assignment.
func TestSyminfoTickerid_AllIdentityAliasesCompileAndReach(t *testing.T) {
	root := projectRootFromCwd()

	cases := []struct {
		prop string
		pine string
	}{
		{
			prop: "tickerid",
			pine: `//@version=4
strategy("syminfo-tickerid", overlay=true)
sz = security(syminfo.tickerid, "D", close)
plot(sz)
`,
		},
		{
			prop: "ticker",
			pine: `//@version=4
strategy("syminfo-ticker", overlay=true)
sz = security(syminfo.ticker, "D", close)
plot(sz)
`,
		},
		{
			prop: "description",
			pine: `//@version=4
strategy("syminfo-description", overlay=true)
s = syminfo.description
plot(close)
`,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.prop, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			built, ok := codegenAndBuild(t, tmpDir, "syminfo_alias_"+tc.prop, tc.pine, root)
			if !ok {
				t.Fatalf("syminfo.%s → syminfo_tickerid: codegen/build failed", tc.prop)
			}

			src, err := os.ReadFile(built.GeneratedPath)
			if err != nil {
				t.Fatalf("read generated file: %v", err)
			}
			assertSyminfoAtPackageScope(t, string(src))
		})
	}
}

// TestSyminfoTickerid_VisibleFromUDFBody verifies that syminfo.tickerid used
// inside a user-defined function body compiles without error. The UDF is emitted
// as a top-level Go function placed before executeStrategy() in the generated
// binary; it references syminfo_tickerid, which must therefore be at package scope
// rather than local to any single function.
func TestSyminfoTickerid_VisibleFromUDFBody(t *testing.T) {
	root := projectRootFromCwd()
	tmpDir := t.TempDir()

	pine := `//@version=4
strategy("udf-syminfo-scope", overlay=true)
my_sym() =>
    syminfo.tickerid
sz = security(my_sym(), "D", close)
plot(sz)
`
	built, ok := codegenAndBuild(t, tmpDir, "udf_syminfo_scope", pine, root)
	if !ok {
		t.Fatal("syminfo.tickerid inside UDF body: codegen/build failed — package-scope visibility broken")
	}

	src, err := os.ReadFile(built.GeneratedPath)
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	assertSyminfoAtPackageScope(t, string(src))
}
