package codegen

import "testing"

func TestBuiltinNamespaceResolver_Resolve(t *testing.T) {
	resolver := NewBuiltinNamespaceResolver()

	tests := []struct {
		name         string
		namespace    string
		prop         string
		expectedCode string
		expectedBool bool
		expectFound  bool
	}{
		/* barstate boolean properties */
		{"barstate.isfirst", "barstate", "isfirst", "(ctx.BarIndex == 0)", true, true},
		{"barstate.islast", "barstate", "islast", "(ctx.BarIndex == len(ctx.Data)-1)", true, true},
		{"barstate.ishistory", "barstate", "ishistory", "true", true, true},
		{"barstate.isrealtime", "barstate", "isrealtime", "false", true, true},
		{"barstate.isnew", "barstate", "isnew", "true", true, true},
		{"barstate.isconfirmed", "barstate", "isconfirmed", "true", true, true},

		/* timeframe boolean properties */
		{"timeframe.ismonthly", "timeframe", "ismonthly", "ctx.IsMonthly", true, true},
		{"timeframe.isdaily", "timeframe", "isdaily", "ctx.IsDaily", true, true},
		{"timeframe.isweekly", "timeframe", "isweekly", "ctx.IsWeekly", true, true},
		{"timeframe.isintraday", "timeframe", "isintraday", "ctx.IsIntraday", true, true},

		/* timeframe non-boolean property */
		{"timeframe.period", "timeframe", "period", "ctx.Timeframe", false, true},

		/* syminfo string properties */
		{"syminfo.tickerid", "syminfo", "tickerid", "syminfo_tickerid", false, true},
		{"syminfo.ticker", "syminfo", "ticker", "syminfo_tickerid", false, true},
		{"syminfo.timezone", "syminfo", "timezone", "ctx.Timezone", false, true},

		/* unknown namespace */
		{"unknown.prop", "unknown", "prop", "", false, false},
		{"strategy.entry", "strategy", "entry", "", false, false},
		{"ta.sma", "ta", "sma", "", false, false},

		/* unknown property within valid namespace */
		{"barstate.unknown", "barstate", "unknown_prop", "", false, false},
		{"timeframe.unknown", "timeframe", "unknown_prop", "", false, false},
		{"syminfo.unknown", "syminfo", "unknown_prop", "", false, false},

		/* case sensitivity */
		{"Barstate uppercase", "Barstate", "isfirst", "", false, false},
		{"BARSTATE uppercase", "BARSTATE", "isfirst", "", false, false},
		{"barstate.ISFIRST uppercase", "barstate", "ISFIRST", "", false, false},
		{"TIMEFRAME uppercase", "TIMEFRAME", "period", "", false, false},
		{"SYMINFO uppercase", "SYMINFO", "tickerid", "", false, false},

		/* empty string */
		{"empty namespace", "", "prop", "", false, false},
		{"empty property", "barstate", "", "", false, false},
		{"both empty", "", "", "", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, found := resolver.Resolve(tt.namespace, tt.prop)
			if found != tt.expectFound {
				t.Fatalf("Resolve(%s, %s) found = %v, want %v", tt.namespace, tt.prop, found, tt.expectFound)
			}
			if !found {
				return
			}
			if res.Code != tt.expectedCode {
				t.Errorf("Resolve(%s, %s) code = %q, want %q", tt.namespace, tt.prop, res.Code, tt.expectedCode)
			}
			if res.IsBool != tt.expectedBool {
				t.Errorf("Resolve(%s, %s) isBool = %v, want %v", tt.namespace, tt.prop, res.IsBool, tt.expectedBool)
			}
		})
	}
}

func TestBuiltinNamespaceResolver_IsNamespace(t *testing.T) {
	resolver := NewBuiltinNamespaceResolver()

	tests := []struct {
		name     string
		expected bool
	}{
		{"barstate", true},
		{"timeframe", true},
		{"syminfo", true},
		{"unknown", false},
		{"close", false},
		{"strategy", false},
		{"ta", false},
		{"Barstate", false},
		{"TIMEFRAME", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.IsNamespace(tt.name)
			if result != tt.expected {
				t.Errorf("IsNamespace(%s) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestBuiltinNamespaceResolver_PropertyExhaustiveness(t *testing.T) {
	resolver := NewBuiltinNamespaceResolver()

	expectedCounts := map[string]int{
		"barstate":  6,
		"timeframe": 5,
		"syminfo":   3,
	}

	namespacePropSets := map[string][]string{
		"barstate":  {"isfirst", "islast", "ishistory", "isrealtime", "isnew", "isconfirmed"},
		"timeframe": {"ismonthly", "isdaily", "isweekly", "isintraday", "period"},
		"syminfo":   {"tickerid", "ticker", "timezone"},
	}

	for ns, props := range namespacePropSets {
		t.Run(ns, func(t *testing.T) {
			resolvedCount := 0
			for _, prop := range props {
				if _, found := resolver.Resolve(ns, prop); found {
					resolvedCount++
				}
			}
			if resolvedCount != expectedCounts[ns] {
				t.Errorf("%s resolved %d properties, want %d", ns, resolvedCount, expectedCounts[ns])
			}
		})
	}
}

func TestBuiltinNamespaceResolver_BoolTypeConsistency(t *testing.T) {
	resolver := NewBuiltinNamespaceResolver()

	boolProperties := []struct {
		ns   string
		prop string
	}{
		{"barstate", "isfirst"},
		{"barstate", "islast"},
		{"barstate", "ishistory"},
		{"barstate", "isrealtime"},
		{"barstate", "isnew"},
		{"barstate", "isconfirmed"},
		{"timeframe", "ismonthly"},
		{"timeframe", "isdaily"},
		{"timeframe", "isweekly"},
		{"timeframe", "isintraday"},
	}

	for _, bp := range boolProperties {
		t.Run(bp.ns+"."+bp.prop, func(t *testing.T) {
			res, found := resolver.Resolve(bp.ns, bp.prop)
			if !found {
				t.Fatalf("%s.%s not found", bp.ns, bp.prop)
			}
			if !res.IsBool {
				t.Errorf("%s.%s should be boolean, got IsBool=false", bp.ns, bp.prop)
			}
		})
	}

	nonBoolProperties := []struct {
		ns   string
		prop string
	}{
		{"timeframe", "period"},
		{"syminfo", "tickerid"},
		{"syminfo", "ticker"},
		{"syminfo", "timezone"},
	}

	for _, nbp := range nonBoolProperties {
		t.Run(nbp.ns+"."+nbp.prop+" non-bool", func(t *testing.T) {
			res, found := resolver.Resolve(nbp.ns, nbp.prop)
			if !found {
				t.Fatalf("%s.%s not found", nbp.ns, nbp.prop)
			}
			if res.IsBool {
				t.Errorf("%s.%s should not be boolean, got IsBool=true", nbp.ns, nbp.prop)
			}
		})
	}
}
