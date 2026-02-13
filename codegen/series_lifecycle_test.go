package codegen

import (
	"strings"
	"testing"
)

type stubLifecycle struct {
	hasUsage      bool
	needsTimezone bool
	tag           string
	seriesNames   []string
}

func (s *stubLifecycle) HasUsage() bool      { return s.hasUsage }
func (s *stubLifecycle) NeedsTimezone() bool { return s.needsTimezone }
func (s *stubLifecycle) GenerateDeclarations(indent string) string {
	return indent + s.tag + "_decl\n"
}
func (s *stubLifecycle) GenerateInitializations(indent string) string {
	return indent + s.tag + "_init\n"
}
func (s *stubLifecycle) GenerateBarPopulation(indent, iterVar string) string {
	return indent + s.tag + "_bar[" + iterVar + "]\n"
}
func (s *stubLifecycle) GenerateAdvancement(indent, iterVar string) string {
	return indent + s.tag + "_adv[" + iterVar + "]\n"
}
func (s *stubLifecycle) GenerateRegistrations(indent string) string {
	return indent + s.tag + "_reg\n"
}
func (s *stubLifecycle) GenerateSuppressUnused(indent string) string {
	return indent + s.tag + "_sup\n"
}
func (s *stubLifecycle) GenerateSymbolTableRegistrations(st SymbolTable) {
	if st == nil {
		return
	}
	for _, name := range s.seriesNames {
		st.Register(name, VariableTypeSeries)
	}
}

/* Zero overhead for composites with no children */
func TestCompositeSeriesLifecycle(t *testing.T) {
	active := &stubLifecycle{hasUsage: true, needsTimezone: true, tag: "A"}
	inactive := &stubLifecycle{hasUsage: false, needsTimezone: false, tag: "B"}

	tests := []struct {
		name     string
		children []SeriesLifecycle
		hasUsage bool
	}{
		{"no children", nil, false},
		{"single active child", []SeriesLifecycle{active}, true},
		{"single inactive child", []SeriesLifecycle{inactive}, false},
		{"mixed children", []SeriesLifecycle{inactive, active}, true},
		{"all inactive", []SeriesLifecycle{inactive, inactive}, false},
		{"all active", []SeriesLifecycle{active, active}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompositeSeriesLifecycle(tt.children...)

			if c.HasUsage() != tt.hasUsage {
				t.Errorf("HasUsage() = %v, want %v", c.HasUsage(), tt.hasUsage)
			}

			methods := map[string]string{
				"declarations":    c.GenerateDeclarations("\t"),
				"initializations": c.GenerateInitializations("\t"),
				"bar population":  c.GenerateBarPopulation("\t", "i"),
				"advancement":     c.GenerateAdvancement("\t", "i"),
				"registrations":   c.GenerateRegistrations("\t"),
				"suppress":        c.GenerateSuppressUnused("\t"),
			}

			for method, output := range methods {
				if len(tt.children) > 0 && output == "" {
					t.Errorf("%s should produce output when children exist", method)
				}
				if len(tt.children) == 0 && output != "" {
					t.Errorf("%s should be empty with no children, got: %q", method, output)
				}
			}
		})
	}
}

/* HasUsage and NeedsTimezone OR across all children */
func TestCompositeSeriesLifecycle_BooleanAggregation(t *testing.T) {
	tests := []struct {
		name         string
		usages       []bool
		timezones    []bool
		wantUsage    bool
		wantTimezone bool
	}{
		{"both false", []bool{false, false}, []bool{false, false}, false, false},
		{"first true", []bool{true, false}, []bool{false, false}, true, false},
		{"second true", []bool{false, true}, []bool{false, true}, true, true},
		{"both true", []bool{true, true}, []bool{true, true}, true, true},
		{"usage without timezone", []bool{true, true}, []bool{false, false}, true, false},
		{"timezone without usage", []bool{false, false}, []bool{true, true}, false, true},
		{"cross: usage-A timezone-B", []bool{true, false}, []bool{false, true}, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			children := make([]SeriesLifecycle, len(tt.usages))
			for i := range tt.usages {
				children[i] = &stubLifecycle{
					hasUsage:      tt.usages[i],
					needsTimezone: tt.timezones[i],
					tag:           string(rune('A' + i)),
				}
			}
			c := NewCompositeSeriesLifecycle(children...)

			if c.HasUsage() != tt.wantUsage {
				t.Errorf("HasUsage() = %v, want %v", c.HasUsage(), tt.wantUsage)
			}
			if c.NeedsTimezone() != tt.wantTimezone {
				t.Errorf("NeedsTimezone() = %v, want %v", c.NeedsTimezone(), tt.wantTimezone)
			}
		})
	}
}

/* Children output concatenated in insertion order */
func TestCompositeSeriesLifecycle_ChildOrdering(t *testing.T) {
	a := &stubLifecycle{tag: "A"}
	b := &stubLifecycle{tag: "B"}
	c := &stubLifecycle{tag: "C"}
	composite := NewCompositeSeriesLifecycle(a, b, c)

	generators := map[string]string{
		"declarations":    composite.GenerateDeclarations("\t"),
		"initializations": composite.GenerateInitializations("\t"),
		"bar population":  composite.GenerateBarPopulation("\t", "i"),
		"advancement":     composite.GenerateAdvancement("\t", "i"),
		"registrations":   composite.GenerateRegistrations("\t"),
		"suppress":        composite.GenerateSuppressUnused("\t"),
	}

	for method, code := range generators {
		posA := strings.Index(code, "A_")
		posB := strings.Index(code, "B_")
		posC := strings.Index(code, "C_")

		if posA >= posB || posB >= posC {
			t.Errorf("%s: children not in insertion order: A@%d B@%d C@%d in:\n%s",
				method, posA, posB, posC, code)
		}
	}
}

/* iterVar propagated to bar population and advancement */
func TestCompositeSeriesLifecycle_IterVarPropagation(t *testing.T) {
	child := &stubLifecycle{tag: "X"}
	c := NewCompositeSeriesLifecycle(child)

	for _, iterVar := range []string{"i", "idx", "barIdx"} {
		bar := c.GenerateBarPopulation("\t", iterVar)
		if !strings.Contains(bar, "X_bar["+iterVar+"]") {
			t.Errorf("GenerateBarPopulation should propagate iterVar %q, got:\n%s", iterVar, bar)
		}
		adv := c.GenerateAdvancement("\t", iterVar)
		if !strings.Contains(adv, "X_adv["+iterVar+"]") {
			t.Errorf("GenerateAdvancement should propagate iterVar %q, got:\n%s", iterVar, adv)
		}
	}
}

/* Indent propagated to all Generate* methods */
func TestCompositeSeriesLifecycle_IndentPropagation(t *testing.T) {
	child := &stubLifecycle{tag: "X"}
	c := NewCompositeSeriesLifecycle(child)

	for _, indent := range []string{"\t", "\t\t", "    "} {
		methods := map[string]string{
			"declarations":    c.GenerateDeclarations(indent),
			"initializations": c.GenerateInitializations(indent),
			"bar population":  c.GenerateBarPopulation(indent, "i"),
			"advancement":     c.GenerateAdvancement(indent, "i"),
			"registrations":   c.GenerateRegistrations(indent),
			"suppress":        c.GenerateSuppressUnused(indent),
		}
		for method, code := range methods {
			if !strings.HasPrefix(code, indent) {
				t.Errorf("%s with indent %q should start with indent, got:\n%s", method, indent, code)
			}
		}
	}
}

func TestCompositeSeriesLifecycle_TimezoneSetup(t *testing.T) {
	tests := []struct {
		name      string
		timezones []bool
		wantEmpty bool
	}{
		{"no children", nil, true},
		{"none need timezone", []bool{false, false}, true},
		{"first needs timezone", []bool{true, false}, false},
		{"second needs timezone", []bool{false, true}, false},
		{"all need timezone", []bool{true, true}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			children := make([]SeriesLifecycle, len(tt.timezones))
			for i, tz := range tt.timezones {
				children[i] = &stubLifecycle{needsTimezone: tz, tag: string(rune('A' + i))}
			}
			c := NewCompositeSeriesLifecycle(children...)
			setup := c.GenerateTimezoneSetup("\t")

			if tt.wantEmpty && setup != "" {
				t.Errorf("expected empty, got: %q", setup)
			}
			if !tt.wantEmpty {
				assertContains(t, setup, "time.LoadLocation")
				assertContains(t, setup, "ctx.Timezone")
				if !strings.HasPrefix(setup, "\t") {
					t.Error("timezone setup should respect indent")
				}
			}
		})
	}
}

func TestCompositeSeriesLifecycle_NilReceiver(t *testing.T) {
	var c *CompositeSeriesLifecycle

	if c.HasUsage() {
		t.Error("nil receiver should report no usage")
	}
	if c.NeedsTimezone() {
		t.Error("nil receiver should not need timezone")
	}

	nilSafeMethods := map[string]string{
		"declarations":    c.GenerateDeclarations("\t"),
		"initializations": c.GenerateInitializations("\t"),
		"bar population":  c.GenerateBarPopulation("\t", "i"),
		"advancement":     c.GenerateAdvancement("\t", "i"),
		"registrations":   c.GenerateRegistrations("\t"),
		"suppress":        c.GenerateSuppressUnused("\t"),
		"timezone setup":  c.GenerateTimezoneSetup("\t"),
	}
	for name, output := range nilSafeMethods {
		if output != "" {
			t.Errorf("nil receiver %s should be empty, got: %q", name, output)
		}
	}

	c.GenerateSymbolTableRegistrations(NewSymbolTable())
	c.GenerateSymbolTableRegistrations(nil)
}

func TestCompositeSeriesLifecycle_SymbolTableRegistration(t *testing.T) {
	a := &stubLifecycle{seriesNames: []string{"alpha", "beta"}}
	b := &stubLifecycle{seriesNames: []string{"gamma"}}
	c := NewCompositeSeriesLifecycle(a, b)

	st := NewSymbolTable()
	c.GenerateSymbolTableRegistrations(st)

	for _, name := range []string{"alpha", "beta", "gamma"} {
		if !st.IsSeries(name) {
			t.Errorf("%s should be registered as series in symbol table", name)
		}
	}
}

func TestCompositeSeriesLifecycle_SymbolTableRegistration_NilReceiver(t *testing.T) {
	var c *CompositeSeriesLifecycle
	st := NewSymbolTable()
	c.GenerateSymbolTableRegistrations(st)

	if st.IsSeries("anything") {
		t.Error("nil composite should not register anything")
	}
}

func TestCompositeSeriesLifecycle_SymbolTableRegistration_NilTable(t *testing.T) {
	child := &stubLifecycle{seriesNames: []string{"x"}}
	c := NewCompositeSeriesLifecycle(child)
	c.GenerateSymbolTableRegistrations(nil)
}
