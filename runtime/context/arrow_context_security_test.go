package context_test

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/request"
)

/* TestSecurityBarMapper_SatisfiesBarIndexMapper verifies interface contract */
func TestSecurityBarMapper_SatisfiesBarIndexMapper(t *testing.T) {
	var mapper context.BarIndexMapper = request.NewSecurityBarMapper()
	if mapper == nil {
		t.Fatal("SecurityBarMapper should satisfy BarIndexMapper interface")
	}
}

/* TestArrowContext_SecurityBridge_SetAndRetrieve verifies set/get roundtrip for security bridge fields */
func TestArrowContext_SecurityBridge_SetAndRetrieve(t *testing.T) {
	ctx := context.New("TEST", "1D", 10)
	ac := context.NewArrowContext(ctx)

	secCtx := context.New("SEC", "1h", 5)
	mapper := request.NewSecurityBarMapper()

	ac.SetSecurityContext("k1", secCtx)
	ac.SetBarMapper("k1", mapper)

	if ac.SecurityContexts["k1"] != secCtx {
		t.Error("SetSecurityContext: stored context mismatch")
	}
	if ac.SecurityBarMappers["k1"] == nil {
		t.Error("SetBarMapper: stored mapper should not be nil")
	}
}

/* TestArrowContext_SecurityBridge_MultipleKeys verifies independent key storage */
func TestArrowContext_SecurityBridge_MultipleKeys(t *testing.T) {
	ctx := context.New("TEST", "1D", 10)
	ac := context.NewArrowContext(ctx)

	sec1 := context.New("SEC1", "1D", 5)
	sec2 := context.New("SEC2", "1h", 3)

	ac.SetSecurityContext("daily", sec1)
	ac.SetSecurityContext("hourly", sec2)
	ac.SetBarMapper("daily", request.NewSecurityBarMapper())
	ac.SetBarMapper("hourly", request.NewSecurityBarMapper())

	if ac.SecurityContexts["daily"] != sec1 {
		t.Error("daily context mismatch")
	}
	if ac.SecurityContexts["hourly"] != sec2 {
		t.Error("hourly context mismatch")
	}
	if len(ac.SecurityContexts) != 2 {
		t.Errorf("expected 2 security contexts, got %d", len(ac.SecurityContexts))
	}
	if len(ac.SecurityBarMappers) != 2 {
		t.Errorf("expected 2 bar mappers, got %d", len(ac.SecurityBarMappers))
	}
}

/* TestArrowContext_SecurityBridge_OverwriteKey verifies last-write-wins semantics */
func TestArrowContext_SecurityBridge_OverwriteKey(t *testing.T) {
	ctx := context.New("TEST", "1D", 10)
	ac := context.NewArrowContext(ctx)

	first := context.New("FIRST", "1D", 5)
	second := context.New("SECOND", "1D", 10)

	ac.SetSecurityContext("k", first)
	ac.SetSecurityContext("k", second)

	if ac.SecurityContexts["k"] != second {
		t.Error("expected overwrite to use latest context")
	}
}

/* TestArrowContext_SecurityBridge_NilByDefault verifies zero-value safety */
func TestArrowContext_SecurityBridge_NilByDefault(t *testing.T) {
	ctx := context.New("TEST", "1D", 10)
	ac := context.NewArrowContext(ctx)

	if ac.SecurityContexts != nil {
		t.Error("SecurityContexts should be nil by default")
	}
	if ac.SecurityBarMappers != nil {
		t.Error("SecurityBarMappers should be nil by default")
	}
}

/* TestArrowContext_SecurityBridge_LazyInit verifies maps initialize on first Set call */
func TestArrowContext_SecurityBridge_LazyInit(t *testing.T) {
	ctx := context.New("TEST", "1D", 10)
	ac := context.NewArrowContext(ctx)

	if ac.SecurityContexts != nil || ac.SecurityBarMappers != nil {
		t.Fatal("maps should be nil before first Set")
	}

	ac.SetSecurityContext("k", context.New("SEC", "1D", 5))
	if ac.SecurityContexts == nil {
		t.Error("SecurityContexts should be non-nil after SetSecurityContext")
	}
	if ac.SecurityBarMappers != nil {
		t.Error("SecurityBarMappers should still be nil (not yet set)")
	}

	ac.SetBarMapper("k", request.NewSecurityBarMapper())
	if ac.SecurityBarMappers == nil {
		t.Error("SecurityBarMappers should be non-nil after SetBarMapper")
	}
}

/* TestArrowContext_SecurityBridge_DoesNotAffectLocalSeries verifies orthogonality */
func TestArrowContext_SecurityBridge_DoesNotAffectLocalSeries(t *testing.T) {
	ctx := context.New("TEST", "1D", 100)
	for i := 0; i < 10; i++ {
		ctx.AddBar(context.OHLCV{Open: 1, High: 2, Low: 0.5, Close: 1.5, Volume: 100})
	}
	ac := context.NewArrowContext(ctx)

	s := ac.GetOrCreateSeries("myVar")
	ac.SetSecurityContext("k", context.New("SEC", "1D", 50))
	ac.SetBarMapper("k", request.NewSecurityBarMapper())

	s2 := ac.GetOrCreateSeries("myVar")
	if s != s2 {
		t.Error("security bridge operations should not affect local series")
	}
}
