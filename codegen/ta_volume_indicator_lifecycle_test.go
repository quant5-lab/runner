package codegen

import (
	"strings"
	"testing"
)

// ── Spec registry ─────────────────────────────────────────────────────────────

func TestVolumeIndicatorMemberKeys_AllEightPresent(t *testing.T) {
	keys := VolumeIndicatorMemberKeys()
	if len(keys) != 8 {
		t.Fatalf("expected 8 member keys, got %d", len(keys))
	}
	expected := map[string]bool{
		"ta.obv": true, "ta.accdist": true, "ta.pvt": true, "ta.iii": true,
		"ta.wvad": true, "ta.nvi": true, "ta.pvi": true, "ta.wad": true,
	}
	got := make(map[string]bool, len(keys))
	for _, k := range keys {
		if !expected[k] {
			t.Errorf("unexpected member key: %q", k)
		}
		got[k] = true
	}
	for k := range expected {
		if !got[k] {
			t.Errorf("missing expected key: %q", k)
		}
	}
}

func TestLookupVolumeIndicator(t *testing.T) {
	cases := []struct {
		propName  string
		wantFound bool
	}{
		{"obv", true},
		{"accdist", true},
		{"pvt", true},
		{"iii", true},
		{"wvad", true},
		{"nvi", true},
		{"pvi", true},
		{"wad", true},
		{"sma", false},
		{"tr", false},
		{"", false},
	}
	for _, tc := range cases {
		t.Run(tc.propName, func(t *testing.T) {
			spec, ok := LookupVolumeIndicator(tc.propName)
			if ok != tc.wantFound {
				t.Fatalf("found=%v want=%v", ok, tc.wantFound)
			}
			if !ok {
				return
			}
			if spec.MemberKey != "ta."+tc.propName {
				t.Errorf("MemberKey=%q want=%q", spec.MemberKey, "ta."+tc.propName)
			}
			if spec.SeriesName != "ta_"+tc.propName+"Series" {
				t.Errorf("SeriesName=%q want=%q", spec.SeriesName, "ta_"+tc.propName+"Series")
			}
			if spec.PropName() != tc.propName {
				t.Errorf("PropName()=%q want=%q", spec.PropName(), tc.propName)
			}
		})
	}
}

// ── HasUsage ──────────────────────────────────────────────────────────────────

func TestVolumeIndicatorLifecycle_HasUsage(t *testing.T) {
	cases := []struct {
		name      string
		detected  map[string]bool
		wantUsage bool
	}{
		{"empty", map[string]bool{}, false},
		{"unrelated keys", map[string]bool{"session.ismarket": true, "time_close": true}, false},
		{"ta.obv", map[string]bool{"ta.obv": true}, true},
		{"ta.accdist", map[string]bool{"ta.accdist": true}, true},
		{"ta.pvt", map[string]bool{"ta.pvt": true}, true},
		{"ta.iii", map[string]bool{"ta.iii": true}, true},
		{"ta.wvad", map[string]bool{"ta.wvad": true}, true},
		{"ta.nvi", map[string]bool{"ta.nvi": true}, true},
		{"ta.pvi", map[string]bool{"ta.pvi": true}, true},
		{"ta.wad", map[string]bool{"ta.wad": true}, true},
		{"all eight", map[string]bool{
			"ta.obv": true, "ta.accdist": true, "ta.pvt": true, "ta.iii": true,
			"ta.wvad": true, "ta.nvi": true, "ta.pvi": true, "ta.wad": true,
		}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := NewVolumeIndicatorLifecycle(tc.detected)
			if l.HasUsage() != tc.wantUsage {
				t.Errorf("HasUsage()=%v want=%v", l.HasUsage(), tc.wantUsage)
			}
		})
	}
}

// ── NeedsTimezone ─────────────────────────────────────────────────────────────

func TestVolumeIndicatorLifecycle_NeedsTimezone(t *testing.T) {
	for _, key := range VolumeIndicatorMemberKeys() {
		t.Run(key, func(t *testing.T) {
			l := NewVolumeIndicatorLifecycle(map[string]bool{key: true})
			if l.NeedsTimezone() {
				t.Error("volume indicators must not require timezone")
			}
		})
	}
}

// ── Empty output when no usage ────────────────────────────────────────────────

func TestVolumeIndicatorLifecycle_EmptyWhenNotDetected(t *testing.T) {
	l := NewVolumeIndicatorLifecycle(map[string]bool{})
	methods := map[string]string{
		"declarations":    l.GenerateDeclarations("\t"),
		"initializations": l.GenerateInitializations("\t"),
		"bar population":  l.GenerateBarPopulation("\t", "i"),
		"advancement":     l.GenerateAdvancement("\t", "i"),
		"registrations":   l.GenerateRegistrations("\t"),
		"suppress unused": l.GenerateSuppressUnused("\t"),
	}
	for name, code := range methods {
		if code != "" {
			t.Errorf("%s: expected empty output, got: %q", name, code)
		}
	}
}

// ── Declarations & Initializations ───────────────────────────────────────────

func TestVolumeIndicatorLifecycle_Declarations(t *testing.T) {
	l := NewVolumeIndicatorLifecycle(map[string]bool{"ta.obv": true, "ta.nvi": true})
	code := l.GenerateDeclarations("\t")

	for _, s := range []string{
		"var ta_obvSeries *series.Series",
		"var ta_nviSeries *series.Series",
	} {
		if !strings.Contains(code, s) {
			t.Errorf("missing %q in:\n%s", s, code)
		}
	}
	if strings.Contains(code, "ta_accdistSeries") {
		t.Errorf("accdist should be absent when not detected:\n%s", code)
	}
}

func TestVolumeIndicatorLifecycle_Initializations(t *testing.T) {
	l := NewVolumeIndicatorLifecycle(map[string]bool{"ta.obv": true, "ta.pvi": true})
	code := l.GenerateInitializations("\t")

	for _, s := range []string{
		"ta_obvSeries = series.NewSeries(len(ctx.Data))",
		"ta_pviSeries = series.NewSeries(len(ctx.Data))",
	} {
		if !strings.Contains(code, s) {
			t.Errorf("missing %q in:\n%s", s, code)
		}
	}
}

// ── Bar population: per-indicator formula completeness ────────────────────────

func TestVolumeIndicatorLifecycle_BarPopulation_OBV(t *testing.T) {
	code := NewVolumeIndicatorLifecycle(map[string]bool{"ta.obv": true}).GenerateBarPopulation("\t", "i")
	for _, s := range []string{
		"ta_obvSeries.Set(",
		"ta_obvSeries.Get(1)",
		"bar.Close > _obv_prevClose",
		"bar.Volume",
		"i == 0",
	} {
		if !strings.Contains(code, s) {
			t.Errorf("OBV: missing %q in:\n%s", s, code)
		}
	}
}

func TestVolumeIndicatorLifecycle_BarPopulation_Accdist(t *testing.T) {
	code := NewVolumeIndicatorLifecycle(map[string]bool{"ta.accdist": true}).GenerateBarPopulation("\t", "i")
	for _, s := range []string{
		"_accdist_hl == 0",
		"(bar.Close - bar.Low) - (bar.High - bar.Close)",
		"ta_accdistSeries.Set(",
		"ta_accdistSeries.Get(1)",
	} {
		if !strings.Contains(code, s) {
			t.Errorf("Accdist: missing %q in:\n%s", s, code)
		}
	}
}

func TestVolumeIndicatorLifecycle_BarPopulation_PVT(t *testing.T) {
	code := NewVolumeIndicatorLifecycle(map[string]bool{"ta.pvt": true}).GenerateBarPopulation("\t", "i")
	for _, s := range []string{
		"_pvt_prevClose",
		"_pvt_prevClose == 0",
		"ta_pvtSeries.Set(",
		"ta_pvtSeries.Get(1)",
		"bar.Volume",
	} {
		if !strings.Contains(code, s) {
			t.Errorf("PVT: missing %q in:\n%s", s, code)
		}
	}
}

func TestVolumeIndicatorLifecycle_BarPopulation_III(t *testing.T) {
	code := NewVolumeIndicatorLifecycle(map[string]bool{"ta.iii": true}).GenerateBarPopulation("\t", "i")
	for _, s := range []string{
		"_iii_hl == 0 || bar.Volume == 0",
		"2*bar.Close-bar.High-bar.Low",
		"_iii_hl*bar.Volume",
		"ta_iiiSeries.Set(",
		"ta_iiiSeries.Get(1)",
	} {
		if !strings.Contains(code, s) {
			t.Errorf("III: missing %q in:\n%s", s, code)
		}
	}
}

func TestVolumeIndicatorLifecycle_BarPopulation_WVAD(t *testing.T) {
	code := NewVolumeIndicatorLifecycle(map[string]bool{"ta.wvad": true}).GenerateBarPopulation("\t", "i")
	for _, s := range []string{
		"_wvad_hl == 0",
		"bar.Close-bar.Open",
		"_wvad_hl*bar.Volume",
		"ta_wvadSeries.Set(",
		"ta_wvadSeries.Get(1)",
	} {
		if !strings.Contains(code, s) {
			t.Errorf("WVAD: missing %q in:\n%s", s, code)
		}
	}
}

func TestVolumeIndicatorLifecycle_BarPopulation_NVI(t *testing.T) {
	code := NewVolumeIndicatorLifecycle(map[string]bool{"ta.nvi": true}).GenerateBarPopulation("\t", "i")
	for _, s := range []string{
		"1000.0",
		"bar.Volume < _nvi_prevBar.Volume",
		"ta_nviSeries.Set(",
		"ta_nviSeries.Get(1)",
	} {
		if !strings.Contains(code, s) {
			t.Errorf("NVI: missing %q in:\n%s", s, code)
		}
	}
}

func TestVolumeIndicatorLifecycle_BarPopulation_PVI(t *testing.T) {
	code := NewVolumeIndicatorLifecycle(map[string]bool{"ta.pvi": true}).GenerateBarPopulation("\t", "i")
	for _, s := range []string{
		"1000.0",
		"bar.Volume > _pvi_prevBar.Volume",
		"ta_pviSeries.Set(",
		"ta_pviSeries.Get(1)",
	} {
		if !strings.Contains(code, s) {
			t.Errorf("PVI: missing %q in:\n%s", s, code)
		}
	}
}

func TestVolumeIndicatorLifecycle_BarPopulation_WAD(t *testing.T) {
	code := NewVolumeIndicatorLifecycle(map[string]bool{"ta.wad": true}).GenerateBarPopulation("\t", "i")
	for _, s := range []string{
		"math.Max(bar.High,",
		"math.Min(bar.Low,",
		"ta_wadSeries.Set(",
		"ta_wadSeries.Get(1)",
	} {
		if !strings.Contains(code, s) {
			t.Errorf("WAD: missing %q in:\n%s", s, code)
		}
	}
}

func TestVolumeIndicatorLifecycle_BarPopulation_IterVarSubstitution(t *testing.T) {
	l := NewVolumeIndicatorLifecycle(map[string]bool{"ta.obv": true})
	code := l.GenerateBarPopulation("\t", "idx")
	if !strings.Contains(code, "idx == 0") {
		t.Errorf("custom iterVar not substituted — expected 'idx == 0' in:\n%s", code)
	}
	if strings.Contains(code, "i == 0") {
		t.Errorf("default iterVar 'i' leaked into generated code:\n%s", code)
	}
}

// ── Advancement ───────────────────────────────────────────────────────────────

func TestVolumeIndicatorLifecycle_Advancement(t *testing.T) {
	l := NewVolumeIndicatorLifecycle(map[string]bool{"ta.obv": true, "ta.nvi": true})
	code := l.GenerateAdvancement("\t", "i")
	for _, s := range []string{
		"i < barCount-1",
		"ta_obvSeries.Next()",
		"ta_nviSeries.Next()",
	} {
		if !strings.Contains(code, s) {
			t.Errorf("missing %q in:\n%s", s, code)
		}
	}
}

func TestVolumeIndicatorLifecycle_Advancement_IterVarSubstitution(t *testing.T) {
	code := NewVolumeIndicatorLifecycle(map[string]bool{"ta.obv": true}).GenerateAdvancement("\t", "j")
	if !strings.Contains(code, "j < barCount-1") {
		t.Errorf("custom iterVar not used in advancement:\n%s", code)
	}
	if strings.Contains(code, "i < barCount-1") {
		t.Errorf("default iterVar 'i' leaked in advancement:\n%s", code)
	}
}

// ── Registrations & SuppressUnused ────────────────────────────────────────────

func TestVolumeIndicatorLifecycle_Registrations(t *testing.T) {
	l := NewVolumeIndicatorLifecycle(map[string]bool{"ta.obv": true, "ta.accdist": true})
	code := l.GenerateRegistrations("\t")
	for _, s := range []string{
		`ctx.RegisterSeries("ta_obvSeries", ta_obvSeries)`,
		`ctx.RegisterSeries("ta_accdistSeries", ta_accdistSeries)`,
	} {
		if !strings.Contains(code, s) {
			t.Errorf("missing %q in:\n%s", s, code)
		}
	}
}

func TestVolumeIndicatorLifecycle_SuppressUnused(t *testing.T) {
	l := NewVolumeIndicatorLifecycle(map[string]bool{"ta.obv": true, "ta.wad": true})
	code := l.GenerateSuppressUnused("\t")
	for _, s := range []string{"_ = ta_obvSeries", "_ = ta_wadSeries"} {
		if !strings.Contains(code, s) {
			t.Errorf("missing %q in:\n%s", s, code)
		}
	}
}

// ── Isolation ─────────────────────────────────────────────────────────────────

func TestVolumeIndicatorLifecycle_Isolation(t *testing.T) {
	cases := []struct {
		detected string
		absent   string
	}{
		{"ta.obv", "ta_accdistSeries"},
		{"ta.nvi", "ta_pviSeries"},
		{"ta.wad", "ta_wvadSeries"},
		{"ta.pvt", "ta_iiiSeries"},
	}
	for _, tc := range cases {
		t.Run(tc.detected, func(t *testing.T) {
			l := NewVolumeIndicatorLifecycle(map[string]bool{tc.detected: true})
			code := l.GenerateDeclarations("\t")
			if strings.Contains(code, tc.absent) {
				t.Errorf("detecting %q should not emit %q:\n%s", tc.detected, tc.absent, code)
			}
		})
	}
}

// ── All eight indicators active ───────────────────────────────────────────────

func TestVolumeIndicatorLifecycle_AllEightActive(t *testing.T) {
	detected := map[string]bool{
		"ta.obv": true, "ta.accdist": true, "ta.pvt": true, "ta.iii": true,
		"ta.wvad": true, "ta.nvi": true, "ta.pvi": true, "ta.wad": true,
	}
	l := NewVolumeIndicatorLifecycle(detected)
	decl := l.GenerateDeclarations("\t")

	for _, name := range []string{
		"ta_obvSeries", "ta_accdistSeries", "ta_pvtSeries", "ta_iiiSeries",
		"ta_wvadSeries", "ta_nviSeries", "ta_pviSeries", "ta_wadSeries",
	} {
		if !strings.Contains(decl, name) {
			t.Errorf("all-eight: missing %q in declarations:\n%s", name, decl)
		}
	}

	if !l.HasUsage() {
		t.Error("HasUsage() should be true when all eight are detected")
	}
}

// ── Symbol table registration ─────────────────────────────────────────────────
//
// Volume indicators are dispatched via evaluateMemberExpressionAtBar, not via
// the identifier var-lookup switch. Registering "ta.obv" in the symbol table
// would cause SecurityExpressionHandler to emit `varSeries = ta.obvSeries`
// which is invalid Go. GenerateSymbolTableRegistrations must therefore be a no-op.

func TestVolumeIndicatorLifecycle_SymbolTableRegistration_NeverRegisters(t *testing.T) {
	detected := map[string]bool{
		"ta.obv": true, "ta.accdist": true, "ta.pvt": true, "ta.iii": true,
		"ta.wvad": true, "ta.nvi": true, "ta.pvi": true, "ta.wad": true,
	}
	l := NewVolumeIndicatorLifecycle(detected)
	st := NewSymbolTable()
	l.GenerateSymbolTableRegistrations(st)

	for _, key := range VolumeIndicatorMemberKeys() {
		if st.IsSeries(key) {
			t.Errorf("%q must not be registered in symbol table — would generate invalid Go", key)
		}
	}
}

func TestVolumeIndicatorLifecycle_SymbolTableRegistration_NilTable(t *testing.T) {
	l := NewVolumeIndicatorLifecycle(map[string]bool{"ta.obv": true})
	l.GenerateSymbolTableRegistrations(nil) // must not panic
}

func TestVolumeIndicatorLifecycle_SymbolTableRegistration_NoUsage(t *testing.T) {
	l := NewVolumeIndicatorLifecycle(map[string]bool{})
	st := NewSymbolTable()
	l.GenerateSymbolTableRegistrations(st)
	for _, key := range VolumeIndicatorMemberKeys() {
		if st.IsSeries(key) {
			t.Errorf("%s should not be registered when not detected", key)
		}
	}
}
