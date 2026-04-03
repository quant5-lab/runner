package codegen

import "fmt"

type NamespaceResolution struct {
	Code   string
	GoType GoValueType
}

type BuiltinNamespaceResolver struct {
	dispatchers map[string]func(string) (NamespaceResolution, bool)
}

func NewBuiltinNamespaceResolver() *BuiltinNamespaceResolver {
	r := &BuiltinNamespaceResolver{}
	r.dispatchers = map[string]func(string) (NamespaceResolution, bool){
		"barstate":  r.resolveBarState,
		"timeframe": r.resolveTimeframe,
		"syminfo":   r.resolveSyminfo,
		"dayofweek": r.resolveDayOfWeek,
		"session":   r.resolveSession,
		"chart":     r.resolveChart,
		"dividends": r.resolveDividends,
		"earnings":  r.resolveEarnings,
		"math":      r.resolveMath,
		"strategy":  r.resolveStrategy,
	}
	return r
}

func (r *BuiltinNamespaceResolver) Resolve(obj, prop string) (NamespaceResolution, bool) {
	if dispatcher, exists := r.dispatchers[obj]; exists {
		return dispatcher(prop)
	}
	return NamespaceResolution{}, false
}

func (r *BuiltinNamespaceResolver) IsNamespace(obj string) bool {
	_, exists := r.dispatchers[obj]
	return exists
}

/* Overrides builtins referencing outer-scope variables; delegates ctx-compatible cases to Resolve */
func (r *BuiltinNamespaceResolver) ResolveForArrow(obj, prop string) (NamespaceResolution, bool) {
	switch obj {
	case "syminfo":
		return r.resolveSyminfoForArrow(prop)
	case "session":
		return r.resolveSessionForArrow(prop)
	default:
		return r.Resolve(obj, prop)
	}
}

func (r *BuiltinNamespaceResolver) resolveSyminfoForArrow(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "tickerid", "ticker":
		return NamespaceResolution{Code: "ctx.Symbol", GoType: GoString}, true
	case "description":
		return NamespaceResolution{Code: "ctx.Symbol", GoType: GoString}, true
	default:
		return r.resolveSyminfo(prop)
	}
}

func (r *BuiltinNamespaceResolver) resolveSessionForArrow(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "ismarket":
		return NamespaceResolution{Code: "true", GoType: GoBool}, true
	case "ispremarket":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "ispostmarket":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "isfirstbar":
		return NamespaceResolution{Code: arrowSessionBoolLookup(SessionIsFirstBarSeriesName), GoType: GoBool}, true
	case "islastbar":
		return NamespaceResolution{Code: arrowSessionBoolLookup(SessionIsLastBarSeriesName), GoType: GoBool}, true
	case "isfirstbar_regular":
		return NamespaceResolution{Code: arrowSessionBoolLookup(SessionIsFirstBarRegularSeriesName), GoType: GoBool}, true
	case "islastbar_regular":
		return NamespaceResolution{Code: arrowSessionBoolLookup(SessionIsLastBarRegularSeriesName), GoType: GoBool}, true
	default:
		return NamespaceResolution{}, false
	}
}

func arrowSessionBoolLookup(seriesName string) string {
	return fmt.Sprintf(
		`func() bool { if s, ok := ctx.LookupSeries(%q); ok { return s.GetCurrent() == 1.0 }; return false }()`,
		seriesName)
}

func (r *BuiltinNamespaceResolver) resolveBarState(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "isfirst":
		return NamespaceResolution{Code: "(ctx.BarIndex == 0)", GoType: GoBool}, true
	case "islast":
		return NamespaceResolution{Code: "(ctx.BarIndex == len(ctx.Data)-1)", GoType: GoBool}, true
	case "ishistory":
		return NamespaceResolution{Code: "true", GoType: GoBool}, true
	case "isrealtime":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "isnew":
		return NamespaceResolution{Code: "true", GoType: GoBool}, true
	case "isconfirmed":
		return NamespaceResolution{Code: "true", GoType: GoBool}, true
	case "islastconfirmedhistory":
		return NamespaceResolution{Code: "(ctx.BarIndex == len(ctx.Data)-1)", GoType: GoBool}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveTimeframe(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "ismonthly":
		return NamespaceResolution{Code: "ctx.IsMonthly", GoType: GoBool}, true
	case "isdaily":
		return NamespaceResolution{Code: "ctx.IsDaily", GoType: GoBool}, true
	case "isweekly":
		return NamespaceResolution{Code: "ctx.IsWeekly", GoType: GoBool}, true
	case "isintraday":
		return NamespaceResolution{Code: "ctx.IsIntraday", GoType: GoBool}, true
	case "isdwm":
		return NamespaceResolution{Code: "(ctx.IsDaily || ctx.IsWeekly || ctx.IsMonthly)", GoType: GoBool}, true
	case "isminutes":
		return NamespaceResolution{Code: "ctx.IsIntraday", GoType: GoBool}, true
	case "isseconds":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "isticks":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "multiplier":
		return NamespaceResolution{Code: "float64(context.TimeframeMultiplier(ctx.Timeframe))"}, true
	case "period":
		return NamespaceResolution{Code: "ctx.Timeframe", GoType: GoString}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveSyminfo(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "tickerid", "ticker":
		return NamespaceResolution{Code: "syminfo_tickerid", GoType: GoString}, true
	case "timezone":
		return NamespaceResolution{Code: "ctx.Timezone", GoType: GoString}, true
	case "type":
		return NamespaceResolution{Code: `"stock"`, GoType: GoString}, true
	case "prefix":
		return NamespaceResolution{Code: `""`, GoType: GoString}, true
	case "session":
		return NamespaceResolution{Code: `"regular"`, GoType: GoString}, true
	case "currency":
		return NamespaceResolution{Code: `"USD"`, GoType: GoString}, true
	case "basecurrency":
		return NamespaceResolution{Code: `""`, GoType: GoString}, true
	case "description":
		return NamespaceResolution{Code: "syminfo_tickerid", GoType: GoString}, true
	case "pointvalue":
		return NamespaceResolution{Code: "1.0"}, true
	case "mintick":
		return NamespaceResolution{Code: "0.01"}, true
	case "volumetype":
		return NamespaceResolution{Code: `"base"`, GoType: GoString}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveDayOfWeek(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "sunday":
		return NamespaceResolution{Code: "1.0"}, true
	case "monday":
		return NamespaceResolution{Code: "2.0"}, true
	case "tuesday":
		return NamespaceResolution{Code: "3.0"}, true
	case "wednesday":
		return NamespaceResolution{Code: "4.0"}, true
	case "thursday":
		return NamespaceResolution{Code: "5.0"}, true
	case "friday":
		return NamespaceResolution{Code: "6.0"}, true
	case "saturday":
		return NamespaceResolution{Code: "7.0"}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveSession(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "ismarket":
		return NamespaceResolution{Code: "true", GoType: GoBool}, true
	case "ispremarket":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "ispostmarket":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "isfirstbar":
		return NamespaceResolution{Code: SessionIsFirstBarSeriesName + ".GetCurrent() == 1.0", GoType: GoBool}, true
	case "islastbar":
		return NamespaceResolution{Code: SessionIsLastBarSeriesName + ".GetCurrent() == 1.0", GoType: GoBool}, true
	case "isfirstbar_regular":
		return NamespaceResolution{Code: SessionIsFirstBarRegularSeriesName + ".GetCurrent() == 1.0", GoType: GoBool}, true
	case "islastbar_regular":
		return NamespaceResolution{Code: SessionIsLastBarRegularSeriesName + ".GetCurrent() == 1.0", GoType: GoBool}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveChart(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "is_standard":
		return NamespaceResolution{Code: "true", GoType: GoBool}, true
	case "is_heikinashi":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "is_kagi":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "is_linebreak":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "is_pnf":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "is_range":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "is_renko":
		return NamespaceResolution{Code: "false", GoType: GoBool}, true
	case "bg_color":
		return NamespaceResolution{Code: `"#FFFFFF"`, GoType: GoString}, true
	case "fg_color":
		return NamespaceResolution{Code: `"#000000"`, GoType: GoString}, true
	case "left_visible_bar_time":
		return NamespaceResolution{Code: "float64(ctx.Data[0].Time * 1000)"}, true
	case "right_visible_bar_time":
		return NamespaceResolution{Code: "float64(ctx.Data[len(ctx.Data)-1].Time * 1000)"}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveDividends(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "future_amount", "future_ex_date", "future_pay_date":
		return NamespaceResolution{Code: "math.NaN()"}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveEarnings(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "future_eps", "future_period_end_time", "future_revenue", "future_time":
		return NamespaceResolution{Code: "math.NaN()"}, true
	default:
		return NamespaceResolution{}, false
	}
}

func (r *BuiltinNamespaceResolver) resolveMath(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "pi":
		return NamespaceResolution{Code: "math.Pi"}, true
	case "e":
		return NamespaceResolution{Code: "math.E"}, true
	case "phi":
		return NamespaceResolution{Code: "1.618033988749895"}, true
	case "rphi":
		return NamespaceResolution{Code: "0.618033988749895"}, true
	default:
		return NamespaceResolution{}, false
	}
}

/* resolveStrategy handles strategy.* flat constants; strategy.entry/close are call sites, not namespaced constants. */
func (r *BuiltinNamespaceResolver) resolveStrategy(prop string) (NamespaceResolution, bool) {
	switch prop {
	case "cash":
		return NamespaceResolution{Code: `"cash"`, GoType: GoString}, true
	case "fixed":
		return NamespaceResolution{Code: `"fixed"`, GoType: GoString}, true
	case "percent_of_equity":
		return NamespaceResolution{Code: `"percent_of_equity"`, GoType: GoString}, true
	case "long":
		return NamespaceResolution{Code: `"long"`, GoType: GoString}, true
	case "short":
		return NamespaceResolution{Code: `"short"`, GoType: GoString}, true
	case "both":
		return NamespaceResolution{Code: `"both"`, GoType: GoString}, true
	case "account_currency":
		return NamespaceResolution{Code: `"USD"`, GoType: GoString}, true
	case "margin_liquidation_price":
		return NamespaceResolution{Code: "math.NaN()"}, true
	default:
		return NamespaceResolution{}, false
	}
}
