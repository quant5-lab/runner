package series_naming

import (
	"testing"

	"github.com/quant5-lab/runner/codegen/source_identity"
)

/* TestStatefulIndicatorNamer_GenerateName tests stateful indicator naming with hash inclusion */
func TestStatefulIndicatorNamer_GenerateName(t *testing.T) {
	namer := NewStatefulIndicatorNamer()

	tests := []struct {
		name           string
		indicatorName  string
		period         int
		sourceHash     string
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:          "rma with hash",
			indicatorName: "rma",
			period:        14,
			sourceHash:    "abc12345",
			wantContains:  []string{"_rma_", "14", "abc12345"},
		},
		{
			name:          "ema with different hash",
			indicatorName: "ema",
			period:        20,
			sourceHash:    "def67890",
			wantContains:  []string{"_ema_", "20", "def67890"},
		},
		{
			name:          "sma with empty hash",
			indicatorName: "sma",
			period:        50,
			sourceHash:    "",
			wantContains:  []string{"_sma_", "50"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := namer.GenerateName(tt.indicatorName, tt.period, tt.sourceHash)

			/* Check all required substrings are present */
			for _, substr := range tt.wantContains {
				if !containsSubstring(result, substr) {
					t.Errorf("GenerateName() = %q, should contain %q", result, substr)
				}
			}

			/* Check forbidden substrings are absent */
			for _, substr := range tt.wantNotContain {
				if containsSubstring(result, substr) {
					t.Errorf("GenerateName() = %q, should not contain %q", result, substr)
				}
			}
		})
	}
}

/* TestStatefulIndicatorNamer_UniqueNamesForDifferentSources tests collision prevention */
func TestStatefulIndicatorNamer_UniqueNamesForDifferentSources(t *testing.T) {
	namer := NewStatefulIndicatorNamer()

	/* Same indicator and period but different source expressions */
	name1 := namer.GenerateName("rma", 14, "source1hash")
	name2 := namer.GenerateName("rma", 14, "source2hash")
	name3 := namer.GenerateName("rma", 14, "source3hash")

	/* All should be unique */
	if name1 == name2 {
		t.Errorf("different sources should produce different names: %q == %q", name1, name2)
	}
	if name2 == name3 {
		t.Errorf("different sources should produce different names: %q == %q", name2, name3)
	}
	if name1 == name3 {
		t.Errorf("different sources should produce different names: %q == %q", name1, name3)
	}
}

/* TestStatefulIndicatorNamer_DeterministicNaming tests naming consistency */
func TestStatefulIndicatorNamer_DeterministicNaming(t *testing.T) {
	namer := NewStatefulIndicatorNamer()

	/* Generate same name multiple times */
	name1 := namer.GenerateName("ema", 20, "testhash")
	name2 := namer.GenerateName("ema", 20, "testhash")
	name3 := namer.GenerateName("ema", 20, "testhash")

	/* All should be identical */
	if name1 != name2 {
		t.Errorf("naming should be deterministic: %q != %q", name1, name2)
	}
	if name2 != name3 {
		t.Errorf("naming should be deterministic: %q != %q", name2, name3)
	}
}

/* TestWindowBasedNamer_GenerateName tests window-based naming without hash */
func TestWindowBasedNamer_GenerateName(t *testing.T) {
	namer := NewWindowBasedNamer()

	tests := []struct {
		name           string
		indicatorName  string
		period         int
		sourceHash     string
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:           "highest without hash",
			indicatorName:  "highest",
			period:         10,
			sourceHash:     "shouldbeignored",
			wantContains:   []string{"_highest_", "10"},
			wantNotContain: []string{"shouldbeignored"},
		},
		{
			name:           "lowest without hash",
			indicatorName:  "lowest",
			period:         5,
			sourceHash:     "alsoignored",
			wantContains:   []string{"_lowest_", "5"},
			wantNotContain: []string{"alsoignored"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := namer.GenerateName(tt.indicatorName, tt.period, tt.sourceHash)

			/* Check required substrings */
			for _, substr := range tt.wantContains {
				if !containsSubstring(result, substr) {
					t.Errorf("GenerateName() = %q, should contain %q", result, substr)
				}
			}

			/* Check hash is not included */
			for _, substr := range tt.wantNotContain {
				if containsSubstring(result, substr) {
					t.Errorf("GenerateName() = %q, should not contain %q", result, substr)
				}
			}
		})
	}
}

/* TestWindowBasedNamer_SameNameForSamePeriod tests that source hash doesn't affect naming */
func TestWindowBasedNamer_SameNameForSamePeriod(t *testing.T) {
	namer := NewWindowBasedNamer()

	/* Same indicator and period but different source hashes */
	name1 := namer.GenerateName("highest", 20, "hash1")
	name2 := namer.GenerateName("highest", 20, "hash2")
	name3 := namer.GenerateName("highest", 20, "hash3")

	/* All should be identical (source hash ignored) */
	if name1 != name2 {
		t.Errorf("source hash should not affect window-based naming: %q != %q", name1, name2)
	}
	if name2 != name3 {
		t.Errorf("source hash should not affect window-based naming: %q != %q", name2, name3)
	}
}

/* TestNamingStrategy_Interface tests both implementations satisfy interface */
func TestNamingStrategy_Interface(t *testing.T) {
	/* Verify both types implement Strategy interface */
	var _ Strategy = NewStatefulIndicatorNamer()
	var _ Strategy = NewWindowBasedNamer()
}

/* TestNamingStrategy_EdgeCases tests edge case handling */
func TestNamingStrategy_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		namer   Strategy
		indName string
		period  int
		hash    string
	}{
		{
			name:    "stateful with zero period",
			namer:   NewStatefulIndicatorNamer(),
			indName: "rma",
			period:  0,
			hash:    "testhash",
		},
		{
			name:    "stateful with negative period",
			namer:   NewStatefulIndicatorNamer(),
			indName: "ema",
			period:  -1,
			hash:    "testhash",
		},
		{
			name:    "stateful with empty indicator name",
			namer:   NewStatefulIndicatorNamer(),
			indName: "",
			period:  14,
			hash:    "testhash",
		},
		{
			name:    "window with very large period",
			namer:   NewWindowBasedNamer(),
			indName: "highest",
			period:  99999,
			hash:    "ignored",
		},
		{
			name:    "stateful with special characters in hash",
			namer:   NewStatefulIndicatorNamer(),
			indName: "sma",
			period:  20,
			hash:    "a!@#$%^&",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			/* Should not panic */
			result := tt.namer.GenerateName(tt.indName, tt.period, tt.hash)

			/* Should produce non-empty result */
			if result == "" {
				t.Error("GenerateName() should not return empty string for edge case")
			}

			/* Should be deterministic even for edge cases */
			result2 := tt.namer.GenerateName(tt.indName, tt.period, tt.hash)
			if result != result2 {
				t.Errorf("edge case naming unstable: %q != %q", result, result2)
			}
		})
	}
}

/* TestNamingStrategy_IntegrationWithSourceIdentity tests naming with real source identifiers */
func TestNamingStrategy_IntegrationWithSourceIdentity(t *testing.T) {
	factory := source_identity.NewIdentifierFactory()
	namer := NewStatefulIndicatorNamer()

	/* Create source identifiers from different expressions */
	id1 := factory.CreateFromExpression(nil) // Simple case
	id2 := factory.CreateFromExpression(nil) // Should be same

	/* Names with same source should be identical */
	name1 := namer.GenerateName("rma", 14, id1.Hash())
	name2 := namer.GenerateName("rma", 14, id2.Hash())

	if name1 != name2 {
		t.Errorf("identical source identifiers should produce same name: %q != %q", name1, name2)
	}
}

/* TestNamingStrategy_PeriodVariations tests naming across period range */
func TestNamingStrategy_PeriodVariations(t *testing.T) {
	namer := NewStatefulIndicatorNamer()
	hash := "constanthash"

	periods := []int{1, 2, 5, 10, 14, 20, 50, 100, 200, 500}
	names := make(map[string]bool)

	for _, period := range periods {
		name := namer.GenerateName("rma", period, hash)

		/* Each period should produce unique name */
		if names[name] {
			t.Errorf("period %d produced duplicate name: %q", period, name)
		}
		names[name] = true

		/* For specific test periods, verify they're present */
		if period == 14 && !containsSubstring(name, "14") {
			t.Errorf("name %q should contain period 14", name)
		}
	}
}

/* Helper function to check substring presence */
func containsSubstring(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
