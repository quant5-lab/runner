package regression

import (
	"os"
	"path/filepath"
	"testing"
)

/* TestTickerNamespace_HeikinashiSecurity_V4Pipeline verifies that v4 bare ticker functions
 * reach the runtime correctly after the preprocessor qualifies them into the ticker.* namespace.
 * Both the official spelling and the community-prevalent typo must resolve to ticker.heikinashi. */
func TestTickerNamespace_HeikinashiSecurity_V4Pipeline(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	cases := []struct {
		name     string
		funcCall string // the symbol argument for security()
	}{
		{"canonical_heikinashi", "heikinashi(syminfo.tickerid)"},
		{"typo_heikenashi", "heikenashi(syminfo.tickerid)"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testDir := t.TempDir()

			/* v4 script — preprocessor transforms heikinashi/heikenashi → ticker.heikinashi,
			 * security → request.security, and study → indicator */
			strategy := "//@version=4\n" +
				"study(\"HA Security Test\", overlay=true)\n" +
				"haClose = security(" + tc.funcCall + ", \"1D\", close)\n" +
				"plot(haClose, \"HA Close\")\n"

			strategyPath := filepath.Join(testDir, "ha_test.pine")
			if err := os.WriteFile(strategyPath, []byte(strategy), 0644); err != nil {
				t.Fatal(err)
			}

			/* Hourly base data — 240 bars (10 days of hourly candles) */
			baseData := generateTestOHLCV(240, 3600)
			basePath := filepath.Join(testDir, "HATEST_1h.json")
			if err := os.WriteFile(basePath, []byte(baseData), 0644); err != nil {
				t.Fatal(err)
			}

			/* Daily base-symbol data for the security() call — the runtime fetches the
			 * base symbol file and applies the heikinashi transform in-memory */
			haData := generateTestOHLCV(10, 86400)
			haPath := filepath.Join(testDir, "HATEST_1D.json")
			if err := os.WriteFile(haPath, []byte(haData), 0644); err != nil {
				t.Fatal(err)
			}

			result := compileAndRun(t, strategyPath, basePath, testDir, projectRoot, "HATEST", testDir)

			haClose, ok := result.Indicators["HA Close"]
			if !ok {
				t.Fatalf("Expected 'HA Close' indicator, got: %v", getIndicatorNames(result.Indicators))
			}

			nonNull := countNonNull(haClose.Data)
			if nonNull == 0 {
				t.Errorf("%s: 'HA Close' has no non-null values — ticker namespace transform did not route to ticker.heikinashi", tc.funcCall)
			}
		})
	}
}
