//go:build integration

package integration

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/*
TestSecurityChartType_TypedConstructors verifies that each chart-type modifier

	with all-literal ticker constructor arguments emits the typed Go constructor
	and the correct mapper initialisation call.
*/
func TestSecurityChartType_TypedConstructors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		script          string
		wantConstructor string
		wantMapper      string
	}{
		{
			name: "heikinashi",
			script: `//@version=5
indicator("HA Test", overlay=true)
haClose = request.security(ticker.heikinashi(syminfo.tickerid), timeframe.period, close)
plot(haClose)
`,
			wantConstructor: "&ticker.HeikinAshiTransformer{}",
			wantMapper:      "BuildIdentityMapping",
		},
		{
			name: "renko",
			script: `//@version=5
indicator("Renko Test", overlay=true)
renkoClose = request.security(ticker.renko(syminfo.tickerid, "ATR", 14), timeframe.period, close)
plot(renkoClose)
`,
			wantConstructor: `ticker.NewRenkoTransformer("ATR", 14)`,
			wantMapper:      "BuildMappingFromTransform",
		},
		{
			name: "kagi",
			script: `//@version=5
indicator("Kagi Test", overlay=true)
kagiClose = request.security(ticker.kagi(syminfo.tickerid, 1), timeframe.period, close)
plot(kagiClose)
`,
			wantConstructor: "ticker.NewKagiTransformer(1)",
			wantMapper:      "BuildMappingFromTransform",
		},
		{
			name: "linebreak",
			script: `//@version=5
indicator("LineBreak Test", overlay=true)
lbClose = request.security(ticker.linebreak(syminfo.tickerid, 3), timeframe.period, close)
plot(lbClose)
`,
			wantConstructor: "ticker.NewLineBreakTransformer(3)",
			wantMapper:      "BuildMappingFromTransform",
		},
		{
			name: "pointfigure",
			script: `//@version=5
indicator("PF Test", overlay=true)
pfClose = request.security(ticker.pointfigure(syminfo.tickerid, "close", "ATR", 14, 3), timeframe.period, close)
plot(pfClose)
`,
			wantConstructor: `ticker.NewPointFigureTransformer("close", "ATR", 14, 3)`,
			wantMapper:      "BuildMappingFromTransform",
		},
	}

	exec := util.NewPineExecutor(t)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			generatedCode, _ := exec.GenerateCode(t, "chart_type_"+tc.name, tc.script)

			if !strings.Contains(generatedCode, tc.wantConstructor) {
				t.Errorf("generated code missing constructor %q", tc.wantConstructor)
			}
			if !strings.Contains(generatedCode, tc.wantMapper) {
				t.Errorf("generated code missing mapper call %q", tc.wantMapper)
			}
			if !strings.Contains(generatedCode, "_transformResult") {
				t.Errorf("generated code missing _transformResult variable")
			}

			if err := exec.CompileCode(t, generatedCode); err != nil {
				t.Fatalf("compilation failed: %v", err)
			}
		})
	}
}

/*
TestSecurityChartType_NonLiteralFallback verifies that a runtime (non-literal)

	ticker constructor argument falls back to NewTransformer with the modifier name
	and still compiles.
*/
func TestSecurityChartType_NonLiteralFallback(t *testing.T) {
	t.Parallel()
	script := `//@version=5
indicator("Renko Runtime", overlay=true)
boxSize = input.float(14, "Box Size")
renkoClose = request.security(ticker.renko(syminfo.tickerid, "ATR", boxSize), timeframe.period, close)
plot(renkoClose)
`
	exec := util.NewPineExecutor(t)
	generatedCode, _ := exec.GenerateCode(t, "chart_type_noliteral", script)

	if !strings.Contains(generatedCode, `ticker.NewTransformer("RENKO")`) {
		t.Errorf("expected fallback to ticker.NewTransformer(\"RENKO\") for non-literal box size")
	}
	if strings.Contains(generatedCode, "ticker.NewRenkoTransformer(") {
		t.Errorf("typed constructor must not be emitted when arguments are not all literals")
	}

	if err := exec.CompileCode(t, generatedCode); err != nil {
		t.Fatalf("compilation failed: %v", err)
	}
}
