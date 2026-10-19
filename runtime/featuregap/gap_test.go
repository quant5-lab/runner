package featuregap

import (
	"math"
	"strings"
	"sync"
	"testing"
)

// recordAt and recordStringAt keep the tests readable while exercising the
// location-carrying recorder entries with an anonymous position.
func recordAt(name, source string, barIdx int) float64 {
	return Record(name, source, "", 0, 0, barIdx)
}

func recordStringAt(name, source string, barIdx int) string {
	return RecordString(name, source, "", 0, 0, barIdx)
}

func TestRecord_ReturnsNaN(t *testing.T) {
	Reset()
	got := recordAt("foo", "test", 0)
	if !math.IsNaN(got) {
		t.Fatalf("Record returned %v, want NaN", got)
	}
}

func TestRecord_CarriesSourceLocation(t *testing.T) {
	Reset()
	Record("gap_fn", "plot_expression", "strategy.pine", 7, 3, 2)

	h := Snapshot()[0]
	if h.File != "strategy.pine" || h.Line != 7 || h.Column != 3 {
		t.Fatalf("location = %s:%d:%d, want strategy.pine:7:3", h.File, h.Line, h.Column)
	}
}

func TestRecordString_CarriesSourceLocation(t *testing.T) {
	Reset()
	RecordString("gap_fn", "generator.comment_unknown", "strategy.pine", 11, 9, 4)

	h := Snapshot()[0]
	if h.File != "strategy.pine" || h.Line != 11 || h.Column != 9 {
		t.Fatalf("location = %s:%d:%d, want strategy.pine:11:9", h.File, h.Line, h.Column)
	}
}

func TestRecordString_ReturnsEmptyStringAndRecordsSubstitution(t *testing.T) {
	Reset()
	got := recordStringAt("foo", "string_context", 3)
	if got != "" {
		t.Fatalf("RecordString returned %q, want empty string", got)
	}
	snap := Snapshot()
	if len(snap) != 1 {
		t.Fatalf("len(Snapshot()) = %d, want 1", len(snap))
	}
	h := snap[0]
	if h.Substitution != "empty-string" {
		t.Fatalf("Substitution = %q, want empty-string", h.Substitution)
	}
	if h.FirstSeenBar != 3 || h.LastSeenBar != 3 || h.HitCount != 1 {
		t.Fatalf("runtime counters = %d..%d hits=%d, want 3..3 hits=1", h.FirstSeenBar, h.LastSeenBar, h.HitCount)
	}
}

func TestRecordString_AccumulatesHits(t *testing.T) {
	Reset()
	recordStringAt("foo", "string_context", 2)
	recordStringAt("foo", "string_context", 5)
	snap := Snapshot()
	if len(snap) != 1 {
		t.Fatalf("len(Snapshot()) = %d, want 1", len(snap))
	}
	h := snap[0]
	if h.FirstSeenBar != 2 || h.LastSeenBar != 5 || h.HitCount != 2 {
		t.Fatalf("runtime counters = %d..%d hits=%d, want 2..5 hits=2", h.FirstSeenBar, h.LastSeenBar, h.HitCount)
	}
}

func TestRecord_AccumulatesHits(t *testing.T) {
	Reset()
	for i := 0; i < 5; i++ {
		recordAt("foo", "test", i)
	}
	snap := Snapshot()
	if len(snap) != 1 {
		t.Fatalf("len(Snapshot()) = %d, want 1", len(snap))
	}
	h := snap[0]
	if h.Name != "foo" || h.Source != "test" {
		t.Errorf("hit identity = %+v, want name=foo source=test", h)
	}
	if h.HitCount != 5 {
		t.Errorf("HitCount = %d, want 5", h.HitCount)
	}
	if h.FirstSeenBar != 0 || h.LastSeenBar != 4 {
		t.Errorf("bar range = %d..%d, want 0..4", h.FirstSeenBar, h.LastSeenBar)
	}
}

func TestRecord_DistinguishesByNameAndSource(t *testing.T) {
	Reset()
	recordAt("foo", "src_a", 1)
	recordAt("foo", "src_b", 2)
	recordAt("bar", "src_a", 3)
	snap := Snapshot()
	if len(snap) != 3 {
		t.Fatalf("len(Snapshot()) = %d, want 3 distinct (name,source) pairs", len(snap))
	}
}

func TestSnapshot_StableSorted(t *testing.T) {
	Reset()
	recordAt("zeta", "x", 0)
	recordAt("alpha", "y", 0)
	recordAt("alpha", "x", 0)
	snap := Snapshot()
	want := []struct{ name, source string }{{"alpha", "x"}, {"alpha", "y"}, {"zeta", "x"}}
	for i, w := range want {
		if snap[i].Name != w.name || snap[i].Source != w.source {
			t.Errorf("snap[%d] = (%s,%s), want (%s,%s)", i, snap[i].Name, snap[i].Source, w.name, w.source)
		}
	}
}

func TestReset_ClearsAll(t *testing.T) {
	Reset()
	recordAt("foo", "test", 0)
	Reset()
	if got := Snapshot(); len(got) != 0 {
		t.Errorf("Snapshot after Reset = %v, want empty", got)
	}
}

func TestRecord_ConcurrencySafe(t *testing.T) {
	Reset()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			recordAt("foo", "test", idx)
		}(i)
	}
	wg.Wait()
	snap := Snapshot()
	if len(snap) != 1 {
		t.Fatalf("len(Snapshot()) = %d, want 1", len(snap))
	}
	if snap[0].HitCount != 50 {
		t.Errorf("HitCount = %d, want 50", snap[0].HitCount)
	}
}

func TestSummary_EmptyWhenNoHits(t *testing.T) {
	Reset()
	if s := Summary(); s != "" {
		t.Errorf("Summary() = %q, want empty string", s)
	}
}

func TestSummary_NonEmptyAfterHit(t *testing.T) {
	Reset()
	recordAt("foo", "src", 7)
	if s := Summary(); s == "" {
		t.Error("Summary() empty after Record; want non-empty")
	}
}

func TestSnapshot_IndependentOfSubsequentState(t *testing.T) {
	Reset()
	recordAt("foo", "test", 0)
	snap := Snapshot()
	Reset()
	if len(snap) != 1 {
		t.Errorf("Snapshot captured before Reset should retain 1 entry; got %d", len(snap))
	}
}

func TestRecord_FirstSeenBarPreserved(t *testing.T) {
	Reset()
	recordAt("foo", "test", 10)
	recordAt("foo", "test", 20)
	recordAt("foo", "test", 5)
	h := Snapshot()[0]
	if h.FirstSeenBar != 10 {
		t.Errorf("FirstSeenBar = %d, want 10 (first call wins)", h.FirstSeenBar)
	}
}

func TestRecord_LastSeenBarMonotone(t *testing.T) {
	Reset()
	recordAt("foo", "test", 10)
	recordAt("foo", "test", 3)
	h := Snapshot()[0]
	if h.LastSeenBar != 10 {
		t.Errorf("LastSeenBar = %d, want 10 (never decreases below observed max)", h.LastSeenBar)
	}
}

func TestSummary_ContainsAllHitFields(t *testing.T) {
	Reset()
	recordAt("request.dividends", "plot_expression", 42)
	recordAt("request.dividends", "plot_expression", 99)
	s := Summary()
	for _, want := range []string{"request.dividends", "plot_expression", "42", "99", "2"} {
		if !strings.Contains(s, want) {
			t.Errorf("Summary() missing %q\nfull summary: %s", want, s)
		}
	}
}

func TestRecordStatic_MergesWithRuntimeObservation(t *testing.T) {
	Reset()
	RecordStaticAt("line.get_price", "chart_namespace_getter", "lines.pine", 4, 5, "codegen", "backtest-critical", "NaN", []string{"strategy.entry"})
	recordAt("line.get_price", "chart_namespace_getter", 3)
	recordAt("line.get_price", "chart_namespace_getter", 5)

	snap := Snapshot()
	if len(snap) != 1 {
		t.Fatalf("len(Snapshot()) = %d, want 1", len(snap))
	}
	h := snap[0]
	if h.Impact != "backtest-critical" {
		t.Fatalf("Impact = %q, want backtest-critical", h.Impact)
	}
	if h.FirstSeenBar != 3 || h.LastSeenBar != 5 {
		t.Fatalf("bar range = %d..%d, want 3..5", h.FirstSeenBar, h.LastSeenBar)
	}
	if h.HitCount != 2 {
		t.Fatalf("HitCount = %d, want 2 runtime observations", h.HitCount)
	}
	if h.File != "lines.pine" || h.Line != 4 || h.Column != 5 {
		t.Fatalf("location = %s:%d:%d, want lines.pine:4:5", h.File, h.Line, h.Column)
	}
}

func TestRecordStaticAt_FillsLocationAfterRuntimeObservation(t *testing.T) {
	Reset()
	recordAt("line.get_price", "chart_namespace_getter", 3)
	RecordStaticAt("line.get_price", "chart_namespace_getter", "lines.pine", 4, 5, "codegen", "backtest-critical", "NaN", []string{"strategy.entry"})

	h := Snapshot()[0]
	if h.File != "lines.pine" || h.Line != 4 || h.Column != 5 {
		t.Fatalf("location = %s:%d:%d, want lines.pine:4:5", h.File, h.Line, h.Column)
	}
	if h.FirstSeenBar != 3 || h.LastSeenBar != 3 || h.HitCount != 1 {
		t.Fatalf("runtime counters = %d..%d hits=%d, want 3..3 hits=1", h.FirstSeenBar, h.LastSeenBar, h.HitCount)
	}
}

func TestRecordDiagnostic_EscalatesImpactAndDeduplicatesSinks(t *testing.T) {
	Reset()
	RecordDiagnostic("gap", "src", "runtime", "observable-non-backtest", "NaN", []string{"strategy.entry", "strategy.entry"}, 1)
	RecordDiagnostic("gap", "src", "runtime", "backtest-critical", "NaN", []string{"strategy.close", "strategy.entry"}, 2)

	h := Snapshot()[0]
	if h.Impact != "backtest-critical" {
		t.Fatalf("Impact = %q, want backtest-critical", h.Impact)
	}
	want := []string{"strategy.close", "strategy.entry"}
	if len(h.Sinks) != len(want) {
		t.Fatalf("Sinks = %v, want %v", h.Sinks, want)
	}
	for i := range want {
		if h.Sinks[i] != want[i] {
			t.Fatalf("Sinks = %v, want %v", h.Sinks, want)
		}
	}
}

func TestBacktestSinks_OnlyBacktestCriticalSortedUnique(t *testing.T) {
	snap := []Hit{
		{Impact: "observable-non-backtest", Sinks: []string{"strategy.entry"}},
		{Impact: "backtest-critical", Sinks: []string{"strategy.close", "strategy.entry"}},
		{Impact: "backtest-critical", Sinks: []string{"strategy.entry"}},
	}

	got := BacktestSinks(snap)
	want := []string{"strategy.close", "strategy.entry"}
	if len(got) != len(want) {
		t.Fatalf("BacktestSinks() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("BacktestSinks() = %v, want %v", got, want)
		}
	}
}
