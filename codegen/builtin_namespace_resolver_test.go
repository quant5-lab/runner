package codegen

import "testing"

func TestBuiltinNamespaceResolver_Resolve(t *testing.T) {
	resolver := NewBuiltinNamespaceResolver()

	tests := []struct {
		name         string
		namespace    string
		prop         string
		expectedCode string
		expectedType GoValueType
		expectFound  bool
	}{
		/* barstate boolean properties */
		{"barstate.isfirst", "barstate", "isfirst", "(ctx.BarIndex == 0)", GoBool, true},
		{"barstate.islast", "barstate", "islast", "(ctx.BarIndex == len(ctx.Data)-1)", GoBool, true},
		{"barstate.ishistory", "barstate", "ishistory", "true", GoBool, true},
		{"barstate.isrealtime", "barstate", "isrealtime", "false", GoBool, true},
		{"barstate.isnew", "barstate", "isnew", "true", GoBool, true},
		{"barstate.isconfirmed", "barstate", "isconfirmed", "true", GoBool, true},
		{"barstate.islastconfirmedhistory", "barstate", "islastconfirmedhistory", "(ctx.BarIndex == len(ctx.Data)-1)", GoBool, true},

		/* timeframe boolean properties */
		{"timeframe.ismonthly", "timeframe", "ismonthly", "ctx.IsMonthly", GoBool, true},
		{"timeframe.isdaily", "timeframe", "isdaily", "ctx.IsDaily", GoBool, true},
		{"timeframe.isweekly", "timeframe", "isweekly", "ctx.IsWeekly", GoBool, true},
		{"timeframe.isintraday", "timeframe", "isintraday", "ctx.IsIntraday", GoBool, true},
		{"timeframe.isdwm", "timeframe", "isdwm", "(ctx.IsDaily || ctx.IsWeekly || ctx.IsMonthly)", GoBool, true},
		{"timeframe.isminutes", "timeframe", "isminutes", "ctx.IsIntraday", GoBool, true},
		{"timeframe.isseconds", "timeframe", "isseconds", "false", GoBool, true},
		{"timeframe.isticks", "timeframe", "isticks", "false", GoBool, true},

		/* timeframe non-boolean properties */
		{"timeframe.period", "timeframe", "period", "ctx.Timeframe", GoString, true},
		{"timeframe.multiplier", "timeframe", "multiplier", "float64(context.TimeframeMultiplier(ctx.Timeframe))", GoFloat64, true},

		/* syminfo string properties */
		{"syminfo.tickerid", "syminfo", "tickerid", "syminfo_tickerid", GoString, true},
		{"syminfo.ticker", "syminfo", "ticker", "syminfo_tickerid", GoString, true},
		{"syminfo.timezone", "syminfo", "timezone", "ctx.Timezone", GoString, true},
		{"syminfo.type", "syminfo", "type", `"stock"`, GoString, true},
		{"syminfo.prefix", "syminfo", "prefix", `""`, GoString, true},
		{"syminfo.session", "syminfo", "session", `"regular"`, GoString, true},
		{"syminfo.currency", "syminfo", "currency", `"USD"`, GoString, true},
		{"syminfo.basecurrency", "syminfo", "basecurrency", `""`, GoString, true},
		{"syminfo.description", "syminfo", "description", "syminfo_tickerid", GoString, true},
		{"syminfo.pointvalue", "syminfo", "pointvalue", "1.0", GoFloat64, true},
		{"syminfo.mintick", "syminfo", "mintick", "0.01", GoFloat64, true},
		{"syminfo.volumetype", "syminfo", "volumetype", `"base"`, GoString, true},

		/* dayofweek constants */
		{"dayofweek.sunday", "dayofweek", "sunday", "1.0", GoFloat64, true},
		{"dayofweek.monday", "dayofweek", "monday", "2.0", GoFloat64, true},
		{"dayofweek.tuesday", "dayofweek", "tuesday", "3.0", GoFloat64, true},
		{"dayofweek.wednesday", "dayofweek", "wednesday", "4.0", GoFloat64, true},
		{"dayofweek.thursday", "dayofweek", "thursday", "5.0", GoFloat64, true},
		{"dayofweek.friday", "dayofweek", "friday", "6.0", GoFloat64, true},
		{"dayofweek.saturday", "dayofweek", "saturday", "7.0", GoFloat64, true},

		/* session properties */
		{"session.ismarket", "session", "ismarket", "true", GoBool, true},
		{"session.ispremarket", "session", "ispremarket", "false", GoBool, true},
		{"session.ispostmarket", "session", "ispostmarket", "false", GoBool, true},
		{"session.isfirstbar", "session", "isfirstbar", "session_isfirstbarSeries.GetCurrent() == 1.0", GoBool, true},
		{"session.islastbar", "session", "islastbar", "session_islastbarSeries.GetCurrent() == 1.0", GoBool, true},
		{"session.isfirstbar_regular", "session", "isfirstbar_regular", "session_isfirstbar_regularSeries.GetCurrent() == 1.0", GoBool, true},
		{"session.islastbar_regular", "session", "islastbar_regular", "session_islastbar_regularSeries.GetCurrent() == 1.0", GoBool, true},

		/* chart properties */
		{"chart.is_standard", "chart", "is_standard", "true", GoBool, true},
		{"chart.is_heikinashi", "chart", "is_heikinashi", "false", GoBool, true},
		{"chart.is_kagi", "chart", "is_kagi", "false", GoBool, true},
		{"chart.is_linebreak", "chart", "is_linebreak", "false", GoBool, true},
		{"chart.is_pnf", "chart", "is_pnf", "false", GoBool, true},
		{"chart.is_range", "chart", "is_range", "false", GoBool, true},
		{"chart.is_renko", "chart", "is_renko", "false", GoBool, true},
		{"chart.bg_color", "chart", "bg_color", `"#FFFFFF"`, GoString, true},
		{"chart.fg_color", "chart", "fg_color", `"#000000"`, GoString, true},
		{"chart.left_visible_bar_time", "chart", "left_visible_bar_time", "float64(ctx.Data[0].Time * 1000)", GoFloat64, true},
		{"chart.right_visible_bar_time", "chart", "right_visible_bar_time", "float64(ctx.Data[len(ctx.Data)-1].Time * 1000)", GoFloat64, true},

		/* dividends (all na) */
		{"dividends.future_amount", "dividends", "future_amount", "math.NaN()", GoFloat64, true},
		{"dividends.future_ex_date", "dividends", "future_ex_date", "math.NaN()", GoFloat64, true},
		{"dividends.future_pay_date", "dividends", "future_pay_date", "math.NaN()", GoFloat64, true},

		/* earnings (all na) */
		{"earnings.future_eps", "earnings", "future_eps", "math.NaN()", GoFloat64, true},
		{"earnings.future_period_end_time", "earnings", "future_period_end_time", "math.NaN()", GoFloat64, true},
		{"earnings.future_revenue", "earnings", "future_revenue", "math.NaN()", GoFloat64, true},
		{"earnings.future_time", "earnings", "future_time", "math.NaN()", GoFloat64, true},

		/* math constants */
		{"math.pi", "math", "pi", "math.Pi", GoFloat64, true},
		{"math.e", "math", "e", "math.E", GoFloat64, true},
		{"math.phi", "math", "phi", "1.618033988749895", GoFloat64, true},
		{"math.rphi", "math", "rphi", "0.618033988749895", GoFloat64, true},

		/* extend constants (drawing objects — numeric enum, GoFloat64) */
		{"extend.none", "extend", "none", "0.0", GoFloat64, true},
		{"extend.right", "extend", "right", "1.0", GoFloat64, true},
		{"extend.left", "extend", "left", "2.0", GoFloat64, true},
		{"extend.both", "extend", "both", "3.0", GoFloat64, true},
		{"extend.unknown", "extend", "unknown_prop", "", GoFloat64, false},

		/* line style constants */
		{"line.style_solid", "line", "style_solid", "0.0", GoFloat64, true},
		{"line.style_dashed", "line", "style_dashed", "1.0", GoFloat64, true},
		{"line.style_dotted", "line", "style_dotted", "2.0", GoFloat64, true},
		{"line.style_arrow_left", "line", "style_arrow_left", "3.0", GoFloat64, true},
		{"line.style_arrow_right", "line", "style_arrow_right", "4.0", GoFloat64, true},
		{"line.style_arrow_both", "line", "style_arrow_both", "5.0", GoFloat64, true},
		{"line.style_cross", "line", "style_cross", "6.0", GoFloat64, true},
		{"line.unknown", "line", "unknown_prop", "", GoFloat64, false},

		/* label style constants */
		{"label.style_none", "label", "style_none", "0.0", GoFloat64, true},
		{"label.style_xcross", "label", "style_xcross", "1.0", GoFloat64, true},
		{"label.style_cross", "label", "style_cross", "2.0", GoFloat64, true},
		{"label.style_triangleup", "label", "style_triangleup", "3.0", GoFloat64, true},
		{"label.style_triangledown", "label", "style_triangledown", "4.0", GoFloat64, true},
		{"label.style_arrowup", "label", "style_arrowup", "7.0", GoFloat64, true},
		{"label.style_arrowdown", "label", "style_arrowdown", "8.0", GoFloat64, true},
		{"label.style_label_up", "label", "style_label_up", "9.0", GoFloat64, true},
		{"label.style_diamond", "label", "style_diamond", "19.0", GoFloat64, true},
		{"label.unknown", "label", "unknown_prop", "", GoFloat64, false},

		/* xloc constants */
		{"xloc.bar_time", "xloc", "bar_time", "0.0", GoFloat64, true},
		{"xloc.bar_index", "xloc", "bar_index", "1.0", GoFloat64, true},
		{"xloc.unknown", "xloc", "unknown_prop", "", GoFloat64, false},

		/* size constants */
		{"size.auto", "size", "auto", "0.0", GoFloat64, true},
		{"size.tiny", "size", "tiny", "1.0", GoFloat64, true},
		{"size.small", "size", "small", "2.0", GoFloat64, true},
		{"size.normal", "size", "normal", "3.0", GoFloat64, true},
		{"size.large", "size", "large", "4.0", GoFloat64, true},
		{"size.huge", "size", "huge", "5.0", GoFloat64, true},
		{"size.unknown", "size", "unknown_prop", "", GoFloat64, false},

		/* shape constants */
		{"shape.xcross", "shape", "xcross", "0.0", GoFloat64, true},
		{"shape.cross", "shape", "cross", "1.0", GoFloat64, true},
		{"shape.triangleup", "shape", "triangleup", "3.0", GoFloat64, true},
		{"shape.arrowup", "shape", "arrowup", "8.0", GoFloat64, true},
		{"shape.diamond", "shape", "diamond", "10.0", GoFloat64, true},
		{"shape.square", "shape", "square", "11.0", GoFloat64, true},
		{"shape.unknown", "shape", "unknown_prop", "", GoFloat64, false},

		/* location constants */
		{"location.abovebar", "location", "abovebar", "0.0", GoFloat64, true},
		{"location.belowbar", "location", "belowbar", "1.0", GoFloat64, true},
		{"location.top", "location", "top", "2.0", GoFloat64, true},
		{"location.bottom", "location", "bottom", "3.0", GoFloat64, true},
		{"location.absolute", "location", "absolute", "4.0", GoFloat64, true},
		{"location.unknown", "location", "unknown_prop", "", GoFloat64, false},

		/* hline style constants */
		{"hline.style_solid", "hline", "style_solid", "0.0", GoFloat64, true},
		{"hline.style_dashed", "hline", "style_dashed", "1.0", GoFloat64, true},
		{"hline.style_dotted", "hline", "style_dotted", "2.0", GoFloat64, true},
		{"hline.unknown", "hline", "unknown_prop", "", GoFloat64, false},

		/* unknown namespace */
		{"unknown.prop", "unknown", "prop", "", GoFloat64, false},
		{"strategy.entry", "strategy", "entry", "", GoFloat64, false},
		{"strategy.unknown", "strategy", "unknown_prop", "", GoFloat64, false},
		{"ta.sma", "ta", "sma", "", GoFloat64, false},

		/* strategy flat constants */
		{"strategy.cash", "strategy", "cash", `"cash"`, GoString, true},
		{"strategy.fixed", "strategy", "fixed", `"fixed"`, GoString, true},
		{"strategy.percent_of_equity", "strategy", "percent_of_equity", `"percent_of_equity"`, GoString, true},
		{"strategy.long", "strategy", "long", `"long"`, GoString, true},
		{"strategy.short", "strategy", "short", `"short"`, GoString, true},
		{"strategy.both", "strategy", "both", `"both"`, GoString, true},
		{"strategy.account_currency", "strategy", "account_currency", `"USD"`, GoString, true},
		{"strategy.margin_liquidation_price", "strategy", "margin_liquidation_price", `math.NaN()`, GoFloat64, true},

		/* unknown property within valid namespace */
		{"barstate.unknown", "barstate", "unknown_prop", "", GoFloat64, false},
		{"timeframe.unknown", "timeframe", "unknown_prop", "", GoFloat64, false},
		{"syminfo.unknown", "syminfo", "unknown_prop", "", GoFloat64, false},
		{"dayofweek.unknown", "dayofweek", "unknown_prop", "", GoFloat64, false},
		{"session.unknown", "session", "unknown_prop", "", GoFloat64, false},
		{"chart.unknown", "chart", "unknown_prop", "", GoFloat64, false},
		{"dividends.unknown", "dividends", "unknown_prop", "", GoFloat64, false},
		{"earnings.unknown", "earnings", "unknown_prop", "", GoFloat64, false},
		{"math.unknown", "math", "unknown_prop", "", GoFloat64, false},

		/* case sensitivity */
		{"Barstate uppercase", "Barstate", "isfirst", "", GoFloat64, false},
		{"BARSTATE uppercase", "BARSTATE", "isfirst", "", GoFloat64, false},
		{"barstate.ISFIRST uppercase", "barstate", "ISFIRST", "", GoFloat64, false},
		{"TIMEFRAME uppercase", "TIMEFRAME", "period", "", GoFloat64, false},
		{"SYMINFO uppercase", "SYMINFO", "tickerid", "", GoFloat64, false},

		/* empty string */
		{"empty namespace", "", "prop", "", GoFloat64, false},
		{"empty property", "barstate", "", "", GoFloat64, false},
		{"both empty", "", "", "", GoFloat64, false},
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
			if res.GoType != tt.expectedType {
				t.Errorf("Resolve(%s, %s) GoType = %v, want %v", tt.namespace, tt.prop, res.GoType, tt.expectedType)
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
		{"dayofweek", true},
		{"session", true},
		{"chart", true},
		{"dividends", true},
		{"earnings", true},
		{"math", true},
		{"strategy", true},
		{"extend", true},
		{"line", true},
		{"label", true},
		{"xloc", true},
		{"size", true},
		{"shape", true},
		{"location", true},
		{"hline", true},
		{"unknown", false},
		{"ta", false},
		{"Barstate", false},
		{"TIMEFRAME", false},
		{"Extend", false},
		{"LINE", false},
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
		"barstate":  7,
		"timeframe": 10,
		"syminfo":   12,
		"dayofweek": 7,
		"session":   7,
		"chart":     11,
		"dividends": 3,
		"earnings":  4,
		"math":      4,
		"strategy":  8,
		"extend":    4,
		"line":      7,
		"label":     20,
		"xloc":      2,
		"size":      6,
		"shape":     12,
		"location":  5,
		"hline":     3,
	}

	namespacePropSets := map[string][]string{
		"barstate":  {"isfirst", "islast", "ishistory", "isrealtime", "isnew", "isconfirmed", "islastconfirmedhistory"},
		"timeframe": {"ismonthly", "isdaily", "isweekly", "isintraday", "isdwm", "isminutes", "isseconds", "isticks", "period", "multiplier"},
		"syminfo":   {"tickerid", "ticker", "timezone", "type", "prefix", "session", "currency", "basecurrency", "description", "pointvalue", "mintick", "volumetype"},
		"dayofweek": {"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"},
		"session":   {"ismarket", "ispremarket", "ispostmarket", "isfirstbar", "islastbar", "isfirstbar_regular", "islastbar_regular"},
		"chart":     {"is_standard", "is_heikinashi", "is_kagi", "is_linebreak", "is_pnf", "is_range", "is_renko", "bg_color", "fg_color", "left_visible_bar_time", "right_visible_bar_time"},
		"dividends": {"future_amount", "future_ex_date", "future_pay_date"},
		"earnings":  {"future_eps", "future_period_end_time", "future_revenue", "future_time"},
		"math":      {"pi", "e", "phi", "rphi"},
		"strategy":  {"cash", "fixed", "percent_of_equity", "long", "short", "both", "account_currency", "margin_liquidation_price"},
		"extend":    {"none", "right", "left", "both"},
		"line":      {"style_solid", "style_dashed", "style_dotted", "style_arrow_left", "style_arrow_right", "style_arrow_both", "style_cross"},
		"label": {
			"style_none", "style_xcross", "style_cross", "style_triangleup", "style_triangledown",
			"style_flag", "style_circle", "style_arrowup", "style_arrowdown", "style_label_up",
			"style_label_down", "style_label_left", "style_label_right", "style_label_lower_left",
			"style_label_lower_right", "style_label_upper_left", "style_label_upper_right",
			"style_label_center", "style_square", "style_diamond",
		},
		"xloc":     {"bar_time", "bar_index"},
		"size":     {"auto", "tiny", "small", "normal", "large", "huge"},
		"shape":    {"xcross", "cross", "circle", "triangleup", "triangledown", "flag", "labelup", "labeldown", "arrowup", "arrowdown", "diamond", "square"},
		"location": {"abovebar", "belowbar", "top", "bottom", "absolute"},
		"hline":    {"style_solid", "style_dashed", "style_dotted"},
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

// TestDrawingNamespaceConstants_AllAreFloat64 verifies that every drawing namespace
// constant resolves to GoFloat64, not GoString or GoBool. Drawing enum values are
// used in arithmetic expressions (plot, comparison), so the numeric contract is load-bearing.
func TestDrawingNamespaceConstants_AllAreFloat64(t *testing.T) {
	resolver := NewBuiltinNamespaceResolver()

	drawingProps := map[string][]string{
		"extend":   {"none", "right", "left", "both"},
		"line":     {"style_solid", "style_dashed", "style_dotted", "style_arrow_left", "style_arrow_right", "style_arrow_both", "style_cross"},
		"xloc":     {"bar_time", "bar_index"},
		"size":     {"auto", "tiny", "small", "normal", "large", "huge"},
		"shape":    {"xcross", "cross", "circle", "triangleup", "triangledown", "flag", "labelup", "labeldown", "arrowup", "arrowdown", "diamond", "square"},
		"location": {"abovebar", "belowbar", "top", "bottom", "absolute"},
		"hline":    {"style_solid", "style_dashed", "style_dotted"},
	}

	for ns, props := range drawingProps {
		for _, prop := range props {
			t.Run(ns+"."+prop, func(t *testing.T) {
				res, found := resolver.Resolve(ns, prop)
				if !found {
					t.Fatalf("Resolve(%s, %s) not found", ns, prop)
				}
				if res.GoType != GoFloat64 {
					t.Errorf("Resolve(%s, %s) GoType = %v, want GoFloat64", ns, prop, res.GoType)
				}
			})
		}
	}
}

// TestDrawingNamespaceConstants_Distinctness verifies that sibling constants within
// each namespace resolve to distinct float64 values, preventing silent comparison bugs
// (e.g. extend.right == extend.left would always be true).
func TestDrawingNamespaceConstants_Distinctness(t *testing.T) {
	resolver := NewBuiltinNamespaceResolver()

	drawingProps := map[string][]string{
		"extend":   {"none", "right", "left", "both"},
		"line":     {"style_solid", "style_dashed", "style_dotted", "style_arrow_left", "style_arrow_right", "style_arrow_both", "style_cross"},
		"xloc":     {"bar_time", "bar_index"},
		"size":     {"auto", "tiny", "small", "normal", "large", "huge"},
		"shape":    {"xcross", "cross", "circle", "triangleup", "triangledown", "flag", "labelup", "labeldown", "arrowup", "arrowdown", "diamond", "square"},
		"location": {"abovebar", "belowbar", "top", "bottom", "absolute"},
		"hline":    {"style_solid", "style_dashed", "style_dotted"},
	}

	for ns, props := range drawingProps {
		t.Run(ns, func(t *testing.T) {
			seen := make(map[string]string)
			for _, prop := range props {
				res, found := resolver.Resolve(ns, prop)
				if !found {
					t.Fatalf("Resolve(%s, %s) not found", ns, prop)
				}
				if prev, dup := seen[res.Code]; dup {
					t.Errorf("%s.%s and %s share the same code %q — constants must be distinct", ns, prop, prev, res.Code)
				}
				seen[res.Code] = prop
			}
		})
	}
}

// TestBuiltinNamespaceResolver_ResolveForArrow asserts the arrow-scope overrides
// for every namespace that has them and that all other namespaces fall through to
// the regular Resolve path unchanged.
//
// UDFs executing inside security() receive an *ArrowContext bound to the secondary
// series, so identity-sensitive builtins must resolve to that context's fields.
func TestBuiltinNamespaceResolver_ResolveForArrow(t *testing.T) {
	resolver := NewBuiltinNamespaceResolver()

	tests := []struct {
		name         string
		namespace    string
		prop         string
		expectedCode string
		expectedType GoValueType
		expectFound  bool
	}{
		{"syminfo.tickerid in arrow", "syminfo", "tickerid", "ctx.Symbol", GoString, true},
		{"syminfo.ticker in arrow", "syminfo", "ticker", "ctx.Symbol", GoString, true},
		{"syminfo.description in arrow", "syminfo", "description", "ctx.Symbol", GoString, true},

		{"syminfo.timezone in arrow", "syminfo", "timezone", "ctx.Timezone", GoString, true},
		{"syminfo.type in arrow", "syminfo", "type", `"stock"`, GoString, true},
		{"syminfo.currency in arrow", "syminfo", "currency", `"USD"`, GoString, true},
		{"syminfo.mintick in arrow", "syminfo", "mintick", "0.01", GoFloat64, true},
		{"syminfo.unknown in arrow", "syminfo", "no_such_prop", "", GoFloat64, false},

		{"session.ismarket in arrow", "session", "ismarket", "true", GoBool, true},
		{"session.ispremarket in arrow", "session", "ispremarket", "false", GoBool, true},
		{"session.ispostmarket in arrow", "session", "ispostmarket", "false", GoBool, true},

		{"dayofweek.sunday in arrow", "dayofweek", "sunday", "1.0", GoFloat64, true},
		{"barstate.isfirst in arrow", "barstate", "isfirst", "(ctx.BarIndex == 0)", GoBool, true},

		{"unknown namespace in arrow", "nosuchns", "prop", "", GoFloat64, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, found := resolver.ResolveForArrow(tt.namespace, tt.prop)
			if found != tt.expectFound {
				t.Fatalf("ResolveForArrow(%s, %s) found = %v, want %v", tt.namespace, tt.prop, found, tt.expectFound)
			}
			if !found {
				return
			}
			if res.Code != tt.expectedCode {
				t.Errorf("ResolveForArrow(%s, %s) code = %q, want %q", tt.namespace, tt.prop, res.Code, tt.expectedCode)
			}
			if res.GoType != tt.expectedType {
				t.Errorf("ResolveForArrow(%s, %s) GoType = %v, want %v", tt.namespace, tt.prop, res.GoType, tt.expectedType)
			}
		})
	}
}

// TestBuiltinNamespaceResolver_SyminfoArrowVsBarLoop asserts the contractual
// difference between the two resolution paths for identity-sensitive syminfo
// properties: bar-loop scope emits the package-level variable name (resolved
// once at startup from the -symbol flag), arrow scope emits ctx.Symbol
// (resolved per-call from the ArrowContext, which may be a secondary series).
func TestBuiltinNamespaceResolver_SyminfoArrowVsBarLoop(t *testing.T) {
	resolver := NewBuiltinNamespaceResolver()

	identityProps := []string{"tickerid", "ticker", "description"}
	for _, prop := range identityProps {
		prop := prop
		t.Run(prop, func(t *testing.T) {
			barLoop, found := resolver.Resolve("syminfo", prop)
			if !found {
				t.Fatalf("Resolve(syminfo, %s) not found", prop)
			}
			if barLoop.Code != "syminfo_tickerid" {
				t.Errorf("Resolve(syminfo, %s) = %q, want \"syminfo_tickerid\"", prop, barLoop.Code)
			}

			arrow, found := resolver.ResolveForArrow("syminfo", prop)
			if !found {
				t.Fatalf("ResolveForArrow(syminfo, %s) not found", prop)
			}
			if arrow.Code != "ctx.Symbol" {
				t.Errorf("ResolveForArrow(syminfo, %s) = %q, want \"ctx.Symbol\"", prop, arrow.Code)
			}

			if barLoop.Code == arrow.Code {
				t.Errorf(
					"bar-loop and arrow resolutions identical (%q) for syminfo.%s — arrow override not active",
					barLoop.Code, prop,
				)
			}
		})
	}
}
