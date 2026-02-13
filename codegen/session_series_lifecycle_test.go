package codegen

import (
	"strings"
	"testing"
)

func TestSessionSeriesLifecycle(t *testing.T) {
	tests := []struct {
		name     string
		first    bool
		last     bool
		firstReg bool
		lastReg  bool
		hasUsage bool
	}{
		{"none", false, false, false, false, false},
		{"isfirstbar only", true, false, false, false, true},
		{"islastbar only", false, true, false, false, true},
		{"isfirstbar_regular only", false, false, true, false, true},
		{"islastbar_regular only", false, false, false, true, true},
		{"all", true, true, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lifecycle := NewSessionSeriesLifecycle(tt.first, tt.last, tt.firstReg, tt.lastReg)

			if lifecycle.HasUsage() != tt.hasUsage {
				t.Errorf("HasUsage() = %v, want %v", lifecycle.HasUsage(), tt.hasUsage)
			}

			methods := map[string]string{
				"declarations":    lifecycle.GenerateDeclarations("\t"),
				"initializations": lifecycle.GenerateInitializations("\t"),
				"bar population":  lifecycle.GenerateBarPopulation("\t", "i"),
				"advancement":     lifecycle.GenerateAdvancement("\t", "i"),
				"registrations":   lifecycle.GenerateRegistrations("\t"),
				"suppress":        lifecycle.GenerateSuppressUnused("\t"),
			}

			for method, output := range methods {
				if tt.hasUsage && output == "" {
					t.Errorf("%s should produce output when builtins are used", method)
				}
				if !tt.hasUsage && output != "" {
					t.Errorf("%s should be empty when no builtins used, got: %q", method, output)
				}
			}
		})
	}
}

func TestSessionSeriesLifecycle_CodeContent(t *testing.T) {
	lifecycle := NewSessionSeriesLifecycle(true, true, true, true)

	t.Run("declarations", func(t *testing.T) {
		code := lifecycle.GenerateDeclarations("\t")
		for _, s := range []string{
			"var session_isfirstbarSeries *series.Series",
			"var session_islastbarSeries *series.Series",
			"var session_isfirstbar_regularSeries *series.Series",
			"var session_islastbar_regularSeries *series.Series",
		} {
			if !strings.Contains(code, s) {
				t.Errorf("missing %q in:\n%s", s, code)
			}
		}
	})

	t.Run("initializations", func(t *testing.T) {
		code := lifecycle.GenerateInitializations("\t")
		for _, s := range []string{
			"session_isfirstbarSeries = series.NewSeries(len(ctx.Data))",
			"session_islastbarSeries = series.NewSeries(len(ctx.Data))",
			"session_isfirstbar_regularSeries = series.NewSeries(len(ctx.Data))",
			"session_islastbar_regularSeries = series.NewSeries(len(ctx.Data))",
		} {
			if !strings.Contains(code, s) {
				t.Errorf("missing %q in:\n%s", s, code)
			}
		}
	})

	t.Run("bar population isfirstbar uses date comparison", func(t *testing.T) {
		code := lifecycle.GenerateBarPopulation("\t", "i")
		for _, s := range []string{
			"currDay",
			"currYear",
			"prevDay",
			"prevYear",
			"sessionIsFirst",
			"session_isfirstbarSeries.Set(",
			"session_isfirstbar_regularSeries.Set(",
		} {
			if !strings.Contains(code, s) {
				t.Errorf("missing %q in:\n%s", s, code)
			}
		}
	})

	t.Run("bar population islastbar uses next bar date comparison", func(t *testing.T) {
		code := lifecycle.GenerateBarPopulation("\t", "i")
		for _, s := range []string{
			"nextDay",
			"nextYear",
			"sessionIsLast",
			"session_islastbarSeries.Set(",
			"session_islastbar_regularSeries.Set(",
			"barCount-1",
		} {
			if !strings.Contains(code, s) {
				t.Errorf("missing %q in:\n%s", s, code)
			}
		}
	})

	t.Run("bar population uses timezone", func(t *testing.T) {
		code := lifecycle.GenerateBarPopulation("\t", "i")
		if !strings.Contains(code, "exchangeLoc") {
			t.Errorf("bar population should use exchangeLoc for timezone-aware dates:\n%s", code)
		}
	})

	t.Run("registrations", func(t *testing.T) {
		code := lifecycle.GenerateRegistrations("\t")
		for _, s := range []string{
			`ctx.RegisterSeries("session_isfirstbarSeries", session_isfirstbarSeries)`,
			`ctx.RegisterSeries("session_islastbarSeries", session_islastbarSeries)`,
			`ctx.RegisterSeries("session_isfirstbar_regularSeries", session_isfirstbar_regularSeries)`,
			`ctx.RegisterSeries("session_islastbar_regularSeries", session_islastbar_regularSeries)`,
		} {
			if !strings.Contains(code, s) {
				t.Errorf("missing %q in:\n%s", s, code)
			}
		}
	})

	t.Run("advancement uses iterVar", func(t *testing.T) {
		code := lifecycle.GenerateAdvancement("\t", "idx")
		if !strings.Contains(code, "idx < barCount-1") {
			t.Errorf("advancement should use custom iterVar, got:\n%s", code)
		}
		for _, s := range []string{
			"session_isfirstbarSeries.Next()",
			"session_islastbarSeries.Next()",
			"session_isfirstbar_regularSeries.Next()",
			"session_islastbar_regularSeries.Next()",
		} {
			if !strings.Contains(code, s) {
				t.Errorf("missing %q in:\n%s", s, code)
			}
		}
	})

	t.Run("suppress unused", func(t *testing.T) {
		code := lifecycle.GenerateSuppressUnused("\t")
		for _, s := range []string{
			"_ = session_isfirstbarSeries",
			"_ = session_islastbarSeries",
			"_ = session_isfirstbar_regularSeries",
			"_ = session_islastbar_regularSeries",
		} {
			if !strings.Contains(code, s) {
				t.Errorf("missing %q in:\n%s", s, code)
			}
		}
	})
}

/* Partial usage should only emit relevant series */
func TestSessionSeriesLifecycle_Isolation(t *testing.T) {
	t.Run("isfirstbar only excludes islastbar", func(t *testing.T) {
		lifecycle := NewSessionSeriesLifecycle(true, false, false, false)
		code := lifecycle.GenerateDeclarations("\t")
		if strings.Contains(code, "islastbar") {
			t.Errorf("should not contain islastbar: %s", code)
		}
		if !strings.Contains(code, "session_isfirstbarSeries") {
			t.Errorf("should contain session_isfirstbarSeries: %s", code)
		}
	})

	t.Run("islastbar only excludes isfirstbar", func(t *testing.T) {
		lifecycle := NewSessionSeriesLifecycle(false, true, false, false)
		code := lifecycle.GenerateDeclarations("\t")
		if strings.Contains(code, "isfirstbar") {
			t.Errorf("should not contain isfirstbar: %s", code)
		}
		if !strings.Contains(code, "session_islastbarSeries") {
			t.Errorf("should contain session_islastbarSeries: %s", code)
		}
	})

	t.Run("regular variants isolated", func(t *testing.T) {
		lifecycle := NewSessionSeriesLifecycle(false, false, true, true)
		code := lifecycle.GenerateDeclarations("\t")
		if !strings.Contains(code, "session_isfirstbar_regularSeries") {
			t.Errorf("should contain regular series: %s", code)
		}
		if strings.Contains(code, "var session_isfirstbarSeries") {
			t.Errorf("should not contain non-regular isfirstbar series: %s", code)
		}
	})
}

/* isfirstbar and isfirstbar_regular share sessionIsFirst computation */
func TestSessionSeriesLifecycle_SharedComputation(t *testing.T) {
	lifecycle := NewSessionSeriesLifecycle(true, false, true, false)
	code := lifecycle.GenerateBarPopulation("\t", "i")

	count := strings.Count(code, "var sessionIsFirst")
	if count != 1 {
		t.Errorf("expected 1 sessionIsFirst declaration, got %d in:\n%s", count, code)
	}

	if !strings.Contains(code, "session_isfirstbarSeries.Set(sessionIsFirst)") {
		t.Errorf("missing session_isfirstbarSeries.Set(sessionIsFirst)")
	}
	if !strings.Contains(code, "session_isfirstbar_regularSeries.Set(sessionIsFirst)") {
		t.Errorf("missing session_isfirstbar_regularSeries.Set(sessionIsFirst)")
	}
}

func TestSessionSeriesLifecycle_NeedsTimezone(t *testing.T) {
	tests := []struct {
		name          string
		first         bool
		last          bool
		needsTimezone bool
	}{
		{"none", false, false, false},
		{"any usage needs timezone", true, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lifecycle := NewSessionSeriesLifecycle(tt.first, tt.last, false, false)
			if lifecycle.NeedsTimezone() != tt.needsTimezone {
				t.Errorf("NeedsTimezone() = %v, want %v", lifecycle.NeedsTimezone(), tt.needsTimezone)
			}
		})
	}
}

func TestSessionSeriesLifecycle_NilReceiver(t *testing.T) {
	var lifecycle *SessionSeriesLifecycle

	if lifecycle.HasUsage() {
		t.Error("nil receiver should report no usage")
	}
	if lifecycle.NeedsTimezone() {
		t.Error("nil receiver should not need timezone")
	}

	nilSafeMethods := map[string]string{
		"declarations":    lifecycle.GenerateDeclarations("\t"),
		"initializations": lifecycle.GenerateInitializations("\t"),
		"bar population":  lifecycle.GenerateBarPopulation("\t", "i"),
		"advancement":     lifecycle.GenerateAdvancement("\t", "i"),
		"registrations":   lifecycle.GenerateRegistrations("\t"),
		"suppress":        lifecycle.GenerateSuppressUnused("\t"),
	}
	for name, output := range nilSafeMethods {
		if output != "" {
			t.Errorf("nil receiver %s should be empty, got: %q", name, output)
		}
	}
}

func TestSessionSeriesLifecycle_SymbolTableRegistration(t *testing.T) {
	lifecycle := NewSessionSeriesLifecycle(true, true, true, true)
	st := NewSymbolTable()
	lifecycle.GenerateSymbolTableRegistrations(st)

	for _, name := range []string{
		"session.isfirstbar",
		"session.islastbar",
		"session.isfirstbar_regular",
		"session.islastbar_regular",
	} {
		if !st.IsSeries(name) {
			t.Errorf("%s should be registered as series in symbol table", name)
		}
	}
}

func TestSessionSeriesLifecycle_SymbolTableRegistration_NilReceiver(t *testing.T) {
	var lifecycle *SessionSeriesLifecycle
	st := NewSymbolTable()

	lifecycle.GenerateSymbolTableRegistrations(st)

	for _, name := range []string{
		"session.isfirstbar",
		"session.islastbar",
		"session.isfirstbar_regular",
		"session.islastbar_regular",
	} {
		if st.IsSeries(name) {
			t.Errorf("nil lifecycle should not register %s", name)
		}
	}
}

func TestSessionSeriesLifecycle_SymbolTableRegistration_NilTable(t *testing.T) {
	lifecycle := NewSessionSeriesLifecycle(true, true, true, true)
	lifecycle.GenerateSymbolTableRegistrations(nil)
}

/* Bar population values: 1.0 for true, 0.0 for false (float64 FSB) */
func TestSessionSeriesLifecycle_BoolAsFloat(t *testing.T) {
	lifecycle := NewSessionSeriesLifecycle(true, true, false, false)
	code := lifecycle.GenerateBarPopulation("\t", "i")

	if !strings.Contains(code, "1.0") {
		t.Errorf("should use 1.0 for true in float64 series:\n%s", code)
	}
}
