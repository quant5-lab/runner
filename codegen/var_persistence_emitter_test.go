package codegen

import (
	"strings"
	"testing"
)

/* TestEmitSeriesGuard validates bar-0 guard structure with init code and carry-forward. */
func TestEmitSeriesGuard(t *testing.T) {
	tests := []struct {
		name         string
		indent       string
		varName      string
		initCode     string
		mustContain  []string
		mustNotAllow []string
	}{
		{
			name:     "standard series guard",
			indent:   "\t",
			varName:  "cumulative",
			initCode: "\t\tcumulativeSeries.Set(0)\n",
			mustContain: []string{
				"if i == 0 {",
				"cumulativeSeries.Set(0)",
				"} else {",
				"cumulativeSeries.Set(cumulativeSeries.Get(1))",
			},
		},
		{
			name:     "double indent propagation",
			indent:   "\t\t",
			varName:  "x",
			initCode: "\t\t\txSeries.Set(0)\n",
			mustContain: []string{
				"\t\tif i == 0 {",
				"xSeries.Set(xSeries.Get(1))",
			},
		},
		{
			name:     "no indent",
			indent:   "",
			varName:  "val",
			initCode: "\tvalSeries.Set(42)\n",
			mustContain: []string{
				"if i == 0 {",
				"valSeries.Set(valSeries.Get(1))",
			},
		},
	}

	emitter := NewVarPersistenceEmitter(NewRuntimeSafetyGuard())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := emitter.EmitSeriesGuard(tt.indent, tt.varName, tt.initCode)
			for _, pattern := range tt.mustContain {
				if !strings.Contains(result, pattern) {
					t.Errorf("missing pattern %q in:\n%s", pattern, result)
				}
			}
		})
	}
}

/* TestEmitStringGuard validates bar-0 guard without carry-forward for string types. */
func TestEmitStringGuard(t *testing.T) {
	emitter := NewVarPersistenceEmitter(NewRuntimeSafetyGuard())
	result := emitter.EmitStringGuard("\t", "\t\tsym = \"BTCUSD\"\n")

	v := &testStringVerifier{t: t, code: result}
	v.mustContain("if i == 0 {", `sym = "BTCUSD"`)
	v.mustNotContain("} else {")
}

/* TestEmitTupleGuard validates per-element carry-forward for tuple destructured var declarations. */
func TestEmitTupleGuard(t *testing.T) {
	tests := []struct {
		name     string
		elements []string
	}{
		{name: "two elements", elements: []string{"upper", "lower"}},
		{name: "three elements", elements: []string{"high", "low", "mid"}},
		{name: "single element", elements: []string{"val"}},
	}

	emitter := NewVarPersistenceEmitter(NewRuntimeSafetyGuard())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := emitter.EmitTupleGuard("\t", tt.elements, "\t\tinitCode\n")
			if !strings.Contains(result, "if i == 0 {") {
				t.Error("missing bar-0 guard")
			}
			for _, name := range tt.elements {
				expected := name + "Series.Set(" + name + "Series.Get(1))"
				if !strings.Contains(result, expected) {
					t.Errorf("missing carry-forward for %q", name)
				}
			}
		})
	}
}

/* TestSeriesCarryForward_Format validates the carry-forward expression format. */
func TestSeriesCarryForward_Format(t *testing.T) {
	tests := []struct {
		varName  string
		expected string
	}{
		{"price", "priceSeries.Set(priceSeries.Get(1))\n"},
		{"x", "xSeries.Set(xSeries.Get(1))\n"},
		{"cumVol", "cumVolSeries.Set(cumVolSeries.Get(1))\n"},
	}

	for _, tt := range tests {
		t.Run(tt.varName, func(t *testing.T) {
			result := seriesCarryForward(tt.varName)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

type testStringVerifier struct {
	t    *testing.T
	code string
}

func (v *testStringVerifier) mustContain(patterns ...string) {
	v.t.Helper()
	for _, p := range patterns {
		if !strings.Contains(v.code, p) {
			v.t.Errorf("missing pattern %q in:\n%s", p, v.code)
		}
	}
}

func (v *testStringVerifier) mustNotContain(patterns ...string) {
	v.t.Helper()
	for _, p := range patterns {
		if strings.Contains(v.code, p) {
			v.t.Errorf("unexpected pattern %q in:\n%s", p, v.code)
		}
	}
}
