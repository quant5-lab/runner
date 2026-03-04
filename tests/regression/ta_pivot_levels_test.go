package regression

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

/*
TestPivotPointLevels_AllTypesCompileAndRun verifies all 6 pivot types survive
the full codegen→compile→execute pipeline and produce non-null P values after
the first anchor fires.
*/
func TestPivotPointLevels_AllTypesCompileAndRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("Pivot All Types", overlay=false)
trad = ta.pivot_point_levels("Traditional", bar_index == 5)
fib  = ta.pivot_point_levels("Fibonacci",   bar_index == 5)
wood = ta.pivot_point_levels("Woodie",      bar_index == 5)
clas = ta.pivot_point_levels("Classic",     bar_index == 5)
dm   = ta.pivot_point_levels("DM",          bar_index == 5)
cam  = ta.pivot_point_levels("Camarilla",   bar_index == 5)
tradP = array.get(trad, 0)
fibP  = array.get(fib,  0)
woodP = array.get(wood, 0)
clasP = array.get(clas, 0)
dmP   = array.get(dm,   0)
camP  = array.get(cam,  0)
plot(tradP, "Traditional P")
plot(fibP,  "Fibonacci P")
plot(woodP, "Woodie P")
plot(clasP, "Classic P")
plot(dmP,   "DM P")
plot(camP,  "Camarilla P")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "pivot-all-types.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "PIVTYPES_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "PIVTYPES", testDir)

	for _, name := range []string{
		"Traditional P", "Fibonacci P", "Woodie P", "Classic P", "DM P", "Camarilla P",
	} {
		t.Run(name, func(t *testing.T) {
			ind, ok := result.Indicators[name]
			if !ok {
				t.Fatalf("indicator %q absent from output", name)
			}
			if countNonNull(ind.Data) == 0 {
				t.Errorf("pivot type %q produced zero non-null P values", name)
			}
		})
	}
}

/*
TestPivotPointLevels_ArrayGetElementAccess verifies that individual level indices
(P=0, R1=1, R2=2, S1=5, S2=6) are all accessible via array.get and produce
non-null float output after the anchor fires.
*/
func TestPivotPointLevels_ArrayGetElementAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("Pivot Element Access", overlay=false)
levels = ta.pivot_point_levels("Traditional", bar_index == 5)
p  = array.get(levels, 0)
r1 = array.get(levels, 1)
r2 = array.get(levels, 2)
s1 = array.get(levels, 5)
s2 = array.get(levels, 6)
plot(p,  "P")
plot(r1, "R1")
plot(r2, "R2")
plot(s1, "S1")
plot(s2, "S2")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "pivot-element-access.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "PIVELM_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "PIVELM", testDir)

	for _, name := range []string{"P", "R1", "R2", "S1", "S2"} {
		t.Run(name, func(t *testing.T) {
			ind, ok := result.Indicators[name]
			if !ok {
				t.Fatalf("element %q absent from output", name)
			}
			vals := extractValues(ind.Data)
			nonNaN := 0
			for _, v := range vals {
				if !math.IsNaN(v) {
					nonNaN++
				}
			}
			if nonNaN == 0 {
				t.Errorf("element %q produced no non-NaN values after anchor", name)
			}
		})
	}
}

/*
TestPivotPointLevels_ArraySizeIs11 verifies array.size(levels) always returns 11
after the anchor fires — confirming the fixed-width contract of ComputePivotLevels.
*/
func TestPivotPointLevels_ArraySizeIs11(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("Pivot Array Size", overlay=false)
levels = ta.pivot_point_levels("Traditional", bar_index == 5)
sz = array.size(levels)
plot(sz, "Size")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "pivot-array-size.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "PIVSIZE_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "PIVSIZE", testDir)

	ind, ok := result.Indicators["Size"]
	if !ok {
		t.Fatal("'Size' indicator absent from output")
	}

	vals := extractValues(ind.Data)
	for i, v := range vals {
		if math.IsNaN(v) {
			continue
		}
		if v != 11 {
			t.Errorf("bar %d: array.size(levels) = %.0f, want 11", i, v)
		}
	}
}

/*
TestPivotPointLevels_PreAnchorBarsAreNaN verifies that levels are NaN on every bar
before the first anchor fires — the ForwardSeriesBuffer initial-nil sentinel.
*/
func TestPivotPointLevels_PreAnchorBarsAreNaN(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	const anchorBar = 5

	strategy := `//@version=5
indicator("Pivot Pre-Anchor NaN", overlay=false)
levels = ta.pivot_point_levels("Traditional", bar_index == 5)
p = array.get(levels, 0)
plot(p, "P")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "pivot-preanchor.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "PIVPRE_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(15, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "PIVPRE", testDir)

	ind, ok := result.Indicators["P"]
	if !ok {
		t.Fatal("'P' indicator absent from output")
	}

	vals := extractValues(ind.Data)
	for i := 0; i < anchorBar && i < len(vals); i++ {
		if !math.IsNaN(vals[i]) {
			t.Errorf("bar %d (pre-anchor): P = %.6f, want NaN", i, vals[i])
		}
	}

	nonNaNAfter := 0
	for i := anchorBar; i < len(vals); i++ {
		if !math.IsNaN(vals[i]) {
			nonNaNAfter++
		}
	}
	if nonNaNAfter == 0 {
		t.Error("no non-NaN P values after anchor fired at bar 5")
	}
}

/*
TestPivotPointLevels_CarryForwardBetweenAnchors verifies that with developing=false,
levels remain constant on every off-anchor bar between two successive anchor events.
*/
func TestPivotPointLevels_CarryForwardBetweenAnchors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("Pivot Carry Forward", overlay=false)
levels = ta.pivot_point_levels("Traditional", bar_index == 5, false)
p = array.get(levels, 0)
plot(p, "P")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "pivot-carry.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "PIVCARRY_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(15, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "PIVCARRY", testDir)

	ind, ok := result.Indicators["P"]
	if !ok {
		t.Fatal("'P' indicator absent from output")
	}

	vals := extractValues(ind.Data)

	/* collect all non-NaN P values from bar 5 onward */
	var postAnchor []float64
	for i := 5; i < len(vals); i++ {
		if !math.IsNaN(vals[i]) {
			postAnchor = append(postAnchor, vals[i])
		}
	}
	if len(postAnchor) < 2 {
		t.Fatalf("need at least 2 post-anchor bars to verify carry-forward, got %d", len(postAnchor))
	}

	ref := postAnchor[0]
	for i, v := range postAnchor {
		if math.Abs(v-ref) > 1e-9 {
			t.Errorf("carry-forward violated at offset %d: P = %.9f, want %.9f", i, v, ref)
		}
	}
}

/*
TestPivotPointLevels_HistoricalSubscriptAccess verifies that array.get(levels[1], 0)
at bar N equals array.get(levels, 0) at bar N-1 — confirming ArraySeries history
indexing is wired through the same ForwardSeriesBuffer offset mechanism as Series.
*/
func TestPivotPointLevels_HistoricalSubscriptAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("Pivot Historical Access", overlay=false)
levels = ta.pivot_point_levels("Traditional", bar_index == 5, false)
p     = array.get(levels,    0)
pPrev = array.get(levels[1], 0)
plot(p,     "P")
plot(pPrev, "P Prev")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "pivot-historical.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "PIVHIST_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "PIVHIST", testDir)

	p, ok := result.Indicators["P"]
	if !ok {
		t.Fatal("'P' indicator absent from output")
	}
	pPrev, ok := result.Indicators["P Prev"]
	if !ok {
		t.Fatal("'P Prev' indicator absent from output")
	}

	pVals := extractValues(p.Data)
	prevVals := extractValues(pPrev.Data)

	for i := 1; i < len(pVals) && i < len(prevVals); i++ {
		if math.IsNaN(pVals[i-1]) || math.IsNaN(prevVals[i]) {
			continue
		}
		if math.Abs(pVals[i-1]-prevVals[i]) > 1e-9 {
			t.Errorf("bar %d: array.get(levels[1],0) = %.9f, array.get(levels,0) at bar %d = %.9f, want equal",
				i, prevVals[i], i-1, pVals[i-1])
		}
	}
}

/*
TestPivotPointLevels_DevelopingTrueRecomputes verifies that developing=true causes
P to change on every off-anchor bar as the accumulator grows, while developing=false
holds P constant — confirming the two modes produce distinct outputs between anchors.
*/
func TestPivotPointLevels_DevelopingTrueRecomputes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	strategy := `//@version=5
indicator("Pivot Developing Mode", overlay=false)
fixed   = ta.pivot_point_levels("Traditional", bar_index == 3, false)
develop = ta.pivot_point_levels("Traditional", bar_index == 3, true)
fixedP   = array.get(fixed,   0)
developP = array.get(develop, 0)
plot(fixedP,   "Fixed P")
plot(developP, "Develop P")
`
	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "pivot-developing.pine")
	if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "PIVDEV_1D.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(15, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	result := compileAndRun(t, strategyPath, dataPath, testDir, projectRoot, "PIVDEV", testDir)

	fixedInd, ok := result.Indicators["Fixed P"]
	if !ok {
		t.Fatal("'Fixed P' indicator absent from output")
	}
	developInd, ok := result.Indicators["Develop P"]
	if !ok {
		t.Fatal("'Develop P' indicator absent from output")
	}

	fixedVals := extractValues(fixedInd.Data)
	developVals := extractValues(developInd.Data)

	/* collect off-anchor values (bars 4+ are between anchors) */
	var fixedOffAnchor, developOffAnchor []float64
	for i := 4; i < len(fixedVals) && i < len(developVals); i++ {
		if !math.IsNaN(fixedVals[i]) && !math.IsNaN(developVals[i]) {
			fixedOffAnchor = append(fixedOffAnchor, fixedVals[i])
			developOffAnchor = append(developOffAnchor, developVals[i])
		}
	}

	if len(fixedOffAnchor) < 2 {
		t.Fatalf("need at least 2 off-anchor bars to compare modes, got %d", len(fixedOffAnchor))
	}

	/* developing=false: all off-anchor values must be identical */
	for i := 1; i < len(fixedOffAnchor); i++ {
		if math.Abs(fixedOffAnchor[i]-fixedOffAnchor[0]) > 1e-9 {
			t.Errorf("fixed mode: P changed at offset %d (%.9f → %.9f), want constant",
				i, fixedOffAnchor[0], fixedOffAnchor[i])
		}
	}

	/* developing=true: at least one off-anchor value must differ from the first */
	allSame := true
	for i := 1; i < len(developOffAnchor); i++ {
		if math.Abs(developOffAnchor[i]-developOffAnchor[0]) > 1e-9 {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("developing=true: P was constant across all off-anchor bars, want recomputation")
	}
}
