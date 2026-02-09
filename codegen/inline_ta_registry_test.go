package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/codegen/series_naming"
)

/* TestInlineTAIIFERegistry_RegisterWithBareAlias validates dual registration pattern */
func TestInlineTAIIFERegistry_RegisterWithBareAlias(t *testing.T) {
	registry := &InlineTAIIFERegistry{
		generators:           make(map[string]InlineTAIIFEGenerator),
		dualPeriodGenerators: make(map[string]InlineTADualPeriodGenerator),
	}

	gen := &SMAIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()}
	registry.RegisterWithBareAlias("ta.sma", gen)

	if !registry.IsSupported("ta.sma") {
		t.Error("RegisterWithBareAlias should register namespaced form")
	}
	if !registry.IsSupported("sma") {
		t.Error("RegisterWithBareAlias should register bare form")
	}
}

/* TestInlineTAIIFERegistry_RegisterWithBareAlias_NoNamespace validates bare-only registration */
func TestInlineTAIIFERegistry_RegisterWithBareAlias_NoNamespace(t *testing.T) {
	registry := &InlineTAIIFERegistry{
		generators:           make(map[string]InlineTAIIFEGenerator),
		dualPeriodGenerators: make(map[string]InlineTADualPeriodGenerator),
	}

	gen := &SMAIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()}
	registry.RegisterWithBareAlias("nz", gen)

	if !registry.IsSupported("nz") {
		t.Error("Bare function should be registered")
	}

	if registry.IsSupported(".nz") {
		t.Error("Should not create spurious registration with leading dot")
	}
}

/* TestInlineTAIIFERegistry_RegisterWithBareAlias_MultiNamespace validates deep namespace extraction */
func TestInlineTAIIFERegistry_RegisterWithBareAlias_MultiNamespace(t *testing.T) {
	registry := &InlineTAIIFERegistry{
		generators:           make(map[string]InlineTAIIFEGenerator),
		dualPeriodGenerators: make(map[string]InlineTADualPeriodGenerator),
	}

	gen := &SMAIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()}
	registry.RegisterWithBareAlias("math.stats.avg", gen)

	if !registry.IsSupported("math.stats.avg") {
		t.Error("Should register full namespaced form")
	}
	if !registry.IsSupported("avg") {
		t.Error("Should register bare form extracted from last dot")
	}
	if registry.IsSupported("stats.avg") {
		t.Error("Should not register intermediate namespace forms")
	}
}

/* TestInlineTAIIFERegistry_RegisterWithBareAlias_GeneratorIdentity validates same generator instance */
func TestInlineTAIIFERegistry_RegisterWithBareAlias_GeneratorIdentity(t *testing.T) {
	registry := &InlineTAIIFERegistry{
		generators:           make(map[string]InlineTAIIFEGenerator),
		dualPeriodGenerators: make(map[string]InlineTADualPeriodGenerator),
	}

	gen := &SMAIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()}
	registry.RegisterWithBareAlias("ta.sma", gen)

	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(10)

	code1, ok1 := registry.Generate("ta.sma", accessor, period, "test")
	code2, ok2 := registry.Generate("sma", accessor, period, "test")

	if !ok1 || !ok2 {
		t.Fatal("Both forms should generate code")
	}

	if code1 != code2 {
		t.Error("Both namespaced and bare forms should produce identical code (same generator instance)")
	}
}

/* TestInlineTAIIFERegistry_RegisterDualPeriodWithBareAlias validates pivot function registration */
func TestInlineTAIIFERegistry_RegisterDualPeriodWithBareAlias(t *testing.T) {
	registry := &InlineTAIIFERegistry{
		generators:           make(map[string]InlineTAIIFEGenerator),
		dualPeriodGenerators: make(map[string]InlineTADualPeriodGenerator),
	}

	gen := &PivotHighIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()}
	registry.RegisterDualPeriodWithBareAlias("ta.pivothigh", gen)

	if !registry.IsSupported("ta.pivothigh") {
		t.Error("RegisterDualPeriodWithBareAlias should register namespaced form")
	}
	if !registry.IsSupported("pivothigh") {
		t.Error("RegisterDualPeriodWithBareAlias should register bare form")
	}
}

/* TestInlineTAIIFERegistry_RegisterDualPeriodWithBareAlias_GeneratorIdentity validates same generator instance */
func TestInlineTAIIFERegistry_RegisterDualPeriodWithBareAlias_GeneratorIdentity(t *testing.T) {
	registry := &InlineTAIIFERegistry{
		generators:           make(map[string]InlineTAIIFEGenerator),
		dualPeriodGenerators: make(map[string]InlineTADualPeriodGenerator),
	}

	gen := &PivotHighIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()}
	registry.RegisterDualPeriodWithBareAlias("ta.pivothigh", gen)

	accessor := NewArrowFunctionParameterAccessor("src")
	left := NewConstantPeriod(5)
	right := NewConstantPeriod(5)

	code1, ok1 := registry.GenerateDualPeriod("ta.pivothigh", accessor, left, right, "test")
	code2, ok2 := registry.GenerateDualPeriod("pivothigh", accessor, left, right, "test")

	if !ok1 || !ok2 {
		t.Fatal("Both forms should generate code")
	}

	if code1 != code2 {
		t.Error("Both namespaced and bare forms should produce identical code (same generator instance)")
	}
}

/* TestInlineTAIIFERegistry_MixedRegistrations validates registry supports both registration types */
func TestInlineTAIIFERegistry_MixedRegistrations(t *testing.T) {
	registry := &InlineTAIIFERegistry{
		generators:           make(map[string]InlineTAIIFEGenerator),
		dualPeriodGenerators: make(map[string]InlineTADualPeriodGenerator),
	}

	smaGen := &SMAIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()}
	pivotGen := &PivotHighIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()}

	registry.RegisterWithBareAlias("ta.sma", smaGen)
	registry.RegisterDualPeriodWithBareAlias("ta.pivothigh", pivotGen)

	if !registry.IsSupported("ta.sma") || !registry.IsSupported("sma") {
		t.Error("Single-period generator should be registered")
	}
	if !registry.IsSupported("ta.pivothigh") || !registry.IsSupported("pivothigh") {
		t.Error("Dual-period generator should be registered")
	}
}

/* TestInlineTAIIFERegistry_DefaultRegistrations validates all default functions registered */
func TestInlineTAIIFERegistry_DefaultRegistrations(t *testing.T) {
	registry := NewInlineTAIIFERegistry()

	expectedSinglePeriod := []string{
		"ta.sma", "sma",
		"ta.wma", "wma",
		"ta.ema", "ema",
		"ta.rma", "rma",
		"ta.rsi", "rsi",
		"ta.atr", "atr",
		"ta.stdev", "stdev",
		"ta.highest", "highest",
		"ta.lowest", "lowest",
		"ta.change", "change",
		"ta.linreg", "linreg",
		"ta.swma", "swma",
		"ta.sum", "sum",
		"math.sum",
	}

	for _, funcName := range expectedSinglePeriod {
		if !registry.IsSupported(funcName) {
			t.Errorf("Default registration missing: %q", funcName)
		}
	}

	expectedDualPeriod := []string{
		"ta.pivothigh", "pivothigh",
		"ta.pivotlow", "pivotlow",
	}

	for _, funcName := range expectedDualPeriod {
		if !registry.IsSupported(funcName) {
			t.Errorf("Default dual-period registration missing: %q", funcName)
		}
	}
}

/* TestInlineTAIIFERegistry_MathSumSpecialCase validates math.sum shares ta.sum generator */
func TestInlineTAIIFERegistry_MathSumSpecialCase(t *testing.T) {
	registry := NewInlineTAIIFERegistry()

	accessor := NewArrowFunctionParameterAccessor("src")
	period := NewConstantPeriod(10)

	mathCode, mathOk := registry.Generate("math.sum", accessor, period, "test")
	taCode, taOk := registry.Generate("ta.sum", accessor, period, "test")
	bareCode, bareOk := registry.Generate("sum", accessor, period, "test")

	if !mathOk || !taOk || !bareOk {
		t.Fatal("All three sum variants should be supported")
	}

	if mathCode != taCode || mathCode != bareCode {
		t.Error("math.sum, ta.sum, and sum should produce identical code (shared generator)")
	}

	if !strings.Contains(mathCode, "sum +=") {
		t.Error("Sum generator should accumulate values")
	}
	if strings.Contains(mathCode, "sum / ") || strings.Contains(mathCode, "sum/") {
		t.Error("Sum generator must not divide (distinguishes from SMA)")
	}
}
