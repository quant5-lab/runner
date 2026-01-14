package context

import (
	"fmt"
	"testing"
)

func TestArrowContext_LazyInitialization(t *testing.T) {
	ctx := New("TEST", "1h", 100)
	for i := 0; i < 100; i++ {
		ctx.AddBar(OHLCV{Time: int64(i), Close: float64(i)})
	}
	for i := 0; i < 10; i++ {
		ctx.AddBar(OHLCV{Time: int64(i), Close: float64(i)})
	}

	arrowCtx := NewArrowContext(ctx)

	if len(arrowCtx.LocalSeries) != 0 {
		t.Errorf("Expected empty LocalSeries on creation, got %d entries", len(arrowCtx.LocalSeries))
	}

	s1 := arrowCtx.GetOrCreateSeries("up")
	if s1 == nil {
		t.Fatal("GetOrCreateSeries returned nil")
	}

	if len(arrowCtx.LocalSeries) != 1 {
		t.Errorf("Expected 1 Series after first access, got %d", len(arrowCtx.LocalSeries))
	}

	s2 := arrowCtx.GetOrCreateSeries("up")
	if s1 != s2 {
		t.Error("Expected same Series instance on repeated access")
	}

	if len(arrowCtx.LocalSeries) != 1 {
		t.Errorf("Expected still 1 Series after repeated access, got %d", len(arrowCtx.LocalSeries))
	}
}

func TestArrowContext_MultipleSeries(t *testing.T) {
	ctx := New("TEST", "1h", 50)
	for i := 0; i < 50; i++ {
		ctx.AddBar(OHLCV{Time: int64(i), Close: float64(i)})
	}
	for i := 0; i < 50; i++ {
		ctx.AddBar(OHLCV{Time: int64(i), Close: float64(i)})
	}
	arrowCtx := NewArrowContext(ctx)

	up := arrowCtx.GetOrCreateSeries("up")
	down := arrowCtx.GetOrCreateSeries("down")
	truerange := arrowCtx.GetOrCreateSeries("truerange")

	if up == down || up == truerange || down == truerange {
		t.Error("Expected distinct Series instances for different variables")
	}

	if len(arrowCtx.LocalSeries) != 3 {
		t.Errorf("Expected 3 Series, got %d", len(arrowCtx.LocalSeries))
	}
}

func TestArrowContext_SeriesCapacity(t *testing.T) {
	ctx := New("TEST", "1h", 100)
	for i := 0; i < 100; i++ {
		ctx.AddBar(OHLCV{Time: int64(i), Close: float64(i)})
	}

	arrowCtx := NewArrowContext(ctx)
	s := arrowCtx.GetOrCreateSeries("test")

	if s.Capacity() != 100 {
		t.Errorf("Expected capacity 100, got %d", s.Capacity())
	}
}

func TestArrowContext_AdvanceAll(t *testing.T) {
	ctx := New("TEST", "1h", 10)
	for i := 0; i < 10; i++ {
		ctx.AddBar(OHLCV{Time: int64(i), Close: float64(i)})
	}
	for i := 0; i < 10; i++ {
		ctx.AddBar(OHLCV{Time: int64(i), Close: float64(i)})
	}
	arrowCtx := NewArrowContext(ctx)

	s1 := arrowCtx.GetOrCreateSeries("var1")
	s2 := arrowCtx.GetOrCreateSeries("var2")

	s1.Set(10.0)
	s2.Set(20.0)

	if s1.Position() != 0 || s2.Position() != 0 {
		t.Error("Expected initial position 0")
	}

	arrowCtx.AdvanceAll()

	if s1.Position() != 1 {
		t.Errorf("Expected var1 position 1 after AdvanceAll, got %d", s1.Position())
	}
	if s2.Position() != 1 {
		t.Errorf("Expected var2 position 1 after AdvanceAll, got %d", s2.Position())
	}

	arrowCtx.AdvanceAll()

	if s1.Position() != 2 || s2.Position() != 2 {
		t.Error("Expected both Series at position 2 after second AdvanceAll")
	}
}

func TestArrowContext_GetSeries_NotFound(t *testing.T) {
	ctx := New("TEST", "1h", 10)
	for i := 0; i < 10; i++ {
		ctx.AddBar(OHLCV{Time: int64(i), Close: float64(i)})
	}
	for i := 0; i < 10; i++ {
		ctx.AddBar(OHLCV{Time: int64(i), Close: float64(i)})
	}
	arrowCtx := NewArrowContext(ctx)

	_, err := arrowCtx.GetSeries("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent Series")
	}

	arrowCtx.GetOrCreateSeries("exists")
	s, err := arrowCtx.GetSeries("exists")
	if err != nil {
		t.Errorf("Unexpected error for existing Series: %v", err)
	}
	if s == nil {
		t.Error("Expected non-nil Series")
	}
}

func TestArrowContext_Reset(t *testing.T) {
	ctx := New("TEST", "1h", 100)
	for i := 0; i < 100; i++ {
		ctx.AddBar(OHLCV{Time: int64(i), Close: float64(i)})
	}
	arrowCtx := NewArrowContext(ctx)

	s1 := arrowCtx.GetOrCreateSeries("var1")
	s2 := arrowCtx.GetOrCreateSeries("var2")

	for i := 0; i < 5; i++ {
		s1.Set(float64(i))
		s2.Set(float64(i * 2))
		arrowCtx.AdvanceAll()
	}

	if s1.Position() != 5 || s2.Position() != 5 {
		t.Error("Expected both Series at position 5")
	}

	arrowCtx.Reset(2)

	if s1.Position() != 2 {
		t.Errorf("Expected var1 position 2 after reset, got %d", s1.Position())
	}
	if s2.Position() != 2 {
		t.Errorf("Expected var2 position 2 after reset, got %d", s2.Position())
	}
}

func TestArrowContext_ContextWrapping(t *testing.T) {
	ctx := New("BTCUSDT", "1D", 50)
	ctx.AddBar(OHLCV{Time: 1000, Close: 100.0})

	arrowCtx := NewArrowContext(ctx)

	if arrowCtx.Context.Symbol != "BTCUSDT" {
		t.Errorf("Expected wrapped Context symbol BTCUSDT, got %s", arrowCtx.Context.Symbol)
	}

	if arrowCtx.Context.Timeframe != "1D" {
		t.Errorf("Expected wrapped Context timeframe 1D, got %s", arrowCtx.Context.Timeframe)
	}

	if len(arrowCtx.Context.Data) != 1 {
		t.Errorf("Expected 1 bar in wrapped Context, got %d", len(arrowCtx.Context.Data))
	}
}

func TestArrowContext_IsolationBetweenInstances(t *testing.T) {
	ctx := New("TEST", "1h", 50)
	for i := 0; i < 50; i++ {
		ctx.AddBar(OHLCV{Time: int64(i), Close: float64(i)})
	}

	arrow1 := NewArrowContext(ctx)
	arrow2 := NewArrowContext(ctx)

	s1 := arrow1.GetOrCreateSeries("shared_name")
	s2 := arrow2.GetOrCreateSeries("shared_name")

	if s1 == s2 {
		t.Error("Expected distinct Series instances for different ArrowContext instances")
	}

	s1.Set(100.0)
	s2.Set(200.0)

	if s1.GetCurrent() == s2.GetCurrent() {
		t.Error("Expected isolated Series values between ArrowContext instances")
	}
}

func BenchmarkArrowContext_GetOrCreateSeries(b *testing.B) {
	ctx := New("TEST", "1h", 1000)
	arrowCtx := NewArrowContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = arrowCtx.GetOrCreateSeries("benchmark_var")
	}
}

func BenchmarkArrowContext_AdvanceAll(b *testing.B) {
	ctx := New("TEST", "1h", 10000)
	arrowCtx := NewArrowContext(ctx)

	for i := 0; i < 10; i++ {
		arrowCtx.GetOrCreateSeries(string(rune('a' + i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arrowCtx.AdvanceAll()
		arrowCtx.Reset(0)
	}
}

/* TestArrowContext_EdgeCases validates boundary conditions and error handling */
func TestArrowContext_EdgeCases(t *testing.T) {
	t.Run("empty context data", func(t *testing.T) {
		ctx := New("TEST", "1h", 0)
		arrowCtx := NewArrowContext(ctx)

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for zero capacity Series")
			}
		}()
		arrowCtx.GetOrCreateSeries("test")
	})

	t.Run("series name collision", func(t *testing.T) {
		ctx := New("TEST", "1h", 10)
		for i := 0; i < 10; i++ {
			ctx.AddBar(OHLCV{Time: int64(i), Close: float64(i)})
		}
		arrowCtx := NewArrowContext(ctx)

		s1 := arrowCtx.GetOrCreateSeries("variable")
		s1.Set(100.0)

		s2 := arrowCtx.GetOrCreateSeries("variable")
		if s1 != s2 {
			t.Error("Expected same Series instance for identical name")
		}
		if s2.GetCurrent() != 100.0 {
			t.Error("Expected value preservation on name collision")
		}
	})

	t.Run("advance without series created", func(t *testing.T) {
		ctx := New("TEST", "1h", 10)
		for i := 0; i < 10; i++ {
			ctx.AddBar(OHLCV{Time: int64(i), Close: float64(i)})
		}
		arrowCtx := NewArrowContext(ctx)

		arrowCtx.AdvanceAll()

		if len(arrowCtx.LocalSeries) != 0 {
			t.Error("Expected empty LocalSeries after AdvanceAll with no Series created")
		}
	})

	t.Run("reset to invalid position", func(t *testing.T) {
		ctx := New("TEST", "1h", 10)
		for i := 0; i < 10; i++ {
			ctx.AddBar(OHLCV{Time: int64(i), Close: float64(i)})
		}
		arrowCtx := NewArrowContext(ctx)
		arrowCtx.GetOrCreateSeries("test")

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for reset beyond capacity")
			}
		}()
		arrowCtx.Reset(100)
	})

	t.Run("get series before create", func(t *testing.T) {
		ctx := New("TEST", "1h", 10)
		for i := 0; i < 10; i++ {
			ctx.AddBar(OHLCV{Time: int64(i), Close: float64(i)})
		}
		arrowCtx := NewArrowContext(ctx)

		_, err := arrowCtx.GetSeries("nonexistent")
		if err == nil {
			t.Error("Expected error for nonexistent Series")
		}
		if err.Error() != `arrow context: Series "nonexistent" not found` {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("special characters in series name", func(t *testing.T) {
		ctx := New("TEST", "1h", 10)
		for i := 0; i < 10; i++ {
			ctx.AddBar(OHLCV{Time: int64(i), Close: float64(i)})
		}
		arrowCtx := NewArrowContext(ctx)

		names := []string{"var-with-dash", "var.with.dot", "var_with_underscore", "var123", ""}
		for _, name := range names {
			s := arrowCtx.GetOrCreateSeries(name)
			if s == nil {
				t.Errorf("Failed to create Series with name %q", name)
			}
		}

		if len(arrowCtx.LocalSeries) != len(names) {
			t.Errorf("Expected %d Series, got %d", len(names), len(arrowCtx.LocalSeries))
		}
	})

	t.Run("massive series count", func(t *testing.T) {
		ctx := New("TEST", "1h", 100)
		for i := 0; i < 100; i++ {
			ctx.AddBar(OHLCV{Time: int64(i), Close: float64(i)})
		}
		arrowCtx := NewArrowContext(ctx)

		for i := 0; i < 100; i++ {
			arrowCtx.GetOrCreateSeries(fmt.Sprintf("var%d", i))
		}

		if len(arrowCtx.LocalSeries) != 100 {
			t.Errorf("Expected 100 Series, got %d", len(arrowCtx.LocalSeries))
		}

		arrowCtx.AdvanceAll()

		for i := 0; i < 100; i++ {
			s, _ := arrowCtx.GetSeries(fmt.Sprintf("var%d", i))
			if s.Position() != 1 {
				t.Errorf("Series var%d: expected position 1, got %d", i, s.Position())
			}
		}
	})
}

/* TestArrowContext_ConcurrentUsage documents thread-safety assumptions (ArrowContext NOT thread-safe by design) */
func TestArrowContext_ConcurrentUsage(t *testing.T) {
	t.Run("concurrent access not supported", func(t *testing.T) {
		ctx := New("TEST", "1h", 100)
		for i := 0; i < 100; i++ {
			ctx.AddBar(OHLCV{Time: int64(i), Close: float64(i)})
		}

		arrow1 := NewArrowContext(ctx)
		arrow2 := NewArrowContext(ctx)

		s1 := arrow1.GetOrCreateSeries("shared_name")
		s2 := arrow2.GetOrCreateSeries("shared_name")

		if s1 == s2 {
			t.Error("Expected isolated Series instances across ArrowContext instances")
		}

		s1.Set(100.0)
		s2.Set(200.0)

		if s1.GetCurrent() == s2.GetCurrent() {
			t.Error("Expected independent values across ArrowContext instances")
		}
	})
}
