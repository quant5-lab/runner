package codegen

import (
	"strings"
	"testing"
)

func mustCompileVersioned(t *testing.T, source string) string {
	t.Helper()
	code, err := compilePineScriptVersioned(t, source)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	return code
}

// extractionContexts lists callers whose source argument is dispatched through
// g.extractSeriesExpression → extractCallExpression → extractTimeBuiltin.
// Eligibility: the host must call g.extractSeriesExpression on its first arg.
// Functions with a dedicated argument extractor (e.g. ta.highest via
// TAArgumentExtractor) route through a different chain and do not belong here.
var extractionContexts = []struct {
	name string
	wrap func(inner string) string
}{
	{
		"change",
		func(inner string) string { return "change(" + inner + ")" },
	},
}

func pineScript(versionHeader, extraDecl, assignRHS string) string {
	return versionHeader + `
strategy("s")
` + extraDecl + `x = ` + assignRHS + `
if x != 0
    strategy.entry("L", strategy.long)
`
}

func TestTimeBuiltinExtractor_ZeroArgEmitsBareTimestamp(t *testing.T) {
	for _, ctx := range extractionContexts {
		ctx := ctx
		t.Run(ctx.name, func(t *testing.T) {
			code := mustCompileVersioned(t, pineScript("//@version=5", "", ctx.wrap("time()")))
			if strings.Contains(code, `featuregap.Record("time"`) {
				t.Errorf("time() must not fall back to featuregap:\n%s", code)
			}
			if !strings.Contains(code, "ctx.Data[ctx.BarIndex].Time * 1000") {
				t.Errorf("zero-arg time() must emit bare bar-timestamp ms expression:\n%s", code)
			}
		})
	}
}

func TestTimeBuiltinExtractor_SingleArgEmitsAlignedBoundary(t *testing.T) {
	cases := []struct {
		name     string
		call     string
		wantInTF string // expected timeframe expression in generated code
	}{
		{"timeframe_period", "time(timeframe.period)", "ctx.Timeframe"},
		{"intraday_numeric_60", `time("60")`, `"60"`},
		{"intraday_numeric_240", `time("240")`, `"240"`},
		{"daily_literal", `time("1D")`, `"1D"`},
	}

	for _, ctx := range extractionContexts {
		for _, tc := range cases {
			ctx, tc := ctx, tc
			t.Run(ctx.name+"/"+tc.name, func(t *testing.T) {
				code := mustCompileVersioned(t, pineScript("//@version=5", "", ctx.wrap(tc.call)))
				if strings.Contains(code, `featuregap.Record("time"`) {
					t.Errorf("time(tf) must not fall back to featuregap:\n%s", code)
				}
				if !strings.Contains(code, "BarOpenTimeAtTimeframe") {
					t.Errorf("time(tf) must emit BarOpenTimeAtTimeframe boundary call:\n%s", code)
				}
				if !strings.Contains(code, tc.wantInTF) {
					t.Errorf("time(tf) must contain timeframe expr %q:\n%s", tc.wantInTF, code)
				}
			})
		}
	}
}

func TestTimeBuiltinExtractor_SessionCall_EmitsVersionedFunction(t *testing.T) {
	for _, ctx := range extractionContexts {
		ctx := ctx
		t.Run(ctx.name, func(t *testing.T) {
			code := mustCompileVersioned(t, pineScript("//@version=5", "",
				ctx.wrap(`time(timeframe.period, "0950-1645")`)))
			if strings.Contains(code, `featuregap.Record("time"`) {
				t.Errorf("[%s] must not emit featuregap:\n%s", ctx.name, code)
			}
			if !strings.Contains(code, "session.TimeFuncWithVersion") {
				t.Errorf("[%s] must emit session.TimeFuncWithVersion:\n%s", ctx.name, code)
			}
			if strings.Contains(code, "session.TimeFunc(") {
				t.Errorf("[%s] unversioned session.TimeFunc( must not appear:\n%s", ctx.name, code)
			}
		})
	}
}

func TestTimeBuiltinExtractor_SessionCall_VersionBinding(t *testing.T) {
	versions := []struct {
		header string
		want   string
	}{
		{"//@version=4", ", 4)"},
		{"//@version=5", ", 5)"},
	}

	sessionForms := []struct {
		name    string
		decl    string
		sessArg string
	}{
		{"literal", "", `"0930-1600"`},
		{"identifier", `mySession = "0930-1600"` + "\n", "mySession"},
	}

	for _, ver := range versions {
		for _, sess := range sessionForms {
			ver, sess := ver, sess
			t.Run(ver.want+"/"+sess.name, func(t *testing.T) {
				code := mustCompileVersioned(t, pineScript(ver.header, sess.decl,
					`change(time(timeframe.period, `+sess.sessArg+`))`))
				if !strings.Contains(code, ver.want) {
					t.Errorf("version suffix %q missing for %s / %s session:\n%s",
						ver.want, ver.header, sess.name, code)
				}
				if strings.Contains(code, "session.TimeFunc(") {
					t.Errorf("unversioned session.TimeFunc( must not appear:\n%s", code)
				}
			})
		}
	}
}

func TestTimeBuiltinExtractor_TimeCloseNoSession_ResolvesSeries(t *testing.T) {
	for _, ctx := range extractionContexts {
		ctx := ctx
		t.Run(ctx.name, func(t *testing.T) {
			code := mustCompileVersioned(t, pineScript("//@version=5", "",
				ctx.wrap("time_close(timeframe.period)")))
			if strings.Contains(code, `featuregap.Record("time_close"`) {
				t.Errorf("[%s] time_close(tf) must not emit featuregap:\n%s", ctx.name, code)
			}
		})
	}
}

// time_close(timeframe, session) is not yet implemented in TimeSeriesLifecycle;
// the named featuregap ensures callers receive an explicit NaN rather than
// silently-wrong output.
func TestTimeBuiltinExtractor_TimeCloseWithSession_EmitsNamedFeaturegap(t *testing.T) {
	code := mustCompileVersioned(t, pineScript("//@version=5", "",
		`change(time_close(timeframe.period, "0950-1645"))`))
	if !strings.Contains(code, `"time_close_session"`) {
		t.Errorf("time_close(tf,sess) must emit named featuregap 'time_close_session':\n%s", code)
	}
}

func TestTimeBuiltinExtractor_NonTimeBuiltins_NotIntercepted(t *testing.T) {
	cases := []struct {
		name string
		call string
	}{
		{"rsi", "rsi(close, 14)"},
		{"sma", "sma(close, 20)"},
		{"ema", "ema(close, 10)"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			code := mustCompileVersioned(t, pineScript("//@version=5", "", "change("+tc.call+")"))
			if strings.Contains(code, `featuregap.Record("time"`) {
				t.Errorf("[%s] extractTimeBuiltin must not intercept non-time calls:\n%s", tc.name, code)
			}
		})
	}
}
