package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

// ohlcvAt builds a minimal OHLCV bar with the given Unix timestamp (seconds).
func ohlcvAt(timeSec int64) context.OHLCV { return context.OHLCV{Time: timeSec, Close: 1} }

// slit builds a string AST literal.
func slit(s string) *ast.Literal { return &ast.Literal{Value: s} }

// taTimeCall builds a ta.time(...) CallExpression matching what codegen emits
// when serialising time() inside a security() value expression.
func taTimeCall(args ...ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "ta"}, Property: &ast.Identifier{Name: "time"}},
		Arguments: args,
	}
}

// bareTimeCall builds a plain time(...) CallExpression (callee is a bare Identifier).
func bareTimeCall(args ...ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "time"},
		Arguments: args,
	}
}

// timedCtx creates a *context.Context with timestamped bars, a named timeframe,
// and a PeriodAnchor for session-aware slot alignment.
func timedCtx(timeframe string, anchor context.PeriodAnchor, bars ...context.OHLCV) *context.Context {
	return &context.Context{
		Data:         bars,
		Timeframe:    timeframe,
		PeriodAnchor: anchor,
	}
}

var (
	nyseAnchor = context.PeriodAnchor{Timezone: "America/New_York", SessionOpenMinute: 9*60 + 30}
	moexAnchor = context.PeriodAnchor{Timezone: "Europe/Moscow", SessionOpenMinute: 7 * 60}
	utcAnchor  = context.PeriodAnchor{}
)

// Reference timestamps derived from 2024-01-18, a regular NYSE trading day.
// nyseOpen4h and nyseOpen4hUTC are the slot starts for the same bar under NYSE
// and UTC anchors respectively; their difference is what the discriminant test asserts.
const (
	nyseOpen4h    = int64(1705588200) // 2024-01-18 14:30:00 UTC (09:30 ET, NYSE 4h slot start)
	nyseOpen4hUTC = int64(1705579200) // 2024-01-18 12:00:00 UTC (preceding UTC 4h boundary)
)

func TestTime_NoArg_ReturnsBarTimestampMilliseconds(t *testing.T) {
	cases := []struct {
		name       string
		barTimeSec int64
	}{
		{"normal_trading_timestamp", nyseOpen4h},
		{"unix_epoch_boundary", 0},
		{"large_recent_timestamp", 1_700_000_000},
		{"single_second", 1},
	}
	ev := NewStreamingBarEvaluator()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := timedCtx("", nyseAnchor, ohlcvAt(c.barTimeSec))
			got, err := ev.EvaluateAtBar(taTimeCall(), ctx, 0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertFloat64(t, "time()", got, float64(c.barTimeSec*1000), 0)
		})
	}
}

func TestTime_WithTimeframeArg_AlignsBarToSlotStart(t *testing.T) {
	cases := []struct {
		name       string
		barTimeSec int64
		wantSec    int64
	}{
		{"bar_exactly_at_slot_start", nyseOpen4h, nyseOpen4h},
		{"bar_one_hour_into_slot", nyseOpen4h + 3600, nyseOpen4h},
		{"bar_two_hours_into_slot", nyseOpen4h + 7200, nyseOpen4h},
		{"bar_one_second_before_next_slot", nyseOpen4h + 4*3600 - 1, nyseOpen4h},
		{"bar_at_next_slot_boundary", nyseOpen4h + 4*3600, nyseOpen4h + 4*3600},
	}
	ev := NewStreamingBarEvaluator()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := timedCtx("4h", nyseAnchor, ohlcvAt(c.barTimeSec))
			got, err := ev.EvaluateAtBar(taTimeCall(slit("4h")), ctx, 0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertFloat64(t, "time(\"4h\")", got, float64(c.wantSec*1000), 0)
		})
	}
}

func TestTime_TimeframeStringForms_PureNumericEquivalentToSuffixed(t *testing.T) {
	// TradingView accepts bare numeric strings ("240") as minute counts alongside
	// suffixed canonical forms ("4h").
	cases := []struct{ canonical, numeric string }{
		{"4h", "240"},
		{"1h", "60"},
		{"30m", "30"},
	}
	ev := NewStreamingBarEvaluator()
	ctx := timedCtx("4h", nyseAnchor, ohlcvAt(nyseOpen4h+3600))
	for _, c := range cases {
		t.Run(c.canonical+"_vs_"+c.numeric, func(t *testing.T) {
			wantV, err := ev.EvaluateAtBar(taTimeCall(slit(c.canonical)), ctx, 0)
			if err != nil {
				t.Fatalf("canonical form error: %v", err)
			}
			gotV, err := ev.EvaluateAtBar(taTimeCall(slit(c.numeric)), ctx, 0)
			if err != nil {
				t.Fatalf("numeric form error: %v", err)
			}
			assertFloat64(t, c.canonical+"=="+c.numeric, gotV, wantV, 0)
		})
	}
}

func TestTime_Float64NumericLiteral_TreatedAsMinutes(t *testing.T) {
	ev := NewStreamingBarEvaluator()
	ctx := timedCtx("4h", nyseAnchor, ohlcvAt(nyseOpen4h))
	wantV, err := ev.EvaluateAtBar(taTimeCall(slit("240")), ctx, 0)
	if err != nil {
		t.Fatalf("string form error: %v", err)
	}
	gotV, err := ev.EvaluateAtBar(taTimeCall(lit(240)), ctx, 0)
	if err != nil {
		t.Fatalf("float64 form error: %v", err)
	}
	assertFloat64(t, "float64 240.0 == string \"240\"", gotV, wantV, 0)
}

func TestTime_TimeframePeriod_ResolvesFromContextTimeframe(t *testing.T) {
	// timeframe.period is the canonical Pine AST form; it must resolve via the
	// secondary context's own Timeframe, not the primary's.
	tfPeriod := &ast.MemberExpression{
		Object:   &ast.Identifier{Name: "timeframe"},
		Property: &ast.Identifier{Name: "period"},
	}
	cases := []struct {
		name      string
		timeframe string
	}{
		{"4h_context", "4h"},
		{"1h_context", "1h"},
		{"1D_context", "1D"},
	}
	ev := NewStreamingBarEvaluator()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := timedCtx(c.timeframe, nyseAnchor, ohlcvAt(nyseOpen4h))
			want, err := ev.EvaluateAtBar(taTimeCall(slit(c.timeframe)), ctx, 0)
			if err != nil {
				t.Fatalf("explicit tf error: %v", err)
			}
			got, err := ev.EvaluateAtBar(taTimeCall(tfPeriod), ctx, 0)
			if err != nil {
				t.Fatalf("timeframe.period error: %v", err)
			}
			assertFloat64(t, "timeframe.period == ctx.Timeframe", got, want, 0)
		})
	}
}

func TestTime_BarIndexOutOfRange_ReturnsNaNWithoutError(t *testing.T) {
	ev := NewStreamingBarEvaluator()
	ctx := timedCtx("4h", nyseAnchor, ohlcvAt(nyseOpen4h))
	cases := []struct {
		name   string
		barIdx int
	}{
		{"one_past_end", 1},
		{"far_past_end", 100},
		{"negative_index", -1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ev.EvaluateAtBar(taTimeCall(), ctx, c.barIdx)
			if err != nil {
				t.Fatalf("expected nil error for out-of-bounds idx %d, got: %v", c.barIdx, err)
			}
			if !math.IsNaN(got) {
				t.Errorf("expected NaN for bar index %d, got %v", c.barIdx, got)
			}
		})
	}
}

func TestTime_DifferentAnchors_ProduceDifferentSlotBoundaries(t *testing.T) {
	// The secondary PeriodAnchor must govern slot boundaries, not the primary's.
	// 2024-01-18 09:30 ET (= 14:30 UTC) falls in different 4h slots depending
	// on anchor origin: NYSE (09:30 ET) vs UTC (00:00 UTC).
	ev := NewStreamingBarEvaluator()

	ctxNYSE := timedCtx("4h", nyseAnchor, ohlcvAt(nyseOpen4h))
	gotNYSE, err := ev.EvaluateAtBar(taTimeCall(slit("4h")), ctxNYSE, 0)
	if err != nil {
		t.Fatalf("NYSE anchor error: %v", err)
	}

	ctxUTC := timedCtx("4h", utcAnchor, ohlcvAt(nyseOpen4h))
	gotUTC, err := ev.EvaluateAtBar(taTimeCall(slit("4h")), ctxUTC, 0)
	if err != nil {
		t.Fatalf("UTC anchor error: %v", err)
	}

	assertFloat64(t, "NYSE slot start (ms)", gotNYSE, float64(nyseOpen4h*1000), 0)
	assertFloat64(t, "UTC slot start (ms)", gotUTC, float64(nyseOpen4hUTC*1000), 0)
	if gotNYSE == gotUTC {
		t.Error("NYSE and UTC anchors produced the same slot start — anchor has no effect")
	}
}

func TestTime_ConsecutiveBarsInSameSlot_AllReturnSameSlotStart(t *testing.T) {
	hourlyBars := []context.OHLCV{
		ohlcvAt(nyseOpen4h),         // 09:30 ET — slot start
		ohlcvAt(nyseOpen4h + 3600),  // 10:30 ET
		ohlcvAt(nyseOpen4h + 7200),  // 11:30 ET
		ohlcvAt(nyseOpen4h + 10800), // 12:30 ET
	}
	ctx := timedCtx("4h", nyseAnchor, hourlyBars...)
	ev := NewStreamingBarEvaluator()
	wantMs := float64(nyseOpen4h * 1000)
	for i := range hourlyBars {
		got, err := ev.EvaluateAtBar(taTimeCall(slit("4h")), ctx, i)
		if err != nil {
			t.Fatalf("bar %d: unexpected error: %v", i, err)
		}
		assertFloat64(t, "bar "+string(rune('0'+i))+" slot start", got, wantMs, 0)
	}
}

func TestTime_AliasSymmetry_BareAndTANamespaceIdentical(t *testing.T) {
	// "time" and "ta.time" are registered aliases and must produce identical results.
	cases := []struct {
		name string
		args []ast.Expression
	}{
		{"no_args", nil},
		{"string_tf", []ast.Expression{slit("4h")}},
		{"numeric_tf", []ast.Expression{slit("240")}},
	}
	ctx := timedCtx("4h", nyseAnchor, ohlcvAt(nyseOpen4h))
	ev := NewStreamingBarEvaluator()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			wantV, err := ev.EvaluateAtBar(taTimeCall(c.args...), ctx, 0)
			if err != nil {
				t.Fatalf("ta.time error: %v", err)
			}
			gotV, err := ev.EvaluateAtBar(bareTimeCall(c.args...), ctx, 0)
			if err != nil {
				t.Fatalf("bare time error: %v", err)
			}
			assertFloat64(t, "time == ta.time", gotV, wantV, 0)
		})
	}
}
