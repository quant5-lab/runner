package regression

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

/*
TestSecurityTA_MathFunctions_AllCompileAndRun verifies that all 18 math.* call
functions survive the full codegen→compile→execute pipeline when evaluated inside
request.security(), producing at least one non-null output bar.

Literal arguments are used for angle and domain-sensitive functions so the smoke
test does not depend on OHLCV shape — only that the dispatch registry routes each
name correctly.

generateTestOHLCV invariants used for inline assertions:
  - close[i] = 50050 + i  (always > 0, integer-valued)
  - close - close = 0     on every bar
  - math.pow(2,10) = 1024, math.sign(close) = 1, math.cos(0) = 1, math.sin(0) = 0
  - math.round_to_mintick is a passthrough in security context → output equals close
*/
func TestSecurityTA_MathFunctions_AllCompileAndRun(t *testing.T) {
	strategy := `//@version=5
indicator("Security Math Functions", overlay=false)
logc    = request.security(syminfo.tickerid, "1D", math.log(close))
log10c  = request.security(syminfo.tickerid, "1D", math.log10(close))
expc    = request.security(syminfo.tickerid, "1D", math.exp(close - close))
sqrtc   = request.security(syminfo.tickerid, "1D", math.sqrt(close))
pow2    = request.security(syminfo.tickerid, "1D", math.pow(2, 10))
rndc    = request.security(syminfo.tickerid, "1D", math.round(close))
mintick = request.security(syminfo.tickerid, "1D", math.round_to_mintick(close))
flrc    = request.security(syminfo.tickerid, "1D", math.floor(close))
celc    = request.security(syminfo.tickerid, "1D", math.ceil(close))
sgnc    = request.security(syminfo.tickerid, "1D", math.sign(close))
sinz    = request.security(syminfo.tickerid, "1D", math.sin(0))
cosz    = request.security(syminfo.tickerid, "1D", math.cos(0))
tanz    = request.security(syminfo.tickerid, "1D", math.tan(0))
asinz   = request.security(syminfo.tickerid, "1D", math.asin(0))
acosq   = request.security(syminfo.tickerid, "1D", math.acos(0))
atanv   = request.security(syminfo.tickerid, "1D", math.atan(1))
torad   = request.security(syminfo.tickerid, "1D", math.toradians(180))
todeg   = request.security(syminfo.tickerid, "1D", math.todegrees(0))
plot(logc,    "LOGC")
plot(log10c,  "LOG10C")
plot(expc,    "EXPC")
plot(sqrtc,   "SQRTC")
plot(pow2,    "POW2")
plot(rndc,    "RNDC")
plot(mintick, "MINTICK")
plot(flrc,    "FLRC")
plot(celc,    "CELC")
plot(sgnc,    "SGNC")
plot(sinz,    "SINZ")
plot(cosz,    "COSZ")
plot(tanz,    "TANZ")
plot(asinz,   "ASINZ")
plot(acosq,   "ACOSQ")
plot(atanv,   "ATANV")
plot(torad,   "TORAD")
plot(todeg,   "TODEG")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "math-functions.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "MATHFN_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(40, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "MATHFN", testDir)

	for _, name := range []string{
		"LOGC", "LOG10C", "EXPC", "SQRTC", "POW2", "RNDC", "MINTICK",
		"FLRC", "CELC", "SGNC", "SINZ", "COSZ", "TANZ", "ASINZ", "ACOSQ",
		"ATANV", "TORAD", "TODEG",
	} {
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

	/* math.pow(2, 10): bar-independent literal expression — must equal 1024 on every
	   non-null bar, proving two-argument dispatch routes correctly. */
	t.Run("POW2_constant_value", func(t *testing.T) {
		ind, ok := result.Indicators["POW2"]
		if !ok {
			t.Fatal("POW2 absent from output")
		}
		for i, bar := range ind.Data {
			if v, ok := getFloatValue(bar); ok {
				if math.Abs(v-1024.0) > 1e-9 {
					t.Errorf("bar %d: math.pow(2,10) = %.6f, want 1024", i, v)
				}
			}
		}
	})

	/* math.sign(close): close[i] = 50050+i > 0 on every bar → sign must always be 1.
	   A failure here means positive-input branch is broken. */
	t.Run("SGNC_positive_input_always_one", func(t *testing.T) {
		ind, ok := result.Indicators["SGNC"]
		if !ok {
			t.Fatal("SGNC absent from output")
		}
		for i, bar := range ind.Data {
			if v, ok := getFloatValue(bar); ok {
				if v != 1.0 {
					t.Errorf("bar %d: math.sign(close) = %.1f, want 1 (close always > 0)", i, v)
				}
			}
		}
	})

	/* math.sin(0) = 0 and math.cos(0) = 1 are exact trig identities.
	   A failure indicates the trig dispatch or argument evaluation is broken. */
	t.Run("SINZ_zero_argument", func(t *testing.T) {
		ind, ok := result.Indicators["SINZ"]
		if !ok {
			t.Fatal("SINZ absent from output")
		}
		for i, bar := range ind.Data {
			if v, ok := getFloatValue(bar); ok {
				if math.Abs(v) > 1e-12 {
					t.Errorf("bar %d: math.sin(0) = %.15f, want 0", i, v)
				}
			}
		}
	})

	t.Run("COSZ_zero_argument", func(t *testing.T) {
		ind, ok := result.Indicators["COSZ"]
		if !ok {
			t.Fatal("COSZ absent from output")
		}
		for i, bar := range ind.Data {
			if v, ok := getFloatValue(bar); ok {
				if math.Abs(v-1.0) > 1e-12 {
					t.Errorf("bar %d: math.cos(0) = %.15f, want 1", i, v)
				}
			}
		}
	})

	/* math.round_to_mintick passes value through unchanged in the security evaluator
	   because tick size is unavailable. Output must equal close on every non-null bar.
	   close[secBar i] = 50050+i; base bar 1 maps to secBar 0 → close = 50050. */
	t.Run("MINTICK_passthrough", func(t *testing.T) {
		mintickInd, ok := result.Indicators["MINTICK"]
		if !ok {
			t.Fatal("MINTICK absent from output")
		}
		rndc, ok := result.Indicators["RNDC"]
		if !ok {
			t.Fatal("RNDC absent from output (used as close proxy)")
		}
		for i := range mintickInd.Data {
			mv, mOK := getFloatValue(mintickInd.Data[i])
			rv, rOK := getFloatValue(rndc.Data[i])
			if mOK && rOK {
				if math.Abs(mv-rv) > 1e-9 {
					t.Errorf("bar %d: round_to_mintick(close) = %.6f, round(close) = %.6f — passthrough broken", i, mv, rv)
				}
			}
		}
	})
}

/*
TestSecurityTA_MathConstants_EvaluateToKnownValues verifies that the six math.*
namespace constants (pi, e, phi, rphi, huge, tiny) resolve to their IEEE 754 values
when used as request.security expressions. Constants have no bar dependency, so
every non-null output bar must carry the same exact value.
*/
func TestSecurityTA_MathConstants_EvaluateToKnownValues(t *testing.T) {
	strategy := `//@version=5
indicator("Math Constants In Security", overlay=false)
pi_   = request.security(syminfo.tickerid, "1D", math.pi)
e_    = request.security(syminfo.tickerid, "1D", math.e)
phi_  = request.security(syminfo.tickerid, "1D", math.phi)
rphi_ = request.security(syminfo.tickerid, "1D", math.rphi)
huge_ = request.security(syminfo.tickerid, "1D", math.huge)
tiny_ = request.security(syminfo.tickerid, "1D", math.tiny)
plot(pi_,   "PI")
plot(e_,    "E")
plot(phi_,  "PHI")
plot(rphi_, "RPHI")
plot(huge_, "HUGE")
plot(tiny_, "TINY")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "math-constants.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "MATHK_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "MATHK", testDir)

	cases := []struct {
		name string
		want float64
		tol  float64
	}{
		{"PI", math.Pi, 1e-14},
		{"E", math.E, 1e-14},
		{"PHI", 1.6180339887498948482, 1e-14},
		{"RPHI", 0.6180339887498948482, 1e-14},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ind, ok := result.Indicators[c.name]
			if !ok {
				t.Fatalf("indicator %q absent from output", c.name)
			}
			if countNonNull(ind.Data) == 0 {
				t.Fatalf("indicator %q produced zero non-null values", c.name)
			}
			for i, bar := range ind.Data {
				if v, ok := getFloatValue(bar); ok {
					if math.Abs(v-c.want) > c.tol {
						t.Errorf("bar %d: math.%s = %.15f, want %.15f", i, c.name, v, c.want)
					}
				}
			}
		})
	}

	/* math.huge and math.tiny are valid float64 but have extreme magnitudes;
	   verify pipeline survives without crashing and produces non-null output. */
	for _, name := range []string{"HUGE", "TINY"} {
		t.Run(name+"_non_null", func(t *testing.T) {
			ind, ok := result.Indicators[name]
			if !ok {
				t.Fatalf("indicator %q absent from output", name)
			}
			if countNonNull(ind.Data) == 0 {
				t.Errorf("indicator %q produced zero non-null values", name)
			}
		})
	}
}

/*
TestSecurityTA_MathLog_DomainProtection verifies that math.log and math.log10 return
NaN (null output) for non-positive inputs when evaluated inside request.security().

generateTestOHLCV: close[i] = 50050+i, so close - 51000 = -(950-i).
With 40 bars (i = 0..39), close - 51000 ranges from -950 to -911 — all negative —
so every security bar must produce null (NaN).
*/
func TestSecurityTA_MathLog_DomainProtection(t *testing.T) {
	strategy := `//@version=5
indicator("Math Log Domain Protection", overlay=false)
logNeg   = request.security(syminfo.tickerid, "1D", math.log(close - 51000))
log10Neg = request.security(syminfo.tickerid, "1D", math.log10(close - 51000))
plot(logNeg,   "LOGNEG")
plot(log10Neg, "LOG10NEG")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "math-log-domain.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "LOGDOM_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(40, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "LOGDOM", testDir)

	for _, name := range []string{"LOGNEG", "LOG10NEG"} {
		t.Run(name, func(t *testing.T) {
			ind, ok := result.Indicators[name]
			if !ok {
				t.Fatalf("indicator %q absent from output", name)
			}
			if n := countNonNull(ind.Data); n != 0 {
				t.Errorf("indicator %q: %d non-null bars for all-negative input, want 0 (domain protection broken)", name, n)
			}
		})
	}
}

/*
TestSecurityTA_ScalarBuiltins_AllCompileAndRun verifies that nz, ta.nz (alias),
na, ta.na (alias), int, and float each survive the full codegen→compile→execute
pipeline when evaluated inside request.security(), producing at least one non-null
output bar.
*/
func TestSecurityTA_ScalarBuiltins_AllCompileAndRun(t *testing.T) {
	strategy := `//@version=5
indicator("Scalar Builtins In Security", overlay=false)
nz_bare  = request.security(syminfo.tickerid, "1D", nz(ta.change(close), 0))
nz_qual  = request.security(syminfo.tickerid, "1D", ta.nz(ta.change(close), 0))
na_func  = request.security(syminfo.tickerid, "1D", na(ta.change(close)))
na_qual  = request.security(syminfo.tickerid, "1D", ta.na(ta.change(close)))
int_c    = request.security(syminfo.tickerid, "1D", int(close))
float_c  = request.security(syminfo.tickerid, "1D", float(close))
plot(nz_bare, "NZ_BARE")
plot(nz_qual, "NZ_QUAL")
plot(na_func, "NA_FUNC")
plot(na_qual, "NA_QUAL")
plot(int_c,   "INT_C")
plot(float_c, "FLOAT_C")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "scalar-builtins.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "SCALAR_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(40, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "SCALAR", testDir)

	for _, name := range []string{"NZ_BARE", "NZ_QUAL", "NA_FUNC", "NA_QUAL", "INT_C", "FLOAT_C"} {
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
TestSecurityTA_Nz_WarmupNaNReplacement verifies that nz(ta.change(close), 0)
evaluated inside request.security() returns the replacement value (0) at the
warmup bar where ta.change produces NaN, and the actual value (1) on all
subsequent bars.

generateTestOHLCV emits monotonically rising closes (Δ = 1/bar), so
ta.change(close) = 1.0 on every post-warmup security bar.

1-bar lookahead_off lag: base bar 1 maps to secBar 0 (change warmup → nz returns 0).
Base bars 2+ map to secBars 1+ (change = 1.0 → nz returns 1.0).
*/
func TestSecurityTA_Nz_WarmupNaNReplacement(t *testing.T) {
	strategy := `//@version=5
indicator("Nz Warmup NaN Replacement", overlay=false)
nz_ = request.security(syminfo.tickerid, "1D", nz(ta.change(close), 0))
plot(nz_, "NZ")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "nz-warmup.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "NZWARM_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(40, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "NZWARM", testDir)

	ind, ok := result.Indicators["NZ"]
	if !ok {
		t.Fatal("NZ indicator absent from output")
	}
	vals := extractValues(ind.Data)

	t.Run("warmup_bar_replaced_with_zero", func(t *testing.T) {
		if len(vals) < 2 {
			t.Fatal("insufficient bars")
		}
		if math.IsNaN(vals[1]) {
			t.Error("bar 1: nz(NaN, 0) produced null, want 0 — NaN replacement broken")
		} else if vals[1] != 0.0 {
			t.Errorf("bar 1: nz(NaN, 0) = %.6f, want 0", vals[1])
		}
	})

	t.Run("post_warmup_passthrough", func(t *testing.T) {
		for i := 2; i < len(vals); i++ {
			if math.IsNaN(vals[i]) {
				t.Errorf("bar %d: unexpected null (nz should pass through ta.change = 1.0)", i)
				continue
			}
			if math.Abs(vals[i]-1.0) > 1e-9 {
				t.Errorf("bar %d: nz(ta.change(close), 0) = %.6f, want 1.0 (change always 1 on monotonic data)", i, vals[i])
			}
		}
	})
}

/*
TestSecurityTA_Na_FunctionFormInSecurity verifies that na(x) evaluated inside
request.security() returns 1.0 when x is NaN and 0.0 when x is a finite value.

na(ta.change(close)): secBar 0 has NaN change (warmup) → 1.0; secBars 1+ have
change = 1.0 → 0.0. With the 1-bar lag, base bar 1 maps to secBar 0.
*/
func TestSecurityTA_Na_FunctionFormInSecurity(t *testing.T) {
	strategy := `//@version=5
indicator("Na Function Form In Security", overlay=false)
na_ = request.security(syminfo.tickerid, "1D", na(ta.change(close)))
plot(na_, "NA")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "na-function.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "NAFUNC_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(40, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "NAFUNC", testDir)

	ind, ok := result.Indicators["NA"]
	if !ok {
		t.Fatal("NA indicator absent from output")
	}
	vals := extractValues(ind.Data)

	t.Run("warmup_nan_detected_as_true", func(t *testing.T) {
		if len(vals) < 2 {
			t.Fatal("insufficient bars")
		}
		if math.IsNaN(vals[1]) {
			t.Error("bar 1: na(NaN) produced null, want 1.0 — na function broken")
		} else if vals[1] != 1.0 {
			t.Errorf("bar 1: na(NaN) = %.6f, want 1.0", vals[1])
		}
	})

	t.Run("post_warmup_finite_value_is_false", func(t *testing.T) {
		for i := 2; i < len(vals); i++ {
			if math.IsNaN(vals[i]) {
				t.Errorf("bar %d: unexpected null from na(1.0)", i)
				continue
			}
			if vals[i] != 0.0 {
				t.Errorf("bar %d: na(ta.change(close)) = %.6f, want 0.0 (change is finite post-warmup)", i, vals[i])
			}
		}
	})
}

/*
TestSecurityTA_Nz_TaNzAliasProducesIdenticalOutput verifies that ta.nz resolves
to the same handler as bare nz and produces bit-identical output on every bar —
catching any registration divergence between the two names.
*/
func TestSecurityTA_Nz_TaNzAliasProducesIdenticalOutput(t *testing.T) {
	strategy := `//@version=5
indicator("Nz Alias Identity", overlay=false)
bare = request.security(syminfo.tickerid, "1D", nz(ta.change(close), 0))
qual = request.security(syminfo.tickerid, "1D", ta.nz(ta.change(close), 0))
plot(bare, "BARE")
plot(qual, "QUAL")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "nz-alias.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "NZALIAS_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(40, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "NZALIAS", testDir)

	bareInd, ok := result.Indicators["BARE"]
	if !ok {
		t.Fatal("BARE indicator absent from output")
	}
	qualInd, ok := result.Indicators["QUAL"]
	if !ok {
		t.Fatal("QUAL indicator absent from output")
	}

	bareVals := extractValues(bareInd.Data)
	qualVals := extractValues(qualInd.Data)

	if len(bareVals) != len(qualVals) {
		t.Fatalf("output length mismatch: nz=%d ta.nz=%d", len(bareVals), len(qualVals))
	}
	for i := range bareVals {
		bNaN, qNaN := math.IsNaN(bareVals[i]), math.IsNaN(qualVals[i])
		if bNaN != qNaN {
			t.Errorf("bar %d: null mismatch — nz=%v ta.nz=%v", i, bNaN, qNaN)
			continue
		}
		if !bNaN && bareVals[i] != qualVals[i] {
			t.Errorf("bar %d: nz=%.9f ta.nz=%.9f — alias diverges", i, bareVals[i], qualVals[i])
		}
	}
}
