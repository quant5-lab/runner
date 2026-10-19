package context

import (
	"math"
	"testing"
)

func makeTestCtx(bars int) *Context {
	ctx := New("TEST", "1h", bars)
	for i := 0; i < bars; i++ {
		ctx.AddBar(OHLCV{Close: float64(i + 1)})
	}
	return ctx
}

func TestGetOrCreateChildContext_CreatesOnce(t *testing.T) {
	parent := NewArrowContext(makeTestCtx(10))
	c1 := parent.GetOrCreateChildContext("calc_1")
	c2 := parent.GetOrCreateChildContext("calc_1")
	if c1 != c2 {
		t.Error("same key must return same child context")
	}
}

func TestGetOrCreateChildContext_DistinctKeys_DistinctContexts(t *testing.T) {
	parent := NewArrowContext(makeTestCtx(10))
	a := parent.GetOrCreateChildContext("fn_1")
	b := parent.GetOrCreateChildContext("fn_2")
	if a == b {
		t.Error("distinct keys must return distinct child contexts")
	}
}

func TestChildContext_SeriesPersistAcrossAdvanceAll(t *testing.T) {
	parent := NewArrowContext(makeTestCtx(10))
	child := parent.GetOrCreateChildContext("inner_1")
	s := child.GetOrCreateSeries("x")
	s.Set(42.0)
	parent.AdvanceAll()
	if got := s.Get(1); got != 42.0 {
		t.Errorf("expected Get(1)=42 after AdvanceAll, got %v", got)
	}
}

func TestAdvanceAll_RecursesIntoGrandchildren(t *testing.T) {
	parent := NewArrowContext(makeTestCtx(10))
	child := parent.GetOrCreateChildContext("rma_1")
	grandchild := child.GetOrCreateChildContext("src_1")
	ps := parent.GetOrCreateSeries("p")
	cs := child.GetOrCreateSeries("c")
	gs := grandchild.GetOrCreateSeries("g")
	ps.Set(1.0)
	cs.Set(2.0)
	gs.Set(3.0)
	parent.AdvanceAll()
	if got := ps.Get(1); got != 1.0 {
		t.Errorf("parent series: expected 1.0, got %v", got)
	}
	if got := cs.Get(1); got != 2.0 {
		t.Errorf("child series: expected 2.0, got %v", got)
	}
	if got := gs.Get(1); got != 3.0 {
		t.Errorf("grandchild series: expected 3.0, got %v", got)
	}
}

func TestChildContext_BeforeSeriesExists_PreHistoryIsNaN(t *testing.T) {
	parent := NewArrowContext(makeTestCtx(10))
	child := parent.GetOrCreateChildContext("fn_1")
	s := child.GetOrCreateSeries("v")
	// Before cursor is advanced at all, offset > 0 reads pre-history (NaN by default).
	if got := s.Get(1); !math.IsNaN(got) {
		t.Errorf("pre-history offset should return NaN, got %v", got)
	}
}

func TestReset_RecursesIntoChildren(t *testing.T) {
	parent := NewArrowContext(makeTestCtx(10))
	child := parent.GetOrCreateChildContext("fn_1")
	s := child.GetOrCreateSeries("v")
	s.Set(99.0)
	parent.AdvanceAll()
	parent.Reset(0)
	if got := s.Get(1); !math.IsNaN(got) {
		t.Errorf("after Reset child series should read NaN, got %v", got)
	}
}
