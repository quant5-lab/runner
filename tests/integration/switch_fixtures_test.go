//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

func TestSwitchFixtures(t *testing.T) {
	t.Parallel()
	fixturesDir := "../fixtures/integration"

	entries, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Fatalf("fixtures directory: %v", err)
	}

	exec := util.NewPineExecutor(t)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".pine" {
			continue
		}

		name := entry.Name()
		if len(name) < 12 || name[:12] != "test-switch-" {
			continue
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			content, err := os.ReadFile(filepath.Join(fixturesDir, name))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}

			output := exec.ExecuteScript(t, name[:len(name)-5], string(content))
			if output == nil {
				t.Error("no output")
			}
		})
	}
}

func TestSwitchFixtures_ExpectedValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		fixture  string
		plot     string
		expected float64
	}{
		{"test-switch-form1-basic.pine", "Result", 20.0},
		{"test-switch-form1-default.pine", "Result", -1.0},
		{"test-switch-form2-boolean.pine", "Result", 2.0},
		{"test-switch-expression-ops.pine", "Result", 50.0},
		{"test-switch-indent-1space.pine", "Result", 20.0},
		{"test-switch-indent-3space.pine", "Result", 100.0},
		{"test-switch-indent-5space.pine", "Result", 30.0},
		{"test-switch-indent-1tab.pine", "Result", 20.0},
		{"test-switch-indent-2tab.pine", "Result", 10.0},
		{"test-switch-indent-2case-3body.pine", "Result", 20.0},
		{"test-switch-indent-3case-7body.pine", "Result", 50.0},
		{"test-switch-form2-indent-3space.pine", "Result", 2.0},
		{"test-switch-inline-form1-basic.pine", "Result", 20.0},
		{"test-switch-inline-form1-default.pine", "Result", -1.0},
		{"test-switch-inline-form2-boolean.pine", "Result", 2.0},
		{"test-switch-inline-expression-ops.pine", "Result", 50.0},
	}

	exec := util.NewPineExecutor(t)

	for _, tt := range tests {
		t.Run(tt.fixture, func(t *testing.T) {
			t.Parallel()
			content, err := os.ReadFile(filepath.Join("../fixtures/integration", tt.fixture))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}

			output := exec.ExecuteScript(t, tt.fixture[:len(tt.fixture)-5], string(content))
			vals := exec.ExtractPlotValues(t, output, tt.plot)

			if len(vals) < 1 {
				t.Fatal("no data points")
			}
			if vals[0] != tt.expected {
				t.Errorf("got %f, want %f", vals[0], tt.expected)
			}
		})
	}
}
