package regression

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

/*
TestSecurityTA_ExpressionSource_BinaryExprCompileAndRun verifies that SMA, EMA,
RMA, STDEV, TSI, and CUM each survive the full codegen→compile→execute pipeline
when their TA source argument is a binary expression rather than a plain
identifier — the capability generalised by the ast.Expression source refactor.
*/
func TestSecurityTA_ExpressionSource_BinaryExprCompileAndRun(t *testing.T) {
	strategy := `//@version=5
indicator("ExprSource BinaryExpr", overlay=false)
sma5   = request.security(syminfo.tickerid, "1D", ta.sma(close - open, 5))
ema5   = request.security(syminfo.tickerid, "1D", ta.ema(close - open, 5))
rma5   = request.security(syminfo.tickerid, "1D", ta.rma(close - open, 5))
stdev5 = request.security(syminfo.tickerid, "1D", ta.stdev(close - open, 5))
tsi    = request.security(syminfo.tickerid, "1D", ta.tsi(close - open, 5, 13))
cum_   = request.security(syminfo.tickerid, "1D", ta.cum(close - open))
plot(sma5,   "SMA5")
plot(ema5,   "EMA5")
plot(rma5,   "RMA5")
plot(stdev5, "STDEV5")
plot(tsi,    "TSI")
plot(cum_,   "CUM")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "exprsrc-compile.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "EXSRC_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(40, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "EXSRC", testDir)

	for _, name := range []string{"SMA5", "EMA5", "RMA5", "STDEV5", "TSI", "CUM"} {
		t.Run(name, func(t *testing.T) {
			ind, ok := result.Indicators[name]
			if !ok {
				t.Fatalf("indicator %q absent from output", name)
			}
			if countNonNull(ind.Data) == 0 {
				t.Errorf("indicator %q produced zero non-null values across 40 bars", name)
			}
		})
	}
}

/*
TestSecurityTA_ExpressionSource_ArithmeticCorrectness verifies exact output
values for TA functions whose source is a binary expression that evaluates to a
known constant (close − open = 50 on every generateTestOHLCV bar).

Invariants:
  - SMA(constant 50, 5)  = 50 for all post-warmup bars
  - STDEV(constant 50, 5) = 0 for all post-warmup bars (zero variance)
  - CUM(constant 50)[bar k] = 50 × k (1-bar lookahead_off lag: secBarIdx = k−1)
*/
func TestSecurityTA_ExpressionSource_ArithmeticCorrectness(t *testing.T) {
	strategy := `//@version=5
indicator("ExprSource Arithmetic", overlay=false)
sma5   = request.security(syminfo.tickerid, "1D", ta.sma(close - open, 5))
stdev5 = request.security(syminfo.tickerid, "1D", ta.stdev(close - open, 5))
cum_   = request.security(syminfo.tickerid, "1D", ta.cum(close - open))
plot(sma5,   "SMA5")
plot(stdev5, "STDEV5")
plot(cum_,   "CUM")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "exprsrc-arith.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "EXARITH_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "EXARITH", testDir)

	t.Run("SMA_constant_source", func(t *testing.T) {
		ind, ok := result.Indicators["SMA5"]
		if !ok {
			t.Fatal("SMA5 indicator absent from output")
		}
		/* SMA(5) warmup: period−1 = 4 secBars + 1-bar lag → first valid base bar = 5. */
		for i, bar := range ind.Data {
			if v, ok := getFloatValue(bar); ok {
				if math.Abs(v-50.0) > 1e-6 {
					t.Errorf("bar %d: sma(close-open,5) = %.9f, want 50.0", i, v)
				}
			}
		}
	})

	t.Run("STDEV_constant_source", func(t *testing.T) {
		ind, ok := result.Indicators["STDEV5"]
		if !ok {
			t.Fatal("STDEV5 indicator absent from output")
		}
		/* A constant source has zero variance; the expression-source path must preserve this. */
		for i, bar := range ind.Data {
			if v, ok := getFloatValue(bar); ok {
				if math.Abs(v-0.0) > 1e-9 {
					t.Errorf("bar %d: stdev(close-open,5) = %.9f, want 0.0", i, v)
				}
			}
		}
	})

	t.Run("CUM_constant_source", func(t *testing.T) {
		ind, ok := result.Indicators["CUM"]
		if !ok {
			t.Fatal("CUM indicator absent from output")
		}
		vals := extractValues(ind.Data)
		/* 1-bar lookahead_off lag: base bar k maps to secBarIdx k−1.
		   cum(close−open)[secBarIdx s] = 50×(s+1), so cum[base bar k] = 50×k. */
		for k := 1; k < len(vals); k++ {
			if math.IsNaN(vals[k]) {
				continue
			}
			want := 50.0 * float64(k)
			if math.Abs(vals[k]-want) > 1e-6 {
				t.Errorf("bar %d: cum(close-open) = %.6f, want %.6f", k, vals[k], want)
			}
		}
	})
}

/*
TestSecurityTA_ExpressionSource_ChainedTASource verifies that the output of one
TA function is correctly accepted as the source expression of another through the
full codegen→compile→execute pipeline.  Two patterns are exercised:

  - sma(ema(close, 3), 3): outer SMA over inner EMA — first valid base bar 5
  - sma(rma(close, 5), 3): outer SMA over inner RMA — first valid base bar 7

Both use a monotonically rising input; the composed output must also be
monotonically rising once both warmup windows are satisfied.
*/
func TestSecurityTA_ExpressionSource_ChainedTASource(t *testing.T) {
	cases := []struct {
		name    string
		expr    string
		plot    string
		warmup  int // first valid base bar (0-based)
		symbol  string
		dataLen int
	}{
		{"sma_of_ema", "ta.sma(ta.ema(close, 3), 3)", "OUT", 5, "CHAIN1", 30},
		{"sma_of_rma", "ta.sma(ta.rma(close, 5), 3)", "OUT", 7, "CHAIN2", 30},
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			strategy := "//@version=5\n" +
				"indicator(\"ChainedTA " + tt.name + "\", overlay=false)\n" +
				"out = request.security(syminfo.tickerid, \"1D\", " + tt.expr + ")\n" +
				"plot(out, \"" + tt.plot + "\")\n"

			testDir := t.TempDir()
			strategyPath := filepath.Join(testDir, "chained.pine")
			if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
				t.Fatal(err)
			}
			dataPath := filepath.Join(testDir, tt.symbol+"_1D.json")
			if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(tt.dataLen, 86400)), 0644); err != nil {
				t.Fatal(err)
			}

			result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, tt.symbol, testDir)

			ind, ok := result.Indicators[tt.plot]
			if !ok {
				t.Fatalf("indicator %q absent from output", tt.plot)
			}
			vals := extractValues(ind.Data)

			for i := 0; i < tt.warmup && i < len(vals); i++ {
				if !math.IsNaN(vals[i]) {
					t.Errorf("bar %d: expected NaN during warmup (first valid = %d), got %.6f",
						i, tt.warmup, vals[i])
				}
			}

			for i := tt.warmup; i < len(vals); i++ {
				if math.IsNaN(vals[i]) {
					t.Errorf("bar %d: unexpected NaN post-warmup", i)
				}
			}

			/* Monotone rising input → composed TA output must be monotone rising. */
			var prev float64
			for i := tt.warmup; i < len(vals); i++ {
				if math.IsNaN(vals[i]) {
					prev = vals[i]
					continue
				}
				if i > tt.warmup && !math.IsNaN(prev) && vals[i] <= prev {
					t.Errorf("bar %d: %s = %.6f not > bar %d = %.6f (monotone violated)",
						i, tt.expr, vals[i], i-1, prev)
				}
				prev = vals[i]
			}
		})
	}
}

/*
TestSecurityTA_ExpressionSource_CacheIsolation verifies that two TA calls with
identical function and period but distinct source expressions maintain independent
ForwardSeriesBuffer state through the full codegen→compile→execute pipeline.

generateTestOHLCV: close[i] = 50050+i, open[i] = 50000+i, so close−open = 50 on
every bar.  With 1-bar lookahead_off lag (secBarIdx = k−1 at base bar k):

	sma(close, 5)[k] = 50047 + k
	sma(open,  5)[k] = 49997 + k
	difference       = 50 (constant, isolates state corruption from value proximity)
*/
func TestSecurityTA_ExpressionSource_CacheIsolation(t *testing.T) {
	strategy := `//@version=5
indicator("ExprSource CacheIsolation", overlay=false)
smaClose = request.security(syminfo.tickerid, "1D", ta.sma(close, 5))
smaOpen  = request.security(syminfo.tickerid, "1D", ta.sma(open,  5))
plot(smaClose, "SMACLOSE")
plot(smaOpen,  "SMAOPEN")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "cache-iso.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "CACISO_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "CACISO", testDir)

	closeInd, ok := result.Indicators["SMACLOSE"]
	if !ok {
		t.Fatal("SMACLOSE indicator absent from output")
	}
	openInd, ok := result.Indicators["SMAOPEN"]
	if !ok {
		t.Fatal("SMAOPEN indicator absent from output")
	}

	closeVals := extractValues(closeInd.Data)
	openVals := extractValues(openInd.Data)

	/* Exact values: sma(close,5)[k] = 50047+k, sma(open,5)[k] = 49997+k, diff = 50.
	   SMA(5) warmup: period−1 = 4 secBars + 1-bar lag → first valid base bar = 6. */
	for k := 6; k < len(closeVals) && k < len(openVals); k++ {
		if math.IsNaN(closeVals[k]) || math.IsNaN(openVals[k]) {
			t.Errorf("bar %d: unexpected NaN post-warmup (close=%.6f open=%.6f)",
				k, closeVals[k], openVals[k])
			continue
		}

		wantClose := 50047.0 + float64(k)
		wantOpen := 49997.0 + float64(k)

		if math.Abs(closeVals[k]-wantClose) > 1e-6 {
			t.Errorf("bar %d: sma(close,5) = %.9f, want %.9f", k, closeVals[k], wantClose)
		}
		if math.Abs(openVals[k]-wantOpen) > 1e-6 {
			t.Errorf("bar %d: sma(open,5) = %.9f, want %.9f", k, openVals[k], wantOpen)
		}

		diff := closeVals[k] - openVals[k]
		if math.Abs(diff-50.0) > 1e-6 {
			t.Errorf("bar %d: sma(close)−sma(open) = %.9f, want 50.0 (state corruption check)", k, diff)
		}
	}
}
