package codegen

import (
	"strings"
	"testing"
)

/* TestDirectionalSplitGenerator_ContextAwareSplitting validates directional decomposition across execution contexts */
func TestDirectionalSplitGenerator_ContextAwareSplitting(t *testing.T) {
	tests := []struct {
		name                string
		context             StatefulIndicatorContext
		gainsName           string
		lossesName          string
		changeVar           string
		expectedGainsUpdate string
		expectedLossUpdate  string
	}{
		{
			name:                "TopLevel context",
			context:             NewTopLevelIndicatorContext(),
			gainsName:           "_rsi_gains",
			lossesName:          "_rsi_losses",
			changeVar:           "priceChange",
			expectedGainsUpdate: "_rsi_gainsSeries.Set(_rsi_",
			expectedLossUpdate:  "_rsi_lossesSeries.Set(_rsi_",
		},
		{
			name:                "Arrow context",
			context:             NewArrowFunctionIndicatorContext(),
			gainsName:           "_upMove",
			lossesName:          "_downMove",
			changeVar:           "delta",
			expectedGainsUpdate: "arrowCtx.GetOrCreateSeries(\"_upMove\").Set(",
			expectedLossUpdate:  "arrowCtx.GetOrCreateSeries(\"_downMove\").Set(",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewDirectionalSplitGenerator(tt.gainsName, tt.lossesName, tt.context)
			code := gen.GenerateSplitCode(tt.changeVar)

			if !strings.Contains(code, tt.expectedGainsUpdate) {
				t.Errorf("Missing gains update %q in:\n%s", tt.expectedGainsUpdate, code)
			}

			if !strings.Contains(code, tt.expectedLossUpdate) {
				t.Errorf("Missing losses update %q in:\n%s", tt.expectedLossUpdate, code)
			}
		})
	}
}

/* TestDirectionalSplitGenerator_EdgeCases validates boundary conditions */
func TestDirectionalSplitGenerator_EdgeCases(t *testing.T) {
	context := NewTopLevelIndicatorContext()
	gen := NewDirectionalSplitGenerator("_g", "_l", context)

	tests := []struct {
		name              string
		changeVar         string
		mustContainChecks []string
		description       string
	}{
		{
			name:      "NaN handling",
			changeVar: "c",
			mustContainChecks: []string{
				"math.IsNaN(c)",
				"gain = 0.0",
				"loss = 0.0",
			},
			description: "NaN change must produce zero gain and loss",
		},
		{
			name:      "Positive change routing",
			changeVar: "c",
			mustContainChecks: []string{
				"else if c > 0",
				"gain = c",
			},
			description: "Positive change becomes gain, loss is zero",
		},
		{
			name:      "Negative change routing",
			changeVar: "c",
			mustContainChecks: []string{
				"} else {",
				"loss = -c",
			},
			description: "Negative change becomes positive loss via negation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := gen.GenerateSplitCode(tt.changeVar)

			for _, check := range tt.mustContainChecks {
				if !strings.Contains(code, check) {
					t.Errorf("%s: missing %q in generated code:\n%s", tt.description, check, code)
				}
			}
		})
	}
}

/* TestDirectionalSplitGenerator_AlgorithmicInvariants validates behavioral properties */
func TestDirectionalSplitGenerator_AlgorithmicInvariants(t *testing.T) {
	context := NewTopLevelIndicatorContext()
	gen := NewDirectionalSplitGenerator("_gain", "_loss", context)
	code := gen.GenerateSplitCode("change")

	t.Run("mutual exclusivity", func(t *testing.T) {
		/* Gain and loss assignments must be in mutually exclusive branches */
		hasIfElseChain := strings.Contains(code, "if math.IsNaN") &&
			strings.Contains(code, "} else if") &&
			strings.Contains(code, "} else {")

		if !hasIfElseChain {
			t.Error("Directional split must use mutually exclusive if-else-if-else chain")
		}
	})

	t.Run("variable declarations", func(t *testing.T) {
		/* gain and loss must be declared with unique prefix */
		t.Logf("Generated code:\n%s", code)
		hasGainLossDecl := strings.Contains(code, "var _gain, _loss float64") ||
			strings.Contains(code, "var gain, loss float64") ||
			strings.Contains(code, "var __gain, __loss float64")
		if !hasGainLossDecl {
			t.Error("Missing gain/loss variable declarations")
		}
	})

	t.Run("series updates", func(t *testing.T) {
		/* Both gain and loss must be stored to series */
		gainsStored := strings.Contains(code, ".Set(_gain)") ||
			strings.Contains(code, ".Set(gain)")
		lossesStored := strings.Contains(code, ".Set(_loss)") ||
			strings.Contains(code, ".Set(loss)")

		if !gainsStored {
			t.Error("Gains not stored to series")
		}
		if !lossesStored {
			t.Error("Losses not stored to series")
		}
	})

	t.Run("sign convention", func(t *testing.T) {
		/* Loss must be negation of negative change (stored as positive) */
		if !strings.Contains(code, "loss = -change") && !strings.Contains(code, "loss = -") {
			t.Error("Loss must use negation to convert negative change to positive value")
		}
	})
}

/* TestDirectionalSplitGenerator_DebugInstrumentation validates code readability through structure */
func TestDirectionalSplitGenerator_DebugInstrumentation(t *testing.T) {
	context := NewTopLevelIndicatorContext()
	gen := NewDirectionalSplitGenerator("_g", "_l", context)
	code := gen.GenerateSplitCode("c")

	/* Code self-documents through control flow structure */
	if !strings.Contains(code, "math.IsNaN(c)") {
		t.Error("Missing self-explanatory NaN check")
	}
	if !strings.Contains(code, "c > 0") {
		t.Error("Missing self-explanatory directional check")
	}
}

/* TestDirectionalSplitGenerator_ZeroChangeHandling validates exact zero behavior */
func TestDirectionalSplitGenerator_ZeroChangeHandling(t *testing.T) {
	context := NewTopLevelIndicatorContext()
	gen := NewDirectionalSplitGenerator("_g", "_l", context)
	code := gen.GenerateSplitCode("change")

	/* Zero change is NOT > 0, so falls into else branch (loss = -0 = 0) */
	t.Run("zero routes to else branch", func(t *testing.T) {
		hasPositiveCheck := strings.Contains(code, "change > 0")
		hasElseBranch := strings.Contains(code, "} else {")

		if !hasPositiveCheck || !hasElseBranch {
			t.Error("Zero change must be handled by else branch (not positive)")
		}
	})
}
