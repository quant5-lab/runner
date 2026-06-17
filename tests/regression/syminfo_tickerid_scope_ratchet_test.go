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
}
