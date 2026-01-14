package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/series"
)

func TestBarEvaluator_ContextHierarchy_ParentVariableResolution(t *testing.T) {
	mainCtx := context.New("AAPL", "1h", 1000)
	for i := 0; i < 10; i++ {
		mainCtx.AddBar(context.OHLCV{Time: int64(i * 3600), Close: float64(i)})
	}
	mainCtx.SetParent(nil, context.NewIdentityAligner())

	mainVarSeries := series.NewSeries(1000)
	for i := 0; i < 10; i++ {
		mainVarSeries.Set(float64(i * 10))
		if i < 9 {
			mainVarSeries.Next()
		}
	}
	mainCtx.RegisterSeries("mainVar", mainVarSeries)

	dailyCtx := context.New("AAPL", "1D", 100)
	for i := 0; i < 3; i++ {
		dailyCtx.AddBar(context.OHLCV{Time: int64(i * 86400), Close: float64(i)})
	}
	aligner := context.NewMappedAligner()
	aligner.SetMapping(0, 0)
	aligner.SetMapping(1, 5)
	aligner.SetMapping(2, 10)
	dailyCtx.SetParent(mainCtx, aligner)

	evaluator := NewStreamingBarEvaluator()

	expr := &ast.Identifier{Name: "mainVar"}

	dailyCtx.BarIndex = 0
	mainCtx.BarIndex = 0
	value, err := evaluator.EvaluateAtBar(expr, dailyCtx, 0)
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}
	if value != 0.0 {
		t.Errorf("expected 0.0 (bar 0), got %.2f", value)
	}

	dailyCtx.BarIndex = 1
	mainCtx.BarIndex = 5
	value, err = evaluator.EvaluateAtBar(expr, dailyCtx, 1)
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}
	if value != 50.0 {
		t.Errorf("expected 50.0 (bar 5 * 10), got %.2f", value)
	}
}

func TestBarEvaluator_ContextHierarchy_ThreeLevels(t *testing.T) {
	mainCtx := context.New("AAPL", "1h", 1000)
	for i := 0; i < 100; i++ {
		mainCtx.AddBar(context.OHLCV{Time: int64(i * 3600), Close: float64(i)})
	}
	mainCtx.SetParent(nil, context.NewIdentityAligner())
	mainSeries := series.NewSeries(1000)
	for i := 0; i < 20; i++ {
		mainSeries.Set(100.0)
		if i < 19 {
			mainSeries.Next()
		}
	}
	mainCtx.RegisterSeries("hourlyVar", mainSeries)

	dailyCtx := context.New("AAPL", "1D", 100)
	for i := 0; i < 20; i++ {
		dailyCtx.AddBar(context.OHLCV{Time: int64(i * 86400), Close: float64(i)})
	}
	dailyAligner := context.NewMappedAligner()
	dailyAligner.SetMapping(5, 10)
	dailyCtx.SetParent(mainCtx, dailyAligner)
	dailySeries := series.NewSeries(100)
	for i := 0; i < 10; i++ {
		dailySeries.Set(200.0)
		if i < 9 {
			dailySeries.Next()
		}
	}
	dailyCtx.RegisterSeries("dailyVar", dailySeries)

	weeklyCtx := context.New("AAPL", "1W", 20)
	for i := 0; i < 10; i++ {
		weeklyCtx.AddBar(context.OHLCV{Time: int64(i * 604800), Close: float64(i)})
	}
	weeklyAligner := context.NewMappedAligner()
	weeklyAligner.SetMapping(0, 5)
	weeklyCtx.SetParent(dailyCtx, weeklyAligner)

	evaluator := NewStreamingBarEvaluator()

	mainCtx.BarIndex = 10
	dailyCtx.BarIndex = 5
	weeklyCtx.BarIndex = 0

	hourlyExpr := &ast.Identifier{Name: "hourlyVar"}
	value, err := evaluator.EvaluateAtBar(hourlyExpr, weeklyCtx, 0)
	if err != nil {
		t.Fatalf("hourly var evaluation failed: %v", err)
	}
	if value != 100.0 {
		t.Errorf("expected hourly value 100.0, got %.2f", value)
	}

	dailyExpr := &ast.Identifier{Name: "dailyVar"}
	value, err = evaluator.EvaluateAtBar(dailyExpr, weeklyCtx, 0)
	if err != nil {
		t.Fatalf("daily var evaluation failed: %v", err)
	}
	if value != 200.0 {
		t.Errorf("expected daily value 200.0, got %.2f", value)
	}
}

func TestBarEvaluator_ContextHierarchy_WarmupPeriod(t *testing.T) {
	mainCtx := context.New("AAPL", "1h", 1000)
	for i := 0; i < 10; i++ {
		mainCtx.AddBar(context.OHLCV{Time: int64(i * 3600), Close: float64(i)})
	}
	mainCtx.SetParent(nil, context.NewIdentityAligner())
	mainSeries := series.NewSeries(1000)
	mainSeries.Set(100.0)
	mainSeries.Next()
	mainCtx.RegisterSeries("mainVar", mainSeries)

	dailyCtx := context.New("AAPL", "1D", 100)
	for i := 0; i < 15; i++ {
		dailyCtx.AddBar(context.OHLCV{Time: int64(i * 86400), Close: float64(i)})
	}
	aligner := context.NewMappedAligner()
	aligner.SetMapping(10, 0)
	dailyCtx.SetParent(mainCtx, aligner)

	evaluator := NewStreamingBarEvaluator()

	mainCtx.BarIndex = 0
	dailyCtx.BarIndex = 5

	expr := &ast.Identifier{Name: "mainVar"}
	value, err := evaluator.EvaluateAtBar(expr, dailyCtx, 5)
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}
	if !math.IsNaN(value) {
		t.Errorf("expected NaN during warmup (unmapped bar), got %.2f", value)
	}
}
