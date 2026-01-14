package context

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/series"
)

func TestContext_ResolveVariable_WithoutParent(t *testing.T) {
	ctx := New("AAPL", "1h", 100)
	ctx.SetParent(nil, NewIdentityAligner())

	localSeries := series.NewSeries(100)
	localSeries.Set(123.45)
	ctx.RegisterSeries("testVar", localSeries)

	result := ctx.ResolveVariable("testVar")

	if !result.Found {
		t.Fatal("expected variable to be found")
	}
	if result.Series != localSeries {
		t.Error("expected same series instance")
	}
}

func TestContext_ResolveVariable_FromParent(t *testing.T) {
	parentCtx := New("AAPL", "1h", 1000)
	parentCtx.SetParent(nil, NewIdentityAligner())

	parentSeries := series.NewSeries(1000)
	parentSeries.Set(100.0)
	parentCtx.RegisterSeries("parentVar", parentSeries)

	childCtx := New("AAPL", "1D", 100)
	childCtx.SetParent(parentCtx, NewIdentityAligner())

	result := childCtx.ResolveVariable("parentVar")

	if !result.Found {
		t.Fatal("expected parent variable to be found")
	}
	if result.Series != parentSeries {
		t.Error("expected parent series instance")
	}
}

func TestContext_ResolveVariable_ThreeLevelHierarchy(t *testing.T) {
	mainCtx := New("AAPL", "1h", 1000)
	mainCtx.SetParent(nil, NewIdentityAligner())
	mainSeries := series.NewSeries(1000)
	mainSeries.Set(50.0)
	mainCtx.RegisterSeries("mainVar", mainSeries)

	dailyCtx := New("AAPL", "1D", 100)
	dailyCtx.SetParent(mainCtx, NewIdentityAligner())
	dailySeries := series.NewSeries(100)
	dailySeries.Set(200.0)
	dailyCtx.RegisterSeries("dailyVar", dailySeries)

	weeklyCtx := New("AAPL", "1W", 20)
	weeklyCtx.SetParent(dailyCtx, NewIdentityAligner())

	mainResult := weeklyCtx.ResolveVariable("mainVar")
	if !mainResult.Found {
		t.Fatal("expected main variable to be found from weekly context")
	}
	if mainResult.Series != mainSeries {
		t.Error("expected main series instance")
	}

	dailyResult := weeklyCtx.ResolveVariable("dailyVar")
	if !dailyResult.Found {
		t.Fatal("expected daily variable to be found from weekly context")
	}
	if dailyResult.Series != dailySeries {
		t.Error("expected daily series instance")
	}
}

func TestContext_ResolveVariable_LocalShadowsParent(t *testing.T) {
	parentCtx := New("AAPL", "1h", 1000)
	parentCtx.SetParent(nil, NewIdentityAligner())
	parentSeries := series.NewSeries(1000)
	parentSeries.Set(100.0)
	parentCtx.RegisterSeries("sharedVar", parentSeries)

	childCtx := New("AAPL", "1D", 100)
	childCtx.SetParent(parentCtx, NewIdentityAligner())
	childSeries := series.NewSeries(100)
	childSeries.Set(200.0)
	childCtx.RegisterSeries("sharedVar", childSeries)

	result := childCtx.ResolveVariable("sharedVar")

	if !result.Found {
		t.Fatal("expected variable to be found")
	}
	if result.Series != childSeries {
		t.Error("expected child series to shadow parent")
	}
}

func TestContext_GetParent(t *testing.T) {
	parentCtx := New("AAPL", "1h", 1000)
	childCtx := New("AAPL", "1D", 100)
	childCtx.SetParent(parentCtx, NewIdentityAligner())

	if childCtx.GetParent() != parentCtx {
		t.Error("expected parent context to be returned")
	}

	if parentCtx.GetParent() != nil {
		t.Error("expected root context to have nil parent")
	}
}
