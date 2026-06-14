package regression

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTop10_StaleSkipGuard uses TV reference CSV presence as the strict-verified
// signal: a strategy with a passing golden but no operator-side CSV is legitimately
// still pending and keeps its .skip without triggering this guard.
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
}
