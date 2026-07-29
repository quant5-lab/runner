package regression

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Strategies here run end-to-end but produce zero trades by design because their signal
// source requires external ±1 wiring absent from the SBERP-1h fixture. A .pine.skip on
// any of them is stale; the guard rejects it.
//
// Strategies absent from both this list and tvAlignmentCases() are legitimately pending
// and are never flagged.
func inertByDesignStrategies() []string {
	return []string{
		"top10/ultima.pine",
	}
}

func TestTop10_StaleSkipGuard(t *testing.T) {
	root := projectRootFromCwd()
	csvFixturesDir := filepath.Join(root, "tests", "regression", "tv_reference", "fixtures")
	strategiesDir := filepath.Join(root, "strategies")

	for _, tc := range tvAlignmentCases() {
		if !strings.HasPrefix(tc.Strategy, "top10/") {
			continue
		}
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			csvPath := filepath.Join(csvFixturesDir, tc.CSV)
			if _, err := os.Stat(csvPath); err != nil {
				return // not yet strict-verified
			}
			skipPath := filepath.Join(strategiesDir, tc.Strategy+".skip")
			if _, err := os.Stat(skipPath); err == nil {
				t.Errorf(
					"stale .pine.skip: %q is strict-verified (TV reference %q present) but %q still exists — delete it",
					tc.Strategy, tc.CSV, skipPath,
				)
			}
		})
	}

	for _, strategy := range inertByDesignStrategies() {
		strategy := strategy
		name := strings.TrimSuffix(filepath.Base(strategy), ".pine")
		t.Run(name, func(t *testing.T) {
			skipPath := filepath.Join(strategiesDir, strategy+".skip")
			if _, err := os.Stat(skipPath); err == nil {
				t.Errorf(
					"stale .pine.skip: %q is inert by design and executes end-to-end, but %q still exists — delete it",
					strategy, skipPath,
				)
			}
		})
	}
}
