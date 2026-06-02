package featuregap

import (
	"math"
	"strings"
	"sync"
	"testing"
)

func TestRecord_ReturnsNaN(t *testing.T) {
	Reset()
	got := Record("foo", "test", 0)
	if !math.IsNaN(got) {
		t.Fatalf("Record returned %v, want NaN", got)
	}
}

func TestRecord_AccumulatesHits(t *testing.T) {
	Reset()
	for i := 0; i < 5; i++ {
		Record("foo", "test", i)
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
	Record("foo", "src_a", 1)
	Record("foo", "src_b", 2)
	Record("bar", "src_a", 3)
	snap := Snapshot()
	if len(snap) != 3 {
		t.Fatalf("len(Snapshot()) = %d, want 3 distinct (name,source) pairs", len(snap))
	}
}

func TestSnapshot_StableSorted(t *testing.T) {
	Reset()
	Record("zeta", "x", 0)
	Record("alpha", "y", 0)
	Record("alpha", "x", 0)
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
	Record("foo", "test", 0)
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
			Record("foo", "test", idx)
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
	Record("foo", "src", 7)
	if s := Summary(); s == "" {
		t.Error("Summary() empty after Record; want non-empty")
	}
}

func TestSnapshot_IndependentOfSubsequentState(t *testing.T) {
	Reset()
	Record("foo", "test", 0)
	snap := Snapshot()
	Reset()
	if len(snap) != 1 {
		t.Errorf("Snapshot captured before Reset should retain 1 entry; got %d", len(snap))
	}
}

func TestRecord_FirstSeenBarPreserved(t *testing.T) {
	Reset()
	Record("foo", "test", 10)
	Record("foo", "test", 20)
	Record("foo", "test", 5)
	h := Snapshot()[0]
	if h.FirstSeenBar != 10 {
		t.Errorf("FirstSeenBar = %d, want 10 (first call wins)", h.FirstSeenBar)
	}
}

func TestRecord_LastSeenBarMonotone(t *testing.T) {
	Reset()
	Record("foo", "test", 10)
	Record("foo", "test", 3)
	h := Snapshot()[0]
	if h.LastSeenBar != 10 {
		t.Errorf("LastSeenBar = %d, want 10 (never decreases below observed max)", h.LastSeenBar)
	}
}

func TestSummary_ContainsAllHitFields(t *testing.T) {
	Reset()
	Record("request.dividends", "plot_expression", 42)
	Record("request.dividends", "plot_expression", 99)
	s := Summary()
	for _, want := range []string{"request.dividends", "plot_expression", "42", "99", "2"} {
		if !strings.Contains(s, want) {
			t.Errorf("Summary() missing %q\nfull summary: %s", want, s)
		}
	}
}
