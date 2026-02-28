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
		{"unknown", false},
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
